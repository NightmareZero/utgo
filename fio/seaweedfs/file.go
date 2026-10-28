package seaweedfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/minio/minio-go/v7"
)

var _ fio.IFile = &SeaweedFSFile{}
var _ fio.IFileWriterAt = &SeaweedFSFile{}

// SeaweedFSFile 实现 fio.IFile。
// 写语义与 minio 后端一致：multipart upload 顺序追加；Close 时完成上传。
// 读取语义：惰性 GetObject，EOF 时关闭流。
// 额外实现 fio.IFileWriterAt：
//   - Write（multipart 覆盖创建）与 WriteAt（位置写）语义不同，互斥不可并发使用
//   - WriteAt 仅支持追加（off == 当前文件大小）；部分写入（off < fileSize）返回 ErrPartialWriteNotSupported
type SeaweedFSFile struct {
	fs   *SeaweedFSFileBucket
	name string
	rb   io.ReadCloser
	// readTried 标记是否曾尝试 Read。用于 Close 时区分"纯写入句柄"与"读取句柄"。
	readTried bool
	// multipart 上传相关字段
	mu       sync.Mutex
	errStore error
	uploadID string
	partSize int64
	buf      *bytes.Buffer
	parts    []minio.CompletePart
	partNum  int
	closed   bool
}

// Read implements fio.IFile
func (f *SeaweedFSFile) Read(p []byte) (n int, err error) {
	f.readTried = true
	// 首次读取时打开对象流
	if f.rb == nil {
		f.rb, err = f.fs.Open(f.name)
		if err != nil {
			return 0, err
		}
	}
	n, err = f.rb.Read(p)
	// 读取到 EOF 时自动关闭流
	if err == io.EOF {
		f.rb.Close()
		f.rb = nil
	}
	return n, err
}

// Write implements fio.IFile
// 使用 multipart upload 实现顺序追加写。缓冲满 partSize 时上传一个 part。
func (f *SeaweedFSFile) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.errStore != nil {
		return 0, f.errStore
	}

	if f.closed {
		return 0, io.ErrClosedPipe
	}

	// 初始化 multipart 状态（首次写入）
	if f.uploadID == "" {
		// default part size 5MB（S3 multipart 最小分片）
		f.partSize = 5 * 1024 * 1024
		f.buf = bytes.NewBuffer(nil)
		f.partNum = 1
		f.parts = nil

		core := minio.Core{Client: f.fs.cl}
		uploadID, err := core.NewMultipartUpload(context.Background(), f.fs.bucket, f.name, minio.PutObjectOptions{})
		if err != nil {
			f.errStore = err
			return 0, err
		}
		f.uploadID = uploadID
	}

	// 写入缓冲
	written, err := f.buf.Write(p)
	if err != nil {
		f.errStore = err
		return 0, err
	}

	// 缓冲达到 partSize 时上传 part
	core := minio.Core{Client: f.fs.cl}
	for int64(f.buf.Len()) >= f.partSize {
		partBytes := make([]byte, f.partSize)
		_, _ = f.buf.Read(partBytes)

		partReader := bytes.NewReader(partBytes)
		objPart, err := core.PutObjectPart(context.Background(), f.fs.bucket, f.name, f.uploadID, f.partNum, partReader, int64(len(partBytes)), minio.PutObjectPartOptions{})
		if err != nil {
			// 出错时中止上传，清理状态
			_ = core.AbortMultipartUpload(context.Background(), f.fs.bucket, f.name, f.uploadID)
			f.errStore = err
			f.uploadID = ""
			f.buf = nil
			f.parts = nil
			f.partNum = 0
			return 0, err
		}
		f.parts = append(f.parts, minio.CompletePart{PartNumber: f.partNum, ETag: objPart.ETag})
		f.partNum++
	}

	return written, nil
}

// Close implements fio.IFile
// 幂等：重复调用安全。先关闭读取流，再处理写入：
//   - 若曾发生 multipart 写入：完成上传（含剩余未满 part）
//   - 若从未读也从未写：按文件系统语义创建空对象（等价于 O_CREATE|O_TRUNC）
func (f *SeaweedFSFile) Close() (err error) {
	// 幂等检查：已关闭直接返回
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return nil
	}
	f.closed = true
	uploadID := f.uploadID
	buf := f.buf
	parts := f.parts
	partNum := f.partNum
	readTried := f.readTried
	f.mu.Unlock()

	// 处理读取流
	if f.rb != nil {
		err = f.rb.Close()
		f.rb = nil
	}

	// 无 multipart 上传
	if uploadID == "" {
		// 从未读也从未写 → 创建空对象，保持与文件系统语义一致
		if !readTried {
			_, err2 := f.fs.cl.PutObject(context.Background(), f.fs.bucket, f.name, bytes.NewReader(nil), 0, minio.PutObjectOptions{})
			if err2 != nil {
				return err2
			}
		}
		return err
	}

	core := minio.Core{Client: f.fs.cl}

	// 上传剩余未满 part
	if buf != nil && buf.Len() > 0 {
		rem := make([]byte, buf.Len())
		_, _ = buf.Read(rem)
		partReader := bytes.NewReader(rem)
		objPart, err2 := core.PutObjectPart(context.Background(), f.fs.bucket, f.name, uploadID, partNum, partReader, int64(len(rem)), minio.PutObjectPartOptions{})
		if err2 != nil {
			_ = core.AbortMultipartUpload(context.Background(), f.fs.bucket, f.name, uploadID)
			return err2
		}
		parts = append(parts, minio.CompletePart{PartNumber: partNum, ETag: objPart.ETag})
	}

	// 完成上传
	_, err = core.CompleteMultipartUpload(context.Background(), f.fs.bucket, f.name, uploadID, parts, minio.PutObjectOptions{})
	if err != nil {
		_ = core.AbortMultipartUpload(context.Background(), f.fs.bucket, f.name, uploadID)
		return err
	}

	// 清理状态
	f.mu.Lock()
	f.uploadID = ""
	f.buf = nil
	f.parts = nil
	f.partNum = 0
	f.mu.Unlock()

	return nil
}

// WriteAt implements fio.IFileWriterAt（io.WriterAt 语义）。
//
// 行为：
//   - 文件不存在 + off==0: PutObject 创建
//   - off == 当前文件大小: 追加（走 Filer ?op=append，小文件自动回退 read-modify-write）
//   - off <  当前文件大小: 返回 ErrPartialWriteNotSupported（待实现）
//   - off >  当前文件大小: 不支持（写洞）
//
// 与 Write（multipart 覆盖）互斥：若已有进行中的 multipart 上传，返回错误。
func (f *SeaweedFSFile) WriteAt(p []byte, off int64) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, io.ErrClosedPipe
	}
	if len(p) == 0 {
		return 0, nil
	}
	// 与 multipart Write 互斥
	if f.uploadID != "" {
		return 0, errors.New("seaweedfs: WriteAt conflicts with in-progress multipart Write")
	}

	ctx := context.Background()

	// 取当前文件大小
	stat, serr := f.fs.Stat(f.name)
	fileExists := serr == nil
	var fileSize int64
	if fileExists {
		fileSize = stat.Size()
	}

	if !fileExists {
		if off != 0 {
			return 0, fmt.Errorf("seaweedfs: WriteAt at offset %d on non-existent file %q", off, f.name)
		}
		// off==0 + 文件不存在 → 创建
		_, err = f.fs.cl.PutObject(ctx, f.fs.bucket, f.name, bytes.NewReader(p), int64(len(p)), minio.PutObjectOptions{})
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}

	if off == fileSize {
		return f.doAppend(ctx, p)
	}
	if off < fileSize {
		return 0, ErrPartialWriteNotSupported
	}
	return 0, fmt.Errorf("seaweedfs: WriteAt beyond EOF (off=%d, size=%d) not supported", off, fileSize)
}

// doAppend 执行追加写。优先走 Filer ?op=append；命中"小文件不支持"限制或 Filer 不可用时回退 RMW。
func (f *SeaweedFSFile) doAppend(ctx context.Context, p []byte) (int, error) {
	if f.fs.fc.available() {
		isSmall, err := f.fs.fc.append(ctx, f.fs.bucket, f.name, p)
		switch {
		case err == nil:
			return len(p), nil
		case isSmall:
			// 命中小文件限制，回退
		default:
			return 0, err
		}
	}
	return f.appendViaRMW(ctx, p)
}

// appendViaRMW 用 read-modify-write 实现追加：读全文件 → 拼接新数据 → PutObject 覆盖。
func (f *SeaweedFSFile) appendViaRMW(ctx context.Context, p []byte) (int, error) {
	var existing []byte
	if f.fs.fc.available() {
		b, err := f.fs.fc.readAll(ctx, f.fs.bucket, f.name)
		if err != nil {
			return 0, fmt.Errorf("seaweedfs: rmw read failed: %w", err)
		}
		existing = b
	} else {
		// Filer 不可用时退化用 S3 GetObject 读取
		rc, err := f.fs.Open(f.name)
		if err != nil {
			return 0, fmt.Errorf("seaweedfs: rmw read failed: %w", err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return 0, fmt.Errorf("seaweedfs: rmw read failed: %w", err)
		}
		existing = b
	}

	combined := append(existing, p...)
	_, err := f.fs.cl.PutObject(ctx, f.fs.bucket, f.name, bytes.NewReader(combined), int64(len(combined)), minio.PutObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("seaweedfs: rmw write failed: %w", err)
	}
	return len(p), nil
}

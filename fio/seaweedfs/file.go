package seaweedfs

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/minio/minio-go/v7"
)

var _ fio.IFile = &SeaweedFSFile{}

// SeaweedFSFile 实现 fio.IFile。
// 写语义与 minio 后端一致：multipart upload 顺序追加；Close 时完成上传。
// 读取语义：惰性 GetObject，EOF 时关闭流。
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

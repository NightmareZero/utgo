package miniofs

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/minio/minio-go/v7"
)

var _ fio.IFile = &MinioFile{}

type MinioFile struct {
	fs   *MinioFileBucket
	name string
	rb   io.ReadCloser
	// multipart 上传相关字段
	mu       sync.Mutex
	errMinio error
	uploadID string
	partSize int64
	buf      *bytes.Buffer
	parts    []minio.CompletePart
	partNum  int
	closed   bool
}

// Read implements fio.IFile
func (m *MinioFile) Read(p []byte) (n int, err error) {
	// 首次读取时打开对象流
	if m.rb == nil {
		m.rb, err = m.fs.Open(m.name)
		if err != nil {
			return 0, err
		}
	}
	n, err = m.rb.Read(p)
	// 如果读取到 EOF，自动关闭流
	if err == io.EOF {
		m.rb.Close()
		m.rb = nil
	}
	return n, err
}

// Write implements fio.IFile
func (m *MinioFile) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMinio != nil {
		return 0, m.errMinio
	}

	if m.closed {
		return 0, io.ErrClosedPipe
	}

	// 初始化 multipart 状态（首次写入）
	if m.uploadID == "" {
		// default part size 5MB
		m.partSize = 5 * 1024 * 1024
		m.buf = bytes.NewBuffer(nil)
		m.partNum = 1
		m.parts = nil

		core := minio.Core{Client: m.fs.cl}
		uploadID, err := core.NewMultipartUpload(context.Background(), m.fs.bucket, m.name, minio.PutObjectOptions{})
		if err != nil {
			m.errMinio = err
			return 0, err
		}
		m.uploadID = uploadID
	}

	// 写入到缓冲
	written, err := m.buf.Write(p)
	if err != nil {
		m.errMinio = err
		return 0, err
	}

	// 当缓冲达到 partSize 时上传该 part
	core := minio.Core{Client: m.fs.cl}
	for int64(m.buf.Len()) >= m.partSize {
		partBytes := make([]byte, m.partSize)
		_, _ = m.buf.Read(partBytes)

		// 上传 part
		partReader := bytes.NewReader(partBytes)
		objPart, err := core.PutObjectPart(context.Background(), m.fs.bucket, m.name, m.uploadID, m.partNum, partReader, int64(len(partBytes)), minio.PutObjectPartOptions{})
		if err != nil {
			// abort multipart
			_ = core.AbortMultipartUpload(context.Background(), m.fs.bucket, m.name, m.uploadID)
			m.errMinio = err
			// clear upload state so future operations don't reuse this uploadID
			m.uploadID = ""
			m.buf = nil
			m.parts = nil
			m.partNum = 0
			return 0, err
		}
		// 记录 part
		m.parts = append(m.parts, minio.CompletePart{PartNumber: m.partNum, ETag: objPart.ETag})
		m.partNum++
	}

	return written, nil
}

// Close implements fio.IFile
// Close closes the file, ensuring that both read and write streams are properly handled.
// It waits for the upload to complete and checks for any errors that occurred during the upload.
func (m *MinioFile) Close() (err error) {
	// 先处理读取流
	if m.rb != nil {
		err = m.rb.Close()
		m.rb = nil
	}

	// 再处理写入：如果没有 multipart，则直接返回
	// mark closed to prevent further writes
	m.mu.Lock()
	m.closed = true
	uploadID := m.uploadID
	buf := m.buf
	parts := m.parts
	partNum := m.partNum
	m.mu.Unlock()

	if uploadID == "" {
		return err
	}

	core := minio.Core{Client: m.fs.cl}

	// 上传剩余未满 part
	if buf != nil && buf.Len() > 0 {
		rem := make([]byte, buf.Len())
		_, _ = buf.Read(rem)
		partReader := bytes.NewReader(rem)
		objPart, err2 := core.PutObjectPart(context.Background(), m.fs.bucket, m.name, uploadID, partNum, partReader, int64(len(rem)), minio.PutObjectPartOptions{})
		if err2 != nil {
			_ = core.AbortMultipartUpload(context.Background(), m.fs.bucket, m.name, uploadID)
			return err2
		}
		parts = append(parts, minio.CompletePart{PartNumber: partNum, ETag: objPart.ETag})
	}

	// Complete multipart
	_, err = core.CompleteMultipartUpload(context.Background(), m.fs.bucket, m.name, uploadID, parts, minio.PutObjectOptions{})
	if err != nil {
		_ = core.AbortMultipartUpload(context.Background(), m.fs.bucket, m.name, uploadID)
		return err
	}

	// cleanup state
	m.mu.Lock()
	m.uploadID = ""
	m.buf = nil
	m.parts = nil
	m.partNum = 0
	m.mu.Unlock()

	return nil
}

// multipart implementation complete

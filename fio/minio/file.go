package miniofs

import (
	"context"
	"io"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/utilp"
	"github.com/minio/minio-go/v7"
)

var _ fio.IFile = &MinioFile{}

type MinioFile struct {
	fs   *MinioFileBucket
	name string
	rb   io.ReadCloser

	errMinio error
	pr       *io.PipeReader
	pw       *io.PipeWriter
}

// Read implements fio.IFile
func (m *MinioFile) Read(p []byte) (n int, err error) {
	rc, err := m.fs.Open(m.name)
	if err != nil {
		return
	}
	defer rc.Close()
	return rc.Read(p)
}

// Write implements fio.IFile
func (m *MinioFile) Write(p []byte) (n int, err error) {
	if m.pr == nil {
		m.pr, m.pw = io.Pipe()
		go utilp.Try(func() {
			// 由于 objectSize=-1 , 所以PutObject会一直等待pr的数据, 直到pr被关闭
			_, m.errMinio = m.fs.cl.PutObject(context.Background(), m.fs.bucket, m.name, m.pr, -1, minio.PutObjectOptions{})
			m.pr.Close()
		})

	}
	if m.errMinio != nil {
		return 0, m.errMinio
	}

	var written int
	written, err = m.pw.Write(p)
	// print("written:", written, "err:", err, "")

	return int(written), err
}

// Close implements fio.IFile
func (m *MinioFile) Close() (err error) {
	defer utilp.Try(func() {
		if m.rb != nil {
			err = m.rb.Close()
			m.rb = nil
		}
	})

	defer utilp.Try(func() {
		if m.pr != nil {
			m.pw.Close()
			m.pw = nil
		}
	})
	return
}

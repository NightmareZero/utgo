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

	errMinio   error
	pr         *io.PipeReader
	pw         *io.PipeWriter
	uploadDone chan struct{} // 用于等待上传完成
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
	if m.pr == nil {
		m.pr, m.pw = io.Pipe()
		m.uploadDone = make(chan struct{})
		go utilp.Try(func() {
			defer close(m.uploadDone) // 上传完成后关闭通道
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
// Close closes the file, ensuring that both read and write streams are properly handled.
// It waits for the upload to complete and checks for any errors that occurred during the upload.
func (m *MinioFile) Close() (err error) {
	// 先处理读取流
	if m.rb != nil {
		err = m.rb.Close()
		m.rb = nil
	}

	// 再处理写入流
	if m.pw != nil {
		m.pw.Close()
		m.pw = nil

		// 等待 Minio 上传完成
		if m.uploadDone != nil {
			<-m.uploadDone
			// 检查上传是否有错误，优先返回上传错误
			if m.errMinio != nil {
				err = m.errMinio
			}
		}
	}
	return
}

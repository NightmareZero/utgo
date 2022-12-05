package fio

import (
	"io"
	"os"
)

var (
	// 默认使用本地文件系统(以后需要被 minio相关实现覆盖)
	FileSystem IFileSystem = &_localFileSystem{
		basePath: "/files/",
	}

	// 临时文件默认使用本地文件系统
	TmpFileSystem IFileSystem = &_localFileSystem{
		basePath: "/tmp/",
	}
)

type IFileSystem interface {
	Open(name string) (io.ReadCloser, error)
	OpenFile(name string) (IFile, error)
	Stat(name string) (os.FileInfo, error)
	Remove(name string) error
}

type IFile interface {
	io.Reader
	io.Writer
	io.Closer
}

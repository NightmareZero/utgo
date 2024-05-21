package fio

import (
	"io"
	"os"
)

var FileSystem IFileSystem

// 默认使用本地文件系统(以后需要被 minio相关实现覆盖)
type IFileSystem interface {
	IsOnline() bool
	Bucket(name string) (IFileBucket, error)
}

type IFileBucket interface {
	Bucket() string
	Info() (stat BucketStat, err error)
	SetConfig(conf BucketConfig) (err error)
	List(path string) ([]os.FileInfo, error)
	Open(name string) (io.ReadSeekCloser, error)
	OpenFile(name string) (IFile, error)
	Stat(name string) (IFileStat, error)
	Remove(name string) error
}

type IFile interface {
	io.Reader
	io.Writer
	io.Closer
}

type IFileStat interface {
	os.FileInfo
	Sha1() string
}

type BucketStat struct {
	Quota int64  // 配额 (单位mb)
	Vers  bool   // 是否开启版本控制
	Size  uint64 // 文件大小 (字节)
	Files uint64 // 文件数量
}

type BucketConfig struct {
	Quota *int64 // 配额 (单位mb)
	Vers  *bool  // 是否开启版本控制
}

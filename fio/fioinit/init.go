package fioinit

import (
	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/fio/localfs"
	miniofs "github.com/NightmareZero/nzgoutil/fio/minio"
	"github.com/NightmareZero/nzgoutil/fio/seaweedfs"
)

func Init(cnf fio.MinIO) error {
	if len(cnf.Endpoint) > 0 {
		return initMinio(cnf)
	} else {
		return initLocal()
	}
}

func initLocal() error {
	fio.FileSystem = &localfs.LocalFileSystem{
		BasePath: "/files/",
	}
	return nil
}

func initMinio(conf fio.MinIO) (err error) {
	fio.FileSystem, err = miniofs.NewMinioFileSystem(conf)
	return err
}

// InitSeaweedFS 初始化 SeaweedFS 后端，与 localfs / minio 并列。
// 调用方显式选择该后端，不与 Init(fio.MinIO) 的二选一逻辑冲突。
func InitSeaweedFS(conf fio.SeaweedFS) (err error) {
	fio.FileSystem, err = seaweedfs.NewSeaweedFSFileSystem(conf)
	return err
}

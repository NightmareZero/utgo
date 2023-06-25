package fioinit

import (
	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/fio/localfs"
	miniofs "github.com/NightmareZero/nzgoutil/fio/minio"
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

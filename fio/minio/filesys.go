package miniofs

import (
	"strings"
	"sync"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/minio/madmin-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioFileSystem(conf fio.MinIO) (ii *MinIOFileSystem, err error) {
	var ifs MinIOFileSystem

	// 创建管理连接
	ifs.mc, err = madmin.New(conf.Endpoint, conf.Access, conf.Secret, conf.Secure)
	if err != nil {
		return nil, err
	}

	// 创建连接
	ifs.cl, err = minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.Access, conf.Secret, ""),
		Secure: conf.Secure,
	})
	ifs.buckets = make(map[string]*MinioFileBucket)

	return &ifs, err
}

var _ fio.IFileSystem = &MinIOFileSystem{}

type MinIOFileSystem struct {
	cl *minio.Client
	mc *madmin.AdminClient

	lock    sync.Mutex
	buckets map[string]*MinioFileBucket
}

// IsOnline implements fio.IFileSystem.
func (m *MinIOFileSystem) IsOnline() bool {
	return m.cl.IsOnline()
}

// Bucket implements fio.IFileSystem.
func (m *MinIOFileSystem) Bucket(name string) (fio.IFileBucket, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	name = strings.ToLower(name)

	if b, ok := m.buckets[name]; ok {
		// 存在则返回
		return b, nil
	} else {
		// 不存则创建后返回
		b = &MinioFileBucket{
			cl:     m.cl,
			mc:     m.mc,
			bucket: name,
		}
		// 初始化
		err := b.Init()
		if err != nil {
			return nil, err
		}
		// 保存
		m.buckets[name] = b
		return b, nil
	}

}

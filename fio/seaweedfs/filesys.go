package seaweedfs

import (
	"errors"
	"strings"
	"sync"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrNotSupported 表示当前 Tier 1 实现尚未对接 SeaweedFS 原生管理 API 的操作。
// 涉及 ClusterInfo / Bucket.Info / Bucket.SetConfig，详见 tmp/todo.md 的 Tier 2 待办。
var ErrNotSupported = errors.New("seaweedfs: operation not supported in Tier 1 implementation")

// ErrPartialWriteNotSupported 表示 WriteAt 的部分写入（off < 当前文件大小）尚未实现。
// SeaweedFS HTTP 路径无法为 chunk 指定写入 offset，部分写入需走 read-modify-write 或 gRPC，留待后续。
var ErrPartialWriteNotSupported = errors.New("seaweedfs: partial write (off < fileSize) not supported yet")

// NewSeaweedFSFileSystem 创建基于 SeaweedFS S3 网关的文件系统后端。
// SeaweedFS 的 S3 网关与标准 S3 协议兼容，可直接复用 minio-go 客户端。
// 注意：SeaweedFS 没有 madmin 等价物，管理层能力（ClusterInfo/Bucket.Info/SetConfig）受限。
func NewSeaweedFSFileSystem(conf fio.SeaweedFS) (*SeaweedFSFileSystem, error) {
	cl, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.Access, conf.Secret, ""),
		Secure: conf.Secure,
	})
	if err != nil {
		return nil, err
	}
	return &SeaweedFSFileSystem{
		cl:      cl,
		fc:      newFilerClient(conf.FilerURL, conf.Access, conf.Secret),
		buckets: make(map[string]*SeaweedFSFileBucket),
	}, nil
}

var _ fio.IFileSystem = &SeaweedFSFileSystem{}

type SeaweedFSFileSystem struct {
	cl      *minio.Client
	fc      *filerClient
	lock    sync.Mutex
	buckets map[string]*SeaweedFSFileBucket
}

// IsOnline implements fio.IFileSystem.
func (s *SeaweedFSFileSystem) IsOnline() bool {
	return s.cl.IsOnline()
}

// Bucket implements fio.IFileSystem.
func (s *SeaweedFSFileSystem) Bucket(name string, create bool) (fio.IFileBucket, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	name = strings.ToLower(name)

	if b, ok := s.buckets[name]; ok {
		// 已缓存
		return b, nil
	}

	// 新建并初始化
	b := &SeaweedFSFileBucket{
		fs:     s,
		cl:     s.cl,
		fc:     s.fc,
		bucket: name,
	}
	if err := b.Init(create); err != nil {
		return nil, err
	}
	s.buckets[name] = b
	return b, nil
}

// removeBucketFromCache 从缓存中移除指定桶（供 Destroy 调用，避免返回失效对象）。
func (s *SeaweedFSFileSystem) removeBucketFromCache(name string) {
	s.lock.Lock()
	delete(s.buckets, name)
	s.lock.Unlock()
}

// ClusterInfo 由 cluster.go 实现（当前为空壳）。

package seaweedfs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/utilp"
	"github.com/minio/minio-go/v7"
)

var _ fio.IFileBucket = &SeaweedFSFileBucket{}

type SeaweedFSFileBucket struct {
	fs     *SeaweedFSFileSystem
	cl     *minio.Client
	fc     *filerClient
	bucket string
}

// MergeFile implements fio.IFileBucket.
// 使用 S3 ComposeObject 实现服务端合并。SeaweedFS S3 网关支持该操作。
func (s *SeaweedFSFileBucket) MergeFile(ctx context.Context, dst fio.MergeOption, srcs ...fio.MergeOption) (stat fio.IFileStat, err error) {
	if len(srcs) == 0 {
		return stat, fmt.Errorf("no source file")
	}

	var (
		dstOpt = minio.CopyDestOptions{
			Bucket: s.bucket,
			Object: dst.Path,
		}
		srcOpts = make([]minio.CopySrcOptions, 0, len(srcs))
	)
	for _, src := range srcs {
		srcOpts = append(srcOpts, minio.CopySrcOptions{
			Bucket: s.bucket,
			Object: src.Path,
		})
	}

	res, err := s.cl.ComposeObject(ctx, dstOpt, srcOpts...)
	if err != nil {
		return stat, fmt.Errorf("MergeFile failed: %w", err)
	}

	var statF = &FileStatHandler{}
	if err := utilp.GobConv(res, statF); err != nil {
		return nil, fmt.Errorf("MergeFile convert stat failed: %w", err)
	}

	return statF, nil
}

// Bucket implements fio.IFileBucket.
func (s *SeaweedFSFileBucket) Bucket() string {
	return s.bucket
}

// Init 检查/创建存储桶。非接口方法，由 SeaweedFSFileSystem.Bucket 内部调用。
func (s *SeaweedFSFileBucket) Init(create bool) (err error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	b, err := s.cl.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !b {
		if create {
			// 不存在则创建。SeaweedFS 无 madmin，无法设置默认配额。
			err = s.cl.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
			if err != nil {
				return err
			}
		} else {
			return fio.ErrBucketNotFound
		}
	}
	return nil
}

// Info implements fio.IFileBucket.
// TODO(Tier 2): SeaweedFS 配额/版本控制/桶大小需要对接原生 HTTP API，暂未实现。
// 当前明确返回 ErrNotSupported，避免调用方误判零值为有效数据。
func (s *SeaweedFSFileBucket) Info() (stat fio.BucketStat, err error) {
	return stat, ErrNotSupported
}

// SetConfig implements fio.IFileBucket.
// TODO(Tier 2): SeaweedFS 配额/版本控制设置需要对接原生 HTTP API，暂未实现。
// 当前明确返回 ErrNotSupported，避免调用方误以为配置已生效。
func (s *SeaweedFSFileBucket) SetConfig(conf fio.BucketConfig) (err error) {
	return ErrNotSupported
}

// OpenFile implements fio.IFileBucket.
// 打开文件（虚拟句柄，支持读写）。
func (s *SeaweedFSFileBucket) OpenFile(name string) (fio.IFile, error) {
	return &SeaweedFSFile{name: name, fs: s}, nil
}

// Open implements fio.IFileBucket.
// 打开文件（只读管道）。
func (s *SeaweedFSFileBucket) Open(name string) (io.ReadSeekCloser, error) {
	o, err := s.cl.GetObject(context.Background(), s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Stat implements fio.IFileBucket.
func (s *SeaweedFSFileBucket) Stat(name string) (fio.IFileStat, error) {
	name = strings.TrimPrefix(name, "/")

	oi, err := s.cl.StatObject(context.Background(), s.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		err2, ok := err.(minio.ErrorResponse)
		// 兼容标准文件系统语义
		if ok && err2.Code == "NoSuchKey" {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	return &FileStatHandler{oi}, nil
}

// Remove implements fio.IFileBucket.
func (s *SeaweedFSFileBucket) Remove(name string) error {
	ctx, cf := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cf()
	return s.cl.RemoveObject(ctx, s.bucket, name, minio.RemoveObjectOptions{})
}

// List implements fio.IFileBucket.
func (s *SeaweedFSFileBucket) List(path string) (res []fs.FileInfo, err error) {
	ctx, cf := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cf()

	// SeaweedFS S3 路径前缀同样不支持前导 "/"
	path = strings.TrimPrefix(path, "/")

	rc := s.cl.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: false,
	})
	for oi := range rc {
		if oi.Err != nil {
			err = oi.Err
			return
		}
		res = append(res, &FileStatHandler{oi})
	}
	return
}

// Destroy implements fio.IFileBucket.
func (s *SeaweedFSFileBucket) Destroy() error {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	err := s.cl.RemoveBucket(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("error on remove bucket: %v", err)
	}

	// 从 filesystem 缓存移除，避免后续 Bucket() 返回失效对象
	if s.fs != nil {
		s.fs.removeBucketFromCache(s.bucket)
	}

	// 清理引用
	s.cl = nil
	s.fs = nil
	return nil
}

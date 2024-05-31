package miniofs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/util"
	"github.com/minio/madmin-go/v2"
	"github.com/minio/minio-go/v7"
)

var _ fio.IFileBucket = &MinioFileBucket{}

type MinioFileBucket struct {
	cl     *minio.Client
	mc     *madmin.AdminClient
	bucket string
}

// MergeFile implements fio.IFileBucket.
func (m *MinioFileBucket) MergeFile(ctx context.Context, dst fio.MergeOption, srcs ...fio.MergeOption) (stat fio.IFileStat, err error) {
	// check src and dst
	if len(srcs) == 0 {
		return stat, fmt.Errorf("no source file")
	}
	// 通过第一个 '/' 拆分dst的bucket和object
	dstPaths := strings.SplitN(dst.Path, "/", 2)
	if len(dstPaths) != 2 {
		return stat, fmt.Errorf("invalid destination path")
	}
	var (
		dstOpt = minio.CopyDestOptions{
			Bucket: dstPaths[0],
			Object: dstPaths[1],
		}
		srcOpts = make([]minio.CopySrcOptions, 0, len(srcs))
	)
	for i, src := range srcs {
		srcPaths := strings.SplitN(src.Path, "/", 2)
		if len(srcPaths) != 2 {
			return stat, fmt.Errorf("invalid source path %d", i)
		}
		srcOpts = append(srcOpts, minio.CopySrcOptions{
			Bucket: srcPaths[0],
			Object: srcPaths[1],
		})
	}

	res, err := m.cl.ComposeObject(context.Background(), dstOpt, srcOpts...)
	if err != nil {
		return stat, fmt.Errorf("MergeFile failed: %w", err)
	}

	var statF = &FileStatHandler{}
	util.GobConv(res, statF)

	return statF, nil
}

func (m *MinioFileBucket) Bucket() string {
	return m.bucket
}

func (m *MinioFileBucket) Init() (err error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	// 检查存储桶
	b, err := m.cl.BucketExists(ctx, m.bucket)
	if err != nil {
		return err
	}
	if !b {
		// 不存在则创建
		err = m.cl.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		var defaultQuota int64 = 1024
		err = m.SetConfig(fio.BucketConfig{
			Quota: &defaultQuota,
		})
		if err != nil {
			m.cl.RemoveBucket(ctx, m.bucket)
			return err
		}

	}

	return nil
}

// Info implements fio.IFileBucket
// 获取存储桶信息 (配额, 版本控制)
func (m *MinioFileBucket) Info() (stat fio.BucketStat, err error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	// 获取版本控制
	bvc, err := m.cl.GetBucketVersioning(ctx, m.bucket)
	if err != nil {
		err = fmt.Errorf("error on get bucket versioning: %v", err)
		return
	}
	stat.Vers = bvc.Enabled()

	// 获取配额
	bq, err := m.mc.GetBucketQuota(ctx, m.bucket)
	if err != nil {
		err = fmt.Errorf("error on get bucket quota: %v", err)
		return
	}
	stat.Quota = int64(bq.Quota) / 1024 / 1024

	// get bucket size
	bi, err := m.mc.AccountInfo(ctx, madmin.AccountOpts{
		PrefixUsage: true,
	})
	if err != nil {
		err = fmt.Errorf("error on get bucket size: %v", err)
	}

	for i, bai := range bi.Buckets {
		if bai.Name == m.bucket {
			stat.Size = bi.Buckets[i].Size / 1024 / 1024
			stat.Files = bi.Buckets[i].Objects
			break
		}
	}

	return
}

func (m *MinioFileBucket) SetConfig(conf fio.BucketConfig) (err error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	// 设置版本控制
	if conf.Vers != nil {
		var versStatus = "Disabled"
		if *conf.Vers {
			versStatus = "Enabled"
		}
		err = m.cl.SetBucketVersioning(ctx, m.bucket, minio.BucketVersioningConfiguration{
			Status: versStatus,
		})
		if err != nil {
			err = fmt.Errorf("error on set bucket versioning: %v", err)
			return
		}
	}

	// 设置配额
	if conf.Quota != nil {
		err = m.mc.SetBucketQuota(ctx, m.bucket, &madmin.BucketQuota{
			Quota: uint64(*conf.Quota) * 1024 * 1024,
			Type:  madmin.HardQuota,
		})
		if err != nil {
			err = fmt.Errorf("error on set bucket quota: %v", err)
			return
		}
	}

	return
}

// OpenFile implements io.IFileSystem
// 打开文件(虚拟)
func (m *MinioFileBucket) OpenFile(name string) (fio.IFile, error) {
	return &MinioFile{name: name, fs: m}, nil
}

// OpenReadOnly implements io.IFileSystem
// 打开文件(只读管道)
func (m *MinioFileBucket) Open(name string) (io.ReadSeekCloser, error) {
	o, err := m.cl.GetObject(context.Background(), m.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Stat implements io.IFileSystem
// 获取文件信息
func (m *MinioFileBucket) Stat(name string) (fio.IFileStat, error) {
	name = strings.TrimPrefix(name, "/")

	oi, err := m.cl.StatObject(context.Background(), m.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		err2, ok := err.(minio.ErrorResponse)
		// 特殊处理以匹配标准文件系统
		if ok && err2.Code == "NoSuchKey" {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	return &FileStatHandler{oi}, nil
}

var _ fs.FileInfo = &FileStatHandler{}

func (m *MinioFileBucket) Remove(name string) error {
	ctx, cf := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cf()
	return m.cl.RemoveObject(ctx, m.bucket, name, minio.RemoveObjectOptions{})
}

func (m *MinioFileBucket) List(path string) (res []fs.FileInfo, err error) {
	ctx, cf := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cf()

	// 处理前置的 "/" (minio不支持)
	path = strings.TrimPrefix(path, "/")

	rc := m.cl.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
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

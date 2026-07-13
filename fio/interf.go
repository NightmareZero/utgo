package fio

import (
	"context"
	"fmt"
	"io"
	"os"
)

var FileSystem IFileSystem
var ErrBucketNotFound = os.ErrNotExist

// 默认使用本地文件系统(以后需要被 minio相关实现覆盖)
type IFileSystem interface {
	IsOnline() bool
	Bucket(name string, create bool) (IFileBucket, error)
	// ClusterInfo 返回集群级汇总信息。
	// 复合类型：通用字段 + Extra(any) 携带 adapter 专有数据。
	//   - minio 实现: Extra 为 *miniofs.MinioClusterExtra
	//   - localfs 实现: Extra 为 nil
	ClusterInfo(ctx context.Context) (ClusterInfo, error)
}

type IFileBucket interface {
	Bucket() string
	Info() (stat BucketStat, err error)
	SetConfig(conf BucketConfig) (err error)
	List(path string) ([]os.FileInfo, error)
	Open(name string) (io.ReadSeekCloser, error)
	OpenFile(name string) (IFile, error)
	// MergeFile merges source files into dst in order and does not remove source files.
	MergeFile(ctx context.Context, dst MergeOption, src ...MergeOption) (IFileStat, error)
	Stat(name string) (IFileStat, error)
	Remove(name string) error
	Destroy() error
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

type MergeOption struct {
	Path string
}

// ClusterInfo 集群级汇总信息
//
// Extra 字段为 adapter 专有数据：
//   - minio 实现: *miniofs.MinioClusterExtra（持有原始 *madmin.InfoMessage）
//   - localfs 实现: nil
type ClusterInfo struct {
	BackendType   string // 后端类型: "Erasure"/"FS"/"Gateway"/"LocalFS"
	Mode          string // 集群运行模式 (minio 专有，localfs 留空)
	TotalCapacity uint64 // 总容量 (字节)
	UsedCapacity  uint64 // 已用容量 (字节)
	FreeCapacity  uint64 // 可用容量 (字节)
	BucketCount   uint64 // 桶总数
	ObjectCount   uint64 // 对象总数
	VersionCount  uint64 // 版本总数 (含历史版本)
	Extra         any    // adapter 专有信息 (miniofs.MinioClusterExtra / nil)
}

// capStr 将字节数格式化为人类可读的容量字符串
func capStr(b uint64) string {
	switch {
	case b == 0:
		return "0 B"
	case b >= 1<<40:
		return fmt.Sprintf("%.2f TiB", float64(b)/(1<<40))
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GiB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.2f MiB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.2f KiB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// String 返回人类可读的集群信息摘要
func (c ClusterInfo) String() string {
	usage := "N/A"
	if c.TotalCapacity > 0 {
		usage = fmt.Sprintf("%s / %s (%.1f%%)",
			capStr(c.UsedCapacity), capStr(c.TotalCapacity),
			100*float64(c.UsedCapacity)/float64(c.TotalCapacity))
	}
	return fmt.Sprintf("ClusterInfo{Backend=%s, Mode=%s, Capacity(used/total)=%s, Free=%s, Buckets=%d, Objects=%d, Versions=%d}",
		c.BackendType, c.Mode, usage, capStr(c.FreeCapacity),
		c.BucketCount, c.ObjectCount, c.VersionCount,
	)
}

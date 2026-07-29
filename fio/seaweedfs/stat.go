package seaweedfs

import (
	"io/fs"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

var _ fs.FileInfo = &FileStatHandler{}

// FileStatHandler 包装 minio.ObjectInfo，实现 fio.IFileStat。
// 与 minio 后端的 FileStatHandler 结构一致：SeaweedFS S3 网关返回的对象信息格式相同。
type FileStatHandler struct {
	minio.ObjectInfo
}

// Sha1 implements fio.IFileStat.
// SeaweedFS S3 网关通常不返回 ChecksumSHA1（与 minio 不同），
// 此处保留字段读取；若为空，调用方可按需自行计算。
func (f *FileStatHandler) Sha1() string {
	return f.ObjectInfo.ChecksumSHA1
}

// IsDir implements fs.FileInfo
func (f *FileStatHandler) IsDir() bool {
	return strings.HasSuffix(f.Key, "/")
}

// ModTime implements fs.FileInfo
func (f *FileStatHandler) ModTime() time.Time {
	return f.LastModified
}

// Mode implements fs.FileInfo
func (f *FileStatHandler) Mode() fs.FileMode {
	return fs.ModeType
}

// Name implements fs.FileInfo
func (f *FileStatHandler) Name() string {
	return f.Key
}

// Size implements fs.FileInfo
func (f *FileStatHandler) Size() int64 {
	return f.ObjectInfo.Size
}

// Sys implements fs.FileInfo
func (f *FileStatHandler) Sys() any {
	return nil
}

package miniofs

import (
	"io/fs"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type FileStatHandler struct {
	minio.ObjectInfo
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

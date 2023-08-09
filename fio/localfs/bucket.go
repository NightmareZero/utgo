package localfs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/NightmareZero/nzgoutil/fio"
)

var _ fio.IFileBucket = &LocalFileBucket{}

// 本地文件系统(需要被 minio相关实现覆盖)
type LocalFileBucket struct {
	bucket   string
	basePath string
}

// Bucket implements fio.IFileBucket.
func (l *LocalFileBucket) Bucket() string {
	return l.bucket
}

// SetConfig implements fio.IFileBucket.
func (*LocalFileBucket) SetConfig(conf fio.BucketConfig) (err error) {
	// TODO
	return
}

// Info implements fio.IFileBucket.
func (*LocalFileBucket) Info() (stat fio.BucketStat, err error) {
	// TODO
	return
}

// List implements fio.IFileBucket.
func (*LocalFileBucket) List(path string) ([]fs.FileInfo, error) {
	de, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var res []fs.FileInfo
	for _, v := range de {
		fi, _ := v.Info()
		res = append(res, fi)
	}
	return res, nil
}

func (l *LocalFileBucket) Init() error {
	return assertDirExist(l.basePath)
}

// OpenFile implements IFileSystem
func (l *LocalFileBucket) OpenFile(name string) (fio.IFile, error) {
	// 处理恶意行为
	name = strings.ReplaceAll(name, "..", "")

	// 拼接并防止路径出现问题
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimPrefix(name, "/")

	// 读取目标路径
	dir := filepath.Dir(absFileName)

	// 检查文件目录是否存在
	if err := assertDirExist(dir); err != nil {
		return nil, err
	}

	// 创建或打开文件
	return os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

}

// OpenReadOnly implements IFileSystem
func (l *LocalFileBucket) Open(name string) (io.ReadCloser, error) {
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimPrefix(name, "/")
	return os.Open(absFileName)
}

// Stat implements IFileSystem
func (l *LocalFileBucket) Stat(name string) (fs.FileInfo, error) {
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimPrefix(name, "/")
	return os.Stat(absFileName)
}

func (l *LocalFileBucket) Remove(name string) error {
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimPrefix(name, "/")
	return os.Remove(absFileName)
}

func assertDirExist(dirpath string) error {
	dir_fi, err := os.Stat(dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			var err error
			fio.TmpFileMask(func() {
				err = os.MkdirAll(dirpath, 0777)
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			return err

		}
	}
	if !dir_fi.IsDir() {
		return fmt.Errorf("target path not a folder")
	}
	return nil
}

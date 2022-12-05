package fio

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// 本地文件系统(需要被 minio相关实现覆盖)
type _localFileSystem struct {
	basePath string
}

// OpenFile implements IFileSystem
func (l *_localFileSystem) OpenFile(name string) (IFile, error) {
	// 拼接并防止路径出现问题
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimLeft(name, "/")

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
func (l *_localFileSystem) Open(name string) (io.ReadCloser, error) {
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimLeft(name, "/")
	return os.Open(absFileName)
}

// Stat implements IFileSystem
func (l *_localFileSystem) Stat(name string) (fs.FileInfo, error) {
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimLeft(name, "/")
	return os.Stat(absFileName)
}

func (l *_localFileSystem) Remove(name string) error {
	absFileName := strings.TrimRight(l.basePath, "/") + string(filepath.Separator) + strings.TrimLeft(name, "/")
	return os.Remove(absFileName)
}

func assertDirExist(dirpath string) error {
	dir_fi, err := os.Stat(dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			mask := syscall.Umask(0)  // 改为 0000 八进制
			defer syscall.Umask(mask) // 改为原来的 umask
			err = os.MkdirAll(dirpath, 0777)
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

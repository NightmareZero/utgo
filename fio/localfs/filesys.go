package localfs

import (
	"sync"

	"github.com/NightmareZero/nzgoutil/fio"
)

type LocalFileSystem struct {
	BasePath string
	lock     sync.RWMutex
	buckets  map[string]*LocalFileBucket
}

func (l *LocalFileSystem) Init() {
	l.buckets = make(map[string]*LocalFileBucket)
}

// Bucket implements fio.IFileSystem.
func (l *LocalFileSystem) Bucket(name string) (fio.IFileBucket, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if b, ok := l.buckets[name]; ok {
		// 存在则返回
		return b, nil
	} else {
		// 不存则创建后返回
		b = &LocalFileBucket{
			bucket:   name,
			basePath: l.BasePath + "/" + name + "/",
		}
		// 初始化
		err := b.Init()
		if err != nil {
			return nil, err
		}
		// 保存
		l.buckets[name] = b
		return b, nil
	}

}

func (l *LocalFileSystem) IsOnline() bool {
	return true
}

//go:build darwin

package localfs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/NightmareZero/nzgoutil/fio"
)

// ClusterInfo implements fio.IFileSystem.
// 读取 BasePath 所在磁盘的容量信息。
func (l *LocalFileSystem) ClusterInfo(_ context.Context) (fio.ClusterInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(l.BasePath, &stat); err != nil {
		return fio.ClusterInfo{}, fmt.Errorf("error on statfs %s: %w", l.BasePath, err)
	}

	bsize := uint64(stat.Bsize)
	total := stat.Blocks * bsize
	// Bavail 是普通用户可用 (扣除 root 保留块)
	free := stat.Bavail * bsize
	var used uint64
	if total > free {
		used = total - free
	}

	return fio.ClusterInfo{
		BackendType:   "LocalFS",
		TotalCapacity: total,
		UsedCapacity:  used,
		FreeCapacity:  free,
	}, nil
}

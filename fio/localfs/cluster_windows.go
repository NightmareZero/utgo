//go:build windows

package localfs

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/NightmareZero/nzgoutil/fio"
)

// 使用 LazyDLL 动态加载 kernel32，确保交叉编译时也能通过
// （syscall.GetDiskFreeSpaceEx 在非 windows 平台上不可见）
var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetDiskFreeSpaceEx = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// ClusterInfo implements fio.IFileSystem.
// 读取 BasePath 所在磁盘的容量信息。
func (l *LocalFileSystem) ClusterInfo(_ context.Context) (fio.ClusterInfo, error) {
	freeBytesAvail, totalBytes, _, err := getDiskFreeSpaceEx(l.BasePath)
	if err != nil {
		return fio.ClusterInfo{}, fmt.Errorf("error on get disk space: %w", err)
	}

	var used uint64
	if totalBytes > freeBytesAvail {
		used = totalBytes - freeBytesAvail
	}

	return fio.ClusterInfo{
		BackendType:   "LocalFS",
		TotalCapacity: totalBytes,
		UsedCapacity:  used,
		FreeCapacity:  freeBytesAvail,
	}, nil
}

// getDiskFreeSpaceEx 包装 GetDiskFreeSpaceExW:
//   - freeBytesAvail  - 当前用户可用字节
//   - totalBytes      - 磁盘总容量
//   - totalFreeBytes  - 磁盘全部空闲字节 (含 root 保留)
func getDiskFreeSpaceEx(path string) (freeBytesAvail, totalBytes, totalFreeBytes uint64, err error) {
	pathPtr, e := syscall.UTF16PtrFromString(path)
	if e != nil {
		return 0, 0, 0, e
	}
	r1, _, e1 := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvail)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if r1 == 0 {
		if e1 != syscall.Errno(0) {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

package fio

import "syscall"

func TmpFileMask(callback func()) {
	mask := syscall.Umask(0)  // 改为 0000 八进制
	defer syscall.Umask(mask) // 改为原来的 umask
	callback()
}

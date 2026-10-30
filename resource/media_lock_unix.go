//go:build unix

package resource

import (
	"os"
	"syscall"
)

func lockCache(filename string) (func(), error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}
	// 保留锁文件，避免删除重建后不同进程锁住不同 inode；关闭或进程退出自动解锁。
	return func() { f.Close() }, nil
}

//go:build windows
// +build windows

package util

import (
	"os"
	"path/filepath"
	"syscall"
)

// AllocateDisk 申请固定大小的磁盘空间，如果申请失败返回error，否则返回*os.File
func AllocateDisk(filePath string, size FileSize) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(filePath), os.ModeDir)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	// 设置文件大小
	if err := syscall.Ftruncate(syscall.Handle(f.Fd()), int64(size)); err != nil {
		defer func() {
			f.Close()
			os.Remove(filePath)
		}()

		return nil, err
	}

	return f, nil
}

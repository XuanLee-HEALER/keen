package fsop

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync/atomic"
)

func copyFileN(dst string, src string, block int64, errflag *atomic.Int32) error {
	f, err := os.Create(dst)
	if err != nil {
		if errflag.CompareAndSwap(0, 1) {
			return err
		} else {
			return nil
		}
	}
	defer f.Close()
	defer f.Sync()

	f0, err := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if errflag.CompareAndSwap(0, 1) {
			return err
		} else {
			return nil
		}
	}
	defer f0.Close()

	for {
		if errflag.Load() == 1 {
			return nil
		}
		_, err := io.CopyN(f, f0, int64(block))
		if err == io.EOF {
			return nil
		}

		if err != nil {
			if errflag.CompareAndSwap(0, 1) {
				return err
			} else {
				return nil
			}
		}
	}
}

// CopyFiles 拷贝文件列表，目标文件列表dsts和源文件列表srcs长度需要相等，否则会panic，
// 通过block指定拷贝单位
func CopyFiles(dsts []string, srcs []string, block int64) error {
	if len(srcs) == 0 || len(dsts) == 0 {
		return errors.New("there is not files will be copied")
	}
	errflag := new(atomic.Int32)
	counter := int32(len(srcs))
	errCh := make(chan error)
	for i := range srcs {
		go func(idx int) {
			defer func() {
				if atomic.AddInt32(&counter, -1) <= 0 {
					close(errCh)
				}
			}()
			errCh <- copyFileN(dsts[idx], srcs[idx], block, errflag)
		}(i)
	}

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// InsufficientDiskSpaceError 判断拷贝文件错误类型是否为磁盘空间不足
func InsufficientDiskSpaceError(err error) bool {
	return strings.Contains(err.Error(), "There is not enough space on the disk.")
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

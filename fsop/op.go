package fsop

import (
	"errors"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
)

type FileSize int64

const (
	KB         FileSize = 1024
	MB         FileSize = 1024 * 1024
	GB         FileSize = 1024 * 1024 * 1024
	WRITE_UNIT FileSize = 4 * KB
)

// CreateRandomFile 创建随机内容的文件
func CreateRandomFile(path string, size FileSize) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync()

	var ruler int
	unit := make([]byte, WRITE_UNIT)
	for i := 0; i < int(size); i++ {
		ruler = i % int(WRITE_UNIT)
		unit[ruler] = byte(rand.Intn(0xff))
		if ruler == int(WRITE_UNIT)-1 {
			_, err := f.Write(unit)
			if err != nil {
				return err
			}
			unit = make([]byte, WRITE_UNIT)
		}
	}

	if ruler != int(WRITE_UNIT)-1 {
		_, err := f.Write(unit[:ruler+1])
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFileN(dst string, src string, block FileSize, errflag *atomic.Int32) error {
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
func CopyFiles(dsts []string, srcs []string, block FileSize) error {
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

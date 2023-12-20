package util

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

type FileSize int64

const (
	B          FileSize = 1
	KB         FileSize = 1024
	MB         FileSize = 1024 * 1024
	GB         FileSize = 1024 * 1024 * 1024
	WRITE_UNIT FileSize = 4 * KB
)

type SpaceNotEnoughErr struct {
	Err error
}

func (e *SpaceNotEnoughErr) Error() string {
	return "There is not enough space on the disk."
}

// CreateFixSizeFile 创建一个固定大小的文件，如果创建失败返回error
func CreateFixSizeFile(fn string, size FileSize) (*os.File, error) {
	f, err := os.Create(fn)
	if err != nil {
		return nil, err
	}

	// 设置文件大小
	if err := syscall.Ftruncate(syscall.Handle(f.Fd()), int64(size)); err != nil {
		defer func() {
			f.Close()
			os.Remove(fn)
		}()

		return nil, err
	}

	return f, nil
}

func IsSpaceNotEnough(err error) bool {
	return err.Error() == "There is not enough space on the disk."
}

type CopyTask struct {
	src string
	dst string
}

func NewCopyTask(src, dst string) CopyTask {
	return CopyTask{src, dst}
}

func CopyFileList(tasks ...CopyTask) {

}

func Copy(task CopyTask, conc int) error {
	// 打开源文件
	srcFile, err := os.Open(task.src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	// 获取源文件大小
	fileSize := srcFileInfo.Size()

	// 创建目标文件
	dstFile, err := CreateFixSizeFile(task.dst, FileSize(fileSize))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 计算每个goroutine需要复制的字节数
	chunkSize := fileSize / int64(conc)

	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(conc)

	// 创建一个channel，用于在任意goroutine失败时通知主goroutine
	errChan := make(chan error, conc)

	// 启动多个goroutine并发地复制文件
	for i := 0; i < conc; i++ {
		go func(i int) {
			// 计算当前goroutine需要复制的起始位置和结束位置
			offset := int64(i) * chunkSize
			start := offset
			end := offset + chunkSize
			if i == concurrency-1 {
				end = fileSize
			}

			// 复制文件
			_, err := srcFile.Seek(start, io.SeekStart)
			if err != nil {
				errChan <- err
				return
			}
			n, err := io.CopyN(dstFile, srcFile, end-start)
			if err != nil {
				errChan <- err
				return
			}
			fmt.Printf("goroutine %d copied %d bytes\n", i, n)

			// 通知等待组
			wg.Done()
		}(i)
	}

	// 等待所有goroutine完成或出错
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 等待任意goroutine失败或全部完成
	for err := range errChan {
		if err != nil {
			fmt.Printf("error: %v\n", err)
			for range errChan {
				// drain the channel
			}
			break
		}
	}

	fmt.Println("copy completed")
}

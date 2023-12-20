package util

import (
	"io"
	"math/rand"
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

func FillFile(f *os.File, size FileSize) error {
	const blockSize = 4096

	randArr := func(arr *[blockSize]byte) {
		for i := 0; i < blockSize; i++ {
			arr[i] = byte(rand.Intn(256))
		}
	}

	block := [blockSize]byte{}
	x, r := size/blockSize, size%blockSize
	if r > 0 {
		x++
	}

	for i := 0; i < int(x); i++ {
		randArr(&block)
		f.Seek(int64(blockSize*i), 0)
		_, err := f.Write(block[:])
		if err != nil {
			return err
		}
	}

	return nil
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

func Copy(task CopyTask, conc int64, buffer int64) error {
	// 打开源文件
	sourceFile, err := os.Open(task.src)
	if err != nil {
		return err
	}

	// 打开目标文件
	destinationFile, err := os.Create(task.dst)
	if err != nil {
		sourceFile.Close()
		return err
	}

	// 创建一个同步器来同步对两个文件的访问
	mutex := sync.Mutex{}

	// 创建一个缓冲区
	buf := make([]byte, buffer)

	// 计算文件的大小
	fileSize, err := sourceFile.Seek(0, io.SeekEnd)
	if err != nil {
		sourceFile.Close()
		destinationFile.Close()
		return err
	}

	// 创建一个通道来传递拷贝任务
	tasks := make(chan int, conc)

	// 启动多个goroutine来拷贝文件
	for i := 0; i < conc; i++ {
		go func() {
			// 从通道中获取拷贝任务
			start := <-tasks
			end := start + 4096

			// 从源文件读取数据
			for n := start; n < end; n += 4096 {
				// 从源文件读取数据
				n, err := sourceFile.Read(buf[n-start:])
				if err != nil {
					if err == io.EOF {
						break
					}
					sourceFile.Close()
					destinationFile.Close()
					return
				}

				// 将数据写入目标文件
				mutex.Lock()
				destinationFile.Write(buf[n-start:])
				mutex.Unlock()
			}
		}()
	}

	// 将所有拷贝任务放入通道
	for i := 0; i < int(fileSize)/4096; i++ {
		tasks <- i * 4096
	}

	// 关闭通道
	close(tasks)

	// 等待所有goroutine执行完毕
	for range tasks {
	}

	// 关闭两个文件
	sourceFile.Close()
	destinationFile.Close()

	return nil
}

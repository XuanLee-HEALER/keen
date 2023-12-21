package util

import (
	"container/heap"
	"fmt"
	"io"
	"math/rand"
	"os"
	"syscall"

	"gitea.fcdm.top/lixuan/keen"
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

// CreateFixSizeFile 向硬盘申请固定大小的空间，如果申请失败返回error
func CreateFixSizeFile(fn string, size FileSize) (*os.File, CleanFunc, error) {
	f, err := os.Create(fn)
	if err != nil {
		return nil, nil, err
	}

	// 设置文件大小
	if err := syscall.Ftruncate(syscall.Handle(f.Fd()), int64(size)); err != nil {
		defer func() {
			f.Close()
			os.Remove(fn)
		}()

		return nil, nil, err
	}

	return f, func() {
		defer f.Close()
	}, nil
}

func IsSpaceNotEnough(err error) bool {
	return err.Error() == "There is not enough space on the disk."
}

type CopyTask struct {
	Src   string
	Dst   string
	sFile *os.File
	dFile *os.File
	Size  FileSize
}

func NewCopyTask(src, dst string) *CopyTask {
	return &CopyTask{
		Src: src,
		Dst: dst,
	}
}

func (t CopyTask) String() string {
	return fmt.Sprintf("Copy Task: source(%s) -> destination(%s)", t.Src, t.Dst)
}

func SetupCopyTasks(tasks ...*CopyTask) error {
	errCh := make(chan error)

	for _, task := range tasks {
		if task.Src == "" || task.Dst == "" {
			errCh <- fmt.Errorf("source file (%s) or destination file (%s) is invalid", task.Src, task.Dst)
		}
		go func(t *CopyTask) {
			rf, err := os.Open(t.Src)
			if err != nil {
				errCh <- err
				return
			}

			fs, err := rf.Seek(0, io.SeekEnd)
			if err != nil {
				errCh <- err
				return
			}

			df, _, err := CreateFixSizeFile(t.Dst, FileSize(fs)*B)
			if err != nil {
				errCh <- err
				return
			}

			t.sFile = rf
			t.dFile = df
			t.Size = FileSize(fs)
			errCh <- nil
		}(task)
	}

	var err error
	clean := false
	for err = range errCh {
		if err != nil {
			clean = true
			break
		}
	}

	if clean {
		for _, t := range tasks {
			if t.sFile != nil {
				t.sFile.Close()
			}
			if t.dFile != nil {
				err := os.Remove(t.Dst)
				if err != nil {
					keen.H(fmt.Sprintf("failed to delete newly created file [%s]: %v", t.Dst, err))
				}
			}
		}
	}

	return err
}

type GroupCopyTask struct {
	SizeSum FileSize
	Tasks   []*CopyTask
}

type CopyTaskHeap []GroupCopyTask

func (h CopyTaskHeap) Len() int {
	return len(h)
}

func (h CopyTaskHeap) Less(i, j int) bool {
	return h[i].SizeSum < h[j].SizeSum
}

func (h CopyTaskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *CopyTaskHeap) Push(x any) {
	*h = append(*h, x.(GroupCopyTask))
}

func (h *CopyTaskHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func SplitTasksToNGroups(tasks []*CopyTask, minGroups int, maxGroups int) ([]GroupCopyTask, bool) {
	if minGroups < 1 || maxGroups > 8 {
		return nil, false
	}

	var ini CopyTaskHeap = make([]GroupCopyTask, 0)
	for i := 0; i < minGroups; i++ {
		ini = append(ini, GroupCopyTask{0, make([]*CopyTask, 0)})
	}

	heap.Init(&ini)

	return nil, true
}

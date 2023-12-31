package util

import (
	"container/heap"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"

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
	return fmt.Sprintf("Copy Task: source(%s)|size(%d MB) -> destination(%s)", t.Src, t.Size/MB, t.Dst)
}

func (t *CopyTask) Clean() error {
	eg := NewErrGroup()

	if t.sFile != nil {
		err := t.sFile.Close()
		if err != nil {
			eg.AddErrs(err)
		}
	}

	if t.dFile != nil {
		err := t.dFile.Close()
		if err != nil {
			eg.AddErrs(err)
		}

		err = os.Remove(t.Dst)
		if err != nil {
			eg.AddErrs(err)
		}
	}

	if eg.IsNil() {
		return nil
	}

	return eg
}

// SetupCopyTasks 初始化拷贝任务，打开所有的源文件、目标文件，包括获取文件大小，申请磁盘空间，如果出现错误，要保证所有的fd（file descriptor）被关闭
func SetupCopyTasks(tasks []*CopyTask) error {
	if len(tasks) <= 0 {
		return errors.New("no task to setup")
	}

	errCh := make(chan error, len(tasks))
	var cl atomic.Int32
	cl.Store(int32(len(tasks)))

	for _, task := range tasks {
		go func(t *CopyTask) {
			defer func() {
				if v := cl.Add(-1); v <= 0 {
					close(errCh)
				}
			}()

			if t.Src == "" || t.Dst == "" {
				errCh <- fmt.Errorf("source file (%s) or destination file (%s) is invalid", task.Src, task.Dst)
				return
			}

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

			_, err = rf.Seek(0, io.SeekStart)
			if err != nil {
				errCh <- err
				return
			}

			df, err := AllocateDisk(t.Dst, FileSize(fs)*B)
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

	eg := NewErrGroup()
	for err := range errCh {
		if err != nil {
			eg.errors = append(eg.errors, err)
		}
	}

	if !eg.IsNil() {
		for _, t := range tasks {
			if err := t.Clean(); err != nil {
				eg.AddErrs(err)
			}
		}
		return eg
	}

	return nil
}

type GroupCopyTask struct {
	SizeSum FileSize
	Tasks   []*CopyTask
}

func (gct GroupCopyTask) String() string {
	return fmt.Sprintf("sum of file size(%d MB), task detail:\n%v", gct.SizeSum/MB, gct.Tasks)
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

// SplitTasksToNGroups 为拷贝任务分组
func SplitTasksToNGroups(tasks []*CopyTask, maxGroups uint) []GroupCopyTask {
	if maxGroups >= 16 {
		keen.Log.Warn("The maximum number of concurrent tasks is set to 16")
		maxGroups = 16
	}
	// 如果tasks为0、1，那么直接返回GroupTask
	// 如果tasks为2（min）~max中间的值，那么返回n个GroupTask
	// 如果task大于max，那么开始尝试分组（贪心）
	tasksNum := len(tasks)
	if tasksNum <= 0 {
		return []GroupCopyTask{}
	} else if tasksNum <= 1 {
		t := tasks[0]
		return []GroupCopyTask{GroupCopyTask{t.Size, []*CopyTask{t}}}
	} else if tasksNum <= int(maxGroups) {
		g := make([]GroupCopyTask, 0, tasksNum)
		for i := 0; i < tasksNum; i++ {
			ct := tasks[i]
			g[i] = GroupCopyTask{ct.Size, []*CopyTask{ct}}
		}
		return g
	} else {
		var hpd CopyTaskHeap = make([]GroupCopyTask, 0)
		for i := 0; i < int(maxGroups); i++ {
			hpd = append(hpd, GroupCopyTask{0, make([]*CopyTask, 0)})
		}

		hp := &hpd
		heap.Init(hp)
		for _, t := range tasks {
			gt := heap.Pop(hp).(GroupCopyTask)
			gt.SizeSum += t.Size
			gt.Tasks = append(gt.Tasks, t)
			heap.Push(hp, gt)
		}

		return []GroupCopyTask(hpd)
	}
}

// RemoveChildren 删除目录下所有的子项，不包括目录自身
func RemoveChildren(path string) error {
	path = filepath.Join(path, "*")
	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	for _, f := range matches {
		err := os.RemoveAll(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveChildren 删除目录下所有的子项，不包括目录自身
func RemoveChildrenWithError(path string) error {
	errg := NewErrGroup()
	path = filepath.Join(path, "*")
	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	for _, f := range matches {
		err := os.RemoveAll(f)
		if err != nil {
			errg.AddErrs(err)
		}
	}

	return errg
}

// func ExecGroupCopyTasks(gtasks []GroupCopyTask, buffer int, ot int64) error {
// 	l := len(gtasks)
// 	errCh := make(chan error, l)
// 	var gl atomic.Int32
// 	gl.Store(int32(l))

// 	tm, err := NewTaskMonitor(l, ot)
// 	if err != nil {
// 		return err
// 	}

// 	resCh := make(chan []TaskStateSummaryMsg)
// 	tm.Register(resCh)

// 	go tm.Run()

// 	for i, gt := range gtasks {
// 		go func(idx int, gt GroupCopyTask) {
// 			defer func() {
// 				if gl.Add(-1) <= 0 {
// 					close(errCh)
// 				}
// 			}()

// 			var (
// 				n      int
// 				err    error
// 				isErr  bool = false
// 				isHalt bool = false
// 			)

// 			buf := make([]byte, buffer)

// 		out:
// 			for _, subT := range gt.Tasks {
// 				keen.Log.Debug("[%s -> %s]start to copy file...", subT.Src, subT.Dst)
// 				st := time.Now()
// 				keen.Log.Debug("[%s -> %s]starttime: %s", subT.Src, subT.Dst, st.Format(ylog2.LOG_TIME_FMT))
// 			inner:
// 				for {
// 					n, err = subT.sFile.Read(buf)
// 					if err != nil {
// 						if errors.Is(err, io.EOF) {
// 							break inner
// 						}
// 						isErr = true
// 						break out
// 					}

// 					_, err = subT.dFile.Write(buf[:n])
// 					if err != nil {
// 						isErr = true
// 						break out
// 					}

// 					// select {
// 					// case <-haltCh:
// 					// 	{
// 					// 		isHalt = true
// 					// 		break inner
// 					// 	}
// 					// default:
// 					continue inner
// 					// }
// 				}

// 				subT.dFile.Sync()
// 				subT.sFile.Close()
// 				subT.dFile.Close()
// 				et := time.Now()
// 				keen.Log.Debug("[%s -> %s]endtime: %s, elapse time: %.2f(s)", subT.Src, subT.Dst, et.Format(ylog2.LOG_TIME_FMT), et.Sub(st).Seconds())
// 			}

// 			if isErr || isHalt {
// 				for _, subT := range gt.Tasks {
// 					err := subT.Clean()
// 					if err != nil {
// 						keen.Log.Error("failed to clean the copy task ([%s]-[%s]) while the error occurred: %v", subT.Src, subT.Dst, err)
// 					}
// 				}
// 				if isErr {
// 					errCh <- err
// 				}
// 			}
// 		}(i, gt)
// 	}

// 	// for range tm.Register()

// 	eg := NewErrGroup()
// 	for err := range errCh {
// 		if err != nil {
// 			eg.AddErrs(err)
// 		}
// 	}

// 	if eg.IsNil() {
// 		return nil
// 	}

// 	return eg
// }

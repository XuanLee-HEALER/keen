package util_test

import (
	"container/heap"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
	"github.com/stretchr/testify/assert"
)

func TestPathExists(t *testing.T) {
	pwd, err := os.Getwd()
	assert.Empty(t, err, "failed to get current directory")

	p1 := filepath.Join(pwd, "test_dir")
	r1 := util.PathExists(p1)
	assert.Equal(t, false, r1, "the destination directory does not exist")

	err = os.Mkdir(p1, os.ModeDir)
	assert.Empty(t, err, "failed to create directory")

	r2 := util.PathExists(p1)
	assert.Equal(t, true, r2, "the distination directory dose exist")

	os.RemoveAll(p1)
}

func TestAllocateDisk(t *testing.T) {
	fn := "./xx1"
	f, err := util.AllocateDisk(fn, 1000*util.MB)
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}

	if fi, err := f.Stat(); err != nil {
		t.Error(err)
	} else {
		if util.FileSize(fi.Size()) != 1000*util.MB {
			t.FailNow()
		}
	}

	f.Close()
	// os.Remove(fn)
}

func TestAllocateDiskOutOfSpace(t *testing.T) {
	fn := "./xx2"
	f, err := util.AllocateDisk(fn, 1000*util.GB)
	if err != nil {
		if util.IsSpaceNotEnough(err) {
			t.SkipNow()
		} else {
			t.FailNow()
		}
	}
	defer func() {
		f.Close()
		os.Remove(fn)
	}()
	t.FailNow()
}

func TestCopyTaskHeap(t *testing.T) {
	gTasks := make(util.CopyTaskHeap, 0)

	for i := 0; i < 50; i++ {
		rs := rand.Intn(10)
		gTasks = append(gTasks, util.GroupCopyTask{util.FileSize(rs), nil})
	}

	heap.Init(&gTasks)

	for gTasks.Len() > 0 {
		t.Log("current value", int64(heap.Pop(&gTasks).(util.GroupCopyTask).SizeSum))
	}
}

func TestCopy(t *testing.T) {
	const (
		SRC  = "SRC"
		DST  = "DST"
		SRCF = "src_"
		DSTF = "dst_"
	)

	_, err := os.Stat(SRC)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(SRC, os.ModeDir)
		} else {
			t.Error(err)
		}
	}

	matches, err := filepath.Glob(filepath.Join(SRC, "*"))
	if err != nil {
		t.Error(err)
	}

	if len(matches) == 0 {
		for _, f := range matches {
			err := os.RemoveAll(f)
			if err != nil {
				t.Logf("failed to remove file/directory [%s]: %v", f, err)
			}
		}
	}

	tasks := make([]*util.CopyTask, 0)

	for i := 0; i < 10; i++ {
		curF := filepath.Join(SRC, SRCF+strconv.Itoa(i))
		df := filepath.Join(DST, DSTF+strconv.Itoa(i))

		t := util.NewCopyTask(curF, df)
		tasks = append(tasks, t)
	}

	err = util.SetupCopyTasks(tasks)
	if err != nil {
		t.Errorf("failed to setup task:\n  %v", err)
		t.FailNow()
	}

	for _, xt := range tasks {
		t.Log("task:", xt)
	}
	t.Log(strings.Repeat("=", 100))
	gct := util.SplitTasksToNGroups(tasks, 1)
	for _, g := range gct {
		t.Log("task group:", g)
	}

	// t.Log(strings.Repeat("=", 100))
	// err = util.ExecGroupCopyTasks(gct, 8096, 10)
	// if err != nil {
	// 	t.Logf("failed to execute copy task:\n  %v", err)
	// }
}

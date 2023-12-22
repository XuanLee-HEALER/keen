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
)

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
	os.Remove(fn)
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
	}

	for _, xt := range tasks {
		t.Log("task:", xt)
	}
	t.Log(strings.Repeat("=", 100))
	gct := util.SplitTasksToNGroups(tasks, 4)
	for _, g := range gct {
		t.Log("task group:", g)
	}

	t.Log(strings.Repeat("=", 100))
	err = util.ExecGroupCopyTasks(gct, 8096)
	if err != nil {
		t.Logf("failed to execute copy task:\n  %v", err)
	}
}

func TestTempCopy(t *testing.T) {
	tk := util.NewCopyTask("SRC/src_0", "DST/dst_0")
	tks := make([]*util.CopyTask, 0)
	tks = append(tks, tk)
	err := util.SetupCopyTasks(tks)
	if err != nil {
		t.Log(err)
	}
	gtk := util.SplitTasksToNGroups(tks, 4)
	err = util.ExecGroupCopyTasks(gtk, 8096)
	if err != nil {
		t.Log(err)
	}
}

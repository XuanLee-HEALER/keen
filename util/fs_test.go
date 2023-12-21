package util_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestCreateFixSizeFile(t *testing.T) {
	fn := "./xx"
	_, clean, err := util.CreateFixSizeFile(fn, 1000*util.GB)
	if err != nil {
		if util.IsSpaceNotEnough(err) {
			t.SkipNow()
		}
		t.Errorf("failed to create file: %v", err)
	}
	defer clean()
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
	cleans := make([]util.CleanFunc, 0)
	defer func() {
		for _, clean := range cleans {
			clean()
		}
	}()

	for i := 0; i < 10; i++ {
		curF := filepath.Join(SRC, SRCF+strconv.Itoa(i))
		df := filepath.Join(DST, DSTF+strconv.Itoa(i))
		// rd := 512*util.MB + util.FileSize(rand.Int63n(int64(512*util.MB+1)))
		// f, clean, err := util.CreateFixSizeFile(curF, rd)
		// if err != nil {
		// 	t.Error(err)
		// }
		// cleans = append(cleans, clean)

		// err = util.FillFile(f, rd)
		// if err != nil {
		// 	t.Error(err)
		// }

		tasks = append(tasks, util.NewCopyTask(curF, df))
	}

	// util.SetupCopyTasks(nt)

	for _, xt := range tasks {
		t.Log("task:", xt)
	}
}

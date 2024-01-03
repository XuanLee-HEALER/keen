//go:build linux || darwin || aix
// +build linux darwin aix

package util_test

import (
	"os"
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

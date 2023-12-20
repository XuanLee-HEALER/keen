package util_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestCreateFixSizeFile(t *testing.T) {
	fn := "./xx"
	f, err := util.CreateFixSizeFile(fn, 1000*util.GB)
	if err != nil {
		if util.IsSpaceNotEnough(err) {
			t.SkipNow()
		}
		t.Errorf("failed to create file: %v", err)
	}
	defer f.Close()
}

func TestCopy(t *testing.T) {
	// size := 650 * util.MB
	// f, err := util.CreateFixSizeFile("src", 650*util.MB)
	// if err != nil {
	// 	t.Error(err)
	// }
	// defer f.Close()

	// err = util.FillFile(f, size)
	// if err != nil {
	// 	t.Error(err)
	// }
	// 8C96E6C189F9202AD127052429A90201
	// 82B0186178EF3B1C1F12A6AD29FCFB16
	task := util.NewCopyTask("src", "dst")
	err := util.Copy(task, 3)
	if err != nil {
		t.Error(err)
	}
}

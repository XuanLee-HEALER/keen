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

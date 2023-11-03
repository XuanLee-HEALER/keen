// go:build aix darwin linux

package ysys_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/ysys"
)

func TestStatusSystemMemory(t *testing.T) {
	res, err := ysys.StatusSystemMemory()
	if err != nil {
		t.Error(err)
	}
	t.Log("total memory: ", ysys.TotalMemory(res))
	t.Logf("total memory: %.2f", float64(ysys.TotalMemory(res))/1024/1024/1024)
}

func TestDirSize(t *testing.T) {
	tdir := "/home/oracle/mount_point"
	size, err := ysys.DirSize(tdir)
	if err != nil {
		t.Logf("computation error: %v", err)
	}
	t.Log("size:", size)
}

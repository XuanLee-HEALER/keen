package nix_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/nix"
)

func TestStatusSystemMemory(t *testing.T) {
	res, err := nix.StatusSystemMemory()
	if err != nil {
		t.Error(err)
	}
	t.Log("total memory: ", nix.TotalMemory(res))
	t.Logf("total memory: %.2f", float64(nix.TotalMemory(res))/1024/1024/1024)
}

func TestDirSize(t *testing.T) {
	tdir := "/home/oracle/mount_point"
	size, err := nix.DirSize(tdir)
	if err != nil {
		t.Logf("computation error: %v", err)
	}
	t.Log("size:", size)
}

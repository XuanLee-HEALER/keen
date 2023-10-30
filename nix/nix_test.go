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

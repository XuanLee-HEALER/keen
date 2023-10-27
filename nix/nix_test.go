package nix_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/nix"
)

func TestStatusSystemMemory(t *testing.T) {
	_, err := nix.StatusSystemMemory()
	if err != nil {
		t.Error(err)
	}
}

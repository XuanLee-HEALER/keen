package win_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/win"
)

func TestPSVersionTable(t *testing.T) {
	v, err := win.PSVersionTable()
	if err != nil {
		t.Fatalf("failed to fetch powershell version: %v", err)
	}
	t.Log("version:", v)
}

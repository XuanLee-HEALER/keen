//go:build linux && aix && darwin
// +build linux,aix,darwin

package util_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestStatusProcess(t *testing.T) {
	res, err := util.QueryProcess("382015")
	if err != nil {
		t.Error("error: ", err)
	}
	t.Log("process detail: ", res)
}

func TestStatusSystemMemory(t *testing.T) {
	res, err := util.QuerySystemMemory()
	if err != nil {
		t.Error(err)
	}
	t.Log("result number: ", len(res))
}

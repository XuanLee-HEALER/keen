//go:build aix || darwin || linux
// +build aix darwin linux

package ysys_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/ysys"
)

func TestStatusProcess(t *testing.T) {
	res, err := ysys.QueryProcess("382015")
	if err != nil {
		t.Error("error: ", err)
	}
	t.Log("process detail: ", res)
}

func TestStatusSystemMemory(t *testing.T) {
	res, err := ysys.QuerySystemMemory()
	if err != nil {
		t.Error(err)
	}
	t.Log("result number: ", len(res))
}

package util_test

import (
	"errors"
	"io/fs"
	"testing"

	"gitea.fcdm.top/lixuan/keen/util"
)

func TestIterDir(t *testing.T) {
	path := "/Users/mouselee/Documents/yzcloud/lib/keen"
	err := util.IterDir(path, func(s string, de fs.DirEntry) bool { return false }, func(s string, de fs.DirEntry) error { println(s); return nil })
	if err != nil {
		t.Error(err)
	}
	t.Log("=============================================")
	err = util.IterDir(path, func(s string, de fs.DirEntry) bool { return de.IsDir() }, func(s string, de fs.DirEntry) error { println(s); return nil })
	if err != nil {
		t.Error(err)
	}
}

func TestDirSize(t *testing.T) {
	// 299565 299709
	path := "/Users/mouselee/Documents/yzcloud/lib/keen/tst"
	r, err := util.DirSize(path)
	if err != nil {
		t.Error(err)
	}
	t.Log("size: ", r)
}

func TestConvertGBKToUtf8(t *testing.T) {
	oc := "这是UTF8编码的字符串"
	ec, err := util.ConvertUtf8ToGBK(oc)
	if err != nil {
		t.Errorf("failed to encode: %v", err)
	}
	t.Log("gbk: ", ec)

	dc, err := util.ConvertGBKToUtf8(ec)
	if err != nil {
		t.Errorf("failed to decode: %v", err)
	}

	t.Log("utf-8:", dc)
}

func TestMinInt(t *testing.T) {
	arr := []int{2, 10, 1, 17, 17, 26, 22, 32, 1, 20, 30, 12, 28, 13, 4, 18, 31, 16, 11, 10}
	min, ok := util.MinInt(arr)
	t.Log("min", min)
	if !ok {
		t.Error(errors.New("valid input"))
	}

	if min != 1 {
		t.FailNow()
	}
}

func TestMinIntInvalidInput(t *testing.T) {
	arr := []int{}
	_, ok := util.MinInt(arr)
	if ok {
		t.Error(errors.New("invalid input"))
	}
}

func TestMaxInt(t *testing.T) {
	arr := []int{2, 10, 1, 17, 17, 26, 22, 32, 1, 20, 30, 12, 28, 13, 4, 18, 31, 16, 11, 10}
	max, ok := util.MaxInt(arr)
	t.Log("max", max)
	if !ok {
		t.Error(errors.New("valid input"))
	}

	if max != 32 {
		t.FailNow()
	}
}

func TestMaxIntInvalidInput(t *testing.T) {
	arr := []int{}
	_, ok := util.MaxInt(arr)
	if ok {
		t.Error(errors.New("invalid input"))
	}
}

func TestMinMaxInt(t *testing.T) {
	arr := []int{2, 10, 1, 17, 17, 26, 22, 32, 1, 20, 30, 12, 28, 13, 4, 18, 31, 16, 11, 10}
	min, max, ok := util.MinMaxInt(arr)
	t.Log("min", min, "max", max)
	if !ok {
		t.Error(errors.New("valid input"))
	}

	if min != 1 {
		t.FailNow()
	}
	if max != 32 {
		t.FailNow()
	}
}

func TestMinMaxIntInvalidInput(t *testing.T) {
	arr := []int{}
	_, _, ok := util.MinMaxInt(arr)
	if ok {
		t.Error(errors.New("invalid input"))
	}
}

package util_test

import (
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

package fsop_test

import (
	"io"
	"os"
	"testing"

	"gitea.fcdm.top/lixuan/keen/fsop"
)

func TestCreateRandomFile(t *testing.T) {
	err := fsop.CreateRandomFile("f4.txt", 2*fsop.MB)
	if err != nil {
		t.Fatal(err)
	}
	err = fsop.CreateRandomFile("f5.txt", 2*fsop.MB)
	if err != nil {
		t.Fatal(err)
	}
	err = fsop.CreateRandomFile("f6.txt", 2*fsop.MB)
	if err != nil {
		t.Fatal(err)
	}
}

// 6G / 107s
func TestCopyFiles(t *testing.T) {
	srcs := []string{"f4.txt", "f5.txt", "f6.txt"}
	dsts := []string{"F:\\go_projects\\keen\\f4.txt", "F:\\go_projects\\keen\\f5.txt", "F:\\go_projects\\keen\\f6.txt"}

	err := fsop.CopyFiles(dsts, srcs, 4*fsop.KB)
	if err != nil {
		if fsop.InsufficientDiskSpaceError(err) {
			t.Log("right error")
		}
		for _, dst := range dsts {
			xer := os.Remove(dst)
			if xer != nil {
				t.Logf("remove error: %v", err)
			}
		}
		t.Fatal(err)
	}
}

func TestCopyFiles1(t *testing.T) {
	srcs := []string{"F:\\go_projects\\keen\\fsop\\op.go", "F:\\go_projects\\keen\\fsop\\op_test.go"}
	dsts := []string{"F:\\go_projects\\keen\\f4.go", "F:\\go_projects\\keen\\f5.go"}

	err := fsop.CopyFiles(dsts, srcs, 4*fsop.KB)
	if err != nil {
		if fsop.InsufficientDiskSpaceError(err) {
			t.Log("right error")
		}
		for _, dst := range dsts {
			xer := os.Remove(dst)
			if xer != nil {
				t.Logf("remove error: %v", err)
			}
		}
		t.Fatal(err)
	}
}

// 6G / 166s
func TestSerialCopyFiles(t *testing.T) {
	srcs := []string{"f4.txt", "f5.txt", "f6.txt"}
	dsts := []string{"F:\\go_projects\\keen\\f4.txt", "F:\\go_projects\\keen\\f5.txt", "F:\\go_projects\\keen\\f6.txt"}

	for i := range srcs {
		f0, _ := os.Open(srcs[i])
		f1, _ := os.Create(dsts[i])
		_, err := io.Copy(f1, f0)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkCreateRandomFile(b *testing.B) {
	err := fsop.CreateRandomFile("f0.txt", 8*fsop.GB)
	if err != nil {
		os.Remove("f0.txt")
		b.Fatal(err)
	}
}

func TestPathExists(t *testing.T) {
	f := "C:\\ProgramData\\mewo.txt"
	if fsop.PathExists(f) {
		t.FailNow()
	}

	f = "C:\\ProgramData"
	if !fsop.PathExists(f) {
		t.FailNow()
	}
}

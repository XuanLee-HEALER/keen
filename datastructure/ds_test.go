package datastructure_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/datastructure"
)

func TestSetMerge(t *testing.T) {
	s1 := make(datastructure.Set[int])
	s2 := make(datastructure.Set[int])
	s1.AddList([]int{1, 2, 3})
	s2.AddList([]int{2, 3, 4})
	s1.Merge(s2)
	if len(s1) != 4 {
		t.Fail()
	}
	t.Logf("merge result: %v", s1)
}

func TestQueue(t *testing.T) {
	q := datastructure.NewQueue()

	q.Push(1)
	q.Push("something")
	q.Push(2)

	for !q.Empty() {
		t.Logf("value: %v", q.Remove())
	}
}

func TestStack(t *testing.T) {
	s := datastructure.NewStack()

	s.Push(1)
	s.Push("something")
	s.Push(2)
	s.Push(3)
	s.Push("anything")

	for !s.Empty() {
		t.Logf("value: %v", s.Remove())
	}
}

func TestDirTree(t *testing.T) {
	tree := datastructure.NewDirTree()
	dir1 := datastructure.NewDir("A")
	dir2 := datastructure.NewDir("B")
	file1 := datastructure.NewFile(&datastructure.SimpleFile{FileName: "B1"})
	file2 := datastructure.NewFile(&datastructure.SimpleFile{FileName: "B2"})
	dir2.AddFile(file1)
	dir2.AddFile(file2)
	file5 := datastructure.NewFile(&datastructure.SimpleFile{FileName: "C"})
	dir4 := datastructure.NewDir("D")
	dir5 := datastructure.NewDir("D1")
	dir6 := datastructure.NewDir("D2")
	file3 := datastructure.NewFile(&datastructure.SimpleFile{FileName: "E1"})
	file4 := datastructure.NewFile(&datastructure.SimpleFile{FileName: "E2"})
	dir5.AddFile(file3)
	dir5.AddFile(file4)
	dir4.AddDir(dir5)
	dir4.AddDir(dir6)
	tree.AddDir(dir1)
	tree.AddDir(dir2, "A")
	tree.AddFile(file5)
	tree.AddDir(dir4)
	tree.AddFile(datastructure.NewFile(&datastructure.SimpleFile{FileName: "B3"}), "A", "B")
	t.Logf("\n%s\n", tree)
}

func TestDirToTree(t *testing.T) {
	p := `F:\yzy\keen`
	tr, err := datastructure.ReadDirTree(p)
	if err != nil {
		t.Error(err)
	}

	t.Logf("\n%s\n", tr)
}

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

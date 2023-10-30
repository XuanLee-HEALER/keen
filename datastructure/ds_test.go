package datastructure_test

import (
	"testing"

	"gitea.fcdm.top/lixuan/keen/datastructure"
)

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

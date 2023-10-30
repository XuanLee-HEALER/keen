package datastructure

import (
	"container/list"
)

type Queue struct {
	l *list.List
}

func NewQueue() Queue {
	return Queue{
		l: new(list.List),
	}
}

func (q Queue) Empty() bool {
	return q.l.Front() == nil
}

func (q Queue) Push(e any) {
	q.l.PushBack(e)
}

func (q Queue) Remove() any {
	if q.l.Front() != nil {
		r := q.l.Front()
		v := r.Value
		q.l.Remove(r)
		return v
	}
	return nil
}

func (q Queue) Peek() any {
	if h := q.l.Front(); h != nil {
		return h.Value
	}
	return nil
}

type Stack struct {
	l *list.List
}

func NewStack() Stack {
	return Stack{
		l: new(list.List),
	}
}

func (s Stack) Empty() bool {
	return s.l.Front() == nil
}

func (s Stack) Push(e any) {
	s.l.PushFront(e)
}

func (s Stack) Remove() any {
	if s.l.Front() != nil {
		r := s.l.Front()
		v := r.Value
		s.l.Remove(r)
		return v
	}
	return nil
}

func (s Stack) Peek() any {
	if h := s.l.Front(); h != nil {
		return h.Value
	}
	return nil
}

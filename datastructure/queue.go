package datastructure

import (
	"container/list"
	"sync"

	pq "github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/emirpasic/gods/utils"
)

type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{
		l: new(list.List),
	}
}

func (q *Queue) Enqueue(v any) {
	q.l.PushBack(v)
}

func (q *Queue) Dequeue() any {
	if q.l.Front() != nil {
		return q.l.Remove(q.l.Front())
	}
	return nil
}

func (q *Queue) Peek() any {
	if h := q.l.Front(); h != nil {
		return h.Value
	}
	return nil
}

func (q *Queue) Empty() bool {
	return q.l.Front() == nil
}

type ConcurrentQueue struct {
	q *Queue

	rwMut *sync.RWMutex
}

func NewConcurrentQueue() *ConcurrentQueue {
	return &ConcurrentQueue{
		NewQueue(),
		&sync.RWMutex{},
	}
}

func (q ConcurrentQueue) Enqueue(v any) {
	q.rwMut.Lock()
	defer q.rwMut.Unlock()
	q.q.Enqueue(v)
}

func (q ConcurrentQueue) Dequeue() any {
	q.rwMut.Lock()
	defer q.rwMut.Unlock()

	return q.q.Dequeue()
}

func (q ConcurrentQueue) Peek() any {
	q.rwMut.RLock()
	defer q.rwMut.RUnlock()

	return q.q.Peek()
}

func (q ConcurrentQueue) Empty() bool {
	q.rwMut.RLock()
	defer q.rwMut.RUnlock()

	return q.q.Empty()
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

type ConcurrentPriorityQueue struct {
	q pq.Queue

	rwMut *sync.RWMutex
}

func NewConcurrentPriorityQueue(cmp utils.Comparator) *ConcurrentPriorityQueue {
	return &ConcurrentPriorityQueue{
		*pq.NewWith(cmp),
		&sync.RWMutex{},
	}
}

func (q *ConcurrentPriorityQueue) Enqueue(v any) {
	q.rwMut.Lock()
	defer q.rwMut.Unlock()

	q.q.Enqueue(v)
}

func (q *ConcurrentPriorityQueue) Dequeue() any {
	q.rwMut.Lock()
	defer q.rwMut.Unlock()

	if v, ok := q.q.Dequeue(); ok {
		return v
	}
	return nil
}

func (q *ConcurrentPriorityQueue) Peek() any {
	q.rwMut.RLock()
	defer q.rwMut.RUnlock()

	if v, ok := q.q.Peek(); ok {
		return v
	}
	return nil
}

func (q *ConcurrentPriorityQueue) Empty() bool {
	q.rwMut.RLock()
	defer q.rwMut.RUnlock()

	return q.q.Empty()
}

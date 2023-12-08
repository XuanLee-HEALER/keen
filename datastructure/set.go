package datastructure

import (
	"fmt"
	"strings"
)

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (set Set[T]) String() string {
	strbuf := new(strings.Builder)
	strbuf.WriteString("(")
	for k := range set {
		strbuf.WriteString(fmt.Sprintf("%v, ", k))
	}
	if strbuf.Len() != 1 {
		strbuf.WriteString("\b\b")
	}
	strbuf.WriteString(")")
	return strbuf.String()
}

func (set Set[T]) Contains(t T) bool {
	if _, ok := set[t]; ok {
		return true
	}
	return false
}

func (set Set[T]) Add(t T) {
	set[t] = struct{}{}
}

func (set Set[T]) AddList(ts []T) {
	for _, t := range ts {
		set.Add(t)
	}
}

func (set Set[T]) Merge(o Set[T]) {
	for e := range o {
		set.Add(e)
	}
}

func (set Set[T]) ToSlice() []T {
	res := make([]T, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	return res
}

func ToSet[T comparable](lst []T) Set[T] {
	s := make(Set[T])
	for _, e := range lst {
		s.Add(e)
	}
	return s
}

package datastructure

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
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

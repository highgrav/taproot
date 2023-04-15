package common

import "fmt"

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](vals ...T) Set[T] {
	set := Set[T]{}
	set.Add(vals...)
	return set
}

func (set Set[T]) Add(vals ...T) {
	for _, val := range vals {
		set[val] = struct{}{}
	}
}

func (set Set[T]) Contains(val T) bool {
	_, ok := set[val]
	return ok
}

func (set Set[T]) All() []T {
	arr := make([]T, len(set))
	x := 0
	for val, _ := range set {
		arr[x] = val
		x++
	}
	return arr
}

func (set Set[T]) String() string {
	return fmt.Sprintf("%v", set.All())
}

// Union() returns the union of sets A and B.
func (set Set[T]) Union(withSet Set[T]) Set[T] {
	res := NewSet[T](set.All()...)
	res.Add(withSet.All()...)
	return res
}

// Intersect() calculates everything that is in A and also in B.
func (set Set[T]) Intersect(withSet Set[T]) Set[T] {
	res := NewSet[T]()
	for val, _ := range set {
		if withSet.Contains(val) {
			res.Add(val)
		}
	}
	return res
}

// Difference() calculates everything that is in A (set) but not B (fromSet).
func (set Set[T]) Difference(fromSet Set[T]) Set[T] {
	res := NewSet[T]()
	for val, _ := range set {
		if !fromSet.Contains(val) {
			res.Add(val)
		}
	}
	return res
}

package nimble

import (
	"cmp"
	"slices"
	"sort"
)

type SortedList[T cmp.Ordered] struct {
	list []T
	min  T
	max  T
}

// NewSortedList constructs a sorted list.  A sorted list maintains order upon every added item, making it ideal
// for accesing the smallest and largest items at any given time.
func NewSortedList[T cmp.Ordered](items ...T) *SortedList[T] {
	l := &SortedList[T]{}

	if len(items) > 0 {
		l.list = items[:]
		sort.Slice(l.list, func(i, j int) bool {
			return l.list[i] < l.list[j]
		})
		l.min = l.list[0]
		l.max = l.list[len(l.list)-1]
	} else {
		l.list = make([]T, 0, 10)
	}

	return l
}

func (l *SortedList[T]) Extend(items ...T) {
	for _, item := range items {
		l.insert(item)
	}
}

func (l *SortedList[T]) Length() int {
	return len(l.list)
}

func (l *SortedList[T]) Get(index int) T {
	return l.list[index]
}

func (l *SortedList[T]) Smallest() T {
	if len(l.list) > 0 {
		return l.list[0]
	}
	var v T
	return v
}

func (l *SortedList[T]) Largest() T {
	if len(l.list) > 0 {
		return l.list[len(l.list)-1]
	}
	var v T
	return v
}

func (l *SortedList[T]) PopSmallest() T {
	v := l.list[0]
	l.list = l.list[1:]
	if len(l.list) > 0 {
		l.min = l.list[0]
	}

	return v
}

func (l *SortedList[T]) PopLargest() T {
	lastElem := len(l.list) - 1
	v := l.list[lastElem]
	l.list = l.list[:lastElem]
	if len(l.list) > 0 {
		l.max = l.list[lastElem-1]
	}

	return v
}

func (l *SortedList[T]) insert(item T) {
	if len(l.list) == 0 {
		l.list = append(l.list, item)
		l.min = item
		l.max = item
		return
	}

	if item <= l.min {
		l.min = item
		l.list = slices.Insert(l.list, 0, item)
		return
	}

	if item >= l.max {
		l.max = item
		l.list = append(l.list, item)
		return
	}

	start := 0
	end := len(l.list) - 1

	for end-start > 1 {
		midPoint := start + (end-start)/2
		value := l.list[midPoint]
		if item > value {
			start = midPoint
		} else if item < value {
			end = midPoint
		} else {
			// if we reach equality, we can stop bisecting and insert here
			start = midPoint
			end = midPoint
		}
	}

	// perform insertion
	l.list = slices.Insert(l.list, end, item)
}

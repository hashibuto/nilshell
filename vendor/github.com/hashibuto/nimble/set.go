package nimble

import "sort"

type Set[T comparable] struct {
	collection map[T]struct{}
}

// NewSet creates a new set, initialized with values
func NewSet[T comparable](values ...T) *Set[T] {
	collection := map[T]struct{}{}
	for _, v := range values {
		collection[v] = struct{}{}
	}
	return &Set[T]{
		collection: collection,
	}
}

func NewSetFromMap[T comparable, X any](values map[T]X) *Set[T] {
	collection := map[T]struct{}{}
	for v := range values {
		collection[v] = struct{}{}
	}
	return &Set[T]{
		collection: collection,
	}
}

// Copy makes a shallow copy of the current set and returns it
func (s *Set[T]) Copy() *Set[T] {
	c := map[T]struct{}{}
	for k, v := range s.collection {
		c[k] = v
	}

	return &Set[T]{
		collection: c,
	}
}

// Add adds a single value to the set
func (s *Set[T]) Add(values ...T) {
	for _, value := range values {
		s.collection[value] = struct{}{}
	}
}

// Has returns true if the set already contains the provided value
func (s *Set[T]) Has(value T) bool {
	_, exists := s.collection[value]
	return exists
}

// Remove removes the specified item from the set
func (s *Set[T]) Remove(value T) {
	delete(s.collection, value)
}

// Size returns the size of the set
func (s *Set[T]) Size() int {
	return len(s.collection)
}

// Items returns a slice of the items in the set
func (s *Set[T]) Items() []T {
	items := make([]T, len(s.collection))
	idx := 0
	for item := range s.collection {
		items[idx] = item
		idx++
	}

	return items
}

func (s *Set[T]) Intersect(set *Set[T]) *Set[T] {
	var setA, setB *Set[T]
	if s.Size() < set.Size() {
		setA = s
		setB = set
	} else {
		setA = set
		setB = s
	}

	collection := map[T]struct{}{}
	for k := range setA.collection {
		if setB.Has(k) {
			collection[k] = struct{}{}
		}
	}

	return &Set[T]{
		collection: collection,
	}
}

func (s *Set[T]) Difference(set ...*Set[T]) *Set[T] {
	if len(set) == 0 {
		ns := NewSet[T]()
		for item := range s.collection {
			ns.collection[item] = struct{}{}
		}
		return ns
	}

	newSet := NewSet[T]()
	for key := range s.collection {
		for _, inset := range set {
			if inset.Has(key) {
				goto skipKey
			}
		}
		newSet.Add(key)
	skipKey:
	}
	return newSet
}

// Pop removes and returns an arbitrary item from the set
func (s *Set[T]) Pop() any {
	if len(s.collection) == 0 {
		return nil
	}

	var selected T
	for key := range s.collection {
		selected = key
		break
	}
	delete(s.collection, selected)

	return selected
}

// Union returns a new set that is a union of the members of the provided sets
func Union[T comparable](set ...*Set[T]) *Set[T] {
	newSet := NewSet[T]()
	for _, s := range set {
		for k := range s.collection {
			newSet.Add(k)
		}
	}

	return newSet
}

// Intersect returns the intersection of all provided sets
func Intersect[T comparable](set ...*Set[T]) *Set[T] {
	if len(set) == 0 {
		return NewSet[T]()
	} else if len(set) == 1 {
		s := NewSet[T]()
		for item := range set[0].collection {
			s.collection[item] = struct{}{}
		}
		return s
	}

	sort.Slice(set, func(i, j int) bool {
		return set[i].Size() < set[j].Size()
	})

	var newSet *Set[T]
	for i := 0; i < len(set)-1; i++ {
		newSet = set[i].Intersect(set[i+1])
	}
	return newSet

}

package nimble

import (
	"strings"
	"time"
)

type IndexedDequeueLink struct {
	CreatedAt time.Time
	Value     string
	prev      *IndexedDequeueLink
	next      *IndexedDequeueLink
	nGrams    []string
}

type IndexedDequeue struct {
	head   *IndexedDequeueLink
	tail   *IndexedDequeueLink
	nGrams map[string]*Set[*IndexedDequeueLink]
	size   int
}

type IndexedDequeIterator struct {
	head *IndexedDequeueLink
	tail *IndexedDequeueLink
	link *IndexedDequeueLink
}

func (iqi *IndexedDequeIterator) Next() bool {
	if iqi.link == nil {
		iqi.link = iqi.tail
		if iqi.link == nil {
			return false
		}
	} else {
		if iqi.link.next == nil {
			return false
		}
		iqi.link = iqi.link.next
	}

	return true
}

func (iqi *IndexedDequeIterator) ReverseNext() bool {
	if iqi.link == nil {
		iqi.link = iqi.head
		if iqi.link == nil {
			return false
		}
	} else {
		if iqi.link.prev == nil {
			return false
		}
		iqi.link = iqi.link.prev
	}

	return true
}

func (iqi *IndexedDequeIterator) Value() string {
	return iqi.link.Value
}

// NewIndexedDequeue creates an indexed deque, whereby items can be accessed in constant time from the head or tail,
// and arbitrary items can be accessed in near constant time using a string lookup, where lookup time varies slightly by
// the number of matching candidates.
// pop => prev(nil) <-- tail <-- prev next --> head --> next(nil) <= push
func NewIndexedDequeue(items ...string) *IndexedDequeue {
	idq := &IndexedDequeue{
		nGrams: map[string]*Set[*IndexedDequeueLink]{},
	}

	for _, item := range items {
		idq.Push(item)
	}

	return idq
}

func (iq *IndexedDequeue) Push(value string) {
	link := &IndexedDequeueLink{
		CreatedAt: time.Now(),
		Value:     value,
		nGrams:    iq.computeNGrams(value),
	}

	if iq.head == nil {
		iq.head = link
		iq.tail = link
	} else {
		link.prev = iq.head
		iq.head.next = link
		iq.head = link
	}

	for _, nGram := range link.nGrams {
		set, ok := iq.nGrams[nGram]
		if !ok {
			set = NewSet[*IndexedDequeueLink]()
			iq.nGrams[nGram] = set
		}
		set.Add(link)
	}

	iq.size++
}

func (iq *IndexedDequeue) Pop() string {
	if iq.size == 0 {
		panic("the collection is empty, nothing to pop")
	}

	value := iq.tail.Value
	iq.RemoveItem(iq.tail)

	return value
}

func (iq *IndexedDequeue) Find(pattern string) []*IndexedDequeueLink {
	var base *Set[*IndexedDequeueLink]
	nGrams := iq.computeNGrams(pattern)
	for _, nGram := range nGrams {
		if set, ok := iq.nGrams[nGram]; ok {
			if base == nil {
				base = set
			} else {
				base = base.Intersect(set)
				if base.Size() == 0 {
					return nil
				}
			}
		} else {
			return nil
		}
	}

	return base.Items()
}

func (iq *IndexedDequeue) RemoveItem(links ...*IndexedDequeueLink) {
	for _, link := range links {
		if link.prev == nil {
			// this is a tail link
			iq.tail = link.next
			if link.next != nil {
				link.next.prev = nil
			}
		} else {
			if link.next != nil {
				link.next.prev = link.prev
				link.prev.next = link.next
			}
		}

		if link.next == nil {
			// this is a head link
			iq.head = link.prev
			if link.prev != nil {
				link.prev.next = nil
			}
		} else {
			if link.prev != nil {
				link.prev.next = link.next
				link.next.prev = link.prev
			}
		}

		iq.size--
		for _, nGram := range link.nGrams {
			set := iq.nGrams[nGram]
			set.Remove(link)
			if set.Size() == 0 {
				delete(iq.nGrams, nGram)
			}
		}
	}
}

func (iq *IndexedDequeue) Size() int {
	return iq.size
}

// GetIter returns an iterator
func (iq *IndexedDequeue) GetIter() *IndexedDequeIterator {
	return &IndexedDequeIterator{
		head: iq.head,
		tail: iq.tail,
	}
}

func (iq *IndexedDequeue) computeNGrams(value string) []string {
	value = strings.ToLower(value)
	ngrams := NewSet[string]()
	for i := 0; i < len(value); i++ {
		for j := i; j < len(value); j++ {
			s := value[i : j+1]
			if ngrams.Has(s) {
				continue
			}

			ngrams.Add(s)
		}
	}

	return ngrams.Items()
}

package nimble

type Link[K comparable, V any] struct {
	Key   K
	Value V
	Next  *Link[K, V]
	Prev  *Link[K, V]
}

type OrderedMap[K comparable, V any] struct {
	oMap      map[K]*Link[K, V]
	oLinkHead *Link[K, V]
	oLinkTail *Link[K, V]
}

type OrderedMapIter[K comparable, V any] struct {
	headLink *Link[K, V]
	tailLink *Link[K, V]
	curLink  *Link[K, V]
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		oMap: map[K]*Link[K, V]{},
	}
}

func (om *OrderedMap[K, V]) Put(key K, value V) {
	// When putting a key that already exists, we update the existing link
	if link, exists := om.oMap[key]; exists {
		link.Value = value
		return
	}

	link := &Link[K, V]{
		Key:   key,
		Value: value,
	}
	om.oMap[key] = link

	if om.oLinkHead == nil {
		om.oLinkHead = link
	} else {
		link.Prev = om.oLinkTail
		om.oLinkTail.Next = link
	}
	om.oLinkTail = link
}

func (om *OrderedMap[K, V]) Get(key K) (V, bool) {
	link, exists := om.oMap[key]
	if !exists {
		var ret V
		return ret, false
	}

	return link.Value, true
}

func (om *OrderedMap[K, V]) Delete(key K) {
	link, exists := om.oMap[key]
	if !exists {
		return
	}

	// unlink the ordered item
	prevLink := link.Prev
	nextLink := link.Next

	if prevLink != nil {
		prevLink.Next = nextLink
	} else {
		om.oLinkHead = nextLink
	}
	if nextLink != nil {
		nextLink.Prev = prevLink
	} else {
		om.oLinkTail = prevLink
	}

	delete(om.oMap, key)
}

func (om *OrderedMap[K, V]) Length() int {
	return len(om.oMap)
}

// GetIter returns an iterator which allows for forward or reverse iteration through the key/value pairs as they were added to the collection
func (om *OrderedMap[K, V]) GetIter() *OrderedMapIter[K, V] {
	return &OrderedMapIter[K, V]{
		headLink: om.oLinkHead,
		tailLink: om.oLinkTail,
	}
}

func (i *OrderedMapIter[K, V]) Key() K {
	return i.curLink.Key
}

func (i *OrderedMapIter[K, V]) Value() V {
	return i.curLink.Value
}

// Next advances to the next position in the iterator.  This must be called before the iterator's Key() or Value() functions can be consumed.
// This lends well to using Next() in a for loop
func (i *OrderedMapIter[K, V]) Next() bool {
	if i.curLink == nil {
		i.curLink = i.headLink
		if i.curLink == nil {
			return false
		}
	} else {
		if i.curLink.Next == nil {
			return false
		}
		i.curLink = i.curLink.Next
	}

	return true
}

// ReverseNext advances to the next position in the iterator, starting from the end going to the beginning.
// This must be called before the iterator's Key() or Value() functions can be consumed. This lends well to using ReverseNext() in a for loop
func (i *OrderedMapIter[K, V]) ReverseNext() bool {
	if i.curLink == nil {
		i.curLink = i.tailLink
		if i.curLink == nil {
			return false
		}
	} else {
		if i.curLink.Prev == nil {
			return false
		}
		i.curLink = i.curLink.Prev
	}

	return true
}

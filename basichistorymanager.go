package ns

import (
	"sort"

	"github.com/hashibuto/nimble"
)

type BasicHistoryManager struct {
	index   *nimble.IndexedDequeue
	maxKeep int
	prev    string
}

type BasicHistoryIterator struct {
	iter *nimble.IndexedDequeIterator
}

func (bhi *BasicHistoryIterator) Forward() string {
	if bhi.iter == nil {
		return ""
	}

	bhi.iter.Next()
	return bhi.iter.Value()
}

func (bhi *BasicHistoryIterator) Backward() string {
	if bhi.iter == nil {
		return ""
	}

	bhi.iter.ReverseNext()
	return bhi.iter.Value()
}

func NewBasicHistoryManager(maxKeep int) *BasicHistoryManager {
	return &BasicHistoryManager{
		index:   nimble.NewIndexedDequeue(),
		maxKeep: maxKeep,
	}
}

func (h *BasicHistoryManager) Push(value string) {
	if value == h.prev {
		return
	}
	h.prev = value
	h.index.Push(value)
	if h.index.Size() > h.maxKeep {
		h.index.Pop()
	}
}

func (h *BasicHistoryManager) GetIterator() HistoryIterator {
	if h.index.Size() > 0 {
		return &BasicHistoryIterator{
			iter: h.index.GetIter(),
		}
	}

	return &BasicHistoryIterator{}
}

func (h *BasicHistoryManager) Search(pattern string) []string {
	if h.index.Size() == 0 {
		return nil
	}

	if len(pattern) == 0 {
		return nil
	}

	links := h.index.Find(pattern)
	if len(links) == 0 {
		return nil
	}

	// sort from most recent to oldest
	sort.Slice(links, func(i, j int) bool {
		return links[i].CreatedAt.After(links[j].CreatedAt)
	})

	strs := make([]string, len(links))
	for i, link := range links {
		strs[i] = link.Value
	}

	return strs
}

func (h *BasicHistoryManager) Exit() {

}

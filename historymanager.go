package ns

import (
	"sort"

	"github.com/hashibuto/nimble"
)

type HistoryIterator interface {
	Backward() string
	Forward() string
}

type HistoryManager interface {
	GetIterator() HistoryIterator
	Push(string)
	Search(string) []string
}

type BasicHistoryManager struct {
	index *nimble.IndexedDequeue
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

func NewBasicHistoryManager() *BasicHistoryManager {
	return &BasicHistoryManager{
		index: nimble.NewIndexedDequeue(),
	}
}

func (h *BasicHistoryManager) Push(value string) {
	h.index.Push(value)
	if h.index.Size() > 100 {
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

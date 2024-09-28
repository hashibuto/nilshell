package ns

import "fmt"

type PersistedHistoryManager struct {
	HistoryManager
	filename string
	isDirty  bool
	fileLock *FileLock
}

func NewPersistedHistoryManager(maxKeep int, filename string) *PersistedHistoryManager {

	pm := &PersistedHistoryManager{
		HistoryManager: NewBasicHistoryManager(maxKeep),
		filename:       filename,
		fileLock:       NewFileLock(fmt.Sprintf("%s.lock", filename)),
	}

	return pm
}

func (pm *PersistedHistoryManager) Push(value string) {
	pm.isDirty = true
	pm.HistoryManager.Push(value)
}

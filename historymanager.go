package ns

type HistoryIterator interface {
	Backward() string
	Forward() string
}

type HistoryManager interface {
	Exit()
	GetIterator() HistoryIterator
	Push(string)
	Search(string) []string
}

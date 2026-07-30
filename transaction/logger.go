package transaction

type EventType int

const (
	_ EventType = iota
	EventPut
	EventDelete
)

type Event struct {
	Sequence uint64
	EventType
	Key   string
	Value string
}

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	ReadEvents() (<-chan Event, <-chan error)
	Run()
}

package log

import (
	"fmt"
	"sync"
)

type wal struct {
	logs                     []*LogEntry
	snapshotLogLength        int
	snapshotLastIncludedTerm int
	mu                       sync.RWMutex
}

func (log *wal) String() string {
	return fmt.Sprintf("logs: %v", log.logs)
}

type CreateLogEntry struct {
	Term    int
	PayLoad interface{}
}

type LogEntry struct {
	Id      int
	Term    int
	Payload interface{}
}

func (e *LogEntry) String() string {
	return fmt.Sprintf("{ Id: %v, term: %v, payload %v }", e.Id, e.Term, e.Payload)
}

func (entry *LogEntry) GetId() int {
	return entry.Id
}

func (entry *LogEntry) GetTerm() int {
	return entry.Term
}

func (entry *LogEntry) GetPayload() interface{} {
	return entry.Payload
}

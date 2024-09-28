package log

import (
	"fmt"
)

func (log *wal) Append(entry *CreateLogEntry) int {
	log.mu.Lock()
	defer log.mu.Unlock()

	return log.appendInternal(entry)
}

func (log *wal) AppendEntries(entries ...*CreateLogEntry) int {
	log.mu.Lock()
	defer log.mu.Unlock()

	for _, e := range entries {
		log.appendInternal(e)
	}

	return log.lastIndexInternal()
}

func (log *wal) ReplaceAt(index int, entries ...*CreateLogEntry) (int, error) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if err := log.assertValidIndex(index); err != nil {
		return -1, err
	}

	return log.replaceAtInternal(index, entries...), nil
}

func (log *wal) GetEntry(index int) (*LogEntry, error) {
	log.mu.RLock()
	defer log.mu.RUnlock()

	if err := log.assertValidIndex(index); err != nil {
		return nil, err
	}

	return log.getEntryInternal(index), nil
}

func (log *wal) GetEntries(start int, end int) ([]*LogEntry, error) {
	log.mu.RLock()
	defer log.mu.RUnlock()

	if err := log.assertValidRange(start, end); err != nil {
		return nil, err
	}

	return log.getEntriesInternal(start, end), nil
}

func (log *wal) GetLast() *LogEntry {
	log.mu.RLock()
	defer log.mu.RUnlock()

	index := log.lastIndexInternal()
	if err := log.assertValidIndex(index); err != nil {
		panic(err)
	}

	return log.getEntryInternal(log.lastIndexInternal())
}

func (log *wal) GetTerm(index int) (int, error) {
	log.mu.RLock()
	defer log.mu.RUnlock()

	if index == log.snapshotLogLength-1 {
		return log.snapshotLastIncludedTerm, nil
	}

	if err := log.assertValidIndex(index); err != nil {
		return -1, err
	}

	return log.getEntryInternal(index).Term, nil
}

// TODO: remove error from here
func (log *wal) GetLastIndexAndTerm() (int, int, error) {
	log.mu.RLock()
	defer log.mu.RUnlock()

	index := log.lastIndexInternal()
	if index == log.snapshotLogLength-1 {
		return index, log.snapshotLastIncludedTerm, nil
	}

	if err := log.assertValidIndex(index); err != nil {
		return -1, -1, err
	}

	return index, log.getEntryInternal(index).Term, nil
}

func (log *wal) Len() int {
	log.mu.RLock()
	defer log.mu.RUnlock()
	return log.snapshotLogLength + len(log.logs)
}

func (log *wal) SnapshotLength() int {
	log.mu.RLock()
	defer log.mu.RUnlock()
	return log.snapshotLogLength
}

func (log *wal) FirstUniqueTermOnTheLeft(index int) (int, int, error) {
	log.mu.RLock()
	defer log.mu.RUnlock()

	if err := log.assertValidIndex(index); err != nil {
		return -1, -1, err
	}

	if len(log.logs) == 1 {
		return log.snapshotLogLength, log.snapshotLastIncludedTerm, nil
	}

	relativeIndex := index - log.snapshotLogLength
	for ; relativeIndex > 0 && log.logs[relativeIndex].Term == log.logs[relativeIndex-1].Term; relativeIndex-- {
	}

	// unique entry not found, but leader and peer MUST agree on entry from snapshot.
	if relativeIndex == 0 && log.logs[relativeIndex].Term == log.logs[relativeIndex+1].Term {
		return log.snapshotLogLength - 1, log.snapshotLastIncludedTerm, nil
	}

	return relativeIndex + log.snapshotLogLength, log.logs[relativeIndex].Term, nil
}

func (log *wal) Encode() []byte {
	log.mu.RLock()
	defer log.mu.RUnlock()
	return nil
}

func (log *wal) Decode(bytes []byte) {

}

func (log *wal) InstallSnapshot(index int, term int) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if index < log.snapshotLogLength {
		panic(&WALError{ErrorType: IndexOutOfBounds, msg: fmt.Sprintf("Index %v is less then snapshot length %v", index, log.snapshotLogLength)})
	}

	log.installSnapshotInternal(index, term)
}

func (log *wal) GetSnapshotLastTerm() int {
	log.mu.RLock()
	defer log.mu.RUnlock()

	return log.snapshotLastIncludedTerm
}

func (log *wal) appendInternal(entry *CreateLogEntry) int {
	newId := log.snapshotLogLength + len(log.logs)
	walEntry := &LogEntry{
		Id:      newId,
		Payload: entry.PayLoad,
		Term:    entry.Term,
	}
	log.logs = append(log.logs, walEntry)

	return newId
}

func (log *wal) lastIndexInternal() int {
	return log.snapshotLogLength + len(log.logs) - 1
}

func (log *wal) assertValidIndex(index int) error {
	if index < log.snapshotLogLength {
		return &WALError{msg: fmt.Sprintf("Index %v is out of snapshot bounds: %v", index, log.snapshotLogLength), ErrorType: IndexOutOfSnapshotBounds}
	}

	if index > log.lastIndexInternal() {
		return &WALError{msg: fmt.Sprintf("Index %v is out of log bounds: last index %v", index, log.lastIndexInternal()), ErrorType: IndexOutOfBounds}
	}

	return nil
}

func (log *wal) assertValidRange(start, end int) error {
	if start > end {
		return &WALError{msg: fmt.Sprintf("Invalid range: start %v >= %v end", start, end), ErrorType: InvalidRange}
	}

	if start < log.snapshotLogLength {
		return &WALError{msg: fmt.Sprintf("Index %v is out of snapshot bounds: %v", start, log.snapshotLogLength), ErrorType: IndexOutOfSnapshotBounds}
	}

	if end < log.snapshotLogLength {
		return &WALError{msg: fmt.Sprintf("Index %v is out of snapshot bounds: %v", end, log.snapshotLogLength), ErrorType: IndexOutOfSnapshotBounds}
	}

	// allow to be equal to length
	if start-1 > log.lastIndexInternal() {
		return &WALError{msg: fmt.Sprintf("Index %v is out of log bounds: last index %v", start, log.lastIndexInternal()), ErrorType: IndexOutOfBounds}
	}

	// allow to be equal to length
	if end-1 > log.lastIndexInternal() {
		return &WALError{msg: fmt.Sprintf("Index %v is out of log bounds: last index %v", end, log.lastIndexInternal()), ErrorType: IndexOutOfBounds}
	}

	return nil
}

func (log *wal) replaceAtInternal(index int, entries ...*CreateLogEntry) int {
	relativeIndex := index - log.snapshotLogLength
	newLog := make([]*LogEntry, relativeIndex)
	copy(newLog, log.logs[:relativeIndex])
	log.logs = newLog

	for _, e := range entries {
		log.appendInternal(e)
	}
	return log.lastIndexInternal()
}

func (log *wal) getEntryInternal(index int) *LogEntry {
	relativeLogIndex := index - log.snapshotLogLength
	return log.logs[relativeLogIndex]
}

func (log *wal) getEntriesInternal(start int, end int) []*LogEntry {
	rStart := start - log.snapshotLogLength
	rEnd := end - log.snapshotLogLength
	return log.logs[rStart:rEnd]
}

func (log *wal) installSnapshotInternal(index int, term int) {
	relativeIndex := min(index-log.snapshotLogLength, len(log.logs)-1)

	discardedLog := log.logs[relativeIndex+1:]
	newLog := make([]*LogEntry, len(discardedLog))
	copy(newLog, discardedLog)

	log.logs = newLog

	log.snapshotLastIncludedTerm = term
	log.snapshotLogLength = index + 1
}

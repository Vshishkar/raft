package log

import "sync"

type Log interface {
	/*
		Appends entry to the log.
		Returns index of new entry.
		(returns absolute index: snapshotLogLength + len(log))
	*/
	Append(entry *CreateLogEntry) int

	/*
		Appends multiple entries to the log
		Returns index of last appended entry
		(returns absolute index: snapshotLogLength + len(log))
	*/
	AppendEntries(entries ...*CreateLogEntry) int

	/*
		Replaces log entries starting from index.
		log[index:] = nil
		log.append(entries)

		Returns the last index of new entry
		(returns absolute index: snapshotLogLength + len(log))

		Returns error if index is not within [snapshotLength, logLength)
	*/
	ReplaceAt(index int, entries ...*CreateLogEntry) (int, error)

	/*
		Retrieves an entry at index.
		Returns error if index is not within [snapshotLength, logLength)
	*/
	GetEntry(index int) (*LogEntry, error)

	/*
		Retrieves entries from [start, end)
		Returns error if start < snapshotLogLength or end >= logLength
	*/
	GetEntries(start, end int) ([]*LogEntry, error)

	/*
		Retrieves last entry if log is not empty.
		Otherwise returns error
	*/
	GetLast() *LogEntry

	/*
		Retrieves term of entry if index is within [snapshotLogLength - 1, logLength)
		Note: index allowed to be snapshotLogLength - 1, because for this entry term is stored in separate variable.
		Otherwise returns an error
	*/
	GetTerm(index int) (int, error)

	/*
		Retrieves last term and last index.
		Note: last index allowed to be snapshotLogLength - 1, because for this entry term is stored in separate variable.
		Otherwise returns an error
	*/
	GetLastIndexAndTerm() (int, int, error)

	/*
		Returns the length of a log
	*/
	Len() int

	/*
		Returns the length of current snapshot
	*/
	SnapshotLength() int

	/*
		For a give index finds the first entry on the left with different term.
		Returns an error if index is out of bounds or there is no such term.
	*/
	FirstUniqueTermOnTheLeft(index int) (int, int, error)

	/*
		Encodes log data
	*/
	Encode() []byte

	/*
		Decodes log data
	*/
	Decode(bytes []byte)

	/*
		Discards log prefix until index.
		log = log[index + 1:]

		Updates snapshot length and snapshot last included term
	*/
	InstallSnapshot(index int, term int)

	/*
		Retrieves snapshot last term
	*/
	GetSnapshotLastTerm() int
}

func InitWriteAheadLog() Log {
	a := &wal{
		logs:              make([]*LogEntry, 0),
		snapshotLogLength: 0,
		mu:                sync.RWMutex{},
	}

	return a
}

type WALError struct {
	ErrorType WALErrorType
	msg       string
}

func (e *WALError) Error() string {
	return e.msg
}

type WALErrorType uint8

const (
	IndexOutOfSnapshotBounds WALErrorType = iota
	IndexOutOfBounds
	InvalidRange
)

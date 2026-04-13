package ops

import (
	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// OpType distinguishes copy/move/delete.
type OpType int

const (
	OpCopy OpType = iota
	OpMove
	OpDelete
)

// OpStatus is the lifecycle state of an operation.
type OpStatus int

const (
	StatusRunning OpStatus = iota
	StatusDone
	StatusFailed
)

// Operation tracks a running file operation.
type Operation struct {
	ID         string
	Type       OpType
	Sources    []string
	Dest       string
	SrcFS      fs.FileSystem
	DstFS      fs.FileSystem
	TotalBytes int64
	DoneBytes  int64
	Status     OpStatus
	Error      error
}

// ProgressMsg is sent to bubbletea to report progress.
type ProgressMsg struct {
	OpID        string
	DoneBytes   int64
	TotalBytes  int64
	Status      OpStatus
	Err         error
	CurrentFile string
}

// ProgressStreamMsg wraps a channel to read continuous progress.
type ProgressStreamMsg struct {
	C chan tea.Msg
}

// ClipOp distinguishes yank vs cut.
type ClipOp int

const (
	ClipCopy ClipOp = iota
	ClipCut
)

// Clipboard holds yanked/cut entries.
type Clipboard struct {
	Entries []string      // absolute paths
	FS      fs.FileSystem // source filesystem
	Op      ClipOp
}

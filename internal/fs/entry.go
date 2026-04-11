package fs

import (
	"os"
	"time"
)

// FileEntry represents a single filesystem entry.
type FileEntry struct {
	Name       string
	Size       int64
	Mode       os.FileMode
	ModTime    time.Time
	IsDir      bool
	IsLink     bool
	LinkTarget string
}

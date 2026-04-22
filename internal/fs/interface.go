package fs

import (
	"io"
	"os"
)

// FileSystem abstracts local and remote filesystems.
type FileSystem interface {
	List(dir string) ([]FileEntry, error)
	Stat(path string) (FileEntry, error)
	Mkdir(path string, perm os.FileMode) error
	Remove(path string) error
	RemoveAll(path string) error
	Rename(oldPath, newPath string) error

	// For copy/move with progress
	Open(path string) (io.ReadCloser, error)
	Create(path string, perm os.FileMode) (io.WriteCloser, error)

	// Metadata
	IsLocal() bool
	Root() string
	Home() string
}

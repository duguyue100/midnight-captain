package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// Copy performs an async copy of sources into destDir on dstFS.
// Progress is reported via tea.Cmd messages.
func Copy(id string, sources []string, destDir string, srcFS, dstFS fs.FileSystem) tea.Cmd {
	return func() tea.Msg {
		total, err := calcTotal(sources, srcFS)
		if err != nil {
			return ProgressMsg{OpID: id, Status: StatusFailed, Err: err}
		}

		var done int64

		for _, src := range sources {
			if err := copyEntry(src, destDir, srcFS, dstFS, &done, total, id); err != nil {
				return ProgressMsg{OpID: id, DoneBytes: done, TotalBytes: total, Status: StatusFailed, Err: err}
			}
		}
		return ProgressMsg{OpID: id, DoneBytes: total, TotalBytes: total, Status: StatusDone}
	}
}

func copyEntry(src, destDir string, srcFS, dstFS fs.FileSystem, done *int64, total int64, id string) error {
	entry, err := srcFS.Stat(src)
	if err != nil {
		return err
	}
	destPath := filepath.Join(destDir, entry.Name)

	if entry.IsDir {
		if err := dstFS.Mkdir(destPath, entry.Mode); err != nil && !os.IsExist(err) {
			return err
		}
		children, err := srcFS.List(src)
		if err != nil {
			return err
		}
		for _, child := range children {
			childSrc := filepath.Join(src, child.Name)
			if err := copyEntry(childSrc, destPath, srcFS, dstFS, done, total, id); err != nil {
				return err
			}
		}
		return nil
	}

	// File copy
	r, err := srcFS.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := dstFS.Create(destPath, entry.Mode)
	if err != nil {
		return err
	}

	counter := &countWriter{w: w, done: done}
	_, copyErr := io.Copy(counter, r)

	// Check Close error explicitly — buffered SFTP writes may flush here
	closeErr := w.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func calcTotal(sources []string, srcFS fs.FileSystem) (int64, error) {
	var total int64
	for _, src := range sources {
		n, err := walkSize(src, srcFS)
		if err != nil {
			return 0, fmt.Errorf("stat %s: %w", src, err)
		}
		total += n
	}
	return total, nil
}

func walkSize(path string, srcFS fs.FileSystem) (int64, error) {
	entry, err := srcFS.Stat(path)
	if err != nil {
		return 0, err
	}
	if !entry.IsDir {
		return entry.Size, nil
	}
	children, err := srcFS.List(path)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, child := range children {
		n, err := walkSize(filepath.Join(path, child.Name), srcFS)
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

type countWriter struct {
	w    io.Writer
	done *int64
}

func (c *countWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	*c.done += int64(n)
	return n, err
}

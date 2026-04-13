package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// Copy performs an async copy of sources into destDir on dstFS.
// Progress is reported via tea.Cmd messages.
func Copy(ctx context.Context, id string, sources []string, destDir string, srcFS, dstFS fs.FileSystem) tea.Cmd {
	return func() tea.Msg {
		total, err := calcTotal(sources, srcFS)
		if err != nil {
			return ProgressMsg{OpID: id, Status: StatusFailed, Err: err}
		}

		ch := make(chan tea.Msg, 100)
		var done int64

		go func() {
			defer close(ch)
			for _, src := range sources {
				if ctx.Err() != nil {
					ch <- ProgressMsg{OpID: id, DoneBytes: done, TotalBytes: total, Status: StatusFailed, Err: context.Canceled}
					return
				}
				if err := copyEntry(ctx, src, destDir, srcFS, dstFS, &done, total, id, ch); err != nil {
					ch <- ProgressMsg{OpID: id, DoneBytes: done, TotalBytes: total, Status: StatusFailed, Err: err}
					return
				}
			}
			ch <- ProgressMsg{OpID: id, DoneBytes: total, TotalBytes: total, Status: StatusDone}
		}()

		return ProgressStreamMsg{C: ch}
	}
}

func copyEntry(ctx context.Context, src, destDir string, srcFS, dstFS fs.FileSystem, done *int64, total int64, id string, ch chan tea.Msg) error {
	if ctx.Err() != nil {
		return context.Canceled
	}
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
			if ctx.Err() != nil {
				return context.Canceled
			}
			childSrc := filepath.Join(src, child.Name)
			if err := copyEntry(ctx, childSrc, destPath, srcFS, dstFS, done, total, id, ch); err != nil {
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

	counter := &countWriter{w: w, done: done, total: total, id: id, ch: ch, currentFile: entry.Name, ctx: ctx}
	_, copyErr := io.Copy(counter, r)

	// Check Close error explicitly — buffered SFTP writes may flush here
	closeErr := w.Close()

	// Clean up if cancelled
	if ctx.Err() != nil {
		_ = dstFS.RemoveAll(destPath)
		return context.Canceled
	}

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
	w           io.Writer
	done        *int64
	total       int64
	id          string
	ch          chan tea.Msg
	last        int64
	currentFile string
	ctx         context.Context
}

func (c *countWriter) Write(p []byte) (int, error) {
	if c.ctx.Err() != nil {
		return 0, context.Canceled
	}
	n, err := c.w.Write(p)
	cur := *c.done + int64(n)
	*c.done = cur

	// report every ~512KB to avoid channel flooding
	if cur-c.last > 512*1024 || cur == c.total {
		c.last = cur
		select {
		case c.ch <- ProgressMsg{OpID: c.id, DoneBytes: cur, TotalBytes: c.total, Status: StatusRunning, CurrentFile: c.currentFile}:
		default: // skip if channel full
		}
	}
	return n, err
}

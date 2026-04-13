package ops

import (
	"fmt"
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// Move performs an async move of sources into destDir.
// Same-FS move tries Rename first; falls back to copy+delete.
func Move(id string, sources []string, destDir string, srcFS, dstFS fs.FileSystem) tea.Cmd {
	return func() tea.Msg {
		// Same local FS: try rename (instant) with rollback on partial failure
		if srcFS.IsLocal() && dstFS.IsLocal() {
			var renamed []struct{ src, dst string }
			allOk := true
			for _, src := range sources {
				entry, err := srcFS.Stat(src)
				if err != nil {
					allOk = false
					break
				}
				dest := filepath.Join(destDir, entry.Name)
				if err := srcFS.Rename(src, dest); err != nil {
					allOk = false
					break
				}
				renamed = append(renamed, struct{ src, dst string }{src, dest})
			}
			if allOk {
				return ProgressMsg{OpID: id, Status: StatusDone}
			}
			// Rollback: undo any renames that succeeded
			for _, r := range renamed {
				_ = srcFS.Rename(r.dst, r.src)
			}
			// Fall through to copy+delete path
		}

		// Cross-FS or rename failed: copy then delete
		total, err := calcTotal(sources, srcFS)
		if err != nil {
			return ProgressMsg{OpID: id, Status: StatusFailed, Err: err}
		}

		ch := make(chan tea.Msg, 100)
		var done int64

		go func() {
			defer close(ch)
			for _, src := range sources {
				if err := copyEntry(src, destDir, srcFS, dstFS, &done, total, id, ch); err != nil {
					ch <- ProgressMsg{OpID: id, DoneBytes: done, TotalBytes: total, Status: StatusFailed, Err: err}
					return
				}
				// Delete source after successful copy — surface errors
				if err := srcFS.RemoveAll(src); err != nil {
					ch <- ProgressMsg{OpID: id, DoneBytes: done, TotalBytes: total, Status: StatusFailed,
						Err: fmt.Errorf("move: copied but failed to remove source %s: %w", src, err)}
					return
				}
			}
			ch <- ProgressMsg{OpID: id, DoneBytes: total, TotalBytes: total, Status: StatusDone}
		}()

		return ProgressStreamMsg{C: ch}
	}
}

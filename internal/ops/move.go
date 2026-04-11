package ops

import (
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// Move performs an async move of sources into destDir.
// Same-FS move tries Rename first; falls back to copy+delete.
func Move(id string, sources []string, destDir string, srcFS, dstFS fs.FileSystem) tea.Cmd {
	return func() tea.Msg {
		// Same local FS: try rename (instant)
		if srcFS.IsLocal() && dstFS.IsLocal() {
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
			}
			if allOk {
				return ProgressMsg{OpID: id, Status: StatusDone}
			}
		}

		// Cross-FS or rename failed: copy then delete
		total, err := calcTotal(sources, srcFS)
		if err != nil {
			return ProgressMsg{OpID: id, Status: StatusFailed, Err: err}
		}
		var done int64
		for _, src := range sources {
			if err := copyEntry(src, destDir, srcFS, dstFS, &done, total, id); err != nil {
				return ProgressMsg{OpID: id, DoneBytes: done, TotalBytes: total, Status: StatusFailed, Err: err}
			}
			// Delete source after successful copy
			_ = srcFS.RemoveAll(src)
		}
		return ProgressMsg{OpID: id, DoneBytes: total, TotalBytes: total, Status: StatusDone}
	}
}

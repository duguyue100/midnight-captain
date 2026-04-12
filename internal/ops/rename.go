package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// Rename renames a single entry on the given filesystem.
// newName must be a plain name with no path separators.
func Rename(oldPath, newName string, filesystem fs.FileSystem) tea.Cmd {
	return func() tea.Msg {
		if strings.ContainsAny(newName, "/\\") {
			return ProgressMsg{OpID: "rename", Status: StatusFailed,
				Err: fmt.Errorf("rename: invalid name %q (contains path separator)", newName)}
		}
		newPath := filepath.Join(parentDir(oldPath), newName)
		if err := filesystem.Rename(oldPath, newPath); err != nil {
			return ProgressMsg{OpID: "rename", Status: StatusFailed, Err: err}
		}
		return ProgressMsg{OpID: "rename", Status: StatusDone}
	}
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}

package ops

import (
	"charm.land/bubbletea/v2"
	"context"
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// Delete removes all sources from srcFS (no progress bar — usually fast).
func Delete(ctx context.Context, id string, sources []string, srcFS fs.FileSystem) tea.Cmd {
	return func() tea.Msg {
		for _, src := range sources {
			if ctx.Err() != nil {
				return ProgressMsg{OpID: id, Status: StatusFailed, Err: context.Canceled}
			}
			if err := srcFS.RemoveAll(src); err != nil {
				return ProgressMsg{OpID: id, Status: StatusFailed, Err: err}
			}
		}
		return ProgressMsg{OpID: id, Status: StatusDone}
	}
}

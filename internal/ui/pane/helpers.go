package pane

import (
	"sort"
	"strings"

	"github.com/dgyhome/midnight-captain/internal/fs"
)

func filterEntries(entries []fs.FileEntry, showHidden bool) []fs.FileEntry {
	if showHidden {
		return entries
	}
	out := make([]fs.FileEntry, 0, len(entries))
	for _, e := range entries {
		if e.Name == ".." || !strings.HasPrefix(e.Name, ".") {
			out = append(out, e)
		}
	}
	return out
}

func sortEntries(entries []fs.FileEntry, by SortField, asc bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		// ".." always first
		if a.Name == ".." {
			return true
		}
		if b.Name == ".." {
			return false
		}
		// Dirs before files
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		var less bool
		switch by {
		case SortBySize:
			less = a.Size < b.Size
		case SortByMtime:
			less = a.ModTime.Before(b.ModTime)
		default: // SortByName
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
		if asc {
			return less
		}
		return !less
	})
}

// parentEntry returns a synthetic ".." directory entry.
func parentEntry() fs.FileEntry {
	return fs.FileEntry{Name: "..", IsDir: true}
}

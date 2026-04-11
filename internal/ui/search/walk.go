package search

import (
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/sahilm/fuzzy"
	"io/fs"
)

// allFiles caches the full file walk for the current session.
var allFiles []string

// startWalk walks baseDir in a goroutine and sends ResultsMsg when done.
func startWalk(baseDir string) tea.Cmd {
	return func() tea.Msg {
		var paths []string
		_ = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable dirs
			}
			if d.IsDir() && d.Name() == ".git" {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				rel, _ := filepath.Rel(baseDir, path)
				paths = append(paths, rel)
			}
			return nil
		})
		allFiles = paths
		return ResultsMsg{Done: true}
	}
}

// filter applies fuzzy matching to allFiles using the query.
func filter(query string) []SearchResult {
	if query == "" || len(allFiles) == 0 {
		// Return all (capped)
		limit := len(allFiles)
		if limit > 100 {
			limit = 100
		}
		results := make([]SearchResult, limit)
		for i, f := range allFiles[:limit] {
			results[i] = SearchResult{Path: f, Score: 0}
		}
		return results
	}

	matches := fuzzy.Find(query, allFiles)
	limit := len(matches)
	if limit > 100 {
		limit = 100
	}
	results := make([]SearchResult, limit)
	for i, m := range matches[:limit] {
		results[i] = SearchResult{Path: m.Str, Score: m.Score}
	}
	return results
}

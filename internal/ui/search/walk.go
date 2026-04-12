package search

import (
	"io/fs"
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/sahilm/fuzzy"
)

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
		return ResultsMsg{Done: true, Files: paths}
	}
}

// filter applies fuzzy matching to the given file list using the query.
func filter(query string, files []string) []SearchResult {
	if query == "" || len(files) == 0 {
		// Return all (capped)
		limit := len(files)
		if limit > 100 {
			limit = 100
		}
		results := make([]SearchResult, limit)
		for i, f := range files[:limit] {
			results[i] = SearchResult{Path: f, Score: 0}
		}
		return results
	}

	matches := fuzzy.Find(query, files)
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

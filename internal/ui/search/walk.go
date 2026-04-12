package search

import (
	"io/fs"
	"path/filepath"

	"charm.land/bubbletea/v2"
	"github.com/sahilm/fuzzy"
)

// startWalk walks baseDir in a goroutine and sends ResultsMsg when done.
// The collected file list is returned in ResultsMsg.Files — never written to
// shared state — so there is no data race.
func startWalk(baseDir string) tea.Cmd {
	return func() tea.Msg {
		var paths []string
		var walkErrors int
		_ = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				walkErrors++
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
		return ResultsMsg{Files: paths, Done: true, WalkErrors: walkErrors}
	}
}

// filter applies fuzzy matching to m.allFiles using the query.
func (m *Model) filter(query string) []SearchResult {
	if query == "" || len(m.allFiles) == 0 {
		limit := len(m.allFiles)
		if limit > 100 {
			limit = 100
		}
		results := make([]SearchResult, limit)
		for i, f := range m.allFiles[:limit] {
			results[i] = SearchResult{Path: f, Score: 0}
		}
		return results
	}

	matches := fuzzy.Find(query, m.allFiles)
	limit := len(matches)
	if limit > 100 {
		limit = 100
	}
	results := make([]SearchResult, limit)
	for i, match := range matches[:limit] {
		results[i] = SearchResult{Path: match.Str, Score: match.Score}
	}
	return results
}

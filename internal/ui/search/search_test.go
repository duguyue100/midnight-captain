package search

import (
	"strings"
	"testing"
)

func makeModel(files []string) Model {
	m := Model{allFiles: files}
	return m
}

// --- filter: empty query ---

func TestFilterEmptyQueryReturnsAll(t *testing.T) {
	m := makeModel([]string{"foo.go", "bar.go", "baz.go"})
	got := m.filter("")
	if len(got) != 3 {
		t.Errorf("empty query: want 3, got %d", len(got))
	}
}

func TestFilterEmptyQueryNoFiles(t *testing.T) {
	m := makeModel(nil)
	got := m.filter("")
	if len(got) != 0 {
		t.Errorf("no files: want 0, got %d", len(got))
	}
}

// --- filter: cap at 100 ---

func TestFilterEmptyQueryCapAt100(t *testing.T) {
	files := make([]string, 150)
	for i := range files {
		files[i] = strings.Repeat("a", i+1) + ".go"
	}
	m := makeModel(files)
	got := m.filter("")
	if len(got) != 100 {
		t.Errorf("empty query cap: want 100, got %d", len(got))
	}
}

func TestFilterWithQueryCapAt100(t *testing.T) {
	// 150 files all matching "go"
	files := make([]string, 150)
	for i := range files {
		files[i] = strings.Repeat("g", i+1) + "o.go"
	}
	m := makeModel(files)
	got := m.filter("go")
	if len(got) > 100 {
		t.Errorf("query cap: want ≤100, got %d", len(got))
	}
}

// --- filter: fuzzy match ---

func TestFilterFuzzyMatch(t *testing.T) {
	m := makeModel([]string{"main.go", "readme.md", "go.mod", "config.yaml"})
	got := m.filter("go")
	// "main.go", "go.mod" should appear; "readme.md" maybe not
	found := false
	for _, r := range got {
		if r.Path == "go.mod" || r.Path == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("fuzzy 'go': expected go.mod or main.go in results")
	}
}

func TestFilterNoMatch(t *testing.T) {
	m := makeModel([]string{"alpha.go", "beta.go"})
	got := m.filter("zzzzzzzzzzzzz")
	if len(got) != 0 {
		t.Errorf("no match: want 0, got %d", len(got))
	}
}

// --- filter: score zero for empty query ---

func TestFilterEmptyQueryScoreZero(t *testing.T) {
	m := makeModel([]string{"x.go", "y.go"})
	for _, r := range m.filter("") {
		if r.Score != 0 {
			t.Errorf("empty query: score should be 0, got %d for %q", r.Score, r.Path)
		}
	}
}

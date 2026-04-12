package pane

import (
	"testing"
	"time"

	"github.com/dgyhome/midnight-captain/internal/fs"
)

// --- filterEntries ---

func TestFilterEntriesShowHidden(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: ".hidden"},
		{Name: "visible"},
		{Name: ".."},
	}
	got := filterEntries(entries, true)
	if len(got) != 3 {
		t.Errorf("showHidden=true: want 3 entries, got %d", len(got))
	}
}

func TestFilterEntriesHideHidden(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: ".hidden"},
		{Name: "visible"},
		{Name: ".."}, // ".." always kept
		{Name: ".gitignore"},
	}
	got := filterEntries(entries, false)
	// ".." + "visible" kept; ".hidden" + ".gitignore" dropped
	if len(got) != 2 {
		t.Errorf("showHidden=false: want 2, got %d", len(got))
	}
	for _, e := range got {
		if e.Name != ".." && e.Name[0] == '.' {
			t.Errorf("hidden entry %q leaked through", e.Name)
		}
	}
}

func TestFilterEntriesEmpty(t *testing.T) {
	got := filterEntries(nil, false)
	if len(got) != 0 {
		t.Errorf("want 0, got %d", len(got))
	}
}

// --- sortEntries ---

func TestSortEntriesDotDotFirst(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "aaa", IsDir: false},
		{Name: ".."},
		{Name: "bbb", IsDir: false},
	}
	sortEntries(entries, SortByName, true)
	if entries[0].Name != ".." {
		t.Errorf("expected '..' first, got %q", entries[0].Name)
	}
}

func TestSortEntriesDirsBeforeFiles(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "file.txt", IsDir: false},
		{Name: "dir", IsDir: true},
	}
	sortEntries(entries, SortByName, true)
	if !entries[0].IsDir {
		t.Errorf("expected dir first")
	}
}

func TestSortEntriesByNameAsc(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "beta", IsDir: false},
		{Name: "Alpha", IsDir: false},
		{Name: "gamma", IsDir: false},
	}
	sortEntries(entries, SortByName, true)
	want := []string{"Alpha", "beta", "gamma"}
	for i, w := range want {
		if entries[i].Name != w {
			t.Errorf("[%d] got %q want %q", i, entries[i].Name, w)
		}
	}
}

func TestSortEntriesByNameDesc(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "alpha", IsDir: false},
		{Name: "beta", IsDir: false},
		{Name: "gamma", IsDir: false},
	}
	sortEntries(entries, SortByName, false)
	if entries[0].Name != "gamma" {
		t.Errorf("desc: expected gamma first, got %q", entries[0].Name)
	}
}

func TestSortEntriesBySize(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "big", Size: 1000, IsDir: false},
		{Name: "small", Size: 10, IsDir: false},
		{Name: "mid", Size: 100, IsDir: false},
	}
	sortEntries(entries, SortBySize, true)
	if entries[0].Name != "small" {
		t.Errorf("size asc: expected small first, got %q", entries[0].Name)
	}
}

func TestSortEntriesByMtime(t *testing.T) {
	t0 := time.Unix(1000, 0)
	t1 := time.Unix(2000, 0)
	t2 := time.Unix(3000, 0)
	entries := []fs.FileEntry{
		{Name: "newest", ModTime: t2, IsDir: false},
		{Name: "oldest", ModTime: t0, IsDir: false},
		{Name: "middle", ModTime: t1, IsDir: false},
	}
	sortEntries(entries, SortByMtime, true)
	if entries[0].Name != "oldest" {
		t.Errorf("mtime asc: expected oldest first, got %q", entries[0].Name)
	}
}

// --- visibleHeight ---

func TestVisibleHeightNormal(t *testing.T) {
	m := Model{Height: 20}
	got := m.visibleHeight()
	if got != 16 { // 20 - 4
		t.Errorf("visibleHeight(20) = %d, want 16", got)
	}
}

func TestVisibleHeightClampMin(t *testing.T) {
	m := Model{Height: 2} // 2-4 = -2 → clamped to 1
	got := m.visibleHeight()
	if got < 1 {
		t.Errorf("visibleHeight must be ≥ 1, got %d", got)
	}
}

func TestVisibleHeightZero(t *testing.T) {
	m := Model{Height: 0}
	got := m.visibleHeight()
	if got < 1 {
		t.Errorf("visibleHeight(0) must be ≥ 1, got %d", got)
	}
}

// --- extendVisualSelection ---

func TestExtendVisualSelectionForward(t *testing.T) {
	m := Model{
		VisualAnchor: 1,
		Cursor:       3,
		Selected:     make(map[int]bool),
	}
	m.extendVisualSelection()
	for _, i := range []int{1, 2, 3} {
		if !m.Selected[i] {
			t.Errorf("index %d should be selected", i)
		}
	}
	if m.Selected[0] || m.Selected[4] {
		t.Error("indices outside range should not be selected")
	}
}

func TestExtendVisualSelectionBackward(t *testing.T) {
	m := Model{
		VisualAnchor: 4,
		Cursor:       2,
		Selected:     make(map[int]bool),
	}
	m.extendVisualSelection()
	for _, i := range []int{2, 3, 4} {
		if !m.Selected[i] {
			t.Errorf("index %d should be selected", i)
		}
	}
}

func TestExtendVisualSelectionSingleRow(t *testing.T) {
	m := Model{
		VisualAnchor: 5,
		Cursor:       5,
		Selected:     make(map[int]bool),
	}
	m.extendVisualSelection()
	if !m.Selected[5] {
		t.Error("anchor==cursor: index 5 should be selected")
	}
	if len(m.Selected) != 1 {
		t.Errorf("want 1 selected, got %d", len(m.Selected))
	}
}

func TestExtendVisualSelectionClearsOld(t *testing.T) {
	m := Model{
		VisualAnchor: 0,
		Cursor:       1,
		Selected:     map[int]bool{99: true}, // stale
	}
	m.extendVisualSelection()
	if m.Selected[99] {
		t.Error("stale selection should be cleared")
	}
}

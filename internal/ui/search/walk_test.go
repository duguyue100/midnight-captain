package search

import (
	"os"
	"path/filepath"
	"testing"
)

// startWalk is internal (same package) — call it directly.

func TestWalkExcludesGitDir(t *testing.T) {
	tmp := t.TempDir()
	// Create .git dir with files
	os.MkdirAll(filepath.Join(tmp, ".git", "objects"), 0755)
	os.WriteFile(filepath.Join(tmp, ".git", "HEAD"), []byte("ref: refs/heads/main"), 0644)
	// Create regular files
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module x"), 0644)

	cmd := startWalk(tmp)
	msg := cmd()
	rm, ok := msg.(ResultsMsg)
	if !ok {
		t.Fatalf("expected ResultsMsg, got %T", msg)
	}
	for _, f := range rm.Files {
		if len(f) >= 4 && f[:4] == ".git" {
			t.Errorf("file under .git should be excluded: %q", f)
		}
	}
	// main.go and go.mod should be present
	found := map[string]bool{}
	for _, f := range rm.Files {
		found[f] = true
	}
	for _, want := range []string{"main.go", "go.mod"} {
		if !found[want] {
			t.Errorf("expected %q in results", want)
		}
	}
}

func TestWalkSkipsUnreadableDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can read all dirs")
	}
	tmp := t.TempDir()
	locked := filepath.Join(tmp, "locked")
	os.Mkdir(locked, 0000)
	os.WriteFile(filepath.Join(tmp, "visible.txt"), nil, 0644)

	defer os.Chmod(locked, 0755) // restore for cleanup

	cmd := startWalk(tmp)
	msg := cmd()
	rm := msg.(ResultsMsg)
	if !rm.Done {
		t.Error("ResultsMsg.Done should be true")
	}
	// visible.txt present; no panic from locked dir
	found := false
	for _, f := range rm.Files {
		if f == "visible.txt" {
			found = true
		}
	}
	if !found {
		t.Error("visible.txt should appear despite locked sibling dir")
	}
}

func TestWalkResultsDone(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.txt"), nil, 0644)

	cmd := startWalk(tmp)
	msg := cmd()
	rm, ok := msg.(ResultsMsg)
	if !ok {
		t.Fatalf("expected ResultsMsg, got %T", msg)
	}
	if !rm.Done {
		t.Error("ResultsMsg.Done should be true")
	}
}

func TestWalkRelativePaths(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "sub"), 0755)
	os.WriteFile(filepath.Join(tmp, "sub", "file.go"), nil, 0644)

	cmd := startWalk(tmp)
	msg := cmd()
	rm := msg.(ResultsMsg)

	for _, f := range rm.Files {
		if filepath.IsAbs(f) {
			t.Errorf("path should be relative, got %q", f)
		}
	}
	found := false
	for _, f := range rm.Files {
		if f == filepath.Join("sub", "file.go") {
			found = true
		}
	}
	if !found {
		t.Errorf("sub/file.go not found in %v", rm.Files)
	}
}

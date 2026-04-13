package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appfs "github.com/dgyhome/midnight-captain/internal/fs"
)

func localFS() *appfs.LocalFS { return appfs.NewLocalFS() }

// --- Copy integration ---

func TestCopyFileIntegration(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create source file
	srcFile := filepath.Join(src, "hello.txt")
	os.WriteFile(srcFile, []byte("hello world"), 0644)

	lfs := localFS()
	cmd := Copy(context.Background(), "cp-int", []string{srcFile}, dst, lfs, lfs)
	msg := cmd()
	psm, ok := msg.(ProgressStreamMsg)
	if !ok {
		t.Fatalf("expected ProgressStreamMsg, got %T", msg)
	}

	var pm ProgressMsg
	for m := range psm.C {
		pm = m.(ProgressMsg)
	}

	if pm.Status != StatusDone {
		t.Fatalf("status=%v err=%v", pm.Status, pm.Err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	if err != nil {
		t.Fatalf("dest file not found: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("content mismatch: %q", string(data))
	}
}

func TestCopyDirIntegration(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create source dir with nested file
	subDir := filepath.Join(src, "mydir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644)

	lfs := localFS()
	cmd := Copy(context.Background(), "cp-dir", []string{subDir}, dst, lfs, lfs)
	msg := cmd()
	psm := msg.(ProgressStreamMsg)

	var pm ProgressMsg
	for m := range psm.C {
		pm = m.(ProgressMsg)
	}

	if pm.Status != StatusDone {
		t.Fatalf("copy dir: status=%v err=%v", pm.Status, pm.Err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "mydir", "nested.txt"))
	if err != nil {
		t.Fatalf("nested file not found: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("content=%q want 'nested'", string(data))
	}
}

func TestCopyMissingSourceIntegration(t *testing.T) {
	dst := t.TempDir()
	lfs := localFS()
	cmd := Copy(context.Background(), "cp-miss", []string{"/no/such/file.txt"}, dst, lfs, lfs)
	msg := cmd()
	pm := msg.(ProgressMsg)
	if pm.Status != StatusFailed {
		t.Errorf("missing source: want StatusFailed, got %v", pm.Status)
	}
}

// --- Move integration ---

func TestMoveFileSameFS(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcFile := filepath.Join(src, "tomove.txt")
	os.WriteFile(srcFile, []byte("move me"), 0644)

	lfs := localFS()
	cmd := Move(context.Background(), "mv-int", []string{srcFile}, dst, lfs, lfs)
	msg := cmd()
	pm := msg.(ProgressMsg)
	if pm.Status != StatusDone {
		t.Fatalf("move: status=%v err=%v", pm.Status, pm.Err)
	}

	// Source gone
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("source file should be gone after move")
	}
	// Dest exists
	data, err := os.ReadFile(filepath.Join(dst, "tomove.txt"))
	if err != nil {
		t.Fatalf("dest file not found: %v", err)
	}
	if string(data) != "move me" {
		t.Errorf("content=%q want 'move me'", string(data))
	}
}

// --- Delete integration ---

func TestDeleteFileIntegration(t *testing.T) {
	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "a.txt")
	f2 := filepath.Join(tmp, "b.txt")
	os.WriteFile(f1, nil, 0644)
	os.WriteFile(f2, nil, 0644)

	lfs := localFS()
	cmd := Delete(context.Background(), "del-int", []string{f1, f2}, lfs)
	msg := cmd()
	pm := msg.(ProgressMsg)
	if pm.Status != StatusDone {
		t.Fatalf("delete: status=%v err=%v", pm.Status, pm.Err)
	}
	for _, f := range []string{f1, f2} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("%s should be deleted", f)
		}
	}
}

func TestDeleteDirIntegration(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "rmdir")
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "file.txt"), nil, 0644)

	lfs := localFS()
	cmd := Delete(context.Background(), "del-dir", []string{dir}, lfs)
	msg := cmd()
	pm := msg.(ProgressMsg)
	if pm.Status != StatusDone {
		t.Fatalf("delete dir: status=%v err=%v", pm.Status, pm.Err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("dir should be deleted")
	}
}

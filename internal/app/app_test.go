package app

import (
	"errors"
	"io"
	"os"
	"testing"

	appfs "github.com/dgyhome/midnight-captain/internal/fs"
)

// --- mockFS for app tests ---

type mockFS struct {
	mkdirFn func(string, os.FileMode) error
	statFn  func(string) (appfs.FileEntry, error)
}

func (m *mockFS) Mkdir(path string, perm os.FileMode) error {
	if m.mkdirFn != nil {
		return m.mkdirFn(path, perm)
	}
	return nil
}
func (m *mockFS) Stat(path string) (appfs.FileEntry, error) {
	if m.statFn != nil {
		return m.statFn(path)
	}
	return appfs.FileEntry{}, os.ErrNotExist
}
func (m *mockFS) List(dir string) ([]appfs.FileEntry, error) { return nil, nil }
func (m *mockFS) Remove(path string) error                   { return nil }
func (m *mockFS) RemoveAll(path string) error                { return nil }
func (m *mockFS) Rename(old, new string) error               { return nil }
func (m *mockFS) Open(path string) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}
func (m *mockFS) Create(path string, perm os.FileMode) (io.WriteCloser, error) {
	return &nopWC{}, nil
}
func (m *mockFS) IsLocal() bool { return true }
func (m *mockFS) Root() string  { return "/" }
func (m *mockFS) Home() string  { return "/" }

type nopWC struct{}

func (n *nopWC) Write(p []byte) (int, error) { return len(p), nil }
func (n *nopWC) Close() error                { return nil }

// --- mkdirAllFS ---

func TestMkdirAllFSRoot(t *testing.T) {
	calls := 0
	fsys := &mockFS{
		mkdirFn: func(path string, perm os.FileMode) error {
			calls++
			return nil
		},
	}
	// "/" alone → no components → zero mkdir calls
	err := mkdirAllFS(fsys, "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 0 {
		t.Errorf("root '/': expected 0 mkdir calls, got %d", calls)
	}
}

func TestMkdirAllFSNestedDirs(t *testing.T) {
	var created []string
	fsys := &mockFS{
		mkdirFn: func(path string, perm os.FileMode) error {
			created = append(created, path)
			return nil
		},
	}
	err := mkdirAllFS(fsys, "/a/b/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should create /a, /a/b, /a/b/c in order
	if len(created) != 3 {
		t.Errorf("want 3 mkdir calls, got %d: %v", len(created), created)
	}
	if created[0] != "/a" {
		t.Errorf("first call should be /a, got %q", created[0])
	}
	if created[2] != "/a/b/c" {
		t.Errorf("last call should be /a/b/c, got %q", created[2])
	}
}

func TestMkdirAllFSAlreadyExists(t *testing.T) {
	// os.IsExist(err) path — should skip silently
	fsys := &mockFS{
		mkdirFn: func(path string, perm os.FileMode) error {
			return os.ErrExist
		},
	}
	err := mkdirAllFS(fsys, "/existing/path")
	if err != nil {
		t.Errorf("IsExist errors should be ignored, got: %v", err)
	}
}

func TestMkdirAllFSFileBlocksDir(t *testing.T) {
	// Mkdir fails with a non-IsExist error, and Stat says it's not a dir
	boom := errors.New("mkdir failed")
	fsys := &mockFS{
		mkdirFn: func(path string, perm os.FileMode) error {
			return boom
		},
		statFn: func(path string) (appfs.FileEntry, error) {
			// exists but is a file, not a dir
			return appfs.FileEntry{Name: "oops", IsDir: false}, nil
		},
	}
	err := mkdirAllFS(fsys, "/oops/child")
	if err == nil {
		t.Error("expected error when non-dir blocks mkdir")
	}
}

func TestMkdirAllFSMkdirFailStatFail(t *testing.T) {
	// Mkdir fails, Stat also fails → return original mkdir error
	boom := errors.New("no permission")
	fsys := &mockFS{
		mkdirFn: func(path string, perm os.FileMode) error {
			return boom
		},
		statFn: func(path string) (appfs.FileEntry, error) {
			return appfs.FileEntry{}, errors.New("stat failed too")
		},
	}
	err := mkdirAllFS(fsys, "/a/b")
	if err == nil {
		t.Error("expected error")
	}
}

func TestMkdirAllFSExistingDirPassthrough(t *testing.T) {
	// Mkdir fails with non-IsExist but Stat shows it IS a dir → should continue
	fsys := &mockFS{
		mkdirFn: func(path string, perm os.FileMode) error {
			return errors.New("some other error")
		},
		statFn: func(path string) (appfs.FileEntry, error) {
			return appfs.FileEntry{Name: "dir", IsDir: true}, nil
		},
	}
	err := mkdirAllFS(fsys, "/dir/sub")
	if err != nil {
		t.Errorf("existing dir should not error, got: %v", err)
	}
}

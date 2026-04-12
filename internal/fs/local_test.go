package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func newLocal() *LocalFS { return NewLocalFS() }

// --- List: symlink detection ---

func TestListSymlink(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "real.txt")
	os.WriteFile(target, []byte("content"), 0644)
	link := filepath.Join(tmp, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink not supported:", err)
	}

	fs := newLocal()
	entries, err := fs.List(tmp)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if e.Name == "link.txt" {
			found = true
			if !e.IsLink {
				t.Error("link.txt: IsLink should be true")
			}
			if e.LinkTarget != target {
				t.Errorf("LinkTarget=%q, want %q", e.LinkTarget, target)
			}
		}
	}
	if !found {
		t.Error("link.txt not found in List output")
	}
}

// --- Create: truncates existing file ---

func TestCreateTruncates(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.txt")

	// Write initial content
	os.WriteFile(path, []byte("initial content here"), 0644)

	fs := newLocal()
	w, err := fs.Create(path, 0644)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("new"))
	w.Close()

	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Errorf("Create should truncate: got %q, want 'new'", string(data))
	}
}

// --- Mkdir: idempotent (MkdirAll) ---

func TestMkdirOnExisting(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "existing")
	os.Mkdir(dir, 0755)

	fs := newLocal()
	if err := fs.Mkdir(dir, 0755); err != nil {
		t.Errorf("Mkdir on existing dir should not error: %v", err)
	}
}

func TestMkdirCreatesNested(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "a", "b", "c")

	fs := newLocal()
	if err := fs.Mkdir(nested, 0755); err != nil {
		t.Errorf("Mkdir nested: %v", err)
	}
	info, err := os.Stat(nested)
	if err != nil || !info.IsDir() {
		t.Error("nested dir should exist after Mkdir")
	}
}

// --- Stat: symlink ---

func TestStatSymlink(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "real.txt")
	os.WriteFile(target, []byte("x"), 0644)
	link := filepath.Join(tmp, "sym")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink not supported:", err)
	}

	fs := newLocal()
	entry, err := fs.Stat(link)
	if err != nil {
		t.Fatal(err)
	}
	if !entry.IsLink {
		t.Error("Stat symlink: IsLink should be true")
	}
	if entry.LinkTarget != target {
		t.Errorf("LinkTarget=%q want %q", entry.LinkTarget, target)
	}
}

// --- Remove ---

func TestRemoveFile(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "todelete.txt")
	os.WriteFile(f, nil, 0644)

	fs := newLocal()
	if err := fs.Remove(f); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("file should be gone after Remove")
	}
}

// --- RemoveAll ---

func TestRemoveAllDir(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "rmdir")
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "file.txt"), nil, 0644)

	fs := newLocal()
	if err := fs.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("dir should be gone after RemoveAll")
	}
}

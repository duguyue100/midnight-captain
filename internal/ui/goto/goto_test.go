package goto_

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

// --- expandTilde ---

func TestExpandTildeNoTilde(t *testing.T) {
	got := expandTilde("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("got %q, want /absolute/path", got)
	}
}

func TestExpandTildeTildeAlone(t *testing.T) {
	u, _ := user.Current()
	got := expandTilde("~")
	if got != u.HomeDir {
		t.Errorf("got %q, want %q", got, u.HomeDir)
	}
}

func TestExpandTildeTildeSlash(t *testing.T) {
	u, _ := user.Current()
	got := expandTilde("~/foo/bar")
	want := filepath.Join(u.HomeDir, "foo/bar")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpandTildeNotAtStart(t *testing.T) {
	got := expandTilde("/foo/~/bar")
	if got != "/foo/~/bar" {
		t.Errorf("tilde not at start should not expand, got %q", got)
	}
}

// --- contractTilde ---

func TestContractTildeHomePath(t *testing.T) {
	u, _ := user.Current()
	input := u.HomeDir + "/documents/file.txt"
	got := contractTilde(input)
	if !strings.HasPrefix(got, "~/") {
		t.Errorf("expected ~/... got %q", got)
	}
}

func TestContractTildeExactHome(t *testing.T) {
	u, _ := user.Current()
	got := contractTilde(u.HomeDir)
	if got != "~" {
		t.Errorf("exact home: got %q want '~'", got)
	}
}

func TestContractTildeNonHome(t *testing.T) {
	got := contractTilde("/usr/local/bin")
	if got != "/usr/local/bin" {
		t.Errorf("non-home path should be unchanged, got %q", got)
	}
}

func TestContractExpandRoundtrip(t *testing.T) {
	original := "~/some/nested/path"
	expanded := expandTilde(original)
	contracted := contractTilde(expanded)
	if contracted != original {
		t.Errorf("roundtrip: got %q, want %q", contracted, original)
	}
}

// --- resolveDir ---

func TestResolveDirEmpty(t *testing.T) {
	got := resolveDir("")
	if got != "" {
		t.Errorf("empty input: want '', got %q", got)
	}
}

func TestResolveDirExistingDir(t *testing.T) {
	tmp := t.TempDir()
	got := resolveDir(tmp)
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

func TestResolveDirExistingFile(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "file.txt")
	os.WriteFile(f, []byte("x"), 0644)
	got := resolveDir(f)
	// file → return parent dir
	if got != tmp {
		t.Errorf("file path: got %q, want parent %q", got, tmp)
	}
}

func TestResolveDirNonExistent(t *testing.T) {
	got := resolveDir("/this/does/not/exist/ever")
	if got != "" {
		t.Errorf("nonexistent: want '', got %q", got)
	}
}

func TestResolveDirTilde(t *testing.T) {
	u, _ := user.Current()
	got := resolveDir("~")
	if got != u.HomeDir {
		t.Errorf("~ resolves to home: got %q, want %q", got, u.HomeDir)
	}
}

// --- listEntries ---

func TestListEntriesEmpty(t *testing.T) {
	tmp := t.TempDir()
	got := listEntries(tmp, "")
	if len(got) != 0 {
		t.Errorf("empty dir: want 0, got %d", len(got))
	}
}

func TestListEntriesAllMatch(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "alpha"), 0755)
	os.Mkdir(filepath.Join(tmp, "beta"), 0755)
	os.WriteFile(filepath.Join(tmp, "gamma.txt"), nil, 0644)
	got := listEntries(tmp, "")
	if len(got) != 3 {
		t.Errorf("want 3, got %d", len(got))
	}
}

func TestListEntriesPrefixFilter(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "foo"), 0755)
	os.Mkdir(filepath.Join(tmp, "bar"), 0755)
	os.WriteFile(filepath.Join(tmp, "foofile.txt"), nil, 0644)
	got := listEntries(tmp, "foo")
	// "foo" dir + "foofile.txt"
	if len(got) != 2 {
		t.Errorf("prefix 'foo': want 2, got %d", len(got))
	}
	for _, e := range got {
		if !strings.HasPrefix(strings.ToLower(e.name), "foo") {
			t.Errorf("entry %q doesn't match prefix 'foo'", e.name)
		}
	}
}

func TestListEntriesDirsFirst(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "afile.txt"), nil, 0644)
	os.Mkdir(filepath.Join(tmp, "zdir"), 0755)
	got := listEntries(tmp, "")
	if len(got) < 2 {
		t.Fatal("expected 2 entries")
	}
	if !got[0].isDir {
		t.Errorf("dirs should come first, got %q (isDir=%v)", got[0].name, got[0].isDir)
	}
}

func TestListEntriesNonExistentDir(t *testing.T) {
	got := listEntries("/no/such/dir/exists", "")
	if len(got) != 0 {
		t.Errorf("nonexistent dir: want 0, got %d", len(got))
	}
}

// --- cycleNext / cyclePrev ---

func modelWithEntries(names []string) Model {
	m := New()
	for _, n := range names {
		m.entries = append(m.entries, entry{name: n, isDir: true})
	}
	m.listDir = "/fake"
	return m
}

func TestCycleNextEmpty(t *testing.T) {
	m := New()
	m.cycleNext() // should not panic
	if m.cursor != -1 {
		t.Errorf("empty entries: cursor should stay -1, got %d", m.cursor)
	}
}

func TestCycleNextAdvances(t *testing.T) {
	m := modelWithEntries([]string{"a", "b", "c"})
	m.cycleNext()
	if m.cursor != 0 {
		t.Errorf("first next: cursor=0, got %d", m.cursor)
	}
	m.cycleNext()
	if m.cursor != 1 {
		t.Errorf("second next: cursor=1, got %d", m.cursor)
	}
}

func TestCycleNextWraps(t *testing.T) {
	m := modelWithEntries([]string{"a", "b"})
	m.cursor = 1
	m.cycleNext()
	if m.cursor != 0 {
		t.Errorf("wrap: cursor should be 0, got %d", m.cursor)
	}
}

func TestCyclePrevEmpty(t *testing.T) {
	m := New()
	m.cyclePrev() // should not panic
}

func TestCyclePrevFromStart(t *testing.T) {
	m := modelWithEntries([]string{"a", "b", "c"})
	m.cursor = 0
	m.cyclePrev()
	// cursor <= 0 → wraps to len-1
	if m.cursor != 2 {
		t.Errorf("prev from 0: cursor=2, got %d", m.cursor)
	}
}

func TestCyclePrevDecrement(t *testing.T) {
	m := modelWithEntries([]string{"a", "b", "c"})
	m.cursor = 2
	m.cyclePrev()
	if m.cursor != 1 {
		t.Errorf("prev from 2: cursor=1, got %d", m.cursor)
	}
}

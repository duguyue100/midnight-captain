package pane

import (
	"io"
	"os"
	"testing"

	appfs "github.com/dgyhome/midnight-captain/internal/fs"
)

// --- mockFS for pane tests ---

type mockFS struct {
	dirs map[string][]appfs.FileEntry
	err  map[string]error
}

func (m *mockFS) List(dir string) ([]appfs.FileEntry, error) {
	if m.err != nil {
		if e, ok := m.err[dir]; ok {
			return nil, e
		}
	}
	if m.dirs != nil {
		if entries, ok := m.dirs[dir]; ok {
			return entries, nil
		}
	}
	return nil, nil
}
func (m *mockFS) Stat(path string) (appfs.FileEntry, error) { return appfs.FileEntry{}, nil }
func (m *mockFS) Mkdir(path string, perm os.FileMode) error { return nil }
func (m *mockFS) Remove(path string) error                  { return nil }
func (m *mockFS) RemoveAll(path string) error               { return nil }
func (m *mockFS) Rename(old, new string) error              { return nil }
func (m *mockFS) Open(path string) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}
func (m *mockFS) Create(path string, perm os.FileMode) (io.WriteCloser, error) {
	return &nopWC{}, nil
}
func (m *mockFS) IsLocal() bool { return true }
func (m *mockFS) Root() string  { return "/" }

type nopWC struct{}

func (n *nopWC) Write(p []byte) (int, error) { return len(p), nil }
func (n *nopWC) Close() error                { return nil }

// --- buildNodes / Reload ---

func TestBuildNodesNoDotDotAtRoot(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/": {{Name: "foo", IsDir: false}},
	}}
	m := New(fsys, "/")
	for _, n := range m.Nodes {
		if n.Entry.Name == ".." {
			t.Error("should have no '..' at root '/'")
		}
	}
}

func TestBuildNodesDotDotNotAtRoot(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/a": {{Name: "file.txt", IsDir: false}},
	}}
	m := New(fsys, "/a")
	if len(m.Nodes) == 0 || m.Nodes[0].Entry.Name != ".." {
		t.Error("non-root: first node should be '..'")
	}
}

func TestBuildNodesEmptyDir(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/empty": {},
	}}
	m := New(fsys, "/empty")
	// only ".." node
	if len(m.Nodes) != 1 || m.Nodes[0].Entry.Name != ".." {
		t.Errorf("empty dir: want only '..', got %d nodes", len(m.Nodes))
	}
}

func TestBuildNodesUnreadableDir(t *testing.T) {
	fsys := &mockFS{
		dirs: map[string][]appfs.FileEntry{},
		err:  map[string]error{"/locked": os.ErrPermission},
	}
	m := New(fsys, "/locked")
	if m.Err == nil {
		t.Error("unreadable dir: m.Err should be set")
	}
}

// --- Reload cursor clamp ---

func TestReloadCursorClamp(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/d": {
			{Name: "a.txt", IsDir: false},
			{Name: "b.txt", IsDir: false},
		},
	}}
	m := New(fsys, "/d")
	// cursor beyond end
	m.Cursor = 999
	m.Reload()
	if m.Cursor >= len(m.Nodes) {
		t.Errorf("cursor not clamped: cursor=%d, nodes=%d", m.Cursor, len(m.Nodes))
	}
}

// --- toggleExpand ---

func TestToggleExpandExpandsDir(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/root":        {{Name: "subdir", IsDir: true}},
		"/root/subdir": {{Name: "child.txt", IsDir: false}},
	}}
	m := New(fsys, "/root")
	// find subdir node (after "..")
	idx := -1
	for i, n := range m.Nodes {
		if n.Entry.Name == "subdir" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("subdir not found in nodes")
	}
	m.toggleExpand(idx)
	// After expand, child.txt should appear
	found := false
	for _, n := range m.Nodes {
		if n.Entry.Name == "child.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("after toggleExpand, child.txt should be visible")
	}
}

func TestToggleExpandCollapses(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/root":        {{Name: "subdir", IsDir: true}},
		"/root/subdir": {{Name: "child.txt", IsDir: false}},
	}}
	m := New(fsys, "/root")
	// Expand first
	idx := -1
	for i, n := range m.Nodes {
		if n.Entry.Name == "subdir" {
			idx = i
			break
		}
	}
	m.toggleExpand(idx)
	// Now collapse — find subdir again after rebuild
	idx = -1
	for i, n := range m.Nodes {
		if n.Entry.Name == "subdir" {
			idx = i
			break
		}
	}
	m.toggleExpand(idx)
	for _, n := range m.Nodes {
		if n.Entry.Name == "child.txt" {
			t.Error("after collapse, child.txt should not be visible")
		}
	}
}

func TestToggleExpandNoop_OnFile(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/root": {{Name: "file.txt", IsDir: false}},
	}}
	m := New(fsys, "/root")
	before := len(m.Nodes)
	for i, n := range m.Nodes {
		if n.Entry.Name == "file.txt" {
			m.toggleExpand(i)
			break
		}
	}
	if len(m.Nodes) != before {
		t.Error("toggleExpand on file should be noop")
	}
}

// --- goParent ---

func TestGoParentFromNonRoot(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/":      {{Name: "child", IsDir: true}},
		"/child": {},
	}}
	m := New(fsys, "/child")
	m2, _ := m.goParent()
	if m2.Cwd != "/" {
		t.Errorf("goParent from /child: want Cwd='/', got %q", m2.Cwd)
	}
}

func TestGoParentFromRootNoOp(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/": {{Name: "a", IsDir: false}},
	}}
	m := New(fsys, "/")
	m2, _ := m.goParent()
	if m2.Cwd != "/" {
		t.Errorf("goParent from '/': Cwd should stay '/', got %q", m2.Cwd)
	}
}

func TestGoParentPositionsCursorOnFormerChild(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/": {
			{Name: "alpha", IsDir: true},
			{Name: "beta", IsDir: true},
		},
		"/beta": {},
	}}
	m := New(fsys, "/beta")
	m2, _ := m.goParent()
	// cursor should land on "beta" in parent
	if m2.Cwd != "/" {
		t.Fatalf("want Cwd='/', got %q", m2.Cwd)
	}
	node, ok := m2.CurrentNode()
	if !ok {
		t.Fatal("no current node after goParent")
	}
	if node.Entry.Name != "beta" {
		t.Errorf("cursor should be on 'beta', got %q", node.Entry.Name)
	}
}

// --- NavigateTo ---

func TestNavigateTo(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/target": {
			{Name: "first.txt", IsDir: false},
			{Name: "second.txt", IsDir: false},
		},
	}}
	m := New(fsys, "/")
	m.NavigateTo("/target", "second.txt")
	if m.Cwd != "/target" {
		t.Errorf("Cwd=%q, want /target", m.Cwd)
	}
	node, ok := m.CurrentNode()
	if !ok {
		t.Fatal("no current node")
	}
	if node.Entry.Name != "second.txt" {
		t.Errorf("cursor on %q, want second.txt", node.Entry.Name)
	}
}

func TestNavigateToMissingName(t *testing.T) {
	fsys := &mockFS{dirs: map[string][]appfs.FileEntry{
		"/target": {{Name: "only.txt", IsDir: false}},
	}}
	m := New(fsys, "/")
	m.NavigateTo("/target", "nonexistent.txt")
	if m.Cwd != "/target" {
		t.Errorf("Cwd=%q, want /target", m.Cwd)
	}
	// cursor defaults to 0 (or "..")
	if m.Cursor < 0 || m.Cursor >= len(m.Nodes) {
		t.Error("cursor out of range after NavigateTo with missing name")
	}
}

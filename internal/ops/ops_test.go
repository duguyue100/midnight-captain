package ops

import (
	"context"
	"bytes"
	"io"
	"os"
	"testing"

	appfs "github.com/dgyhome/midnight-captain/internal/fs"
)

// --- parentDir ---

func TestParentDir(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", "."},
		{".", "."},
		{"file", "."},
		{"/file", "/"},
		{"/foo/bar", "/foo"},
		{"/foo/bar/baz", "/foo/bar"},
		{"/", "/"},
	}
	for _, tc := range cases {
		got := parentDir(tc.input)
		if got != tc.want {
			t.Errorf("parentDir(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- countWriter ---

func TestCountWriter(t *testing.T) {
	var buf bytes.Buffer
	var done int64
	cw := &countWriter{w: &buf, done: &done, ctx: context.Background()}

	data := []byte("hello world")
	n, err := cw.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Errorf("Write returned n=%d, want %d", n, len(data))
	}
	if done != int64(len(data)) {
		t.Errorf("done=%d, want %d", done, len(data))
	}
	if buf.String() != "hello world" {
		t.Errorf("buf=%q", buf.String())
	}
}

func TestCountWriterAccumulates(t *testing.T) {
	var buf bytes.Buffer
	var done int64
	cw := &countWriter{w: &buf, done: &done, ctx: context.Background()}

	cw.Write([]byte("abc"))
	cw.Write([]byte("de"))
	cw.Write([]byte("f"))

	if done != 6 {
		t.Errorf("done=%d, want 6", done)
	}
}

func TestCountWriterPartialWrite(t *testing.T) {
	// limitWriter only accepts at most N bytes per Write.
	var done int64
	lw := &limitWriter{max: 3}
	cw := &countWriter{w: lw, done: &done, ctx: context.Background()}

	// Write 5 bytes, but underlying writer only accepts 3.
	n, err := cw.Write([]byte("hello"))
	// countWriter counts only actually written bytes.
	if int64(n) != done {
		t.Errorf("done=%d but Write returned n=%d; should match", done, n)
	}
	_ = err
}

// limitWriter accepts at most max bytes per call.
type limitWriter struct {
	max     int
	written []byte
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if len(p) > l.max {
		p = p[:l.max]
	}
	l.written = append(l.written, p...)
	return len(p), nil
}

// --- walkSize ---

func TestWalkSizeNonExistent(t *testing.T) {
	fsys := &mockFS{statErr: io.ErrUnexpectedEOF}
	_, err := walkSize("/nonexistent", fsys)
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestWalkSizeFile(t *testing.T) {
	fsys := &mockFS{
		statResult: fsEntry{name: "file.txt", size: 42, isDir: false},
	}
	size, err := walkSize("/file.txt", fsys)
	if err != nil {
		t.Fatal(err)
	}
	if size != 42 {
		t.Errorf("got size %d, want 42", size)
	}
}

func TestWalkSizeEmptyDir(t *testing.T) {
	fsys := &mockFS{
		statResult: fsEntry{name: "dir", size: 0, isDir: true},
		listResult: nil,
	}
	size, err := walkSize("/dir", fsys)
	if err != nil {
		t.Fatal(err)
	}
	if size != 0 {
		t.Errorf("got size %d, want 0", size)
	}
}

// --- delete ---

func TestDeleteSuccess(t *testing.T) {
	var removed []string
	fsys := &mockFS{
		removeAllFn: func(path string) error {
			removed = append(removed, path)
			return nil
		},
	}
	cmd := Delete(context.Background(), "del1", []string{"/a", "/b"}, fsys)
	msg := cmd()
	pm, ok := msg.(ProgressMsg)
	if !ok {
		t.Fatalf("expected ProgressMsg, got %T", msg)
	}
	if pm.Status != StatusDone {
		t.Errorf("status=%v want Done", pm.Status)
	}
	if len(removed) != 2 {
		t.Errorf("removed %d items, want 2", len(removed))
	}
}

func TestDeleteStopsOnFirstError(t *testing.T) {
	calls := 0
	fsys := &mockFS{
		removeAllFn: func(path string) error {
			calls++
			return io.ErrUnexpectedEOF // always fail
		},
	}
	cmd := Delete(context.Background(), "del2", []string{"/a", "/b", "/c"}, fsys)
	msg := cmd()
	pm := msg.(ProgressMsg)
	if pm.Status != StatusFailed {
		t.Errorf("status=%v want Failed", pm.Status)
	}
	if calls != 1 {
		t.Errorf("removeAll called %d times, want 1 (stops on first error)", calls)
	}
}

// --- Copy ---

func TestCopyFile(t *testing.T) {
	content := []byte("file content")
	src := &mockFS{
		statResult: fsEntry{name: "file.txt", size: int64(len(content)), isDir: false},
		openData:   content,
	}
	var written []byte
	dst := &mockFS{
		createFn: func(path string, perm os.FileMode) (io.WriteCloser, error) {
			return &captureWriter{buf: &written}, nil
		},
	}
	cmd := Copy(context.Background(), "cp1", []string{"/file.txt"}, "/dst", src, dst)
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
		t.Errorf("status=%v want Done", pm.Status)
	}
	if string(written) != string(content) {
		t.Errorf("written=%q want %q", written, content)
	}
}

func TestCopyStatError(t *testing.T) {
	src := &mockFS{statErr: io.ErrUnexpectedEOF}
	dst := &mockFS{}
	cmd := Copy(context.Background(), "cp2", []string{"/missing"}, "/dst", src, dst)
	msg := cmd()

	// This returns the stream immediately
	psm, ok := msg.(ProgressStreamMsg)
	if !ok {
		// Or it returns ProgressMsg immediately if total calculation fails early
		if pm, ok := msg.(ProgressMsg); ok {
			if pm.Status != StatusFailed {
				t.Errorf("status=%v want Failed", pm.Status)
			}
			return
		}
		t.Fatalf("expected ProgressStreamMsg or ProgressMsg, got %T", msg)
	}

	var pm ProgressMsg
	for m := range psm.C {
		pm = m.(ProgressMsg)
	}

	if pm.Status != StatusFailed {
		t.Errorf("status=%v want Failed", pm.Status)
	}
}

// --- mockFS ---

type fsEntry struct {
	name  string
	size  int64
	isDir bool
}

type mockFS struct {
	statResult  fsEntry
	statErr     error
	listResult  []appfs.FileEntry
	listErr     error
	removeAllFn func(string) error
	mkdirFn     func(string, os.FileMode) error
	renameFn    func(string, string) error
	createFn    func(string, os.FileMode) (io.WriteCloser, error)
	openData    []byte
}

func (m *mockFS) List(dir string) ([]appfs.FileEntry, error) {
	return m.listResult, m.listErr
}
func (m *mockFS) Stat(path string) (appfs.FileEntry, error) {
	if m.statErr != nil {
		return appfs.FileEntry{}, m.statErr
	}
	return appfs.FileEntry{
		Name:  m.statResult.name,
		Size:  m.statResult.size,
		IsDir: m.statResult.isDir,
	}, nil
}
func (m *mockFS) Mkdir(path string, perm os.FileMode) error {
	if m.mkdirFn != nil {
		return m.mkdirFn(path, perm)
	}
	return nil
}
func (m *mockFS) Remove(path string) error { return nil }
func (m *mockFS) RemoveAll(path string) error {
	if m.removeAllFn != nil {
		return m.removeAllFn(path)
	}
	return nil
}
func (m *mockFS) Rename(old, new string) error {
	if m.renameFn != nil {
		return m.renameFn(old, new)
	}
	return nil
}
func (m *mockFS) Open(path string) (io.ReadCloser, error) {
	data := m.openData
	if data == nil {
		data = []byte{}
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}
func (m *mockFS) Create(path string, perm os.FileMode) (io.WriteCloser, error) {
	if m.createFn != nil {
		return m.createFn(path, perm)
	}
	return &nopWriteCloser{}, nil
}
func (m *mockFS) IsLocal() bool { return true }
func (m *mockFS) Root() string  { return "/" }

type nopWriteCloser struct{}

func (n *nopWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (n *nopWriteCloser) Close() error                { return nil }

type captureWriter struct {
	buf *[]byte
}

func (c *captureWriter) Write(p []byte) (int, error) {
	*c.buf = append(*c.buf, p...)
	return len(p), nil
}
func (c *captureWriter) Close() error { return nil }

package fs

import (
	"io"
	"os"
)

// LocalFS implements FileSystem for the local OS.
type LocalFS struct{}

// NewLocalFS returns a new LocalFS.
func NewLocalFS() *LocalFS { return &LocalFS{} }

func (l *LocalFS) List(dir string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		fe := FileEntry{
			Name:    e.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   e.IsDir(),
			IsLink:  e.Type()&os.ModeSymlink != 0,
		}
		if fe.IsLink {
			if target, err := os.Readlink(dir + "/" + e.Name()); err == nil {
				fe.LinkTarget = target
			}
		}
		result = append(result, fe)
	}
	return result, nil
}

func (l *LocalFS) Stat(path string) (FileEntry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return FileEntry{}, err
	}
	fe := FileEntry{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		IsLink:  info.Mode()&os.ModeSymlink != 0,
	}
	if fe.IsLink {
		if target, err := os.Readlink(path); err == nil {
			fe.LinkTarget = target
		}
	}
	return fe, nil
}

func (l *LocalFS) Mkdir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (l *LocalFS) Remove(path string) error {
	return os.Remove(path)
}

func (l *LocalFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (l *LocalFS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (l *LocalFS) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (l *LocalFS) Create(path string, perm os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
}

func (l *LocalFS) IsLocal() bool { return true }
func (l *LocalFS) Root() string  { return "/" }

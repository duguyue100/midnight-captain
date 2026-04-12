package fs

import (
	"io"
	"os"

	"github.com/pkg/sftp"
)

// RemoteFS implements FileSystem over SFTP.
type RemoteFS struct {
	client *sftp.Client
	root   string
	label  string // "user@host"
}

// NewRemoteFS wraps an sftp.Client as a FileSystem.
func NewRemoteFS(client *sftp.Client, label string) *RemoteFS {
	return &RemoteFS{client: client, root: "/", label: label}
}

func (r *RemoteFS) List(dir string) ([]FileEntry, error) {
	infos, err := r.client.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]FileEntry, 0, len(infos))
	for _, info := range infos {
		fe := infoToEntry(info)
		if fe.IsLink {
			if target, err := r.client.ReadLink(dir + "/" + info.Name()); err == nil {
				fe.LinkTarget = target
			}
		}
		entries = append(entries, fe)
	}
	return entries, nil
}

func (r *RemoteFS) Stat(path string) (FileEntry, error) {
	info, err := r.client.Lstat(path)
	if err != nil {
		return FileEntry{}, err
	}
	fe := infoToEntry(info)
	if fe.IsLink {
		if target, err := r.client.ReadLink(path); err == nil {
			fe.LinkTarget = target
		}
	}
	return fe, nil
}

func (r *RemoteFS) Mkdir(path string, perm os.FileMode) error {
	if err := r.client.MkdirAll(path); err != nil {
		return err
	}
	// Best-effort chmod to honor requested permissions
	_ = r.client.Chmod(path, perm)
	return nil
}

func (r *RemoteFS) Remove(path string) error {
	return r.client.Remove(path)
}

func (r *RemoteFS) RemoveAll(path string) error {
	return r.client.RemoveAll(path)
}

func (r *RemoteFS) Rename(oldPath, newPath string) error {
	return r.client.Rename(oldPath, newPath)
}

func (r *RemoteFS) Open(path string) (io.ReadCloser, error) {
	return r.client.Open(path)
}

func (r *RemoteFS) Create(path string, perm os.FileMode) (io.WriteCloser, error) {
	f, err := r.client.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return nil, err
	}
	// Best-effort: remote umask may restrict, so ignore chmod errors.
	_ = r.client.Chmod(path, perm)
	return f, nil
}

func (r *RemoteFS) IsLocal() bool { return false }
func (r *RemoteFS) Root() string  { return r.label }

func infoToEntry(info os.FileInfo) FileEntry {
	return FileEntry{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		IsLink:  info.Mode()&os.ModeSymlink != 0,
	}
}

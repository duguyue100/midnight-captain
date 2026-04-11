package app

import (
	"charm.land/bubbletea/v2"
	appfs "github.com/dgyhome/midnight-captain/internal/fs"
	appssh "github.com/dgyhome/midnight-captain/internal/ssh"
)

// SSHConnectedMsg is sent when SSH connection succeeds.
type SSHConnectedMsg struct {
	Label  string
	Remote *appfs.RemoteFS
}

// SSHErrorMsg is sent when SSH connection fails.
type SSHErrorMsg struct {
	Err error
}

// sshConnect starts an async SSH connection and switches the active pane's FS.
func sshConnect(m *Model, target string) tea.Cmd {
	return func() tea.Msg {
		sess, err := appssh.Connect(target)
		if err != nil {
			return SSHErrorMsg{Err: err}
		}
		key := sess.User + "@" + sess.Host
		appssh.Global.Set(key, sess)
		remote := appfs.NewRemoteFS(sess.SFTP, key)
		return SSHConnectedMsg{Label: key, Remote: remote}
	}
}

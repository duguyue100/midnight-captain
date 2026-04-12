package appssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Session holds an active SSH+SFTP connection.
type Session struct {
	client    *ssh.Client
	SFTP      *sftp.Client
	Host      string
	User      string
	Connected bool
}

// Connect establishes an SSH connection to user@host[:port].
// Auth order: SSH agent → default key files → password (not implemented here).
func Connect(target string) (*Session, error) {
	user, host, port := parseTarget(target)

	authMethods, cleanup := collectAuthMethods()
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth methods available")
	}
	defer cleanup()

	hostKeyCallback, _ := buildHostKeyCallback()

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("sftp init: %w", err)
	}

	return &Session{
		client:    client,
		SFTP:      sftpClient,
		Host:      host,
		User:      user,
		Connected: true,
	}, nil
}

// Close terminates the SSH session.
func (s *Session) Close() {
	if s.SFTP != nil {
		s.SFTP.Close()
	}
	if s.client != nil {
		s.client.Close()
	}
	s.Connected = false
}

// parseTarget splits "user@host:port" into components.
func parseTarget(target string) (user, host, port string) {
	port = "22"
	if idx := strings.Index(target, "@"); idx >= 0 {
		user = target[:idx]
		target = target[idx+1:]
	} else {
		user = os.Getenv("USER")
	}
	if h, p, err := net.SplitHostPort(target); err == nil {
		host = h
		port = p
	} else {
		host = target
	}
	return
}

func collectAuthMethods() ([]ssh.AuthMethod, func()) {
	var methods []ssh.AuthMethod
	var agentConn net.Conn // track for cleanup

	// SSH agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
			agentConn = conn
		}
	}

	// Default key files
	home, _ := os.UserHomeDir()
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		keyPath := filepath.Join(home, ".ssh", name)
		if key := tryLoadKey(keyPath); key != nil {
			methods = append(methods, key)
		}
	}

	cleanup := func() {
		if agentConn != nil {
			agentConn.Close()
		}
	}
	return methods, cleanup
}

func tryLoadKey(path string) ssh.AuthMethod {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(signer)
}

func buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return ssh.InsecureIgnoreHostKey(), nil //nolint
	}
	khPath := filepath.Join(home, ".ssh", "known_hosts")
	cb, err := knownhosts.New(khPath)
	if err != nil {
		// No known_hosts — fall back to insecure (dev convenience)
		return ssh.InsecureIgnoreHostKey(), nil //nolint
	}
	return cb, nil
}

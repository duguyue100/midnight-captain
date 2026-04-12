package appssh

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
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
	if user == "" {
		return nil, fmt.Errorf("ssh: cannot determine username (no user@ prefix and os user lookup failed)")
	}

	authMethods, cleanup := collectAuthMethods()
	defer cleanup()

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth methods available")
	}

	hostKeyCallback, err := buildHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("ssh host key verification: %w", err)
	}

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
func parseTarget(target string) (userName, host, port string) {
	port = "22"
	if idx := strings.Index(target, "@"); idx >= 0 {
		userName = target[:idx]
		target = target[idx+1:]
	} else {
		// Fallback: try os/user.Current(), then $USER env
		if u, err := user.Current(); err == nil {
			userName = u.Username
		} else {
			userName = os.Getenv("USER")
		}
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
	var closers []func()

	// SSH agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			closers = append(closers, func() { conn.Close() })
			methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}
	}

	// Default key files
	home, _ := os.UserHomeDir()
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		keyPath := filepath.Join(home, ".ssh", name)
		if key, err := tryLoadKey(keyPath); err != nil {
			// Log non-trivial errors (file exists but can't parse — e.g. passphrase-protected)
			if !os.IsNotExist(err) {
				log.Printf("ssh: skipping key %s: %v", keyPath, err)
			}
		} else if key != nil {
			methods = append(methods, key)
		}
	}

	cleanup := func() {
		for _, c := range closers {
			c()
		}
	}
	return methods, cleanup
}

// tryLoadKey loads an SSH private key from disk.
// Returns (nil, nil) if file doesn't exist.
// Returns (nil, err) if file exists but can't be parsed (e.g. passphrase-protected).
func tryLoadKey(path string) (ssh.AuthMethod, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}
	return ssh.PublicKeys(signer), nil
}

// buildHostKeyCallback returns a host-key callback using known_hosts.
// Returns an error instead of silently falling back to insecure when known_hosts
// is missing or unreadable — callers must handle this explicitly.
func buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	khPath := filepath.Join(home, ".ssh", "known_hosts")
	cb, err := knownhosts.New(khPath)
	if err != nil {
		return nil, fmt.Errorf("known_hosts (%s): %w — create the file or verify permissions", khPath, err)
	}
	return cb, nil
}

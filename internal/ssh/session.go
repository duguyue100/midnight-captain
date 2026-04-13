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

// Connect establishes an SSH connection to user@host[:port] or an SSH config
// alias.  Auth order: config IdentityFile → SSH agent → default key files.
func Connect(target string) (*Session, error) {
	user, host, port, explicitUser := parseTarget(target)

	// Resolve ~/.ssh/config for this host alias *before* overriding with
	// explicit CLI values.
	hostCfg := resolveSSHConfig(host)
	if hostCfg != nil {
		if hostCfg.HostName != "" {
			host = hostCfg.HostName
		}
		if hostCfg.Port != "" && port == "22" {
			port = hostCfg.Port
		}
		if hostCfg.User != "" && !explicitUser {
			user = hostCfg.User
		}
	}

	if user == "" {
		return nil, fmt.Errorf("ssh: cannot determine username (no user@ prefix and os user lookup failed)")
	}

	authMethods, cleanup := collectAuthMethods(hostCfg)
	defer cleanup()

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth methods available (check SSH_AUTH_SOCK and key files)")
	}

	hostKeyCallback, err := buildHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("ssh host key verification: %w", err)
	}

	clientCfg := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, port)
	client, err := ssh.Dial("tcp", addr, clientCfg)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s (user=%s): %w", addr, user, err)
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
// Returns explicitUser=true when a user@ prefix was provided.
func parseTarget(target string) (userName, host, port string, explicitUser bool) {
	port = "22"
	if idx := strings.Index(target, "@"); idx >= 0 {
		userName = target[:idx]
		target = target[idx+1:]
		explicitUser = true
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

func collectAuthMethods(hostCfg *sshHostConfig) ([]ssh.AuthMethod, func()) {
	var methods []ssh.AuthMethod
	var closers []func()

	// Put specific IdentityFiles FIRST (before agent) so they get priority!
	var agentClient agent.Agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			closers = append(closers, func() { conn.Close() })
			agentClient = agent.NewClient(conn)
		}
	}

	// IdentityFiles from ~/.ssh/config
	if hostCfg != nil {
		for _, keyPath := range hostCfg.IdentityFiles {
			if m := loadKeyOrAgent(keyPath, agentClient); m != nil {
				methods = append(methods, m)
			}
		}
	}

	// SSH agent (try agent keys AFTER explicit keys)
	if agentClient != nil {
		methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
	}

	home, _ := os.UserHomeDir()
	// Default key files
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		keyPath := filepath.Join(home, ".ssh", name)
		if key, err := tryLoadKey(keyPath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("ssh: skipping key %s: %v", keyPath, err)
			}
		} else if key != nil {
			methods = append(methods, key)
		}
	}

	// keyboard-interactive
	methods = append(methods, ssh.KeyboardInteractive(
		func(name, instruction string, questions []string, echos []bool) ([]string, error) {
			return make([]string, len(questions)), nil
		},
	))

	cleanup := func() {
		for _, c := range closers {
			c()
		}
	}
	return methods, cleanup
}

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

func loadKeyOrAgent(keyPath string, ag agent.Agent) ssh.AuthMethod {
	// Try direct load first (unencrypted key)
	if m, err := tryLoadKey(keyPath); err == nil && m != nil {
		return m
	}

	// Fallback: match via agent using the .pub sidecar file
	if ag == nil {
		return nil
	}
	pubPath := keyPath + ".pub"
	pubData, err := os.ReadFile(pubPath)
	if err != nil {
		return nil
	}
	wantPub, _, _, _, err := ssh.ParseAuthorizedKey(pubData)
	if err != nil {
		return nil
	}
	wantFP := ssh.FingerprintSHA256(wantPub)

	agentKeys, err := ag.List()
	if err != nil {
		return nil
	}
	for _, ak := range agentKeys {
		if ssh.FingerprintSHA256(ak) == wantFP {
			return ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
				signers, err := ag.Signers()
				if err != nil {
					return nil, err
				}
				for _, s := range signers {
					if ssh.FingerprintSHA256(s.PublicKey()) == wantFP {
						return []ssh.Signer{s}, nil
					}
				}
				return nil, fmt.Errorf("agent key vanished")
			})
		}
	}
	return nil
}

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

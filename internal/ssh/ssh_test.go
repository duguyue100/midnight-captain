package appssh

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestParseTargetUserAtHost(t *testing.T) {
	user, host, port, explicit := parseTarget("alice@example.com")
	if user != "alice" {
		t.Errorf("user=%q want 'alice'", user)
	}
	if host != "example.com" {
		t.Errorf("host=%q want 'example.com'", host)
	}
	if port != "22" {
		t.Errorf("port=%q want '22'", port)
	}
	if !explicit {
		t.Error("explicitUser should be true")
	}
}

func TestParseTargetUserAtHostPort(t *testing.T) {
	user, host, port, explicit := parseTarget("bob@myserver.io:2222")
	if user != "bob" {
		t.Errorf("user=%q want 'bob'", user)
	}
	if host != "myserver.io" {
		t.Errorf("host=%q want 'myserver.io'", host)
	}
	if port != "2222" {
		t.Errorf("port=%q want '2222'", port)
	}
	if !explicit {
		t.Error("explicitUser should be true")
	}
}

func TestParseTargetHostOnly(t *testing.T) {
	// No "@" — falls back to user.Current()
	u, err := user.Current()
	if err != nil {
		t.Skip("cannot determine current user")
	}
	userName, host, port, explicit := parseTarget("myhost.local")
	if userName != u.Username {
		t.Errorf("user=%q want %q (from user.Current)", userName, u.Username)
	}
	if host != "myhost.local" {
		t.Errorf("host=%q want 'myhost.local'", host)
	}
	if port != "22" {
		t.Errorf("port=%q want '22'", port)
	}
	if explicit {
		t.Error("explicitUser should be false")
	}
}

func TestParseTargetHostPortOnly(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Skip("cannot determine current user")
	}
	userName, host, port, explicit := parseTarget("remotehost:9922")
	if userName != u.Username {
		t.Errorf("user=%q want %q (from user.Current)", userName, u.Username)
	}
	if host != "remotehost" {
		t.Errorf("host=%q want 'remotehost'", host)
	}
	if port != "9922" {
		t.Errorf("port=%q want '9922'", port)
	}
	if explicit {
		t.Error("explicitUser should be false")
	}
}

func TestParseTargetDefaultPort(t *testing.T) {
	_, _, port, _ := parseTarget("user@host")
	if port != "22" {
		t.Errorf("default port: got %q want '22'", port)
	}
}

// ---------------------------------------------------------------------------
// SSH config parser tests
// ---------------------------------------------------------------------------

// helper: write a fake ~/.ssh/config and point resolveSSHConfig at it via
// overriding HOME so os.UserHomeDir returns our temp dir.
func withFakeSSHConfig(t *testing.T, content string, fn func()) {
	t.Helper()
	tmp := t.TempDir()
	sshDir := filepath.Join(tmp, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	old := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", old)
	fn()
}

func TestResolveSSHConfigBasic(t *testing.T) {
	cfg := `
Host lfwork
  HostName 10.0.0.42
  Port 2222
  User deploy
  IdentityFile ~/.ssh/id_work
`
	withFakeSSHConfig(t, cfg, func() {
		got := resolveSSHConfig("lfwork")
		if got == nil {
			t.Fatal("expected config, got nil")
		}
		if got.HostName != "10.0.0.42" {
			t.Errorf("HostName=%q want '10.0.0.42'", got.HostName)
		}
		if got.Port != "2222" {
			t.Errorf("Port=%q want '2222'", got.Port)
		}
		if got.User != "deploy" {
			t.Errorf("User=%q want 'deploy'", got.User)
		}
		if len(got.IdentityFiles) != 1 {
			t.Fatalf("IdentityFiles len=%d want 1", len(got.IdentityFiles))
		}
		// ~ expanded to fake HOME
		home := os.Getenv("HOME")
		want := filepath.Join(home, ".ssh", "id_work")
		if got.IdentityFiles[0] != want {
			t.Errorf("IdentityFiles[0]=%q want %q", got.IdentityFiles[0], want)
		}
	})
}

func TestResolveSSHConfigNoMatch(t *testing.T) {
	cfg := `
Host production
  HostName prod.example.com
  Port 2222
`
	withFakeSSHConfig(t, cfg, func() {
		got := resolveSSHConfig("staging")
		if got != nil {
			t.Errorf("expected nil for non-matching host, got %+v", got)
		}
	})
}

func TestResolveSSHConfigWildcard(t *testing.T) {
	cfg := `
Host myserver
  HostName myserver.example.com
  Port 3333

Host *
  Port 9999
  User fallback
`
	withFakeSSHConfig(t, cfg, func() {
		// Specific match: myserver gets HostName + Port from its block,
		// User from wildcard (first-match-wins per keyword)
		got := resolveSSHConfig("myserver")
		if got == nil {
			t.Fatal("expected config, got nil")
		}
		if got.HostName != "myserver.example.com" {
			t.Errorf("HostName=%q want 'myserver.example.com'", got.HostName)
		}
		if got.Port != "3333" {
			t.Errorf("Port=%q want '3333' (first match wins)", got.Port)
		}
		if got.User != "fallback" {
			t.Errorf("User=%q want 'fallback' (from wildcard)", got.User)
		}

		// Unknown host: only wildcard matches
		got2 := resolveSSHConfig("unknown")
		if got2 == nil {
			t.Fatal("expected wildcard config, got nil")
		}
		if got2.Port != "9999" {
			t.Errorf("Port=%q want '9999'", got2.Port)
		}
		if got2.User != "fallback" {
			t.Errorf("User=%q want 'fallback'", got2.User)
		}
	})
}

func TestResolveSSHConfigFirstMatchWins(t *testing.T) {
	cfg := `
Host dev
  HostName dev1.example.com
  Port 1111

Host dev
  HostName dev2.example.com
  Port 2222
`
	withFakeSSHConfig(t, cfg, func() {
		got := resolveSSHConfig("dev")
		if got == nil {
			t.Fatal("expected config, got nil")
		}
		// First matching block values win
		if got.HostName != "dev1.example.com" {
			t.Errorf("HostName=%q want 'dev1.example.com'", got.HostName)
		}
		if got.Port != "1111" {
			t.Errorf("Port=%q want '1111'", got.Port)
		}
	})
}

func TestResolveSSHConfigMultipleIdentityFiles(t *testing.T) {
	cfg := `
Host multi
  HostName multi.example.com
  IdentityFile ~/.ssh/id_ed25519
  IdentityFile ~/.ssh/id_rsa_legacy
`
	withFakeSSHConfig(t, cfg, func() {
		got := resolveSSHConfig("multi")
		if got == nil {
			t.Fatal("expected config, got nil")
		}
		if len(got.IdentityFiles) != 2 {
			t.Fatalf("IdentityFiles len=%d want 2", len(got.IdentityFiles))
		}
		home := os.Getenv("HOME")
		want0 := filepath.Join(home, ".ssh", "id_ed25519")
		want1 := filepath.Join(home, ".ssh", "id_rsa_legacy")
		if got.IdentityFiles[0] != want0 {
			t.Errorf("IdentityFiles[0]=%q want %q", got.IdentityFiles[0], want0)
		}
		if got.IdentityFiles[1] != want1 {
			t.Errorf("IdentityFiles[1]=%q want %q", got.IdentityFiles[1], want1)
		}
	})
}

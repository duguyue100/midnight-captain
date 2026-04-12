package appssh

import (
	"os"
	"testing"
)

func TestParseTargetUserAtHost(t *testing.T) {
	user, host, port := parseTarget("alice@example.com")
	if user != "alice" {
		t.Errorf("user=%q want 'alice'", user)
	}
	if host != "example.com" {
		t.Errorf("host=%q want 'example.com'", host)
	}
	if port != "22" {
		t.Errorf("port=%q want '22'", port)
	}
}

func TestParseTargetUserAtHostPort(t *testing.T) {
	user, host, port := parseTarget("bob@myserver.io:2222")
	if user != "bob" {
		t.Errorf("user=%q want 'bob'", user)
	}
	if host != "myserver.io" {
		t.Errorf("host=%q want 'myserver.io'", host)
	}
	if port != "2222" {
		t.Errorf("port=%q want '2222'", port)
	}
}

func TestParseTargetHostOnly(t *testing.T) {
	// No "@" — falls back to $USER env
	os.Setenv("USER", "testuser")
	user, host, port := parseTarget("myhost.local")
	if user != "testuser" {
		t.Errorf("user=%q want 'testuser'", user)
	}
	if host != "myhost.local" {
		t.Errorf("host=%q want 'myhost.local'", host)
	}
	if port != "22" {
		t.Errorf("port=%q want '22'", port)
	}
}

func TestParseTargetHostPortOnly(t *testing.T) {
	os.Setenv("USER", "envuser")
	user, host, port := parseTarget("remotehost:9922")
	if user != "envuser" {
		t.Errorf("user=%q want 'envuser'", user)
	}
	if host != "remotehost" {
		t.Errorf("host=%q want 'remotehost'", host)
	}
	if port != "9922" {
		t.Errorf("port=%q want '9922'", port)
	}
}

func TestParseTargetDefaultPort(t *testing.T) {
	_, _, port := parseTarget("user@host")
	if port != "22" {
		t.Errorf("default port: got %q want '22'", port)
	}
}

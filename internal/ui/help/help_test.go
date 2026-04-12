package help

import (
	"strings"
	"testing"
)

func TestViewInvisible(t *testing.T) {
	m := New()
	got := m.View()
	if got != "" {
		t.Errorf("invisible: want empty string, got %q", got)
	}
}

func TestViewVisible(t *testing.T) {
	m := New()
	m.Open()
	got := string(m.View())
	if got == "" {
		t.Error("visible: want non-empty view")
	}
}

func TestViewVersionShown(t *testing.T) {
	m := New()
	m.Version = "1.2.3"
	m.Open()
	got := string(m.View())
	if !strings.Contains(got, "1.2.3") {
		t.Error("view should contain version '1.2.3'")
	}
}

func TestViewVersionDefaultDev(t *testing.T) {
	m := New()
	m.Version = "" // no version set
	m.Open()
	got := string(m.View())
	if !strings.Contains(got, "dev") {
		t.Error("empty version: view should contain 'dev'")
	}
}

func TestViewCloseThenOpen(t *testing.T) {
	m := New()
	m.Open()
	m.Close()
	if m.Visible {
		t.Error("Visible should be false after Close()")
	}
	if m.View() != "" {
		t.Error("closed: view should be empty")
	}
	m.Open()
	if !m.Visible {
		t.Error("Visible should be true after re-Open()")
	}
	if m.View() == "" {
		t.Error("re-opened: view should be non-empty")
	}
}

func TestViewContainsKeyBindings(t *testing.T) {
	m := New()
	m.Open()
	got := string(m.View())
	// Key bindings always present
	for _, want := range []string{"Navigation", "Actions", "Commands"} {
		if !strings.Contains(got, want) {
			t.Errorf("view missing section %q", want)
		}
	}
}

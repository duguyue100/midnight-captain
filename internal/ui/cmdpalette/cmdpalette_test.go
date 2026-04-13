package cmdpalette

import (
	"strings"
	"testing"

	"charm.land/bubbletea/v2"
)

// helpers

func newModelWithInput(val string) Model {
	m := New()
	m.input.SetValue(val)
	m.refilter()
	return m
}

func keyMsg(s string) tea.KeyPressMsg {
	if len(s) == 1 {
		return tea.KeyPressMsg{Text: s, Code: rune(s[0])}
	}
	// Special keys like "esc", "enter" — Code only
	switch s {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	return tea.KeyPressMsg{}
}

// --- refilter ---

func TestRefilterEmpty(t *testing.T) {
	m := New()
	m.refilter()
	if len(m.filtered) != len(m.commands) {
		t.Errorf("empty input: want all %d commands, got %d", len(m.commands), len(m.filtered))
	}
	if m.Cursor != 0 {
		t.Errorf("cursor should reset to 0")
	}
}

func TestRefilterPrefixMatch(t *testing.T) {
	m := newModelWithInput("so")
	found := false
	for _, c := range m.filtered {
		if c.Name == "sort" {
			found = true
		}
		if !strings.HasPrefix(c.Name, "so") {
			t.Errorf("unexpected command %q in filtered", c.Name)
		}
	}
	if !found {
		t.Error("'sort' should appear for prefix 'so'")
	}
}

func TestRefilterNoMatch(t *testing.T) {
	m := newModelWithInput("zzz")
	if len(m.filtered) != 0 {
		t.Errorf("no match: want 0, got %d", len(m.filtered))
	}
}

func TestRefilterCursorClampedOnShrink(t *testing.T) {
	m := New()
	m.Cursor = 5 // beyond any filtered result
	m.input.SetValue("sort")
	m.refilter()
	if m.Cursor != 0 {
		t.Errorf("cursor should clamp to 0 when filtered shrinks, got %d", m.Cursor)
	}
}

func TestRefilterExactMatch(t *testing.T) {
	m := newModelWithInput("hidden")
	if len(m.filtered) != 1 {
		t.Errorf("exact 'hidden': want 1 result, got %d", len(m.filtered))
	}
	if m.filtered[0].Name != "hidden" {
		t.Errorf("want 'hidden', got %q", m.filtered[0].Name)
	}
}

// --- execute ---

func TestExecuteEmptyClosesNoop(t *testing.T) {
	m := New()
	m.Visible = true
	m2, cmd := m.execute()
	if m2.Visible {
		t.Error("execute on empty should close palette")
	}
	if cmd != nil {
		t.Error("execute on empty should return nil cmd")
	}
}

func TestExecuteExactMatch(t *testing.T) {
	m := New()
	m.Visible = true
	m.input.SetValue("hidden")
	m.refilter()
	m2, cmd := m.execute()
	if m2.Visible {
		t.Error("should close after execute")
	}
	if cmd == nil {
		t.Fatal("expected a cmd from hidden command")
	}
	msg := cmd()
	em, ok := msg.(ExecuteMsg)
	if !ok {
		t.Fatalf("expected ExecuteMsg, got %T", msg)
	}
	if em.Name != "hidden" {
		t.Errorf("got name %q, want 'hidden'", em.Name)
	}
}

func TestExecuteSinglePrefixMatch(t *testing.T) {
	// "find" is the only command starting with "fin"
	m := New()
	m.Visible = true
	m.input.SetValue("fin")
	m.refilter()
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered result for 'fin', got %d", len(m.filtered))
	}
	_, cmd := m.execute()
	if cmd == nil {
		t.Fatal("expected cmd for single prefix match")
	}
	msg := cmd()
	em, ok := msg.(ExecuteMsg)
	if !ok {
		t.Fatalf("expected ExecuteMsg, got %T", msg)
	}
	if em.Name != "find" {
		t.Errorf("got %q, want 'find'", em.Name)
	}
}

func TestExecuteWithArgs(t *testing.T) {
	m := New()
	m.Visible = true
	m.input.SetValue("sort size")
	m.refilter()
	_, cmd := m.execute()
	if cmd == nil {
		t.Fatal("expected cmd")
	}
	msg := cmd()
	em, ok := msg.(ExecuteMsg)
	if !ok {
		t.Fatalf("expected ExecuteMsg, got %T", msg)
	}
	if em.Name != "sort" {
		t.Errorf("got name %q, want 'sort'", em.Name)
	}
	if len(em.Args) == 0 || em.Args[0] != "size" {
		t.Errorf("got args %v, want [size]", em.Args)
	}
}

// --- first key after open ---

func TestFirstKeyAfterOpenProcessed(t *testing.T) {
	m := New()
	m.Open()
	// First key should NOT be dropped — trigger char `:` is consumed
	// by app layer before palette sees it.
	m2, _ := m.Update(keyMsg("s"))
	if m2.input.Value() != "s" {
		t.Errorf("first key should be processed, input=%q want 's'", m2.input.Value())
	}
}

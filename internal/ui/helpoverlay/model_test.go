package helpoverlay

import (
	"strings"
	"testing"
)

func TestNewReturnsHiddenModel(t *testing.T) {
	m := New()
	if m.Visible() {
		t.Error("help overlay should be hidden by default")
	}
}

func TestToggleVisibility(t *testing.T) {
	m := New()

	m.Toggle()
	if !m.Visible() {
		t.Error("help overlay should be visible after first toggle")
	}

	m.Toggle()
	if m.Visible() {
		t.Error("help overlay should be hidden after second toggle")
	}
}

func TestHide(t *testing.T) {
	m := New()
	m.Toggle() // show
	m.Hide()
	if m.Visible() {
		t.Error("help overlay should be hidden after Hide()")
	}
}

func TestSetSize(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Errorf("SetSize failed: got width=%d height=%d, want 120x40", m.width, m.height)
	}
}

func TestViewWhenHidden(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	view := m.View()
	if view != "" {
		t.Error("View() should return empty string when hidden")
	}
}

func TestViewContainsKeybindingSections(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Toggle()
	view := m.View()

	// Check section headers exist
	sections := []string{"Navigation", "Staging", "View", "Commit", "Push"}
	for _, section := range sections {
		if !strings.Contains(view, section) {
			t.Errorf("View() should contain section header %q", section)
		}
	}
}

func TestViewContainsKeyBindings(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Toggle()
	view := m.View()

	// Check some key bindings are present
	bindings := []string{"j/k", "Space", "quit", "help", "push"}
	for _, binding := range bindings {
		if !strings.Contains(view, binding) {
			t.Errorf("View() should contain key binding text %q", binding)
		}
	}
}

func TestViewContainsCloseHint(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Toggle()
	view := m.View()

	if !strings.Contains(view, "?") && !strings.Contains(view, "Esc") {
		t.Error("View() should contain close hint mentioning ? or Esc")
	}
}

func TestViewContainsKeybindingsTitle(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Toggle()
	view := m.View()

	if !strings.Contains(view, "Keybindings") {
		t.Error("View() should contain 'Keybindings' title")
	}
}

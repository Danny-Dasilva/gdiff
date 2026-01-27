package spinner

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
)

func TestNew(t *testing.T) {
	m := New()

	// Should not be spinning initially
	if m.IsSpinning() {
		t.Error("expected spinner to not be spinning initially")
	}

	// Should have empty message initially
	if m.Message() != "" {
		t.Errorf("expected empty message, got %q", m.Message())
	}
}

func TestStart(t *testing.T) {
	m := New()

	cmd := m.Start("Loading...")

	if !m.IsSpinning() {
		t.Error("expected spinner to be spinning after Start")
	}

	if m.Message() != "Loading..." {
		t.Errorf("expected message %q, got %q", "Loading...", m.Message())
	}

	// Should return a command (the spinner tick)
	if cmd == nil {
		t.Error("expected Start to return a command")
	}
}

func TestStop(t *testing.T) {
	m := New()

	m.Start("Loading...")
	m.Stop()

	if m.IsSpinning() {
		t.Error("expected spinner to stop after Stop")
	}

	if m.Message() != "" {
		t.Errorf("expected empty message after Stop, got %q", m.Message())
	}
}

func TestView(t *testing.T) {
	m := New()

	// When not spinning, should return empty
	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when not spinning, got %q", view)
	}

	// When spinning, should return spinner frame + message
	m.Start("Processing...")
	view = m.View()

	// View should contain the message
	if view == "" {
		t.Error("expected non-empty view when spinning")
	}
}

func TestUpdate(t *testing.T) {
	m := New()
	m.Start("Working...")

	// Simulate a spinner tick message
	tickMsg := spinner.TickMsg{}
	newModel, cmd := m.Update(tickMsg)

	// Should still be spinning
	if !newModel.IsSpinning() {
		t.Error("expected spinner to still be spinning after tick")
	}

	// Should return another tick command
	if cmd == nil {
		t.Error("expected Update to return a command for next tick")
	}
}

func TestUpdateWhenNotSpinning(t *testing.T) {
	m := New()

	// When not spinning, Update should be a no-op
	tickMsg := spinner.TickMsg{}
	newModel, cmd := m.Update(tickMsg)

	if newModel.IsSpinning() {
		t.Error("expected spinner to not start spinning from tick when stopped")
	}

	// Should not return a command when stopped
	if cmd != nil {
		t.Error("expected no command when spinner is stopped")
	}
}

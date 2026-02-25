package spinner

import (
	"testing"
)

func TestNew(t *testing.T) {
	m := New()

	if m.IsSpinning() {
		t.Error("expected spinner to not be spinning initially")
	}
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

	if view := m.View(); view != "" {
		t.Errorf("expected empty view when not spinning, got %q", view)
	}

	m.Start("Processing...")
	if view := m.View(); view == "" {
		t.Error("expected non-empty view when spinning")
	}
}

func TestUpdate(t *testing.T) {
	m := New()
	m.Start("Working...")

	newModel, cmd := m.Update(m.spinner.Tick())

	if !newModel.IsSpinning() {
		t.Error("expected spinner to still be spinning after tick")
	}
	if cmd == nil {
		t.Error("expected Update to return a command for next tick")
	}
}

func TestUpdateWhenNotSpinning(t *testing.T) {
	m := New()
	newModel, cmd := m.Update(m.spinner.Tick())

	if newModel.IsSpinning() {
		t.Error("expected spinner to not start spinning from tick when stopped")
	}
	if cmd != nil {
		t.Error("expected no command when spinner is stopped")
	}
}

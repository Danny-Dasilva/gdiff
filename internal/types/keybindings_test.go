package types

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func makeKeyMsg(k string) tea.KeyPressMsg {
	switch k {
	case "ctrl+f":
		return tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl}
	case "ctrl+b":
		return tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "pgup":
		return tea.KeyPressMsg{Code: tea.KeyPgUp}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	default:
		r := []rune(k)
		if len(r) == 1 {
			return tea.KeyPressMsg{Code: r[0], Text: k}
		}
		return tea.KeyPressMsg{}
	}
}

func TestDefaultKeyMap_ModifierKeyBindings(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("FullPageUp binding exists with pgup", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("pgup"), km.FullPageUp) {
			t.Error("FullPageUp should match pgup")
		}
	})

	t.Run("ToggleSidebar binding exists with ctrl+b", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("ctrl+b"), km.ToggleSidebar) {
			t.Error("ToggleSidebar should match ctrl+b")
		}
	})

	t.Run("FullPageDown binding exists with ctrl+f", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("ctrl+f"), km.FullPageDown) {
			t.Error("FullPageDown should match ctrl+f")
		}
	})
}

func TestDefaultKeyMap_RevertItemUsesD(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("RevertItem uses d key", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("d"), km.RevertItem) {
			t.Error("RevertItem should match 'd' key")
		}
	})

	t.Run("RevertItem does not use x key", func(t *testing.T) {
		if key.Matches(makeKeyMsg("x"), km.RevertItem) {
			t.Error("RevertItem should NOT match 'x' key")
		}
	})
}

func TestDefaultKeyMap_SpaceToggle(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("SpaceToggle binding exists with space", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("space"), km.SpaceToggle) {
			t.Error("SpaceToggle should match space key")
		}
	})
}

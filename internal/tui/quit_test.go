package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestQuitKeyQuitsFromTheNormalView(t *testing.T) {
	m := newTestModel(t)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if cmd == nil {
		t.Fatal("expected q to return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestCtrlCQuitsEvenWithTheEggShowing(t *testing.T) {
	m := press(t, newTestModel(t), "C")
	if !m.Egg() {
		t.Fatal("expected the egg to be showing")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("expected ctrl+c to return a quit command even while the egg is showing")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestQKeyWhileEggShowingJustDismisses(t *testing.T) {
	m := press(t, newTestModel(t), "C")
	if !m.Egg() {
		t.Fatal("expected the egg to be showing")
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = next.(Model)

	if cmd != nil {
		t.Error("q while the egg is showing should just dismiss it, not quit")
	}
	if m.Egg() {
		t.Error("q should have dismissed the egg")
	}
}

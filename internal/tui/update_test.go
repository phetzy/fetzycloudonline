package tui

import (
	"io"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	site "github.com/phetzy/fetzycloudonline"
)

// testRenderer returns a renderer with its color profile explicitly forced
// to Ascii (no color), so the golden-file tests are deterministic regardless
// of whatever TTY (or lack of one) the test process happens to have — and,
// crucially, so they build styles from an explicit renderer rather than the
// package-level lipgloss default, which is the renderer this branch stops
// using in the TUI itself. See view_color_test.go for the color-forced
// regression test.
func testRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.Ascii)
	return r
}

// press sends one keystroke through Update and returns the resulting model.
func press(t *testing.T, m Model, keys ...string) Model {
	t.Helper()
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "shift+tab":
			msg = tea.KeyMsg{Type: tea.KeyShiftTab}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case " ":
			// A lone space arrives from bubbletea as its own KeyType, not
			// KeyRunes — see key.go's detectOneMsg — so it must be modeled
			// that way here too, not folded into the default KeyRunes case.
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func newTestModel(t *testing.T) Model {
	t.Helper()
	return New(site.MustLoad(), testRenderer(), os.Stdout).SetSize(120, 40)
}

func TestSelectionMovesAndDoesNotWrap(t *testing.T) {
	m := press(t, newTestModel(t), "tab") // projects
	if m.Selected() != "mapwright" {
		t.Fatalf("after tab, selected = %q, want mapwright", m.Selected())
	}

	m = press(t, m, "j")
	if m.Selected() != "3dpass" {
		t.Errorf("after j, selected = %q, want 3dpass", m.Selected())
	}

	m = press(t, m, "k", "k") // already at the top; must not wrap
	if m.Selected() != "mapwright" {
		t.Errorf("after k k, selected = %q, want mapwright (no wrap)", m.Selected())
	}

	m = press(t, m, "G", "j") // at the end; must not wrap
	if m.Selected() != "oss" {
		t.Errorf("after G j, selected = %q, want oss (no wrap)", m.Selected())
	}
}

func TestFocusSwitchesWithHAndL(t *testing.T) {
	m := newTestModel(t)
	if m.Focus() != FocusList {
		t.Fatalf("initial focus = %v, want list", m.Focus())
	}
	if m = press(t, m, "l"); m.Focus() != FocusViewport {
		t.Errorf("after l, focus = %v, want viewport", m.Focus())
	}
	if m = press(t, m, "h"); m.Focus() != FocusList {
		t.Errorf("after h, focus = %v, want list", m.Focus())
	}
}

func TestTabCyclesAndSelectsFirstSection(t *testing.T) {
	m := newTestModel(t)
	m = press(t, m, "tab")
	if m.Tab() != "projects" || m.Selected() != "mapwright" {
		t.Errorf("got tab=%q selected=%q, want projects/mapwright", m.Tab(), m.Selected())
	}
	m = press(t, m, "tab")
	if m.Tab() != "work" || m.Selected() != "c1" {
		t.Errorf("got tab=%q selected=%q, want work/c1", m.Tab(), m.Selected())
	}
	m = press(t, m, "shift+tab")
	if m.Tab() != "projects" {
		t.Errorf("after shift+tab, tab = %q, want projects", m.Tab())
	}
}

func TestJKScrollsViewportWhenItHasFocus(t *testing.T) {
	// Deliberately narrower than newTestModel's 120x40: at 120x40 the detail
	// viewport is 35 rows tall and no section's rendered body — mapwright
	// included, the longest at 23 lines — exceeds that, so nothing could
	// ever scroll regardless of Update's correctness. 80x24 (the floor size
	// pinned by Task 4's golden) gives mapwright's 32 wrapped lines a 19-row
	// viewport to overflow, which is what this test needs to exercise.
	m := press(t, New(site.MustLoad(), testRenderer(), os.Stdout).SetSize(80, 24), "tab") // projects/mapwright has long content
	m = press(t, m, "l")                                                                  // focus the viewport

	before := m.ViewportOffset()
	m = press(t, m, "j")
	if m.ViewportOffset() <= before {
		t.Errorf("viewport did not scroll: %d -> %d", before, m.ViewportOffset())
	}
	if m.Selected() != "mapwright" {
		t.Errorf("selection changed while the viewport had focus: %q", m.Selected())
	}
}

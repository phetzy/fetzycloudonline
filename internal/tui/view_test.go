package tui

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	site "github.com/phetzy/fetzycloudonline"
)

var update = flag.Bool("update", false, "update golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run with -update to create)", err)
	}
	if got != string(want) {
		t.Errorf("view does not match %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func TestViewAt80x24(t *testing.T) {
	m := New(site.MustLoad(), os.Stdout).SetSize(80, 24)
	golden(t, "view_80x24", m.View())
}

func TestViewAtWideTerminal(t *testing.T) {
	m := New(site.MustLoad(), os.Stdout).SetSize(140, 40)
	golden(t, "view_140x40", m.View())
}

func TestViewNeverExceedsTheTerminalWidth(t *testing.T) {
	for _, w := range []int{60, 80, 100, 140} {
		m := New(site.MustLoad(), os.Stdout).SetSize(w, 24)
		for i, line := range splitLines(m.View()) {
			if width := visibleWidth(line); width > w {
				t.Errorf("at %d columns, line %d is %d wide: %q", w, i, width, line)
			}
		}
	}
}

// TestListPaneScrollsSelectionIntoView reproduces the narrow-terminal trap:
// below wideThreshold the stacked layout gives the list pane only a few
// rows, so hard-truncating at the pane height (rather than scrolling) can
// leave the current selection — and its "›" marker — entirely off-screen
// with no on-screen sign anything changed. At 60x24, the list pane has far
// fewer than 5 rows, so selecting the last of projects' 5 sections requires
// the window to have scrolled.
func TestListPaneScrollsSelectionIntoView(t *testing.T) {
	m := press(t, New(site.MustLoad(), os.Stdout).SetSize(60, 24), "tab") // projects
	m = press(t, m, "j", "j", "j", "j")                                   // mapwright -> open-source (id "oss")
	if m.Selected() != "oss" {
		t.Fatalf("selected = %q, want oss (open-source)", m.Selected())
	}

	view := m.View()
	if !containsLine(view, "open-source") {
		t.Errorf("selected row %q not present in the rendered frame:\n%s", "open-source", view)
	}
	if !containsLine(view, "›") {
		t.Errorf("selection marker not present in the rendered frame:\n%s", view)
	}
}

func containsLine(view, needle string) bool {
	for _, line := range splitLines(view) {
		if strings.Contains(line, needle) {
			return true
		}
	}
	return false
}

func TestViewNeverExceedsTheTerminalHeight(t *testing.T) {
	for _, size := range []struct{ w, h int }{{80, 24}, {100, 30}, {140, 40}} {
		m := New(site.MustLoad(), os.Stdout).SetSize(size.w, size.h)
		if got := len(splitLines(m.View())); got > size.h {
			t.Errorf("at %dx%d the frame is %d lines, which overflows the terminal",
				size.w, size.h, got)
		}
	}
}

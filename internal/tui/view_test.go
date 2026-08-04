package tui

import (
	"flag"
	"os"
	"path/filepath"
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
	for _, w := range []int{80, 100, 140} {
		m := New(site.MustLoad(), os.Stdout).SetSize(w, 24)
		for i, line := range splitLines(m.View()) {
			if width := visibleWidth(line); width > w {
				t.Errorf("at %d columns, line %d is %d wide: %q", w, i, width, line)
			}
		}
	}
}

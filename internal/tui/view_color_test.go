package tui

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	site "github.com/phetzy/fetzycloudonline"
)

// leadingSGR matches an ANSI SGR escape sequence at the very start of a
// string, e.g. "\x1b[38;2;138;173;243m".
var leadingSGR = regexp.MustCompile(`^\x1b\[[0-9;]*m`)

// TestViewUsesColorWhenTheRendererSupportsIt is a regression test for the
// bug where every style was built from the package-level lipgloss default
// renderer instead of a renderer bound to the connecting SSH session. The
// default renderer detects color support from this *process's* own stdout;
// under systemd that's a journald socket, not a TTY, so it silently strips
// color for every visitor no matter what their terminal supports. The
// golden-file tests in view_test.go render through testRenderer(), which is
// deliberately colorless (see update_test.go) and so cannot catch this: a
// renderer that never emits color passes those goldens whether or not the
// production code path is wired up correctly. This test instead forces a
// color-capable profile and asserts the actual ANSI escape sequences for
// specific palette values appear in the rendered frame — the exact
// assertion that would have failed while the bug was live, since
// NewStyles/New took no renderer argument at all and every style silently
// fell back to whatever this test process's own stdout reported.
func TestViewUsesColorWhenTheRendererSupportsIt(t *testing.T) {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)

	m := New(site.MustLoad(), r, os.Stdout).SetSize(80, 24)
	view := m.View()

	// escapeOpen renders a style with the same renderer the view was built
	// from, then extracts the leading ANSI SGR escape sequence it opens
	// with. Deriving the expected code this way — rather than hardcoding
	// the hex-to-decimal RGB conversion by hand — ties the assertion to
	// whatever lipgloss/termenv actually compute for that palette constant,
	// not to a possibly-mistaken manual calculation.
	escapeOpen := func(st lipgloss.Style) string {
		t.Helper()
		sample := st.Render("X")
		open := leadingSGR.FindString(sample)
		if open == "" {
			t.Fatalf("rendered sample %q does not start with an SGR escape", sample)
		}
		if !strings.Contains(open, ";2;") {
			t.Fatalf("rendered sample %q does not look like a truecolor escape", sample)
		}
		return open
	}

	cases := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Title foreground (colAccent)", r.NewStyle().Foreground(colAccent).Bold(true)},
		{"TabActive background (colLavender)", r.NewStyle().Background(colLavender)},
		{"PaneFocused border (colBorderAcc)", r.NewStyle().BorderForeground(colBorderAcc).Border(lipgloss.RoundedBorder())},
	}
	for _, c := range cases {
		want := escapeOpen(c.style)
		if !strings.Contains(view, want) {
			t.Errorf("%s: rendered view does not contain the color escape %q\n--- view ---\n%s", c.name, want, view)
		}
	}
}

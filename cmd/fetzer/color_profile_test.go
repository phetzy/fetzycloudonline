package main

import (
	"testing"

	"github.com/muesli/termenv"
)

// TestSessionColorProfile pins sessionColorProfile's precedence rules. It is
// a regression test for the fix that replaced wishbubbletea.MakeRenderer's
// OSC 11 / device-attributes terminal query (see newSessionRenderer's doc
// comment for why that query can hang forever) with a color profile derived
// purely from the session's environment.
func TestSessionColorProfile(t *testing.T) {
	cases := []struct {
		name    string
		environ []string
		want    termenv.Profile
	}{
		{
			name:    "COLORTERM truecolor",
			environ: []string{"TERM=xterm", "COLORTERM=truecolor"},
			want:    termenv.TrueColor,
		},
		{
			name:    "COLORTERM 24bit",
			environ: []string{"TERM=xterm", "COLORTERM=24bit"},
			want:    termenv.TrueColor,
		},
		{
			name:    "TERM xterm-256color",
			environ: []string{"TERM=xterm-256color"},
			want:    termenv.ANSI256,
		},
		{
			name:    "TERM screen-256color",
			environ: []string{"TERM=screen-256color"},
			want:    termenv.ANSI256,
		},
		{
			name:    "TERM plain xterm",
			environ: []string{"TERM=xterm"},
			want:    termenv.ANSI,
		},
		{
			name:    "TERM dumb",
			environ: []string{"TERM=dumb"},
			want:    termenv.Ascii,
		},
		{
			name:    "TERM empty",
			environ: []string{"TERM="},
			want:    termenv.Ascii,
		},
		{
			name:    "TERM missing entirely",
			environ: []string{},
			want:    termenv.Ascii,
		},
		{
			name:    "COLORTERM truecolor overrides a 256color TERM",
			environ: []string{"TERM=xterm-256color", "COLORTERM=truecolor"},
			want:    termenv.TrueColor,
		},
		{
			name:    "COLORTERM present but not truecolor/24bit falls back to TERM",
			environ: []string{"TERM=xterm-256color", "COLORTERM=yes"},
			want:    termenv.ANSI256,
		},

		// termsAlwaysTrueColor: TERM values termenv's own ColorProfile()
		// (termenv_unix.go) always treats as truecolor, regardless of
		// COLORTERM. OpenSSH does not forward COLORTERM by default, so a
		// visitor on one of these terminals typically arrives with only
		// TERM set. Every entry in the real table is covered below.
		{
			name:    "TERM alacritty (name table, no COLORTERM)",
			environ: []string{"TERM=alacritty"},
			want:    termenv.TrueColor,
		},
		{
			name:    "TERM contour (name table, no COLORTERM)",
			environ: []string{"TERM=contour"},
			want:    termenv.TrueColor,
		},
		{
			name:    "TERM rio (name table, no COLORTERM)",
			environ: []string{"TERM=rio"},
			want:    termenv.TrueColor,
		},
		{
			name:    "TERM wezterm (name table, no COLORTERM)",
			environ: []string{"TERM=wezterm"},
			want:    termenv.TrueColor,
		},
		{
			name:    "TERM xterm-ghostty (name table, no COLORTERM)",
			environ: []string{"TERM=xterm-ghostty"},
			want:    termenv.TrueColor,
		},
		{
			name:    "TERM xterm-kitty (name table, no COLORTERM)",
			environ: []string{"TERM=xterm-kitty"},
			want:    termenv.TrueColor,
		},

		// Precedence: the name table is checked ahead of the generic
		// "256color" substring branch and the ANSI default. None of
		// termenv's own always-truecolor TERM names happen to contain
		// "256color" (so there is no real-world name that would otherwise
		// fall into the 256color branch), but xterm-kitty does contain
		// "color" and would otherwise land on the generic ANSI default —
		// this pins that the name table wins ahead of that fallthrough.
		{
			name:    "name table wins over the generic default ANSI branch",
			environ: []string{"TERM=xterm-kitty"},
			want:    termenv.TrueColor,
		},

		// Not in the name table, and shouldn't be regressed by adding it:
		// these fall through to the existing 256color/default logic exactly
		// as before.
		{
			name:    "TERM foot is not in the name table",
			environ: []string{"TERM=foot"},
			want:    termenv.ANSI,
		},
		{
			name:    "TERM screen is not in the name table",
			environ: []string{"TERM=screen"},
			want:    termenv.ANSI,
		},
		{
			name:    "TERM tmux-256color is not in the name table",
			environ: []string{"TERM=tmux-256color"},
			want:    termenv.ANSI256,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := sessionColorProfile(c.environ)
			if got != c.want {
				t.Errorf("sessionColorProfile(%v) = %v, want %v", c.environ, got, c.want)
			}
		})
	}
}

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

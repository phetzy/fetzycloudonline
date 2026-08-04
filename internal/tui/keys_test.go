package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func flatten(groups [][]key.Binding) []key.Binding {
	var out []key.Binding
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

func TestShortHelpCoversTheDocumentedBindings(t *testing.T) {
	km := DefaultKeyMap()
	var help []string
	for _, b := range km.ShortHelp() {
		help = append(help, b.Help().Key)
	}
	joined := strings.Join(help, " ")

	for _, want := range []string{"j/k", "h/l", "tab", "/", "g/G", "enter"} {
		if !strings.Contains(joined, want) {
			t.Errorf("short help is missing %q; got %q", want, joined)
		}
	}
}

func TestEasterEggIsUndocumented(t *testing.T) {
	km := DefaultKeyMap()
	shown := append(km.ShortHelp(), flatten(km.FullHelp())...)

	for _, b := range shown {
		if strings.Contains(strings.ToLower(b.Help().Desc), "catppuccin") {
			t.Errorf("the egg is described in help: %+v", b.Help())
		}
	}

	// The egg's key must not be reachable through any documented binding.
	eggKeys := map[string]bool{}
	for _, k := range km.Egg.Keys() {
		eggKeys[k] = true
	}
	for _, b := range shown {
		for _, k := range b.Keys() {
			if eggKeys[k] {
				t.Errorf("egg key %q appears in a documented binding: %+v", k, b.Help())
			}
		}
	}
}

package tui

import (
	"bytes"
	"strings"
	"testing"

	site "github.com/phetzy/fetzycloudonline"
)

func TestOSC52EncodesBase64(t *testing.T) {
	got := OSC52("https://mapwright.io")
	want := "\x1b]52;c;aHR0cHM6Ly9tYXB3cmlnaHQuaW8=\x07"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEnterRevealsTheFirstLink(t *testing.T) {
	var out bytes.Buffer
	m := New(site.MustLoad(), &out).SetSize(120, 40)
	m = press(t, m, "tab") // projects/mapwright, which has two links

	m = press(t, m, "enter")

	if got := m.Revealed(); got != "https://mapwright.io" {
		t.Errorf("revealed = %q, want the section's first link", got)
	}
	if !strings.Contains(out.String(), OSC52("https://mapwright.io")) {
		t.Error("the OSC 52 sequence was not written to the session")
	}
	if !strings.Contains(m.View(), "https://mapwright.io") {
		t.Error("the URL should also be displayed, since OSC 52 is not universal")
	}
}

func TestEnterDoesNothingWithoutLinks(t *testing.T) {
	var out bytes.Buffer
	m := New(site.MustLoad(), &out).SetSize(120, 40) // readme has no links

	m = press(t, m, "enter")

	if got := m.Revealed(); got != "" {
		t.Errorf("revealed = %q, want empty for a section with no links", got)
	}
	if out.Len() != 0 {
		t.Errorf("wrote %q to the session for a section with no links", out.String())
	}
}

func TestEggShowsAndAnyKeyDismisses(t *testing.T) {
	m := press(t, newTestModel(t), "C")
	if !m.Egg() {
		t.Fatal("expected the egg to be showing")
	}
	if !strings.Contains(m.View(), "Catppuccin") {
		t.Error("the egg's line should be rendered")
	}

	m = press(t, m, "x")
	if m.Egg() {
		t.Error("any key should dismiss the egg")
	}
}

func TestEggIsNotInTheHelpFooter(t *testing.T) {
	m := newTestModel(t)
	view := m.View()
	if strings.Contains(strings.ToLower(view), "catppuccin") {
		t.Error("the egg must not be discoverable from the default view")
	}
	for _, k := range DefaultKeyMap().Egg.Keys() {
		if strings.Contains(m.HelpView(), k) {
			t.Errorf("egg key %q appears in the help footer", k)
		}
	}
}

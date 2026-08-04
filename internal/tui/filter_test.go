package tui

import "testing"

func TestFilterOpensAndNarrows(t *testing.T) {
	m := press(t, newTestModel(t), "tab") // projects
	m = press(t, m, "/")
	if !m.Filtering() {
		t.Fatal("expected the filter input to be open")
	}
	m = press(t, m, "m", "a", "p")
	if got := m.Filter(); got != "map" {
		t.Errorf("filter = %q, want %q", got, "map")
	}
	if got := m.Status(); got != "1/1 filtered" {
		t.Errorf("status = %q, want %q", got, "1/1 filtered")
	}
}

func TestFilterSearchesEveryTabWhileOpen(t *testing.T) {
	m := newTestModel(t) // starts on readme
	m = press(t, m, "/", "s", "t", "a", "c", "k")

	if m.Selected() != "stack" {
		t.Errorf("selected = %q, want stack — the filter must cross tabs", m.Selected())
	}
	if m.Tab() != "work" {
		t.Errorf("tab = %q, want work — selecting a match carries the tab", m.Tab())
	}
}

func TestEscapeCancelsAndClearsTheFilter(t *testing.T) {
	m := press(t, newTestModel(t), "/", "s", "t", "a", "c", "k")
	m = press(t, m, "esc")

	if m.Filtering() {
		t.Error("filter input should be closed")
	}
	if m.Filter() != "" {
		t.Errorf("filter = %q, want empty", m.Filter())
	}
	// Escape clears the filter but does not rewind where it carried you —
	// this matches the prototype, whose escape sets only filtering and filter.
	if m.Tab() != "work" || m.Selected() != "stack" {
		t.Errorf("got tab=%q selected=%q, want work/stack left in place",
			m.Tab(), m.Selected())
	}
}

func TestEnterKeepsTheFilter(t *testing.T) {
	m := press(t, newTestModel(t), "/", "s", "t", "a", "c", "k")
	m = press(t, m, "enter")

	if m.Filtering() {
		t.Error("filter input should be closed")
	}
	if m.Filter() != "stack" {
		t.Errorf("filter = %q, want it kept as %q", m.Filter(), "stack")
	}
}

func TestSwitchingTabsClearsTheFilter(t *testing.T) {
	m := press(t, newTestModel(t), "/", "s", "t", "a", "c", "k")
	m = press(t, m, "enter") // filter kept, tab is now work
	m = press(t, m, "tab")   // move to contact

	if m.Filter() != "" {
		t.Errorf("filter = %q, want cleared on tab switch", m.Filter())
	}
	if got := m.Status(); got == "no match" {
		t.Error("switching tabs left a stale filter, stranding the list on 'no match'")
	}
}

// TestTypingWhileFilteringDoesNotNavigate guards against Task 5's flat
// switch reclaiming j/k as navigation once the filter input is open: they
// must be typed into the query, not move the selection.
func TestTypingWhileFilteringDoesNotNavigate(t *testing.T) {
	m := press(t, newTestModel(t), "tab") // projects: mapwright, transfer, ...
	before := m.Selected()

	m = press(t, m, "/", "j", "k")
	if m.Filter() != "jk" {
		t.Errorf("filter = %q, want %q — j/k must be typed, not treated as navigation", m.Filter(), "jk")
	}
	if m.Selected() != before {
		t.Errorf("selected changed to %q while typing j/k into the filter", m.Selected())
	}
}

package tui

import (
	"testing"

	site "github.com/phetzy/fetzycloudonline"
)

func ids(sections []site.Section) []string {
	out := make([]string, 0, len(sections))
	for _, s := range sections {
		out = append(out, s.ID)
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestVisibleSectionsNoFilterShowsActiveTab(t *testing.T) {
	c := site.MustLoad()
	got := ids(VisibleSections(c, "projects", "", false))
	want := []string{"mapwright", "3dpass", "transfer", "platform", "hat", "oss"}
	if !equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestVisibleSectionsFilterOpenSearchesEveryTab(t *testing.T) {
	c := site.MustLoad()
	got := ids(VisibleSections(c, "readme", "stack", true))
	if !equal(got, []string{"stack"}) {
		t.Errorf("got %v, want [stack] — the filter must cross tabs while open", got)
	}
}

func TestVisibleSectionsFilterClosedStaysInTab(t *testing.T) {
	c := site.MustLoad()
	got := ids(VisibleSections(c, "readme", "stack", false))
	if len(got) != 0 {
		t.Errorf("got %v, want none — a closed filter stays within the active tab", got)
	}
}

func TestVisibleSectionsMatchesCaseInsensitively(t *testing.T) {
	c := site.MustLoad()
	if got := ids(VisibleSections(c, "projects", "MAPWRIGHT", true)); !equal(got, []string{"mapwright"}) {
		t.Errorf("got %v, want [mapwright]", got)
	}
	if got := ids(VisibleSections(c, "projects", "secure element", true)); !equal(got, []string{"hat"}) {
		t.Errorf("got %v, want [hat] — matching covers title as well as label", got)
	}
}

func TestListStatus(t *testing.T) {
	c := site.MustLoad()
	visible := VisibleSections(c, "projects", "", false)

	if got := ListStatus(visible, "mapwright", ""); got != "1/6" {
		t.Errorf("got %q, want %q", got, "1/6")
	}
	if got := ListStatus(visible, "oss", ""); got != "6/6" {
		t.Errorf("got %q, want %q", got, "6/6")
	}
	filtered := VisibleSections(c, "projects", "map", true)
	if got := ListStatus(filtered, "mapwright", "map"); got != "1/1 filtered" {
		t.Errorf("got %q, want %q", got, "1/1 filtered")
	}
	if got := ListStatus(nil, "mapwright", "zzz"); got != "no match" {
		t.Errorf("got %q, want %q", got, "no match")
	}
}

func TestPromptSegments(t *testing.T) {
	c := site.MustLoad()
	var readme, mapwright, contact site.Section
	for _, s := range c.Sections {
		switch s.ID {
		case "readme":
			readme = s
		case "mapwright":
			mapwright = s
		case "contact":
			contact = s
		}
	}

	if got := PromptSegments(readme); len(got) != 2 {
		t.Errorf("readme: got %d segments, want 2 (directory and branch)", len(got))
	}
	if got := PromptSegments(contact); len(got) != 2 {
		t.Errorf("contact: got %d segments, want 2 — no language module", len(got))
	}
	got := PromptSegments(mapwright)
	if len(got) != 3 {
		t.Fatalf("mapwright: got %d segments, want 3", len(got))
	}
	if got[0].Text != "~/site/projects/mapwright" {
		t.Errorf("first segment: got %q, want the section path", got[0].Text)
	}
}

func TestClampIndexDoesNotWrap(t *testing.T) {
	cases := []struct{ current, delta, length, want int }{
		{0, -1, 5, 0},
		{4, 1, 5, 4},
		{0, 99, 5, 4},
		{4, -99, 5, 0},
		{1, 1, 5, 2},
	}
	for _, tc := range cases {
		if got := ClampIndex(tc.current, tc.delta, tc.length); got != tc.want {
			t.Errorf("ClampIndex(%d, %d, %d) = %d, want %d",
				tc.current, tc.delta, tc.length, got, tc.want)
		}
	}
}

package site

import "testing"

func TestLoadReturnsNineSectionsInOrder(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	want := []string{
		"readme", "mapwright", "transfer", "platform", "hat", "oss", "c1", "stack", "contact",
	}
	if len(c.Sections) != len(want) {
		t.Fatalf("got %d sections, want %d", len(c.Sections), len(want))
	}
	for i, id := range want {
		if c.Sections[i].ID != id {
			t.Errorf("section %d: got %q, want %q", i, c.Sections[i].ID, id)
		}
	}
}

func TestLoadReturnsFourTabs(t *testing.T) {
	c := MustLoad()
	want := []string{"readme", "projects", "work", "contact"}
	if len(c.Tabs) != len(want) {
		t.Fatalf("got %d tabs, want %d", len(c.Tabs), len(want))
	}
	for i, id := range want {
		if c.Tabs[i].ID != id {
			t.Errorf("tab %d: got %q, want %q", i, c.Tabs[i].ID, id)
		}
	}
}

func TestEverySectionBelongsToAKnownTab(t *testing.T) {
	c := MustLoad()
	known := map[string]bool{}
	for _, tb := range c.Tabs {
		known[tb.ID] = true
	}
	for _, s := range c.Sections {
		if !known[s.Tab] {
			t.Errorf("section %q has unknown tab %q", s.ID, s.Tab)
		}
	}
}

func TestSectionIDsAreUnique(t *testing.T) {
	c := MustLoad()
	seen := map[string]bool{}
	for _, s := range c.Sections {
		if seen[s.ID] {
			t.Errorf("duplicate section id %q", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestRequiredFieldsArePresent(t *testing.T) {
	c := MustLoad()
	for _, s := range c.Sections {
		if s.Label == "" || s.Path == "" || s.Title == "" {
			t.Errorf("section %q has an empty required field", s.ID)
		}
		for _, l := range s.Links {
			if l.Label == "" || l.Href == "" {
				t.Errorf("section %q has an incomplete link", s.ID)
			}
		}
	}
}

func TestContactCarriesThreeLinks(t *testing.T) {
	c := MustLoad()
	var contact Section
	for _, s := range c.Sections {
		if s.ID == "contact" {
			contact = s
		}
	}
	want := []string{
		"mailto:david.j.fetzer@gmail.com",
		"https://github.com/phetzy",
		"https://linkedin.com/in/fetzy",
	}
	if len(contact.Links) != len(want) {
		t.Fatalf("got %d links, want %d", len(contact.Links), len(want))
	}
	for i, href := range want {
		if contact.Links[i].Href != href {
			t.Errorf("link %d: got %q, want %q", i, contact.Links[i].Href, href)
		}
	}
}

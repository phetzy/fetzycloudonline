package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	site "github.com/phetzy/fetzycloudonline"
)

// Segment is one piece of the starship-style prompt in the detail pane header.
type Segment struct {
	Text  string
	Color lipgloss.Color
	Bold  bool
}

// langModules mirrors the web build: a starship language module per section,
// in that language's starship colour mapped to Macchiato.
var langModules = map[string]Segment{
	"mapwright": {Text: "\uF308 docker", Color: lipgloss.Color("#8aadf4"), Bold: true},
	"transfer":  {Text: "\uE627 go", Color: lipgloss.Color("#91d7e3"), Bold: true},
	"platform":  {Text: "\uE235 python", Color: lipgloss.Color("#eed49f"), Bold: true},
	"hat":       {Text: "\uF2DB i2c", Color: lipgloss.Color("#8bd5ca"), Bold: true},
	"c1":        {Text: "\uE627 go", Color: lipgloss.Color("#91d7e3"), Bold: true},
	"stack":     {Text: "\uE781 node", Color: lipgloss.Color("#a6da95"), Bold: true},
}

// VisibleSections mirrors src/selectors.ts. While the filter input is open the
// search covers every section, not just the active tab — that is what lets
// selecting a match carry you to another tab. With a filter string but the
// input closed, the active tab still bounds the list.
func VisibleSections(c site.Content, tab, filter string, filtering bool) []site.Section {
	q := strings.ToLower(strings.TrimSpace(filter))

	inTab := make([]site.Section, 0, len(c.Sections))
	for _, s := range c.Sections {
		if s.Tab == tab {
			inTab = append(inTab, s)
		}
	}
	if q == "" {
		return inTab
	}

	pool := inTab
	if filtering {
		pool = c.Sections
	}

	out := make([]site.Section, 0, len(pool))
	for _, s := range pool {
		hay := strings.ToLower(s.Label + " " + s.Title)
		if strings.Contains(hay, q) {
			out = append(out, s)
		}
	}
	return out
}

func ListStatus(visible []site.Section, selectedID, filter string) string {
	if len(visible) == 0 {
		return "no match"
	}
	idx := 0
	for i, s := range visible {
		if s.ID == selectedID {
			idx = i
			break
		}
	}
	status := fmt.Sprintf("%d/%d", idx+1, len(visible))
	if filter != "" {
		status += " filtered"
	}
	return status
}

func PromptSegments(s site.Section) []Segment {
	segs := []Segment{
		{Text: s.Path, Color: lipgloss.Color("#b7bdf8"), Bold: true},
		{Text: "\uE725 main", Color: lipgloss.Color("#c6a0f6"), Bold: true},
	}
	if lang, ok := langModules[s.ID]; ok {
		segs = append(segs, lang)
	}
	return segs
}

func FirstSectionOfTab(c site.Content, tab string) site.Section {
	for _, s := range c.Sections {
		if s.Tab == tab {
			return s
		}
	}
	return c.Sections[0]
}

func ClampIndex(current, delta, length int) int {
	next := current + delta
	if next < 0 {
		return 0
	}
	if next > length-1 {
		return length - 1
	}
	return next
}

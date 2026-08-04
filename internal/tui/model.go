// Package tui implements the Bubble Tea program that serves the same
// content as the web front end over SSH.
package tui

import (
	"io"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	site "github.com/phetzy/fetzycloudonline"
)

// Focus names which pane currently owns j/k and the rest of the movement
// keys: the list of sections, or the detail viewport.
type Focus int

const (
	FocusList Focus = iota
	FocusViewport
)

// Model holds the full state of the program: the frozen content, where the
// visitor currently is within it, and the bubbles components used to render
// that state.
type Model struct {
	content site.Content
	styles  Styles
	keys    KeyMap

	// out is where OSC 52 sequences are written — the SSH session, or
	// stdout when run locally. Task 7 uses it; this task only stores it.
	out io.Writer

	tabIdx    int
	selected  string
	focus     Focus
	filtering bool
	filter    string
	egg       bool
	revealed  string

	viewport viewport.Model
	help     help.Model

	width  int
	height int
}

// New builds a Model rooted at the first tab's first section. w is where
// OSC 52 escape sequences are written once link reveal lands in Task 7.
func New(c site.Content, w io.Writer) Model {
	var tab string
	if len(c.Tabs) > 0 {
		tab = c.Tabs[0].ID
	}

	m := Model{
		content: c,
		styles:  NewStyles(),
		keys:    DefaultKeyMap(),
		out:     w,
		tabIdx:  0,
		focus:   FocusList,

		viewport: viewport.New(0, 0),
		help:     help.New(),
	}
	m.selected = FirstSectionOfTab(c, tab).ID
	m.syncViewport()
	return m
}

// Init satisfies tea.Model. There is nothing to kick off yet.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetSize records the terminal dimensions and resizes the bubbles
// components accordingly. Every layout dimension in view.go derives from
// this — nothing is hardcoded to a particular terminal size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.help.Width = width

	lay := m.computeLayout()
	m.viewport.Width = paneTextWidth(lay.detailOuterW)
	m.viewport.Height = paneStyleDim(lay.detailOuterH)
	m.syncViewport()
	return m
}

// currentTab returns the active tab's id.
func (m Model) currentTab() string {
	if len(m.content.Tabs) == 0 {
		return ""
	}
	if m.tabIdx < 0 || m.tabIdx >= len(m.content.Tabs) {
		return m.content.Tabs[0].ID
	}
	return m.content.Tabs[m.tabIdx].ID
}

// visibleSections returns the sections the list pane currently shows.
func (m Model) visibleSections() []site.Section {
	return VisibleSections(m.content, m.currentTab(), m.filter, m.filtering)
}

// selectedSection returns the currently selected section, falling back to
// the zero value if content is somehow empty.
func (m Model) selectedSection() site.Section {
	for _, s := range m.content.Sections {
		if s.ID == m.selected {
			return s
		}
	}
	return site.Section{}
}

// syncViewport rewrites the viewport's content from the selected section,
// wrapped to the viewport's current width. This takes a pointer receiver
// deliberately: Model's other methods are value receivers so that Update
// and SetSize can return a modified copy in the Elm architecture style, but
// that means a value-receiver syncViewport would mutate a throwaway copy.
func (m *Model) syncViewport() {
	m.viewport.SetContent(renderDetailBody(m.styles, m.selectedSection(), m.viewport.Width))
}

// Accessors used by later tasks and their tests.

func (m Model) Selected() string    { return m.selected }
func (m Model) Tab() string         { return m.currentTab() }
func (m Model) Focus() Focus        { return m.focus }
func (m Model) ViewportOffset() int { return m.viewport.YOffset }
func (m Model) Filtering() bool     { return m.filtering }
func (m Model) Filter() string      { return m.filter }
func (m Model) Egg() bool           { return m.egg }
func (m Model) Revealed() string    { return m.revealed }
func (m Model) HelpView() string    { return m.help.View(m.keys) }

func (m Model) Status() string {
	return ListStatus(m.visibleSections(), m.selected, m.filter)
}

// Package tui implements the Bubble Tea program that serves the same
// content as the web front end over SSH.
package tui

import (
	"io"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	content  site.Content
	styles   Styles
	renderer *lipgloss.Renderer
	keys     KeyMap

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
//
// r is the renderer every style is built from. It must be bound to the
// connecting session (see wishbubbletea.MakeRenderer) rather than being the
// package-level default renderer, whose color detection reads this
// process's own stdout — under systemd that's a journald socket, not a TTY,
// which would strip color for every visitor regardless of their terminal.
// Deliberately explicit rather than a package-level variable: a global
// renderer would be shared across concurrent sessions with different
// terminal capabilities.
func New(c site.Content, r *lipgloss.Renderer, w io.Writer) Model {
	var tab string
	if len(c.Tabs) > 0 {
		tab = c.Tabs[0].ID
	}

	m := Model{
		content:  c,
		styles:   NewStyles(r),
		renderer: r,
		keys:     DefaultKeyMap(),
		out:      w,
		tabIdx:   0,
		focus:    FocusList,

		viewport: viewport.New(0, 0),
		help:     help.New(),
	}
	// help.New() builds its own Styles from the package-level lipgloss
	// default renderer with no way to inject one — see newHelpStyles for
	// why that needs correcting the same way every other style here was.
	m.help.Styles = newHelpStyles(r)
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
	m.viewport.SetContent(renderDetailBody(m.renderer, m.styles, m.selectedSection(), m.viewport.Width))
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

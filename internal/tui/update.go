package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	site "github.com/phetzy/fetzycloudonline"
)

// Update satisfies tea.Model. This task wires up navigation and pane focus;
// filtering and link reveal land in later tasks.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.SetSize(msg.Width, msg.Height), nil
	case tea.KeyMsg:
		return m.handleKey(msg), nil
	}
	return m, nil
}

// handleKey dispatches a single keystroke. It mirrors the web build: j/k
// move the selection when the list has focus and scroll the viewport when
// the viewport has focus; h/l switch focus; tab/shift+tab cycle tabs and
// select the new tab's first section; g/G jump to the ends; selection never
// wraps.
func (m Model) handleKey(msg tea.KeyMsg) Model {
	// While the filter input is open, every keystroke belongs to it — j/k
	// included — so this branch must run before any navigation case gets a
	// chance to claim them.
	if m.filtering {
		return m.handleFilterKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Filter):
		m.filtering = true
		return m
	case key.Matches(msg, m.keys.NextTab):
		return m.switchTab(1)
	case key.Matches(msg, m.keys.PrevTab):
		return m.switchTab(-1)
	case key.Matches(msg, m.keys.Left):
		m.focus = FocusList
		return m
	case key.Matches(msg, m.keys.Right):
		m.focus = FocusViewport
		return m
	case key.Matches(msg, m.keys.Down):
		return m.move(1)
	case key.Matches(msg, m.keys.Up):
		return m.move(-1)
	case key.Matches(msg, m.keys.Top):
		return m.jump(true)
	case key.Matches(msg, m.keys.Bottom):
		return m.jump(false)
	case key.Matches(msg, m.keys.PageDn):
		return m.page(1)
	case key.Matches(msg, m.keys.PageUp):
		return m.page(-1)
	}
	return m
}

// handleFilterKey dispatches a keystroke while the filter input is open.
// Escape cancels: it closes the input and clears the query, but does not
// rewind the tab or selection to wherever they were before — the visitor
// stays where the filter carried them, matching the prototype. Enter closes
// the input and keeps the query. Backspace and printable runes edit the
// query and re-run the cross-tab search each time.
func (m Model) handleFilterKey(msg tea.KeyMsg) Model {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.filtering = false
		m.filter = ""
		return m
	case key.Matches(msg, m.keys.Accept):
		m.filtering = false
		return m
	case msg.Type == tea.KeyBackspace:
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
		}
		return m.reselectAfterFilterChange()
	case msg.Type == tea.KeyRunes:
		m.filter += string(msg.Runes)
		return m.reselectAfterFilterChange()
	}
	return m
}

// reselectAfterFilterChange mirrors the web build's onFilter: whenever the
// query changes, if the current selection fell out of the now-narrower
// cross-tab list, jump to the first remaining match — switching the active
// tab to match it, since that match may live on another tab entirely.
func (m Model) reselectAfterFilterChange() Model {
	visible := m.visibleSections()
	if len(visible) == 0 {
		return m
	}
	for _, s := range visible {
		if s.ID == m.selected {
			return m
		}
	}
	return m.selectAcrossTabs(visible[0])
}

// selectAcrossTabs selects s, switching the active tab to the one s belongs
// to if it differs from the current one. Filtering while the input is open
// searches every section, so a match can live on a tab other than the one
// currently on screen.
func (m Model) selectAcrossTabs(s site.Section) Model {
	for i, t := range m.content.Tabs {
		if t.ID == s.Tab {
			m.tabIdx = i
			break
		}
	}
	m.selected = s.ID
	m.syncViewport()
	return m
}

// move advances or retreats the selection by one when the list has focus,
// or scrolls the viewport by one line when the viewport has focus.
func (m Model) move(delta int) Model {
	if m.focus == FocusViewport {
		if delta > 0 {
			m.viewport.ScrollDown(1)
		} else {
			m.viewport.ScrollUp(1)
		}
		return m
	}
	return m.moveSelection(delta)
}

// jump moves to the first or last item (list focus) or the top/bottom of
// the viewport (viewport focus).
func (m Model) jump(first bool) Model {
	if m.focus == FocusViewport {
		if first {
			m.viewport.GotoTop()
		} else {
			m.viewport.GotoBottom()
		}
		return m
	}
	visible := m.visibleSections()
	if len(visible) == 0 {
		return m
	}
	if first {
		return m.selectByIndex(visible, 0)
	}
	return m.selectByIndex(visible, len(visible)-1)
}

// page pages the focused pane by a full page in the given direction.
func (m Model) page(dir int) Model {
	if m.focus == FocusViewport {
		if dir > 0 {
			m.viewport.PageDown()
		} else {
			m.viewport.PageUp()
		}
		return m
	}
	visible := m.visibleSections()
	if len(visible) == 0 {
		return m
	}
	page := m.viewport.Height
	if page < 1 {
		page = 1
	}
	return m.moveSelection(dir * page)
}

// moveSelection clamps the selection within the visible list — it never
// wraps — using ClampIndex from Task 2.
func (m Model) moveSelection(delta int) Model {
	visible := m.visibleSections()
	if len(visible) == 0 {
		return m
	}
	idx := 0
	for i, s := range visible {
		if s.ID == m.selected {
			idx = i
			break
		}
	}
	next := ClampIndex(idx, delta, len(visible))
	return m.selectByIndex(visible, next)
}

// selectByIndex selects the section at idx within visible and resyncs the
// viewport to show it.
func (m Model) selectByIndex(visible []site.Section, idx int) Model {
	m.selected = visible[idx].ID
	m.syncViewport()
	return m
}

// switchTab cycles the active tab by delta (1 for next, -1 for previous)
// and selects the new tab's first section.
func (m Model) switchTab(delta int) Model {
	n := len(m.content.Tabs)
	if n == 0 {
		return m
	}
	m.tabIdx = ((m.tabIdx+delta)%n + n) % n
	m.selected = FirstSectionOfTab(m.content, m.currentTab()).ID
	m.focus = FocusList
	m.filter = "" // switching tabs clears the filter — do not strand the list on "no match".
	m.syncViewport()
	return m
}

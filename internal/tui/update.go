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
	switch {
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
	m.syncViewport()
	return m
}

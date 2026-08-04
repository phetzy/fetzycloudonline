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
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey dispatches a single keystroke. It mirrors the web build: j/k
// move the selection when the list has focus and scroll the viewport when
// the viewport has focus; h/l switch focus; tab/shift+tab cycle tabs and
// select the new tab's first section; g/G jump to the ends; selection never
// wraps.
func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	// ctrl+c always quits, regardless of mode — including while the filter
	// input is open or the egg is showing. It is the terminal's universal
	// escape hatch; having an undocumented easter egg swallow it, or the
	// filter guard eat it as text, would be a bad surprise. This check must
	// run before both of those guards.
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// While the filter input is open, every keystroke belongs to it — j/k
	// included — so this branch must run before any navigation case gets a
	// chance to claim them. It also outranks the egg: typing "C" into a
	// filter must insert a "C", not trigger the egg.
	if m.filtering {
		return m.handleFilterKey(msg), nil
	}

	// While the egg is showing, any key dismisses it and does nothing else.
	// (ctrl+c already returned above, so this only ever sees other keys,
	// such as "q" — which just dismisses, matching "any key returns".)
	if m.egg {
		m.egg = false
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Egg):
		m.egg = true
		return m, nil
	case key.Matches(msg, m.keys.Accept):
		return m.revealLink(), nil
	case key.Matches(msg, m.keys.Filter):
		m.filtering = true
		return m, nil
	case key.Matches(msg, m.keys.Cancel):
		// Reachable here whenever the filter input is closed but a stale
		// query is still narrowing (or emptying) the list — e.g. after
		// enter closed the input on a no-match query. handleFilterKey
		// covers esc while the input is open; this covers esc afterward,
		// matching the web build's global Escape handler, which clears the
		// filter unconditionally rather than only while its own input has
		// focus.
		m.filter = ""
		return m, nil
	case key.Matches(msg, m.keys.NextTab):
		return m.switchTab(1), nil
	case key.Matches(msg, m.keys.PrevTab):
		return m.switchTab(-1), nil
	case key.Matches(msg, m.keys.Left):
		m.focus = FocusList
		return m, nil
	case key.Matches(msg, m.keys.Right):
		m.focus = FocusViewport
		return m, nil
	case key.Matches(msg, m.keys.Down):
		return m.move(1), nil
	case key.Matches(msg, m.keys.Up):
		return m.move(-1), nil
	case key.Matches(msg, m.keys.Top):
		return m.jump(true), nil
	case key.Matches(msg, m.keys.Bottom):
		return m.jump(false), nil
	case key.Matches(msg, m.keys.PageDn):
		return m.page(1), nil
	case key.Matches(msg, m.keys.PageUp):
		return m.page(-1), nil
	}
	return m, nil
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
	case msg.Type == tea.KeyRunes, msg.Type == tea.KeySpace:
		// tea.KeySpace is a distinct Type from KeyRunes for a lone space (see
		// bubbletea's key.go): a run of runes is reported as KeyRunes, but a
		// run of exactly one rune that is a space is reported as KeySpace
		// instead, with Runes still populated. Without this case, a space
		// typed into the filter (as opposed to pasted as part of a longer
		// run, which stays KeyRunes) is silently swallowed.
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
	return m.selectID(s.ID)
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
	return m.selectID(visible[idx].ID)
}

// selectID is the single funnel every selection change passes through: it
// sets m.selected, resyncs the viewport, and clears any previously revealed
// link. A revealed URL in the footer describes the section that was
// current when enter was pressed; once the selection moves on, leaving it
// up would misleadingly imply it belongs to the new section.
func (m Model) selectID(id string) Model {
	m.selected = id
	m.revealed = ""
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
	m.focus = FocusList
	m.filter = "" // switching tabs clears the filter — do not strand the list on "no match".
	return m.selectID(FirstSectionOfTab(m.content, m.currentTab()).ID)
}

// revealLink is enter's handler: a server cannot open the visitor's
// browser, so it copies the selected section's first link to the visitor's
// clipboard via an OSC 52 escape sequence written to the model's writer,
// and records the URL so the view can also display it — OSC 52 is not
// universally supported, so the display is what always works. Sections
// with no links (most of them) do nothing.
func (m Model) revealLink() Model {
	links := m.selectedSection().Links
	if len(links) == 0 {
		return m
	}
	url := links[0].Href
	m.revealed = url
	if m.out != nil {
		_, _ = m.out.Write([]byte(OSC52(url)))
	}
	return m
}

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	site "github.com/phetzy/fetzycloudonline"
)

// listPaneWidth is the outer width of the list pane once the terminal is
// wide enough to show both panes side by side.
const listPaneWidth = 26

// wideThreshold is the narrowest terminal width at which the panes sit side
// by side. Below it they stack vertically.
const wideThreshold = 70

// The pane styles (see styles.go) are Border(RoundedBorder()).Padding(0, 1):
// one column of border on each side, one column of padding on each side,
// and no vertical padding.
//
// lipgloss.Style.Width(n) treats n as the padded content box — the border
// adds 2 more columns on top of it — so the value to pass to .Width() is
// the outer pane width minus the border only. The text actually available
// for content (what a viewport or a truncated label must fit within) is
// narrower still, by the horizontal padding on top of that.
//
// Style.Height(n) has no such split: Padding(0, 1) carries no vertical
// padding, so the value passed to .Height() and the number of content rows
// are the same number.
const (
	paneBorderOverhead  = 2 // one column/row per side, width and height
	panePaddingOverhead = 2 // one column per side, width only
)

// paneStyleDim is the value to give lipgloss.Style.Width or .Height for a
// pane whose outer (bordered) size should be outer.
func paneStyleDim(outer int) int {
	d := outer - paneBorderOverhead
	if d < 0 {
		return 0
	}
	return d
}

// paneTextWidth is how many columns of actual text a pane of outer width
// outer can hold — narrower than paneStyleDim(outer) by the horizontal
// padding, which paneStyleDim's return value still has to include.
func paneTextWidth(outer int) int {
	d := outer - paneBorderOverhead - panePaddingOverhead
	if d < 0 {
		return 0
	}
	return d
}

// frameLayout is the terminal size broken down into the frame's regions.
// Every field derives from Model.width/height; nothing here is a constant
// tied to a specific terminal size.
type frameLayout struct {
	stacked      bool
	listOuterW   int
	listOuterH   int
	detailOuterW int
	detailOuterH int
}

// computeLayout works out where the title block, tab bar, panes, and help
// footer sit for the model's current width and height. It falls back to
// 80x24 only as a defensive default for the pathological case of View being
// called before any size has ever been set — every real render path goes
// through SetSize first, which is what tea.WindowSizeMsg drives.
func (m Model) computeLayout() frameLayout {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	fixed := lipgloss.Height(m.renderTitle(w)) +
		lipgloss.Height(m.renderTabBar(w)) +
		lipgloss.Height(m.renderFooter(w))
	mainH := h - fixed
	if mainH < paneBorderOverhead+1 {
		mainH = paneBorderOverhead + 1
	}

	if w >= wideThreshold {
		listW := listPaneWidth
		if listW > w {
			listW = w
		}
		detailW := w - listW
		if detailW < 0 {
			detailW = 0
		}
		return frameLayout{
			stacked:      false,
			listOuterW:   listW,
			listOuterH:   mainH,
			detailOuterW: detailW,
			detailOuterH: mainH,
		}
	}

	listH := mainH / 3
	minPane := paneBorderOverhead + 1
	if listH < minPane {
		listH = minPane
	}
	detailH := mainH - listH
	if detailH < minPane {
		detailH = minPane
	}
	return frameLayout{
		stacked:      true,
		listOuterW:   w,
		listOuterH:   listH,
		detailOuterW: w,
		detailOuterH: detailH,
	}
}

// renderTitle is the identity line: the readme section's title and kicker,
// which is the same identity block regardless of which tab is active.
func (m Model) renderTitle(width int) string {
	id := m.identitySection()
	line := m.styles.Title.Render(strings.ToUpper(id.Title))
	if id.Kicker != "" {
		line += m.styles.TitleSub.Render(" · " + id.Kicker)
	}
	return truncateToWidth(line, width)
}

func (m Model) identitySection() site.Section {
	for _, s := range m.content.Sections {
		if s.ID == "readme" {
			return s
		}
	}
	if len(m.content.Sections) > 0 {
		return m.content.Sections[0]
	}
	return site.Section{}
}

// renderTabBar renders one tab per entry in content.Tabs, marking the
// active one with a leading bar, and never grows past width.
func (m Model) renderTabBar(width int) string {
	parts := make([]string, 0, len(m.content.Tabs))
	for i, t := range m.content.Tabs {
		if i == m.tabIdx {
			parts = append(parts, m.styles.TabActive.Render("▌ "+t.Label))
			continue
		}
		parts = append(parts, m.styles.TabInactive.Render(t.Label))
	}
	return truncateToWidth(strings.Join(parts, " "), width)
}

// renderHelp is the key-binding footer, generated from the same KeyMap that
// handles input so it cannot drift from actual behavior. While the filter
// input is open it is replaced by renderFilterBar — see View.
func (m Model) renderHelp() string {
	return m.styles.Help.Render(m.help.View(m.keys))
}

// renderFilterBar is the footer shown while the filter input is open,
// mirroring the design reference's "filter> " prompt and live query.
func (m Model) renderFilterBar(width int) string {
	prompt := m.renderer.NewStyle().Foreground(colGreen).Render("filter> ")
	query := m.styles.Help.Render(m.filter)
	return truncateToWidth(prompt+query, width)
}

// renderRevealedBar is the footer shown after enter reveals a link: the URL
// that was written to the clipboard via OSC 52. Since OSC 52 support is not
// universal, this is what always works — the copy is a bonus, not the only
// way the visitor sees the address.
func (m Model) renderRevealedBar(width int) string {
	label := m.renderer.NewStyle().Foreground(colGreen).Render("copied ")
	url := m.styles.Help.Render(m.revealed)
	return truncateToWidth(label+url, width)
}

// renderFooter is the frame's bottom line: the filter input while it is
// open, the just-revealed URL after enter, the key-binding help otherwise.
// All three are always exactly one line, so which one is showing does not
// change computeLayout's fixed height.
func (m Model) renderFooter(width int) string {
	if m.filtering {
		return m.renderFilterBar(width)
	}
	if m.revealed != "" {
		return m.renderRevealedBar(width)
	}
	return m.renderHelp()
}

// scrollWindow returns the index of the first row to show in a list of
// length total, height rows tall, so that idx stays visible — the same
// keep-selection-in-view intent as the web build's ListPane effect (which
// nudges box.scrollTop just far enough to bring the selected button back
// into its viewport). Recomputed from scratch on every render rather than
// carried as persisted scroll state, so it centers idx in the window
// whenever there's room, and otherwise clamps to the start or end of the
// list — idx is always within [start, start+height) for any height > 0 and
// total > 0.
func scrollWindow(idx, total, height int) int {
	if height <= 0 || total <= height {
		return 0
	}
	start := idx - height/2
	if start < 0 {
		start = 0
	}
	if maxStart := total - height; start > maxStart {
		start = maxStart
	}
	return start
}

// renderListPane renders the section list: one line per visible section,
// marking the current selection, scrolled so the selection is always in the
// visible window, plus one status line (ListStatus) pinned to the bottom of
// the pane — mirroring the web build's ListPane, which keeps the selected
// button in its scroll viewport and shows the same status line beneath it.
func (m Model) renderListPane(outerW, outerH int, focused bool) string {
	style := m.styles.PaneUnfocused
	if focused {
		style = m.styles.PaneFocused
	}
	styleW := paneStyleDim(outerW)
	styleH := paneStyleDim(outerH)
	textW := paneTextWidth(outerW)

	sections := m.visibleSections()

	rowsH := styleH - 1 // one row reserved for the status line below
	if rowsH < 0 {
		rowsH = 0
	}

	idx := 0
	for i, s := range sections {
		if s.ID == m.selected {
			idx = i
			break
		}
	}
	start := scrollWindow(idx, len(sections), rowsH)
	end := start + rowsH
	if end > len(sections) {
		end = len(sections)
	}

	lines := make([]string, 0, end-start+1)
	for _, s := range sections[start:end] {
		marker := "  "
		if s.ID == m.selected {
			marker = "› "
		}
		line := truncateToWidth(marker+s.Label, textW)
		if s.ID == m.selected {
			line = m.styles.List.Bold(true).Render(line)
		} else {
			line = m.styles.List.Render(line)
		}
		lines = append(lines, line)
	}
	lines = append(lines, truncateToWidth(m.styles.Help.Render(m.Status()), textW))

	if len(lines) > styleH && styleH >= 0 {
		lines = lines[:styleH]
	}

	return style.Width(styleW).Height(styleH).Render(strings.Join(lines, "\n"))
}

// renderDetailPane renders the starship-style prompt header and the
// selected section's body inside the bubbles viewport, which owns
// scrolling from Task 5 on. While the egg is showing, it replaces the
// viewport's content entirely; any key dismisses it and returns to the
// section that was on screen.
func (m Model) renderDetailPane(outerW, outerH int, focused bool) string {
	style := m.styles.PaneUnfocused
	if focused {
		style = m.styles.PaneFocused
	}
	styleW := paneStyleDim(outerW)
	styleH := paneStyleDim(outerH)

	body := m.viewport.View()
	if m.egg {
		body = renderEgg(m.renderer, m.styles)
	}
	return style.Width(styleW).Height(styleH).Render(body)
}

// renderEgg renders the undocumented easter egg's line: "Catppuccin" in the
// accent colour, italic, the rest in subtext0.
func renderEgg(r *lipgloss.Renderer, styles Styles) string {
	accent := r.NewStyle().Foreground(colAccent).Italic(true).Render("Catppuccin")
	rest := styles.TitleSub.Render(" enjoyer")
	lead := styles.TitleSub.Render("Yes, I am a ")
	return lead + accent + rest
}

// renderDetailBody builds the detail viewport's content: the prompt line,
// then the section's kicker, paragraphs, rows, and links, each word-wrapped
// to width rather than left to overflow the pane.
func renderDetailBody(r *lipgloss.Renderer, styles Styles, s site.Section, width int) string {
	if width <= 0 {
		return ""
	}
	wrap := r.NewStyle().Width(width)

	var b strings.Builder
	b.WriteString(renderPrompt(r, s))
	b.WriteString("\n\n")

	if s.Kicker != "" {
		b.WriteString(styles.TitleSub.Render(wrap.Render(s.Kicker)))
		b.WriteString("\n\n")
	}

	for i, p := range s.Paras {
		b.WriteString(styles.Detail.Render(wrap.Render(p)))
		if i != len(s.Paras)-1 {
			b.WriteString("\n\n")
		}
	}

	if len(s.Rows) > 0 {
		b.WriteString("\n\n")
		for i, r := range s.Rows {
			b.WriteString(styles.Detail.Render(wrap.Render(r.Label + ": " + r.Body)))
			if i != len(s.Rows)-1 {
				b.WriteString("\n")
			}
		}
	}

	if len(s.Links) > 0 {
		b.WriteString("\n\n")
		for i, l := range s.Links {
			b.WriteString(styles.Detail.Render(wrap.Render(l.Label + " -> " + l.Href)))
			if i != len(s.Links)-1 {
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

// renderPrompt is the starship-style header: directory, branch, an optional
// language module, then the cat glyph and the prompt arrow.
func renderPrompt(r *lipgloss.Renderer, s site.Section) string {
	segs := PromptSegments(s)
	parts := make([]string, 0, len(segs)+2)
	for _, seg := range segs {
		st := r.NewStyle().Foreground(seg.Color)
		if seg.Bold {
			st = st.Bold(true)
		}
		parts = append(parts, st.Render(seg.Text))
	}
	// The cat glyph, matching src/components/Prompt.tsx's "U+F0D1B". Written
	// as a Go escape rather than pasted so a copy cannot silently flatten it.
	parts = append(parts, r.NewStyle().Foreground(colGreen).Render("\U000F0D1B"))
	parts = append(parts, r.NewStyle().Foreground(colAccent).Bold(true).Render("❯"))
	return strings.Join(parts, " ")
}

// View composes the full frame: title block, tab bar, the list and detail
// panes (side by side or stacked, per computeLayout), and the help footer.
func (m Model) View() string {
	lay := m.computeLayout()

	w := m.width
	if w <= 0 {
		w = 80
	}

	title := m.renderTitle(w)
	tabs := m.renderTabBar(w)
	helpFooter := m.renderFooter(w)

	list := m.renderListPane(lay.listOuterW, lay.listOuterH, m.focus == FocusList)
	detail := m.renderDetailPane(lay.detailOuterW, lay.detailOuterH, m.focus == FocusViewport)

	var main string
	if lay.stacked {
		main = lipgloss.JoinVertical(lipgloss.Left, list, detail)
	} else {
		main = lipgloss.JoinHorizontal(lipgloss.Top, list, detail)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, tabs, main, helpFooter)
}

// truncateToWidth cuts s to fit within width display columns, appending an
// ellipsis when it had to cut — long values are truncated rather than left
// to widen the frame. It is ANSI-aware, since the strings it is given are
// already styled and a byte- or rune-level cut could sever an escape
// sequence.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(s, width, "…")
}

// splitLines splits a rendered frame into its display lines.
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// visibleWidth measures a line's display width, excluding ANSI escapes and
// accounting for wide runes — lipgloss.Width, not len(), because byte
// length is meaningless once color codes and wide characters are involved.
func visibleWidth(s string) int {
	return lipgloss.Width(s)
}

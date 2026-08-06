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

// tallThreshold is the shortest terminal height at which the header renders
// as the full two-card form (renderNameCard + renderMetaCard, side by side
// or stacked per wideThreshold). Below it — the visitor's terminal has few
// rows to spare — the header switches to renderCompactCardHeader: the same
// content merged into a single card, trading the second card's border for
// two more rows handed back to the detail pane. 30 is comfortably above the
// 80x24 minimum this project supports while leaving ample headroom below
// any terminal worth calling "generous" (the 140x40 fixture the header's
// full form must keep matching exactly); it is compared against
// Model.height on every render, never assumed.
const tallThreshold = 30

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

// cardPaddingOverhead is the header cards' horizontal padding (Padding(0,
// 2): two columns on each side), wider than the panes' so the name/meta
// boxes read as their own distinct chrome rather than matching pane insets
// exactly — the same split paneStyleDim/paneTextWidth apply to panes.
const cardPaddingOverhead = 4

// cardStyleDim is the value to give a header card's lipgloss.Style.Width for
// an outer (bordered) size of outer.
func cardStyleDim(outer int) int {
	d := outer - paneBorderOverhead
	if d < 0 {
		return 0
	}
	return d
}

// cardTextWidth is how many columns of text a header card of outer width
// outer can hold.
func cardTextWidth(outer int) int {
	d := outer - paneBorderOverhead - cardPaddingOverhead
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

// computeLayout works out where the header row, tab bar, panes, and help
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

	fixed := lipgloss.Height(m.renderHeaderRow(w, h)) +
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

// identitySection returns the readme section, which carries the site-wide
// identity (name, role/kicker) regardless of which tab is active.
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

// rowByLabel finds a row by its label (case-insensitive), so the header
// card can pull "role" and "status" out of the readme section's own rows
// rather than duplicating their copy.
func rowByLabel(s site.Section, label string) string {
	for _, r := range s.Rows {
		if strings.EqualFold(r.Label, label) {
			return r.Body
		}
	}
	return ""
}

// letterSpace inserts a thin space between every rune of s, and widens
// existing word spaces, approximating the web build's letter-spacing on
// "DAVID FETZER" within a monospace grid.
func letterSpace(s string) string {
	runes := []rune(s)
	parts := make([]string, 0, len(runes))
	for _, r := range runes {
		if r == ' ' {
			parts = append(parts, "  ")
			continue
		}
		parts = append(parts, string(r))
	}
	return strings.Join(parts, " ")
}

// renderNameCard renders the bordered "DAVID FETZER" identity box: the
// title, letterspaced when there is room for it, and the readme section's
// kicker beneath it as the subtitle. Both strings come straight from
// content.json (Title and Kicker) — nothing here is invented copy.
func (m Model) renderNameCard(outerW int) string {
	id := m.identitySection()
	title := strings.ToUpper(id.Title)
	spaced := letterSpace(title)

	textW := cardTextWidth(outerW)
	styleW := cardStyleDim(outerW)

	titleLine := title
	if textW >= lipgloss.Width(spaced) {
		titleLine = spaced
	}

	lines := []string{
		m.styles.Title.Render(truncateToWidth(titleLine, textW)),
	}
	if id.Kicker != "" {
		lines = append(lines, m.styles.TitleSub.Render(truncateToWidth(id.Kicker, textW)))
	}

	return m.styles.NameCard.Width(styleW).Render(strings.Join(lines, "\n"))
}

// shortRole trims content.json's role row down to the clause before the
// first comma — "Software Engineer at C1, March 2022 – present" becomes
// "Software Engineer at C1". It is a substring of the exact content.json
// text, not paraphrased copy: the dates stay intact in the metadata table
// below, and repeating them in the meta card unbalanced it against the
// name card beside it.
func shortRole(role string) string {
	if i := strings.Index(role, ","); i >= 0 {
		return role[:i]
	}
	return role
}

// buildsLine is the meta card's third line. It is not in content.json —
// it is hardcoded, verbatim, in the web build's src/components/HeaderRow.tsx
// (lines 21-22). Reproduced here character-for-character (including the
// middot separators) because it is real, already-shipped site copy, not
// invented text; see the report for why it lives here as a constant
// instead of coming from content.json like everything else in this file.
const buildsLine = "map servers · CLI tools · self-hosted platforms · hardware"

// renderMetaCard renders the bordered role/status/builds box beside the
// name card: "role" and "status" pulled from the readme section's own rows
// in content.json, with a green status dot, plus the "builds" line from
// HeaderRow.tsx (see buildsLine).
func (m Model) renderMetaCard(outerW int) string {
	id := m.identitySection()
	role := shortRole(rowByLabel(id, "role"))
	status := rowByLabel(id, "status")

	textW := cardTextWidth(outerW)
	styleW := cardStyleDim(outerW)

	roleLine := m.styles.TitleSub.Render("role ") + m.styles.Detail.Render(role)
	dot := m.renderer.NewStyle().Foreground(colGreen).Render("● ")
	statusLine := m.styles.TitleSub.Render("status ") + dot +
		m.renderer.NewStyle().Foreground(colGreen).Render(status)
	buildsLineRendered := m.styles.TitleSub.Render("builds ") + m.styles.Detail.Render(buildsLine)

	lines := []string{
		truncateToWidth(roleLine, textW),
		truncateToWidth(statusLine, textW),
		truncateToWidth(buildsLineRendered, textW),
	}

	return m.styles.MetaCard.Width(styleW).Render(strings.Join(lines, "\n"))
}

// renderHeaderRow lays out the name card and the meta card beside it when
// there is room, or stacked full-width when there is not. Below
// wideThreshold — the same column count at which the list and detail panes
// themselves give up sitting side by side — there simply is not enough
// horizontal room for two bordered, multi-line cards, so the header
// collapses to the single compact identity line the TUI always had. Below
// tallThreshold — plenty of columns, but few rows — the two cards would
// still fit sideways but cost too many of the visitor's scarce rows when
// stacked, so they merge into the single bordered card
// renderCompactCardHeader renders, which keeps every line of content but
// drops the second border. Both decisions are derived from the model's
// current width and height, never a constant tied to a specific terminal
// size.
func (m Model) renderHeaderRow(width, height int) string {
	id := m.identitySection()
	if width < wideThreshold {
		return m.renderCompactHeader(width)
	}
	if height < tallThreshold {
		return m.renderCompactCardHeader(width)
	}
	title := strings.ToUpper(id.Title)

	nameContentW := lipgloss.Width(letterSpace(title))
	if w := lipgloss.Width(id.Kicker); w > nameContentW {
		nameContentW = w
	}
	nameOuterW := nameContentW + cardPaddingOverhead + paneBorderOverhead

	role := shortRole(rowByLabel(id, "role"))
	status := rowByLabel(id, "status")
	metaContentW := lipgloss.Width("role " + role)
	if w := lipgloss.Width("status ● " + status); w > metaContentW {
		metaContentW = w
	}
	if w := lipgloss.Width("builds " + buildsLine); w > metaContentW {
		metaContentW = w
	}
	metaOuterW := metaContentW + cardPaddingOverhead + paneBorderOverhead

	const gap = 1
	sideBySide := width >= nameOuterW+gap+metaOuterW

	if sideBySide {
		if nameOuterW > width {
			nameOuterW = width
		}
		metaOuterW = width - nameOuterW - gap
		name := m.renderNameCard(nameOuterW)
		meta := m.renderMetaCard(metaOuterW)
		return lipgloss.JoinHorizontal(lipgloss.Top, name, strings.Repeat(" ", gap), meta)
	}

	name := m.renderNameCard(width)
	meta := m.renderMetaCard(width)
	return lipgloss.JoinVertical(lipgloss.Left, name, meta)
}

// identityLine renders the title/kicker identity line shared by
// renderCompactHeader and renderCompactCardHeader: the title in Title style,
// then, if present, the kicker in TitleSub style after a " · " separator.
// Unstyled truncation is left to the caller, since the two callers truncate
// to different widths (the bare terminal width vs. a card's inset text
// width).
func (m Model) identityLine() string {
	id := m.identitySection()
	line := m.styles.Title.Render(strings.ToUpper(id.Title))
	if id.Kicker != "" {
		line += m.styles.TitleSub.Render(" · " + id.Kicker)
	}
	return line
}

// renderCompactHeader is the header row shown below wideThreshold: the
// identity line the TUI always had, no bordered cards. Both strings are
// content.json's Title and Kicker for the readme section. This is the one
// tier that drops content rather than just chrome: the role/status/builds
// lines the other tiers carry are absent here, because below wideThreshold
// there is no longer room to lay them out legibly without truncating
// mid-word — see the report for the exact width math.
func (m Model) renderCompactHeader(width int) string {
	return truncateToWidth(m.identityLine(), width)
}

// renderCompactCardHeader is the header row shown at wideThreshold or wider
// but below tallThreshold: every line the full name/meta cards carry —
// title, kicker, role, status, builds — merged into a single bordered card
// instead of two stacked ones. It drops one card's border (four rows: two
// border lines plus the blank line JoinVertical would otherwise need to
// separate the cards) while keeping every string content.json (or
// buildsLine) provides, unlike renderCompactHeader's width-driven fallback
// below, which does drop content.
func (m Model) renderCompactCardHeader(width int) string {
	id := m.identitySection()
	role := shortRole(rowByLabel(id, "role"))
	status := rowByLabel(id, "status")

	textW := cardTextWidth(width)
	styleW := cardStyleDim(width)

	roleLine := m.styles.TitleSub.Render("role ") + m.styles.Detail.Render(role)
	dot := m.renderer.NewStyle().Foreground(colGreen).Render("● ")
	statusLine := m.styles.TitleSub.Render("status ") + dot +
		m.renderer.NewStyle().Foreground(colGreen).Render(status)
	buildsLineRendered := m.styles.TitleSub.Render("builds ") + m.styles.Detail.Render(buildsLine)

	lines := []string{
		truncateToWidth(m.identityLine(), textW),
		truncateToWidth(roleLine, textW),
		truncateToWidth(statusLine, textW),
		truncateToWidth(buildsLineRendered, textW),
	}

	return m.styles.NameCard.Width(styleW).Render(strings.Join(lines, "\n"))
}

// renderTabBar renders one bordered chip per entry in content.Tabs, marking
// the active one with a leading bar and a distinct border color, and never
// grows past width. Below wideThreshold it falls back to plain unbordered
// labels, for the same reason renderHeaderRow falls back to a single line
// there — see its comment.
func (m Model) renderTabBar(width int) string {
	if width < wideThreshold {
		return m.renderCompactTabBar(width)
	}
	// The active chip's "▌ " marker is a prefix on top of the label, not a
	// substitute for it, so it does not need compensating leading spaces on
	// the inactive chips — those get plain, symmetrically padded labels
	// (Padding(0, 1) on TabInactive handles both sides equally). Only the
	// active chip is asymmetric, by design: it alone carries the bar.
	chips := make([]string, 0, len(m.content.Tabs))
	for i, t := range m.content.Tabs {
		if i == m.tabIdx {
			text := m.styles.TabActive.Render("▌ " + t.Label)
			chips = append(chips, m.styles.ChipActiveBorder.Render(text))
			continue
		}
		text := m.styles.TabInactive.Render(t.Label)
		chips = append(chips, m.styles.ChipInactiveBorder.Render(text))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, chips...)
	return truncateBlockToWidth(row, width)
}

// renderCompactTabBar renders one plain-text tab per entry in
// content.Tabs, marking the active one with a leading bar, and never grows
// past width.
func (m Model) renderCompactTabBar(width int) string {
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

// creditLine is the right-aligned "bubbletea · lipgloss" credit shown
// opposite the footer's key bindings, mirroring HelpFooter.tsx's
// margin-left: auto span. It is decorative chrome, not content.json copy.
const creditLine = "bubbletea · lipgloss"

// withCredit right-pads left with the credit line so it lands flush right
// within width. If there is no room for both, the credit is dropped rather
// than truncating the (more important) key bindings.
func (m Model) withCredit(left string, width int) string {
	credit := m.styles.Help.Render(creditLine)
	leftW := lipgloss.Width(left)
	creditW := lipgloss.Width(credit)
	gap := width - leftW - creditW
	if gap < 1 {
		return left
	}
	return left + strings.Repeat(" ", gap) + credit
}

// renderHelp is the key-binding footer, generated from the same KeyMap that
// handles input so it cannot drift from actual behavior. While the filter
// input is open it is replaced by renderFilterBar — see View.
func (m Model) renderHelp(width int) string {
	left := m.styles.Help.Render(m.help.View(m.keys))
	return truncateToWidth(m.withCredit(left, width), width)
}

// renderFilterBar is the footer shown while the filter input is open,
// mirroring the design reference's "filter> " prompt and live query.
func (m Model) renderFilterBar(width int) string {
	prompt := m.renderer.NewStyle().Foreground(colGreen).Render("filter> ")
	query := m.styles.Help.Render(m.filter)
	left := prompt + query
	return truncateToWidth(m.withCredit(left, width), width)
}

// renderRevealedBar is the footer shown after enter reveals a link: the URL
// that was written to the clipboard via OSC 52. Since OSC 52 support is not
// universal, this is what always works — the copy is a bonus, not the only
// way the visitor sees the address.
func (m Model) renderRevealedBar(width int) string {
	label := m.renderer.NewStyle().Foreground(colGreen).Render("copied ")
	url := m.styles.Help.Render(m.revealed)
	left := label + url
	return truncateToWidth(m.withCredit(left, width), width)
}

// renderFooter is the frame's bottom line: the filter input while it is
// open, the just-revealed URL after enter, the key-binding help otherwise —
// each with the "bubbletea · lipgloss" credit right-aligned opposite it.
// All three are always exactly one line, so which one is showing does not
// change computeLayout's fixed height.
func (m Model) renderFooter(width int) string {
	if m.filtering {
		return m.renderFilterBar(width)
	}
	if m.revealed != "" {
		return m.renderRevealedBar(width)
	}
	return m.renderHelp(width)
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

// renderListPane renders the section list: a README-style header label
// naming the active tab, one line per visible section, marking the current
// selection with a filled background, scrolled so the selection is always
// in the visible window, plus one status line (ListStatus) pinned to the
// bottom of the pane — mirroring the web build's ListPane, which shows the
// tab name above the rows and keeps the selected button in its scroll
// viewport.
func (m Model) renderListPane(outerW, outerH int, focused bool) string {
	style := m.styles.PaneUnfocused
	if focused {
		style = m.styles.PaneFocused
	}
	styleW := paneStyleDim(outerW)
	styleH := paneStyleDim(outerH)
	textW := paneTextWidth(outerW)

	sections := m.visibleSections()

	header := truncateToWidth(m.styles.ListHeader.Render(strings.ToUpper(m.currentTab())), textW)
	rule := truncateToWidth(m.styles.Rule.Render(strings.Repeat("─", textW)), textW)

	rowsH := styleH - 3 // header label, its rule, and the status line below
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

	lines := []string{header, rule}
	for _, s := range sections[start:end] {
		if s.ID == m.selected {
			line := truncateToWidth("› "+s.Label, textW)
			lines = append(lines, m.styles.SelectedRow.Width(textW).Render(line))
			continue
		}
		line := truncateToWidth("  "+s.Label, textW)
		lines = append(lines, m.styles.UnselectedRow.Render(line))
	}

	// Pad with blank rows so the status line lands pinned to the pane's
	// bottom edge (mirroring ListPane.tsx's flex layout: header flex-none,
	// rows flex-1, status flex-none) instead of sitting directly under
	// whatever row happened to be last.
	for len(lines) < rowsH+2 {
		lines = append(lines, strings.Repeat(" ", textW))
	}
	lines = append(lines, truncateToWidth(m.styles.Help.Render(m.Status()), textW))

	if len(lines) > styleH && styleH >= 0 {
		lines = lines[:styleH]
	}

	return style.Width(styleW).Height(styleH).Render(strings.Join(lines, "\n"))
}

// renderDetailPane renders the starship-style prompt header — on its own
// background-filled bar — and the selected section's body inside the
// bubbles viewport, which owns scrolling from Task 5 on. While the egg is
// showing, it replaces the viewport's content entirely; any key dismisses
// it and returns to the section that was on screen.
func (m Model) renderDetailPane(outerW, outerH int, focused bool) string {
	style := m.styles.PaneUnfocused
	if focused {
		style = m.styles.PaneFocused
	}
	styleW := paneStyleDim(outerW)
	styleH := paneStyleDim(outerH)
	textW := paneTextWidth(outerW)

	prompt := renderPrompt(m.renderer, m.selectedSection())
	promptBar := m.styles.PromptBar.Width(textW).Render(truncateToWidth(prompt, textW))

	body := m.viewport.View()
	if m.egg {
		body = renderEgg(m.renderer, m.styles)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, promptBar, body)
	return style.Width(styleW).Height(styleH).Render(content)
}

// renderEgg renders the undocumented easter egg's line: "Catppuccin" in the
// accent colour, italic, the rest in subtext0.
func renderEgg(r *lipgloss.Renderer, styles Styles) string {
	accent := r.NewStyle().Foreground(colAccent).Italic(true).Render("Catppuccin")
	rest := styles.TitleSub.Render(" enjoyer")
	lead := styles.TitleSub.Render("Yes, I am a ")
	return lead + accent + rest
}

// renderMetaRow renders one label/value row of the metadata table: a
// lavender label column of fixed width beside a wrapped subtext1 value
// column, mirroring DetailPane.tsx's 17ch/1fr grid.
func renderMetaRow(r *lipgloss.Renderer, styles Styles, labelW, gapW, bodyW int, row site.Row) string {
	label := styles.RowLabel.Width(labelW).Render(row.Label)
	spacer := strings.Repeat(" ", gapW)
	body := styles.RowBody.Width(bodyW).Render(row.Body)
	return lipgloss.JoinHorizontal(lipgloss.Top, label, spacer, body)
}

// renderMetaTable renders the section's Rows as a two-column ruled table:
// a top rule, then each row separated by a hairline, matching the web
// build's border-top/border-b grid rather than plain "label: value" lines.
func renderMetaTable(r *lipgloss.Renderer, styles Styles, rows []site.Row, width int) string {
	if len(rows) == 0 || width <= 0 {
		return ""
	}
	const gap = 2
	labelW := width / 3
	if labelW > 17 {
		labelW = 17
	}
	if labelW < 4 {
		labelW = 4
	}
	bodyW := width - labelW - gap
	if bodyW < 1 {
		bodyW = 1
		labelW = width - gap - bodyW
		if labelW < 0 {
			labelW = 0
		}
	}

	tableW := labelW + gap + bodyW
	topRule := styles.Rule.Render(strings.Repeat("─", tableW))
	hairline := styles.TableRule.Render(strings.Repeat("─", tableW))

	var b strings.Builder
	b.WriteString(topRule)
	for _, row := range rows {
		b.WriteString("\n")
		b.WriteString(renderMetaRow(r, styles, labelW, gap, bodyW, row))
		b.WriteString("\n")
		b.WriteString(hairline)
	}
	return b.String()
}

// renderDetailBody builds the detail viewport's content: a large heading
// (the section title, matching DetailPane.tsx's <h2>), the kicker,
// paragraphs word-wrapped to width, then the metadata table and links.
func renderDetailBody(r *lipgloss.Renderer, styles Styles, s site.Section, width int) string {
	if width <= 0 {
		return ""
	}
	wrap := r.NewStyle().Width(width)

	var b strings.Builder
	if s.Title != "" {
		b.WriteString(styles.Title.Render(wrap.Render(s.Title)))
		b.WriteString("\n")
	}

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
		b.WriteString(renderMetaTable(r, styles, s.Rows, width))
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

	// The web build closes the detail body with a small blinking caret
	// (DetailPane.tsx's aria-hidden accent block). Terminals here don't
	// blink it, but the filled accent block itself carries across as a
	// two-cell background fill on its own trailing line.
	b.WriteString("\n\n")
	b.WriteString(r.NewStyle().Background(colAccent).Render("  "))

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

// View composes the full frame: header row, tab bar, the list and detail
// panes (side by side or stacked, per computeLayout), and the help footer.
func (m Model) View() string {
	lay := m.computeLayout()

	w := m.width
	if w <= 0 {
		w = 80
	}
	h := m.height
	if h <= 0 {
		h = 24
	}

	header := m.renderHeaderRow(w, h)
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

	return lipgloss.JoinVertical(lipgloss.Left, header, tabs, main, helpFooter)
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

// truncateBlockToWidth applies truncateToWidth to every line of a
// (possibly multi-line, possibly bordered) block, so a block that spans
// several rows — like the bordered tab chips — never widens the frame past
// width even though it isn't a single line.
func truncateBlockToWidth(s string, width int) string {
	lines := splitLines(s)
	for i, line := range lines {
		lines[i] = truncateToWidth(line, width)
	}
	return strings.Join(lines, "\n")
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

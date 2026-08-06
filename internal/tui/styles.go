package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Macchiato palette, matching the web build.
const (
	colBase      = lipgloss.Color("#24273a")
	colMantle    = lipgloss.Color("#1e2030")
	colSurface0  = lipgloss.Color("#363a4f")
	colSurface1  = lipgloss.Color("#494d64")
	colRule      = lipgloss.Color("#2f3348")
	colText      = lipgloss.Color("#cad3f5")
	colSubtext1  = lipgloss.Color("#b8c0e0")
	colSubtext0  = lipgloss.Color("#a5adcb")
	colLavender  = lipgloss.Color("#b7bdf8")
	colGreen     = lipgloss.Color("#a6da95")
	colAccent    = lipgloss.Color("#f5a97f")
	colBorderAcc = lipgloss.Color("#8aadf4")
)

// Styles holds every lipgloss style used by the TUI, built once from the
// Catppuccin Macchiato palette above.
type Styles struct {
	Title    lipgloss.Style
	TitleSub lipgloss.Style

	// NameCard and MetaCard are the two bordered boxes in the header row:
	// the "DAVID FETZER" identity card and the role/status/builds card
	// beside it. Border-only + padding, applied around already-styled
	// inner content, so they never touch the color-tested Title style.
	NameCard lipgloss.Style
	MetaCard lipgloss.Style

	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// ChipActiveBorder and ChipInactiveBorder wrap the already-rendered
	// TabActive/TabInactive text in a bordered box (border-only, no
	// foreground/background of their own) so the tested TabActive color
	// combination keeps rendering unchanged underneath the border.
	ChipActiveBorder   lipgloss.Style
	ChipInactiveBorder lipgloss.Style

	PaneFocused   lipgloss.Style
	PaneUnfocused lipgloss.Style

	// ListHeader is the "README"-style label above the list pane's rows.
	ListHeader lipgloss.Style
	// Rule draws a faint horizontal separator in the pane border color.
	Rule lipgloss.Style
	// TableRule draws the hairline between metadata table rows.
	TableRule lipgloss.Style

	List          lipgloss.Style
	SelectedRow   lipgloss.Style
	UnselectedRow lipgloss.Style
	Detail        lipgloss.Style

	// RowLabel and RowBody style the two-column metadata table's columns.
	RowLabel lipgloss.Style
	RowBody  lipgloss.Style

	// PromptBar gives the detail pane's starship-prompt header row its own
	// background fill, distinct from the pane body behind it.
	PromptBar lipgloss.Style

	Help lipgloss.Style
}

// NewStyles builds the Styles used across the TUI, from styles created by r
// rather than the package-level lipgloss.NewStyle(). The default renderer
// that lipgloss.NewStyle() implies detects color support from this process's
// own stdout; under systemd that is a journald socket, not a TTY, so it
// would strip color for every visitor regardless of what their terminal
// supports. r must instead be a renderer bound to the connecting session
// (see wishbubbletea.MakeRenderer), so color detection reflects the client,
// not the server.
//
// It deliberately does not set a background colour on the root view: this
// program runs inside the visitor's own terminal, which paints its own
// background, unlike the web build's simulated terminal window.
func NewStyles(r *lipgloss.Renderer) Styles {
	pane := r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	return Styles{
		Title: r.NewStyle().
			Foreground(colAccent).
			Bold(true),

		TitleSub: r.NewStyle().
			Foreground(colSubtext0),

		NameCard: r.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorderAcc).
			Padding(0, 2),

		MetaCard: r.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colSurface0).
			Padding(0, 2),

		TabActive: r.NewStyle().
			Foreground(colBase).
			Background(colLavender).
			Bold(true).
			Padding(0, 1),

		TabInactive: r.NewStyle().
			Foreground(colSubtext0).
			Padding(0, 1),

		ChipActiveBorder: r.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorderAcc),

		ChipInactiveBorder: r.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colSurface0),

		PaneFocused: pane.
			BorderForeground(colBorderAcc),

		PaneUnfocused: pane.
			BorderForeground(colSurface0),

		ListHeader: r.NewStyle().
			Foreground(colSubtext0),

		Rule: r.NewStyle().
			Foreground(colSurface0),

		TableRule: r.NewStyle().
			Foreground(colRule),

		List: r.NewStyle().
			Foreground(colText),

		SelectedRow: r.NewStyle().
			Foreground(colAccent).
			Background(colSurface1),

		UnselectedRow: r.NewStyle().
			Foreground(colSubtext1),

		Detail: r.NewStyle().
			Foreground(colText),

		RowLabel: r.NewStyle().
			Foreground(colLavender),

		RowBody: r.NewStyle().
			Foreground(colSubtext1),

		PromptBar: r.NewStyle().
			Background(colMantle),

		Help: r.NewStyle().
			Foreground(colSubtext1),
	}
}

// helpKeyColor, helpDescColor, and helpSepColor mirror the values
// charmbracelet/bubbles' help.New() hardcodes for its default Styles (see
// bubbles/help.New). They are reproduced here, unchanged, purely so
// newHelpStyles can rebuild the same look through the session renderer —
// this is not a palette change.
var (
	helpKeyColor  = lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"}
	helpDescColor = lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}
	helpSepColor  = lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
)

// newHelpStyles rebuilds help.Model's default Styles from r instead of the
// package-level lipgloss default renderer. help.New() builds its Styles
// with bare lipgloss.NewStyle() internally and offers no way to inject a
// renderer, so left alone its key/description/separator text is styled by
// whatever color profile this process's own stdout reports — the same bug
// class this branch fixes everywhere else: under systemd that stdout is a
// journald socket, not a TTY, so the help footer's key/description
// two-tone styling would silently collapse to flat text regardless of what
// the connecting client's terminal supports.
func newHelpStyles(r *lipgloss.Renderer) help.Styles {
	keyStyle := r.NewStyle().Foreground(helpKeyColor)
	descStyle := r.NewStyle().Foreground(helpDescColor)
	sepStyle := r.NewStyle().Foreground(helpSepColor)

	return help.Styles{
		ShortKey:       keyStyle,
		ShortDesc:      descStyle,
		ShortSeparator: sepStyle,
		Ellipsis:       sepStyle,
		FullKey:        keyStyle,
		FullDesc:       descStyle,
		FullSeparator:  sepStyle,
	}
}

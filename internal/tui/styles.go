package tui

import "github.com/charmbracelet/lipgloss"

// Catppuccin Macchiato palette, matching the web build.
const (
	colBase      = lipgloss.Color("#24273a")
	colSurface0  = lipgloss.Color("#363a4f")
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

	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	PaneFocused   lipgloss.Style
	PaneUnfocused lipgloss.Style

	List   lipgloss.Style
	Detail lipgloss.Style

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

		TabActive: r.NewStyle().
			Foreground(colBase).
			Background(colLavender).
			Bold(true).
			Padding(0, 1),

		TabInactive: r.NewStyle().
			Foreground(colSubtext0).
			Padding(0, 1),

		PaneFocused: pane.
			BorderForeground(colBorderAcc),

		PaneUnfocused: pane.
			BorderForeground(colSurface0),

		List: r.NewStyle().
			Foreground(colText),

		Detail: r.NewStyle().
			Foreground(colText),

		Help: r.NewStyle().
			Foreground(colSubtext1),
	}
}

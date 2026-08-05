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

// NewStyles builds the Styles used across the TUI. It deliberately does not
// set a background colour on the root view: this program runs inside the
// visitor's own terminal, which paints its own background, unlike the web
// build's simulated terminal window.
func NewStyles() Styles {
	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	return Styles{
		Title: lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true),

		TitleSub: lipgloss.NewStyle().
			Foreground(colSubtext0),

		TabActive: lipgloss.NewStyle().
			Foreground(colBase).
			Background(colLavender).
			Bold(true).
			Padding(0, 1),

		TabInactive: lipgloss.NewStyle().
			Foreground(colSubtext0).
			Padding(0, 1),

		PaneFocused: pane.
			BorderForeground(colBorderAcc),

		PaneUnfocused: pane.
			BorderForeground(colSurface0),

		List: lipgloss.NewStyle().
			Foreground(colText),

		Detail: lipgloss.NewStyle().
			Foreground(colText),

		Help: lipgloss.NewStyle().
			Foreground(colSubtext1),
	}
}

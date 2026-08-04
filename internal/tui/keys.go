package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	PageUp  key.Binding
	PageDn  key.Binding
	Top     key.Binding
	Bottom  key.Binding
	Left    key.Binding
	Right   key.Binding
	NextTab key.Binding
	PrevTab key.Binding
	Filter  key.Binding
	Accept  key.Binding
	Cancel  key.Binding
	Quit    key.Binding

	// Egg is deliberately absent from ShortHelp and FullHelp.
	Egg key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:      key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/↓ j/k", "select")),
		Down:    key.NewBinding(key.WithKeys("j", "down")),
		PageUp:  key.NewBinding(key.WithKeys("pgup")),
		PageDn:  key.NewBinding(key.WithKeys("pgdown", " ")),
		Top:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g/G", "first/last")),
		Bottom:  key.NewBinding(key.WithKeys("G")),
		Left:    key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/l ←/→", "pane")),
		Right:   key.NewBinding(key.WithKeys("l", "right")),
		NextTab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
		PrevTab: key.NewBinding(key.WithKeys("shift+tab")),
		Filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Accept:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show link")),
		Cancel:  key.NewBinding(key.WithKeys("esc")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Egg:     key.NewBinding(key.WithKeys("C")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Left, k.NextTab, k.Filter, k.Top, k.Accept, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

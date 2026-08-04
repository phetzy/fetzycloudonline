package tui

import "encoding/base64"

// OSC52 builds the terminal escape sequence that asks the terminal emulator
// to copy payload to the visitor's system clipboard. OSC 52 is widely but
// not universally supported, so callers must also display the payload —
// the copy is a convenience, not the only way the value reaches the
// visitor.
//
// The sequence is ESC ] 52 ; c ; <base64 payload> BEL. "c" selects the
// clipboard (as opposed to the primary selection); BEL (\x07) terminates
// it, matching what actual terminal emulators expect.
func OSC52(payload string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(payload)) + "\x07"
}

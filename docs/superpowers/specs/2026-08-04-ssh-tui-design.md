# Design: SSH TUI — the Go program

Date: 2026-08-04
Status: approved, ready for implementation planning
Branch: `ssh-tui`

## Scope

Sub-project 2a. A Go program that serves the same content as the website over SSH,
as a real Bubble Tea terminal UI, so a visitor who runs `ssh <host>` lands in the TUI.

Sub-project 2b — OpenTofu infrastructure and a GitHub Actions deploy path onto a
personal AWS EC2 instance — gets its own spec and plan after this lands. This spec
covers the program only. **Nothing here deploys anything.**

The split is deliberate: this program is fully verifiable on a laptop by running it and
connecting with a real `ssh` client, while the infrastructure has entirely different
failure modes and needs an AWS account.

## Source of truth

- `~/Downloads/tuiSite/design_handoff_tui_ssh/PROMPT.md` — Deliverable 2
- `~/Downloads/tuiSite/design_handoff_tui_ssh/SPEC.md` — the "SSH build notes" section,
  the palette, and the interaction model
- `~/Downloads/tuiSite/design_handoff_tui_ssh/TUI.dc.html` — the prototype, authoritative
  for behavior the web build already implements
- `content.json` in this repository — the content, already transcribed verbatim and
  shipped in the web front end

## Content constraint

Content is frozen. Every string comes from `content.json`. No invented metrics, user
counts, download numbers, client names, testimonials, tool versions, or project details.
The employer paragraph is not expanded. No IP addresses, hostnames, ports, or network
topology in the AI platform section. No fifth project. The veteran line is stated plainly.

## Repository layout

The Go program lives alongside the web front end. The web app stays at the repository
root exactly as it is, so the Vercel deployment is untouched.

```
content.json              shared content — unchanged
content_embed.go          package site: //go:embed content.json, types, accessors
go.mod                    module github.com/phetzy/fetzycloudonline
cmd/fetzer/main.go        Wish server: config, middleware, signals
internal/tui/
  model.go                Bubble Tea model and state
  update.go               message handling and the keymap
  view.go                 layout and rendering
  keys.go                 key.Binding set, feeding both input and the help footer
  styles.go               lipgloss styles from the Catppuccin palette
  selectors.go            visible sections, list status, prompt segments
internal/ratelimit/
  ratelimit.go            per-IP concurrent and rate limits
```

### Why the embed file sits at the repository root

`go:embed` cannot reference a path outside its own package directory, so
`internal/tui` cannot embed `../../content.json`. The alternatives were copying the file
into the Go tree at build time — two sources of truth, which is exactly what the shared
content requirement exists to prevent — or embedding it in place from a package at the
root. This design takes the second. A Go package sitting beside `package.json` reads
oddly, and it is the smaller cost.

## Dependencies

- `github.com/charmbracelet/wish` — SSH server, with `bubbletea`, `activeterm`, and
  `logging` middleware
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles` — `list`, `viewport`, `help`, `key`
- `github.com/charmbracelet/lipgloss`

Go 1.26. Standard library `log/slog` for structured logging; no logging dependency.

## Palette

Catppuccin Macchiato, the same values the web build uses, as lipgloss colors:

| Role | Hex |
| --- | --- |
| base | `#24273a` |
| mantle | `#1e2030` |
| surface0 | `#363a4f` |
| surface1 | `#494d64` |
| text | `#cad3f5` |
| subtext1 | `#b8c0e0` |
| subtext0 | `#a5adcb` |
| lavender | `#b7bdf8` |
| mauve | `#c6a0f6` |
| green | `#a6da95` |
| accent (peach) | `#f5a97f` |
| border accent (blue) | `#8aadf4` |

The SSH build renders inside the visitor's terminal, so it inherits their font and their
background. It does not paint a window frame, traffic lights, or a title bar — those
exist on the web to *depict* a terminal. Here we are in one.

Glyphs assume a Nerd Font but must degrade legibly without one.

## Layout

```
title block      DAVID FETZER · software engineer · boise, id · remote
tab bar          readme · projects · work · contact   (active: ▌ label)
main             list pane (26 cols) │ detail viewport (remainder)
help footer      generated from the key bindings
```

Below roughly 70 columns the panes stack vertically, list above detail. The layout must
work at 80×24 and must never assume more. Every size comes from `tea.WindowSizeMsg`;
nothing is hardcoded to a terminal size.

The detail pane header is the starship prompt from the web build — directory in bold
lavender, ` main` in bold mauve, an optional language module per section, then the cat
glyph in green and `❯` in the accent.

## State

Mirrors the web build exactly:

- `tab` — active tab index
- `selected` — section id
- `focus` — list or viewport
- `filtering` — whether the filter input is open
- `filter` — the filter string
- `egg` — whether the easter egg is showing

Plus the bubbles components' own state: `list.Model`, `viewport.Model`, `help.Model`.

## Interaction

| Key | Action |
| --- | --- |
| `j` / `k` / `↓` / `↑` | move selection (list focused) or scroll (viewport focused) |
| `PageDown` / `PageUp` / `space` | page the focused pane |
| `g` / `G` | first/last item, or top/bottom of the viewport |
| `h` / `l` / `←` / `→` | move focus between panes |
| `tab` / `shift+tab` | cycle tabs |
| `/` | open the filter |
| `esc` | cancel the filter, or dismiss the easter egg |
| `enter` | in the filter, keep it; otherwise reveal the selected section's first link |
| `q` / `ctrl+c` | quit |

Two behaviors carried from the web build because they are load-bearing:

1. **The filter searches every section while it is open**, not just the active tab, and
   selecting a match moves the tab with it. `bubbles/list` filters only its own items, so
   the model swaps the list's items to all nine sections while filtering and restores the
   tab's subset on cancel.
2. **Focus decides what `j`/`k` mean** — selection in the list, scrolling in the
   viewport — and the help footer's hint reflects which.

Switching tabs clears the filter and selects the tab's first section.

### Two behaviors the web design cannot specify

**`enter` on a section with links.** A server cannot open the visitor's browser. The
program writes the URL to their clipboard with an OSC 52 escape sequence and shows it in
the footer. OSC 52 is widely but not universally supported, so displaying the URL is the
fallback that always works; the copy is the convenience.

**The easter egg.** The web build reveals it behind window minimize, which has no meaning
inside a terminal. Here an undocumented key swaps the detail viewport for the same line —
`Yes, I am a Catppuccin enjoyer`, "Catppuccin" italic in the accent, the rest in
`subtext0` — and any key returns to the previous view. It does not appear in the help
footer.

## Server

`wish.NewServer` with `wish.WithAddress` and `wish.WithHostKeyPath`. Middleware order is
`logging → bubbletea → activeterm`, because Wish applies middleware in reverse.

This is an **unauthenticated listener on the public internet**, which is a different risk
posture from anything else in this repository. The design treats it accordingly:

- **No shell, ever.** Every session is handled by the Bubble Tea middleware.
  `activeterm` rejects sessions without a PTY. There is no exec path, no subsystem, and
  no port forwarding.
- **Every public key is accepted** — it is public by design — and nothing is
  authenticated against anything.
- **Idle timeout and maximum session duration**, so a connection cannot be held open
  indefinitely.
- **Per-IP limits:** a concurrent connection cap and a connection rate limit, so a single
  source cannot exhaust the process.
- **A persisted host key**, so returning visitors do not see key-changed warnings.
- **Structured logging** of connects and disconnects with the remote address, via
  `log/slog`.

Configuration comes from flags with environment variable fallbacks: listen address, host
key path, idle timeout, max session duration, per-IP concurrent limit, per-IP rate.
Defaults are sensible for local development — a high port, a host key under the working
directory.

Shutdown is graceful on SIGINT/SIGTERM: stop accepting, let live sessions finish within a
bounded window, then exit.

## Testing

**Go unit tests:**

- Content loading: nine sections in the documented order, four tabs, ids unique, every
  section's tab known, links non-empty. The same invariants the web build asserts.
- Selectors: the three-way filter behavior (no filter → active tab; filter open → all
  sections; filter string with the input closed → active tab), list status text, prompt
  segments including which sections carry a language module.
- Rate limiter: concurrent cap, rate limit, and release on disconnect.
- Keymap: the bindings that feed the help footer are the same ones that handle input, so
  the footer cannot drift from behavior.

**Golden-file view tests** of the rendered frame at 80×24 and at a wide terminal. The
80×24 degradation is the requirement most likely to break silently, and a golden file
makes a layout regression visible in a diff.

**Manual verification, which is the real gate:** run the server on a high port and
connect with a real `ssh` client, at 80×24 and at a large window, exercising every
binding, the filter, the link reveal, and the easter egg.

## Out of scope

Deployment of any kind — that is sub-project 2b: OpenTofu for AWS EC2, an S3 state
backend with native locking, SSM Session Manager for administration, a GitHub OIDC role,
and a GitHub Actions workflow that cross-compiles for arm64 and ships the binary.

Also out of scope: any change to the web front end, and any change to `content.json`.

# fetzycloudonline

David Fetzer's personal site, rendered as a Bubble Tea–style terminal program
running inside a terminal window (Catppuccin Macchiato palette). Tabs, a
list/detail pane layout, and keyboard navigation stand in for a normal web
page.

A Go SSH front end (`cmd/fetzer`) serves the same content as a Bubble Tea
program over `ssh`. See [SSH TUI](#ssh-tui) below.

**`content.json` is the shared content source.** This React front end reads
it (via `src/content.ts`), and the SSH app reads the same file. Content is
frozen: do not add projects, metrics, or details that aren't already in
`content.json`.

## Development

```bash
pnpm install
pnpm dev
```

## Commands

| Command                 | Description                                         |
| ----------------------- | --------------------------------------------------- |
| `pnpm dev`              | Vite dev server                                     |
| `pnpm build`            | Type-check, bundle, and prerender to `dist/`        |
| `pnpm preview`          | Serve the production build                          |
| `pnpm test`             | Vitest unit tests, then Playwright end-to-end tests |
| `pnpm test:unit`        | Vitest unit tests only                              |
| `pnpm test:integration` | Playwright end-to-end tests only                    |
| `pnpm lint`             | Prettier check and ESLint                           |
| `pnpm format`           | Rewrite files with Prettier                         |

## Structure

- `content.json` — shared content source (tabs, sections, rows, links)
- `src/App.tsx` — the program: tab switching, keyboard wiring, pane state
- `src/content.ts` — typed accessors over `content.json`
- `src/selectors.ts` — derived view state
- `src/hooks/` — terminal dimensions, reduced-motion detection
- `src/components/` — title bar, tab bar, list pane, detail pane, header row,
  help footer, prompt, and minimized-window note
- `prerender.ts` — injects prerendered markup into `dist/index.html` at build
  time
- `tests/` — Playwright end-to-end specs

The design comes from an external design handoff at
`~/Downloads/tuiSite/design_handoff_tui_ssh/`, which is not checked into this
repo.

## Deployment

Vercel, static build. No environment variables. `vercel.json` pins
`framework: null` and `outputDirectory: dist` — the Vercel project was
originally created in 2024 for a SvelteKit app, and its stale framework
preset would otherwise try to serve that instead of this static build.

## SSH TUI

`ssh`-ing to the host drops you into the same content as the website,
rendered as a Bubble Tea program instead of a browser page. `cmd/fetzer` is
a [Wish](https://github.com/charmbracelet/wish) SSH server; `internal/tui`
is the Bubble Tea model, view, and update loop; `internal/ratelimit` is the
per-IP concurrency and rate limiter it's built on.

### Why there's a Go package next to `package.json`

`content.json` is the single shared source for both front ends. The React
site reads it via `src/content.ts`; the Go program embeds it at compile
time with `go:embed` from the `site` package at the module root
(`content_embed.go`). `go:embed` can only reference a path inside its own
package's directory, so that package has to sit beside `content.json` —
i.e. at the repo root, next to `package.json` — rather than under
`internal/`. Copying the file into the Go tree at build time was
deliberately avoided: that would create a second source of truth, which is
exactly what sharing the file is meant to prevent.

### Running it locally

```bash
go build -o fetzer ./cmd/fetzer

# Host key: created automatically on first run at ./.ssh/fetzer_ed25519 if
# it doesn't already exist (via wish.WithHostKeyPath). Nothing to generate
# by hand.
./fetzer

# In another terminal:
ssh -p 23234 localhost
```

### Flags

Every flag falls back to an environment variable, then a hardcoded default,
in that order (read from `cmd/fetzer/main.go`'s `parseConfig`):

| Flag                 | Env var                     | Default                   | Meaning                                                          |
| --------------------- | ---------------------------- | -------------------------- | ------------------------------------------------------------------ |
| `-addr`               | `FETZER_ADDR`                | `:23234`                   | Address to listen on                                                |
| `-host-key`           | `FETZER_HOST_KEY`            | `./.ssh/fetzer_ed25519`    | Path to the SSH host key (created if it does not exist)             |
| `-idle-timeout`       | `FETZER_IDLE_TIMEOUT`        | `5m`                       | Disconnect a session after this long with no activity               |
| `-max-session`        | `FETZER_MAX_SESSION`         | `30m`                      | Hard cap on a single session's total duration, active or not        |
| `-max-conns-per-ip`   | `FETZER_MAX_CONNS_PER_IP`    | `3`                        | Maximum concurrent connections allowed from a single IP             |
| `-rate-per-min`       | `FETZER_RATE_PER_MIN`        | `20`                       | Maximum new connections allowed from a single IP per minute         |

A non-positive `-idle-timeout`, `-max-session`, `-max-conns-per-ip`, or
`-rate-per-min` — whether from a flag or an environment variable — makes
the server refuse to start (`validateConfig`, exit code 2), because each of
those values silently disables the control it names if it's allowed to be
zero or negative.

### Keybindings (`internal/tui/keys.go`)

| Keys              | Action                    |
| ------------------ | -------------------------- |
| `↑`/`↓`, `j`/`k`    | Move selection              |
| `pgup`/`pgdown`/`space` | Page up / down          |
| `g` / `G`           | Jump to first / last item   |
| `h`/`l`, `←`/`→`    | Switch pane focus            |
| `tab` / `shift+tab` | Next / previous tab         |
| `/`                 | Filter                       |
| `enter`             | Show link (reveals via OSC 52) |
| `esc`               | Clear the filter — while the filter input is open, or afterward if a stale query is still narrowing (or emptying) the list |
| `q`, `ctrl+c`       | Quit                         |

(There's also an undocumented easter egg bound to `C` — deliberately absent
from both `ShortHelp` and `FullHelp`.)

### It's an unauthenticated public listener — what protects it

Every public key is accepted; there is no login and no notion of an
authorized user, by design. What keeps that safe:

- **PTY-only sessions.** There is no shell handler, no exec handler, and no
  subsystem handler registered — a client cannot run a command, use SFTP,
  or forward a port. `activeterm` middleware additionally rejects any
  session that doesn't request a PTY. The only way a session produces
  output is through the Bubble Tea middleware chain.
- **Per-IP concurrency and rate limits** (`-max-conns-per-ip`,
  `-rate-per-min`), enforced before the SSH handshake even begins.
- **A per-connection session gate** caps each TCP connection to a single
  granted session channel, since neither the connection limiter nor the
  underlying SSH libraries cap session channels per connection on their
  own.
- **A per-connection session-channel-open cap** (4) rejects a "session"
  channel outright — before it is ever accepted, so no goroutine or `env`
  slice is created for it — once a connection has opened that many. This
  closes the gap one level below the session gate above:
  `DefaultSessionHandler` accepts the raw channel before the gate is ever
  consulted, so without this cap a connection inside the per-IP budget
  could otherwise open session channels without limit.
- **Idle and max-session timeouts** (`-idle-timeout`, `-max-session`) bound
  how long any one session can stay open, active or not.
- **A persisted host key**, generated once and reused, so the server's
  identity doesn't change (and doesn't trigger client host-key warnings)
  across restarts.
- **Panic recovery** wraps the Bubble Tea program and the middleware
  bodies, so a panic anywhere in that path is caught and logged instead of
  taking down the whole process.
- **Fail-fast config validation**: a non-positive value for any of the
  above timeouts or limits refuses to start rather than silently running
  with that control disabled.

### Known residual risks

- Panic recovery covers the Bubble Tea program and the wish middleware
  bodies, but not panics inside the underlying SSH library's own
  goroutines (outside that recovered path).
- Rejection logging is bounded per event, but not per connection. A new
  TCP connection that fails the per-IP concurrency or rate limit logs one
  "connection rejected" line (`rateLimitConnCallback`); a "session"
  channel opened past the per-connection cap logs one "session channel
  rejected" line (`limitSessionChannels`, at most once per connection, not
  once per attempt). Both are still attacker-triggerable at whatever rate
  the client can generate the underlying event — a new TCP connection each
  time, in the first case. The one path still unbounded even per
  connection: once a "session" channel is accepted (up to
  `maxSessionChannelsPerConn` of them) but the connection's session gate
  has already granted a different channel, every subsequent request on
  that channel — "shell", "pty-req", "env", and so on — re-triggers
  `sessionRequestCallback` and logs its own "session rejected" line, at
  whatever rate the client sends requests within that one channel.

### Deployment

The AWS infrastructure that puts this server on a public address — EC2, an
Elastic IP, an S3 state backend and artifact bucket, SSM Session Manager for
administrative access, a GitHub OIDC deploy role, and the GitHub Actions
workflow that builds and ships `cmd/fetzer` — lives entirely under
[`terraform/`](terraform/README.md). It is **not applied automatically**;
see that README for the bootstrap-then-apply order, the GitHub repository
secrets/variables the deploy workflow needs, and known first-apply caveats.

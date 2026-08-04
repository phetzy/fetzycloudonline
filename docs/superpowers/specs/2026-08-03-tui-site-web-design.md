# Design: TUI site — shared content + web front end

Date: 2026-08-03
Status: approved, ready for implementation planning
Branch: `tui-site`

## Scope

This is sub-project 1 of 2. It replaces the current "manual page" site with a new
design: the personal site presented as a Bubble Tea program running inside a
terminal window.

Sub-project 2 — a Go program serving the same content over SSH via Wish and Bubble
Tea — gets its own spec and plan after this one lands. The two share `content.json`,
which this sub-project creates.

The work decomposes this way because the deliverables are independent: the web front
end ships to Vercel and needs no Go, while the SSH server is a different language, a
different runtime, and a different deployment story. Each produces working, testable
software on its own.

## Source of truth

- `~/Downloads/tuiSite/design_handoff_tui_ssh/SPEC.md` — authoritative design spec
- `~/Downloads/tuiSite/design_handoff_tui_ssh/PROMPT.md` — the deliverables
- `~/Downloads/tuiSite/design_handoff_tui_ssh/TUI.dc.html` — working prototype: layout,
  styling, interaction logic, and the content
- `~/Downloads/tuiSite/design_handoff_tui_ssh/favicons/` — favicon set

`support.js` is the prototype runtime only and is not ported. The handoff folder is
not checked into the repository, matching how the previous handoff was treated.

Fidelity is high: colors, sizes, spacing, motion, and keybindings are final.

## Content constraint

All copy comes verbatim from the prototype's `SECTIONS` array. Do not invent metrics,
user counts, download numbers, client names, testimonials, tool versions, or project
details. Do not expand the employer paragraph or infer what the internal work
involves. No IP addresses, hostnames, ports, or network topology in the AI platform
section. No fifth project. The veteran line is stated plainly — no flag imagery, no
iconography, no "thank you for your service" framing. If a section runs long for its
container, cut from the end rather than padding.

## What is removed

The manual-page implementation goes entirely — the prompt is explicit that the old
implementation is removed rather than left alongside the new one:

`src/Manual.tsx`, `src/Manual.test.tsx`, `src/content.ts` and `src/content.test.ts`
(the path `src/content.ts` is reused by this design for a different module — a typed
loader over `content.json` — so it is rewritten rather than merely deleted),
`src/components/` (Section, rows, TerminalButton, StatusBar, Overlay, CopyButton and
their tests), `src/hooks/` (useManualNav, useScrollPercent, useSearch and their
tests), `tests/manual.spec.ts`, and `tests/no-js.spec.ts`.

`public/favicons/` stays — the handoff ships the same caret mark already installed.

`src/hooks/prefersReducedMotion.ts` is **kept and extended**. It already exposes
`scrollBehavior()` with the SSR guard this design needs; the motion work adds a
`prefersReducedMotion()` boolean alongside it rather than writing a second module.

## What is kept

The toolchain and deploy pipeline, all of it recently repaired and verified:

- React 19 + Vite 5 + TypeScript, Tailwind 3, Vitest + React Testing Library,
  Playwright 1.61.0 (pinned to match the cached chromium build)
- `prerender.ts` and the `hydrateRoot` client, including the `data-hydrated` marker
  that end-to-end tests wait on
- `vercel.json` pinning `framework: null` and `outputDirectory: dist`
- `pnpm-workspace.yaml` with both `packages` and `allowBuilds`, which is what makes
  the build work under Vercel's pnpm 9 and local pnpm 11 alike
- `engines.node` at 24.x

## Repository layout

The web app stays at the repository root so Vercel's configuration is untouched. Go
lands in subdirectories alongside it in sub-project 2.

```
content.json          shared content, 9 sections
index.html            head, meta, Open Graph, favicons, fonts
prerender.ts          unchanged
vercel.json           unchanged
public/
  favicons/           new Catppuccin caret set
  fonts/              vendored Symbols Nerd Font
src/
  main.tsx            hydrateRoot + data-hydrated marker
  index.css           font faces, CSS custom properties, keyframes
  content.ts          typed loader over content.json
  App.tsx             window shell, keydown wiring
  components/
    TitleBar.tsx      traffic lights, window title, cols×rows readout
    HeaderRow.tsx     title block + metadata block
    TabBar.tsx        four tabs, active reads "▌ label"
    ListPane.tsx      header, rows, n/m footer
    DetailPane.tsx    prompt header, article stack, caret
    Prompt.tsx        starship segments
    HelpFooter.tsx    key hints, or the filter input when filtering
  hooks/
    useTuiState.ts    tab, selected, focus, filtering, filter + derived lists
    useTerminalDims.ts
    useSlideMotion.ts
cmd/, internal/, go.mod   sub-project 2
```

## Palette and typography

Catppuccin Macchiato, exposed as Tailwind theme colors so components never carry hex
literals:

| Token | Hex | Use |
| --- | --- | --- |
| crust | `#181926` | desk behind the window |
| base | `#24273a` | terminal window body |
| mantle | `#1e2030` | title bar, panes |
| surface0 | `#363a4f` | borders, rules, active tab fill |
| surface1 | `#494d64` | selected row fill, scrollbar thumb, link underline |
| rule | `#2f3348` | faint row separators |
| text | `#cad3f5` | body copy |
| subtext1 | `#b8c0e0` | detail values, unselected rows |
| subtext0 | `#a5adcb` | labels, chrome, help footer — dimmest permitted for text |
| lavender | `#b7bdf8` | detail row labels, prompt directory, `::selection` |
| mauve | `#c6a0f6` | prompt git branch |
| green | `#a6da95` | status dot, filter prompt, prompt cat glyph |
| red / yellow | `#ed8796` / `#eed49f` | traffic lights |

Two accents drive the theme as CSS custom properties with literal fallbacks:
`--acc` (peach `#f5a97f`) for headings, selection text, caret, links, scrollbar hover,
and the prompt `❯`; `--acc2` (blue `#8aadf4`) for focused pane borders, the active tab
border, and the title block frame.

Every text color clears WCAG AA on its ground. `#a5adcb` is the dimmest permitted for
text; darker violets are for borders and rules only.

Font stack: `'CaskaydiaCove NF', 'Cascadia Code', 'Symbols Nerd Font', ui-monospace,
monospace`. Self-hosted, per the spec's explicit allowance and consistent with the
decision already made for the previous site: `@fontsource/cascadia-code` for the
webfont, the Symbols Nerd Font TTF vendored into `public/fonts/`, and
`CaskaydiaCove NF` declared as a `local()`-only face so visitors who have it installed
get it. No third-party font requests.

Sizes: 11.5–12.5px chrome and help, 13.5px list rows and detail values, 14px
paragraphs at line-height 1.7, `clamp(20px, 2.6vw, 30px)` detail heading,
`clamp(18px, 2.4vw, 26px)` title block.

## Layout

```
desk (crust, 100vh, no page scroll, padding clamp(8px, 2.4vw, 32px))
└─ window (base, 1px surface0, radius 10, max-width 1220px, height 100%, column flex)
   ├─ title bar (mantle, bottom border)
   │    traffic lights · "fetzer@boise: ~/site — go run ./cmd/fetzer" · cols×rows
   └─ body (padding clamp(12px, 2vw, 20px), gap 12)
      ├─ header row: title block + metadata block
      ├─ tab bar: readme · projects · work · contact
      ├─ main grid: list pane (26ch) | detail pane (1fr), gap 14, independent scroll
      └─ help footer
```

Below 700px the grid collapses to a single column with the list capped at `30vh`. The
prototype does this with a ResizeObserver because inline styles cannot carry media
queries; this build uses a Tailwind breakpoint instead. The breakpoint is
viewport-based where the prototype measured the grid element — an acceptable
difference at these sizes, and the spec explicitly directs using a breakpoint.

## Content

`content.json` holds the prototype's `SECTIONS` verbatim. Shape:

```jsonc
{
  "id": "mapwright",
  "tab": "projects",                     // readme | projects | work | contact
  "label": "mapwright",
  "path": "~/site/projects/mapwright",   // prompt directory
  "meta": "flagship",
  "title": "mapwright",
  "kicker": "self-hosted map infrastructure · mapwright.io",
  "paras": ["…"],
  "rows": [{ "label": "01 tiles", "body": "…" }],
  "links": [{ "label": "mapwright.io", "href": "https://mapwright.io" }]
}
```

Nine sections in order: `readme`; projects `mapwright`, `transfer`, `platform`, `hat`,
`oss`; work `c1`, `stack`; `contact`. Four tabs: readme, projects, work, contact.

`src/content.ts` imports the JSON, declares the `Section`, `Row`, `Link`, and `TabId`
types, and re-exports the parsed array. Vite handles JSON imports natively; Go will
`//go:embed` the same file, so neither front end owns the content.

## Interaction

| Key | Action |
| --- | --- |
| `j` / `k` / `↓` / `↑` | move selection (list focused) or scroll ~56px (viewport focused) |
| `PageDown` / `PageUp` / `space` | page the focused pane — 4 items, or 85% of viewport height |
| `g` / `G` | first/last item, or top/bottom of the viewport |
| `h` / `l` / `←` / `→` | move focus between panes |
| `tab` / `shift+tab` | cycle tabs |
| `/` | open the filter |
| `esc` | cancel filter and clear it |
| `enter` | in the filter, keep it; otherwise open the selected section's first link |

Keys are ignored while any modifier is held and while focus is in an input.

Clicking a pane focuses it; clicking a row selects it. The list auto-scrolls to keep
the selection visible.

Two behaviors that the spec's summary table understates, both taken from the
prototype's logic and both load-bearing:

1. **The filter searches every section while it is open**, not only the active tab.
   The prototype narrows to the active tab only when a filter string exists without
   the filter being open. Selecting a match from another tab switches the tab with it.
2. **Focus determines what `j`/`k` mean.** The help footer's first hint reads `select`
   when the list holds focus and `scroll` when the viewport does.

Selecting a section resets the detail viewport to the top. Switching tabs clears the
filter and selects that tab's first section.

The list footer reads `n/m`, with ` filtered` appended when a filter string is
active, or `no match` when nothing matches.

## Motion

All motion is skipped when `prefers-reduced-motion: reduce` matches. This is checked
in JavaScript via `matchMedia`, not only through a CSS media query — programmatic
scrolling and JS-driven animation ignore the CSS property, which was a real defect on
the previous site.

- Detail content slides in on selection change: 10px from below when moving down, from
  above when moving up, `170ms cubic-bezier(0.22, 1, 0.36, 1)`.
- Tab switches slide list rows in from the left, `190ms`, staggered `24ms` per row.
- Selected row background and color cross-fade over `130ms`.
- The detail caret blinks at `1.1s step-end infinite`.

## State

Local component state only. No routing, no data fetching.

- `tab` — active tab id
- `selected` — section id
- `focus` — `"list" | "viewport"`
- `filtering` — boolean, whether the filter input is open
- `filter` — the filter string
- `dims` — the `cols×rows` readout, derived from `innerWidth / 8.4` and
  `innerHeight / 19`

`dims` starts as the empty string and is measured in an effect. A computed initial
value would differ between the prerender and the client and break hydration.

Configurable props `accent` and `borderAccent` from the prototype are not carried
over — they exist for the prototype's config panel. The two accents are CSS custom
properties in `index.css`, changeable in one line.

## Rendering without JavaScript

The spec requires content to render as semantic HTML and be readable without
JavaScript. This is in tension with an interface that shows one section at a time.

The approach: prerender **all nine sections** as semantic `<article>` elements inside
the detail region. With scripting disabled they render stacked and fully readable, and
the page scrolls normally. On hydration the client adds a class to the root that hides
all but the selected article and switches the window to its fixed, non-scrolling
layout.

This keeps a single copy of every string in the DOM — no duplicated content, no
`<noscript>` block, nothing hidden from crawlers — and it means the prerendered
document is genuinely complete rather than a snapshot of one pane.

`prerender.ts` continues to run `renderToString` over the app and inject the result
into `dist/index.html`, with its guard that throws when the `#root` placeholder is not
found.

## Head and metadata

From the prototype, kept accurate because the LinkedIn preview is the first impression
for most inbound traffic:

- `<title>fetzer — TUI</title>`
- `description`, `og:title`, `og:type`, `og:description`
- `og:url` and `og:image` remain absent pending a production image; the current site's
  TODO comment carries forward

Favicons need no work. The handoff's `favicons/` directory is byte-identical to the
set already installed at `public/favicons/` (verified by checksum across all seven
files), and the existing `index.html` links and `site.webmanifest` already match its
README. The caret mark carries over unchanged; only the manifest's `name` and
`short_name` are revisited alongside the new `<title>`.

## Testing

**Vitest + React Testing Library** — logic that does not need a real viewport:

- Filter matching: matches against `label + " " + title`, case-insensitive; searches
  all sections while the filter is open; narrows to the active tab otherwise.
- `move()` clamps at both ends of the visible list and does not wrap.
- List footer text: `n/m`, `n/m filtered`, and `no match`.
- Prompt segments: directory and branch always present; the language module appears
  only for the six sections that define one, with the right glyph and color.
- Tab switching clears the filter and selects the tab's first section.
- Content integrity: nine sections in the documented order, ids unique, every `tab`
  value is one of the four known tabs, every link `href` non-empty.

**Playwright** — behavior that needs a browser:

- The full keymap, including `j`/`k` meaning different things per focus.
- `h`/`l` move focus and the focused pane's border changes.
- The two panes scroll independently while the document itself does not.
- `/` opens the filter, `esc` cancels it, `enter` keeps it.
- `enter` on a section with links opens the first one.
- With JavaScript disabled, all nine sections are present and readable.
- No horizontal overflow at 375px.

Tests that drive the keyboard wait on the `data-hydrated` marker, since the
prerendered document paints before React attaches its listeners.

## Out of scope

Routing, a theme switcher, analytics beyond Vercel's existing integration, deploying
the SSH server, and any content not present in the prototype.

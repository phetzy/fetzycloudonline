# Design: personal site refactor — SvelteKit → React, "Manual" direction

Date: 2026-08-03
Status: approved, ready for implementation planning

## Goal

Replace the current SvelteKit "under construction" placeholder with the finished
personal site described in `design_handoff_manual_site/` — a single static page styled
as a Unix man page rendered in a terminal pager. Audience is recruiters, hiring
managers, and engineers arriving from LinkedIn or GitHub.

This is a complete refactor: SvelteKit is removed, React + Vite takes its place in the
same repository, preserving git history.

## Source of truth

- `~/Downloads/personalSite/design_handoff_manual_site/README.md` — design handoff
- `~/Downloads/personalSite/design_handoff_manual_site/Manual.dc.html` — design
  reference (markup, styling, interaction logic)
- `support.js` in that bundle is prototype runtime only. Not ported.

Fidelity is high: colors, type sizes, spacing, and interactions are final and
reproduced exactly.

## Content constraint

All copy comes verbatim from `Manual.dc.html`. Nothing is invented — no metrics, user
counts, client names, testimonials, or project details. The employer paragraph is not
expanded. No IP addresses, hostnames, ports, or network topology added to the AI
platform section. No fifth project. If a section overflows its container, cut from the
end rather than pad.

## Stack decisions

| Decision | Choice |
| --- | --- |
| Framework | React 19 + Vite + TypeScript |
| Styling | Tailwind 3 (already in repo) + CSS custom properties |
| Fonts | IBM Plex Mono via Google Fonts CDN (`<link>` in `index.html`) |
| Deploy | Vercel, static Vite build (auto-detected, no `vercel.json`) |
| Analytics | `@vercel/analytics/react` retained |
| Tests | Vitest + React Testing Library (unit), Playwright (e2e) |

## Repository changes

### Removed

`svelte.config.js`, `src/app.html`, `src/routes/`, `src/index.test.ts`,
`static/construction.svg`, `static/web/`.

Dependencies: `@sveltejs/kit`, `@sveltejs/adapter-auto`,
`@sveltejs/vite-plugin-svelte`, `svelte`, `svelte-check`, `flowbite`,
`flowbite-svelte`, `flowbite-svelte-icons`, `eslint-plugin-svelte`,
`prettier-plugin-svelte`.

### Added

Dependencies: `react`, `react-dom`. Dev: `@vitejs/plugin-react`, `@types/react`,
`@types/react-dom`, `eslint-plugin-react-hooks`, `@testing-library/react`,
`@testing-library/user-event`, `jsdom`.

`index.html` moves to the repository root (Vite convention) and carries all head
metadata.

### Structure

```
index.html
src/
  main.tsx              mount + <Analytics /> from @vercel/analytics/react
  index.css             tailwind directives, CSS vars, keyframes, selection/link rules
  Manual.tsx            page shell: header, sections, footer, status bar, overlay
  content.ts            SECTIONS, HELP, subsystems, transferDetails,
                        platformDetails, stack
  components/
    Section.tsx         <h2> heading + indented body wrapper
    rows.tsx            LabelRow, DetailRow, SubsystemRow
    StatusBar.tsx
    Overlay.tsx
    CopyButton.tsx
  hooks/
    useManualNav.ts     idx, goTo, IntersectionObserver + 900ms lock
    useScrollPercent.ts
    useSearch.ts        matches, matchIdx, highlight/clear, showMatch
tests/                  Playwright specs (existing directory)
```

Prose paragraphs live in JSX — they are structure, not data. Repeating row data
(subsystems, detail rows, stack rows, TOC, help keys) lives in `content.ts`.

### Dropped from the reference

The `phosphor` and `scanlines` props are not carried over. They exist for the
prototype's config panel, which the production site does not have. The accent is the
CSS custom property `--ph: #E0845C` declared once in `index.css`; changing the accent
is a one-line edit. The scanline overlay element is omitted entirely (it defaults to
off and has no toggle).

## Styling layer

`src/index.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root { --ph: #E0845C; }
body { margin: 0; background: #0D0D0E; }
html { scroll-behavior: smooth; }
@media (prefers-reduced-motion: reduce) { html { scroll-behavior: auto; } }
@keyframes blink { 0%, 49% { opacity: 1 } 50%, 100% { opacity: 0 } }
::selection { background: var(--ph); color: #0D0D0E; }
```

`tailwind.config.ts` exposes the palette as named colors so JSX reads semantically
rather than as hex literals:

| Tailwind name | Value | Use |
| --- | --- | --- |
| `ground` | `#0D0D0E` | page background, inverted text on highlight |
| `ph` | `var(--ph)` | headings, caret, links, status highlights |
| `bright` | `#F2EEE4` | h1, subsystem titles, key names |
| `body` | `#D5D0C4` | primary paragraphs |
| `muted` | `#A9A498` | detail bodies, values |
| `faint` | `#8E8A80` | subsystem descriptions |
| `dim` | `#949086` | labels, gutter numbers, footer |
| `chrome` | `#7E7A70` | manual header, status bar |
| `sky` | `#8FB8DE` | detail row labels, synopsis values |
| `rule` | `#2A2A28` | header/footer/section rules, buttons, overlay border |
| `rule-faint` | `#1E1E1C` | detail row separators |
| `surface` | `#17171A` | status bar |
| `panel` | `#101012` | overlay background |

`fontFamily.mono: ["IBM Plex Mono", "monospace"]`, applied at the root.

Off-scale values stay as Tailwind arbitrary values so the handoff numbers survive
literally: `text-[14.5px]`, `text-[clamp(22px,3.4vw,36px)]`, `max-w-[96ch]`,
`px-[clamp(14px,4vw,48px)]`, `pl-[clamp(16px,4vw,44px)]`, `tracking-[0.16em]`,
`pb-[34px]`, `scroll-mt-[24px]`.

Hover states become real Tailwind variants (`hover:border-ph hover:text-ph`) rather
than the prototype's `style-hover` attribute. No border radius, no shadows, no
gradients, no cards anywhere.

## Behavior

Three hooks, each owning one concern. They share nothing except `goTo`, which
`useSearch` does not use.

### `useScrollPercent()`

Returns `"37%"`-style label, or `"END"` at ≥99%. Passive `scroll` listener on window;
`round(scrollY / (scrollHeight - innerHeight) * 100)`. Guards divide-by-zero when the
page is shorter than the viewport (returns 100).

### `useManualNav()`

Returns `{ idx, goTo }`.

- `IntersectionObserver` with `rootMargin: "-24px 0px -70% 0px"` observes all 11
  sections; the topmost intersecting section wins.
- Observer updates are ignored while `Date.now() < lockUntil`. `lockUntil` is a
  `useRef`, not state — it must not trigger a rerender.
- `goTo(i)` clamps to `[0, 10]`, scrolls to `elementTop - 24` with
  `behavior: "smooth"`, sets `lockUntil = Date.now() + 900`, sets `idx`, and closes
  any open overlay.

The 900 ms lock exists so the status bar readout matches the key that was just pressed
instead of being overwritten mid-scroll by the observer.

### `useSearch(contentRef)`

Returns `{ query, setQuery, matches, matchIdx, showMatch, clear, label }`.

- Matching runs only for queries longer than one character after trimming.
- Candidates: `p, h1, h2, h3, span, a` inside `contentRef` that contain no matching
  descendant (leaf elements only). Case-insensitive substring match.
- The active match is highlighted by mutating that DOM node's style directly —
  `background: var(--ph)`, `color: #0D0D0E` — matching the reference. React does not
  own that text, so direct mutation is correct here rather than a rendering concern.
- Only one match is highlighted at a time. Highlights clear on a new query, on Esc,
  and on unmount.
- `showMatch(i)` wraps modulo `matches.length` and scrolls the match into view with a
  120 px offset.
- `label` is `"n/m"` when there are matches, `"no match"` when the query is >1 char
  with none, and `"enter ↵ next"` otherwise.

### Keyboard

The `keydown` listener lives in `Manual.tsx` because it needs all three hooks. It
bails when any of meta/ctrl/alt is held, and when the event target is an `input` or
`textarea`.

| Key | Action |
| --- | --- |
| `j` / `k` | next / previous section |
| `g` / `G` | first / last section |
| `q` | jump to SEE ALSO (last section) |
| `/` | open search, focus the input |
| `n` / `N` | next / previous match |
| `Enter` / `Shift+Enter` in search | next / previous match |
| `t` | toggle table of contents |
| `?` | toggle key help |
| `Esc` | close overlay, close search, clear highlight |

### Overlays

One state value: `null | "help" | "toc"`. Full-viewport scrim `rgba(8,8,9,0.86)` at
`z-index: 60`; panel `1px solid #2A2A28` on `#101012`, `max-width: 60ch`, padding
`22px 24px`. Clicking the scrim closes. TOC lists the 11 sections numbered `01`–`11`.

The overlay gets `role="dialog"` and `aria-modal="true"` — an addition over the
reference, no visual change.

### Copy button

`navigator.clipboard.writeText("david.j.fetzer@gmail.com")`, label flips to `copied`
for 1600 ms. The timeout is cleared on unmount and on repeat clicks. Failure is
swallowed silently (as in the reference) — the address is visible as a `mailto:` link
regardless.

## Sections

Eleven sections in order, ids matching the reference: `name`, `synopsis`,
`description`, `mapwright`, `transfer`, `platform`, `hardware`, `oss`, `history`,
`stack`, `contact`.

Every section follows the same pattern: uppercase `<h2>` at `14.5px / 600 /
letter-spacing 0.16em` in the accent, then a body indented by
`clamp(16px, 4vw, 44px)`. Section bottom padding `34px`, `scroll-margin-top: 24px`.

Two body row patterns recur:

- **Label/value** — grid `minmax(0,16ch) minmax(0,1fr)`, gap 16px, label `dim`, value
  `muted`.
- **Detail rows** (TRANSFER-IT-CLI, SELF-HOSTED AI PLATFORM) — same grid, 8px vertical
  padding, top border `rule-faint`, label an `<h3>` in `sky` at `14.5px/500`, body
  `muted`.

MAPWRIGHT additionally uses a `4ch / minmax(0,26ch) / 1fr` grid for its eight
subsystem rows.

The HISTORY veteran line is stated plainly — no flag imagery, no military iconography,
no "thank you for your service" framing.

## Metadata

`index.html` carries, verbatim from the reference: `<title>FETZER(1) — Software
Engineer</title>`, `description`, `og:title`, `og:type`, `og:url`, `og:description`,
`twitter:card`, `twitter:title`, `twitter:description`, viewport, the
`fonts.gstatic.com` preconnect, and the IBM Plex Mono stylesheet link.

`og:image` is left as a TODO comment — no preview image asset exists, and inventing
one would violate the content constraint. `og:url` stays `https://mapwright.io` as in
the reference; this is Mapwright's domain rather than a personal one, and is flagged
for the owner to change when a personal production domain exists.

## Accessibility

- Semantic `<header>`, `<main>`, `<section>`, `<footer>`; exactly one `<h1>`.
- `aria-label` on the status bar's terse buttons (`k ↑`, `j ↓`, `?`) and on the search
  input.
- Blinking caret is `aria-hidden`.
- Overlay is `role="dialog" aria-modal="true"`, closable with Esc and scrim click.
- Palette is unchanged from the handoff, so all text clears WCAG AA (≥4.5:1) on
  `#0D0D0E`. `#949086` is the dimmest permitted for text.
- Reduced motion disables smooth scrolling.
- Without JavaScript, all content is present and readable. Only navigation aids
  (search, overlays, status bar readouts) require JS.

## Testing

### Vitest + React Testing Library (`environment: "jsdom"`)

- `useSearch`: single-character query yields no matches; 2+ characters match leaves
  only and not their parents; matching is case-insensitive; `showMatch` wraps at both
  ends; `clear` removes highlight styles.
- Keyboard routing: `j`/`k` change the active index; keys are ignored while focus is
  in an input; keys are ignored when ctrl/meta/alt is held; `Esc` closes the overlay
  and clears search; `t` and `?` toggle their overlays.
- `CopyButton`: label reads `copy`, then `copied` after click, then `copy` again after
  1600 ms (fake timers, stubbed `navigator.clipboard`).
- Content integrity: 11 sections render, and their ids match `SECTIONS` order.

### Playwright (`tests/`)

IntersectionObserver and scroll math do not work meaningfully in jsdom, so they are
covered here instead:

- Scroll to the bottom → status bar reads `END`.
- Press `j` five times → status bar section label matches the sixth section.
- Press `/`, type a query → match counter reads `n/m` and the active match is in the
  viewport.
- Press `t` → overlay lists `01` through `11`; clicking the scrim closes it.
- With JavaScript disabled → all 11 section headings are present in the DOM.

`npm test` keeps its current meaning: `test:integration && test:unit`.

## Out of scope

Routing, light/dark toggle, CMS or content backend, resume PDF, a fifth project,
analytics beyond Vercel's default, and any content not present in the reference file.

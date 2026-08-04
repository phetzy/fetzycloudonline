# fetzycloudonline

David Fetzer's personal site, rendered as a Bubble Tea–style terminal program
running inside a terminal window (Catppuccin Macchiato palette). Tabs, a
list/detail pane layout, and keyboard navigation stand in for a normal web
page.

A Go SSH front end that serves the same content over `ssh` is planned as a
separate sub-project. It has not been built yet.

**`content.json` is the shared content source.** This React front end reads
it (via `src/content.ts`), and the future SSH app will read the same file.
Content is frozen: do not add projects, metrics, or details that aren't
already in `content.json`.

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

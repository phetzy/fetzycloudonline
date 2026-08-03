# fetzycloudonline

David Fetzer's personal site — a single static page styled as a Unix man page
rendered in a terminal pager.

## Development

```bash
pnpm install
pnpm dev
```

## Commands

| Command        | Description                                         |
| -------------- | --------------------------------------------------- |
| `pnpm dev`     | Vite dev server                                     |
| `pnpm build`   | Type-check, bundle, and prerender to `dist/`        |
| `pnpm preview` | Serve the production build                          |
| `pnpm test`    | Vitest unit tests, then Playwright end-to-end tests |
| `pnpm lint`    | Prettier check and ESLint                           |
| `pnpm format`  | Rewrite files with Prettier                         |

## Structure

- `src/Manual.tsx` — the page: eleven sections and the keyboard wiring
- `src/content.ts` — content for the repeating rows
- `src/hooks/` — scroll percentage, section navigation, search
- `src/components/` — section, row, status bar, overlay, and button primitives

The design comes from `design_handoff_manual_site/`. Content is fixed: do not add
projects, metrics, or details that are not in that reference.

## Deployment

Vercel, static build. No environment variables.

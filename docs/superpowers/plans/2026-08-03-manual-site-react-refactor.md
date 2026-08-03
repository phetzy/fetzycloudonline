# Manual Site React Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the SvelteKit placeholder site with a single static React page styled as a Unix man page rendered in a terminal pager.

**Architecture:** One page component (`Manual.tsx`) composes eleven sections from small presentational components. Three hooks own the interactive behavior — scroll percentage, section navigation, and search — and the page component wires them to a global keydown handler. All content comes from a design reference; repeating rows live in a data module, prose lives in JSX.

**Tech Stack:** React 19, Vite 6, TypeScript, Tailwind 3, Vitest + React Testing Library, Playwright, Vercel.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-08-03-manual-site-react-refactor-design.md`. Read it before starting.
- Design reference: `~/Downloads/personalSite/design_handoff_manual_site/Manual.dc.html`. Every string in this plan is copied from it verbatim.
- **Content is frozen.** Do not invent metrics, user counts, client names, testimonials, or project details. Do not expand the employer paragraph. Do not add IP addresses, hostnames, ports, or network topology to the AI platform section. Do not add a fifth project. Do not add images, icons, logos, or SVG.
- Package manager is **pnpm** (`pnpm-lock.yaml` is committed). Use `pnpm`, never `npm`.
- Prettier config is authoritative: tabs, single quotes, no semicolons, `printWidth: 100`. Code in this plan is already formatted that way. Run `pnpm format` before any commit if unsure.
- No border radius, no shadows, no gradients, no cards anywhere in the styling.
- Body text is never smaller than 12px.
- The accent color is only ever referenced as the CSS custom property `--ph` (Tailwind color `ph`), never as a hardcoded `#E0845C` in a component.
- Commit after every task. Work on branch `react-refactor`.

---

## File Structure

| File | Responsibility |
| --- | --- |
| `index.html` | Document shell, all head metadata, font links |
| `src/main.tsx` | React root mount, `<Analytics />` |
| `src/index.css` | Tailwind directives, CSS custom properties, keyframes, global element rules |
| `src/Manual.tsx` | Page shell, all eleven sections, keydown wiring |
| `src/content.ts` | `SECTIONS`, `HELP`, `SUBSYSTEMS`, `TRANSFER_DETAILS`, `PLATFORM_DETAILS`, `STACK`, `EMAIL` |
| `src/components/Section.tsx` | Section heading + indented body wrapper |
| `src/components/rows.tsx` | `LabelGrid`, `LabelRow`, `DetailRow`, `SubsystemRow` |
| `src/components/TerminalButton.tsx` | Shared bordered button styling |
| `src/components/StatusBar.tsx` | Fixed bottom bar: readout, search input, buttons |
| `src/components/Overlay.tsx` | TOC / key help modal |
| `src/components/CopyButton.tsx` | Clipboard copy with transient `copied` label |
| `src/hooks/useScrollPercent.ts` | Scroll percentage label + `formatPercent` helper |
| `src/hooks/useManualNav.ts` | Active section index, `goTo`, IntersectionObserver |
| `src/hooks/useSearch.ts` | Match finding, highlighting, `findMatches` helper |
| `prerender.ts` | Build step that writes the rendered document into `dist/index.html` |
| `tests/manual.spec.ts` | Playwright end-to-end suite |
| `tests/no-js.spec.ts` | Playwright check that the page reads with scripting disabled |

---

### Task 1: Toolchain swap — SvelteKit out, React + Vite in

**Files:**
- Delete: `svelte.config.js`, `src/app.html`, `src/app.css`, `src/routes/+layout.svelte`, `src/routes/+page.svelte`, `src/lib/index.ts`, `src/index.test.ts`, `static/construction.svg`, `static/web/icons8-svelte-doodle-32.png`
- Rename: `static/` → `public/`
- Create: `index.html`, `src/main.tsx`, `src/index.css`, `src/Manual.tsx`, `src/test-setup.ts`, `src/Manual.test.tsx`
- Modify: `package.json`, `vite.config.ts`, `tsconfig.json`, `eslint.config.js`, `.prettierrc`, `playwright.config.ts`, `tests/test.ts`

**Interfaces:**
- Consumes: nothing.
- Produces: `Manual` — a named export from `src/Manual.tsx`, `function Manual(): JSX.Element`. Every later task edits this component.

- [ ] **Step 1: Remove Svelte files and dependencies**

```bash
git rm -r src/routes src/lib src/app.html src/app.css src/index.test.ts svelte.config.js
git rm static/construction.svg static/web/icons8-svelte-doodle-32.png
git mv static public
pnpm remove @sveltejs/kit @sveltejs/adapter-auto @sveltejs/vite-plugin-svelte svelte \
  svelte-check flowbite flowbite-svelte flowbite-svelte-icons eslint-plugin-svelte \
  prettier-plugin-svelte @tailwindcss/typography
```

- [ ] **Step 2: Add React dependencies**

```bash
pnpm add react react-dom
pnpm add -D @vitejs/plugin-react @types/react @types/react-dom eslint-plugin-react-hooks \
  @testing-library/react @testing-library/user-event @testing-library/jest-dom jsdom @eslint/js
```

`@eslint/js` is imported by the existing `eslint.config.js` but was only ever a transitive dependency. Make it explicit now.

- [ ] **Step 3: Rewrite `package.json` scripts**

Replace the `scripts` block (leave `dependencies` and `devDependencies` as pnpm left them):

```json
{
	"scripts": {
		"dev": "vite",
		"build": "tsc --noEmit && vite build",
		"preview": "vite preview",
		"test": "pnpm test:unit && pnpm test:integration",
		"test:unit": "vitest run",
		"test:integration": "playwright test",
		"lint": "prettier --check . && eslint .",
		"format": "prettier --write ."
	}
}
```

Also change `"name"` from `"my-app"` to `"fetzycloudonline"`. Delete the `"check"` and `"check:watch"` scripts — they were `svelte-check`.

- [ ] **Step 4: Rewrite `vite.config.ts`**

```ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
	plugins: [react()],
	test: {
		environment: 'jsdom',
		globals: true,
		setupFiles: './src/test-setup.ts',
		include: ['src/**/*.{test,spec}.{ts,tsx}']
	}
})
```

- [ ] **Step 5: Rewrite `tsconfig.json`**

It currently extends `./.svelte-kit/tsconfig.json`, which no longer exists. Replace the whole file:

```json
{
	"compilerOptions": {
		"target": "ES2022",
		"lib": ["ES2022", "DOM", "DOM.Iterable"],
		"module": "ESNext",
		"moduleResolution": "bundler",
		"jsx": "react-jsx",
		"strict": true,
		"noEmit": true,
		"skipLibCheck": true,
		"esModuleInterop": true,
		"forceConsistentCasingInFileNames": true,
		"resolveJsonModule": true,
		"isolatedModules": true,
		"types": ["vitest/globals", "@testing-library/jest-dom"]
	},
	"include": ["src", "tests", "vite.config.ts", "tailwind.config.ts", "playwright.config.ts"]
}
```

- [ ] **Step 6: Rewrite `eslint.config.js`**

```js
import js from '@eslint/js'
import ts from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'
import prettier from 'eslint-config-prettier'
import globals from 'globals'

/** @type {import('eslint').Linter.Config[]} */
export default [
	js.configs.recommended,
	...ts.configs.recommended,
	{
		files: ['**/*.{ts,tsx}'],
		plugins: { 'react-hooks': reactHooks },
		rules: reactHooks.configs.recommended.rules,
		languageOptions: {
			globals: { ...globals.browser, ...globals.node }
		}
	},
	prettier,
	{ ignores: ['dist/', 'playwright-report/', 'test-results/'] }
]
```

- [ ] **Step 7: Strip Svelte from `.prettierrc`**

Remove `"prettier-plugin-svelte"` from `plugins` and delete the entire `"overrides"` array. Result:

```json
{
	"useTabs": true,
	"singleQuote": true,
	"trailingComma": "none",
	"printWidth": 100,
	"semi": false,
	"plugins": ["prettier-plugin-tailwindcss"]
}
```

- [ ] **Step 8: Point Playwright at pnpm**

In `playwright.config.ts`, change the `webServer.command` from `'npm run build && npm run preview'` to `'pnpm build && pnpm preview'`. Leave `port: 4173` and `testDir: 'tests'` alone — Vite's preview server defaults to 4173.

- [ ] **Step 9: Create `index.html` at the repository root**

All metadata is copied verbatim from the design reference. Note the `og:image` TODO — do not invent an image.

```html
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<link rel="icon" href="/favicon.png" />
		<title>FETZER(1) — Software Engineer</title>
		<meta
			name="description"
			content="Software engineer in Boise, Idaho. I build infrastructure — map servers, command-line tools, and the hardware underneath them."
		/>
		<meta property="og:title" content="David Fetzer — Software Engineer" />
		<meta property="og:type" content="website" />
		<meta property="og:url" content="https://mapwright.io" />
		<meta
			property="og:description"
			content="I build infrastructure — map servers, command-line tools, and the hardware underneath them. Founder of Mapwright, self-hosted map infrastructure."
		/>
		<!-- TODO: add og:image once a preview image exists, and update og:url if a personal domain is registered. -->
		<meta name="twitter:card" content="summary_large_image" />
		<meta name="twitter:title" content="David Fetzer — Software Engineer" />
		<meta
			name="twitter:description"
			content="I build infrastructure — map servers, command-line tools, and the hardware underneath them."
		/>
		<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
		<link
			href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&display=swap"
			rel="stylesheet"
		/>
	</head>
	<body>
		<div id="root"></div>
		<script type="module" src="/src/main.tsx"></script>
	</body>
</html>
```

- [ ] **Step 10: Create `src/index.css`**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
	--ph: #e0845c;
}

body {
	margin: 0;
	background: #0d0d0e;
}

html {
	scroll-behavior: smooth;
}

@media (prefers-reduced-motion: reduce) {
	html {
		scroll-behavior: auto;
	}
}

::selection {
	background: var(--ph);
	color: #0d0d0e;
}
```

- [ ] **Step 11: Create `src/test-setup.ts`**

```ts
import '@testing-library/jest-dom/vitest'
```

- [ ] **Step 12: Create `src/main.tsx`**

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Analytics } from '@vercel/analytics/react'
import './index.css'
import { Manual } from './Manual'

const root = document.getElementById('root')
if (!root) throw new Error('#root not found')

createRoot(root).render(
	<StrictMode>
		<Manual />
		<Analytics />
	</StrictMode>
)
```

Note: `StrictMode` double-invokes effects in development. Every effect added in later tasks must clean up after itself so a double mount is harmless.

- [ ] **Step 13: Write the failing test**

Create `src/Manual.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import { Manual } from './Manual'

test('renders the manual page header', () => {
	render(<Manual />)
	expect(screen.getAllByText('FETZER(1)').length).toBeGreaterThan(0)
})
```

- [ ] **Step 14: Run the test to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./Manual"`.

- [ ] **Step 15: Create the minimal `src/Manual.tsx`**

```tsx
export function Manual() {
	return (
		<div>
			<header>
				<span>FETZER(1)</span>
			</header>
		</div>
	)
}
```

- [ ] **Step 16: Run the test to verify it passes**

Run: `pnpm test:unit`
Expected: PASS, 1 test.

- [ ] **Step 17: Replace the Playwright smoke test**

`tests/test.ts` asserts an `h1` that no longer exists yet. Replace its contents:

```ts
import { expect, test } from '@playwright/test'

test('page loads with the manual title', async ({ page }) => {
	await page.goto('/')
	await expect(page).toHaveTitle('FETZER(1) — Software Engineer')
})
```

- [ ] **Step 18: Verify the full toolchain**

Run: `pnpm build`
Expected: `tsc --noEmit` clean, then Vite writes `dist/`.

Run: `pnpm lint`
Expected: no errors.

Run: `pnpm test`
Expected: unit test passes, Playwright test passes.

- [ ] **Step 19: Commit**

```bash
git add -A
git commit -m "build: replace SvelteKit with React + Vite

Removes the SvelteKit app, adapter, and flowbite dependencies. Adds React,
@vitejs/plugin-react, and a jsdom Vitest environment. index.html moves to the
repository root with the full head metadata, and static/ becomes public/."
```

---

### Task 2: Design tokens and page chrome

**Files:**
- Modify: `tailwind.config.ts`, `src/Manual.tsx`, `src/Manual.test.tsx`

**Interfaces:**
- Consumes: `Manual` from Task 1.
- Produces: the Tailwind color names `ground`, `ph`, `bright`, `body`, `muted`, `faint`, `dim`, `chrome`, `sky`, `rule`, `rule-faint`, `surface`, `panel`, and `font-mono`. Every later task uses these names instead of hex literals.

- [ ] **Step 1: Rewrite `tailwind.config.ts`**

The current file is a flowbite config. Replace the whole file:

```ts
import type { Config } from 'tailwindcss'

export default {
	content: ['./index.html', './src/**/*.{ts,tsx}'],
	theme: {
		extend: {
			colors: {
				ground: '#0D0D0E',
				ph: 'var(--ph)',
				bright: '#F2EEE4',
				body: '#D5D0C4',
				muted: '#A9A498',
				faint: '#8E8A80',
				dim: '#949086',
				chrome: '#7E7A70',
				sky: '#8FB8DE',
				rule: '#2A2A28',
				'rule-faint': '#1E1E1C',
				surface: '#17171A',
				panel: '#101012'
			},
			fontFamily: {
				mono: ['"IBM Plex Mono"', 'monospace']
			},
			keyframes: {
				blink: {
					'0%, 49%': { opacity: '1' },
					'50%, 100%': { opacity: '0' }
				}
			},
			animation: {
				blink: 'blink 1.1s step-end infinite'
			}
		}
	},
	plugins: []
} as Config
```

The caret's blink lives in the Tailwind theme rather than in `index.css`, so the
component can use a plain `animate-blink` class with no arbitrary-property escape
hatch and no specificity conflict against Tailwind's `animation` reset.

- [ ] **Step 2: Write the failing test**

Add to `src/Manual.test.tsx`:

```tsx
test('renders the manual footer', () => {
	render(<Manual />)
	expect(screen.getByText('General Commands Manual')).toBeInTheDocument()
	expect(screen.getByText('Boise, Idaho')).toBeInTheDocument()
	expect(screen.getByText('2026')).toBeInTheDocument()
})
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find an element with the text: General Commands Manual`.

- [ ] **Step 4: Build the page shell**

Replace `src/Manual.tsx`:

```tsx
export function Manual() {
	return (
		<div className="relative min-h-screen bg-ground px-[clamp(14px,4vw,48px)] pb-[108px] font-mono text-[14.5px] leading-[1.65] text-body">
			<div className="mx-auto max-w-[96ch]">
				<header className="flex flex-wrap justify-between gap-4 border-b border-rule py-4 text-[12.5px] tracking-[0.08em] text-chrome">
					<span>FETZER(1)</span>
					<span>General Commands Manual</span>
					<span>FETZER(1)</span>
				</header>

				<main className="pt-[clamp(28px,5vh,52px)]">
					<footer className="flex flex-wrap justify-between gap-4 border-t border-rule pt-4 text-[12.5px] text-dim">
						<span>Boise, Idaho</span>
						<span>2026</span>
						<span>FETZER(1)</span>
					</footer>
				</main>
			</div>
		</div>
	)
}
```

The footer sits inside `<main>` — that matches the reference, and it keeps the footer inside the search root established in Task 7.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS, 2 tests.

- [ ] **Step 6: Check it visually**

Run: `pnpm dev`, open the printed URL. Expect near-black background, warm grey monospace header and footer rules, content centered at 96 characters wide. If the font falls back to a system monospace, the Google Fonts link in `index.html` is wrong.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add man page design tokens and page chrome

Replaces the flowbite Tailwind theme with the terminal palette as named
colors, and builds the manual header, footer, and centered page shell."
```

---

### Task 3: Content module and row components

**Files:**
- Create: `src/content.ts`, `src/components/Section.tsx`, `src/components/rows.tsx`, `src/content.test.ts`
- Test: `src/content.test.ts`

**Interfaces:**
- Consumes: Tailwind color names from Task 2.
- Produces:
  - `SECTIONS: SectionMeta[]` where `type SectionMeta = { id: string; label: string }` — 11 entries, order is the page order.
  - `HELP: KeyRow[]` where `type KeyRow = { key: string; label: string }` — 8 entries.
  - `SUBSYSTEMS: Subsystem[]` where `type Subsystem = { n: string; title: string; body: string }` — 8 entries.
  - `TRANSFER_DETAILS: Detail[]` and `PLATFORM_DETAILS: Detail[]` where `type Detail = { label: string; body: string }` — 6 and 5 entries.
  - `STACK: StackRow[]` where `type StackRow = { label: string; items: string }` — 6 entries.
  - `EMAIL: string`.
  - `Section({ id, heading, children })`, `LabelGrid({ variant?, children })`, `LabelRow({ label, children })`, `DetailRow({ label, body })`, `SubsystemRow({ n, title, body })`.

- [ ] **Step 1: Write the failing test**

Create `src/content.test.ts`:

```ts
import {
	HELP,
	PLATFORM_DETAILS,
	SECTIONS,
	STACK,
	SUBSYSTEMS,
	TRANSFER_DETAILS
} from './content'

test('there are eleven sections in manual order', () => {
	expect(SECTIONS.map((s) => s.id)).toEqual([
		'name',
		'synopsis',
		'description',
		'mapwright',
		'transfer',
		'platform',
		'hardware',
		'oss',
		'history',
		'stack',
		'contact'
	])
})

test('section ids are unique', () => {
	expect(new Set(SECTIONS.map((s) => s.id)).size).toBe(SECTIONS.length)
})

test('row collections have the counts the design specifies', () => {
	expect(HELP).toHaveLength(8)
	expect(SUBSYSTEMS).toHaveLength(8)
	expect(TRANSFER_DETAILS).toHaveLength(6)
	expect(PLATFORM_DETAILS).toHaveLength(5)
	expect(STACK).toHaveLength(6)
})
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./content"`.

- [ ] **Step 3: Create `src/content.ts`**

Every string below is verbatim from the design reference. Do not reword.

```ts
export type SectionMeta = { id: string; label: string }
export type KeyRow = { key: string; label: string }
export type Subsystem = { n: string; title: string; body: string }
export type Detail = { label: string; body: string }
export type StackRow = { label: string; items: string }

export const EMAIL = 'david.j.fetzer@gmail.com'

export const SECTIONS: SectionMeta[] = [
	{ id: 'name', label: 'NAME' },
	{ id: 'synopsis', label: 'SYNOPSIS' },
	{ id: 'description', label: 'DESCRIPTION' },
	{ id: 'mapwright', label: 'MAPWRIGHT' },
	{ id: 'transfer', label: 'TRANSFER-IT-CLI' },
	{ id: 'platform', label: 'SELF-HOSTED AI PLATFORM' },
	{ id: 'hardware', label: 'SECURE ELEMENT HAT' },
	{ id: 'oss', label: 'OPEN SOURCE' },
	{ id: 'history', label: 'HISTORY' },
	{ id: 'stack', label: 'ENVIRONMENT' },
	{ id: 'contact', label: 'SEE ALSO' }
]

export const HELP: KeyRow[] = [
	{ key: 'j / k', label: 'next / previous section' },
	{ key: 'g / G', label: 'jump to top / end' },
	{ key: '/', label: 'search the manual' },
	{ key: 'n / N', label: 'next / previous match' },
	{ key: 't', label: 'table of contents' },
	{ key: 'q', label: 'jump to SEE ALSO (contact)' },
	{ key: '?', label: 'this help' },
	{ key: 'esc', label: 'close overlay or search' }
]

export const SUBSYSTEMS: Subsystem[] = [
	{ n: '01', title: 'vector & raster tiles', body: 'Planet-scale tile serving from a single container.' },
	{ n: '02', title: 'style editor', body: 'Visual editor with five standard styles.' },
	{ n: '03', title: 'geocoding', body: 'Forward, reverse, and POI search.' },
	{
		n: '04',
		title: 'routing',
		body: 'Directions, matrix, isochrone, map-matching, optimization.'
	},
	{
		n: '05',
		title: 'static maps',
		body: 'Server-rendered images for reports, email, and embeds.'
	},
	{ n: '06', title: 'wmts / wms', body: 'Standards-based endpoints for existing GIS clients.' },
	{ n: '07', title: 'api keys', body: 'Scopes, origin allowlists, metering, and rate limits.' },
	{
		n: '08',
		title: 'admin console',
		body: 'Operate and observe the whole server from one place.'
	}
]

export const TRANSFER_DETAILS: Detail[] = [
	{
		label: 'protocol',
		body: "The upload path is reverse-engineered from transfer.it's WebSocket protocol."
	},
	{
		label: 'throughput',
		body: 'Multi-pool parallel connections, in-session retry with exponential backoff, and token-bucket bandwidth limiting over a shared HTTP/2 client with TLS session resumption.'
	},
	{
		label: 'memory',
		body: 'Transfers stream, so memory stays bounded by the chunk size — under a megabyte, even on 20+ GiB files.'
	},
	{
		label: 'resume',
		body: 'Uploads, downloads, and recursive folder uploads all resume after process restart, tracked by a per-transfer chunk bitmap with per-chunk MAC verification that stays forward-compatible across binary upgrades.'
	},
	{
		label: 'packaging',
		body: 'Six cross-compiled targets plus native .deb, .rpm, and .apk packages.'
	},
	{
		label: 'ci',
		body: 'staticcheck, govulncheck, and race-detector tests across Linux, macOS, and Windows, alongside a weekly check that hashes upstream bundles and opens an issue when the service changes underneath it.'
	}
]

export const PLATFORM_DETAILS: Detail[] = [
	{
		label: 'ops agent',
		body: 'A read-only operations agent answers infrastructure questions from runbooks and live cluster state.'
	},
	{
		label: 'safety',
		body: 'Every request passes a default-deny safety router: an LLM intent classifier and a regex gate run independently, and either one flagging write intent stops execution before it starts. Writes come back as a scoped review checklist — the agent proposes, a human approves.'
	},
	{
		label: 'architecture',
		body: 'Mixed-architecture throughout. Cluster workloads are arm64 images built and served from a private registry with upstream images pinned by digest, while inference runs on x86 with a passed-through consumer AMD GPU.'
	},
	{
		label: 'inference',
		body: 'Vulkan rules out flash attention and KV-cache quantization, so throughput comes from context and keep-alive tuning instead.'
	},
	{
		label: 'services',
		body: 'Python and FastAPI services, a Go backend-for-frontend, React and TypeScript console.'
	}
]

export const STACK: StackRow[] = [
	{ label: 'languages', items: 'Go, TypeScript, JavaScript, Python' },
	{
		label: 'backend',
		items: 'Node.js, Express, PostgreSQL, PostGIS, REST APIs, WebSockets'
	},
	{ label: 'frontend', items: 'React, Tailwind' },
	{ label: 'infra', items: 'AWS, Terraform, OpenTofu, Docker, Kubernetes, Linux' },
	{ label: 'hardware', items: 'KiCad, embedded Linux, I2C' },
	{ label: 'domains', items: 'geospatial, embedded systems, self-hosted infrastructure' }
]
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `pnpm test:unit`
Expected: PASS, 5 tests.

- [ ] **Step 5: Create `src/components/Section.tsx`**

```tsx
import type { ReactNode } from 'react'

type Props = {
	id: string
	heading: string
	children: ReactNode
	/** Only the last section uses a shorter bottom pad. */
	tight?: boolean
}

export function Section({ id, heading, children, tight = false }: Props) {
	return (
		<section id={id} className={`scroll-mt-[24px] ${tight ? 'pb-[20px]' : 'pb-[34px]'}`}>
			<h2 className="mb-[14px] text-[14.5px] font-semibold tracking-[0.16em] text-ph">
				{heading}
			</h2>
			<div className="pl-[clamp(16px,4vw,44px)]">{children}</div>
		</section>
	)
}
```

- [ ] **Step 6: Create `src/components/rows.tsx`**

The `variant` strings must stay as complete literal class names — Tailwind scans source text and will not see a dynamically concatenated class.

```tsx
import type { ReactNode } from 'react'
import type { Detail, Subsystem } from '../content'

export function LabelGrid({
	variant = 'wide',
	children
}: {
	variant?: 'wide' | 'narrow'
	children: ReactNode
}) {
	const cols =
		variant === 'narrow'
			? 'grid-cols-[minmax(0,14ch)_minmax(0,1fr)] gap-y-[4px]'
			: 'grid-cols-[minmax(0,16ch)_minmax(0,1fr)] gap-y-[6px]'
	return <div className={`grid gap-x-4 text-muted ${cols}`}>{children}</div>
}

export function LabelRow({ label, children }: { label: string; children: ReactNode }) {
	return (
		<>
			<span className="text-dim">{label}</span>
			<span>{children}</span>
		</>
	)
}

export function DetailRow({ label, body }: Detail) {
	return (
		<div className="grid grid-cols-[minmax(0,16ch)_minmax(0,1fr)] items-start gap-x-4 border-t border-rule-faint py-2">
			<h3 className="text-[14.5px] font-medium text-sky">{label}</h3>
			<p className="text-pretty text-muted">{body}</p>
		</div>
	)
}

export function SubsystemRow({ n, title, body }: Subsystem) {
	return (
		<div className="grid grid-cols-[4ch_minmax(0,26ch)_minmax(0,1fr)] items-baseline gap-x-3 py-[5px]">
			<span className="text-dim">{n}</span>
			<span className="text-bright">{title}</span>
			<span className="text-pretty text-faint">{body}</span>
		</div>
	)
}
```

- [ ] **Step 7: Verify types and lint**

Run: `pnpm build`
Expected: `tsc --noEmit` clean.

Run: `pnpm lint`
Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add manual content module and row components

Content strings are copied verbatim from the design reference. Section, and
the three recurring row patterns, are extracted as components."
```

---

### Task 4: The eleven sections

**Files:**
- Modify: `src/Manual.tsx`, `src/Manual.test.tsx`

**Interfaces:**
- Consumes: `Section`, `LabelGrid`, `LabelRow`, `DetailRow`, `SubsystemRow` from Task 3; all content exports from Task 3.
- Produces: eleven `<section>` elements with the ids from `SECTIONS`, and a `contentRef` on `<main>` that Task 7's search uses as its root.

- [ ] **Step 1: Write the failing test**

Add to `src/Manual.test.tsx`:

```tsx
import { SECTIONS } from './content'

test('every section in SECTIONS renders with its heading', () => {
	const { container } = render(<Manual />)
	for (const section of SECTIONS) {
		expect(container.querySelector(`#${section.id}`)).toBeInTheDocument()
		expect(screen.getByRole('heading', { level: 2, name: section.label })).toBeInTheDocument()
	}
})

test('there is exactly one h1', () => {
	render(<Manual />)
	expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1)
})

test('contact links point at the right destinations', () => {
	render(<Manual />)
	expect(screen.getByRole('link', { name: 'david.j.fetzer@gmail.com' })).toHaveAttribute(
		'href',
		'mailto:david.j.fetzer@gmail.com'
	)
	expect(screen.getByRole('link', { name: 'github.com/phetzy' })).toHaveAttribute(
		'href',
		'https://github.com/phetzy'
	)
	expect(screen.getByRole('link', { name: 'linkedin.com/in/fetzy' })).toHaveAttribute(
		'href',
		'https://linkedin.com/in/fetzy'
	)
})
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find an accessible element with the role "heading"`.

- [ ] **Step 3: Fill in the sections**

Replace `src/Manual.tsx`. Prose stays inline; repeating rows come from `content.ts`.

```tsx
import { useRef } from 'react'
import { Section } from './components/Section'
import { DetailRow, LabelGrid, LabelRow, SubsystemRow } from './components/rows'
import { EMAIL, PLATFORM_DETAILS, STACK, SUBSYSTEMS, TRANSFER_DETAILS } from './content'

export function Manual() {
	const contentRef = useRef<HTMLElement>(null)

	return (
		<div className="relative min-h-screen bg-ground px-[clamp(14px,4vw,48px)] pb-[108px] font-mono text-[14.5px] leading-[1.65] text-body">
			<div className="mx-auto max-w-[96ch]">
				<header className="flex flex-wrap justify-between gap-4 border-b border-rule py-4 text-[12.5px] tracking-[0.08em] text-chrome">
					<span>FETZER(1)</span>
					<span>General Commands Manual</span>
					<span>FETZER(1)</span>
				</header>

				<main ref={contentRef} className="pt-[clamp(28px,5vh,52px)]">
					<Section id="name" heading="NAME">
						<h1 className="mb-[10px] text-[clamp(22px,3.4vw,36px)] font-semibold leading-[1.25] tracking-[-0.01em] text-bright">
							david-fetzer
							<span className="text-chrome"> — </span>
							software engineer
							<span
								aria-hidden="true"
								className="ml-[0.3em] inline-block h-[1em] w-[0.6em] animate-blink bg-ph align-[-0.12em]"
							/>
						</h1>
						<p className="max-w-[74ch] text-pretty text-muted">
							Builds infrastructure — map servers, command-line tools, self-hosted platforms, and
							the hardware underneath them.
						</p>
					</Section>

					<Section id="synopsis" heading="SYNOPSIS">
						<p className="mb-1 text-bright">
							<span className="text-ph">david-fetzer</span> [
							<span className="text-muted">--location</span> <span className="text-sky">boise-id</span>
							] [<span className="text-muted">--remote</span>]
						</p>
						<p className="mb-1 text-bright">
							<span className="text-ph">david-fetzer</span> [
							<span className="text-muted">--role</span>{' '}
							<span className="text-sky">&quot;software engineer @ C1&quot;</span>]
						</p>
						<p className="text-bright">
							<span className="text-ph">david-fetzer</span> [
							<span className="text-muted">--status</span>{' '}
							<span className="text-sky">open-to-interesting-conversations</span>]
						</p>
					</Section>

					<Section id="description" heading="DESCRIPTION">
						<p className="mb-[14px] max-w-[76ch] text-pretty text-body">
							Software engineer in Boise, Idaho. Founder of Mapwright, self-hosted map
							infrastructure. Projects are documented below; employment is summarized under
							HISTORY.
						</p>
						<LabelGrid variant="narrow">
							<LabelRow label="--location">Boise, ID · remote</LabelRow>
							<LabelRow label="--role">Software Engineer at C1</LabelRow>
							<LabelRow label="--status">
								<span className="text-ph">open to interesting conversations</span>
							</LabelRow>
						</LabelGrid>
					</Section>

					<Section id="mapwright" heading="MAPWRIGHT">
						<p className="mb-[14px] max-w-[76ch] text-pretty text-body">
							A self-hosted map server. The full Mapbox stack — tiles, styles, geocoding, routing,
							static maps — running on your own hardware as a single Docker image with embedded
							Postgres.
						</p>
						<p className="mb-[22px] max-w-[76ch] text-pretty text-muted">
							Mapwright speaks Mapbox and MapTiler URL shapes, so existing MapLibre applications
							migrate by changing a base URL rather than rewriting against a new SDK. It runs
							air-gapped for on-premise and data-residency requirements.
						</p>
						{SUBSYSTEMS.map((s) => (
							<SubsystemRow key={s.n} {...s} />
						))}
						<div className="mt-[22px] grid grid-cols-[repeat(auto-fit,minmax(200px,1fr))] gap-x-6 gap-y-[6px] border-t border-rule pt-[14px] text-muted">
							<span>
								<span className="text-dim">tiles </span>127 GB, full planet
							</span>
							<span>
								<span className="text-dim">routing </span>93 GB, full planet
							</span>
							<span>
								<span className="text-dim">isolation </span>air-gapped capable
							</span>
						</div>
						<p className="mt-[18px]">
							<a
								href="https://mapwright.io"
								target="_blank"
								rel="noopener"
								className="text-ph no-underline hover:bg-ph hover:text-ground"
							>
								https://mapwright.io
							</a>{' '}
							<span className="text-dim">·</span>{' '}
							<a
								href="https://mapwright.io/docs"
								target="_blank"
								rel="noopener"
								className="text-ph no-underline hover:bg-ph hover:text-ground"
							>
								/docs
							</a>
						</p>
					</Section>

					<Section id="transfer" heading="TRANSFER-IT-CLI">
						<p className="mb-[22px] max-w-[76ch] text-pretty text-body">
							A single static Go binary for MEGA&apos;s transfer.it, shipping as both an
							interactive TUI and a scriptable CLI. Built for machines without a browser — servers,
							embedded boxes, SSH-only sessions. Zero runtime dependencies.
						</p>
						{TRANSFER_DETAILS.map((d) => (
							<DetailRow key={d.label} {...d} />
						))}
					</Section>

					<Section id="platform" heading="SELF-HOSTED AI PLATFORM">
						<p className="mb-[14px] max-w-[76ch] text-pretty text-body">
							Private LLM and RAG stack on Kubernetes. A multi-node K3s cluster running across
							Raspberry Pis, with GPU inference offloaded to a virtualized host elsewhere on the
							same network. Open WebUI, Ollama, a FastAPI retrieval service over a Qdrant vector
							store, Prometheus and Grafana observability, and automated backup-freshness checks.
						</p>
						{PLATFORM_DETAILS.map((d) => (
							<DetailRow key={d.label} {...d} />
						))}
					</Section>

					<Section id="hardware" heading="SECURE ELEMENT HAT">
						<p className="mb-[18px] max-w-[76ch] text-pretty text-body">
							A Raspberry Pi add-on board: a custom PCB designed in KiCad around an
							ATECC608B-TNGTLS secure element.
						</p>
						<LabelGrid>
							<LabelRow label="interface">I2C</LabelRow>
							<LabelRow label="keys">hardware-backed key storage, TLS provisioning</LabelRow>
							<LabelRow label="integration">
								device tree overlay, EEPROM configuration for Raspberry Pi
							</LabelRow>
						</LabelGrid>
					</Section>

					<Section id="oss" heading="OPEN SOURCE">
						<p className="text-body">Contributions to Jellyfin and Websurfx.</p>
					</Section>

					<Section id="history" heading="HISTORY">
						<p className="mb-[10px] text-bright">
							C1 — Software Engineer <span className="text-dim">·</span> March 2022 – present{' '}
							<span className="text-dim">·</span> Remote
						</p>
						<p className="mb-[18px] max-w-[76ch] text-pretty text-muted">
							Distributed systems and embedded software. Backend services in Go and TypeScript,
							React frontends, and AWS infrastructure managed as code with Terraform. Also builds
							internal tooling used across engineering teams.
						</p>
						<p className="border-t border-rule-faint pt-[14px] text-muted">
							US Army infantry veteran, 2012–2015.
						</p>
					</Section>

					<Section id="stack" heading="ENVIRONMENT">
						<div className="grid grid-cols-[minmax(0,16ch)_minmax(0,1fr)] gap-x-4">
							{STACK.map((g) => (
								<div key={g.label} className="contents">
									<span className="py-[5px] text-dim">{g.label}</span>
									<span className="text-pretty py-[5px] text-body">{g.items}</span>
								</div>
							))}
						</div>
					</Section>

					<Section id="contact" heading="SEE ALSO" tight>
						<div className="grid gap-y-[6px]">
							<span>
								<span className="inline-block w-[12ch] text-dim">email</span>
								<a
									href={`mailto:${EMAIL}`}
									className="text-ph no-underline hover:bg-ph hover:text-ground"
								>
									{EMAIL}
								</a>
							</span>
							<span>
								<span className="inline-block w-[12ch] text-dim">github</span>
								<a
									href="https://github.com/phetzy"
									target="_blank"
									rel="noopener"
									className="text-ph no-underline hover:bg-ph hover:text-ground"
								>
									github.com/phetzy
								</a>
							</span>
							<span>
								<span className="inline-block w-[12ch] text-dim">linkedin</span>
								<a
									href="https://linkedin.com/in/fetzy"
									target="_blank"
									rel="noopener"
									className="text-ph no-underline hover:bg-ph hover:text-ground"
								>
									linkedin.com/in/fetzy
								</a>
							</span>
						</div>
					</Section>

					<footer className="flex flex-wrap justify-between gap-4 border-t border-rule pt-4 text-[12.5px] text-dim">
						<span>Boise, Idaho</span>
						<span>2026</span>
						<span>FETZER(1)</span>
					</footer>
				</main>
			</div>
		</div>
	)
}
```

The `contentRef` is unused until Task 7. TypeScript will not complain — it is passed to `<main>`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS, 8 tests.

- [ ] **Step 5: Compare against the reference**

Serve the design reference in a second terminal:

```bash
cd ~/Downloads/personalSite/design_handoff_manual_site && npx serve .
```

Open `Manual.dc.html` from that server next to `pnpm dev`, and compare the three screenshots in `screenshots/`. Section spacing, indent depth, and the MAPWRIGHT subsystem column widths should match.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: render all eleven manual sections

Content is verbatim from the design reference. No behavior yet — the page is
a static document that reads correctly without JavaScript."
```

---

### Task 5: Scroll percentage and the status bar

**Files:**
- Create: `src/hooks/useScrollPercent.ts`, `src/hooks/useScrollPercent.test.ts`, `src/components/TerminalButton.tsx`, `src/components/StatusBar.tsx`
- Modify: `src/Manual.tsx`
- Test: `src/hooks/useScrollPercent.test.ts`

**Interfaces:**
- Consumes: `SECTIONS` from Task 3.
- Produces:
  - `formatPercent(scrollY: number, scrollHeight: number, innerHeight: number): string`
  - `useScrollPercent(): string`
  - `TerminalButton(props: ButtonHTMLAttributes<HTMLButtonElement>)`
  - `StatusBar(props: StatusBarProps)` — see the type below. Tasks 6–8 fill in the props that are inert here.

- [ ] **Step 1: Write the failing test**

Create `src/hooks/useScrollPercent.test.ts`:

```ts
import { formatPercent } from './useScrollPercent'

test('reports a rounded percentage of the scrollable distance', () => {
	expect(formatPercent(0, 2000, 1000)).toBe('0%')
	expect(formatPercent(500, 2000, 1000)).toBe('50%')
	expect(formatPercent(333, 2000, 1000)).toBe('33%')
})

test('reads END from 99 percent onward', () => {
	expect(formatPercent(990, 2000, 1000)).toBe('END')
	expect(formatPercent(1000, 2000, 1000)).toBe('END')
})

test('reports END when the page is not scrollable', () => {
	expect(formatPercent(0, 800, 1000)).toBe('END')
})
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./useScrollPercent"`.

- [ ] **Step 3: Create `src/hooks/useScrollPercent.ts`**

```ts
import { useEffect, useState } from 'react'

export function formatPercent(scrollY: number, scrollHeight: number, innerHeight: number): string {
	const distance = scrollHeight - innerHeight
	const percent = distance > 0 ? Math.round((scrollY / distance) * 100) : 100
	return percent >= 99 ? 'END' : `${percent}%`
}

export function useScrollPercent(): string {
	const [percent, setPercent] = useState('0%')

	useEffect(() => {
		const onScroll = () => {
			setPercent(
				formatPercent(window.scrollY, document.documentElement.scrollHeight, window.innerHeight)
			)
		}
		onScroll()
		window.addEventListener('scroll', onScroll, { passive: true })
		return () => window.removeEventListener('scroll', onScroll)
	}, [])

	return percent
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `pnpm test:unit`
Expected: PASS, 11 tests.

- [ ] **Step 5: Create `src/components/TerminalButton.tsx`**

```tsx
import type { ButtonHTMLAttributes } from 'react'

export function TerminalButton({
	className = '',
	...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
	return (
		<button
			type="button"
			className={`border border-rule bg-transparent px-[10px] py-[3px] font-mono text-[12px] text-muted hover:border-ph hover:text-ph ${className}`}
			{...props}
		/>
	)
}
```

- [ ] **Step 6: Create `src/components/StatusBar.tsx`**

The search half and the overlay callbacks are wired in Tasks 7 and 8; the props exist now so the component's shape does not churn.

Note the ref types: React 19's `useRef<T>(null)` produces `RefObject<T | null>`, so the
prop types must include `| null` or the call site will not type-check.

Note the `data-testid` attributes: the readout renders several values inside one line, so
tests cannot select them by text without matching the whole line. They are the only test
hooks in the codebase.

```tsx
import type { KeyboardEvent, RefObject } from 'react'
import { TerminalButton } from './TerminalButton'

export type StatusBarProps = {
	current: string
	position: string
	percent: string
	searching: boolean
	query: string
	matchLabel: string
	searchRef: RefObject<HTMLInputElement | null>
	onQueryChange: (value: string) => void
	onSearchKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void
	onPrev: () => void
	onNext: () => void
	onToc: () => void
	onFind: () => void
	onHelp: () => void
}

export function StatusBar({
	current,
	position,
	percent,
	searching,
	query,
	matchLabel,
	searchRef,
	onQueryChange,
	onSearchKeyDown,
	onPrev,
	onNext,
	onToc,
	onFind,
	onHelp
}: StatusBarProps) {
	return (
		<div className="fixed bottom-0 left-0 right-0 z-[55] border-t border-rule bg-surface px-[clamp(14px,4vw,48px)] py-2">
			<div className="mx-auto flex max-w-[96ch] flex-wrap items-center justify-between gap-4 text-[12px] text-chrome">
				{searching ? (
					<span className="flex flex-1 basis-[240px] items-center gap-2">
						<span className="text-ph">/</span>
						<input
							ref={searchRef}
							value={query}
							onChange={(event) => onQueryChange(event.target.value)}
							onKeyDown={onSearchKeyDown}
							placeholder="search the manual"
							aria-label="Search the manual"
							className="min-w-0 flex-1 border-0 bg-transparent font-mono text-[13px] text-bright outline-none placeholder:text-dim"
						/>
						<span data-testid="match-label" className="whitespace-nowrap text-dim">
							{matchLabel}
						</span>
					</span>
				) : (
					<span className="flex-1 basis-[200px]">
						<span data-testid="current-section" className="text-ph">
							{current}
						</span>{' '}
						<span className="text-dim">·</span>{' '}
						<span data-testid="position">{position}</span> <span className="text-dim">·</span>{' '}
						<span data-testid="percent">{percent}</span>
					</span>
				)}
				<span className="flex flex-wrap items-center gap-[10px]">
					<TerminalButton onClick={onPrev} aria-label="Previous section">
						k ↑
					</TerminalButton>
					<TerminalButton onClick={onNext} aria-label="Next section">
						j ↓
					</TerminalButton>
					<TerminalButton onClick={onToc}>t toc</TerminalButton>
					<TerminalButton onClick={onFind}>/ find</TerminalButton>
					<TerminalButton onClick={onHelp} aria-label="Keyboard help">
						?
					</TerminalButton>
				</span>
			</div>
		</div>
	)
}
```

- [ ] **Step 7: Mount the status bar in `Manual.tsx`**

Add the imports:

```tsx
import { useRef } from 'react'
import { StatusBar } from './components/StatusBar'
import { SECTIONS } from './content'
import { useScrollPercent } from './hooks/useScrollPercent'
```

Inside `Manual`, above the `return`:

```tsx
const searchRef = useRef<HTMLInputElement>(null)
const percent = useScrollPercent()
const idx = 0
const noop = () => {}
```

Then, immediately before the closing `</div>` of the outermost element (after the `max-w-[96ch]` wrapper closes):

```tsx
<StatusBar
	current={SECTIONS[idx].label}
	position={`${idx + 1}/${SECTIONS.length}`}
	percent={percent}
	searching={false}
	query=""
	matchLabel="enter ↵ next"
	searchRef={searchRef}
	onQueryChange={noop}
	onSearchKeyDown={noop}
	onPrev={noop}
	onNext={noop}
	onToc={noop}
	onFind={noop}
	onHelp={noop}
/>
```

The `idx = 0` and `noop` placeholders are replaced in Tasks 6–8.

- [ ] **Step 8: Verify**

Run: `pnpm test:unit`
Expected: PASS, 11 tests.

Run: `pnpm dev`, scroll the page. The bar should read `NAME · 1/11 · 43%`, ending at `END`.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: add the fixed status bar and scroll percentage

Percentage math is a pure helper so it can be tested without a viewport.
Navigation, search, and overlay callbacks are stubbed for now."
```

---

### Task 6: Section navigation and keyboard control

**Files:**
- Create: `src/hooks/useManualNav.ts`
- Modify: `src/Manual.tsx`, `src/Manual.test.tsx`

**Interfaces:**
- Consumes: `SECTIONS` from Task 3; `StatusBar` from Task 5.
- Produces: `useManualNav(): { idx: number; goTo: (i: number) => void }`.

- [ ] **Step 1: Write the failing test**

Add to `src/Manual.test.tsx`:

```tsx
import userEvent from '@testing-library/user-event'

test('j and k move between sections', async () => {
	const user = userEvent.setup()
	render(<Manual />)
	expect(screen.getByTestId('current-section')).toHaveTextContent('NAME')

	await user.keyboard('j')
	expect(screen.getByTestId('position')).toHaveTextContent('2/11')
	expect(screen.getByTestId('current-section')).toHaveTextContent('SYNOPSIS')

	await user.keyboard('j')
	expect(screen.getByTestId('position')).toHaveTextContent('3/11')

	await user.keyboard('k')
	expect(screen.getByTestId('position')).toHaveTextContent('2/11')
})

test('g and G jump to the first and last section', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('{Shift>}G{/Shift}')
	expect(screen.getByTestId('position')).toHaveTextContent('11/11')

	await user.keyboard('g')
	expect(screen.getByTestId('position')).toHaveTextContent('1/11')
})

test('q jumps to SEE ALSO', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('q')
	expect(screen.getByTestId('current-section')).toHaveTextContent('SEE ALSO')
})

test('navigation keys are ignored while a modifier is held', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('{Control>}j{/Control}')
	expect(screen.getByTestId('position')).toHaveTextContent('1/11')
})
```

`window.scrollTo` is not implemented in jsdom and logs a noisy error. Add this stub to the top of `src/test-setup.ts`:

```ts
import '@testing-library/jest-dom/vitest'

window.scrollTo = () => {}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `pnpm test:unit`
Expected: FAIL — the `position` testid still reads `1/11` after pressing `j`, because `idx` is hardcoded to `0`.

- [ ] **Step 3: Create `src/hooks/useManualNav.ts`**

```ts
import { useCallback, useEffect, useRef, useState } from 'react'
import { SECTIONS } from '../content'

/** Milliseconds after an explicit jump during which observer updates are ignored. */
const LOCK_MS = 900

export function useManualNav() {
	const [idx, setIdx] = useState(0)
	const lockUntil = useRef(0)

	const goTo = useCallback((i: number) => {
		const n = Math.max(0, Math.min(i, SECTIONS.length - 1))
		const el = document.getElementById(SECTIONS[n].id)
		if (el) {
			window.scrollTo({
				top: el.getBoundingClientRect().top + window.scrollY - 24,
				behavior: 'smooth'
			})
		}
		lockUntil.current = Date.now() + LOCK_MS
		setIdx(n)
	}, [])

	useEffect(() => {
		// jsdom has no IntersectionObserver; unit tests drive idx through goTo instead.
		if (typeof IntersectionObserver === 'undefined') return

		const observer = new IntersectionObserver(
			(entries) => {
				if (Date.now() < lockUntil.current) return
				const visible = entries
					.filter((e) => e.isIntersecting)
					.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
				const top = visible[0]
				if (!top) return
				const i = SECTIONS.findIndex((s) => s.id === top.target.id)
				if (i >= 0) setIdx((current) => (current === i ? current : i))
			},
			{ rootMargin: '-24px 0px -70% 0px' }
		)

		for (const section of SECTIONS) {
			const el = document.getElementById(section.id)
			if (el) observer.observe(el)
		}
		return () => observer.disconnect()
	}, [])

	return { idx, goTo }
}
```

- [ ] **Step 4: Wire it into `Manual.tsx`**

Replace `const idx = 0` with:

```tsx
const { idx, goTo } = useManualNav()
```

Add the import:

```tsx
import { useManualNav } from './hooks/useManualNav'
```

Add the keydown effect below the hook calls:

```tsx
useEffect(() => {
	const onKey = (event: KeyboardEvent) => {
		if (event.metaKey || event.ctrlKey || event.altKey) return
		const tag = (event.target as HTMLElement | null)?.tagName?.toLowerCase()
		if (tag === 'input' || tag === 'textarea') return

		switch (event.key) {
			case 'j':
				event.preventDefault()
				goTo(idx + 1)
				break
			case 'k':
				event.preventDefault()
				goTo(idx - 1)
				break
			case 'g':
				event.preventDefault()
				goTo(0)
				break
			case 'G':
			case 'q':
				event.preventDefault()
				goTo(SECTIONS.length - 1)
				break
		}
	}

	window.addEventListener('keydown', onKey)
	return () => window.removeEventListener('keydown', onKey)
}, [goTo, idx])
```

Import `useEffect` from React. Replace the `onPrev` and `onNext` props on `<StatusBar>`:

```tsx
onPrev={() => goTo(idx - 1)}
onNext={() => goTo(idx + 1)}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS, 15 tests.

- [ ] **Step 6: Check the observer in a real browser**

Run: `pnpm dev`. Scroll slowly — the status bar section label should update as each heading passes the top. Press `j` repeatedly and confirm the label never flickers back to the previous section mid-scroll (that is the 900 ms lock working).

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add section navigation with j/k/g/G/q

An IntersectionObserver tracks the active section, suppressed for 900ms after
an explicit jump so the readout matches the key that was pressed."
```

---

### Task 7: Search

**Files:**
- Create: `src/hooks/useSearch.ts`, `src/hooks/useSearch.test.ts`
- Modify: `src/Manual.tsx`, `src/Manual.test.tsx`

**Interfaces:**
- Consumes: the `contentRef` on `<main>` from Task 4; `StatusBar` from Task 5.
- Produces:
  - `findMatches(root: HTMLElement, query: string): HTMLElement[]`
  - `useSearch(contentRef: RefObject<HTMLElement>): { query, setQuery, matchCount, matchLabel, showMatch, clear }` where `showMatch(offset: number): void` moves relative to the current match and `clear(): void` removes highlights and resets the query.

- [ ] **Step 1: Write the failing test**

Create `src/hooks/useSearch.test.ts`:

```ts
import { findMatches } from './useSearch'

function root(html: string): HTMLElement {
	const el = document.createElement('div')
	el.innerHTML = html
	return el
}

test('ignores queries of one character or less', () => {
	const el = root('<p>routing</p>')
	expect(findMatches(el, '')).toHaveLength(0)
	expect(findMatches(el, 'r')).toHaveLength(0)
	expect(findMatches(el, '  ')).toHaveLength(0)
})

test('matches leaf elements only, not their containers', () => {
	const el = root('<p><span>routing</span></p>')
	const matches = findMatches(el, 'rou')
	expect(matches).toHaveLength(1)
	expect(matches[0].tagName).toBe('SPAN')
})

test('matches case-insensitively', () => {
	const el = root('<p>Routing</p><p>ROUTING</p>')
	expect(findMatches(el, 'routing')).toHaveLength(2)
})

test('returns matches in document order', () => {
	const el = root('<p>first routing</p><p>second routing</p>')
	expect(findMatches(el, 'routing').map((m) => m.textContent)).toEqual([
		'first routing',
		'second routing'
	])
})

test('ignores elements that do not contain the query', () => {
	const el = root('<p>routing</p><p>geocoding</p>')
	expect(findMatches(el, 'geo')).toHaveLength(1)
})
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./useSearch"`.

- [ ] **Step 3: Create `src/hooks/useSearch.ts`**

The active match is highlighted by mutating the matched node's inline style. React does not own that text, so direct mutation is correct here.

```ts
import { useCallback, useEffect, useRef, useState, type RefObject } from 'react'

const SELECTOR = 'p, h1, h2, h3, span, a'

export function findMatches(root: HTMLElement, query: string): HTMLElement[] {
	const needle = query.trim().toLowerCase()
	if (needle.length < 2) return []
	return Array.from(root.querySelectorAll<HTMLElement>(SELECTOR)).filter(
		(el) => !el.querySelector(SELECTOR) && (el.textContent ?? '').toLowerCase().includes(needle)
	)
}

export function useSearch(contentRef: RefObject<HTMLElement | null>) {
	const [query, setQueryState] = useState('')
	const [matches, setMatches] = useState<HTMLElement[]>([])
	const [matchIdx, setMatchIdx] = useState(0)
	const highlighted = useRef<HTMLElement | null>(null)

	const clearHighlight = useCallback(() => {
		if (highlighted.current) {
			highlighted.current.style.background = ''
			highlighted.current.style.color = ''
			highlighted.current = null
		}
	}, [])

	const highlight = useCallback(
		(el: HTMLElement) => {
			clearHighlight()
			el.style.background = 'var(--ph)'
			el.style.color = '#0D0D0E'
			highlighted.current = el
			window.scrollTo({
				top: el.getBoundingClientRect().top + window.scrollY - 120,
				behavior: 'smooth'
			})
		},
		[clearHighlight]
	)

	const setQuery = useCallback(
		(value: string) => {
			clearHighlight()
			const found = contentRef.current ? findMatches(contentRef.current, value) : []
			setQueryState(value)
			setMatches(found)
			setMatchIdx(0)
			if (found.length > 0) highlight(found[0])
		},
		[clearHighlight, contentRef, highlight]
	)

	// Deliberately not a functional setState: highlight() mutates the DOM, and
	// StrictMode invokes updater functions twice.
	const showMatch = useCallback(
		(offset: number) => {
			if (matches.length === 0) return
			const next = (matchIdx + offset + matches.length) % matches.length
			highlight(matches[next])
			setMatchIdx(next)
		},
		[highlight, matchIdx, matches]
	)

	const clear = useCallback(() => {
		clearHighlight()
		setQueryState('')
		setMatches([])
		setMatchIdx(0)
	}, [clearHighlight])

	useEffect(() => clearHighlight, [clearHighlight])

	const matchLabel =
		matches.length > 0
			? `${matchIdx + 1}/${matches.length}`
			: query.trim().length > 1
				? 'no match'
				: 'enter ↵ next'

	return { query, setQuery, matchCount: matches.length, matchLabel, showMatch, clear }
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `pnpm test:unit`
Expected: PASS, 20 tests.

- [ ] **Step 5: Write the failing integration test**

Add to `src/Manual.test.tsx`:

```tsx
test('slash opens search and typing reports the match count', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('/')
	const input = screen.getByLabelText('Search the manual')
	expect(input).toHaveFocus()

	await user.type(input, 'geocoding')
	expect(screen.getByTestId('match-label')).toHaveTextContent(/^\d+\/\d+$/)
})

test('search reports no match for a query that is absent', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Search the manual'), 'zzzzqqq')
	expect(screen.getByTestId('match-label')).toHaveTextContent('no match')
})

test('escape closes search and restores the section readout', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Search the manual'), 'geocoding')
	await user.keyboard('{Escape}')

	expect(screen.queryByLabelText('Search the manual')).not.toBeInTheDocument()
	expect(screen.getByTestId('position')).toHaveTextContent('1/11')
})
```

- [ ] **Step 6: Run the tests to verify they fail**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find a label with the text of: Search the manual` (the `searching` prop is hardcoded `false`).

- [ ] **Step 7: Wire search into `Manual.tsx`**

Add the import and hook call:

```tsx
import { useSearch } from './hooks/useSearch'

const [searching, setSearching] = useState(false)
const search = useSearch(contentRef)
```

Add a callback that opens search and focuses the input:

```tsx
const startSearch = useCallback(() => {
	setSearching(true)
	// The input mounts in this same commit; focus after paint.
	requestAnimationFrame(() => searchRef.current?.focus())
}, [])
```

Extend the keydown switch with these cases, before the navigation cases:

```tsx
case 'Escape':
	event.preventDefault()
	setSearching(false)
	search.clear()
	break
case '/':
	event.preventDefault()
	startSearch()
	break
case 'n':
	event.preventDefault()
	search.showMatch(1)
	break
case 'N':
	event.preventDefault()
	search.showMatch(-1)
	break
```

Add `search` and `startSearch` to the effect's dependency array.

Replace the corresponding `<StatusBar>` props:

```tsx
searching={searching}
query={search.query}
matchLabel={search.matchLabel}
onQueryChange={search.setQuery}
onSearchKeyDown={(event) => {
	if (event.key === 'Enter') {
		event.preventDefault()
		search.showMatch(event.shiftKey ? -1 : 1)
	} else if (event.key === 'Escape') {
		event.preventDefault()
		setSearching(false)
		search.clear()
	}
}}
onFind={startSearch}
```

Import `useCallback` and `useState` from React.

- [ ] **Step 8: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS, 23 tests.

If the focus assertion fails, `requestAnimationFrame` did not run before the assertion. jsdom implements it, but if it proves flaky, swap `requestAnimationFrame` for `queueMicrotask`.

- [ ] **Step 9: Check it in a real browser**

Run: `pnpm dev`. Press `/`, type `routing`. The first match should invert to accent-on-black and scroll into view 120 px below the top. Press `n` and `N` to cycle. Press `Esc` — the highlight must disappear.

- [ ] **Step 10: Commit**

```bash
git add -A
git commit -m "feat: add less-style search over the manual

Matches leaf elements only, highlights one at a time by inverting it, and
cycles with n/N or enter. Match finding is a pure helper for testability."
```

---

### Task 8: Table of contents and key help overlays

**Files:**
- Create: `src/components/Overlay.tsx`
- Modify: `src/Manual.tsx`, `src/Manual.test.tsx`

**Interfaces:**
- Consumes: `SECTIONS`, `HELP` from Task 3.
- Produces: `Overlay({ title, rows, onClose })` where `rows: { key: string; label: string }[]`.

- [ ] **Step 1: Write the failing test**

Add to `src/Manual.test.tsx`:

```tsx
import { HELP } from './content'

test('t opens the table of contents with all eleven sections', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('t')
	const dialog = screen.getByRole('dialog')
	expect(dialog).toHaveTextContent('TABLE OF CONTENTS')
	expect(dialog).toHaveTextContent('01')
	expect(dialog).toHaveTextContent('11')
})

test('t toggles the table of contents closed again', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('t')
	expect(screen.getByRole('dialog')).toBeInTheDocument()
	await user.keyboard('t')
	expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})

test('question mark opens key help listing every binding', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('?')
	const dialog = screen.getByRole('dialog')
	expect(dialog).toHaveTextContent('KEYS')
	for (const row of HELP) {
		expect(dialog).toHaveTextContent(row.label)
	}
})

test('escape closes an open overlay', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('?')
	await user.keyboard('{Escape}')
	expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find an accessible element with the role "dialog"`.

- [ ] **Step 3: Create `src/components/Overlay.tsx`**

```tsx
export type OverlayRow = { key: string; label: string }

export function Overlay({
	title,
	rows,
	onClose
}: {
	title: string
	rows: OverlayRow[]
	onClose: () => void
}) {
	return (
		<div
			role="dialog"
			aria-modal="true"
			aria-label={title}
			onClick={onClose}
			className="fixed inset-0 z-[60] grid place-items-center bg-[rgba(8,8,9,0.86)] p-5"
		>
			<div
				onClick={(event) => event.stopPropagation()}
				className="w-full max-w-[60ch] border border-rule bg-panel px-6 py-[22px]"
			>
				<p className="mb-4 text-[12.5px] tracking-[0.16em] text-ph">{title}</p>
				{rows.map((row) => (
					<div
						key={row.key}
						className="grid grid-cols-[minmax(0,12ch)_minmax(0,1fr)] gap-x-4 py-[5px]"
					>
						<span className="text-bright">{row.key}</span>
						<span className="text-muted">{row.label}</span>
					</div>
				))}
				<p className="mt-4 text-[12.5px] text-dim">esc to close</p>
			</div>
		</div>
	)
}
```

Clicking the scrim closes; the inner panel stops propagation so clicks inside do not.

- [ ] **Step 4: Wire overlays into `Manual.tsx`**

Add the imports and state:

```tsx
import { Overlay, type OverlayRow } from './components/Overlay'
import { HELP, SECTIONS } from './content'

const [overlay, setOverlay] = useState<'help' | 'toc' | null>(null)
```

Derive the rows:

```tsx
const tocRows: OverlayRow[] = SECTIONS.map((section, n) => ({
	key: String(n + 1).padStart(2, '0'),
	label: section.label
}))
const overlayRows = overlay === 'toc' ? tocRows : HELP
const overlayTitle = overlay === 'toc' ? 'TABLE OF CONTENTS' : 'KEYS'
```

Extend the keydown switch:

```tsx
case 't':
	event.preventDefault()
	setOverlay((current) => (current === 'toc' ? null : 'toc'))
	break
case '?':
	event.preventDefault()
	setOverlay((current) => (current === 'help' ? null : 'help'))
	break
```

Add `setOverlay(null)` to the existing `Escape` case, and to `startSearch`.

`goTo` already closes overlays in the reference — add that here, in the keydown navigation cases and the status bar handlers, by calling `setOverlay(null)` alongside `goTo`. Extract a small wrapper so it is not repeated:

```tsx
const jump = useCallback(
	(i: number) => {
		setOverlay(null)
		goTo(i)
	},
	[goTo]
)
```

Replace every `goTo(...)` call site in `Manual.tsx` with `jump(...)`, and update the effect dependency array from `goTo` to `jump`.

Render the overlay just before `<StatusBar>`:

```tsx
{overlay && (
	<Overlay title={overlayTitle} rows={overlayRows} onClose={() => setOverlay(null)} />
)}
```

Wire the remaining status bar props:

```tsx
onToc={() => setOverlay((current) => (current === 'toc' ? null : 'toc'))}
onHelp={() => setOverlay((current) => (current === 'help' ? null : 'help'))}
```

Delete the now-unused `noop` constant.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS, 27 tests.

- [ ] **Step 6: Check it in a real browser**

Run: `pnpm dev`. Press `t`, then click the dark area outside the panel — it should close. Press `?`, click inside the panel — it should stay open. Press `Esc` — it should close.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add table of contents and key help overlays

Both are one piece of state. The scrim closes on click, the panel does not,
and the dialog carries role and aria-modal that the reference lacked."
```

---

### Task 9: Copy-to-clipboard on the email row

**Files:**
- Create: `src/components/CopyButton.tsx`, `src/components/CopyButton.test.tsx`
- Modify: `src/Manual.tsx`

**Interfaces:**
- Consumes: `TerminalButton` from Task 5; `EMAIL` from Task 3.
- Produces: `CopyButton({ value }: { value: string })`.

- [ ] **Step 1: Write the failing test**

Create `src/components/CopyButton.test.tsx`:

These tests use `fireEvent` rather than `userEvent`. `userEvent.setup()` installs its own
`navigator.clipboard` stub, which would fight the one under test.

```tsx
import { act, fireEvent, render, screen } from '@testing-library/react'
import { CopyButton } from './CopyButton'

function setClipboard(value: { writeText: (text: string) => Promise<void> } | undefined) {
	Object.defineProperty(navigator, 'clipboard', {
		value,
		configurable: true,
		writable: true
	})
}

afterEach(() => {
	setClipboard(undefined)
	vi.useRealTimers()
})

test('writes the value to the clipboard and confirms for 1600ms', async () => {
	const writeText = vi.fn().mockResolvedValue(undefined)
	setClipboard({ writeText })

	render(<CopyButton value="a@b.com" />)
	expect(screen.getByRole('button')).toHaveTextContent('copy')

	fireEvent.click(screen.getByRole('button'))
	expect(writeText).toHaveBeenCalledWith('a@b.com')

	// Let the writeText promise settle so setCopied(true) runs.
	await act(async () => {})
	expect(screen.getByRole('button')).toHaveTextContent('copied')
})

test('reverts to copy after the confirmation window', async () => {
	setClipboard({ writeText: vi.fn().mockResolvedValue(undefined) })

	render(<CopyButton value="a@b.com" />)
	fireEvent.click(screen.getByRole('button'))
	await act(async () => {})
	expect(screen.getByRole('button')).toHaveTextContent('copied')

	vi.useFakeTimers()
	await act(async () => {
		vi.advanceTimersByTime(1600)
	})
	expect(screen.getByRole('button')).toHaveTextContent('copy')
})

test('does nothing visible when the clipboard is unavailable', () => {
	setClipboard(undefined)

	render(<CopyButton value="a@b.com" />)
	fireEvent.click(screen.getByRole('button'))
	expect(screen.getByRole('button')).toHaveTextContent('copy')
})
```

`vi.useFakeTimers()` is installed only after the click in the second test — installing it
first would freeze the promise microtask queue that `writeText` resolves on.

- [ ] **Step 2: Run the test to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./CopyButton"`.

- [ ] **Step 3: Create `src/components/CopyButton.tsx`**

```tsx
import { useEffect, useRef, useState } from 'react'
import { TerminalButton } from './TerminalButton'

const CONFIRM_MS = 1600

export function CopyButton({ value }: { value: string }) {
	const [copied, setCopied] = useState(false)
	const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

	useEffect(() => () => clearTimeout(timer.current), [])

	const onClick = () => {
		if (!navigator.clipboard) return
		navigator.clipboard
			.writeText(value)
			.then(() => {
				setCopied(true)
				clearTimeout(timer.current)
				timer.current = setTimeout(() => setCopied(false), CONFIRM_MS)
			})
			.catch(() => {})
	}

	return (
		<TerminalButton onClick={onClick} className="ml-3 text-chrome" aria-label={`Copy ${value}`}>
			{copied ? 'copied' : 'copy'}
		</TerminalButton>
	)
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `pnpm test:unit`
Expected: PASS, 30 tests.

- [ ] **Step 5: Add it to the contact section**

In `src/Manual.tsx`, inside the `email` row of the SEE ALSO section, immediately after the closing `</a>`:

```tsx
<CopyButton value={EMAIL} />
```

Add the import:

```tsx
import { CopyButton } from './components/CopyButton'
```

Note: the existing Task 4 test `contact links point at the right destinations` finds the link by its accessible name. Adding a sibling button does not change that name, so it keeps passing.

- [ ] **Step 6: Verify**

Run: `pnpm test:unit`
Expected: PASS, 30 tests.

Run: `pnpm dev`, click `copy` next to the email. It should read `copied` and revert after about 1.6 seconds. Paste somewhere to confirm the address made it to the clipboard.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add copy-to-clipboard on the email row

Reverts to 'copy' after 1600ms, clears its timer on unmount, and degrades
silently where the clipboard API is unavailable."
```

---

### Task 10: Prerender the page so it reads without JavaScript

**Files:**
- Create: `prerender.ts`
- Modify: `src/main.tsx`, `package.json`
- Test: `tests/no-js.spec.ts`

**Interfaces:**
- Consumes: `Manual` from Task 4.
- Produces: a `dist/index.html` whose `#root` contains the fully rendered document. Task 11's end-to-end tests run against this build.

The spec requires that all content is readable with JavaScript disabled, and that only
the navigation aids need JS. A client-only Vite build ships an empty `#root`, so this
must be solved before the end-to-end suite can pass.

- [ ] **Step 1: Write the failing test**

Create `tests/no-js.spec.ts`:

```ts
import { expect, test } from '@playwright/test'

test('all content is readable without JavaScript', async ({ browser }) => {
	const context = await browser.newContext({ javaScriptEnabled: false })
	const page = await context.newPage()
	await page.goto('/')

	await expect(page.getByRole('heading', { level: 2 })).toHaveCount(11)
	await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
	await expect(page.getByText('Contributions to Jellyfin and Websurfx.')).toBeVisible()
	await expect(page.getByRole('link', { name: 'github.com/phetzy' })).toBeVisible()

	await context.close()
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:integration`
Expected: FAIL — `Expected: 11, Received: 0`. With scripting off, `#root` is empty.

- [ ] **Step 3: Add the prerender step**

```bash
pnpm add -D vite-node
```

`react-dom/server` ships with `react-dom` — no extra dependency. Create `prerender.ts` at
the repository root:

```ts
import { readFileSync, writeFileSync } from 'node:fs'
import { createElement } from 'react'
import { renderToString } from 'react-dom/server'
import { Manual } from './src/Manual'

const template = readFileSync('dist/index.html', 'utf8')
const html = renderToString(createElement(Manual))

writeFileSync(
	'dist/index.html',
	template.replace('<div id="root"></div>', `<div id="root">${html}</div>`)
)
```

`vite-node` is used rather than plain `tsc` output so the script runs through Vite's
transform pipeline and resolves the TypeScript and JSX in `src/` without a second build
config.

- [ ] **Step 4: Run it after the client build**

In `package.json`, change the build script:

```json
"build": "tsc --noEmit && vite build && vite-node prerender.ts"
```

- [ ] **Step 5: Switch the client from render to hydrate**

Replace `src/main.tsx`:

```tsx
import { StrictMode } from 'react'
import { hydrateRoot } from 'react-dom/client'
import { Analytics } from '@vercel/analytics/react'
import './index.css'
import { Manual } from './Manual'

const root = document.getElementById('root')
if (!root) throw new Error('#root not found')

hydrateRoot(
	root,
	<StrictMode>
		<Manual />
		<Analytics />
	</StrictMode>
)
```

`renderToString` runs `Manual` where `useEffect` never fires, so the observer, scroll
listener, and keydown handler attach only on hydration — exactly the behavior the spec
describes. `<Analytics />` renders no DOM of its own (it injects its script from an
effect), so prerendering `Manual` alone still produces markup that matches the hydrated
tree.

- [ ] **Step 6: Run the test to verify it passes**

Run: `pnpm test:integration`
Expected: PASS.

Confirm the markup is really in the file:

Run: `grep -c 'MAPWRIGHT' dist/index.html`
Expected: at least 1.

- [ ] **Step 7: Check for hydration mismatches**

Run: `pnpm preview`, open the page, and check the browser console. There must be no
hydration warnings. If one appears, the server and client rendered different markup —
the likely culprit is a hook whose initial state is computed rather than constant.
`useScrollPercent` must start at the literal `'0%'`, not at a measured value.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "build: prerender the page to static markup

renderToString writes the document into dist/index.html and the client
hydrates it, so the whole manual is readable with JavaScript disabled."
```

---

### Task 11: End-to-end suite and final verification

**Files:**
- Create: `tests/manual.spec.ts`
- Delete: `tests/test.ts`
- Modify: `README.md`

**Interfaces:**
- Consumes: the complete page from Tasks 1–10.
- Produces: nothing consumed by later tasks — this is the last one.

- [ ] **Step 1: Write the failing test**

The behaviors here depend on a real viewport, real scrolling, and a real `IntersectionObserver`, none of which jsdom provides. Create `tests/manual.spec.ts`:

```ts
import { expect, test } from '@playwright/test'

test('head metadata is intact', async ({ page }) => {
	await page.goto('/')
	await expect(page).toHaveTitle('FETZER(1) — Software Engineer')
	await expect(page.locator('meta[property="og:title"]')).toHaveAttribute(
		'content',
		'David Fetzer — Software Engineer'
	)
})

test('scrolling to the bottom reports END', async ({ page }) => {
	await page.goto('/')
	await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
	await expect(page.getByTestId('percent')).toHaveText('END')
})

test('pressing j five times lands on the sixth section', async ({ page }) => {
	await page.goto('/')
	for (let i = 0; i < 5; i++) await page.keyboard.press('j')
	await expect(page.getByTestId('position')).toHaveText('6/11')
	await expect(page.getByTestId('current-section')).toHaveText('SELF-HOSTED AI PLATFORM')
})

test('search finds a match and scrolls it into view', async ({ page }) => {
	await page.goto('/')
	await page.keyboard.press('/')
	await page.getByLabel('Search the manual').fill('geocoding')
	await expect(page.getByTestId('match-label')).toHaveText(/^\d+\/\d+$/)
	await expect(page.getByText('Forward, reverse, and POI search.')).toBeInViewport()
})

test('the table of contents lists all eleven sections and closes on scrim click', async ({
	page
}) => {
	await page.goto('/')
	await page.keyboard.press('t')
	const dialog = page.getByRole('dialog')
	await expect(dialog).toBeVisible()
	await expect(dialog.getByText('01')).toBeVisible()
	await expect(dialog.getByText('11')).toBeVisible()

	await page.mouse.click(5, 5)
	await expect(dialog).not.toBeVisible()
})
```

The no-JavaScript case already lives in `tests/no-js.spec.ts` from Task 10.

- [ ] **Step 2: Delete the old smoke test**

```bash
git rm tests/test.ts
```

Its title assertion is now covered by `head metadata is intact`.

- [ ] **Step 3: Run the suite**

Run: `pnpm test:integration`
Expected: all 6 pass — 5 from `manual.spec.ts`, 1 from `no-js.spec.ts`.

- [ ] **Step 4: Rewrite `README.md`**

It is still the `create-svelte` boilerplate. Replace its entire contents with:

````markdown
# fetzycloudonline

David Fetzer's personal site — a single static page styled as a Unix man page
rendered in a terminal pager.

## Development

```bash
pnpm install
pnpm dev
```

## Commands

| Command | Description |
| --- | --- |
| `pnpm dev` | Vite dev server |
| `pnpm build` | Type-check, bundle, and prerender to `dist/` |
| `pnpm preview` | Serve the production build |
| `pnpm test` | Vitest unit tests, then Playwright end-to-end tests |
| `pnpm lint` | Prettier check and ESLint |
| `pnpm format` | Rewrite files with Prettier |

## Structure

- `src/Manual.tsx` — the page: eleven sections and the keyboard wiring
- `src/content.ts` — content for the repeating rows
- `src/hooks/` — scroll percentage, section navigation, search
- `src/components/` — section, row, status bar, overlay, and button primitives

The design comes from `design_handoff_manual_site/`. Content is fixed: do not add
projects, metrics, or details that are not in that reference.

## Deployment

Vercel, static build. No environment variables.
````

- [ ] **Step 5: Full verification**

Run: `pnpm lint`
Expected: no errors.

Run: `pnpm build`
Expected: type-check clean, bundle written, prerender writes markup into `dist/index.html`.

Run: `pnpm test`
Expected: 30 unit tests pass, 6 Playwright tests pass.

- [ ] **Step 6: Manual pass against the reference**

Open `pnpm preview` alongside the served `Manual.dc.html`. Walk the three screenshots in `design_handoff_manual_site/screenshots/`. Then check, at a narrow viewport (375 px):

- The status bar wraps rather than overflowing.
- No horizontal scrollbar appears.
- Every status bar button is tappable.
- No body text renders below 12px.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "test: add end-to-end suite and rewrite the README

Playwright covers scroll percentage, section jumps, search, overlays, and
head metadata. The README replaces create-svelte boilerplate."
```

---

## Done

At this point the branch is ready for a pull request into `main`. Use the `superpowers:finishing-a-development-branch` skill to decide how to integrate.

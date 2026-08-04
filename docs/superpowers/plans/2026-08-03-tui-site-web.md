# TUI Site — Web Front End Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the manual-page site with a React front end that presents the personal site as a Bubble Tea program running inside a terminal window, reading from a shared `content.json`.

**Architecture:** One `App` owns the state the prototype owns — tab, selected section, focus, filter — and hands derived values to small presentational components. Pure selectors (visible list, list status, prompt segments) live outside React so they can be unit-tested directly. All nine sections are prerendered as semantic articles; hydration hides the inactive ones and switches the window to its fixed, non-scrolling layout.

**Tech Stack:** React 19, Vite 5, TypeScript, Tailwind 3, Vitest + React Testing Library, Playwright 1.61.0, deployed static to Vercel.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-08-03-tui-site-web-design.md`. Read it before starting.
- Design reference: `~/Downloads/tuiSite/design_handoff_tui_ssh/` — `SPEC.md` is authoritative, `TUI.dc.html` is the working prototype and the content source. `support.js` is prototype runtime only and must never be ported.
- **Content is frozen.** Every string comes verbatim from the prototype's `SECTIONS` array. No invented metrics, user counts, download numbers, client names, testimonials, tool versions, or project details. Do not expand the employer paragraph. No IP addresses, hostnames, ports, or network topology in the AI platform section. No fifth project. The veteran line is plain — no flag imagery, no iconography, no "thank you for your service" framing.
- Palette is Catppuccin Macchiato, exposed as Tailwind theme colors. Components never carry hex literals. The two accents are the CSS custom properties `--acc` (peach `#f5a97f`) and `--acc2` (blue `#8aadf4`), always written with their literal fallback.
- `#a5adcb` is the dimmest permitted for text. Darker violets are borders and rules only.
- Prettier is authoritative: tabs, single quotes, no semicolons, printWidth 100.
- Package manager is **pnpm**, never npm.
- No third-party font or asset requests. Fonts are self-hosted.
- `dims` and every other piece of state must have a constant initial value — the page is prerendered and hydrated, and a computed initial value causes a hydration mismatch.
- One commit per task. Work on branch `tui-site`.

## What already exists and must not be broken

The deploy pipeline was repaired in a previous session and is verified working. Do not modify without cause:

- `vercel.json` — pins `framework: null`, `outputDirectory: dist`. The project's Vercel framework preset is still SvelteKit from 2024; this file is what overrides it.
- `pnpm-workspace.yaml` — carries both `packages` (required by Vercel's pnpm 9) and `allowBuilds` (required by local pnpm 11 so esbuild can build).
- `prerender.ts` — `renderToString` into `dist/index.html`, with a guard that throws when the `#root` placeholder is missing.
- `src/main.tsx` — `hydrateRoot` plus the `data-hydrated` marker that end-to-end tests wait on.
- `public/favicons/` and `public/site.webmanifest` — the handoff's favicon set is byte-identical to what is installed. No favicon work is needed.
- `src/hooks/prefersReducedMotion.ts` — keep and extend.

---

## File Structure

| File | Responsibility |
| --- | --- |
| `content.json` | Shared content, nine sections. Also consumed by the Go SSH app later. |
| `src/content.ts` | Typed loader: `Section`, `Row`, `Link`, `TabId`, `SECTIONS`, `TABS` |
| `src/selectors.ts` | Pure derived logic: visible list, list status, prompt segments |
| `src/App.tsx` | State, keyboard wiring, window shell |
| `src/components/TitleBar.tsx` | Traffic lights, window title, `cols×rows` readout |
| `src/components/HeaderRow.tsx` | Title block + metadata block |
| `src/components/TabBar.tsx` | Four tabs; active reads `▌ label` |
| `src/components/ListPane.tsx` | Header, rows, `n/m` footer |
| `src/components/DetailPane.tsx` | Prompt header, the nine articles, caret |
| `src/components/Prompt.tsx` | Starship segments |
| `src/components/HelpFooter.tsx` | Key hints, or the filter input |
| `src/hooks/useTerminalDims.ts` | `cols×rows` from viewport size |
| `src/hooks/prefersReducedMotion.ts` | Existing; gains a `prefersReducedMotion()` boolean |
| `tests/tui.spec.ts` | Playwright end-to-end suite |
| `tests/no-js.spec.ts` | Rewritten for the nine-article fallback |

---

### Task 1: Strip the manual site, add the shared content source

**Files:**
- Delete: `src/Manual.tsx`, `src/Manual.test.tsx`, `src/content.test.ts`, `src/components/Section.tsx`, `src/components/rows.tsx`, `src/components/TerminalButton.tsx`, `src/components/StatusBar.tsx`, `src/components/Overlay.tsx`, `src/components/Overlay.test.tsx`, `src/components/CopyButton.tsx`, `src/components/CopyButton.test.tsx`, `src/hooks/useManualNav.ts`, `src/hooks/useScrollPercent.ts`, `src/hooks/useScrollPercent.test.ts`, `src/hooks/useSearch.ts`, `src/hooks/useSearch.test.ts`, `tests/manual.spec.ts`
- Create: `content.json`, `src/content.test.ts` (new), `src/App.tsx`
- Rewrite: `src/content.ts`
- Modify: `src/main.tsx`, `prerender.ts`, `tests/no-js.spec.ts`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `content.json` — the shared content file, also read by the Go app in sub-project 2.
  - From `src/content.ts`: `type TabId = 'readme' | 'projects' | 'work' | 'contact'`; `type Link = { label: string; href: string }`; `type Row = { label: string; body: string }`; `type Section = { id: string; tab: TabId; label: string; path: string; meta: string; title: string; kicker: string; paras: string[]; rows: Row[]; links: Link[] }`; `SECTIONS: Section[]`; `TABS: { id: TabId; label: string }[]`.
  - `App` — named export from `src/App.tsx`, `function App(): JSX.Element`.

**THE CRITICAL PART OF THIS TASK:** `content.json` is prose for a real person's site, frozen by contract. Transcribe it **character for character**: exact wording, punctuation, em dashes (—), en dashes (–), curly apostrophes (’), and capitalization. Do not paraphrase, "fix" grammar, or invent. The authoritative source is `~/Downloads/tuiSite/design_handoff_tui_ssh/TUI.dc.html`, lines 148–271.

- [ ] **Step 1: Write the failing test**

Create `src/content.test.ts`:

```ts
import { SECTIONS, TABS, type TabId } from './content'

test('nine sections in the documented order', () => {
	expect(SECTIONS.map((s) => s.id)).toEqual([
		'readme',
		'mapwright',
		'transfer',
		'platform',
		'hat',
		'oss',
		'c1',
		'stack',
		'contact'
	])
})

test('four tabs', () => {
	expect(TABS.map((t) => t.id)).toEqual(['readme', 'projects', 'work', 'contact'])
})

test('section ids are unique', () => {
	expect(new Set(SECTIONS.map((s) => s.id)).size).toBe(SECTIONS.length)
})

test('every section belongs to a known tab', () => {
	const known = new Set<TabId>(TABS.map((t) => t.id))
	for (const s of SECTIONS) expect(known.has(s.tab)).toBe(true)
})

test('every section has the required fields', () => {
	for (const s of SECTIONS) {
		expect(s.label).not.toBe('')
		expect(s.path).not.toBe('')
		expect(s.title).not.toBe('')
		expect(Array.isArray(s.paras)).toBe(true)
		expect(Array.isArray(s.rows)).toBe(true)
		expect(Array.isArray(s.links)).toBe(true)
	}
})

test('every link has a label and a non-empty href', () => {
	for (const s of SECTIONS) {
		for (const l of s.links) {
			expect(l.label).not.toBe('')
			expect(l.href).not.toBe('')
		}
	}
})

test('the contact section carries three links', () => {
	const contact = SECTIONS.find((s) => s.id === 'contact')!
	expect(contact.links.map((l) => l.href)).toEqual([
		'mailto:david.j.fetzer@gmail.com',
		'https://github.com/phetzy',
		'https://linkedin.com/in/fetzy'
	])
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — the old `content.ts` exports `SECTIONS` with different ids, so the first test fails on the id list.

- [ ] **Step 3: Delete the manual implementation**

```bash
git rm src/Manual.tsx src/Manual.test.tsx \
  src/components/Section.tsx src/components/rows.tsx src/components/TerminalButton.tsx \
  src/components/StatusBar.tsx src/components/Overlay.tsx src/components/Overlay.test.tsx \
  src/components/CopyButton.tsx src/components/CopyButton.test.tsx \
  src/hooks/useManualNav.ts src/hooks/useScrollPercent.ts src/hooks/useScrollPercent.test.ts \
  src/hooks/useSearch.ts src/hooks/useSearch.test.ts \
  tests/manual.spec.ts
```

Keep `src/hooks/prefersReducedMotion.ts` and its test — later tasks extend it.

- [ ] **Step 4: Create `content.json`**

Transcribed verbatim from the prototype. Every string below is final.

```json
{
	"tabs": [
		{ "id": "readme", "label": "readme" },
		{ "id": "projects", "label": "projects" },
		{ "id": "work", "label": "work" },
		{ "id": "contact", "label": "contact" }
	],
	"sections": [
		{
			"id": "readme",
			"tab": "readme",
			"label": "README.md",
			"path": "~/site/README.md",
			"meta": "overview",
			"title": "David Fetzer",
			"kicker": "software engineer · boise, id · remote",
			"paras": [
				"I build infrastructure — map servers, command-line tools, self-hosted platforms, and the hardware underneath them.",
				"Four independent projects carry this page: a self-hosted map server sold as a Mapbox alternative, a static Go binary for browserless file transfer, a private LLM and RAG platform on Kubernetes, and a secure element board for the Raspberry Pi."
			],
			"rows": [
				{ "label": "location", "body": "Boise, ID · remote" },
				{ "label": "role", "body": "Software Engineer at C1, March 2022 – present" },
				{ "label": "status", "body": "Open to interesting conversations" }
			],
			"links": []
		},
		{
			"id": "mapwright",
			"tab": "projects",
			"label": "mapwright",
			"path": "~/site/projects/mapwright",
			"meta": "flagship",
			"title": "mapwright",
			"kicker": "self-hosted map infrastructure · mapwright.io",
			"paras": [
				"A self-hosted map server. The full Mapbox stack — tiles, styles, geocoding, routing, static maps — running on your own hardware as a single Docker image with embedded Postgres.",
				"Mapwright speaks Mapbox and MapTiler URL shapes, so existing MapLibre applications migrate by changing a base URL rather than rewriting against a new SDK. It runs air-gapped for on-premise and data-residency requirements."
			],
			"rows": [
				{ "label": "01 tiles", "body": "Vector and raster tiles. 127 GB for full-planet coverage." },
				{ "label": "02 styles", "body": "Visual style editor with five standard styles." },
				{ "label": "03 geocoding", "body": "Forward and reverse geocoding with POI search." },
				{
					"label": "04 routing",
					"body": "Directions, matrix, isochrone, map-matching, optimization. 93 GB, full planet."
				},
				{ "label": "05 static maps", "body": "Server-rendered static maps." },
				{ "label": "06 wmts/wms", "body": "WMTS and WMS for standards-based GIS clients." },
				{ "label": "07 api keys", "body": "Scopes, origin allowlists, metering, and rate limits." },
				{ "label": "08 admin", "body": "Admin console." }
			],
			"links": [
				{ "label": "mapwright.io", "href": "https://mapwright.io" },
				{ "label": "docs", "href": "https://mapwright.io/docs" }
			]
		},
		{
			"id": "transfer",
			"tab": "projects",
			"label": "transfer-it-cli",
			"path": "~/site/projects/transfer-it-cli",
			"meta": "go · TUI + CLI",
			"title": "transfer-it-cli",
			"kicker": "single static go binary · MIT",
			"paras": [
				"A single static Go binary for MEGA's transfer.it, shipping as both an interactive TUI and a scriptable CLI. Built for machines without a browser — servers, embedded boxes, SSH-only sessions. Zero runtime dependencies."
			],
			"rows": [
				{
					"label": "protocol",
					"body": "The upload path is reverse-engineered from transfer.it's WebSocket protocol."
				},
				{
					"label": "throughput",
					"body": "Multi-pool parallel connections, in-session retry with exponential backoff, and token-bucket bandwidth limiting over a shared HTTP/2 client with TLS session resumption."
				},
				{
					"label": "memory",
					"body": "Transfers stream, so memory stays bounded by the chunk size — under a megabyte, even on 20+ GiB files."
				},
				{
					"label": "resume",
					"body": "Uploads, downloads, and recursive folder uploads all resume after process restart, tracked by a per-transfer chunk bitmap with per-chunk MAC verification that stays forward-compatible across binary upgrades."
				},
				{
					"label": "packaging",
					"body": "Six cross-compiled targets plus native .deb, .rpm, and .apk packages."
				},
				{
					"label": "ci",
					"body": "staticcheck, govulncheck, and race-detector tests across Linux, macOS, and Windows, alongside a weekly check that hashes upstream bundles and opens an issue when the service changes underneath it."
				}
			],
			"links": []
		},
		{
			"id": "platform",
			"tab": "projects",
			"label": "ai-platform",
			"path": "~/site/projects/ai-platform",
			"meta": "k3s · rag",
			"title": "self-hosted ai platform",
			"kicker": "private LLM and RAG stack on kubernetes",
			"paras": [
				"A multi-node K3s cluster running across Raspberry Pis, with GPU inference offloaded to a virtualized host elsewhere on the same network. Open WebUI, Ollama, a FastAPI retrieval service over a Qdrant vector store, Prometheus and Grafana observability, and automated backup-freshness checks."
			],
			"rows": [
				{
					"label": "ops agent",
					"body": "A read-only operations agent answers infrastructure questions from runbooks and live cluster state."
				},
				{
					"label": "safety",
					"body": "Every request passes a default-deny safety router: an LLM intent classifier and a regex gate run independently, and either one flagging write intent stops execution before it starts. Writes come back as a scoped review checklist — the agent proposes, a human approves."
				},
				{
					"label": "architecture",
					"body": "Mixed-architecture throughout. Cluster workloads are arm64 images built and served from a private registry with upstream images pinned by digest, while inference runs on x86 with a passed-through consumer AMD GPU."
				},
				{
					"label": "inference",
					"body": "Vulkan rules out flash attention and KV-cache quantization, so throughput comes from context and keep-alive tuning instead."
				},
				{
					"label": "services",
					"body": "Python and FastAPI services, a Go backend-for-frontend, React and TypeScript console."
				}
			],
			"links": []
		},
		{
			"id": "hat",
			"tab": "projects",
			"label": "secure-element-hat",
			"path": "~/site/projects/secure-element-hat",
			"meta": "kicad · pcb",
			"title": "secure element HAT",
			"kicker": "raspberry pi add-on board",
			"paras": ["A custom PCB designed in KiCad around an ATECC608B-TNGTLS secure element."],
			"rows": [
				{ "label": "interface", "body": "I2C." },
				{ "label": "keys", "body": "Hardware-backed key storage and TLS provisioning." },
				{
					"label": "integration",
					"body": "Device tree overlay and EEPROM configuration for Raspberry Pi."
				}
			],
			"links": []
		},
		{
			"id": "oss",
			"tab": "projects",
			"label": "open-source",
			"path": "~/site/projects/open-source",
			"meta": "contributions",
			"title": "open source",
			"kicker": "contributions",
			"paras": ["Contributions to Jellyfin and Websurfx."],
			"rows": [],
			"links": []
		},
		{
			"id": "c1",
			"tab": "work",
			"label": "c1",
			"path": "~/site/work/c1",
			"meta": "mar 2022 – present",
			"title": "C1 — Software Engineer",
			"kicker": "march 2022 – present · remote",
			"paras": [
				"Distributed systems and embedded software. Backend services in Go and TypeScript, React frontends, and AWS infrastructure managed as code with Terraform. Also builds internal tooling used across engineering teams."
			],
			"rows": [{ "label": "prior", "body": "US Army infantry veteran, 2012–2015." }],
			"links": []
		},
		{
			"id": "stack",
			"tab": "work",
			"label": "stack",
			"path": "~/site/work/stack",
			"meta": "index",
			"title": "stack",
			"kicker": "languages, services, and tools",
			"paras": [],
			"rows": [
				{ "label": "languages", "body": "Go, TypeScript, JavaScript, Python" },
				{
					"label": "backend",
					"body": "Node.js, Express, PostgreSQL, PostGIS, REST APIs, WebSockets"
				},
				{ "label": "frontend", "body": "React, Tailwind" },
				{
					"label": "infrastructure",
					"body": "AWS, Terraform, OpenTofu, Docker, Kubernetes, Linux"
				},
				{ "label": "hardware", "body": "KiCad, embedded Linux, I2C" },
				{
					"label": "domains",
					"body": "geospatial, embedded systems, self-hosted infrastructure"
				}
			],
			"links": []
		},
		{
			"id": "contact",
			"tab": "contact",
			"label": "contact",
			"path": "~/site/contact",
			"meta": "3 entries",
			"title": "contact",
			"kicker": "no phone, no form",
			"paras": [],
			"rows": [
				{ "label": "email", "body": "david.j.fetzer@gmail.com" },
				{ "label": "github", "body": "github.com/phetzy" },
				{ "label": "linkedin", "body": "linkedin.com/in/fetzy" }
			],
			"links": [
				{ "label": "david.j.fetzer@gmail.com", "href": "mailto:david.j.fetzer@gmail.com" },
				{ "label": "github.com/phetzy", "href": "https://github.com/phetzy" },
				{ "label": "linkedin.com/in/fetzy", "href": "https://linkedin.com/in/fetzy" }
			]
		}
	]
}
```

- [ ] **Step 5: Rewrite `src/content.ts`**

```ts
import data from '../content.json'

export type TabId = 'readme' | 'projects' | 'work' | 'contact'
export type Link = { label: string; href: string }
export type Row = { label: string; body: string }

export type Section = {
	id: string
	tab: TabId
	label: string
	path: string
	meta: string
	title: string
	kicker: string
	paras: string[]
	rows: Row[]
	links: Link[]
}

export type Tab = { id: TabId; label: string }

export const SECTIONS: Section[] = data.sections as Section[]
export const TABS: Tab[] = data.tabs as Tab[]
```

`resolveJsonModule` is already enabled in `tsconfig.json`, so the import type-checks.

- [ ] **Step 6: Create a placeholder `src/App.tsx`**

Later tasks build this out. It exists now so the build and the prerender keep working.

```tsx
import { SECTIONS } from './content'

export function App() {
	return (
		<div>
			<h1>David Fetzer</h1>
			{SECTIONS.map((s) => (
				<article key={s.id}>
					<h2>{s.title}</h2>
				</article>
			))}
		</div>
	)
}
```

- [ ] **Step 7: Repoint `main.tsx` and `prerender.ts`**

In `src/main.tsx`, change the import from `./Manual` to `./App` and the element from `<Manual />` to `<App />`. Leave `hydrateRoot`, `<Analytics />`, and the `data-hydrated` marker exactly as they are.

In `prerender.ts`, change the import from `./src/Manual` to `./src/App` and `createElement(Manual)` to `createElement(App)`. Leave the placeholder guard alone.

- [ ] **Step 8: Replace `tests/no-js.spec.ts`**

The old assertions were written for the manual page. Replace its contents:

```ts
import { expect, test } from '@playwright/test'

test('all nine sections are readable without JavaScript', async ({ browser }) => {
	const context = await browser.newContext({ javaScriptEnabled: false })
	const page = await context.newPage()
	await page.goto('/')

	await expect(page.locator('article')).toHaveCount(9)
	await expect(page.getByText('Contributions to Jellyfin and Websurfx.')).toBeVisible()

	await context.close()
})
```

- [ ] **Step 9: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS — 7 content tests plus the 3 existing `prefersReducedMotion` tests.

Run: `pnpm build`
Expected: type-check clean, bundle written, prerender injects markup.

Run: `pnpm test:integration`
Expected: PASS.

- [ ] **Step 10: Audit the transcription**

Before committing, re-read `content.json` against the prototype string by string. Confirm every one matches character for character, including dashes and apostrophes. Report what you checked and what you found.

- [ ] **Step 11: Commit**

```bash
git add -A
git commit -m "feat: replace manual site content with the shared TUI content source

Removes the manual-page implementation and adds content.json, transcribed
verbatim from the design prototype. The Go SSH app will embed the same file.
App.tsx is a placeholder that later tasks build out."
```

---

### Task 2: Palette, self-hosted fonts, and the window chrome

**Files:**
- Create: `src/components/TitleBar.tsx`, `src/hooks/useTerminalDims.ts`, `src/hooks/useTerminalDims.test.ts`
- Modify: `tailwind.config.ts`, `src/index.css`, `src/App.tsx`, `index.html`, `public/site.webmanifest`
- Add: `public/fonts/SymbolsNerdFont-Regular.ttf`

**Interfaces:**
- Consumes: `SECTIONS` from Task 1.
- Produces:
  - Tailwind colors: `crust`, `base`, `mantle`, `surface0`, `surface1`, `rule`, `text`, `subtext1`, `subtext0`, `lavender`, `mauve`, `green`, `red`, `yellow`, `acc`, `acc2`
  - `formatDims(width: number, height: number): string` and `useTerminalDims(): string` from `src/hooks/useTerminalDims.ts`
  - `TitleBar({ dims }: { dims: string })`

- [ ] **Step 1: Install the webfont**

```bash
pnpm add @fontsource/cascadia-code
```

Vendor the symbols font rather than pulling it from a CDN:

```bash
mkdir -p public/fonts
curl -fsSL -o public/fonts/SymbolsNerdFont-Regular.ttf \
  https://cdn.jsdelivr.net/gh/ryanoasis/nerd-fonts@v3.2.1/patched-fonts/NerdFontsSymbolsOnly/SymbolsNerdFont-Regular.ttf
```

Verify it downloaded as a real font, not an error page:

```bash
file public/fonts/SymbolsNerdFont-Regular.ttf
```

Expected: TrueType font data. If it is HTML or the download fails, stop and report — do not fall back to the CDN link.

- [ ] **Step 2: Write the failing test**

Create `src/hooks/useTerminalDims.test.ts`:

```ts
import { formatDims } from './useTerminalDims'

test('derives a cols×rows readout from the viewport', () => {
	expect(formatDims(1680, 1050)).toBe('200×55')
	expect(formatDims(840, 380)).toBe('100×20')
})

test('never reports a non-positive dimension', () => {
	expect(formatDims(0, 0)).toBe('1×1')
})
```

- [ ] **Step 3: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./useTerminalDims"`.

- [ ] **Step 4: Create `src/hooks/useTerminalDims.ts`**

```ts
import { useEffect, useState } from 'react'

/** The prototype's cell metrics: 8.4px per column, 19px per row. */
const CELL_W = 8.4
const CELL_H = 19

export function formatDims(width: number, height: number): string {
	const cols = Math.max(1, Math.round(width / CELL_W))
	const rows = Math.max(1, Math.round(height / CELL_H))
	return `${cols}×${rows}`
}

/**
 * Starts empty and measures after mount. The document is prerendered, so a
 * computed initial value would not match what the server rendered.
 */
export function useTerminalDims(): string {
	const [dims, setDims] = useState('')

	useEffect(() => {
		const measure = () => setDims(formatDims(window.innerWidth, window.innerHeight))
		measure()
		window.addEventListener('resize', measure)
		return () => window.removeEventListener('resize', measure)
	}, [])

	return dims
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `pnpm test:unit`
Expected: PASS.

- [ ] **Step 6: Rewrite `tailwind.config.ts`**

```ts
import type { Config } from 'tailwindcss'

export default {
	content: ['./index.html', './src/**/*.{ts,tsx}'],
	theme: {
		extend: {
			colors: {
				crust: '#181926',
				base: '#24273a',
				mantle: '#1e2030',
				surface0: '#363a4f',
				surface1: '#494d64',
				rule: '#2f3348',
				text: '#cad3f5',
				subtext1: '#b8c0e0',
				subtext0: '#a5adcb',
				lavender: '#b7bdf8',
				mauve: '#c6a0f6',
				green: '#a6da95',
				red: '#ed8796',
				yellow: '#eed49f',
				acc: 'var(--acc, #f5a97f)',
				acc2: 'var(--acc2, #8aadf4)'
			},
			fontFamily: {
				mono: [
					'CaskaydiaCove NF',
					'Cascadia Code',
					'Symbols Nerd Font',
					'ui-monospace',
					'monospace'
				]
			},
			keyframes: {
				blink: { '0%, 49%': { opacity: '1' }, '50%, 100%': { opacity: '0' } },
				slidedown: {
					from: { opacity: '0', transform: 'translateY(10px)' },
					to: { opacity: '1', transform: 'none' }
				},
				slideup: {
					from: { opacity: '0', transform: 'translateY(-10px)' },
					to: { opacity: '1', transform: 'none' }
				},
				slidex: {
					from: { opacity: '0', transform: 'translateX(-10px)' },
					to: { opacity: '1', transform: 'none' }
				}
			},
			animation: {
				blink: 'blink 1.1s step-end infinite',
				slidedown: 'slidedown 170ms cubic-bezier(0.22, 1, 0.36, 1)',
				slideup: 'slideup 170ms cubic-bezier(0.22, 1, 0.36, 1)',
				slidex: 'slidex 190ms cubic-bezier(0.22, 1, 0.36, 1) both'
			}
		}
	},
	plugins: []
} as Config
```

- [ ] **Step 7: Rewrite `src/index.css`**

```css
@import '@fontsource/cascadia-code/latin-300.css';
@import '@fontsource/cascadia-code/latin-400.css';
@import '@fontsource/cascadia-code/latin-600.css';
@import '@fontsource/cascadia-code/latin-700.css';

@tailwind base;
@tailwind components;
@tailwind utilities;

/* Installed locally by some visitors; never fetched over the network. */
@font-face {
	font-family: 'CaskaydiaCove NF';
	src:
		local('CaskaydiaCove Nerd Font'),
		local('CaskaydiaCove NF'),
		local('CaskaydiaCoveNerdFont-Regular');
	font-display: swap;
}

/* Self-hosted so no CDN sees a visitor request. */
@font-face {
	font-family: 'Symbols Nerd Font';
	src:
		local('Symbols Nerd Font'),
		url('/fonts/SymbolsNerdFont-Regular.ttf') format('truetype');
	font-display: swap;
}

:root {
	--acc: #f5a97f;
	--acc2: #8aadf4;
}

body {
	margin: 0;
	background: #181926;
}

::selection {
	background: #b7bdf8;
	color: #24273a;
}

/* Pane scrollbars, styled to match the terminal. */
[data-vp]::-webkit-scrollbar {
	width: 10px;
}
[data-vp]::-webkit-scrollbar-track {
	background: #1e2030;
}
[data-vp]::-webkit-scrollbar-thumb {
	background: #494d64;
}
[data-vp]::-webkit-scrollbar-thumb:hover {
	background: var(--acc, #f5a97f);
}
```

If `@fontsource/cascadia-code` does not publish a `latin-300` or `latin-700` file, list the package's actual CSS entry points and use the closest available weights, then say which you used in your report. Do not silently drop a weight.

- [ ] **Step 8: Create `src/components/TitleBar.tsx`**

```tsx
export function TitleBar({ dims }: { dims: string }) {
	return (
		<div className="flex flex-none items-center gap-3 border-b border-surface0 bg-mantle px-[14px] py-[10px]">
			<span className="flex gap-[7px]" aria-hidden="true">
				<span className="h-[11px] w-[11px] rounded-full bg-red" />
				<span className="h-[11px] w-[11px] rounded-full bg-yellow" />
				<span className="h-[11px] w-[11px] rounded-full bg-green" />
			</span>
			<span className="flex-1 truncate text-center text-[12px] text-subtext0">
				fetzer@boise: ~/site — go run ./cmd/fetzer
			</span>
			<span className="text-[11.5px] text-subtext0">{dims}</span>
		</div>
	)
}
```

- [ ] **Step 9: Build the window shell in `src/App.tsx`**

```tsx
import { TitleBar } from './components/TitleBar'
import { SECTIONS } from './content'
import { useTerminalDims } from './hooks/useTerminalDims'

export function App() {
	const dims = useTerminalDims()

	return (
		<div className="h-screen overflow-hidden bg-crust p-[clamp(8px,2.4vw,32px)] font-mono text-text">
			<div className="mx-auto flex h-full max-w-[1220px] flex-col overflow-hidden rounded-[10px] border border-surface0 bg-base">
				<TitleBar dims={dims} />
				<div className="flex min-h-0 flex-1 flex-col gap-3 p-[clamp(12px,2vw,20px)]">
					{SECTIONS.map((s) => (
						<article key={s.id}>
							<h2>{s.title}</h2>
						</article>
					))}
				</div>
			</div>
		</div>
	)
}
```

- [ ] **Step 10: Update the head in `index.html`**

Change the title to `fetzer — TUI`. Replace the description, `og:title`, and `og:description` with the prototype's:

```html
<title>fetzer — TUI</title>
<meta
	name="description"
	content="Software engineer in Boise, Idaho. I build infrastructure — map servers, command-line tools, self-hosted platforms, and the hardware underneath them."
/>
<meta property="og:title" content="David Fetzer — Software Engineer" />
<meta property="og:type" content="website" />
<meta
	property="og:description"
	content="I build infrastructure — map servers, command-line tools, self-hosted platforms, and the hardware underneath them."
/>
```

Leave the favicon links, the manifest link, and the `theme-color` meta alone — the handoff ships the same favicon set already installed. Update `theme-color` to `#181926` to match the new ground, and update `background_color` and `theme_color` in `public/site.webmanifest` to `#181926` as well.

Remove the Twitter card tags if present; the prototype does not carry them.

- [ ] **Step 11: Verify**

Run: `pnpm lint`, `pnpm build`, `pnpm test`
Expected: all clean.

Confirm the fonts are self-hosted and no CDN reference survives:

```bash
grep -rn "jsdelivr\|googleapis\|gstatic" index.html src/ || echo "no third-party font refs"
ls dist/assets/*.woff2 | head
```

- [ ] **Step 12: Commit**

```bash
git add -A
git commit -m "feat: add the Catppuccin palette, self-hosted fonts, and window chrome

Cascadia Code comes from @fontsource and the Symbols Nerd Font TTF is
vendored, so the page makes no third-party font request. Adds the terminal
window shell with its title bar and live cols×rows readout."
```

---

### Task 3: Pure selectors

**Files:**
- Create: `src/selectors.ts`, `src/selectors.test.ts`

**Interfaces:**
- Consumes: `SECTIONS`, `TABS`, `Section`, `TabId` from Task 1.
- Produces, all from `src/selectors.ts`:
  - `visibleSections(tab: TabId, filter: string, filtering: boolean): Section[]`
  - `listStatus(visible: Section[], selectedId: string, filter: string): string`
  - `promptSegments(section: Section): { text: string; color: string }[]`
  - `firstSectionOfTab(tab: TabId): Section`
  - `clampIndex(current: number, delta: number, length: number): number`

These are the behaviors the whole interface turns on, so they live outside React and are tested directly.

- [ ] **Step 1: Write the failing test**

Create `src/selectors.test.ts`:

```ts
import {
	clampIndex,
	firstSectionOfTab,
	listStatus,
	promptSegments,
	visibleSections
} from './selectors'
import { SECTIONS } from './content'

test('with no filter, shows only the active tab', () => {
	expect(visibleSections('projects', '', false).map((s) => s.id)).toEqual([
		'mapwright',
		'transfer',
		'platform',
		'hat',
		'oss'
	])
	expect(visibleSections('work', '', false).map((s) => s.id)).toEqual(['c1', 'stack'])
})

test('while the filter is open, searches every section regardless of tab', () => {
	const hits = visibleSections('readme', 'stack', true).map((s) => s.id)
	expect(hits).toContain('stack')
})

test('with a filter string but the filter closed, stays within the active tab', () => {
	expect(visibleSections('readme', 'stack', false).map((s) => s.id)).toEqual([])
})

test('filter matches label and title, case-insensitively', () => {
	expect(visibleSections('projects', 'MAPWRIGHT', true).map((s) => s.id)).toEqual(['mapwright'])
	expect(visibleSections('projects', 'secure element', true).map((s) => s.id)).toEqual(['hat'])
})

test('filter with no match returns nothing', () => {
	expect(visibleSections('projects', 'zzzzqqq', true)).toEqual([])
})

test('list status reports position and total', () => {
	const visible = visibleSections('projects', '', false)
	expect(listStatus(visible, 'mapwright', '')).toBe('1/5')
	expect(listStatus(visible, 'oss', '')).toBe('5/5')
})

test('list status marks an active filter', () => {
	const visible = visibleSections('projects', 'map', true)
	expect(listStatus(visible, 'mapwright', 'map')).toBe('1/1 filtered')
})

test('list status reports no match on an empty list', () => {
	expect(listStatus([], 'mapwright', 'zzz')).toBe('no match')
})

test('prompt always carries directory and branch', () => {
	const readme = SECTIONS.find((s) => s.id === 'readme')!
	const segs = promptSegments(readme)
	expect(segs[0].text).toBe('~/site/README.md')
	expect(segs[1].text).toContain('main')
	expect(segs).toHaveLength(2)
})

test('prompt adds a language module only for sections that define one', () => {
	const mapwright = SECTIONS.find((s) => s.id === 'mapwright')!
	expect(promptSegments(mapwright)).toHaveLength(3)
	expect(promptSegments(mapwright)[2].text).toContain('docker')

	const contact = SECTIONS.find((s) => s.id === 'contact')!
	expect(promptSegments(contact)).toHaveLength(2)
})

test('first section of a tab', () => {
	expect(firstSectionOfTab('projects').id).toBe('mapwright')
	expect(firstSectionOfTab('work').id).toBe('c1')
})

test('clampIndex stops at both ends and does not wrap', () => {
	expect(clampIndex(0, -1, 5)).toBe(0)
	expect(clampIndex(4, 1, 5)).toBe(4)
	expect(clampIndex(0, 99, 5)).toBe(4)
	expect(clampIndex(4, -99, 5)).toBe(0)
	expect(clampIndex(1, 1, 5)).toBe(2)
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Failed to resolve import "./selectors"`.

- [ ] **Step 3: Create `src/selectors.ts`**

The `LANG` map and the filter semantics come straight from the prototype.

```ts
import { SECTIONS, type Section, type TabId } from './content'

/** Starship language modules, per section, in Macchiato-mapped colors. */
const LANG: Record<string, { text: string; color: string }> = {
	mapwright: { text: ' docker', color: '#8aadf4' },
	transfer: { text: ' go', color: '#91d7e3' },
	platform: { text: ' python', color: '#eed49f' },
	hat: { text: ' i2c', color: '#8bd5ca' },
	c1: { text: ' go', color: '#91d7e3' },
	stack: { text: ' node', color: '#a6da95' }
}

/**
 * While the filter input is open the search covers every section, not just the
 * active tab — selecting a match is what moves you to another tab. With a
 * filter string but the input closed, the active tab still bounds the list.
 */
export function visibleSections(tab: TabId, filter: string, filtering: boolean): Section[] {
	const q = filter.trim().toLowerCase()
	const inTab = SECTIONS.filter((s) => s.tab === tab)
	if (!q) return inTab
	const pool = filtering ? SECTIONS : inTab
	return pool.filter((s) => `${s.label} ${s.title}`.toLowerCase().includes(q))
}

export function listStatus(visible: Section[], selectedId: string, filter: string): string {
	if (visible.length === 0) return 'no match'
	const i = visible.findIndex((s) => s.id === selectedId)
	return `${i + 1}/${visible.length}${filter ? ' filtered' : ''}`
}

export function promptSegments(section: Section): { text: string; color: string }[] {
	const segs = [
		{ text: section.path, color: '#b7bdf8' },
		{ text: ' main', color: '#c6a0f6' }
	]
	const lang = LANG[section.id]
	if (lang) segs.push(lang)
	return segs
}

export function firstSectionOfTab(tab: TabId): Section {
	return SECTIONS.filter((s) => s.tab === tab)[0] ?? SECTIONS[0]
}

export function clampIndex(current: number, delta: number, length: number): number {
	return Math.max(0, Math.min(current + delta, length - 1))
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS — 12 selector tests added.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add the pure selectors behind the TUI

Filter scope, list status, prompt segments, and index clamping live outside
React so the behaviors the interface turns on are directly testable."
```

---

### Task 4: Header row and tab bar

**Files:**
- Create: `src/components/HeaderRow.tsx`, `src/components/TabBar.tsx`
- Modify: `src/App.tsx`, `src/App.test.tsx` (create)

**Interfaces:**
- Consumes: Tailwind colors from Task 2; `TABS`, `TabId` from Task 1; `firstSectionOfTab` from Task 3.
- Produces: `HeaderRow()`, `TabBar({ tab, onSelect }: { tab: TabId; onSelect: (t: TabId) => void })`.

- [ ] **Step 1: Write the failing test**

Create `src/App.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App } from './App'

test('renders the title block and metadata', () => {
	render(<App />)
	expect(screen.getByText('DAVID FETZER')).toBeInTheDocument()
	expect(screen.getByText('software engineer · boise, id · remote')).toBeInTheDocument()
	expect(screen.getByText(/Software Engineer at C1/)).toBeInTheDocument()
})

test('renders four tabs with readme active', () => {
	render(<App />)
	expect(screen.getByRole('button', { name: '▌ readme' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'projects' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'work' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'contact' })).toBeInTheDocument()
})

test('clicking a tab makes it active', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))
	expect(screen.getByRole('button', { name: '▌ projects' })).toBeInTheDocument()
})
```

Note the accessible names: the active tab's label is prefixed `▌ `, and inactive labels are prefixed with two spaces in the prototype, which the accessible-name calculation collapses — so query inactive tabs by their bare label.

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find an element with the text: DAVID FETZER`.

- [ ] **Step 3: Create `src/components/HeaderRow.tsx`**

Strings come from the prototype's markup and are frozen.

```tsx
export function HeaderRow() {
	return (
		<div className="flex flex-none flex-wrap items-stretch gap-[14px]">
			<div className="rounded-[6px] border border-acc2 bg-mantle px-4 py-[10px]">
				<div className="text-[clamp(18px,2.4vw,26px)] font-bold leading-[1.15] tracking-[0.14em] text-acc">
					DAVID FETZER
				</div>
				<div className="mt-[3px] text-[12px] text-subtext0">
					software engineer · boise, id · remote
				</div>
			</div>
			<div className="grid content-center gap-[3px] rounded-[6px] border border-surface0 bg-mantle px-4 py-[10px] text-[12px] text-text">
				<span>
					<span className="text-subtext0">role </span>Software Engineer at C1
				</span>
				<span>
					<span className="text-subtext0">status </span>
					<span className="text-green">● open to interesting conversations</span>
				</span>
				<span>
					<span className="text-subtext0">builds </span>map servers · CLI tools · self-hosted
					platforms · hardware
				</span>
			</div>
		</div>
	)
}
```

- [ ] **Step 4: Create `src/components/TabBar.tsx`**

```tsx
import { TABS, type TabId } from '../content'

export function TabBar({ tab, onSelect }: { tab: TabId; onSelect: (t: TabId) => void }) {
	return (
		<div className="flex flex-none flex-wrap gap-[6px]">
			{TABS.map((t) => {
				const active = t.id === tab
				return (
					<button
						key={t.id}
						type="button"
						onClick={() => onSelect(t.id)}
						className={`rounded-t-[6px] border border-b-0 px-[14px] py-[6px] font-mono text-[12.5px] ${
							active
								? 'border-acc2 bg-surface0 text-acc'
								: 'border-surface0 bg-mantle text-subtext0'
						}`}
					>
						{active ? `▌ ${t.label}` : `  ${t.label}`}
					</button>
				)
			})}
			<span className="min-w-[40px] flex-1 self-end border-b border-surface0" />
		</div>
	)
}
```

- [ ] **Step 5: Wire them into `src/App.tsx`**

Add the imports and a `tab` state, and render `HeaderRow` and `TabBar` above the section list:

```tsx
const [tab, setTab] = useState<TabId>('readme')
const [selected, setSelected] = useState('readme')

const selectTab = useCallback((t: TabId) => {
	setTab(t)
	setSelected(firstSectionOfTab(t).id)
}, [])
```

Render `<HeaderRow />` then `<TabBar tab={tab} onSelect={selectTab} />` inside the body div, before the articles. Import `useCallback` and `useState` from React, `TabId` from `./content`, and `firstSectionOfTab` from `./selectors`. `selected` is unused until Task 5 — keep it.

- [ ] **Step 6: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add the title block, metadata block, and tab bar"
```

---

### Task 5: List pane

**Files:**
- Create: `src/components/ListPane.tsx`
- Modify: `src/App.tsx`, `src/App.test.tsx`

**Interfaces:**
- Consumes: `visibleSections`, `listStatus` from Task 3.
- Produces: `ListPane({ tab, sections, selectedId, focused, status, onSelect, onFocus })` where `sections: Section[]`, `focused: boolean`, `status: string`, `onSelect: (id: string) => void`, `onFocus: () => void`.

- [ ] **Step 1: Write the failing test**

Add to `src/App.test.tsx`:

```tsx
test('the list shows the active tab name and its sections', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	expect(screen.getByText('PROJECTS')).toBeInTheDocument()
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'transfer-it-cli' })).toBeInTheDocument()
	expect(screen.getByText('1/5')).toBeInTheDocument()
})

test('clicking a list row selects it', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))
	await user.click(screen.getByRole('button', { name: 'transfer-it-cli' }))

	expect(screen.getByRole('button', { name: '› transfer-it-cli' })).toBeInTheDocument()
	expect(screen.getByText('2/5')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find an element with the text: PROJECTS`.

- [ ] **Step 3: Create `src/components/ListPane.tsx`**

```tsx
import { useEffect, useRef } from 'react'
import type { Section } from '../content'

type Props = {
	tab: string
	sections: Section[]
	selectedId: string
	focused: boolean
	status: string
	onSelect: (id: string) => void
	onFocus: () => void
}

export function ListPane({
	tab,
	sections,
	selectedId,
	focused,
	status,
	onSelect,
	onFocus
}: Props) {
	const boxRef = useRef<HTMLDivElement>(null)

	// Keep the selection in view as it moves.
	useEffect(() => {
		const box = boxRef.current
		if (!box) return
		const i = sections.findIndex((s) => s.id === selectedId)
		const el = box.children[i] as HTMLElement | undefined
		if (!el) return
		const top = el.offsetTop - box.offsetTop
		const bottom = top + el.offsetHeight
		if (top < box.scrollTop) box.scrollTop = top
		else if (bottom > box.scrollTop + box.clientHeight) box.scrollTop = bottom - box.clientHeight
	}, [sections, selectedId])

	return (
		<div
			className={`flex min-h-0 flex-col overflow-hidden rounded-[6px] border bg-mantle ${
				focused ? 'border-acc2' : 'border-surface0'
			}`}
		>
			<div className="flex-none border-b border-surface0 px-3 py-2 text-[11.5px] tracking-[0.1em] text-subtext0">
				{tab.toUpperCase()}
			</div>
			<div data-vp ref={boxRef} onClick={onFocus} className="flex-1 overflow-y-auto py-[6px]">
				{sections.map((s) => {
					const active = s.id === selectedId
					return (
						<button
							key={s.id}
							type="button"
							onClick={() => onSelect(s.id)}
							className={`block w-full border-0 px-3 py-[5px] text-left font-mono text-[13.5px] transition-colors duration-[130ms] ${
								active ? 'bg-surface1 text-acc' : 'bg-transparent text-subtext1'
							}`}
						>
							{active ? `› ${s.label}` : `  ${s.label}`}
						</button>
					)
				})}
			</div>
			<div className="flex-none border-t border-surface0 px-3 py-[7px] text-[11.5px] text-subtext0">
				{status}
			</div>
		</div>
	)
}
```

- [ ] **Step 4: Wire it into `src/App.tsx`**

Add `focus` state and the derived list, and render the grid:

```tsx
const [focus, setFocus] = useState<'list' | 'viewport'>('list')
const visible = visibleSections(tab, filter, filtering)
const status = listStatus(visible, selected, filter)
```

For now `filter` is `''` and `filtering` is `false` — declare them as state here so Task 8 only has to wire the input:

```tsx
const [filtering, setFiltering] = useState(false)
const [filter, setFilter] = useState('')
```

Replace the placeholder article list with a grid containing `ListPane` on the left; the detail pane arrives in Task 6:

```tsx
<div className="grid min-h-0 flex-1 grid-cols-1 gap-[14px] md:grid-cols-[minmax(0,26ch)_minmax(0,1fr)]">
	<ListPane
		tab={tab}
		sections={visible}
		selectedId={selected}
		focused={focus === 'list'}
		status={status}
		onSelect={setSelected}
		onFocus={() => setFocus('list')}
	/>
</div>
```

`setFiltering` and `setFilter` are unused until Task 8 — that is expected. If ESLint flags them, leave them and note it; Task 8 uses them.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add the list pane with selection and auto-scroll"
```

---

### Task 6: Detail pane, prompt, and the nine articles

**Files:**
- Create: `src/components/DetailPane.tsx`, `src/components/Prompt.tsx`
- Modify: `src/App.tsx`, `src/App.test.tsx`

**Interfaces:**
- Consumes: `promptSegments` from Task 3; `SECTIONS`, `Section` from Task 1.
- Produces: `Prompt({ section }: { section: Section })` and `DetailPane({ section, focused, hydrated, onFocus }: { section: Section; focused: boolean; hydrated: boolean; onFocus: () => void })`.
- Later tasks extend `DetailPane`'s props: Task 7 adds `viewportRef: RefObject<HTMLDivElement | null>` so the keyboard handler can scroll it, and Task 9 adds `direction: number` to pick the slide animation. Both are modifications to this file, not new components.

**This task implements the no-JavaScript strategy**, so read it carefully. `DetailPane` renders **all nine sections** as `<article>` elements. Every article except the selected one carries the `hidden` attribute — but only once the client has hydrated. Before hydration, and forever with JavaScript disabled, all nine render stacked and readable.

The mechanism: `App` sets a `data-hydrated` attribute on the root element (already done in `main.tsx`), and `DetailPane` only applies `hidden` when a `hydrated` prop is true. `App` derives that from a `useEffect` that flips a state flag after mount — so the server render and the first client render agree, and the second client render hides the extra articles.

- [ ] **Step 1: Write the failing test**

Add to `src/App.test.tsx`:

```tsx
test('renders every section as an article', () => {
	const { container } = render(<App />)
	expect(container.querySelectorAll('article')).toHaveLength(9)
})

test('after hydration only the selected article is visible', async () => {
	render(<App />)
	// The mount effect has run by the time render() returns.
	const articles = screen.getAllByRole('article', { hidden: true })
	const visible = articles.filter((a) => !a.hasAttribute('hidden'))
	expect(visible).toHaveLength(1)
	expect(visible[0]).toHaveAttribute('id', 'section-readme')
})

test('the detail pane shows the selected section', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	expect(screen.getByRole('heading', { name: 'mapwright', level: 2 })).toBeVisible()
	expect(screen.getByText('self-hosted map infrastructure · mapwright.io')).toBeVisible()
})

test('the prompt shows the section path and branch', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	expect(screen.getByText('~/site/projects/mapwright')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — the placeholder renders no `<article>` elements with those ids.

- [ ] **Step 3: Create `src/components/Prompt.tsx`**

```tsx
import { promptSegments } from '../selectors'
import type { Section } from '../content'

export function Prompt({ section }: { section: Section }) {
	return (
		<div className="flex flex-none items-center gap-2 overflow-x-auto whitespace-nowrap border-b border-surface0 px-3 py-[7px] text-[12.5px]">
			<span className="flex items-baseline gap-3">
				{promptSegments(section).map((s) => (
					<span key={s.text} style={{ color: s.color }} className="font-bold">
						{s.text}
					</span>
				))}
			</span>
			<span className="inline-flex items-baseline gap-[6px]">
				<span className="text-green">&#xf0d1b;</span>
				<span className="text-acc">❯</span>
			</span>
		</div>
	)
}
```

The two prompt colors that are not palette tokens (`#91d7e3` teal for go, `#8bd5ca` for i2c) come from `promptSegments` as inline styles — they are starship's own language colors, not theme colors, which is why they are data rather than Tailwind classes.

- [ ] **Step 4: Create `src/components/DetailPane.tsx`**

```tsx
import { SECTIONS, type Section } from '../content'
import { Prompt } from './Prompt'

type Props = {
	section: Section
	focused: boolean
	hydrated: boolean
	onFocus: () => void
}

export function DetailPane({ section, focused, hydrated, onFocus }: Props) {
	return (
		<div
			className={`flex min-h-0 flex-col overflow-hidden rounded-[6px] border bg-mantle ${
				focused ? 'border-acc2' : 'border-surface0'
			}`}
		>
			<Prompt section={section} />
			<div
				data-vp
				onClick={onFocus}
				className="flex-1 overflow-y-auto p-[clamp(14px,2vw,22px)]"
			>
				{SECTIONS.map((s) => (
					<article
						key={s.id}
						id={`section-${s.id}`}
						// Before hydration every article renders, so the document is
						// complete without JavaScript. After hydration only the selected
						// one stays visible.
						hidden={hydrated && s.id !== section.id}
					>
						<h2 className="mb-[6px] text-[clamp(20px,2.6vw,30px)] font-bold tracking-[-0.02em] text-acc">
							{s.title}
						</h2>
						<p className="mb-[18px] text-[12px] tracking-[0.06em] text-subtext0">{s.kicker}</p>
						{s.paras.map((p) => (
							<p
								key={p}
								className="mb-[14px] max-w-[74ch] text-pretty text-[14px] leading-[1.7] text-text"
							>
								{p}
							</p>
						))}
						{s.rows.length > 0 && (
							<div className="mt-2 border-t border-surface0">
								{s.rows.map((r) => (
									<div
										key={r.label}
										className="grid grid-cols-[minmax(0,17ch)_minmax(0,1fr)] items-start gap-[14px] border-b border-rule py-[9px]"
									>
										<span className="text-[12.5px] text-lavender">{r.label}</span>
										<span className="text-pretty text-[13.5px] leading-[1.6] text-subtext1">
											{r.body}
										</span>
									</div>
								))}
							</div>
						)}
						{s.links.length > 0 && (
							<div className="mt-[18px] flex flex-wrap gap-[18px] text-[13px]">
								{s.links.map((l) => (
									<a
										key={l.href}
										href={l.href}
										target="_blank"
										rel="noopener noreferrer"
										className="border-b border-surface1 pb-[2px] text-acc no-underline"
									>
										{l.label} ↗
									</a>
								))}
							</div>
						)}
						<span
							aria-hidden="true"
							className="mt-[18px] inline-block h-[1.05em] w-[0.55em] animate-blink bg-acc align-[-0.16em]"
						/>
					</article>
				))}
			</div>
		</div>
	)
}
```

- [ ] **Step 5: Wire it into `src/App.tsx`**

Add the hydration flag and the current section, and render `DetailPane` as the grid's second child:

```tsx
const [hydrated, setHydrated] = useState(false)
useEffect(() => setHydrated(true), [])

const current = SECTIONS.find((s) => s.id === selected) ?? SECTIONS[0]
```

```tsx
<DetailPane
	section={current}
	focused={focus === 'viewport'}
	hydrated={hydrated}
	onFocus={() => setFocus('viewport')}
/>
```

Import `useEffect` from React and `SECTIONS` from `./content`.

- [ ] **Step 6: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS.

- [ ] **Step 7: Confirm the no-JS document is complete**

Run: `pnpm build`
Then:

```bash
grep -o "<article" dist/index.html | wc -l
grep -c "Contributions to Jellyfin and Websurfx." dist/index.html
grep -c "hidden" dist/index.html
```

Expected: 9 articles, the oss copy present, and **no `hidden` attributes** — the server render has `hydrated` false, so nothing is hidden in the prerendered HTML. If `hidden` appears, the flag is being computed during the server render and the no-JS fallback is broken.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add the detail pane, starship prompt, and the nine articles

All sections render as semantic articles so the prerendered document is
complete without JavaScript; hydration hides all but the selected one."
```

---

### Task 7: Keyboard model and focus

**Files:**
- Modify: `src/App.tsx`, `src/App.test.tsx`, `src/components/DetailPane.tsx`

**Interfaces:**
- Consumes: everything from Tasks 3–6.
- Produces: no new exports. `App` gains its `keydown` handler, and `DetailPane` gains a `viewportRef: RefObject<HTMLDivElement | null>` prop that it attaches to its `data-vp` scroller so the handler can scroll it.

- [ ] **Step 1: Write the failing test**

Add to `src/App.test.tsx`:

```tsx
test('j and k move the selection when the list has focus', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('j')
	expect(screen.getByRole('button', { name: '› transfer-it-cli' })).toBeInTheDocument()

	await user.keyboard('k')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
})

test('selection does not wrap past either end', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('k')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()

	await user.keyboard('{Shift>}G{/Shift}')
	expect(screen.getByRole('button', { name: '› open-source' })).toBeInTheDocument()
	await user.keyboard('j')
	expect(screen.getByRole('button', { name: '› open-source' })).toBeInTheDocument()
})

test('g and G jump to the first and last item', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('{Shift>}G{/Shift}')
	expect(screen.getByText('5/5')).toBeInTheDocument()

	await user.keyboard('g')
	expect(screen.getByText('1/5')).toBeInTheDocument()
})

test('h and l move focus between panes', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('l')
	expect(screen.getByText('scroll')).toBeInTheDocument()

	await user.keyboard('h')
	expect(screen.getByText('select')).toBeInTheDocument()
})

test('tab cycles tabs forward and shift+tab backward', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('{Tab}')
	expect(screen.getByRole('button', { name: '▌ projects' })).toBeInTheDocument()

	await user.keyboard('{Shift>}{Tab}{/Shift}')
	expect(screen.getByRole('button', { name: '▌ readme' })).toBeInTheDocument()
})

test('keys are ignored while a modifier is held', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('{Control>}j{/Control}')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
})
```

The `select`/`scroll` assertions depend on the help footer, which arrives in this task — add a minimal footer now (Task 9 styles it fully).

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — no keydown handler exists.

- [ ] **Step 3: Add the keydown handler to `src/App.tsx`**

```tsx
const move = useCallback(
	(delta: number) => {
		const list = visibleSections(tab, filter, filtering)
		if (list.length === 0) return
		const i = Math.max(
			0,
			list.findIndex((s) => s.id === selected)
		)
		const next = list[clampIndex(i, delta, list.length)]
		setSelected(next.id)
		setTab(next.tab)
	},
	[tab, filter, filtering, selected]
)

const scrollViewport = useCallback((px: number) => {
	const vp = viewportRef.current
	if (!vp) return
	vp.scrollTop = Math.max(0, Math.min(vp.scrollTop + px, vp.scrollHeight - vp.clientHeight))
}, [])

const cycleTab = useCallback(
	(dir: number) => {
		const i = TABS.findIndex((t) => t.id === tab)
		selectTab(TABS[(i + dir + TABS.length) % TABS.length].id)
	},
	[tab, selectTab]
)

useEffect(() => {
	const onKey = (event: KeyboardEvent) => {
		if (event.metaKey || event.ctrlKey || event.altKey) return
		const tagName = (event.target as HTMLElement | null)?.tagName?.toLowerCase()
		if (tagName === 'input' || tagName === 'textarea') return

		const onList = focus === 'list'
		const vp = viewportRef.current
		const page = (vp?.clientHeight ?? 400) * 0.85

		switch (event.key) {
			case 'j':
			case 'ArrowDown':
				event.preventDefault()
				onList ? move(1) : scrollViewport(56)
				break
			case 'k':
			case 'ArrowUp':
				event.preventDefault()
				onList ? move(-1) : scrollViewport(-56)
				break
			case 'PageDown':
			case ' ':
				event.preventDefault()
				onList ? move(4) : scrollViewport(page)
				break
			case 'PageUp':
				event.preventDefault()
				onList ? move(-4) : scrollViewport(-page)
				break
			case 'g':
				event.preventDefault()
				onList ? move(-99) : scrollViewport(-1e7)
				break
			case 'G':
				event.preventDefault()
				onList ? move(99) : scrollViewport(1e7)
				break
			case 'Tab':
				event.preventDefault()
				cycleTab(event.shiftKey ? -1 : 1)
				break
			case 'l':
			case 'ArrowRight':
				event.preventDefault()
				setFocus('viewport')
				break
			case 'h':
			case 'ArrowLeft':
				event.preventDefault()
				setFocus('list')
				break
			case 'Enter': {
				const link = current.links[0]
				if (link) window.open(link.href, '_blank', 'noopener')
				break
			}
		}
	}

	window.addEventListener('keydown', onKey)
	return () => window.removeEventListener('keydown', onKey)
}, [focus, move, scrollViewport, cycleTab, current])
```

Add `const viewportRef = useRef<HTMLDivElement>(null)` in `App` and pass it to `DetailPane` as a `viewportRef` prop, which `DetailPane` puts on its `data-vp` scroller. Import `useRef`, `TABS`, `clampIndex`, and `visibleSections`.

Add a minimal help footer inside the body div, below the grid:

```tsx
<div className="flex flex-none flex-wrap items-center gap-[14px] pt-[2px] text-[11.5px] text-subtext0">
	<span>
		<span className="text-lavender">↑/↓ j/k</span> {focus === 'list' ? 'select' : 'scroll'}
	</span>
	<span>
		<span className="text-lavender">h/l ←/→</span> pane
	</span>
	<span>
		<span className="text-lavender">tab</span> next tab
	</span>
	<span>
		<span className="text-lavender">/</span> filter
	</span>
	<span>
		<span className="text-lavender">g/G</span> first/last
	</span>
	<span>
		<span className="text-lavender">enter</span> open link
	</span>
	<span className="ml-auto text-subtext0">bubbletea · lipgloss</span>
</div>
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add the keyboard model and pane focus

j/k act on the focused pane, h/l switch panes, tab cycles tabs, g/G jump,
enter opens the selected section's first link."
```

---

### Task 8: Filter

**Files:**
- Create: `src/components/HelpFooter.tsx` (moves the footer out of `App`)
- Modify: `src/App.tsx`, `src/App.test.tsx`

**Interfaces:**
- Consumes: `visibleSections` from Task 3.
- Produces: `HelpFooter({ focus, filtering, filter, inputRef, onFilterChange, onFilterKeyDown })`.

- [ ] **Step 1: Write the failing test**

Add to `src/App.test.tsx`:

```tsx
test('slash opens the filter and typing narrows the list', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	const input = screen.getByLabelText('Filter sections')
	expect(input).toHaveFocus()

	await user.type(input, 'mapwright')
	expect(screen.getByText('1/1 filtered')).toBeInTheDocument()
})

test('the filter searches every tab, not just the active one', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'stack')
	expect(screen.getByRole('button', { name: '› stack' })).toBeInTheDocument()
})

test('escape cancels the filter and clears it', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'mapwright')
	await user.keyboard('{Escape}')

	expect(screen.queryByLabelText('Filter sections')).not.toBeInTheDocument()
	expect(screen.getByText('1/1')).toBeInTheDocument()
})

test('enter keeps the filter and closes the input', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'mapwright')
	await user.keyboard('{Enter}')

	expect(screen.queryByLabelText('Filter sections')).not.toBeInTheDocument()
})

test('a filter with no match reports no match', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'zzzzqqq')
	expect(screen.getByText('no match')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `Unable to find a label with the text of: Filter sections`.

- [ ] **Step 3: Create `src/components/HelpFooter.tsx`**

Move the footer markup from Task 7 into this component and add the filtering branch:

```tsx
import type { KeyboardEvent, RefObject } from 'react'

type Props = {
	focus: 'list' | 'viewport'
	filtering: boolean
	filter: string
	inputRef: RefObject<HTMLInputElement | null>
	onFilterChange: (value: string) => void
	onFilterKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void
}

const HINTS: [string, string][] = [
	['h/l ←/→', 'pane'],
	['tab', 'next tab'],
	['/', 'filter'],
	['g/G', 'first/last'],
	['enter', 'open link']
]

export function HelpFooter({
	focus,
	filtering,
	filter,
	inputRef,
	onFilterChange,
	onFilterKeyDown
}: Props) {
	return (
		<div className="flex flex-none flex-wrap items-center gap-[14px] pt-[2px] text-[11.5px] text-subtext0">
			{filtering ? (
				<span className="flex flex-1 basis-[220px] items-center gap-2 text-text">
					<span className="text-green">filter&gt;</span>
					<input
						ref={inputRef}
						value={filter}
						onChange={(e) => onFilterChange(e.target.value)}
						onKeyDown={onFilterKeyDown}
						aria-label="Filter sections"
						placeholder="type to filter, esc to cancel"
						className="min-w-0 flex-1 border-0 bg-transparent font-mono text-[12.5px] text-text outline-none"
					/>
				</span>
			) : (
				<span className="flex flex-wrap gap-[14px]">
					<span>
						<span className="text-lavender">↑/↓ j/k</span>{' '}
						{focus === 'list' ? 'select' : 'scroll'}
					</span>
					{HINTS.map(([keys, label]) => (
						<span key={keys}>
							<span className="text-lavender">{keys}</span> {label}
						</span>
					))}
				</span>
			)}
			<span className="ml-auto text-subtext0">bubbletea · lipgloss</span>
		</div>
	)
}
```

- [ ] **Step 4: Wire the filter into `src/App.tsx`**

```tsx
const filterRef = useRef<HTMLInputElement>(null)

const startFilter = useCallback(() => {
	setFiltering(true)
	// The input mounts in this same commit; focus once it exists.
	queueMicrotask(() => filterRef.current?.focus())
}, [])

const onFilterChange = useCallback(
	(value: string) => {
		setFilter(value)
		const list = visibleSections(tab, value, true)
		if (list.length > 0 && !list.some((s) => s.id === selected)) {
			setSelected(list[0].id)
			setTab(list[0].tab)
		}
	},
	[tab, selected]
)

const onFilterKeyDown = useCallback((event: KeyboardEvent<HTMLInputElement>) => {
	if (event.key === 'Escape') {
		event.preventDefault()
		setFiltering(false)
		setFilter('')
	} else if (event.key === 'Enter') {
		event.preventDefault()
		setFiltering(false)
	}
}, [])
```

Import `KeyboardEvent` as a type from React (`import { useCallback, useRef, type KeyboardEvent } from 'react'`) — writing `React.KeyboardEvent` would require a React namespace import that this file does not have.

Add two cases to the keydown switch, before the navigation cases:

```tsx
case '/':
	event.preventDefault()
	startFilter()
	break
case 'Escape':
	event.preventDefault()
	setFiltering(false)
	setFilter('')
	break
```

Add `startFilter` to the effect's dependency array. Replace the inline footer with `<HelpFooter ... />`.

`queueMicrotask` rather than `requestAnimationFrame`: React's commit is flushed on a microtask enqueued before this one, so FIFO ordering guarantees the input exists. `requestAnimationFrame` needs an animation-frame tick that `user.keyboard()` never awaits, which fails deterministically in tests — this exact trap cost a fix round on the previous site.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `pnpm test:unit`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add the filter

Searches every section while the input is open, so selecting a match moves
across tabs. Esc cancels and clears, enter keeps the filter."
```

---

### Task 9: Motion and the responsive breakpoint

**Files:**
- Modify: `src/hooks/prefersReducedMotion.ts`, `src/hooks/prefersReducedMotion.test.ts`, `src/components/DetailPane.tsx`, `src/components/ListPane.tsx`, `src/App.tsx`

**Interfaces:**
- Consumes: the Tailwind keyframes from Task 2.
- Produces: `prefersReducedMotion(): boolean` added to `src/hooks/prefersReducedMotion.ts`.
- Extends two existing components: `DetailPane` gains `direction: number` (which slide to play) and `ListPane` gains `tabEpoch: number` (a counter `App` increments on every tab change, used to retrigger the staggered row animation). `App` owns both values.

- [ ] **Step 1: Write the failing test**

Add to `src/hooks/prefersReducedMotion.test.ts`:

```ts
import { prefersReducedMotion } from './prefersReducedMotion'

test('reports the reduced-motion preference', () => {
	const stub = (matches: boolean) =>
		vi.stubGlobal('matchMedia', () => ({ matches, media: '', addEventListener() {}, removeEventListener() {} }))

	stub(true)
	expect(prefersReducedMotion()).toBe(true)

	stub(false)
	expect(prefersReducedMotion()).toBe(false)

	vi.unstubAllGlobals()
})

test('reports false when matchMedia is unavailable', () => {
	vi.stubGlobal('matchMedia', undefined)
	expect(prefersReducedMotion()).toBe(false)
	vi.unstubAllGlobals()
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `pnpm test:unit`
Expected: FAIL — `prefersReducedMotion is not a function`.

- [ ] **Step 3: Extend the hook module**

Add to `src/hooks/prefersReducedMotion.ts`, keeping `scrollBehavior` as it is:

```ts
/**
 * Whether the visitor asked for reduced motion. Safe during SSR. Checked in
 * JavaScript rather than only in CSS because JS-driven animation ignores the
 * CSS media query.
 */
export function prefersReducedMotion(): boolean {
	if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
	return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}
```

- [ ] **Step 4: Animate the detail pane on selection change**

In `DetailPane`, accept a `direction: number` prop and apply the slide animation to the article wrapper when the selection changes:

```tsx
const paneRef = useRef<HTMLDivElement>(null)

useEffect(() => {
	const el = paneRef.current
	if (!el || prefersReducedMotion()) return
	el.style.animation = 'none'
	void el.offsetWidth // force reflow so the animation restarts
	el.style.animation =
		direction >= 0
			? 'slidedown 170ms cubic-bezier(0.22, 1, 0.36, 1)'
			: 'slideup 170ms cubic-bezier(0.22, 1, 0.36, 1)'
}, [section.id, direction])
```

Wrap the articles in `<div ref={paneRef}>`. `App` tracks `direction` alongside `selected`: `+1` when the new index is greater than the old, `-1` otherwise, `+1` on tab switches.

Also reset the viewport scroll to the top whenever the selection changes:

```tsx
useEffect(() => {
	if (viewportRef.current) viewportRef.current.scrollTop = 0
}, [section.id])
```

- [ ] **Step 5: Stagger the list rows on tab change**

In `ListPane`, accept a `tabEpoch: number` prop that `App` increments on every tab change, and run:

```tsx
useEffect(() => {
	const box = boxRef.current
	if (!box || prefersReducedMotion()) return
	Array.from(box.children).forEach((child, i) => {
		const el = child as HTMLElement
		el.style.animation = 'none'
		void el.offsetWidth
		el.style.animation = `slidex 190ms cubic-bezier(0.22, 1, 0.36, 1) ${i * 24}ms both`
	})
}, [tabEpoch])
```

- [ ] **Step 6: Confirm the responsive breakpoint**

The grid from Task 5 already uses `grid-cols-1 md:grid-cols-[minmax(0,26ch)_minmax(0,1fr)]`. Tailwind's `md` is 768px; the spec calls for 700px. Add a custom screen to `tailwind.config.ts` rather than accepting the wrong number:

```ts
screens: {
	tui: '700px'
}
```

and change the grid classes to `grid-cols-1 tui:grid-cols-[minmax(0,26ch)_minmax(0,1fr)]`. Below the breakpoint the list is capped: add `max-h-[30vh] tui:max-h-none` to `ListPane`'s root.

- [ ] **Step 7: Verify**

Run: `pnpm lint`, `pnpm build`, `pnpm test`
Expected: all clean.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add slide motion, row stagger, and the 700px breakpoint

All motion is gated on a JS matchMedia check, not only the CSS media query,
because JS-driven animation ignores the CSS property."
```

---

### Task 10: End-to-end suite

**Files:**
- Create: `tests/tui.spec.ts`
- Modify: `tests/no-js.spec.ts`

**Interfaces:**
- Consumes: the complete app.
- Produces: nothing downstream.

- [ ] **Step 1: Write the suite**

Create `tests/tui.spec.ts`:

```ts
import { expect, test, type Page } from '@playwright/test'

/** The document is prerendered, so it paints before React attaches listeners. */
async function gotoHydrated(page: Page) {
	await page.goto('/')
	await page.waitForSelector('html[data-hydrated="true"]')
}

test('head metadata is intact', async ({ page }) => {
	await page.goto('/')
	await expect(page).toHaveTitle('fetzer — TUI')
	await expect(page.locator('meta[property="og:title"]')).toHaveAttribute(
		'content',
		'David Fetzer — Software Engineer'
	)
})

test('the page itself never scrolls', async ({ page }) => {
	await gotoHydrated(page)
	const { scrollHeight, clientHeight } = await page.evaluate(() => ({
		scrollHeight: document.documentElement.scrollHeight,
		clientHeight: document.documentElement.clientHeight
	}))
	expect(scrollHeight).toBeLessThanOrEqual(clientHeight + 1)
})

test('the detail pane scrolls independently of the list', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('Tab') // projects
	await page.keyboard.press('l') // focus viewport
	await page.keyboard.press('G')

	const scrolled = await page.evaluate(() => {
		const panes = document.querySelectorAll('[data-vp]')
		return (panes[1] as HTMLElement).scrollTop
	})
	expect(scrolled).toBeGreaterThan(0)
})

test('j and k move the selection, and the status follows', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('Tab')
	await expect(page.getByText('1/5')).toBeVisible()

	await page.keyboard.press('j')
	await expect(page.getByText('2/5')).toBeVisible()
	await expect(page.getByRole('button', { name: '› transfer-it-cli' })).toBeVisible()
})

test('h and l move focus, and the help hint follows', async ({ page }) => {
	await gotoHydrated(page)
	await expect(page.getByText('select')).toBeVisible()
	await page.keyboard.press('l')
	await expect(page.getByText('scroll')).toBeVisible()
	await page.keyboard.press('h')
	await expect(page.getByText('select')).toBeVisible()
})

test('the filter opens, narrows, and cancels', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('/')
	await page.getByLabel('Filter sections').fill('mapwright')
	await expect(page.getByText('1/1 filtered')).toBeVisible()

	await page.keyboard.press('Escape')
	await expect(page.getByLabel('Filter sections')).toHaveCount(0)
})

test('enter opens the selected section link', async ({ page, context }) => {
	await gotoHydrated(page)
	await page.keyboard.press('Tab') // projects, mapwright selected

	const [popup] = await Promise.all([
		context.waitForEvent('page'),
		page.keyboard.press('Enter')
	])
	expect(popup.url()).toContain('mapwright.io')
	await popup.close()
})

test('only one article is visible after hydration', async ({ page }) => {
	await gotoHydrated(page)
	const visible = await page.evaluate(
		() => Array.from(document.querySelectorAll('article')).filter((a) => !a.hasAttribute('hidden')).length
	)
	expect(visible).toBe(1)
})

test('no horizontal overflow at a 375px viewport', async ({ page }) => {
	await page.setViewportSize({ width: 375, height: 800 })
	await gotoHydrated(page)
	const { scrollWidth, clientWidth } = await page.evaluate(() => ({
		scrollWidth: document.documentElement.scrollWidth,
		clientWidth: document.documentElement.clientWidth
	}))
	expect(scrollWidth).toBeLessThanOrEqual(clientWidth)
})
```

- [ ] **Step 2: Run the suite**

Run: `pnpm test:integration`
Expected: all pass, plus the no-JS spec from Task 1.

If the "page never scrolls" test fails, the desk is taller than the viewport — check that the root is `h-screen overflow-hidden` and that no child forces height.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "test: add the end-to-end suite for the TUI front end"
```

---

### Task 11: README and final verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Rewrite `README.md`**

Replace the body with a description of the new site. Keep the commands table accurate against `package.json`, note that the design handoff lives outside the repository, and add a line stating that `content.json` is the shared content source that the SSH app will also read.

- [ ] **Step 2: Full verification**

Run: `pnpm lint` — expect clean.
Run: `pnpm build` — expect type-check clean and the prerender to inject markup.
Run: `pnpm test` — expect all unit and Playwright tests passing.

Confirm the prerendered document is complete:

```bash
grep -o "<article" dist/index.html | wc -l   # expect 9
grep -c "hidden" dist/index.html             # expect 0
```

Confirm no third-party requests:

```bash
grep -rn "jsdelivr\|googleapis\|gstatic" index.html src/ dist/index.html || echo clean
```

- [ ] **Step 3: Manual check against the prototype**

Serve the reference and compare side by side:

```bash
cd ~/Downloads/tuiSite/design_handoff_tui_ssh && npx serve .
```

Walk the four tabs, every key in the table, both panes' scrolling, and the 700px collapse. Report anything that differs from the prototype.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "docs: rewrite the README for the TUI site"
```

---

## Done

The web front end is complete and reviewable. Sub-project 2 — the Go SSH server reading the same `content.json` — gets its own spec and plan.

Do not deploy. The prompt is explicit that deployment is the owner's call, and `main` currently serves the previous site.

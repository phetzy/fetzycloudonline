import { useCallback, useEffect, useRef, useState } from 'react'
import { CopyButton } from './components/CopyButton'
import { Overlay, type OverlayRow } from './components/Overlay'
import { Section } from './components/Section'
import { StatusBar } from './components/StatusBar'
import { DetailRow, LabelGrid, LabelRow, SubsystemRow } from './components/rows'
import {
	EMAIL,
	HELP,
	PLATFORM_DETAILS,
	SECTIONS,
	STACK,
	SUBSYSTEMS,
	TRANSFER_DETAILS
} from './content'
import { useManualNav } from './hooks/useManualNav'
import { useScrollPercent } from './hooks/useScrollPercent'
import { useSearch } from './hooks/useSearch'

export function Manual() {
	const contentRef = useRef<HTMLElement>(null)
	const searchRef = useRef<HTMLInputElement>(null)
	const percent = useScrollPercent()
	const { idx, goTo } = useManualNav()

	const [searching, setSearching] = useState(false)
	const search = useSearch(contentRef)
	const [overlay, setOverlay] = useState<'help' | 'toc' | null>(null)

	const tocRows: OverlayRow[] = SECTIONS.map((section, n) => ({
		key: String(n + 1).padStart(2, '0'),
		label: section.label
	}))
	const overlayRows = overlay === 'toc' ? tocRows : HELP
	const overlayTitle = overlay === 'toc' ? 'TABLE OF CONTENTS' : 'KEYS'

	const jump = useCallback(
		(i: number) => {
			setOverlay(null)
			goTo(i)
		},
		[goTo]
	)

	const startSearch = useCallback(() => {
		setSearching(true)
		setOverlay(null)
		// The input mounts in this same commit; focus after paint.
		queueMicrotask(() => searchRef.current?.focus())
	}, [])

	useEffect(() => {
		const onKey = (event: KeyboardEvent) => {
			if (event.metaKey || event.ctrlKey || event.altKey) return
			const tag = (event.target as HTMLElement | null)?.tagName?.toLowerCase()
			if (tag === 'input' || tag === 'textarea') return

			switch (event.key) {
				case 'Escape':
					event.preventDefault()
					setSearching(false)
					search.clear()
					setOverlay(null)
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
				case 'j':
					event.preventDefault()
					jump(idx + 1)
					break
				case 'k':
					event.preventDefault()
					jump(idx - 1)
					break
				case 'g':
					event.preventDefault()
					jump(0)
					break
				case 'G':
				case 'q':
					event.preventDefault()
					jump(SECTIONS.length - 1)
					break
				case 't':
					event.preventDefault()
					setOverlay((current) => (current === 'toc' ? null : 'toc'))
					break
				case '?':
					event.preventDefault()
					setOverlay((current) => (current === 'help' ? null : 'help'))
					break
			}
		}

		window.addEventListener('keydown', onKey)
		return () => window.removeEventListener('keydown', onKey)
	}, [idx, jump, search, startSearch])

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
							<span className="text-muted">--location</span>{' '}
							<span className="text-sky">boise-id</span>] [
							<span className="text-muted">--remote</span>]
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
							infrastructure. Projects are documented below; employment is summarized under HISTORY.
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
							A single static Go binary for MEGA&apos;s transfer.it, shipping as both an interactive
							TUI and a scriptable CLI. Built for machines without a browser — servers, embedded
							boxes, SSH-only sessions. Zero runtime dependencies.
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
							A Raspberry Pi add-on board: a custom PCB designed in KiCad around an ATECC608B-TNGTLS
							secure element.
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
								<CopyButton value={EMAIL} />
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

			{overlay && (
				<Overlay title={overlayTitle} rows={overlayRows} onClose={() => setOverlay(null)} />
			)}

			<StatusBar
				current={SECTIONS[idx].label}
				position={`${idx + 1}/${SECTIONS.length}`}
				percent={percent}
				searching={searching}
				query={search.query}
				matchLabel={search.matchLabel}
				searchRef={searchRef}
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
				onPrev={() => jump(idx - 1)}
				onNext={() => jump(idx + 1)}
				onToc={() => setOverlay((current) => (current === 'toc' ? null : 'toc'))}
				onFind={startSearch}
				onHelp={() => setOverlay((current) => (current === 'help' ? null : 'help'))}
			/>
		</div>
	)
}

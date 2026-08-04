import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react'
import { DetailPane } from './components/DetailPane'
import { HeaderRow } from './components/HeaderRow'
import { ListPane } from './components/ListPane'
import { TabBar } from './components/TabBar'
import { TitleBar } from './components/TitleBar'
import { SECTIONS, TABS, type TabId } from './content'
import { useTerminalDims } from './hooks/useTerminalDims'
import { clampIndex, firstSectionOfTab, listStatus, visibleSections } from './selectors'

export function App() {
	const dims = useTerminalDims()
	const [tab, setTab] = useState<TabId>('readme')
	const [selected, setSelected] = useState('readme')
	const [focus, setFocus] = useState<'list' | 'viewport'>('list')
	const viewportRef = useRef<HTMLDivElement>(null)
	// False during the server render and the hydrating client render, true
	// afterwards. Both renders must agree or hydration breaks, which is why this
	// is not plain state set in an effect.
	const hydrated = useSyncExternalStore(
		() => () => {},
		() => true,
		() => false
	)
	// Become state in Task 8, when the filter input lands.
	const filter = ''
	const filtering = false

	const selectTab = useCallback((t: TabId) => {
		setTab(t)
		setSelected(firstSectionOfTab(t).id)
	}, [])

	const visible = visibleSections(tab, filter, filtering)
	const status = listStatus(visible, selected, filter)
	const current = SECTIONS.find((s) => s.id === selected) ?? SECTIONS[0]

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
					if (onList) move(1)
					else scrollViewport(56)
					break
				case 'k':
				case 'ArrowUp':
					event.preventDefault()
					if (onList) move(-1)
					else scrollViewport(-56)
					break
				case 'PageDown':
				case ' ':
					event.preventDefault()
					if (onList) move(4)
					else scrollViewport(page)
					break
				case 'PageUp':
					event.preventDefault()
					if (onList) move(-4)
					else scrollViewport(-page)
					break
				case 'g':
					event.preventDefault()
					if (onList) move(-99)
					else scrollViewport(-1e7)
					break
				case 'G':
					event.preventDefault()
					if (onList) move(99)
					else scrollViewport(1e7)
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

	return (
		<div className="h-screen overflow-hidden bg-crust p-[clamp(8px,2.4vw,32px)] font-mono text-text">
			<div className="mx-auto flex h-full max-w-[1220px] flex-col overflow-hidden rounded-[10px] border border-surface0 bg-base">
				<TitleBar dims={dims} />
				<div className="flex min-h-0 flex-1 flex-col gap-3 p-[clamp(12px,2vw,20px)]">
					<HeaderRow />
					<TabBar tab={tab} onSelect={selectTab} />
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
						<DetailPane
							section={current}
							focused={focus === 'viewport'}
							hydrated={hydrated}
							onFocus={() => setFocus('viewport')}
							viewportRef={viewportRef}
						/>
					</div>
					<div className="flex flex-none flex-wrap items-center gap-[14px] pt-[2px] text-[11.5px] text-subtext0">
						<span>
							<span className="text-lavender">↑/↓ j/k</span>{' '}
							{focus === 'list' ? 'select' : 'scroll'}
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
				</div>
			</div>
		</div>
	)
}

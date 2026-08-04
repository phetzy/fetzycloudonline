import {
	useCallback,
	useEffect,
	useRef,
	useState,
	useSyncExternalStore,
	type KeyboardEvent
} from 'react'
import { DetailPane } from './components/DetailPane'
import { HeaderRow } from './components/HeaderRow'
import { HelpFooter } from './components/HelpFooter'
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
	const [direction, setDirection] = useState(1)
	const [tabEpoch, setTabEpoch] = useState(0)
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
	const [filter, setFilter] = useState('')
	const [filtering, setFiltering] = useState(false)
	const filterRef = useRef<HTMLInputElement>(null)

	const selectTab = useCallback((t: TabId) => {
		setTab(t)
		setSelected(firstSectionOfTab(t).id)
		setDirection(1)
		setTabEpoch((e) => e + 1)
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
			const nextIndex = clampIndex(i, delta, list.length)
			const next = list[nextIndex]
			setDirection(nextIndex > i ? 1 : -1)
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

	useEffect(() => {
		const onKey = (event: globalThis.KeyboardEvent) => {
			if (event.metaKey || event.ctrlKey || event.altKey) return
			const tagName = (event.target as HTMLElement | null)?.tagName?.toLowerCase()
			if (tagName === 'input' || tagName === 'textarea') return

			const onList = focus === 'list'
			const vp = viewportRef.current
			const page = (vp?.clientHeight ?? 400) * 0.85

			switch (event.key) {
				case '/':
					event.preventDefault()
					startFilter()
					break
				case 'Escape':
					event.preventDefault()
					setFiltering(false)
					setFilter('')
					break
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
	}, [focus, move, scrollViewport, cycleTab, current, startFilter])

	return (
		<div className="h-screen overflow-hidden bg-crust p-[clamp(8px,2.4vw,32px)] font-mono text-text">
			<div className="mx-auto flex h-full max-w-[1220px] flex-col overflow-hidden rounded-[10px] border border-surface0 bg-base">
				<TitleBar dims={dims} />
				<div className="flex min-h-0 flex-1 flex-col gap-3 p-[clamp(12px,2vw,20px)]">
					<HeaderRow />
					<TabBar tab={tab} onSelect={selectTab} />
					<div className="grid min-h-0 flex-1 grid-cols-1 gap-[14px] tui:grid-cols-[minmax(0,26ch)_minmax(0,1fr)]">
						<ListPane
							tab={tab}
							sections={visible}
							selectedId={selected}
							focused={focus === 'list'}
							status={status}
							tabEpoch={tabEpoch}
							onSelect={setSelected}
							onFocus={() => setFocus('list')}
						/>
						<DetailPane
							section={current}
							focused={focus === 'viewport'}
							hydrated={hydrated}
							direction={direction}
							onFocus={() => setFocus('viewport')}
							viewportRef={viewportRef}
						/>
					</div>
					<HelpFooter
						focus={focus}
						filtering={filtering}
						filter={filter}
						inputRef={filterRef}
						onFilterChange={onFilterChange}
						onFilterKeyDown={onFilterKeyDown}
					/>
				</div>
			</div>
		</div>
	)
}

import { useCallback, useState, useSyncExternalStore } from 'react'
import { DetailPane } from './components/DetailPane'
import { HeaderRow } from './components/HeaderRow'
import { ListPane } from './components/ListPane'
import { TabBar } from './components/TabBar'
import { TitleBar } from './components/TitleBar'
import { SECTIONS, type TabId } from './content'
import { useTerminalDims } from './hooks/useTerminalDims'
import { firstSectionOfTab, listStatus, visibleSections } from './selectors'

export function App() {
	const dims = useTerminalDims()
	const [tab, setTab] = useState<TabId>('readme')
	const [selected, setSelected] = useState('readme')
	const [focus, setFocus] = useState<'list' | 'viewport'>('list')
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
						/>
					</div>
				</div>
			</div>
		</div>
	)
}

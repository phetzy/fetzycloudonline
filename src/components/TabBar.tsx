import { TABS, type TabId } from '../content'

export function TabBar({ tab, onSelect }: { tab: TabId; onSelect: (t: TabId) => void }) {
	return (
		<div role="tablist" className="flex flex-none flex-wrap gap-[6px]">
			{TABS.map((t) => {
				const active = t.id === tab
				return (
					<button
						key={t.id}
						type="button"
						role="tab"
						aria-selected={active}
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

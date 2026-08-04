import { useEffect, useRef } from 'react'
import type { Section } from '../content'
import { prefersReducedMotion } from '../hooks/prefersReducedMotion'

type Props = {
	tab: string
	sections: Section[]
	selectedId: string
	focused: boolean
	status: string
	tabEpoch: number
	onSelect: (id: string) => void
	onFocus: () => void
}

export function ListPane({
	tab,
	sections,
	selectedId,
	focused,
	status,
	tabEpoch,
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

	// Stagger the row entrance animation whenever the active tab changes.
	useEffect(() => {
		const box = boxRef.current
		if (!box || prefersReducedMotion()) return
		Array.from(box.children).forEach((child, i) => {
			const el = child as HTMLElement
			el.style.animation = 'none'
			void el.offsetWidth // force reflow so the animation restarts
			el.style.animation = `slidex 190ms cubic-bezier(0.22, 1, 0.36, 1) ${i * 24}ms both`
		})
	}, [tabEpoch])

	return (
		<div
			className={`flex max-h-[30vh] min-h-0 flex-col overflow-hidden rounded-[6px] border bg-mantle tui:max-h-none ${
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

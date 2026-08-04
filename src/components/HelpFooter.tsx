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
						<span className="text-lavender">↑/↓ j/k</span> {focus === 'list' ? 'select' : 'scroll'}
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

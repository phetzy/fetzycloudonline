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
						<span className="text-dim">·</span> <span data-testid="position">{position}</span>{' '}
						<span className="text-dim">·</span> <span data-testid="percent">{percent}</span>
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

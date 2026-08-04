type Props = {
	dims: string
	minimized: boolean
	onMinimize: () => void
	onRestore: () => void
}

export function TitleBar({ dims, minimized, onMinimize, onRestore }: Props) {
	const light = 'h-[11px] w-[11px] rounded-full transition-[filter] duration-[120ms] ease-out'
	const live = 'cursor-pointer hover:brightness-125'
	const inert = 'pointer-events-none'

	return (
		<div className="flex flex-none items-center gap-3 border-b border-surface0 bg-mantle px-[14px] py-[10px]">
			<span className="flex gap-[7px]">
				<span className={`${light} bg-red`} aria-hidden="true" />
				<button
					type="button"
					onClick={onMinimize}
					disabled={minimized}
					title="Minimize the terminal window"
					aria-label="Minimize the terminal window"
					className={`${light} bg-yellow ${minimized ? inert : live}`}
				/>
				<button
					type="button"
					onClick={onRestore}
					disabled={!minimized}
					title="Restore the terminal window"
					aria-label="Restore the terminal window"
					className={`${light} bg-green ${minimized ? live : inert}`}
				/>
			</span>
			<span className="flex-1 truncate text-center text-[12px] text-subtext0">
				fetzer@boise: ~/site — go run ./cmd/fetzer
			</span>
			<span className="text-[11.5px] text-subtext0">{dims}</span>
		</div>
	)
}

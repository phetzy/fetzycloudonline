export type OverlayRow = { key: string; label: string }

export function Overlay({
	title,
	rows,
	onClose
}: {
	title: string
	rows: OverlayRow[]
	onClose: () => void
}) {
	return (
		<div
			role="dialog"
			aria-modal="true"
			aria-label={title}
			onClick={onClose}
			className="fixed inset-0 z-[60] grid place-items-center bg-[rgba(8,8,9,0.86)] p-5"
		>
			<div
				onClick={(event) => event.stopPropagation()}
				className="w-full max-w-[60ch] border border-rule bg-panel px-6 py-[22px]"
			>
				<p className="mb-4 text-[12.5px] tracking-[0.16em] text-ph">{title}</p>
				{rows.map((row) => (
					<div
						key={row.key}
						className="grid grid-cols-[minmax(0,12ch)_minmax(0,1fr)] gap-x-4 py-[5px]"
					>
						<span className="text-bright">{row.key}</span>
						<span className="text-muted">{row.label}</span>
					</div>
				))}
				<p className="mt-4 text-[12.5px] text-dim">esc to close</p>
			</div>
		</div>
	)
}

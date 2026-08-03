import type { ReactNode } from 'react'
import type { Detail, Subsystem } from '../content'

export function LabelGrid({
	variant = 'wide',
	children
}: {
	variant?: 'wide' | 'narrow'
	children: ReactNode
}) {
	const cols =
		variant === 'narrow'
			? 'grid-cols-[minmax(0,14ch)_minmax(0,1fr)] gap-y-[4px]'
			: 'grid-cols-[minmax(0,16ch)_minmax(0,1fr)] gap-y-[6px]'
	return <div className={`grid gap-x-4 text-muted ${cols}`}>{children}</div>
}

export function LabelRow({ label, children }: { label: string; children: ReactNode }) {
	return (
		<>
			<span className="text-dim">{label}</span>
			<span>{children}</span>
		</>
	)
}

export function DetailRow({ label, body }: Detail) {
	return (
		<div className="grid grid-cols-[minmax(0,16ch)_minmax(0,1fr)] items-start gap-x-4 border-t border-rule-faint py-2">
			<h3 className="text-[14.5px] font-medium text-sky">{label}</h3>
			<p className="text-pretty text-muted">{body}</p>
		</div>
	)
}

export function SubsystemRow({ n, title, body }: Subsystem) {
	return (
		<div className="grid grid-cols-[4ch_minmax(0,1fr)] items-baseline gap-x-3 py-[5px] sm:grid-cols-[4ch_minmax(0,26ch)_minmax(0,1fr)]">
			<span className="text-dim">{n}</span>
			<span className="text-bright">{title}</span>
			<span className="col-start-2 text-pretty text-faint sm:col-start-3">{body}</span>
		</div>
	)
}

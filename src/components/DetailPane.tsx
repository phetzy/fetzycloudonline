import { useEffect, useRef, type RefObject } from 'react'
import { SECTIONS, type Section } from '../content'
import { prefersReducedMotion } from '../hooks/prefersReducedMotion'
import { Prompt } from './Prompt'

type Props = {
	section: Section
	focused: boolean
	hydrated: boolean
	direction: number
	onFocus: () => void
	viewportRef: RefObject<HTMLDivElement | null>
}

export function DetailPane({ section, focused, hydrated, direction, onFocus, viewportRef }: Props) {
	const paneRef = useRef<HTMLDivElement>(null)

	useEffect(() => {
		const el = paneRef.current
		if (!el || prefersReducedMotion()) return
		el.style.animation = 'none'
		void el.offsetWidth // force reflow so the animation restarts
		el.style.animation =
			direction >= 0
				? 'slidedown 170ms cubic-bezier(0.22, 1, 0.36, 1)'
				: 'slideup 170ms cubic-bezier(0.22, 1, 0.36, 1)'
	}, [section.id, direction])

	useEffect(() => {
		if (viewportRef.current) viewportRef.current.scrollTop = 0
	}, [section.id, viewportRef])

	return (
		<div
			className={`flex min-h-0 flex-col overflow-hidden rounded-[6px] border bg-mantle ${
				focused ? 'border-acc2' : 'border-surface0'
			}`}
		>
			<Prompt section={section} />
			<div
				data-vp
				data-testid="detail-viewport"
				ref={viewportRef}
				tabIndex={-1}
				onClick={onFocus}
				className="flex-1 overflow-y-auto p-[clamp(14px,2vw,22px)] outline-none"
			>
				<div ref={paneRef}>
					{SECTIONS.map((s) => (
						<article
							key={s.id}
							id={`section-${s.id}`}
							// Before hydration every article renders, so the document is
							// complete without JavaScript. After hydration only the selected
							// one stays visible.
							hidden={hydrated && s.id !== section.id}
						>
							<h2 className="mb-[6px] text-[clamp(20px,2.6vw,30px)] font-bold tracking-[-0.02em] text-acc">
								{s.title}
							</h2>
							<p className="mb-[18px] text-[12px] tracking-[0.06em] text-subtext0">{s.kicker}</p>
							{s.paras.map((p) => (
								<p
									key={p}
									className="mb-[14px] max-w-[74ch] text-pretty text-[14px] leading-[1.7] text-text"
								>
									{p}
								</p>
							))}
							{s.rows.length > 0 && (
								<div className="mt-2 border-t border-surface0">
									{s.rows.map((r) => (
										<div
											key={r.label}
											className="grid grid-cols-[minmax(0,17ch)_minmax(0,1fr)] items-start gap-[14px] border-b border-rule py-[9px]"
										>
											<span className="text-[12.5px] text-lavender">{r.label}</span>
											<span className="text-pretty text-[13.5px] leading-[1.6] text-subtext1">
												{r.body}
											</span>
										</div>
									))}
								</div>
							)}
							{s.links.length > 0 && (
								<div className="mt-[18px] flex flex-wrap gap-[18px] text-[13px]">
									{s.links.map((l) => (
										<a
											key={l.href}
											href={l.href}
											target="_blank"
											rel="noopener noreferrer"
											className="border-b border-surface1 pb-[2px] text-acc no-underline"
										>
											{l.label} ↗
										</a>
									))}
								</div>
							)}
							<span
								aria-hidden="true"
								className="mt-[18px] inline-block h-[1.05em] w-[0.55em] animate-blink bg-acc align-[-0.16em]"
							/>
						</article>
					))}
				</div>
			</div>
		</div>
	)
}

import type { ReactNode } from 'react'

type Props = {
	id: string
	heading: string
	children: ReactNode
	/** Only the last section uses a shorter bottom pad. */
	tight?: boolean
}

export function Section({ id, heading, children, tight = false }: Props) {
	return (
		<section id={id} className={`scroll-mt-[24px] ${tight ? 'pb-[20px]' : 'pb-[34px]'}`}>
			<h2 className="mb-[14px] text-[14.5px] font-semibold tracking-[0.16em] text-ph">{heading}</h2>
			<div className="pl-[clamp(16px,4vw,44px)]">{children}</div>
		</section>
	)
}

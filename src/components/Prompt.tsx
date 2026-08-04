import { promptSegments } from '../selectors'
import type { Section } from '../content'

export function Prompt({ section }: { section: Section }) {
	return (
		<div className="flex flex-none items-center gap-2 overflow-x-auto whitespace-nowrap border-b border-surface0 px-3 py-[7px] text-[12.5px]">
			<span className="flex items-baseline gap-3">
				{promptSegments(section).map((s) => (
					<span key={s.text} style={{ color: s.color }} className="font-bold">
						{s.text}
					</span>
				))}
			</span>
			<span className="inline-flex items-baseline gap-[6px]">
				<span className="text-green">&#xf0d1b;</span>
				<span className="text-acc">❯</span>
			</span>
		</div>
	)
}

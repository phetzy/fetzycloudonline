import { SECTIONS } from '../content'

// Every string in this header except `builds` comes from content.json, which
// is the single source of truth shared with the SSH build in cmd/fetzer.
// These used to be hardcoded here, and drifted: the status line read
// "open to interesting conversations" while content.json said
// "Open to interesting conversations", so the web and SSH builds disagreed.
// Deriving them is what stops that happening again.
const README = SECTIONS.find((s) => s.id === 'readme')!

const row = (label: string): string => README.rows.find((r) => r.label === label)?.body ?? ''

// The meta card shows the role without its date clause — the dates appear in
// the metadata table lower down the page, and repeating them here unbalances
// the card against the name card beside it. This is a substring of the
// content.json value, not a second copy of the copy. cmd/fetzer's shortRole
// trims at the same boundary so both builds show the same text.
const shortRole = (r: string): string => (r.includes(',') ? r.slice(0, r.indexOf(',')) : r)

export function HeaderRow() {
	return (
		<div data-testid="header-row" className="flex flex-none flex-wrap items-stretch gap-[14px]">
			<div className="rounded-[6px] border border-acc2 bg-mantle px-4 py-[10px]">
				{/* uppercase is presentation, so it lives in CSS rather than in the
				    content: content.json carries "David Fetzer". */}
				<h1 className="text-[clamp(18px,2.4vw,26px)] font-bold uppercase leading-[1.15] tracking-[0.14em] text-acc">
					{README.title}
				</h1>
				<div className="mt-[3px] text-[12px] text-subtext0">{README.kicker}</div>
			</div>
			<div className="grid content-center gap-[3px] rounded-[6px] border border-surface0 bg-mantle px-4 py-[10px] text-[12px] text-text">
				<span>
					<span className="text-subtext0">role </span>
					{shortRole(row('role'))}
				</span>
				<span>
					<span className="text-subtext0">status </span>
					<span className="text-green">● {row('status')}</span>
				</span>
				{/* `builds` has no content.json field. It is deliberately literal here,
				    and cmd/fetzer takes it verbatim from this file for the same reason. */}
				<span>
					<span className="text-subtext0">builds </span>map servers · CLI tools · self-hosted
					platforms · hardware
				</span>
			</div>
		</div>
	)
}

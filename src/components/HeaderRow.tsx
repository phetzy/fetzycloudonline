export function HeaderRow() {
	return (
		<div data-testid="header-row" className="flex flex-none flex-wrap items-stretch gap-[14px]">
			<div className="rounded-[6px] border border-acc2 bg-mantle px-4 py-[10px]">
				<h1 className="text-[clamp(18px,2.4vw,26px)] font-bold leading-[1.15] tracking-[0.14em] text-acc">
					DAVID FETZER
				</h1>
				<div className="mt-[3px] text-[12px] text-subtext0">
					software engineer · boise, id · remote
				</div>
			</div>
			<div className="grid content-center gap-[3px] rounded-[6px] border border-surface0 bg-mantle px-4 py-[10px] text-[12px] text-text">
				<span>
					<span className="text-subtext0">role </span>Software Engineer at C1
				</span>
				<span>
					<span className="text-subtext0">status </span>
					<span className="text-green">● open to interesting conversations</span>
				</span>
				<span>
					<span className="text-subtext0">builds </span>map servers · CLI tools · self-hosted
					platforms · hardware
				</span>
			</div>
		</div>
	)
}

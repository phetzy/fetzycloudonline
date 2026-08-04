export function TitleBar({ dims }: { dims: string }) {
	return (
		<div className="flex flex-none items-center gap-3 border-b border-surface0 bg-mantle px-[14px] py-[10px]">
			<span className="flex gap-[7px]" aria-hidden="true">
				<span className="h-[11px] w-[11px] rounded-full bg-red" />
				<span className="h-[11px] w-[11px] rounded-full bg-yellow" />
				<span className="h-[11px] w-[11px] rounded-full bg-green" />
			</span>
			<span className="flex-1 truncate text-center text-[12px] text-subtext0">
				fetzer@boise: ~/site — go run ./cmd/fetzer
			</span>
			<span className="text-[11.5px] text-subtext0">{dims}</span>
		</div>
	)
}

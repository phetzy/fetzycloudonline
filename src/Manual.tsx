export function Manual() {
	return (
		<div className="relative min-h-screen bg-ground px-[clamp(14px,4vw,48px)] pb-[108px] font-mono text-[14.5px] leading-[1.65] text-body">
			<div className="mx-auto max-w-[96ch]">
				<header className="flex flex-wrap justify-between gap-4 border-b border-rule py-4 text-[12.5px] tracking-[0.08em] text-chrome">
					<span>FETZER(1)</span>
					<span>General Commands Manual</span>
					<span>FETZER(1)</span>
				</header>

				<main className="pt-[clamp(28px,5vh,52px)]">
					<footer className="flex flex-wrap justify-between gap-4 border-t border-rule pt-4 text-[12.5px] text-dim">
						<span>Boise, Idaho</span>
						<span>2026</span>
						<span>FETZER(1)</span>
					</footer>
				</main>
			</div>
		</div>
	)
}

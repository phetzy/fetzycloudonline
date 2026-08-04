import { prefersReducedMotion } from '../hooks/prefersReducedMotion'

export function MinimizedNote() {
	return (
		<div className="flex flex-1 items-center justify-center">
			<p
				className={`text-center text-[clamp(16px,2.4vw,26px)] leading-[1.5] text-subtext0 ${
					prefersReducedMotion() ? '' : 'animate-risein'
				}`}
			>
				Yes, I am a <em className="italic text-acc">Catppuccin</em> enjoyer
				<span
					aria-hidden="true"
					className="ml-[0.35em] inline-block h-[1.05em] w-[0.55em] animate-blink bg-acc align-[-0.16em]"
				/>
			</p>
		</div>
	)
}

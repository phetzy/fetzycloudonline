import { TitleBar } from './components/TitleBar'
import { SECTIONS } from './content'
import { useTerminalDims } from './hooks/useTerminalDims'

export function App() {
	const dims = useTerminalDims()

	return (
		<div className="h-screen overflow-hidden bg-crust p-[clamp(8px,2.4vw,32px)] font-mono text-text">
			<div className="mx-auto flex h-full max-w-[1220px] flex-col overflow-hidden rounded-[10px] border border-surface0 bg-base">
				<TitleBar dims={dims} />
				<div className="flex min-h-0 flex-1 flex-col gap-3 p-[clamp(12px,2vw,20px)]">
					{SECTIONS.map((s) => (
						<article key={s.id}>
							<h2>{s.title}</h2>
						</article>
					))}
				</div>
			</div>
		</div>
	)
}

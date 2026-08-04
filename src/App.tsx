import { SECTIONS } from './content'

export function App() {
	return (
		<div>
			<h1>David Fetzer</h1>
			{SECTIONS.map((s) => (
				<article key={s.id}>
					<h2>{s.title}</h2>
					{s.paras.map((p) => (
						<p key={p}>{p}</p>
					))}
				</article>
			))}
		</div>
	)
}

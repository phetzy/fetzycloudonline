import { useEffect, useState } from 'react'

/** The prototype's cell metrics: 8.4px per column, 19px per row. */
const CELL_W = 8.4
const CELL_H = 19

export function formatDims(width: number, height: number): string {
	const cols = Math.max(1, Math.round(width / CELL_W))
	const rows = Math.max(1, Math.round(height / CELL_H))
	return `${cols}×${rows}`
}

/**
 * Starts empty and measures after mount. The document is prerendered, so a
 * computed initial value would not match what the server rendered.
 */
export function useTerminalDims(): string {
	const [dims, setDims] = useState('')

	useEffect(() => {
		const measure = () => setDims(formatDims(window.innerWidth, window.innerHeight))
		measure()
		window.addEventListener('resize', measure)
		return () => window.removeEventListener('resize', measure)
	}, [])

	return dims
}

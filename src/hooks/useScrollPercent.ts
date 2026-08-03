import { useEffect, useState } from 'react'

export function formatPercent(scrollY: number, scrollHeight: number, innerHeight: number): string {
	const distance = scrollHeight - innerHeight
	const percent = distance > 0 ? Math.round((scrollY / distance) * 100) : 100
	return percent >= 99 ? 'END' : `${percent}%`
}

export function useScrollPercent(): string {
	const [percent, setPercent] = useState('0%')

	useEffect(() => {
		const onScroll = () => {
			setPercent(
				formatPercent(window.scrollY, document.documentElement.scrollHeight, window.innerHeight)
			)
		}
		onScroll()
		window.addEventListener('scroll', onScroll, { passive: true })
		return () => window.removeEventListener('scroll', onScroll)
	}, [])

	return percent
}

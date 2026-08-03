import { useCallback, useEffect, useRef, useState } from 'react'
import { SECTIONS } from '../content'
import { scrollBehavior } from './prefersReducedMotion'

/** Milliseconds after an explicit jump during which observer updates are ignored. */
const LOCK_MS = 900

export function useManualNav() {
	const [idx, setIdx] = useState(0)
	const lockUntil = useRef(0)

	const goTo = useCallback((i: number) => {
		const n = Math.max(0, Math.min(i, SECTIONS.length - 1))
		const el = document.getElementById(SECTIONS[n].id)
		if (el) {
			window.scrollTo({
				top: el.getBoundingClientRect().top + window.scrollY - 24,
				behavior: scrollBehavior()
			})
		}
		lockUntil.current = Date.now() + LOCK_MS
		setIdx(n)
	}, [])

	useEffect(() => {
		// jsdom has no IntersectionObserver; unit tests drive idx through goTo instead.
		if (typeof IntersectionObserver === 'undefined') return

		const observer = new IntersectionObserver(
			(entries) => {
				if (Date.now() < lockUntil.current) return
				const visible = entries
					.filter((e) => e.isIntersecting)
					.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
				const top = visible[0]
				if (!top) return
				const i = SECTIONS.findIndex((s) => s.id === top.target.id)
				if (i >= 0) setIdx((current) => (current === i ? current : i))
			},
			{ rootMargin: '-24px 0px -70% 0px' }
		)

		for (const section of SECTIONS) {
			const el = document.getElementById(section.id)
			if (el) observer.observe(el)
		}
		return () => observer.disconnect()
	}, [])

	return { idx, goTo }
}

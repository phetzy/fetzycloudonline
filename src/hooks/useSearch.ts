import { useCallback, useEffect, useRef, useState, type RefObject } from 'react'
import { scrollBehavior } from './prefersReducedMotion'

const SELECTOR = 'p, h1, h2, h3, span, a'

export function findMatches(root: HTMLElement, query: string): HTMLElement[] {
	const needle = query.trim().toLowerCase()
	if (needle.length < 2) return []
	return Array.from(root.querySelectorAll<HTMLElement>(SELECTOR)).filter(
		(el) => !el.querySelector(SELECTOR) && (el.textContent ?? '').toLowerCase().includes(needle)
	)
}

export function useSearch(contentRef: RefObject<HTMLElement | null>) {
	const [query, setQueryState] = useState('')
	const [matches, setMatches] = useState<HTMLElement[]>([])
	const [matchIdx, setMatchIdx] = useState(0)
	const highlighted = useRef<HTMLElement | null>(null)

	const clearHighlight = useCallback(() => {
		if (highlighted.current) {
			highlighted.current.style.background = ''
			highlighted.current.style.color = ''
			highlighted.current = null
		}
	}, [])

	const highlight = useCallback(
		(el: HTMLElement) => {
			clearHighlight()
			el.style.background = 'var(--ph)'
			el.style.color = '#0D0D0E'
			highlighted.current = el
			window.scrollTo({
				top: el.getBoundingClientRect().top + window.scrollY - 120,
				behavior: scrollBehavior()
			})
		},
		[clearHighlight]
	)

	const setQuery = useCallback(
		(value: string) => {
			clearHighlight()
			const found = contentRef.current ? findMatches(contentRef.current, value) : []
			setQueryState(value)
			setMatches(found)
			setMatchIdx(0)
			if (found.length > 0) highlight(found[0])
		},
		[clearHighlight, contentRef, highlight]
	)

	// Deliberately not a functional setState: highlight() mutates the DOM, and
	// StrictMode invokes updater functions twice.
	const showMatch = useCallback(
		(offset: number) => {
			if (matches.length === 0) return
			const next = (matchIdx + offset + matches.length) % matches.length
			highlight(matches[next])
			setMatchIdx(next)
		},
		[highlight, matchIdx, matches]
	)

	const clear = useCallback(() => {
		clearHighlight()
		setQueryState('')
		setMatches([])
		setMatchIdx(0)
	}, [clearHighlight])

	useEffect(() => clearHighlight, [clearHighlight])

	const matchLabel =
		matches.length > 0
			? `${matchIdx + 1}/${matches.length}`
			: query.trim().length > 1
				? 'no match'
				: 'enter ↵ next'

	return { query, setQuery, matchCount: matches.length, matchLabel, showMatch, clear }
}

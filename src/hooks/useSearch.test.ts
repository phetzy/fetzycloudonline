import { act, renderHook } from '@testing-library/react'
import { findMatches, useSearch } from './useSearch'

function root(html: string): HTMLElement {
	const el = document.createElement('div')
	el.innerHTML = html
	return el
}

test('ignores queries of one character or less', () => {
	const el = root('<p>routing</p>')
	expect(findMatches(el, '')).toHaveLength(0)
	expect(findMatches(el, 'r')).toHaveLength(0)
	expect(findMatches(el, '  ')).toHaveLength(0)
})

test('matches leaf elements only, not their containers', () => {
	const el = root('<p><span>routing</span></p>')
	const matches = findMatches(el, 'rou')
	expect(matches).toHaveLength(1)
	expect(matches[0].tagName).toBe('SPAN')
})

test('matches case-insensitively', () => {
	const el = root('<p>Routing</p><p>ROUTING</p>')
	expect(findMatches(el, 'routing')).toHaveLength(2)
})

test('returns matches in document order', () => {
	const el = root('<p>first routing</p><p>second routing</p>')
	expect(findMatches(el, 'routing').map((m) => m.textContent)).toEqual([
		'first routing',
		'second routing'
	])
})

test('ignores elements that do not contain the query', () => {
	const el = root('<p>routing</p><p>geocoding</p>')
	expect(findMatches(el, 'geo')).toHaveLength(1)
})

test('showMatch(1) wraps from the last match back to the first', () => {
	const el = root('<p>routing</p><p>routing</p><p>routing</p>')
	const ref = { current: el }
	const { result } = renderHook(() => useSearch(ref))

	act(() => result.current.setQuery('routing'))
	expect(result.current.matchLabel).toBe('1/3')

	act(() => result.current.showMatch(1))
	expect(result.current.matchLabel).toBe('2/3')

	act(() => result.current.showMatch(1))
	expect(result.current.matchLabel).toBe('3/3')

	act(() => result.current.showMatch(1))
	expect(result.current.matchLabel).toBe('1/3')
})

test('showMatch(-1) wraps from the first match back to the last', () => {
	const el = root('<p>routing</p><p>routing</p><p>routing</p>')
	const ref = { current: el }
	const { result } = renderHook(() => useSearch(ref))

	act(() => result.current.setQuery('routing'))
	expect(result.current.matchLabel).toBe('1/3')

	act(() => result.current.showMatch(-1))
	expect(result.current.matchLabel).toBe('3/3')
})

import { scrollBehavior } from './prefersReducedMotion'

function stubMatchMedia(matches: boolean) {
	window.matchMedia = ((query: string) => ({
		matches,
		media: query,
		onchange: null,
		addListener: () => {},
		removeListener: () => {},
		addEventListener: () => {},
		removeEventListener: () => {},
		dispatchEvent: () => false
	})) as unknown as typeof window.matchMedia
}

test('returns auto when the user prefers reduced motion', () => {
	stubMatchMedia(true)
	expect(scrollBehavior()).toBe('auto')
})

test('returns smooth when the user has no motion preference', () => {
	stubMatchMedia(false)
	expect(scrollBehavior()).toBe('smooth')
})

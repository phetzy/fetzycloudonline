import { prefersReducedMotion, scrollBehavior } from './prefersReducedMotion'

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

test('reports the reduced-motion preference', () => {
	const stub = (matches: boolean) =>
		vi.stubGlobal('matchMedia', () => ({
			matches,
			media: '',
			addEventListener() {},
			removeEventListener() {}
		}))

	stub(true)
	expect(prefersReducedMotion()).toBe(true)

	stub(false)
	expect(prefersReducedMotion()).toBe(false)

	vi.unstubAllGlobals()
})

test('reports false when matchMedia is unavailable', () => {
	vi.stubGlobal('matchMedia', undefined)
	expect(prefersReducedMotion()).toBe(false)
	vi.unstubAllGlobals()
})

/**
 * ScrollOptions.behavior to use for a programmatic scroll, respecting the
 * user's `prefers-reduced-motion: reduce` setting.
 *
 * Safe to call during SSR (prerender.ts runs in Node, where `window` and
 * `matchMedia` don't exist) as long as it's only invoked inside browser-only
 * callbacks — it never touches `window` at module load time.
 */
export function scrollBehavior(): ScrollBehavior {
	if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return 'smooth'
	return window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth'
}

/**
 * Whether the visitor asked for reduced motion. Safe during SSR. Checked in
 * JavaScript rather than only in CSS because JS-driven animation ignores the
 * CSS media query.
 */
export function prefersReducedMotion(): boolean {
	if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
	return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

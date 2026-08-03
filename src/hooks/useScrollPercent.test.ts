import { formatPercent } from './useScrollPercent'

test('reports a rounded percentage of the scrollable distance', () => {
	expect(formatPercent(0, 2000, 1000)).toBe('0%')
	expect(formatPercent(500, 2000, 1000)).toBe('50%')
	expect(formatPercent(333, 2000, 1000)).toBe('33%')
})

test('reads END from 99 percent onward', () => {
	expect(formatPercent(990, 2000, 1000)).toBe('END')
	expect(formatPercent(1000, 2000, 1000)).toBe('END')
})

test('reports END when the page is not scrollable', () => {
	expect(formatPercent(0, 800, 1000)).toBe('END')
})

import { formatDims } from './useTerminalDims'

test('derives a cols×rows readout from the viewport', () => {
	expect(formatDims(1680, 1050)).toBe('200×55')
	expect(formatDims(840, 380)).toBe('100×20')
})

test('never reports a non-positive dimension', () => {
	expect(formatDims(0, 0)).toBe('1×1')
})

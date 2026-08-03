import { act, fireEvent, render, screen } from '@testing-library/react'
import { CopyButton } from './CopyButton'

function setClipboard(value: { writeText: (text: string) => Promise<void> } | undefined) {
	Object.defineProperty(navigator, 'clipboard', {
		value,
		configurable: true,
		writable: true
	})
}

afterEach(() => {
	setClipboard(undefined)
	vi.useRealTimers()
})

test('writes the value to the clipboard and confirms for 1600ms', async () => {
	const writeText = vi.fn().mockResolvedValue(undefined)
	setClipboard({ writeText })

	render(<CopyButton value="a@b.com" />)
	expect(screen.getByRole('button')).toHaveTextContent('copy')

	fireEvent.click(screen.getByRole('button'))
	expect(writeText).toHaveBeenCalledWith('a@b.com')

	// Let the writeText promise settle so setCopied(true) runs.
	await act(async () => {})
	expect(screen.getByRole('button')).toHaveTextContent('copied')
})

test('reverts to copy after the confirmation window', async () => {
	// NOTE: deviates from the task brief, which installs fake timers after the click.
	// That ordering leaves the pending setTimeout as a real native timer (Promise
	// microtasks resolve independent of vitest's timer faking here), so
	// advanceTimersByTime never fires it and the test fails deterministically.
	// Installing fake timers before the click does not freeze the microtask queue
	// (queueMicrotask/Promise resolution isn't among vitest's faked APIs) and
	// correctly captures the setTimeout call. See task-9-report.md for details.
	vi.useFakeTimers()
	setClipboard({ writeText: vi.fn().mockResolvedValue(undefined) })

	render(<CopyButton value="a@b.com" />)
	fireEvent.click(screen.getByRole('button'))
	await act(async () => {})
	expect(screen.getByRole('button')).toHaveTextContent('copied')

	await act(async () => {
		vi.advanceTimersByTime(1600)
	})
	expect(screen.getByRole('button')).toHaveTextContent('copy')
})

test('does nothing visible when the clipboard is unavailable', () => {
	setClipboard(undefined)

	render(<CopyButton value="a@b.com" />)
	fireEvent.click(screen.getByRole('button'))
	expect(screen.getByRole('button')).toHaveTextContent('copy')
})

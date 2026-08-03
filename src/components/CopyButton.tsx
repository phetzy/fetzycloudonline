import { useEffect, useRef, useState } from 'react'
import { TerminalButton } from './TerminalButton'

const CONFIRM_MS = 1600

export function CopyButton({ value }: { value: string }) {
	const [copied, setCopied] = useState(false)
	const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

	useEffect(() => () => clearTimeout(timer.current), [])

	const onClick = () => {
		if (!navigator.clipboard) return
		navigator.clipboard
			.writeText(value)
			.then(() => {
				setCopied(true)
				clearTimeout(timer.current)
				timer.current = setTimeout(() => setCopied(false), CONFIRM_MS)
			})
			.catch(() => {})
	}

	return (
		<TerminalButton onClick={onClick} className="ml-3 text-chrome" aria-label={`Copy ${value}`}>
			{copied ? 'copied' : 'copy'}
		</TerminalButton>
	)
}

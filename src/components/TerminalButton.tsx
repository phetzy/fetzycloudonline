import type { ButtonHTMLAttributes } from 'react'

export function TerminalButton({
	className = '',
	...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
	return (
		<button
			type="button"
			className={`border border-rule bg-transparent px-[10px] py-[3px] font-mono text-[12px] text-muted hover:border-ph hover:text-ph ${className}`}
			{...props}
		/>
	)
}

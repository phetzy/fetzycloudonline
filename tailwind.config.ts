import type { Config } from 'tailwindcss'

export default {
	content: ['./index.html', './src/**/*.{ts,tsx}'],
	theme: {
		extend: {
			colors: {
				crust: '#181926',
				base: '#24273a',
				mantle: '#1e2030',
				surface0: '#363a4f',
				surface1: '#494d64',
				rule: '#2f3348',
				text: '#cad3f5',
				subtext1: '#b8c0e0',
				subtext0: '#a5adcb',
				lavender: '#b7bdf8',
				mauve: '#c6a0f6',
				green: '#a6da95',
				red: '#ed8796',
				yellow: '#eed49f',
				acc: 'var(--acc, #f5a97f)',
				acc2: 'var(--acc2, #8aadf4)'
			},
			fontFamily: {
				mono: [
					'CaskaydiaCove NF',
					'Cascadia Code',
					'Symbols Nerd Font',
					'ui-monospace',
					'monospace'
				]
			},
			keyframes: {
				blink: { '0%, 49%': { opacity: '1' }, '50%, 100%': { opacity: '0' } },
				slidedown: {
					from: { opacity: '0', transform: 'translateY(10px)' },
					to: { opacity: '1', transform: 'none' }
				},
				slideup: {
					from: { opacity: '0', transform: 'translateY(-10px)' },
					to: { opacity: '1', transform: 'none' }
				},
				slidex: {
					from: { opacity: '0', transform: 'translateX(-10px)' },
					to: { opacity: '1', transform: 'none' }
				}
			},
			animation: {
				blink: 'blink 1.1s step-end infinite',
				slidedown: 'slidedown 170ms cubic-bezier(0.22, 1, 0.36, 1)',
				slideup: 'slideup 170ms cubic-bezier(0.22, 1, 0.36, 1)',
				slidex: 'slidex 190ms cubic-bezier(0.22, 1, 0.36, 1) both'
			}
		}
	},
	plugins: []
} as Config

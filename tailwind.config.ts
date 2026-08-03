import type { Config } from 'tailwindcss'

export default {
	content: ['./index.html', './src/**/*.{ts,tsx}'],
	theme: {
		extend: {
			colors: {
				ground: '#0D0D0E',
				ph: 'var(--ph)',
				bright: '#F2EEE4',
				body: '#D5D0C4',
				muted: '#A9A498',
				faint: '#8E8A80',
				dim: '#949086',
				chrome: '#7E7A70',
				sky: '#8FB8DE',
				rule: '#2A2A28',
				'rule-faint': '#1E1E1C',
				surface: '#17171A',
				panel: '#101012'
			},
			fontFamily: {
				mono: ['"IBM Plex Mono"', 'monospace']
			},
			keyframes: {
				blink: {
					'0%, 49%': { opacity: '1' },
					'50%, 100%': { opacity: '0' }
				}
			},
			animation: {
				blink: 'blink 1.1s step-end infinite'
			}
		}
	},
	plugins: []
} as Config

import js from '@eslint/js'
import ts from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'
import prettier from 'eslint-config-prettier'
import globals from 'globals'

/** @type {import('eslint').Linter.Config[]} */
export default [
	js.configs.recommended,
	...ts.configs.recommended,
	{
		files: ['**/*.{ts,tsx}'],
		plugins: { 'react-hooks': reactHooks },
		rules: reactHooks.configs.recommended.rules,
		languageOptions: {
			globals: { ...globals.browser, ...globals.node }
		}
	},
	prettier,
	{ ignores: ['dist/', 'playwright-report/', 'test-results/'] }
]

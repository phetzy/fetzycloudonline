import { readFileSync, writeFileSync } from 'node:fs'
import { createElement } from 'react'
import { renderToString } from 'react-dom/server'
import { App } from './src/App'

const template = readFileSync('dist/index.html', 'utf8')
const html = renderToString(createElement(App))

const output = template.replace('<div id="root"></div>', `<div id="root">${html}</div>`)
if (output === template) {
	throw new Error('prerender: could not find #root placeholder in dist/index.html')
}

writeFileSync('dist/index.html', output)

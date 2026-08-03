import { readFileSync, writeFileSync } from 'node:fs'
import { createElement } from 'react'
import { renderToString } from 'react-dom/server'
import { Manual } from './src/Manual'

const template = readFileSync('dist/index.html', 'utf8')
const html = renderToString(createElement(Manual))

writeFileSync(
	'dist/index.html',
	template.replace('<div id="root"></div>', `<div id="root">${html}</div>`)
)

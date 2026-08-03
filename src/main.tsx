import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Analytics } from '@vercel/analytics/react'
import './index.css'
import { Manual } from './Manual'

const root = document.getElementById('root')
if (!root) throw new Error('#root not found')

createRoot(root).render(
	<StrictMode>
		<Manual />
		<Analytics />
	</StrictMode>
)

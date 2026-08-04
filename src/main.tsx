import { StrictMode } from 'react'
import { hydrateRoot } from 'react-dom/client'
import { Analytics } from '@vercel/analytics/react'
import './index.css'
import { Manual } from './Manual'

const root = document.getElementById('root')
if (!root) throw new Error('#root not found')

hydrateRoot(
	root,
	<StrictMode>
		<Manual />
		<Analytics />
	</StrictMode>
)

// The prerendered document paints before hydration, so the page looks ready
// while the keyboard and scroll listeners are not attached yet. This marks the
// moment they are, which end-to-end tests wait on. Set on the root element,
// outside React's tree, so it cannot cause a hydration mismatch.
document.documentElement.dataset.hydrated = 'true'

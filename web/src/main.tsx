import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'

const rootEl = document.getElementById('root')
if (!rootEl) {
	throw new Error('Root element not found: make sure web/index.html contains <div id="root"></div>')
}

const root = createRoot(rootEl)
root.render(
	<React.StrictMode>
		<App />
	</React.StrictMode>
)

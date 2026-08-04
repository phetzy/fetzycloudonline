import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App } from './App'

test('renders the title block and metadata', () => {
	render(<App />)
	expect(screen.getByText('DAVID FETZER')).toBeInTheDocument()
	expect(screen.getByText('software engineer · boise, id · remote')).toBeInTheDocument()
	expect(screen.getByText(/Software Engineer at C1/)).toBeInTheDocument()
})

test('renders four tabs with readme active', () => {
	render(<App />)
	expect(screen.getByRole('button', { name: '▌ readme' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'projects' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'work' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'contact' })).toBeInTheDocument()
})

test('clicking a tab makes it active', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))
	expect(screen.getByRole('button', { name: '▌ projects' })).toBeInTheDocument()
})

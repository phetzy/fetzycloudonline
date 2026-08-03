import { render, screen } from '@testing-library/react'
import { Manual } from './Manual'

test('renders the manual page header', () => {
	render(<Manual />)
	expect(screen.getAllByText('FETZER(1)').length).toBeGreaterThan(0)
})

test('renders the manual footer', () => {
	render(<Manual />)
	expect(screen.getByText('General Commands Manual')).toBeInTheDocument()
	expect(screen.getByText('Boise, Idaho')).toBeInTheDocument()
	expect(screen.getByText('2026')).toBeInTheDocument()
})

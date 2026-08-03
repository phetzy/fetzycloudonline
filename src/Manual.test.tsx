import { render, screen } from '@testing-library/react'
import { Manual } from './Manual'

test('renders the manual page header', () => {
	render(<Manual />)
	expect(screen.getAllByText('FETZER(1)').length).toBeGreaterThan(0)
})

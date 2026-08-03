import { render, screen } from '@testing-library/react'
import { Manual } from './Manual'
import { SECTIONS } from './content'

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

test('every section in SECTIONS renders with its heading', () => {
	const { container } = render(<Manual />)
	for (const section of SECTIONS) {
		expect(container.querySelector(`#${section.id}`)).toBeInTheDocument()
		expect(screen.getByRole('heading', { level: 2, name: section.label })).toBeInTheDocument()
	}
})

test('there is exactly one h1', () => {
	render(<Manual />)
	expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1)
})

test('contact links point at the right destinations', () => {
	render(<Manual />)
	expect(screen.getByRole('link', { name: 'david.j.fetzer@gmail.com' })).toHaveAttribute(
		'href',
		'mailto:david.j.fetzer@gmail.com'
	)
	expect(screen.getByRole('link', { name: 'github.com/phetzy' })).toHaveAttribute(
		'href',
		'https://github.com/phetzy'
	)
	expect(screen.getByRole('link', { name: 'linkedin.com/in/fetzy' })).toHaveAttribute(
		'href',
		'https://linkedin.com/in/fetzy'
	)
})

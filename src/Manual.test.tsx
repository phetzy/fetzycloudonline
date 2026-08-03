import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Manual } from './Manual'
import { HELP, SECTIONS } from './content'

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

test('j and k move between sections', async () => {
	const user = userEvent.setup()
	render(<Manual />)
	expect(screen.getByTestId('current-section')).toHaveTextContent('NAME')

	await user.keyboard('j')
	expect(screen.getByTestId('position')).toHaveTextContent('2/11')
	expect(screen.getByTestId('current-section')).toHaveTextContent('SYNOPSIS')

	await user.keyboard('j')
	expect(screen.getByTestId('position')).toHaveTextContent('3/11')

	await user.keyboard('k')
	expect(screen.getByTestId('position')).toHaveTextContent('2/11')
})

test('g and G jump to the first and last section', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('{Shift>}G{/Shift}')
	expect(screen.getByTestId('position')).toHaveTextContent('11/11')

	await user.keyboard('g')
	expect(screen.getByTestId('position')).toHaveTextContent('1/11')
})

test('q jumps to SEE ALSO', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('q')
	expect(screen.getByTestId('current-section')).toHaveTextContent('SEE ALSO')
})

test('navigation keys are ignored while a modifier is held', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('{Control>}j{/Control}')
	expect(screen.getByTestId('position')).toHaveTextContent('1/11')
})

test('slash opens search and typing reports the match count', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('/')
	const input = screen.getByLabelText('Search the manual')
	expect(input).toHaveFocus()

	await user.type(input, 'geocoding')
	expect(screen.getByTestId('match-label')).toHaveTextContent(/^\d+\/\d+$/)
})

test('search reports no match for a query that is absent', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Search the manual'), 'zzzzqqq')
	expect(screen.getByTestId('match-label')).toHaveTextContent('no match')
})

test('escape closes search and restores the section readout', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Search the manual'), 'geocoding')
	await user.keyboard('{Escape}')

	expect(screen.queryByLabelText('Search the manual')).not.toBeInTheDocument()
	expect(screen.getByTestId('position')).toHaveTextContent('1/11')
})

test('t opens the table of contents with all eleven sections', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('t')
	const dialog = screen.getByRole('dialog')
	expect(dialog).toHaveTextContent('TABLE OF CONTENTS')
	expect(dialog).toHaveTextContent('01')
	expect(dialog).toHaveTextContent('11')
})

test('t toggles the table of contents closed again', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('t')
	expect(screen.getByRole('dialog')).toBeInTheDocument()
	await user.keyboard('t')
	expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})

test('question mark opens key help listing every binding', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('?')
	const dialog = screen.getByRole('dialog')
	expect(dialog).toHaveTextContent('KEYS')
	for (const row of HELP) {
		expect(dialog).toHaveTextContent(row.label)
	}
})

test('escape closes an open overlay', async () => {
	const user = userEvent.setup()
	render(<Manual />)

	await user.keyboard('?')
	await user.keyboard('{Escape}')
	expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})

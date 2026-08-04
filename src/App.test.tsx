import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App } from './App'

test('renders the title block and metadata', () => {
	render(<App />)
	const header = within(screen.getByTestId('header-row'))
	expect(header.getByText('DAVID FETZER')).toBeInTheDocument()
	expect(header.getByText('software engineer · boise, id · remote')).toBeInTheDocument()
	expect(header.getByText(/Software Engineer at C1/)).toBeInTheDocument()
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

test('the list shows the active tab name and its sections', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	expect(screen.getByText('PROJECTS')).toBeInTheDocument()
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'transfer-it-cli' })).toBeInTheDocument()
	expect(screen.getByText('1/5')).toBeInTheDocument()
})

test('clicking a list row selects it', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))
	await user.click(screen.getByRole('button', { name: 'transfer-it-cli' }))

	expect(screen.getByRole('button', { name: '› transfer-it-cli' })).toBeInTheDocument()
	expect(screen.getByText('2/5')).toBeInTheDocument()
})

test('renders every section as an article', () => {
	const { container } = render(<App />)
	expect(container.querySelectorAll('article')).toHaveLength(9)
})

test('after hydration only the selected article is visible', async () => {
	render(<App />)
	// The mount effect has run by the time render() returns.
	const articles = screen.getAllByRole('article', { hidden: true })
	const visible = articles.filter((a) => !a.hasAttribute('hidden'))
	expect(visible).toHaveLength(1)
	expect(visible[0]).toHaveAttribute('id', 'section-readme')
})

test('the detail pane shows the selected section', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	expect(screen.getByRole('heading', { name: 'mapwright', level: 2 })).toBeVisible()
	expect(screen.getByText('self-hosted map infrastructure · mapwright.io')).toBeVisible()
})

test('the prompt shows the section path and branch', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	expect(screen.getByText('~/site/projects/mapwright')).toBeInTheDocument()
})

test('j and k move the selection when the list has focus', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('j')
	expect(screen.getByRole('button', { name: '› transfer-it-cli' })).toBeInTheDocument()

	await user.keyboard('k')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
})

test('selection does not wrap past either end', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('k')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()

	await user.keyboard('{Shift>}G{/Shift}')
	expect(screen.getByRole('button', { name: '› open-source' })).toBeInTheDocument()
	await user.keyboard('j')
	expect(screen.getByRole('button', { name: '› open-source' })).toBeInTheDocument()
})

test('g and G jump to the first and last item', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('{Shift>}G{/Shift}')
	expect(screen.getByText('5/5')).toBeInTheDocument()

	await user.keyboard('g')
	expect(screen.getByText('1/5')).toBeInTheDocument()
})

test('h and l move focus between panes', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('l')
	expect(screen.getByText('scroll')).toBeInTheDocument()

	await user.keyboard('h')
	expect(screen.getByText('select')).toBeInTheDocument()
})

test('tab cycles tabs forward and shift+tab backward', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('{Tab}')
	expect(screen.getByRole('button', { name: '▌ projects' })).toBeInTheDocument()

	await user.keyboard('{Shift>}{Tab}{/Shift}')
	expect(screen.getByRole('button', { name: '▌ readme' })).toBeInTheDocument()
})

test('keys are ignored while a modifier is held', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('button', { name: 'projects' }))

	await user.keyboard('{Control>}j{/Control}')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
})

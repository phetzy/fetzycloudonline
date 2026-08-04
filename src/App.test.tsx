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
	expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()
	expect(screen.getByRole('tab', { name: 'projects' })).toBeInTheDocument()
	expect(screen.getByRole('tab', { name: 'work' })).toBeInTheDocument()
	expect(screen.getByRole('tab', { name: 'contact' })).toBeInTheDocument()
})

test('clicking a tab makes it active', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))
	expect(screen.getByRole('tab', { name: '▌ projects' })).toBeInTheDocument()
})

test('the list shows the active tab name and its sections', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))

	expect(screen.getByText('PROJECTS')).toBeInTheDocument()
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
	expect(screen.getByRole('button', { name: 'transfer-it-cli' })).toBeInTheDocument()
	expect(screen.getByText('1/5')).toBeInTheDocument()
})

test('clicking a list row selects it', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))
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
	await user.click(screen.getByRole('tab', { name: 'projects' }))

	expect(screen.getByRole('heading', { name: 'mapwright', level: 2 })).toBeVisible()
	expect(screen.getByText('self-hosted map infrastructure · mapwright.io')).toBeVisible()
})

test('the prompt shows the section path and branch', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))

	expect(screen.getByText('~/site/projects/mapwright')).toBeInTheDocument()
})

test('j and k move the selection when the list has focus', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))

	await user.keyboard('j')
	expect(screen.getByRole('button', { name: '› transfer-it-cli' })).toBeInTheDocument()

	await user.keyboard('k')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
})

test('selection does not wrap past either end', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))

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
	await user.click(screen.getByRole('tab', { name: 'projects' }))

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
	expect(screen.getByRole('tab', { name: '▌ projects' })).toBeInTheDocument()

	await user.keyboard('{Shift>}{Tab}{/Shift}')
	expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()
})

test('tab does not cycle tabs while the viewport is focused', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('l')
	expect(screen.getByText('scroll')).toBeInTheDocument()

	await user.keyboard('{Tab}')
	expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()

	await user.keyboard('{Shift>}{Tab}{/Shift}')
	expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()
})

test('keys are ignored while a modifier is held', async () => {
	const user = userEvent.setup()
	render(<App />)
	await user.click(screen.getByRole('tab', { name: 'projects' }))

	await user.keyboard('{Control>}j{/Control}')
	expect(screen.getByRole('button', { name: '› mapwright' })).toBeInTheDocument()
})

test('slash opens the filter and typing narrows the list', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	const input = screen.getByLabelText('Filter sections')
	expect(input).toHaveFocus()

	await user.type(input, 'mapwright')
	expect(screen.getByText('1/1 filtered')).toBeInTheDocument()
})

test('the filter searches every tab, not just the active one', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'stack')
	expect(screen.getByRole('button', { name: '› stack' })).toBeInTheDocument()
})

test('escape cancels the filter and clears it, leaving you where it carried you', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'mapwright')
	await user.keyboard('{Escape}')

	// Typing "mapwright" from readme crosses into projects and selects mapwright,
	// since readme drops out of the match list partway through. Escape only clears
	// filtering/filter — it does not rewind the tab or selection — so afterward the
	// list is the full, unfiltered projects tab (5 sections) with mapwright first:
	// 1/5, not 1/1.
	expect(screen.queryByLabelText('Filter sections')).not.toBeInTheDocument()
	expect(screen.getByText('1/5')).toBeInTheDocument()
})

test('enter keeps the filter and closes the input', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'mapwright')
	await user.keyboard('{Enter}')

	expect(screen.queryByLabelText('Filter sections')).not.toBeInTheDocument()
})

test('switching tabs while a filter is active clears it and shows the new tab full list', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'stack')
	await user.keyboard('{Enter}')
	expect(screen.getByRole('button', { name: '› stack' })).toBeInTheDocument()
	expect(screen.getByText('1/1 filtered')).toBeInTheDocument()

	// work -> contact
	await user.keyboard('{Tab}')

	expect(screen.getByRole('button', { name: '› contact' })).toBeInTheDocument()
	expect(screen.getByText('1/1')).toBeInTheDocument()
	expect(screen.queryByText('no match')).not.toBeInTheDocument()
})

test('a filter with no match reports no match', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.keyboard('/')
	await user.type(screen.getByLabelText('Filter sections'), 'zzzzqqq')
	expect(screen.getByText('no match')).toBeInTheDocument()
})

test('the yellow light minimizes the window and reveals the note', async () => {
	const user = userEvent.setup()
	render(<App />)

	expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()

	await user.click(screen.getByRole('button', { name: 'Minimize the terminal window' }))

	expect(screen.queryByRole('tab', { name: '▌ readme' })).not.toBeInTheDocument()
	expect(screen.getByText(/Yes, I am a/)).toBeInTheDocument()
	expect(screen.getByText('Catppuccin')).toBeInTheDocument()
})

test('the green light restores the window', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.click(screen.getByRole('button', { name: 'Minimize the terminal window' }))
	await user.click(screen.getByRole('button', { name: 'Restore the terminal window' }))

	expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()
	expect(screen.queryByText(/Yes, I am a/)).not.toBeInTheDocument()
})

test('each light is disabled when it has nothing to do', async () => {
	const user = userEvent.setup()
	render(<App />)

	expect(screen.getByRole('button', { name: 'Restore the terminal window' })).toBeDisabled()
	expect(screen.getByRole('button', { name: 'Minimize the terminal window' })).toBeEnabled()

	await user.click(screen.getByRole('button', { name: 'Minimize the terminal window' }))

	expect(screen.getByRole('button', { name: 'Minimize the terminal window' })).toBeDisabled()
	expect(screen.getByRole('button', { name: 'Restore the terminal window' })).toBeEnabled()
})

test('esc, enter, and space restore; other keys do nothing while minimized', async () => {
	const user = userEvent.setup()
	render(<App />)

	for (const key of ['{Escape}', '{Enter}', ' ']) {
		await user.click(screen.getByRole('button', { name: 'Minimize the terminal window' }))
		expect(screen.getByText(/Yes, I am a/)).toBeInTheDocument()
		await user.keyboard(key)
		expect(screen.getByRole('tab', { name: '▌ readme' })).toBeInTheDocument()
	}
})

test('TUI keybindings are suspended while minimized', async () => {
	const user = userEvent.setup()
	render(<App />)

	await user.click(screen.getByRole('button', { name: 'Minimize the terminal window' }))
	await user.keyboard('j')
	await user.keyboard('/')

	// Still minimized, and no filter input appeared.
	expect(screen.getByText(/Yes, I am a/)).toBeInTheDocument()
	expect(screen.queryByLabelText('Filter sections')).not.toBeInTheDocument()
})

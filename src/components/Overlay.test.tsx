import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { Overlay } from './Overlay'

function Harness() {
	const [open, setOpen] = useState(false)
	return (
		<div>
			<button onClick={() => setOpen(true)}>open overlay</button>
			{open && (
				<Overlay title="KEYS" rows={[{ key: 'j', label: 'down' }]} onClose={() => setOpen(false)} />
			)}
		</div>
	)
}

test('moves focus into the panel on open', async () => {
	const user = userEvent.setup()
	render(<Harness />)

	await user.click(screen.getByRole('button', { name: 'open overlay' }))

	expect(screen.getByRole('dialog')).toHaveFocus()
})

test('restores focus to the trigger when the overlay closes', async () => {
	const user = userEvent.setup()
	render(<Harness />)

	const trigger = screen.getByRole('button', { name: 'open overlay' })
	await user.click(trigger)
	expect(screen.getByRole('dialog')).toHaveFocus()

	await user.click(screen.getByRole('dialog'))
	expect(trigger).toHaveFocus()
})

import { expect, test, type Page } from '@playwright/test'

/** The document is prerendered, so it paints before React attaches listeners. */
async function gotoHydrated(page: Page) {
	await page.goto('/')
	await page.waitForSelector('html[data-hydrated="true"]')
}

test('head metadata is intact', async ({ page }) => {
	await page.goto('/')
	await expect(page).toHaveTitle('fetzer — TUI')
	await expect(page.locator('meta[property="og:title"]')).toHaveAttribute(
		'content',
		'David Fetzer — Software Engineer'
	)
})

test('the page itself never scrolls', async ({ page }) => {
	await gotoHydrated(page)
	const { scrollHeight, clientHeight } = await page.evaluate(() => ({
		scrollHeight: document.documentElement.scrollHeight,
		clientHeight: document.documentElement.clientHeight
	}))
	expect(scrollHeight).toBeLessThanOrEqual(clientHeight + 1)
})

test('the detail pane scrolls independently of the list', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('Tab') // projects
	await page.keyboard.press('l') // focus viewport
	await page.keyboard.press('G')

	const scrolled = await page.evaluate(() => {
		const panes = document.querySelectorAll('[data-vp]')
		return (panes[1] as HTMLElement).scrollTop
	})
	expect(scrolled).toBeGreaterThan(0)
})

test('j and k move the selection, and the status follows', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('Tab')
	await expect(page.getByText('1/5')).toBeVisible()

	await page.keyboard.press('j')
	await expect(page.getByText('2/5')).toBeVisible()
	await expect(page.getByRole('button', { name: '› transfer-it-cli' })).toBeVisible()
})

test('h and l move focus, and the help hint follows', async ({ page }) => {
	await gotoHydrated(page)
	await expect(page.getByText('select')).toBeVisible()
	await page.keyboard.press('l')
	await expect(page.getByText('scroll')).toBeVisible()
	await page.keyboard.press('h')
	await expect(page.getByText('select')).toBeVisible()
})

test('the filter opens, narrows, and cancels', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('/')
	await page.getByLabel('Filter sections').fill('mapwright')
	await expect(page.getByText('1/1 filtered')).toBeVisible()

	await page.keyboard.press('Escape')
	await expect(page.getByLabel('Filter sections')).toHaveCount(0)
})

test('enter opens the selected section link', async ({ page, context }) => {
	await gotoHydrated(page)
	await page.keyboard.press('Tab') // projects, mapwright selected

	const [popup] = await Promise.all([context.waitForEvent('page'), page.keyboard.press('Enter')])
	expect(popup.url()).toContain('mapwright.io')
	await popup.close()
})

test('only one article is visible after hydration', async ({ page }) => {
	await gotoHydrated(page)
	const visible = await page.evaluate(
		() =>
			Array.from(document.querySelectorAll('article')).filter((a) => !a.hasAttribute('hidden'))
				.length
	)
	expect(visible).toBe(1)
})

test('the yellow light minimizes and the green light restores', async ({ page }) => {
	await gotoHydrated(page)

	await page.getByRole('button', { name: 'Minimize the terminal window' }).click()
	await expect(page.getByText(/Yes, I am a/)).toBeVisible()
	await expect(page.getByRole('button', { name: '▌ readme' })).toHaveCount(0)

	await page.keyboard.press('Escape')
	await expect(page.getByRole('button', { name: '▌ readme' })).toBeVisible()
	await expect(page.getByText(/Yes, I am a/)).toHaveCount(0)
})

test('no horizontal overflow at a 375px viewport', async ({ page }) => {
	await page.setViewportSize({ width: 375, height: 800 })
	await gotoHydrated(page)
	const { scrollWidth, clientWidth } = await page.evaluate(() => ({
		scrollWidth: document.documentElement.scrollWidth,
		clientWidth: document.documentElement.clientWidth
	}))
	expect(scrollWidth).toBeLessThanOrEqual(clientWidth)
})

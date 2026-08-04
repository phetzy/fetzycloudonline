import { expect, test, type Page } from '@playwright/test'

/**
 * The document is prerendered, so it paints — and page.goto resolves — before
 * React hydrates. Any test that drives the keyboard or reads a JS-computed
 * readout has to wait for the listeners to be attached, or it races hydration.
 */
async function gotoHydrated(page: Page) {
	await page.goto('/')
	await page.waitForSelector('html[data-hydrated="true"]')
}

test('head metadata is intact', async ({ page }) => {
	await page.goto('/')
	await expect(page).toHaveTitle('FETZER(1) — Software Engineer')
	await expect(page.locator('meta[property="og:title"]')).toHaveAttribute(
		'content',
		'David Fetzer — Software Engineer'
	)
})

test('scrolling to the bottom reports END', async ({ page }) => {
	await gotoHydrated(page)
	await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
	await expect(page.getByTestId('percent')).toHaveText('END')
})

test('pressing j five times lands on the sixth section', async ({ page }) => {
	await gotoHydrated(page)
	for (let i = 0; i < 5; i++) await page.keyboard.press('j')
	await expect(page.getByTestId('position')).toHaveText('6/11')
	await expect(page.getByTestId('current-section')).toHaveText('SELF-HOSTED AI PLATFORM')
})

test('search finds a match and scrolls it into view', async ({ page }) => {
	await gotoHydrated(page)
	await page.keyboard.press('/')
	await page.getByLabel('Search the manual').fill('geocoding')
	await expect(page.getByTestId('match-label')).toHaveText(/^\d+\/\d+$/)
	await expect(page.getByText('Forward, reverse, and POI search.')).toBeInViewport()
})

test('the table of contents lists all eleven sections and closes on scrim click', async ({
	page
}) => {
	await gotoHydrated(page)
	await page.keyboard.press('t')
	const dialog = page.getByRole('dialog')
	await expect(dialog).toBeVisible()
	await expect(dialog.getByText('01')).toBeVisible()
	await expect(dialog.getByText('11')).toBeVisible()

	await page.mouse.click(5, 5)
	await expect(dialog).not.toBeVisible()
})

test('no horizontal overflow at a 375px viewport', async ({ page }) => {
	await page.setViewportSize({ width: 375, height: 800 })
	await page.goto('/')
	const { scrollWidth, clientWidth } = await page.evaluate(() => ({
		scrollWidth: document.documentElement.scrollWidth,
		clientWidth: document.documentElement.clientWidth
	}))
	expect(scrollWidth).toBeLessThanOrEqual(clientWidth)
})

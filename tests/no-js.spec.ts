import { expect, test } from '@playwright/test'

test('all content is readable without JavaScript', async ({ browser }) => {
	const context = await browser.newContext({ javaScriptEnabled: false })
	const page = await context.newPage()
	await page.goto('/')

	await expect(page.getByRole('heading', { level: 2 })).toHaveCount(11)
	await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
	await expect(page.getByText('Contributions to Jellyfin and Websurfx.')).toBeVisible()
	await expect(page.getByRole('link', { name: 'github.com/phetzy' })).toBeVisible()

	await context.close()
})

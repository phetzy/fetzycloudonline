import { expect, test } from '@playwright/test'

test('all thirteen sections are readable without JavaScript', async ({ browser }) => {
	const context = await browser.newContext({ javaScriptEnabled: false })
	const page = await context.newPage()
	await page.goto('/')

	await expect(page.locator('article')).toHaveCount(13)
	await expect(page.getByText('Contributions to Jellyfin, Websurfx, and jev-router.')).toBeVisible()

	await context.close()
})

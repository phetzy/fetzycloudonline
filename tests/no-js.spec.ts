import { expect, test } from '@playwright/test'

test('all nine sections are readable without JavaScript', async ({ browser }) => {
	const context = await browser.newContext({ javaScriptEnabled: false })
	const page = await context.newPage()
	await page.goto('/')

	await expect(page.locator('article')).toHaveCount(9)
	await expect(page.getByText('Contributions to Jellyfin and Websurfx.')).toBeVisible()

	await context.close()
})

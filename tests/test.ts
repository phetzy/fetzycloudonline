import { expect, test } from '@playwright/test'

test('page loads with the manual title', async ({ page }) => {
	await page.goto('/')
	await expect(page).toHaveTitle('FETZER(1) — Software Engineer')
})

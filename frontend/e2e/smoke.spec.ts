import { test, expect } from '@playwright/test'

test('authentication shell and offline project draft affordance', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('登录线禾')).toBeVisible()
  await page.getByRole('button', { name: '创建账号' }).click()
  await expect(page.getByText('创建线禾账号')).toBeVisible()
})

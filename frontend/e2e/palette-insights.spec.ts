import { test, expect } from '@playwright/test'

// 1×1 的 PNG，用来当线材照片（纹理只需要能解码）
const tinyPhoto = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFAAH/q842iQAAAABJRU5ErkJggg=='

test.beforeEach(async ({ page }) => {
  await page.goto('/')
  await page.getByPlaceholder('name@example.com').fill('admin@line.local')
  await page.getByPlaceholder('输入密码').fill('line_admin_2026')
  await page.getByRole('button', { name: '登录并继续' }).click()
  await expect(page.getByRole('button', { name: '配色推荐' })).toBeVisible()
  await page.getByRole('button', { name: '配色推荐' }).click()
  await expect(page.getByRole('heading', { name: '内置配色模板' })).toBeVisible()
})

async function chooseScheme(page: import('@playwright/test').Page, name: string) {
  await page.locator('.palette-controls .ant-select').first().click()
  // antd 的 role=listbox 是无障碍占位（宽高 0），真正可点的是 .ant-select-item-option
  await page.locator('.ant-select-dropdown .ant-select-item-option', { hasText: name }).first().click()
}

test('色卡文字按亮度自动反色，区分度不够时给出提醒', async ({ page }) => {
  // 浅色底：三色块会被算法压成同一个灰，文字必须是深色
  await page.locator('.palette-controls input[type=color]').fill('#f0f0f0')
  await chooseScheme(page, '互补色')

  const blocks = page.locator('.palette-result > div')
  await expect(blocks).toHaveCount(3)
  for (let index = 0; index < 3; index++) {
    await expect(blocks.nth(index)).toHaveCSS('color', 'rgb(17, 17, 17)')
  }

  // 三个颜色几乎相同 → 需要给出「色差太小」的提醒
  const insights = page.locator('.palette-insights')
  await expect(insights).toHaveClass(/warn/)
  await expect(insights).toContainText('色差')
  await expect(page.getByText(/处需要注意/)).toBeVisible()

  // 深色底：文字必须切成白色
  await page.locator('.palette-controls input[type=color]').fill('#2b2b2b')
  await chooseScheme(page, '邻近色')
  for (let index = 0; index < 3; index++) {
    await expect(blocks.nth(index)).toHaveCSS('color', 'rgb(255, 255, 255)')
  }
})

test('线材有照片时，色块用真实线材纹理渲染', async ({ page }) => {
  const yarnName = `纹理线 ${Date.now()}`
  const created = await page.request.post('/api/yarns', {
    data: { photoData: tinyPhoto, name: yarnName, brand: '', batchCode: '', color: '灰', colorHex: '#B8B8B8', material: '棉', thickness: 'DK', plies: 2, remainingGrams: 120, unitPrice: 0 },
  })
  expect(created.ok()).toBeTruthy()

  await page.reload()
  await page.getByRole('button', { name: '配色推荐' }).click()
  await page.locator('.palette-controls input[type=color]').fill('#f0f0f0')
  await chooseScheme(page, '互补色')

  const textured = page.locator('.palette-result > div[data-texture="yarn"]')
  await expect(textured.first()).toBeVisible()
  await expect(textured.first()).toContainText('实物')
  await expect(textured.first()).toContainText(yarnName)
})

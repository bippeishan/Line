import { test, expect } from '@playwright/test'

test('配色模板库：内置方案可套用、可收藏、可删除', async ({ page }) => {
  await page.goto('/')
  await page.getByPlaceholder('name@example.com').fill('admin@line.local')
  await page.getByPlaceholder('输入密码').fill('line_admin_2026')
  await page.getByRole('button', { name: '登录并继续' }).click()
  await expect(page.getByRole('button', { name: '配色推荐' })).toBeVisible()

  const listed = await page.request.get('/api/palette-templates')
  expect(listed.ok()).toBeTruthy()
  const builtins = (await listed.json()).data.filter((item: { isBuiltin: boolean }) => item.isBuiltin)
  expect(builtins.length).toBeGreaterThanOrEqual(10)
  for (const item of builtins) {
    expect(item.colors.length).toBeGreaterThanOrEqual(2)
    expect(item.colors[0]).toMatch(/^#[0-9A-F]{6}$/)
  }

  await page.getByRole('button', { name: '配色推荐' }).click()
  await expect(page.getByRole('heading', { name: '内置配色模板' })).toBeVisible()

  // 套用一组内置模板：色块必须变成模板自带的颜色
  const target = builtins[0] as { name: string; colors: string[] }
  await page.getByRole('button', { name: target.name, exact: true }).click()
  const result = page.locator('.palette-result')
  for (const color of target.colors) {
    await expect(result.getByText(color)).toBeVisible()
  }
  await expect(page.getByText(`当前使用模板「${target.name}」`, { exact: false })).toBeVisible()

  // 收藏当前配色，出现在模板库并可删除
  const customName = `我的配色 ${Date.now()}`
  await page.getByRole('button', { name: '收藏这组配色' }).click()
  await page.getByPlaceholder('例如：我的奶咖小熊').fill(customName)
  await page.getByRole('button', { name: '保存到模板库' }).click()
  await expect(page.getByRole('button', { name: customName, exact: true })).toBeVisible()

  const afterSave = await page.request.get('/api/palette-templates')
  const mine = (await afterSave.json()).data.find((item: { name: string }) => item.name === customName)
  expect(mine).toBeTruthy()
  expect(mine.isBuiltin).toBe(false)
  expect(mine.colors).toEqual(target.colors)

  // 重复名称会被拒绝
  const duplicate = await page.request.post('/api/palette-templates', { data: { name: customName, colors: target.colors } })
  expect(duplicate.status()).toBe(409)

  // 界面删除：卡片上的删除按钮 + 二次确认
  await page.getByRole('button', { name: `删除 ${customName}` }).click()
  await page.locator('.ant-popconfirm').getByRole('button', { name: /删/ }).click()
  await expect(page.getByRole('button', { name: customName, exact: true })).toHaveCount(0)

  const afterDelete = await page.request.get('/api/palette-templates')
  expect((await afterDelete.json()).data.some((item: { name: string }) => item.name === customName)).toBe(false)

  // 内置模板不允许删除
  const builtinDelete = await page.request.delete(`/api/palette-templates/${builtins[0].id}`)
  expect(builtinDelete.status()).toBe(404)
})

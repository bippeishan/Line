import { test, expect } from '@playwright/test'
import fs from 'node:fs'

const samplePdf = '/tmp/show-me-sample.pdf'

test('pattern file uploads through object storage and renders from its url', async ({ page }) => {
  fs.writeFileSync(samplePdf, '%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 200 200]>>endobj\ntrailer<</Root 1 0 R>>\n%%EOF\n')

  await page.goto('/')
  await page.getByPlaceholder('name@example.com').fill('admin@line.local')
  await page.getByPlaceholder('输入密码').fill('line_admin_2026')
  await page.getByRole('button', { name: '登录并继续' }).click()
  await expect(page.getByRole('button', { name: '图解与内容' })).toBeVisible()

  await page.getByRole('button', { name: '图解与内容' }).click()
  await page.setInputFiles('input[type=file]', samplePdf)

  const iframe = page.locator('.pattern-preview iframe')
  await expect(iframe).toBeVisible()
  const src = await iframe.getAttribute('src')
  expect(src).toMatch(/^\/api\/pattern-files\/\d+\/file$/)

  const shell = await page.evaluate(async () => (await fetch('/api/pattern-files')).json())
  const stored = shell.data[0]
  expect(stored.url).toBe(src)
  expect(stored.fileData).toBeUndefined()
  expect(stored.sizeBytes).toBeGreaterThan(0)

  // 已登录会话可以取回文件；匿名请求必须被拒绝。
  const authorized = await page.request.get(src)
  expect(authorized.status()).toBe(200)
  expect(authorized.headers()['content-type']).toContain('application/pdf')
  expect((await page.request.get(src, { headers: { Cookie: '' } })).status()).not.toBe(200)
})

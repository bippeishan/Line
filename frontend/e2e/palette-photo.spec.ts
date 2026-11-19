import { test, expect } from '@playwright/test'
import zlib from 'node:zlib'

// 用纯 Node 生成一张三色块的 PNG，避免引入二进制测试素材。
function crc32(buffer: Buffer): number {
  let crc = ~0
  for (const byte of buffer) {
    crc ^= byte
    for (let bit = 0; bit < 8; bit++) crc = (crc >>> 1) ^ (0xEDB88320 & -(crc & 1))
  }
  return ~crc >>> 0
}
function chunk(type: string, data: Buffer): Buffer {
  const length = Buffer.alloc(4)
  length.writeUInt32BE(data.length)
  const body = Buffer.concat([Buffer.from(type, 'ascii'), data])
  const crc = Buffer.alloc(4)
  crc.writeUInt32BE(crc32(body))
  return Buffer.concat([length, body, crc])
}
function yarnPhoto(colors: string[], size = 48): Buffer {
  const rgb = colors.map((hex) => [1, 3, 5].map((index) => parseInt(hex.slice(index, index + 2), 16)))
  const width = size * colors.length
  const raw = Buffer.alloc((width * 3 + 1) * size)
  for (let y = 0; y < size; y++) {
    let offset = y * (width * 3 + 1)
    raw[offset++] = 0
    for (let x = 0; x < width; x++) {
      const color = rgb[Math.min(colors.length - 1, Math.floor((x / width) * colors.length))]
      raw[offset++] = color[0]
      raw[offset++] = color[1]
      raw[offset++] = color[2]
    }
  }
  const header = Buffer.alloc(13)
  header.writeUInt32BE(width, 0)
  header.writeUInt32BE(size, 4)
  header[8] = 8
  header[9] = 2
  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', header),
    chunk('IDAT', zlib.deflateSync(raw)),
    chunk('IEND', Buffer.alloc(0)),
  ])
}

const photoColors = ['#e8a0a8', '#f6e7dd', '#6a8065']

test('照片取色：识别主色 → 生成组合 → 套用', async ({ page }) => {
  await page.goto('/')
  await page.getByPlaceholder('name@example.com').fill('admin@line.local')
  await page.getByPlaceholder('输入密码').fill('line_admin_2026')
  await page.getByRole('button', { name: '登录并继续' }).click()
  await expect(page.getByRole('button', { name: '配色推荐' })).toBeVisible()

  // 接口层：3 个颜色应返回按分数排序的组合
  const suggest = await page.request.post('/api/palette-suggest', { data: { colors: photoColors.map((color) => color.toUpperCase()) } })
  expect(suggest.ok()).toBeTruthy()
  const payload = (await suggest.json()).data
  expect(payload.suggestions.length).toBeGreaterThan(0)
  expect(payload.suggestions[0].score).toBeGreaterThan(0)
  expect(payload.suggestions[0].colors.length).toBeGreaterThanOrEqual(2)
  expect(payload.suggestions[0].colors.every((color: string) => /^#[0-9A-F]{6}$/.test(color))).toBe(true)

  const tooFew = await page.request.post('/api/palette-suggest', { data: { colors: ['#E8A0A8'] } })
  expect(tooFew.status()).toBe(400)

  // 界面层：上传照片 → 出现主色卡
  await page.getByRole('button', { name: '配色推荐' }).click()
  await expect(page.getByRole('heading', { name: '拍一张线材，让它挑颜色' })).toBeVisible()
  await page.setInputFiles('.palette-photo input[type=file]', {
    name: 'yarn.png',
    mimeType: 'image/png',
    buffer: yarnPhoto(photoColors),
  })

  const swatches = page.locator('.photo-color')
  await expect(swatches.first()).toBeVisible()
  expect(await swatches.count()).toBeGreaterThanOrEqual(3)
  await expect(page.locator('.photo-preview')).toBeVisible()

  // 识别出的主色应当接近照片里的三块颜色（允许量化与边缘混色的偏差）
  const distance = (left: string, right: string) => {
    const parse = (hex: string) => [1, 3, 5].map((index) => parseInt(hex.slice(index, index + 2), 16))
    const [a, b] = [parse(left), parse(right)]
    return Math.hypot(a[0] - b[0], a[1] - b[1], a[2] - b[2])
  }
  const extracted = (await swatches.locator('b').allTextContents()).map((text) => text.trim().toUpperCase())
  for (const color of photoColors) {
    expect(Math.min(...extracted.map((item) => distance(item, color.toUpperCase())))).toBeLessThan(60)
  }

  // 默认选中前三个，直接生成组合
  await expect(page.getByRole('button', { name: /生成配色组合（已选 3）/ })).toBeVisible()
  await page.getByRole('button', { name: /生成配色组合/ }).click()

  const suggestions = page.locator('.suggestion')
  await expect(suggestions.first()).toBeVisible()
  const count = await suggestions.count()
  expect(count).toBeGreaterThan(0)

  // 套用到下方结果区
  const firstColors = await suggestions.first().locator('.suggestion-colors span').all()
  await suggestions.first().getByRole('button', { name: '套用到下方' }).click()
  await expect(page.locator('.palette-result > div')).toHaveCount(firstColors.length)
})

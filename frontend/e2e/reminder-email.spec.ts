import { test, expect } from '@playwright/test'

// 需要本地或 CI 里运行 Mailpit：SMTP 1025 / HTTP 8025。
// 未设置 MAILPIT_URL 时跳过，避免没有邮件服务的环境误报。
const mailpitURL = process.env.MAILPIT_URL ?? ''
const smtpConfigured = !!process.env.SMTP_HOST

test.describe('提醒邮件触达', () => {
  test.skip(!mailpitURL || !smtpConfigured, '需要 MAILPIT_URL 与 SMTP_HOST 才能验证邮件投递')

  test('到点提醒通过邮件送达，并在界面显示投递状态', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('name@example.com').fill('admin@line.local')
    await page.getByPlaceholder('输入密码').fill('line_admin_2026')
    await page.getByRole('button', { name: '登录并继续' }).click()
    await expect(page.getByRole('button', { name: '图解与内容' })).toBeVisible()

    // 清空 Mailpit 收件箱，保证断言只看到本轮邮件。
    await page.request.delete(`${mailpitURL}/api/v1/messages`)

    const projectName = `提醒验证 ${Date.now()}`
    const created = await page.request.post('/api/projects', {
      data: { name: projectName, patternName: '', targetSize: '', hookSize: '', yarnNames: [], toolNames: [], palette: [], workTags: [], totalRows: 40, breakpointNote: '' },
    })
    expect(created.ok()).toBeTruthy()
    const projectId = (await created.json()).data.id as number

    const remindAt = new Date(Date.now() + 5 * 60 * 1000).toISOString()
    const reminder = await page.request.post(`/api/projects/${projectId}/reminders`, {
      data: { targetRow: 24, message: '在这里换浅色线', remindAt, channel: 'email' },
    })
    expect(reminder.ok()).toBeTruthy()
    const reminderId = (await reminder.json()).data.id as number

    const listed = await page.request.get(`/api/projects/${projectId}/reminders`)
    const pending = (await listed.json()).data.find((item: { id: number }) => item.id === reminderId)
    expect(pending.channel).toBe('email')
    expect(pending.emailSentAt).toBeNull()
    expect(pending.attempts).toBe(0)

    const sent = await page.request.post(`/api/projects/${projectId}/reminders/${reminderId}/send`)
    expect(sent.ok()).toBeTruthy()

    const after = await page.request.get(`/api/projects/${projectId}/reminders`)
    const delivered = (await after.json()).data.find((item: { id: number }) => item.id === reminderId)
    expect(delivered.emailSentAt).not.toBeNull()
    expect(delivered.attempts).toBe(1)
    expect(delivered.lastError).toBe('')

    const inbox = await page.request.get(`${mailpitURL}/api/v1/messages?limit=20`)
    const messages = (await inbox.json()).messages as Array<{ ID: string; Subject: string; To: Array<{ Address: string }> }>
    const mail = messages.find((item) => item.Subject.includes(projectName))
    expect(mail, '提醒邮件应该出现在 Mailpit 收件箱').toBeTruthy()
    expect(mail!.Subject).toContain('第 24 行')
    expect(mail!.To[0].Address).toBe('admin@line.local')

    const html = await page.request.get(`${mailpitURL}/api/v1/message/${mail!.ID}`)
    expect(await html.text()).toContain('在这里换浅色线')

    // 界面上的投递状态
    await page.goto('/')
    await page.getByRole('button', { name: '我的项目' }).click()
    await page.getByRole('heading', { name: projectName }).click()
    await page.getByRole('tab', { name: '提醒' }).click()
    await expect(page.getByText('已发送', { exact: false })).toBeVisible()
  })
})

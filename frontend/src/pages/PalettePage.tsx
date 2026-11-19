import { useEffect, useState } from 'react'
import { Button, Empty, Form, Input, Modal, Popconfirm, Select, Tag, message } from 'antd'

type Yarn = { id: number; name: string; color: string; colorHex: string; remainingGrams: number }
type Project = { id: number; name: string; palette: string[]; status: string; [key: string]: unknown }
type PaletteTemplate = { id: number; name: string; style: string; occasion: string; description: string; colors: string[]; isBuiltin: boolean; createdAt: string }
type TemplateForm = { name: string; style?: string; occasion?: string; description?: string }

const styles = ['甜系', '清新', '大地色', '低饱和', '自然', '清爽', '暖调', '糖果', '极简', '明亮']
const schemes = ['同色系', '邻近色', '互补色', '三角色']

export default function PalettePage({ onBack }: { onBack: () => void }) {
  const [base, setBase] = useState('#d97150')
  const [scheme, setScheme] = useState('邻近色')
  const [copied, setCopied] = useState(false)
  const [yarns, setYarns] = useState<Yarn[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [projectId, setProjectId] = useState<number | undefined>()
  const [saving, setSaving] = useState(false)
  const [templates, setTemplates] = useState<PaletteTemplate[]>([])
  const [templateId, setTemplateId] = useState<number>()
  const [styleFilter, setStyleFilter] = useState<string>()
  const [saveOpen, setSaveOpen] = useState(false)
  const [savingTemplate, setSavingTemplate] = useState(false)
  const [templateForm] = Form.useForm<TemplateForm>()
  const [messageApi, holder] = message.useMessage()

  const template = templates.find((item) => item.id === templateId)
  const colors = template ? template.colors : palette(base, scheme)
  const availableYarns = yarns.filter((item) => /^#[0-9a-fA-F]{6}$/.test(item.colorHex))
  const matches = colors.map((color) => recommendYarn(color, availableYarns))
  const visibleTemplates = styleFilter ? templates.filter((item) => item.style === styleFilter) : templates
  const styleOptions = Array.from(new Set(templates.map((item) => item.style).filter(Boolean)))

  const loadTemplates = () => fetch('/api/palette-templates').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setTemplates(data)).catch(() => messageApi.error('无法加载配色模板'))

  useEffect(() => {
    fetch('/api/yarns').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setYarns(data)).catch(() => undefined)
    fetch('/api/projects').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setProjects(data)).catch(() => undefined)
    loadTemplates()
  }, [])

  const applyTemplate = (item: PaletteTemplate) => {
    setTemplateId(item.id)
    if (item.colors[0]) setBase(item.colors[0])
  }
  const editColor = (value: string) => { setTemplateId(undefined); setBase(value) }
  const chooseScheme = (value: string) => { setTemplateId(undefined); setScheme(value) }
  const chooseYarn = (id: number) => {
    const yarn = availableYarns.find((item) => item.id === id)
    if (yarn) editColor(yarn.colorHex)
  }

  const copy = async () => {
    await navigator.clipboard?.writeText(colors.join(', '))
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }
  const saveToProject = async () => {
    const project = projects.find((item) => item.id === projectId)
    if (!project) return messageApi.error('请选择项目')
    setSaving(true)
    try {
      const response = await fetch('/api/projects/' + project.id, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...project, palette: colors }) })
      const payload = await response.json().catch(() => ({}))
      if (!response.ok) throw new Error(payload.error || '保存失败')
      setProjects((items) => items.map((item) => item.id === project.id ? { ...item, palette: colors } : item))
      messageApi.success('配色已关联到项目')
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }
  const saveTemplate = async (values: TemplateForm) => {
    setSavingTemplate(true)
    try {
      const response = await fetch('/api/palette-templates', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...values, colors }) })
      const payload = await response.json().catch(() => ({}))
      if (!response.ok) throw new Error(payload.error || '保存配色失败')
      await loadTemplates()
      setTemplateId(payload.data.id)
      setSaveOpen(false)
      templateForm.resetFields()
      messageApi.success('配色已收藏到模板库')
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '保存配色失败')
    } finally {
      setSavingTemplate(false)
    }
  }
  const removeTemplate = async (item: PaletteTemplate) => {
    try {
      const response = await fetch('/api/palette-templates/' + item.id, { method: 'DELETE' })
      if (!response.ok) throw new Error((await response.json().catch(() => ({}))).error || '删除失败')
      if (templateId === item.id) setTemplateId(undefined)
      await loadTemplates()
      messageApi.success('已删除收藏的配色')
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '删除失败')
    }
  }

  return <main className="palette-page">{holder}
    <nav><button className="brand brand-button" onClick={onBack}>线<span>禾</span></button><div className="admin-nav">创作工具 <span>/ 配色推荐</span></div><Button type="text" onClick={onBack}>返回首页</Button></nav>
    <section className="palette-head">
      <p className="eyebrow">COLOR HARMONY</p>
      <h1>配色推荐</h1>
      <p>先挑一组现成模板，再用算法微调；每一组都能落到真实库存上。</p>
    </section>

    <section className="palette-templates">
      <header>
        <div>
          <p className="eyebrow">TEMPLATES · {templates.length}</p>
          <h2>内置配色模板</h2>
        </div>
        <div className="palette-template-filter">
          <Select allowClear value={styleFilter} onChange={setStyleFilter} placeholder="按风格筛选" style={{ minWidth: 140 }} options={styleOptions.map((value) => ({ value }))} />
          <Button onClick={() => setSaveOpen(true)}>收藏当前配色</Button>
        </div>
      </header>
      {visibleTemplates.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有配色模板"/> : <div className="template-grid">{visibleTemplates.map((item) =>
        <div key={item.id} role="button" tabIndex={0} aria-label={item.name} aria-pressed={item.id === templateId} className={item.id === templateId ? 'template-card active' : 'template-card'} onClick={() => applyTemplate(item)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); applyTemplate(item) } }}>
          <span className="template-swatches">{item.colors.map((color) => <i key={color} style={{ background: color }} />)}</span>
          <span className="template-meta">
            <b>{item.name}</b>
            <small>{item.colors.map((color) => color.toUpperCase()).join(' · ')}</small>
          </span>
          <span className="template-tags">{item.style && <Tag>{item.style}</Tag>}{item.occasion && <Tag>{item.occasion}</Tag>}{!item.isBuiltin && <Tag color="volcano">我的</Tag>}</span>
          {item.description && <span className="template-note">{item.description}</span>}
          {!item.isBuiltin && <span className="template-delete" onClick={(event) => event.stopPropagation()}>
            <Popconfirm title="删除这组收藏？" okText="删除" cancelText="取消" onConfirm={() => void removeTemplate(item)}>
              <Button size="small" type="text" aria-label={`删除 ${item.name}`}>删除</Button>
            </Popconfirm>
          </span>}
        </div>)}</div>}
    </section>

    <section className="palette-work">
      <div className="palette-controls">
        <label>基准色 <input type="color" value={base} onChange={(event) => editColor(event.target.value)} /><code>{base.toUpperCase()}</code></label>
        <Select value={scheme} onChange={chooseScheme} options={schemes.map((value) => ({ value }))} disabled={!!template} />
        <Select value={undefined} onChange={chooseYarn} placeholder="从我的线材选择" options={availableYarns.map((item) => ({ value: item.id, label: item.name + ' · ' + (item.color || item.colorHex) + ` · ${item.remainingGrams}g` }))} />
        <Button type="primary" onClick={copy}>{copied ? '已复制色值' : '复制这组配色'}</Button>
        <Button onClick={() => setSaveOpen(true)}>收藏这组配色</Button>
        <p>{template ? `当前使用模板「${template.name}」，改动基准色会自动切回算法生成。` : availableYarns.length ? '按 HSL 色相、饱和度、明度匹配，并优先推荐库存充足的线材。' : '先在线材库中补充标准色值，即可匹配真实库存。'}</p>
      </div>
      <div className="palette-result">{colors.map((color, index) => <div key={color + index} style={{ background: color }}><span>0{index + 1}</span><strong>{color.toUpperCase()}</strong><small>{matches[index] ? <>{matches[index]!.yarn.name} · HSL 距离 {matches[index]!.distance}<br />库存 {matches[index]!.yarn.remainingGrams}g · {matches[index]!.reason}</> : '库存中未找到可用线材'}</small></div>)}</div>
    </section>

    <section className="palette-project-link">
      <Select value={projectId} onChange={setProjectId} placeholder="关联到项目" options={projects.filter((item) => item.status !== 'archived').map((item) => ({ value: item.id, label: item.name }))} />
      <Button type="primary" onClick={saveToProject} loading={saving}>保存到项目</Button>
    </section>

    <section className="palette-note">
      <b>{template ? template.name : scheme}</b>
      <span>{template ? (template.description || '来自模板库的现成配色，可直接套用到项目。') : scheme === '同色系' ? '同一色相的明暗变化，适合柔和有层次的成品。' : scheme === '邻近色' ? '相邻色相自然过渡，适合花束、条纹和渐变。' : scheme === '互补色' ? '对比色突出重点，适合小配件与五官细节。' : '三角关系均衡鲜明，适合玩偶和拼花。'}</span>
    </section>

    <Modal open={saveOpen} title="收藏这组配色" onCancel={() => setSaveOpen(false)} onOk={() => templateForm.submit()} confirmLoading={savingTemplate} okText="保存到模板库">
      <Form form={templateForm} layout="vertical" onFinish={saveTemplate}>
        <Form.Item name="name" label="模板名称" rules={[{ required: true, min: 1, max: 60, message: '给这组配色起个名字' }]}>
          <Input placeholder="例如：我的奶咖小熊"/>
        </Form.Item>
        <Form.Item name="style" label="风格">
          <Select allowClear placeholder="可选" options={styles.map((value) => ({ value }))} />
        </Form.Item>
        <Form.Item name="occasion" label="适用题材">
          <Select allowClear placeholder="可选" options={['玩偶', '花束', '服饰', '居家', '挂饰', '穿搭'].map((value) => ({ value }))} />
        </Form.Item>
        <Form.Item name="description" label="一句话说明">
          <Input maxLength={200} placeholder="可选，例如：深棕收边，适合秋冬款"/>
        </Form.Item>
        <div className="template-preview">{colors.map((color) => <i key={color} style={{ background: color }} title={color.toUpperCase()} />)}</div>
      </Form>
    </Modal>
  </main>
}

function palette(hex: string, scheme: string) {
  const red = parseInt(hex.slice(1, 3), 16), green = parseInt(hex.slice(3, 5), 16), blue = parseInt(hex.slice(5, 7), 16)
  const max = Math.max(red, green, blue), min = Math.min(red, green, blue)
  let hue = 0
  if (max !== min) hue = max === red ? (60 * ((green - blue) / (max - min)) + 360) % 360 : max === green ? 60 * ((blue - red) / (max - min)) + 120 : 60 * ((red - green) / (max - min)) + 240
  const saturation = max === 0 ? 0 : (max - min) / max, lightness = max / 255
  const offsets = scheme === '同色系' ? [0, 0, 0] : scheme === '邻近色' ? [-28, 0, 28] : scheme === '互补色' ? [0, 180, 180] : [0, 120, 240]
  return offsets.map((offset, index) => hslToHex((hue + offset + 360) % 360, scheme === '同色系' ? saturation * (.78 + index * .1) : saturation, scheme === '同色系' ? [Math.min(.86, lightness + .25), lightness, Math.max(.18, lightness - .22)][index] : Math.min(.72, Math.max(.32, lightness + (index - 1) * .08))))
}
type YarnRecommendation = { yarn: Yarn; distance: number; score: number; reason: string }
function recommendYarn(color: string, yarns: Yarn[]): YarnRecommendation | null {
  if (!yarns.length) return null
  const target = hexToHsl(color)
  return yarns.map((yarn) => {
    const candidate = hexToHsl(yarn.colorHex)
    const hueGap = Math.min(Math.abs(target.h - candidate.h), 360 - Math.abs(target.h - candidate.h)) / 180
    const saturationGap = Math.abs(target.s - candidate.s)
    const lightnessGap = Math.abs(target.l - candidate.l)
    const hslDistance = (hueGap * 55 + saturationGap * 25 + lightnessGap * 20)
    const stock = Math.max(0, yarn.remainingGrams || 0)
    const stockPenalty = stock === 0 ? 100 : stock < 20 ? 28 : stock < 50 ? 10 : stock >= 100 ? -12 : 0
    const score = hslDistance + stockPenalty
    const reason = stock >= 100 ? '库存充足' : stock >= 50 ? '库存可用' : stock >= 20 ? '库存偏少' : stock > 0 ? '低库存' : '无库存'
    return { yarn, score, distance: Math.round(hslDistance * 10) / 10, reason }
  }).filter((item) => item.yarn.remainingGrams > 0).sort((left, right) => left.score - right.score)[0] ?? null
}
function hexToHsl(hex: string) {
  const red = parseInt(hex.slice(1, 3), 16) / 255, green = parseInt(hex.slice(3, 5), 16) / 255, blue = parseInt(hex.slice(5, 7), 16) / 255
  const max = Math.max(red, green, blue), min = Math.min(red, green, blue), delta = max - min
  let hue = 0
  if (delta) hue = max === red ? 60 * (((green - blue) / delta + 6) % 6) : max === green ? 60 * ((blue - red) / delta + 2) : 60 * ((red - green) / delta + 4)
  const lightness = (max + min) / 2, saturation = delta === 0 ? 0 : delta / (1 - Math.abs(2 * lightness - 1))
  return { h: hue, s: saturation, l: lightness }
}
function hslToHex(hue: number, saturation: number, lightness: number) {
  const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation, value = chroma * (1 - Math.abs((hue / 60) % 2 - 1)), match = lightness - chroma / 2
  const [red, green, blue] = hue < 60 ? [chroma, value, 0] : hue < 120 ? [value, chroma, 0] : hue < 180 ? [0, chroma, value] : hue < 240 ? [0, value, chroma] : hue < 300 ? [value, 0, chroma] : [chroma, 0, value]
  return '#' + [red, green, blue].map((item) => Math.round((item + match) * 255).toString(16).padStart(2, '0')).join('')
}

import { useEffect, useState } from 'react'
import { Button, Select, message } from 'antd'

type Yarn = { id: number; name: string; color: string; colorHex: string; remainingGrams: number }
type Project = { id: number; name: string; palette: string[]; status: string; [key: string]: unknown }

export default function PalettePage({ onBack }: { onBack: () => void }) {
  const [base, setBase] = useState('#d97150')
  const [scheme, setScheme] = useState('邻近色')
  const [copied, setCopied] = useState(false)
  const [yarns, setYarns] = useState<Yarn[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [projectId, setProjectId] = useState<number | undefined>()
  const [saving, setSaving] = useState(false)
  const [messageApi, holder] = message.useMessage()
  const colors = palette(base, scheme)
  const availableYarns = yarns.filter((item) => /^#[0-9a-fA-F]{6}$/.test(item.colorHex))
  const matches = colors.map((color) => recommendYarn(color, availableYarns))

  useEffect(() => {
    fetch('/api/yarns').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setYarns(data)).catch(() => undefined)
    fetch('/api/projects').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setProjects(data)).catch(() => undefined)
  }, [])

  const copy = async () => {
    await navigator.clipboard?.writeText(colors.join(', '))
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }
  const chooseYarn = (id: number) => {
    const yarn = availableYarns.find((item) => item.id === id)
    if (yarn) setBase(yarn.colorHex)
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

  return <main className="palette-page">{holder}<nav><button className="brand brand-button" onClick={onBack}>线<span>禾</span></button><div className="admin-nav">创作工具 <span>/ 配色推荐</span></div><Button type="text" onClick={onBack}>返回首页</Button></nav><section className="palette-head"><p className="eyebrow">COLOR HARMONY</p><h1>配色推荐</h1><p>从已有线材出发，为下一件作品生成真正可用的三色方案。</p></section><section className="palette-work"><div className="palette-controls"><label>基准色 <input type="color" value={base} onChange={(event) => setBase(event.target.value)} /><code>{base.toUpperCase()}</code></label><Select value={scheme} onChange={setScheme} options={['同色系', '邻近色', '互补色', '三角色'].map((value) => ({ value }))} /><Select value={undefined} onChange={chooseYarn} placeholder="从我的线材选择" options={availableYarns.map((item) => ({ value: item.id, label: item.name + ' · ' + (item.color || item.colorHex) + ` · ${item.remainingGrams}g` }))} /><Button type="primary" onClick={copy}>{copied ? '已复制色值' : '复制这组配色'}</Button><p>{availableYarns.length ? '按 HSL 色相、饱和度、明度匹配，并优先推荐库存充足的线材。' : '先在线材库中补充标准色值，即可匹配真实库存。'}</p></div><div className="palette-result">{colors.map((color, index) => <div key={color} style={{ background: color }}><span>0{index + 1}</span><strong>{color.toUpperCase()}</strong><small>{matches[index] ? <>{matches[index]!.yarn.name} · HSL 距离 {matches[index]!.distance}<br />库存 {matches[index]!.yarn.remainingGrams}g · {matches[index]!.reason}</> : '库存中未找到可用线材'}</small></div>)}</div></section><section className="palette-project-link"><Select value={projectId} onChange={setProjectId} placeholder="关联到项目" options={projects.filter((item) => item.status !== 'archived').map((item) => ({ value: item.id, label: item.name }))} /><Button type="primary" onClick={saveToProject} loading={saving}>保存到项目</Button></section><section className="palette-note"><b>{scheme}</b><span>{scheme === '同色系' ? '同一色相的明暗变化，适合柔和有层次的成品。' : scheme === '邻近色' ? '相邻色相自然过渡，适合花束、条纹和渐变。' : scheme === '互补色' ? '对比色突出重点，适合小配件与五官细节。' : '三角关系均衡鲜明，适合玩偶和拼花。'}</span></section></main>
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

import { useEffect, useState, type ReactNode } from 'react'
import { CalculatorOutlined } from '@ant-design/icons'
import { Alert, Button, Form, InputNumber, Select, Tabs, message } from 'antd'

type CircleResult = { rounds: { round: number; stitches: number }[]; finalStitches: number; formula: string }
type SizeResult = { castOnStitches: number; rows: number; stitchGaugePerCM: number; rowGaugePerCM: number; formula: string }
type YarnResult = { baseGrams: number; recommendedGrams: number; recommendedSkeins: number; formula: string }
type Project = { id: number; name: string; status: 'planning' | 'in_progress' | 'completed' | 'archived' }
type CalculationDraft = { mode: 'circle' | 'size' | 'yarn'; input: Record<string, number> }

export default function CalculatorPage({ onBack }: { onBack: () => void }) {
  const [circle, setCircle] = useState<CircleResult | null>(null)
  const [size, setSize] = useState<SizeResult | null>(null)
  const [yarn, setYarn] = useState<YarnResult | null>(null)
  const [projects, setProjects] = useState<Project[]>([])
  const [draft, setDraft] = useState<CalculationDraft | null>(null)
  const [projectId, setProjectId] = useState<number | undefined>()
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState<string | null>(null)
  const [messageApi, contextHolder] = message.useMessage()

  useEffect(() => {
    fetch('/api/projects').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setProjects(data)).catch(() => undefined)
  }, [])

  const calculate = async <T,>(mode: CalculationDraft['mode'], values: Record<string, number>, setResult: (result: T) => void) => {
    setLoading(mode)
    try {
      const response = await fetch('/api/calculator/' + mode, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(values) })
      const payload = await response.json().catch(() => ({})) as { data?: T; error?: string }
      if (!response.ok || !payload.data) throw new Error(payload.error ?? '计算失败')
      setResult(payload.data)
      setDraft({ mode, input: values })
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '计算失败')
    } finally {
      setLoading(null)
    }
  }

  const saveCalculation = async () => {
    if (!draft) return
    setSaving(true)
    try {
      const response = await fetch('/api/calculations', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...draft, projectId }) })
      const payload = await response.json().catch(() => ({}))
      if (!response.ok) throw new Error(payload.error ?? '保存失败')
      messageApi.success(projectId ? '计算记录已关联项目' : '计算记录已保存')
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return <main className="calculator-page">{contextHolder}
    <nav><button className="brand brand-button" onClick={onBack}>线<span>禾</span></button><div className="admin-nav"><CalculatorOutlined /> 创作工具 <span>/ 钩织计算器</span></div><Button type="text" onClick={onBack}>返回首页</Button></nav>
    <section className="calculator-head"><div><p className="eyebrow">STITCH BY STITCH</p><h1>钩织计算器</h1><p>把密度、尺寸和用线量化为下一针可执行的数字。每项结果都保留计算依据，方便回看与调整。</p></div><div className="calculator-stamp">✿<span>CHECK<br />YOUR<br />GAUGE</span></div></section>
    <section className="calculator-workbench"><Tabs defaultActiveKey="circle" items={[
      { key: 'circle', label: '起针与圆形', children: <div className="calculator-grid"><Form layout="vertical" initialValues={{ startStitches: 6, rounds: 8, increasePerRound: 6 }} onFinish={(values) => calculate<CircleResult>('circle', values, setCircle)}><h2>圆形加针</h2><p className="form-hint">适用于魔法环或常规圆形起针；每一圈以固定针数增加。</p><NumberField name="startStitches" label="首圈针数" suffix="针" min={3} max={200} /><NumberField name="rounds" label="计算圈数" suffix="圈" min={1} max={100} /><NumberField name="increasePerRound" label="每圈增加" suffix="针" min={0} max={200} /><Button type="primary" htmlType="submit" loading={loading === 'circle'}>生成每圈针数</Button></Form><CalculatorResult title="每圈针数" formula={circle?.formula}>{circle ? <><strong className="big-number">{circle.finalStitches}<small>针 / 第 {circle.rounds.length} 圈</small></strong><div className="rounds">{circle.rounds.map((item) => <span key={item.round}>第 {item.round} 圈 <b>{item.stitches} 针</b></span>)}</div></> : '输入参数后，查看每圈的针数变化。'}</CalculatorResult></div> },
      { key: 'size', label: '密度与尺寸', children: <div className="calculator-grid"><Form layout="vertical" initialValues={{ stitchesPer10cm: 20, rowsPer10cm: 18, widthCm: 30, heightCm: 25 }} onFinish={(values) => calculate<SizeResult>('size', values, setSize)}><h2>尺寸换算</h2><p className="form-hint">先钩一块 10 × 10 cm 样片，再输入实测密度。</p><NumberField name="stitchesPer10cm" label="每 10 cm 针数" suffix="针" min={0.1} /><NumberField name="rowsPer10cm" label="每 10 cm 行数" suffix="行" min={0.1} /><NumberField name="widthCm" label="目标宽度" suffix="cm" min={0.1} /><NumberField name="heightCm" label="目标高度" suffix="cm" min={0.1} /><Button type="primary" htmlType="submit" loading={loading === 'size'}>换算起针与行数</Button></Form><CalculatorResult title="建议尺寸" formula={size?.formula}>{size ? <div className="size-results"><div><strong>{size.castOnStitches}</strong><span>建议起针</span></div><div><strong>{size.rows}</strong><span>建议行数</span></div><p>每厘米约 {size.stitchGaugePerCM.toFixed(2)} 针 / {size.rowGaugePerCM.toFixed(2)} 行</p></div> : '输入样片密度和目标尺寸，得到向上取整的建议。'}</CalculatorResult></div> },
      { key: 'yarn', label: '材料估算', children: <div className="calculator-grid"><Form layout="vertical" initialValues={{ metersPer100g: 125, patternMeters: 180, reserveRate: 10 }} onFinish={(values) => calculate<YarnResult>('yarn', values, setYarn)}><h2>用线量估算</h2><p className="form-hint">根据图解标注的长度，按所选线材的每 100 克米数折算。</p><NumberField name="metersPer100g" label="线材每 100 克长度" suffix="m" min={0.1} /><NumberField name="patternMeters" label="图解预计用线" suffix="m" min={0.1} /><NumberField name="reserveRate" label="余线预留" suffix="%" min={0} max={100} /><Button type="primary" htmlType="submit" loading={loading === 'yarn'}>估算建议用量</Button></Form><CalculatorResult title="材料建议" formula={yarn?.formula}>{yarn ? <div className="size-results"><div><strong>{yarn.recommendedGrams}</strong><span>建议准备克数</span></div><div><strong>{yarn.recommendedSkeins}</strong><span>按 100g / 团估算</span></div><p>不含余量的理论用量：{yarn.baseGrams} g</p></div> : '输入线材参数与图解用线量，预留余量后生成建议。'}</CalculatorResult></div> },
    ]} /></section>
    {draft && <section className="calculation-save"><div><p className="eyebrow">SAVE THIS RESULT</p><h2>把这次计算留下来</h2><p>保存的输入、公式和结果会在项目详情中持续可查。</p></div><div className="calculation-save-actions"><Select allowClear value={projectId} onChange={setProjectId} placeholder="关联到项目（可选）" options={projects.filter((item) => item.status !== 'archived').map((item) => ({ value: item.id, label: item.name }))} /><Button type="primary" onClick={saveCalculation} loading={saving}>保存计算</Button></div></section>}
    <Alert className="calculator-note" type="info" showIcon message="计算结果为辅助建议" description="实际成品会受针法、松紧度、线材批次和填充方式影响；开始前建议先完成样片测试。" />
  </main>
}

function NumberField({ name, label, suffix, min, max }: { name: string; label: string; suffix: string; min: number; max?: number }) {
  return <Form.Item name={name} label={label} rules={[{ required: true, message: '请输入' + label }]}><InputNumber min={min} max={max} precision={Number.isInteger(min) ? 0 : 2} addonAfter={suffix} style={{ width: '100%' }} /></Form.Item>
}

function CalculatorResult({ title, formula, children }: { title: string; formula?: string; children: ReactNode }) {
  return <aside className="calculator-result"><p className="eyebrow">RESULT / {title}</p><div className="result-content">{children}</div>{formula && <p className="formula">{formula}</p>}</aside>
}

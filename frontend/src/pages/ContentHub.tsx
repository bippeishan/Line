import { useEffect, useState } from 'react'
import { DeleteOutlined, UploadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Input, Popconfirm, Select, Upload, message } from 'antd'

type PatternFile = { id: number; title: string; fileData: string; fileType: 'pdf' | 'image'; createdAt: string }

export default function ContentHub({ onBack }: { onBack: () => void }) {
  const [source, setSource] = useState('')
  const [sourceType, setSourceType] = useState<'image' | 'pdf' | null>(null)
  const [stitches, setStitches] = useState('第 1 行：6 短针\n第 2 行：(1 短针，加针) × 3\n第 3 行：12 短针')
  const [question, setQuestion] = useState('')
  const [answer, setAnswer] = useState('')
  const [files, setFiles] = useState<PatternFile[]>([])
  const [selectedFileId, setSelectedFileId] = useState<number>()
  const [uploading, setUploading] = useState(false)
  const [asking, setAsking] = useState(false)
  const [messageApi, contextHolder] = message.useMessage()
  const selectedFile = files.find((file) => file.id === selectedFileId)

  const loadFiles = () => fetch('/api/pattern-files').then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setFiles(data)).catch(() => messageApi.error('无法加载图解文件库'))
  useEffect(() => { loadFiles() }, [])

  const loadFile = (file: File) => {
    const reader = new FileReader()
    reader.onload = async () => {
      const fileData = String(reader.result)
      const fileType = file.type === 'application/pdf' ? 'pdf' : 'image'
      setUploading(true)
      try {
        const response = await fetch('/api/pattern-files', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ title: file.name, fileData, fileType }) })
        if (!response.ok) throw new Error((await response.json().catch(() => ({}))).error || '上传失败')
        await loadFiles()
        setSource(fileData)
        setSourceType(fileType)
        messageApi.success('图解已保存到文件库')
      } catch (error) {
        messageApi.error(error instanceof Error ? error.message : '上传失败')
      } finally {
        setUploading(false)
      }
    }
    reader.readAsDataURL(file)
    return false
  }

  const selectFile = (id: number) => { const file = files.find((item) => item.id === id); if (!file) return; setSelectedFileId(id); setSource(file.fileData); setSourceType(file.fileType) }
  const removeFile = async (id: number) => { const response = await fetch('/api/pattern-files/' + id, { method: 'DELETE' }); if (!response.ok) return messageApi.error('删除失败'); if (selectedFileId === id) { setSelectedFileId(undefined); setSource(''); setSourceType(null) }; await loadFiles() }
  const ask = async () => { if (!question.trim()) return; setAsking(true); try { const response = await fetch('/api/content/ask', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ question, patternFileId: selectedFileId }) }); const payload = await response.json().catch(() => ({})); if (!response.ok) throw new Error(payload.error || '问答服务暂不可用'); setAnswer(payload.data.answer) } catch (error) { messageApi.error(error instanceof Error ? error.message : '问答失败') } finally { setAsking(false) } }

  return <main className="content-page">{contextHolder}
    <nav>
      <button className="brand brand-button" onClick={onBack}>线<span>禾</span></button>
      <div className="admin-nav">创作与学习 <span>/ 图解 · 内容</span></div>
      <Button type="text" onClick={onBack}>返回首页</Button>
    </nav>
    <section className="content-head">
      <p className="eyebrow">PATTERN · AI · LEARN</p>
      <h1>图解与内容</h1>
      <p>读懂每一行，也把自己的针法整理成可以分享的步骤。</p>
    </section>
    <section className="content-grid">
      <Card className="content-card" title="图解阅读器">
        <p>导入 PDF、图片或长图，在当前页面集中阅读。</p>
        <Upload accept="application/pdf,image/*" showUploadList={false} beforeUpload={loadFile}>
          <Button icon={<UploadOutlined />} loading={uploading}>导入并保存图解</Button>
        </Upload>
        <Select value={selectedFileId} onChange={selectFile} placeholder="从文件库打开图解" options={files.map((file) => ({ value: file.id, label: file.title }))} />
        {selectedFile && <Popconfirm title="删除这份图解？" onConfirm={() => removeFile(selectedFile.id)}><Button danger type="text" icon={<DeleteOutlined />}>删除当前图解</Button></Popconfirm>}
        {source && <div className="pattern-preview">{sourceType === 'pdf' ? <iframe src={source} title="图解 PDF" /> : <img src={source} alt="导入图解" />}</div>}
        <Alert type="info" showIcon message="已保存至文件库" description="图解可在项目工作台中关联，并使用图解阅读器添加持久化标注。OCR 将在启用 AI 服务后提供。" />
      </Card>
      <Card className="content-card" title="图解创作草稿">
        <p>按行输入文字针法，生成可复制的编号步骤。</p>
        <Input.TextArea value={stitches} onChange={(event) => setStitches(event.target.value)} rows={7} />
        <ol className="stitch-output">{stitches.split('\n').filter(Boolean).map((line, index) => <li key={index}>{line}</li>)}</ol>
      </Card>
      <Card className="content-card" title="智能问答 · 规则版">
        <p>先用可解释的建议帮你定位针数与尺寸问题。</p>
        <Input.TextArea value={question} onChange={(event) => setQuestion(event.target.value)} placeholder="例如：这一行针数为什么不一致？" rows={3} />
        <Button type="primary" onClick={ask} loading={asking}>获得建议</Button>
        {answer && <div className="answer-box">{answer}</div>}
        <Alert type="warning" showIcon message="需要配置 AI 服务" description="后端已预留问答接口；设置 OPENAI_API_KEY 后可接入基于已授权图解的问答。" />
      </Card>
    </section>
    <section className="learn-strip">
      <div>
        <p className="eyebrow">LEARN BY MAKING</p>
        <h2>从入门针法，到完成第一件作品。</h2>
      </div>
      <div>
        <span>01 · 基础针法</span>
        <span>02 · 图解符号</span>
        <span>03 · 玩偶圆形加针</span>
      </div>
    </section>
  </main>
}

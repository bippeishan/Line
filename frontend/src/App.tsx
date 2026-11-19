import { lazy, Suspense, useEffect, useState } from 'react'
import { ArrowDownOutlined, ArrowRightOutlined, DeleteOutlined, LockOutlined, LoginOutlined, LogoutOutlined, PlusOutlined, UploadOutlined } from '@ant-design/icons'
import { Badge, Button, Card, ConfigProvider, Empty, Form, Input, InputNumber, message, Modal, Popconfirm, Select, Tag, Upload } from 'antd'
import { discardLatestUndo, latestUndo, pushUndo, subscribeUndo } from './undo'

type Pattern = { id: number; title: string; category: string; difficulty: string; yarn: string; hookSize: string; durationMinutes: number }
type User = { id: number; name: string; email: string; role: 'admin' | 'user'; createdAt: string }
type Yarn = { id: number; photoData: string; name: string; brand: string; batchCode: string; color: string; colorHex: string; material: string; thickness: string; plies: number; remainingGrams: number; unitPrice: number; createdAt: string }
type YarnForm = Omit<Yarn, 'id' | 'createdAt' | 'photoData'>
type Tool = { id: number; photoData: string; name: string; type: string; thickness: string; createdAt: string }
type ToolForm = Omit<Tool, 'id' | 'createdAt' | 'photoData'>
type Project = { id:number; version:number; coverData:string; completionPhotoData:string; name:string; patternName:string; patternFileId?:number; targetSize:string; hookSize:string; yarnNames:string[]; toolNames:string[]; palette:string[]; workTags:string[]; shareToken:string; shareShowMaterials:boolean; shareShowNote:boolean; status:'planning'|'in_progress'|'completed'|'archived'; currentRow:number; totalRows:number; breakpointNote:string; completedAt?:string; completionNote:string; actualMinutes:number; targetDate?:string; weeklyFocusMinutes:number; createdAt:string; materialCost:number }
type UpcomingReminder = { projectId:number; projectName:string; targetRow:number; message:string }
type Insights = { weeklyFocus:{date:string;minutes:number}[]; monthlyCompleted:number; monthlyMinutes:number; monthlyCost:number; projectDistribution:{id:number;name:string;minutes:number}[]; inactiveProjects:{id:number;name:string;status:string;daysIdle:number}[] }
type ShoppingItem = { id:number; yarnId?:number; yarnName:string; targetGrams:number; estimatedPrice:number; priority:'low'|'normal'|'high'; status:'pending'|'purchased'; createdAt:string }
type InventoryShortage = { yarnId:number; yarnName:string; remainingGrams:number; suggestedGrams:number; projectNames:string[] }
const PublicWorkPage = lazy(() => import('./pages/PublicWorkPage'))
const UserManagement = lazy(() => import('./pages/UserManagement'))
const ContentHub = lazy(() => import('./pages/ContentHub'))
const CalculatorScreen = lazy(() => import('./pages/CalculatorPage'))
const PalettePage = lazy(() => import('./pages/PalettePage'))
const ProjectWorkspace = lazy(() => import('./pages/ProjectWorkspace'))

const fallbackPatterns: Pattern[] = [
  { id: 1, title: '初夏小花束', category: '居家装饰', difficulty: 'beginner', yarn: '棉线 2 股', hookSize: '3.0 mm', durationMinutes: 90 },
  { id: 2, title: '奶油吐司收纳篮', category: '居家装饰', difficulty: 'intermediate', yarn: '再生棉绳', hookSize: '5.0 mm', durationMinutes: 180 },
  { id: 3, title: '海风渔夫帽', category: '穿搭配饰', difficulty: 'intermediate', yarn: '亚麻棉混纺', hookSize: '4.0 mm', durationMinutes: 240 },
]
const icons = ['✿', '⌁', '◒']
const difficulty = { beginner: '入门', intermediate: '进阶', advanced: '挑战' }

export default function App() {
  const shareToken = window.location.pathname.match(/^\/share\/([a-f0-9]{32})\/?$/i)?.[1]
  const [view, setView] = useState<'home' | 'users' | 'calculator' | 'yarns' | 'tools' | 'projects' | 'content' | 'palette'>('home')
  const [account, setAccount] = useState<User | null>(null)
  const [resumeProjectId, setResumeProjectId] = useState<number | null>(null)
  const [checkingSession, setCheckingSession] = useState(true)
  useEffect(() => { fetch('/api/auth/me').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => setAccount(data)).catch(() => setAccount(null)).finally(() => setCheckingSession(false)) }, [])
  const logout = async () => { await fetch('/api/auth/logout', { method: 'POST' }); setAccount(null); setView('home') }
  return <ConfigProvider theme={{ token: { colorPrimary: '#d97150', colorText: '#23312a', colorTextBase: '#23312a', colorBgContainer: '#fcf8ef', borderRadius: 0, fontFamily: 'Noto Serif SC, serif' }, components: { Button: { fontWeight: 600, defaultShadow: 'none', primaryShadow: 'none' }, Card: { bodyPadding: 0 }, Table: { headerBg: '#f0dfc5' } } }}>
    {shareToken ? <Suspense fallback={<main className="session-loading">正在打开作品档案...</main>}><PublicWorkPage token={shareToken} /></Suspense> : checkingSession ? <main className="session-loading">正在准备你的钩织手帐…</main> : !account ? <LoginPage onLogin={(user) => setAccount(user)} /> : view === 'home' ? <Home account={account} onOpenUsers={() => setView('users')} onOpenCalculator={() => setView('calculator')} onOpenYarns={() => setView('yarns')} onOpenTools={() => setView('tools')} onOpenProjects={() => { setResumeProjectId(null); setView('projects') }} onContinueProject={(id) => { setResumeProjectId(id); setView('projects') }} onOpenContent={() => setView('content')} onOpenPalette={() => setView('palette')} onLogout={logout} /> : view === 'users' ? <Suspense fallback={<main className="session-loading">正在打开用户管理...</main>}><UserManagement onBack={() => setView('home')} /></Suspense> : view === 'calculator' ? <Suspense fallback={<main className="session-loading">正在打开钩织计算器...</main>}><CalculatorScreen onBack={() => setView('home')} /></Suspense> : view === 'yarns' ? <YarnLibrary onBack={() => setView('home')} /> : view === 'tools' ? <ToolLibrary onBack={() => setView('home')} /> : view === 'projects' ? <Suspense fallback={<main className="session-loading">正在打开项目工作台...</main>}><ProjectWorkspace onBack={() => setView('home')} initialProjectId={resumeProjectId} onInitialProjectOpened={() => setResumeProjectId(null)} /></Suspense> : view === 'content' ? <Suspense fallback={<main className="session-loading">正在打开图解与内容...</main>}><ContentHub onBack={() => setView('home')} /></Suspense> : <Suspense fallback={<main className="session-loading">正在打开配色推荐...</main>}><PalettePage onBack={() => setView('home')} /></Suspense>}
    {!shareToken && account && <UndoToast />}
  </ConfigProvider>
}

function UndoToast() {
  const [action,setAction]=useState(latestUndo())
  const [running,setRunning]=useState(false)
  useEffect(()=>subscribeUndo(()=>setAction(latestUndo())),[])
  if(!action)return null
  const undo=async()=>{setRunning(true);try{await action.undo();discardLatestUndo()}finally{setRunning(false)}}
  return <aside className="undo-toast"><span>{action.label}</span><Button type="primary" size="small" loading={running} onClick={undo}>撤销</Button></aside>
}

function Home({ account, onOpenUsers, onOpenCalculator, onOpenYarns, onOpenTools, onOpenProjects, onContinueProject, onOpenContent, onOpenPalette, onLogout }: { account: User; onOpenUsers: () => void; onOpenCalculator: () => void; onOpenYarns: () => void; onOpenTools: () => void; onOpenProjects: () => void; onContinueProject: (id: number) => void; onOpenContent: () => void; onOpenPalette: () => void; onLogout: () => void }) {
  const [patterns, setPatterns] = useState<Pattern[]>(fallbackPatterns)
  const [online, setOnline] = useState(false)
  const [projects, setProjects] = useState<Project[]>([])
  const [upcomingReminder, setUpcomingReminder] = useState<UpcomingReminder | null>(null)
  const [shoppingItems, setShoppingItems] = useState<ShoppingItem[]>([])
  const [insights, setInsights] = useState<Insights | null>(null)
  useEffect(() => { fetch('/api/patterns').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => { setPatterns(data); setOnline(true) }).catch(() => undefined) }, [])
  useEffect(() => { fetch('/api/projects').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => setProjects(data)).catch(() => undefined) }, [])
  useEffect(() => { fetch('/api/reminders/upcoming').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => setUpcomingReminder(data)).catch(() => undefined) }, [])
  useEffect(() => { fetch('/api/shopping-list').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => setShoppingItems(data)).catch(() => undefined) }, [])
  useEffect(() => { fetch('/api/insights/monthly').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => setInsights(data)).catch(() => undefined) }, [])
  const resume = (upcomingReminder ? projects.find((item) => item.id === upcomingReminder.projectId) : undefined) ?? projects.find((item) => item.status === 'in_progress') ?? projects.find((item) => item.status === 'planning')
  const weekStart = startOfWeek()
  const activeProjects = projects.filter((item) => item.status === 'planning' || item.status === 'in_progress')
  const completedThisWeek = projects.filter((item) => item.status === 'completed' && item.completedAt && new Date(item.completedAt) >= weekStart)
  const recentWorks = projects.filter((item) => item.status === 'completed').sort((a, b) => new Date(b.completedAt || b.createdAt).getTime() - new Date(a.completedAt || a.createdAt).getTime()).slice(0, 3)
  const activeRows = activeProjects.reduce((total, item) => total + item.currentRow, 0)
  const weekMinutes = completedThisWeek.reduce((total, item) => total + item.actualMinutes, 0)
  const pendingShopping = shoppingItems.filter((item) => item.status === 'pending')
  const shoppingBudget = pendingShopping.reduce((total, item) => total + item.estimatedPrice, 0)
  const urgentProjects=projects.filter((item)=>{if((item.status!=='planning'&&item.status!=='in_progress')||!item.targetDate)return false;const days=Math.ceil((new Date(item.targetDate).getTime()-Date.now())/86400000);return days<=7}).sort((a,b)=>new Date(a.targetDate!).getTime()-new Date(b.targetDate!).getTime()).slice(0,2)
  const maxFocus=Math.max(1,...(insights?.weeklyFocus.map((item)=>item.minutes)??[0]))
  return <main>
    <nav>
<a className="brand" href="#top">线<span>禾</span>
</a>
<div className="nav-links">
<button className="nav-action" onClick={onOpenCalculator}>钩织计算器</button>
<button className="nav-action" onClick={onOpenYarns}>我的线材</button>
<button className="nav-action" onClick={onOpenTools}>我的工具</button>
<button className="nav-action" onClick={onOpenContent}>图解与内容</button>
<button className="nav-action" onClick={onOpenPalette}>配色推荐</button>
<button className="nav-action" onClick={onOpenProjects}>我的项目</button>{account.role === 'admin' && <button className="nav-action" onClick={onOpenUsers}>用户管理</button>}</div>
<div className="account-menu">
<span>{account.name}</span>
<Button type="text" icon={<LogoutOutlined />} onClick={onLogout}>退出</Button>
</div>
</nav>
    <section className="hero" id="top">
<div className="hero-copy">
<p className="eyebrow">CROCHET, AT YOUR OWN PACE</p>
<h1>把时间<br />
<em>织</em>进日常。</h1>
<p className="intro">线禾是你的钩织小宇宙。收藏图解、记录针目，也收藏每一段被双手温柔编织的时光。</p>
<Button className="primary-link" type="text" href="#patterns">探索图解 <ArrowDownOutlined />
</Button>
</div>
<div className="hero-art" aria-label="毛线团装饰插图">
<div className="loop loop-one" />
<div className="loop loop-two" />
<div className="yarn-ball">
<i />
<i />
<i />
</div>
<div className="needle" />
<p>slow stitches<br />warm days</p>
</div>
<aside className="hero-note hero-resume">{resume ? <>
<span>继续钩织</span>
<strong>{resume.currentRow} <small>/ {resume.totalRows || '—'} 行</small>
</strong>
<p>{upcomingReminder?.projectId === resume.id ? `第 ${upcomingReminder.targetRow} 行：${upcomingReminder.message}` : resume.breakpointNote || '从上次停下的地方继续'}</p>
<Button type="primary" size="small" onClick={() => onContinueProject(resume.id)}>回到项目 <ArrowRightOutlined />
</Button>
</> : <>
<span>准备开工</span>
<strong>0 <small>个进行中</small>
</strong>
<p>建立第一个项目，<br />从第一针开始。</p>
<Button type="primary" size="small" onClick={onOpenProjects}>新建项目 <ArrowRightOutlined />
</Button>
</>}</aside>
</section>
    <section className="story-strip">
<p>“手作不是抵达终点，而是让手与心重新相遇。”</p>
<span>— 线禾手帐</span>
</section>
    <section className="review-board">
<header><div><p className="eyebrow">WEEKLY REVIEW</p><h2>这一周，织到了哪里？</h2></div><Button type="link" onClick={onOpenProjects}>查看项目 <ArrowRightOutlined /></Button></header>
<div className="review-layout"><div className="review-stats"><article><span>进行中项目</span><strong>{activeProjects.length}</strong><small>已推进 {activeRows} 行</small></article><article><span>本周完成</span><strong>{completedThisWeek.length}</strong><small>{weekMinutes ? `投入 ${formatHours(weekMinutes)}` : '等待一件作品收尾'}</small></article><article><span>下一条提醒</span><strong>{upcomingReminder ? `第 ${upcomingReminder.targetRow} 行` : '—'}</strong><small>{upcomingReminder?.projectName || '暂时没有待办提醒'}</small></article></div><div className="review-recent"><div className="review-title"><span>最近完成</span><small>{recentWorks.length ? '作品档案' : '尚无归档'}</small></div>{recentWorks.length ? recentWorks.map((work) => <button key={work.id} className="recent-work" onClick={() => onContinueProject(work.id)}><span className="recent-work-photo">{work.completionPhotoData || work.coverData ? <img src={work.completionPhotoData || work.coverData} alt="" /> : '线'}</span><span><b>{work.name}</b><small>{work.completedAt ? new Date(work.completedAt).toLocaleDateString('zh-CN') : '已完成'} · {work.actualMinutes ? formatHours(work.actualMinutes) : '未记录用时'}</small></span><ArrowRightOutlined /></button>) : <p className="review-empty">完成一件作品后，它会出现在这里。</p>}<button className="shopping-glance" onClick={onOpenYarns}><span>待补货材料</span><b>{pendingShopping.length} 项 · ¥ {shoppingBudget.toFixed(0)}</b><ArrowRightOutlined /></button></div></div>
</section>
    <section className="action-center"><header><div><p className="eyebrow">TODAY'S NEXT STITCH</p><h2>今天，先完成这几件事</h2></div><Button type="link" onClick={onOpenProjects}>打开项目 <ArrowRightOutlined/></Button></header><div>{upcomingReminder&&<button onClick={()=>onContinueProject(upcomingReminder.projectId)}><b>关键行提醒</b><span>{upcomingReminder.projectName} · 第 {upcomingReminder.targetRow} 行：{upcomingReminder.message}</span><ArrowRightOutlined/></button>}{pendingShopping.length>0&&<button onClick={onOpenYarns}><b>材料待购</b><span>{pendingShopping.length} 项待补货，预计 ¥ {shoppingBudget.toFixed(0)}</span><ArrowRightOutlined/></button>}{urgentProjects.map((item)=>{const days=Math.ceil((new Date(item.targetDate!).getTime()-Date.now())/86400000);return <button key={item.id} onClick={()=>onContinueProject(item.id)}><b>{days<0?'已超过目标日期':days===0?'目标日期是今天':`还有 ${days} 天`}</b><span>{item.name} · 第 {item.currentRow} / {item.totalRows||'—'} 行</span><ArrowRightOutlined/></button>})}{!upcomingReminder&&!pendingShopping.length&&!urgentProjects.length&&<p>今天没有紧急事项，可以按自己的节奏继续创作。</p>}</div></section>
    {insights&&<section className="insights-board"><header><div><p className="eyebrow">MAKING RHYTHM</p><h2>这一月的手作节奏</h2></div><span>累计专注 {formatHours(insights.monthlyMinutes)}</span></header><div className="insight-layout"><section className="focus-chart"><div className="chart-bars">{recentWeekDays().map((date)=>{const minutes=insights.weeklyFocus.find((item)=>item.date===date)?.minutes??0;return <div key={date}><i style={{height:`${Math.max(minutes?10:3,minutes/maxFocus*100)}%`}} title={`${minutes} 分钟`}/><span>{date.slice(5).replace('-','/')}</span></div>})}</div><p>近 7 天专注时长</p></section><section className="month-totals"><div><strong>{insights.monthlyCompleted}</strong><span>本月完成</span></div><div><strong>{formatHours(insights.monthlyMinutes)}</strong><span>专注投入</span></div><div><strong>¥ {insights.monthlyCost.toFixed(0)}</strong><span>完成作品材料</span></div></section><section className="project-insight"><h3>投入最多</h3>{insights.projectDistribution.length?insights.projectDistribution.map((item)=>{const total=Math.max(1,...insights.projectDistribution.map((entry)=>entry.minutes));return <button key={item.id} onClick={()=>onContinueProject(item.id)}><span>{item.name}</span><i><b style={{width:`${item.minutes/total*100}%`}}/></i><small>{formatHours(item.minutes)}</small></button>}):<p>开始一次专注后，这里会显示项目投入。</p>}</section><section className="inactive-projects"><h3>等待继续</h3>{insights.inactiveProjects.length?insights.inactiveProjects.map((item)=><button key={item.id} onClick={()=>onContinueProject(item.id)}><span>{item.name}</span><small>{item.daysIdle} 天未推进 <ArrowRightOutlined/></small></button>):<p>没有长期搁置的项目。</p>}</section></div></section>}
    <section className="patterns" id="patterns">
<header>
<div>
<p className="eyebrow">PATTERN SHELF <Badge status={online ? 'success' : 'default'} text={online ? '已连接工坊' : '本地预览'} />
</p>
<h2>今日想织点什么？</h2>
</div>
<Button type="link" href="#patterns">查看全部 <ArrowRightOutlined />
</Button>
</header>
<div className="pattern-grid">{patterns.map((item, index) => <Card className={`pattern-card card-${index % 3}`} key={item.id} hoverable cover={<div className="card-illustration">
<span>{icons[index % icons.length]}</span>
<small>{item.category}</small>
</div>}>
<div className="card-details">
<span>{difficulty[item.difficulty as keyof typeof difficulty] ?? item.difficulty} · {item.durationMinutes} 分钟</span>
<h3>{item.title}</h3>
<p>{item.yarn} &nbsp; / &nbsp; {item.hookSize}</p>
</div>
</Card>)}</div>
</section>
    <section className="footer-card" id="projects">
<div>
<p className="eyebrow">MAKE IT YOURS</p>
<h2>从第一针，<br />到属于你的作品。</h2>
</div>
<Button type="primary" size="large">开始记录 <ArrowRightOutlined />
</Button>
</section>
  </main>
}

function startOfWeek() { const date = new Date(); const day = date.getDay() || 7; date.setHours(0, 0, 0, 0); date.setDate(date.getDate() - day + 1); return date }
function formatHours(minutes: number) { return `${Math.round(minutes / 60 * 10) / 10} 小时` }
function recentWeekDays() { return Array.from({length:7},(_,index)=>{const date=new Date();date.setDate(date.getDate()-6+index);return date.toISOString().slice(0,10)}) }


function YarnLibrary({ onBack }: { onBack: () => void }) {
  const [yarns, setYarns] = useState<Yarn[]>([])
  const [shoppingItems, setShoppingItems] = useState<ShoppingItem[]>([])
  const [shortages, setShortages] = useState<InventoryShortage[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [purchaseOpen, setPurchaseOpen] = useState(false)
  const [photoData, setPhotoData] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm<YarnForm>()
  const [purchaseForm] = Form.useForm<Omit<ShoppingItem, 'id' | 'createdAt'>>()
  const [messageApi, contextHolder] = message.useMessage()
  const loadYarns = () => { setLoading(true); Promise.all([fetch('/api/yarns').then((r) => r.ok ? r.json() : Promise.reject()),fetch('/api/shopping-list').then((r) => r.ok ? r.json() : Promise.reject()),fetch('/api/inventory/shortages').then((r) => r.ok ? r.json() : Promise.reject())]).then(([yarnData,shoppingData,shortageData]) => { setYarns(yarnData.data); setShoppingItems(shoppingData.data); setShortages(shortageData.data) }).catch(() => messageApi.error('无法加载线材与采购数据，请重新登录后再试')).finally(() => setLoading(false)) }
  useEffect(() => { loadYarns() }, [])
  const closeModal = () => { setOpen(false); setPhotoData(''); form.resetFields() }
  const uploadMedia = async (file: File) => { if (!file.type.startsWith('image/')) throw new Error('请上传图片文件'); if (file.size > 8 * 1024 * 1024) throw new Error('照片不能超过 8 MB'); const body = new FormData(); body.append('file', file); const response = await fetch('/api/media/images', { method: 'POST', body }); const payload = await response.json().catch(() => ({})) as { data?: { url: string }; error?: string }; if (!response.ok || !payload.data?.url) throw new Error(payload.error ?? '图片上传失败'); return payload.data.url }
  const onPhotoSelect = (file: File) => { void uploadMedia(file).then(setPhotoData).catch((error) => messageApi.error(error instanceof Error ? error.message : '图片上传失败')); return false }
  const createYarn = async (values: YarnForm) => { setSubmitting(true); try { const response = await fetch('/api/yarns', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...values, photoData }) }); const payload = await response.json().catch(() => ({})) as { data?: Yarn; error?: string }; if (!response.ok || !payload.data) throw new Error(payload.error ?? '添加线材失败'); pushUndo({label:`已添加线材「${payload.data.name}」`,undo:async()=>{await fetch(`/api/yarns/${payload.data!.id}`,{method:'DELETE'});loadYarns()}}); messageApi.success('线材已收进你的线材库'); closeModal(); loadYarns() } catch (error) { messageApi.error(error instanceof Error ? error.message : '添加线材失败') } finally { setSubmitting(false) } }
  const removeYarn = async (item: Yarn) => { try { const response = await fetch(`/api/yarns/${item.id}`, { method: 'DELETE' }); if (!response.ok) { const payload = await response.json().catch(() => ({})) as { error?: string }; throw new Error(payload.error ?? '删除线材失败') }; pushUndo({label:`已移除线材「${item.name}」`,undo:async()=>{const result=await fetch('/api/yarns',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(item)});if(!result.ok)throw new Error('恢复线材失败');loadYarns()}}); messageApi.success(`已移除 ${item.name}`); loadYarns() } catch (error) { messageApi.error(error instanceof Error ? error.message : '删除线材失败') } }
  const createShoppingItem = async (values: Omit<ShoppingItem, 'id' | 'createdAt'>) => { try { const response=await fetch('/api/shopping-list',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)}); const payload=await response.json().catch(()=>({})) as {data?:{id:number};error?:string}; if(!response.ok||!payload.data) throw new Error(payload.error??'添加采购项失败'); pushUndo({label:`已添加采购项「${values.yarnName}」`,undo:async()=>{await fetch(`/api/shopping-list/${payload.data!.id}`,{method:'DELETE'});loadYarns()}}); messageApi.success('已加入采购清单'); setPurchaseOpen(false); purchaseForm.resetFields(); loadYarns() } catch(error) { messageApi.error(error instanceof Error?error.message:'添加采购项失败') } }
  const togglePurchased = async (item: ShoppingItem) => { try { const next={...item,status:item.status==='pending'?'purchased':'pending'} as ShoppingItem; const response=await fetch(`/api/shopping-list/${item.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify(next)}); if(!response.ok) throw new Error('更新采购状态失败'); pushUndo({label:`已更新采购项「${item.yarnName}」`,undo:async()=>{await fetch(`/api/shopping-list/${item.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify(item)});loadYarns()}}); loadYarns() } catch(error) { messageApi.error(error instanceof Error?error.message:'更新采购状态失败') } }
  const useSuggestion = (item: InventoryShortage) => { purchaseForm.setFieldsValue({yarnId:item.yarnId,yarnName:item.yarnName,targetGrams:item.suggestedGrams,estimatedPrice:0,priority:item.projectNames.length?'high':'normal',status:'pending'}); setPurchaseOpen(true) }
  const pendingItems=shoppingItems.filter((item)=>item.status==='pending')
  const shoppingBudget=pendingItems.reduce((total,item)=>total+item.estimatedPrice,0)
  return <main className="yarn-page">{contextHolder}<nav>
<button className="brand brand-button" onClick={onBack}>线<span>禾</span>
</button>
<div className="admin-nav">我的材料 <span>/ 线材库</span>
</div>
<Button type="text" onClick={onBack}>返回首页</Button>
</nav>
<section className="yarn-head">
<div>
<p className="eyebrow">MY YARN STASH</p>
<h1>我的线材</h1>
<p>把每一团线的来处、余量和成本记下来。</p>
</div>
<div className="yarn-head-actions"><Button onClick={() => { purchaseForm.setFieldsValue({targetGrams:100,estimatedPrice:0,priority:'normal',status:'pending'}); setPurchaseOpen(true) }}>采购清单 {pendingItems.length}</Button><Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => { form.setFieldsValue({ plies: 1, remainingGrams: 0, unitPrice: 0 }); setOpen(true) }}>录入线材</Button></div>
</section>
<section className="yarn-shelf">{loading ? <div className="yarn-loading">正在整理线材库…</div> : <><section className="purchase-overview"><div><p className="eyebrow">RESTOCK PLAN</p><h2>待补货 {pendingItems.length} 项</h2><span>预计预算 ¥ {shoppingBudget.toFixed(2)}</span></div><div className="purchase-suggestions">{shortages.length ? shortages.map((item)=><button key={item.yarnId} onClick={()=>useSuggestion(item)}><b>{item.yarnName}</b><span>余 {item.remainingGrams}g，建议补 {item.suggestedGrams}g{item.projectNames.length?` · ${item.projectNames.join('、')}`:''}</span><PlusOutlined/></button>) : <p>库存充足，暂无补货建议。</p>}</div></section>{yarns.length === 0 ? <Empty className="yarn-empty" image={Empty.PRESENTED_IMAGE_SIMPLE} description="你的线材库还是空的">
<Button type="primary" onClick={() => setOpen(true)}>录入第一团线</Button>
</Empty> : <div className="yarn-grid">{yarns.map((item) => <Card key={item.id} className="yarn-card" hoverable cover={item.photoData ? <img src={item.photoData} alt={item.name} /> : <div className="yarn-placeholder" style={{ backgroundColor: colorToSwatch(item.colorHex || item.color) }}>
<span>⌁</span>
</div>}>
<div className="yarn-card-body">
<div className="yarn-card-title">
<div>
<h3>{item.name}</h3>
<p>{item.brand || '未记录品牌'} {item.batchCode && `· ${item.batchCode}`}</p>
</div>
<Popconfirm title="移除线材" description={`确认从线材库移除“${item.name}”吗？`} onConfirm={() => removeYarn(item)} okText="移除" cancelText="取消">
<Button type="text" danger icon={<DeleteOutlined />} />
</Popconfirm>
</div>
<div className="yarn-tags">
<Tag>{item.color || item.colorHex || '未标注颜色'}</Tag>
<Tag>{item.material || '未标注材质'}</Tag>
</div>
<dl>
<div>
<dt>余量</dt>
<dd>{item.remainingGrams} g</dd>
</div>
<div>
<dt>单价</dt>
<dd>¥ {item.unitPrice}</dd>
</div>
</dl>
</div>
</Card>)}</div>}</>}</section>
<Modal open={open} title="录入一团线材" onCancel={closeModal} footer={null} destroyOnHidden>
<Form form={form} layout="vertical" initialValues={{ plies: 1, remainingGrams: 0, unitPrice: 0 }} onFinish={createYarn}>
<Form.Item label="线材照片"><div className="yarn-upload">{photoData ? <img src={photoData} alt="线材预览" /> : <Upload accept="image/*" maxCount={1} showUploadList={false} beforeUpload={onPhotoSelect}><Button icon={<UploadOutlined />}>选择照片</Button></Upload>}{photoData && <Button type="link" danger onClick={() => setPhotoData('')}>移除照片</Button>}</div></Form.Item>
<Form.Item name="name" label="线材名称" rules={[{ required: true }]}>
<Input />
</Form.Item>
<div className="two-fields">
<Form.Item name="brand" label="品牌">
<Input />
</Form.Item>
<Form.Item name="batchCode" label="批次">
<Input />
</Form.Item>
</div>
<div className="two-fields">
<Form.Item name="color" label="颜色">
<Input placeholder="奶油白" />
</Form.Item>
<Form.Item name="colorHex" label="标准色值">
<Input placeholder="#F6E7D4" />
</Form.Item>
</div>
<div className="two-fields">
<Form.Item name="remainingGrams" label="剩余克数">
<InputNumber min={0} addonAfter="g" style={{width:'100%'}} />
</Form.Item>
<Form.Item name="unitPrice" label="单价">
<InputNumber min={0} addonAfter="元" style={{width:'100%'}} />
</Form.Item>
</div>
<div className="two-fields">
<Form.Item name="material" label="材质">
<Input />
</Form.Item>
<Form.Item name="thickness" label="粗细">
<Input />
</Form.Item>
</div>
<Form.Item name="plies" label="股数">
<InputNumber min={1} max={50} addonAfter="股" style={{ width: '100%' }} />
</Form.Item>
<div className="modal-actions">
<Button onClick={closeModal}>取消</Button>
<Button type="primary" htmlType="submit" loading={submitting}>保存线材</Button>
</div>
</Form>
</Modal>
<Modal open={purchaseOpen} title="采购清单" onCancel={()=>{setPurchaseOpen(false);purchaseForm.resetFields()}} footer={null} destroyOnHidden>
<Form form={purchaseForm} layout="vertical" onFinish={createShoppingItem} initialValues={{targetGrams:100,estimatedPrice:0,priority:'normal',status:'pending'}}>
<Form.Item name="yarnId" hidden><InputNumber /></Form.Item>
<Form.Item name="yarnName" label="材料名称" rules={[{required:true,message:'请输入材料名称'}]}><Input placeholder="例如：云朵棉奶油白" /></Form.Item>
<div className="two-fields"><Form.Item name="targetGrams" label="计划购买" rules={[{required:true}]}><InputNumber min={1} addonAfter="g" style={{width:'100%'}} /></Form.Item><Form.Item name="estimatedPrice" label="预计预算" rules={[{required:true}]}><InputNumber min={0} addonAfter="元" style={{width:'100%'}} /></Form.Item></div>
<Form.Item name="priority" label="优先级"><Select options={[{value:'high',label:'高：项目即将需要'},{value:'normal',label:'普通'},{value:'low',label:'低'}]} /></Form.Item>
<Form.Item name="status" hidden><Input /></Form.Item>
<div className="modal-actions"><Button onClick={()=>setPurchaseOpen(false)}>取消</Button><Button type="primary" htmlType="submit">加入清单</Button></div>
</Form>
<div className="shopping-list">{shoppingItems.map((item)=><div key={item.id} className={item.status==='purchased'?'purchased':''}><div><b>{item.yarnName}</b><span>{item.targetGrams}g · ¥ {item.estimatedPrice.toFixed(2)} · {({high:'高优先级',normal:'普通',low:'低优先级'} as Record<string,string>)[item.priority]}</span></div><Button type={item.status==='pending'?'primary':'default'} onClick={()=>togglePurchased(item)}>{item.status==='pending'?'标为已购':'恢复待购'}</Button></div>)}</div>
</Modal>
</main>
}

function colorToSwatch(color: string) { return /^#[0-9a-fA-F]{6}$/.test(color.trim()) ? color.trim() : '#d9b49a' }

function ToolLibrary({ onBack }: { onBack: () => void }) {
  const [tools, setTools] = useState<Tool[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [photoData, setPhotoData] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm<ToolForm>()
  const [messageApi, contextHolder] = message.useMessage()
  const loadTools = () => { setLoading(true); fetch('/api/tools').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => setTools(data)).catch(() => messageApi.error('无法加载工具库，请重新登录后再试')).finally(() => setLoading(false)) }
  useEffect(() => { loadTools() }, [])
  const closeModal = () => { setOpen(false); setPhotoData(''); form.resetFields() }
  const uploadMedia = async (file: File) => { if (!file.type.startsWith('image/')) throw new Error('请上传图片文件'); if (file.size > 8 * 1024 * 1024) throw new Error('照片不能超过 8 MB'); const body = new FormData(); body.append('file', file); const response = await fetch('/api/media/images', { method: 'POST', body }); const payload = await response.json().catch(() => ({})) as { data?: { url: string }; error?: string }; if (!response.ok || !payload.data?.url) throw new Error(payload.error ?? '图片上传失败'); return payload.data.url }
  const onPhotoSelect = (file: File) => { void uploadMedia(file).then(setPhotoData).catch((error) => messageApi.error(error instanceof Error ? error.message : '图片上传失败')); return false }
  const createTool = async (values: ToolForm) => { setSubmitting(true); try { const response = await fetch('/api/tools', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...values, photoData }) }); const payload = await response.json().catch(() => ({})) as { data?: Tool; error?: string }; if (!response.ok || !payload.data) throw new Error(payload.error ?? '添加工具失败'); pushUndo({label:`已添加工具「${payload.data.name}」`,undo:async()=>{await fetch(`/api/tools/${payload.data!.id}`,{method:'DELETE'});loadTools()}}); messageApi.success('工具已收进你的工具库'); closeModal(); loadTools() } catch (error) { messageApi.error(error instanceof Error ? error.message : '添加工具失败') } finally { setSubmitting(false) } }
  const removeTool = async (item: Tool) => { try { const response = await fetch(`/api/tools/${item.id}`, { method: 'DELETE' }); if (!response.ok) { const payload = await response.json().catch(() => ({})) as { error?: string }; throw new Error(payload.error ?? '删除工具失败') }; pushUndo({label:`已移除工具「${item.name}」`,undo:async()=>{const result=await fetch('/api/tools',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(item)});if(!result.ok)throw new Error('恢复工具失败');loadTools()}}); messageApi.success(`已移除 ${item.name}`); loadTools() } catch (error) { messageApi.error(error instanceof Error ? error.message : '删除工具失败') } }
  return <main className="tool-page">{contextHolder}<nav>
<button className="brand brand-button" onClick={onBack}>线<span>禾</span>
</button>
<div className="admin-nav">我的材料 <span>/ 工具库</span>
</div>
<Button type="text" onClick={onBack}>返回首页</Button>
</nav>
<section className="tool-head">
<div>
<p className="eyebrow">MY TOOL KIT</p>
<h1>我的工具</h1>
<p>记录每一支钩针、棒针和辅助工具。开工时，随时确认你手边有什么、还缺什么。</p>
</div>
<Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => setOpen(true)}>录入工具</Button>
</section>
<section className="tool-shelf">{loading ? <div className="yarn-loading">正在整理工具库…</div> : tools.length === 0 ? <Empty className="yarn-empty" image={Empty.PRESENTED_IMAGE_SIMPLE} description="你的工具库还是空的">
<Button type="primary" onClick={() => setOpen(true)}>录入第一件工具</Button>
</Empty> : <div className="tool-grid">{tools.map((item) => <Card key={item.id} className="tool-card" hoverable cover={item.photoData ? <img src={item.photoData} alt={item.name} /> : <div className="tool-placeholder">
<span>{toolGlyph(item.type)}</span>
</div>}>
<div className="tool-card-body">
<div className="yarn-card-title">
<div>
<h3>{item.name}</h3>
<p>{item.type || '未标注类型'}</p>
</div>
<Popconfirm title="移除工具" description={`确认从工具库移除“${item.name}”吗？`} onConfirm={() => removeTool(item)} okText="移除" cancelText="取消">
<Button type="text" danger icon={<DeleteOutlined />} aria-label={`删除 ${item.name}`} />
</Popconfirm>
</div>
<div className="tool-thickness">
<span>规格 / 粗细</span>
<strong>{item.thickness || '—'}</strong>
</div>
</div>
</Card>)}</div>}</section>
<Modal open={open} title="录入一件工具" onCancel={closeModal} footer={null} destroyOnHidden>
<Form form={form} layout="vertical" onFinish={createTool}>
<Form.Item label="照片">
<div className="yarn-upload">{photoData ? <img src={photoData} alt="工具预览" /> : <Upload accept="image/png,image/jpeg,image/webp" maxCount={1} showUploadList={false} beforeUpload={onPhotoSelect}>
<Button icon={<UploadOutlined />}>选择照片</Button>
</Upload>}{photoData && <Button type="link" danger onClick={() => setPhotoData('')}>移除照片</Button>}</div>
</Form.Item>
<Form.Item name="name" label="工具名称" rules={[{ required: true, message: '请输入工具名称' }]}>
<Input placeholder="例如：樱花 4.0mm 钩针" maxLength={120} />
</Form.Item>
<Form.Item name="type" label="类型">
<Input placeholder="例如：钩针、棒针、剪刀、记号扣" maxLength={80} />
</Form.Item>
<Form.Item name="thickness" label="粗细 / 规格">
<Input placeholder="例如：4.0 mm" maxLength={80} />
</Form.Item>
<div className="modal-actions">
<Button onClick={closeModal}>取消</Button>
<Button type="primary" htmlType="submit" loading={submitting}>保存工具</Button>
</div>
</Form>
</Modal>
</main>
}

function toolGlyph(type: string) { if (type.includes('剪')) return '✂'; if (type.includes('针')) return '⌁'; if (type.includes('尺')) return '⌗'; return '✦' }

function LoginPage({ onLogin }: { onLogin: (user: User) => void }) {
  const [form] = Form.useForm<any>()
  const [submitting, setSubmitting] = useState(false)
  const [mode, setMode] = useState<'login' | 'register' | 'forgot' | 'reset'>(() => window.location.pathname === '/reset-password' ? 'reset' : 'login')
  const [messageApi, contextHolder] = message.useMessage()
  const token = new URLSearchParams(window.location.search).get('token') ?? ''
  useEffect(() => {
    if (window.location.pathname !== '/verify-email') return
    const verify = async () => {
      const response = await fetch(`/api/auth/verify?token=${encodeURIComponent(token)}`, { method: 'POST' })
      const payload = await response.json().catch(() => ({})) as { error?: string }
      if (response.ok) messageApi.success('邮箱验证成功，现在可以登录了')
      else messageApi.error(payload.error ?? '邮箱验证失败')
      window.history.replaceState({}, '', '/')
    }
    if (token) void verify()
  }, [token, messageApi])
  const submit = async (values: any) => {
    setSubmitting(true)
    try {
      const endpoint = mode === 'login' ? '/api/auth/login' : mode === 'register' ? '/api/auth/register' : mode === 'forgot' ? '/api/auth/password-reset/request' : '/api/auth/password-reset/confirm'
      const body = mode === 'reset' ? { token, password: values.password } : values
      const response = await fetch(endpoint, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      const payload = await response.json().catch(() => ({})) as { data?: User; error?: string }
      if (!response.ok) throw new Error(payload.error ?? '操作失败')
      if (mode === 'login' && payload.data) onLogin(payload.data)
      else if (mode === 'register') { messageApi.success('注册成功，请查收验证邮件'); setMode('login'); form.resetFields() }
      else if (mode === 'forgot') messageApi.success('如果邮箱已注册，重置链接已发送')
      else { messageApi.success('密码已重置，请登录'); setMode('login'); form.resetFields() }
    } catch (error) { messageApi.error(error instanceof Error ? error.message : '操作失败') } finally { setSubmitting(false) }
  }
  const title = mode === 'login' ? '登录线禾' : mode === 'register' ? '创建线禾账号' : mode === 'forgot' ? '找回密码' : '设置新密码'
  return <main className="login-page">{contextHolder}<section className="login-copy">
<a className="brand" href="#top">线<span>禾</span>
</a>
<p className="eyebrow">YOUR CROCHET WORKBENCH</p>
<h1>从一针一线，<br />回到自己的节奏。</h1>
<p>记录项目、管理材料、阅读图解。登录后，继续你的钩织手帐。</p>
<div className="login-mark">⌁</div>
</section>
<section className="login-panel">
<div className="login-card">
<p className="eyebrow">WELCOME BACK</p>
<h2>{title}</h2>
<p className="login-subtitle">{mode === 'login' ? '使用你的邮箱和密码继续。' : mode === 'register' ? '注册后验证邮箱即可开始记录作品。' : mode === 'forgot' ? '输入注册邮箱，我们会发送重置链接。' : '请输入至少 8 个字符的新密码。'}</p>
<Form form={form} layout="vertical" onFinish={submit}>
{(mode === 'login' || mode === 'register' || mode === 'forgot') && <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
<Input size="large" placeholder="name@example.com" autoComplete="email" />
</Form.Item>}
{mode === 'register' && <Form.Item name="name" label="姓名" rules={[{ required: true, min: 2, message: '请输入姓名' }]}><Input size="large" placeholder="例如：林小满" /></Form.Item>}
{(mode === 'login' || mode === 'register' || mode === 'reset') && <Form.Item name="password" label="密码" rules={[{ required: true, min: mode === 'login' ? 1 : 8, message: mode === 'login' ? '请输入密码' : '密码至少 8 个字符' }]}>
<Input.Password size="large" prefix={<LockOutlined />} placeholder="输入密码" autoComplete="current-password" />
</Form.Item>}
<Button type="primary" size="large" block htmlType="submit" loading={submitting} icon={mode === 'login' ? <LoginOutlined /> : undefined}>{mode === 'login' ? '登录并继续' : mode === 'register' ? '注册账号' : mode === 'forgot' ? '发送重置链接' : '确认新密码'}</Button>
</Form>
<div className="login-switches">{mode === 'login' && <><Button type="link" onClick={() => setMode('register')}>创建账号</Button><Button type="link" onClick={() => setMode('forgot')}>忘记密码</Button></>}{mode !== 'login' && <Button type="link" onClick={() => setMode('login')}>返回登录</Button>}</div>
</div>
</section>
</main>
}

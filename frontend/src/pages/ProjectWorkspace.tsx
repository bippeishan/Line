import { useEffect, useRef, useState } from 'react'
import { getDocument, GlobalWorkerOptions } from 'pdfjs-dist/legacy/build/pdf.mjs'
import pdfWorker from 'pdfjs-dist/build/pdf.worker.min.mjs?url'
import { InboxOutlined, DeleteOutlined, EditOutlined, ExpandOutlined, FileAddOutlined, FilePdfOutlined, FullscreenOutlined, PlusOutlined, RotateRightOutlined, UploadOutlined, ZoomInOutlined, ZoomOutOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, Modal, Popconfirm, Select, Switch, Tabs, Tag, Upload, message } from 'antd'
import { pendingMutationCount, syncEvent, syncOfflineMutations } from '../offline'
import { pushUndo } from '../undo'

GlobalWorkerOptions.workerSrc = pdfWorker

type Tool = { id: number; photoData: string; name: string; type: string; thickness: string; createdAt: string }
type Yarn = { id:number; name:string; remainingGrams:number; unitPrice:number }
type Project = { id:number; version:number; coverData:string; completionPhotoData:string; name:string; patternName:string; patternFileId?:number; targetSize:string; hookSize:string; yarnNames:string[]; toolNames:string[]; palette:string[]; workTags:string[]; shareToken:string; shareShowMaterials:boolean; shareShowNote:boolean; status:'planning'|'in_progress'|'completed'|'archived'; currentRow:number; totalRows:number; breakpointNote:string; completedAt?:string; completionNote:string; actualMinutes:number; targetDate?:string; weeklyFocusMinutes:number; createdAt:string; materialCost:number }
type ProjectForm = { coverData:string; completionPhotoData:string; name:string; patternName:string; patternFileId?:number; targetSize:string; hookSize:string; yarnNames:string[]; toolNames:string[]; palette:string[]; workTags:string[]; status:'planning'|'in_progress'|'completed'|'archived'; currentRow:number; totalRows:number; breakpointNote:string; completionNote:string; actualMinutes:number; targetDate?:string; weeklyFocusMinutes:number }
type PatternFile = { id:number; title:string; fileData:string; fileType:'pdf'|'image'; createdAt:string }
type PatternAnnotation = { id:number; pageNumber:number; rowNumber:number; x:number; y:number; note:string; createdAt:string }
type StitchTerm = { id:number; code:string; nameZh:string; nameEn:string; category:string; description:string; instruction:string; caution:string; isFavorite:boolean }
type ProjectStep = { id:number; title:string; isDone:boolean; sortOrder:number }
type ProjectNote = { id:number; body:string; photoData:string; createdAt:string }
type ProjectActivity = { id:number; kind:string; detail:string; createdAt:string }
type AuditEvent = { id:number; action:string; resourceType:string; resourceId:number; details:Record<string,unknown>; createdAt:string }
type ProjectReminder = { id:number; targetRow:number; message:string; isDone:boolean }
type CalculationRecord = { id:number; projectId?:number; mode:'circle'|'size'|'yarn'; input:Record<string,number>; result:Record<string,unknown>; formula:string; createdAt:string }
type ProjectYarnUsage = { id:number; yarnId:number; yarnName:string; grams:number; cost:number; createdAt:string }
type ProjectAccessory = { id:number; name:string; quantity:number; unit:string; unitPrice:number; cost:number; createdAt:string }
type ProjectYarnPlan = { yarnId:number; yarnName:string; plannedGrams:number; usedGrams:number; remainingGrams:number }
type ProjectTimer = { id:number; startedAt:string; stoppedAt?:string; secondsElapsed:number }
type ProjectForecast = { remainingRows:number; remainingMinutes?:number; recentMinutes:number; weeklyTargetMinutes:number; requiredWeeklyMinutes?:number; forecastDate?:string; targetDate?:string; status:'needs_data'|'at_risk'|'on_track'|'complete'; detail:string }
type ProjectReadinessItem = { key:string; label:string; ready:boolean; detail:string; yarnId?:number; missingGrams?:number }
type ProjectConflict = { local: Omit<ProjectForm,'targetDate'> & { targetDate?: string | null }; server: Project; choices: Record<string,'local'|'server'> }
function formatHours(minutes: number) { return `${Math.round(minutes / 60 * 10) / 10} 小时` }
const conflictFields = [
  { key:'name', label:'项目名称' }, { key:'patternName', label:'图解名称' }, { key:'targetSize', label:'目标尺寸' }, { key:'hookSize', label:'针号' },
  { key:'status', label:'项目状态' }, { key:'currentRow', label:'当前行数' }, { key:'totalRows', label:'总行数' }, { key:'targetDate', label:'目标完成日期' },
  { key:'weeklyFocusMinutes', label:'每周专注目标' }, { key:'yarnNames', label:'关联线材' }, { key:'toolNames', label:'关联工具' }, { key:'workTags', label:'作品标签' },
  { key:'breakpointNote', label:'断点备注' }, { key:'completionNote', label:'完成感想' }, { key:'actualMinutes', label:'实际用时' }, { key:'palette', label:'配色' },
] as const
function readData(file: File, done: (data: string) => void) { if (file.size > 8 * 1024 * 1024) return false; const reader = new FileReader(); reader.onload = () => done(String(reader.result)); reader.readAsDataURL(file); return false }
function formatFocusTime(seconds: number) { const hours=Math.floor(seconds/3600); const minutes=Math.floor(seconds%3600/60); const rest=seconds%60; return `${hours?`${String(hours).padStart(2,'0')}:`:''}${String(minutes).padStart(2,'0')}:${String(rest).padStart(2,'0')}` }

export default function ProjectWorkspace({ onBack, initialProjectId, onInitialProjectOpened }: { onBack: () => void; initialProjectId: number | null; onInitialProjectOpened: () => void }) {
  const [projects,setProjects]=useState<Project[]>([]),[yarns,setYarns]=useState<Yarn[]>([]),[tools,setTools]=useState<Tool[]>([]),[files,setFiles]=useState<PatternFile[]>([]),[selected,setSelected]=useState<Project|null>(null),[editor,setEditor]=useState(false),[archiveView,setArchiveView]=useState(false),[archiveSearch,setArchiveSearch]=useState(''),[archiveTag,setArchiveTag]=useState<string|undefined>(),[archiveYarn,setArchiveYarn]=useState<string|undefined>(),[archivePeriod,setArchivePeriod]=useState<'all'|'week'|'month'|'year'>('all'),[reader,setReader]=useState<PatternFile|null>(null),[annotations,setAnnotations]=useState<PatternAnnotation[]>([]),[steps,setSteps]=useState<ProjectStep[]>([]),[notes,setNotes]=useState<ProjectNote[]>([]),[activities,setActivities]=useState<ProjectActivity[]>([]),[reminders,setReminders]=useState<ProjectReminder[]>([]),[calculations,setCalculations]=useState<CalculationRecord[]>([]),[yarnUsages,setYarnUsages]=useState<ProjectYarnUsage[]>([]),[accessories,setAccessories]=useState<ProjectAccessory[]>([]),[yarnPlans,setYarnPlans]=useState<ProjectYarnPlan[]>([]),[timers,setTimers]=useState<ProjectTimer[]>([]),[forecast,setForecast]=useState<ProjectForecast|null>(null),[clock,setClock]=useState(Date.now()),[shareOpen,setShareOpen]=useState(false),[readinessProject,setReadinessProject]=useState<Project|null>(null),[readinessItems,setReadinessItems]=useState<ProjectReadinessItem[]>([]),[readinessReady,setReadinessReady]=useState(false),[form]=Form.useForm<ProjectForm>(),[shareForm]=Form.useForm<{workTags:string[];showMaterials:boolean;showNote:boolean}>(),[messageApi,holder]=message.useMessage()
  const [auditEvents,setAuditEvents]=useState<AuditEvent[]>([])
  const [online,setOnline]=useState(navigator.onLine),[queuedCount,setQueuedCount]=useState(0),[syncBlocked,setSyncBlocked]=useState(false),[draftSavedAt,setDraftSavedAt]=useState<string>(),[undoState,setUndoState]=useState<Project|null>(null),[conflict,setConflict]=useState<ProjectConflict|null>(null)
	useEffect(()=>{const on=()=>setOnline(true),off=()=>setOnline(false);window.addEventListener('online',on);window.addEventListener('offline',off);return()=>{window.removeEventListener('online',on);window.removeEventListener('offline',off)}},[])
	useEffect(()=>{void pendingMutationCount().then(setQueuedCount);const update=(event:Event)=>{const detail=(event as CustomEvent<{pending:number;blocked?:boolean}>).detail;setQueuedCount(detail.pending);setSyncBlocked(!!detail.blocked)};window.addEventListener(syncEvent,update);return()=>window.removeEventListener(syncEvent,update)},[])
  const request=async(url:string,options?:RequestInit)=>{const r=await fetch(url,options);const p=await r.json().catch(()=>({}));if(!r.ok){const error=new Error(p.error||'操作失败') as Error & {code?:string;currentVersion?:number;currentProject?:Project};error.code=p.code;error.currentVersion=p.currentVersion;error.currentProject=p.currentProject;throw error}return p}
  const load=()=>{request('/api/projects').then(p=>setProjects(p.data)).catch(()=>messageApi.error('无法加载项目'));request('/api/yarns').then(p=>setYarns(p.data)).catch(()=>undefined);request('/api/tools').then(p=>setTools(p.data)).catch(()=>undefined);request('/api/pattern-files').then(p=>setFiles(p.data)).catch(()=>undefined)}
  useEffect(()=>{load()},[])
  const loadDetail=(p:Project)=>{setSelected(p);Promise.all([request(`/api/projects/${p.id}/steps`),request(`/api/projects/${p.id}/notes`),request(`/api/projects/${p.id}/activities`),request(`/api/projects/${p.id}/reminders`),request(`/api/projects/${p.id}/calculations`),request(`/api/projects/${p.id}/materials`),request(`/api/projects/${p.id}/yarn-plans`),request(`/api/projects/${p.id}/timers`),request(`/api/projects/${p.id}/forecast`),request('/api/audit')]).then(([a,b,c,d,e,f,g,h,i,j])=>{setSteps(a.data);setNotes(b.data);setActivities(c.data);setReminders(d.data);setCalculations(e.data);setYarnUsages(f.data.yarnUsages);setAccessories(f.data.accessories);setYarnPlans(g.data);setTimers(h.data);setForecast(i.data);setAuditEvents(j.data.filter((event:AuditEvent)=>event.resourceType==='project'&&event.resourceId===p.id));setClock(Date.now())}).catch(()=>messageApi.error('无法加载项目执行记录'))}
  useEffect(()=>{const active=timers.some((item)=>!item.stoppedAt);if(!active)return;const interval=window.setInterval(()=>setClock(Date.now()),1000);return()=>window.clearInterval(interval)},[timers])
  useEffect(()=>{if(!initialProjectId||projects.length===0)return;const project=projects.find((item)=>item.id===initialProjectId);if(project){loadDetail(project);onInitialProjectOpened()}},[initialProjectId,projects])
  const save=async(v:ProjectForm)=>{const payload={...v,targetDate:v.targetDate?(v.targetDate.includes('T')?v.targetDate:`${v.targetDate}T00:00:00Z`):null};const editing=!!selected;try{let result:any;if(editing){setUndoState(selected);result=await request(`/api/projects/${selected!.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...payload,version:selected!.version})})}else result=await request('/api/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});localStorage.removeItem(`line:project-draft:${selected?.id??'new'}`);setEditor(false);if(result.queued){messageApi.info('当前离线，修改已加入同步队列');void pendingMutationCount().then(setQueuedCount);return}if(editing)pushUndo({label:`已修改项目「${selected!.name}」`,undo:undoProject});messageApi.success(editing?'项目已更新':'项目已创建');load();if(selected)loadDetail({...selected,...payload,version:selected.version+1,targetDate:payload.targetDate??undefined})}catch(e){const conflictError=e as Error & {code?:string;currentProject?:Project};if(conflictError.code==='conflict'&&conflictError.currentProject){const choices=Object.fromEntries(conflictFields.map((field)=>[field.key,'local'])) as Record<string,'local'|'server'>;setConflict({local:payload,server:conflictError.currentProject,choices});return}messageApi.error(e instanceof Error?e.message:'保存失败')}}
  const resolveConflict=async()=>{if(!conflict)return;const merged:any={...conflict.server};conflictFields.forEach((field)=>{if(conflict.choices[field.key]==='local')merged[field.key]=(conflict.local as any)[field.key]});try{await request(`/api/projects/${conflict.server.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...merged,version:conflict.server.version})});setConflict(null);setEditor(false);messageApi.success('已合并并保存项目修改');load();loadDetail({...merged,version:conflict.server.version+1})}catch(e){messageApi.error(e instanceof Error?e.message:'合并保存失败')}}
  const openNew=()=>{form.resetFields();form.setFieldsValue({coverData:'',completionPhotoData:'',yarnNames:[],toolNames:[],palette:[],workTags:[],status:'planning',currentRow:0,totalRows:0,breakpointNote:'',completionNote:'',actualMinutes:0,weeklyFocusMinutes:0});setEditor(false);setSelected(null);setEditor(true)}
  const openEdit=(p:Project)=>{const draft=localStorage.getItem(`line:project-draft:${p.id}`);form.setFieldsValue(draft?JSON.parse(draft):{...p,targetDate:p.targetDate?.slice(0,10)});setDraftSavedAt(draft?new Date().toISOString():undefined);setSelected(p);setEditor(true)}
  const undoProject=async()=>{if(!selected||!undoState)return;try{await request(`/api/projects/${selected.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...undoState,version:selected.version})});setUndoState(null);messageApi.success('已撤销上一次项目修改');load();loadDetail({...undoState,version:selected.version+1})}catch(e){messageApi.error(e instanceof Error?e.message:'撤销失败')}}
  const updateProgress=async(p:Project,delta:number)=>{try{const current=Math.max(0,Math.min(p.totalRows,p.currentRow+delta));const status=current>0&&p.status==='planning'?'in_progress':current===p.totalRows&&p.totalRows>0?'completed':p.status;await request(`/api/projects/${p.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...p,currentRow:current,status})});const reached=selected?.id===p.id?reminders.filter((item)=>!item.isDone&&item.targetRow===current):[];reached.forEach((item)=>messageApi.info(`第 ${item.targetRow} 行提醒：${item.message}`,5));load();if(selected?.id===p.id)loadDetail({...p,currentRow:current,status})}catch(e){messageApi.error(e instanceof Error?e.message:'更新进度失败')}}
  const archive=async(p:Project)=>{try{await request(`/api/projects/${p.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...p,status:'archived'})});messageApi.success('项目已归档');setSelected(null);load()}catch(e){messageApi.error(e instanceof Error?e.message:'归档失败')}}
  const remove=async(p:Project)=>{try{await request(`/api/projects/${p.id}`,{method:'DELETE'});setSelected(null);load()}catch(e){messageApi.error(e instanceof Error?e.message:'删除失败')}}
  const openReadiness=async(p:Project)=>{try{const response=await request(`/api/projects/${p.id}/readiness`);setReadinessProject(p);setReadinessItems(response.data.items);setReadinessReady(response.data.ready)}catch(e){messageApi.error(e instanceof Error?e.message:'无法完成开工检查')}}
  const addReadinessPurchase=async(item:ProjectReadinessItem)=>{if(!item.yarnId||!item.missingGrams||!readinessProject)return;const yarn=yarns.find((value)=>value.id===item.yarnId);try{await request('/api/shopping-list',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({yarnId:item.yarnId,yarnName:yarn?.name??item.label,targetGrams:item.missingGrams,estimatedPrice:0,priority:'high',status:'pending'})});messageApi.success(`${item.label} 已加入采购清单`);setReadinessItems((items)=>items.map((value)=>value.key===item.key?{...value,detail:`${value.detail} · 已加入采购清单`}:value))}catch(e){messageApi.error(e instanceof Error?e.message:'加入采购清单失败')}}
  const startProject=async()=>{if(!readinessProject||!readinessReady)return;try{await request(`/api/projects/${readinessProject.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...readinessProject,status:'in_progress'})});messageApi.success('项目已开始');setReadinessProject(null);load()}catch(e){messageApi.error(e instanceof Error?e.message:'无法开始项目')}}
  const startTimer=async()=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/timers`,{method:'POST'});messageApi.success('专注计时已开始');loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'无法开始计时')}}
  const stopTimer=async(timer:ProjectTimer)=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/timers/${timer.id}/stop`,{method:'POST'});messageApi.success('本次专注已记录');load();loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'无法停止计时')}}
  const duplicateProject=async(p:Project)=>{try{await request('/api/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({coverData:p.coverData,name:`${p.name} · 新项目`,patternName:p.patternName,patternFileId:p.patternFileId,targetSize:p.targetSize,hookSize:p.hookSize,yarnNames:p.yarnNames,toolNames:p.toolNames,palette:p.palette,workTags:p.workTags,totalRows:p.totalRows,breakpointNote:''})});messageApi.success('已复制为新的计划项目');setArchiveView(false);load()}catch(e){messageApi.error(e instanceof Error?e.message:'复制项目失败')}}
  const uploadMedia=async(file:File)=>{if(!file.type.startsWith('image/'))throw new Error('请上传图片文件');if(file.size>8*1024*1024)throw new Error('图片不能超过 8 MB');const body=new FormData();body.append('file',file);const response=await fetch('/api/media/images',{method:'POST',body});const payload=await response.json().catch(()=>({})) as {data?:{url:string};error?:string};if(!response.ok||!payload.data?.url)throw new Error(payload.error??'图片上传失败');return payload.data.url}
  const uploadCover=(file:File)=>{void uploadMedia(file).then((url)=>form.setFieldValue('coverData',url)).catch((e)=>messageApi.error(e instanceof Error?e.message:'图片上传失败'));return false}
  const uploadCompletionPhoto=(file:File)=>{void uploadMedia(file).then((url)=>form.setFieldValue('completionPhotoData',url)).catch((e)=>messageApi.error(e instanceof Error?e.message:'图片上传失败'));return false}
  const uploadPattern=(file:File)=>readData(file,async(data)=>{try{await request('/api/pattern-files',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:file.name,fileData:data,fileType:file.type==='application/pdf'?'pdf':'image'})});messageApi.success('图解已加入文件库');load()}catch(e){messageApi.error(e instanceof Error?e.message:'图解上传失败')}})
  const openPatternReader=async(file:PatternFile)=>{try{const response=await request(`/api/pattern-files/${file.id}/annotations`);setAnnotations(response.data);setReader(file)}catch(e){messageApi.error(e instanceof Error?e.message:'无法读取图解标注')}}
  const addAnnotation=async(values:Omit<PatternAnnotation,'id'|'createdAt'>)=>{if(!reader)return;try{await request(`/api/pattern-files/${reader.id}/annotations`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)});const response=await request(`/api/pattern-files/${reader.id}/annotations`);setAnnotations(response.data);messageApi.success('图解标记已保存')}catch(e){messageApi.error(e instanceof Error?e.message:'保存标记失败')}}
  const deleteAnnotation=async(item:PatternAnnotation)=>{if(!reader)return;try{await request(`/api/pattern-files/${reader.id}/annotations/${item.id}`,{method:'DELETE'});setAnnotations((items)=>items.filter((value)=>value.id!==item.id))}catch(e){messageApi.error(e instanceof Error?e.message:'删除标记失败')}}
  const addStep=async(title:string)=>{if(!selected||!title.trim())return;await request(`/api/projects/${selected.id}/steps`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title})});loadDetail(selected)}
  const toggleStep=async(s:ProjectStep)=>{if(!selected)return;await request(`/api/projects/${selected.id}/steps/${s.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...s,isDone:!s.isDone})});loadDetail(selected)}
  const addNote=async(values:{body:string})=>{if(!selected)return;await request(`/api/projects/${selected.id}/notes`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)});loadDetail(selected)}
  const useYarn=async(values:{yarnId:number;grams:number})=>{if(!selected)return;await request(`/api/projects/${selected.id}/usages`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)});messageApi.success('材料用量已记录，并更新线材余量');load();loadDetail(selected)}
  const saveYarnPlan=async(values:{yarnId:number;plannedGrams:number})=>{if(!selected)return;try{const existing=yarnPlans.filter((item)=>item.yarnId!==values.yarnId).map((item)=>({yarnId:item.yarnId,plannedGrams:item.plannedGrams}));await request(`/api/projects/${selected.id}/yarn-plans`,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({plans:[...existing,values]})});messageApi.success('预计用量已保存');loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'保存预计用量失败')}}
  const removeYarnPlan=async(plan:ProjectYarnPlan)=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/yarn-plans`,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({plans:yarnPlans.filter((item)=>item.yarnId!==plan.yarnId).map((item)=>({yarnId:item.yarnId,plannedGrams:item.plannedGrams}))})});loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'移除预计用量失败')}}
  const addReminder=async(values:{targetRow:number;message:string})=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/reminders`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)});messageApi.success('关键行提醒已保存');loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'保存提醒失败')}}
  const toggleReminder=async(item:ProjectReminder)=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/reminders/${item.id}`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({...item,isDone:!item.isDone})});loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'更新提醒失败')}}
  const deleteReminder=async(item:ProjectReminder)=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/reminders/${item.id}`,{method:'DELETE'});loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'删除提醒失败')}}
  const deleteCalculation=async(item:CalculationRecord)=>{if(!selected)return;try{await request(`/api/calculations/${item.id}`,{method:'DELETE'});messageApi.success('计算记录已删除');loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'删除失败')}}
  const addAccessory=async(values:{name:string;quantity:number;unit:string;unitPrice:number})=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/accessories`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)});messageApi.success('辅料已计入项目成本');load();loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'添加辅料失败')}}
  const deleteAccessory=async(item:ProjectAccessory)=>{if(!selected)return;try{await request(`/api/projects/${selected.id}/accessories/${item.id}`,{method:'DELETE'});load();loadDetail(selected)}catch(e){messageApi.error(e instanceof Error?e.message:'删除辅料失败')}}
  const linkedPattern=selected?.patternFileId ? files.find((file)=>file.id===selected.patternFileId) : undefined
  const materialTotal = yarnUsages.reduce((total, item) => total + item.cost, 0) + accessories.reduce((total, item) => total + item.cost, 0)
  const activeTimer = timers.find((item)=>!item.stoppedAt)
  const activeTimerSeconds = activeTimer ? activeTimer.secondsElapsed + Math.max(0,Math.floor((clock-new Date(activeTimer.startedAt).getTime())/1000)) : 0
  const archiveTags=[...new Set(projects.flatMap((item)=>item.workTags??[]))].sort()
  const archiveYarns=[...new Set(projects.flatMap((item)=>item.yarnNames??[]))].sort()
  const archiveSince = archivePeriod==='week' ? new Date(Date.now()-7*24*60*60*1000) : archivePeriod==='month' ? new Date(Date.now()-30*24*60*60*1000) : archivePeriod==='year' ? new Date(Date.now()-365*24*60*60*1000) : null
  const visibleProjects = projects.filter((item) => { if (!archiveView) return item.status === 'planning' || item.status === 'in_progress'; const searchable=`${item.name} ${item.patternName} ${(item.workTags??[]).join(' ')} ${(item.yarnNames??[]).join(' ')}`.toLocaleLowerCase(); const date=item.completedAt ? new Date(item.completedAt) : new Date(item.createdAt); return (item.status === 'completed' || item.status === 'archived') && (!archiveSearch.trim() || searchable.includes(archiveSearch.trim().toLocaleLowerCase())) && (!archiveTag || item.workTags?.includes(archiveTag)) && (!archiveYarn || item.yarnNames?.includes(archiveYarn)) && (!archiveSince || date >= archiveSince) }).sort((a,b)=>archiveView ? new Date(b.completedAt||b.createdAt).getTime()-new Date(a.completedAt||a.createdAt).getTime() : 0)
  const exportProjectPDF = async () => { if (!selected) return; try { const response = await fetch('/api/projects/' + selected.id + '/export.pdf'); if (!response.ok) throw new Error('导出失败'); const url = URL.createObjectURL(await response.blob()); const link = document.createElement('a'); link.href = url; link.download = selected.name + '.pdf'; link.click(); URL.revokeObjectURL(url) } catch (error) { messageApi.error(error instanceof Error ? error.message : '导出失败') } }
  const openShareSettings = () => { if (!selected) return; shareForm.setFieldsValue({workTags:selected.workTags ?? [],showMaterials:selected.shareShowMaterials ?? true,showNote:selected.shareShowNote ?? true}); setShareOpen(true) }
  const copyShareLink = async (token: string) => { const url=`${window.location.origin}/share/${token}`; try { await navigator.clipboard.writeText(url); messageApi.success('只读分享链接已复制') } catch { messageApi.info(`分享链接：${url}`) } }
  const shareProject = async (values:{workTags:string[];showMaterials:boolean;showNote:boolean}) => { if (!selected || selected.status !== 'completed') return; try { const response=await request(`/api/projects/${selected.id}/share`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(values)}); const updated={...selected,workTags:values.workTags,shareShowMaterials:values.showMaterials,shareShowNote:values.showNote,shareToken:response.data.token}; setSelected(updated); setProjects((items)=>items.map((item)=>item.id===updated.id?updated:item)); setShareOpen(false); await copyShareLink(response.data.token) } catch (error) { messageApi.error(error instanceof Error?error.message:'无法生成分享链接') } }
  const closeShare = async () => { if (!selected) return; try { await request(`/api/projects/${selected.id}/share`,{method:'DELETE'}); const updated={...selected,shareToken:''}; setSelected(updated); setProjects((items)=>items.map((item)=>item.id===updated.id?updated:item)); messageApi.success('分享链接已关闭') } catch (error) { messageApi.error(error instanceof Error?error.message:'无法关闭分享链接') } }
  return <main className="project-page">{holder}<nav>
<button className="brand brand-button" onClick={onBack}>线<span>禾</span>
</button>
<div className="admin-nav">我的创作 <span>/ 项目工作台</span>
</div>
<Button type="text" onClick={onBack}>返回首页</Button>
</nav>
<section className="project-head">
<div>
<p className="eyebrow">MY MAKES</p>
<h1>我的项目</h1>
<p>从图解、材料到停针处，都留在同一件作品里。</p>
{(!online||queuedCount>0)&&<div className="offline-notice"><span>{online?'待同步':'离线工作中'}</span><b>{queuedCount?`${queuedCount} 项修改待同步`:'本机草稿会在联网后同步'}</b>{online&&queuedCount>0&&<Button type="link" onClick={()=>void syncOfflineMutations()}>立即同步</Button>}{syncBlocked&&<small>存在冲突，请重新打开项目并确认最新内容。</small>}</div>}
</div>
<Button type="primary" size="large" icon={<PlusOutlined/>} onClick={openNew}>新建项目</Button>
</section>
<section className="project-shelf">
<div className="project-view-toggle">
<Button type={archiveView?'default':'primary'} onClick={()=>setArchiveView(false)}>进行中 {projects.filter((item)=>item.status==='planning'||item.status==='in_progress').length}</Button>
<Button type={archiveView?'primary':'default'} onClick={()=>setArchiveView(true)}>作品归档 {projects.filter((item)=>item.status==='completed'||item.status==='archived').length}</Button>
</div>{archiveView&&<div className="archive-filters"><Input allowClear value={archiveSearch} onChange={(event)=>setArchiveSearch(event.target.value)} placeholder="搜索作品、图解、标签或线材"/><Select allowClear value={archiveTag} onChange={setArchiveTag} placeholder="全部标签" options={archiveTags.map((tag)=>({value:tag,label:tag}))}/><Select allowClear value={archiveYarn} onChange={setArchiveYarn} placeholder="全部线材" options={archiveYarns.map((yarn)=>({value:yarn,label:yarn}))}/><Select value={archivePeriod} onChange={setArchivePeriod} options={[{value:'all',label:'全部时间'},{value:'week',label:'最近 7 天'},{value:'month',label:'最近 30 天'},{value:'year',label:'最近一年'}]}/></div>}{projects.length===0?<Empty className="yarn-empty" description="还没有项目记录">
<Button type="primary" onClick={openNew}>从第一针开始</Button>
</Empty>:visibleProjects.length===0?<Empty className="yarn-empty" description={archiveView?'还没有完成的作品':'还没有进行中的项目'}/>:<div className="project-grid">{visibleProjects.map(p=>
<Card key={p.id} className="project-card" cover={(archiveView?(p.completionPhotoData||p.coverData):p.coverData)?<img src={archiveView?(p.completionPhotoData||p.coverData):p.coverData} alt=""/>:undefined}>
<div className="project-card-top">
<Tag color={p.status==='completed'?'green':p.status==='in_progress'?'volcano':p.status==='archived'?'default':'gold'}>{({planning:'计划中',in_progress:'进行中',completed:'已完成',archived:'已归档'} as Record<string,string>)[p.status]}</Tag>
<div>
<Button type="text" icon={<EditOutlined/>} onClick={()=>openEdit(p)} aria-label="编辑项目"/>
<Popconfirm title="删除项目" onConfirm={()=>remove(p)}>
<Button type="text" danger icon={<DeleteOutlined/>} aria-label="删除项目"/>
</Popconfirm>
</div>
</div>
<button className="project-open" onClick={()=>loadDetail(p)}>
<h3>{p.name}</h3>
<p>{p.patternName||'未关联图解'} · {p.hookSize||'未填针号'}</p>
</button>
<div className="project-progress">
<span>第 {p.currentRow} / {p.totalRows||'—'} 行</span>
<div>
<i style={{width:`${p.totalRows?Math.min(100,p.currentRow/p.totalRows*100):0}%`}}/>
</div>
</div>
<div className="project-links">{archiveView&&p.workTags?.map(x=><Tag className="work-tag" key={`tag-${x}`}>{x}</Tag>)}{p.yarnNames.map(x=>
<Tag key={x}>{x}</Tag>)}{p.toolNames.map(x=>
<Tag key={x}>{x}</Tag>)}{p.palette?.length>0&&<span className="project-palette">{p.palette.map((color)=><i key={color} style={{backgroundColor:color}} title={color}/>)}</span>}</div>
<div className="project-actions">
{archiveView?<><Button onClick={()=>loadDetail(p)}>查看档案</Button><Button type="primary" onClick={()=>duplicateProject(p)}>复制为新项目</Button></>:p.status==='planning'?<><Button onClick={()=>openReadiness(p)}>开工检查</Button><Button type="primary" onClick={()=>openReadiness(p)}>开始项目</Button></>:<><Button onClick={()=>updateProgress(p,-1)} disabled={p.currentRow===0||p.status==='archived'}>− 1 行</Button>
<Button type="primary" onClick={()=>updateProgress(p,1)} disabled={p.status==='archived'||(p.totalRows>0&&p.currentRow>=p.totalRows)}>＋ 1 行</Button></>}
</div>{p.status==='completed'&&<div className="completion-summary">
<span>{p.completedAt?new Date(p.completedAt).toLocaleDateString():'已完成'}</span>
<b>{p.actualMinutes?`${Math.round(p.actualMinutes/60)} 小时`:'未记录用时'}</b>
<p>{p.completionNote||'这件作品已经完成。'}</p>
</div>}{p.breakpointNote&&<p className="breakpoint">断点：{p.breakpointNote}</p>}</Card>)}</div>}</section>
<Modal open={editor} title={selected?'编辑项目':'新建项目'} onCancel={()=>setEditor(false)} footer={null} destroyOnHidden>
<Form form={form} layout="vertical" onValuesChange={(_, values)=>{localStorage.setItem(`line:project-draft:${selected?.id??'new'}`,JSON.stringify(values));setDraftSavedAt(new Date().toISOString())}} onFinish={save}>
<Form.Item name="coverData" hidden>
<Input/>
</Form.Item>
<Form.Item label="项目封面">
<Upload accept="image/*" showUploadList={false} beforeUpload={uploadCover}>
<Button icon={<UploadOutlined/>}>上传封面</Button>
</Upload>
</Form.Item>
<Form.Item name="name" label="项目名称" rules={[{required:true,message:'请输入项目名称'}]}>
<Input/>
</Form.Item>
<div className="two-fields">
<Form.Item name="patternName" label="图解名称">
<Input placeholder="例如：小熊玩偶"/>
</Form.Item>
<Form.Item name="patternFileId" label="关联图解文件">
<Select allowClear options={files.map(f=>({value:f.id,label:f.title}))} placeholder="从文件库选择"/>
</Form.Item>
</div>
<Upload accept="application/pdf,image/*" showUploadList={false} beforeUpload={uploadPattern}>
<Button type="link" icon={<FileAddOutlined/>}>上传新图解到文件库</Button>
</Upload>
<div className="two-fields">
<Form.Item name="targetSize" label="目标尺寸">
<Input placeholder="例如：高 18cm"/>
</Form.Item>
<Form.Item name="hookSize" label="针号">
<Input placeholder="例如：4.0 mm"/>
</Form.Item>
</div>
<div className="two-fields">
<Form.Item name="totalRows" label="总行数">
<InputNumber min={0} style={{width:'100%'}}/>
</Form.Item>
<Form.Item name="status" label="状态">
<Select options={[{value:'planning',label:'计划中'},{value:'in_progress',label:'进行中'},{value:'completed',label:'已完成'},{value:'archived',label:'已归档'}]}/>
</Form.Item>
</div>
<div className="two-fields">
<Form.Item name="targetDate" label="目标完成日期"><Input type="date" /></Form.Item>
<Form.Item name="weeklyFocusMinutes" label="每周专注目标"><InputNumber min={0} addonAfter="分钟" style={{width:'100%'}} /></Form.Item>
</div>
<Form.Item name="yarnNames" label="关联线材">
<Select mode="multiple" options={yarns.map(x=>({value:x.name,label:x.name}))}/>
</Form.Item>
<Form.Item name="toolNames" label="关联工具">
<Select mode="multiple" options={tools.map(x=>({value:x.name,label:x.name}))}/>
</Form.Item>
<Form.Item name="workTags" label="作品标签">
<Select mode="tags" tokenSeparators={[',','，']} placeholder="例如：玩偶、送礼、春日" />
</Form.Item>
<Form.Item name="breakpointNote" label="断点 / 备注">
<Input.TextArea rows={2}/>
</Form.Item>
<section className="completion-fields">
<p className="eyebrow">FINISHING RECORD</p>
<Form.Item name="completionPhotoData" hidden>
<Input/>
</Form.Item>
<Form.Item label="成品照">
<Upload accept="image/*" showUploadList={false} beforeUpload={uploadCompletionPhoto}>
<Button icon={<UploadOutlined/>}>上传成品照</Button>
</Upload>
</Form.Item>
<div className="two-fields">
<Form.Item name="actualMinutes" label="实际用时">
<InputNumber min={0} addonAfter="分钟" style={{width:'100%'}}/>
</Form.Item>
<Form.Item name="completionNote" label="完成感想">
<Input.TextArea rows={2} placeholder="记录这件作品完成时的想法"/>
</Form.Item>
</div>
</section>
<div className="modal-actions">
<Button onClick={()=>setEditor(false)}>取消</Button>
<span className="draft-status">{online?'在线':'离线'}{draftSavedAt?` · 草稿 ${new Date(draftSavedAt).toLocaleTimeString()}`:''}</span>
<Button type="primary" htmlType="submit">保存项目</Button>
</div>
</Form>
</Modal>
<Modal open={!!conflict} title="合并项目修改" onCancel={()=>setConflict(null)} footer={<div className="modal-actions"><Button onClick={()=>setConflict(null)}>暂不处理</Button><Button type="primary" onClick={resolveConflict}>确认合并并保存</Button></div>} width={760}>{conflict&&<section className="conflict-panel"><p>此项目在你编辑期间已被其他窗口更新。请为每一项保留本地修改或服务端最新值。</p>{conflictFields.filter((field)=>JSON.stringify((conflict.local as any)[field.key])!==JSON.stringify((conflict.server as any)[field.key])).map((field)=>{const local=String((conflict.local as any)[field.key]??'未填写');const server=String((conflict.server as any)[field.key]??'未填写');return <article key={field.key}><b>{field.label}</b><div><Button type={conflict.choices[field.key]==='local'?'primary':'default'} onClick={()=>setConflict({...conflict,choices:{...conflict.choices,[field.key]:'local'}})}><small>我的修改</small>{local}</Button><Button type={conflict.choices[field.key]==='server'?'primary':'default'} onClick={()=>setConflict({...conflict,choices:{...conflict.choices,[field.key]:'server'}})}><small>服务器最新值</small>{server}</Button></div></article>})}</section>}</Modal>
<Modal open={!!selected&&!editor} title={selected?.name} onCancel={()=>setSelected(null)} footer={null} width={860}>{selected&&<>
<Tabs items={[{key:'progress',label:'执行',children:<section className="workspace-tab">
<div className="workspace-summary">
<b>第 {selected.currentRow} / {selected.totalRows||'—'} 行</b>
<span>{selected.breakpointNote||'还没有断点备注'}</span>
</div>
{forecast&&<section className={`project-forecast ${forecast.status}`}><div><span>{({on_track:'进度正常',at_risk:'可能延期',needs_data:'等待数据',complete:'已完成'} as Record<string,string>)[forecast.status]}</span><strong>{forecast.detail}</strong>{forecast.remainingMinutes!==undefined&&<small>剩余约 {formatHours(forecast.remainingMinutes)} · 近 7 天已专注 {forecast.recentMinutes} 分钟</small>}</div><Button type="text" onClick={()=>openEdit(selected)}>调整计划</Button></section>}
<section className="focus-timer"><div><span>FOCUS TIMER</span><strong>{activeTimer?formatFocusTime(activeTimerSeconds):`${selected.actualMinutes} 分钟`}</strong><small>{activeTimer?'本次专注进行中':'累计专注时长'}</small></div>{selected.status==='in_progress'&&(activeTimer?<Button danger onClick={()=>stopTimer(activeTimer)}>结束专注</Button>:<Button type="primary" onClick={startTimer}>开始专注</Button>)}</section>
<div className="timer-history">{timers.filter((item)=>item.stoppedAt).slice(0,3).map((item)=><span key={item.id}>{new Date(item.startedAt).toLocaleDateString('zh-CN')} · {formatFocusTime(item.secondsElapsed)}</span>)}</div>
<Form layout="inline" onFinish={addStep}>
<Form.Item name="title" rules={[{required:true}]}>
<Input placeholder="添加下一步"/>
</Form.Item>
<Button htmlType="submit">添加步骤</Button>
</Form>
<div className="steps-list">{steps.map(s=>
<div key={s.id}>
<Button type={s.isDone?'primary':'default'} onClick={()=>toggleStep(s)}>{s.isDone?'已完成':'待完成'}</Button>
<span>{s.title}</span>
</div>)}</div>
</section>},{key:'reminders',label:'提醒',children:<section className="workspace-tab">
<Form layout="inline" onFinish={addReminder}>
<Form.Item name="targetRow" rules={[{required:true,message:'请输入行数'}]}>
<InputNumber min={1} addonBefore="第" addonAfter="行" placeholder="关键行"/>
</Form.Item>
<Form.Item name="message" rules={[{required:true,message:'请输入提醒内容'}]}>
<Input placeholder="例如：换色并填充头部"/>
</Form.Item>
<Button type="primary" htmlType="submit">添加提醒</Button>
</Form>
<div className="reminder-list">{reminders.length===0?<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有关键行提醒"/>:reminders.map((item)=>
<div key={item.id} className={item.isDone?'done':''}>
<span>第 {item.targetRow} 行</span>
<p>{item.message}</p>
<Button type={item.isDone?'default':'primary'} onClick={()=>toggleReminder(item)}>{item.isDone?'恢复':'完成'}</Button>
<Popconfirm title="删除这条提醒？" onConfirm={()=>deleteReminder(item)}>
<Button type="text" danger icon={<DeleteOutlined/>} aria-label="删除提醒"/>
</Popconfirm>
</div>)}</div>
</section>},{key:'notes',label:'笔记',children:<section className="workspace-tab">
<Form layout="vertical" onFinish={addNote}>
<Form.Item name="body" rules={[{required:true,message:'请输入笔记'}]}>
<Input.TextArea rows={3} placeholder="记录当前行、改动或下次继续的位置"/>
</Form.Item>
<Button htmlType="submit" type="primary">保存笔记</Button>
</Form>
<div className="notes-list">{notes.map(n=>
<article key={n.id}>
<time>{new Date(n.createdAt).toLocaleString()}</time>
<p>{n.body}</p>{n.photoData&&<img src={n.photoData} alt="笔记照片"/>}</article>)}</div>
 </section>},{key:'pattern',label:'图解',children:<section className="workspace-tab pattern-tab">{linkedPattern ? <>
<div className="pattern-linked-card">
<div>
<span>{linkedPattern.fileType==='pdf'?'PDF 图解':'图片图解'}</span>
<strong>{linkedPattern.title}</strong>
</div>
<Button type="primary" icon={<ExpandOutlined/>} onClick={()=>openPatternReader(linkedPattern)}>打开阅读器</Button>
</div>
<p>阅读进度与项目进度可并列查看；图片支持缩放、旋转和全屏。</p>
</> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="尚未关联图解文件">
<Button icon={<EditOutlined/>} onClick={()=>openEdit(selected)}>关联图解</Button>
</Empty>}</section>},{key:'materials',label:'材料',children:<section className="workspace-tab">
<section className="material-plan"><div className="material-plan-head"><div><p className="eyebrow">MATERIAL BUDGET</p><h3>预计用量</h3></div><span>按项目计划核对开工前缺口</span></div>
<Form layout="inline" onFinish={saveYarnPlan}><Form.Item name="yarnId" rules={[{required:true}]}><Select style={{minWidth:180}} placeholder="选择关联线材" options={yarns.filter((item)=>selected.yarnNames.includes(item.name)).map((item)=>({value:item.id,label:item.name}))}/></Form.Item><Form.Item name="plannedGrams" rules={[{required:true}]}><InputNumber min={1} addonAfter="g" placeholder="预计用量"/></Form.Item><Button type="primary" htmlType="submit">保存预计</Button></Form>
<div className="material-plan-list">{yarnPlans.length===0?<p>尚未填写预计用量。开工前先为关联线材设定需求。</p>:yarnPlans.map((plan)=>{const gap=Math.max(0,plan.plannedGrams-plan.usedGrams-plan.remainingGrams);return <div key={plan.yarnId}><b>{plan.yarnName}</b><span>预计 {plan.plannedGrams}g</span><span>已用 {plan.usedGrams}g</span><span>余 {plan.remainingGrams}g</span><strong className={gap?'shortage':''}>{gap?`还差 ${gap}g`:'库存可覆盖'}</strong><Button type="text" danger icon={<DeleteOutlined/>} onClick={()=>removeYarnPlan(plan)} aria-label={`移除 ${plan.yarnName} 的预计用量`}/></div>})}</div></section>
<div className="material-total">
<span>项目材料成本</span>
<strong>¥ {materialTotal.toFixed(2)}</strong>
</div>{yarns.filter((item)=>selected.yarnNames.includes(item.name)&&item.remainingGrams<20).map((item)=>
<Alert key={item.id} className="material-low-stock" type="warning" showIcon message={`${item.name} 仅剩 ${item.remainingGrams}g`} description="低于 20g，请确认是否需要补货。"/>)}<h3 className="material-section-title">线材用量</h3>
<Form layout="inline" onFinish={useYarn}>
<Form.Item name="yarnId" rules={[{required:true}]}>
<Select style={{minWidth:180}} placeholder="选择线材" options={yarns.map(y=>({value:y.id,label:`${y.name} · 余 ${y.remainingGrams}g`}))}/>
</Form.Item>
<Form.Item name="grams" rules={[{required:true}]}>
<InputNumber min={0.01} addonAfter="g" placeholder="使用量"/>
</Form.Item>
<Button type="primary" htmlType="submit">记录用量</Button>
</Form>
<div className="material-history">{yarnUsages.length===0?<p>尚未记录线材消耗。</p>:yarnUsages.map((item)=>
<div key={item.id}>
<span>{item.yarnName}</span>
<span>{item.grams}g</span>
<b>¥ {item.cost.toFixed(2)}</b>
</div>)}</div>
<h3 className="material-section-title">辅料</h3>
<Form layout="inline" initialValues={{quantity:1,unit:'个',unitPrice:0}} onFinish={addAccessory}>
<Form.Item name="name" rules={[{required:true,message:'请输入辅料'}]}>
<Input placeholder="例如：安全眼"/>
</Form.Item>
<Form.Item name="quantity" rules={[{required:true}]}>
<InputNumber min={0.01} placeholder="数量"/>
</Form.Item>
<Form.Item name="unit">
<Input placeholder="单位"/>
</Form.Item>
<Form.Item name="unitPrice" rules={[{required:true}]}>
<InputNumber min={0} addonAfter="元" placeholder="单价"/>
</Form.Item>
<Button htmlType="submit">添加辅料</Button>
</Form>
<div className="material-history">{accessories.length===0?<p>尚未记录辅料。</p>:accessories.map((item)=>
<div key={item.id}>
<span>{item.name}</span>
<span>{item.quantity} {item.unit || '个'}</span>
<b>¥ {item.cost.toFixed(2)}</b>
<Popconfirm title="删除这项辅料？" onConfirm={()=>deleteAccessory(item)}>
<Button type="text" danger icon={<DeleteOutlined/>} aria-label="删除辅料"/>
</Popconfirm>
</div>)}</div>
</section>},{key:'calculations',label:'计算',children:<section className="workspace-tab">
<div className="calculation-history">{calculations.length===0?<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="尚未保存与该项目关联的计算"/>:calculations.map((item)=>
<article key={item.id}>
<div>
<span>{({circle:'圆形加针',size:'尺寸换算',yarn:'用线量估算'} as Record<string,string>)[item.mode]}</span>
<time>{new Date(item.createdAt).toLocaleString()}</time>
</div>
<strong>{item.formula}</strong>
<dl>{Object.entries(item.result).filter(([key])=>key!=='rounds'&&key!=='formula').map(([key,value])=>
<div key={key}>
<dt>{calculationLabel(key)}</dt>
<dd>{String(value)}</dd>
</div>)}</dl>
<Popconfirm title="删除这条计算记录？" onConfirm={()=>deleteCalculation(item)}>
<Button type="text" danger icon={<DeleteOutlined/>}>删除</Button>
</Popconfirm>
</article>)}</div>
</section>},{key:'activity',label:'活动',children:<section className="workspace-tab">
<div className="activity-list">{activities.map(a=>
<div key={a.id}>
<time>{new Date(a.createdAt).toLocaleString()}</time>
<span>{a.detail}</span>
</div>)}</div>
</section>},{key:'audit',label:'审计',children:<section className="workspace-tab"><div className="activity-list">{auditEvents.length?auditEvents.map((event)=><div key={event.id}><time>{new Date(event.createdAt).toLocaleString()}</time><span>{event.action} · 版本 {String(event.details?.version??'—')}</span></div>):<p>暂无审计记录</p>}</div></section>}]}/>
<div className="workspace-footer">
 {selected.status==='completed'&&<Button onClick={openShareSettings}>{selected.shareToken?'管理分享':'分享作品'}</Button>}
<Button icon={<FilePdfOutlined/>} onClick={exportProjectPDF}>导出 PDF</Button>
<Button icon={<EditOutlined/>} onClick={()=>openEdit(selected)}>编辑项目</Button>
{undoState&&<Button onClick={undoProject}>撤销上次修改</Button>}
<Popconfirm title="归档后仍可查看，需要恢复时再编辑状态" onConfirm={()=>archive(selected)}>
<Button icon={<InboxOutlined/>} disabled={selected.status==='archived'}>归档</Button>
</Popconfirm>
</div>
	</>}</Modal>
		<Modal open={shareOpen} title="分享作品" onCancel={()=>setShareOpen(false)} footer={null} destroyOnHidden>
			<Form form={shareForm} layout="vertical" onFinish={shareProject} initialValues={{workTags:[],showMaterials:true,showNote:true}}>
				<p className="share-modal-intro">分享页仅展示作品资料，不提供评论、点赞或关注入口。</p>
				<Form.Item name="workTags" label="作品标签"><Select mode="tags" tokenSeparators={[',','，']} placeholder="例如：玩偶、送礼、春日" /></Form.Item>
				<Form.Item name="showMaterials" label="展示材料" valuePropName="checked"><Switch checkedChildren="显示" unCheckedChildren="隐藏" /></Form.Item>
				<Form.Item name="showNote" label="展示完成感想" valuePropName="checked"><Switch checkedChildren="显示" unCheckedChildren="隐藏" /></Form.Item>
				<div className="modal-actions">{selected?.shareToken&&<Popconfirm title="关闭分享链接？" description="已复制的链接将无法再打开。" onConfirm={closeShare} okText="关闭" cancelText="取消"><Button danger>关闭链接</Button></Popconfirm>}<Button type="primary" htmlType="submit">{selected?.shareToken?'更新并复制链接':'生成并复制链接'}</Button></div>
			</Form>
		</Modal>
		<Modal open={!!readinessProject} title={`${readinessProject?.name ?? ''} · 开工检查`} onCancel={()=>setReadinessProject(null)} footer={null} destroyOnHidden>
			<div className="readiness-list">{readinessItems.map((item)=><div key={item.key} className={item.ready?'ready':'blocked'}><span>{item.ready?'已就绪':'待完善'}</span><div><b>{item.label}</b><p>{item.detail}</p></div>{!item.ready&&item.yarnId&&item.missingGrams?<Button size="small" onClick={()=>addReadinessPurchase(item)}>加入采购清单</Button>:null}</div>)}</div>
			<div className="modal-actions"><Button onClick={()=>setReadinessProject(null)}>稍后再说</Button><Button type="primary" disabled={!readinessReady} onClick={startProject}>确认开始项目</Button></div>
		</Modal>
		<PatternReader file={reader} annotations={annotations} currentRow={selected?.currentRow ?? 0} onCreateAnnotation={addAnnotation} onDeleteAnnotation={deleteAnnotation} onClose={()=>setReader(null)}/>
	</main>
}


function calculationLabel(key: string) { return ({ finalStitches:'最终针数', castOnStitches:'建议起针', rows:'建议行数', stitchGaugePerCM:'每厘米针数', rowGaugePerCM:'每厘米行数', baseGrams:'理论用量(g)', recommendedGrams:'建议用量(g)', recommendedSkeins:'建议团数' } as Record<string,string>)[key] ?? key }

function PatternReader({ file, annotations, currentRow, onCreateAnnotation, onDeleteAnnotation, onClose }: { file: PatternFile | null; annotations: PatternAnnotation[]; currentRow: number; onCreateAnnotation: (value: Omit<PatternAnnotation,'id'|'createdAt'>) => Promise<void>; onDeleteAnnotation: (value: PatternAnnotation) => Promise<void>; onClose: () => void }) {
  const [zoom, setZoom] = useState(1)
  const [rotation, setRotation] = useState(0)
  const [pageNumber, setPageNumber] = useState(1)
  const [pending, setPending] = useState<{x:number;y:number}|null>(null)
  const [annotationForm] = Form.useForm<{rowNumber:number;note:string}>()
  useEffect(() => { setZoom(1); setRotation(0); setPageNumber(1); setPending(null); annotationForm.resetFields() }, [file?.id])
  const fullscreen = () => document.getElementById('project-pattern-reader')?.requestFullscreen?.()
  const selectPoint = (event: React.MouseEvent<HTMLDivElement>) => { if (!file || file.fileType !== 'image') return; const rect=event.currentTarget.getBoundingClientRect(); setPending({x:Math.min(1,Math.max(0,(event.clientX-rect.left)/rect.width)),y:Math.min(1,Math.max(0,(event.clientY-rect.top)/rect.height))}); annotationForm.setFieldsValue({rowNumber:currentRow||1,note:''}) }
  const saveAnnotation = async (values:{rowNumber:number;note:string}) => { if (!pending) return; await onCreateAnnotation({pageNumber:file?.fileType === 'pdf' ? pageNumber : 1,rowNumber:values.rowNumber,x:pending.x,y:pending.y,note:values.note}); setPending(null); annotationForm.resetFields() }
  return <Modal open={!!file} title={file?.title} onCancel={onClose} footer={null} width="min(1180px, calc(100vw - 32px))" wrapClassName="pattern-reader-modal">
<div id="project-pattern-reader" className="pattern-reader">
<div className="pattern-reader-tools">
<span>{file?.fileType === 'pdf' ? 'PDF 阅读器 · 页码备注' : `${Math.round(zoom * 100)}% · 点击图解添加标记`}</span>{file?.fileType === 'pdf' && <InputNumber min={1} value={pageNumber} onChange={(value)=>setPageNumber(Number(value)||1)} addonBefore="第" addonAfter="页" aria-label="PDF 页码"/>}{file?.fileType === 'image' && <>
<Button type="text" icon={<ZoomOutOutlined/>} onClick={()=>setZoom((value)=>Math.max(.5,value-.15))} aria-label="缩小"/>
<Button type="text" icon={<ZoomInOutlined/>} onClick={()=>setZoom((value)=>Math.min(2.5,value+.15))} aria-label="放大"/>
<Button type="text" icon={<RotateRightOutlined/>} onClick={()=>setRotation((value)=>(value+90)%360)} aria-label="旋转"/>
</>}<Button type="text" icon={<PlusOutlined/>} onClick={()=>{setPending({x:.5,y:.1});annotationForm.setFieldsValue({rowNumber:currentRow||1,note:''})}}>添加备注</Button>
<Button type="text" icon={<FullscreenOutlined/>} onClick={fullscreen}>全屏</Button>
</div>
<div className="pattern-reader-layout">
<div className="pattern-reader-canvas">{file?.fileType === 'pdf' ? <PDFPageCanvas file={file} pageNumber={pageNumber} onPageChange={setPageNumber} onSelectPoint={(point)=>{setPending(point);annotationForm.setFieldsValue({rowNumber:currentRow||1,note:''})}}/> : file && <div className="pattern-image-stage" onClick={selectPoint} style={{transform:`scale(${zoom}) rotate(${rotation}deg)`}}>
<img src={file.fileData} alt={file.title}/>{annotations.map((item)=>
<button key={item.id} className={`pattern-marker ${item.rowNumber===currentRow?'current':''}`} style={{left:`${item.x*100}%`,top:`${item.y*100}%`}} onClick={(event)=>event.stopPropagation()} title={`第 ${item.rowNumber||'—'} 行：${item.note}`}>{item.rowNumber||'•'}</button>)}</div>}</div>
<aside className="annotation-panel">{pending ? <Form form={annotationForm} layout="vertical" onFinish={saveAnnotation}>
<p className="eyebrow">NEW MARKER</p>
<Form.Item name="rowNumber" label="关联行号" rules={[{required:true,min:1,message:'请输入行号'}]}>
<InputNumber min={1} style={{width:'100%'}}/>
</Form.Item>
<Form.Item name="note" label="备注" rules={[{required:true,message:'请输入备注'}]}>
<Input.TextArea rows={3} placeholder="例如：在这里换成浅色线"/>
</Form.Item>
<div className="annotation-actions">
<Button onClick={()=>setPending(null)}>取消</Button>
<Button type="primary" htmlType="submit">保存标记</Button>
</div>
</Form> : <>
<p className="eyebrow">MARKERS · {annotations.filter((item)=>file?.fileType !== 'pdf' || item.pageNumber === pageNumber).length}</p>
<div className="annotation-list">{annotations.filter((item)=>file?.fileType !== 'pdf' || item.pageNumber === pageNumber).length===0?<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前页尚未添加标记"/>:annotations.filter((item)=>file?.fileType !== 'pdf' || item.pageNumber === pageNumber).map((item)=>
<article key={item.id} className={item.rowNumber===currentRow?'current':''}>
<b>第 {item.rowNumber||'—'} 行</b>
<p>{item.note}</p>
<Button type="text" danger icon={<DeleteOutlined/>} onClick={()=>onDeleteAnnotation(item)} aria-label="删除标记"/>
</article>)}</div>
<StitchDictionaryPanel/>
</>}</aside>
</div>
</div>
</Modal>
}

function PDFPageCanvas({ file, pageNumber, onPageChange, onSelectPoint }: { file: PatternFile; pageNumber: number; onPageChange: (page: number) => void; onSelectPoint: (point: { x: number; y: number }) => void }) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [pageCount, setPageCount] = useState(1)
  const [error, setError] = useState('')
  useEffect(() => {
    let cancelled = false
    const render = async () => {
      try {
        setError('')
        const bytes = new Uint8Array(await (await fetch(file.fileData)).arrayBuffer())
        const document = await getDocument({ data: bytes }).promise
        if (cancelled) return
        setPageCount(document.numPages)
        const safePage = Math.min(Math.max(1, pageNumber), document.numPages)
        if (safePage !== pageNumber) onPageChange(safePage)
        const page = await document.getPage(safePage)
        const viewport = page.getViewport({ scale: 1.35 })
        const canvas = canvasRef.current
        if (!canvas || cancelled) return
        canvas.width = Math.ceil(viewport.width)
        canvas.height = Math.ceil(viewport.height)
        const context = canvas.getContext('2d')
        if (!context) return
        await page.render({ canvas, canvasContext: context, viewport }).promise
      } catch {
        if (!cancelled) setError('无法渲染该 PDF 页面')
      }
    }
    render()
    return () => { cancelled = true }
  }, [file.id, file.fileData, pageNumber, onPageChange])
  if (error) return <Empty description={error} />
  return <div className="pdf-page-canvas"><div className="pdf-page-controls"><Button onClick={() => onPageChange(Math.max(1, pageNumber - 1))} disabled={pageNumber <= 1}>上一页</Button><span>第 {pageNumber} / {pageCount} 页</span><Button onClick={() => onPageChange(Math.min(pageCount, pageNumber + 1))} disabled={pageNumber >= pageCount}>下一页</Button></div><canvas ref={canvasRef} onClick={(event) => { const rect = event.currentTarget.getBoundingClientRect(); onSelectPoint({ x: (event.clientX - rect.left) / rect.width, y: (event.clientY - rect.top) / rect.height }) }} /></div>
}

function StitchDictionaryPanel() {
  const [terms, setTerms] = useState<StitchTerm[]>([])
  const [query, setQuery] = useState('')
  const [category, setCategory] = useState('')
  const [favoritesOnly, setFavoritesOnly] = useState(false)
  const [selected, setSelected] = useState<StitchTerm | null>(null)
  useEffect(() => { const params = new URLSearchParams(); if (query) params.set('q', query); if (category) params.set('category', category); fetch(`/api/stitch-terms?${params}`).then((response)=>response.ok?response.json():Promise.reject()).then(({data})=>setTerms(data)).catch(()=>setTerms([])) }, [query, category])
  const toggleFavorite = async (term:StitchTerm) => { const method=term.isFavorite?'DELETE':'POST'; const response=await fetch(`/api/stitch-terms/${term.id}/favorite`,{method}); if (!response.ok) return; const next={...term,isFavorite:!term.isFavorite}; setTerms((items)=>items.map((item)=>item.id===term.id?next:item)); if(selected?.id===term.id)setSelected(next) }
  const visible = favoritesOnly ? terms.filter((term)=>term.isFavorite) : terms
  return <section className="stitch-panel"><p className="eyebrow">STITCH DICTIONARY</p><Input value={query} onChange={(event)=>setQuery(event.target.value)} placeholder="搜索 sc、短针、加针..."/><div className="stitch-filters"><Select value={category} onChange={setCategory} options={[{value:'',label:'全部'},{value:'基础',label:'基础'},{value:'增减针',label:'增减针'},{value:'长针',label:'长针'},{value:'进阶',label:'进阶'}]}/><Button type={favoritesOnly?'primary':'default'} onClick={()=>setFavoritesOnly((value)=>!value)}>收藏</Button></div><div className="stitch-list">{visible.map((term)=><button key={term.id} className={selected?.id===term.id?'selected':''} onClick={()=>setSelected(term)}><b>{term.code}</b><span>{term.nameZh}</span><i>{term.isFavorite?'*':''}</i></button>)}</div>{selected&&<article className="stitch-detail"><header><div><b>{selected.code}</b><strong>{selected.nameZh}</strong></div><Button type="text" onClick={()=>toggleFavorite(selected)}>{selected.isFavorite?'取消收藏':'收藏'}</Button></header><p>{selected.description}</p><dl><div><dt>做法</dt><dd>{selected.instruction}</dd></div><div><dt>注意</dt><dd>{selected.caution}</dd></div></dl></article>}</section>
}

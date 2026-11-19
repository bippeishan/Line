import { useEffect, useState } from 'react'
import { Tag } from 'antd'

type PublicShare = { name:string; photoData:string; completionNote:string; actualMinutes:number; workTags:string[]; yarnNames:string[]; toolNames:string[]; showMaterials:boolean; showNote:boolean }

export default function PublicWorkPage({ token }: { token: string }) {
  const [work, setWork] = useState<PublicShare | null>(null)
  const [failed, setFailed] = useState(false)
  useEffect(() => { fetch(`/api/share/${token}`).then((response) => response.ok ? response.json() : Promise.reject()).then(({ data }) => setWork(data)).catch(() => setFailed(true)) }, [token])
  if (failed) return <main className="public-work public-work-empty"><p className="eyebrow">LINE / PRIVATE WORK</p><h1>这份作品分享已不可用</h1><p>链接可能已被关闭，或作品尚未公开。</p></main>
  if (!work) return <main className="public-work public-work-empty">正在打开作品档案...</main>
  return <main className="public-work"><nav><a className="brand" href="/">线<span>禾</span></a><span className="public-nav-label">PRIVATE WORK SHARE</span></nav><section className="public-work-hero"><div><p className="eyebrow">A HANDMADE FINISH</p><h1>{work.name}</h1><div className="public-tags">{work.workTags.map((tag) => <Tag key={tag}>{tag}</Tag>)}</div>{work.showNote && work.completionNote && <p className="public-note">{work.completionNote}</p>}<dl><div><dt>制作时长</dt><dd>{work.actualMinutes ? `${Math.round(work.actualMinutes / 60 * 10) / 10} 小时` : '未记录'}</dd></div>{work.showMaterials && <><div><dt>线材</dt><dd>{work.yarnNames.length ? work.yarnNames.join('、') : '未记录'}</dd></div><div><dt>工具</dt><dd>{work.toolNames.length ? work.toolNames.join('、') : '未记录'}</dd></div></>}</dl></div><div className="public-photo">{work.photoData ? <img src={work.photoData} alt={`${work.name} 成品照`} /> : <span>作品已完成</span>}</div></section><footer>线禾 · 一份只读的作品档案</footer></main>
}

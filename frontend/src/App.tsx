import { useEffect, useState } from 'react'

type Pattern = { id: number; title: string; category: string; difficulty: string; yarn: string; hookSize: string; durationMinutes: number }

const fallbackPatterns: Pattern[] = [
  { id: 1, title: '初夏小花束', category: '居家装饰', difficulty: 'beginner', yarn: '棉线 2 股', hookSize: '3.0 mm', durationMinutes: 90 },
  { id: 2, title: '奶油吐司收纳篮', category: '居家装饰', difficulty: 'intermediate', yarn: '再生棉绳', hookSize: '5.0 mm', durationMinutes: 180 },
  { id: 3, title: '海风渔夫帽', category: '穿搭配饰', difficulty: 'intermediate', yarn: '亚麻棉混纺', hookSize: '4.0 mm', durationMinutes: 240 },
]

const icons = ['✿', '⌁', '◒']
const difficulty = { beginner: '入门', intermediate: '进阶', advanced: '挑战' }

export default function App() {
  const [patterns, setPatterns] = useState<Pattern[]>(fallbackPatterns)
  const [online, setOnline] = useState(false)
  useEffect(() => { fetch('/api/patterns').then((r) => r.ok ? r.json() : Promise.reject()).then(({ data }) => { setPatterns(data); setOnline(true) }).catch(() => undefined) }, [])
  return <main>
    <nav><a className="brand" href="#top">线<span>禾</span></a><div className="nav-links"><a href="#patterns">找图解</a><a href="#projects">我的在钩</a><a href="#journal">灵感手帐</a></div><button className="round-button" aria-label="新建作品">＋</button></nav>
    <section className="hero" id="top">
      <div className="hero-copy"><p className="eyebrow">CROCHET, AT YOUR OWN PACE</p><h1>把时间<br /><em>织</em>进日常。</h1><p className="intro">线禾是你的钩织小宇宙。收藏图解、记录针目，也收藏每一段被双手温柔编织的时光。</p><a className="primary-link" href="#patterns">探索图解 <span>↘</span></a></div>
      <div className="hero-art" aria-label="毛线团装饰插图"><div className="loop loop-one" /><div className="loop loop-two" /><div className="yarn-ball"><i /><i /><i /></div><div className="needle" /><p>slow stitches<br />warm days</p></div>
      <aside className="hero-note"><span>本周进度</span><strong>12 <small>行</small></strong><p>海风渔夫帽<br />还有 38%</p><div className="progress"><i /></div></aside>
    </section>
    <section className="story-strip"><p>“手作不是抵达终点，而是让手与心重新相遇。”</p><span>— 线禾手帐</span></section>
    <section className="patterns" id="patterns"><header><div><p className="eyebrow">PATTERN SHELF <b className={online ? 'online' : ''}>{online ? '· 已连接工坊' : '· 本地预览'}</b></p><h2>今日想织点什么？</h2></div><a href="#patterns">查看全部 →</a></header><div className="pattern-grid">{patterns.map((item, index) => <article className={`pattern-card card-${index % 3}`} key={item.id}><div className="card-illustration"><span>{icons[index % icons.length]}</span><small>{item.category}</small></div><div className="card-details"><span>{difficulty[item.difficulty as keyof typeof difficulty] ?? item.difficulty} · {item.durationMinutes} 分钟</span><h3>{item.title}</h3><p>{item.yarn} &nbsp; / &nbsp; {item.hookSize}</p></div></article>)}</div></section>
    <section className="footer-card" id="projects"><div><p className="eyebrow">MAKE IT YOURS</p><h2>从第一针，<br />到属于你的作品。</h2></div><button>开始记录 <span>→</span></button></section>
  </main>
}

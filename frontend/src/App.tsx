import { useState } from 'react'

const modes = [
  {
    id: 'final-sprint',
    label: '📚 期末突击',
    title: '期末突击',
    description: '上传往年题，分析重点题型和考点，制定适合你的期末冲刺计划。',
  },
  {
    id: 'homework',
    label: '📝 日常作业辅助',
    title: '日常作业辅助',
    description: '上传作业题目，先学习解题思路和方法，再完成自己的答案。',
  },
  {
    id: 'general-question',
    label: '❓ 其它问题',
    title: '其它问题',
    description: '提出任何学习相关的问题，由通用学习 Agent 为你解答。',
  },
  {
    id: 'agent-settings',
    label: '🤖 Agent 设置',
    title: 'Agent 设置',
    description: '配置你的个人学习助手，调整回答风格和学习方式。',
  },
]

function App() {
  const [currentMode, setCurrentMode] = useState('final-sprint')
  const current = modes.find((m) => m.id === currentMode) || modes[0]

  return (
    <div className="app-layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-title">灵犀-Claw</div>
          <div className="sidebar-subtitle">Study with 灵犀-Claw</div>
        </div>
        <nav className="sidebar-nav">
          {modes.map((mode) => (
            <div
              key={mode.id}
              className={`sidebar-item ${currentMode === mode.id ? 'active' : ''}`}
              onClick={() => setCurrentMode(mode.id)}
            >
              {mode.label}
            </div>
          ))}
        </nav>
      </aside>
      <main className="main-content">
        <div className="mode-panel">
          <div className="mode-icon">{current.label.split(' ')[0]}</div>
          <h1>{current.title}</h1>
          <p className="subtitle">{current.description}</p>
        </div>
      </main>
    </div>
  )
}

export default App

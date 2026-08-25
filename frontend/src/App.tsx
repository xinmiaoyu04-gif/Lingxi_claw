import { useState, useRef } from 'react'

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
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [analysisMessage, setAnalysisMessage] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  if (currentMode === 'final-sprint') {
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

            <div
              className="upload-area"
              onClick={() => fileInputRef.current?.click()}
            >
              <div className="upload-hint">点击选择学习资料</div>
              <div className="upload-types">支持 PDF、图片、Word、文本文件</div>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden-input"
                onChange={(e) => {
                  const files = Array.from(e.target.files || [])
                  setSelectedFiles(files)
                  setAnalysisMessage('')
                }}
              />
            </div>

            {selectedFiles.length > 0 && (
              <div className="file-summary">
                <div className="file-summary-title">已选择 {selectedFiles.length} 个文件</div>
                <div className="file-list">
                  {selectedFiles.map((file, index) => (
                    <div key={index} className="file-item">
                      <div className="file-name">{file.name}</div>
                      <div className="file-size">{(file.size / 1024).toFixed(1)} KB</div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <button
              className="primary-button"
              disabled={selectedFiles.length === 0 || !!analysisMessage}
              onClick={() => setAnalysisMessage('资料已准备好，下一步将进行分析')}
            >
              开始分析资料
            </button>

            {analysisMessage && (
              <div className="analysis-message">{analysisMessage}</div>
            )}
          </div>
        </main>
      </div>
    )
  }

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

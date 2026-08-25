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

const analysisResult = {
  knowledgePoints: ['二重积分', '条件概率', '随机变量分布', '无穷级数'],
  questionTypes: [
    { name: '判别级数敛散性', level: '高频' },
    { name: '幂级数求和', level: '高频' },
    { name: '傅里叶展开', level: '中等' },
  ],
}

function App() {
  const [currentMode, setCurrentMode] = useState('final-sprint')
  const current = modes.find((m) => m.id === currentMode) || modes[0]
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [step, setStep] = useState<'idle' | 'analyzing' | 'analyzed'>('idle')
  const [days, setDays] = useState(3)
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

            {step === 'idle' && (
              <>
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
                  disabled={selectedFiles.length === 0}
                  onClick={() => {
                    setStep('analyzing')
                    setTimeout(() => setStep('analyzed'), 1500)
                  }}
                >
                  开始分析资料
                </button>
              </>
            )}

            {step === 'analyzing' && (
              <div className="analyze-loading">
                <div className="analyze-loading-text">正在分析学习资料...</div>
                <div className="analyze-loading-sub">正在处理 {selectedFiles.length} 份资料</div>
              </div>
            )}

            {step === 'analyzed' && (
              <div className="analyze-result">
                <div className="analyze-result-header">
                  <div className="analyze-result-title">资料分析完成</div>
                  <div className="analyze-result-sub">已分析 {selectedFiles.length} 份学习资料</div>
                </div>

                <div className="result-section">
                  <div className="result-title">高频考点</div>
                  <div className="result-list">
                    {analysisResult.knowledgePoints.map((item) => (
                      <div key={item} className="result-item">{item}</div>
                    ))}
                  </div>
                </div>

                <div className="result-section">
                  <div className="result-title">高频题型</div>
                  <div className="result-list">
                    {analysisResult.questionTypes.map((item) => (
                      <div key={item.name} className="result-item">
                        <span>{item.name}</span>
                        <span className="result-badge">{item.level}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="result-section">
                  <div className="result-title">你还有多少天准备考试？</div>
                  <input
                    className="days-input"
                    type="number"
                    min={1}
                    value={days}
                    onChange={(e) => setDays(Number(e.target.value))}
                  />
                </div>

                <div className="result-actions">
                  <button className="primary-button" onClick={() => alert('冲刺计划将在下一步生成')}>
                    生成我的冲刺计划
                  </button>
                  <button className="secondary-button" onClick={() => { setStep('idle'); setSelectedFiles([]) }}>
                    重新选择资料
                  </button>
                </div>
              </div>
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

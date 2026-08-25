import { useState, useRef, useEffect } from 'react'

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

const basePlans = [
  {
    title: '高频考点复习',
    items: ['二重积分', '条件概率', '完成基础计算题', '整理核心公式'],
  },
  {
    title: '综合题训练',
    items: ['随机变量分布', '无穷级数', '完成综合练习', '整理错题'],
  },
  {
    title: '最后冲刺',
    items: ['模拟考试', '高频题型回顾', '完成一套模拟题', '查漏补缺'],
  },
]

function buildPlans(days: number) {
  const safe = Math.max(1, days)
  const plans = basePlans.slice(0, Math.min(safe, 3))
  if (safe > 3) {
    for (let i = 4; i <= safe; i++) {
      plans.push({
        title: `第 ${i} 天：自主复习`,
        items: ['复习错题本', '回顾高频公式', '保持做题手感'],
      })
    }
  }
  return plans
}

function App() {
  const [currentMode, setCurrentMode] = useState('final-sprint')
  const current = modes.find((m) => m.id === currentMode) || modes[0]
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [step, setStep] = useState<'idle' | 'analyzing' | 'analyzed' | 'planning' | 'planned'>('idle')
  const [days, setDays] = useState(3)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [homeworkFiles, setHomeworkFiles] = useState<File[]>([])
  const [homeworkStep, setHomeworkStep] = useState<'idle' | 'analyzing' | 'analyzed'>('idle')
  const homeworkInputRef = useRef<HTMLInputElement>(null)

  const [chatMessages, setChatMessages] = useState<{ role: 'user' | 'assistant'; content: string }[]>([
    { role: 'assistant', content: '有什么学习问题想问我？' },
  ])
  const [chatInput, setChatInput] = useState('')
  const [isThinking, setIsThinking] = useState(false)
  const chatEndRef = useRef<HTMLDivElement>(null)

  const [agentResponseStyle, setAgentResponseStyle] = useState<'concise' | 'normal' | 'detailed'>('normal')
  const [agentGuideFirst, setAgentGuideFirst] = useState(true)
  const [agentNoDirectAnswer, setAgentNoDirectAnswer] = useState(true)
  const [agentStepByStep, setAgentStepByStep] = useState(true)
  const [agentSaved, setAgentSaved] = useState(false)

  const resetAgentSettings = () => {
    setAgentResponseStyle('normal')
    setAgentGuideFirst(true)
    setAgentNoDirectAnswer(true)
    setAgentStepByStep(true)
  }

  const saveAgentSettings = () => {
    setAgentSaved(true)
    setTimeout(() => setAgentSaved(false), 2000)
  }

  const getMockReply = (text: string) => {
    if (text.includes('概率')) return '概率是研究随机事件发生可能性的数学分支，常用条件概率、贝叶斯公式等工具来分析事件之间的关系。'
    if (text.includes('积分')) return '积分是微积分中的核心概念之一，常用于求面积、体积、平均值等，常见方法包括换元法和分部积分法。'
    if (text.includes('Python')) return '学习 Python 建议先从变量、条件、循环和函数入手，再逐步学习数据结构、面向对象和常用标准库。'
    return '这是一个很好的学习问题。我们可以先分析题目的核心概念，再一步一步解决它。'
  }

  const sendChat = () => {
    const text = chatInput.trim()
    if (!text || isThinking) return
    setChatMessages((s) => [...s, { role: 'user', content: text }])
    setChatInput('')
    setIsThinking(true)
    setTimeout(() => {
      setChatMessages((s) => [...s, { role: 'assistant', content: getMockReply(text) }])
      setIsThinking(false)
    }, 1000)
  }

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatMessages, isThinking])

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
                    onChange={(e) => {
                      const raw = e.target.value
                      if (raw === '') {
                        setDays(1)
                        return
                      }
                      const num = Number(raw)
                      setDays(Number.isNaN(num) || num < 1 ? 1 : Math.round(num))
                    }}
                  />
                </div>

                <div className="result-actions">
                  <button className="primary-button" onClick={() => {
                    setStep('planning')
                    setTimeout(() => setStep('planned'), 1000)
                  }}>
                    生成我的冲刺计划
                  </button>
                  <button className="secondary-button" onClick={() => { setStep('idle'); setSelectedFiles([]) }}>
                    重新选择资料
                  </button>
                </div>
              </div>
            )}

            {step === 'planning' && (
              <div className="analyze-loading">
                <div className="analyze-loading-text">正在为你制定冲刺计划...</div>
                <div className="analyze-loading-sub">剩余复习天数：{days} 天</div>
              </div>
            )}

            {step === 'planned' && (
              <div className="plan-result">
                <div className="plan-header">
                  <div className="plan-title">📅 我的期末冲刺计划</div>
                  <div className="plan-subtitle">距离考试还有 {days} 天</div>
                </div>

                <div className="plan-list">
                  {buildPlans(days).map((plan, index) => (
                    <div key={index} className="plan-card">
                      <div className="plan-day">Day {index + 1}</div>
                      <div className="plan-day-title">{plan.title}</div>
                      <ul className="plan-items">
                        {plan.items.map((item) => (
                          <li key={item}>{item}</li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>

                <div className="result-actions">
                  <button className="secondary-button" onClick={() => setStep('analyzed')}>
                    重新制定计划
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

  if (currentMode === 'homework') {
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

            {homeworkStep === 'idle' && (
              <>
                <div
                  className="upload-area"
                  onClick={() => homeworkInputRef.current?.click()}
                >
                  <div className="upload-hint">点击上传作业题目</div>
                  <div className="upload-types">支持图片和常见文档格式</div>
                  <input
                    ref={homeworkInputRef}
                    type="file"
                    multiple
                    className="hidden-input"
                    onChange={(e) => {
                      const files = Array.from(e.target.files || [])
                      setHomeworkFiles(files)
                    }}
                  />
                </div>

                {homeworkFiles.length > 0 && (
                  <div className="file-summary">
                    <div className="file-summary-title">已选择 {homeworkFiles.length} 个文件</div>
                    <div className="file-list">
                      {homeworkFiles.map((file, index) => (
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
                  disabled={homeworkFiles.length === 0}
                  onClick={() => {
                    setHomeworkStep('analyzing')
                    setTimeout(() => setHomeworkStep('analyzed'), 1000)
                  }}
                >
                  开始解析题目
                </button>
              </>
            )}

            {homeworkStep === 'analyzing' && (
              <div className="analyze-loading">
                <div className="analyze-loading-text">正在解析你的作业题目...</div>
                <div className="analyze-loading-sub">正在处理 {homeworkFiles.length} 个文件</div>
              </div>
            )}

            {homeworkStep === 'analyzed' && (
              <div className="homework-result">
                <div className="analyze-result-header">
                  <div className="analyze-result-title">题目解析完成</div>
                  <div className="analyze-result-sub">已解析 {homeworkFiles.length} 个文件</div>
                </div>

                <div className="result-section">
                  <div className="result-title">题目类型</div>
                  <div className="result-item">计算题</div>
                </div>

                <div className="result-section">
                  <div className="result-title">涉及知识点</div>
                  <div className="result-list">
                    <div className="result-item">条件概率</div>
                    <div className="result-item">贝叶斯公式</div>
                  </div>
                </div>

                <div className="result-section">
                  <div className="result-title">解题思路</div>
                  <div className="homework-steps">
                    <div className="homework-step">1. 先确定已知条件和要求的目标事件</div>
                    <div className="homework-step">2. 将题目转换为概率表达式</div>
                    <div className="homework-step">3. 判断是否需要使用条件概率公式</div>
                    <div className="homework-step">4. 代入数据并完成计算</div>
                  </div>
                </div>

                <div className="result-actions">
                  <button className="primary-button" onClick={() => alert('请先自己尝试完成，这里不直接展示最终答案')}>
                    我已经尝试完成
                  </button>
                  <button className="secondary-button" onClick={() => { setHomeworkStep('idle'); setHomeworkFiles([]) }}>
                    换一道题
                  </button>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>
    )
  }

  if (currentMode === 'general-question') {
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
          <div className="mode-panel chat-panel">
            <div className="mode-icon">{current.label.split(' ')[0]}</div>
            <h1>{current.title}</h1>
            <p className="subtitle">{current.description}</p>

            <div className="chat-area">
              {chatMessages.map((msg, index) => (
                <div key={index} className={`chat-message ${msg.role}`}>
                  <div className="chat-bubble">{msg.content}</div>
                </div>
              ))}
              {isThinking && (
                <div className="chat-message assistant">
                  <div className="chat-bubble">正在思考...</div>
                </div>
              )}
              <div ref={chatEndRef} />
            </div>

            <div className="chat-input-area">
              <input
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && sendChat()}
                className="chat-input"
                placeholder="输入你的学习问题..."
              />
              <button className="primary-button" onClick={sendChat} disabled={!chatInput.trim() || isThinking}>
                发送
              </button>
            </div>
          </div>
        </main>
      </div>
    )
  }

  if (currentMode === 'agent-settings') {
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
          <div className="mode-panel settings-panel">
            <div className="mode-icon">{current.label.split(' ')[0]}</div>
            <h1>{current.title}</h1>
            <p className="subtitle">{current.description}</p>

            <div className="settings-section">
              <div className="settings-title">回答风格</div>
              <div className="settings-options">
                {[
                  { value: 'concise', label: '极速简洁' },
                  { value: 'normal', label: '标准讲解' },
                  { value: 'detailed', label: '详细教学' },
                ].map((option) => (
                  <button
                    key={option.value}
                    className={`settings-option ${agentResponseStyle === option.value ? 'active' : ''}`}
                    onClick={() => setAgentResponseStyle(option.value as typeof agentResponseStyle)}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="settings-section">
              <div className="settings-title">学习方式</div>
              <div className="settings-toggles">
                <label className="settings-toggle">
                  <input
                    type="checkbox"
                    checked={agentGuideFirst}
                    onChange={(e) => setAgentGuideFirst(e.target.checked)}
                  />
                  <span>优先引导我思考</span>
                </label>
                <label className="settings-toggle">
                  <input
                    type="checkbox"
                    checked={agentNoDirectAnswer}
                    onChange={(e) => setAgentNoDirectAnswer(e.target.checked)}
                  />
                  <span>不直接给出最终答案</span>
                </label>
                <label className="settings-toggle">
                  <input
                    type="checkbox"
                    checked={agentStepByStep}
                    onChange={(e) => setAgentStepByStep(e.target.checked)}
                  />
                  <span>使用分步骤讲解</span>
                </label>
              </div>
            </div>

            <div className="settings-actions">
              <button className="primary-button" onClick={saveAgentSettings}>
                保存我的 Agent 设置
              </button>
              <button className="secondary-button" onClick={resetAgentSettings}>
                恢复默认设置
              </button>
            </div>

            {agentSaved && <div className="settings-saved">✓ Agent 设置已保存</div>}
          </div>
        </main>
      </div>
    )
  }
}

export default App

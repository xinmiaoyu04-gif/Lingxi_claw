import { useState, useRef, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { api, waitForTask, type AgentSettings, type Analysis, type Dataset, type Homework, type HomeworkResult, type Plan } from './api/client'
import AppShell from './components/AppShell'
import Home from './pages/Home'
import Courses from './pages/Courses'
import CourseSpace from './pages/CourseSpace'
import Analytics from './pages/Analytics'
import Settings from './pages/Settings'

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

function LegacyApp() {
  const [currentMode, setCurrentMode] = useState('final-sprint')
  const current = modes.find((m) => m.id === currentMode) || modes[0]
  const [datasetName, setDatasetName] = useState('高等数学期末突击')
  const [datasetCourse, setDatasetCourse] = useState('高等数学')
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [step, setStep] = useState<'idle' | 'analyzing' | 'analyzed' | 'planning' | 'planned'>('idle')
  const [days, setDays] = useState(3)
  const [dataset, setDataset] = useState<Dataset | null>(null)
  const [analysis, setAnalysis] = useState<Analysis | null>(null)
  const [plan, setPlan] = useState<Plan | null>(null)
  const [fsError, setFsError] = useState<string | null>(null)
  const [fsMessage, setFsMessage] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const runAnalysis = async () => {
    if (selectedFiles.length === 0) return
    setFsError(null)
    setStep('analyzing')
    try {
      setFsMessage('正在创建数据集...')
      const ds = await api.createDataset(datasetName.trim() || '期末突击', datasetCourse.trim() || '高等数学')
      setDataset(ds)
      setFsMessage('正在上传并解析资料...')
      const upload = await api.uploadDatasetFiles(ds.dataset_id, selectedFiles)
      await waitForTask(upload.task_id)
      setFsMessage('正在分析历年题...')
      const analysisTask = await api.startAnalysis(ds.dataset_id)
      await waitForTask(analysisTask.task_id)
      setAnalysis(await api.getAnalysis(ds.dataset_id))
      setStep('analyzed')
    } catch (e) {
      setFsError(e instanceof Error ? e.message : String(e))
      setStep('idle')
    }
  }

  const runPlan = async () => {
    if (!dataset) return
    setFsError(null)
    setStep('planning')
    try {
      const examDate = new Date(Date.now() + days * 86400000).toISOString().slice(0, 10)
      const planTask = await api.startPlan(dataset.dataset_id, {
        exam_date: examDate,
        daily_study_hours: 4,
        current_level: 'medium',
      })
      await waitForTask(planTask.task_id)
      setPlan(await api.getPlan(dataset.dataset_id))
      setStep('planned')
    } catch (e) {
      setFsError(e instanceof Error ? e.message : String(e))
      setStep('analyzed')
    }
  }

  const [homeworkCourse, setHomeworkCourse] = useState('高等数学')
  const [homeworkFiles, setHomeworkFiles] = useState<File[]>([])
  const [homeworkStep, setHomeworkStep] = useState<'idle' | 'processing' | 'ready'>('idle')
  const [homework, setHomework] = useState<Homework | null>(null)
  const [hwError, setHwError] = useState<string | null>(null)
  const [hwMessage, setHwMessage] = useState('')
  const [hints, setHints] = useState<Record<string, string>>({})
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [results, setResults] = useState<Record<string, HomeworkResult>>({})
  const [busyQuestion, setBusyQuestion] = useState<string | null>(null)
  const homeworkInputRef = useRef<HTMLInputElement>(null)

  const runHomeworkAnalysis = async () => {
    const file = homeworkFiles[0]
    if (!file) return
    setHwError(null)
    setHomeworkStep('processing')
    try {
      setHwMessage('正在上传作业...')
      const upload = await api.uploadHomework(homeworkCourse.trim() || '高等数学', file)
      setHwMessage('正在识别题目...')
      await waitForTask(upload.task_id)
      const hw = await api.getHomework(upload.homework_id)
      setHomework(hw)
      setHints({})
      setAnswers({})
      setResults({})
      setHomeworkStep('ready')
    } catch (e) {
      setHwError(e instanceof Error ? e.message : String(e))
      setHomeworkStep('idle')
    }
  }

  const askHint = async (questionId: string) => {
    if (!homework) return
    setBusyQuestion(questionId)
    try {
      const hint = await api.homeworkHint(homework.homework_id, questionId, '我不知道从哪里开始')
      setHints((s) => ({ ...s, [questionId]: hint.response }))
    } catch (e) {
      setHwError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusyQuestion(null)
    }
  }

  const submitAnswer = async (questionId: string) => {
    if (!homework) return
    const answer = (answers[questionId] ?? '').trim()
    if (!answer) return
    setBusyQuestion(questionId)
    try {
      const result = await api.homeworkAnswer(homework.homework_id, questionId, answer)
      setResults((s) => ({ ...s, [questionId]: result }))
    } catch (e) {
      setHwError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusyQuestion(null)
    }
  }

  const [chatMessages, setChatMessages] = useState<{ role: 'user' | 'assistant'; content: string }[]>([
    { role: 'assistant', content: '有什么学习问题想问我？' },
  ])
  const [chatInput, setChatInput] = useState('')
  const [isThinking, setIsThinking] = useState(false)
  const chatEndRef = useRef<HTMLDivElement>(null)

  const [settings, setSettings] = useState<AgentSettings>({ response_style: 'detailed', personality: 'encouraging', answer_policy: 'hint_first' })
  const [settingsLoaded, setSettingsLoaded] = useState(false)
  const [settingsSaving, setSettingsSaving] = useState(false)
  const [settingsSaved, setSettingsSaved] = useState(false)
  const [settingsError, setSettingsError] = useState<string | null>(null)

  useEffect(() => {
    api.getAgentSettings()
      .then(setSettings)
      .catch((e: unknown) => setSettingsError(e instanceof Error ? e.message : String(e)))
      .finally(() => setSettingsLoaded(true))
  }, [])

  const resetAgentSettings = () => {
    setSettings({ response_style: 'detailed', personality: 'encouraging', answer_policy: 'hint_first' })
  }

  const saveAgentSettings = () => {
    setSettingsSaving(true)
    setSettingsError(null)
    api.updateAgentSettings(settings)
      .then((s) => {
        setSettings(s)
        setSettingsSaved(true)
        setTimeout(() => setSettingsSaved(false), 2000)
      })
      .catch((e: unknown) => setSettingsError(e instanceof Error ? e.message : String(e)))
      .finally(() => setSettingsSaving(false))
  }

  const sendChat = () => {
    const text = chatInput.trim()
    if (!text || isThinking) return
    setChatMessages((s) => [...s, { role: 'user', content: text }])
    setChatInput('')
    setIsThinking(true)
    api.chat(text)
      .then((reply) => {
        setChatMessages((s) => [...s, { role: 'assistant', content: reply.message }])
      })
      .catch((e: unknown) => {
        setChatMessages((s) => [...s, { role: 'assistant', content: `⚠️ ${e instanceof Error ? e.message : '请求失败'}` }])
      })
      .finally(() => setIsThinking(false))
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
                <div className="result-section">
                  <div className="result-title">数据集名称</div>
                  <input
                    className="days-input"
                    value={datasetName}
                    onChange={(e) => setDatasetName(e.target.value)}
                    placeholder="例如：高等数学期末突击"
                    style={{ width: '100%', maxWidth: 360 }}
                  />
                </div>

                <div className="result-section">
                  <div className="result-title">所属课程</div>
                  <input
                    className="days-input"
                    value={datasetCourse}
                    onChange={(e) => setDatasetCourse(e.target.value)}
                    placeholder="例如：高等数学"
                    style={{ width: '100%', maxWidth: 360 }}
                  />
                </div>

                <div
                  className="upload-area"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <div className="upload-hint">点击选择学习资料</div>
                  <div className="upload-types">支持 PDF、Word、图片</div>
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

                {fsError && (
                  <div className="analysis-message" style={{ color: '#b91c1c', borderColor: '#fecaca', background: '#fef2f2' }}>⚠️ {fsError}</div>
                )}

                <button
                  className="primary-button"
                  disabled={selectedFiles.length === 0}
                  onClick={runAnalysis}
                >
                  开始分析资料
                </button>
              </>
            )}

            {step === 'analyzing' && (
              <div className="analyze-loading">
                <div className="analyze-loading-text">{fsMessage || '正在分析学习资料...'}</div>
                <div className="analyze-loading-sub">正在处理 {selectedFiles.length} 份资料</div>
              </div>
            )}

            {step === 'analyzed' && analysis && (
              <div className="analyze-result">
                <div className="analyze-result-header">
                  <div className="analyze-result-title">资料分析完成</div>
                  <div className="analyze-result-sub">{analysis.course} · 共识别 {analysis.total_questions} 道题</div>
                </div>

                <div className="result-section">
                  <div className="result-title">高频考点</div>
                  <div className="result-list">
                    {analysis.knowledge_points.map((item) => (
                      <div key={item.name} className="result-item">
                        <span>{item.name}</span>
                        <span className="result-badge">出现 {item.frequency} 次</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="result-section">
                  <div className="result-title">高频题型</div>
                  <div className="result-list">
                    {analysis.question_types.map((item) => (
                      <div key={item.name} className="result-item">
                        <span>{item.name}</span>
                        <span className="result-badge">{item.count} 题</span>
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

                {fsError && (
                  <div className="analysis-message" style={{ color: '#b91c1c', borderColor: '#fecaca', background: '#fef2f2' }}>⚠️ {fsError}</div>
                )}

                <div className="result-actions">
                  <button className="primary-button" onClick={runPlan}>
                    生成我的冲刺计划
                  </button>
                  <button className="secondary-button" onClick={() => { setStep('idle'); setSelectedFiles([]); setAnalysis(null); setDataset(null) }}>
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

            {step === 'planned' && plan && (
              <div className="plan-result">
                <div className="plan-header">
                  <div className="plan-title">📅 我的期末冲刺计划</div>
                  <div className="plan-subtitle">距离考试还有 {plan.days_remaining} 天</div>
                </div>

                <div className="plan-list">
                  {plan.daily_plan.map((day) => (
                    <div key={day.day} className="plan-card">
                      <div className="plan-day">Day {day.day} · 约 {day.estimated_hours} 小时 · {day.practice_count} 题</div>
                      <ul className="plan-items">
                        {day.focus.map((item) => (
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
                  <button className="secondary-button" onClick={() => { setStep('idle'); setSelectedFiles([]); setAnalysis(null); setDataset(null); setPlan(null) }}>
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
                <div className="result-section">
                  <div className="result-title">所属课程</div>
                  <input
                    className="days-input"
                    value={homeworkCourse}
                    onChange={(e) => setHomeworkCourse(e.target.value)}
                    placeholder="例如：高等数学"
                    style={{ width: '100%', maxWidth: 360 }}
                  />
                </div>

                <div
                  className="upload-area"
                  onClick={() => homeworkInputRef.current?.click()}
                >
                  <div className="upload-hint">点击上传作业题目</div>
                  <div className="upload-types">支持图片、PDF、Word（单张作业）</div>
                  <input
                    ref={homeworkInputRef}
                    type="file"
                    className="hidden-input"
                    onChange={(e) => {
                      const files = Array.from(e.target.files || [])
                      setHomeworkFiles(files)
                    }}
                  />
                </div>

                {homeworkFiles.length > 0 && (
                  <div className="file-summary">
                    <div className="file-summary-title">已选择 1 个文件</div>
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

                {hwError && (
                  <div className="analysis-message" style={{ color: '#b91c1c', borderColor: '#fecaca', background: '#fef2f2' }}>⚠️ {hwError}</div>
                )}

                <button
                  className="primary-button"
                  disabled={homeworkFiles.length === 0}
                  onClick={runHomeworkAnalysis}
                >
                  开始解析题目
                </button>
              </>
            )}

            {homeworkStep === 'processing' && (
              <div className="analyze-loading">
                <div className="analyze-loading-text">{hwMessage || '正在解析你的作业题目...'}</div>
                <div className="analyze-loading-sub">请稍候</div>
              </div>
            )}

            {homeworkStep === 'ready' && homework && (
              <div className="homework-result">
                <div className="analyze-result-header">
                  <div className="analyze-result-title">题目解析完成</div>
                  <div className="analyze-result-sub">识别出 {homework.questions?.length ?? 0} 道题</div>
                </div>

                {homework.questions?.map((q) => {
                  const result = results[q.question_id]
                  const answer = answers[q.question_id] ?? ''
                  return (
                    <div key={q.question_id} className="homework-step" style={{ padding: 14, marginTop: 12 }}>
                      <div style={{ fontWeight: 600 }}>{q.content}</div>
                      <div style={{ fontSize: 13, color: '#64748b', marginTop: 4 }}>{q.knowledge_point} · 难度 {q.difficulty}</div>

                      <div style={{ marginTop: 10 }}>
                        <button className="secondary-button" onClick={() => askHint(q.question_id)} disabled={busyQuestion === q.question_id}>
                          获取提示
                        </button>
                      </div>
                      {hints[q.question_id] && <div style={{ marginTop: 8, color: '#065f46' }}>💡 {hints[q.question_id]}</div>}

                      <textarea
                        className="chat-input"
                        style={{ marginTop: 10, width: '100%', minHeight: 64, resize: 'vertical', boxSizing: 'border-box' }}
                        placeholder="写下你的答案..."
                        value={answer}
                        onChange={(e) => setAnswers((s) => ({ ...s, [q.question_id]: e.target.value }))}
                      />

                      <div style={{ marginTop: 10 }}>
                        <button
                          className="primary-button"
                          style={{ marginTop: 0 }}
                          onClick={() => submitAnswer(q.question_id)}
                          disabled={busyQuestion === q.question_id || !answer.trim()}
                        >
                          提交答案
                        </button>
                      </div>

                      {result && (
                        <div style={{ marginTop: 10 }}>
                          <div className="homework-step" style={{ background: result.correct ? '#ecfdf5' : '#fef2f2', borderColor: result.correct ? '#bbf7d0' : '#fecaca' }}>
                            {result.correct ? '✓ 回答正确' : '✗ 还需改进'}（得分 {result.score}）
                          </div>
                          {result.feedback.map((fb) => (
                            <div key={fb.step} className="homework-step" style={{ marginTop: 6 }}>
                              第 {fb.step} 步 {fb.correct ? '✓' : '✗'}：{fb.message}
                            </div>
                          ))}
                          {result.final_answer && (
                            <div className="homework-step" style={{ marginTop: 6 }}>参考答案：{result.final_answer}</div>
                          )}
                        </div>
                      )}
                    </div>
                  )
                })}

                {hwError && (
                  <div className="analysis-message" style={{ color: '#b91c1c', borderColor: '#fecaca', background: '#fef2f2' }}>⚠️ {hwError}</div>
                )}

                <div className="result-actions">
                  <button className="secondary-button" onClick={() => { setHomeworkStep('idle'); setHomeworkFiles([]); setHomework(null) }}>
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

            {!settingsLoaded ? (
              <div className="settings-saved">正在加载 Agent 设置...</div>
            ) : (
              <>
                <div className="settings-section">
                  <div className="settings-title">回答风格</div>
                  <div className="settings-options">
                    {([
                      { value: 'detailed', label: '详细教学' },
                      { value: 'concise', label: '极速简洁' },
                    ] as const).map((option) => (
                      <button
                        key={option.value}
                        className={`settings-option ${settings.response_style === option.value ? 'active' : ''}`}
                        onClick={() => setSettings((s) => ({ ...s, response_style: option.value }))}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="settings-section">
                  <div className="settings-title">讲解性格</div>
                  <div className="settings-options">
                    {([
                      { value: 'encouraging', label: '鼓励型' },
                      { value: 'strict', label: '严格型' },
                      { value: 'neutral', label: '中立型' },
                    ] as const).map((option) => (
                      <button
                        key={option.value}
                        className={`settings-option ${settings.personality === option.value ? 'active' : ''}`}
                        onClick={() => setSettings((s) => ({ ...s, personality: option.value }))}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="settings-section">
                  <div className="settings-title">答题策略</div>
                  <div className="settings-options">
                    {([
                      { value: 'hint_first', label: '先给提示' },
                      { value: 'direct_answer', label: '直接给答案' },
                    ] as const).map((option) => (
                      <button
                        key={option.value}
                        className={`settings-option ${settings.answer_policy === option.value ? 'active' : ''}`}
                        onClick={() => setSettings((s) => ({ ...s, answer_policy: option.value }))}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="settings-actions">
                  <button className="primary-button" onClick={saveAgentSettings} disabled={settingsSaving}>
                    {settingsSaving ? '保存中...' : '保存我的 Agent 设置'}
                  </button>
                  <button className="secondary-button" onClick={resetAgentSettings}>
                    恢复默认设置
                  </button>
                </div>

                {settingsSaved && <div className="settings-saved">✓ Agent 设置已保存</div>}
                {settingsError && <div className="settings-saved" style={{ color: '#c62828' }}>⚠️ {settingsError}</div>}
              </>
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

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/home" replace />} />
        <Route element={<AppShell />}>
          <Route path="/home" element={<Home />} />
          <Route path="/courses" element={<Courses />} />
          <Route path="/courses/:courseId" element={<CourseSpace />} />
          <Route path="/analytics" element={<Analytics />} />
          <Route path="/settings" element={<Settings />} />
        </Route>
        <Route path="*" element={<LegacyApp />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App

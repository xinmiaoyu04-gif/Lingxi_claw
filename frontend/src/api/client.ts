// API 客户端：统一处理 API.md §4 的返回信封 { success, data, error }，
// 并把 success:false 或网络错误转成抛错，页面只关心业务数据。
//
// Base URL 走 import.meta.env.VITE_API_BASE_URL（API.md §16 要求可配置）；
// 开发期留空，配合 vite.config.ts 的 /api 代理转发到后端 8080。

export interface ChatRoute {
  mode: string
  complexity: string
  handler: string
}

export interface ChatReply {
  message: string
  route: ChatRoute
}

export type ResponseStyle = 'detailed' | 'concise'
export type Personality = 'encouraging' | 'strict' | 'neutral'
export type AnswerPolicy = 'hint_first' | 'direct_answer'

export interface AgentSettings {
  response_style: ResponseStyle
  personality: Personality
  answer_policy: AnswerPolicy
}

export interface Dataset {
  dataset_id: string
  name: string
  course: string
  file_count: number
  status: string
  created_at?: string
}

export interface FailedFile {
  name: string
  reason: string
}

export interface Task {
  task_id: string
  type: string
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'partial_success'
  progress: number
  processed_files: number
  total_files: number
  message?: string
  failed_files?: FailedFile[]
}

export interface KnowledgePoint {
  name: string
  frequency: number
  importance: string
  difficulty: string
}

export interface QuestionType {
  name: string
  count: number
}

export interface Analysis {
  dataset_id: string
  course: string
  total_questions: number
  knowledge_points: KnowledgePoint[]
  question_types: QuestionType[]
}

export interface DailyPlan {
  day: number
  focus: string[]
  practice_count: number
  estimated_hours: number
}

export interface Plan {
  dataset_id: string
  days_remaining: number
  daily_plan: DailyPlan[]
}

export interface Question {
  question_id: string
  content: string
  knowledge_point: string
  difficulty: string
}

export interface Homework {
  homework_id: string
  course: string
  status: string
  created_at: string
  questions?: Question[]
}

export interface Hint {
  question_id: string
  help_level: string
  response: string
}

export interface StepFeedback {
  step: number
  correct: boolean
  message: string
}

export interface HomeworkResult {
  question_id: string
  correct: boolean
  score: number
  feedback: StepFeedback[]
  final_answer: string
}

interface Envelope<T> {
  success: boolean
  data: T
  error: { code: string; message: string } | null
}

const BASE = import.meta.env.VITE_API_BASE_URL ?? ''

async function parseEnvelope<T>(res: Response): Promise<T> {
  let body: Envelope<T>
  try {
    body = (await res.json()) as Envelope<T>
  } catch {
    throw new Error(`后端返回了非 JSON 响应（HTTP ${res.status}）`)
  }
  if (!body.success) {
    throw new Error(body.error?.message ?? `请求失败（HTTP ${res.status}）`)
  }
  return body.data
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE}/api/v1${path}`, {
      headers: { 'Content-Type': 'application/json' },
      ...init,
    })
  } catch {
    throw new Error('无法连接后端服务，请确认它已在 8080 端口启动')
  }
  return parseEnvelope<T>(res)
}

// requestForm 用于 multipart 上传：不设置 Content-Type，让浏览器自动带上 boundary。
async function requestForm<T>(path: string, form: FormData): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE}/api/v1${path}`, { method: 'POST', body: form })
  } catch {
    throw new Error('无法连接后端服务，请确认它已在 8080 端口启动')
  }
  return parseEnvelope<T>(res)
}

export const api = {
  chat: (message: string, course?: string) =>
    request<ChatReply>('/chat', {
      method: 'POST',
      body: JSON.stringify({ message, course }),
    }),

  getAgentSettings: () => request<AgentSettings>('/settings/agent'),
  updateAgentSettings: (settings: AgentSettings) =>
    request<AgentSettings>('/settings/agent', { method: 'PUT', body: JSON.stringify(settings) }),

  getTask: (taskId: string) => request<Task>(`/tasks/${taskId}`),

  createDataset: (name: string, course: string) =>
    request<Dataset>('/final-sprint/datasets', { method: 'POST', body: JSON.stringify({ name, course }) }),

  uploadDatasetFiles: (datasetId: string, files: File[]) => {
    const form = new FormData()
    files.forEach((f) => form.append('files', f))
    return requestForm<{ dataset_id: string; task_id: string; total_files: number; status: string }>(
      `/final-sprint/datasets/${datasetId}/files`, form)
  },

  startAnalysis: (datasetId: string) =>
    request<{ task_id: string; dataset_id: string; status: string }>(
      `/final-sprint/datasets/${datasetId}/analyze`, { method: 'POST', body: '{}' }),

  getAnalysis: (datasetId: string) => request<Analysis>(`/final-sprint/datasets/${datasetId}/analysis`),

  startPlan: (datasetId: string, input: { exam_date: string; daily_study_hours: number; current_level: string }) =>
    request<{ task_id: string; status: string }>(
      `/final-sprint/datasets/${datasetId}/plan`, { method: 'POST', body: JSON.stringify(input) }),

  getPlan: (datasetId: string) => request<Plan>(`/final-sprint/datasets/${datasetId}/plan`),

  uploadHomework: (course: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    form.append('course', course)
    return requestForm<{ homework_id: string; task_id: string; status: string }>('/homework', form)
  },

  getHomework: (homeworkId: string) => request<Homework>(`/homework/${homeworkId}`),

  homeworkHint: (homeworkId: string, questionId: string, userMessage: string) =>
    request<Hint>(`/homework/${homeworkId}/hint`, {
      method: 'POST',
      body: JSON.stringify({ question_id: questionId, user_message: userMessage }),
    }),

  homeworkAnswer: (homeworkId: string, questionId: string, userAnswer: string) =>
    request<HomeworkResult>(`/homework/${homeworkId}/answer`, {
      method: 'POST',
      body: JSON.stringify({ question_id: questionId, user_answer: userAnswer }),
    }),
}

// waitForTask 轮询任务直到终态。partial_success 视为可继续（部分文件成功），
// failed 抛错并带上后端返回的 message。
export async function waitForTask(
  taskId: string,
  { intervalMs = 500, timeoutMs = 60000 } = {},
): Promise<Task> {
  const start = Date.now()
  for (;;) {
    const task = await api.getTask(taskId)
    if (task.status === 'completed' || task.status === 'partial_success') return task
    if (task.status === 'failed') throw new Error(task.message || '任务处理失败，请检查上传的文件格式')
    if (Date.now() - start > timeoutMs) throw new Error('任务处理超时，请稍后重试')
    await new Promise((r) => setTimeout(r, intervalMs))
  }
}

# Study with Lingxi-claw - Data Schema 数据结构定义

## 1. 文档目的

本文档定义 Lingxi-claw 项目中的核心数据结构。

系统中所有模块：

```text
Frontend
Backend
Workflow
Router
Agent
Mock Data
Test
```

都应该尽量使用本文档定义的数据结构。

本文档与以下文档配套使用：

```text
PRODUCT.md
    ↓
定义产品功能

WORKFLOW.md
    ↓
定义业务流程

ARCHITECTURE.md
    ↓
定义系统架构

DATA_SCHEMA.md
    ↓
定义数据结构

API.md
    ↓
定义接口通信方式
```

关系如下：

```text
用户操作
    ↓
API Request
    ↓
Data Schema
    ↓
Workflow / Router / Agent
    ↓
Data Schema
    ↓
API Response
```

---

# 2. 当前阶段说明

当前项目处于 **V02 页面架构 + 前后端协作 + Mock 联调阶段**。

当前数据结构的主要目标不是建立完整数据库，而是：

```text
统一 Frontend / Backend / Mock / Test 的数据格式
```

因此当前阶段：

```text
✅ 定义核心 Entity
✅ 定义 TypeScript 类型
✅ 定义 Mock Data 结构
✅ 定义 API Request / Response 使用的数据
```

暂不要求：

```text
❌ 完整数据库设计
❌ 真实用户认证系统
❌ 真实 AI 数据持久化
❌ 真实向量数据库
❌ 复杂知识图谱
❌ 完整学习行为数据仓库
```

未来 Backend 可以根据实际数据库选型，将本文档中的逻辑数据结构映射到具体数据库。

---

# 3. 核心数据模型

Lingxi-claw 当前核心数据关系：

```text
User
 │
 ├── LearningProfile
 │
 ├── AIPreference
 │
 ├── GeneralSettings
 │
 └── Courses
       │
       └── Course
             │
             ├── CourseOverview
             │
             ├── CourseMaterial
             │
             ├── Homework
             │
             ├── Mistake
             │
             ├── LearningRecord
             │
             └── CourseAnalytics
```

整体关系：

```text
                       User
                         │
          ┌──────────────┼──────────────┐
          ↓              ↓              ↓
 LearningProfile   AIPreference   GeneralSettings
                         │
                         ↓
                      Courses
                         │
                         ↓
                    Course Space
                         │
       ┌─────────────────┼─────────────────┐
       ↓                 ↓                 ↓
   Materials         Homework           Mistakes
       │                 │                 │
       └─────────────────┼─────────────────┘
                         ↓
                  Course Analytics
```

---

# 4. ID 规范

所有核心 Entity 必须具有唯一 `id`。

统一使用：

```ts
id: string
```

例如：

```text
user-001
course-math
material-001
homework-001
mistake-001
task-001
```

当前 Mock 环境可以使用字符串 ID。

未来 Backend 可以映射为：

```text
UUID
数据库自增 ID
Snowflake ID
```

但 Frontend 不应该依赖 ID 的具体生成方式。

---

# 5. User

User 表示系统用户。

```ts
interface User {
  id: string;
  name: string;
  avatar?: string;
  createdAt: string;
  updatedAt: string;
}
```

示例：

```json
{
  "id": "user-001",
  "name": "Student",
  "avatar": "",
  "createdAt": "2026-08-30T09:00:00Z",
  "updatedAt": "2026-08-30T09:00:00Z"
}
```

---

# 6. LearningProfile

LearningProfile 对应产品中的：

> 首次进入 / 学习画像

包括：

```text
专业
学习习惯
作业习惯
AI 使用习惯
学习目标
```

数据结构：

```ts
interface LearningProfile {
  userId: string;
  major: string;
  learningHabits: string[];
  homeworkHabits: string[];
  aiUsageHabits: string[];
  learningGoals: string[];
  completed: boolean;
  updatedAt: string;
}
```

示例：

```json
{
  "userId": "user-001",
  "major": "软件工程",
  "learningHabits": [
    "晚上学习",
    "集中学习"
  ],
  "homeworkHabits": [
    "先自己尝试"
  ],
  "aiUsageHabits": [
    "概念解释",
    "作业辅导"
  ],
  "learningGoals": [
    "提高课程成绩"
  ],
  "completed": true,
  "updatedAt": "2026-08-30T09:00:00Z"
}
```

---

# 7. AIPreference

对应：

> Settings → AI 偏好

```ts
interface AIPreference {
  userId: string;
  responseStyle: ResponseStyle;
  teachingMode: TeachingMode;
  responseLength: ResponseLength;
  customPrompt?: string;
  updatedAt: string;
}
```

## 7.1 ResponseStyle

```ts
type ResponseStyle =
  | "concise"
  | "standard"
  | "detailed";
```

对应：

```text
极速简洁
标准讲解
详细推导
```

---

## 7.2 TeachingMode

```ts
type TeachingMode =
  | "direct_answer"
  | "hint_first"
  | "guided_question"
  | "teacher_style";
```

对应：

```text
直接给出答案
先提示，再给答案
通过提问引导思考
像老师一样详细讲解
```

---

## 7.3 ResponseLength

```ts
type ResponseLength =
  | "short"
  | "medium"
  | "long";
```

---

# 8. GeneralSettings

对应：

> Settings → 通用设置

```ts
interface GeneralSettings {
  userId: string;
  language: string;
  theme: "light" | "dark" | "system";
  notifications: boolean;
  updatedAt: string;
}
```

---

# 9. Course

Course 是当前产品最重要的核心实体。

用户进入：

```text
我的课程
    ↓
高等数学
    ↓
Course Space
```

因此其他课程相关数据都应该通过：

```text
courseId
```

与 Course 建立关系。

```ts
interface Course {
  id: string;
  userId: string;
  name: string;
  code?: string;
  description?: string;
  progress: number;
  mastery: number;
  createdAt: string;
  updatedAt: string;
}
```

示例：

```json
{
  "id": "math",
  "userId": "user-001",
  "name": "高等数学",
  "code": "MATH101",
  "description": "大学高等数学课程",
  "progress": 68,
  "mastery": 72,
  "createdAt": "2026-08-01T09:00:00Z",
  "updatedAt": "2026-08-30T09:00:00Z"
}
```

---

# 10. Course Space

Course Space 不是独立数据库实体，而是：

> **以 Course 为核心组织的一组学习数据和功能上下文。**

结构：

```text
Course Space
 │
 ├── Overview
 ├── Materials
 ├── Homework
 ├── Mistakes
 └── Analytics
```

统一通过：

```ts
courseId: string
```

关联。

因此：

```text
Course
    │
    ├── courseId
    │
    ├── Materials
    ├── Homework
    ├── Mistakes
    └── Analytics
```

---

# 11. CourseOverview

Course Overview 是 Course Space 的首页数据。

```ts
interface CourseOverview {
  courseId: string;
  progress: number;
  mastery: number;
  todayTasks: number;
  completedTasks: number;
  pendingHomework: number;
  mistakeCount: number;
  weakTopics: WeakTopic[];
  recentActivities: LearningActivity[];
}
```

---

# 12. CourseMaterial

CourseMaterial 表示课程资料知识库中的一个资料。

支持：

```text
PDF
DOCX
JPG
PNG
```

数据结构：

```ts
interface CourseMaterial {
  id: string;
  courseId: string;
  name: string;
  type: MaterialType;
  size?: number;
  status: MaterialStatus;
  url?: string;
  createdAt: string;
  updatedAt: string;
}
```

---

## 12.1 MaterialType

```ts
type MaterialType =
  | "pdf"
  | "docx"
  | "jpg"
  | "png";
```

---

## 12.2 MaterialStatus

```ts
type MaterialStatus =
  | "uploaded"
  | "processing"
  | "ready"
  | "failed";
```

文件处理流程：

```text
uploaded
    ↓
processing
    ↓
ready
```

失败：

```text
processing
    ↓
failed
```

---

# 13. Homework

Homework 表示课程中的一个作业任务。

```ts
interface Homework {
  id: string;
  courseId: string;
  title: string;
  status: HomeworkStatus;
  questionCount: number;
  completedCount: number;
  createdAt: string;
  updatedAt: string;
}
```

---

## 13.1 HomeworkStatus

```ts
type HomeworkStatus =
  | "pending"
  | "in_progress"
  | "completed";
```

---

# 14. HomeworkQuestion

如果一个 Homework 包含多道题，则使用 HomeworkQuestion。

```ts
interface HomeworkQuestion {
  id: string;
  homeworkId: string;
  courseId: string;
  content: string;
  type?: QuestionType;
  topic?: string;
  userAnswer?: string;
  status: QuestionStatus;
}
```

---

## 14.1 QuestionType

当前允许：

```ts
type QuestionType =
  | "calculation"
  | "proof"
  | "concept"
  | "programming"
  | "multiple_choice"
  | "other";
```

---

## 14.2 QuestionStatus

```ts
type QuestionStatus =
  | "pending"
  | "attempting"
  | "submitted"
  | "completed";
```

---

# 15. Mistake

Mistake 表示用户的一条错题记录。

```ts
interface Mistake {
  id: string;
  courseId: string;
  questionId?: string;
  title: string;
  content?: string;
  topic: string;
  difficulty: Difficulty;
  mistakeCount: number;
  status: MistakeStatus;
  createdAt: string;
  updatedAt: string;
}
```

---

## 15.1 Difficulty

```ts
type Difficulty =
  | "easy"
  | "medium"
  | "hard";
```

---

## 15.2 MistakeStatus

```ts
type MistakeStatus =
  | "unreviewed"
  | "reviewed";
```

---

# 16. LearningRecord

LearningRecord 用于记录用户的学习活动。

```ts
interface LearningRecord {
  id: string;
  userId: string;
  courseId?: string;
  type: LearningRecordType;
  duration?: number;
  metadata?: Record<string, unknown>;
  createdAt: string;
}
```

---

## 16.1 LearningRecordType

```ts
type LearningRecordType =
  | "study"
  | "homework"
  | "review"
  | "mistake_review"
  | "material_read"
  | "ai_assist";
```

---

# 17. LearningActivity

用于 Course Overview 中的最近活动。

```ts
interface LearningActivity {
  id: string;
  courseId: string;
  type: LearningRecordType;
  title: string;
  description?: string;
  createdAt: string;
}
```

示例：

```json
{
  "id": "activity-001",
  "courseId": "math",
  "type": "mistake_review",
  "title": "复习了二重积分错题",
  "description": "完成 3 道错题复习",
  "createdAt": "2026-08-30T10:00:00Z"
}
```

---

# 18. Analytics

Analytics 是产品中的学习分析数据。

分为：

```text
Global Analytics
Course Analytics
```

---

# 19. GlobalAnalytics

对应：

> 学习分析 → 整体学习情况

```ts
interface GlobalAnalytics {
  userId: string;
  overallMastery: number;
  studyHours: number;
  completedTasks: number;
  weakTopics: WeakTopic[];
  courseMastery: CourseMastery[];
  trend: LearningTrend[];
}
```

---

# 20. CourseAnalytics

对应 Course Space：

> 学习分析

```ts
interface CourseAnalytics {
  courseId: string;
  mastery: number;
  studyHours: number;
  completedHomework: number;
  mistakeCount: number;
  weakTopics: WeakTopic[];
  homeworkPerformance: HomeworkPerformance;
  trend: LearningTrend[];
}
```

---

# 21. WeakTopic

用于表示用户薄弱知识点。

```ts
interface WeakTopic {
  topic: string;
  courseId?: string;
  courseName?: string;
  mastery: number;
}
```

示例：

```json
{
  "topic": "二重积分",
  "courseId": "math",
  "courseName": "高等数学",
  "mastery": 42
}
```

---

# 22. CourseMastery

用于整体学习分析中的课程掌握度。

```ts
interface CourseMastery {
  courseId: string;
  courseName: string;
  mastery: number;
}
```

---

# 23. HomeworkPerformance

```ts
interface HomeworkPerformance {
  accuracy: number;
  completionRate: number;
}
```

---

# 24. LearningTrend

用于学习趋势图表。

```ts
interface LearningTrend {
  date: string;
  studyHours?: number;
  mastery?: number;
  completedTasks?: number;
}
```

示例：

```json
[
  {
    "date": "2026-08-26",
    "studyHours": 2,
    "mastery": 61
  },
  {
    "date": "2026-08-27",
    "studyHours": 3,
    "mastery": 64
  },
  {
    "date": "2026-08-28",
    "studyHours": 4,
    "mastery": 68
  }
]
```

---

# 25. LearningSuggestion

用于 Home 和 Analytics 中的：

> AI 学习建议

```ts
interface LearningSuggestion {
  id: string;
  type: SuggestionType;
  courseId?: string;
  title: string;
  description: string;
  priority: SuggestionPriority;
}
```

---

## 25.1 SuggestionType

```ts
type SuggestionType =
  | "review"
  | "homework"
  | "study"
  | "mistake"
  | "course";
```

---

## 25.2 SuggestionPriority

```ts
type SuggestionPriority =
  | "low"
  | "medium"
  | "high";
```

---

# 26. AI Task

AI Task 用于未来统一承载 AI 请求。

当前阶段可以使用 Mock。

```ts
interface AITask {
  taskId: string;
  userId: string;
  courseId?: string;
  type: AITaskType;
  status: AITaskStatus;
  input: AITaskInput;
  result?: AITaskResult;
  routing?: RoutingInfo;
  createdAt: string;
  updatedAt: string;
}
```

---

# 27. AITaskType

```ts
type AITaskType =
  | "homework"
  | "question"
  | "explanation"
  | "analysis"
  | "study_plan"
  | "general";
```

---

# 28. AITaskStatus

```ts
type AITaskStatus =
  | "queued"
  | "processing"
  | "completed"
  | "failed";
```

---

# 29. AITaskInput

```ts
interface AITaskInput {
  text?: string;
  fileIds?: string[];
  questionId?: string;
}
```

---

# 30. AITaskResult

```ts
interface AITaskResult {
  type: "answer" | "hint" | "analysis" | "plan";
  content: string;
  metadata?: Record<string, unknown>;
}
```

当前阶段可以使用：

```text
Mock Result
```

未来再接入：

```text
Workflow
Router
Agent
Skill
Model
```

---

# 31. RoutingInfo

RoutingInfo 用于描述 AI Task 的内部路由结果。

Frontend 默认不依赖这些字段。

```ts
interface RoutingInfo {
  scene?: string;
  workflow?: string;
  skill?: string;
  model?: string;
  inputType?: string;
}
```

示例：

```json
{
  "scene": "course_homework",
  "workflow": "homework",
  "skill": "math",
  "model": "mock-model",
  "inputType": "text"
}
```

RoutingInfo 的主要用途：

```text
调试
Explainable AI
Demo
成本分析
延迟分析
```

---

# 32. File 与 Material 的关系

文件不是独立的业务学习对象。

当前产品中：

```text
用户上传文件
       ↓
Course Material
       ↓
课程知识库
```

因此：

```text
File
 ↓
Material
 ↓
Course
```

文件本身可以包含：

```text
原始文件
解析状态
文件类型
文件大小
```

而 CourseMaterial 负责表达：

> 这个文件属于哪门课程，以及当前资料状态。

---

# 33. Home Data

Home 页面需要的不是单独的数据库实体，而是多个数据的聚合结果。

```ts
interface HomeData {
  todayStudy: TodayStudy;
  recentCourses: Course[];
  recentHomework: Homework[];
  learningProgress: LearningProgress;
  aiSuggestion?: LearningSuggestion;
}
```

---

# 34. TodayStudy

```ts
interface TodayStudy {
  completed: number;
  total: number;
}
```

---

# 35. LearningProgress

```ts
interface LearningProgress {
  overall: number;
}
```

---

# 36. Settings Data

Settings 页面数据由三个部分组成：

```text
Settings
 ├── LearningProfile
 ├── AIPreference
 └── GeneralSettings
```

不创建一个重复的 Settings 数据模型。

---

# 37. Entity Relationship

核心关系：

```text
User
 │
 ├─────────────── LearningProfile
 │
 ├─────────────── AIPreference
 │
 ├─────────────── GeneralSettings
 │
 └─────────────── Course
                    │
                    ├──────── CourseMaterial
                    │
                    ├──────── Homework
                    │              │
                    │              └── HomeworkQuestion
                    │
                    ├──────── Mistake
                    │
                    ├──────── LearningRecord
                    │
                    └──────── CourseAnalytics
```

其中：

```text
User 1 ── N Course

Course 1 ── N CourseMaterial

Course 1 ── N Homework

Homework 1 ── N HomeworkQuestion

Course 1 ── N Mistake

Course 1 ── N LearningRecord

Course 1 ── 1 CourseAnalytics

User 1 ── 1 LearningProfile

User 1 ── 1 AIPreference

User 1 ── 1 GeneralSettings
```

---

# 38. Product Information Architecture 与 Data Schema

产品结构：

```text
LINGXI
│
├── 首次进入 / 学习画像
│       └── LearningProfile
│
├── Home
│       ├── TodayStudy
│       ├── RecentCourses
│       ├── RecentHomework
│       ├── LearningProgress
│       └── LearningSuggestion
│
├── Courses
│       └── Course
│
├── Course Space
│       ├── Overview
│       │     └── CourseOverview
│       │
│       ├── 课程资料知识库
│       │     └── CourseMaterial
│       │
│       ├── AI 作业辅导
│       │     └── Homework
│       │
│       ├── 错题记录
│       │     └── Mistake
│       │
│       └── 学习分析
│             └── CourseAnalytics
│
├── Analytics
│       └── GlobalAnalytics
│
└── Settings
        ├── LearningProfile
        ├── AIPreference
        └── GeneralSettings
```

---

# 39. API 与 Data Schema 对应关系

```text
GET /api/profile
        ↓
LearningProfile

GET /api/home
        ↓
HomeData

GET /api/courses
        ↓
Course[]

GET /api/courses/:courseId
        ↓
Course

GET /api/courses/:courseId/overview
        ↓
CourseOverview

GET /api/courses/:courseId/materials
        ↓
CourseMaterial[]

GET /api/courses/:courseId/homework
        ↓
Homework[]

GET /api/courses/:courseId/mistakes
        ↓
Mistake[]

GET /api/courses/:courseId/analytics
        ↓
CourseAnalytics

GET /api/analytics
        ↓
GlobalAnalytics

GET /api/settings/ai
        ↓
AIPreference

GET /api/settings/general
        ↓
GeneralSettings

POST /api/ai/tasks
        ↓
AITask
```

---

# 40. Frontend TypeScript 使用原则

Frontend 应尽量直接使用本文档对应的 TypeScript 类型。

例如：

```ts
interface Course {
  id: string;
  userId: string;
  name: string;
  code?: string;
  description?: string;
  progress: number;
  mastery: number;
  createdAt: string;
  updatedAt: string;
}
```

页面：

```tsx
const [courses, setCourses] = useState<Course[]>([]);
```

而不是自行定义：

```ts
interface MyCourse {
  courseName: string;
  percent: number;
}
```

避免：

```text
Data Schema
      ↓
API
      ↓
Frontend 自己重新定义一套数据
```

---

# 41. Mock Data 使用原则

Mock Data 必须严格遵循 Data Schema。

正确：

```text
DATA_SCHEMA
      ↓
TypeScript Type
      ↓
Mock Data
      ↓
Mock API
      ↓
Frontend
```

错误：

```text
Frontend 临时需要一个字段
      ↓
Mock 随便加字段
      ↓
API 与 Schema 不一致
```

如果确实需要新的数据字段：

```text
需求
 ↓
修改 DATA_SCHEMA.md
 ↓
团队确认
 ↓
修改 API.md
 ↓
修改 TypeScript Type
 ↓
修改 Mock
 ↓
修改 Backend
 ↓
修改 Test
```

---

# 42. 数据字段命名规范

统一使用：

```text
camelCase
```

例如：

```text
courseId
createdAt
updatedAt
questionCount
completedCount
mistakeCount
learningGoals
```

不要混用：

```text
course_id
courseID
CourseId
```

---

# 43. 时间字段规范

所有时间字段统一使用：

```text
ISO 8601
```

例如：

```text
2026-08-30T10:00:00Z
```

Frontend 不应该假设 Backend 返回的是某种本地时间格式。

---

# 44. 数值范围规范

百分比字段统一使用：

```text
0 - 100
```

例如：

```json
{
  "progress": 68,
  "mastery": 72
}
```

不要使用：

```text
0.68
0.72
```

除非 API.md 明确规定。

---

# 45. 当前阶段的数据简化原则

由于 V02 当前主要完成：

```text
页面框架
路由
UI
Mock
```

因此数据模型允许简化。

例如：

```text
Course
 ├── progress
 └── mastery
```

可以直接使用 Mock 数值。

未来真实系统中：

```text
学习记录
    ↓
作业表现
    ↓
错题
    ↓
知识点
    ↓
学习分析
    ↓
mastery
```

再逐步计算真实数据。

---

# 46. Workflow / Router / Agent 数据关系

未来 AI 系统：

```text
User
 ↓
AITask
 ↓
Router
 ↓
Workflow / Agent
 ↓
Skill
 ↓
Model
 ↓
AITaskResult
```

其中：

```text
AITask
```

是 AI 请求的数据载体。

```text
RoutingInfo
```

是内部路由结果。

```text
AITaskResult
```

是最终 AI 输出。

因此 Frontend 不应该直接依赖：

```text
Workflow 内部对象
Router 内部对象
Agent 内部对象
Model 内部对象
```

Frontend 只依赖：

```text
AITask
AITaskResult
RoutingInfo（可选）
```

---

# 47. 当前禁止事项

当前阶段禁止因为页面开发而随意修改核心数据结构：

```text
❌ Frontend 自己创造重复 Entity
❌ Mock 自己创造字段
❌ Backend 随意修改字段名称
❌ API 返回与 Data Schema 不一致
❌ 删除已有 services
❌ 删除已有 API Types
❌ 删除 Mock API
❌ 删除 LegacyApp
```

如果确实需要修改：

```text
发现需求
    ↓
修改 DATA_SCHEMA.md
    ↓
团队确认
    ↓
同步 API.md
    ↓
同步 TypeScript Types
    ↓
同步 Mock
    ↓
同步 Backend
    ↓
同步 Test
```

---

# 48. 当前优先级

考虑当前开发时间有限，数据结构优先保证：

## P0

```text
User
LearningProfile
Course
CourseOverview
CourseMaterial
Homework
Mistake
GlobalAnalytics
CourseAnalytics
AIPreference
GeneralSettings
HomeData
```

## P1

```text
HomeworkQuestion
LearningRecord
LearningSuggestion
LearningActivity
```

## P2

```text
AITask
AITaskResult
RoutingInfo
```

P2 数据结构主要为后续真实 AI 联调准备。

---

# 49. 最终数据架构

Lingxi-claw 当前的数据核心可以概括为：

```text
                         User
                           │
            ┌──────────────┼──────────────┐
            ↓              ↓              ↓
       学习画像          AI 偏好        通用设置
            │
            ↓
         Courses
            │
            ↓
       ┌────────────┐
       │ Course     │
       └─────┬──────┘
             │
      ┌──────┼──────┬──────────┐
      ↓      ↓      ↓          ↓
  Materials Homework Mistakes Analytics
             │
             ↓
       HomeworkQuestion

             ↓
       LearningRecord
             ↓
      Global Analytics
             ↓
      Learning Suggestion
```

AI 能力作为独立的数据流：

```text
User
 ↓
AITask
 ↓
Router
 ↓
Workflow / Agent
 ↓
Skill / Model
 ↓
AITaskResult
```

---

# 50. 一句话总结

> **DATA_SCHEMA.md 定义 Lingxi-claw “数据长什么样”，以 User → Course → Course Space 为核心数据关系，并围绕课程资料、作业、错题、学习记录和学习分析建立统一的数据模型；AI 部分通过 AITask 与 AITaskResult 进行隔离，为后续 Workflow、Router、Agent 和 Model 联调提供稳定的数据基础。**

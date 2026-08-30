# Study with Lingxi-claw - API 接口定义

# 1. 文档目的

本文档定义 Lingxi-claw 前端与后端之间的 API 接口。

所有团队成员必须遵守：

```text
Frontend
    ↓ HTTP Request
Backend API
    ↓
Business Service
    ↓
Workflow / Router / Agent
    ↓
Service / Tool / Model
    ↓
HTTP Response
    ↓
Frontend
```

本文档是：

> **前端、后端、测试之间的接口合同。**

前端开发者不应该猜测后端接口。

后端开发者不应该随意修改 Response 字段。

测试人员根据本文档编写测试。

如果接口需要修改：

```text
修改 API.md
    ↓
团队确认
    ↓
修改代码
    ↓
Frontend / Backend / QA 同步
```

而不是直接修改代码。

---

# 2. 当前阶段说明

当前项目处于 **V02 页面架构与前后端协作阶段**。

当前目标：

```text
完成页面
    ↓
完成路由
    ↓
定义 API Contract
    ↓
使用 Mock Data
    ↓
前后端并行开发
    ↓
后续进行真实 API 联调
```

当前阶段：

```text
✅ 页面框架
✅ 路由
✅ API Contract
✅ Mock API
✅ Mock Data
```

暂不要求：

```text
❌ 真实 AI 模型
❌ 真实 Agent
❌ 真实 Workflow
❌ 真实 OCR
❌ 真实知识库
❌ 真实向量检索
❌ 真实学习分析算法
```

因此所有 API 都必须设计成：

> **现在可以 Mock，未来可以无缝替换为真实 Backend。**

---

# 3. API 基础规范

## 3.1 Base URL

开发环境：

```text
/api
```

例如：

```text
GET /api/courses
```

生产环境 Base URL 根据部署环境配置。

Frontend 不应该在业务代码中硬编码完整域名。

---

# 4. HTTP 方法规范

| 方法     | 用途          |
| ------ | ----------- |
| GET    | 获取数据        |
| POST   | 创建数据 / 执行任务 |
| PUT    | 完整更新        |
| PATCH  | 部分更新        |
| DELETE | 删除数据        |

---

# 5. Response 基础结构

所有 API 推荐使用统一 Response：

```json
{
  "success": true,
  "data": {},
  "message": null
}
```

失败：

```json
{
  "success": false,
  "data": null,
  "message": "课程不存在"
}
```

---

# 6. HTTP Status Code

| Status | 含义             |
| ------ | -------------- |
| 200    | 请求成功           |
| 201    | 创建成功           |
| 400    | 请求参数错误         |
| 401    | 未认证            |
| 403    | 无权限            |
| 404    | 资源不存在          |
| 409    | 数据冲突           |
| 422    | 参数校验失败         |
| 500    | 服务端错误          |
| 503    | AI / 外部服务暂时不可用 |

Frontend 不应该仅依赖 HTTP Status 判断业务状态。

优先读取：

```json
{
  "success": false,
  "message": "..."
}
```

---

# 7. 核心资源模型

当前 API 主要围绕以下资源：

```text
User
 │
 ├── LearningProfile
 │
 ├── AIPreference
 │
 └── Course
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
 ↓
Courses
 ↓
Course Space
 ├── Overview
 ├── Materials
 ├── Homework
 ├── Mistakes
 └── Analytics
```

---

# 8. 用户与学习画像 API

## 8.1 获取学习画像

```http
GET /api/profile
```

Response：

```json
{
  "success": true,
  "data": {
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
    ]
  },
  "message": null
}
```

---

## 8.2 更新学习画像

```http
PATCH /api/profile
```

Request：

```json
{
  "major": "软件工程",
  "learningHabits": [
    "晚上学习"
  ],
  "homeworkHabits": [
    "先自己尝试"
  ],
  "aiUsageHabits": [
    "概念解释"
  ],
  "learningGoals": [
    "提高课程成绩"
  ]
}
```

Response：

```json
{
  "success": true,
  "data": {
    "major": "软件工程",
    "learningHabits": [
      "晚上学习"
    ],
    "homeworkHabits": [
      "先自己尝试"
    ],
    "aiUsageHabits": [
      "概念解释"
    ],
    "learningGoals": [
      "提高课程成绩"
    ]
  },
  "message": null
}
```

---

# 9. Home API

Home 用于展示用户整体学习状态。

---

## 9.1 获取首页数据

```http
GET /api/home
```

Response：

```json
{
  "success": true,
  "data": {
    "todayStudy": {
      "completed": 2,
      "total": 4
    },
    "recentCourses": [
      {
        "id": "math",
        "name": "高等数学",
        "progress": 68
      },
      {
        "id": "physics",
        "name": "大学物理",
        "progress": 52
      }
    ],
    "recentHomework": [],
    "learningProgress": {
      "overall": 61
    },
    "aiSuggestion": {
      "title": "建议继续复习高等数学",
      "description": "当前高等数学存在部分薄弱知识点。"
    }
  },
  "message": null
}
```

当前阶段允许全部使用 Mock Data。

---

# 10. Courses API

Courses 是用户的课程管理入口。

---

## 10.1 获取课程列表

```http
GET /api/courses
```

Response：

```json
{
  "success": true,
  "data": [
    {
      "id": "math",
      "name": "高等数学",
      "code": "MATH101",
      "progress": 68,
      "mastery": 72,
      "color": null
    },
    {
      "id": "physics",
      "name": "大学物理",
      "code": "PHY101",
      "progress": 52,
      "mastery": 61,
      "color": null
    },
    {
      "id": "linear-algebra",
      "name": "线性代数",
      "code": "MATH102",
      "progress": 45,
      "mastery": 55,
      "color": null
    }
  ],
  "message": null
}
```

---

## 10.2 获取单门课程

```http
GET /api/courses/:courseId
```

例如：

```http
GET /api/courses/math
```

Response：

```json
{
  "success": true,
  "data": {
    "id": "math",
    "name": "高等数学",
    "code": "MATH101",
    "description": "大学高等数学课程",
    "progress": 68,
    "mastery": 72,
    "totalMaterials": 12,
    "pendingHomework": 2,
    "mistakeCount": 15
  },
  "message": null
}
```

---

## 10.3 创建课程

```http
POST /api/courses
```

Request：

```json
{
  "name": "高等数学",
  "code": "MATH101",
  "description": "大学高等数学课程"
}
```

Response：

```json
{
  "success": true,
  "data": {
    "id": "math",
    "name": "高等数学",
    "code": "MATH101",
    "description": "大学高等数学课程"
  },
  "message": null
}
```

---

## 10.4 更新课程

```http
PATCH /api/courses/:courseId
```

Request：

```json
{
  "name": "高等数学",
  "description": "大学高等数学"
}
```

---

## 10.5 删除课程

```http
DELETE /api/courses/:courseId
```

Response：

```json
{
  "success": true,
  "data": null,
  "message": "课程删除成功"
}
```

---

# 11. Course Space API

Course Space 是产品的核心二级架构。

所有 Course Space API 都必须携带：

```text
courseId
```

例如：

```text
/courses/math/...
```

Course Space 包含：

```text
Overview
Materials
Homework
Mistakes
Analytics
```

---

# 12. Course Overview API

## 12.1 获取课程概览

```http
GET /api/courses/:courseId/overview
```

例如：

```http
GET /api/courses/math/overview
```

Response：

```json
{
  "success": true,
  "data": {
    "course": {
      "id": "math",
      "name": "高等数学"
    },
    "progress": 68,
    "mastery": 72,
    "todayTasks": 3,
    "completedTasks": 2,
    "pendingHomework": 2,
    "mistakeCount": 15,
    "weakTopics": [
      "二重积分",
      "微分方程"
    ],
    "recentActivities": []
  },
  "message": null
}
```

---

# 13. Course Materials API

课程资料对应 Course Space 中的：

> **课程资料知识库**

---

## 13.1 获取课程资料

```http
GET /api/courses/:courseId/materials
```

Response：

```json
{
  "success": true,
  "data": [
    {
      "id": "material-001",
      "courseId": "math",
      "name": "高等数学期末复习资料.pdf",
      "type": "pdf",
      "size": 1024000,
      "status": "ready",
      "createdAt": "2026-08-30T10:00:00Z"
    }
  ],
  "message": null
}
```

---

## 13.2 上传课程资料

```http
POST /api/courses/:courseId/materials
```

使用：

```text
multipart/form-data
```

字段：

```text
file
```

支持：

```text
PDF
DOCX
JPG
PNG
```

Response：

```json
{
  "success": true,
  "data": {
    "id": "material-002",
    "name": "高等数学笔记.pdf",
    "type": "pdf",
    "status": "processing"
  },
  "message": "文件上传成功"
}
```

---

## 13.3 获取资料处理状态

```http
GET /api/courses/:courseId/materials/:materialId
```

Response：

```json
{
  "success": true,
  "data": {
    "id": "material-002",
    "status": "processing",
    "progress": 65
  },
  "message": null
}
```

`status`：

```text
uploaded
processing
ready
failed
```

---

## 13.4 删除资料

```http
DELETE /api/courses/:courseId/materials/:materialId
```

---

# 14. Homework API

Course Space 中的：

> **AI 作业辅导**

---

## 14.1 获取课程作业

```http
GET /api/courses/:courseId/homework
```

Response：

```json
{
  "success": true,
  "data": [
    {
      "id": "homework-001",
      "courseId": "math",
      "title": "二重积分作业",
      "status": "in_progress",
      "questionCount": 10,
      "completedCount": 6,
      "createdAt": "2026-08-30T09:00:00Z"
    }
  ],
  "message": null
}
```

作业状态：

```text
pending
in_progress
completed
```

---

## 14.2 获取单个作业

```http
GET /api/courses/:courseId/homework/:homeworkId
```

---

## 14.3 创建作业任务

```http
POST /api/courses/:courseId/homework
```

Request：

```json
{
  "title": "二重积分作业"
}
```

---

# 15. AI Homework API

真实 AI 能力暂不要求实现，但 API Contract 可以提前定义。

---

## 15.1 请求 AI 作业辅导

```http
POST /api/courses/:courseId/homework/:homeworkId/assist
```

Request：

```json
{
  "questionId": "question-001",
  "userAnswer": "..."
}
```

Response：

```json
{
  "success": true,
  "data": {
    "level": 1,
    "type": "hint",
    "content": "可以先考虑使用换元法。",
    "nextAction": "continue"
  },
  "message": null
}
```

提示等级：

```text
1 = 方向提示
2 = 方法提示
3 = 关键步骤
4 = 完整解析
```

当前 V02：

```text
使用 Mock Response
```

---

# 16. Mistakes API

Course Space 中的：

> **错题记录**

---

## 16.1 获取错题

```http
GET /api/courses/:courseId/mistakes
```

Response：

```json
{
  "success": true,
  "data": [
    {
      "id": "mistake-001",
      "courseId": "math",
      "questionId": "question-001",
      "title": "二重积分计算",
      "topic": "二重积分",
      "difficulty": "medium",
      "mistakeCount": 2,
      "status": "unreviewed"
    }
  ],
  "message": null
}
```

---

## 16.2 获取单个错题

```http
GET /api/courses/:courseId/mistakes/:mistakeId
```

---

## 16.3 标记错题已复习

```http
PATCH /api/courses/:courseId/mistakes/:mistakeId
```

Request：

```json
{
  "status": "reviewed"
}
```

---

# 17. Analytics API

Analytics 分成：

```text
Global Analytics
Course Analytics
```

---

# 18. Global Analytics

## 18.1 获取整体学习分析

```http
GET /api/analytics
```

Response：

```json
{
  "success": true,
  "data": {
    "overallMastery": 68,
    "studyHours": 24,
    "completedTasks": 18,
    "weakTopics": [
      {
        "topic": "二重积分",
        "course": "高等数学",
        "mastery": 42
      }
    ],
    "courseMastery": [
      {
        "courseId": "math",
        "courseName": "高等数学",
        "mastery": 72
      },
      {
        "courseId": "physics",
        "courseName": "大学物理",
        "mastery": 61
      }
    ],
    "trend": []
  },
  "message": null
}
```

---

# 19. Course Analytics

## 19.1 获取单课程学习分析

```http
GET /api/courses/:courseId/analytics
```

Response：

```json
{
  "success": true,
  "data": {
    "mastery": 72,
    "studyHours": 12,
    "completedHomework": 8,
    "mistakeCount": 15,
    "weakTopics": [
      {
        "topic": "二重积分",
        "mastery": 42
      },
      {
        "topic": "微分方程",
        "mastery": 51
      }
    ],
    "homeworkPerformance": {
      "accuracy": 76,
      "completionRate": 84
    },
    "trend": []
  },
  "message": null
}
```

---

# 20. AI Learning Suggestion API

Home 和 Analytics 可以使用 AI 学习建议。

---

## 20.1 获取学习建议

```http
GET /api/learning/suggestions
```

Response：

```json
{
  "success": true,
  "data": [
    {
      "id": "suggestion-001",
      "type": "review",
      "courseId": "math",
      "title": "建议复习二重积分",
      "description": "近期错题中二重积分占比较高。",
      "priority": "high"
    }
  ],
  "message": null
}
```

当前阶段使用 Mock Data。

---

# 21. Settings API

Settings 分成：

```text
学习画像
AI 偏好
通用设置
```

---

# 22. AI Preference API

## 22.1 获取 AI 偏好

```http
GET /api/settings/ai
```

Response：

```json
{
  "success": true,
  "data": {
    "responseStyle": "standard",
    "teachingMode": "hint_first",
    "responseLength": "medium",
    "customPrompt": ""
  },
  "message": null
}
```

---

## 22.2 更新 AI 偏好

```http
PATCH /api/settings/ai
```

Request：

```json
{
  "responseStyle": "standard",
  "teachingMode": "hint_first",
  "responseLength": "medium",
  "customPrompt": ""
}
```

允许值：

### responseStyle

```text
concise
standard
detailed
```

### teachingMode

```text
direct_answer
hint_first
guided_question
teacher_style
```

### responseLength

```text
short
medium
long
```

---

# 23. General Settings API

## 23.1 获取通用设置

```http
GET /api/settings/general
```

Response：

```json
{
  "success": true,
  "data": {
    "language": "zh-CN",
    "theme": "system",
    "notifications": true
  },
  "message": null
}
```

---

## 23.2 更新通用设置

```http
PATCH /api/settings/general
```

Request：

```json
{
  "language": "zh-CN",
  "theme": "system",
  "notifications": true
}
```

---

# 24. AI Task API

未来所有 AI 请求建议统一经过 Task API。

基本结构：

```text
Frontend
    ↓
POST /api/ai/tasks
    ↓
Task Router
    ↓
Workflow / Agent
    ↓
Skill / Tool
    ↓
Model
```

---

## 24.1 创建 AI Task

```http
POST /api/ai/tasks
```

Request：

```json
{
  "courseId": "math",
  "type": "homework",
  "input": {
    "text": "请帮我分析这道题"
  }
}
```

Response：

```json
{
  "success": true,
  "data": {
    "taskId": "task-001",
    "status": "queued"
  },
  "message": null
}
```

---

# 25. AI Task 状态

```http
GET /api/ai/tasks/:taskId
```

Response：

```json
{
  "success": true,
  "data": {
    "taskId": "task-001",
    "status": "completed",
    "result": {
      "type": "hint",
      "content": "可以先尝试分析积分区域。"
    }
  },
  "message": null
}
```

任务状态：

```text
queued
processing
completed
failed
```

---

# 26. Router 信息

Lingxi-claw 的 Router 是内部实现。

Frontend 默认不需要知道 Router 如何工作。

但是为了：

* Demo
* 调试
* Explainable AI
* 后续成本 / 延迟展示

可以提供可选的 routing 信息。

例如：

```json
{
  "success": true,
  "data": {
    "taskId": "task-001",
    "status": "completed",
    "result": {},
    "routing": {
      "scene": "course_homework",
      "workflow": "homework",
      "skill": "math",
      "model": "mock-model",
      "inputType": "text"
    }
  },
  "message": null
}
```

Frontend 不应该依赖这些字段完成核心业务。

---

# 27. 文件上传与 AI Task 的关系

文件处理流程：

```text
Frontend
    ↓
Upload
    ↓
Material API
    ↓
Backend
    ↓
File Router
    ↓
Parser / OCR / Vision
    ↓
Structured Data
    ↓
Course Knowledge
```

AI Task：

```text
Frontend
    ↓
AI Task API
    ↓
Task Router
    ↓
Workflow / Agent
    ↓
Course Knowledge
    ↓
Model
```

二者不要混成一个 API。

---

# 28. API 与产品信息架构对应关系

```text
产品页面
        API
────────────────────────────────

Onboarding
        ↓
/api/profile

Home
        ↓
/api/home

Courses
        ↓
/api/courses

Course Space
        │
        ├── Overview
        │      ↓
        │  /api/courses/:courseId/overview
        │
        ├── Materials
        │      ↓
        │  /api/courses/:courseId/materials
        │
        ├── Homework
        │      ↓
        │  /api/courses/:courseId/homework
        │
        ├── Mistakes
        │      ↓
        │  /api/courses/:courseId/mistakes
        │
        └── Analytics
               ↓
           /api/courses/:courseId/analytics

Analytics
        ↓
/api/analytics

Settings
        ├── Profile
        │      ↓
        │  /api/profile
        │
        ├── AI
        │      ↓
        │  /api/settings/ai
        │
        └── General
               ↓
           /api/settings/general
```

---

# 29. Mock API 规范

V02 阶段 Frontend 可以使用 Mock API。

Mock API 必须：

1. 与 API Contract 保持一致
2. Request 参数与真实 API 一致
3. Response 字段与真实 API 一致
4. 不创建与 API.md 不一致的临时数据结构

例如：

```text
真实：

GET /api/courses

Mock：

mockApi.getCourses()
```

但 Mock 返回的数据结构必须与：

```text
GET /api/courses
```

完全一致。

---

# 30. API 类型定义

Frontend 和 Backend 应尽可能共享或同步 TypeScript 类型。

例如：

```ts
interface Course {
  id: string;
  name: string;
  code?: string;
  description?: string;
  progress: number;
  mastery: number;
}
```

API Response：

```ts
interface ApiResponse<T> {
  success: boolean;
  data: T | null;
  message: string | null;
}
```

---

# 31. 前后端协作原则

## Frontend

Frontend 只关心：

```text
Request
 ↓
Response
 ↓
UI
```

不应该关心：

```text
数据库怎么查
Workflow 怎么实现
模型怎么调用
OCR 怎么实现
Router 怎么实现
```

---

## Backend

Backend 负责：

```text
Request
 ↓
参数验证
 ↓
业务逻辑
 ↓
Workflow / Service / AI
 ↓
Response
```

Backend 不应该要求 Frontend 理解内部实现。

---

## QA

QA 根据 API.md 验证：

```text
URL
Method
Request
Response
Status Code
Error Case
```

---

# 32. API 修改流程

任何接口变化必须遵循：

```text
发现需求变化
      ↓
修改 API.md
      ↓
团队确认
      ↓
Backend 修改
      ↓
Frontend 修改
      ↓
QA 更新测试
      ↓
npm run build
      ↓
Integration Test
```

禁止：

```text
Backend：
“我顺手把字段名改了。”
```

或者：

```text
Frontend：
“这个字段没有，我自己猜一个。”
```

---

# 33. 当前 API 优先级

由于当前只有 2 天开发时间，API 分为三个等级。

## P0：必须完成

```text
GET    /api/home

GET    /api/courses
GET    /api/courses/:courseId
GET    /api/courses/:courseId/overview

GET    /api/courses/:courseId/materials
GET    /api/courses/:courseId/homework
GET    /api/courses/:courseId/mistakes
GET    /api/courses/:courseId/analytics

GET    /api/analytics

GET    /api/profile
GET    /api/settings/ai
GET    /api/settings/general
```

这些接口主要用于让 V02 页面可以完整展示。

---

## P1：第二优先级

```text
POST   /api/courses

PATCH  /api/courses/:courseId

POST   /api/courses/:courseId/materials

POST   /api/courses/:courseId/homework

PATCH  /api/courses/:courseId/mistakes/:mistakeId

PATCH  /api/profile

PATCH  /api/settings/ai

PATCH  /api/settings/general
```

用于基本交互。

---

## P2：后续 AI 联调

```text
POST   /api/ai/tasks

GET    /api/ai/tasks/:taskId

POST   /api/courses/:courseId/homework/:homeworkId/assist
```

这些接口可以先定义 Contract，暂时使用 Mock。

---

# 34. 当前禁止事项

当前阶段禁止因为实现 API 而：

```text
❌ 删除已有 services
❌ 删除已有 API Types
❌ 删除 Mock API
❌ 删除 LegacyApp
❌ 重写整个 Frontend
❌ 重写整个 Backend
❌ 提前接入真实模型
❌ 为了一个页面创建复杂的新架构
```

如果发现现有 API 与新产品架构冲突：

```text
先记录
    ↓
修改 API.md
    ↓
团队确认
    ↓
再修改代码
```

---

# 35. API Contract 最终目标

Lingxi-claw 的 API 应最终形成：

```text
                         Frontend
                            │
                            │ API
                            ↓
                    ┌──────────────┐
                    │    Backend   │
                    └──────┬───────┘
                           │
             ┌─────────────┼─────────────┐
             ↓             ↓             ↓
          Course        Learning       AI Task
           Service       Service       Service
             │             │             │
             ↓             ↓             ↓
         Materials      Analytics    Router
         Homework       Mistakes       │
         Course Data    Records        ↓
                                  Workflow / Agent
                                        │
                                        ↓
                                   Skill / Model
```

---

# 36. 一句话总结

> **API.md 是 Lingxi-claw 前端、后端和测试之间的接口合同。当前 V02 以 Course / Course Space 为核心组织 API，优先保证页面和 Mock 联调；未来再通过统一的 AI Task API 接入 Workflow、Router、Agent、Skill 和 Model，而不让前端直接依赖 AI 内部实现。**

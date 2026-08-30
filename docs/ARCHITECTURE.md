# Study with Lingxi-claw - 系统架构设计

## 1. 文档目的

本文档定义 Lingxi-claw 的整体技术架构，以及前端、后端、Workflow、Agent、文件处理和模型之间的关系。

核心目标：

1. 明确前后端之间的职责边界
2. 明确 Workflow 与 Agent 的调用关系
3. 明确文件、题目、学习状态和模型如何参与路由
4. 为团队协作和 Vibe Coding 提供统一的技术上下文
5. 明确当前 V02 阶段的实现范围，避免过早实现真实 AI 能力

---

# 2. 系统总体架构

Lingxi-claw 是一个面向大学生的个人 AI 学习系统。

系统整体分为五个主要层次：

```text
┌──────────────────────────────────────────────────────┐
│                    Frontend                          │
│                                                      │
│  Onboarding / Home / Courses / Course Space          │
│  Analytics / Settings                                │
│                                                      │
└────────────────────────┬─────────────────────────────┘
                         │
                    REST / API
                         │
                         ↓
┌──────────────────────────────────────────────────────┐
│                    Backend                           │
│                                                      │
│  API / Business Logic / User / Course / Task         │
│                                                      │
└────────────────────────┬─────────────────────────────┘
                         │
                         ↓
┌──────────────────────────────────────────────────────┐
│                AI Orchestration Layer                │
│                                                      │
│  Task Router                                         │
│      ↓                                               │
│  Workflow / General Agent                            │
│      ↓                                               │
│  Skill / Tool / Model                                │
│                                                      │
└────────────────────────┬─────────────────────────────┘
                         │
                         ↓
┌──────────────────────────────────────────────────────┐
│                  Data / AI Services                  │
│                                                      │
│  Course Data / Files / Questions / Wrong Answers     │
│  Learning Records / Knowledge Base / Models          │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

# 3. 产品架构与技术架构的关系

Lingxi-claw 的产品结构以**课程**为核心。

用户侧：

```text
首次进入
    ↓
学习画像
    ↓
Home
    ↓
我的课程
    ↓
Course Space
    ↓
具体学习任务
```

技术侧：

```text
用户操作
    ↓
Frontend
    ↓
API
    ↓
Backend
    ↓
Task / Workflow
    ↓
AI / Data Services
```

因此：

> **产品层围绕课程组织，技术层围绕任务和数据处理组织。**

---

# 4. Frontend 架构

## 4.1 Frontend 职责

Frontend 负责：

* 页面展示
* 路由
* 用户交互
* 状态展示
* 表单输入
* 文件上传入口
* API 调用
* Loading / Empty / Error 状态
* Workflow 状态展示
* AI 结果展示

Frontend 不负责：

* 数据库
* AI 模型调用
* Workflow 核心逻辑
* Agent 核心逻辑
* OCR
* 文件解析
* 向量检索
* 模型路由

---

# 5. Frontend 页面架构

当前 V02 页面结构：

```text
Frontend
│
├── Onboarding / 学习画像
│
├── Home
│
├── Courses
│   │
│   ├── 高等数学
│   ├── 大学物理
│   ├── 线性代数
│   └── 添加课程
│
├── Course Space
│   │
│   ├── 概览
│   ├── 课程资料知识库
│   ├── AI 作业辅导
│   ├── 错题记录
│   └── 学习分析
│
├── Analytics
│
└── Settings
    ├── 学习画像
    ├── AI 偏好
    └── 通用设置
```

---

# 6. Frontend 路由架构

当前核心路由：

```text
/home
/courses
/courses/:courseId
/analytics
/settings
```

未来可以进一步扩展：

```text
/onboarding

/courses/:courseId/overview
/courses/:courseId/materials
/courses/:courseId/homework
/courses/:courseId/mistakes
/courses/:courseId/analytics
```

但具体 URL 结构以实际前端实现为准。

原则：

> 路由用于表达产品信息架构，不应该直接暴露后端 Workflow 内部实现。

例如不应该设计：

```text
/final-sprint-agent
/document-parser
/task-router
```

因为这些属于系统内部能力。

---

# 7. AppShell 架构

Frontend 使用统一 AppShell：

```text
App
 │
 └── AppShell
      │
      ├── Sidebar
      │
      └── Main Content
            │
            └── Router Outlet
                  │
                  ├── Home
                  ├── Courses
                  ├── CourseSpace
                  ├── Analytics
                  └── Settings
```

AppShell 负责：

* 全局页面布局
* Sidebar
* 页面主体区域
* 全局导航
* 页面容器

具体页面只负责自己的内容。

---

# 8. Course Space 架构

Course Space 是整个产品的核心二级架构。

用户：

```text
Courses
    ↓
选择课程
    ↓
Course Space
```

Course Space：

```text
Course Space
│
├── Overview
│
├── Materials
│
├── Homework
│
├── Mistakes
│
└── Analytics
```

其中：

```text
Overview
    ↓
课程整体状态

Materials
    ↓
课程资料 / 知识库

Homework
    ↓
AI 作业辅导

Mistakes
    ↓
错题记录

Analytics
    ↓
课程学习分析
```

Course Space 的所有数据都必须绑定当前课程。

例如：

```text
/course/math
```

其中：

```text
courseId = math
```

后续 API 请求必须携带对应课程上下文。

---

# 9. Backend 架构

Backend 是 Frontend 与 AI / 数据服务之间的业务层。

Backend 主要负责：

```text
API
 ↓
身份 / 用户
 ↓
课程
 ↓
学习任务
 ↓
Workflow
 ↓
AI 能力
 ↓
数据持久化
```

Backend 不应该把所有逻辑全部塞进 API Controller。

推荐：

```text
API Layer
    ↓
Service Layer
    ↓
Workflow / AI Layer
    ↓
Data Layer
```

---

# 10. Backend API Layer

API Layer 负责：

* 接收 Frontend 请求
* 参数校验
* 身份校验
* 调用对应 Service
* 返回统一格式
* 错误处理

例如：

```text
GET /api/courses
```

用于获取课程列表。

```text
GET /api/courses/:courseId
```

用于获取课程信息。

```text
GET /api/courses/:courseId/materials
```

用于获取课程资料。

```text
GET /api/courses/:courseId/homework
```

用于获取课程作业。

```text
GET /api/courses/:courseId/mistakes
```

用于获取错题。

```text
GET /api/courses/:courseId/analytics
```

用于获取课程学习分析。

具体接口定义以 `API.md` 为准。

---

# 11. AI Orchestration Layer

AI Orchestration Layer 是系统 AI 能力的核心。

整体结构：

```text
用户任务
    ↓
Task Router
    ↓
判断任务类型
    ↓
┌───────────────┬────────────────┐
│               │                │
Workflow     General Agent       │
│               │                │
↓               ↓                │
结构化执行     动态规划           │
└───────────────┴────────────────┘
                ↓
             Skill / Tool
                ↓
              Model
```

---

# 12. Task Router

Task Router 负责判断：

```text
当前是什么任务？
```

可以使用以下上下文：

```text
用户
课程
当前页面
用户输入
文件类型
任务类型
学习状态
任务复杂度
```

例如：

```text
用户在：

高等数学
    ↓
AI 作业辅导
    ↓
上传一道数学题
```

Router 可以判断：

```text
课程 = 高等数学
任务 = 作业辅导
输入 = 图片
题型 = 数学题
```

然后选择：

```text
Homework Workflow
    ↓
Vision / OCR
    ↓
Math Skill
    ↓
对应模型
```

---

# 13. Workflow 层

Workflow 用于处理明确、结构化的学习任务。

典型 Workflow：

```text
AI Homework Workflow
Materials Workflow
Mistake Review Workflow
Learning Analysis Workflow
```

Workflow 负责：

* 定义步骤
* 控制执行顺序
* 调用对应 Skill
* 调用工具
* 保存任务状态
* 返回结构化结果

例如 AI 作业辅导：

```text
Homework Workflow
    ↓
识别题目
    ↓
识别知识点
    ↓
判断用户状态
    ↓
选择提示等级
    ↓
生成提示
    ↓
用户继续作答
    ↓
检查
```

---

# 14. General Agent

General Agent 用于处理无法被固定 Workflow 覆盖的问题。

例如：

```text
用户：
为什么二重积分需要换元？
```

如果不属于明确的固定任务：

```text
Task Router
    ↓
General Agent
    ↓
理解问题
    ↓
决定是否需要：
├── 直接回答
├── 检索资料
├── 工具调用
└── 多步骤规划
```

General Agent 是系统的兜底能力。

原则：

> **Workflow 优先，Agent 兜底。**

---

# 15. Skill 层

Skill 是 Workflow / Agent 可以调用的具体能力。

例如：

```text
Math Skill
Physics Skill
Question Analysis Skill
Learning Plan Skill
Grading Skill
Hint Skill
```

Skill 与 Workflow 的关系：

```text
Workflow
    ↓
调用 Skill
    ↓
完成具体任务
```

例如：

```text
Homework Workflow
    ↓
Question Recognition
    ↓
Math Skill
    ↓
Hint Service
```

Skill 不负责决定整个用户任务应该怎么执行。

---

# 16. 文件处理架构

Lingxi-claw 支持多种学习资料。

文件进入系统后：

```text
Upload
    ↓
File Type Detection
    ↓
Parser Router
    ↓
┌────────┬────────┬──────────┬──────────┐
│ DOCX   │ PDF    │ Scan PDF │ Image    │
└───┬────┴───┬────┴────┬─────┴────┬─────┘
    ↓        ↓         ↓          ↓
 Text      PDF      OCR        OCR/Vision
 Parser    Parser
    └────────┴─────────┴──────────┘
                  ↓
             Structured Data
                  ↓
             Course Knowledge
```

文件处理属于 Backend / AI 服务能力。

Frontend 只负责：

```text
选择文件
    ↓
上传
    ↓
显示上传状态
    ↓
显示处理状态
```

---

# 17. 数据架构

核心数据对象：

```text
User
Course
CourseMaterial
Homework
Question
WrongAnswer
LearningRecord
LearningProfile
LearningAnalysis
AIPreference
```

基本关系：

```text
User
 │
 ├── LearningProfile
 │
 └── Courses
       │
       ├── Materials
       │
       ├── Homework
       │      └── Questions
       │
       ├── WrongAnswers
       │
       └── LearningRecords
```

课程是重要的数据隔离单位。

例如：

```text
User
 ├── 高等数学
 │    ├── 资料
 │    ├── 作业
 │    └── 错题
 │
 ├── 大学物理
 │    ├── 资料
 │    ├── 作业
 │    └── 错题
 │
 └── 线性代数
      ├── 资料
      ├── 作业
      └── 错题
```

具体数据字段以 `DATA_SCHEMA.md` 为准。

---

# 18. 学习数据闭环

系统最终形成：

```text
课程资料
    ↓
学习任务
    ↓
作业 / 练习
    ↓
用户作答
    ↓
学习记录
    ↓
错题
    ↓
学习分析
    ↓
发现薄弱知识点
    ↓
AI 学习建议
    ↓
新的学习任务
```

数据最终服务于：

```text
Home
Course Space
Analytics
AI Learning Suggestion
```

---

# 19. Analytics 架构

Analytics 分为两个层级。

## 19.1 Course Analytics

针对单门课程：

```text
Course Space
    ↓
Learning Analytics
    ↓
当前课程数据
```

分析：

* 当前课程掌握度
* 知识点
* 错题
* 作业
* 学习趋势

---

## 19.2 Global Analytics

针对用户整体：

```text
Analytics
    ↓
多课程数据
    ↓
整体学习状态
```

分析：

* 整体学习情况
* 各课程掌握度
* 知识薄弱点
* 作业表现
* 学习趋势

---

# 20. AI Model Layer

模型层不应该直接暴露给 Frontend。

整体：

```text
Frontend
    ↓
Backend
    ↓
Router
    ↓
Workflow / Agent
    ↓
Model Service
```

模型选择未来可以根据：

```text
任务复杂度
输入模态
任务类型
响应速度
成本
模型能力
```

进行动态选择。

例如：

```text
简单文本问题
    ↓
轻量模型

普通学习问题
    ↓
通用模型

复杂数学 / 多模态问题
    ↓
高能力模型
```

---

# 21. Token / 成本优化架构

Lingxi-claw 可以保留原有 Token 优化设计。

未来可以：

```text
用户输入
    ↓
轻量预处理
    ↓
Token Pruning
    ↓
保留有效上下文
    ↓
Task Router
    ↓
选择模型
```

同时记录：

```text
输入 Token
输出 Token
模型
响应时间
估算成本
```

用于未来的：

```text
Cost Analytics
Latency Analytics
```

当前 V02 阶段：

> 不要求重新实现 Token Pruning 或模型成本优化。

---

# 22. API Contract

Frontend 与 Backend 必须通过 API Contract 协作。

基本原则：

```text
Frontend
    ↓
API Contract
    ↓
Backend
```

双方不能通过：

```text
直接读取数据库
直接调用内部 Python 模块
直接修改对方代码
```

进行耦合。

API 的请求和响应结构统一记录在：

```text
docs/API.md
```

如果 API 发生变化：

```text
Backend 修改
    ↓
更新 API.md
    ↓
Frontend 根据 Contract 调整
    ↓
QA 回归测试
```

---

# 23. Mock Mode

当前 V02 阶段必须保留 Mock Mode。

系统可以：

```text
Frontend
    ↓
Mock API
    ↓
Mock Data
```

或者：

```text
Frontend
    ↓
Real API
    ↓
Backend
```

通过配置切换。

Mock Mode 用于：

* UI 开发
* 页面联调
* 路由测试
* API 占位
* Demo
* QA 测试

Mock 数据必须与未来 API 数据结构尽可能一致。

---

# 24. LegacyApp 兼容架构

项目原有 LegacyApp 必须保留。

当前架构允许：

```text
                    Lingxi
                       │
             ┌─────────┴─────────┐
             │                   │
          V02 App              LegacyApp
             │                   │
        新产品架构          原有完整 Demo
```

LegacyApp 中已有能力包括：

* 期末突击
* 日常作业辅助
* General Question
* Agent Settings
* 文件上传
* 文件分析 Mock
* 复习计划 Mock
* Chat Mock

V02 重构不能因为新的产品架构而删除这些能力。

尤其禁止未经确认删除：

```text
services/
API Types
API Calls
Mock API
LegacyApp
```

---

# 25. 当前 V02 技术范围

当前阶段目标：

> **建立完整的产品信息架构、前端页面架构和前后端协作边界。**

---

## 25.1 当前已实现 / 应实现

```text
Frontend
├── React
├── React Router
├── AppShell
├── Sidebar
├── PageHeader
├── Card
├── Home
├── Courses
├── Course Space
├── Analytics
└── Settings
```

并且：

```text
npm run build
```

必须保持通过。

---

## 25.2 当前只需要页面框架

```text
学习画像
Home
Courses
Course Space
    ├── 概览
    ├── 课程资料知识库
    ├── AI 作业辅导
    ├── 错题记录
    └── 学习分析
Analytics
Settings
```

这些页面当前可以使用 Mock Data。

---

# 26. 当前暂不实现

V02 阶段暂不要求：

```text
❌ 真实 AI 模型调用
❌ 真实 Workflow
❌ 真实 General Agent
❌ 真实 Task Router
❌ 真实 OCR
❌ 真实 Vision
❌ 真实文件解析
❌ 真实向量数据库
❌ 真实知识库
❌ 真实学习分析算法
❌ 真实个性化推荐
❌ 真实 Token Pruning
❌ 真实模型成本优化
```

这些属于后续阶段。

---

# 27. 三人团队职责边界

项目由三类角色协作：

```text
Frontend
Backend
QA / Testing
```

---

## 27.1 Frontend

负责：

```text
pages/
components/
routes/
styles/
UI state
API integration
```

核心目标：

> 将产品信息架构实现成可交互页面。

Frontend 不应：

* 修改 Backend 核心逻辑
* 修改数据库结构
* 修改 AI Workflow
* 删除 services
* 删除已有 API 类型

---

## 27.2 Backend

负责：

```text
API
Business Logic
Data
Workflow
Agent
AI Services
File Processing
```

核心目标：

> 为 Frontend 提供稳定的 API 和未来 AI 能力。

Backend 不应：

* 直接修改 Frontend UI
* 改变前端页面结构而不更新 API Contract
* 删除现有 Mock API
* 未沟通修改 Frontend 文件

---

## 27.3 QA / Testing

负责：

```text
Functional Testing
API Testing
Integration Testing
Regression Testing
UI Testing
```

核心目标：

> 确保 Frontend + Backend + API + Workflow 在用户视角下能够正常工作。

QA 重点验证：

```text
页面
 ↓
路由
 ↓
API
 ↓
数据
 ↓
Workflow
 ↓
结果
```

---

# 28. 文件修改边界

三人协作时必须明确文件所有权。

推荐：

```text
Frontend
├── frontend/pages/
├── frontend/components/
├── frontend/styles/
└── frontend/routes/

Backend
├── backend/
├── services/
└── AI / Workflow

QA
├── tests/
├── test reports
└── test documentation
```

共享文件：

```text
docs/
API Types
package.json
配置文件
```

共享文件必须：

> **先约定，再修改。**

---

# 29. Vibe Coding 协作规则

由于项目使用 Kilo Code 等 AI Coding Agent，必须限制 AI 的修改范围。

每次任务应该明确：

```text
1. 修改哪些文件
2. 不允许修改哪些文件
3. 当前任务目标
4. 不实现哪些功能
5. 完成后的验证方式
```

例如 Frontend：

```text
任务：
完善 CourseSpace.tsx 页面。

允许修改：
frontend/pages/CourseSpace.tsx

禁止修改：
App.tsx
services/
API types
backend/
其他 pages/

要求：
1. 保持现有路由
2. 使用已有 AppShell
3. 使用已有 Card / PageHeader
4. 使用 Mock Data
5. 不实现真实 AI
6. npm run build 必须通过
```

---

# 30. 开发流程

推荐开发流程：

```text
Product
    ↓
Workflow
    ↓
Architecture
    ↓
API Contract
    ↓
Frontend / Backend
    ↓
Integration
    ↓
QA
```

具体：

```text
产品定义
    ↓
确定页面与用户流程
    ↓
确定 Workflow
    ↓
确定数据模型
    ↓
确定 API
    ↓
Frontend / Backend 并行开发
    ↓
API 联调
    ↓
QA 测试
    ↓
修复问题
    ↓
Demo
```

---

# 31. 系统最终架构

最终 Lingxi-claw 的完整技术架构：

```text
                         User
                           │
                           ↓
                    ┌─────────────┐
                    │  Frontend   │
                    │             │
                    │ Onboarding  │
                    │ Home        │
                    │ Courses     │
                    │ CourseSpace │
                    │ Analytics   │
                    │ Settings    │
                    └──────┬──────┘
                           │
                         API
                           │
                           ↓
                    ┌─────────────┐
                    │   Backend   │
                    │             │
                    │ User        │
                    │ Course      │
                    │ Task        │
                    │ Data        │
                    └──────┬──────┘
                           │
                           ↓
                 ┌───────────────────┐
                 │ AI Orchestration  │
                 │                   │
                 │ Task Router       │
                 │      ↓            │
                 │ Workflow / Agent  │
                 │      ↓            │
                 │ Skill / Tool      │
                 └────────┬──────────┘
                          │
             ┌────────────┼────────────┐
             ↓            ↓            ↓
        File Parser    Knowledge     Model
        OCR / Vision     Base       Service
             │            │            │
             └────────────┼────────────┘
                          ↓
                    Learning Data
                          │
             ┌────────────┼────────────┐
             ↓            ↓            ↓
          Homework      Mistakes   Analytics
             │            │            │
             └────────────┼────────────┘
                          ↓
                   AI Suggestions
                          │
                          ↓
                       Home
```

---

# 32. 架构核心原则

Lingxi-claw 遵循以下架构原则：

### 1. Product First

技术架构服务于产品信息架构。

### 2. Course Centered

课程是用户学习数据组织的核心单位。

### 3. Frontend / Backend Separation

Frontend 负责展示与交互，Backend 负责业务与 AI 能力。

### 4. Workflow First

明确、结构化的任务优先进入 Workflow。

### 5. Agent Fallback

开放、复杂、未知任务由 General Agent 处理。

### 6. API Contract

Frontend 与 Backend 通过明确 API Contract 协作。

### 7. Mock First

早期开发优先使用 Mock，降低并行开发耦合。

### 8. Preserve Existing Capability

新架构不能破坏已有 services、API、Mock 和 LegacyApp。

### 9. Small Scope for Vibe Coding

AI Coding Agent 每次只处理明确、小范围的任务。

### 10. Testable

每一个页面、API 和 Workflow 都必须能够独立验证。

---

# 33. 当前阶段的一句话架构总结

> **Lingxi-claw 以课程为产品核心，以 Course Space 为主要学习工作空间，通过 Frontend + Backend + API Contract 解耦协作，并在后端通过 Workflow + Agent + Router 逐步构建真正的 AI 学习能力；V02 阶段优先完成产品架构、页面框架、路由和 Mock 联调，不提前实现真实 AI 能力。**

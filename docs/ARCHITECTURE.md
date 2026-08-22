# Study with Lingxi-claw - 系统架构设计

## 1. 文档目的

本文档定义 Lingxi-claw 的整体技术架构，以及前端、后端、Workflow、Router、Agent、文件处理和模型之间的关系。

核心目标：

1. 明确前后端之间的职责边界
2. 明确 Workflow 与 Agent 的调用关系
3. 明确 Router 在系统中的位置
4. 明确文件、题目、学习状态和模型如何参与路由
5. 为团队协作和 Vibe Coding 提供统一的技术上下文

---

## 2. 核心架构理念

Lingxi-claw 不是简单的：

```text
用户
 ↓
大模型
 ↓
回答
````

而是：

```text
用户
 ↓
前端 Sidebar
 ↓
选择学习模式
 ↓
Workflow / Agent
 ↓
内部 Router
 ↓
根据任务进行精准分流
 ↓
文件 / 题型 / 学习状态 / 模型
 ↓
执行对应能力
 ↓
生成学习结果
 ↓
返回前端
```

核心设计原则：

> **用户选择大场景，Workflow 负责流程，Router 负责分流，Agent 负责复杂任务。**

---

## 3. 系统整体架构

```text
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                            │
│                                                             │
│  Study with Lingxi-claw                                    │
│                                                             │
│  Sidebar                                                    │
│  ├── 📚 期末突击                                            │
│  ├── 📝 日常作业辅助                                        │
│  ├── ❓ 其它问题                                            │
│  └── 🤖 Agent 设定                                          │
│                                                             │
│                     ↓ HTTP / API ↓                         │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────┐
│                         Backend                             │
│                                                             │
│                         API Layer                           │
│                              ↓                              │
│                         Main Router                         │
│                              ↓                              │
│                    Workflow / Agent Layer                   │
│                              ↓                              │
│                    Internal Router Layer                    │
│                              ↓                              │
│       ┌────────────┬────────────┬────────────┬──────────┐  │
│       ↓            ↓            ↓            ↓          │  │
│   File Router  Question      State Router  Model      │  │
│                Router                       Router      │  │
│       ↓            ↓            ↓            ↓          │  │
│   File Parser   Question     Learning      LLM /      │  │
│   OCR / DOCX    Classifier    State        Model       │  │
│                / Skill                    Provider      │  │
│       └────────────┴────────────┴────────────┴──────────┘  │
│                              ↓                              │
│                       Services Layer                        │
│                              ↓                              │
│                    Data / Knowledge Base                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. 用户请求的完整生命周期

一个请求从用户输入到最终返回，大致经历以下过程：

```text
用户
 ↓
Frontend
 ↓
API Request
 ↓
Main Router
 ↓
识别当前学习模式
 ↓
Workflow / Agent
 ↓
内部 Router
 ↓
选择具体处理能力
 ↓
执行 Service / Tool / Model
 ↓
生成结果
 ↓
Schema 校验
 ↓
API Response
 ↓
Frontend
 ↓
展示结果
```

---

## 5. 第一层：Frontend

Frontend 负责：

* 页面展示
* Sidebar 模式切换
* 文件上传
* 用户输入
* 对话展示
* 学习计划展示
* 题目展示
* 作答交互
* Agent 设置
* Loading / Error 状态

Frontend **不负责**：

* 判断使用哪个模型
* 判断调用哪个 Agent
* 解析复杂文件
* 统计历年考点
* 生成复习计划
* 决定 Workflow 的具体执行逻辑

这些逻辑统一放在 Backend。

---

## 6. Sidebar：第一层场景路由

用户通过 Sidebar 主动选择学习场景。

```text
Sidebar
   │
   ├── 📚 期末突击
   │       ↓
   │   final_sprint
   │
   ├── 📝 日常作业辅助
   │       ↓
   │   homework
   │
   ├── ❓ 其它问题
   │       ↓
   │   general
   │
   └── 🤖 Agent 设定
           ↓
       settings
```

Frontend 向 Backend 发送：

```json
{
  "mode": "final_sprint"
}
```

或者：

```json
{
  "mode": "homework"
}
```

---

## 7. API Layer

Backend 的 API Layer 是前后端之间的统一入口。

建议按照功能划分 API：

```text
backend/app/api/
```

例如：

```text
POST /api/final-sprint/upload
POST /api/final-sprint/analyze
POST /api/final-sprint/plan
POST /api/final-sprint/practice

POST /api/homework/upload
POST /api/homework/analyze
POST /api/homework/check

POST /api/chat

GET  /api/settings
POST /api/settings
```

API Layer 的主要职责：

```text
接收请求
 ↓
参数校验
 ↓
调用对应 Workflow / Agent
 ↓
返回标准 Response
```

API Layer 不应该直接编写复杂业务逻辑。

错误示例：

```text
API
 ↓
直接解析 PDF
 ↓
直接调用 LLM
 ↓
直接生成复习计划
```

正确方式：

```text
API
 ↓
Workflow
 ↓
Router
 ↓
Service
 ↓
Model
```

---

## 8. Main Router：主场景路由

用户通过 Sidebar 选择模式之后，Backend 首先进行主场景路由。

```text
用户请求
 ↓
Main Router
 ↓
┌────────────────────────────────┐
│                                │
│ final_sprint → Final Sprint    │
│ homework     → Homework        │
│ general      → General Agent   │
│ settings     → Settings        │
│                                │
└────────────────────────────────┘
```

Main Router 只负责：

> **决定用户应该进入哪个主 Workflow / Agent。**

它不负责处理具体题目。

---

## 9. Workflow Layer

Workflow 是系统中最重要的业务流程层。

目录：

```text
backend/app/workflows/
```

建议：

```text
workflows/
├── final_sprint.py
├── homework.py
└── ...
```

Workflow 负责组织多个步骤。

例如：

```text
Final Sprint Workflow

文件上传
 ↓
文件解析
 ↓
题目提取
 ↓
题型分类
 ↓
考点分析
 ↓
复习计划
 ↓
刷题
 ↓
学习反馈
```

Workflow 本身不应该实现所有底层功能。

例如：

```text
Workflow
 ↓
调用 File Service
 ↓
调用 Question Service
 ↓
调用 Exam Analysis Service
 ↓
调用 Study Planner
```

---

## 10. Agent Layer

目录：

```text
backend/app/agents/
```

Agent 用于处理：

* 开放式问题
* 非固定流程
* 多步骤复杂任务
* 需要自主决定工具的任务
* Workflow 无法覆盖的请求

核心原则：

> **Workflow 解决确定性问题，Agent 解决开放性问题。**

---

## 11. Workflow + Agent 关系

系统不是 Workflow 和 Agent 二选一。

而是：

```text
                 用户
                   ↓
              主场景选择
                   ↓
        ┌──────────┴──────────┐
        ↓                     ↓
    Workflow              Agent
        ↓                     ↓
确定流程任务             开放复杂任务
        ↓                     ↓
内部 Router              Agent Router
        ↓                     ↓
Services / Tools         Tools / Models
```

例如：

#### 期末突击

```text
Final Sprint Workflow
        ↓
File Router
        ↓
Question Router
        ↓
Exam Analysis
        ↓
Study Planner
```

#### 日常作业

```text
Homework Workflow
        ↓
File Router
        ↓
Question Router
        ↓
Homework State Router
        ↓
Answer Checker
```

#### 其它问题

```text
General Agent
        ↓
Request Router
        ↓
Complexity Router
        ↓
Tool / Model Router
        ↓
Agent
```

---

## 12. Internal Router Layer

Router 是 Lingxi-claw 的核心技术之一。

Router 的目标不是简单地：

> “选择一个大模型。”

而是根据不同维度，把请求分配给最合适的处理模块。

主要包括：

```text
Router
│
├── Main Router
│
├── File Router
│
├── Question Router
│
├── State Router
│
├── Model Router
│
└── Tool Router
```

---

## 13. File Router

File Router 判断输入文件应该使用什么解析方式。

```text
文件
 ↓
File Router
 ↓
┌────────────┬────────────┬────────────┬────────────┐
↓            ↓            ↓            ↓
DOCX         PDF          Scan PDF     Image
↓            ↓            ↓            ↓
DOCX Parser  PDF Parser   OCR          OCR/Vision
```

例如：

```json
{
  "file_type": "pdf",
  "is_scanned": true
}
```

Router：

```text
PDF + scanned
 ↓
OCR
```

而：

```json
{
  "file_type": "pdf",
  "is_scanned": false
}
```

则：

```text
PDF
 ↓
Text Parser
```

---

## 14. Question Router

Question Router 判断题目的：

* 学科
* 题型
* 知识点
* 难度
* 是否需要多模态能力
* 是否需要复杂模型

例如：

```text
题目
 ↓
Question Router
 ↓
数学
 ↓
二重积分
 ↓
计算题
 ↓
Medium
```

然后将结果交给对应的 Skill / Service。

---

## 15. State Router

State Router 判断用户当前处于什么学习状态。

主要应用于日常作业辅助。

```text
用户状态
 ↓
State Router
 ↓
┌────────────────┬────────────────┬────────────────┐
↓                ↓                ↓
刚开始做题       正在尝试         已提交答案
↓                ↓                ↓
方向提示         方法提示         答案检查
```

如果用户多次尝试仍然无法解决：

```text
多次失败
 ↓
State Router
 ↓
升级帮助等级
 ↓
完整解题过程
```

---

## 16. Model Router

Model Router 决定：

> 当前任务需要什么级别的模型？

核心原则：

> **简单任务使用低成本处理，复杂任务才调用更强模型。**

```text
请求
 ↓
Complexity Analysis
 ↓
┌──────────────┬──────────────┬──────────────┐
↓              ↓              ↓
简单            普通            复杂
↓              ↓              ↓
轻量模型        标准模型        大模型 / Agent
```

例如：

```text
“什么是概率？”
 ↓
轻量模型
```

```text
“解释贝叶斯公式并举例”
 ↓
标准模型
```

```text
“分析我上传的五年期末试卷并制定复习计划”
 ↓
Workflow
 ↓
多个 Service
 ↓
必要时调用强模型
```

---

## 17. Tool Router

某些请求不需要直接调用模型，而是需要工具。

例如：

```text
用户请求
 ↓
Tool Router
 ↓
┌───────────────┬───────────────┬───────────────┐
↓               ↓               ↓
文件解析         知识库检索       计算工具
```

未来可以继续扩展：

```text
Tool Router
├── File Parser
├── OCR
├── Knowledge Retrieval
├── Calculator
├── Code Execution
└── Search
```

---

## 18. Services Layer

目录：

```text
backend/app/services/
```

Services 是具体能力的实现层。

建议按照职责拆分：

```text
services/
├── file_service.py
├── ocr_service.py
├── question_service.py
├── knowledge_service.py
├── exam_analysis_service.py
├── study_plan_service.py
├── question_selector.py
├── answer_checker.py
└── llm_service.py
```

---

## 19. Services 与 Workflow 的关系

Workflow 负责：

> **决定先做什么、后做什么。**

Service 负责：

> **具体把某件事情做好。**

例如：

```text
Final Sprint Workflow
        ↓
File Service
        ↓
Question Service
        ↓
Exam Analysis Service
        ↓
Study Plan Service
        ↓
Question Selector
```

不要把所有代码全部写进 Workflow。

---

## 20. Schemas Layer

目录：

```text
backend/app/schemas/
```

Schemas 用于定义系统内部统一的数据结构。

核心对象包括：

```text
Dataset
File
Question
KnowledgePoint
ExamAnalysis
StudyPlan
PracticeRecord
UserLearningState
AgentProfile
```

---

## 21. Dataset 数据关系

期末突击的核心数据结构：

```text
Dataset
 │
 ├── Files
 │
 ├── Questions
 │
 ├── Knowledge Points
 │
 ├── Exam Analysis
 │
 └── Study Plan
```

更完整：

```text
Dataset
   │
   ├── File[]
   │      ↓
   │   ParsedContent
   │
   ├── Question[]
   │      ↓
   │   QuestionClassification
   │
   ├── KnowledgePoint[]
   │
   ├── ExamAnalysis
   │
   └── StudyPlan
```

---

## 22. Backend 目录与架构对应关系

当前项目结构：

```text
backend/
└── app/
    ├── agents/
    ├── api/
    ├── routers/
    ├── schemas/
    ├── services/
    └── workflows/
```

对应关系：

```text
agents/
    ↓
复杂开放任务

api/
    ↓
前后端接口

routers/
    ↓
内部路由与任务分流

schemas/
    ↓
统一数据结构

services/
    ↓
具体能力

workflows/
    ↓
业务流程
```

---

## 23. 推荐的后端结构

随着项目开发，可以逐渐形成：

```text
backend/
│
├── app/
│   │
│   ├── agents/
│   │   ├── general_agent.py
│   │   └── ...
│   │
│   ├── api/
│   │   ├── final_sprint.py
│   │   ├── homework.py
│   │   ├── chat.py
│   │   └── settings.py
│   │
│   ├── routers/
│   │   ├── main_router.py
│   │   ├── file_router.py
│   │   ├── question_router.py
│   │   ├── state_router.py
│   │   ├── model_router.py
│   │   └── tool_router.py
│   │
│   ├── schemas/
│   │   ├── dataset.py
│   │   ├── file.py
│   │   ├── question.py
│   │   ├── study_plan.py
│   │   └── agent.py
│   │
│   ├── services/
│   │   ├── file_service.py
│   │   ├── ocr_service.py
│   │   ├── question_service.py
│   │   ├── exam_analysis_service.py
│   │   ├── study_plan_service.py
│   │   ├── question_selector.py
│   │   ├── answer_checker.py
│   │   └── llm_service.py
│   │
│   └── workflows/
│       ├── final_sprint.py
│       └── homework.py
│
└── tests/
```

---

## 24. 期末突击的代码调用关系

```text
POST /api/final-sprint/upload
        ↓
FinalSprintWorkflow
        ↓
BatchFileProcessor
        ↓
FileRouter
        ↓
FileService / OCRService
        ↓
QuestionService
        ↓
QuestionRouter
        ↓
QuestionBank
        ↓
ExamAnalysisService
        ↓
StudyPlanService
        ↓
QuestionSelector
        ↓
API Response
```

---

## 25. 日常作业的代码调用关系

```text
POST /api/homework/upload
        ↓
HomeworkWorkflow
        ↓
FileRouter
        ↓
FileService / OCRService
        ↓
QuestionService
        ↓
QuestionRouter
        ↓
HomeworkStateRouter
        ↓
AnswerChecker
        ↓
必要时调用 LLM / Agent
        ↓
API Response
```

---

## 26. 其它问题的代码调用关系

```text
POST /api/chat
        ↓
Main Router
        ↓
General Agent
        ↓
Request Router
        ↓
Complexity Router
        ↓
Tool Router / Model Router
        ↓
调用对应能力
        ↓
General Agent
        ↓
API Response
```

---

## 27. Agent 与 Router 的关系

Agent 不应该直接决定所有底层实现。

推荐：

```text
General Agent
      ↓
Router
      ↓
Tool / Model / Knowledge
      ↓
返回结果
      ↓
Agent 综合
      ↓
最终回答
```

Agent 更关注：

> **“我要完成什么任务？”**

Router 更关注：

> **“这个任务应该交给谁处理？”**

---

## 28. 前后端接口原则

前后端通过 API 进行通信。

Frontend 不直接访问：

* LLM API
* OCR API
* 文件解析器
* 数据库
* Agent

统一：

```text
Frontend
   ↓
Backend API
   ↓
Workflow / Agent
```

这样可以避免 API Key 暴露在浏览器端。

---

## 29. API 请求示例

### 29.1 创建期末复习任务

```json
POST /api/final-sprint/create

{
  "course": "高等数学",
  "exam_date": "2026-08-30",
  "daily_study_hours": 4
}
```

返回：

```json
{
  "dataset_id": "dataset_001",
  "status": "created"
}
```

---

### 29.2 批量上传文件

```text
POST /api/final-sprint/upload
```

请求：

```text
dataset_id
files[]
```

返回：

```json
{
  "dataset_id": "dataset_001",
  "files": [
    {
      "file_id": "file_001",
      "name": "2024期末试题.pdf",
      "status": "processing"
    },
    {
      "file_id": "file_002",
      "name": "2025期末试题.docx",
      "status": "processing"
    }
  ]
}
```

---

## 30. API Response 统一原则

所有 API 返回尽量保持统一结构。

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

失败时：

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "FILE_PARSE_ERROR",
    "message": "文件解析失败"
  }
}
```

这样前端不需要针对每个接口设计完全不同的错误处理方式。

---

## 31. 错误处理原则

任何一个子任务失败，不应该默认导致整个 Workflow 崩溃。

例如：

```text
上传 10 个文件
        ↓
其中 1 个 PDF OCR 失败
        ↓
其他 9 个文件继续处理
        ↓
最终告诉用户：
9 个文件成功
1 个文件失败
```

而不是：

```text
1 个文件失败
 ↓
整个任务失败
```

批处理任务应该尽量支持：

```text
partial success
```

即：

> **允许部分成功。**

---

## 32. 异步任务原则

批量文件解析、OCR、题目分类等任务可能耗时较长。

因此不要让前端一直等待一个 HTTP 请求。

推荐：

```text
用户上传文件
        ↓
API
        ↓
创建后台任务
        ↓
返回 task_id
        ↓
后台处理
        ↓
Frontend 查询任务状态
```

例如：

```json
{
  "task_id": "task_001",
  "status": "processing"
}
```

前端可以显示：

```text
正在处理资料...

████████████░░░░ 75%

已处理 6 / 8 个文件
```

---

## 33. Vibe Coding 开发原则

团队使用 AI Coding 工具开发时，所有成员必须遵守：

> **先阅读 docs，再写代码。**

AI 在生成代码前应该知道：

```text
PRODUCT.md
WORKFLOW.md
ARCHITECTURE.md
API.md
DATA_SCHEMA.md
AI_CONTEXT.md
```

避免每个人根据自己的理解生成不同架构。

---

## 34. 模块职责边界

### Frontend

负责：

```text
UI
交互
文件选择
API 调用
结果展示
```

不负责：

```text
业务 Workflow
模型选择
API Key
核心 AI 逻辑
```

---

### API

负责：

```text
接收请求
参数验证
调用 Workflow
返回 Response
```

---

### Workflow

负责：

```text
组织业务流程
决定执行顺序
调用 Router / Service
```

---

### Router

负责：

```text
任务分类
能力选择
模型选择
工具选择
```

---

### Service

负责：

```text
完成具体能力
```

---

### Agent

负责：

```text
复杂任务理解
任务规划
多步骤执行
开放问题回答
```

---

## 35. 最重要的架构原则

Lingxi-claw 的核心调用链：

```text
                USER
                  │
                  ↓
             FRONTEND
                  │
                  ↓
              API LAYER
                  │
                  ↓
             MAIN ROUTER
                  │
          ┌───────┴────────┐
          ↓                ↓
      WORKFLOW           AGENT
          │                │
          └───────┬────────┘
                  ↓
           INTERNAL ROUTERS
                  │
       ┌──────────┼──────────┐
       ↓          ↓          ↓
     FILE       TASK       MODEL
    ROUTER      ROUTER     ROUTER
       │          │          │
       ↓          ↓          ↓
   PARSER       SKILL       LLM
   / OCR        / TOOL     PROVIDER
       │          │          │
       └──────────┼──────────┘
                  ↓
              SERVICES
                  ↓
             DATA / KB
                  ↓
              RESULT
                  ↓
             API RESPONSE
                  ↓
              FRONTEND
                  ↓
                USER
```

---

## 36. 最终架构总结

Lingxi-claw 的系统架构可以概括为：

> **用户通过 Sidebar 主动选择学习场景，Backend 根据场景进入对应 Workflow 或 General Agent；Workflow / Agent 在执行过程中通过多个轻量 Router，对文件类型、题型、用户学习状态、工具和模型进行精准分流，再由 Services 完成具体任务，最终将结构化结果返回前端。**

核心链路：

```text
用户点击 Sidebar
        ↓
选择学习模式
        ↓
Workflow / Agent
        ↓
内部 Router
        ↓
文件 / 题型 / 状态 / 工具 / 模型
        ↓
Services
        ↓
数据 / 知识库 / LLM
        ↓
生成结果
        ↓
API Response
        ↓
Frontend
        ↓
用户
```

核心原则：

```text
Workflow 优先
        +
Agent 兜底
        +
Router 分流
        +
Service 执行
        +
Schema 统一数据
        +
API 连接前后端
```

最终形成一个：

> **低成本、低延迟、模块化、可扩展的 Agent 学习系统。**

````

### 这个文件写完以后，你们三个人的分工就能真正对上了

```text
A：后端核心逻辑
│
├── workflows/
├── routers/
├── services/
├── agents/
└── schemas/

B：前端
│
└── frontend/

C：数据 / 测试 / 部署
│
├── data/
├── tests/
└── infra/

# Lingxi-claw Data Schema

> Version: v0.1.0  
> Status: Hackathon MVP  
> Backend: Go  
> API Data Format: JSON  
> Naming Convention: snake_case

---

# 1. 文档目的

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
````

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

# 2. 命名规范

## 2.1 API JSON

所有 JSON 字段使用：

```text
snake_case
```

正确：

```json
{
  "dataset_id": "ds_001",
  "file_count": 5,
  "created_at": "2026-08-22T10:00:00Z"
}
```

错误：

```json
{
  "datasetId": "ds_001",
  "fileCount": 5
}
```

---

## 2.2 ID 命名

不同对象使用不同前缀。

```text
dataset_id      ds_
file_id         file_
task_id         task_
question_id     q_
plan_id         plan_
session_id      practice_
homework_id     hw_
message_id      msg_
```

示例：

```text
ds_001
file_001
task_001
q_001
practice_001
hw_001
```

---

## 2.3 时间格式

所有时间使用 ISO 8601。

```text
2026-08-22T10:30:00Z
```

推荐 Go 后端使用：

```go
time.Time
```

JSON 输出统一为：

```text
RFC3339
```

---

# 3. 通用 API Response

所有 API 使用统一响应结构。

## 3.1 成功响应

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

---

## 3.2 失败响应

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求参数错误"
  }
}
```

---

# 4. Error

表示 API 错误信息。

```json
{
  "code": "INVALID_FILE",
  "message": "不支持该文件格式"
}
```

字段定义：

| 字段      | 类型     | 描述       |
| ------- | ------ | -------- |
| code    | string | 机器可识别错误码 |
| message | string | 用户可读错误信息 |

推荐错误码：

```text
INVALID_REQUEST
INVALID_FILE
FILE_TOO_LARGE
FILE_PARSE_ERROR
DATASET_NOT_FOUND
FILE_NOT_FOUND
TASK_NOT_FOUND
QUESTION_NOT_FOUND
HOMEWORK_NOT_FOUND
SESSION_NOT_FOUND
MODEL_UNAVAILABLE
INTERNAL_ERROR
```

---

# 5. Dataset

Dataset 表示一次完整的课程复习资料集合。

例如：

```text
概率论期末突击
│
├── 2023期末试卷.pdf
├── 2024期末试卷.docx
├── 2025期末试卷.pdf
└── 教师复习资料.pdf
```

对应一个 Dataset。

---

## Schema

```json
{
  "dataset_id": "ds_001",
  "name": "概率论期末突击",
  "course": "概率论",
  "file_count": 4,
  "status": "ready",
  "created_at": "2026-08-22T10:00:00Z",
  "updated_at": "2026-08-22T10:30:00Z"
}
```

字段：

| 字段         | 类型      | 描述            |
| ---------- | ------- | ------------- |
| dataset_id | string  | Dataset 唯一 ID |
| name       | string  | 复习任务名称        |
| course     | string  | 课程名称          |
| file_count | integer | 文件数量          |
| status     | string  | 当前状态          |
| created_at | string  | 创建时间          |
| updated_at | string  | 更新时间          |

---

## Dataset Status

```text
created
processing
ready
failed
```

状态流：

```text
created
    ↓
processing
    ↓
ready
```

如果失败：

```text
processing
    ↓
failed
```

---

# 6. StudyFile

StudyFile 表示 Dataset 中的一个文件。

例如：

```text
2024年概率论期末试卷.pdf
```

---

## Schema

```json
{
  "file_id": "file_001",
  "dataset_id": "ds_001",
  "file_name": "2024概率论期末试卷.pdf",
  "file_type": "pdf",
  "file_size": 2048000,
  "source_type": "exam_paper",
  "parse_status": "completed",
  "question_count": 20,
  "created_at": "2026-08-22T10:00:00Z"
}
```

字段：

| 字段             | 类型      | 描述            |
| -------------- | ------- | ------------- |
| file_id        | string  | 文件 ID         |
| dataset_id     | string  | 所属 Dataset    |
| file_name      | string  | 原始文件名         |
| file_type      | string  | 文件格式          |
| file_size      | integer | 文件大小，单位 bytes |
| source_type    | string  | 文件用途          |
| parse_status   | string  | 解析状态          |
| question_count | integer | 识别出的题目数量      |
| created_at     | string  | 上传时间          |

---

## File Type

```text
pdf
docx
image
```

图片包括：

```text
jpg
jpeg
png
webp
```

---

## Source Type

```text
exam_paper
review_material
homework
other
```

含义：

```text
exam_paper
    往年题 / 考试题

review_material
    教师复习资料 / 课程资料

homework
    日常作业

other
    其它学习资料
```

---

## Parse Status

```text
pending
processing
completed
failed
```

---

# 7. Task

Task 表示异步处理任务。

适用于：

```text
批量文件解析
OCR
题目提取
题型分类
历年题分析
复习计划生成
```

---

## Schema

```json
{
  "task_id": "task_001",
  "type": "file_processing",
  "status": "processing",
  "progress": 65,
  "message": "正在分析第 3 个文件",
  "processed_items": 3,
  "total_items": 5,
  "result": null,
  "created_at": "2026-08-22T10:00:00Z",
  "updated_at": "2026-08-22T10:01:00Z"
}
```

字段：

| 字段              | 类型          | 描述       |
| --------------- | ----------- | -------- |
| task_id         | string      | Task ID  |
| type            | string      | 任务类型     |
| status          | string      | 任务状态     |
| progress        | integer     | 进度，0-100 |
| message         | string      | 当前任务信息   |
| processed_items | integer     | 已处理数量    |
| total_items     | integer     | 总数量      |
| result          | object/null | 完成后的结果   |
| created_at      | string      | 创建时间     |
| updated_at      | string      | 更新时间     |

---

## Task Type

```text
file_processing
ocr
question_extraction
exam_analysis
study_plan
homework_analysis
```

---

## Task Status

```text
pending
processing
completed
failed
partial_success
```

---

# 8. Question

Question 是系统中的核心对象。

无论题目来自：

```text
往年题
日常作业
PDF
Word
图片
OCR
```

最终都尽量转换为统一的 Question。

---

## Schema

```json
{
  "question_id": "q_001",
  "dataset_id": "ds_001",
  "file_id": "file_001",
  "content": "计算以下二重积分……",
  "question_type": "calculation",
  "knowledge_points": [
    "二重积分",
    "积分区域"
  ],
  "difficulty": "medium",
  "source_year": 2024,
  "source_page": 2,
  "source_type": "exam_paper"
}
```

字段：

| 字段               | 类型           | 描述    |
| ---------------- | ------------ | ----- |
| question_id      | string       | 题目 ID |
| dataset_id       | string       | 所属资料集 |
| file_id          | string       | 来源文件  |
| content          | string       | 题目内容  |
| question_type    | string       | 题型    |
| knowledge_points | array        | 知识点   |
| difficulty       | string       | 难度    |
| source_year      | integer/null | 来源年份  |
| source_page      | integer/null | 来源页码  |
| source_type      | string       | 来源类型  |

---

## Question Type

MVP 第一版统一使用：

```text
choice
fill_blank
calculation
proof
application
short_answer
other
```

注意：

不同课程题型可能不同。

例如：

```text
高数
├── calculation
├── proof
└── application

概率论
├── calculation
├── choice
└── proof

程序设计
├── coding
├── debugging
└── theory
```

MVP 阶段可以先使用通用分类。

---

## Difficulty

```text
easy
medium
hard
```

---

# 9. KnowledgePoint

KnowledgePoint 表示知识点。

例如：

```text
概率论
├── 条件概率
├── 贝叶斯公式
├── 全概率公式
└── 随机变量
```

---

## Schema

```json
{
  "knowledge_point_id": "kp_001",
  "name": "贝叶斯公式",
  "course": "概率论",
  "description": "用于根据条件概率反推事件概率"
}
```

字段：

| 字段                 | 类型     | 描述     |
| ------------------ | ------ | ------ |
| knowledge_point_id | string | 知识点 ID |
| name               | string | 知识点名称  |
| course             | string | 所属课程   |
| description        | string | 简单描述   |

---

# 10. ExamAnalysis

ExamAnalysis 表示对 Dataset 中往年题的综合分析结果。

这是：

```text
期末突击
```

最核心的输出之一。

---

## Schema

```json
{
  "analysis_id": "analysis_001",
  "dataset_id": "ds_001",
  "course": "概率论",
  "total_files": 5,
  "total_questions": 120,
  "knowledge_points": [
    {
      "name": "贝叶斯公式",
      "frequency": 15,
      "importance": "high",
      "difficulty": "medium"
    },
    {
      "name": "随机变量",
      "frequency": 12,
      "importance": "high",
      "difficulty": "high"
    }
  ],
  "question_types": [
    {
      "name": "calculation",
      "count": 65,
      "percentage": 54.2
    },
    {
      "name": "proof",
      "count": 20,
      "percentage": 16.7
    }
  ],
  "summary": "贝叶斯公式、随机变量和数字特征是近几年出现频率较高的重点内容。",
  "created_at": "2026-08-22T11:00:00Z"
}
```

---

## Importance

```text
high
medium
low
```

---

# 11. StudyProfile

StudyProfile 表示生成复习计划时用户提供的学习情况。

---

## Schema

```json
{
  "exam_date": "2026-08-30",
  "daily_study_hours": 4,
  "current_level": "medium"
}
```

字段：

| 字段                | 类型     | 描述     |
| ----------------- | ------ | ------ |
| exam_date         | string | 考试日期   |
| daily_study_hours | number | 每日学习时间 |
| current_level     | string | 当前水平   |

---

## Current Level

```text
low
medium
high
```

---

# 12. StudyPlan

StudyPlan 表示根据：

```text
剩余时间
+
用户水平
+
每日学习时间
+
历年题分析
```

生成的个性化复习计划。

---

## Schema

```json
{
  "plan_id": "plan_001",
  "dataset_id": "ds_001",
  "days_remaining": 7,
  "daily_study_hours": 4,
  "daily_plan": [
    {
      "day": 1,
      "date": "2026-08-23",
      "focus": [
        "条件概率",
        "贝叶斯公式"
      ],
      "practice_count": 20,
      "estimated_hours": 4,
      "priority": "high"
    },
    {
      "day": 2,
      "date": "2026-08-24",
      "focus": [
        "随机变量"
      ],
      "practice_count": 15,
      "estimated_hours": 4,
      "priority": "high"
    }
  ],
  "created_at": "2026-08-22T11:30:00Z"
}
```

---

# 13. StudyPlanDay

StudyPlan 中的单日计划。

```json
{
  "day": 1,
  "date": "2026-08-23",
  "focus": [
    "贝叶斯公式"
  ],
  "practice_count": 20,
  "estimated_hours": 4,
  "priority": "high"
}
```

---

# 14. PracticeSession

PracticeSession 表示一次刷题会话。

---

## Schema

```json
{
  "session_id": "practice_001",
  "dataset_id": "ds_001",
  "knowledge_points": [
    "贝叶斯公式"
  ],
  "question_count": 5,
  "difficulty": "medium",
  "status": "active",
  "questions": [
    "q_001",
    "q_002",
    "q_003"
  ],
  "created_at": "2026-08-22T12:00:00Z"
}
```

---

## Session Status

```text
active
completed
abandoned
```

---

# 15. PracticeAnswer

表示用户对一道练习题的回答。

---

## Schema

```json
{
  "answer_id": "answer_001",
  "session_id": "practice_001",
  "question_id": "q_001",
  "user_answer": "用户输入的答案",
  "correct": false,
  "score": 0.6,
  "feedback": "积分区域判断正确，但是上下限设置错误。",
  "knowledge_gaps": [
    "积分区域转换"
  ],
  "submitted_at": "2026-08-22T12:10:00Z"
}
```

---

# 16. Homework

Homework 表示一次用户上传的日常作业。

---

## Schema

```json
{
  "homework_id": "hw_001",
  "course": "高等数学",
  "file_id": "file_010",
  "status": "analyzed",
  "question_count": 5,
  "created_at": "2026-08-22T14:00:00Z"
}
```

---

## Homework Status

```text
uploaded
processing
analyzed
failed
```

---

# 17. HomeworkQuestion

Homework 中识别出的题目。

```json
{
  "question_id": "q_hw_001",
  "homework_id": "hw_001",
  "content": "计算以下积分……",
  "question_type": "calculation",
  "knowledge_points": [
    "不定积分"
  ],
  "difficulty": "medium"
}
```

---

# 18. HintRequest

表示用户请求解题帮助。

注意：

Lingxi-claw 的作业辅助核心策略是：

> 默认不给完整答案，优先帮助用户自己完成。

---

## Schema

```json
{
  "question_id": "q_hw_001",
  "user_message": "我不知道第一步怎么做",
  "help_level": "direction"
}
```

---

## Help Level

```text
direction
method
step
full_solution
```

含义：

```text
direction
    只告诉思考方向

method
    告诉使用什么方法

step
    分步骤指导

full_solution
    给出完整解答
```

默认：

```text
direction
```

---

# 19. HomeworkFeedback

表示系统对用户答案的反馈。

---

## Schema

```json
{
  "question_id": "q_hw_001",
  "correct": false,
  "score": 0.6,
  "feedback": [
    {
      "step": 1,
      "correct": true,
      "message": "第一步判断正确"
    },
    {
      "step": 2,
      "correct": false,
      "message": "这里计算符号错误"
    }
  ],
  "final_answer": "正确答案",
  "summary": "你的方法是正确的，主要问题出现在第二步计算。"
}
```

注意：

`final_answer` 在提示阶段可以为空。

```json
{
  "final_answer": null
}
```

只有：

```text
用户主动提交答案
```

或者：

```text
用户明确请求完整答案
```

时才返回。

---

# 20. AgentSettings

AgentSettings 表示用户对 AI 学习助手行为的设置。

---

## Schema

```json
{
  "response_style": "detailed",
  "personality": "encouraging",
  "answer_policy": "hint_first"
}
```

---

## Response Style

```text
concise
normal
detailed
```

---

## Personality

```text
encouraging
strict
friendly
professional
```

---

## Answer Policy

```text
hint_first
balanced
direct_answer
```

默认：

```text
hint_first
```

---

# 21. ChatMessage

表示通用 Agent 的一次消息。

---

## Schema

```json
{
  "message_id": "msg_001",
  "role": "user",
  "content": "什么是贝叶斯公式？",
  "created_at": "2026-08-22T15:00:00Z"
}
```

AI 回复：

```json
{
  "message_id": "msg_002",
  "role": "assistant",
  "content": "贝叶斯公式用于……",
  "created_at": "2026-08-22T15:00:02Z"
}
```

---

## Role

```text
user
assistant
system
```

---

# 22. RoutingInfo

RoutingInfo 是 Lingxi-claw 的核心技术数据结构之一。

它记录：

```text
用户请求
    ↓
前置分类
    ↓
Workflow / Agent
    ↓
文件路由
    ↓
任务路由
    ↓
模型路由
```

最终系统是如何处理这个请求的。

---

## Schema

```json
{
  "input_type": "file_and_text",
  "intent": "final_sprint",
  "handler_type": "workflow",
  "handler": "final_sprint_workflow",
  "file_route": "ocr",
  "complexity": "medium",
  "model_route": "lightweight_model",
  "fallback": false
}
```

字段：

| 字段           | 类型          | 描述               |
| ------------ | ----------- | ---------------- |
| input_type   | string      | 输入类型             |
| intent       | string      | 用户意图             |
| handler_type | string      | Workflow 或 Agent |
| handler      | string      | 实际处理器            |
| file_route   | string/null | 文件处理路径           |
| complexity   | string      | 任务复杂度            |
| model_route  | string      | 模型路由结果           |
| fallback     | boolean     | 是否发生降级           |

---

## Input Type

```text
text
file
image
file_and_text
image_and_text
```

---

## Intent

```text
final_sprint
homework
general
agent_settings
unknown
```

---

## Handler Type

```text
workflow
agent
```

核心逻辑：

```text
高频明确需求
    ↓
Workflow

模糊开放需求
    ↓
Agent
```

---

## File Route

```text
docx_parser
pdf_parser
ocr
vision_model
none
```

---

## Complexity

```text
low
medium
high
```

---

## Model Route

```text
mock
rule_based
lightweight_model
standard_model
multimodal_model
```

---

# 23. RouterDecision

RouterDecision 用于记录前置轻量分类器的决策。

---

## Schema

```json
{
  "intent": "homework",
  "confidence": 0.92,
  "input_type": "image",
  "recommended_handler": "homework_workflow"
}
```

字段：

| 字段                  | 类型     | 描述    |
| ------------------- | ------ | ----- |
| intent              | string | 分类结果  |
| confidence          | number | 分类置信度 |
| input_type          | string | 输入类型  |
| recommended_handler | string | 推荐处理器 |

---

## Router Decision Flow

```text
用户输入
    ↓
轻量前置分类器
    ↓
RouterDecision
    │
    ├── 高置信度 + 高频场景
    │       ↓
    │    Workflow
    │
    ├── 中等置信度
    │       ↓
    │    Router 二次判断
    │
    └── 低置信度 / 开放问题
            ↓
          Agent
```

---

# 24. TokenOptimization

用于记录 Token Pruning 和成本优化效果。

主要用于：

```text
黑客松 Demo
实时成本看板
系统性能分析
```

---

## Schema

```json
{
  "original_input_tokens": 1200,
  "pruned_input_tokens": 450,
  "token_saved": 750,
  "token_saved_percent": 62.5
}
```

字段：

| 字段                    | 类型      | 描述        |
| --------------------- | ------- | --------- |
| original_input_tokens | integer | 原始 Token  |
| pruned_input_tokens   | integer | 剪枝后 Token |
| token_saved           | integer | 节省 Token  |
| token_saved_percent   | number  | 节省比例      |

---

# 25. ModelMetrics

用于记录一次模型调用的性能数据。

---

## Schema

```json
{
  "model_name": "lightweight_model",
  "input_tokens": 450,
  "output_tokens": 300,
  "latency_ms": 1200,
  "estimated_cost": 0.008
}
```

字段：

| 字段             | 类型      | 描述       |
| -------------- | ------- | -------- |
| model_name     | string  | 模型名称     |
| input_tokens   | integer | 输入 Token |
| output_tokens  | integer | 输出 Token |
| latency_ms     | integer | 延迟       |
| estimated_cost | number  | 预估成本     |

---

# 26. RequestMetrics

RequestMetrics 用于完整记录一次请求的优化效果。

---

## Schema

```json
{
  "request_id": "req_001",
  "routing": {
    "intent": "final_sprint",
    "handler": "final_sprint_workflow",
    "model_route": "lightweight_model"
  },
  "token_optimization": {
    "original_input_tokens": 1200,
    "pruned_input_tokens": 450,
    "token_saved_percent": 62.5
  },
  "model_metrics": {
    "latency_ms": 1200,
    "estimated_cost": 0.008
  }
}
```

---

# 27. DemoMetrics

DemoMetrics 用于前端实时成本与延迟对比看板。

---

## Schema

```json
{
  "baseline": {
    "input_tokens": 2450,
    "output_tokens": 800,
    "latency_ms": 3200,
    "estimated_cost": 0.032
  },
  "lingxi_claw": {
    "input_tokens": 650,
    "output_tokens": 500,
    "latency_ms": 1200,
    "estimated_cost": 0.008
  },
  "improvement": {
    "token_saved_percent": 73.5,
    "latency_saved_percent": 62.5,
    "cost_saved_percent": 75.0
  }
}
```

---

# 28. 数据对象关系

系统核心对象关系：

```text
Dataset
    │
    ├── StudyFile
    │       │
    │       └── Question
    │
    ├── ExamAnalysis
    │       │
    │       └── KnowledgePoint
    │
    └── StudyPlan
            │
            └── PracticeSession
                    │
                    └── PracticeAnswer
```

作业模块：

```text
Homework
    │
    ├── StudyFile
    │
    └── HomeworkQuestion
            │
            ├── HintRequest
            │
            └── HomeworkFeedback
```

路由模块：

```text
User Request
    │
    ├── RouterDecision
    │
    ├── RoutingInfo
    │
    ├── TokenOptimization
    │
    └── ModelMetrics
```

---

# 29. 前后端字段统一规则

禁止出现：

```text
前端：

datasetId
fileName
createdAt
```

后端：

```text
dataset_id
file_name
created_at
```

正确做法：

```text
API JSON
    ↓
全部 snake_case
```

统一：

```text
dataset_id
file_id
task_id
question_id
created_at
updated_at
```

---

# 30. Go Struct 示例

Go 后端推荐：

```go
type Dataset struct {
    DatasetID string `json:"dataset_id"`
    Name      string `json:"name"`
    Course    string `json:"course"`
    FileCount int    `json:"file_count"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}
```

RoutingInfo：

```go
type RoutingInfo struct {
    InputType   string  `json:"input_type"`
    Intent      string  `json:"intent"`
    HandlerType string  `json:"handler_type"`
    Handler     string  `json:"handler"`
    FileRoute   *string `json:"file_route"`
    Complexity  string  `json:"complexity"`
    ModelRoute  string  `json:"model_route"`
    Fallback    bool    `json:"fallback"`
}
```

---

# 31. Mock Data 规则

Mock 数据必须遵守 DATA_SCHEMA.md。

例如 Mock Dataset：

```json
{
  "dataset_id": "ds_mock_001",
  "name": "概率论期末突击",
  "course": "概率论",
  "file_count": 3,
  "status": "ready",
  "created_at": "2026-08-22T10:00:00Z",
  "updated_at": "2026-08-22T10:00:00Z"
}
```

禁止：

```json
{
  "id": 1,
  "title": "概率论",
  "files": 3
}
```

因为 Mock 数据必须与真实后端返回格式兼容。

---

# 32. Schema 修改规则

如果新增字段：

```text
修改 DATA_SCHEMA.md
    ↓
修改 API.md
    ↓
后端修改 Go Struct
    ↓
前端修改 TypeScript Type
    ↓
更新 Mock Data
    ↓
更新 Test
```

禁止：

```text
直接修改后端 Struct
```

而不更新文档。

---

# 33. MVP 必须实现的数据对象

## P0

```text
Dataset
StudyFile
Task
Question
ExamAnalysis
StudyProfile
StudyPlan
Homework
HomeworkQuestion
HomeworkFeedback
ChatMessage
RoutingInfo
RouterDecision
```

---

## P1

```text
KnowledgePoint
PracticeSession
PracticeAnswer
AgentSettings
TokenOptimization
ModelMetrics
```

---

## P2 / Demo

```text
RequestMetrics
DemoMetrics
```

---

# 34. 最终数据流

```text
用户输入
    ↓
User Request
    ↓
RouterDecision
    ↓
RoutingInfo
    ↓
┌─────────────────────┐
│                     │
Workflow            Agent
│                     │
↓                     ↓
Dataset             ChatMessage
StudyFile
Question
Homework
│
↓
ExamAnalysis
│
↓
StudyPlan
│
↓
PracticeSession
│
↓
PracticeAnswer
│
↓
Response
```

---

# 35. 核心原则

Lingxi-claw 的数据设计遵循：

> **外部接口统一、内部模块解耦、Mock 与真实数据兼容。**

最终：

```text
DATA_SCHEMA.md
        ↓
定义数据是什么

API.md
        ↓
定义数据怎么传输

ARCHITECTURE.md
        ↓
定义数据经过哪些模块

WORKFLOW.md
        ↓
定义数据如何流动
```

这四个文档共同构成 Lingxi-claw 的核心技术规范。

````

我特别建议你们**暂时不要让队友一次性把这里所有对象都实现成数据库表**。这是 `Data Schema`，不等于数据库 Schema。

例如黑客松 MVP 可以先这样：

```text
Dataset / Task / Question
        ↓
Go 内存 / JSON 文件 / SQLite

复杂 AI 分析结果
        ↓
直接 JSON

Router Metrics
        ↓
内存记录 / 日志
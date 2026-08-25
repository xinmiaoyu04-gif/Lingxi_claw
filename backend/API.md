1. 文档目的
本文档定义 Lingxi-claw 前端与后端之间的 API 接口。

所有团队成员必须遵守：

Frontend
    ↓ HTTP Request
Backend API
    ↓
Workflow / Router / Agent
    ↓
Service / Tool / Model
    ↓
HTTP Response
    ↓
Frontend
本文档是：

前端、后端、测试之间的接口合同。

前端开发者不应该猜测后端接口。

后端开发者不应该随意修改 Response 字段。

测试人员根据本文档编写测试。

如果接口需要修改：

修改 API.md
    ↓
团队确认
    ↓
修改代码
而不是直接修改代码。

2. 基础约定
2.1 Base URL
开发环境：

http://localhost:8080
所有 API 使用：

/api/v1
因此完整接口示例：

POST http://localhost:8080/api/v1/chat
3. Content-Type
普通 JSON 请求：

Content-Type: application/json
文件上传：

Content-Type: multipart/form-data
4. 统一 Response 格式
所有接口尽量使用统一返回结构。

成功
{
  "success": true,
  "data": {},
  "error": null
}
失败
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述"
  }
}
5. 通用错误码
HTTP Status	Code	含义
400	INVALID_REQUEST	请求参数错误
400	INVALID_FILE	不支持的文件
404	DATASET_NOT_FOUND	数据集不存在
404	TASK_NOT_FOUND	任务不存在
404	HOMEWORK_NOT_FOUND	作业不存在
422	FILE_PARSE_ERROR	文件解析失败
422	QUESTION_PARSE_ERROR	题目识别失败
500	INTERNAL_ERROR	服务器内部错误
503	MODEL_UNAVAILABLE	模型服务不可用
6. 核心数据对象
6.1 Dataset
Dataset 表示用户一次期末复习任务中的资料集合。

例如：

高等数学期末复习
├── 2023年期末试卷.pdf
├── 2024年期末试卷.docx
├── 2025年期末试卷.pdf
└── 教师复习资料.pdf
对应一个 Dataset。

结构：

{
  "dataset_id": "ds_xxxxx",
  "name": "高等数学期末复习",
  "course": "高等数学",
  "file_count": 4,
  "status": "ready",
  "created_at": "2026-08-22T10:00:00Z"
}
6.2 Task
Task 表示后台异步任务。

例如：

OCR
批量文件解析
题目分类
考点分析
结构：

{
  "task_id": "task_xxxxx",
  "type": "file_processing",
  "status": "processing",
  "progress": 65,
  "message": "正在分析第 3 个文件"
}
Task 状态：

pending
processing
completed
failed
partial_success
7. API 总览
📚 期末突击
POST   /api/v1/final-sprint/datasets
POST   /api/v1/final-sprint/datasets/{dataset_id}/files
GET    /api/v1/tasks/{task_id}

POST   /api/v1/final-sprint/datasets/{dataset_id}/analyze
GET    /api/v1/final-sprint/datasets/{dataset_id}/analysis

POST   /api/v1/final-sprint/datasets/{dataset_id}/plan
GET    /api/v1/final-sprint/datasets/{dataset_id}/plan

POST   /api/v1/final-sprint/datasets/{dataset_id}/practice
POST   /api/v1/practice/{session_id}/answer
📝 日常作业辅助
POST   /api/v1/homework
POST   /api/v1/homework/{homework_id}/hint
POST   /api/v1/homework/{homework_id}/answer
❓ 其它问题
POST   /api/v1/chat
🤖 Agent 设定
GET    /api/v1/settings/agent
PUT    /api/v1/settings/agent
8. 期末突击 API
8.1 创建 Dataset
创建一个新的期末复习任务。

Request
POST /api/v1/final-sprint/datasets
Content-Type: application/json
Body：

{
  "name": "高等数学期末突击",
  "course": "高等数学"
}
字段：

字段	类型	必填	描述
name	string	是	Dataset 名称
course	string	是	课程名称
Response
{
  "success": true,
  "data": {
    "dataset_id": "ds_001",
    "name": "高等数学期末突击",
    "course": "高等数学",
    "file_count": 0,
    "status": "created"
  },
  "error": null
}
8.2 批量上传文件
用户可以一次上传多个文件。

支持：

PDF
DOCX
图片
扫描版 PDF
Request
POST /api/v1/final-sprint/datasets/{dataset_id}/files
Content-Type: multipart/form-data
FormData：

files[]: 2023期末.pdf
files[]: 2024期末.docx
files[]: 2025期末.pdf
files[]: 教师复习资料.pdf
Response
{
  "success": true,
  "data": {
    "dataset_id": "ds_001",
    "task_id": "task_001",
    "total_files": 4,
    "status": "processing"
  },
  "error": null
}
后端逻辑
API
 ↓
FinalSprintWorkflow
 ↓
File Router
 ↓
判断文件类型
 ↓
┌───────────┬──────────────┬───────────┐
↓           ↓              ↓
DOCX        Text PDF       Scan PDF
↓           ↓              ↓
Parser      Parser         OCR
 ↓
统一文本 / 图片内容
 ↓
Question Service
 ↓
Question Bank
8.3 查询异步任务
用于前端获取文件处理进度。

Request
GET /api/v1/tasks/{task_id}
Response
{
  "success": true,
  "data": {
    "task_id": "task_001",
    "type": "file_processing",
    "status": "processing",
    "progress": 75,
    "processed_files": 3,
    "total_files": 4,
    "message": "正在分析第 4 个文件"
  },
  "error": null
}
处理完成：

{
  "success": true,
  "data": {
    "task_id": "task_001",
    "status": "completed",
    "progress": 100,
    "processed_files": 4,
    "total_files": 4
  },
  "error": null
}
部分成功：

{
  "success": true,
  "data": {
    "task_id": "task_001",
    "status": "partial_success",
    "processed_files": 3,
    "total_files": 4,
    "failed_files": [
      {
        "name": "损坏文件.pdf",
        "reason": "文件无法解析"
      }
    ]
  },
  "error": null
}
8.4 分析历年题
启动历年题分析。

系统分析：

高频考点
高频题型
题目出现次数
难度
重点内容
Request
POST /api/v1/final-sprint/datasets/{dataset_id}/analyze
Body：

{}
Response
{
  "success": true,
  "data": {
    "task_id": "task_002",
    "dataset_id": "ds_001",
    "status": "processing"
  },
  "error": null
}
8.5 获取分析结果
Request
GET /api/v1/final-sprint/datasets/{dataset_id}/analysis
Response
{
  "success": true,
  "data": {
    "dataset_id": "ds_001",
    "course": "高等数学",
    "total_questions": 120,
    "knowledge_points": [
      {
        "name": "二重积分",
        "frequency": 15,
        "importance": "high",
        "difficulty": "medium"
      },
      {
        "name": "无穷级数",
        "frequency": 12,
        "importance": "high",
        "difficulty": "high"
      }
    ],
    "question_types": [
      {
        "name": "计算题",
        "count": 65
      },
      {
        "name": "证明题",
        "count": 20
      }
    ]
  },
  "error": null
}
8.6 生成复习计划
根据用户剩余时间生成个性化复习计划。

Request
POST /api/v1/final-sprint/datasets/{dataset_id}/plan
Content-Type: application/json
Body：

{
  "exam_date": "2026-08-30",
  "daily_study_hours": 4,
  "current_level": "medium"
}
字段：

字段	类型	必填	描述
exam_date	string	是	考试日期
daily_study_hours	number	是	每日复习时间
current_level	string	否	用户当前水平
current_level：

low
medium
high
Response
{
  "success": true,
  "data": {
    "task_id": "task_003",
    "status": "processing"
  },
  "error": null
}
8.7 获取复习计划
Request
GET /api/v1/final-sprint/datasets/{dataset_id}/plan
Response
{
  "success": true,
  "data": {
    "dataset_id": "ds_001",
    "days_remaining": 7,
    "daily_plan": [
      {
        "day": 1,
        "focus": [
          "二重积分",
          "三重积分"
        ],
        "practice_count": 20,
        "estimated_hours": 4
      },
      {
        "day": 2,
        "focus": [
          "无穷级数"
        ],
        "practice_count": 15,
        "estimated_hours": 4
      }
    ]
  },
  "error": null
}
8.8 开始刷题
根据当前学习计划选择题目。

Request
POST /api/v1/final-sprint/datasets/{dataset_id}/practice
Content-Type: application/json
Body：

{
  "knowledge_points": [
    "二重积分"
  ],
  "question_count": 5,
  "difficulty": "medium"
}
Response
{
  "success": true,
  "data": {
    "session_id": "practice_001",
    "questions": [
      {
        "question_id": "q_001",
        "content": "计算以下二重积分……",
        "knowledge_point": "二重积分",
        "difficulty": "medium"
      }
    ]
  },
  "error": null
}
8.9 提交刷题答案
Request
POST /api/v1/practice/{session_id}/answer
Content-Type: application/json
Body：

{
  "question_id": "q_001",
  "user_answer": "用户答案"
}
Response
{
  "success": true,
  "data": {
    "question_id": "q_001",
    "correct": false,
    "feedback": "你的积分区域判断正确，但是积分上下限设置有问题。",
    "knowledge_gap": [
      "积分区域转换"
    ]
  },
  "error": null
}
9. 日常作业辅助 API
9.1 上传并分析作业
用户上传：

图片
PDF
DOCX
Request
POST /api/v1/homework
Content-Type: multipart/form-data
FormData：

file: homework.jpg
course: 高等数学
Response
{
  "success": true,
  "data": {
    "homework_id": "hw_001",
    "task_id": "task_010",
    "status": "processing"
  },
  "error": null
}
前端随后通过：

GET /api/v1/tasks/{task_id}
查询处理状态。

9.2 请求解题提示
默认不直接给最终答案。

Request
POST /api/v1/homework/{homework_id}/hint
Content-Type: application/json
Body：

{
  "question_id": "q_001",
  "user_message": "我不知道从哪里开始"
}
Response
{
  "success": true,
  "data": {
    "question_id": "q_001",
    "help_level": "direction",
    "response": "先判断这个题属于什么类型，再考虑是否需要使用换元法。"
  },
  "error": null
}
9.3 提交答案
用户自己完成题目后提交。

Request
POST /api/v1/homework/{homework_id}/answer
Content-Type: application/json
Body：

{
  "question_id": "q_001",
  "user_answer": "我的计算过程和答案"
}
Response
{
  "success": true,
  "data": {
    "question_id": "q_001",
    "correct": false,
    "score": 0.6,
    "feedback": [
      {
        "step": 1,
        "correct": true,
        "message": "第一步正确"
      },
      {
        "step": 2,
        "correct": false,
        "message": "这里的积分上下限有误"
      }
    ],
    "final_answer": "正确答案"
  },
  "error": null
}
10. 通用 Agent API
用于：

其它问题
开放式学习问题
复杂问题
10.1 Chat
Request
POST /api/v1/chat
Content-Type: application/json
Body：

{
  "message": "帮我解释一下什么是贝叶斯公式",
  "course": "概率论",
  "agent_settings": {
    "response_style": "detailed"
  }
}
Response
{
  "success": true,
  "data": {
    "message": "贝叶斯公式用于……",
    "route": {
      "mode": "general",
      "complexity": "medium",
      "handler": "general_agent"
    }
  },
  "error": null
}
注意：

route 主要用于：

前端 Debug
成本看板
黑客松 Demo 展示
可以显示：

当前路由：

General Agent
    ↓
Medium Complexity
    ↓
Standard Model
11. Agent 设置 API
11.1 获取 Agent 设置
Request
GET /api/v1/settings/agent
Response
{
  "success": true,
  "data": {
    "response_style": "detailed",
    "personality": "encouraging",
    "answer_policy": "hint_first"
  },
  "error": null
}
11.2 更新 Agent 设置
Request
PUT /api/v1/settings/agent
Content-Type: application/json
Body：

{
  "response_style": "concise",
  "personality": "strict",
  "answer_policy": "hint_first"
}
12. 路由信息标准
为了展示项目的核心技术，关键接口可以返回路由信息。

统一格式：

{
  "routing": {
    "workflow": "final_sprint",
    "file_route": "ocr",
    "question_type": "calculation",
    "model_route": "standard_model"
  }
}
复杂任务：

{
  "routing": {
    "mode": "agent",
    "agent": "general_agent",
    "complexity": "high",
    "tool": "vision_model",
    "model_route": "multimodal_model"
  }
}
13. Demo 成本与延迟数据
为了黑客松演示，可以增加一个模拟接口。

GET /api/v1/demo/metrics
Response：

{
  "success": true,
  "data": {
    "baseline": {
      "model": "general_large_model",
      "input_tokens": 2450,
      "output_tokens": 800,
      "latency_ms": 3200,
      "estimated_cost": 0.032
    },
    "lingxi_claw": {
      "pruned_tokens": 1800,
      "input_tokens": 650,
      "output_tokens": 500,
      "latency_ms": 1200,
      "estimated_cost": 0.008,
      "route": "lightweight_model"
    },
    "improvement": {
      "token_saved_percent": 73.5,
      "latency_saved_percent": 62.5,
      "cost_saved_percent": 75.0
    }
  },
  "error": null
}
前端可以展示：

┌──────────────────────────────┐
│ 通用大模型                    │
│ Token: 2450                 │
│ 延迟: 3.2s                  │
│ 成本: $0.032                │
└──────────────────────────────┘

              VS

┌──────────────────────────────┐
│ Lingxi-claw                  │
│ Token: 650 ↓73.5%           │
│ 延迟: 1.2s ↓62.5%           │
│ 成本: $0.008 ↓75%           │
└──────────────────────────────┘
14. Mock Mode
在后端真实 AI 能力尚未完成时，必须支持 Mock 数据。

推荐：

APP_MODE=mock
Mock 模式：

API Request
    ↓
Mock Service
    ↓
固定格式 Response
真实模式：

API Request
    ↓
Workflow / Router
    ↓
Real Service / Model
    ↓
Response
保证：

Mock Response 与真实 Response 字段必须完全兼容。

这样前端不需要等待真实后端完成。

15. Go 后端实现建议
推荐：

backend/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── handler/
│   ├── service/
│   ├── router/
│   ├── workflow/
│   ├── agent/
│   ├── model/
│   └── repository/
│
└── pkg/
职责：

handler
    ↓
HTTP Request / Response

service
    ↓
具体业务逻辑

router
    ↓
任务路由

workflow
    ↓
业务流程

agent
    ↓
复杂开放任务

model
    ↓
数据结构
16. Vibe Coding 前端约束
前端 AI 生成代码时，必须遵守：

1. 只调用 API.md 中定义的接口
2. 不自己编造接口地址
3. 不自己编造字段
4. 不直接调用模型 API
5. API Base URL 必须可配置
6. 必须支持 Loading 状态
7. 必须支持 Error 状态
8. 必须支持 Mock Mode
推荐环境变量：

VITE_API_BASE_URL=http://localhost:8080
VITE_APP_MODE=mock
17. Vibe Coding 测试约束
测试 AI 生成测试时：

所有 API 测试以 API.md 为准。
测试至少覆盖：

正常请求
空参数
错误参数
文件不存在
文件格式错误
Task 不存在
Dataset 不存在
部分文件解析失败
模型不可用
18. 接口修改规则
禁止：

后端修改字段
    ↓
不通知前端
禁止：

前端自己猜字段
正确流程：

需要修改接口
        ↓
修改 API.md
        ↓
确认 Request
        ↓
确认 Response
        ↓
更新 DATA_SCHEMA.md
        ↓
前后端同步修改
        ↓
测试
19. MVP 优先级
P0：必须完成
POST /api/v1/final-sprint/datasets

POST /api/v1/final-sprint/datasets/{dataset_id}/files

GET /api/v1/tasks/{task_id}

POST /api/v1/final-sprint/datasets/{dataset_id}/analyze

GET /api/v1/final-sprint/datasets/{dataset_id}/analysis

POST /api/v1/final-sprint/datasets/{dataset_id}/plan

POST /api/v1/homework

POST /api/v1/homework/{homework_id}/hint

POST /api/v1/homework/{homework_id}/answer

POST /api/v1/chat
P1：建议完成
GET /api/v1/final-sprint/datasets/{dataset_id}/plan

POST /api/v1/final-sprint/datasets/{dataset_id}/practice

POST /api/v1/practice/{session_id}/answer

GET /api/v1/settings/agent

PUT /api/v1/settings/agent
P2：演示增强
GET /api/v1/demo/metrics
20. 最终接口调用关系
Frontend
    │
    │ HTTP API
    ↓
Go Backend
    │
    ├── Handler
    │
    ├── Workflow
    │
    ├── Router
    │
    ├── Agent
    │
    └── Service
            │
            ├── File Parser
            ├── OCR
            ├── Knowledge Base
            ├── Model API
            └── Mock Service
                    │
                    ↓
                 Response
                    │
                    ↓
                 Frontend
21. 团队最重要的约定
API.md 是前端、后端和测试之间的唯一接口标准。

对于 API：

前端以 API.md 为准
后端以 API.md 为准
测试以 API.md 为准
Vibe Coding Prompt 以 API.md 为上下文
任何代码与 API.md 不一致时：

优先修正文档或统一修改代码，不能各自维护不同版本。
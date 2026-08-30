# Study with Lingxi-claw - Workflow 定义

# 1. 文档目的

本文档定义 Lingxi-claw 的产品工作流（Workflow）与用户学习任务处理流程。

Workflow 用于描述：

* 用户如何进入学习任务
* 系统如何理解当前任务
* 不同学习任务对应什么处理流程
* Workflow 与 Agent 如何协作
* 课程资料、作业、错题和学习分析之间如何形成闭环
* 当前 V02 阶段哪些 Workflow 只展示页面，哪些属于未来真实 AI 能力

本文档主要用于：

1. 产品设计
2. 前后端开发
3. AI Agent 开发
4. 测试与验收
5. Vibe Coding AI 的项目上下文理解

---

# 2. Workflow 总体原则

Lingxi-claw 的 Workflow 不作为用户需要理解的独立产品功能。

用户看到的是：

```text
学习画像
    ↓
Home
    ↓
我的课程
    ↓
具体课程
    ↓
Course Space
    ↓
具体学习功能
```

系统内部才进入对应 Workflow：

```text
用户学习任务
    ↓
任务识别
    ↓
Workflow / Agent
    ↓
对应能力
    ↓
学习结果
```

因此：

> **产品层围绕“课程”组织，AI 层围绕“学习任务”组织。**

---

# 3. 产品整体工作流

Lingxi-claw 的整体用户工作流：

```text
首次进入
    ↓
建立学习画像
    ↓
Home
    ↓
查看今日学习 / 最近作业 / 学习进度 / AI 建议
    ↓
我的课程
    ↓
选择具体课程
    ↓
Course Space
    ↓
┌──────────────┬──────────────┬──────────────┬──────────────┐
│     概览     │ 课程资料知识库 │  AI 作业辅导  │   错题记录    │
└──────────────┴──────────────┴──────────────┴──────────────┘
                                      │
                                      ↓
                                学习分析
                                      │
                                      ↓
                                个性化建议
```

其中：

* **学习画像**负责提供用户背景与偏好
* **Home**负责聚合学习状态
* **Courses**负责课程组织
* **Course Space**负责具体课程学习
* **Analytics**负责跨课程分析
* **Settings**负责用户偏好管理

---

# 4. 学习画像 Workflow

## 4.1 首次进入流程

用户第一次使用 Lingxi-claw：

```text
用户首次进入
    ↓
Onboarding
    ↓
填写学习画像
    ↓
保存学习偏好
    ↓
进入 Home
```

---

## 4.2 学习画像信息

学习画像包括：

```text
专业
学习习惯
作业习惯
AI 使用习惯
学习目标
```

例如：

```text
专业：
软件工程

学习习惯：
喜欢先理解概念，再刷题

作业习惯：
遇到不会的题目希望获得提示

AI 使用习惯：
概念解释 + 解题提示

学习目标：
提高课程成绩
```

---

## 4.3 学习画像的作用

学习画像未来可以作为 AI Workflow 的上下文。

例如：

```text
学习画像
    ↓
用户偏好
    ↓
AI 作业辅导
    ↓
选择更合适的教学方式
```

当前 V02 阶段：

> 学习画像只实现页面和 Mock 数据，不实现真实个性化算法。

---

# 5. Home Workflow

Home 是学习状态聚合中心。

用户进入 Home 后：

```text
进入 Home
    ↓
读取用户学习状态
    ↓
展示：
├── 今日学习
├── 我的课程
├── 最近作业
├── 学习进度
└── AI 学习建议
```

当前 V02 阶段可以使用 Mock 数据。

未来：

```text
课程数据
+
作业数据
+
错题数据
+
学习行为
    ↓
学习状态计算
    ↓
AI 学习建议
    ↓
Home
```

---

# 6. Courses Workflow

Courses 用于管理用户课程。

```text
进入 Courses
    ↓
获取课程列表
    ↓
展示课程
    ↓
用户选择课程
    ↓
进入 Course Space
```

例如：

```text
Courses
    ↓
高等数学
    ↓
/courses/math
    ↓
高等数学 Course Space
```

---

## 6.1 添加课程

未来课程添加流程：

```text
点击「＋添加课程」
    ↓
填写课程信息
    ↓
创建课程
    ↓
课程进入 Courses
    ↓
生成对应 Course Space
```

课程可以包含：

* 课程名称
* 教师
* 学期
* 学习目标
* 课程资料

当前 V02 阶段：

> 只实现 UI 和页面跳转，可以使用 Mock 数据。

---

# 7. Course Space 总 Workflow

Course Space 是 Lingxi-claw 的核心工作空间。

用户：

```text
Courses
    ↓
选择课程
    ↓
Course Space
```

进入课程空间后，可以在以下模块之间切换：

```text
概览
课程资料知识库
AI 作业辅导
错题记录
学习分析
```

整体 Workflow：

```text
                 Course Space
                      │
       ┌──────────────┼──────────────┐
       ↓              ↓              ↓
     资料            作业            错题
       │              │              │
       └──────────────┼──────────────┘
                      ↓
                  学习分析
                      ↓
                 学习建议
```

---

# 8. Course Space - 概览 Workflow

概览用于展示当前课程的学习状态。

```text
进入课程空间
    ↓
概览
    ↓
查看课程状态
```

可以展示：

* 课程掌握度
* 最近学习内容
* 待完成作业
* 错题数量
* AI 学习建议

示例：

```text
高等数学

掌握度：72%

近期学习：
二重积分

待完成作业：
2 项

错题：
8 道

AI 建议：
优先复习二重积分换元法
```

当前阶段：

> 使用 Mock 数据。

---

# 9. Course Space - 课程资料知识库 Workflow

课程资料知识库用于管理当前课程的学习资料。

---

## 9.1 资料上传 Workflow

未来真实流程：

```text
用户点击上传
    ↓
选择文件
    ↓
上传文件
    ↓
识别文件类型
    ↓
文件解析
    ↓
提取课程内容
    ↓
建立课程知识内容
    ↓
进入课程资料知识库
```

支持的资料类型包括：

```text
PDF
Word
PPT
图片
扫描 PDF
课程笔记
历年试题
```

---

## 9.2 文件处理 Workflow

未来可以根据文件类型选择不同处理方式：

```text
输入文件
    ↓
文件类型识别
    │
    ├── DOCX
    │      ↓
    │   文本解析
    │
    ├── 普通 PDF
    │      ↓
    │   PDF 文本提取
    │
    ├── 扫描 PDF
    │      ↓
    │   OCR
    │
    └── 图片
           ↓
       OCR / Vision
```

然后：

```text
解析结果
    ↓
内容结构化
    ↓
课程资料
```

当前 V02 阶段：

> 不实现真实文件解析、OCR 和知识库构建。

页面只展示：

* 文件列表
* 文件名称
* 文件类型
* 文件状态
* 上传入口

---

# 10. Course Space - AI 作业辅导 Workflow

AI 作业辅导是课程空间中的核心学习 Workflow。

核心目标：

> **帮助学生完成学习，而不是简单替学生完成作业。**

---

## 10.1 基础流程

未来完整流程：

```text
进入 AI 作业辅导
    ↓
选择 / 上传作业
    ↓
识别题目
    ↓
识别知识点
    ↓
判断题目类型
    ↓
选择对应教学策略
    ↓
提供解题帮助
    ↓
用户尝试
    ↓
提交答案 / 解题过程
    ↓
检查
    ↓
反馈
    ↓
记录学习结果
```

---

## 10.2 分层提示 Workflow

作业辅导采用逐级帮助机制。

```text
用户表示不会
    ↓
Level 1：方向提示
    ↓
用户继续尝试
    ↓
Level 2：方法提示
    ↓
用户继续尝试
    ↓
Level 3：关键步骤
    ↓
用户仍无法完成
    ↓
Level 4：完整解析
```

例如：

### Level 1

```text
先观察这个题目是否可以使用换元法。
```

### Level 2

```text
可以尝试令 t = x²。
```

### Level 3

```text
完成换元后，积分可以转换为……
```

### Level 4

```text
提供完整推导过程。
```

---

## 10.3 作业结果记录

未来作业辅导结束后：

```text
作业辅导
    ↓
用户结果
    ↓
记录：
├── 是否完成
├── 是否正确
├── 错误知识点
├── 使用提示等级
└── 错题
```

这些数据可以进一步进入：

```text
错题记录
学习分析
AI 学习建议
```

当前 V02 阶段：

> 只展示作业列表、作业状态和 AI 辅导入口，不实现真实 AI 作业辅导。

---

# 11. Course Space - 错题记录 Workflow

错题记录用于形成课程学习反馈。

未来：

```text
AI 作业辅导
    ↓
发现错误
    ↓
判断是否形成错题
    ↓
记录错题
    ↓
进入错题记录
```

用户查看错题：

```text
错题记录
    ↓
选择错题
    ↓
查看题目
    ↓
查看错误情况
    ↓
重新练习
```

未来可以进一步形成：

```text
错题
 ↓
知识点
 ↓
错误原因
 ↓
错误次数
 ↓
掌握状态
 ↓
复习建议
```

当前 V02 阶段：

> 使用 Mock 数据展示错题列表和状态。

---

# 12. Course Space - 学习分析 Workflow

课程学习分析负责分析单门课程。

数据来源未来包括：

```text
课程资料
+
作业
+
错题
+
学习行为
+
学习任务
```

进入：

```text
Course Space
    ↓
学习分析
    ↓
分析当前课程
```

展示：

```text
课程掌握度
知识点掌握情况
错题分布
作业表现
学习趋势
```

未来可以形成：

```text
学习数据
    ↓
分析
    ↓
发现薄弱知识点
    ↓
生成建议
    ↓
返回 Course Space
```

当前 V02 阶段：

> 使用 Mock 数据和静态图表。

---

# 13. 全局 Analytics Workflow

全局 Analytics 与 Course Space 中的学习分析不同。

```text
Analytics
    ↓
跨课程分析
```

而：

```text
Course Space
    ↓
单课程分析
```

---

## 13.1 数据流

未来：

```text
高等数学数据 ─┐
大学物理数据 ─┼→ 全局学习分析
线性代数数据 ─┘
                     ↓
             用户整体学习状态
                     ↓
        ┌────────────┼────────────┐
        ↓            ↓            ↓
      掌握度        薄弱点       学习趋势
```

Analytics 可以展示：

* 整体学习情况
* 各课程掌握度
* 知识薄弱点
* 作业表现
* 学习趋势

当前 V02 阶段：

> 使用 Mock 数据。

---

# 14. AI 学习建议 Workflow

AI 学习建议是连接学习数据和用户行动的能力。

未来：

```text
学习画像
+
课程数据
+
作业数据
+
错题数据
+
学习行为
    ↓
学习状态分析
    ↓
识别问题
    ↓
生成建议
    ↓
Home / Course Space / Analytics
```

例如：

```text
发现：
用户最近在二重积分相关题目中错误较多。

        ↓

生成：

建议优先复习：
「二重积分换元法」

        ↓

用户点击建议

        ↓

进入：

高等数学
    ↓
AI 作业辅导 / 课程资料
```

当前 V02 阶段：

> 只展示 Mock AI 建议。

---

# 15. General Agent Workflow

General Agent 用于处理无法被固定 Workflow 覆盖的开放式学习需求。

它不是一级导航入口，而是系统内部能力。

例如：

```text
用户提出开放问题
    ↓
任务理解
    ↓
判断是否存在对应 Workflow
    │
    ├── 有
    │    ↓
    │ 进入对应 Workflow
    │
    └── 没有
         ↓
      General Agent
         ↓
      理解需求
         ↓
    ┌────┼────┬────┐
    ↓    ↓    ↓    ↓
  直接  检索  工具  多步骤
  回答  资料  调用  规划
    └────┴────┴────┘
         ↓
       生成结果
```

当前 V02 阶段：

> 不实现新的 General Agent 功能，只保留已有 LegacyApp 和 Mock 能力。

---

# 16. Workflow 与 Agent 的关系

Lingxi-claw 使用：

```text
Workflow + Agent
```

混合模式。

---

## 16.1 Workflow

适用于：

* 高频
* 明确
* 结构化
* 可预测

例如：

```text
AI 作业辅导
资料处理
错题复习
```

---

## 16.2 Agent

适用于：

* 开放问题
* 复杂问题
* 非固定任务
* 无法提前定义完整步骤的任务

---

## 16.3 混合关系

```text
用户学习任务
      ↓
判断任务类型
      ↓
┌─────┴─────┐
↓           ↓
Workflow    Agent
↓           ↓
结构化执行  动态规划
└─────┬─────┘
      ↓
学习结果
```

---

# 17. Router Workflow

Router 属于内部系统能力。

用户不需要看到 Router，也不需要手动选择 Router。

未来系统可以：

```text
学习任务
    ↓
Task Router
    ↓
判断：
├── 当前课程
├── 当前功能
├── 任务类型
├── 输入类型
├── 学习状态
└── 任务复杂度
    ↓
选择：
├── Workflow
├── Agent
├── Skill
├── Parser
└── Model
```

Router 的目标是：

> **将任务交给最适合的处理能力。**

当前 V02 阶段：

> Router 只作为架构概念保留，不要求在前端页面中展示，也不要求重新实现已有 Router。

---

# 18. 端到端学习闭环

Lingxi-claw 最终希望形成完整的学习闭环：

```text
学习画像
    ↓
课程
    ↓
课程资料
    ↓
学习任务
    ↓
AI 作业辅导
    ↓
用户完成任务
    ↓
产生学习数据
    ↓
错题记录
    ↓
学习分析
    ↓
发现薄弱点
    ↓
AI 学习建议
    ↓
新的学习任务
    ↓
继续学习
```

最终形成：

```text
学习
 ↓
记录
 ↓
分析
 ↓
建议
 ↓
行动
 ↓
再次学习
```

这是 Lingxi-claw 与单次 AI Chatbot 的重要区别之一。

---

# 19. 当前 V02 Workflow 范围

当前开发阶段为：

> **Lingxi V02 — Product Architecture & UI Foundation**

当前目标不是完成完整 AI Workflow，而是建立完整的产品工作流页面框架。

---

## 19.1 当前必须实现

```text
✅ 学习画像页面框架
✅ Home
✅ Courses
✅ Course Space
✅ Course Space 五个功能区域
✅ Analytics
✅ Settings
✅ 页面路由
✅ 页面之间跳转
✅ Mock 数据
```

---

## 19.2 当前暂不实现

```text
❌ 真实 AI Workflow
❌ 真实 Agent
❌ 真实模型调用
❌ 真实 Router
❌ 真实知识库
❌ 真实文件解析
❌ OCR
❌ Vision
❌ 真实错题识别
❌ 真实学习分析
❌ 真实个性化推荐
```

---

# 20. Mock Mode 规则

当前开发必须保留 Mock Mode。

Mock Mode 用于：

* 页面展示
* 路由验证
* 前后端接口占位
* UI 联调
* Demo 演示
* 测试

Mock 数据不应被误认为真实 AI 结果。

例如：

```text
课程掌握度：72%
```

当前仅代表：

```text
Mock Data
```

不代表系统已经实现真实掌握度计算。

---

# 21. Legacy 功能兼容原则

项目已有 LegacyApp 和相关功能必须保留。

已有能力包括：

* 期末突击
* 日常作业辅助
* General Question
* Agent Settings
* 文件上传
* 文件分析 Mock
* 复习计划 Mock
* Chat Mock
* Agent 设置 Mock

V02 重构过程中：

```text
新产品 UI
      +
LegacyApp
```

可以并存。

原则：

> **先建立新产品架构，再逐步迁移已有能力，不因为 UI 重构而删除已有功能。**

不得因为新 Workflow 设计而删除：

* services
* API 类型
* API 调用代码
* Mock Mode
* LegacyApp

---

# 22. Workflow 与前后端职责

## Frontend

Frontend 负责：

```text
页面
 ↓
路由
 ↓
用户操作
 ↓
展示 Workflow 状态
 ↓
调用 API
 ↓
展示结果
```

Frontend 当前阶段不负责：

* AI Workflow 实现
* 模型调用
* 数据库
* 文件解析
* OCR

---

## Backend

Backend 负责：

```text
API
 ↓
任务处理
 ↓
Workflow
 ↓
Agent
 ↓
数据
 ↓
AI 能力
```

未来包括：

* Task Router
* Workflow
* Agent
* 文件处理
* 知识库
* 学习数据
* 学习分析

---

## QA / Testing

QA 负责验证：

```text
用户操作
 ↓
Frontend
 ↓
API
 ↓
Backend
 ↓
Workflow
 ↓
返回结果
 ↓
Frontend 展示
```

重点测试：

* 路由
* 页面状态
* API
* Mock Mode
* 前后端联调
* 异常情况
* 回归测试
* Demo 主流程

---

# 23. Workflow 验收标准

一个 Workflow 在产品层面完成，需要满足：

```text
① 用户知道从哪里进入
② 页面入口存在
③ 路由正确
④ 状态能够展示
⑤ Mock 数据能够正常工作
⑥ API 接口契约明确
⑦ 异常状态有对应处理
⑧ QA 能够验证
```

当前 V02 阶段：

> **只要求完成 ①～⑧ 的产品与 UI 骨架，不要求真实 AI 能力完成。**

---

# 24. 核心 Workflow 总览

Lingxi-claw 当前核心 Workflow 可以概括为：

```text
                    LINGXI
                       │
                学习画像 Workflow
                       │
                       ↓
                     Home
                       │
                 Courses Workflow
                       │
                       ↓
                  Course Space
                       │
       ┌───────────────┼────────────────┐
       ↓               ↓                ↓
     课程资料         AI 作业辅导       错题记录
       │               │                │
       └───────────────┼────────────────┘
                       ↓
                  学习分析 Workflow
                       │
                       ↓
                 AI 学习建议
                       │
                       ↓
                    Home
```

AI 内部：

```text
学习任务
   ↓
Task Router
   ↓
┌──────────────┐
│ Workflow     │
│      或      │
│ GeneralAgent │
└──────┬───────┘
       ↓
对应 AI 能力
       ↓
学习结果
       ↓
学习数据
       ↓
分析 / 建议
```

---

# 25. 最终 Workflow 原则

Lingxi-claw 的 Workflow 遵循以下原则：

1. **课程优先**：学习内容围绕课程组织。
2. **任务驱动**：AI 能力由具体学习任务触发。
3. **Workflow 优先**：明确、高频任务优先采用结构化 Workflow。
4. **Agent 兜底**：开放、复杂任务由 General Agent 处理。
5. **数据闭环**：学习、作业、错题和分析形成持续反馈。
6. **用户无感路由**：Router、Skill、Model 等内部能力不要求用户理解。
7. **Mock 优先开发**：V02 阶段优先完成页面、路由和交互框架。
8. **兼容已有功能**：不因新架构删除已有 services、API、Mock 或 LegacyApp。
9. **前后端解耦**：通过 API Contract 连接 Frontend 与 Backend。
10. **可测试**：每一个用户 Workflow 都必须能够被 QA 验证。

最终目标：

> **让 Lingxi-claw 从“一个可以聊天的 AI”，逐渐成为一个围绕课程持续工作的个人学习系统。**

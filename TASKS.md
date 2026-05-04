# TASKS.md — 信息汇总 Agent 开发清单

[简体中文](./TASKS.md) | [English](./TASKS.en.md)

## 0. 全局执行规则（GLOBAL RULES）

### 执行流程（必须遵守）

1. 阅读 AGENTS.md
2. 明确当前任务范围
3. 输出执行计划（3-6步）
4. 再开始编码

---

### 通用约束

* 每次只完成一个任务模块
* 不允许跨模块修改代码
* 所有代码必须可编译
* 必须运行测试（如果存在）
* 修改必须最小化

---

## 1. 初始化项目（Project Bootstrap）

### 任务目标

初始化一个 Go 项目骨架，符合 AGENTS.md 结构

---

### 执行步骤

1. 创建项目目录结构：

```bash
mkdir -p cmd internal/{crawler,processor,ai,service,repository,scheduler,delivery} pkg configs tests
```

2. 初始化 Go module：

```bash
go mod init InfoHub-agent
```

3. 创建基础入口：

* `cmd/main.go`

---

### 验证标准

* `go run cmd/main.go` 能成功运行
* 项目结构符合 AGENTS.md

---

## 2. 定义核心数据模型（Model Layer）

### 任务目标

实现核心数据结构 `NewsItem`

---

### 执行步骤

1. 创建文件：

```text
internal/model/news.go
```

2. 定义结构体：

* ID
* Title
* Content
* Source
* URL
* PublishTime
* Tags
* Score

---

### 验证标准

* 代码可编译
* 可被其他模块引用

---

## 3. 实现采集模块（Crawler）

### 任务目标

实现一个基础数据采集器（先用 mock / 简单数据源）

---

### 执行步骤

1. 定义接口：

```go
type Crawler interface {
    Fetch() ([]NewsItem, error)
}
```

2. 实现一个 Demo crawler：

* 返回模拟数据（不要接真实 API）

---

### 验证标准

* 能返回至少 3 条数据
* service 层可以调用

---

## 4. 实现处理模块（Processor）

### 任务目标

实现基础去重逻辑

---

### 执行步骤

1. 创建：

```text
internal/processor/deduplicate.go
```

2. 实现：

* 基于 Title 去重（map）
* 保留第一条

---

### 验证标准

* 输入重复数据 → 输出唯一数据
* 单元测试通过

---

## 5. 实现 AI 模块（AI Layer）

### 任务目标

封装 AI 调用接口（先 mock）

---

### 执行步骤

1. 定义接口：

```go
type AIProcessor interface {
    Summarize(item NewsItem) (NewsItem, error)
}
```

2. mock 实现：

* 给 Content 添加“摘要”
* 生成评分（1-5）

---

### 验证标准

* 每条数据都有 Score
* 输出结构符合 AGENTS.md

---

## 6. 实现业务流程（Service Orchestration）

### 任务目标

串联整个流程

---

### 执行步骤

创建：

```text
internal/service/pipeline.go
```

实现流程：

```text
Fetch → Deduplicate → AI → Return
```

---

### 验证标准

* 能返回处理后的完整数据列表
* 无 panic

---

## 7. 实现输出模块（Delivery）

### 任务目标

输出 Markdown 报告

---

### 执行步骤

1. 创建：

```text
internal/delivery/markdown.go
```

2. 实现：

输出格式：

```md
# 今日信息

## ⭐⭐⭐⭐⭐
- 标题
- 摘要
```

---

### 验证标准

* 能生成 markdown 字符串
* 内容结构正确

---

## 8. 集成主流程（Main Integration）

### 任务目标

打通完整链路

---

### 执行步骤

1. 在 `main.go` 中：

* 初始化 crawler
* 初始化 processor
* 初始化 AI
* 调用 pipeline
* 输出 markdown

---

### 验证标准

```bash
go run cmd/main.go
```

输出：

* 一份完整日报

---

## 9. 增加定时任务（Scheduler）

### 任务目标

实现自动执行

---

### 执行步骤

1. 使用 cron：

```go
cron.New().AddFunc("@every 1h", job)
```

2. 调用 pipeline

---

### 验证标准

* 每小时执行一次
* 无重复 crash

---

## 10. 引入真实数据源（Crawler Upgrade）

### 任务目标

接入真实数据（如 RSS / GitHub）

---

### 执行步骤

* 替换 mock crawler
* 支持至少一个真实来源

---

### 验证标准

* 数据来源真实
* 能正常解析

---

## 11. AI 接入真实模型（AI Upgrade）

### 任务目标

接入真实 LLM

---

### 执行步骤

* 封装 API client
* 实现 prompt：

```text
总结以下内容：
输出格式：
【发生了什么】
【为什么重要】
【影响】
【评分】
```

---

### 验证标准

* 输出真实摘要
* 格式正确

---

## 12. 推送能力（Delivery Upgrade）

### 任务目标

支持通知

---

### 执行步骤

实现至少一种：

* 邮件发送
* Webhook
* Bot

---

### 验证标准

* 成功发送消息
* 内容正确

---

## 13. 去重升级（Advanced）

### 任务目标

提升去重准确性

---

### 执行步骤

* 引入 embedding
* 计算相似度

---

### 验证标准

* 相似内容被合并
* 准确率提升

---

## 14. 排序策略（Scoring）

### 任务目标

实现智能排序

---

### 执行步骤

评分模型：

```text
score = 热度 + 时间 + AI评分
```

---

### 验证标准

* 高价值内容排前面

---

## 15. 个性化（Optional）

### 任务目标

支持用户偏好

---

### 执行步骤

* 标签过滤
* 订阅机制

---

### 验证标准

* 用户只看到感兴趣内容

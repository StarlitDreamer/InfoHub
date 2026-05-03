# AGENTS.md - 信息汇总 Agent 项目规范

## 1. 项目概览

本项目是一个“信息汇总与决策 Agent”，用于自动采集多源信息，完成去重、分类、摘要、打分，并输出高质量、可决策的信息日报。

核心目标：

- 减少用户筛选信息的时间
- 提供可直接辅助判断和行动的决策价值
- 支持自动推送到邮件、Webhook 或 Bot

## 2. 技术栈

- 语言：Go >= 1.22
- Web 框架：Gin
- 主存储：MySQL
- 缓存与去重：Redis
- AI 能力：OpenAI、Claude 或其他 LLM API
- 搜索能力：Elasticsearch，可选
- 定时任务：cron 或自定义 scheduler

## 3. 推荐项目结构

```text
project-root/
├── cmd/
│   └── main.go
├── internal/
│   ├── crawler/
│   ├── processor/
│   ├── ai/
│   ├── service/
│   ├── repository/
│   ├── scheduler/
│   └── delivery/
├── pkg/
├── configs/
├── scripts/
├── tests/
└── AGENTS.md
```

## 4. 核心数据结构

```go
type NewsItem struct {
    ID          int64
    Title       string
    Content     string
    Source      string
    URL         string
    PublishTime time.Time
    Tags        []string
    Score       float64
}
```

## 5. 核心模块职责

### crawler

- 从多个数据源采集信息
- 必须保证采集幂等
- 必须支持插件式扩展数据源

### processor

- 按标题、URL、内容指纹或 embedding 去重
- 清洗正文内容
- 聚合相似内容

### ai

必须提供以下能力：

- 分类
- 结构化摘要
- 重要性评分

AI 输出格式：

```text
【标题】
【发生了什么】
【为什么重要】
【影响】
【评分】1-5
```

### service

负责串联主业务流程：

```text
采集 -> 去重 -> AI 处理 -> 存储 -> 输出
```

### scheduler

- 支持 cron 定时触发
- 支持手动触发

### delivery

至少支持：

- Markdown 日报
- 邮件推送
- Webhook 或 Bot 推送

## 6. Agent 执行流程

Agent 执行信息处理时必须遵循：

1. 获取数据源列表
2. 执行采集
3. 执行去重
4. 调用 AI 处理
5. 存储结果
6. 按评分排序
7. 输出日报或推送结果

## 7. 代码规范

- 遵循 Go idiomatic 风格
- 函数保持短小，原则上不超过 50 行
- 模块职责单一
- 禁止巨型函数
- 业务能力必须有接口抽象
- 写代码必须写注释，注释统一使用简体中文
- 新增功能必须包含测试
- 不引入未声明依赖
- 不修改无关模块

## 8. Codex 工作规则

Codex 执行任务时必须：

1. 先读取项目结构
2. 明确目标模块
3. 输出 3-6 步执行计划
4. 列出预计影响文件
5. 再进行编码
6. 写代码时必须补充必要注释，且注释必须使用简体中文
7. 小步修改，避免一次性大范围改动
8. 完成后运行相关测试或说明未运行原因

禁止行为：

- 跳过分析阶段直接编码
- 无计划编码
- 一次性修改大量文件
- 修改与任务无关的模块
- 引入未声明依赖

## 9. 开发优先级

### 第一阶段：MVP

- 数据采集
- AI 摘要
- Markdown 输出

### 第二阶段

- 去重
- 打分排序

### 第三阶段

- 推送系统
- 个性化推荐

## 10. 未来扩展

- 用户兴趣推荐
- GitHub 项目分析
- 自动生成报告
- 自动执行动作，例如收藏项目

## 11. 成功标准

- 每日自动生成信息摘要
- 无重复内容
- 输出结构化
- 用户阅读时间不超过 5 分钟

## 12. 常用命令

```bash
# 启动服务
go run cmd/main.go

# 测试
go test ./...

# 构建
go build -o app cmd/main.go

# 格式化
go fmt ./...
```

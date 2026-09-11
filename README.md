# B2B 营销 Agent Harness

> **与领域无关的 Agent Harness + 半导体营销参考实现。**
> 一个面向 B2B 半导体 / 电子元器件 / 工业科技营销自动化的开源 Agent 脚手架。

```
研究 → 情报 → 内容 → 审批 → 发布 → 分析 → 优化
```

适用对象：

- 半导体企业
- 电子元器件分销商
- B2B 制造企业
- 工业科技公司
- 技术营销团队

> 首个渠道是 **LinkedIn**（参考 Microchip 等企业运营官网 + LinkedIn 主页的方式）。
> LinkedIn 是 *其中一个* 渠道，而不是产品本身。

---

## 为什么

B2B 技术营销不是「LinkedIn 自动发帖」的问题，而是：

- **研究（Research）** — 行业新闻、竞品帖子、技术趋势
- **情报（Intelligence）** — 主题聚类、选题发现、打分
- **内容（Content）** — 基于产品知识的多语言草稿生成
- **审核（Review）** — 技术事实与品牌口吻检查
- **审批（Approval）** — 人工决定哪些内容可以发布
- **发布（Publishing）** — 排期与多渠道发布
- **分析（Analytics）** — 把结果反馈到下一轮循环

本仓库是 **大脑 + 神经系统**：`Agent` / `Workflow` / `Tool` / `Memory` / `Knowledge` / `Policy` / `Approval` / `Observability`。

**手** 是数据采集层 — 例如 [`crawl-worker-redis`](https://github.com/esanwu-bot/crawl-worker-redis) —
通过干净的 `Tool` 接口接入；可以换成自己的爬虫。

---

## 架构

```
                     ┌──────────────────────────┐
                     │   管理控制台 (Vue 3)       │
                     └────────────┬─────────────┘
                                  │ REST + SSE
                     ┌────────────▼─────────────┐
                     │        API Gateway        │
                     └────────────┬─────────────┘
         ┌────────────────────────┼─────────────────────────┐
         ▼                        ▼                         ▼
   Agent Manager            Workflow Engine             Knowledge
         └────────────────────────┼─────────────────────────┘
                                  ▼
                       ┌───────────────────┐
                       │   Agent Runtime   │  规划 → 上下文 → 执行 → 观察 → 反思
                       └─────────┬─────────┘
                    ┌────────────┴────────────┐
                    ▼                         ▼
              Tool Registry             Policy Engine
                    │
     ┌──────────────┼───────────────────────────┐
     ▼              ▼                            ▼
 Crawl Tool    LinkedIn Tool              Knowledge Tool
     │              │                            │
crawl-worker    LinkedIn API            PostgreSQL (+pgvector)
  -redis(MySQL)                              Redis Queue/Event
```

**原则：** `Workflow` 决定「何时 / 做什么」，`Agent` 决定「怎么做」，`Tool` 真正「执行」，
`Memory` 负责记忆，`Knowledge` 负责知识，`Policy` 负责策略放行。

详见 [`docs/architecture.md`](docs/architecture.md)、[`项目需求.md`](项目需求.md)、
[`开发计划.md`](开发计划.md)。

---

## 快速开始

```bash
# 1. 准备 PostgreSQL
psql -U postgres -c "CREATE DATABASE b2b_marketing_agent;"
psql -U postgres -d b2b_marketing_agent -f schema.sql

# 2. 配置（可选；默认以 mock 模式开箱即用）
cp .env.example .env

# 3. 构建并跑通离线 demo（mock LLM + 内存存储，无需 Redis）
make demo

# 4. 启动 API
make run-api        # http://localhost:8090/api/v1/health

# 5. 启动 Workbench 控制台（Vue 3，UI 全部简体中文）
cd web && npm install && npm run dev   # http://localhost:5173

# 或者直接一键启动前后端
start_all.bat
```

### Mock 优先

Harness 在离线环境下即可完整运行，使用确定性的 **mock LLM** 与 **内存存储**，
可以在接入 PostgreSQL / Redis / 真实 LLM / LinkedIn 之前，先把整条流水线（研究 → … → 分析）
跑通验证。Workbench 在后端未启动时，也会自动用内置演示数据兜底渲染。

```bash
make demo     # 端到端跑通半导体参考流水线
make test     # 运行单元测试
```

---

## 技术栈

| 层 | 选型 |
|----|------|
| 语言 | Go 1.23+ |
| API | `net/http`（ServeMux）+ SSE |
| 数据库 | PostgreSQL 16（主数据源）+ pgvector（可选） |
| 队列 / 缓存 / 事件 / 锁 | Redis Streams |
| LLM | OpenAI / Anthropic / Gemini / OpenAI-compatible（内置 `mock`） |
| 工具协议 | Native Tool + MCP（规划中） |
| 爬虫 | crawl-worker-redis（通过 Tool 适配器接入） |
| 前端 | Vue 3 + Ant Design Vue |
| 部署 | Docker Compose → Kubernetes |

---

## 仓库结构

```
cmd/{api,worker,scheduler,harness}
internal/
  domain/        契约层（零外部依赖）
  agent/         runtime、planner、context builder、registry
  tool/          registry + crawler/content/knowledge/linkedin/notification/browser
  llm/           provider、mock、openai-compatible、router
  workflow/      定义、引擎、调度
  policy/        策略引擎（自动 / AI 审核 / 人工审批）
  store/         store 接口 + 内存实现
  builtin/       6 个参考 Agent + 每日情报 Workflow（半导体）
  observability/ 日志 / Trace / 成本
  app/           依赖装配
  transport/httpapi
web/                       Workbench 控制台（Vue 3 + Ant Design Vue）
schema.sql                 PostgreSQL 建表脚本
```

---

## 内置 Agent（参考实现）

| Agent | 职责 |
|-------|------|
| 研究（Research） | 采集行业 / 竞品情报 |
| 趋势（Trend） | 主题发现与聚类 |
| 策略（Strategy） | 选题打分与选择 |
| 内容（Content） | 生成多语言草稿 |
| 审核（Review） | 技术事实 + 品牌口吻审核 |
| 分析（Analytics） | 表现反馈与优化建议 |

---

## 与爬虫的数据边界

```
b2b-marketing-agent-harness   （为何 / 做什么 / 怎么决策）   PostgreSQL
        │  Tool / API / Event
crawl-worker-redis            （如何采集）                  MySQL
```

Harness **永远不**直接读取爬虫的数据库；它通过抽象为 `Tool` 的 `Crawler` 接口调用，
因此可以替换为 Firecrawl / Browserbase / Apify 或你自己的爬虫。

---

## 许可证

MIT — 详见 [LICENSE](LICENSE)。

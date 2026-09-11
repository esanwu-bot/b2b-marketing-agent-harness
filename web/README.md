# 营销 Agent 工作台（Workbench）

> B2B 营销 Agent Harness 的控制台前端，界面依据根目录 `preview.html` 原型实现，文案全部为**简体中文**。
> 消费后端 `crawler harness API`（`/api/v1`，默认 `:8090`）。

## 技术栈

- Vue 3 + TypeScript + Vite
- Ant Design Vue 4
- Vue Router 4
- Axios（REST）

## 快速开始

```bash
cd web
npm install

# 另开终端启动后端 API（默认 :8090）
#   cd .. && go run ./cmd/api -config configs/config.yaml

npm run dev        # http://localhost:5173
```

开发期 Vite 将 `/api` 代理到 `VITE_API_TARGET`（默认 `http://localhost:8090`）。
后端未启动时，界面使用内置演示数据兜底渲染（顶栏显示「本地演示数据」）。

## 构建

```bash
npm run build      # 类型检查 + 产物到 dist/
npm run preview
```

生产环境用 Nginx 托管 `dist/`，并将 `/api` 反向代理到 crawler harness API。

## 页面

| 页面 | 路由 | 说明 |
| --- | --- | --- |
| Agent 工作台 | `/workbench` | 总览：活跃运行、内容流水线、实时轨迹、审批、Agent 表现 |
| Agent 管理 | `/agents` | 6 个内置 Agent，可手动运行 |
| 工作流 | `/workflows` | Workflow 定义与运行记录 |
| 工具注册表 | `/tools` | Tool 列表、风险等级、是否需审批 |
| 采集任务 | `/tasks` | Agent 触发的采集任务 |
| 内容工作室 | `/content` | 草稿 / 内容日历 / 发布记录 |
| 知识库 | `/knowledge` | 文档检索与选题候选 |
| 数据分析 | `/analytics` | Agent 表现、内容状态分布、发布指标 |
| 审批中心 | `/approvals` | Human-in-the-loop 审批 |
| 可观测性 | `/observability` | 运行轨迹（步骤级）+ 领域事件 |

## 与后端接口对照

`GET /health`、`GET /metrics`、`GET /agents`、`GET /tools`、`GET /workflows`、
`POST /workflows/{key}/run`、`GET /runs`、`GET /workflow-runs`、`GET /tasks`、
`GET /approvals`、`POST /approvals/{id}/approve|reject`、`GET /knowledge`、
`GET /knowledge/search`、`GET /topics`、`GET /content`、`GET /publications`、`GET /events`。

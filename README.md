# B2B Marketing Agent Harness

> **Domain-independent Agent Harness + Semiconductor Marketing Reference Implementation.**
> An open-source agent harness for B2B semiconductor / electronic-component / industrial-technology marketing automation.

```
Research → Intelligence → Content → Approval → Publishing → Analytics → Optimization
```

Built for:

- Semiconductor companies
- Electronic component distributors
- B2B manufacturers
- Industrial technology companies
- Technical marketing teams

> The first channel is **LinkedIn** (inspired by how companies like Microchip Taiwan run
> their corporate site + LinkedIn presence). LinkedIn is *one* channel, not the product.

---

## Why

B2B technical marketing is not a "LinkedIn auto-poster" problem. It is:

- **Research** — industry news, competitor posts, technology trends
- **Intelligence** — clustering, topic detection, scoring
- **Content** — multilingual drafting grounded in your product knowledge
- **Review** — technical-fact and brand-voice checking
- **Approval** — humans decide what gets published
- **Publishing** — schedule and post to channels
- **Analytics** — feed results back into the next cycle

This repository is the **brain + nervous system**: `Agent` / `Workflow` / `Tool` /
`Memory` / `Knowledge` / `Policy` / `Approval` / `Observability`.

The **hands** are a data-acquisition layer — e.g. [`crawl-worker-redis`](https://github.com/esanwu-bot/crawl-worker-redis) —
integrated through a clean `Tool` interface. Bring your own crawler.

---

## Architecture

```
                    ┌──────────────────────────┐
                    │   Admin Console (Vue 3)   │
                    └────────────┬─────────────┘
                                 │ REST + SSE
                    ┌────────────▼─────────────┐
                    │        API Gateway        │
                    └────────────┬─────────────┘
        ┌────────────────────────┼─────────────────────────┐
        ▼                        ▼                         ▼
  Agent Manager            Workflow Engine            Knowledge
        └────────────────────────┼─────────────────────────┘
                                 ▼
                       ┌───────────────────┐
                       │   Agent Runtime   │  Plan → Context → Execute → Observe → Reflect
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

**Principle:** `Workflow` decides *when/what*, `Agent` decides *how*, `Tool` *executes*,
`Memory` remembers, `Knowledge` knows, `Policy` permits.

See [`docs/architecture.md`](docs/architecture.md), [`项目需求.md`](项目需求.md),
[`开发计划.md`](开发计划.md).

---

## Quick start

```bash
# 1. prepare PostgreSQL
psql -U postgres -c "CREATE DATABASE b2b_marketing_agent;"
psql -U postgres -d b2b_marketing_agent -f schema.sql

# 2. configure (optional; defaults work in mock mode)
cp .env.example .env

# 3. build & run the offline demo (mock LLM + in-memory store, no Redis needed)
make demo

# 4. start the API
make run-api        # http://localhost:8090/api/v1/health

# 5. start the Workbench console (Vue 3, all UI in Simplified Chinese)
cd web && npm install && npm run dev   # http://localhost:5173
```

### Mock-first

The harness runs fully offline with a deterministic **mock LLM** and an **in-memory store**,
so you can exercise the whole pipeline (research → … → analytics) before wiring PostgreSQL,
Redis, real LLMs or LinkedIn. The Workbench also falls back to bundled demo data when the API
is not running.

```bash
make demo     # runs the semiconductor reference pipeline end-to-end
make test
```

---

## Tech stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.23+ |
| API | `net/http` (ServeMux) + SSE |
| DB | PostgreSQL 16 (source of truth) + pgvector (optional) |
| Queue/Cache/Event/Lock | Redis Streams |
| LLM | OpenAI / Anthropic / Gemini / OpenAI-compatible (built-in `mock`) |
| Tool protocol | Native Tool + MCP (planned) |
| Crawler | crawl-worker-redis (via Tool adapter) |
| Frontend | Vue 3 + Vben Admin |
| Deployment | Docker Compose → Kubernetes |

---

## Repository layout

```
cmd/{api,worker,scheduler,harness}
internal/
  domain/        contracts (zero external deps)
  agent/         runtime, planner, context builder, registry
  tool/          registry + crawler/content/knowledge/linkedin/notification/browser
  llm/           provider, mock, openai-compatible, router
  workflow/      definition, engine, scheduler
  policy/        policy engine (auto / AI-review / human approval)
  store/         store interfaces + in-memory impl
  builtin/       6 reference agents + daily workflow (semiconductor)
  observability/ logging / trace / cost
  app/           dependency wiring
  transport/httpapi
web/                       Workbench console (Vue 3 + Ant Design Vue)
schema.sql                 PostgreSQL schema
```

---

## Built-in agents (reference)

| Agent | Responsibility |
|-------|----------------|
| Research | collect industry / competitor intelligence |
| Trend | detect and cluster topics |
| Strategy | score and select topics |
| Content | generate multilingual drafts |
| Review | technical-fact + brand-voice review |
| Analytics | performance feedback and optimization |

---

## Data boundary with the crawler

```
b2b-marketing-agent-harness   (Why / What / Decide)   PostgreSQL
        │  Tool / API / Event
crawl-worker-redis            (How to collect)         MySQL
```

The harness **never** reads the crawler's database directly. It calls a `Crawler` interface
abstracted as a Tool, so you can swap in Firecrawl / Browserbase / Apify / your own crawler.

---

## License

MIT — see [LICENSE](LICENSE).

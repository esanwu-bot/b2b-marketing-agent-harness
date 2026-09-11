-- ============================================================================
-- B2B Marketing Agent Harness — PostgreSQL Schema (V1)
-- ----------------------------------------------------------------------------
-- 定位：Agent Harness 的 Source of Truth（PostgreSQL 独占）。
--       crawler 的 MySQL 是独立 bounded context，本库不直接访问它，
--       两者只通过 Tool / API / Event 交互。
--
-- 依赖扩展：
--   pgcrypto  -> gen_random_uuid()
--   pg_trgm   -> 文本模糊检索
--   vector    -> pgvector 向量检索（安装 pgvector 后可用；无则见文末降级说明）
--
-- 约定：
--   * 主键统一 UUID
--   * 半结构化数据（Agent State / Tool IO / Workflow 定义）统一 JSONB
--   * 时间统一 TIMESTAMPTZ，默认 now()
--   * 业务枚举用 VARCHAR + CHECK，不建独立字典表（V1 从简）
--   * 所有表带 created_at/updated_at，updated_at 由触发器维护
-- ============================================================================

-- 扩展创建容错：缺少扩展（尤其 pgvector）不影响建库。
-- pgcrypto  -> gen_random_uuid()
-- pg_trgm   -> 文本模糊检索
-- vector    -> pgvector 向量检索（可选；未装则 embedding 以 JSONB 存储，见文末升级说明）
DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS pgcrypto;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pgcrypto 不可用，跳过：%', SQLERRM;
END $$;
DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS pg_trgm;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pg_trgm 不可用，跳过：%', SQLERRM;
END $$;
DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS vector;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pgvector 未安装，跳过；embedding 仍以 JSONB 存储';
END $$;

-- ---------------------------------------------------------------------------
-- 0. 通用触发器：自动维护 updated_at
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- 1. 工作区 / 渠道 / 凭据
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS workspaces (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(200) NOT NULL,
    slug             VARCHAR(120) NOT NULL UNIQUE,
    industry         VARCHAR(120) NOT NULL DEFAULT 'semiconductor',
    region           VARCHAR(80)  NOT NULL DEFAULT 'TW',
    default_language VARCHAR(16)  NOT NULL DEFAULT 'zh-TW',
    timezone         VARCHAR(64)  NOT NULL DEFAULT 'Asia/Taipei',
    description      TEXT         NOT NULL DEFAULT '',
    settings         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    status           VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_workspaces_updated BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS channels (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    kind         VARCHAR(20) NOT NULL CHECK (kind IN ('linkedin','website','youtube','newsletter','x','email','facebook')),
    name         VARCHAR(200) NOT NULL,
    external_id  VARCHAR(200) NOT NULL DEFAULT '',
    status       VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled','error')),
    config       JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, kind, external_id)
);
CREATE TRIGGER trg_channels_updated BEFORE UPDATE ON channels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS channel_credentials (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id        UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    provider          VARCHAR(50)  NOT NULL,
    access_token_enc  TEXT         NOT NULL DEFAULT '',
    refresh_token_enc TEXT         NOT NULL DEFAULT '',
    token_type        VARCHAR(30)  NOT NULL DEFAULT 'Bearer',
    scopes            TEXT[]       NOT NULL DEFAULT '{}',
    expires_at        TIMESTAMPTZ,
    extra             JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_channel_credentials_updated BEFORE UPDATE ON channel_credentials
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 2. Agent 定义与版本
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS agents (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id     UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    key              VARCHAR(100) NOT NULL,               -- research / trend / strategist ...
    name             VARCHAR(200) NOT NULL,
    role             VARCHAR(100) NOT NULL DEFAULT '',
    description      TEXT         NOT NULL DEFAULT '',
    system_prompt    TEXT         NOT NULL DEFAULT '',
    llm_provider     VARCHAR(50)  NOT NULL DEFAULT '',    -- 空 => 用工作区/全局默认
    llm_model        VARCHAR(100) NOT NULL DEFAULT '',
    temperature      NUMERIC(4,2) NOT NULL DEFAULT 0.30,
    max_steps        INT          NOT NULL DEFAULT 12,
    max_tool_calls   INT          NOT NULL DEFAULT 20,
    max_cost         NUMERIC(12,4) NOT NULL DEFAULT 1.0,
    require_approval BOOLEAN      NOT NULL DEFAULT false,
    allowed_tools    TEXT[]       NOT NULL DEFAULT '{}',
    config           JSONB        NOT NULL DEFAULT '{}'::jsonb,
    status           VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, key)
);
CREATE TRIGGER trg_agents_updated BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS agent_versions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id   UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    version    INT  NOT NULL,
    spec       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (agent_id, version)
);

-- ---------------------------------------------------------------------------
-- 3. Tool 注册表与执行记录
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tools (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key               VARCHAR(200) NOT NULL UNIQUE,       -- crawler.create_job / content.generate ...
    name              VARCHAR(200) NOT NULL,
    description       TEXT         NOT NULL DEFAULT '',
    kind              VARCHAR(30)  NOT NULL DEFAULT 'native' CHECK (kind IN ('native','mcp','http','grpc')),
    input_schema      JSONB        NOT NULL DEFAULT '{}'::jsonb,
    output_schema     JSONB        NOT NULL DEFAULT '{}'::jsonb,
    requires_approval BOOLEAN      NOT NULL DEFAULT false,
    risk              VARCHAR(20)  NOT NULL DEFAULT 'read' CHECK (risk IN ('read','write','external')),
    config            JSONB        NOT NULL DEFAULT '{}'::jsonb,
    status            VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_tools_updated BEFORE UPDATE ON tools
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS tool_executions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id      UUID,
    run_id            UUID,
    step_id           UUID,
    tool_key          VARCHAR(200) NOT NULL,
    arguments         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    output            JSONB,
    status            VARCHAR(20)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','success','failed','rejected','awaiting_approval')),
    error             TEXT         NOT NULL DEFAULT '',
    requires_approval BOOLEAN      NOT NULL DEFAULT false,
    approval_id       UUID,
    latency_ms        INT          NOT NULL DEFAULT 0,
    started_at        TIMESTAMPTZ,
    finished_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tool_executions_run ON tool_executions(run_id);
CREATE INDEX IF NOT EXISTS idx_tool_executions_tool ON tool_executions(tool_key, created_at DESC);

-- ---------------------------------------------------------------------------
-- 4. Task / Run / Step（Agent 执行核心：最重要的链路）
--    task → task_run → task_step → tool_execution / llm_call
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    workflow_run_id UUID,
    agent_id        UUID REFERENCES agents(id) ON DELETE SET NULL,
    type            VARCHAR(100) NOT NULL,                -- collect_industry_news / generate_linkedin_posts ...
    goal            TEXT         NOT NULL DEFAULT '',
    input           JSONB        NOT NULL DEFAULT '{}'::jsonb,
    priority        INT          NOT NULL DEFAULT 0,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','queued','running','succeeded','failed','cancelled','awaiting_approval')),
    scheduled_at    TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_tasks_updated BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status, priority DESC, created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_workflow_run ON tasks(workflow_run_id);

CREATE TABLE IF NOT EXISTS task_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id       UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    agent_id      UUID REFERENCES agents(id) ON DELETE SET NULL,
    agent_version INT          NOT NULL DEFAULT 1,
    status        VARCHAR(20)  NOT NULL DEFAULT 'running' CHECK (status IN ('running','succeeded','failed','cancelled','awaiting_approval')),
    state         JSONB        NOT NULL DEFAULT '{}'::jsonb,   -- AgentState（可 Resume）
    result        JSONB,
    error         TEXT         NOT NULL DEFAULT '',
    steps         INT          NOT NULL DEFAULT 0,
    prompt_tokens INT          NOT NULL DEFAULT 0,
    comp_tokens   INT          NOT NULL DEFAULT 0,
    cost          NUMERIC(12,4) NOT NULL DEFAULT 0,
    started_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    finished_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_task_runs_updated BEFORE UPDATE ON task_runs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_task_runs_task ON task_runs(task_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_runs_status ON task_runs(status);

CREATE TABLE IF NOT EXISTS task_steps (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id     UUID NOT NULL REFERENCES task_runs(id) ON DELETE CASCADE,
    seq        INT  NOT NULL,
    kind       VARCHAR(20) NOT NULL CHECK (kind IN ('plan','llm','tool','observe','reflect','approval','output','error')),
    name       VARCHAR(200) NOT NULL DEFAULT '',
    input      JSONB,
    output     JSONB,
    status     VARCHAR(20) NOT NULL DEFAULT 'success' CHECK (status IN ('running','success','failed','skipped')),
    latency_ms INT  NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (run_id, seq)
);
CREATE INDEX IF NOT EXISTS idx_task_steps_run ON task_steps(run_id, seq);

CREATE TABLE IF NOT EXISTS agent_messages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id       UUID NOT NULL REFERENCES task_runs(id) ON DELETE CASCADE,
    seq          INT  NOT NULL,
    role         VARCHAR(20) NOT NULL CHECK (role IN ('system','user','assistant','tool')),
    name         VARCHAR(120) NOT NULL DEFAULT '',
    content      TEXT         NOT NULL DEFAULT '',
    tool_call_id VARCHAR(120) NOT NULL DEFAULT '',
    tool_calls   JSONB,
    tokens       INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (run_id, seq)
);

CREATE TABLE IF NOT EXISTS decisions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id     UUID REFERENCES task_runs(id) ON DELETE CASCADE,
    step_id    UUID REFERENCES task_steps(id) ON DELETE SET NULL,
    kind       VARCHAR(60)  NOT NULL,
    summary    TEXT         NOT NULL DEFAULT '',
    rationale  TEXT         NOT NULL DEFAULT '',
    confidence NUMERIC(4,3) NOT NULL DEFAULT 0,
    payload    JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_decisions_run ON decisions(run_id);

CREATE TABLE IF NOT EXISTS llm_calls (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id            UUID REFERENCES task_runs(id) ON DELETE CASCADE,
    step_id           UUID REFERENCES task_steps(id) ON DELETE SET NULL,
    provider          VARCHAR(50)  NOT NULL,
    model             VARCHAR(120) NOT NULL,
    request           JSONB        NOT NULL DEFAULT '{}'::jsonb,
    response          JSONB,
    prompt_tokens     INT          NOT NULL DEFAULT 0,
    completion_tokens INT          NOT NULL DEFAULT 0,
    total_tokens      INT          NOT NULL DEFAULT 0,
    cost              NUMERIC(12,6) NOT NULL DEFAULT 0,
    latency_ms        INT          NOT NULL DEFAULT 0,
    status            VARCHAR(20)  NOT NULL DEFAULT 'success',
    error             TEXT         NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_llm_calls_run ON llm_calls(run_id);
CREATE INDEX IF NOT EXISTS idx_llm_calls_created ON llm_calls(created_at DESC);

-- ---------------------------------------------------------------------------
-- 5. Workflow 定义与运行
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS workflows (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    key          VARCHAR(120) NOT NULL,
    name         VARCHAR(200) NOT NULL,
    description  TEXT         NOT NULL DEFAULT '',
    definition   JSONB        NOT NULL DEFAULT '{"steps":[]}'::jsonb,
    trigger      JSONB        NOT NULL DEFAULT '{}'::jsonb,     -- {type:"cron",schedule:"0 8 * * *"}
    status       VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, key)
);
CREATE TRIGGER trg_workflows_updated BEFORE UPDATE ON workflows
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS workflow_runs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id  UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running','succeeded','failed','cancelled','awaiting_approval')),
    trigger_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    context      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    error        TEXT        NOT NULL DEFAULT '',
    started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_workflow_runs_updated BEFORE UPDATE ON workflow_runs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_workflow_runs_wf ON workflow_runs(workflow_id, started_at DESC);

CREATE TABLE IF NOT EXISTS workflow_steps (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_run_id  UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    seq              INT  NOT NULL,
    kind             VARCHAR(20) NOT NULL CHECK (kind IN ('agent','approval','channel','delay','condition')),
    ref              VARCHAR(200) NOT NULL DEFAULT '',
    input            JSONB,
    output           JSONB,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','success','failed','skipped','awaiting_approval')),
    task_id          UUID REFERENCES tasks(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workflow_run_id, seq)
);
CREATE TRIGGER trg_workflow_steps_updated BEFORE UPDATE ON workflow_steps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 6. Memory（Agent/用户/工作区“记得什么”）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS memories (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    scope        VARCHAR(20) NOT NULL CHECK (scope IN ('agent','user','workspace','task')),
    scope_id     VARCHAR(200) NOT NULL DEFAULT '',
    kind         VARCHAR(30) NOT NULL DEFAULT 'fact' CHECK (kind IN ('preference','fact','summary','episode')),
    key          VARCHAR(300) NOT NULL DEFAULT '',
    content      TEXT        NOT NULL,
    embedding    JSONB,                                    -- 见文末 pgvector 升级说明
    importance   NUMERIC(4,3) NOT NULL DEFAULT 0.5,
    metadata     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_memories_updated BEFORE UPDATE ON memories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_memories_scope ON memories(workspace_id, scope, scope_id);
DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_memories_content_trgm ON memories USING gin (content gin_trgm_ops);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'pg_trgm 索引跳过：%', SQLERRM; END $$;

-- ---------------------------------------------------------------------------
-- 7. Knowledge（“知道什么”：文档 / 分块 / 实体 / 关系）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS knowledge_documents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    source_type  VARCHAR(40) NOT NULL DEFAULT 'crawl' CHECK (source_type IN ('crawl','rss','website','manual','linkedin','upload','api')),
    source_id    VARCHAR(300) NOT NULL DEFAULT '',
    title        VARCHAR(500) NOT NULL DEFAULT '',
    url          VARCHAR(1000) NOT NULL DEFAULT '',
    language     VARCHAR(16)  NOT NULL DEFAULT '',
    summary      TEXT         NOT NULL DEFAULT '',
    content      TEXT         NOT NULL DEFAULT '',
    doc_metadata JSONB        NOT NULL DEFAULT '{}'::jsonb,
    hash         VARCHAR(64)  NOT NULL DEFAULT '',        -- 去重指纹
    status       VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived','error')),
    fetched_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, hash)
);
CREATE TRIGGER trg_knowledge_documents_updated BEFORE UPDATE ON knowledge_documents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_kd_workspace ON knowledge_documents(workspace_id, created_at DESC);
DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_kd_title_trgm ON knowledge_documents USING gin (title gin_trgm_ops);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'pg_trgm 索引跳过：%', SQLERRM; END $$;

CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    seq         INT  NOT NULL,
    content     TEXT NOT NULL,
    token_count INT  NOT NULL DEFAULT 0,
    embedding   JSONB,                                     -- 见文末 pgvector 升级说明
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (document_id, seq)
);
CREATE INDEX IF NOT EXISTS idx_kc_document ON knowledge_chunks(document_id, seq);

CREATE TABLE IF NOT EXISTS knowledge_entities (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id   UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    type           VARCHAR(40) NOT NULL DEFAULT 'topic',   -- product/technology/company/person/application/topic
    name           VARCHAR(300) NOT NULL,
    canonical_name VARCHAR(300) NOT NULL,
    metadata       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, type, canonical_name)
);
CREATE TRIGGER trg_knowledge_entities_updated BEFORE UPDATE ON knowledge_entities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS knowledge_relations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    from_entity  UUID NOT NULL REFERENCES knowledge_entities(id) ON DELETE CASCADE,
    to_entity    UUID NOT NULL REFERENCES knowledge_entities(id) ON DELETE CASCADE,
    relation     VARCHAR(60) NOT NULL,
    weight       NUMERIC(6,3) NOT NULL DEFAULT 1.0,
    metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (from_entity, to_entity, relation)
);

-- ---------------------------------------------------------------------------
-- 8. 内容域：Topic → Content → Version/Asset → Publication → Metrics
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS topics (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    title        VARCHAR(500) NOT NULL,
    summary      TEXT         NOT NULL DEFAULT '',
    category     VARCHAR(100) NOT NULL DEFAULT '',
    keywords     TEXT[]       NOT NULL DEFAULT '{}',
    source_refs  JSONB        NOT NULL DEFAULT '[]'::jsonb,  -- 来自哪些知识/采集结果
    score        NUMERIC(5,2) NOT NULL DEFAULT 0,
    cluster_key  VARCHAR(200) NOT NULL DEFAULT '',
    status       VARCHAR(20)  NOT NULL DEFAULT 'candidate' CHECK (status IN ('candidate','approved','rejected','used')),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_topics_updated BEFORE UPDATE ON topics
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_topics_workspace ON topics(workspace_id, score DESC);

CREATE TABLE IF NOT EXISTS content (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    topic_id        UUID REFERENCES topics(id) ON DELETE SET NULL,
    channel_kind    VARCHAR(20) NOT NULL DEFAULT 'linkedin',
    format          VARCHAR(30) NOT NULL DEFAULT 'post' CHECK (format IN ('post','article','video_script','newsletter','comment_reply')),
    language        VARCHAR(16) NOT NULL DEFAULT 'zh-TW',
    title           VARCHAR(500) NOT NULL DEFAULT '',
    body            TEXT        NOT NULL DEFAULT '',
    hashtags        TEXT[]      NOT NULL DEFAULT '{}',
    status          VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','in_review','approved','scheduled','published','rejected','failed')),
    author_agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    created_by      VARCHAR(100) NOT NULL DEFAULT '',
    scheduled_at    TIMESTAMPTZ,
    published_at    TIMESTAMPTZ,
    external_url    VARCHAR(1000) NOT NULL DEFAULT '',
    metadata        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_content_updated BEFORE UPDATE ON content
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_content_workspace ON content(workspace_id, status, created_at DESC);
DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_content_title_trgm ON content USING gin (title gin_trgm_ops);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'pg_trgm 索引跳过：%', SQLERRM; END $$;

CREATE TABLE IF NOT EXISTS content_versions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id  UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    version     INT  NOT NULL,
    title       VARCHAR(500) NOT NULL DEFAULT '',
    body        TEXT NOT NULL DEFAULT '',
    editor      VARCHAR(100) NOT NULL DEFAULT '',
    change_note TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (content_id, version)
);

CREATE TABLE IF NOT EXISTS content_assets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id  UUID REFERENCES content(id) ON DELETE CASCADE,
    kind        VARCHAR(20) NOT NULL DEFAULT 'image' CHECK (kind IN ('image','video','doc','cover')),
    url         VARCHAR(1000) NOT NULL DEFAULT '',
    object_key  VARCHAR(500)  NOT NULL DEFAULT '',
    mime        VARCHAR(120)  NOT NULL DEFAULT '',
    width       INT NOT NULL DEFAULT 0,
    height      INT NOT NULL DEFAULT 0,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_content_assets_content ON content_assets(content_id);

CREATE TABLE IF NOT EXISTS publications (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id       UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    channel_id       UUID REFERENCES channels(id) ON DELETE SET NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','scheduled','publishing','published','failed','deleted')),
    external_id      VARCHAR(300) NOT NULL DEFAULT '',
    external_url     VARCHAR(1000) NOT NULL DEFAULT '',
    scheduled_at     TIMESTAMPTZ,
    published_at     TIMESTAMPTZ,
    error            TEXT        NOT NULL DEFAULT '',
    metrics_synced_at TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_publications_updated BEFORE UPDATE ON publications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_publications_content ON publications(content_id);
CREATE INDEX IF NOT EXISTS idx_publications_status ON publications(status, scheduled_at);

CREATE TABLE IF NOT EXISTS content_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publication_id  UUID NOT NULL REFERENCES publications(id) ON DELETE CASCADE,
    measured_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    impressions     INT NOT NULL DEFAULT 0,
    reach           INT NOT NULL DEFAULT 0,
    likes           INT NOT NULL DEFAULT 0,
    comments        INT NOT NULL DEFAULT 0,
    shares          INT NOT NULL DEFAULT 0,
    clicks          INT NOT NULL DEFAULT 0,
    engagement_rate NUMERIC(7,4) NOT NULL DEFAULT 0,
    ctr             NUMERIC(7,4) NOT NULL DEFAULT 0,
    raw             JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (publication_id, measured_at)
);
CREATE INDEX IF NOT EXISTS idx_content_metrics_pub ON content_metrics(publication_id, measured_at DESC);

-- ---------------------------------------------------------------------------
-- 9. 互动 / 线索（LinkedIn 评论、私信、提及）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS engagements (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id         UUID REFERENCES channels(id) ON DELETE CASCADE,
    publication_id     UUID REFERENCES publications(id) ON DELETE SET NULL,
    kind               VARCHAR(20) NOT NULL CHECK (kind IN ('comment','mention','message')),
    external_id        VARCHAR(300) NOT NULL DEFAULT '',
    author_name        VARCHAR(200) NOT NULL DEFAULT '',
    author_profile     VARCHAR(1000) NOT NULL DEFAULT '',
    author_external_id VARCHAR(200) NOT NULL DEFAULT '',
    content            TEXT NOT NULL DEFAULT '',
    sentiment          VARCHAR(20) NOT NULL DEFAULT 'neutral' CHECK (sentiment IN ('positive','neutral','negative','unknown')),
    intent             VARCHAR(20) NOT NULL DEFAULT 'other' CHECK (intent IN ('question','lead','praise','complaint','spam','other')),
    score              NUMERIC(5,2) NOT NULL DEFAULT 0,
    status             VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new','drafted','replied','ignored','escalated')),
    assigned_to        VARCHAR(100) NOT NULL DEFAULT '',
    received_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_engagements_updated BEFORE UPDATE ON engagements
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_engagements_status ON engagements(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_engagements_author ON engagements(author_external_id);

CREATE TABLE IF NOT EXISTS engagement_replies (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id UUID NOT NULL REFERENCES engagements(id) ON DELETE CASCADE,
    content       TEXT NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','sent','failed')),
    sent_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_engagement_replies_eng ON engagement_replies(engagement_id);

CREATE TABLE IF NOT EXISTS leads (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    channel_id    UUID REFERENCES channels(id) ON DELETE SET NULL,
    engagement_id UUID REFERENCES engagements(id) ON DELETE SET NULL,
    name          VARCHAR(200) NOT NULL DEFAULT '',
    title         VARCHAR(200) NOT NULL DEFAULT '',
    company       VARCHAR(300) NOT NULL DEFAULT '',
    email         VARCHAR(300) NOT NULL DEFAULT '',
    profile_url   VARCHAR(1000) NOT NULL DEFAULT '',
    score         NUMERIC(5,2) NOT NULL DEFAULT 0,
    stage         VARCHAR(30) NOT NULL DEFAULT 'new' CHECK (stage IN ('new','qualified','nurturing','handoff','disqualified')),
    owner         VARCHAR(100) NOT NULL DEFAULT '',
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_leads_updated BEFORE UPDATE ON leads
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 10. Human-in-the-loop 审批
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS approvals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    run_id          UUID REFERENCES task_runs(id) ON DELETE SET NULL,
    workflow_run_id UUID REFERENCES workflow_runs(id) ON DELETE SET NULL,
    kind            VARCHAR(60) NOT NULL,                 -- publish_linkedin / send_message / delete_content
    subject_type    VARCHAR(60) NOT NULL DEFAULT '',
    subject_id      UUID,
    summary         TEXT        NOT NULL DEFAULT '',
    payload         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    risk            VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (risk IN ('low','medium','high')),
    confidence      NUMERIC(4,3) NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','expired','cancelled')),
    requested_by    VARCHAR(100) NOT NULL DEFAULT '',
    decided_by      VARCHAR(100) NOT NULL DEFAULT '',
    decision_note   TEXT        NOT NULL DEFAULT '',
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at      TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_approvals_updated BEFORE UPDATE ON approvals
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX IF NOT EXISTS idx_approvals_status ON approvals(status, created_at DESC);

-- ---------------------------------------------------------------------------
-- 11. 调度 / 事件 / 审计
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schedules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name         VARCHAR(200) NOT NULL,
    target_type  VARCHAR(20) NOT NULL CHECK (target_type IN ('workflow','task')),
    target_id    UUID,
    target_key   VARCHAR(200) NOT NULL DEFAULT '',
    cron         VARCHAR(120) NOT NULL,
    timezone     VARCHAR(64)  NOT NULL DEFAULT 'Asia/Taipei',
    enabled      BOOLEAN      NOT NULL DEFAULT true,
    next_run_at  TIMESTAMPTZ,
    last_run_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_schedules_updated BEFORE UPDATE ON schedules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS events (
    id             BIGSERIAL PRIMARY KEY,
    workspace_id   UUID,
    type           VARCHAR(120) NOT NULL,
    aggregate_type VARCHAR(60)  NOT NULL DEFAULT '',
    aggregate_id   VARCHAR(200) NOT NULL DEFAULT '',
    payload        JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_aggregate ON events(aggregate_type, aggregate_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    workspace_id UUID,
    actor_type  VARCHAR(30) NOT NULL DEFAULT 'system' CHECK (actor_type IN ('user','agent','system','api')),
    actor_id    VARCHAR(200) NOT NULL DEFAULT '',
    action      VARCHAR(120) NOT NULL,
    target_type VARCHAR(60)  NOT NULL DEFAULT '',
    target_id   VARCHAR(200) NOT NULL DEFAULT '',
    before      JSONB,
    after       JSONB,
    ip          VARCHAR(64)  NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action, created_at DESC);

-- ---------------------------------------------------------------------------
-- 12. 视图：内容表现（供 Analytics Agent / 后台查询）
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_content_performance AS
SELECT
    c.workspace_id,
    c.id            AS content_id,
    c.title,
    c.channel_kind,
    c.language,
    c.status,
    t.id            AS topic_id,
    t.category      AS topic_category,
    p.id            AS publication_id,
    p.published_at,
    COALESCE(m.impressions, 0)     AS impressions,
    COALESCE(m.reach, 0)           AS reach,
    COALESCE(m.likes, 0)           AS likes,
    COALESCE(m.comments, 0)        AS comments,
    COALESCE(m.shares, 0)          AS shares,
    COALESCE(m.clicks, 0)          AS clicks,
    COALESCE(m.engagement_rate, 0) AS engagement_rate,
    COALESCE(m.ctr, 0)             AS ctr
FROM content c
LEFT JOIN topics t ON t.id = c.topic_id
LEFT JOIN publications p ON p.content_id = c.id
LEFT JOIN LATERAL (
    SELECT * FROM content_metrics cm
    WHERE cm.publication_id = p.id
    ORDER BY cm.measured_at DESC LIMIT 1
) m ON true;

-- ============================================================================
-- pgvector 升级说明（安装 pgvector 后执行）：
--   ALTER TABLE knowledge_chunks ALTER COLUMN embedding TYPE vector(1536) USING NULL;
--   ALTER TABLE memories        ALTER COLUMN embedding TYPE vector(1536) USING NULL;
--   CREATE INDEX ON knowledge_chunks USING hnsw (embedding vector_cosine_ops);
--   CREATE INDEX ON memories        USING hnsw (embedding vector_cosine_ops);
-- 说明：V1 统一用 JSONB 存 embedding，保证在没有 pgvector 的 PostgreSQL
--       也能直接建库；接入向量检索时再执行上面的 ALTER。
-- ============================================================================

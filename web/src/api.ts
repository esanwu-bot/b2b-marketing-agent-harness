import axios from 'axios';

const baseURL = (import.meta.env.VITE_API_BASE as string) || '/api/v1';
const http = axios.create({ baseURL, timeout: 15000 });

export interface Metrics {
  active_runs: number;
  total_runs: number;
  tasks_completed: number;
  content_ready: number;
  llm_cost: number;
  documents: number;
  topics: number;
  publications: number;
  approvals_pending: number;
  content_by_status: Record<string, number>;
  agent_performance: Record<string, number>;
  workspace: string;
}

export interface AgentSpec {
  id: string;
  name: string;
  role: string;
  description: string;
  system_prompt: string;
  tools: string[];
  planner: string;
}

export interface ToolInfo {
  name: string;
  description: string;
  risk: string;
  requires_approval: boolean;
}

export interface WorkflowStepDef {
  id: string;
  kind: string;
  agent?: string;
  task?: string;
  channel?: string;
  action?: string;
}
export interface WorkflowDef {
  key: string;
  name: string;
  description: string;
  trigger?: Record<string, unknown>;
  steps: WorkflowStepDef[];
}

export interface RunStep {
  seq: number;
  kind: string;
  name: string;
  input?: Record<string, unknown>;
  output?: Record<string, unknown>;
  status: string;
  latency_ms: number;
  at: string;
}
export interface Run {
  id: string;
  task_id: string;
  agent_id: string;
  status: string;
  steps?: RunStep[];
  prompt_tokens: number;
  started_at: string;
  finished_at?: string;
}

export interface WorkflowStepRun {
  seq: number;
  kind: string;
  ref: string;
  status: string;
  task_id?: string;
  run_id?: string;
  output?: Record<string, unknown>;
  at: string;
}
export interface WorkflowRun {
  id: string;
  workflow_key: string;
  status: string;
  trigger_type: string;
  steps?: WorkflowStepRun[];
  started_at: string;
  finished_at?: string;
}

export interface Task {
  id: string;
  agent_id: string;
  type: string;
  goal: string;
  status: string;
  created_at: string;
}

export interface Approval {
  id: string;
  kind: string;
  summary: string;
  risk: string;
  status: string;
  requested_by: string;
  requested_at: string;
}

export interface Document {
  id: string;
  title: string;
  url: string;
  summary: string;
  content: string;
  source_type: string;
}
export interface Topic {
  id: string;
  title: string;
  category: string;
  score: number;
  status: string;
}
export interface Content {
  id: string;
  title: string;
  body: string;
  status: string;
  channel_kind: string;
  language: string;
}
export interface Publication {
  id: string;
  content_id: string;
  status: string;
  external_id: string;
  external_url: string;
}
export interface Event {
  type: string;
  aggregate_type: string;
  aggregate_id: string;
  payload: Record<string, unknown>;
  at: string;
}
export interface Health {
  status: string;
  store: string;
  llm: string;
  workspace: string;
  auto_approve: boolean;
}

// ---------- mock 兜底数据（后端未启动时也能渲染界面） ----------
const MOCK: {
  metrics: Metrics;
  agents: AgentSpec[];
  tools: ToolInfo[];
  workflows: WorkflowDef[];
  runs: Run[];
  workflowRuns: WorkflowRun[];
  tasks: Task[];
  approvals: Approval[];
  documents: Document[];
  topics: Topic[];
  content: Content[];
  publications: Publication[];
  events: Event[];
} = {
  metrics: {
    active_runs: 8, total_runs: 42, tasks_completed: 126, content_ready: 24, llm_cost: 18.42,
    documents: 12842, topics: 16, publications: 87, approvals_pending: 4,
    content_by_status: { draft: 18, in_review: 6, approved: 12, published: 87 },
    agent_performance: { research: 96.8, content: 91.4, review: 98.2, analytics: 94.7 },
    workspace: 'Microchip Taiwan Demo',
  },
  agents: [
    { id: 'research', name: '研究 Agent', role: 'Research', description: '采集行业新闻与竞品动态并入库', system_prompt: '', tools: ['crawler.search', 'knowledge.store_batch'], planner: 'rule' },
    { id: 'trend', name: '趋势 Agent', role: 'Trend', description: '主题聚类，产出选题候选', system_prompt: '', tools: ['trend.analyze'], planner: 'rule' },
    { id: 'strategy', name: '策略 Agent', role: 'Strategy', description: '选题打分与选择', system_prompt: '', tools: ['strategy.select'], planner: 'rule' },
    { id: 'content', name: '内容 Agent', role: 'Content', description: '生成多渠道内容草稿', system_prompt: '', tools: ['content.generate', 'content.save'], planner: 'rule' },
    { id: 'review', name: '审核 Agent', role: 'Review', description: '技术事实与品牌口吻审核', system_prompt: '', tools: ['review.check'], planner: 'rule' },
    { id: 'analytics', name: '分析 Agent', role: 'Analytics', description: '内容表现分析与优化建议', system_prompt: '', tools: ['analytics.insights'], planner: 'rule' },
  ],
  tools: [
    { name: 'crawler.search', description: '检索行业信息', risk: 'read', requires_approval: false },
    { name: 'content.generate', description: '生成渠道内容', risk: 'read', requires_approval: false },
    { name: 'knowledge.store', description: '写入知识库', risk: 'write', requires_approval: false },
    { name: 'linkedin.create_post', description: '发布到 LinkedIn 公司主页', risk: 'external', requires_approval: true },
    { name: 'notification.send', description: '发送通知', risk: 'external', requires_approval: false },
  ],
  workflows: [
    {
      key: 'daily_semiconductor_intelligence', name: '每日半导体情报与内容流水线',
      description: '采集 → 选题 → 内容 → 审核 → 审批 → 发布 → 分析',
      trigger: { type: 'cron', schedule: '0 8 * * *' },
      steps: [
        { id: 'research', kind: 'agent', agent: 'research', task: '采集今日行业与竞品情报' },
        { id: 'trend', kind: 'agent', agent: 'trend', task: '主题聚类' },
        { id: 'strategy', kind: 'agent', agent: 'strategy', task: '选题打分' },
        { id: 'content', kind: 'agent', agent: 'content', task: '生成内容草稿' },
        { id: 'review', kind: 'agent', agent: 'review', task: '技术审核' },
        { id: 'approval', kind: 'approval' },
        { id: 'publish', kind: 'channel', channel: 'linkedin', action: 'publish' },
        { id: 'analytics', kind: 'agent', agent: 'analytics', task: '数据分析' },
      ],
    },
  ],
  runs: [
    { id: 'run-1', task_id: 't1', agent_id: 'research', status: 'running', prompt_tokens: 3821, started_at: new Date().toISOString() },
    { id: 'run-2', task_id: 't2', agent_id: 'content', status: 'awaiting_approval', prompt_tokens: 2100, started_at: new Date().toISOString() },
    { id: 'run-3', task_id: 't3', agent_id: 'trend', status: 'succeeded', prompt_tokens: 1500, started_at: new Date().toISOString() },
    { id: 'run-4', task_id: 't4', agent_id: 'analytics', status: 'succeeded', prompt_tokens: 900, started_at: new Date().toISOString() },
  ],
  workflowRuns: [],
  tasks: [
    { id: 'task-1', agent_id: 'research', type: 'collect', goal: '采集今日情报', status: 'succeeded', created_at: new Date().toISOString() },
    { id: 'task-2', agent_id: 'content', type: 'generate', goal: '生成 LinkedIn 草稿', status: 'running', created_at: new Date().toISOString() },
  ],
  approvals: [
    { id: 'apr-1', kind: 'linkedin.create_post', summary: 'Edge AI：5 个工程师需要关注的趋势', risk: 'external', status: 'pending', requested_by: 'content', requested_at: new Date().toISOString() },
    { id: 'apr-2', kind: 'linkedin.create_post', summary: '10BASE-T1S 对比 100BASE-T1', risk: 'external', status: 'pending', requested_by: 'content', requested_at: new Date().toISOString() },
  ],
  documents: [
    { id: 'd1', title: '边缘 AI 在工业视觉检测中的落地加速', url: 'https://example.com/1', summary: '推理下沉到设备端，低功耗与实时性要求提升。', content: '', source_type: 'crawl' },
    { id: 'd2', title: '10BASE-T1S 车载以太网进入规模量产', url: 'https://example.com/2', summary: '单对以太网成本优势显现。', content: '', source_type: 'crawl' },
  ],
  topics: [
    { id: 'tp1', title: '边缘 AI 与 TinyML 的工程化落地', category: '边缘AI', score: 80, status: 'approved' },
    { id: 'tp2', title: '车载以太网与软件定义汽车', category: '车载网络', score: 80, status: 'candidate' },
    { id: 'tp3', title: '低功耗电源与能效管理', category: '电源管理', score: 60, status: 'candidate' },
  ],
  content: [
    { id: 'c1', title: '技术观察：边缘 AI 的工程化落地', body: '【技术观察】……', status: 'published', channel_kind: 'linkedin', language: 'zh-CN' },
    { id: 'c2', title: '技术观察：车载以太网', body: '【技术观察】……', status: 'in_review', channel_kind: 'linkedin', language: 'zh-CN' },
  ],
  publications: [
    { id: 'p1', content_id: 'c1', status: 'published', external_id: 'li-001', external_url: 'https://www.linkedin.com/feed/update/mock' },
  ],
  events: [
    { type: 'workflow.run.succeeded', aggregate_type: 'workflow_run', aggregate_id: 'wfr-1', payload: {}, at: new Date().toISOString() },
    { type: 'agent.run.succeeded', aggregate_type: 'run', aggregate_id: 'run-1', payload: { agent: 'research' }, at: new Date().toISOString() },
  ],
};

export let backendOnline = false;

async function safe<T>(p: Promise<{ data: T }>, fallback: T): Promise<T> {
  try {
    const r = await p;
    backendOnline = true;
    return r.data;
  } catch {
    backendOnline = false;
    return fallback;
  }
}

const get = <T>(url: string, fallback: T, params?: Record<string, unknown>) =>
  safe(http.get<T>(url, { params }), fallback);

export const api = {
  health: () => get<Health>('/health', { status: 'offline', store: 'mock', llm: 'mock', workspace: MOCK.metrics.workspace, auto_approve: true }),
  metrics: () => get<Metrics>('/metrics', MOCK.metrics),
  agents: async (): Promise<AgentSpec[]> => (await get<{ agents: AgentSpec[] }>('/agents', { agents: MOCK.agents })).agents ?? MOCK.agents,
  tools: async (): Promise<ToolInfo[]> => (await get<{ tools: ToolInfo[] }>('/tools', { tools: MOCK.tools })).tools ?? MOCK.tools,
  workflows: async (): Promise<WorkflowDef[]> => (await get<{ workflows: WorkflowDef[] }>('/workflows', { workflows: MOCK.workflows })).workflows ?? MOCK.workflows,
  runs: async (): Promise<Run[]> => (await get<{ runs: Run[] }>('/runs', { runs: MOCK.runs })).runs ?? MOCK.runs,
  workflowRuns: async (): Promise<WorkflowRun[]> => (await get<{ workflow_runs: WorkflowRun[] }>('/workflow-runs', { workflow_runs: MOCK.workflowRuns })).workflow_runs ?? [],
  tasks: async (): Promise<Task[]> => (await get<{ tasks: Task[] }>('/tasks', { tasks: MOCK.tasks })).tasks ?? MOCK.tasks,
  approvals: async (status?: string): Promise<Approval[]> =>
    (await get<{ approvals: Approval[] }>('/approvals', { approvals: MOCK.approvals }, status ? { status } : undefined)).approvals ?? MOCK.approvals,
  knowledge: async (): Promise<{ documents: Document[]; total: number }> =>
    get<{ documents: Document[]; total: number }>('/knowledge', { documents: MOCK.documents, total: MOCK.documents.length }),
  searchKnowledge: async (q: string) =>
    (await get<{ hits: { title: string; snippet: string; score: number }[] }>('/knowledge/search', { hits: [] }, { q })).hits ?? [],
  topics: async (): Promise<Topic[]> => (await get<{ topics: Topic[] }>('/topics', { topics: MOCK.topics })).topics ?? MOCK.topics,
  content: async (): Promise<Content[]> => (await get<{ content: Content[] }>('/content', { content: MOCK.content })).content ?? MOCK.content,
  publications: async (): Promise<Publication[]> =>
    (await get<{ publications: Publication[] }>('/publications', { publications: MOCK.publications })).publications ?? MOCK.publications,
  events: async (): Promise<Event[]> => (await get<{ events: Event[] }>('/events', { events: MOCK.events })).events ?? MOCK.events,
  runWorkflow: (key: string) => http.post(`/workflows/${key}/run`).then((r) => r.data),
  runAgent: (id: string, goal: string) => http.post(`/agents/${id}/run`, { goal }).then((r) => r.data),
  approve: (id: string) => http.post(`/approvals/${id}/approve`).then((r) => r.data),
  reject: (id: string) => http.post(`/approvals/${id}/reject`).then((r) => r.data),
};

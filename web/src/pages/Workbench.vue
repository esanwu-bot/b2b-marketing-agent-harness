<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>Agent 工作台</h1>
        <p>研究 → 情报 → 内容 → 审批 → 发布 → 分析</p>
      </div>
      <a-space>
        <a-button @click="load">刷新数据</a-button>
        <a-button type="primary" :loading="running" @click="runPipeline">＋ 新建运行</a-button>
      </a-space>
    </div>

    <div class="cw-grid-4">
      <div class="cw-stat">
        <div class="top"><span>活跃运行</span><span>◉</span></div>
        <div class="value">{{ metrics.active_runs }}</div>
        <div class="trend">运行中 / 待审批</div>
      </div>
      <div class="cw-stat">
        <div class="top"><span>已完成任务</span><span>✓</span></div>
        <div class="value">{{ metrics.tasks_completed }}</div>
        <div class="trend">累计成功任务</div>
      </div>
      <div class="cw-stat">
        <div class="top"><span>待审内容</span><span>✦</span></div>
        <div class="value">{{ metrics.content_ready }}</div>
        <div class="trend">{{ metrics.approvals_pending }} 条待审批</div>
      </div>
      <div class="cw-stat">
        <div class="top"><span>LLM 成本</span><span>◫</span></div>
        <div class="value">${{ metrics.llm_cost.toFixed(2) }}</div>
        <div class="trend">当日预算内</div>
      </div>
    </div>

    <div class="layout">
      <div>
        <div class="cw-card">
          <div class="card-head">
            <h3>Agent 运行</h3>
            <router-link to="/observability"><a-button size="small">查看全部 →</a-button></router-link>
          </div>
          <div class="card-body">
            <div v-for="r in runs.slice(0, 6)" :key="r.id" class="run">
              <div class="run-icon">{{ agentIcon(r.agent_id) }}</div>
              <div class="run-main">
                <div class="run-title">{{ agentName(r.agent_id) }}</div>
                <div class="run-meta">{{ r.id }} · tokens {{ r.prompt_tokens }}</div>
              </div>
              <a-tag :color="runColor(r.status)">{{ runText(r.status) }}</a-tag>
            </div>
            <a-empty v-if="!runs.length" description="暂无运行记录" />
          </div>
        </div>

        <div class="cw-card section">
          <div class="card-head"><h3>内容流水线</h3><router-link to="/content"><a-button size="small">打开工作室 →</a-button></router-link></div>
          <div class="card-body">
            <div class="pipeline">
              <div v-for="s in pipeline" :key="s.label">
                <div class="pipe-label">{{ s.label }}</div>
                <b class="pipe-value">{{ s.value }}</b>
                <a-progress :percent="s.percent" :show-info="false" size="small" />
              </div>
            </div>
          </div>
        </div>

        <div class="cw-card section">
          <div class="card-head"><h3>内容日历</h3><router-link to="/content"><a-button size="small">＋ 排期</a-button></router-link></div>
          <div class="card-body">
            <div class="calendar">
              <div v-for="d in 14" :key="d" class="day" :class="{ muted: d > 12 }">
                <strong>{{ d }}</strong>
                <div v-if="calendarPill(d)" class="pill" :class="calendarPill(d)!.type">{{ calendarPill(d)!.text }}</div>
              </div>
            </div>
            <div class="footer-note">九月 · 已排期 {{ publications.length }} 条发布</div>
          </div>
        </div>
      </div>

      <div>
        <div class="cw-card">
          <div class="card-head"><h3>实时 Agent 轨迹</h3><a-tag color="processing">● 实时</a-tag></div>
          <div class="card-body">
            <a-empty v-if="!trace.length" description="暂无轨迹" />
            <div class="cw-timeline">
              <div v-for="(t, i) in trace" :key="i" class="cw-event">
                <span class="cw-dot"></span>
                <h4>{{ t.title }}</h4>
                <p>{{ t.desc }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="cw-card section">
          <div class="card-head"><h3>数据源</h3><router-link to="/tasks"><a-button size="small">管理</a-button></router-link></div>
          <div class="card-body">
            <div v-for="s in sources" :key="s.name" class="source">
              <div class="source-logo">{{ s.logo }}</div>
              <div class="source-main"><b>{{ s.name }}</b><small>{{ s.desc }}</small></div>
              <span class="dot-online"></span>
            </div>
          </div>
        </div>

        <div class="cw-card section">
          <div class="card-head"><h3>审批中心</h3><a-tag color="warning">{{ metrics.approvals_pending }} 条待办</a-tag></div>
          <div class="card-body">
            <a-empty v-if="!approvals.length" description="暂无待审批" />
            <div v-for="a in approvals.slice(0, 4)" :key="a.id" class="approval-row">
              <b>{{ a.summary }}</b>
              <div class="approval-meta">{{ a.kind }} · {{ a.requested_by }}</div>
            </div>
            <router-link to="/approvals"><a-button type="primary" block style="margin-top: 10px">前往审批</a-button></router-link>
          </div>
        </div>
      </div>
    </div>

    <div class="two section">
      <div class="cw-card">
        <div class="card-head"><h3>Agent 表现</h3><router-link to="/analytics"><a-button size="small">分析 →</a-button></router-link></div>
        <div class="card-body">
          <div v-for="(v, k) in agentPerf" :key="k" class="perf-row">
            <div class="perf-head"><span>{{ agentName(String(k)) }}</span><b>{{ v.toFixed(1) }}%</b></div>
            <a-progress :percent="Number(v.toFixed(1))" :show-info="false" size="small" />
          </div>
          <a-empty v-if="!Object.keys(agentPerf).length" description="暂无数据" />
        </div>
      </div>
      <div class="cw-card">
        <div class="card-head"><h3>知识库增长</h3><router-link to="/knowledge"><a-button size="small">知识库 →</a-button></router-link></div>
        <div class="card-body">
          <div class="kb-head"><span>文档总数</span><b>{{ metrics.documents }}</b></div>
          <a-progress :percent="Math.min(100, Math.round((metrics.documents / 15000) * 100))" :show-info="false" />
          <div class="kb-grid">
            <div><b>{{ topics.length }}</b><small>选题</small></div>
            <div><b>{{ publications.length }}</b><small>发布</small></div>
            <div><b>{{ content.length }}</b><small>内容</small></div>
          </div>
        </div>
      </div>
    </div>

    <div class="footer-note">B2B 营销 Agent Harness · PostgreSQL + pgvector · Redis · crawl-worker-redis 采集适配</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { message } from 'ant-design-vue';
import { api, type Approval, type Content, type Metrics, type Publication, type Run, type Topic, type WorkflowRun } from '@/api';

const metrics = ref<Metrics>({
  active_runs: 0, total_runs: 0, tasks_completed: 0, content_ready: 0, llm_cost: 0,
  documents: 0, topics: 0, publications: 0, approvals_pending: 0, content_by_status: {}, agent_performance: {}, workspace: '',
});
const runs = ref<Run[]>([]);
const approvals = ref<Approval[]>([]);
const topics = ref<Topic[]>([]);
const content = ref<Content[]>([]);
const publications = ref<Publication[]>([]);
const workflowRuns = ref<WorkflowRun[]>([]);
const running = ref(false);

const agentPerf = computed(() => {
  if (Object.keys(metrics.value.agent_performance || {}).length) return metrics.value.agent_performance;
  // 兜底：由运行成功率推导
  const map: Record<string, number> = {};
  const total: Record<string, number> = {};
  for (const r of runs.value) {
    total[r.agent_id] = (total[r.agent_id] || 0) + 1;
    if (r.status === 'succeeded') map[r.agent_id] = (map[r.agent_id] || 0) + 1;
  }
  const out: Record<string, number> = {};
  for (const k of Object.keys(total)) out[k] = (map[k] || 0) / total[k] * 100;
  return out;
});

const pipeline = computed(() => {
  const s = metrics.value.content_by_status || {};
  const draft = s.draft || 0;
  const review = s.in_review || 0;
  const approved = s.approved || 0;
  const published = s.published || 0;
  return [
    { label: '选题', value: metrics.value.topics, percent: Math.min(100, metrics.value.topics * 5) },
    { label: '草稿', value: draft, percent: Math.min(100, draft * 5) },
    { label: '审核中', value: review, percent: Math.min(100, review * 8) },
    { label: '已排期', value: approved, percent: Math.min(100, approved * 8) },
    { label: '已发布', value: published, percent: Math.min(100, published) },
  ];
});

const trace = computed(() => {
  const items: { title: string; desc: string }[] = [];
  const wr = workflowRuns.value[0];
  if (wr?.steps?.length) {
    for (const st of wr.steps.slice(-6)) {
      items.push({ title: stepTitle(st.kind, st.ref), desc: `状态：${st.status}` });
    }
  } else {
    items.push({ title: '研究 Agent · 规划中', desc: '正在分析信息源，准备竞品监控。' });
    items.push({ title: '工具调用 · crawler.search', desc: '检索行业与竞品内容。' });
    items.push({ title: '观察 · 知识入库', desc: '情报写入知识库。' });
    items.push({ title: '下一步 · 趋势 Agent', desc: '进行主题聚类与内容机会评分…' });
  }
  return items;
});

const sources = [
  { name: 'Microchip LinkedIn', logo: 'M', desc: '最新帖子 · 企业主页' },
  { name: 'EE Times', logo: 'EE', desc: '行业新闻' },
  { name: '科技新报', logo: 'TW', desc: '行业动态' },
  { name: '竞品动态', logo: 'C', desc: '竞争对手内容监控' },
];

function agentName(id: string) {
  const map: Record<string, string> = { research: '研究 Agent', trend: '趋势 Agent', strategy: '策略 Agent', content: '内容 Agent', review: '审核 Agent', analytics: '分析 Agent' };
  return map[id] || id || '未知 Agent';
}
function agentIcon(id: string) {
  const map: Record<string, string> = { research: '⌁', trend: '◇', strategy: '◈', content: '✦', review: '✓', analytics: '↗' };
  return map[id] || '◌';
}
function runColor(s: string) {
  return s === 'running' ? 'processing' : s === 'succeeded' ? 'success' : s === 'awaiting_approval' ? 'warning' : 'error';
}
function runText(s: string) {
  const map: Record<string, string> = { running: '运行中', succeeded: '已完成', awaiting_approval: '待审批', failed: '失败', cancelled: '已取消' };
  return map[s] || s;
}
function stepTitle(kind: string, ref: string) {
  const map: Record<string, string> = { agent: 'Agent', tool: '工具调用', approval: '人工审批', channel: '渠道发布', plan: '规划', output: '产出' };
  return `${map[kind] || kind} · ${ref}`;
}
function calendarPill(d: number) {
  const pills = [
    { text: '边缘 AI 趋势', type: 'blue' }, { text: 'MCU 产品', type: 'green' },
    { text: '技术文章', type: 'purple' }, { text: '行业快讯', type: 'orange' },
  ];
  if (d <= 12) return pills[(d - 1) % pills.length];
  return null;
}

async function runPipeline() {
  running.value = true;
  try {
    await api.runWorkflow('daily_semiconductor_intelligence');
    message.success('已触发每日情报流水线');
    await new Promise((r) => setTimeout(r, 400));
    await load();
  } catch {
    message.error('触发失败（后端可能未启动）');
  } finally {
    running.value = false;
  }
}

async function load() {
  metrics.value = await api.metrics();
  const [r, a, t, c, p, wr] = await Promise.all([
    api.runs(), api.approvals('pending'), api.topics(), api.content(), api.publications(), api.workflowRuns(),
  ]);
  runs.value = r;
  approvals.value = a;
  topics.value = t;
  content.value = c;
  publications.value = p;
  workflowRuns.value = wr;
}

onMounted(load);
</script>

<style scoped>
.layout {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(340px, 0.85fr);
  gap: 16px;
}
.cw-card .card-head {
  padding: 14px 16px;
  border-bottom: 1px solid #e6e9ef;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.cw-card .card-head h3 {
  margin: 0;
  font-size: 14px;
}
.cw-card .card-body {
  padding: 14px 16px;
}
.section {
  margin-top: 16px;
}
.run {
  display: flex;
  gap: 12px;
  padding: 13px 0;
  border-bottom: 1px solid #eef0f4;
  align-items: center;
}
.run:last-child {
  border-bottom: 0;
}
.run-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: #eff6ff;
  color: #2563eb;
  display: grid;
  place-items: center;
}
.run-main {
  flex: 1;
  min-width: 0;
}
.run-title {
  font-weight: 620;
}
.run-meta {
  font-size: 12px;
  color: #7b8496;
}
.pipeline {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 10px;
}
.pipe-label {
  font-size: 12px;
  color: #7b8496;
}
.pipe-value {
  font-size: 20px;
  display: block;
  margin: 4px 0 6px;
}
.calendar {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 6px;
}
.day {
  min-height: 64px;
  border: 1px solid #edf0f4;
  border-radius: 7px;
  padding: 7px;
  background: #fff;
}
.day.muted {
  background: #fafbfc;
  color: #a4adbc;
}
.day strong {
  font-size: 11px;
}
.pill {
  font-size: 10px;
  padding: 4px 5px;
  border-radius: 5px;
  margin-top: 7px;
  background: #eff6ff;
  color: #1d4ed8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pill.green {
  background: #ecfdf3;
  color: #15803d;
}
.pill.purple {
  background: #f5f3ff;
  color: #6d28d9;
}
.pill.orange {
  background: #fff7ed;
  color: #c2410c;
}
.source {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid #eef0f4;
}
.source:last-child {
  border: 0;
}
.source-logo {
  width: 32px;
  height: 32px;
  border-radius: 7px;
  background: #f1f5f9;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #475569;
  font-size: 12px;
}
.source-main {
  flex: 1;
}
.source-main b {
  font-size: 13px;
}
.source-main small {
  display: block;
  color: #7b8496;
  margin-top: 3px;
}
.dot-online {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #22c55e;
}
.approval-row {
  padding: 9px 0;
  border-bottom: 1px solid #eef0f4;
}
.approval-meta {
  font-size: 11px;
  color: #7b8496;
  margin-top: 4px;
}
.two {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.perf-row {
  margin-bottom: 12px;
}
.perf-head {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 6px;
}
.kb-head {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}
.kb-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 16px;
}
.kb-grid small {
  display: block;
  color: #7b8496;
}
.footer-note {
  font-size: 11px;
  color: #9aa3b2;
  margin-top: 16px;
}
@media (max-width: 1100px) {
  .layout,
  .two {
    grid-template-columns: 1fr;
  }
}
</style>

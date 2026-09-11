<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>数据分析</h1>
        <p>内容表现回流，形成对下一轮选题与策略的反馈</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>

    <div class="cw-grid-4">
      <div class="cw-stat"><div class="top"><span>文档</span><span>◎</span></div><div class="value">{{ metrics.documents }}</div></div>
      <div class="cw-stat"><div class="top"><span>选题</span><span>✦</span></div><div class="value">{{ metrics.topics }}</div></div>
      <div class="cw-stat"><div class="top"><span>发布</span><span>↗</span></div><div class="value">{{ metrics.publications }}</div></div>
      <div class="cw-stat"><div class="top"><span>LLM 成本</span><span>◫</span></div><div class="value">${{ metrics.llm_cost.toFixed(2) }}</div></div>
    </div>

    <div class="two">
      <a-card :bordered="false" class="cw-card" title="Agent 表现">
        <div v-for="(v, k) in metrics.agent_performance" :key="k" class="perf-row">
          <div class="perf-head"><span>{{ agentName(String(k)) }}</span><b>{{ v.toFixed(1) }}%</b></div>
          <a-progress :percent="Number(v.toFixed(1))" :show-info="false" size="small" />
        </div>
        <a-empty v-if="!Object.keys(metrics.agent_performance).length" description="暂无数据" />
      </a-card>

      <a-card :bordered="false" class="cw-card" title="内容状态分布">
        <div v-for="(v, k) in metrics.content_by_status" :key="k" class="perf-row">
          <div class="perf-head"><span>{{ statusText(String(k)) }}</span><b>{{ v }}</b></div>
          <a-progress :percent="Math.min(100, Number(v) * 4)" :show-info="false" size="small" />
        </div>
        <a-empty v-if="!Object.keys(metrics.content_by_status).length" description="暂无数据" />
      </a-card>
    </div>

    <a-card :bordered="false" class="cw-card" title="已发布内容" style="margin-top: 16px">
      <a-table :columns="columns" :data-source="publications" row-key="id" size="small" :pagination="{ pageSize: 8 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'"><a-tag color="success">{{ record.status }}</a-tag></template>
          <template v-else-if="column.key === 'external_url'"><a :href="record.external_url" target="_blank">{{ record.external_url || '-' }}</a></template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type Metrics, type Publication } from '@/api';

const metrics = ref<Metrics>({
  active_runs: 0, total_runs: 0, tasks_completed: 0, content_ready: 0, llm_cost: 0,
  documents: 0, topics: 0, publications: 0, approvals_pending: 0, content_by_status: {}, agent_performance: {}, workspace: '',
});
const publications = ref<Publication[]>([]);
const columns: TableColumnsType = [
  { title: '发布 ID', dataIndex: 'id', key: 'id', width: 220 },
  { title: '内容 ID', dataIndex: 'content_id', key: 'content_id', width: 220 },
  { title: '状态', key: 'status', width: 100 },
  { title: '链接', key: 'external_url' },
];
function agentName(id: string) {
  const m: Record<string, string> = { research: '研究 Agent', trend: '趋势 Agent', strategy: '策略 Agent', content: '内容 Agent', review: '审核 Agent', analytics: '分析 Agent' };
  return m[id] || id;
}
function statusText(s: string) {
  const m: Record<string, string> = { draft: '草稿', in_review: '审核中', approved: '已通过', scheduled: '已排期', published: '已发布', rejected: '已拒绝', failed: '失败' };
  return m[s] || s;
}
async function load() {
  metrics.value = await api.metrics();
  publications.value = await api.publications();
}
onMounted(load);
</script>

<style scoped>
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
</style>

<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>可观测性</h1>
        <p>Agent 思考 / 工具调用 / LLM 成本 全量留痕，可回放</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="11">
        <a-card :bordered="false" class="cw-card" title="Agent 运行">
          <a-table
            :columns="columns"
            :data-source="runs"
            row-key="id"
            size="small"
            :pagination="{ pageSize: 10 }"
            :custom-row="rowProps"
            :row-class-name="rowClass"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'agent'">{{ agentName(record.agent_id) }}</template>
              <template v-else-if="column.key === 'status'"><a-tag :color="color(record.status)">{{ text(record.status) }}</a-tag></template>
            </template>
          </a-table>
        </a-card>
      </a-col>

      <a-col :xs="24" :lg="13">
        <a-card :bordered="false" class="cw-card" :title="selected ? `运行轨迹 · ${selected.id}` : '运行轨迹'">
          <a-empty v-if="!selected" description="选择左侧运行查看轨迹" />
          <div v-else class="cw-timeline">
            <div v-for="s in selected.steps || []" :key="s.seq" class="cw-event">
              <span class="cw-dot" :style="dotStyle(s.status)"></span>
              <h4>{{ stepTitle(s) }}</h4>
              <p>
                状态 {{ s.status }}
                <span v-if="s.latency_ms"> · 耗时 {{ s.latency_ms }}ms</span>
                <span v-if="s.output && (s.output as any).summary"> · {{ (s.output as any).summary }}</span>
              </p>
            </div>
            <a-empty v-if="!(selected.steps || []).length" description="该运行暂无步骤明细" />
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-card :bordered="false" class="cw-card" title="领域事件" style="margin-top: 16px">
      <a-table :columns="eventColumns" :data-source="events" row-key="at" size="small" :pagination="{ pageSize: 10 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'"><span class="cw-mono">{{ record.type }}</span></template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type Event, type Run, type RunStep } from '@/api';

const runs = ref<Run[]>([]);
const events = ref<Event[]>([]);
const selected = ref<Run | null>(null);

const columns: TableColumnsType = [
  { title: '运行 ID', dataIndex: 'id', key: 'id', width: 210 },
  { title: 'Agent', key: 'agent', width: 110 },
  { title: '状态', key: 'status', width: 100 },
  { title: '步骤', key: 'steps', width: 70, customRender: ({ record }: any) => record.steps?.length || 0 },
];
const eventColumns: TableColumnsType = [
  { title: '事件类型', key: 'type', width: 260 },
  { title: '聚合', dataIndex: 'aggregate_id', key: 'aggregate_id', width: 240 },
  { title: '时间', dataIndex: 'at', key: 'at' },
];

function rowProps(record: Run) {
  return { onClick: () => (selected.value = record), style: { cursor: 'pointer' } };
}
function rowClass(record: Run) {
  return selected.value?.id === record.id ? 'selected-row' : '';
}
function agentName(id: string) {
  const m: Record<string, string> = { research: '研究 Agent', trend: '趋势 Agent', strategy: '策略 Agent', content: '内容 Agent', review: '审核 Agent', analytics: '分析 Agent' };
  return m[id] || id;
}
function color(s: string) {
  return s === 'succeeded' ? 'success' : s === 'failed' ? 'error' : s === 'awaiting_approval' ? 'warning' : 'processing';
}
function text(s: string) {
  const m: Record<string, string> = { succeeded: '已完成', failed: '失败', running: '运行中', awaiting_approval: '待审批', cancelled: '已取消' };
  return m[s] || s;
}
function stepTitle(s: RunStep) {
  const m: Record<string, string> = { plan: '规划', tool: '工具调用', output: '产出', approval: '审批', error: '错误', reflect: '反思', observe: '观察', llm: '模型调用' };
  return `${m[s.kind] || s.kind} · ${s.name}`;
}
function dotStyle(status: string) {
  if (status === 'failed') return { background: '#dc2626', borderColor: '#fee2e2' };
  if (status === 'awaiting_approval') return { background: '#f59e0b', borderColor: '#fef3c7' };
  return {};
}
async function load() {
  runs.value = await api.runs();
  events.value = await api.events();
  if (!selected.value && runs.value.length) selected.value = runs.value[0];
}
onMounted(load);
</script>

<style scoped>
:deep(.selected-row) {
  background: #eff6ff;
}
</style>

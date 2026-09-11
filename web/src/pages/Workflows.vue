<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>工作流</h1>
        <p>决定「什么时候做什么」：Agent / 审批 / 渠道发布 的编排</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>

    <a-card v-for="w in workflows" :key="w.key" :bordered="false" class="cw-card" style="margin-bottom: 16px">
      <template #title>{{ w.name }}</template>
      <template #extra>
        <a-button type="primary" size="small" @click="run(w.key)">运行工作流</a-button>
      </template>
      <p style="color: #7b8496">{{ w.description }}</p>
      <a-steps :current="-1" size="small" style="margin: 12px 0 16px">
        <a-step v-for="s in w.steps" :key="s.id" :title="stepLabel(s)" />
      </a-steps>
    </a-card>

    <a-card :bordered="false" class="cw-card" title="最近工作流运行">
      <a-table :columns="columns" :data-source="runs" row-key="id" size="small" :pagination="{ pageSize: 8 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'"><a-tag :color="color(record.status)">{{ text(record.status) }}</a-tag></template>
          <template v-else-if="column.key === 'steps'">{{ record.steps?.length || 0 }} 步</template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { message } from 'ant-design-vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type WorkflowDef, type WorkflowStepDef, type WorkflowRun } from '@/api';

const workflows = ref<WorkflowDef[]>([]);
const runs = ref<WorkflowRun[]>([]);
const columns: TableColumnsType = [
  { title: '运行 ID', dataIndex: 'id', key: 'id', width: 220 },
  { title: '触发方式', dataIndex: 'trigger_type', key: 'trigger_type', width: 110 },
  { title: '状态', key: 'status', width: 110 },
  { title: '步骤', key: 'steps', width: 90 },
  { title: '开始时间', dataIndex: 'started_at', key: 'started_at' },
];

function stepLabel(s: WorkflowStepDef) {
  if (s.kind === 'agent') return `Agent: ${s.agent}`;
  if (s.kind === 'approval') return '人工审批';
  if (s.kind === 'channel') return `发布: ${s.channel}`;
  return s.id;
}
function color(s: string) {
  return s === 'succeeded' ? 'success' : s === 'failed' ? 'error' : s === 'awaiting_approval' ? 'warning' : 'processing';
}
function text(s: string) {
  const m: Record<string, string> = { succeeded: '已完成', failed: '失败', running: '运行中', awaiting_approval: '待审批' };
  return m[s] || s;
}
async function run(key: string) {
  try {
    await api.runWorkflow(key);
    message.success('已触发工作流');
    await load();
  } catch {
    message.error('触发失败（后端可能未启动）');
  }
}
async function load() {
  workflows.value = await api.workflows();
  runs.value = await api.workflowRuns();
}
onMounted(load);
</script>

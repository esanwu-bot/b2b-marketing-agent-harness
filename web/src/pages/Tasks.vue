<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>采集任务</h1>
        <p>由 crawl-worker-redis 采集层提供数据；这里展示 Agent 触发的采集任务</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>
    <a-card :bordered="false" class="cw-card">
      <a-table :columns="columns" :data-source="tasks" row-key="id" size="middle" :pagination="{ pageSize: 15 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'"><a-tag :color="color(record.status)">{{ text(record.status) }}</a-tag></template>
          <template v-else-if="column.key === 'agent'">{{ agentName(record.agent_id) }}</template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type Task } from '@/api';

const tasks = ref<Task[]>([]);
const columns: TableColumnsType = [
  { title: '任务 ID', dataIndex: 'id', key: 'id', width: 220 },
  { title: 'Agent', key: 'agent', width: 130 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 130 },
  { title: '目标', dataIndex: 'goal', key: 'goal' },
  { title: '状态', key: 'status', width: 110 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 190 },
];
function color(s: string) {
  return s === 'succeeded' ? 'success' : s === 'failed' ? 'error' : s === 'running' ? 'processing' : 'default';
}
function text(s: string) {
  const m: Record<string, string> = { succeeded: '已完成', failed: '失败', running: '运行中', pending: '待执行', awaiting_approval: '待审批', queued: '排队中' };
  return m[s] || s;
}
function agentName(id: string) {
  const m: Record<string, string> = { research: '研究 Agent', trend: '趋势 Agent', strategy: '策略 Agent', content: '内容 Agent', review: '审核 Agent', analytics: '分析 Agent' };
  return m[id] || id || '-';
}
async function load() {
  tasks.value = await api.tasks();
}
onMounted(load);
</script>

<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>工具注册表</h1>
        <p>Agent 只能通过工具与外部世界交互；写/外部动作可配置审批</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>
    <a-card :bordered="false" class="cw-card">
      <a-table :columns="columns" :data-source="tools" row-key="name" size="middle" :pagination="{ pageSize: 20 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'"><span class="cw-mono">{{ record.name }}</span></template>
          <template v-else-if="column.key === 'risk'"><a-tag :color="riskColor(record.risk)">{{ record.risk }}</a-tag></template>
          <template v-else-if="column.key === 'approval'">
            <a-tag :color="record.requires_approval ? 'warning' : 'default'">{{ record.requires_approval ? '需审批' : '可直执行' }}</a-tag>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type ToolInfo } from '@/api';

const tools = ref<ToolInfo[]>([]);
const columns: TableColumnsType = [
  { title: '工具名', key: 'name', width: 240 },
  { title: '说明', dataIndex: 'description', key: 'description' },
  { title: '风险', key: 'risk', width: 100 },
  { title: '审批', key: 'approval', width: 120 },
];
function riskColor(r: string) {
  return r === 'external' ? 'red' : r === 'write' ? 'orange' : 'green';
}
async function load() {
  tools.value = await api.tools();
}
onMounted(load);
</script>

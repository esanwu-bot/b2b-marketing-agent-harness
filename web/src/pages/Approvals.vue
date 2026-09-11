<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>审批中心</h1>
        <p>Human-in-the-loop：发布、发送、删除等高风险动作需人工确认</p>
      </div>
      <a-space>
        <a-radio-group v-model:value="status" button-style="solid" @change="load">
          <a-radio-button value="pending">待审批</a-radio-button>
          <a-radio-button value="">全部</a-radio-button>
        </a-radio-group>
        <a-button @click="load">刷新</a-button>
      </a-space>
    </div>

    <a-card :bordered="false" class="cw-card">
      <a-alert v-if="!approvals.length" type="success" show-icon message="当前没有待审批事项" style="margin-bottom: 12px" />
      <a-table :columns="columns" :data-source="approvals" row-key="id" size="middle" :pagination="{ pageSize: 12 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'risk'"><a-tag :color="record.risk === 'external' ? 'red' : 'orange'">{{ record.risk }}</a-tag></template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'approved' ? 'success' : record.status === 'rejected' ? 'error' : 'warning'">{{ statusText(record.status) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space v-if="record.status === 'pending'">
              <a-button type="primary" size="small" @click="decide(record, 'approve')">通过</a-button>
              <a-button danger size="small" @click="decide(record, 'reject')">拒绝</a-button>
            </a-space>
            <span v-else style="color: #9aa3b2">已处理</span>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { message } from 'ant-design-vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type Approval } from '@/api';

const approvals = ref<Approval[]>([]);
const status = ref('pending');
const columns: TableColumnsType = [
  { title: '事项', dataIndex: 'summary', key: 'summary' },
  { title: '类型', dataIndex: 'kind', key: 'kind', width: 190 },
  { title: '发起方', dataIndex: 'requested_by', key: 'requested_by', width: 110 },
  { title: '风险', key: 'risk', width: 100 },
  { title: '状态', key: 'status', width: 100 },
  { title: '操作', key: 'actions', width: 170 },
];

function statusText(s: string) {
  const m: Record<string, string> = { pending: '待审批', approved: '已通过', rejected: '已拒绝', expired: '已过期' };
  return m[s] || s;
}
async function decide(a: Approval, action: 'approve' | 'reject') {
  try {
    if (action === 'approve') await api.approve(a.id);
    else await api.reject(a.id);
    message.success(action === 'approve' ? '已通过' : '已拒绝');
    await load();
  } catch {
    message.error('操作失败（后端可能未启动）');
  }
}
async function load() {
  approvals.value = await api.approvals(status.value || undefined);
}
onMounted(load);
</script>

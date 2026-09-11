<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>内容工作室</h1>
        <p>选题 → 草稿 → 审核 → 排期 → 发布，全部留痕</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>

    <a-card :bordered="false" class="cw-card">
      <a-tabs v-model:activeKey="tab">
        <a-tab-pane key="content" tab="内容草稿">
          <a-table :columns="columns" :data-source="content" row-key="id" size="middle" :pagination="{ pageSize: 10 }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'title'"><a @click="open(record)">{{ record.title }}</a></template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
              </template>
            </template>
          </a-table>
        </a-tab-pane>
        <a-tab-pane key="calendar" tab="内容日历">
          <div class="calendar">
            <div v-for="d in 30" :key="d" class="day">
              <strong>{{ d }}</strong>
              <div v-if="pill(d)" class="pill" :class="pill(d)!.type">{{ pill(d)!.text }}</div>
            </div>
          </div>
        </a-tab-pane>
        <a-tab-pane key="publications" tab="发布记录">
          <a-table :columns="pubColumns" :data-source="publications" row-key="id" size="small" :pagination="{ pageSize: 10 }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'"><a-tag color="success">{{ record.status }}</a-tag></template>
              <template v-else-if="column.key === 'url'"><a :href="record.external_url" target="_blank">{{ record.external_url || '-' }}</a></template>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <a-drawer v-model:open="drawer" :title="current?.title" width="680">
      <template v-if="current">
        <a-space style="margin-bottom: 12px">
          <a-tag :color="statusColor(current.status)">{{ statusText(current.status) }}</a-tag>
          <a-tag>{{ current.channel_kind }}</a-tag>
          <a-tag>{{ current.language }}</a-tag>
        </a-space>
        <div style="white-space: pre-wrap; line-height: 1.8">{{ current.body }}</div>
      </template>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type Content, type Publication } from '@/api';

const tab = ref('content');
const content = ref<Content[]>([]);
const publications = ref<Publication[]>([]);
const drawer = ref(false);
const current = ref<Content | null>(null);

const columns: TableColumnsType = [
  { title: '标题', key: 'title' },
  { title: '渠道', dataIndex: 'channel_kind', key: 'channel_kind', width: 110 },
  { title: '语言', dataIndex: 'language', key: 'language', width: 90 },
  { title: '状态', key: 'status', width: 110 },
];
const pubColumns: TableColumnsType = [
  { title: '发布 ID', dataIndex: 'id', key: 'id', width: 220 },
  { title: '状态', key: 'status', width: 100 },
  { title: '链接', key: 'url' },
];

function open(c: Content) {
  current.value = c;
  drawer.value = true;
}
function statusColor(s: string) {
  return s === 'published' ? 'success' : s === 'approved' ? 'cyan' : s === 'in_review' ? 'warning' : s === 'rejected' ? 'error' : 'default';
}
function statusText(s: string) {
  const m: Record<string, string> = { draft: '草稿', in_review: '审核中', approved: '已通过', scheduled: '已排期', published: '已发布', rejected: '已拒绝', failed: '失败' };
  return m[s] || s;
}
function pill(d: number) {
  const pills = [
    { text: '边缘 AI', type: 'blue' }, { text: 'MCU 产品', type: 'green' },
    { text: '技术文章', type: 'purple' }, { text: '行业快讯', type: 'orange' },
  ];
  return d % 3 === 0 ? pills[(d / 3) % pills.length | 0] : null;
}
async function load() {
  content.value = await api.content();
  publications.value = await api.publications();
}
onMounted(load);
</script>

<style scoped>
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
.pill.green { background: #ecfdf3; color: #15803d; }
.pill.purple { background: #f5f3ff; color: #6d28d9; }
.pill.orange { background: #fff7ed; color: #c2410c; }
</style>

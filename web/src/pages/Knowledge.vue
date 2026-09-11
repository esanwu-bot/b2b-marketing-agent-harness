<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>知识库</h1>
        <p>“世界上有什么知识”：文档 / 分块 / 实体（V1 使用 PostgreSQL）</p>
      </div>
      <a-input-search v-model:value="q" placeholder="搜索知识" style="width: 260px" @search="search" />
    </div>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="14">
        <a-card :bordered="false" class="cw-card" :title="`文档（共 ${total} 篇）`">
          <a-table :columns="columns" :data-source="documents" row-key="id" size="middle" :pagination="{ pageSize: 10 }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'title'">
                <a @click="open(record)">{{ record.title }}</a>
              </template>
              <template v-else-if="column.key === 'source_type'"><a-tag>{{ record.source_type }}</a-tag></template>
              <template v-else-if="column.key === 'url'">
                <a :href="record.url" target="_blank">{{ record.url }}</a>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>
      <a-col :xs="24" :lg="10">
        <a-card :bordered="false" class="cw-card" title="选题候选">
          <a-table :columns="topicColumns" :data-source="topics" row-key="id" size="small" :pagination="false">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'"><a-tag :color="record.status === 'approved' ? 'success' : 'default'">{{ record.status }}</a-tag></template>
              <template v-else-if="column.key === 'score'">{{ record.score.toFixed(0) }}</template>
            </template>
          </a-table>
        </a-card>
      </a-col>
    </a-row>

    <a-drawer v-model:open="drawer" :title="current?.title" width="640">
      <template v-if="current">
        <a-descriptions :column="1" size="small" bordered style="margin-bottom: 12px">
          <a-descriptions-item label="来源">{{ current.source_type }}</a-descriptions-item>
          <a-descriptions-item label="链接">{{ current.url || '-' }}</a-descriptions-item>
          <a-descriptions-item label="摘要">{{ current.summary || '-' }}</a-descriptions-item>
        </a-descriptions>
        <div class="cw-mono" style="font-size: 13px; line-height: 1.7; white-space: pre-wrap">{{ current.content }}</div>
      </template>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { TableColumnsType } from 'ant-design-vue';
import { api, type Document, type Topic } from '@/api';

const q = ref('');
const documents = ref<Document[]>([]);
const topics = ref<Topic[]>([]);
const total = ref(0);
const drawer = ref(false);
const current = ref<Document | null>(null);

const columns: TableColumnsType = [
  { title: '标题', key: 'title' },
  { title: '来源', key: 'source_type', width: 100 },
  { title: '链接', key: 'url', width: 220 },
];
const topicColumns: TableColumnsType = [
  { title: '选题', dataIndex: 'title', key: 'title' },
  { title: '分数', key: 'score', width: 70 },
  { title: '状态', key: 'status', width: 90 },
];

function open(d: Document) {
  current.value = d;
  drawer.value = true;
}
async function search() {
  if (!q.value) return load();
  const hits = await api.searchKnowledge(q.value);
  documents.value = hits.map((h, i) => ({ id: String(i), title: h.title, summary: h.snippet, url: '', content: h.snippet, source_type: '检索' }));
  total.value = hits.length;
}
async function load() {
  const res = await api.knowledge();
  documents.value = res.documents;
  total.value = res.total;
  topics.value = await api.topics();
}
onMounted(load);
</script>

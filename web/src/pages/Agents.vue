<template>
  <div class="cw-page">
    <div class="cw-page-head">
      <div>
        <h1>Agent 管理</h1>
        <p>内置 6 个参考 Agent，规格驱动，无需改运行时即可扩展</p>
      </div>
      <a-button @click="load">刷新</a-button>
    </div>
    <a-row :gutter="[16, 16]">
      <a-col v-for="a in agents" :key="a.id" :xs="24" :sm="12" :lg="8">
        <a-card :bordered="false" class="cw-card">
          <template #title>
            <a-space><span>{{ a.name }}</span><a-tag color="blue">{{ a.role }}</a-tag></a-space>
          </template>
          <p style="color: #7b8496; min-height: 40px">{{ a.description }}</p>
          <div style="margin-bottom: 10px">
            <a-tag v-for="t in a.tools" :key="t" class="cw-mono" style="margin-bottom: 4px">{{ t }}</a-tag>
          </div>
          <a-button type="primary" size="small" :loading="running === a.id" @click="run(a)">运行 Agent</a-button>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { message } from 'ant-design-vue';
import { api, type AgentSpec } from '@/api';

const agents = ref<AgentSpec[]>([]);
const running = ref('');

async function load() {
  agents.value = await api.agents();
}
async function run(a: AgentSpec) {
  running.value = a.id;
  try {
    await api.runAgent(a.id, `手动运行：${a.name}`);
    message.success(`${a.name} 运行完成`);
  } catch {
    message.error('运行失败（后端可能未启动）');
  } finally {
    running.value = '';
  }
}
onMounted(load);
</script>

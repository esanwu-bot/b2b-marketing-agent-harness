<template>
  <div class="app">
    <aside class="sidebar">
      <div class="logo"><div class="logo-mark">A</div><span>营销 Agent OS</span></div>
      <div class="workspace">工作区 · {{ workspace }}</div>
      <nav class="nav">
        <router-link v-for="m in menus" :key="m.path" :to="m.path" :class="{ active: isActive(m.path) }">
          <span class="icon">{{ m.icon }}</span>
          <span>{{ m.label }}</span>
          <span v-if="m.badge" class="badge">{{ m.badge }}</span>
        </router-link>
      </nav>
      <div class="sidebar-bottom">
        <div class="user">
          <div class="avatar">管</div>
          <div>
            <b style="color: #fff; font-size: 12px">管理员</b>
            <small style="display: block; color: #778399">{{ llmLabel }}</small>
          </div>
        </div>
      </div>
    </aside>

    <main class="main">
      <header class="topbar">
        <div class="crumb">{{ pageTitle }}</div>
        <span style="color: #c3cad5">/</span>
        <span style="color: #7b8496">{{ workspace }}</span>
        <div class="top-actions">
          <a-tag :color="online ? 'success' : 'default'">{{ online ? '已连接' : '本地演示数据' }}</a-tag>
          <a-button @click="refresh">刷新</a-button>
        </div>
      </header>

      <section class="content">
        <router-view v-slot="{ Component }">
          <keep-alive :max="10"><component :is="Component" /></keep-alive>
        </router-view>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { api, backendOnline } from '@/api';

const route = useRoute();
const online = ref(backendOnline);
const workspace = ref('Microchip Taiwan Demo');
const llmLabel = ref('mock · 本地');

const menus = [
  { path: '/workbench', label: 'Agent 工作台', icon: '⌂' },
  { path: '/agents', label: 'Agent 管理', icon: '◈', badge: 6 },
  { path: '/workflows', label: '工作流', icon: '◇' },
  { path: '/tools', label: '工具注册表', icon: '⚙' },
  { path: '/tasks', label: '采集任务', icon: '⌁' },
  { path: '/content', label: '内容工作室', icon: '✦' },
  { path: '/knowledge', label: '知识库', icon: '◎' },
  { path: '/analytics', label: '数据分析', icon: '↗' },
  { path: '/approvals', label: '审批中心', icon: '✓' },
  { path: '/observability', label: '可观测性', icon: '◌' },
];

const pageTitle = computed(() => (route.meta?.title as string) || 'Agent 工作台');
function isActive(path: string) {
  return route.path === path || route.path.startsWith(path + '/');
}
async function refresh() {
  const h = await api.health();
  online.value = h.status !== 'offline';
  if (h.workspace) workspace.value = h.workspace;
  if (h.llm) llmLabel.value = h.llm === 'mock' ? 'mock · 本地' : h.llm;
}
onMounted(refresh);
</script>

<style scoped>
.app {
  height: 100vh;
  display: flex;
  overflow: hidden;
}
.sidebar {
  width: 246px;
  background: #101827;
  color: #cbd5e1;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.logo {
  height: 64px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  gap: 11px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  color: #fff;
  font-weight: 700;
  font-size: 16px;
}
.logo-mark {
  width: 30px;
  height: 30px;
  border-radius: 9px;
  background: linear-gradient(135deg, #60a5fa, #7c3aed);
  display: grid;
  place-items: center;
  font-weight: 800;
}
.workspace {
  padding: 16px 14px 10px;
  font-size: 11px;
  color: #718096;
  letter-spacing: 0.08em;
}
.nav {
  padding: 0 10px;
  overflow: auto;
}
.nav a {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 10px 12px;
  margin: 2px 0;
  border-radius: 8px;
  color: #aeb9ca;
  text-decoration: none;
  font-size: 14px;
}
.nav a:hover,
.nav a.active {
  background: #182235;
  color: #fff;
}
.nav .icon {
  width: 20px;
  text-align: center;
}
.badge {
  margin-left: auto;
  background: #263247;
  color: #9fb4d3;
  padding: 2px 7px;
  border-radius: 10px;
  font-size: 10px;
}
.sidebar-bottom {
  margin-top: auto;
  padding: 14px;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
}
.user {
  display: flex;
  gap: 10px;
  align-items: center;
}
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #334155;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 600;
}
.main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.topbar {
  height: 64px;
  background: #fff;
  border-bottom: 1px solid #e6e9ef;
  display: flex;
  align-items: center;
  padding: 0 22px;
  gap: 10px;
}
.crumb {
  font-weight: 650;
}
.top-actions {
  margin-left: auto;
  display: flex;
  gap: 9px;
  align-items: center;
}
.content {
  padding: 0;
  overflow: auto;
  flex: 1;
}
</style>

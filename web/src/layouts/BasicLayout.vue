<template>
  <div class="min-h-screen bg-bg-canvas font-body-base text-body-base text-on-surface antialiased">
    <!-- Sidebar -->
    <aside
      class="fixed left-0 top-0 h-full w-[246px] bg-sidebar-base z-50 flex flex-col justify-between select-none shadow-[0_1px_8px_rgba(0,0,0,0.04)]"
    >
      <div class="flex flex-col flex-1 min-h-0">
        <div class="h-16 px-space-md flex items-center justify-between">
          <div class="flex items-center gap-space-sm">
            <div class="w-8 h-8 rounded-lg bg-primary-container flex items-center justify-center">
              <span class="material-symbols-outlined text-on-primary text-[18px]">memory</span>
            </div>
            <div class="flex flex-col">
              <span class="font-title-md text-title-md text-sidebar-text tracking-tight">芯智营销 OS</span>
              <span class="font-tag-counter text-tag-counter text-sidebar-muted uppercase">TikChip AI Matrix</span>
            </div>
          </div>
          <span class="w-2 h-2 rounded-full bg-status-success-dot animate-pulse" :title="online ? '后端已连接' : '本地演示数据'"></span>
        </div>

        <template v-for="g in groups" :key="g.label">
          <div class="px-space-md pt-space-sm pb-space-xs">
            <span class="font-label-caps text-label-caps text-sidebar-muted tracking-wider uppercase">{{ g.label }}</span>
          </div>
          <div class="px-space-sm space-y-0.5" :class="g.last ? '' : 'pb-2'">
            <nav class="flex flex-col gap-1">
              <router-link
                v-for="m in g.items"
                :key="m.path"
                :to="m.path"
                class="group flex items-center justify-between px-space-sm py-2 rounded-lg transition-all duration-150"
                :class="
                  isActive(m.path)
                    ? 'bg-sidebar-pill text-on-primary font-bold'
                    : 'text-sidebar-text hover:bg-sidebar-hover hover:text-on-primary'
                "
              >
                <div class="flex items-center gap-space-sm min-w-0">
                  <span
                    class="material-symbols-outlined text-[18px]"
                    :class="isActive(m.path) ? 'text-primary-fixed-dim' : 'text-sidebar-muted group-hover:text-on-primary'"
                    >{{ m.icon }}</span
                  >
                  <span class="font-subheading-sm text-subheading-sm truncate">{{ m.label }}</span>
                </div>
                <span
                  v-if="m.badge"
                  class="px-1.5 py-0.5 rounded-full bg-sidebar-pill text-sidebar-text font-tag-counter text-tag-counter"
                  >{{ m.badge }}</span
                >
                <span
                  v-else-if="isActive(m.path)"
                  class="w-1.5 h-1.5 rounded-full bg-primary-fixed-dim shadow-[0_0_6px_rgba(180,197,255,0.8)]"
                ></span>
              </router-link>
            </nav>
          </div>
        </template>
      </div>

      <div class="px-space-md py-space-md border-t border-white/5">
        <div class="flex items-center gap-space-sm">
          <div class="w-8 h-8 rounded-full bg-sidebar-pill flex items-center justify-center text-sidebar-text font-subheading-sm text-subheading-sm">
            管
          </div>
          <div class="flex flex-col min-w-0">
            <span class="font-subheading-sm text-subheading-sm text-sidebar-text truncate">管理员</span>
            <span class="font-tag-counter text-tag-counter text-sidebar-muted truncate">{{ llmLabel }} · {{ workspace }}</span>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main -->
    <main class="ml-[246px] p-6 min-h-screen">
      <router-view v-slot="{ Component }">
        <keep-alive :max="10"><component :is="Component" /></keep-alive>
      </router-view>
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

const groups = computed(() => [
  {
    label: '核心工作区',
    last: false,
    items: [
      { path: '/workbench', label: 'Agent 工作台', icon: 'space_dashboard' },
      { path: '/agents', label: 'Agent 管理', icon: 'smart_toy', badge: 6 },
      { path: '/workflows', label: '工作流编排', icon: 'schema' },
      { path: '/tools', label: '工具注册表', icon: 'build' },
    ],
  },
  {
    label: '内容与增长',
    last: false,
    items: [
      { path: '/tasks', label: '采集任务', icon: 'radar' },
      { path: '/content', label: '内容工作室', icon: 'edit_note' },
      { path: '/knowledge', label: '知识库', icon: 'memory' },
    ],
  },
  {
    label: '治理与洞察',
    last: true,
    items: [
      { path: '/analytics', label: '数据分析', icon: 'query_stats' },
      { path: '/approvals', label: '审批中心', icon: 'gavel' },
      { path: '/observability', label: '可观测性', icon: 'monitoring' },
    ],
  },
]);

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

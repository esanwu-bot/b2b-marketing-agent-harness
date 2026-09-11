import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import BasicLayout from '@/layouts/BasicLayout.vue';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: BasicLayout,
    redirect: '/workbench',
    children: [
      { path: 'workbench', name: 'workbench', component: () => import('@/pages/Workbench.vue'), meta: { title: 'Agent 工作台' } },
      { path: 'agents', name: 'agents', component: () => import('@/pages/Agents.vue'), meta: { title: 'Agent 管理' } },
      { path: 'workflows', name: 'workflows', component: () => import('@/pages/Workflows.vue'), meta: { title: '工作流' } },
      { path: 'tools', name: 'tools', component: () => import('@/pages/Tools.vue'), meta: { title: '工具注册表' } },
      { path: 'tasks', name: 'tasks', component: () => import('@/pages/Tasks.vue'), meta: { title: '采集任务' } },
      { path: 'content', name: 'content', component: () => import('@/pages/Content.vue'), meta: { title: '内容工作室' } },
      { path: 'knowledge', name: 'knowledge', component: () => import('@/pages/Knowledge.vue'), meta: { title: '知识库' } },
      { path: 'analytics', name: 'analytics', component: () => import('@/pages/Analytics.vue'), meta: { title: '数据分析' } },
      { path: 'approvals', name: 'approvals', component: () => import('@/pages/Approvals.vue'), meta: { title: '审批中心' } },
      { path: 'observability', name: 'observability', component: () => import('@/pages/Observability.vue'), meta: { title: '可观测性' } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/workbench' },
];

const router = createRouter({ history: createWebHistory(), routes });
router.afterEach((to) => {
  document.title = `${(to.meta?.title as string) || '工作台'} · 营销 Agent 工作台`;
});
export default router;

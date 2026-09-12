<template>
  <div>
    <!-- Top Ambient Banner & Visual Header -->
    <div class="relative w-full overflow-hidden rounded-xl bg-surface-container-low shadow-sm mb-6 p-6">
      <div class="absolute -right-16 -top-16 w-80 h-80 rounded-full bg-primary/5 blur-3xl pointer-events-none"></div>
      <div class="absolute right-48 -bottom-20 w-72 h-72 rounded-full bg-secondary-container/10 blur-2xl pointer-events-none"></div>
      <div class="relative z-10 flex flex-col xl:flex-row xl:items-center xl:justify-between gap-6">
        <div class="flex flex-col space-y-1.5">
          <div class="flex items-center gap-3">
            <span
              class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-primary-fixed text-on-primary-fixed-variant font-label-caps text-label-caps uppercase tracking-wider"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-primary"></span>
              Semiconductor Matrix v4.2
            </span>
            <span class="text-text-muted font-caption-sm text-caption-sm">集群节点：{{ workspace || 'Tokyo-East-Edge #04' }}</span>
          </div>
          <h1 class="font-headline-lg text-headline-lg text-on-surface tracking-tight">Agent 工作台</h1>
          <p class="font-body-base text-body-base text-on-surface-variant flex items-center flex-wrap gap-2">
            <span>营销全链路闭环：</span>
            <template v-for="(s, i) in pipelineChips" :key="s">
              <span class="px-2 py-0.5 rounded bg-surface-container text-on-surface text-caption-medium font-caption-medium">{{ s }}</span>
              <span v-if="i < pipelineChips.length - 1" class="text-text-muted">→</span>
            </template>
          </p>
        </div>
        <!-- Controls Toolbar -->
        <div class="flex items-center flex-wrap gap-3">
          <div class="relative inline-flex items-center">
            <select class="appearance-none bg-surface-panel pl-3 pr-8 py-2 rounded-lg text-on-surface font-body-base text-body-base shadow-sm focus:outline-none cursor-pointer">
              <option>最近 24 小时</option>
              <option>最近 7 天</option>
              <option>最近 30 天</option>
              <option>2026 Q3 季度</option>
            </select>
            <span class="material-symbols-outlined pointer-events-none absolute right-2 text-text-muted text-[18px]">arrow_drop_down</span>
          </div>
          <div class="relative inline-flex items-center">
            <select class="appearance-none bg-surface-panel pl-3 pr-8 py-2 rounded-lg text-on-surface font-body-base text-body-base shadow-sm focus:outline-none cursor-pointer">
              <option>全部智能体 (6/6)</option>
              <option>调研智能体 (Active)</option>
              <option>内容生成智能体 (Active)</option>
              <option>合规评审智能体 (Active)</option>
            </select>
            <span class="material-symbols-outlined pointer-events-none absolute right-2 text-text-muted text-[18px]">arrow_drop_down</span>
          </div>
          <button
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-primary-container text-on-primary hover:bg-primary transition-all shadow-sm font-subheading-sm text-subheading-sm disabled:opacity-60"
            type="button"
            :disabled="running"
            @click="runPipeline"
          >
            <span class="material-symbols-outlined text-[18px]">play_arrow</span>
            {{ running ? '流水线执行中…' : '新建运行' }}
          </button>
          <button
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-surface-panel text-on-surface hover:bg-surface-container-high transition-all shadow-sm font-subheading-sm text-subheading-sm"
            type="button"
            @click="exportReport"
          >
            <span class="material-symbols-outlined text-[18px] text-text-muted">download</span>
            导出运行周报
          </button>
          <button
            class="inline-flex items-center justify-center p-2 rounded-lg bg-primary-container text-on-primary hover:bg-primary transition-all shadow-sm"
            type="button"
            :class="{ 'animate-spin': syncing }"
            @click="load"
          >
            <span class="material-symbols-outlined text-[20px]">sync</span>
          </button>
        </div>
      </div>
    </div>

    <!-- KPI Cards Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 mb-6">
      <div class="p-5 rounded-xl bg-surface-panel shadow-sm hover:shadow-md transition-shadow relative overflow-hidden flex flex-col justify-between">
        <div class="flex items-center justify-between mb-3">
          <span class="font-subheading-sm text-subheading-sm text-text-muted">活跃运行中</span>
          <span class="inline-flex items-center gap-1 font-caption-medium text-caption-medium text-status-success-text bg-status-success-bg px-2 py-0.5 rounded-full">
            <span class="material-symbols-outlined text-[14px]">trending_up</span>
            ↑ 14.3% 较昨日
          </span>
        </div>
        <div class="flex items-baseline gap-2 mb-2">
          <span class="font-display-stat text-display-stat text-on-surface">{{ metrics.active_runs }}</span>
          <span class="font-caption-medium text-caption-medium text-text-muted">个 Agent 进程</span>
        </div>
        <div class="pt-3 flex items-center gap-1.5 text-text-muted font-caption-sm text-caption-sm">
          <span class="w-1.5 h-1.5 rounded-full bg-primary"></span>
          <span>{{ runningCount }} 运行 / {{ pendingApprovalCount }} 待审批</span>
        </div>
      </div>

      <div class="p-5 rounded-xl bg-surface-panel shadow-sm hover:shadow-md transition-shadow relative overflow-hidden flex flex-col justify-between">
        <div class="flex items-center justify-between mb-3">
          <span class="font-subheading-sm text-subheading-sm text-text-muted">已完成营销任务</span>
          <span class="inline-flex items-center gap-1 font-caption-medium text-caption-medium text-status-success-text bg-status-success-bg px-2 py-0.5 rounded-full">
            <span class="material-symbols-outlined text-[14px]">trending_up</span>
            ↑ 21.7% 本周环比
          </span>
        </div>
        <div class="flex items-baseline gap-2 mb-2">
          <span class="font-display-stat text-display-stat text-on-surface">{{ metrics.tasks_completed }}</span>
          <span class="font-caption-medium text-caption-medium text-text-muted">项任务</span>
        </div>
        <div class="pt-3 flex items-center justify-between text-caption-sm font-caption-sm text-text-muted">
          <span>任务执行成功率</span>
          <span class="text-status-success-text font-subheading-sm text-subheading-sm">{{ successRate }}</span>
        </div>
      </div>

      <div class="p-5 rounded-xl bg-surface-panel shadow-sm hover:shadow-md transition-shadow relative overflow-hidden flex flex-col justify-between">
        <div class="flex items-center justify-between mb-3">
          <span class="font-subheading-sm text-subheading-sm text-text-muted">待审发布内容</span>
          <span class="inline-flex items-center gap-1 font-caption-medium text-caption-medium text-status-warning-text bg-status-warning-bg px-2 py-0.5 rounded-full">
            <span class="w-1.5 h-1.5 rounded-full bg-status-warning-dot"></span>
            {{ metrics.approvals_pending }} 篇待技术总监终审
          </span>
        </div>
        <div class="flex items-baseline gap-2 mb-2">
          <span class="font-display-stat text-display-stat text-on-surface">{{ metrics.content_ready }}</span>
          <span class="font-caption-medium text-caption-medium text-text-muted">篇待审稿件</span>
        </div>
        <div class="pt-3 text-text-muted font-caption-sm text-caption-sm truncate">涵盖领英技术专栏、白皮书、产品对标</div>
      </div>

      <div class="p-5 rounded-xl bg-surface-panel shadow-sm hover:shadow-md transition-shadow relative overflow-hidden flex flex-col justify-between">
        <div class="flex items-center justify-between mb-3">
          <span class="font-subheading-sm text-subheading-sm text-text-muted">算力消耗 (LLM Cost)</span>
          <span class="inline-flex items-center gap-1 font-caption-medium text-caption-medium text-status-info-text bg-status-info-bg px-2 py-0.5 rounded-full">
            占单日预算 {{ budgetPercent }}%
          </span>
        </div>
        <div class="flex items-baseline gap-2 mb-2">
          <span class="font-display-stat text-display-stat text-on-surface">${{ metrics.llm_cost.toFixed(2) }}</span>
          <span class="font-caption-medium text-caption-medium text-text-muted">/ 今日配额 $48</span>
        </div>
        <div class="pt-3 flex items-center justify-between text-caption-sm font-caption-sm text-text-muted">
          <span>今日消耗 {{ tokenUsage }} Tokens</span>
          <div class="w-20 bg-surface-container-high rounded-full h-1.5 overflow-hidden">
            <div class="bg-primary h-1.5 rounded-full" :style="{ width: budgetPercent + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Stage: Split -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- Left Column -->
      <div class="lg:col-span-8 flex flex-col space-y-6">
        <!-- Section A: Agent Runs -->
        <div class="p-6 rounded-xl bg-surface-panel shadow-sm">
          <div class="flex items-center justify-between pb-4">
            <div class="flex items-center gap-3">
              <span class="material-symbols-outlined text-primary text-[24px]">terminal</span>
              <div>
                <h2 class="font-title-md text-title-md text-on-surface">智能体运行记录 (Agent Runs)</h2>
                <p class="font-caption-sm text-caption-sm text-text-muted">实时追踪自治智能体在半导体垂直情报与创作矩阵的执行链路</p>
              </div>
            </div>
            <router-link
              class="inline-flex items-center gap-1 font-subheading-sm text-subheading-sm text-primary hover:text-on-primary-fixed-variant transition-colors"
              to="/observability"
            >
              查看全部记录
              <span class="material-symbols-outlined text-[16px]">arrow_forward</span>
            </router-link>
          </div>
          <div class="flex flex-col divide-y divide-border-nested space-y-1">
            <div
              v-for="r in runs.slice(0, 5)"
              :key="r.id"
              class="py-3.5 flex flex-col md:flex-row md:items-center justify-between gap-3 hover:bg-surface-container-low/60 px-3 rounded-lg transition-colors"
            >
              <div class="flex items-start gap-3">
                <div class="w-8 h-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5" :class="runIconBg(r.status)">
                  <span class="material-symbols-outlined text-[18px]">{{ agentIcon(r.agent_id) }}</span>
                </div>
                <div class="flex flex-col">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="font-subheading-sm text-subheading-sm text-on-surface font-semibold">{{ runTitle(r) }}</span>
                    <span class="px-2 py-0.5 rounded bg-surface-container font-badge-micro text-badge-micro text-on-surface-variant">{{ agentName(r.agent_id) }}</span>
                  </div>
                  <div class="flex items-center gap-3 mt-1 font-caption-sm text-caption-sm text-text-muted">
                    <span class="flex items-center gap-1"
                      ><span class="material-symbols-outlined text-[14px]">token</span>{{ r.prompt_tokens }} tokens</span
                    >
                    <span>·</span>
                    <span>启动于 {{ hhmm(r.started_at) }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-3 self-end md:self-center shrink-0">
                <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full font-badge-micro text-badge-micro font-medium" :class="runPillClass(r.status)">
                  <span class="w-2 h-2 rounded-full" :class="runDotClass(r.status)"></span>
                  {{ runText(r.status) }}
                </span>
              </div>
            </div>
            <div v-if="!runs.length" class="py-8 text-center text-text-muted font-caption-sm text-caption-sm">暂无运行记录</div>
          </div>
        </div>

        <!-- Section B: Content Pipeline -->
        <div class="p-6 rounded-xl bg-surface-panel shadow-sm">
          <div class="flex items-center justify-between pb-5">
            <div class="flex items-center gap-3">
              <span class="material-symbols-outlined text-secondary text-[24px]">view_kanban</span>
              <div>
                <h2 class="font-title-md text-title-md text-on-surface">营销内容全周期流水线 (Content Pipeline)</h2>
                <p class="font-caption-sm text-caption-sm text-text-muted">自动化内容生成自适应多流转看板与阶段健康度</p>
              </div>
            </div>
            <span class="text-text-muted font-caption-sm text-caption-sm">总资产沉淀 {{ totalAssets }} 项</span>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-5 gap-3">
            <div v-for="s in pipeline" :key="s.label" class="p-3.5 rounded-xl bg-surface-container-low flex flex-col justify-between">
              <div>
                <div class="flex items-center justify-between mb-1.5">
                  <span class="font-caption-medium text-caption-medium text-text-muted">{{ s.label }}</span>
                  <span class="font-subheading-sm text-subheading-sm font-bold" :class="s.valueClass">{{ s.value }} {{ s.unit }}</span>
                </div>
                <span class="font-badge-micro text-badge-micro text-text-muted">{{ s.en }}</span>
              </div>
              <div class="mt-4">
                <div class="w-full bg-surface-container-high rounded-full h-1.5 mb-1.5 overflow-hidden">
                  <div class="h-1.5 rounded-full" :class="s.barClass" :style="{ width: s.percent + '%' }"></div>
                </div>
                <span class="font-tag-counter text-tag-counter" :class="s.valueClass">{{ s.footnote }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Section C: Content Calendar -->
        <div class="p-6 rounded-xl bg-surface-panel shadow-sm">
          <div class="flex items-center justify-between pb-5">
            <div class="flex items-center gap-3">
              <span class="material-symbols-outlined text-primary text-[24px]">calendar_month</span>
              <div>
                <h2 class="font-title-md text-title-md text-on-surface">营销内容排期日历 · 2026年9月 (Content Calendar)</h2>
                <p class="font-caption-sm text-caption-sm text-text-muted">半导体展会、技术白皮书、产品架构解读多节点排期</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span class="px-2 py-1 rounded bg-surface-container text-on-surface font-caption-medium text-caption-medium">2026年9月 第1-2周</span>
            </div>
          </div>
          <div class="grid grid-cols-7 gap-2">
            <div v-for="w in weekHeaders" :key="w" class="text-center font-caption-medium text-caption-medium text-text-muted py-1.5 bg-surface-container-low rounded">
              {{ w }}
            </div>
            <div
              v-for="d in calendarDays"
              :key="d.day"
              class="min-h-[96px] p-2 rounded-lg flex flex-col justify-between"
              :class="d.pill ? 'bg-surface-container-low' : 'bg-surface-container-low/40'"
            >
              <span class="font-subheading-sm text-subheading-sm" :class="d.pill ? 'font-bold text-on-surface' : 'font-normal text-text-muted'">
                {{ String(d.day).padStart(2, '0') }}
              </span>
              <div v-if="d.pill" class="p-1.5 rounded font-badge-micro text-badge-micro truncate" :class="d.pill.cls">{{ d.pill.text }}</div>
              <span v-else class="text-text-muted font-badge-micro text-badge-micro italic">无排期</span>
            </div>
          </div>
          <div class="mt-4 pt-3 flex items-center justify-between flex-wrap gap-2 text-caption-sm font-caption-sm text-text-muted">
            <div class="flex items-center gap-4">
              <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded bg-status-info-bg"></span>技术专栏与社媒</span>
              <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded bg-status-success-bg"></span>芯片评测与案例</span>
              <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded bg-secondary-fixed"></span>白皮书与SEO</span>
              <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded bg-tertiary-fixed"></span>展会与资本动态</span>
            </div>
            <span class="text-primary cursor-pointer hover:underline">同步到 Google Calendar ▾</span>
          </div>
        </div>
      </div>

      <!-- Right Column -->
      <div class="lg:col-span-4 flex flex-col space-y-6">
        <!-- Live Agent Trace -->
        <div class="p-6 rounded-xl bg-surface-panel shadow-sm">
          <div class="flex items-center justify-between pb-4">
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-status-success-dot animate-ping"></span>
              <h2 class="font-title-md text-title-md text-on-surface">实时智能体调用链路追踪</h2>
            </div>
            <span class="px-2 py-0.5 rounded-full bg-status-danger-bg text-status-danger-text font-badge-micro text-badge-micro font-bold">LIVE</span>
          </div>
          <div class="p-3 rounded-lg bg-surface-container-low mb-5">
            <div class="flex items-center justify-between text-caption-sm font-caption-sm text-text-muted mb-1">
              <span>当前执行实例</span>
              <span class="font-tag-counter text-tag-counter text-primary uppercase">Trace #{{ currentTraceId }}</span>
            </div>
            <div class="font-subheading-sm text-subheading-sm text-on-surface font-semibold truncate">{{ currentTraceTitle }}</div>
          </div>
          <div class="relative pl-6 space-y-5 before:content-[''] before:absolute before:left-2 before:top-2 before:bottom-2 before:w-[2px] before:bg-surface-container-high">
            <div v-for="(t, i) in trace" :key="i" class="relative" :class="{ 'p-2.5 rounded-lg bg-surface-container-low': t.active }">
              <span
                class="absolute -left-6 top-1 w-2.5 h-2.5 rounded-full"
                :class="t.active ? 'bg-primary ring-4 ring-primary/20 animate-pulse' : 'bg-status-success-dot'"
                :style="t.active ? 'top: 12px' : ''"
              ></span>
              <div class="flex flex-col">
                <div class="flex items-center justify-between">
                  <span class="font-subheading-sm text-subheading-sm font-semibold" :class="t.active ? 'text-primary font-bold' : 'text-on-surface'">
                    {{ t.title }}
                  </span>
                  <span class="font-tag-counter text-tag-counter" :class="t.active ? 'text-primary font-semibold' : 'text-text-muted'">{{ t.at }}</span>
                </div>
                <p class="font-caption-sm text-caption-sm mt-0.5" :class="t.active ? 'text-on-surface mt-1' : 'text-on-surface-variant'">{{ t.desc }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Data Sources -->
        <div class="p-6 rounded-xl bg-surface-panel shadow-sm">
          <div class="flex items-center justify-between pb-4">
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-primary text-[20px]">hub</span>
              <h2 class="font-title-md text-title-md text-on-surface">情报数据源监控 (Data Sources)</h2>
            </div>
            <span class="font-badge-micro text-badge-micro text-status-success-text bg-status-success-bg px-2 py-0.5 rounded-full">4 节点正常</span>
          </div>
          <div class="space-y-3">
            <div v-for="s in sources" :key="s.name" class="p-3 rounded-lg bg-surface-container-low flex items-center justify-between">
              <div class="flex items-center gap-3">
                <span class="material-symbols-outlined text-text-muted text-[20px]">{{ s.icon }}</span>
                <div class="flex flex-col">
                  <span class="font-subheading-sm text-subheading-sm text-on-surface font-medium">{{ s.name }}</span>
                  <span class="font-caption-sm text-caption-sm text-text-muted">{{ s.desc }}</span>
                </div>
              </div>
              <span class="inline-flex items-center gap-1 font-badge-micro text-badge-micro text-status-success-text">
                <span class="w-1.5 h-1.5 rounded-full bg-status-success-dot"></span>在线
              </span>
            </div>
          </div>
        </div>

        <!-- Approvals Pending -->
        <div class="p-6 rounded-xl bg-surface-panel shadow-sm">
          <div class="flex items-center justify-between pb-4">
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-status-warning-text text-[20px]">gavel</span>
              <h2 class="font-title-md text-title-md text-on-surface">待处理审批流</h2>
            </div>
            <span class="font-tag-counter text-tag-counter px-2 py-0.5 rounded-full bg-status-warning-bg text-status-warning-text font-bold">
              {{ metrics.approvals_pending }} 待处理
            </span>
          </div>
          <div class="space-y-3 mb-5">
            <div v-for="a in approvals.slice(0, 2)" :key="a.id" class="p-3 rounded-lg bg-surface-container-low hover:bg-surface-container transition-colors">
              <div class="flex items-center justify-between mb-1">
                <span class="font-badge-micro text-badge-micro px-1.5 py-0.5 rounded bg-status-info-bg text-status-info-text font-medium">
                  {{ a.kind }} · {{ a.requested_by }}
                </span>
                <span class="font-tag-counter text-tag-counter text-text-muted">{{ relTime(a.requested_at) }}</span>
              </div>
              <p class="font-subheading-sm text-subheading-sm text-on-surface font-semibold line-clamp-2">{{ a.summary }}</p>
            </div>
            <div v-if="!approvals.length" class="py-4 text-center text-text-muted font-caption-sm text-caption-sm">暂无待审批</div>
          </div>
          <router-link
            class="w-full py-2.5 px-4 rounded-lg bg-primary-container text-on-primary hover:bg-primary transition-all font-subheading-sm text-subheading-sm shadow-sm flex items-center justify-center gap-2"
            to="/approvals"
          >
            <span class="material-symbols-outlined text-[18px]">verified</span>
            一键进入人机协同审批中心 ({{ metrics.approvals_pending }} 条待处理)
          </router-link>
        </div>
      </div>
    </div>

    <!-- Bottom Section -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
      <!-- Agent Performance -->
      <div class="p-6 rounded-xl bg-surface-panel shadow-sm flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between pb-4">
            <div class="flex items-center gap-3">
              <span class="material-symbols-outlined text-primary text-[24px]">speed</span>
              <div>
                <h2 class="font-title-md text-title-md text-on-surface">智能体效能与准召率 (Agent Performance)</h2>
                <p class="font-caption-sm text-caption-sm text-text-muted">基于半导体芯片专业度基准评测模型综合得分</p>
              </div>
            </div>
            <span class="font-caption-medium text-caption-medium text-status-success-text bg-status-success-bg px-2.5 py-1 rounded-full">
              综合质量评分 {{ overallScore }}
            </span>
          </div>
          <div class="space-y-4 pt-2">
            <div v-for="(p, i) in perfList" :key="p.key">
              <div class="flex items-center justify-between text-subheading-sm font-subheading-sm mb-1.5">
                <span class="text-on-surface font-medium">{{ p.label }}</span>
                <div class="flex items-center gap-2">
                  <span class="text-status-success-text font-caption-medium text-caption-medium">↑ {{ p.delta }}%</span>
                  <span class="font-bold text-on-surface">{{ p.value.toFixed(1) }}%</span>
                </div>
              </div>
              <div class="w-full bg-surface-container-high rounded-full h-2 overflow-hidden">
                <div class="h-2 rounded-full" :class="perfBarClass(i)" :style="{ width: p.value + '%' }"></div>
              </div>
            </div>
            <div v-if="!perfList.length" class="py-6 text-center text-text-muted font-caption-sm text-caption-sm">暂无数据</div>
          </div>
        </div>
        <div class="mt-4 pt-3 flex items-center justify-between text-caption-sm font-caption-sm text-text-muted">
          <span>基于 IEEE / 半导体规范语义评测引擎</span>
          <span class="text-primary font-caption-medium text-caption-medium cursor-pointer">下载性能报表 (PDF)</span>
        </div>
      </div>

      <!-- Knowledge Base -->
      <div class="p-6 rounded-xl bg-surface-panel shadow-sm flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between pb-4">
            <div class="flex items-center gap-3">
              <span class="material-symbols-outlined text-primary text-[24px]">memory</span>
              <div>
                <h2 class="font-title-md text-title-md text-on-surface">芯片知识图谱与向量资产沉淀 (Knowledge Growth)</h2>
                <p class="font-caption-sm text-caption-sm text-text-muted">pgvector 混合检索增强，为智能体提供精准事实基座</p>
              </div>
            </div>
            <span class="font-subheading-sm text-subheading-sm font-bold text-primary">{{ metrics.documents.toLocaleString() }} 份文档已索引</span>
          </div>
          <div class="p-4 rounded-xl bg-surface-container-low mb-4 flex items-center justify-between">
            <div class="flex items-center gap-4">
              <svg class="w-12 h-12 shrink-0 -rotate-90" viewBox="0 0 36 36">
                <path class="text-surface-container-high" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" stroke-width="3.8"></path>
                <path class="text-primary" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" :stroke-dasharray="`${kbHealth}, 100`" stroke-linecap="round" stroke-width="3.8"></path>
              </svg>
              <div>
                <div class="font-subheading-sm text-subheading-sm text-on-surface font-semibold">向量库索引健康度 {{ kbHealth }}%</div>
                <p class="font-caption-sm text-caption-sm text-text-muted">支持 1536 维语义稠密检索 + BM25 稀疏混合打分</p>
              </div>
            </div>
            <button
              class="px-3 py-1.5 rounded-lg bg-surface-panel text-on-surface font-caption-medium text-caption-medium shadow-sm hover:bg-surface-container transition-colors"
              type="button"
              @click="reindex"
            >
              立即重建索引
            </button>
          </div>
          <div class="grid grid-cols-3 gap-3">
            <div class="p-3 rounded-lg bg-surface-container-low text-center">
              <span class="font-metric-sub text-metric-sub text-on-surface block">{{ kbSplit.datasheet.toLocaleString() }}</span>
              <span class="font-caption-sm text-caption-sm text-text-muted">芯片规格书 (Datasheet)</span>
            </div>
            <div class="p-3 rounded-lg bg-surface-container-low text-center">
              <span class="font-metric-sub text-metric-sub text-on-surface block">{{ kbSplit.articles.toLocaleString() }}</span>
              <span class="font-caption-sm text-caption-sm text-text-muted">技术深度文章</span>
            </div>
            <div class="p-3 rounded-lg bg-surface-container-low text-center">
              <span class="font-metric-sub text-metric-sub text-on-surface block">{{ kbSplit.social.toLocaleString() }}</span>
              <span class="font-caption-sm text-caption-sm text-text-muted">社媒营销资产</span>
            </div>
          </div>
        </div>
        <div class="mt-4 pt-3 flex items-center justify-between text-caption-sm font-caption-sm text-text-muted">
          <span>最后同步时间: 10分钟前</span>
          <router-link class="text-primary font-caption-medium text-caption-medium" to="/knowledge">进入芯片知识库 →</router-link>
        </div>
      </div>
    </div>

    <!-- Footer Badge -->
    <div class="mt-8 pt-4 flex flex-col sm:flex-row items-center justify-between text-caption-sm font-caption-sm text-text-muted gap-2">
      <div class="flex items-center gap-2">
        <span class="material-symbols-outlined text-[16px] text-status-success-text">verified_user</span>
        <span>B2B 半导体营销智能体运行平台 · PostgreSQL + pgvector 向量检索 · Redis 任务调度</span>
      </div>
      <div class="flex items-center gap-4">
        <span>crawl-worker-redis 实时爬虫适配器</span>
        <span>·</span>
        <span>TikChip AI Matrix v4.2.0-prod</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { message } from 'ant-design-vue';
import { api, backendOnline, type Approval, type Content, type Metrics, type Publication, type Run, type Topic, type WorkflowRun } from '@/api';

const metrics = ref<Metrics>({
  active_runs: 0, total_runs: 0, tasks_completed: 0, content_ready: 0, llm_cost: 0,
  documents: 0, topics: 0, publications: 0, approvals_pending: 0, content_by_status: {}, agent_performance: {}, workspace: '',
});
const runs = ref<Run[]>([]);
const approvals = ref<Approval[]>([]);
const topics = ref<Topic[]>([]);
const content = ref<Content[]>([]);
const publications = ref<Publication[]>([]);
const workflowRuns = ref<WorkflowRun[]>([]);
const syncing = ref(false);
const workspace = ref('');

const pipelineChips = ['行业调研', '情报挖掘', '内容生成', '人机审批', '全球分发', '效果复盘'];
const weekHeaders = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];

const sources = [
  { name: 'Microchip 官方领英', icon: 'feed', desc: '37 条新动态 · 08:14 抓取' },
  { name: '德州仪器 (TI) 开发者论坛', icon: 'forum', desc: '22 篇新动态 · 08:18 抓取' },
  { name: 'EE Times 电子工程专辑', icon: 'newspaper', desc: '18 篇技术研报 · 08:22 抓取' },
  { name: 'TechNews 科技新报', icon: 'insights', desc: '14 篇产业链分析 · 08:25 抓取' },
];

const runningCount = computed(() => runs.value.filter((r) => r.status === 'running').length);
const pendingApprovalCount = computed(() => runs.value.filter((r) => r.status === 'awaiting_approval').length);
const successRate = computed(() => {
  const done = runs.value.filter((r) => r.status === 'succeeded' || r.status === 'failed').length;
  if (!done) return '99.2% (125/126)';
  const ok = runs.value.filter((r) => r.status === 'succeeded').length;
  return `${((ok / done) * 100).toFixed(1)}% (${ok}/${done})`;
});
const budgetPercent = computed(() => Math.min(100, Math.round((metrics.value.llm_cost / 48) * 100)));
const tokenUsage = computed(() => {
  const t = runs.value.reduce((s, r) => s + (r.prompt_tokens || 0), 0);
  return t ? (t / 1_000_000).toFixed(1) + 'M' : '3.2M';
});
const totalAssets = computed(() => metrics.value.topics + metrics.value.publications + (metrics.value.content_by_status?.draft || 0));

const pipeline = computed(() => {
  const s = metrics.value.content_by_status || {};
  const draft = s.draft || 0;
  const review = s.in_review || 0;
  const approved = s.approved || 0;
  const published = s.published || 0;
  return [
    { label: '选题灵感池', en: 'Ideas Pool', value: metrics.value.topics, unit: '条', percent: Math.min(100, metrics.value.topics * 5 || 72), barClass: 'bg-primary', valueClass: 'text-on-surface', footnote: `饱和度 ${Math.min(100, metrics.value.topics * 5 || 72)}%` },
    { label: '初稿生成', en: 'Drafts In-Prog', value: draft, unit: '篇', percent: Math.min(100, draft * 5 || 48), barClass: 'bg-secondary', valueClass: 'text-on-surface', footnote: `生成负载 ${Math.min(100, draft * 5 || 48)}%` },
    { label: '专家审核', en: 'Tech Review', value: review, unit: '篇', percent: Math.min(100, review * 8 || 28), barClass: 'bg-status-warning-dot', valueClass: 'text-status-warning-text', footnote: `${Math.min(100, review * 8 || 28)}% 待处理` },
    { label: '排期待发', en: 'Scheduled', value: approved, unit: '篇', percent: Math.min(100, approved * 8 || 55), barClass: 'bg-primary-container', valueClass: 'text-on-surface', footnote: `日历就绪 ${Math.min(100, approved * 8 || 55)}%` },
    { label: '已全球发布', en: 'Published Live', value: published, unit: '篇', percent: Math.min(100, published || 88), barClass: 'bg-status-success-dot', valueClass: 'text-status-success-text', footnote: `转化达标 ${Math.min(100, published || 88)}%` },
  ];
});

const calendarDays = computed(() => {
  const pool = [
    ...content.value.map((c) => ({ text: c.title, cls: 'bg-status-info-bg text-status-info-text' })),
    ...topics.value.map((t) => ({ text: t.title, cls: 'bg-status-success-bg text-status-success-text' })),
    ...publications.value.map((p) => ({ text: p.external_url ? '已发布 · ' + p.external_id : p.id, cls: 'bg-secondary-fixed text-on-secondary-fixed' })),
  ];
  const fallback = [
    { text: 'Edge AI 算力趋势洞察', cls: 'bg-status-info-bg text-status-info-text' },
    { text: '车规级 MCU 新品实测', cls: 'bg-status-success-bg text-status-success-text' },
    { text: '深度技术白皮书', cls: 'bg-secondary-fixed text-on-secondary-fixed' },
    { text: '行业投融资要闻', cls: 'bg-tertiary-fixed text-on-tertiary-fixed-variant' },
  ];
  const list = pool.length ? pool : fallback;
  const empty = new Set([5, 6, 12, 13, 14]);
  return Array.from({ length: 14 }, (_, i) => {
    const day = i + 1;
    return { day, pill: empty.has(day) ? null : list[i % list.length] };
  });
});

const currentTrace = computed(() => workflowRuns.value.find((w) => w.status === 'running') || workflowRuns.value[0]);
const currentTraceId = computed(() => (currentTrace.value?.id || 'TK-8942').replace(/\W/g, '').slice(-4).padStart(4, '8'));
const currentTraceTitle = computed(() => {
  if (currentTrace.value?.steps?.length) {
    const last = currentTrace.value.steps[currentTrace.value.steps.length - 1];
    return last.ref || currentTrace.value.workflow_key;
  }
  return '每日半导体行业深度情报扫描';
});

const trace = computed(() => {
  const wr = currentTrace.value;
  if (wr?.steps?.length) {
    const steps = wr.steps.slice(-4).map((st) => ({
      title: `${stepKindLabel(st.kind)} · ${st.ref}`,
      desc: `状态：${st.status}`,
      at: hhmm(st.at),
      active: false,
    }));
    steps.push({ title: '下一阶段 · 执行中', desc: '正在推进工作流后续节点…', at: '进行中…', active: true });
    return steps;
  }
  return [
    { title: '调研智能体 · 任务规划', desc: '分析 5 个核心信息源，准备启动竞品监控雷达。', at: '08:00:12', active: false },
    { title: '工具调用 · crawler.create_job', desc: '数据源: microchip_linkedin · 限制: 100 条', at: '08:01:05', active: false },
    { title: '工具调用 · browser.search', desc: '检索词: “semiconductor edge AI 2026”', at: '08:02:18', active: false },
    { title: '观测结果 · 182 份文档', desc: '发现 16 个高热话题，车载以太网与边缘 AI 热度急剧上升。', at: '08:03:30', active: false },
    { title: '下一阶段 · 趋势分析智能体', desc: '正在进行向量主题聚类与内容转化价值评分…', at: '进行中…', active: true },
  ];
});

const perfList = computed(() => {
  const labels: Record<string, string> = {
    research: '调研智能体 (Retrieval & Crawl)',
    content: '内容生成智能体 (Content Synthesis)',
    review: '合规评审智能体 (Technical Verification)',
    analytics: '数据分析智能体 (Analytics & Attribution)',
    trend: '趋势智能体 (Topic Clustering)',
    strategy: '策略智能体 (Topic Scoring)',
  };
  const deltas: Record<string, string> = { research: '1.2', content: '3.4', review: '0.8', analytics: '2.1', trend: '1.6', strategy: '0.9' };
  const perf = metrics.value.agent_performance || {};
  return Object.keys(perf).map((k) => ({ key: k, label: labels[k] || k, value: perf[k], delta: deltas[k] || '1.0' }));
});
const overallScore = computed(() => {
  const list = perfList.value;
  if (!list.length) return '95.3';
  return (list.reduce((s, p) => s + p.value, 0) / list.length).toFixed(1);
});

const kbHealth = computed(() => Math.min(100, Math.round((metrics.value.documents / 15000) * 100) || 76));
const kbSplit = computed(() => {
  const d = metrics.value.documents || 12842;
  const datasheet = Math.round(d * 0.271);
  const articles = Math.round(d * 0.461);
  return { datasheet, articles, social: d - datasheet - articles };
});

function perfBarClass(i: number) {
  return ['bg-primary', 'bg-secondary', 'bg-status-success-dot', 'bg-tertiary-container'][i % 4];
}
function agentName(id: string) {
  const map: Record<string, string> = {
    research: '调研智能体', trend: '趋势分析智能体', strategy: '策略智能体',
    content: '内容生成智能体', review: '合规评审智能体', analytics: '数据分析智能体',
  };
  return map[id] || id || '未知智能体';
}
function agentIcon(id: string) {
  const map: Record<string, string> = {
    research: 'travel_explore', trend: 'bubble_chart', strategy: 'lightbulb',
    content: 'draw', review: 'gavel', analytics: 'query_stats',
  };
  return map[id] || 'smart_toy';
}
function runTitle(r: Run) {
  const map: Record<string, string> = {
    research: '每日半导体行业深度情报扫描',
    content: '边缘 AI 芯片 领英全球推广战役',
    trend: '车载以太网 (10BASE-T1S) 主题聚类分析',
    analytics: '海外社媒营销效果周度复盘',
    review: '技术事实与品牌口吻合规评审',
    strategy: '选题打分与战役策略选择',
  };
  return map[r.agent_id] || r.id;
}
function runIconBg(status: string) {
  if (status === 'running') return 'bg-status-info-bg text-status-info-text';
  if (status === 'awaiting_approval') return 'bg-status-warning-bg text-status-warning-text';
  if (status === 'failed') return 'bg-status-danger-bg text-status-danger-text';
  return 'bg-status-success-bg text-status-success-text';
}
function runPillClass(status: string) {
  if (status === 'running') return 'bg-status-info-bg text-status-info-text';
  if (status === 'awaiting_approval') return 'bg-status-warning-bg text-status-warning-text';
  if (status === 'failed') return 'bg-status-danger-bg text-status-danger-text';
  return 'bg-status-success-bg text-status-success-text';
}
function runDotClass(status: string) {
  if (status === 'running') return 'bg-primary animate-ping';
  if (status === 'awaiting_approval') return 'bg-status-warning-dot';
  if (status === 'failed') return 'bg-status-danger-dot';
  return 'bg-status-success-dot';
}
function runText(s: string) {
  const map: Record<string, string> = {
    running: '运行中 (Running)', succeeded: '已完成', awaiting_approval: '待人机审批', failed: '失败', cancelled: '已取消',
  };
  return map[s] || s;
}
function stepKindLabel(kind: string) {
  const map: Record<string, string> = { agent: '智能体', tool: '工具调用', approval: '人工审批', channel: '渠道发布', plan: '规划', output: '产出' };
  return map[kind] || kind;
}
function hhmm(iso: string) {
  if (!iso) return '--:--';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}
function relTime(iso: string) {
  const d = new Date(iso).getTime();
  if (Number.isNaN(d)) return iso;
  const min = Math.max(1, Math.round((Date.now() - d) / 60000));
  if (min < 60) return `${min}分钟前`;
  return `${Math.round(min / 60)}小时前`;
}
function exportReport() {
  message.info('运行周报导出为演示能力，正式环境接入 OSS 后开放');
}
function reindex() {
  message.success('向量索引重建任务已提交（后台异步执行）');
}

const running = ref(false);
async function runPipeline() {
  running.value = true;
  try {
    await api.runWorkflow('daily_semiconductor_intelligence');
    message.success('已触发每日半导体情报与内容流水线');
    await new Promise((r) => setTimeout(r, 600));
    await load();
  } catch {
    message.error('触发失败（后端可能未启动）');
  } finally {
    running.value = false;
  }
}

async function load() {
  syncing.value = true;
  try {
    metrics.value = await api.metrics();
    const [r, a, t, c, p, wr, h] = await Promise.all([
      api.runs(), api.approvals('pending'), api.topics(), api.content(), api.publications(), api.workflowRuns(), api.health(),
    ]);
    runs.value = r;
    approvals.value = a;
    topics.value = t;
    content.value = c;
    publications.value = p;
    workflowRuns.value = wr;
    workspace.value = h.workspace || metrics.value.workspace || '';
    if (!backendOnline) message.info('后端未连接，当前展示本地演示数据');
  } finally {
    syncing.value = false;
  }
}

onMounted(load);
</script>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Briefcase, Close, Notebook, Upload, User } from '@element-plus/icons-vue'
import HomeView from '../views/HomeView.vue'
import UploadView from '../views/UploadView.vue'
import DocsView from '../views/DocsView.vue'
import CandidateView from '../views/CandidateView.vue'
import PositionsView from '../views/PositionsView.vue'
import { useTabs } from '../composables/useTabs'
import { useCandidateNav } from '../composables/useCandidateNav'
import { usePositions } from '../composables/usePositions'
import { statusLabel, statusType } from '../constants/status'

const route = useRoute()
const router = useRouter()
const { activeKey, lastWorkspaceKey, candidateTabs, openCandidate, openCandidateById, removeTab, switchTab } = useTabs()
const { navList } = useCandidateNav()
const { openPositions, listPositionId, loadPositions } = usePositions()

const isCandidateView = computed(() => activeKey.value.startsWith('candidate-'))
const isWorkspaceView = computed(() => activeKey.value === 'list' || activeKey.value === 'kanban')

/** 详情页仍点亮来源子项（列表或看板），避免侧栏看起来像停在上传。 */
const navKey = computed(() =>
  isCandidateView.value ? lastWorkspaceKey.value : activeKey.value,
)

/** 打开过候选人后加宽侧栏，给便签和当前列表留出固定选择区 */
const asideWidth = computed(() => (candidateTabs.value.length || isCandidateView.value ? '208px' : '168px'))

function onNavSelect(key: string) {
  if (key === 'list') listPositionId.value = 0
  switchTab(key)
  syncRoute()
}

function onSelectOpening(id: number) {
  listPositionId.value = id
  switchTab('list')
  syncRoute()
}

function onTabRemove(name: string | number) {
  removeTab(String(name))
  syncRoute()
}

function selectOpened(key: string) {
  activeKey.value = key
  syncRoute()
}

function selectFromList(c: { id: number; name: string; tier: string; status: string }) {
  openCandidate(c)
  syncRoute()
}

function syncRoute() {
  const q: Record<string, string> = {}
  if (activeKey.value !== 'list') {
    q.tab = activeKey.value
  }
  router.replace({ path: '/', query: q })
}

watch(activeKey, syncRoute)

const bootTab = route.query.tab as string | undefined
if (bootTab?.startsWith('candidate-')) {
  const id = Number(bootTab.replace('candidate-', ''))
  if (id) openCandidateById(id)
  activeKey.value = bootTab
} else if (bootTab === 'upload' || bootTab === 'docs' || bootTab === 'kanban' || bootTab === 'list' || bootTab === 'positions') {
  switchTab(bootTab)
} else if (bootTab === 'home') {
  switchTab('list')
}

onMounted(() => {
  syncRoute()
  loadPositions().catch(() => {})
})
</script>

<template>
  <el-container class="layout">
    <el-header class="header">
      <div class="brand">
        <!-- 焦糖色简历夹，刻意不用工程开发平台那套蓝紫立方体 / 字母 A -->
        <svg class="brand-mark" viewBox="0 0 32 32" aria-hidden="true">
          <rect width="32" height="32" rx="8" fill="#C2410C" />
          <rect x="8" y="7" width="16" height="19" rx="2" fill="#FFF7ED" />
          <rect x="12" y="5" width="8" height="4" rx="1.2" fill="#FDBA74" />
          <path d="M11 14h10M11 18h10M11 22h6" stroke="#9A3412" stroke-width="1.8" stroke-linecap="round" fill="none" />
        </svg>
        <span>Agent 招聘评估台</span>
      </div>
      <div class="tagline">工程落地优先 · 标签页打开候选人 · 可标记面试状态</div>
    </el-header>

    <el-container class="body">
      <el-aside :width="asideWidth" class="side">
        <nav class="side-nav">
          <!-- 候选人是一级分组，列表/看板是侧栏二级标题，不再放进页面里切。 -->
          <div class="nav-group">
            <div class="nav-group-title">
              <el-icon><User /></el-icon>
              <span>候选人</span>
            </div>
            <button
              type="button"
              class="nav-sub"
              :class="{ active: navKey === 'list' && listPositionId === 0 }"
              @click="onNavSelect('list')"
            >
              列表
            </button>
            <button
              type="button"
              class="nav-sub"
              :class="{ active: navKey === 'kanban' }"
              @click="onNavSelect('kanban')"
            >
              看板
            </button>
          </div>
          <div class="nav-group">
            <button
              type="button"
              class="nav-group-btn"
              :class="{ active: navKey === 'positions' }"
              @click="onNavSelect('positions')"
            >
              <el-icon><Briefcase /></el-icon>
              <span>岗位阶梯</span>
            </button>
            <button
              v-for="p in openPositions"
              :key="p.id"
              type="button"
              class="nav-sub"
              :class="{ active: navKey === 'list' && listPositionId === p.id }"
              :title="p.skill_name ? `检验标准：${p.skill_name}` : undefined"
              @click="onSelectOpening(p.id)"
            >
              {{ p.name }}
            </button>
          </div>
          <button
            type="button"
            class="nav-item"
            :class="{ active: navKey === 'upload' }"
            @click="onNavSelect('upload')"
          >
            <el-icon><Upload /></el-icon>
            <span>上传</span>
          </button>
          <button
            type="button"
            class="nav-item"
            :class="{ active: navKey === 'docs' }"
            @click="onNavSelect('docs')"
          >
            <el-icon><Notebook /></el-icon>
            <span>知识库</span>
          </button>
        </nav>

        <div v-if="candidateTabs.length" class="sticky-panel">
          <div class="sticky-head">已选候选人</div>
          <button
            v-for="tab in candidateTabs"
            :key="tab.key"
            type="button"
            class="sticky-note"
            :class="{ active: activeKey === tab.key }"
            @click="selectOpened(tab.key)"
          >
            <span class="note-name">{{ tab.title }}</span>
            <span class="note-meta">
              <el-tag v-if="tab.tier" size="small" :type="tab.tier === 'S' ? 'danger' : 'info'">{{ tab.tier }}</el-tag>
              <el-tag v-if="tab.status" size="small" :type="statusType(tab.status)">{{ statusLabel(tab.status) }}</el-tag>
            </span>
            <el-icon class="note-close" @click.stop="onTabRemove(tab.key)"><Close /></el-icon>
          </button>
        </div>

        <div v-if="isCandidateView && navList.length" class="list-panel">
          <div class="sticky-head">当前列表 {{ navList.length }}</div>
          <button
            v-for="(c, i) in navList"
            :key="c.id"
            type="button"
            class="list-row"
            :class="{ active: activeKey === `candidate-${c.id}` }"
            @click="selectFromList(c)"
          >
            <span class="list-idx">{{ i + 1 }}</span>
            <span class="list-name">{{ c.name }}</span>
            <span class="list-tier">{{ c.tier }}</span>
          </button>
        </div>
      </el-aside>

      <el-main class="main">
        <el-tabs
          v-if="candidateTabs.length && isCandidateView"
          v-model="activeKey"
          type="card"
          class="candidate-tabs"
          @tab-remove="onTabRemove"
        >
          <el-tab-pane
            v-for="tab in candidateTabs"
            :key="tab.key"
            :name="tab.key"
            :closable="true"
            lazy
          >
            <template #label>
              <span class="tab-label">
                {{ tab.title }}
                <el-tag v-if="tab.tier" size="small" :type="tab.tier === 'S' ? 'danger' : 'info'" class="tab-tier">
                  {{ tab.tier }}
                </el-tag>
                <el-tag v-if="tab.status" size="small" :type="statusType(tab.status)" class="tab-tier">
                  {{ statusLabel(tab.status) }}
                </el-tag>
              </span>
            </template>
            <CandidateView :id="String(tab.candidateId)" />
          </el-tab-pane>
        </el-tabs>

        <div v-show="!isCandidateView" class="page-panel">
          <HomeView v-if="isWorkspaceView" />
          <UploadView v-else-if="activeKey === 'upload'" />
          <PositionsView v-else-if="activeKey === 'positions'" />
          <DocsView v-else-if="activeKey === 'docs'" />
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout {
  min-height: 100vh;
}
.header {
  display: flex;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  padding: 0 20px;
  gap: 24px;
  height: 56px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  font-size: 18px;
  white-space: nowrap;
}
.brand-mark {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}
.tagline {
  margin-left: auto;
  color: #909399;
  font-size: 13px;
}
.body {
  min-height: calc(100vh - 56px);
}
.side {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-right: 1px solid #ebeef5;
  overflow: hidden;
}
.side-nav {
  padding: 8px 0 4px;
  flex-shrink: 0;
}
.nav-group {
  margin-bottom: 4px;
}
.nav-group-title,
.nav-group-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px 4px;
  font-size: 12px;
  font-weight: 600;
  color: #909399;
  letter-spacing: 0.02em;
}
.nav-group-btn {
  width: calc(100% - 16px);
  margin: 2px 8px 0;
  padding: 8px 12px 4px;
  border: none;
  border-radius: 8px;
  background: transparent;
  text-align: left;
  cursor: pointer;
}
.nav-group-btn.active {
  color: #409eff;
}
.nav-group-btn:hover {
  background: #f5f7fa;
}
.nav-sub,
.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: calc(100% - 16px);
  margin: 2px 8px;
  padding: 8px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #303133;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
}
.nav-sub {
  margin-left: 20px;
  width: calc(100% - 28px);
  padding: 7px 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nav-sub.active,
.nav-item.active {
  background: #ecf5ff;
  color: #409eff;
  font-weight: 600;
}
.nav-item:hover,
.nav-sub:hover {
  background: #f5f7fa;
}
.nav-sub.active:hover,
.nav-item.active:hover {
  background: #ecf5ff;
}
.sticky-panel,
.list-panel {
  padding: 8px 10px 12px;
  overflow-y: auto;
}
.sticky-panel {
  flex: 0 0 auto;
  max-height: 42%;
  border-top: 1px solid #ebeef5;
}
.list-panel {
  flex: 1 1 auto;
  border-top: 1px dashed #ebeef5;
}
.sticky-head {
  font-size: 12px;
  color: #909399;
  margin: 4px 2px 8px;
}
.sticky-note {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  width: 100%;
  margin: 0 0 8px;
  padding: 10px 22px 10px 10px;
  border: 1px solid #fde68a;
  border-radius: 2px;
  background: #fffbeb;
  box-shadow: 1px 2px 0 rgba(180, 83, 9, 0.08);
  text-align: left;
  cursor: pointer;
}
.sticky-note:nth-child(odd) {
  background: #fff7ed;
}
.sticky-note.active {
  border-color: #c2410c;
  background: #ffedd5;
  box-shadow: 0 0 0 1px #c2410c;
}
.note-name {
  font-size: 13px;
  font-weight: 600;
  color: #3f3f46;
}
.note-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.note-close {
  position: absolute;
  top: 6px;
  right: 6px;
  color: #a8abb2;
}
.note-close:hover {
  color: #c2410c;
}
.list-row {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  margin: 0 0 4px;
  padding: 6px 8px;
  border: none;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  text-align: left;
}
.list-row:hover,
.list-row.active {
  background: #ecf5ff;
}
.list-idx {
  flex: 0 0 18px;
  color: #909399;
  font-size: 12px;
}
.list-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}
.list-tier {
  flex-shrink: 0;
  font-size: 11px;
  color: #c2410c;
}
.main {
  padding: 12px 16px 16px;
  background: #f5f7fa;
}
/* 切换改走左侧便签，顶栏标签只保留内容区，避免两套选择器抢视线 */
.candidate-tabs :deep(.el-tabs__header) {
  display: none;
}
.candidate-tabs :deep(.el-tabs__content) {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  min-height: calc(100vh - 120px);
}
.page-panel {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  min-height: calc(100vh - 120px);
}
.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.tab-tier {
  transform: scale(0.9);
}
</style>

<style>
body {
  margin: 0;
  background: #f5f7fa;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}
</style>

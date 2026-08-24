<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { User, Upload, Notebook } from '@element-plus/icons-vue'
import HomeView from '../views/HomeView.vue'
import UploadView from '../views/UploadView.vue'
import DocsView from '../views/DocsView.vue'
import CandidateView from '../views/CandidateView.vue'
import { useTabs } from '../composables/useTabs'
import { statusLabel, statusType } from '../constants/status'

const route = useRoute()
const router = useRouter()
const { activeKey, candidateTabs, openCandidateById, removeTab } = useTabs()

const navItems = [
  { key: 'home', label: '候选人', icon: User },
  { key: 'upload', label: '上传', icon: Upload },
  { key: 'docs', label: '知识库', icon: Notebook },
] as const

const navKey = computed(() =>
  activeKey.value.startsWith('candidate-') ? 'home' : activeKey.value,
)

const isCandidateView = computed(() => activeKey.value.startsWith('candidate-'))

const candidateTabActive = computed({
  get: () => (isCandidateView.value ? activeKey.value : ''),
  set: (key: string) => {
    if (key) activeKey.value = key
  },
})

function onNavSelect(key: string) {
  activeKey.value = key
  syncRoute()
}

function onTabRemove(name: string | number) {
  removeTab(String(name))
  syncRoute()
}

function syncRoute() {
  const q: Record<string, string> = {}
  if (activeKey.value !== 'home') {
    q.tab = activeKey.value
  }
  router.replace({ path: '/', query: q })
}

watch(activeKey, syncRoute)

onMounted(() => {
  const tab = route.query.tab as string | undefined
  if (tab?.startsWith('candidate-')) {
    const id = Number(tab.replace('candidate-', ''))
    if (id) openCandidateById(id)
    activeKey.value = tab
  } else if (tab === 'upload') {
    activeKey.value = 'upload'
  } else if (tab === 'docs') {
    activeKey.value = 'docs'
  }
})
</script>

<template>
  <el-container class="layout">
    <el-header class="header">
      <div class="brand">Agent 招聘评估台</div>
      <div class="tagline">工程落地优先 · 标签页打开候选人 · 可标记面试状态</div>
    </el-header>

    <el-container class="body">
      <el-aside width="148px" class="side">
        <el-menu
          :default-active="navKey"
          class="side-menu"
          @select="onNavSelect"
        >
          <el-menu-item v-for="item in navItems" :key="item.key" :index="item.key">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-main class="main">
        <el-tabs
          v-if="candidateTabs.length"
          v-model="candidateTabActive"
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
          <HomeView v-if="activeKey === 'home'" />
          <UploadView v-else-if="activeKey === 'upload'" />
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
  font-weight: 700;
  font-size: 18px;
  white-space: nowrap;
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
  background: #fff;
  border-right: 1px solid #ebeef5;
}
.side-menu {
  border-right: none;
  padding-top: 8px;
}
.side-menu .el-menu-item {
  height: 44px;
  margin: 4px 8px;
  border-radius: 8px;
}
.side-menu .el-menu-item.is-active {
  background: #ecf5ff;
}
.main {
  padding: 12px 16px 16px;
  background: #f5f7fa;
}
.candidate-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}
.candidate-tabs :deep(.el-tabs__content) {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-top: none;
  border-radius: 0 0 8px 8px;
  padding: 16px;
  min-height: calc(100vh - 180px);
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

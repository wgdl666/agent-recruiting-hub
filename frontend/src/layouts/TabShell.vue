<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import UploadView from '../views/UploadView.vue'
import DocsView from '../views/DocsView.vue'
import CandidateView from '../views/CandidateView.vue'
import { useTabs } from '../composables/useTabs'
import { statusLabel, statusType } from '../constants/status'

const route = useRoute()
const router = useRouter()
const { activeKey, candidateTabs, openCandidateById, removeTab } = useTabs()

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
    <el-main class="main">
      <el-tabs
        v-model="activeKey"
        type="card"
        class="app-tabs"
        @tab-remove="onTabRemove"
      >
        <el-tab-pane label="候选人" name="home" :closable="false">
          <HomeView />
        </el-tab-pane>
        <el-tab-pane label="上传" name="upload" :closable="false">
          <UploadView />
        </el-tab-pane>
        <el-tab-pane label="知识库" name="docs" :closable="false">
          <DocsView />
        </el-tab-pane>
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
    </el-main>
  </el-container>
</template>

<style>
body {
  margin: 0;
  background: #f5f7fa;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}
.layout { min-height: 100vh; }
.header {
  display: flex;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  padding: 0 20px;
  gap: 24px;
}
.brand { font-weight: 700; font-size: 18px; white-space: nowrap; }
.tagline { margin-left: auto; color: #909399; font-size: 13px; }
.el-header { height: 56px; }
.main { padding-top: 12px; }
.app-tabs > .el-tabs__header { margin-bottom: 0; }
.app-tabs > .el-tabs__content {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-top: none;
  border-radius: 0 0 8px 8px;
  padding: 16px;
  min-height: calc(100vh - 120px);
}
.tab-label { display: inline-flex; align-items: center; gap: 6px; }
.tab-tier { transform: scale(0.9); }
</style>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { Upload } from '@element-plus/icons-vue'
import {
  fetchCandidates, fetchPipelineStats, fetchStats,
} from '../api/client'
import type { Candidate, PipelineStats, Stats } from '../types'
import { useTabs } from '../composables/useTabs'
import { useActiveBatch } from '../composables/useActiveBatch'
import { PIPELINE_ORDER, statusLabel } from '../constants/status'
import StatsBar from '../components/StatsBar.vue'
import CandidateTable from '../components/CandidateTable.vue'
import UploadDropzone from '../components/UploadDropzone.vue'
import ToolbarActions from '../components/ToolbarActions.vue'
import WorkflowGuide from '../components/WorkflowGuide.vue'
import BatchSelector from '../components/pipeline/BatchSelector.vue'
import PipelineFunnel from '../components/pipeline/PipelineFunnel.vue'
import PipelineKanban from '../components/pipeline/PipelineKanban.vue'

const { switchTab } = useTabs()
const { activeBatchId, uploadPeriodType } = useActiveBatch()
const tier = ref('S')
const status = ref('screening')
const query = ref('')
const viewMode = ref<'list' | 'kanban'>('list')
const candidates = ref<Candidate[]>([])
const kanbanCandidates = ref<Candidate[]>([])
const stats = ref<Stats | null>(null)
const pipelineStats = ref<PipelineStats | null>(null)
const loading = ref(false)
const uploadExpanded = ref<string[]>([])

async function load() {
  loading.value = true
  try {
    const batchId = activeBatchId.value
    const [list, st, pipe, allForKanban] = await Promise.all([
      fetchCandidates(tier.value, query.value, status.value, batchId),
      fetchStats(),
      fetchPipelineStats(batchId),
      viewMode.value === 'kanban'
        ? fetchCandidates('all', query.value, 'all', batchId)
        : Promise.resolve([] as Candidate[]),
    ])
    candidates.value = list
    stats.value = st
    pipelineStats.value = pipe
    if (viewMode.value === 'kanban') kanbanCandidates.value = allForKanban
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([tier, status, query, activeBatchId, viewMode], load)

function onUploadDone() {
  uploadExpanded.value = []
  status.value = 'screening'
  load()
}
</script>

<template>
  <div>
    <WorkflowGuide />
    <a-card :bordered="false" class="pipeline-card">
      <BatchSelector @change="load" />
      <PipelineFunnel :stats="pipelineStats" :loading="loading" />
    </a-card>

    <StatsBar :stats="stats" />

    <el-card shadow="never">
      <div class="toolbar">
        <div class="filters">
          <a-segmented
            v-model:value="viewMode"
            :options="[
              { value: 'list', label: '列表' },
              { value: 'kanban', label: '看板' },
            ]"
          />
          <div v-if="viewMode === 'list'" class="filter-row">
            <span class="filter-label">状态</span>
            <el-radio-group v-model="status" size="small">
              <el-radio-button value="all">全部</el-radio-button>
              <el-radio-button v-for="s in PIPELINE_ORDER" :key="s" :value="s">
                {{ statusLabel(s) }}
              </el-radio-button>
            </el-radio-group>
          </div>
          <div v-if="viewMode === 'list'" class="filter-row">
            <span class="filter-label">档位</span>
            <el-radio-group v-model="tier" size="small">
              <el-radio-button value="all">全部</el-radio-button>
              <el-radio-button value="S">S</el-radio-button>
              <el-radio-button value="A">A</el-radio-button>
              <el-radio-button value="淘汰">淘汰</el-radio-button>
            </el-radio-group>
          </div>
          <el-input
            v-model="query"
            placeholder="搜索姓名 / 项目"
            clearable
            size="small"
            class="search"
          />
        </div>

        <div class="primary-actions">
          <el-button type="primary" :icon="Upload" @click="switchTab('upload')">上传简历</el-button>
          <ToolbarActions @refresh="load" />
        </div>
      </div>

      <el-collapse v-model="uploadExpanded" class="quick-upload">
        <el-collapse-item name="upload" title="快速上传（拖拽 PDF / ZIP，归入当前周期批次）">
          <UploadDropzone
            :batch-id="activeBatchId"
            :period-type="uploadPeriodType"
            @done="onUploadDone"
          />
        </el-collapse-item>
      </el-collapse>

      <PipelineKanban
        v-if="viewMode === 'kanban'"
        :candidates="kanbanCandidates"
        :loading="loading"
        @updated="load"
      />

      <template v-else>
        <CandidateTable
          :candidates="candidates"
          :loading="loading"
          :show-order="tier === 'S' && status !== 'completed' && status !== 'rejected'"
          @updated="load"
        />

        <el-empty
          v-if="!loading && candidates.length === 0"
          description="没有匹配的候选人"
        >
          <el-button type="primary" @click="switchTab('upload')">去上传简历</el-button>
        </el-empty>
      </template>
    </el-card>
  </div>
</template>

<style scoped>
.pipeline-card {
  margin-bottom: 16px;
}
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 12px;
}
.filters {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  min-width: 280px;
}
.filter-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.filter-label {
  font-size: 13px;
  color: #909399;
  min-width: 32px;
}
.search { max-width: 220px; }
.primary-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.quick-upload {
  margin-bottom: 12px;
  border: none;
}
:deep(.quick-upload .el-collapse-item__header) {
  font-size: 13px;
  color: #606266;
  border-bottom: none;
  height: 40px;
}
:deep(.quick-upload .el-collapse-item__wrap) {
  border-bottom: none;
}
</style>

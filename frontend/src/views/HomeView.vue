<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  createdAfterISO, fetchCandidates,
  type CandidateSort, type CreatedRange,
} from '../api/client'
import type { Candidate, PipelineStats } from '../types'
import { useTabs } from '../composables/useTabs'
import { useCandidateNav } from '../composables/useCandidateNav'
import { useActiveBatch } from '../composables/useActiveBatch'
import { usePositions } from '../composables/usePositions'
import { PIPELINE_ORDER, statusLabel } from '../constants/status'
import CandidateTable from '../components/CandidateTable.vue'
import UploadDropzone from '../components/UploadDropzone.vue'
import ToolbarActions from '../components/ToolbarActions.vue'
import WorkflowGuide from '../components/WorkflowGuide.vue'
import BatchSelector from '../components/pipeline/BatchSelector.vue'
import PipelineFunnel from '../components/pipeline/PipelineFunnel.vue'
import PipelineKanban from '../components/pipeline/PipelineKanban.vue'

const { switchTab, activeKey } = useTabs()
const { setNavFromCandidates } = useCandidateNav()
const { activeBatchId } = useActiveBatch()
const { allPositions, listPositionId } = usePositions()
const tier = ref('all')
const status = ref('all')
const query = ref('')
const sort = ref<CandidateSort>('imported_desc')
const createdRange = ref<CreatedRange>('')
const isKanban = computed(() => activeKey.value === 'kanban')
const candidates = ref<Candidate[]>([])
const kanbanCandidates = ref<Candidate[]>([])
const loading = ref(false)
const uploadExpanded = ref<string[]>([])
const page = ref(1)
const pageSize = ref(20)

const pagedCandidates = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return candidates.value.slice(start, start + pageSize.value)
})
const orderOffset = computed(() => (page.value - 1) * pageSize.value)

function summarize(list: Candidate[]): PipelineStats {
  const by_status: Record<string, number> = {}
  const by_tier: Record<string, number> = {}
  for (const c of list) {
    by_status[c.status] = (by_status[c.status] || 0) + 1
    by_tier[c.tier] = (by_tier[c.tier] || 0) + 1
  }
  return { total: list.length, by_status, by_tier }
}

const listStats = computed(() => summarize(candidates.value))
const kanbanStats = computed(() => summarize(kanbanCandidates.value))

async function load() {
  loading.value = true
  try {
    if (isKanban.value) {
      const allForKanban = await fetchCandidates(
        'all',
        query.value,
        'all',
        activeBatchId.value,
        sort.value,
        createdAfterISO(createdRange.value),
      )
      kanbanCandidates.value = allForKanban
      setNavFromCandidates(allForKanban, 'all', 'all', sort.value, createdRange.value, 0)
    } else {
      // 列表是独立办事视图：不跟当前批次走，避免切回列表时被看板选中的批次悄悄缩小范围。
      const list = await fetchCandidates(
        tier.value,
        query.value,
        status.value,
        0,
        sort.value,
        createdAfterISO(createdRange.value),
        listPositionId.value,
      )
      candidates.value = list
      if (page.value > 1 && (page.value - 1) * pageSize.value >= list.length) {
        page.value = Math.max(1, Math.ceil(list.length / pageSize.value) || 1)
      }
      setNavFromCandidates(list, tier.value, status.value, sort.value, createdRange.value, listPositionId.value)
    }
  } catch (err) {
    console.error('failed to load candidates', err)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([tier, status, query, sort, createdRange, isKanban, listPositionId], () => {
  page.value = 1
  load()
})
watch(activeBatchId, () => {
  if (isKanban.value) load()
})

function onUploadDone() {
  uploadExpanded.value = []
  status.value = 'screening'
  load()
}
</script>

<template>
  <div>
    <WorkflowGuide v-if="!isKanban" />

    <el-card v-if="!isKanban" shadow="never">
      <div class="toolbar">
        <div class="filters">
          <div class="filter-row">
            <span class="list-count">
              当前筛选 <strong>{{ listStats.total }}</strong> 人
            </span>
            <a-tag color="red">S {{ listStats.by_tier?.S ?? 0 }}</a-tag>
            <a-tag color="orange">A {{ listStats.by_tier?.A ?? 0 }}</a-tag>
            <a-tag>待评 {{ listStats.by_tier?.待评 ?? 0 }}</a-tag>
            <a-tag>淘汰 {{ listStats.by_tier?.淘汰 ?? 0 }}</a-tag>
          </div>
          <div class="filter-row">
            <span class="filter-label">岗位</span>
            <el-radio-group v-model="listPositionId" size="small">
              <el-radio-button :value="0">全部</el-radio-button>
              <el-radio-button v-for="p in allPositions" :key="p.id" :value="p.id">
                {{ p.name }}
              </el-radio-button>
            </el-radio-group>
          </div>
          <div class="filter-row">
            <span class="filter-label">状态</span>
            <el-radio-group v-model="status" size="small">
              <el-radio-button value="all">全部</el-radio-button>
              <el-radio-button v-for="s in PIPELINE_ORDER" :key="s" :value="s">
                {{ statusLabel(s) }}
              </el-radio-button>
            </el-radio-group>
          </div>
          <div class="filter-row">
            <span class="filter-label">档位</span>
            <el-radio-group v-model="tier" size="small">
              <el-radio-button value="all">全部</el-radio-button>
              <el-radio-button value="S">S</el-radio-button>
              <el-radio-button value="A">A</el-radio-button>
              <el-radio-button value="待评">待评</el-radio-button>
              <el-radio-button value="淘汰">淘汰</el-radio-button>
            </el-radio-group>
          </div>
          <div class="filter-row">
            <span class="filter-label">创建时间</span>
            <el-radio-group v-model="createdRange" size="small">
              <el-radio-button value="">全部</el-radio-button>
              <el-radio-button value="24h">最近24h</el-radio-button>
              <el-radio-button value="2d">2天</el-radio-button>
              <el-radio-button value="7d">一周</el-radio-button>
            </el-radio-group>
          </div>
          <div class="filter-row">
            <span class="filter-label">排序</span>
            <el-select v-model="sort" size="small" class="sort-select">
              <el-option value="imported_desc" label="导入时间（新→旧）" />
              <el-option value="imported_asc" label="导入时间（旧→新）" />
              <el-option value="eng_first" label="工程优先" />
            </el-select>
            <el-input
              v-model="query"
              placeholder="搜索姓名 / 项目"
              clearable
              size="small"
              class="search"
            />
          </div>
        </div>

        <div class="primary-actions">
          <ToolbarActions @refresh="load" />
        </div>
      </div>

      <el-collapse v-model="uploadExpanded" class="quick-upload">
        <el-collapse-item name="upload" title="快速上传（拖拽 PDF / ZIP，归入今日批次）">
          <UploadDropzone period-type="daily" @done="onUploadDone" />
        </el-collapse-item>
      </el-collapse>

      <CandidateTable
        :candidates="pagedCandidates"
        :loading="loading"
        :show-order="tier === 'S' && status !== 'completed' && status !== 'rejected'"
        :order-offset="orderOffset"
        @updated="load"
      />
      <div v-if="candidates.length > 0" class="pager">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="candidates.length"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          background
          small
        />
      </div>
      <el-empty
        v-if="!loading && (candidates?.length ?? 0) === 0"
        description="没有匹配的候选人"
      >
        <el-button type="primary" @click="switchTab('upload')">去上传简历</el-button>
      </el-empty>
    </el-card>

    <template v-else>
      <a-card :bordered="false" class="batch-card">
        <BatchSelector @change="load" />
      </a-card>
      <el-card shadow="never">
        <div class="toolbar">
          <div class="filters">
            <div class="filter-row">
              <span class="filter-label">创建时间</span>
              <el-radio-group v-model="createdRange" size="small">
                <el-radio-button value="">全部</el-radio-button>
                <el-radio-button value="24h">最近24h</el-radio-button>
                <el-radio-button value="2d">2天</el-radio-button>
                <el-radio-button value="7d">一周</el-radio-button>
              </el-radio-group>
            </div>
            <div class="filter-row">
              <span class="filter-label">排序</span>
              <el-select v-model="sort" size="small" class="sort-select">
                <el-option value="imported_desc" label="导入时间（新→旧）" />
                <el-option value="imported_asc" label="导入时间（旧→新）" />
                <el-option value="eng_first" label="工程优先" />
              </el-select>
              <el-input
                v-model="query"
                placeholder="搜索姓名 / 项目"
                clearable
                size="small"
                class="search"
              />
            </div>
          </div>
          <div class="primary-actions">
            <ToolbarActions @refresh="load" />
          </div>
        </div>
        <PipelineFunnel :stats="kanbanStats" :loading="loading" />
        <PipelineKanban
          :candidates="kanbanCandidates"
          :loading="loading"
          @updated="load"
        />
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.batch-card {
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
  min-width: 56px;
}
.list-count {
  font-size: 13px;
  color: #606266;
}
.list-count strong {
  font-size: 16px;
  margin: 0 2px;
}
.search { max-width: 220px; }
.sort-select { width: 180px; }
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
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>

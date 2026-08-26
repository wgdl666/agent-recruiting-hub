<script setup lang="ts">
import { CircleCheck, Document } from '@element-plus/icons-vue'
import type { Candidate } from '../types'
import TierEditor from './TierEditor.vue'
import StatusEditor from './StatusEditor.vue'
import { useTabs } from '../composables/useTabs'

defineProps<{
  candidates: Candidate[]
  loading?: boolean
  showOrder?: boolean
  orderOffset?: number
}>()
const emit = defineEmits<{ updated: [] }>()
const { openCandidate } = useTabs()

function rowClass({ row }: { row: Candidate }) {
  return row.tier === 'S' ? 'row-s' : ''
}

function onRowClick(row: Candidate) {
  openCandidate(row)
}

function onTierUpdated() {
  emit('updated')
}

function formatImported(s?: string) {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}
</script>

<template>
  <el-table
    :data="candidates"
    v-loading="loading"
    stripe
    size="small"
    :header-cell-style="{ background: '#fafafa' }"
    @row-click="onRowClick"
    :row-class-name="rowClass"
    style="width: 100%; cursor: pointer"
  >
    <el-table-column v-if="showOrder" label="序" width="56">
      <!-- 序按当前筛选结果的行号连号，不展示手工 interview_order 留下的 3、5、7 空档 -->
      <template #default="{ $index }">
        {{ (orderOffset ?? 0) + $index + 1 }}
      </template>
    </el-table-column>
    <el-table-column label="状态" width="108">
      <template #default="{ row }">
        <StatusEditor :candidate="row" compact @updated="onTierUpdated" />
      </template>
    </el-table-column>
    <el-table-column label="档位" width="100">
      <template #default="{ row }">
        <TierEditor :candidate="row" compact @updated="onTierUpdated" />
      </template>
    </el-table-column>
    <el-table-column prop="name" label="姓名" width="100" />
    <el-table-column label="导入" width="108">
      <template #default="{ row }">
        <span class="imported">{{ formatImported(row.created_at) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="eng_summary" label="传统工程" min-width="160" show-overflow-tooltip />
    <el-table-column prop="project_summary" label="深挖项目" min-width="180" show-overflow-tooltip />
    <el-table-column prop="one_liner" label="摘要" min-width="120" show-overflow-tooltip />
    <el-table-column label="自动分" width="88">
      <template #default="{ row }">
        <el-tooltip v-if="row.score_total < 0" content="图片PDF或文本过少，自动分不可信，看档位与摘要">
          <el-tag size="small" type="danger">待人工</el-tag>
        </el-tooltip>
        <span v-else class="score-mini">E{{ row.eng_score }} A{{ row.agent_score }}</span>
      </template>
    </el-table-column>
    <el-table-column label="实习" width="72">
      <template #default="{ row }">
        <el-tag v-if="row.flags?.includes('no_intern')" size="small" type="info">无</el-tag>
        <el-tag v-else size="small" type="success">有</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="岗位" width="100">
      <template #default="{ row }">
        <span class="position">{{ row.position_name || '实习生' }}</span>
      </template>
    </el-table-column>
    <el-table-column label="面评" min-width="140" show-overflow-tooltip>
      <template #default="{ row }">
        <span class="note">{{ row.interview_note || '—' }}</span>
      </template>
    </el-table-column>
    <el-table-column label="简历" width="72">
      <template #default="{ row }">
        <el-icon v-if="row.has_resume" color="#67c23a"><CircleCheck /></el-icon>
        <el-icon v-else color="#c0c4cc"><Document /></el-icon>
      </template>
    </el-table-column>
  </el-table>
</template>

<style scoped>
:deep(.row-s) { background: #fef0f0 !important; }
.score-mini { font-size: 12px; color: #606266; }
.imported { font-size: 12px; color: #909399; white-space: nowrap; }
.position { font-size: 12px; color: #606266; }
.note { font-size: 12px; color: #606266; }
</style>

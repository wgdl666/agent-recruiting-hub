<script setup lang="ts">
import { computed } from 'vue'
import type { Candidate } from '../types'
import { explainScores } from '../utils/scoreExplain'

const props = defineProps<{ candidate: Candidate }>()
const ex = computed(() => explainScores(props.candidate))
</script>

<template>
  <div class="score-panel">
    <div class="row">
      <span class="label">传统工程</span>
      <el-tag size="small" type="success">{{ ex.engLabel }}</el-tag>
      <span class="hint">{{ ex.engDesc }}</span>
    </div>
    <div class="row">
      <span class="label">Agent 能力</span>
      <el-tag size="small" type="primary">{{ ex.agentLabel }}</el-tag>
      <span class="hint">{{ ex.agentDesc }}</span>
    </div>
    <div class="row">
      <span class="label">自动总分</span>
      <el-tag size="small" :type="candidate.score_total < 0 ? 'danger' : 'info'">{{ candidate.score_total }}</el-tag>
      <span class="hint">{{ ex.totalHint }}</span>
    </div>
    <el-alert
      v-if="ex.verdict"
      :title="ex.verdict"
      type="warning"
      :closable="false"
      show-icon
      class="verdict"
    />
    <div v-if="ex.reason && !ex.verdict" class="reason">系统备注：{{ ex.reason }}</div>
  </div>
</template>

<style scoped>
.score-panel { font-size: 13px; line-height: 1.6; }
.row { display: flex; align-items: flex-start; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.label { font-weight: 600; min-width: 72px; color: #303133; }
.hint { color: #909399; flex: 1; min-width: 200px; }
.verdict { margin-top: 8px; }
.reason { margin-top: 4px; color: #909399; font-size: 12px; }
</style>

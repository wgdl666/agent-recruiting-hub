<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchCandidate, updateCandidate } from '../api/client'
import type { CandidateDetail } from '../types'
import TierEditor from '../components/TierEditor.vue'
import ResumeViewer from '../components/ResumeViewer.vue'
import QuestionPanel from '../components/QuestionPanel.vue'
import StatusEditor from '../components/StatusEditor.vue'
import ScorePanel from '../components/ScorePanel.vue'
import { useTabs } from '../composables/useTabs'

const props = defineProps<{ id: string }>()
const { updateCandidateTab } = useTabs()
const detail = ref<CandidateDetail | null>(null)
const loading = ref(true)
const orderInput = ref<number | undefined>()

async function load() {
  loading.value = true
  try {
    detail.value = await fetchCandidate(Number(props.id))
    orderInput.value = detail.value.interview_order || undefined
    updateCandidateTab(detail.value.id, { name: detail.value.name, tier: detail.value.tier, status: detail.value.status })
  } finally {
    loading.value = false
  }
}

async function saveOrder() {
  if (!detail.value) return
  const updated = await updateCandidate(detail.value.id, {
    interview_order: orderInput.value || 0,
  })
  detail.value = updated
  ElMessage.success('面试顺序已更新')
}

async function onTierUpdated(c: CandidateDetail) {
  detail.value = c
  updateCandidateTab(c.id, { tier: c.tier, status: c.status })
}

async function onStatusUpdated(c: CandidateDetail) {
  detail.value = c
  updateCandidateTab(c.id, { status: c.status })
}

async function resetAutoTier() {
  if (!detail.value) return
  detail.value = await updateCandidate(detail.value.id, { clear_manual: true })
  ElMessage.info('已恢复自动档位（下次重评生效）')
}

onMounted(load)
watch(() => props.id, load)
</script>

<template>
  <div v-loading="loading" class="candidate-pane">
    <template v-if="detail">
      <div class="pane-title">
        <h2>{{ detail.name }}</h2>
        <el-tag v-if="detail.tier" :type="detail.tier === 'S' ? 'danger' : 'info'">{{ detail.tier }} 档</el-tag>
      </div>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="面试状态">
          <StatusEditor :candidate="detail" @updated="onStatusUpdated" />
        </el-descriptions-item>
        <el-descriptions-item label="档位">
          <TierEditor :candidate="detail" @updated="onTierUpdated" />
          <el-tag v-if="detail.tier_manual" size="small" type="warning" style="margin-left: 8px">手动</el-tag>
          <el-button v-if="detail.tier_manual" link type="primary" size="small" @click="resetAutoTier">恢复自动</el-button>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.tier === 'S'" label="面试顺序">
          <el-input-number v-model="orderInput" :min="0" :max="99" size="small" />
          <el-button size="small" style="margin-left: 8px" @click="saveOrder">保存</el-button>
        </el-descriptions-item>
        <el-descriptions-item label="传统工程">{{ detail.eng_summary }}</el-descriptions-item>
        <el-descriptions-item label="深挖项目">{{ detail.project_summary }}</el-descriptions-item>
        <el-descriptions-item label="摘要">{{ detail.one_liner }}</el-descriptions-item>
        <el-descriptions-item label="实习经历" :span="2">
          <template v-if="detail.flags?.includes('no_intern')">
            <el-tag type="warning">未检出</el-tag>
            <span class="field-hint">简历里没有明显的实习/上线/独立开发等关键词（可能漏检，以人工为准）</span>
          </template>
          <template v-else>
            <el-tag type="success">已检出</el-tag>
            <span class="field-hint">简历有关键词信号（实习段、上线、独立项目等），不等于已核实，面试仍需确认</span>
          </template>
        </el-descriptions-item>
        <el-descriptions-item label="建议">{{ detail.action }}</el-descriptions-item>
        <el-descriptions-item label="自动评分" :span="2">
          <ScorePanel :candidate="detail" />
        </el-descriptions-item>
        <el-descriptions-item label="来源">{{ detail.source }}</el-descriptions-item>
      </el-descriptions>

      <el-row :gutter="16" style="margin-top: 16px">
        <el-col :span="14">
          <el-card v-if="detail.has_resume" shadow="never">
            <template #header>简历预览</template>
            <ResumeViewer :candidate-id="detail.id" />
          </el-card>
          <el-empty v-else description="暂无简历文件，可重新上传" />
        </el-col>
        <el-col :span="10">
          <QuestionPanel :questions="detail.questions" />
        </el-col>
      </el-row>
    </template>
  </div>
</template>

<style scoped>
.candidate-pane { min-height: 400px; }
.pane-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.pane-title h2 { margin: 0; font-size: 20px; font-weight: 600; }
.field-hint { margin-left: 8px; color: #909399; font-size: 12px; }
</style>

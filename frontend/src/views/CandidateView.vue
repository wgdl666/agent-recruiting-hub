<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { fetchCandidate, updateCandidate } from '../api/client'
import type { CandidateDetail } from '../types'
import TierEditor from '../components/TierEditor.vue'
import ResumeViewer from '../components/ResumeViewer.vue'
import QuestionPanel from '../components/QuestionPanel.vue'
import StatusEditor from '../components/StatusEditor.vue'
import ScorePanel from '../components/ScorePanel.vue'
import { useTabs } from '../composables/useTabs'
import { useCandidateNav } from '../composables/useCandidateNav'

const props = defineProps<{ id: string }>()
const { updateCandidateTab, replaceActiveCandidate } = useTabs()
const { hasPrev, hasNext, prevId, nextId, navLabel, ensureNav, indexOf, reloadNav } = useCandidateNav()
const detail = ref<CandidateDetail | null>(null)
const loading = ref(true)
const questionsOpen = ref(false)

const currentId = computed(() => Number(props.id))
const canPrev = computed(() => hasPrev(currentId.value))
const canNext = computed(() => hasNext(currentId.value))
const positionLabel = computed(() => navLabel.value(currentId.value))
const autoOrder = computed(() => {
  const i = indexOf(currentId.value)
  return i >= 0 ? i + 1 : (detail.value?.interview_order || 0)
})

async function load() {
  loading.value = true
  try {
    await ensureNav(currentId.value)
    let data = await fetchCandidate(currentId.value)
    if (data.status === 'screening') {
      data = await updateCandidate(data.id, { status: 'read' })
    }
    detail.value = data
    updateCandidateTab(detail.value.id, {
      name: detail.value.name,
      tier: detail.value.tier,
      status: detail.value.status,
    })
  } finally {
    loading.value = false
  }
}

function goSibling(targetId: number | null) {
  if (!targetId || targetId === currentId.value) return
  replaceActiveCandidate({
    id: targetId,
    name: `候选人 #${targetId}`,
    tier: '',
    status: 'screening',
  })
}

function onKeydown(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.key === 'ArrowLeft' && canPrev.value) {
    e.preventDefault()
    goSibling(prevId(currentId.value))
  } else if (e.key === 'ArrowRight' && canNext.value) {
    e.preventDefault()
    goSibling(nextId(currentId.value))
  }
}

async function moveOrder(delta: number) {
  if (!detail.value || detail.value.tier !== 'S') return
  const next = autoOrder.value + delta
  if (next < 1) return
  const updated = await updateCandidate(detail.value.id, { interview_order: next })
  detail.value = updated
  await reloadNav()
  ElMessage.success(`面试顺序已调整为 ${indexOf(currentId.value) + 1}`)
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
  ElMessage.success('已恢复自动档位')
}

onMounted(() => {
  load()
  window.addEventListener('keydown', onKeydown)
})
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
watch(() => props.id, load)
</script>

<template>
  <div class="candidate-shell">
    <button
      type="button"
      class="nav-btn nav-prev"
      :disabled="!canPrev"
      title="上一个 (←)"
      @click="goSibling(prevId(currentId))"
    >
      <el-icon><ArrowLeft /></el-icon>
    </button>

    <div v-loading="loading" class="candidate-pane">
      <template v-if="detail">
        <div class="pane-title">
          <div>
            <h2>{{ detail.name }}</h2>
            <span v-if="positionLabel" class="position">{{ positionLabel }}</span>
          </div>
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
            <span class="auto-order">{{ autoOrder }}</span>
            <span class="field-hint">按当前列表自动连号</span>
            <el-button size="small" :disabled="autoOrder <= 1" style="margin-left: 8px" @click="moveOrder(-1)">上移</el-button>
            <el-button size="small" :disabled="!canNext" @click="moveOrder(1)">下移</el-button>
          </el-descriptions-item>
          <el-descriptions-item label="传统工程">{{ detail.eng_summary }}</el-descriptions-item>
          <el-descriptions-item label="深挖项目">{{ detail.project_summary }}</el-descriptions-item>
          <el-descriptions-item label="摘要">{{ detail.one_liner }}</el-descriptions-item>
        <el-descriptions-item label="实习经历" :span="2">
          <template v-if="detail.flags?.includes('no_intern')">
            <el-tag type="warning">未达标</el-tag>
            <span class="field-hint">标准：真实公司实习 + 项目上线/生产落地（日活、生产环境等）。仅校内/个人项目或实习无上线描述不算。</span>
          </template>
          <template v-else>
            <el-tag type="success">已达标</el-tag>
            <span class="field-hint">有真实实习且简历里有上线/落地信号，面试仍需核实细节。</span>
          </template>
        </el-descriptions-item>
          <el-descriptions-item label="建议">{{ detail.action }}</el-descriptions-item>
          <el-descriptions-item label="自动评分" :span="2">
            <ScorePanel :candidate="detail" />
          </el-descriptions-item>
          <el-descriptions-item label="来源">{{ detail.source }}</el-descriptions-item>
        </el-descriptions>

        <div class="content-row">
          <aside class="questions-side" :class="{ open: questionsOpen }">
            <button
              type="button"
              class="questions-toggle"
              :title="questionsOpen ? '收起面试题' : '展开面试题'"
              @click="questionsOpen = !questionsOpen"
            >
              <span class="toggle-icon">{{ questionsOpen ? '‹' : '›' }}</span>
              <span v-if="!questionsOpen" class="toggle-label">面试题</span>
            </button>
            <div v-show="questionsOpen" class="questions-body">
              <QuestionPanel :questions="detail.questions" />
            </div>
          </aside>

          <div class="resume-side">
            <el-card v-if="detail.has_resume" shadow="never" class="resume-card">
              <template #header>简历预览</template>
              <ResumeViewer :candidate-id="detail.id" />
            </el-card>
            <el-empty v-else description="暂无简历文件，可重新上传" />
          </div>
        </div>
      </template>
    </div>

    <button
      type="button"
      class="nav-btn nav-next"
      :disabled="!canNext"
      title="下一个 (→)"
      @click="goSibling(nextId(currentId))"
    >
      <el-icon><ArrowRight /></el-icon>
    </button>
  </div>
</template>

<style scoped>
.candidate-shell {
  display: flex;
  align-items: stretch;
  gap: 8px;
  min-height: 400px;
}
.candidate-pane {
  flex: 1;
  min-width: 0;
}
.nav-btn {
  flex: 0 0 40px;
  align-self: center;
  height: 72px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  color: #606266;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  transition: background 0.15s, border-color 0.15s;
}
.nav-btn:hover:not(:disabled) {
  background: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}
.nav-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.pane-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}
.pane-title h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
.position {
  color: #909399;
  font-size: 13px;
}
.field-hint { margin-left: 8px; color: #909399; font-size: 12px; }
.auto-order { font-weight: 600; margin-right: 4px; }

.content-row {
  display: flex;
  align-items: stretch;
  gap: 12px;
  margin-top: 16px;
  min-height: calc(100vh - 320px);
}
.questions-side {
  flex: 0 0 auto;
  display: flex;
  min-width: 36px;
  max-width: 36px;
  transition: max-width 0.2s ease;
}
.questions-side.open {
  max-width: 300px;
  min-width: 300px;
}
.questions-toggle {
  flex: 0 0 36px;
  width: 36px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #f5f7fa;
  color: #606266;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 4px;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}
.questions-toggle:hover {
  background: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}
.toggle-icon {
  font-size: 18px;
  line-height: 1;
}
.toggle-label {
  writing-mode: vertical-rl;
  font-size: 12px;
  letter-spacing: 2px;
}
.questions-body {
  flex: 1;
  min-width: 0;
  overflow: auto;
}
.resume-side {
  flex: 1;
  min-width: 0;
}
.resume-card {
  height: 100%;
}
.resume-card :deep(.el-card__body) {
  padding: 0 12px 12px;
}
</style>

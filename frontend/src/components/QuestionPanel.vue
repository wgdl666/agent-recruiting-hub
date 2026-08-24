<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { InterviewQuestion } from '../types'
import { generateQuestions } from '../api/client'
import { renderMarkdown } from '../utils/markdown'
import axios from 'axios'

const props = defineProps<{
  questions: InterviewQuestion[]
  candidateId: number
  tier?: string
}>()

const emit = defineEmits<{
  updated: [questions: InterviewQuestion[]]
}>()

const generating = ref(false)
const active = ref<InterviewQuestion | null>(null)

const dialogVisible = computed({
  get: () => active.value != null,
  set: (v: boolean) => {
    if (!v) active.value = null
  },
})

function levelColor(level?: string) {
  if (level === 'L1') return 'blue'
  if (level === 'L2') return 'orange'
  if (level === 'L3') return 'red'
  return 'default'
}

function openAnswer(q: InterviewQuestion) {
  active.value = q
}

const answerHtml = computed(() => {
  if (!active.value?.answer) return ''
  return renderMarkdown(active.value.answer)
})

async function onRegenerate() {
  await ElMessageBox.confirm(
    '将按当前简历重新出 3 道题，并写给不一定熟悉该方向的面试官看的参考答案。需调用模型，大约半分钟。',
    '按简历重新生成',
    { confirmButtonText: '生成', cancelButtonText: '取消', type: 'warning' },
  )
  generating.value = true
  try {
    const detail = await generateQuestions(props.candidateId)
    emit('updated', detail.questions || [])
    ElMessage.success('面试题已更新')
  } catch (err: unknown) {
    let msg = '生成失败'
    if (axios.isAxiosError(err)) {
      const data = err.response?.data as { error?: string } | undefined
      msg = data?.error || err.message || msg
    }
    ElMessage.error(msg)
  } finally {
    generating.value = false
  }
}
</script>

<template>
  <el-card shadow="never" class="panel">
    <template #header>
      <div class="header">
        <span>定制面试题（L1→L2→L3 深挖）</span>
        <el-button
          v-if="tier === 'S'"
          size="small"
          :loading="generating"
          @click="onRegenerate"
        >
          按简历重新生成
        </el-button>
      </div>
    </template>
    <el-empty v-if="!questions.length" description="暂无定制题，S 档候选人会按简历自动生成" />
    <div v-else class="qa-list">
      <button
        v-for="q in questions"
        :key="q.id"
        type="button"
        class="qa-item"
        @click="openAnswer(q)"
      >
        <div class="qa-title">
          <a-tag v-if="q.level" :color="levelColor(q.level)" size="small">{{ q.level }}</a-tag>
          <span>{{ q.question }}</span>
        </div>
        <span class="qa-action">{{ q.answer ? '查看参考答案' : '暂无参考答案' }}</span>
      </button>
    </div>
  </el-card>

  <el-dialog
    v-model="dialogVisible"
    class="answer-dialog"
    width="720px"
    top="8vh"
    append-to-body
    destroy-on-close
  >
    <template #header>
      <div class="dialog-head">
        <a-tag v-if="active?.level" :color="levelColor(active.level)" size="small">{{ active.level }}</a-tag>
        <span>{{ active?.question }}</span>
      </div>
    </template>
    <div v-if="active?.answer" class="answer-body markdown-body" v-html="answerHtml" />
    <el-text v-else type="info">暂无参考答案。S 档可点「按简历重新生成」。 </el-text>
  </el-dialog>
</template>

<style scoped>
.panel { margin-top: 0; }
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.qa-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.qa-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  width: 100%;
  text-align: left;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #fafafa;
  padding: 10px 12px;
  cursor: pointer;
}
.qa-item:hover {
  border-color: #409eff;
  background: #f5f9ff;
}
.qa-title {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  line-height: 1.5;
  color: #303133;
}
.qa-action {
  font-size: 12px;
  color: #409eff;
}
.dialog-head {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  line-height: 1.5;
  padding-right: 24px;
  font-size: 15px;
  font-weight: 600;
}
.answer-body {
  font-size: 14px;
  line-height: 1.75;
  color: #303133;
  max-height: min(70vh, 640px);
  overflow: auto;
}
.markdown-body :deep(h3) {
  font-size: 15px;
  margin: 1.1em 0 0.4em;
}
.markdown-body :deep(p),
.markdown-body :deep(li) {
  line-height: 1.75;
}
.markdown-body :deep(ul) {
  padding-left: 1.2em;
}
.markdown-body :deep(code) {
  background: #f5f5f5;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 0.9em;
}
</style>

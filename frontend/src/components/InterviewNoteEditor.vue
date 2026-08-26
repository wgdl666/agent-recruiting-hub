<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { updateCandidate } from '../api/client'
import type { Candidate, CandidateDetail } from '../types'

const props = defineProps<{ candidate: Pick<Candidate, 'id' | 'interview_note'> }>()
const emit = defineEmits<{ updated: [CandidateDetail] }>()

const draft = ref(props.candidate.interview_note || '')
const saving = ref(false)
const savedFlash = ref(false)

watch(
  () => [props.candidate.id, props.candidate.interview_note] as const,
  ([, note]) => {
    draft.value = note || ''
  },
)

const dirty = computed(() => (draft.value || '').trim() !== (props.candidate.interview_note || '').trim())

async function save() {
  if (!dirty.value || saving.value) return
  saving.value = true
  try {
    const updated = await updateCandidate(props.candidate.id, { interview_note: draft.value })
    draft.value = updated.interview_note || ''
    emit('updated', updated)
    savedFlash.value = true
    window.setTimeout(() => { savedFlash.value = false }, 1500)
  } catch {
    ElMessage.error('面评保存失败')
  } finally {
    saving.value = false
  }
}

function onBlur() {
  void save()
}

function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    void save()
  }
}
</script>

<template>
  <div class="note-editor" @click.stop>
    <el-input
      v-model="draft"
      type="textarea"
      :autosize="{ minRows: 3, maxRows: 10 }"
      placeholder="面试后写下评价：项目是否站得住、工程意识、风险、是否过…"
      :disabled="saving"
      @blur="onBlur"
      @keydown="onKeydown"
    />
    <div class="note-bar">
      <span class="hint">失焦或 ⌘/Ctrl + Enter 保存；重评不会覆盖</span>
      <span v-if="savedFlash" class="saved">已保存</span>
      <el-button size="small" type="primary" :loading="saving" :disabled="!dirty" @click="save">保存</el-button>
    </div>
  </div>
</template>

<style scoped>
.note-editor { width: 100%; }
.note-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
.hint { color: #909399; font-size: 12px; flex: 1; }
.saved { color: #67c23a; font-size: 12px; }
</style>

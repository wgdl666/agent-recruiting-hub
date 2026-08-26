<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchPosition, updatePosition } from '../api/client'
import type { Position } from '../types'
import { usePositions } from '../composables/usePositions'
import { renderMarkdown } from '../utils/markdown'

const { jdPositionId, loadPositions } = usePositions()
const position = ref<Position | null>(null)
const loading = ref(false)
const editing = ref(false)
const draft = ref('')
const saving = ref(false)

const html = computed(() => renderMarkdown(position.value?.jd || ''))

async function load() {
  if (!jdPositionId.value) {
    position.value = null
    return
  }
  loading.value = true
  editing.value = false
  try {
    position.value = await fetchPosition(jdPositionId.value)
    draft.value = position.value.jd || ''
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载岗位 JD 失败')
  } finally {
    loading.value = false
  }
}

watch(jdPositionId, load, { immediate: true })

function startEdit() {
  draft.value = position.value?.jd || ''
  editing.value = true
}

function cancelEdit() {
  draft.value = position.value?.jd || ''
  editing.value = false
}

async function save() {
  if (!position.value) return
  saving.value = true
  try {
    position.value = await updatePosition(position.value.id, { jd: draft.value })
    editing.value = false
    await loadPositions()
    ElMessage.success('JD 已保存')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div v-loading="loading" class="jd-page">
    <template v-if="position">
      <div class="head">
        <div>
          <h2>{{ position.name }}</h2>
          <p class="sub">岗位 JD · 检验标准 {{ position.skill_name || position.skill_id }}</p>
        </div>
        <div class="actions">
          <el-tag :type="position.is_open ? 'success' : 'info'" size="small">
            {{ position.is_open ? '在招' : '已停' }}
          </el-tag>
          <el-button v-if="!editing" type="primary" @click="startEdit">编辑 JD</el-button>
          <template v-else>
            <el-button @click="cancelEdit">取消</el-button>
            <el-button type="primary" :loading="saving" @click="save">保存</el-button>
          </template>
        </div>
      </div>

      <el-input
        v-if="editing"
        v-model="draft"
        type="textarea"
        :autosize="{ minRows: 16, maxRows: 36 }"
        placeholder="用 Markdown 写岗位职责、要求、加分项…"
      />
      <article v-else-if="position.jd" class="markdown-body" v-html="html" />
      <el-empty v-else description="还没有岗位 JD">
        <el-button type="primary" @click="startEdit">写 JD</el-button>
      </el-empty>
    </template>
  </div>
</template>

<style scoped>
.jd-page { max-width: 800px; }
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}
h2 { margin: 0 0 6px; font-size: 20px; }
.sub { margin: 0; font-size: 13px; color: #909399; }
.actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.markdown-body :deep(h1) {
  font-size: 1.5em;
  margin: 0.4em 0 0.6em;
}
.markdown-body :deep(h2) {
  font-size: 1.2em;
  margin: 1.3em 0 0.5em;
  padding-bottom: 0.3em;
  border-bottom: 1px solid #eee;
}
.markdown-body :deep(h3) {
  font-size: 1.05em;
  margin: 1.1em 0 0.4em;
}
.markdown-body :deep(p),
.markdown-body :deep(li) {
  line-height: 1.75;
  color: #434343;
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  margin: 12px 0;
  width: 100%;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #ebeef5;
  padding: 8px 10px;
  text-align: left;
}
.markdown-body :deep(th) { background: #fafafa; }
</style>

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
  const id = jdPositionId.value
  if (!id) {
    position.value = null
    return
  }
  loading.value = true
  editing.value = false
  try {
    const next = await fetchPosition(id)
    // 抽屉里连点两个岗位时，只采用仍指向当前岗的响应，避免旧 JD 盖住新岗。
    if (jdPositionId.value !== id) return
    position.value = next
    draft.value = next.jd || ''
  } catch (e: unknown) {
    if (jdPositionId.value !== id) return
    ElMessage.error(e instanceof Error ? e.message : '加载岗位 JD 失败')
  } finally {
    if (jdPositionId.value === id) loading.value = false
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
  <div v-loading="loading" class="jd-body">
    <template v-if="position">
      <div class="toolbar">
        <p class="sub">检验标准 {{ position.skill_name || position.skill_id }}</p>
        <div class="actions">
          <el-tag :type="position.is_open ? 'success' : 'info'" size="small">
            {{ position.is_open ? '在招' : '已停' }}
          </el-tag>
          <el-button v-if="!editing" type="primary" size="small" @click="startEdit">编辑 JD</el-button>
          <template v-else>
            <el-button size="small" @click="cancelEdit">取消</el-button>
            <el-button type="primary" size="small" :loading="saving" @click="save">保存</el-button>
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
.jd-body { min-height: 240px; }
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.sub { margin: 0; font-size: 13px; color: #909399; }
.actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.markdown-body :deep(h1) {
  font-size: 1.45em;
  margin: 0 0 0.7em;
  line-height: 1.35;
}
.markdown-body :deep(h2) {
  font-size: 1.15em;
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
.markdown-body :deep(strong) { color: #303133; }
.markdown-body :deep(table) {
  border-collapse: collapse;
  margin: 12px 0;
  width: 100%;
  font-size: 13px;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #ebeef5;
  padding: 8px 10px;
  text-align: left;
}
.markdown-body :deep(th) { background: #fafafa; }
.markdown-body :deep(code) {
  background: #f5f5f5;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 0.9em;
}
</style>

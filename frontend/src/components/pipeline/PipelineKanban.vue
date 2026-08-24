<script setup lang="ts">
import { computed, ref } from 'vue'
import { message } from 'ant-design-vue'
import { updateCandidate } from '../../api/client'
import type { Candidate } from '../../types'
import { PIPELINE_ORDER, statusColor, statusLabel } from '../../constants/status'
import { useTabs } from '../../composables/useTabs'

const props = defineProps<{
  candidates: Candidate[]
  loading?: boolean
}>()
const emit = defineEmits<{ updated: [] }>()
const { openCandidate } = useTabs()

const dragId = ref<number | null>(null)
const saving = ref(false)

const grouped = computed(() => {
  const map: Record<string, Candidate[]> = {}
  for (const s of PIPELINE_ORDER) map[s] = []
  for (const c of props.candidates) {
    const key = PIPELINE_ORDER.includes(c.status as typeof PIPELINE_ORDER[number])
      ? c.status
      : 'screening'
    map[key].push(c)
  }
  return map
})

function onDragStart(id: number) {
  dragId.value = id
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
}

async function onDrop(status: string) {
  const id = dragId.value
  dragId.value = null
  if (!id) return
  const cand = props.candidates.find((c) => c.id === id)
  if (!cand || cand.status === status) return
  saving.value = true
  try {
    await updateCandidate(id, { status })
    message.success(`${cand.name} → ${statusLabel(status)}`)
    emit('updated')
  } catch {
    message.error('状态更新失败')
  } finally {
    saving.value = false
  }
}

function tierColor(tier: string) {
  if (tier === 'S') return 'red'
  if (tier === 'A') return 'orange'
  if (tier === '淘汰') return 'default'
  return 'blue'
}
</script>

<template>
  <a-spin :spinning="loading || saving">
    <div class="kanban">
      <div
        v-for="status in PIPELINE_ORDER"
        :key="status"
        class="column"
        @dragover="onDragOver"
        @drop="onDrop(status)"
      >
        <div class="col-head">
          <a-tag :color="statusColor(status)">{{ statusLabel(status) }}</a-tag>
          <span class="col-count">{{ grouped[status]?.length ?? 0 }}</span>
        </div>
        <div class="col-body">
          <a-card
            v-for="c in grouped[status]"
            :key="c.id"
            size="small"
            class="card"
            draggable="true"
            @dragstart="onDragStart(c.id)"
            @click="openCandidate(c)"
          >
            <div class="card-top">
              <strong>{{ c.name }}</strong>
              <a-tag :color="tierColor(c.tier)" size="small">{{ c.tier }}</a-tag>
            </div>
            <div class="card-sub">{{ c.one_liner || c.eng_summary }}</div>
            <div v-if="c.batch_name" class="card-batch">{{ c.batch_name }}</div>
          </a-card>
          <a-empty v-if="!grouped[status]?.length" :image-style="{ height: 40 }" description="暂无" />
        </div>
      </div>
    </div>
  </a-spin>
</template>

<style scoped>
.kanban {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 8px;
  min-height: 360px;
}
.column {
  flex: 0 0 200px;
  background: #f5f5f5;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  max-height: 70vh;
}
.col-head {
  padding: 8px 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e8e8e8;
}
.col-count {
  font-size: 12px;
  color: #999;
}
.col-body {
  padding: 8px;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.card {
  cursor: grab;
}
.card:active {
  cursor: grabbing;
}
.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 6px;
}
.card-sub {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-batch {
  font-size: 11px;
  color: #aaa;
  margin-top: 4px;
}
</style>

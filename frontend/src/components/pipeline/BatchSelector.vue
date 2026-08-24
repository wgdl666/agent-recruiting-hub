<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { createBatch, fetchBatches } from '../../api/client'
import type { Batch } from '../../types'
import { useActiveBatch } from '../../composables/useActiveBatch'

const emit = defineEmits<{ change: [] }>()
const { activeBatchId, uploadPeriodType } = useActiveBatch()
const batches = ref<Batch[]>([])
const loading = ref(false)

const selectValue = computed({
  get: () => (activeBatchId.value > 0 ? activeBatchId.value : 'all'),
  set: (v: number | 'all') => {
    activeBatchId.value = v === 'all' ? 0 : v
    emit('change')
  },
})

const options = computed(() => [
  { value: 'all', label: '全部批次' },
  ...batches.value.map((b) => ({
    value: b.id,
    label: `${b.name}（${b.stats?.total ?? 0}）`,
  })),
])

async function load() {
  loading.value = true
  try {
    batches.value = await fetchBatches(true)
    if (activeBatchId.value > 0 && !batches.value.some((b) => b.id === activeBatchId.value)) {
      activeBatchId.value = 0
    }
  } finally {
    loading.value = false
  }
}

async function ensureBatch(period: string) {
  loading.value = true
  try {
    const { id } = await createBatch({ auto: true, period_type: period })
    uploadPeriodType.value = period
    activeBatchId.value = id
    await load()
    message.success('已切换到当前批次')
    emit('change')
  } catch {
    message.error('创建批次失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(activeBatchId, () => emit('change'))

defineExpose({ reload: load })
</script>

<template>
  <div class="batch-bar">
    <div class="left">
      <span class="label">招聘批次</span>
      <a-select
        v-model:value="selectValue"
        :loading="loading"
        style="min-width: 220px"
        :options="options"
      />
      <a-radio-group v-model:value="uploadPeriodType" size="small" button-style="solid">
        <a-radio-button value="daily">按日</a-radio-button>
        <a-radio-button value="weekly">按周</a-radio-button>
        <a-radio-button value="monthly">按月</a-radio-button>
      </a-radio-group>
    </div>
    <a-space>
      <a-button size="small" @click="ensureBatch('daily')">今日批次</a-button>
      <a-button size="small" @click="ensureBatch('weekly')">本周批次</a-button>
      <a-button size="small" @click="ensureBatch('monthly')">本月批次</a-button>
      <a-button size="small" type="link" @click="load">刷新</a-button>
    </a-space>
  </div>
  <a-typography-text type="secondary" class="hint">
    新上传的简历会归入「{{ uploadPeriodType === 'weekly' ? '周' : uploadPeriodType === 'monthly' ? '月' : '日' }}」批次；也可在上方选定批次后只查看该批进度。
  </a-typography-text>
</template>

<style scoped>
.batch-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.label {
  font-size: 13px;
  color: #666;
  white-space: nowrap;
}
.hint {
  display: block;
  font-size: 12px;
  margin-bottom: 12px;
}
</style>

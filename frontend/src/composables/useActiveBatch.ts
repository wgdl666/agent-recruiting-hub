import { ref, watch } from 'vue'

const STORAGE_KEY = 'recruiting-hub:active-batch'
const PERIOD_KEY = 'recruiting-hub:upload-period'

function readBatch(): number {
  const v = localStorage.getItem(STORAGE_KEY)
  return v ? Number(v) : 0
}

function readPeriod(): string {
  return localStorage.getItem(PERIOD_KEY) || 'daily'
}

const activeBatchId = ref(readBatch())
const uploadPeriodType = ref(readPeriod())

watch(activeBatchId, (v) => {
  localStorage.setItem(STORAGE_KEY, String(v))
})

watch(uploadPeriodType, (v) => {
  localStorage.setItem(PERIOD_KEY, v)
})

export function useActiveBatch() {
  return { activeBatchId, uploadPeriodType }
}

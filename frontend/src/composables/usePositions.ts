import { ref } from 'vue'
import { fetchPositions } from '../api/client'
import type { Position } from '../types'

const openPositions = ref<Position[]>([])
const allPositions = ref<Position[]>([])
const listPositionId = ref(0)
const jdPositionId = ref(0)

export function usePositions() {
  async function loadPositions() {
    const [open, all] = await Promise.all([fetchPositions(true), fetchPositions(false)])
    openPositions.value = open
    allPositions.value = all
    if (listPositionId.value > 0 && !all.some((p) => p.id === listPositionId.value)) {
      listPositionId.value = 0
    }
    if (jdPositionId.value > 0 && !all.some((p) => p.id === jdPositionId.value)) {
      jdPositionId.value = 0
    }
  }

  return { openPositions, allPositions, listPositionId, jdPositionId, loadPositions }
}

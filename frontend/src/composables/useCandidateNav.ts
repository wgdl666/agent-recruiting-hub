import { computed, ref } from 'vue'
import type { Candidate } from '../types'
import { fetchCandidates } from '../api/client'
import { useActiveBatch } from './useActiveBatch'

const navIds = ref<number[]>([])
const navTier = ref('S')
const navStatus = ref('screening')

export function useCandidateNav() {
  function setNavFromCandidates(list: Candidate[], tier = navTier.value, status = navStatus.value) {
    navIds.value = list.map((c) => c.id)
    navTier.value = tier
    navStatus.value = status
  }

  function indexOf(id: number) {
    return navIds.value.indexOf(id)
  }

  const hasPrev = (id: number) => indexOf(id) > 0
  const hasNext = (id: number) => {
    const i = indexOf(id)
    return i >= 0 && i < navIds.value.length - 1
  }

  const prevId = (id: number) => {
    const i = indexOf(id)
    return i > 0 ? navIds.value[i - 1] : null
  }

  const nextId = (id: number) => {
    const i = indexOf(id)
    return i >= 0 && i < navIds.value.length - 1 ? navIds.value[i + 1] : null
  }

  const navLabel = computed(() => {
    return (id: number) => {
      const i = indexOf(id)
      if (i < 0 || navIds.value.length === 0) return ''
      return `${i + 1} / ${navIds.value.length}`
    }
  })

  async function ensureNav(currentId: number) {
    if (navIds.value.length > 0 && indexOf(currentId) >= 0) return
    const { activeBatchId } = useActiveBatch()
    const list = await fetchCandidates(navTier.value, '', navStatus.value, activeBatchId.value)
    setNavFromCandidates(list, navTier.value, navStatus.value)
  }

  return {
    navIds,
    setNavFromCandidates,
    hasPrev,
    hasNext,
    prevId,
    nextId,
    navLabel,
    ensureNav,
  }
}

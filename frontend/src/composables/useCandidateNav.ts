import { computed, ref } from 'vue'
import type { Candidate } from '../types'
import { fetchCandidates, type CandidateSort } from '../api/client'
import { useActiveBatch } from './useActiveBatch'

/** 左侧固定选择栏只需要切换身份的字段，避免把整份简历摘要挂在导航状态里 */
export type NavCandidate = Pick<Candidate, 'id' | 'name' | 'tier' | 'status'>

const navList = ref<NavCandidate[]>([])
const navTier = ref('all')
const navStatus = ref('all')
const navSort = ref<CandidateSort>('imported_desc')

export function useCandidateNav() {
  function setNavFromCandidates(
    list: Candidate[] | null | undefined,
    tier = navTier.value,
    status = navStatus.value,
    sort: CandidateSort = navSort.value,
  ) {
    navList.value = (list ?? []).map((c) => ({
      id: c.id,
      name: c.name,
      tier: c.tier,
      status: c.status,
    }))
    navTier.value = tier
    navStatus.value = status
    navSort.value = sort
  }

  function indexOf(id: number) {
    return navList.value.findIndex((c) => c.id === id)
  }

  const hasPrev = (id: number) => indexOf(id) > 0
  const hasNext = (id: number) => {
    const i = indexOf(id)
    return i >= 0 && i < navList.value.length - 1
  }

  const prevId = (id: number) => {
    const i = indexOf(id)
    return i > 0 ? navList.value[i - 1].id : null
  }

  const nextId = (id: number) => {
    const i = indexOf(id)
    return i >= 0 && i < navList.value.length - 1 ? navList.value[i + 1].id : null
  }

  const navLabel = computed(() => {
    return (id: number) => {
      const i = indexOf(id)
      if (i < 0 || navList.value.length === 0) return ''
      return `${i + 1} / ${navList.value.length}`
    }
  })

  async function ensureNav(currentId: number) {
    if (navList.value.length > 0 && indexOf(currentId) >= 0) return
    await reloadNav()
  }

  async function reloadNav() {
    const { activeBatchId } = useActiveBatch()
    const list = await fetchCandidates(navTier.value, '', navStatus.value, activeBatchId.value, navSort.value)
    setNavFromCandidates(list, navTier.value, navStatus.value, navSort.value)
  }

  return {
    navList,
    setNavFromCandidates,
    indexOf,
    hasPrev,
    hasNext,
    prevId,
    nextId,
    navLabel,
    ensureNav,
    reloadNav,
  }
}

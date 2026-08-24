import { ref } from 'vue'
import type { Candidate } from '../types'

export type AppTab =
  | { key: 'home'; title: '候选人'; kind: 'home' }
  | { key: 'upload'; title: '上传'; kind: 'upload' }
  | { key: 'docs'; title: '知识库'; kind: 'docs' }
  | { key: string; title: string; kind: 'candidate'; candidateId: number; tier?: string; status?: string }

const fixedTabs: AppTab[] = [
  { key: 'home', title: '候选人', kind: 'home' },
  { key: 'upload', title: '上传', kind: 'upload' },
  { key: 'docs', title: '知识库', kind: 'docs' },
]

const candidateTabs = ref<Extract<AppTab, { kind: 'candidate' }>[]>([])
const activeKey = ref('home')

export function useTabs() {
  function allTabs(): AppTab[] {
    return [...fixedTabs, ...candidateTabs.value]
  }

  function openCandidate(c: Pick<Candidate, 'id' | 'name' | 'tier' | 'status'>) {
    const key = `candidate-${c.id}`
    const existing = candidateTabs.value.find((t) => t.key === key)
    if (existing) {
      existing.title = c.name
      existing.tier = c.tier
      existing.status = c.status
    } else {
      candidateTabs.value.push({
        key,
        title: c.name,
        kind: 'candidate',
        candidateId: c.id,
        tier: c.tier,
        status: c.status,
      })
    }
    activeKey.value = key
  }

  function openCandidateById(id: number, name = `候选人 #${id}`, tier?: string, status?: string) {
    openCandidate({ id, name, tier: tier ?? '', status: status ?? 'screening' })
  }

  function switchTab(key: string) {
    activeKey.value = key
  }

  function removeTab(key: string) {
    if (key === 'home' || key === 'upload' || key === 'docs') return
    const idx = candidateTabs.value.findIndex((t) => t.key === key)
    if (idx < 0) return
    candidateTabs.value.splice(idx, 1)
    if (activeKey.value === key) {
      activeKey.value = candidateTabs.value.length ? candidateTabs.value[candidateTabs.value.length - 1].key : 'home'
    }
  }

  function updateCandidateTab(id: number, patch: { name?: string; tier?: string; status?: string }) {
    const tab = candidateTabs.value.find((t) => t.candidateId === id)
    if (!tab) return
    if (patch.name) tab.title = patch.name
    if (patch.tier) tab.tier = patch.tier
    if (patch.status) tab.status = patch.status
  }

  return {
    activeKey,
    candidateTabs,
    allTabs,
    openCandidate,
    openCandidateById,
    switchTab,
    removeTab,
    updateCandidateTab,
  }
}

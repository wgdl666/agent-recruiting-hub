import { ref, watch } from 'vue'
import type { Candidate } from '../types'

export type WorkspaceKey = 'list' | 'kanban'

export type AppTab =
  | { key: 'list'; title: '列表'; kind: 'list' }
  | { key: 'kanban'; title: '看板'; kind: 'kanban' }
  | { key: 'upload'; title: '上传'; kind: 'upload' }
  | { key: 'docs'; title: '知识库'; kind: 'docs' }
  | { key: string; title: string; kind: 'candidate'; candidateId: number; tier?: string; status?: string }

const WORKSPACE_KEY = 'recruiting-hub:view-mode'

function readWorkspace(): WorkspaceKey {
  return localStorage.getItem(WORKSPACE_KEY) === 'kanban' ? 'kanban' : 'list'
}

const lastWorkspaceKey = ref<WorkspaceKey>(readWorkspace())
const candidateTabs = ref<Extract<AppTab, { kind: 'candidate' }>[]>([])
const activeKey = ref<string>(lastWorkspaceKey.value)

watch(lastWorkspaceKey, (v) => localStorage.setItem(WORKSPACE_KEY, v))

function isWorkspace(key: string): key is WorkspaceKey {
  return key === 'list' || key === 'kanban'
}

function rememberWorkspace(key: string) {
  if (isWorkspace(key)) lastWorkspaceKey.value = key
}

export function useTabs() {
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
    // 旧入口 home = 候选人列表；关详情时回到上次的列表或看板。
    if (key === 'home') key = 'list'
    rememberWorkspace(key)
    activeKey.value = key
  }

  function removeTab(key: string) {
    if (key === 'list' || key === 'kanban' || key === 'upload' || key === 'docs' || key === 'home') return
    const idx = candidateTabs.value.findIndex((t) => t.key === key)
    if (idx < 0) return
    candidateTabs.value.splice(idx, 1)
    if (activeKey.value === key) {
      activeKey.value = candidateTabs.value.length
        ? candidateTabs.value[candidateTabs.value.length - 1].key
        : lastWorkspaceKey.value
    }
  }

  function updateCandidateTab(id: number, patch: { name?: string; tier?: string; status?: string }) {
    const tab = candidateTabs.value.find((t) => t.candidateId === id)
    if (!tab) return
    if (patch.name) tab.title = patch.name
    if (patch.tier) tab.tier = patch.tier
    if (patch.status) tab.status = patch.status
  }

  function replaceActiveCandidate(c: Pick<Candidate, 'id' | 'name' | 'tier' | 'status'>) {
    const key = `candidate-${c.id}`
    const idx = candidateTabs.value.findIndex((t) => t.key === activeKey.value)
    const tab = {
      key,
      title: c.name,
      kind: 'candidate' as const,
      candidateId: c.id,
      tier: c.tier,
      status: c.status,
    }
    if (idx >= 0) {
      candidateTabs.value[idx] = tab
    } else {
      candidateTabs.value.push(tab)
    }
    activeKey.value = key
  }

  return {
    activeKey,
    lastWorkspaceKey,
    candidateTabs,
    openCandidate,
    openCandidateById,
    replaceActiveCandidate,
    switchTab,
    removeTab,
    updateCandidateTab,
  }
}

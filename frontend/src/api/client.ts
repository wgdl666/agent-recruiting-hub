import axios from 'axios'
import type { Batch, Candidate, CandidateDetail, DocDetail, DocEntry, EvalSkill, PipelineStats, Position, Stats, UploadResult } from '../types'

const api = axios.create({ baseURL: '/api' })

function asList<T>(data: T[] | null | undefined): T[] {
  return Array.isArray(data) ? data : []
}

export async function fetchStats() {
  const { data } = await api.get<Stats>('/stats')
  return data
}

export type CandidateSort = 'imported_desc' | 'imported_asc' | 'eng_first'
export type CreatedRange = '' | '24h' | '2d' | '7d'

/** 「最近 24h / 2天 / 一周」转成 UTC RFC3339 下界，和库里 created_at 文本比较。 */
export function createdAfterISO(range: CreatedRange): string | undefined {
  if (!range) return undefined
  const hours = range === '24h' ? 24 : range === '2d' ? 48 : 168
  return new Date(Date.now() - hours * 3600_000).toISOString().replace(/\.\d{3}Z$/, 'Z')
}

export async function fetchCandidates(
  tier = 'all',
  q = '',
  status = 'all',
  batchId = 0,
  sort: CandidateSort = 'imported_desc',
  createdAfter?: string,
  positionId = 0,
) {
  const { data } = await api.get<Candidate[] | null>('/candidates', {
    params: {
      tier: tier === 'all' ? '' : tier,
      status: status === 'all' ? '' : status,
      q,
      batch_id: batchId > 0 ? batchId : undefined,
      position_id: positionId > 0 ? positionId : undefined,
      sort,
      created_after: createdAfter || undefined,
    },
  })
  return asList(data)
}

export async function fetchBatches(active = true) {
  const { data } = await api.get<Batch[] | null>('/batches', { params: { active } })
  return asList(data)
}

export async function createBatch(opts: { auto?: boolean; period_type?: string; name?: string; tag?: string }) {
  const { data } = await api.post<{ id: number }>('/batches', opts)
  return data
}

export async function fetchPipelineStats(batchId = 0) {
  const { data } = await api.get<PipelineStats>('/pipeline/stats', {
    params: { batch_id: batchId > 0 ? batchId : undefined },
  })
  return data
}

export async function fetchDocList() {
  const { data } = await api.get<DocEntry[]>('/docs')
  return data
}

export async function fetchDoc(slug: string) {
  const { data } = await api.get<DocDetail>(`/docs/${slug}`)
  return data
}

export async function fetchSkills() {
  const { data } = await api.get<EvalSkill[] | null>('/skills')
  return asList(data)
}

export async function fetchPositions(openOnly = false) {
  const { data } = await api.get<Position[] | null>('/positions', {
    params: openOnly ? { open: 'true' } : undefined,
  })
  return asList(data)
}

export async function createPosition(payload: { name: string; skill_id: string; description?: string; is_open?: boolean }) {
  const { data } = await api.post<Position>('/positions', payload)
  return data
}

export async function updatePosition(
  id: number,
  patch: { name?: string; skill_id?: string; description?: string; is_open?: boolean; sort_order?: number },
) {
  const { data } = await api.patch<Position>(`/positions/${id}`, patch)
  return data
}

export async function fetchCandidate(id: number) {
  const { data } = await api.get<CandidateDetail>(`/candidates/${id}`)
  return data
}

export function resumeUrl(id: number) {
  return `/api/candidates/${id}/resume`
}

export async function uploadFiles(
  files: File[],
  source = 'upload',
  opts?: { batchId?: number; periodType?: string; positionId?: number; skillId?: string },
) {
  const form = new FormData()
  files.forEach((f) => form.append('files', f))
  form.append('source', source)
  if (opts?.positionId && opts.positionId > 0) {
    form.append('position_id', String(opts.positionId))
  } else if (opts?.skillId) {
    form.append('skill_id', opts.skillId)
  }
  if (opts?.batchId && opts.batchId > 0) {
    form.append('batch_id', String(opts.batchId))
  } else if (opts?.periodType) {
    form.append('period_type', opts.periodType)
  }
  const { data } = await api.post<UploadResult>('/upload', form)
  return data
}

export async function importSeed() {
  const { data } = await api.post<{ imported: number }>('/import/seed')
  return data
}

export async function rescreenAll() {
  const { data } = await api.post<{ updated: number }>('/rescreen')
  return data
}

export async function syncQuestions() {
  const { data } = await api.post<{ updated: number }>('/sync/questions', null, { timeout: 300000 })
  return data
}

export async function generateQuestions(id: number) {
  const { data } = await api.post<CandidateDetail>(`/candidates/${id}/questions/generate`, null, { timeout: 180000 })
  return data
}

export async function exportFeishu() {
  const { data } = await api.post<{ ok: boolean; lark?: { data?: { document?: { url?: string } } }; markdown?: string; error?: string }>('/export/feishu')
  return data
}

export async function exportMarkdown() {
  const { data } = await api.get<string>('/export/markdown', { responseType: 'text' as never })
  const blob = new Blob([data as unknown as string], { type: 'text/markdown' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'screening-report.md'
  a.click()
  URL.revokeObjectURL(url)
}

export async function updateCandidate(
  id: number,
  patch: { tier?: string; action?: string; interview_order?: number; status?: string; clear_manual?: boolean },
) {
  const { data } = await api.patch<CandidateDetail>(`/candidates/${id}`, patch)
  return data
}

export async function exportAll() {
  const { data } = await api.get('/export')
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'candidates-export.json'
  a.click()
  URL.revokeObjectURL(url)
}

import axios from 'axios'
import type { Batch, Candidate, CandidateDetail, DocDetail, DocEntry, PipelineStats, Stats, UploadResult } from '../types'

const api = axios.create({ baseURL: '/api' })

function asList<T>(data: T[] | null | undefined): T[] {
  return Array.isArray(data) ? data : []
}

export async function fetchStats() {
  const { data } = await api.get<Stats>('/stats')
  return data
}

export type CandidateSort = 'imported_desc' | 'imported_asc' | 'eng_first'

export async function fetchCandidates(
  tier = 'all',
  q = '',
  status = 'all',
  batchId = 0,
  sort: CandidateSort = 'imported_desc',
) {
  const { data } = await api.get<Candidate[] | null>('/candidates', {
    params: {
      tier: tier === 'all' ? '' : tier,
      status: status === 'all' ? '' : status,
      q,
      batch_id: batchId > 0 ? batchId : undefined,
      sort,
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

export async function fetchCandidate(id: number) {
  const { data } = await api.get<CandidateDetail>(`/candidates/${id}`)
  return data
}

export function resumeUrl(id: number) {
  return `/api/candidates/${id}/resume`
}

export async function uploadFiles(files: File[], source = 'upload', opts?: { batchId?: number; periodType?: string }) {
  const form = new FormData()
  files.forEach((f) => form.append('files', f))
  form.append('source', source)
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
  const { data } = await api.post<{ updated: number }>('/sync/questions')
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

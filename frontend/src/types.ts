export interface Candidate {
  id: number
  name: string
  source: string
  tier: string
  eng_summary: string
  project_summary: string
  one_liner: string
  action: string
  resume_path?: string
  score_total: number
  eng_score: number
  agent_score: number
  reason?: string
  flags?: string[]
  status: string
  tier_manual?: boolean
  interview_order?: number
  batch_id?: number
  batch_name?: string
  has_resume: boolean
}

export interface InterviewQuestion {
  id: number
  candidate_id: number
  sort_order: number
  question: string
  answer?: string
  level?: string
}

export interface CandidateDetail extends Candidate {
  questions: InterviewQuestion[]
}

export interface Stats {
  total: number
  by_tier: Record<string, number>
  by_status: Record<string, number>
}

export interface BatchStats {
  total: number
  by_status: Record<string, number>
  by_tier: Record<string, number>
}

export interface Batch {
  id: number
  name: string
  tag: string
  period_type: string
  is_active: boolean
  created_at: string
  updated_at: string
  stats?: BatchStats
}

export interface PipelineStats {
  batch_id?: number
  batch_name?: string
  total: number
  by_status: Record<string, number>
  by_tier: Record<string, number>
}

export interface DocEntry {
  slug: string
  title: string
  order: number
}

export interface DocDetail {
  slug: string
  title: string
  markdown: string
}

export interface UploadResult {
  imported: number
  batch_id?: number
  results: ScreenResult[]
  errors?: string[]
}

export interface ScreenResult {
  name: string
  tier: string
  score_total: number
  eng_score: number
  agent_score: number
  reason: string
  candidate_id?: number
}

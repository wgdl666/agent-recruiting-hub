export type CandidateStatus =
  | 'screening'
  | 'read'
  | 'to_interview'
  | 'interviewing'
  | 'passed'
  | 'completed'
  | 'rejected'

export const PIPELINE_ORDER: CandidateStatus[] = [
  'screening',
  'read',
  'to_interview',
  'interviewing',
  'passed',
  'completed',
  'rejected',
]

export const STATUS_OPTIONS: {
  value: CandidateStatus
  label: string
  type: '' | 'success' | 'warning' | 'info' | 'danger'
  color: string
}[] = [
  { value: 'screening', label: '简历筛选', type: 'info', color: 'default' },
  { value: 'read', label: '已阅', type: 'info', color: 'blue' },
  { value: 'to_interview', label: '待约面', type: 'info', color: 'processing' },
  { value: 'interviewing', label: '面试中', type: 'warning', color: 'warning' },
  { value: 'passed', label: '通过', type: 'success', color: 'success' },
  { value: 'completed', label: '面试完成', type: 'success', color: 'cyan' },
  { value: 'rejected', label: '本轮淘汰', type: 'danger', color: 'error' },
]

export function statusLabel(s: string) {
  return STATUS_OPTIONS.find((o) => o.value === s)?.label ?? '简历筛选'
}

export function statusType(s: string) {
  return STATUS_OPTIONS.find((o) => o.value === s)?.type ?? 'info'
}

export function statusColor(s: string) {
  return STATUS_OPTIONS.find((o) => o.value === s)?.color ?? 'default'
}

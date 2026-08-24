import type { Candidate } from '../types'

/** 自动评分说明（0–4 为维度分，不是百分制） */
export function explainScores(c: Pick<Candidate, 'eng_score' | 'agent_score' | 'score_total' | 'reason' | 'flags' | 'tier' | 'tier_manual' | 'auto_tier'>) {
  const engMax = 4
  const agentMax = 4
  const thin = c.flags?.includes('thin')
  const noIntern = c.flags?.includes('no_intern')

  const engLabel = `${c.eng_score}/${engMax} 维`
  const agentLabel = `${c.agent_score}/${agentMax} 维`

  let totalHint = ''
  if (c.score_total < 0) {
    totalHint = thin ? 'PDF 文本过少（多为图片简历），自动分不可信' : '自动评分偏低'
  } else if (c.score_total >= 20) {
    totalHint = '自动评分较高，建议结合面试验证'
  } else {
    totalHint = '自动评分一般，以面试深挖为准'
  }

  let verdict = ''
  if (thin && (c.tier_manual || c.tier === 'S' || c.tier === 'A')) {
    verdict = '档位已人工/种子数据校正，请以下方「传统工程」「深挖项目」和简历为准，勿看自动分。'
  } else if (thin) {
    verdict = '图片版 PDF，系统读不出文字，需人工看简历或 OCR。'
  } else if (c.tier_manual) {
    const auto = c.auto_tier?.trim()
    if (auto && auto !== c.tier) {
      verdict = `档位已手动锁定为 ${c.tier}；自动建议 ${auto} 档。上传/重评会继续更新自动分，不改手动档位。`
    } else {
      verdict = '档位已手动锁定；上传/重评会继续更新自动分，不改手动档位。'
    }
  }

  return {
    engLabel,
    agentLabel,
    engDesc: '传统工程：后端栈、负责落地、可靠性、接口设计等关键词是否出现',
    agentDesc: 'Agent：Agent 框架、RAG、Prompt/评测等关键词是否出现',
    totalHint,
    verdict,
    thin,
    noIntern,
    reason: c.reason || (thin ? 'thin（简历文本过少）' : ''),
  }
}

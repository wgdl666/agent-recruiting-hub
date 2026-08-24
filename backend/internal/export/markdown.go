package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/caden/agent-recruiting-hub/internal/seed"
)

func MarkdownReport(candidates []models.CandidateDetail) string {
	var b strings.Builder
	now := time.Now().Format("2006-01-02 15:04")
	fmt.Fprintf(&b, "# Agent 实习招聘筛选报告\n\n> 生成时间：%s · 标准：实习经历 + 传统工程底座 + 单项目深挖\n\n", now)

	byTier := map[string][]models.CandidateDetail{}
	for _, c := range candidates {
		byTier[c.Tier] = append(byTier[c.Tier], c)
	}

	if ss := byTier["S"]; len(ss) > 0 {
		b.WriteString("## S 档 — 优先面试\n\n")
		b.WriteString("| 序 | 姓名 | 传统工程 | 深挖项目 | 摘要 | E/A |\n")
		b.WriteString("|---:|------|----------|----------|------|-----|\n")
		for _, c := range sortS(ss) {
			order := seed.InterviewOrder[c.Name]
			fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %d/%d |\n",
				order, c.Name, c.EngSummary, c.ProjectSummary, c.OneLiner, c.EngScore, c.AgentScore)
		}
		b.WriteString("\n### S 档定制面试题\n\n")
		for _, c := range sortS(ss) {
			if len(c.Questions) == 0 {
				continue
			}
			fmt.Fprintf(&b, "#### %s\n\n", c.Name)
			for i, q := range c.Questions {
				fmt.Fprintf(&b, "%d. %s\n", i+1, q.Question)
			}
			b.WriteByte('\n')
		}
	}

	for _, tier := range []string{"A", "B", "C", "淘汰"} {
		list := byTier[tier]
		if len(list) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s 档（%d）\n\n", tier, len(list))
		b.WriteString("| 姓名 | 状态 | 传统工程 | 深挖项目 | 摘要 | 建议 |\n")
		b.WriteString("|------|------|----------|----------|------|------|\n")
		for _, c := range list {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
				c.Name, models.StatusLabel(c.Status), c.EngSummary, c.ProjectSummary, c.OneLiner, c.Action)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func sortS(list []models.CandidateDetail) []models.CandidateDetail {
	out := make([]models.CandidateDetail, len(list))
	copy(out, list)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			oi := seed.InterviewOrder[out[i].Name]
			oj := seed.InterviewOrder[out[j].Name]
			if oi == 0 {
				oi = 999
			}
			if oj == 0 {
				oj = 999
			}
			if oj < oi {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

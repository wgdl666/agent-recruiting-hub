package scanner

import (
	"fmt"
	"strings"
	"time"
)

// InterviewQA is one generated question plus a briefing for the interviewer.
type InterviewQA struct {
	Level    string
	Question string
	Answer   string
}

type llmQuestionsPayload struct {
	Questions []struct {
		Level    string `json:"level"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"questions"`
}

// GenerateInterviewQA writes L1–L3 questions a non-specialist interviewer can run.
func GenerateInterviewQA(name, engSummary, projectSummary, resumeText string) ([]InterviewQA, error) {
	client := modelHubClient()
	if client == nil {
		return nil, errNoModelHub
	}
	resumeText = strings.TrimSpace(resumeText)
	if resumeText == "" {
		return nil, fmt.Errorf("empty resume text")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "候选人：%s\n", strings.TrimSpace(name))
	if s := strings.TrimSpace(engSummary); s != "" {
		fmt.Fprintf(&b, "工程摘要：%s\n", s)
	}
	if s := strings.TrimSpace(projectSummary); s != "" {
		fmt.Fprintf(&b, "深挖项目：%s\n", s)
	}
	b.WriteString("\n简历正文：\n")
	b.WriteString(truncateResume(resumeText, 80000))

	var out llmQuestionsPayload
	// 参考答案要写细，输出比评分 JSON 长很多。
	if err := client.GenerateJSONOpts(questionsPrompt, b.String(), &out, 8192, 120*time.Second); err != nil {
		return nil, err
	}
	items := normalizeInterviewQA(out)
	if len(items) == 0 {
		return nil, fmt.Errorf("modelhub returned no interview questions")
	}
	return items, nil
}

func normalizeInterviewQA(out llmQuestionsPayload) []InterviewQA {
	seen := map[string]bool{}
	items := make([]InterviewQA, 0, 3)
	for _, q := range out.Questions {
		level := normalizeQALevel(q.Level)
		question := strings.TrimSpace(q.Question)
		answer := strings.TrimSpace(q.Answer)
		if question == "" || answer == "" {
			continue
		}
		if seen[level] {
			continue
		}
		seen[level] = true
		items = append(items, InterviewQA{Level: level, Question: question, Answer: answer})
		if len(items) == 3 {
			break
		}
	}
	return sortQALevels(items)
}

func normalizeQALevel(level string) string {
	s := strings.ToUpper(strings.TrimSpace(level))
	s = strings.ReplaceAll(s, "L", "")
	switch s {
	case "1":
		return "L1"
	case "2":
		return "L2"
	default:
		return "L3"
	}
}

func sortQALevels(items []InterviewQA) []InterviewQA {
	order := map[string]int{"L1": 0, "L2": 1, "L3": 2}
	sorted := append([]InterviewQA(nil), items...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if order[sorted[j].Level] < order[sorted[i].Level] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

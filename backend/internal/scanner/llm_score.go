package scanner

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/llm"
)

var (
	modelHubOnce sync.Once
	modelHubCli  *llm.ModelHub
)

func modelHubClient() *llm.ModelHub {
	modelHubOnce.Do(func() {
		modelHubCli = llm.NewFromEnv()
	})
	return modelHubCli
}

// ModelHubEnabled reports whether HUB_MODELHUB_ADDRESS is set.
func ModelHubEnabled() bool {
	return modelHubClient() != nil
}

// GeminiEnabled is an alias for ModelHubEnabled (legacy health field).
func GeminiEnabled() bool {
	return ModelHubEnabled()
}

type llmScorePayload struct {
	EngScore       int      `json:"eng_score"`
	AgentScore     int      `json:"agent_score"`
	InternScore    int      `json:"intern_score"`
	DepthScore     int      `json:"depth_score"`
	ScoreTotal     int      `json:"score_total"`
	Tier           string   `json:"tier"`
	Flags          []string `json:"flags"`
	Reason         string   `json:"reason"`
	EngSummary     string   `json:"eng_summary"`
	ProjectSummary string   `json:"project_summary"`
	OneLiner       string   `json:"one_liner"`
	Action         string   `json:"action"`
}

const visionScoreUser = `这是候选人简历的 PDF/页面截图（可能是图片版简历，文字层为空）。请阅读图像内容，按系统要求的 JSON 打分。看不清的字段不要编造。`

// ScoreFromResume 先抽文字；不够再把 PDF/截图交给模型看图评分。仍不够才标待评，不淘汰。
func ScoreFromResume(path, skillID string) (text string, score ScoreBreakdown, err error) {
	text, err = ExtractText(path)
	if err != nil {
		text = ""
	}
	if textEnough(text) {
		return text, ScoreUploadForSkill(text, skillID), nil
	}
	if score, ok := tryModelHubScorePDF(path, skillID); ok {
		if strings.TrimSpace(score.Text) != "" {
			text = betterText(text, score.Text)
		}
		return text, score, nil
	}
	if score, ok := tryModelHubScorePages(path, skillID); ok {
		return text, score, nil
	}
	log.Printf("scanner: resume text thin (%d runes), leave in screening: %s", len([]rune(strings.TrimSpace(text))), path)
	return text, thinScore(text), nil
}

func scoreWithModelHub(text, skillID string) (ScoreBreakdown, error) {
	client := modelHubClient()
	if client == nil {
		return ScoreBreakdown{}, errNoModelHub
	}

	user := truncateResume(text, 120000)
	var out llmScorePayload
	if err := client.GenerateJSON(scoringPromptFor(skillID), user, &out); err != nil {
		return ScoreBreakdown{}, err
	}
	return normalizeLLMScore(text, out, false), nil
}

func tryModelHubScorePDF(path, skillID string) (ScoreBreakdown, bool) {
	client := modelHubClient()
	if client == nil {
		return ScoreBreakdown{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 || len(data) > 12<<20 {
		return ScoreBreakdown{}, false
	}
	var out llmScorePayload
	if err := client.GenerateJSONParts(scoringPromptFor(skillID), visionScoreUser, []llm.UserMedia{{
		MIME: "application/pdf",
		Data: data,
	}}, &out, 4096, 120*time.Second); err != nil {
		log.Printf("scanner: modelhub pdf score failed: %v", err)
		return ScoreBreakdown{}, false
	}
	return normalizeLLMScore("", out, true), true
}

var errNoModelHub = errors.New("modelhub not configured")

func normalizeLLMScore(text string, out llmScorePayload, fromDocument bool) ScoreBreakdown {
	flags := out.Flags
	if flags == nil {
		flags = []string{}
	}
	if !fromDocument && !textEnough(text) {
		flags = appendUnique(flags, "thin")
		out.Tier = "待评"
		out.ScoreTotal = -10
		out.Reason = "thin"
		if strings.TrimSpace(out.OneLiner) == "" {
			out.OneLiner = "图片简历抽字不足，待筛选"
		}
	}
	if contains(flags, "thin") || looksUnreadable(out.Reason, out.OneLiner) {
		flags = appendUnique(flags, "thin")
		out.Tier = "待评"
		if out.ScoreTotal >= 0 {
			out.ScoreTotal = -10
		}
		if strings.TrimSpace(out.OneLiner) == "" {
			out.OneLiner = "图片简历抽字不足，待筛选"
		}
	}

	tier := normalizeTier(out.Tier)
	action := strings.TrimSpace(out.Action)
	if action == "" {
		action = ActionForTier(tier)
	}

	reason := strings.TrimSpace(out.Reason)
	if reason == "" {
		reason = "modelhub"
	}

	return ScoreBreakdown{
		Text:           text,
		EngScore:       clamp(out.EngScore, 0, 4),
		AgentScore:     clamp(out.AgentScore, 0, 4),
		InternScore:    clamp(out.InternScore, 0, 4),
		DepthScore:     clamp(out.DepthScore, 0, 3),
		Total:          out.ScoreTotal,
		Tier:           tier,
		Reason:         reason,
		Flags:          flags,
		EngSummary:     strings.TrimSpace(out.EngSummary),
		ProjectSummary: strings.TrimSpace(out.ProjectSummary),
		OneLiner:       strings.TrimSpace(out.OneLiner),
		Action:         action,
		Source:         "modelhub",
	}
}

func normalizeTier(t string) string {
	return NormalizeTier(t)
}

func truncateResume(text string, maxRunes int) string {
	r := []rune(text)
	if len(r) <= maxRunes {
		return text
	}
	return string(r[:maxRunes]) + "\n\n[简历已截断]"
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func appendUnique(flags []string, s string) []string {
	if contains(flags, s) {
		return flags
	}
	return append(flags, s)
}

func tryModelHubScore(text, skillID string) (ScoreBreakdown, bool) {
	if modelHubClient() == nil {
		return ScoreBreakdown{}, false
	}
	score, err := scoreWithModelHub(text, skillID)
	if err != nil {
		// port-forward 偶发 reset，重试一次再退回启发式。
		score, err = scoreWithModelHub(text, skillID)
	}
	if err != nil {
		log.Printf("scanner: modelhub score failed, fallback heuristic: %v", err)
		return ScoreBreakdown{}, false
	}
	return score, true
}

func tryModelHubScorePages(path, skillID string) (ScoreBreakdown, bool) {
	client := modelHubClient()
	if client == nil {
		return ScoreBreakdown{}, false
	}
	pages := renderPDFPages(path, 3)
	if len(pages) == 0 {
		return ScoreBreakdown{}, false
	}
	media := make([]llm.UserMedia, 0, len(pages))
	for _, p := range pages {
		media = append(media, llm.UserMedia{MIME: "image/png", Data: p})
	}
	var out llmScorePayload
	if err := client.GenerateJSONParts(scoringPromptFor(skillID), visionScoreUser, media, &out, 4096, 0); err != nil {
		log.Printf("scanner: modelhub page score failed: %v", err)
		return ScoreBreakdown{}, false
	}
	return normalizeLLMScore("", out, true), true
}

func looksUnreadable(parts ...string) bool {
	blob := strings.Join(parts, " ")
	for _, k := range []string{"内容极少", "内容为空", "无法评估", "抽字不足", "看不清"} {
		if strings.Contains(blob, k) {
			return true
		}
	}
	return false
}

package scanner

import (
	"errors"
	"log"
	"strings"
	"sync"

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

func scoreWithModelHub(text string) (ScoreBreakdown, error) {
	client := modelHubClient()
	if client == nil {
		return ScoreBreakdown{}, errNoModelHub
	}

	user := truncateResume(text, 120000)
	var out llmScorePayload
	if err := client.GenerateJSON(scoringPrompt, user, &out); err != nil {
		return ScoreBreakdown{}, err
	}
	return normalizeLLMScore(text, out), nil
}

var errNoModelHub = errors.New("modelhub not configured")

func normalizeLLMScore(text string, out llmScorePayload) ScoreBreakdown {
	flags := out.Flags
	if flags == nil {
		flags = []string{}
	}
	if len(strings.TrimSpace(text)) < 300 {
		flags = appendUnique(flags, "thin")
		out.Tier = "淘汰"
		out.ScoreTotal = -10
		out.Reason = "thin"
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

func tryModelHubScore(text string) (ScoreBreakdown, bool) {
	if modelHubClient() == nil {
		return ScoreBreakdown{}, false
	}
	score, err := scoreWithModelHub(text)
	if err != nil {
		log.Printf("scanner: modelhub score failed, fallback heuristic: %v", err)
		return ScoreBreakdown{}, false
	}
	return score, true
}

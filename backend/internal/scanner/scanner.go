package scanner

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

var internPatterns = map[string]*regexp.Regexp{
	"intern_section": regexp.MustCompile(`实习经历|实习经验|工作经历|工作/项目经历|实习单位`),
	"intern_role":    regexp.MustCompile(`(?i)实习|实习生|intern`),
	"ownership":      regexp.MustCompile(`(?i)独立(开发|负责|设计|实现)|主导|负责.*(开发|实现|设计|架构)|核心开发`),
	"equivalent":     regexp.MustCompile(`(?i)上线|已上线|上线运营|投产|生产环境|真实用户|内部使用|日活|活跃用户|sole dev|独立(完成|开发|交付)|个人项目.*上线`),
}

var engPatterns = map[string]*regexp.Regexp{
	"backend_stack": regexp.MustCompile(`(?i)Redis|MySQL|PostgreSQL|MongoDB|Kafka|RabbitMQ|消息队列|队列|微服务|gRPC|REST|FastAPI|Flask|Go|Gin|Spring|Docker|K8s|Kubernetes|部署|并发|幂等|分布式`),
	"ownership":     regexp.MustCompile(`(?i)独立(开发|负责|设计|实现)|从零|主导|负责.*(开发|实现|设计|架构)|自研|上线|生产环境|完整链路`),
	"reliability":   regexp.MustCompile(`(?i)超时|重试|降级|容错|高可用|监控|日志|Trace|压测|QPS|延迟|P99|性能优化|耗时|P95`),
	"api_design":    regexp.MustCompile(`(?i)接口|API|schema|鉴权|限流|中间件|ORM`),
}

var agentPatterns = map[string]*regexp.Regexp{
	"agent_core": regexp.MustCompile(`(?i)Agent|ReAct|Function Calling|tool call|MCP|LangGraph|LangChain|AutoGen|多Agent|workflow|工作流|编排`),
	"rag":        regexp.MustCompile(`(?i)RAG|向量|embedding|检索|知识库|chunk|rerank|Milvus|Faiss|Elasticsearch`),
	"llm_eng":    regexp.MustCompile(`(?i)prompt|Prompt|微调|fine-tun|评测|eval|轨迹|guardrail|上下文|token`),
}

var metricPattern = regexp.MustCompile(`\d+[%％]|\d+→\d+|P9[59]|QPS|MRR|Recall|准确率|压缩|耗时|降幅|\d+\s*秒`)

var keywordStuff = regexp.MustCompile(`(?i)精通.*精通|熟练掌握.*熟练掌握`)

type ScoreBreakdown struct {
	Text           string
	EngScore       int
	AgentScore     int
	InternScore    int
	DepthScore     int
	Total          int
	Tier           string
	Reason         string
	Flags          []string
	EngSummary     string
	ProjectSummary string
	OneLiner       string
	Action         string
	Source         string // "gemini" | "heuristic"
}

func ExtractText(path string) (string, error) {
	text, err := extractPDFText(path)
	if err != nil || len(strings.TrimSpace(text)) < 200 {
		if ocr, ocrErr := tryOCR(path); ocrErr == nil && len(ocr) > 200 {
			return ocr, nil
		}
		if err != nil {
			return "", err
		}
		return text, nil
	}
	return text, nil
}

func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	total := r.NumPage()
	for i := 1; i <= total; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteByte('\n')
	}
	return buf.String(), nil
}

func tryOCR(path string) (string, error) {
	script := filepath.Join("scripts", "ocr_resume.py")
	cmd := exec.Command("python3", script, path)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ExtractName(filename, text string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	if m := regexp.MustCompile(`】(.+?)\s+\d`).FindStringSubmatch(base); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := regexp.MustCompile(`^(.+?)[-_ ]`).FindStringSubmatch(base); len(m) > 1 && len([]rune(m[1])) <= 6 {
		return strings.TrimSpace(m[1])
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines[:min(8, len(lines))] {
		line = strings.TrimSpace(line)
		if len([]rune(line)) >= 2 && len([]rune(line)) <= 4 && !strings.ContainsAny(line, "0123456789@") {
			return line
		}
	}
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

func Score(text string) ScoreBreakdown {
	return ScoreUpload(text)
}

// ScoreUpload runs ModelHub when configured (upload / screen paths only).
func ScoreUpload(text string) ScoreBreakdown {
	text = strings.TrimSpace(text)
	if len(text) < 300 {
		return thinScore(text)
	}
	if score, ok := tryModelHubScore(text); ok {
		return score
	}
	return scoreHeuristic(text)
}

// ScoreHeuristic keyword scoring without ModelHub (bulk rescreen only).
func ScoreHeuristic(text string) ScoreBreakdown {
	text = strings.TrimSpace(text)
	if len(text) < 300 {
		return thinScore(text)
	}
	return scoreHeuristic(text)
}

func thinScore(text string) ScoreBreakdown {
	return ScoreBreakdown{
		Text: text, Total: -10, Tier: "淘汰", Reason: "thin", Flags: []string{"thin"},
		Action: ActionForTier("淘汰"), Source: "heuristic",
	}
}

func scoreHeuristic(text string) ScoreBreakdown {
	flags := []string{}

	if keywordStuff.MatchString(text) {
		flags = append(flags, "keyword_stuff")
	}

	eng := 0
	for _, re := range engPatterns {
		if re.MatchString(text) {
			eng++
		}
	}
	agent := 0
	for _, re := range agentPatterns {
		if re.MatchString(text) {
			agent++
		}
	}

	intern := 0
	for _, re := range internPatterns {
		if re.MatchString(text) {
			intern++
		}
	}
	if !hasRealInternship(text) {
		flags = append(flags, "no_intern")
	}

	depth := 0
	if metricPattern.MatchString(text) {
		depth++
	}
	if strings.Contains(text, "LangGraph") || strings.Contains(text, "Harness") {
		depth++
	}
	if strings.Contains(text, "checkpoint") || strings.Contains(text, "幂等") {
		depth++
	}

	total := eng*3 + agent*3 + depth*2 + intern*2
	if contains(flags, "keyword_stuff") {
		total -= 15
	}
	if contains(flags, "no_intern") {
		total -= 8
	}
	if len(text) < 800 {
		total -= 5
	}

	tier, reason := tierFromScores(eng, agent, intern, depth, total, flags)
	return ScoreBreakdown{
		Text: text, EngScore: eng, AgentScore: agent, InternScore: intern, DepthScore: depth,
		Total: total, Tier: tier, Reason: reason, Flags: flags,
		Action: ActionForTier(tier), Source: "heuristic",
	}
}

func tierFromScores(eng, agent, intern, depth, total int, flags []string) (string, string) {
	if contains(flags, "thin") {
		return "淘汰", "thin"
	}
	if contains(flags, "keyword_stuff") && eng < 2 {
		return "淘汰", "keyword_stuff"
	}
	if contains(flags, "no_intern") && eng < 2 && agent < 2 {
		return "淘汰", "no_intern"
	}
	if eng >= 3 && agent >= 2 && depth >= 2 && intern >= 3 && !contains(flags, "no_intern") {
		return "S", fmt.Sprintf("eng=%d agent=%d intern=%d depth=%d", eng, agent, intern, depth)
	}
	if eng >= 2 && agent >= 2 && intern >= 2 && !contains(flags, "no_intern") {
		return "A", fmt.Sprintf("eng=%d agent=%d intern=%d", eng, agent, intern)
	}
	if eng >= 2 || agent >= 2 {
		return "A", fmt.Sprintf("score=%d", total)
	}
	if total < 5 {
		return "淘汰", "weak"
	}
	return "淘汰", fmt.Sprintf("score=%d", total)
}

func NormalizeTier(tier string) string {
	switch strings.TrimSpace(tier) {
	case "S", "A", "淘汰":
		return strings.TrimSpace(tier)
	case "B":
		return "A"
	case "C":
		return "淘汰"
	default:
		return "淘汰"
	}
}

func hasRealInternship(text string) bool {
	hasSection := internPatterns["intern_section"].MatchString(text)
	hasRole := internPatterns["intern_role"].MatchString(text)
	hasLaunch := internPatterns["equivalent"].MatchString(text) ||
		internPatterns["ownership"].MatchString(text)
	return hasSection && hasRole && hasLaunch
}

func ActionForTier(tier string) string {
	switch NormalizeTier(tier) {
	case "S":
		return "优先排期"
	case "A":
		return "第二批"
	default:
		return "暂不约"
	}
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

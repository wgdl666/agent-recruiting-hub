package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/caden/agent-recruiting-hub/internal/scanner"
	"github.com/caden/agent-recruiting-hub/internal/seed"
	"github.com/gin-gonic/gin"
)

func toSeedQA(items []scanner.InterviewQA) []seed.QA {
	out := make([]seed.QA, 0, len(items))
	for _, q := range items {
		out = append(out, seed.QA{Level: q.Level, Question: q.Question, Answer: q.Answer})
	}
	return out
}

func (s *Server) resumeTextOf(c *models.Candidate) string {
	if c == nil {
		return ""
	}
	local := s.st.ResolveResumePath(c.ResumePath)
	if local == "" {
		return ""
	}
	text, err := scanner.ExtractText(local)
	if err != nil {
		return ""
	}
	return text
}

// tryWriteInterviewQuestions 给 S 档出题：优先 LLM 写细答案，失败再回落种子题。
func (s *Server) tryWriteInterviewQuestions(id int64, name, tier, eng, proj, resumeText, skillID string) {
	qs, err := s.buildInterviewQuestions(name, tier, eng, proj, resumeText, skillID)
	if err != nil {
		log.Printf("interview questions %s: %v", name, err)
		return
	}
	if len(qs) == 0 {
		return
	}
	if err := s.st.SetQuestions(id, qs); err != nil {
		log.Printf("save interview questions %s: %v", name, err)
	}
}

func (s *Server) buildInterviewQuestions(name, tier, eng, proj, resumeText, skillID string) ([]seed.QA, error) {
	resumeText = strings.TrimSpace(resumeText)
	var genErr error
	// 只有 S 档走模型：题要务实，答案要写给不一定懂这块的面试官。
	if scanner.NormalizeTier(tier) == "S" && resumeText != "" && scanner.ModelHubEnabled() {
		items, err := scanner.GenerateInterviewQA(name, eng, proj, resumeText, skillID)
		if err == nil && len(items) > 0 {
			return toSeedQA(items), nil
		}
		genErr = err
		if err != nil {
			log.Printf("llm interview questions %s: %v", name, err)
		}
	}
	if qs, ok := seed.Questions[name]; ok {
		return qs, nil
	}
	if genErr != nil {
		return nil, genErr
	}
	if scanner.NormalizeTier(tier) != "S" {
		return nil, nil
	}
	if resumeText == "" {
		return nil, fmt.Errorf("no resume text")
	}
	return nil, fmt.Errorf("modelhub not configured")
}

func (s *Server) generateQuestions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	d, err := s.st.GetCandidate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "candidate not found"})
		return
	}
	text := s.resumeTextOf(&d.Candidate)
	qs, err := s.buildInterviewQuestions(d.Name, d.Tier, d.EngSummary, d.ProjectSummary, text, d.SkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(qs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅 S 档生成定制题"})
		return
	}
	if err := s.st.SetQuestions(id, qs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	d, err = s.st.GetCandidate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

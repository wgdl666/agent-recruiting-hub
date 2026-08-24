package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/export"
	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/caden/agent-recruiting-hub/internal/scanner"
	"github.com/caden/agent-recruiting-hub/internal/seed"
	"github.com/caden/agent-recruiting-hub/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	st   *store.Store
	root string
}

func New(st *store.Store, root string) *Server {
	return &Server{st: st, root: root}
}

func (s *Server) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/docs", s.listDocs)
		api.GET("/docs/:slug", s.getDoc)
		api.GET("/health", s.health)
		api.GET("/batches", s.listBatches)
		api.POST("/batches", s.createBatch)
		api.GET("/pipeline/stats", s.pipelineStats)
		api.GET("/stats", s.stats)
		api.GET("/candidates", s.listCandidates)
		api.GET("/candidates/:id", s.getCandidate)
		api.GET("/candidates/:id/resume", s.getResume)
		api.PATCH("/candidates/:id", s.patchCandidate)
		api.DELETE("/candidates/:id", s.deleteCandidate)
		api.POST("/upload", s.uploadFiles)
		api.POST("/screen", s.screenPath)
		api.POST("/import/seed", s.importSeed)
		api.POST("/rescreen", s.rescreenAll)
		api.POST("/sync/questions", s.syncQuestions)
		api.POST("/export/feishu", s.exportFeishu)
		api.GET("/export/markdown", s.exportMarkdown)
		api.GET("/export", s.exportAll)
	}

	dist := filepath.Join(s.root, "frontend", "dist")
	if info, err := os.Stat(dist); err == nil && info.IsDir() {
		r.Static("/assets", filepath.Join(dist, "assets"))
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(filepath.Join(dist, "index.html"))
		})
	}

	return r
}

func (s *Server) health(c *gin.Context) {
	n, _ := s.st.Count()
	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"candidates": n,
		"db":         s.st.DBPath(),
		"resumes":    s.st.ResumeDir(),
	})
}

func (s *Server) stats(c *gin.Context) {
	st, err := s.st.Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, st)
}

func (s *Server) listCandidates(c *gin.Context) {
	tier := c.Query("tier")
	status := c.Query("status")
	q := c.Query("q")
	engFirst := c.DefaultQuery("eng_first", "true") == "true"
	batchID, _ := strconv.ParseInt(c.Query("batch_id"), 10, 64)
	list, err := s.st.ListCandidates(tier, status, q, batchID, engFirst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *Server) getCandidate(c *gin.Context) {
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
	c.JSON(http.StatusOK, d)
}

func (s *Server) patchCandidate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Tier           *string `json:"tier"`
		Action         *string `json:"action"`
		InterviewOrder *int    `json:"interview_order"`
		Status         *string `json:"status"`
		ClearManual    bool    `json:"clear_manual"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Tier == nil && req.Action == nil && req.InterviewOrder == nil && req.Status == nil && !req.ClearManual {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	before, _ := s.st.GetCandidate(id)
	if err := s.st.PatchCandidate(id, store.PatchInput{
		Tier: req.Tier, Action: req.Action, InterviewOrder: req.InterviewOrder, Status: req.Status, ClearManual: req.ClearManual,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Tier != nil && *req.Tier == "S" && before != nil {
		if qs, ok := seed.Questions[before.Name]; ok {
			_ = s.st.SetQuestions(id, qs)
		}
	}
	d, err := s.st.GetCandidate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (s *Server) getResume(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	d, err := s.st.GetCandidate(id)
	if err != nil || d.ResumePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "resume not found"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(d.ResumePath)))
	c.File(d.ResumePath)
}

func (s *Server) deleteCandidate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.st.DeleteCandidate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) exportAll(c *gin.Context) {
	data, err := s.st.ExportAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="candidates-export.json"`)
	c.JSON(http.StatusOK, data)
}

func (s *Server) uploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files"})
		return
	}
	source := c.DefaultPostForm("source", "upload")
	batchID, err := s.resolveUploadBatch(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch_id"})
		return
	}
	result := models.UploadResult{BatchID: batchID}

	for _, fh := range files {
		tmp, err := os.CreateTemp("", "upload-*"+filepath.Ext(fh.Filename))
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		src, err := fh.Open()
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		_, _ = io.Copy(tmp, src)
		src.Close()
		tmp.Close()

		ext := strings.ToLower(filepath.Ext(fh.Filename))
		switch ext {
		case ".zip":
			rs, errs := s.processZip(tmp.Name(), source, batchID)
			result.Results = append(result.Results, rs...)
			result.Errors = append(result.Errors, errs...)
			result.Imported += len(rs)
		case ".pdf":
			sr, err := s.processPDF(tmp.Name(), fh.Filename, source, batchID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", fh.Filename, err))
			} else {
				result.Results = append(result.Results, *sr)
				result.Imported++
			}
		default:
			result.Errors = append(result.Errors, fmt.Sprintf("unsupported: %s", fh.Filename))
		}
		os.Remove(tmp.Name())
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) screenPath(c *gin.Context) {
	var req struct {
		Path       string `json:"path"`
		Source     string `json:"source"`
		BatchID    int64  `json:"batch_id"`
		PeriodType string `json:"period_type"`
	}
	if err := c.BindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}
	if req.Source == "" {
		req.Source = "cursor"
	}
	batchID := req.BatchID
	if batchID <= 0 {
		pt := req.PeriodType
		if pt == "" {
			pt = "daily"
		}
		batchID, _ = s.st.GetOrCreateBatchForDate(time.Now(), pt)
	}
	info, err := os.Stat(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var results []models.ScreenResult
	var errs []string
	if info.IsDir() {
		_ = filepath.Walk(req.Path, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() || strings.ToLower(filepath.Ext(path)) != ".pdf" {
				return nil
			}
			sr, e := s.processPDF(path, filepath.Base(path), req.Source, batchID)
			if e != nil {
				errs = append(errs, e.Error())
			} else {
				results = append(results, *sr)
			}
			return nil
		})
	} else {
		sr, e := s.processPDF(req.Path, filepath.Base(req.Path), req.Source, batchID)
		if e != nil {
			errs = append(errs, e.Error())
		} else {
			results = append(results, *sr)
		}
	}
	c.JSON(http.StatusOK, models.UploadResult{Imported: len(results), Results: results, Errors: errs})
}

func (s *Server) processZip(zipPath, source string, batchID int64) ([]models.ScreenResult, []string) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, []string{err.Error()}
	}
	defer r.Close()
	var results []models.ScreenResult
	var errs []string
	for _, f := range r.File {
		if f.FileInfo().IsDir() || strings.ToLower(filepath.Ext(f.Name)) != ".pdf" {
			continue
		}
		if strings.Contains(f.Name, "__MACOSX") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		tmp, err := os.CreateTemp("", "zip-pdf-*.pdf")
		if err != nil {
			rc.Close()
			errs = append(errs, err.Error())
			continue
		}
		_, _ = io.Copy(tmp, rc)
		rc.Close()
		tmp.Close()
		sr, err := s.processPDF(tmp.Name(), filepath.Base(f.Name), source, batchID)
		os.Remove(tmp.Name())
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f.Name, err))
		} else {
			results = append(results, *sr)
		}
	}
	return results, errs
}

func (s *Server) processPDF(path, filename, source string, batchID int64) (*models.ScreenResult, error) {
	text, err := scanner.ExtractText(path)
	if err != nil {
		return nil, err
	}
	name := scanner.ExtractName(filename, text)
	score := scanner.Score(text)
	tier := score.Tier
	if t, ok := seed.ManualTier[name]; ok {
		tier = t
	}
	eng, proj, one := score.Reason, "", fmt.Sprintf("自动分 %d", score.Total)
	if sum, ok := seed.Summaries[name]; ok {
		eng, proj, one = sum.Eng, sum.Proj, sum.One
	}
	resumePath, err := s.st.SaveResume(name, path)
	if err != nil {
		return nil, err
	}
	id, err := s.st.UpsertCandidate(store.UpsertInput{
		Name: name, Source: source, Tier: tier,
		EngSummary: eng, ProjectSummary: proj, OneLiner: one,
		Action: scanner.ActionForTier(tier), ResumePath: resumePath,
		ScoreTotal: score.Total, EngScore: score.EngScore, AgentScore: score.AgentScore,
		Reason: score.Reason, Flags: score.Flags,
		InterviewOrder: seed.InterviewOrder[name],
		BatchID:        batchID,
		Status:         models.StatusScreening,
	})
	if err != nil {
		return nil, err
	}
	if qs, ok := seed.Questions[name]; ok {
		_ = s.st.SetQuestions(id, qs)
	}
	return &models.ScreenResult{
		Name: name, Tier: tier, ScoreTotal: score.Total,
		EngScore: score.EngScore, AgentScore: score.AgentScore,
		Reason: score.Reason, CandidateID: id,
	}, nil
}

func (s *Server) importSeed(c *gin.Context) {
	count, err := s.RunSeedImport()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": count})
}

func (s *Server) RunSeedImport() (int, error) {
	seedPath := filepath.Join(s.root, "seed", "batch.json")
	data, err := os.ReadFile(seedPath)
	if err != nil {
		return 0, err
	}
	var batch struct {
		Rows [][]string `json:"rows"`
	}
	if err := json.Unmarshal(data, &batch); err != nil {
		return 0, err
	}
	scored, _ := seed.LoadScored(s.root)
	imported := 0
	defaultBatch, _ := s.st.EnsureDefaultBatch()
	for _, row := range batch.Rows {
		if len(row) < 7 {
			continue
		}
		name := row[0]
		in := store.UpsertInput{
			Name: name, Source: row[1], Tier: row[2],
			EngSummary: row[3], ProjectSummary: row[4], OneLiner: row[5],
			Action: row[6], InterviewOrder: seed.InterviewOrder[name],
			BatchID: defaultBatch, Status: models.StatusScreening,
		}
		if sc, ok := scored[name]; ok {
			in.ScoreTotal = sc.Total
			in.EngScore = sc.EngScore
			in.AgentScore = sc.AgentScore
			in.Reason = sc.Reason
			in.Flags = sc.Flags
		}
		id, err := s.st.UpsertCandidate(in)
		if err != nil {
			continue
		}
		if qs, ok := seed.Questions[name]; ok {
			_ = s.st.SetQuestions(id, qs)
		}
		imported++
	}
	resumeDir := filepath.Join(s.root, "seed", "resumes")
	_ = filepath.Walk(resumeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".pdf" {
			return nil
		}
		text, _ := scanner.ExtractText(path)
		name := scanner.ExtractName(info.Name(), text)
		cand, err := s.st.GetCandidateByName(name)
		if err != nil || cand == nil {
			return nil
		}
		dest, err := s.st.SaveResume(name, path)
		if err != nil {
			return nil
		}
		in := store.UpsertInput{
			Name: name, Source: cand.Source, Tier: cand.Tier,
			EngSummary: cand.EngSummary, ProjectSummary: cand.ProjectSummary,
			OneLiner: cand.OneLiner, Action: cand.Action, ResumePath: dest,
			InterviewOrder: seed.InterviewOrder[name],
			BatchID: cand.BatchID, Status: cand.Status,
		}
		if sc, ok := scored[name]; ok {
			in.ScoreTotal, in.EngScore, in.AgentScore = sc.Total, sc.EngScore, sc.AgentScore
			in.Reason, in.Flags = sc.Reason, sc.Flags
		} else {
			score := scanner.Score(text)
			in.ScoreTotal, in.EngScore, in.AgentScore = score.Total, score.EngScore, score.AgentScore
			in.Reason, in.Flags = score.Reason, score.Flags
		}
		_, _ = s.st.UpsertCandidate(in)
		return nil
	})
	return imported, nil
}

func (s *Server) rescreenAll(c *gin.Context) {
	list, err := s.st.ListCandidates("", "", "", 0, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	scored, _ := seed.LoadScored(s.root)
	updated := 0
	for _, cand := range list {
		if cand.ResumePath == "" {
			continue
		}
		text, err := scanner.ExtractText(cand.ResumePath)
		if err != nil {
			continue
		}
		score := scanner.Score(text)
		tier := score.Tier
		if cand.TierManual {
			tier = cand.Tier
		} else if t, ok := seed.ManualTier[cand.Name]; ok {
			tier = t
		}
		if score.Total < 0 {
			if sc, ok := scored[cand.Name]; ok {
				score.Total, score.EngScore, score.AgentScore = sc.Total, sc.EngScore, sc.AgentScore
				score.Reason, score.Flags = sc.Reason, sc.Flags
				if !cand.TierManual {
					tier = cand.Tier
					if t, ok := seed.ManualTier[cand.Name]; ok {
						tier = t
					}
				}
			} else if !cand.TierManual {
				if t, ok := seed.ManualTier[cand.Name]; ok {
					tier = t
					score.Reason = "manual_tier"
				}
			}
		}
		eng, proj, one := cand.EngSummary, cand.ProjectSummary, cand.OneLiner
		if sum, ok := seed.Summaries[cand.Name]; ok {
			eng, proj, one = sum.Eng, sum.Proj, sum.One
		}
		_, err = s.st.UpsertCandidate(store.UpsertInput{
			Name: cand.Name, Source: cand.Source, Tier: tier,
			EngSummary: eng, ProjectSummary: proj, OneLiner: one,
			Action: func() string {
				if cand.TierManual {
					return cand.Action
				}
				return scanner.ActionForTier(tier)
			}(),
			ResumePath: cand.ResumePath,
			ScoreTotal: score.Total, EngScore: score.EngScore, AgentScore: score.AgentScore,
			Reason: score.Reason, Flags: score.Flags,
			InterviewOrder: cand.InterviewOrder,
			BatchID:        cand.BatchID,
			Status:         cand.Status,
		})
		if err == nil {
			updated++
		}
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (s *Server) syncQuestions(c *gin.Context) {
	list, err := s.st.ListCandidates("", "", "", 0, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated := 0
	for _, cand := range list {
		qs, ok := seed.Questions[cand.Name]
		if !ok {
			continue
		}
		if err := s.st.SetQuestions(cand.ID, qs); err == nil {
			updated++
		}
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (s *Server) exportMarkdown(c *gin.Context) {
	data, err := s.st.ExportAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	md := export.MarkdownReport(data)
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, md)
}

func (s *Server) exportFeishu(c *gin.Context) {
	data, err := s.st.ExportAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	md := export.MarkdownReport(data)
	title := fmt.Sprintf("Agent实习招聘筛选 %s", time.Now().Format("2006-01-02"))
	mdPath := filepath.Join(s.root, "data", "export.md")
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rel := "data/export.md"
	cmd := exec.Command("lark-cli", "docs", "+create",
		"--title", title,
		"--doc-format", "markdown",
		"--content", "@"+rel,
		"--as", "user",
		"--format", "json",
	)
	cmd.Dir = s.root
	out, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"ok":       false,
			"markdown": mdPath,
			"error":    string(out),
			"hint":     "lark-cli 未授权或不可用，已保存本地 Markdown",
		})
		return
	}
	var resp map[string]any
	_ = json.Unmarshal(out, &resp)
	c.JSON(http.StatusOK, gin.H{"ok": true, "lark": resp, "markdown": mdPath})
}

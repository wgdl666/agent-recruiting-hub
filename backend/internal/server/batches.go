package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) listBatches(c *gin.Context) {
	activeOnly := c.DefaultQuery("active", "true") == "true"
	list, err := s.st.ListBatches(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *Server) createBatch(c *gin.Context) {
	var req struct {
		Name       string `json:"name"`
		Tag        string `json:"tag"`
		PeriodType string `json:"period_type"`
		Auto       bool   `json:"auto"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int64
	var err error
	if req.Auto {
		pt := req.PeriodType
		if pt == "" {
			pt = "daily"
		}
		id, err = s.st.GetOrCreateBatchForDate(time.Now(), pt)
	} else {
		id, err = s.st.CreateBatch(req.Name, req.Tag, req.PeriodType)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Server) pipelineStats(c *gin.Context) {
	batchID, _ := strconv.ParseInt(c.Query("batch_id"), 10, 64)
	st, err := s.st.BatchStats(batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := models.PipelineStats{
		BatchID:  batchID,
		Total:    st.Total,
		ByStatus: st.ByStatus,
		ByTier:   st.ByTier,
	}
	if batchID > 0 {
		out.BatchName, _ = s.st.BatchName(batchID)
	} else {
		out.BatchName = "全部批次"
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) resolveUploadBatch(c *gin.Context) (int64, error) {
	if v := c.PostForm("batch_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	period := c.DefaultPostForm("period_type", "daily")
	return s.st.GetOrCreateBatchForDate(time.Now(), period)
}

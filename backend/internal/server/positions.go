package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/caden/agent-recruiting-hub/internal/store"
	"github.com/gin-gonic/gin"
)

func (s *Server) listPositions(c *gin.Context) {
	openOnly := c.Query("open") == "true"
	list, err := s.st.ListPositions(openOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *Server) getPosition(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := s.st.GetPosition(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "position not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (s *Server) createPosition(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		SkillID     string `json:"skill_id"`
		Description string `json:"description"`
		IsOpen      *bool  `json:"is_open"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := s.st.CreatePosition(store.PositionInput{
		Name: req.Name, SkillID: req.SkillID, Description: req.Description, IsOpen: req.IsOpen,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := s.st.GetPosition(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (s *Server) patchPosition(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name        *string `json:"name"`
		SkillID     *string `json:"skill_id"`
		Description *string `json:"description"`
		JD          *string `json:"jd"`
		IsOpen      *bool   `json:"is_open"`
		SortOrder   *int    `json:"sort_order"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.st.PatchPosition(id, store.PositionPatch{
		Name: req.Name, SkillID: req.SkillID, Description: req.Description, JD: req.JD, IsOpen: req.IsOpen, SortOrder: req.SortOrder,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := s.st.GetPosition(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// resolveOpening 上传优先用岗位；岗位上绑了检验 Skill。旧客户端只传 skill_id 时按该 Skill 找在招岗。
func (s *Server) resolveOpening(positionID int64, skillID string) (int64, string, error) {
	if positionID > 0 {
		p, err := s.st.GetPosition(positionID)
		if err != nil {
			return 0, "", fmt.Errorf("unknown position_id")
		}
		if !p.IsOpen {
			return 0, "", fmt.Errorf("position is closed")
		}
		if _, err := parseSkillID(p.SkillID); err != nil {
			return 0, "", fmt.Errorf("position has invalid skill_id")
		}
		return p.ID, p.SkillID, nil
	}
	sid, err := parseSkillID(skillID)
	if err != nil {
		return 0, "", fmt.Errorf("position_id or skill_id required")
	}
	p, err := s.st.FindPositionBySkill(sid)
	if err != nil {
		return 0, "", fmt.Errorf("no position for skill_id")
	}
	return p.ID, sid, nil
}

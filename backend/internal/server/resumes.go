package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (s *Server) syncResumes(c *gin.Context) {
	if s.oss == nil || !s.oss.Enabled() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OSS not configured (set HUB_OSS_* env)"})
		return
	}
	res, err := s.st.SyncResumesToOSS(s.oss)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) persistResume(name, localRelPath string) (string, string) {
	resumeKey := ""
	if s.oss != nil && s.oss.Enabled() {
		full := s.st.ResolveResumePath(localRelPath)
		if full != "" {
			key := s.oss.ObjectKey(filepath.Base(full))
			if err := s.oss.PutFile(full, key); err == nil {
				resumeKey = key
			}
		}
	}
	return localRelPath, resumeKey
}

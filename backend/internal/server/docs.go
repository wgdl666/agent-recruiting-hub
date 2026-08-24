package server

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type docEntry struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

var docCatalog = []docEntry{
	{Slug: "overview", Title: "项目概览", Order: 1},
	{Slug: "quickstart", Title: "快速开始", Order: 2},
	{Slug: "architecture", Title: "架构说明", Order: 3},
	{Slug: "api", Title: "REST API", Order: 4},
	{Slug: "deploy", Title: "部署指南", Order: 5},
	{Slug: "extend", Title: "扩展开发", Order: 6},
}

var slugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

func (s *Server) listDocs(c *gin.Context) {
	out := make([]docEntry, 0, len(docCatalog))
	docsDir := filepath.Join(s.root, "docs")
	for _, e := range docCatalog {
		if _, err := os.Stat(filepath.Join(docsDir, e.Slug+".md")); err == nil {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	c.JSON(http.StatusOK, out)
}

func (s *Server) getDoc(c *gin.Context) {
	slug := c.Param("slug")
	if !slugRe.MatchString(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug"})
		return
	}
	path := filepath.Join(s.root, "docs", slug+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doc not found"})
		return
	}
	title := slug
	for _, e := range docCatalog {
		if e.Slug == slug {
			title = e.Title
			break
		}
	}
	// First # heading in file overrides display title if present
	if lines := strings.SplitN(string(data), "\n", 2); len(lines) > 0 {
		if h := strings.TrimPrefix(strings.TrimSpace(lines[0]), "# "); lines[0] != h {
			title = h
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"slug":     slug,
		"title":    title,
		"markdown": string(data),
	})
}

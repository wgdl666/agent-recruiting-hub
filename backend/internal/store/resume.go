package store

import (
	"path/filepath"
	"strings"
)

// RelativeResumePath is stored in DB (portable across machines).
func RelativeResumePath(filename string) string {
	base := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	return filepath.Join("data", "resumes", base)
}

func (s *Store) ResolveResumePath(stored string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	if filepath.IsAbs(stored) {
		if fileExists(stored) {
			return stored
		}
		// Broken absolute path from another machine — try local data dir by basename.
		stored = RelativeResumePath(stored)
	}
	full := filepath.Join(s.root, filepath.FromSlash(stored))
	if fileExists(full) {
		return full
	}
	return ""
}

func (s *Store) HasResumeFile(resumePath, resumeKey string) bool {
	if strings.TrimSpace(resumeKey) != "" {
		return true
	}
	return s.ResolveResumePath(resumePath) != ""
}

package store

import (
	"path/filepath"

	"github.com/caden/agent-recruiting-hub/internal/storage"
)

type ResumeSyncResult struct {
	Uploaded int      `json:"uploaded"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

func (s *Store) SyncResumesToOSS(oss *storage.ResumeStore) (*ResumeSyncResult, error) {
	out := &ResumeSyncResult{}
	if oss == nil || !oss.Enabled() {
		return out, nil
	}
	rows, err := s.db.Query(`SELECT id, name, resume_path, resume_key FROM candidates`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name, path, key string
		if err := rows.Scan(&id, &name, &path, &key); err != nil {
			return nil, err
		}
		if key != "" {
			out.Skipped++
			continue
		}
		local := s.ResolveResumePath(path)
		if local == "" {
			continue
		}
		objKey := oss.ObjectKey(filepath.Base(local))
		if err := oss.PutFile(local, objKey); err != nil {
			out.Errors = append(out.Errors, name+": "+err.Error())
			continue
		}
		rel := RelativeResumePath(local)
		if _, err := s.db.Exec(`UPDATE candidates SET resume_path = ?, resume_key = ? WHERE id = ?`, rel, objKey, id); err != nil {
			out.Errors = append(out.Errors, name+": db "+err.Error())
			continue
		}
		out.Uploaded++
	}
	return out, rows.Err()
}

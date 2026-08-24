package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/models"
)

func (s *Store) migrateBatches() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS batches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  tag TEXT NOT NULL UNIQUE,
  period_type TEXT DEFAULT 'custom',
  is_active INTEGER DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_batches_active ON batches(is_active);
`)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec(`ALTER TABLE candidates ADD COLUMN batch_id INTEGER DEFAULT 0`)
	return nil
}

func (s *Store) EnsureDefaultBatch() (int64, error) {
	tag := "seed-2026-08"
	var id int64
	err := s.db.QueryRow(`SELECT id FROM batches WHERE tag = ?`, tag).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`INSERT INTO batches (name, tag, period_type, is_active, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`, "2026-08 初筛批次", tag, "custom", 1, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetOrCreateBatchForDate(t time.Time, periodType string) (int64, error) {
	var tag, name string
	switch periodType {
	case "weekly":
		y, w := t.ISOWeek()
		tag = fmt.Sprintf("%d-W%02d", y, w)
		name = fmt.Sprintf("%d年第%d周", y, w)
	case "monthly":
		tag = t.Format("2006-01")
		name = t.Format("2006年1月")
	default:
		tag = t.Format("2006-01-02")
		name = t.Format("1月2日") + " 上传"
		periodType = "daily"
	}
	var id int64
	err := s.db.QueryRow(`SELECT id FROM batches WHERE tag = ?`, tag).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`INSERT INTO batches (name, tag, period_type, is_active, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`, name, tag, periodType, 1, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListBatches(activeOnly bool) ([]models.Batch, error) {
	q := `SELECT id, name, tag, period_type, is_active, created_at, updated_at FROM batches`
	if activeOnly {
		q += ` WHERE is_active = 1`
	}
	q += ` ORDER BY created_at DESC`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Batch, 0)
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		st, _ := s.BatchStats(b.ID)
		b.Stats = st
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) CreateBatch(name, tag, periodType string) (int64, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return 0, fmt.Errorf("tag required")
	}
	if name == "" {
		name = tag
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`INSERT INTO batches (name, tag, period_type, is_active, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`, name, tag, periodType, 1, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) BatchStats(batchID int64) (*models.BatchStats, error) {
	where := ""
	args := []any{}
	if batchID > 0 {
		where = ` WHERE batch_id = ?`
		args = append(args, batchID)
	}
	st := &models.BatchStats{ByStatus: map[string]int{}, ByTier: map[string]int{}}
	rows, err := s.db.Query(`SELECT status, COUNT(*) FROM candidates`+where+` GROUP BY status`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		if status == "" {
			status = models.StatusScreening
		}
		st.ByStatus[status] = n
		st.Total += n
	}
	rows2, err := s.db.Query(`SELECT tier, COUNT(*) FROM candidates`+where+` GROUP BY tier`, args...)
	if err != nil {
		return st, nil
	}
	defer rows2.Close()
	for rows2.Next() {
		var tier string
		var n int
		if err := rows2.Scan(&tier, &n); err != nil {
			return nil, err
		}
		st.ByTier[tier] = n
	}
	return st, rows2.Err()
}

func scanBatch(rows *sql.Rows) (models.Batch, error) {
	var b models.Batch
	var active int
	var created, updated string
	err := rows.Scan(&b.ID, &b.Name, &b.Tag, &b.PeriodType, &active, &created, &updated)
	if err != nil {
		return b, err
	}
	b.IsActive = active == 1
	b.CreatedAt, _ = time.Parse(time.RFC3339, created)
	b.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return b, nil
}

func (s *Store) AssignBatchToOrphans(batchID int64) error {
	_, err := s.db.Exec(`UPDATE candidates SET batch_id = ? WHERE batch_id IS NULL OR batch_id = 0`, batchID)
	return err
}

func (s *Store) BatchName(id int64) (string, error) {
	var name string
	err := s.db.QueryRow(`SELECT name FROM batches WHERE id = ?`, id).Scan(&name)
	return name, err
}

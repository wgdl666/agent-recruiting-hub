package store

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/caden/agent-recruiting-hub/internal/seed"
	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	root string
}

func Open(root string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(root, "data", "resumes"), 0o755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(root, "data", "hub.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, root: root}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	if err := s.migrateBatches(); err != nil {
		return nil, err
	}
	defaultBatch, _ := s.EnsureDefaultBatch()
	if defaultBatch > 0 {
		_ = s.AssignBatchToOrphans(defaultBatch)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Root() string { return s.root }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS candidates (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  source TEXT DEFAULT '',
  tier TEXT DEFAULT 'B',
  eng_summary TEXT DEFAULT '',
  project_summary TEXT DEFAULT '',
  one_liner TEXT DEFAULT '',
  action TEXT DEFAULT '',
  resume_path TEXT DEFAULT '',
  score_total INTEGER DEFAULT 0,
  eng_score INTEGER DEFAULT 0,
  agent_score INTEGER DEFAULT 0,
  reason TEXT DEFAULT '',
  flags_json TEXT DEFAULT '[]',
  interview_order INTEGER DEFAULT 0,
  tier_manual INTEGER DEFAULT 0,
  status TEXT DEFAULT 'screening',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_candidates_name ON candidates(name);
CREATE TABLE IF NOT EXISTS interview_questions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  candidate_id INTEGER NOT NULL,
  sort_order INTEGER NOT NULL,
  question TEXT NOT NULL,
  FOREIGN KEY(candidate_id) REFERENCES candidates(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_questions_candidate ON interview_questions(candidate_id);
`)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec(`ALTER TABLE candidates ADD COLUMN interview_order INTEGER DEFAULT 0`)
	_, _ = s.db.Exec(`ALTER TABLE candidates ADD COLUMN tier_manual INTEGER DEFAULT 0`)
	_, _ = s.db.Exec(`ALTER TABLE candidates ADD COLUMN status TEXT DEFAULT 'screening'`)
	_, _ = s.db.Exec(`UPDATE candidates SET status = 'screening' WHERE status IS NULL OR status = ''`)
	_, _ = s.db.Exec(`ALTER TABLE interview_questions ADD COLUMN answer TEXT DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE interview_questions ADD COLUMN level TEXT DEFAULT ''`)
	return nil
}

func (s *Store) ListCandidates(tier, status, q string, batchID int64, engFirst bool) ([]models.Candidate, error) {
	query := `SELECT c.id, c.name, c.source, c.tier, c.eng_summary, c.project_summary, c.one_liner, c.action,
		c.resume_path, c.score_total, c.eng_score, c.agent_score, c.reason, c.flags_json, c.interview_order, c.tier_manual, c.status, c.batch_id,
		COALESCE(b.name,''), c.created_at, c.updated_at
		FROM candidates c LEFT JOIN batches b ON c.batch_id = b.id WHERE 1=1`
	args := []any{}
	if tier != "" && tier != "all" {
		query += ` AND c.tier = ?`
		args = append(args, tier)
	}
	if status != "" && status != "all" {
		query += ` AND c.status = ?`
		args = append(args, status)
	}
	if batchID > 0 {
		query += ` AND c.batch_id = ?`
		args = append(args, batchID)
	}
	if q != "" {
		query += ` AND (c.name LIKE ? OR c.eng_summary LIKE ? OR c.project_summary LIKE ? OR c.one_liner LIKE ?)`
		like := "%" + q + "%"
		args = append(args, like, like, like, like)
	}
	if engFirst {
		query += ` ORDER BY CASE c.status WHEN 'interviewing' THEN 0 WHEN 'to_interview' THEN 1 WHEN 'screening' THEN 2 WHEN 'passed' THEN 3 WHEN 'completed' THEN 4 ELSE 5 END,
			CASE c.tier WHEN 'S' THEN 0 WHEN 'A' THEN 1 WHEN 'B' THEN 2 WHEN 'C' THEN 3 ELSE 4 END,
			CASE WHEN c.interview_order > 0 THEN c.interview_order ELSE 999 END,
			c.eng_score DESC, c.agent_score DESC, c.score_total DESC`
	} else {
		query += ` ORDER BY CASE tier WHEN 'S' THEN 0 WHEN 'A' THEN 1 WHEN 'B' THEN 2 WHEN 'C' THEN 3 ELSE 4 END,
			score_total DESC`
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Candidate
	for rows.Next() {
		c, err := scanCandidate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCandidate(id int64) (*models.CandidateDetail, error) {
	row := s.db.QueryRow(`SELECT c.id, c.name, c.source, c.tier, c.eng_summary, c.project_summary, c.one_liner, c.action,
		c.resume_path, c.score_total, c.eng_score, c.agent_score, c.reason, c.flags_json, c.interview_order, c.tier_manual, c.status, c.batch_id,
		COALESCE(b.name,''), c.created_at, c.updated_at
		FROM candidates c LEFT JOIN batches b ON c.batch_id = b.id WHERE c.id = ?`, id)
	c, err := scanCandidateRow(row)
	if err != nil {
		return nil, err
	}
	qs, err := s.ListQuestions(id)
	if err != nil {
		return nil, err
	}
	return &models.CandidateDetail{Candidate: c, Questions: qs}, nil
}

func (s *Store) GetCandidateByName(name string) (*models.Candidate, error) {
	row := s.db.QueryRow(`SELECT c.id, c.name, c.source, c.tier, c.eng_summary, c.project_summary, c.one_liner, c.action,
		c.resume_path, c.score_total, c.eng_score, c.agent_score, c.reason, c.flags_json, c.interview_order, c.tier_manual, c.status, c.batch_id,
		COALESCE(b.name,''), c.created_at, c.updated_at
		FROM candidates c LEFT JOIN batches b ON c.batch_id = b.id WHERE c.name = ?`, name)
	c, err := scanCandidateRow(row)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListQuestions(candidateID int64) ([]models.InterviewQuestion, error) {
	rows, err := s.db.Query(`SELECT id, candidate_id, sort_order, question, answer, level FROM interview_questions
		WHERE candidate_id = ? ORDER BY sort_order`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.InterviewQuestion
	for rows.Next() {
		var q models.InterviewQuestion
		if err := rows.Scan(&q.ID, &q.CandidateID, &q.SortOrder, &q.Question, &q.Answer, &q.Level); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Store) Stats() (*models.Stats, error) {
	rows, err := s.db.Query(`SELECT tier, COUNT(*) FROM candidates GROUP BY tier`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	st := &models.Stats{ByTier: map[string]int{}, ByStatus: map[string]int{}}
	for rows.Next() {
		var tier string
		var n int
		if err := rows.Scan(&tier, &n); err != nil {
			return nil, err
		}
		st.ByTier[tier] = n
		st.Total += n
	}
	rows.Close()
	srows, err := s.db.Query(`SELECT status, COUNT(*) FROM candidates GROUP BY status`)
	if err != nil {
		return st, nil
	}
	defer srows.Close()
	for srows.Next() {
		var status string
		var n int
		if err := srows.Scan(&status, &n); err != nil {
			return nil, err
		}
		if status == "" {
			status = models.StatusScreening
		}
		st.ByStatus[status] = n
	}
	return st, srows.Err()
}

type UpsertInput struct {
	Name           string
	Source         string
	Tier           string
	EngSummary     string
	ProjectSummary string
	OneLiner       string
	Action         string
	ResumePath     string
	ScoreTotal     int
	EngScore       int
	AgentScore     int
	Reason         string
	Flags          []string
	InterviewOrder int
	BatchID        int64
	Status         string
}

func defaultStatus(s string) string {
	if s == "" {
		return models.StatusScreening
	}
	return s
}

func (s *Store) UpsertCandidate(in UpsertInput) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	flags, _ := json.Marshal(in.Flags)
	var id int64
	err := s.db.QueryRow(`SELECT id FROM candidates WHERE name = ?`, in.Name).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := s.db.Exec(`INSERT INTO candidates
			(name, source, tier, eng_summary, project_summary, one_liner, action, resume_path,
			 score_total, eng_score, agent_score, reason, flags_json, interview_order, status, batch_id, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			in.Name, in.Source, in.Tier, in.EngSummary, in.ProjectSummary, in.OneLiner, in.Action,
			in.ResumePath, in.ScoreTotal, in.EngScore, in.AgentScore, in.Reason, string(flags), in.InterviewOrder,
			defaultStatus(in.Status), in.BatchID, now, now)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	if err != nil {
		return 0, err
	}
	resumePath := in.ResumePath
	status := defaultStatus(in.Status)
	batchID := in.BatchID
	_ = s.db.QueryRow(`SELECT resume_path, status, batch_id FROM candidates WHERE id = ?`, id).
		Scan(&resumePath, &status, &batchID)
	if in.ResumePath != "" {
		resumePath = in.ResumePath
	}
	if in.Status != "" {
		status = defaultStatus(in.Status)
	}
	if in.BatchID > 0 {
		batchID = in.BatchID
	}
	_, err = s.db.Exec(`UPDATE candidates SET source=?, tier=?, eng_summary=?, project_summary=?, one_liner=?,
		action=?, resume_path=?, score_total=?, eng_score=?, agent_score=?, reason=?, flags_json=?, interview_order=?, status=?, batch_id=?, updated_at=?
		WHERE id=?`,
		in.Source, in.Tier, in.EngSummary, in.ProjectSummary, in.OneLiner, in.Action,
		resumePath, in.ScoreTotal, in.EngScore, in.AgentScore, in.Reason, string(flags), in.InterviewOrder, status, batchID, now, id)
	return id, err
}

func (s *Store) SetQuestions(candidateID int64, items []seed.QA) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM interview_questions WHERE candidate_id = ?`, candidateID); err != nil {
		return err
	}
	for i, item := range items {
		if strings.TrimSpace(item.Question) == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO interview_questions (candidate_id, sort_order, question, answer, level) VALUES (?,?,?,?,?)`,
			candidateID, i+1, item.Question, item.Answer, item.Level); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ResumeDir() string {
	return filepath.Join(s.root, "data", "resumes")
}

func (s *Store) SaveResume(name, srcPath string) (string, error) {
	ext := filepath.Ext(srcPath)
	if ext == "" {
		ext = ".pdf"
	}
	safe := sanitizeFilename(name) + ext
	dest := filepath.Join(s.ResumeDir(), safe)
	if err := copyFile(srcPath, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func scanCandidate(rows *sql.Rows) (models.Candidate, error) {
	var c models.Candidate
	var flagsJSON, created, updated string
	var interviewOrder, tierManual int
	var status string
	var batchID int64
	var batchName string
	err := rows.Scan(&c.ID, &c.Name, &c.Source, &c.Tier, &c.EngSummary, &c.ProjectSummary, &c.OneLiner,
		&c.Action, &c.ResumePath, &c.ScoreTotal, &c.EngScore, &c.AgentScore, &c.Reason, &flagsJSON,
		&interviewOrder, &tierManual, &status, &batchID, &batchName, &created, &updated)
	if err != nil {
		return c, err
	}
	c.InterviewOrder = interviewOrder
	c.TierManual = tierManual == 1
	c.Status = defaultStatus(status)
	c.BatchID = batchID
	c.BatchName = batchName
	_ = json.Unmarshal([]byte(flagsJSON), &c.Flags)
	c.HasResume = c.ResumePath != "" && fileExists(c.ResumePath)
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return c, nil
}

func scanCandidateRow(row *sql.Row) (models.Candidate, error) {
	var c models.Candidate
	var flagsJSON, created, updated string
	var interviewOrder, tierManual int
	var status string
	var batchID int64
	var batchName string
	err := row.Scan(&c.ID, &c.Name, &c.Source, &c.Tier, &c.EngSummary, &c.ProjectSummary, &c.OneLiner,
		&c.Action, &c.ResumePath, &c.ScoreTotal, &c.EngScore, &c.AgentScore, &c.Reason, &flagsJSON,
		&interviewOrder, &tierManual, &status, &batchID, &batchName, &created, &updated)
	if err != nil {
		return c, err
	}
	c.InterviewOrder = interviewOrder
	c.TierManual = tierManual == 1
	c.Status = defaultStatus(status)
	c.BatchID = batchID
	c.BatchName = batchName
	_ = json.Unmarshal([]byte(flagsJSON), &c.Flags)
	c.HasResume = c.ResumePath != "" && fileExists(c.ResumePath)
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return c, nil
}

func sanitizeFilename(name string) string {
	re := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return re.Replace(name)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

func (s *Store) ExportAll() ([]models.CandidateDetail, error) {
	list, err := s.ListCandidates("", "", "", 0, true)
	if err != nil {
		return nil, err
	}
	out := make([]models.CandidateDetail, 0, len(list))
	for _, c := range list {
		d, err := s.GetCandidate(c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, nil
}

func (s *Store) DeleteCandidate(id int64) error {
	var path string
	_ = s.db.QueryRow(`SELECT resume_path FROM candidates WHERE id = ?`, id).Scan(&path)
	_, err := s.db.Exec(`DELETE FROM candidates WHERE id = ?`, id)
	if err == nil && path != "" {
		_ = os.Remove(path)
	}
	return err
}

func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM candidates`).Scan(&n)
	return n, err
}

func (s *Store) DBPath() string {
	return filepath.Join(s.root, "data", "hub.db")
}

package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/caden/agent-recruiting-hub/internal/skills"
)

//go:embed intern_jd.md
var internJD string

//go:embed image_software_test_jd.md
var imageSoftwareTestJD string

const internPositionSlug = "intern"
const imageSoftwareTestSlug = "image-software-test"
const imageSoftwareTestName = "影像 && 软件 测试"

func (s *Store) migratePositions() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS positions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  skill_id TEXT NOT NULL,
  description TEXT DEFAULT '',
  jd TEXT DEFAULT '',
  sort_order INTEGER DEFAULT 0,
  is_open INTEGER DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_positions_open ON positions(is_open, sort_order);
`)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec(`ALTER TABLE candidates ADD COLUMN position_id INTEGER DEFAULT 0`)
	_, _ = s.db.Exec(`ALTER TABLE positions ADD COLUMN jd TEXT DEFAULT ''`)
	internID, err := s.EnsureInternPosition()
	if err != nil {
		return err
	}
	// 历史候选人没有岗位字段，一律归到「实习生」
	_, _ = s.db.Exec(`UPDATE candidates SET position_id = ? WHERE position_id IS NULL OR position_id = 0`, internID)
	// 仅给空 JD 填默认文案，已手改过的不覆盖
	_, _ = s.db.Exec(`UPDATE positions SET jd = ? WHERE slug = ? AND (jd IS NULL OR jd = '')`, internJD, internPositionSlug)
	if err := s.cleanupEmptyInternDuplicates(); err != nil {
		return err
	}
	if _, err := s.EnsureImageSoftwareTestPosition(); err != nil {
		return err
	}
	return nil
}

// cleanupEmptyInternDuplicates 清掉误点「新增岗位」留下的空实习生岗（已停、无人），避免侧栏/列表出现两条同名岗。
func (s *Store) cleanupEmptyInternDuplicates() error {
	_, err := s.db.Exec(`
DELETE FROM positions
 WHERE name = ?
   AND slug != ?
   AND is_open = 0
   AND NOT EXISTS (SELECT 1 FROM candidates c WHERE c.position_id = positions.id)`,
		"实习生", internPositionSlug)
	return err
}

func (s *Store) EnsureInternPosition() (int64, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM positions WHERE slug = ?`, internPositionSlug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`INSERT INTO positions (name, slug, skill_id, description, jd, sort_order, is_open, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		"实习生", internPositionSlug, skills.Intern, "Agent 工程实习：上手就能干活", internJD, 0, 1, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) EnsureImageSoftwareTestPosition() (int64, error) {
	// 启动时补齐「影像 && 软件 测试」在招岗；已有同名岗只填空 JD，不覆盖手改。
	var id int64
	err := s.db.QueryRow(`SELECT id FROM positions WHERE slug = ? OR name = ? ORDER BY id ASC LIMIT 1`,
		imageSoftwareTestSlug, imageSoftwareTestName).Scan(&id)
	if err == nil {
		// 已有同名岗时只补空 JD / 技能，不覆盖手改过的文案
		_, _ = s.db.Exec(`UPDATE positions SET skill_id = ?, updated_at = ? WHERE id = ? AND skill_id != ?`,
			skills.ImageSoftwareTest, time.Now().UTC().Format(time.RFC3339), id, skills.ImageSoftwareTest)
		_, _ = s.db.Exec(`UPDATE positions SET jd = ? WHERE id = ? AND (jd IS NULL OR jd = '')`, imageSoftwareTestJD, id)
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	sortOrder := 1
	_ = s.db.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM positions`).Scan(&sortOrder)
	res, err := s.db.Exec(`INSERT INTO positions (name, slug, skill_id, description, jd, sort_order, is_open, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		imageSoftwareTestName, imageSoftwareTestSlug, skills.ImageSoftwareTest,
		"影像质量与软件测试：上手就能验", imageSoftwareTestJD, sortOrder, 1, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func fillPositionSkill(p *models.Position) {
	if sk, ok := skills.Lookup(p.SkillID); ok {
		p.SkillName = sk.Name
	}
}

func scanPosition(scanner interface {
	Scan(dest ...any) error
}) (models.Position, error) {
	var p models.Position
	var created, updated string
	var isOpen int
	err := scanner.Scan(&p.ID, &p.Name, &p.Slug, &p.SkillID, &p.Description, &p.JD, &p.SortOrder, &isOpen, &p.CandidateCount, &created, &updated)
	if err != nil {
		return p, err
	}
	p.IsOpen = isOpen == 1
	p.CreatedAt, _ = time.Parse(time.RFC3339, created)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	fillPositionSkill(&p)
	return p, nil
}

const positionSelect = `SELECT p.id, p.name, p.slug, p.skill_id, p.description, COALESCE(p.jd,''), p.sort_order, p.is_open,
	(SELECT COUNT(*) FROM candidates c WHERE c.position_id = p.id), p.created_at, p.updated_at
	FROM positions p`

func (s *Store) ListPositions(openOnly bool) ([]models.Position, error) {
	q := positionSelect + ` WHERE 1=1`
	args := []any{}
	if openOnly {
		q += ` AND p.is_open = 1`
	}
	q += ` ORDER BY p.sort_order ASC, p.id ASC`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Position, 0)
	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetPosition(id int64) (*models.Position, error) {
	row := s.db.QueryRow(positionSelect+` WHERE p.id = ?`, id)
	p, err := scanPosition(row)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) FindPositionBySkill(skillID string) (*models.Position, error) {
	skillID = strings.TrimSpace(skillID)
	row := s.db.QueryRow(positionSelect+` WHERE p.skill_id = ? AND p.is_open = 1 ORDER BY p.sort_order ASC, p.id ASC LIMIT 1`, skillID)
	p, err := scanPosition(row)
	if err == nil {
		return &p, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	if skillID == skills.Intern {
		id, e := s.EnsureInternPosition()
		if e != nil {
			return nil, e
		}
		return s.GetPosition(id)
	}
	row = s.db.QueryRow(positionSelect+` WHERE p.skill_id = ? ORDER BY p.id ASC LIMIT 1`, skillID)
	p, err = scanPosition(row)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type PositionInput struct {
	Name        string
	SkillID     string
	Description string
	JD          string
	SortOrder   *int
	IsOpen      *bool
}

type PositionPatch struct {
	Name        *string
	SkillID     *string
	Description *string
	JD          *string
	SortOrder   *int
	IsOpen      *bool
}

func (s *Store) CreatePosition(in PositionInput) (int64, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return 0, fmt.Errorf("name required")
	}
	if !skills.Valid(in.SkillID) {
		return 0, fmt.Errorf("unknown skill_id")
	}
	sortOrder := 0
	if in.SortOrder != nil {
		sortOrder = *in.SortOrder
	} else {
		_ = s.db.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM positions`).Scan(&sortOrder)
	}
	isOpen := 1
	if in.IsOpen != nil && !*in.IsOpen {
		isOpen = 0
	}
	now := time.Now().UTC().Format(time.RFC3339)
	slug := fmt.Sprintf("pos-%d", time.Now().UnixNano())
	res, err := s.db.Exec(`INSERT INTO positions (name, slug, skill_id, description, jd, sort_order, is_open, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, name, slug, in.SkillID, strings.TrimSpace(in.Description), strings.TrimSpace(in.JD), sortOrder, isOpen, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) PatchPosition(id int64, in PositionPatch) error {
	cur, err := s.GetPosition(id)
	if err != nil {
		return err
	}
	if in.Name == nil && in.SkillID == nil && in.Description == nil && in.JD == nil && in.SortOrder == nil && in.IsOpen == nil {
		return fmt.Errorf("no fields to update")
	}
	name := cur.Name
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return fmt.Errorf("name required")
		}
	}
	skillID := cur.SkillID
	if in.SkillID != nil {
		skillID = strings.TrimSpace(*in.SkillID)
	}
	if !skills.Valid(skillID) {
		return fmt.Errorf("unknown skill_id")
	}
	desc := cur.Description
	if in.Description != nil {
		desc = strings.TrimSpace(*in.Description)
	}
	jd := cur.JD
	if in.JD != nil {
		jd = strings.TrimSpace(*in.JD)
	}
	sortOrder := cur.SortOrder
	if in.SortOrder != nil {
		sortOrder = *in.SortOrder
	}
	isOpen := 0
	if cur.IsOpen {
		isOpen = 1
	}
	if in.IsOpen != nil {
		if *in.IsOpen {
			isOpen = 1
		} else {
			isOpen = 0
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`UPDATE positions SET name=?, skill_id=?, description=?, jd=?, sort_order=?, is_open=?, updated_at=? WHERE id=?`,
		name, skillID, desc, jd, sortOrder, isOpen, now, id)
	return err
}

// DeletePosition 先解绑候选人再删岗，避免列表还挂着已不存在的岗位名；候选人本身不删。
func (s *Store) DeletePosition(id int64) error {
	if _, err := s.GetPosition(id); err != nil {
		return err
	}
	if _, err := s.db.Exec(`UPDATE candidates SET position_id = 0 WHERE position_id = ?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM positions WHERE id = ?`, id)
	return err
}

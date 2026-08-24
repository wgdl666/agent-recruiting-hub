package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/caden/agent-recruiting-hub/internal/models"
	"github.com/caden/agent-recruiting-hub/internal/scanner"
)

var validTiers = map[string]bool{"S": true, "A": true, "淘汰": true}

var validStatuses = map[string]bool{
	models.StatusScreening:    true,
	models.StatusRead:         true,
	models.StatusToInterview:  true,
	models.StatusInterviewing: true,
	models.StatusPassed:       true,
	models.StatusCompleted:    true,
	models.StatusRejected:     true,
}

type PatchInput struct {
	Tier           *string
	Action         *string
	InterviewOrder *int
	Status         *string
	ClearManual    bool
}

func (s *Store) PatchCandidate(id int64, patch PatchInput) error {
	cand, err := s.GetCandidate(id)
	if err != nil {
		return err
	}
	tier := cand.Tier
	action := cand.Action
	order := cand.InterviewOrder
	manual := cand.TierManual
	status := cand.Status

	if patch.Tier != nil {
		t := scanner.NormalizeTier(*patch.Tier)
		if !validTiers[t] {
			return fmt.Errorf("invalid tier: %s", t)
		}
		tier = t
		action = scanner.ActionForTier(tier)
		manual = true
		if tier != "S" {
			order = 0
		}
	}
	if patch.Action != nil {
		action = strings.TrimSpace(*patch.Action)
	}
	if patch.InterviewOrder != nil {
		order = *patch.InterviewOrder
	}
	if patch.Status != nil {
		st := strings.TrimSpace(*patch.Status)
		if !validStatuses[st] {
			return fmt.Errorf("invalid status: %s", st)
		}
		status = st
	}
	if patch.ClearManual {
		manual = false
		if strings.TrimSpace(cand.AutoTier) != "" {
			tier = scanner.NormalizeTier(cand.AutoTier)
			action = scanner.ActionForTier(tier)
		}
	}
	manualInt := 0
	if manual {
		manualInt = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`UPDATE candidates SET tier=?, action=?, interview_order=?, tier_manual=?, status=?, updated_at=? WHERE id=?`,
		tier, action, order, manualInt, status, now, id)
	if err != nil {
		return err
	}
	// 档位或面试顺序变化后，把同批 S 档压成 1..n，避免列表出现 3、5、7 空档
	if patch.InterviewOrder != nil || patch.Tier != nil {
		movedID := int64(0)
		desired := 0
		if patch.InterviewOrder != nil && tier == "S" {
			movedID = id
			desired = *patch.InterviewOrder
		}
		return s.compactSInterviewOrders(cand.BatchID, movedID, desired)
	}
	return nil
}

func (s *Store) compactAllSInterviewOrders() error {
	rows, err := s.db.Query(`SELECT DISTINCT batch_id FROM candidates WHERE tier='S'`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var batchIDs []int64
	for rows.Next() {
		var batchID int64
		if err := rows.Scan(&batchID); err != nil {
			return err
		}
		batchIDs = append(batchIDs, batchID)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, batchID := range batchIDs {
		if err := s.compactSInterviewOrders(batchID, 0, 0); err != nil {
			return err
		}
	}
	return nil
}

// compactSInterviewOrders 按现有顺序把同批 S 档编号成连续 1..n。
// movedID>0 时先抽出该人，再插入 desired（1-based）位置，用于「上移/下移」而不留下空号。
func (s *Store) compactSInterviewOrders(batchID, movedID int64, desired int) error {
	query := `SELECT id FROM candidates WHERE tier='S'`
	args := []any{}
	if batchID > 0 {
		query += ` AND batch_id=?`
		args = append(args, batchID)
	}
	query += ` ORDER BY CASE WHEN interview_order>0 THEN interview_order ELSE 999 END, eng_score DESC, id`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	ids := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if movedID > 0 && id == movedID {
			continue
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if movedID > 0 {
		pos := desired - 1
		if pos < 0 || pos > len(ids) {
			pos = len(ids)
		}
		ids = append(ids[:pos], append([]int64{movedID}, ids[pos:]...)...)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for i, id := range ids {
		if _, err := tx.Exec(`UPDATE candidates SET interview_order=?, updated_at=? WHERE id=?`, i+1, now, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

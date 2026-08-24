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
	return err
}

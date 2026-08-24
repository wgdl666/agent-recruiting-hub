package models

import "time"

type Candidate struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Source         string    `json:"source"`
	Tier           string    `json:"tier"`
	AutoTier       string    `json:"auto_tier,omitempty"`
	EngSummary     string    `json:"eng_summary"`
	ProjectSummary string    `json:"project_summary"`
	OneLiner       string    `json:"one_liner"`
	Action         string    `json:"action"`
	ResumePath     string    `json:"resume_path,omitempty"`
	ResumeKey      string    `json:"resume_key,omitempty"`
	ScoreTotal     int       `json:"score_total"`
	EngScore       int       `json:"eng_score"`
	AgentScore     int       `json:"agent_score"`
	Reason         string    `json:"reason,omitempty"`
	Flags          []string  `json:"flags,omitempty"`
	TierManual     bool      `json:"tier_manual"`
	InterviewOrder int       `json:"interview_order"`
	BatchID        int64     `json:"batch_id"`
	BatchName      string    `json:"batch_name,omitempty"`
	Status         string    `json:"status"`
	HasResume      bool      `json:"has_resume"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type InterviewQuestion struct {
	ID          int64  `json:"id"`
	CandidateID int64  `json:"candidate_id"`
	SortOrder   int    `json:"sort_order"`
	Question    string `json:"question"`
	Answer      string `json:"answer,omitempty"`
	Level       string `json:"level,omitempty"`
}

type CandidateDetail struct {
	Candidate
	Questions []InterviewQuestion `json:"questions"`
}

type UploadResult struct {
	Imported int              `json:"imported"`
	BatchID  int64            `json:"batch_id,omitempty"`
	Results  []ScreenResult   `json:"results"`
	Errors   []string         `json:"errors,omitempty"`
}

type ScreenResult struct {
	Name       string `json:"name"`
	Tier       string `json:"tier"`
	ScoreTotal int    `json:"score_total"`
	EngScore   int    `json:"eng_score"`
	AgentScore int    `json:"agent_score"`
	Reason     string `json:"reason"`
	CandidateID int64 `json:"candidate_id,omitempty"`
}

type Stats struct {
	Total    int            `json:"total"`
	ByTier   map[string]int `json:"by_tier"`
	ByStatus map[string]int `json:"by_status"`
}

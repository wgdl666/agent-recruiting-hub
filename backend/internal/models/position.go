package models

import "time"

// Position 正在招聘的岗位。检验标准走 Skill（评分 prompt），岗位名才是列表/上传里看到的「岗」。
type Position struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	SkillID        string    `json:"skill_id"`
	SkillName      string    `json:"skill_name,omitempty"`
	Description    string    `json:"description,omitempty"`
	JD             string    `json:"jd,omitempty"`
	SortOrder      int       `json:"sort_order"`
	IsOpen         bool      `json:"is_open"`
	CandidateCount int       `json:"candidate_count,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

package models

import "time"

type Batch struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Tag        string    `json:"tag"`
	PeriodType string    `json:"period_type"` // daily | weekly | monthly | custom
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Stats      *BatchStats `json:"stats,omitempty"`
}

type BatchStats struct {
	Total        int            `json:"total"`
	ByStatus     map[string]int `json:"by_status"`
	ByTier       map[string]int `json:"by_tier"`
}

type PipelineStats struct {
	BatchID   int64          `json:"batch_id,omitempty"`
	BatchName string         `json:"batch_name,omitempty"`
	Total     int            `json:"total"`
	ByStatus  map[string]int `json:"by_status"`
	ByTier    map[string]int `json:"by_tier"`
}

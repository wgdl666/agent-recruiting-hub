package seed

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ScoredRecord struct {
	Name       string   `json:"name"`
	Batch      string   `json:"batch"`
	Text       string   `json:"text"`
	EngScore   int      `json:"eng_score"`
	AgentScore int      `json:"agent_score"`
	Total      int      `json:"total"`
	Tier       string   `json:"tier"`
	Reason     string   `json:"reason"`
	Flags      []string `json:"flags"`
}

func LoadScored(root string) (map[string]ScoredRecord, error) {
	path := filepath.Join(root, "seed", "scored.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []ScoredRecord
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	out := make(map[string]ScoredRecord, len(rows))
	for _, r := range rows {
		out[r.Name] = r
	}
	return out, nil
}

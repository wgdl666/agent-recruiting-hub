package skills

import "strings"

// Skill 是检验标准（评分 prompt），不是招聘岗位本身。岗位在 positions 表，上传时选岗位再套本 Skill。
// 目前只开放 intern，对应仓库内 .cursor/skills/agent-intern-recruiting。

const Intern = "intern"

type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func All() []Skill {
	return []Skill{
		{
			ID:          Intern,
			Name:        "实习生",
			Description: "Agent 工程实习：上手就能干活",
		},
	}
}

func Lookup(id string) (Skill, bool) {
	id = strings.TrimSpace(id)
	for _, s := range All() {
		if s.ID == id {
			return s, true
		}
	}
	return Skill{}, false
}

func Valid(id string) bool {
	_, ok := Lookup(id)
	return ok
}

package skills

import "strings"

// 评估岗位标准就是招聘 Skill：上传前必须选定，评分按该岗的标准走。
// 目前只开放实习生岗，对应仓库内 .cursor/skills/agent-intern-recruiting。

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

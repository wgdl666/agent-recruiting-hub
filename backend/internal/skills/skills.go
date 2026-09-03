package skills

import "strings"

// Skill 是检验标准（评分 prompt），不是招聘岗位本身。岗位在 positions 表，上传时选岗位再套本 Skill。
// intern 对应仓库内 .cursor/skills/agent-intern-recruiting。
// image_software_test 用独立评分 prompt，禁止按 Agent 实习生标准把测试简历整批打成淘汰。

const Intern = "intern"
const ImageSoftwareTest = "image_software_test"

// ImageSoftwareTestDescription 内置岗简述，positions 表启动时与此对齐。
const ImageSoftwareTestDescription = "约3年执行型 QA：IQ 画质（AWB/AE/清晰度/影调 + 实验室指标）+ 完整安卓测试"

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
		{
			ID:          ImageSoftwareTest,
			Name:        "影像 && 软件 测试",
			Description: ImageSoftwareTestDescription,
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

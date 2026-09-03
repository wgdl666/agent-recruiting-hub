package scanner

import (
	"strings"

	_ "embed"

	"github.com/caden/agent-recruiting-hub/internal/skills"
)

//go:embed scoring_prompt.txt
var scoringPrompt string

//go:embed scoring_prompt_image_software_test.txt
var scoringPromptImageSoftwareTest string

//go:embed questions_prompt.txt
var questionsPrompt string

//go:embed questions_prompt_image_software_test.txt
var questionsPromptImageSoftwareTest string

// scoringPromptFor 按岗位 Skill 选评分说明；测开岗不能再用实习生 Agent 标准，否则会被整批误打成淘汰。
func scoringPromptFor(skillID string) string {
	switch strings.TrimSpace(skillID) {
	case skills.ImageSoftwareTest:
		return scoringPromptImageSoftwareTest
	default:
		return scoringPrompt
	}
}

func questionsPromptFor(skillID string) string {
	switch strings.TrimSpace(skillID) {
	case skills.ImageSoftwareTest:
		return questionsPromptImageSoftwareTest
	default:
		return questionsPrompt
	}
}

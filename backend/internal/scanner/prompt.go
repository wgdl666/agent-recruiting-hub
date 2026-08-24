package scanner

import _ "embed"

//go:embed scoring_prompt.txt
var scoringPrompt string

//go:embed questions_prompt.txt
var questionsPrompt string

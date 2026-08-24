package models

// Candidate workflow status (stored in DB as English key).
const (
	StatusScreening    = "screening"
	StatusRead         = "read"
	StatusToInterview  = "to_interview"
	StatusInterviewing = "interviewing"
	StatusPassed       = "passed"
	StatusCompleted    = "completed"
	StatusRejected     = "rejected"
)

func ValidStatuses() []string {
	return []string{
		StatusScreening,
		StatusRead,
		StatusToInterview,
		StatusInterviewing,
		StatusPassed,
		StatusCompleted,
		StatusRejected,
	}
}

func StatusLabel(s string) string {
	switch s {
	case StatusScreening:
		return "简历筛选"
	case StatusRead:
		return "已阅"
	case StatusToInterview:
		return "待约面"
	case StatusInterviewing:
		return "面试中"
	case StatusPassed:
		return "通过"
	case StatusCompleted:
		return "面试完成"
	case StatusRejected:
		return "本轮淘汰"
	default:
		return "简历筛选"
	}
}

// PipelineOrder for Kanban columns left-to-right.
func PipelineOrder() []string {
	return []string{
		StatusScreening,
		StatusRead,
		StatusToInterview,
		StatusInterviewing,
		StatusPassed,
		StatusCompleted,
		StatusRejected,
	}
}

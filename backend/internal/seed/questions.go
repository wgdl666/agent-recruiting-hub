package seed

// ManualTier overrides auto-scoring tier for known candidates.
var ManualTier = map[string]string{
	"苏志祺": "S", "钟嘉漩": "S", "黄浩翔": "S", "何鑫奎": "S", "尤曙宏": "S",
	"林嘉翰": "S", "王嘉麟": "S", "夏钰林": "S", "郭轩": "S", "吴鹏举": "S", "肖昌未": "S",
	"张尘朴": "A", "傅鸿东": "A", "聂武永": "A", "张恒滔": "A", "屈玉超": "A", "彭泰烨": "A",
	"钱亮亮": "C", "梁贤云": "C",
}

var Summaries = map[string]struct{ Eng, Proj, One string }{
	"苏志祺": {"Redis/RocketMQ/PG、幂等、降级", "AI Chat四层记忆+Agent Loop", "后端+Agent双强"},
	"钟嘉漩": {"FastAPI、Redis/MySQL、MCP队列DLQ", "MindBridge Harness+EchoMind", "Agent工程化完整"},
	"黄浩翔": {"FlyGPT平台、上下文预算", "小天才87%压缩、检索-75%", "上线闭环有数字"},
	"何鑫奎": {"FastAPI/SSE/Redis", "LangGraph NL2SQL+Multi-Agent评测", "问数Agent深挖"},
	"尤曙宏": {"FastAPI、权限链、Session", "CodePaceX eval 20/20", "Coding Agent强"},
	"林嘉翰": {"RocketMQ事务+Lua幂等", "实验室Agent/RAG", "后端深挖+Agent"},
	"王嘉麟": {"多租户/限流/沙箱", "JLClaw Harness", "Harness匹配"},
	"夏钰林": {"Redis/ES/压测", "119 LangGraph Agent", "多轮Agent闭环"},
	"郭轩":   {"Redis/MySQL、tool熔断", "InsightForge多Agent", "并发tool分区"},
	"吴鹏举": {"SQLite checkpoint、工具Schema", "Mini-Coding Agent+RAG/Milvus", "OCR：Agent框架深挖"},
	"肖昌未": {"FastAPI/MySQL/Redis/Celery", "AI学习平台 sole dev", "OCR：RAG+KG+Agent"},
}

// InterviewOrder is the recommended S-tier interview sequence (lower = earlier).
var InterviewOrder = map[string]int{
	"苏志祺": 1, "钟嘉漩": 2, "黄浩翔": 3, "何鑫奎": 4,
	"吴鹏举": 5, "肖昌未": 6, "尤曙宏": 7, "林嘉翰": 8,
	"王嘉麟": 9, "夏钰林": 10, "郭轩": 11,
}

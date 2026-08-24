package seed

// Questions keyed by candidate name (S-tier from manual screening).
var Questions = map[string][]string{
	"苏志祺": {
		"Like&Visit 任务 Redis 积压时，你的 Zset 游标消费方案和传统 scan 入队比，解决了什么具体问题？",
		"AI Chat 四层记忆（L0-L3）各存什么？strength 遗忘曲线怎么参与 topK 重排？",
		"同会话并发写坏记忆的问题，LaneManager 串行+合并消息具体怎么实现？失败时怎么办？",
	},
	"钟嘉漩": {
		"MindBridge 里 MCP 异步工具队列：幂等创建、DLQ、限流各在什么环节触发？",
		"Harness 把 Runtime、落库、tool 后处理拆开——state/契约在各层之间传什么？",
		"EchoMind 三路意图融合，低置信度降级策略是什么？有没有线上 bad case？",
	},
	"黄浩翔": {
		"数据库工具 99 万→12 万字符：按 CREATE TABLE 边界原子截断，如果截断处破坏语义怎么办？",
		"混合检索误用 rerank 原始分而非融合分的 bug，你的新融合公式是什么？怎么回归验证？",
		"上下文污染（iframe 会话 ID 未重置）——你怎么定位根因？上线后怎么防再发？",
	},
	"何鑫奎": {
		"NL2SQL LangGraph：多路召回后 SQL 校验失败，纠错节点怎么回灌？最多重试几轮？",
		"Multi-Agent 研究助手：checkpointer 存哪些字段？服务重启后怎么恢复不串 thread？",
		"5 层评测体系里，哪一层最能发现「答案对但 tool 乱调」？举一个失败 case。",
	},
	"尤曙宏": {
		"CodePaceX 的 ToolSearch 延迟激活：50 个重型 tool 场景下，初始 schema 怎么降 90%？",
		"权限链 deny/ask/allow 对 shell 写文件怎么裁决？误拦/漏拦怎么测？",
		"20-task eval subset 20/20 scorable——ground truth 怎么定义？会不会过拟合？",
	},
	"林嘉翰": {
		"热门名额抢购：RocketMQ 事务消息 + Redis Lua 预占 + MySQL 异步确认，任一步失败怎么回滚/补偿？",
		"requestId 幂等 + 唯一约束 + 条件更新——三道防线分别防什么重复？",
		"Agent/RAG 扩展：Outbox 驱动索引构建，消费失败重试会不会重复 embedding？怎么保证一致？",
	},
	"王嘉麟": {
		"JLClaw Harness Profile 按 strong/standard/cheap 路由——fallback 触发条件是什么？",
		"多租户 (tenant, channel, session) → thread_id：飞书 WebSocket 和 CLI 怎么保证不串会话？",
		"隔离沙箱 gVisor/microVM 分级——tool 执行超时且不知副作用时怎么处理？",
	},
	"夏钰林": {
		"119 接警 LangGraph：场景切换时 Redis 里槽位/历史怎么清理？最大轮次到了怎么降级？",
		"ES 地址检索 Top-3 93%——和 LLM 抽取的地址字段冲突时以谁为准？",
		"20 并发 10 分钟压测 P99——瓶颈在哪一层？失败重试会不会放大延迟？",
	},
	"郭轩": {
		"InsightForge 工具并发分区：哪些 tool 可并行、哪些必须串行？注册时怎么声明？",
		"ReAct loop max 5 次 tool call + 熔断——超时后用户看到什么？状态存哪？",
		"云鹤智标 8.5x 加速的「预编排+并发+审计」——如果审计失败，已生成的片段怎么处理？",
	},
	"吴鹏举": {
		"Mini-Coding Agent 双层架构 QueryEngine + Query()：session 恢复时 checkpoint 存哪些字段？",
		"auto-compact 渐进压缩：什么条件下触发？压缩后 tool 结果/pruning 策略会不会丢关键栈信息？",
		"RAG 混合检索 Strategy + 自动 fallback：fallback 触发条件？RAGAS 四指标哪个最常暴露问题？",
	},
	"肖昌未": {
		"AI 学习平台「向量+BM25 粗排→综合精排」：MRR 0.497→0.684 的主要改动是哪一步？",
		"Neo4j + FAISS + MySQL 四层存储：一次 Tutor 问答读/write 路径是什么？一致性怎么保证？",
		"LangGraph 学术助手动态 research plan：某一步 search 失败，图里 blocked 还是 replan？怎么实现？",
	},
}

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

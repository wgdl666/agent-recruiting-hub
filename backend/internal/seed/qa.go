package seed

// QA is a leveled interview question with interviewer reference answer.
type QA struct {
	Level    string // L1 | L2 | L3
	Question string
	Answer   string
}

// Questions keyed by candidate name (S-tier).
var Questions = map[string][]QA{
	"苏志祺": {
		{
			Level:    "L1",
			Question: "Like&Visit 任务在 Redis 里堆了很多，你最后怎么消费的？和一开始直接 scan 全部 key 比，好在哪里？",
			Answer:   "好答案：scan 会一下子塞满队列、还可能重复消费；他用 Zset 按时间片增量拉、配合锁或消费组保证不丢不重。加分：积压时调批量大小、死信队列、看 lag。",
		},
		{
			Level:    "L2",
			Question: "AI Chat 四层记忆，每层大概放什么？快「忘掉」的记忆怎么参与排序？",
			Answer:   "L0 当前对话；L1 会话摘要；L2 用户偏好；L3 长期事实。越久 strength 越低，召回时相似度 × strength 再取 topK。可追问：新旧记忆冲突怎么处理。",
		},
		{
			Level:    "L2",
			Question: "同一会话里用户连发好几条，记忆会不会写乱？你怎么保证顺序？失败了怎么办？",
			Answer:   "按 session 分队列，同会话串行处理；短时间里多条消息合并再写。失败重试要带幂等 key，多次失败进死信并告警，读的时候可以回退到上一个 checkpoint。",
		},
		{
			Level:    "L3",
			Question: "如果 Agent 要调工具，你会从哪几层记忆里取内容？工具跑完的结果写回哪一层？",
			Answer:   "一般在 system 或第一条 user 前注入 L2/L3 摘要 + L1 会话摘要；每轮 tool 结果先放 L0，再异步提炼进 L1/L3。注意：原始 tool 输出不要全塞进长期记忆，只存结构化字段。",
		},
	},
	"钟嘉漩": {
		{
			Level:    "L1",
			Question: "MindBridge 里 MCP 工具要排队跑，重复提交、失败进死信、限流，分别在哪一步做的？",
			Answer:   "入队前用 request_id 防重复（Redis SETNX 或 DB 唯一键）；worker 失败 N 次进 DLQ；限流在网关和消费速率两边。重复提交应返回同一个 job_id，不重复执行。",
		},
		{
			Level:    "L2",
			Question: "Harness 把运行时、落库、工具后处理拆开了，这几块之间具体传什么？",
			Answer:   "运行时传 messages、tool_calls、trace_id；落库持久化 thread/checkpoint；后处理传校验过的 tool 结果（JSON schema）。用固定结构体/Pydantic，带版本号防兼容问题。",
		},
		{
			Level:    "L2",
			Question: "EchoMind 三路判断用户想干什么，置信度很低时你怎么处理？线上有没有踩过坑？",
			Answer:   "规则、小模型、LLM 加权或投票；太低就反问澄清或走默认 FAQ，别硬调工具。常见坑：口语省略主语导致 LLM 很自信但判错，要靠规则兜底或多问一句。",
		},
		{
			Level:    "L3",
			Question: "工具超时 30 秒了，但模型已经想好下一步，你怎么保证流程不乱？",
			Answer:   "tool 节点标 pending → 超时写 ToolMessage(error)；图里走 replan 或让用户确认，不能悄悄跳过。checkpoint 记下哪个 tool 失败了，重试要有明确策略。",
		},
	},
	"黄浩翔": {
		{
			Level:    "L1",
			Question: "数据库工具从 99 万字符压到 12 万，你怎么截的？会不会把表结构截断截坏了？",
			Answer:   "按表/语句边界截，保留完整 DDL；表太多就留 schema + 说明采样行数。还超长就按表重要性丢数据留结构，并在 prompt 里写明「已截断」。",
		},
		{
			Level:    "L2",
			Question: "混合检索那次 bug，rerank 分和融合分搞混了——你最后怎么算的？怎么确认修好了？",
			Answer:   "常见做法：dense、BM25、rerank 各自归一化后再加权融合。验证：固定一批 golden query，对比 MRR/Recall@K，改完跑回归别回退。",
		},
		{
			Level:    "L2",
			Question: "iframe 会话串了那次，你怎么查出来的？上线后怎么防再犯？",
			Answer:   "多 tab/iframe 共用 storage 导致 session_id 混了；看日志对比 session 和请求链。修复：路由切换时 reset session；加 E2E 测多会话隔离。",
		},
		{
			Level:    "L3",
			Question: "上下文压了 87%，哪些内容绝对不能压？压不下去怎么办？",
			Answer:   "必须留：system 约束、最近用户意图、没跑完的 tool 链、安全规则。可压：闲聊、重复检索片段。还不行就先摘要再截最旧几轮，最后拒答并提示缩小范围。",
		},
	},
	"何鑫奎": {
		{
			Level:    "L1",
			Question: "NL2SQL 生成的 SQL 校验没过，你的纠错流程怎么走？最多试几次？",
			Answer:   "校验节点给出具体错误（语法/表不存在/列不对）；纠错节点把错误 + 原 SQL + schema + 用户问题再喂模型改写，回到校验节点。一般 2–3 轮，还不行就转人工或说明查不了。",
		},
		{
			Level:    "L2",
			Question: "Multi-Agent 研究助手，checkpoint 里存什么？服务重启后怎么恢复、不串会话？",
			Answer:   "存 thread_id、checkpoint_id、messages/plan/tool 状态、user_id 等。恢复只按 thread_id 拉最新；多租户用 (tenant, thread_id) 做 key；重启后必须从 store 加载，不能靠内存。",
		},
		{
			Level:    "L2",
			Question: "五层评测里，哪一层最容易发现「答案看着对但工具乱调」？举个你自己的例子。",
			Answer:   "第 3–4 层（tool 轨迹/执行是否靠谱）最能抓。例子：问库存却查了订单表，数字碰巧接近；或同样 search 调了 5 次。看 tool 命中率、必要调用数、参数是否合法。",
		},
		{
			Level:    "L3",
			Question: "多路召回（schema、样例、关键词）结果对不上，比如表名冲突，你怎么合并？",
			Answer:   "各路打分后合并再 rerank；冲突的表/列可投票或用小模型 rerank；不确定的标成「候选」让模型自选。最好记每路贡献，方便线下归因。",
		},
		{
			Level:    "L3",
			Question: "SSE 往前端推 Agent 中间状态，断线重连会不会重复显示工具卡片？你怎么处理？",
			Answer:   "事件带递增 event_id；重连用 Last-Event-ID 补拉或 REST 拉 diff。前端按 event_id 去重，工具卡片用 call_id 做 key 幂等更新。",
		},
	},
	"尤曙宏": {
		{
			Level:    "L1",
			Question: "CodePaceX 有 50 个重型工具，你怎么让模型一开始不用看全部 schema？",
			Answer:   "首轮只暴露「搜工具/看工具详情」这类 meta-tool，或按意图检索 top-k schema；真用到再 lazy load 完整定义。也可用 embedding 找相关工具子集。",
		},
		{
			Level:    "L2",
			Question: "权限链里 deny/ask/allow，shell 写文件怎么判？你怎么测有没有拦错或漏拦？",
			Answer:   "路径白名单 + 读/写/执行类型；敏感路径直接 deny，未知的 ask 用户。用 fixture 命令集（rm、curl、写 /etc）自动化测；红队 case 测漏拦。",
		},
		{
			Level:    "L2",
			Question: "20 个 eval 任务全满分，标准答案怎么定的？会不会过拟合这 20 个？",
			Answer:   "GT 看预期文件变更、测试通过、输出匹配；多解任务用 rubric 打分。防过拟合：留 held-out 任务、换模型复测、别把 eval 样例写进 few-shot。",
		},
		{
			Level:    "L3",
			Question: "同一个 Session 里又有 coding agent 又有聊天，上下文怎么隔开？",
			Answer:   "子 session 或分支 thread；coding 独立 message 通道和沙箱；合并只通过显式「引用 diff」，不共享全量历史。",
		},
	},
	"林嘉翰": {
		{
			Level:    "L1",
			Question: "热门名额抢购：RocketMQ 事务 + Redis 预占 + MySQL 确认，哪一步失败了怎么回滚？",
			Answer:   "Redis 预占失败直接售罄；MQ 提交后 MySQL 写失败走补偿释放库存；消费幂等靠订单号。定时对账 Redis 和 DB 修库存漂移。",
		},
		{
			Level:    "L2",
			Question: "requestId 幂等、DB 唯一约束、条件更新，这三道分别防什么重复？",
			Answer:   "requestId 防客户端重试重复下单；unique 防并发双写；条件更新（where status=待支付 and stock>0）防状态机乱序。",
		},
		{
			Level:    "L2",
			Question: "Agent/RAG 用 Outbox 建索引，消费失败重试会不会重复 embedding？怎么保证一致？",
			Answer:   "事件带 doc_version；consumer 幂等 key=(doc_id, version)；embedding 前查索引是否已是同 hash。失败重试安全，靠 version 单调递增保证最终一致。",
		},
		{
			Level:    "L3",
			Question: "从抢购系统做到 Agent 服务，哪些后端经验你会保留？哪些不能照搬？",
			Answer:   "保留：幂等、限流、异步解耦、对账。别照搬：强一致库存模型套 LLM 状态；Agent 更适合 checkpoint + 可恢复长任务。",
		},
	},
	"王嘉麟": {
		{
			Level:    "L1",
			Question: "JLClaw 按 strong/standard/cheap 选模型，什么时候会往更强的模型 fallback？",
			Answer:   "超时、置信度低、tool 连续失败、token 超限、或高风险任务升级 strong。cheap 失败升 standard，还不行升 strong 或转人工。",
		},
		{
			Level:    "L2",
			Question: "飞书 WebSocket 和 CLI 共用一套 thread，怎么保证不串会话？",
			Answer:   "thread_id = hash(tenant, channel, session)；飞书用 open_chat_id + user_id，CLI 用 profile 路径。存储带 tenant 前缀，别用纯内存 map。",
		},
		{
			Level:    "L2",
			Question: "沙箱里 tool 超时了，又不知道有没有副作用，你怎么处理？",
			Answer:   "超时 kill 进程，标记 side_effect_unknown；后续默认只读直到用户确认；checkpoint 记 partial；写操作换新沙箱实例。",
		},
		{
			Level:    "L3",
			Question: "一次 agent 跑下来，你最少会打哪些日志/span 方便排查？",
			Answer:   "trace_id 贯穿；至少：llm_call、tool_exec、checkpoint_save、routing；带上 model、耗时、token、tool 名、错误码。",
		},
	},
	"夏钰林": {
		{
			Level:    "L1",
			Question: "119 接警 Agent 换场景时，Redis 里的槽位和历史怎么处理？聊太多轮了怎么办？",
			Answer:   "换场景清 slot，必要身份信息保留；历史太长先摘要再归档。超轮次：简短确认 + 转人工话术，打 flag 停止自动追问。",
		},
		{
			Level:    "L2",
			Question: "ES 地址检索 Top-3 命中率 93%，和模型抽出来的地址对不上时听谁的？",
			Answer:   "ES 结构化结果高置信优先，模型抽取作补全；冲突就并列候选让用户选，或有 geocode 的优先。记录冲突率线下调权。",
		},
		{
			Level:    "L2",
			Question: "20 并发压了 10 分钟，P99 卡在哪？失败重试会不会把延迟越拖越长？",
			Answer:   "常见瓶颈：LLM 排队、Redis、ES；看最慢 span。重试要退避 + jitter，限制总次数，否则雪崩。",
		},
		{
			Level:    "L3",
			Question: "接警场景误派单代价很高，你会加什么护栏？",
			Answer:   "高危 tool 二次确认；dry-run；权限和审计日志；模拟环境 replay。关键步骤必须经人工确认节点。",
		},
	},
	"郭轩": {
		{
			Level:    "L1",
			Question: "InsightForge 里哪些 tool 可以并行、哪些必须排队？注册时怎么声明？",
			Answer:   "只读检索可并行；写库、发消息、依赖上一步输出的要串行。注册时标 parallel_safe、requires_lock、side_effects，调度按 DAG 或 partition 执行。",
		},
		{
			Level:    "L2",
			Question: "ReAct 最多调 5 次 tool 就熔断，超时后用户看到什么？状态存哪？",
			Answer:   "展示已完成部分 +「步骤超限/超时，请缩小问题」+ 可继续问；状态存 thread checkpoint（messages + 半成品），可从中断点接着跑。",
		},
		{
			Level:    "L2",
			Question: "云鹤智标加速那套，如果审计没通过，已经生成的内容怎么处理？",
			Answer:   "标 draft/未发布；审计失败整批不入库或回滚上一版；留 audit trail 人工复核，不静默覆盖线上。",
		},
		{
			Level:    "L3",
			Question: "多个 Agent 互相调用，怎么防止死循环？",
			Answer:   "调用深度上限、记录已访问 agent、总 token/时间预算；检测到环就熔断并上报。",
		},
	},
	"吴鹏举": {
		{
			Level:    "L1",
			Question: "Mini-Coding Agent 恢复 session 时，checkpoint 里存了哪些东西？",
			Answer:   "messages、tool 结果、工作区快照指针、文件/cursor 状态、未完成的 tool call。恢复时 hydrate 工作区并重放没跑完的 tool。",
		},
		{
			Level:    "L2",
			Question: "上下文太长触发 auto-compact，压缩后会不会把关键报错栈弄丢？你怎么取舍？",
			Answer:   "token 或轮次超阈值触发；最近 K 轮完整保留 tool 输入输出，更早的只留摘要和文件路径；错误栈、测试失败日志标为不可删。",
		},
		{
			Level:    "L2",
			Question: "RAG 混合检索不行时会 fallback，什么时候触发？RAGAS 四个指标哪个最常暴露问题？",
			Answer:   "低置信、空召回、超时触发 fallback。RAGAS 里 context_precision、faithfulness 最常抓「检索偏了但答得像对的」。",
		},
		{
			Level:    "L3",
			Question: "简历是图片 PDF、信息不全，你会怎么设计人工补全和再评分？",
			Answer:   "标 thin/OCR 队列；人工补 tier 和摘要；tier_manual 锁定；重评不覆盖。面试以人工档位和现场深挖为准。",
		},
	},
	"肖昌未": {
		{
			Level:    "L1",
			Question: "学习平台检索 MRR 从 0.497 提到 0.684，你改的最关键一步是什么？",
			Answer:   "能说清具体改动：加 cross-encoder 精排、调融合权重、query 改写或 chunk 策略等，并说明哪步贡献最大、怎么 offline 验证。",
		},
		{
			Level:    "L2",
			Question: "Neo4j + FAISS + MySQL 四层存储，学生问一题 Tutor 的读写路径是什么？",
			Answer:   "读：MySQL 进度 → KG 查概念 → FAISS 召回片段 → 拼 prompt。写：答题记 MySQL，掌握度更新 KG，异步建索引。事件驱动 + version，最终一致。",
		},
		{
			Level:    "L2",
			Question: "学术助手某一步 search 失败了，流程是卡住还是换计划？你怎么实现的？",
			Answer:   "失败且未超重试次数就重试 search；否则 replan 换关键词或只用已有资料。只有等人工输入时才 blocked。",
		},
		{
			Level:    "L3",
			Question: "一个人开发，你怎么给 Agent 链路做最小可用的评测？",
			Answer:   "10–20 条 golden task（含失败模式）；看 tool 轨迹 + 最终答案；每次改 prompt/graph 跑回归；不追求大而全 benchmark。",
		},
	},
}

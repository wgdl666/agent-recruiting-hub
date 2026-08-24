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
			Question: "Like&Visit 任务 Redis 积压时，你的 Zset 游标消费方案和传统 scan 入队比，解决了什么具体问题？",
			Answer:   "期望听到：scan 会把全量 key 一次性推入队列导致内存尖峰和重复消费；Zset 游标按 score/时间片增量拉取，配合 consumer group 或分布式锁保证 at-least-once 不丢不重。加分：积压时动态调 batch size、死信队列、监控 lag。",
		},
		{
			Level:    "L2",
			Question: "AI Chat 四层记忆（L0-L3）各存什么？strength 遗忘曲线怎么参与 topK 重排？",
			Answer:   "L0 当前轮上下文；L1 会话摘要；L2 用户画像/偏好；L3 长期事实库。strength 随时间衰减（如指数/半衰期），召回时 score = 相似度 × strength × 时效权重，再 topK。追问：新事实与旧记忆冲突时 merge 还是版本化。",
		},
		{
			Level:    "L2",
			Question: "同会话并发写坏记忆的问题，LaneManager 串行+合并消息具体怎么实现？失败时怎么办？",
			Answer:   "按 session_id 分 lane，同 lane 内单 goroutine/队列串行处理；短时间多条 user 消息合并成一批再写记忆。失败：重试带幂等 key，超限进 DLQ 并告警，读路径可回退到上一 checkpoint。",
		},
		{
			Level:    "L3",
			Question: "如果要把四层记忆接到 Agent tool loop，你会在哪一层注入 retrieved memory？tool 结果写回哪一层？",
			Answer:   "注入：system/首条 user 前注入 L2-L3 摘要 + L1 会话摘要；每轮 tool 后把结构化 observation 写 L0，异步提炼进 L1/L3。关键：tool 原始输出不全进长期层，只存结构化槽位，防污染。",
		},
	},
	"钟嘉漩": {
		{
			Level:    "L1",
			Question: "MindBridge 里 MCP 异步工具队列：幂等创建、DLQ、限流各在什么环节触发？",
			Answer:   "入队前用 request_id 幂等（Redis SETNX/DB unique）；worker 消费失败 N 次进 DLQ；限流在 API 网关 + 队列消费速率（token bucket/按 tenant）。幂等创建指重复 submit 返回同一 job_id 不重复执行。",
		},
		{
			Level:    "L2",
			Question: "Harness 把 Runtime、落库、tool 后处理拆开——state/契约在各层之间传什么？",
			Answer:   "Runtime 传 messages + tool_calls + trace_id；落库层持久化 thread/checkpoint；后处理层传 normalized tool_result（JSON schema 校验后）。契约用 Pydantic/TypedDict 固定字段，版本号防兼容问题。",
		},
		{
			Level:    "L2",
			Question: "EchoMind 三路意图融合，低置信度降级策略是什么？有没有线上 bad case？",
			Answer:   "三路（规则/小模型/LLM）加权或投票；低于阈值走澄清问句或默认 FAQ，不强行调 tool。bad case 例：口语省略主语导致 LLM 高置信错误 intent，应用规则兜底或多轮确认。",
		},
		{
			Level:    "L3",
			Question: "MCP 工具超时 30s 但 LLM 已生成后续 plan，你怎么保证图状态一致？",
			Answer:   "tool 节点标记 pending→timeout 写 ToolMessage(error)；图 conditional edge 走 replan 或用户确认分支，禁止静默跳过。checkpoint 记录 last_failed_tool，重试需显式用户或自动 policy。",
		},
	},
	"黄浩翔": {
		{
			Level:    "L1",
			Question: "数据库工具 99 万→12 万字符：按 CREATE TABLE 边界原子截断，如果截断处破坏语义怎么办？",
			Answer:   "按表/语句边界截断，保留完整 DDL 块；超长表只留 schema + 采样行数说明。破坏语义时：优先丢 data 留 schema；仍超长则按表重要性排序截断，并在 prompt 声明「已截断」。",
		},
		{
			Level:    "L2",
			Question: "混合检索误用 rerank 原始分而非融合分的 bug，你的新融合公式是什么？怎么回归验证？",
			Answer:   "常见：score = α·dense + β·BM25 + γ·rerank_norm（各分量 min-max 或 z-score 后再融合）。回归：固定 golden query 集对比 MRR/Recall@K，CI 跑 diff 防回退。",
		},
		{
			Level:    "L2",
			Question: "上下文污染（iframe 会话 ID 未重置）——你怎么定位根因？上线后怎么防再发？",
			Answer:   "复现：多 tab/iframe 共享 storage；日志对比 session_id 与请求链路。修复：路由切换 reset session；防再发：E2E 测多会话隔离 + 监控异常 session 复用率。",
		},
		{
			Level:    "L3",
			Question: "87% 上下文压缩后，哪些信息绝对不能压？压缩失败时 fallback 策略？",
			Answer:   "保留：system 约束、最近 user 意图、未完成的 tool 链、安全策略。可压：历史闲聊、重复检索片段。fallback：超 token 则先 summarization，再截断最旧轮，最后拒答并提示缩小范围。",
		},
	},
	"何鑫奎": {
		{
			Level:    "L1",
			Question: "NL2SQL LangGraph：多路召回后 SQL 校验失败，纠错节点怎么回灌？最多重试几轮？",
			Answer:   "校验节点输出 structured error（语法/表不存在/列不匹配）；纠错节点把 error + 原 SQL + schema 片段 + 用户问题再喂 LLM rewrite。回灌走 graph 边回到 validate，不覆盖 checkpoint 历史。一般 max 2-3 轮，仍失败转人工或降级为「解释无法查询」。",
		},
		{
			Level:    "L2",
			Question: "Multi-Agent 研究助手：checkpointer 存哪些字段？服务重启后怎么恢复不串 thread？",
			Answer:   "存：thread_id、checkpoint_id、channel values（messages/plan/tool_state）、metadata（user_id、version）。恢复：仅用 thread_id 拉最新 checkpoint；多租户 key 用 (tenant, thread_id)；重启后 worker 无内存状态，必须从 store  hydrate。",
		},
		{
			Level:    "L2",
			Question: "5 层评测体系里，哪一层最能发现「答案对但 tool 乱调」？举一个失败 case。",
			Answer:   "第 3-4 层（tool trajectory / execution fidelity）最能抓：答案碰巧对但调错 API 或多余调用。case：问库存却查了订单表，返回数字碰巧接近；或并行调了 5 次相同 search。指标：tool 精确率、必要调用数、参数 schema 合法率。",
		},
		{
			Level:    "L3",
			Question: "NL2SQL 多路召回（schema link / example / keyword）结果冲突时，合并策略是什么？",
			Answer:   "分路打分后 union + rerank；冲突表/列用投票或 learned reranker；低置信度列进 prompt 标注「候选」让 LLM 自选。记录每路贡献便于离线归因。",
		},
		{
			Level:    "L3",
			Question: "SSE 推流到前端的 Agent 中间态，断线重连怎么补发？会不会重复渲染 tool 卡片？",
			Answer:   "事件带 monotonic event_id / sequence；重连用 Last-Event-ID 补拉或 REST 拉 checkpoint diff。前端按 event_id 去重；tool 卡片用 call_id 作 key 幂等更新。",
		},
	},
	"尤曙宏": {
		{
			Level:    "L1",
			Question: "CodePaceX 的 ToolSearch 延迟激活：50 个重型 tool 场景下，初始 schema 怎么降 90%？",
			Answer:   "首轮只暴露 meta-tool（search_tools / describe_tool）或按意图检索 top-k tool schema；用到再 lazy load 完整 JSON schema。或用 embedding 检索相关 tool 子集注入 prompt。",
		},
		{
			Level:    "L2",
			Question: "权限链 deny/ask/allow 对 shell 写文件怎么裁决？误拦/漏拦怎么测？",
			Answer:   "规则：路径白名单 + 操作类型（读/写/执行）；写敏感路径 deny，未知 ask 用户。测试：fixture 命令集（rm、curl、写 /etc）自动化断言；红队 case 测漏拦；误拦率用真实任务回放统计。",
		},
		{
			Level:    "L2",
			Question: "20-task eval subset 20/20 scorable——ground truth 怎么定义？会不会过拟合？",
			Answer:   "GT：预期文件变更/测试通过/stdout 匹配；部分任务多解用 rubric 打分。防过拟合：held-out 任务集、换模型复测、禁止把 eval 样例写进 prompt few-shot。",
		},
		{
			Level:    "L3",
			Question: "Session 里同时有 coding agent 和 chat，上下文隔离怎么做？",
			Answer:   "子 session / branch thread；coding 用独立 message channel 和 tool 沙箱；合并只通过显式「引用 diff」而非共享全量 history。",
		},
	},
	"林嘉翰": {
		{
			Level:    "L1",
			Question: "热门名额抢购：RocketMQ 事务消息 + Redis Lua 预占 + MySQL 异步确认，任一步失败怎么回滚/补偿？",
			Answer:   "Lua 预占失败直接返回售罄；MQ 半消息 commit 后 MySQL 写失败走补偿任务释放 Redis 库存；消费者幂等用订单号。定时对账 Redis vs DB 修库存漂移。",
		},
		{
			Level:    "L2",
			Question: "requestId 幂等 + 唯一约束 + 条件更新——三道防线分别防什么重复？",
			Answer:   "requestId：防客户端重试重复下单；DB unique：防并发双写；条件更新（where status=待支付 and stock>0）：防状态机乱序。",
		},
		{
			Level:    "L2",
			Question: "Agent/RAG 扩展：Outbox 驱动索引构建，消费失败重试会不会重复 embedding？怎么保证一致？",
			Answer:   "Outbox 事件带 doc_version；consumer 幂等 key=(doc_id, version)；embedding 前先查索引是否已同 hash。失败重试安全；最终一致靠 version 单调递增。",
		},
		{
			Level:    "L3",
			Question: "从抢购系统迁到 Agent 服务，你会保留哪些后端模式？哪些不该照搬？",
			Answer:   "保留：幂等、限流、异步解耦、对账。别照搬：强一致库存模型套到 LLM 状态；Agent 更适合 checkpoint + 可恢复长任务而非短事务。",
		},
	},
	"王嘉麟": {
		{
			Level:    "L1",
			Question: "JLClaw Harness Profile 按 strong/standard/cheap 路由——fallback 触发条件是什么？",
			Answer:   "触发：超时、置信度低、tool 连续失败、token 超限、或 policy 标记高风险任务升级 strong。cheap 失败自动升 standard；仍失败升 strong 或人工。",
		},
		{
			Level:    "L2",
			Question: "多租户 (tenant, channel, session) → thread_id：飞书 WebSocket 和 CLI 怎么保证不串会话？",
			Answer:   "thread_id = hash(tenant, channel, session_key)；飞书用 open_chat_id + user_id；CLI 用 profile 路径。存储分区带 tenant 前缀；绝不用纯内存 map 无 tenant。",
		},
		{
			Level:    "L2",
			Question: "隔离沙箱 gVisor/microVM 分级——tool 执行超时且不知副作用时怎么处理？",
			Answer:   "超时发 SIGKILL + 标记 side_effect_unknown；后续 tool 默认 read-only 直到用户确认；checkpoint 记录 partial；写操作需新 sandbox 实例。",
		},
		{
			Level:    "L3",
			Question: "Harness 如何做「可观测」：一次 agent run 你最少打哪些 span？",
			Answer:   "trace_id 贯穿；span：llm_call、tool_exec、checkpoint_save、routing；属性含 model、latency、token、tool_name、error_code。便于排 tool 乱调与延迟。",
		},
	},
	"夏钰林": {
		{
			Level:    "L1",
			Question: "119 接警 LangGraph：场景切换时 Redis 里槽位/历史怎么清理？最大轮次到了怎么降级？",
			Answer:   "场景切换清 slot key 保留必要身份字段；历史 summarization 后归档。超轮次：转简短确认 + 人工接管话术，写 flag 停止自动追问。",
		},
		{
			Level:    "L2",
			Question: "ES 地址检索 Top-3 93%——和 LLM 抽取的地址字段冲突时以谁为准？",
			Answer:   "结构化 ES 高置信优先；LLM 抽取作补全。冲突时并列展示候选让用户选，或规则：有 geocode 结果优先。记录冲突率离线调权。",
		},
		{
			Level:    "L2",
			Question: "20 并发 10 分钟压测 P99——瓶颈在哪一层？失败重试会不会放大延迟？",
			Answer:   "常见瓶颈：LLM 排队或 Redis/ES；P99 看最慢 span。重试需 exponential backoff + jitter，限总重试次数；否则雪崩放大延迟。",
		},
		{
			Level:    "L3",
			Question: "接警场景误触发 tool（如误派单）风险高，你会加什么护栏？",
			Answer:   "高危 tool 二次确认；dry-run 模式；RBAC；审计日志；模拟环境 replay。图里 critical 边必须经 human-in-the-loop 节点。",
		},
	},
	"郭轩": {
		{
			Level:    "L1",
			Question: "InsightForge 工具并发分区：哪些 tool 可并行、哪些必须串行？注册时怎么声明？",
			Answer:   "只读检索可并行；写库、发消息、依赖前步输出的串行。注册 metadata：parallel_safe、requires_lock、side_effects。调度器按 DAG 或 partition key 执行。",
		},
		{
			Level:    "L2",
			Question: "ReAct loop max 5 次 tool call + 熔断——超时后用户看到什么？状态存哪？",
			Answer:   "用户看到：已完成部分 +「步骤超限/超时，请缩小问题」+ 可续问按钮。状态存 thread checkpoint（messages + partial artifacts），可从中断点 resume。",
		},
		{
			Level:    "L2",
			Question: "云鹤智标 8.5x 加速的「预编排+并发+审计」——如果审计失败，已生成的片段怎么处理？",
			Answer:   "片段标记 draft/未发布；审计失败整批不入库或回滚到上一版本；保留 audit trail 供人工复核，不静默覆盖线上。",
		},
		{
			Level:    "L3",
			Question: "多 Agent 互相调用时，怎么防止循环调用？",
			Answer:   "调用深度上限、visited agent set、总 token/时间预算；检测到环则熔断并上报。",
		},
	},
	"吴鹏举": {
		{
			Level:    "L1",
			Question: "Mini-Coding Agent 双层架构 QueryEngine + Query()：session 恢复时 checkpoint 存哪些字段？",
			Answer:   "messages、tool_results、workspace snapshot 指针、cursor/file state、pending tool call。恢复时 hydrate 工作区 + 重放未完成 tool。",
		},
		{
			Level:    "L2",
			Question: "auto-compact 渐进压缩：什么条件下触发？压缩后 tool 结果/pruning 策略会不会丢关键栈信息？",
			Answer:   "触发：token > 阈值或轮次 > N。保留最近 K 轮完整 tool I/O，更早的只留 summary + 文件路径；错误栈、测试失败日志标为 do_not_prune。",
		},
		{
			Level:    "L2",
			Question: "RAG 混合检索 Strategy + 自动 fallback：fallback 触发条件？RAGAS 四指标哪个最常暴露问题？",
			Answer:   "fallback：低置信、空召回、超时。RAGAS 里 context_precision / faithfulness 最常抓「检索偏了但答得像对的」。",
		},
		{
			Level:    "L3",
			Question: "图片 PDF 简历信息不全时，你会如何设计人工补全与再评分流程？",
			Answer:   "标 thin/OCR 队列；人工补 tier 与摘要；tier_manual 锁定；重评不覆盖。面试以人工档位和现场深挖为准。",
		},
	},
	"肖昌未": {
		{
			Level:    "L1",
			Question: "AI 学习平台「向量+BM25 粗排→综合精排」：MRR 0.497→0.684 的主要改动是哪一步？",
			Answer:   "期望具体说：加 cross-encoder 精排、或调融合权重、或 query 改写、或 chunk 策略。要能说出哪一步贡献最大及 offline 验证方式。",
		},
		{
			Level:    "L2",
			Question: "Neo4j + FAISS + MySQL 四层存储：一次 Tutor 问答读/write 路径是什么？一致性怎么保证？",
			Answer:   "读：MySQL 用户进度 → KG 查概念关系 → FAISS 召回片段 → 组装 prompt。写：答题记录 MySQL，掌握度更新 KG；异步建索引。一致：事件驱动 + version，最终一致。",
		},
		{
			Level:    "L2",
			Question: "LangGraph 学术助手动态 research plan：某一步 search 失败，图里 blocked 还是 replan？怎么实现？",
			Answer:   "conditional edge：失败且 retry<max 回 search；否则 replan 换关键词或降级仅用已有资料。blocked 仅用于等人工输入。",
		},
		{
			Level:    "L3",
			Question: "sole dev 情况下，你如何给 Agent 链路做最小可行评测？",
			Answer:   "10-20 条 golden task（含失败模式）；看 tool 轨迹 + 最终答案；每次改 prompt/graph 跑回归；不追求大而全 benchmark。",
		},
	},
}

# 扩展开发

## 修改 S 档面试题

S 档上传后会按简历 **LLM 出题**（prompt：`backend/internal/scanner/questions_prompt.txt`）。题目要求务实、绑项目；参考答案写成 briefing，让不一定做过 Agent 的面试官也能问、能判。

页面上点题目会弹出参考答案；详情里可「按简历重新生成」。

无模型或生成失败时，回落 `backend/internal/seed/qa.go` 里的种子题：

```go
var Questions = map[string][]QA{
  "候选人姓名": {
    {Level: "L1", Question: "...", Answer: "面试官期望要点..."},
  },
}
```

批量刷新（S 档走模型，失败再用种子题）：

```bash
curl -X POST http://localhost:8080/api/sync/questions
```

单人重新生成：

```bash
curl -X POST http://localhost:8080/api/candidates/1/questions/generate
```

## 调整人工档位

`backend/internal/seed/questions.go` 中 `ManualTier`、`Summaries`、`InterviewOrder`。

重评时始终更新自动分与 `auto_tier`；`tier_manual=1` 时仅锁定展示档位，不阻止自动评估。上传/重传简历会清除手动锁定并重新自动定档。

## 改评分规则

用人标准（上手就做 / 聪明度 / 广度与深度 / 工程意识）见 `docs/hiring.md`。

上传前必须选定 **招聘岗位**。每个岗位绑定一个检验标准（招聘 Skill）。当前仅 `intern` / 实习生，定义在 `backend/internal/skills/skills.go`；在招岗位在 `positions` 表，侧栏「岗位阶梯」可增改。新增检验标准时在 skills.go 登记并补对应评分 prompt。

- LLM：`backend/internal/scanner/scoring_prompt.txt`（上传/拖入时走 ModelHub）
- 启发式：`backend/internal/scanner/scanner.go` — 关键词与权重。

改完后：

```bash
curl -X POST http://localhost:8080/api/rescreen   # 启发式重评，不走 ModelHub
```

**ModelHub LLM 评分仅在上传 PDF/ZIP 或 `POST /api/screen` 时触发。**

## 新增 API

1. 在 `backend/internal/server/server.go` 注册路由  
2. 业务逻辑放 `store/` 或新 package  
3. 前端 `frontend/src/api/client.ts` 增加方法  

## 新增前端页面

1. `frontend/src/views/` 新建 Vue 组件  
2. `TabShell.vue` 增加固定标签，或 `router/index.ts` 加路由  
3. 复用 `useTabs()` 切换标签  

## 更新知识库

直接编辑 `docs/*.md`，部署后刷新网页「知识库」即可（API 读磁盘，无需改前端）。

新增文档：

1. 在 `docs/` 添加 `your-topic.md`  
2. 在 `backend/internal/server/docs.go` 的 `docCatalog` 增加条目（标题 + slug + 排序）  

## 本地调试远端 API

```bash
export HUB_URL=http://127.0.0.1:8808/api   # 隧道后
./scripts/hub list S
```

## 提交规范

日常在 **`main`** 开发并发布，详见 [分支与发布](workflow.md)。

```bash
git checkout main
git add ...
git commit -m "简述 why"
git push origin main
```

确认可上线后 `make deploy`。

## 安全提示

- 当前 **无登录鉴权**，仅适合内网 / SSH 隧道  
- 勿将 `data/hub.db` 提交到 Git  
- 简历含个人信息，注意访问控制  

如需公网 HTTPS，可在 nginx 反代 8808 并加 Basic Auth 或 SSO。

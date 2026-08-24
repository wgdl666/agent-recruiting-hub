# 扩展开发

## 修改 S 档面试题

编辑 `backend/internal/seed/qa.go`：

```go
var Questions = map[string][]QA{
  "候选人姓名": {
    {Level: "L1", Question: "...", Answer: "面试官期望要点..."},
  },
}
```

同步到数据库：

```bash
curl -X POST http://localhost:8080/api/sync/questions
```

## 调整人工档位

`backend/internal/seed/questions.go` 中 `ManualTier`、`Summaries`、`InterviewOrder`。

重评时 `tier_manual=0` 的候选人会重新自动分；`ManualTier` 在 rescreen 时仍生效。

## 改评分规则

`backend/internal/scanner/scanner.go` — 关键词与权重。

改完后：

```bash
curl -X POST http://localhost:8080/api/rescreen
```

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

```bash
git add ...
git commit -m "简述 why"
git push origin main
make deploy    # 需要发布时
```

## 安全提示

- 当前 **无登录鉴权**，仅适合内网 / SSH 隧道  
- 勿将 `data/hub.db` 提交到 Git  
- 简历含个人信息，注意访问控制  

如需公网 HTTPS，可在 nginx 反代 8808 并加 Basic Auth 或 SSO。

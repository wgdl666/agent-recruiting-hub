# Agent Recruiting Hub — Cursor Agent Guide

本地 Agent 实习招聘评估台。标准：**上手就能干活** — 聪明度与潜力、广度与深度、工程意识（对耗时/失败敏感）；硬门槛仍是真实实习 + 项目上线。

**Git 分支**：日常开发与发布都在 `main`，确认后 `make deploy`（详见知识库「分支与发布」）。

## 启动

```bash
make install    # 首次
make backend    # :8080 API + 生产静态资源（需 make build）
make frontend   # :5173 开发前端（代理到 8080）
```

生产单端口：`make run` → http://localhost:8080  
远端生产：https://recruit.wgdl.tech（`make deploy` → mirror_zby）

## 数据位置

| 路径 | 说明 |
|------|------|
| `data/hub.db` | SQLite 候选人库 |
| `data/resumes/` | 简历 PDF 文件 |
| `seed/batch.json` | 种子筛选结果（47人） |

## REST API（Cursor 可直接 curl）

```bash
# 健康检查
curl http://localhost:8080/api/health

# 列表（工程优先排序：eng_first=true；按批次 batch_id；按岗位 position_id）
curl 'http://localhost:8080/api/candidates?tier=S&batch_id=1&eng_first=true'
curl 'http://localhost:8080/api/candidates?position_id=1'

# 岗位阶梯（正在招聘的岗位；每个岗位绑定一个检验 Skill）
curl http://localhost:8080/api/positions
curl 'http://localhost:8080/api/positions?open=true'
curl -X POST http://localhost:8080/api/positions \
  -H 'Content-Type: application/json' \
  -d '{"name":"实习生","skill_id":"intern"}'

# 检验标准（招聘 Skill；岗位绑定后上传时自动选用）
curl http://localhost:8080/api/skills

# 招聘批次（持续上传：按日/周/月自动归批）
curl http://localhost:8080/api/batches
curl -X POST http://localhost:8080/api/batches \
  -H 'Content-Type: application/json' \
  -d '{"auto":true,"period_type":"daily"}'
curl 'http://localhost:8080/api/pipeline/stats?batch_id=1'

# 调整进度状态（screening → read → to_interview → interviewing → passed/completed/rejected）
curl -X PATCH http://localhost:8080/api/candidates/1 \
  -H 'Content-Type: application/json' \
  -d '{"status":"to_interview"}'

# 调整档位（手动锁定，重评不覆盖）
curl -X PATCH http://localhost:8080/api/candidates/1 \
  -H 'Content-Type: application/json' \
  -d '{"tier":"A"}'

# 恢复自动档位
curl -X PATCH http://localhost:8080/api/candidates/1 \
  -H 'Content-Type: application/json' \
  -d '{"clear_manual":true}'

# 拖入等价：上传 PDF/ZIP（必须指定招聘岗位 position_id，或兼容旧客户端的 skill_id；默认归入当日批次）
curl -F 'files=@resume.pdf' -F 'source=cursor' -F 'position_id=1' -F 'period_type=daily' http://localhost:8080/api/upload
curl -F 'files=@resume.pdf' -F 'skill_id=intern' -F 'batch_id=2' http://localhost:8080/api/upload

# 批量评估本地目录（Cursor 常用；skill_id 会映射到对应在招岗位）
curl -X POST http://localhost:8080/api/screen \
  -H 'Content-Type: application/json' \
  -d '{"path":"/Users/caden/Downloads/resumes","source":"cursor","skill_id":"intern"}'

# 导入种子数据
curl -X POST http://localhost:8080/api/import/seed

# 导出飞书文档（需 lark-cli 已登录）
curl -X POST http://localhost:8080/api/export/feishu

# 重评所有已存简历
curl -X POST http://localhost:8080/api/rescreen

# CLI 快捷脚本
./scripts/hub list S
./scripts/hub screen ~/Downloads/resumes.zip
./scripts/hub feishu

# 简历 PDF
open http://localhost:8080/api/candidates/1/resume
```

## 评分逻辑

- 配置 `HUB_MODELHUB_ADDRESS` 后，**上传/拖入简历**时走 **wgModelHub** → `gemini-2.5-flash`（评分 prompt：`backend/internal/scanner/scoring_prompt.txt`；S 档面试题 prompt：`backend/internal/scanner/questions_prompt.txt`）
- 未配置或 RPC 失败时回退关键词启发式
- **全库重评**（`POST /api/rescreen`）仅用启发式，不调用 ModelHub
- `GET /api/health` 的 `scanner` 字段：`modelhub` | `heuristic`
- `seed/questions.go`：S 档 11 人各 3 条定制面试题
- `seed/questions.go` ManualTier：人工档位覆盖
- 图片 PDF：尝试 `python3 backend/scripts/ocr_resume.py`（需 pymupdf）

## 典型 Cursor 工作流

1. 用户丢 zip 到 `~/Downloads/` → Agent 调 `/api/screen` 批量评估
2. Agent `GET /api/candidates?tier=S` 拉 S 档列表
3. Agent `GET /api/candidates/:id` 读面试题，辅助面试
4. `GET /api/export` 同步到飞书文档

## 前端组件（Vue + Element Plus）

- `CandidateTable` — 列表（含岗位列），工程分 E / Agent 分 A
- `ResumeViewer` — iframe 简历
- `QuestionPanel` — 定制面试题
- `UploadDropzone` — 拖拽 PDF/ZIP（先选招聘岗位，自动套用该岗 Skill）
- `PositionsView` — 岗位阶梯：在招岗位与检验标准
- `DocsView` — 开发知识库（`docs/*.md`，API `/api/docs`）

## 知识库

- 网页顶部 **知识库** 标签，或 `http://localhost:8080/?tab=docs`
- 源码：`docs/` 目录（快速开始、部署、API、扩展开发）
- 更新文档后 `make deploy` 即可，无需改前端

# Agent Recruiting Hub — Cursor Agent Guide

本地 Agent 实习招聘评估台。标准：**实习经历 + 传统工程底座 + 单项目深挖**，工程落地经验优先。

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

# 列表（工程优先排序：eng_first=true；按批次 batch_id）
curl 'http://localhost:8080/api/candidates?tier=S&batch_id=1&eng_first=true'

# 招聘批次（持续上传：按日/周/月自动归批）
curl http://localhost:8080/api/batches
curl -X POST http://localhost:8080/api/batches \
  -H 'Content-Type: application/json' \
  -d '{"auto":true,"period_type":"daily"}'
curl 'http://localhost:8080/api/pipeline/stats?batch_id=1'

# 调整进度状态（screening → to_interview → interviewing → passed/completed/rejected）
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

# 拖入等价：上传 PDF/ZIP（默认归入当日批次；可指定 period_type 或 batch_id）
curl -F 'files=@resume.pdf' -F 'source=cursor' -F 'period_type=daily' http://localhost:8080/api/upload
curl -F 'files=@resume.pdf' -F 'batch_id=2' http://localhost:8080/api/upload

# 批量评估本地目录（Cursor 常用）
curl -X POST http://localhost:8080/api/screen \
  -H 'Content-Type: application/json' \
  -d '{"path":"/Users/caden/Downloads/resumes","source":"cursor"}'

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

- 后端 `backend/internal/scanner/`：关键词启发式（实习 4 维 + 工程 4 维 + Agent 3 维 + 深度指标）
- 简历无实习关键词 → 标 `no_intern`，自动降权；有实习+工程+Agent 双强 → 优先
- `seed/questions.go`：S 档 11 人各 3 条定制面试题
- `seed/questions.go` ManualTier：人工档位覆盖
- 图片 PDF：尝试 `python3 backend/scripts/ocr_resume.py`（需 pymupdf）

## 典型 Cursor 工作流

1. 用户丢 zip 到 `~/Downloads/` → Agent 调 `/api/screen` 批量评估
2. Agent `GET /api/candidates?tier=S` 拉 S 档列表
3. Agent `GET /api/candidates/:id` 读面试题，辅助面试
4. `GET /api/export` 同步到飞书文档

## 前端组件（Vue + Element Plus）

- `CandidateTable` — 列表，工程分 E / Agent 分 A
- `ResumeViewer` — iframe 简历
- `QuestionPanel` — 定制面试题
- `UploadDropzone` — 拖拽 PDF/ZIP
- `DocsView` — 开发知识库（`docs/*.md`，API `/api/docs`）

## 知识库

- 网页顶部 **知识库** 标签，或 `http://localhost:8080/?tab=docs`
- 源码：`docs/` 目录（快速开始、部署、API、扩展开发）
- 更新文档后 `make deploy` 即可，无需改前端

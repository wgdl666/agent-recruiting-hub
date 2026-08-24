# Agent 招聘评估台

Go + Vue 本地招聘筛选网站：看简历、看面试题、拖拽评估，REST API 供 Cursor 操作。

## 功能

- **候选人列表**：按档位筛选，工程分优先排序（有落地经验排前）
- **简历预览**：浏览器内嵌 PDF
- **定制面试题**：S 档候选人每人 3 条深挖题
- **上传评估**：拖拽 PDF 或 ZIP，自动解压+评分+入库
- **Cursor 集成**：`AGENTS.md` 中的 REST API

## 快速开始

```bash
make install
make backend   # 终端 1
make frontend  # 终端 2 → http://localhost:5173
```

首次打开点「导入种子数据」，或：

```bash
curl -X POST http://localhost:8080/api/import/seed
```

生产构建（单端口 8080）：

```bash
make run
```

## 技术栈

- Backend: Go, Gin, SQLite
- Frontend: Vue 3, Element Plus, Vite

# Agent 招聘评估台

面向 **Agent 工程实习** 招聘的本地评估台：简历筛选、批次管理、面试进度看板、定制面试题。

## 适用角色

| 角色 | 常用能力 |
|------|----------|
| HR / 面试官 | 网页上传简历、看板拖拽改状态、查看 S 档面试题与参考答案 |
| 开发 / Cursor | REST API、`scripts/hub` CLI、批量 `/api/screen` |
| 运维 | `make deploy` 部署到 mirror_zby |

## 招聘标准（内置）

1. **实习经历** — 有真实实习/上线/独立交付信号优先  
2. **传统工程底座** — 后端、存储、并发、幂等、可观测  
3. **单项目深挖** — Agent / RAG / LangGraph / Tool 闭环  

自动评分为启发式关键词，**图片 PDF 可能不准**，以人工档位和简历为准。

## 仓库

- GitHub（私有）：`https://github.com/Zhan-boyi/agent-recruiting-hub`
- 远端实例：`mirror_zby`（112.74.38.5:8808，建议 SSH 隧道访问）

## 文档导航

左侧目录可切换：快速开始、架构、API、部署、扩展开发。

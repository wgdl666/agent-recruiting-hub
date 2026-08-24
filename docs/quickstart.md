# 快速开始

## 环境要求

- Go 1.22+
- Node.js 18+
- （可选）`lark-cli` — 飞书文档导出
- （可选）`pymupdf` — 图片 PDF OCR：`pip install pymupdf`

## 首次安装

```bash
git clone https://github.com/wgdl666/agent-recruiting-hub.git
cd agent-recruiting-hub
make install
```

## 开发模式（推荐）

两个终端：

```bash
# 终端 1 — API :8080
make backend

# 终端 2 — 前端 :5173（代理到 8080）
make frontend
```

浏览器打开 http://localhost:5173

## 生产单端口

```bash
make run   # 构建前端 + 启动 http://localhost:8080
```

## 初始化数据

首次空库可导入 47 人种子数据：

```bash
curl -X POST http://localhost:8080/api/import/seed
# 或网页：数据与导出 → 导入种子数据
```

## CLI 快捷命令

```bash
export HUB_URL=http://localhost:8080/api   # 远端时改成对应地址

./scripts/hub health
./scripts/hub list S
./scripts/hub get 1
./scripts/hub screen ~/Downloads/resumes.zip
./scripts/hub open
```

## 日常招聘流程

1. 顶部选择 **今日/本周/本月批次**
2. 上传 PDF 或 ZIP → 自动评分入库，状态为「简历筛选」
3. **列表** 或 **看板** 推进状态
4. S 档点击候选人 → 查看 **定制面试题 + 参考答案**

## 常见问题

**Q：自动分显示「待人工」？**  
A：多为图片 PDF，文本提取失败。看档位、摘要和 PDF 原文，可手动调档。

**Q：如何更新面试题？**  
A：改 `backend/internal/seed/qa.go` 后执行 `POST /api/sync/questions` 或网页「同步面试题与参考答案」。

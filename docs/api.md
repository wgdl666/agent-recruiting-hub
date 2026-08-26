# REST API 参考

Base URL：`http://localhost:8080/api`（远端替换主机与端口）

## 健康与统计

```bash
GET /health
GET /stats
GET /pipeline/stats?batch_id=1
```

## 候选人

```bash
# 列表（eng_first 工程优先；batch_id 筛批次；position_id 筛岗位）
GET /candidates?tier=S&status=screening&batch_id=1&position_id=1&eng_first=true&q=张三

GET /candidates/:id
GET /candidates/:id/resume          # PDF inline

PATCH /candidates/:id
  {"tier":"A"}
  {"status":"to_interview"}
  {"interview_order":3}
  {"clear_manual":true}               # 恢复自动档位

DELETE /candidates/:id
```

## 批次

```bash
GET /batches?active=true

POST /batches
  {"auto":true,"period_type":"daily"}   # 今日批次
  {"name":"校招专场","tag":"campus-2026","period_type":"custom"}
```

## 岗位与检验标准

```bash
GET /positions                         # 岗位阶梯；?open=true 仅在招
POST /positions
  {"name":"实习生","skill_id":"intern","description":"...","is_open":true}
PATCH /positions/:id
  {"name":"...","skill_id":"intern","is_open":false}

GET /skills                            # 检验标准（招聘 Skill）目录；岗位绑定后上传自动选用
```

历史候选人在迁移时岗位一律写成「实习生」。

## 上传与评估

```bash
# 表单上传（网页传 position_id；CLI 仍可只传 skill_id，会映射到对应在招岗）
POST /upload
  files=@resume.pdf
  position_id=1                        # 推荐：招聘岗位
  skill_id=intern                      # 兼容旧客户端
  source=cursor
  period_type=daily|weekly|monthly
  batch_id=2                            # 可选，显式指定批次

# 批量评估本地路径（Cursor 常用）
POST /screen
  {"path":"/path/to/dir","source":"cursor","position_id":1,"skill_id":"intern","period_type":"daily","batch_id":0}
```

## 维护

```bash
POST /import/seed       # 导入 seed/batch.json + 关联 PDF
POST /rescreen          # 启发式重评全部（不走 ModelHub；不覆盖 tier_manual）
POST /sync/questions    # S 档按简历生成面试题（失败回落 seed/qa.go）
POST /candidates/:id/questions/generate
```

## 导出

```bash
GET /export                    # JSON 全量
GET /export/markdown           # Markdown 报告
POST /export/feishu            # 飞书文档（需 lark-cli）
```

## 知识库

```bash
GET /docs                      # 文档目录
GET /docs/quickstart           # 单篇 Markdown 正文
```

## 状态枚举

| 值 | 含义 |
|----|------|
| `screening` | 简历筛选 |
| `read` | 已阅 |
| `to_interview` | 待约面 |
| `interviewing` | 面试中 |
| `passed` | 通过 |
| `completed` | 面试完成 |
| `rejected` | 本轮淘汰 |

## Cursor 典型链路

```bash
curl -X POST http://localhost:8080/api/screen \
  -H 'Content-Type: application/json' \
  -d '{"path":"/Users/you/Downloads/resumes.zip","source":"cursor"}'

curl 'http://localhost:8080/api/candidates?tier=S&eng_first=true'
curl http://localhost:8080/api/candidates/12
```

完整说明见仓库根目录 `AGENTS.md`（给 Cursor Agent 用）。

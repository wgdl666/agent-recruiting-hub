# Git 分支与发布

## 分支约定

| 分支 | 用途 |
|------|------|
| **`dev`** | 日常开发。后续功能、评分规则、前端改动均在此分支提交 |
| **`main`** | 生产基线。与 https://recruit.wgdl.tech 当前部署对齐；发布前从 `dev` 合并 |

## 日常开发

```bash
git checkout dev
git pull origin dev

# 开发、自测
make backend    # 终端 1
make frontend   # 终端 2

git add ...
git commit -m "简述 why"
git push origin dev
```

## 发布到生产

确认 `dev` 可发布后：

```bash
git checkout main
git pull origin main
git merge dev
git push origin main
make deploy     # → mirror_zby / recruit.wgdl.tech

git checkout dev   # 回到开发分支
```

## 首次克隆

```bash
git clone https://github.com/wgdl666/agent-recruiting-hub.git
cd agent-recruiting-hub
git checkout dev
make install
```

## 说明

- **不要**在 `main` 上直接做功能开发（热修复除外）
- `make deploy` 部署的是**当前检出分支**的代码，发布前请确认在 `main` 且已合并 `dev`
- 远端生产环境配置（`.env`、ModelHub 端口转发等）在服务器上，不随 Git 分支切换；见 [部署指南](deploy.md)

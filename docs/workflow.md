# Git 分支与发布

日常开发和生产发布都在 **`main`**。https://recruit.wgdl.tech 只从这个分支部署。

| 分支 | 用途 |
|------|------|
| **`main`** | 日常开发与生产发布。与 https://recruit.wgdl.tech 当前部署对齐 |

历史分支 `dev` 不再作为开发入口；已有改动合并进 `main` 后，后续直接在 `main` 提交。

## 日常开发

```bash
git checkout main
git pull origin main

# 开发、自测
make backend    # 终端 1
make frontend   # 终端 2

git add ...
git commit -m "简述 why"
git push origin main
```

## 发布到生产

确认 `main` 可发布后：

```bash
git checkout main
git pull origin main
make deploy     # → mirror_zby / recruit.wgdl.tech
```

## 首次克隆

```bash
git clone https://github.com/wgdl666/agent-recruiting-hub.git
cd agent-recruiting-hub
make install
```

默认即 `main`，不必再切 `dev`。

## 说明

- 功能开发直接在 `main` 小步提交
- `make deploy` 部署的是**当前检出分支**的代码，发布前请确认在 `main`
- 远端生产环境配置（`.env`、ModelHub 端口转发等）在服务器上，不随 Git 分支切换；见 [部署指南](deploy.md)

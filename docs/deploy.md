# 部署指南（mirror_zby）

默认部署目标：**mirror_zby**（与 `mirror` 同机，112.74.38.5）。

## 前置条件

本机已配置 SSH：

```sshconfig
Host mirror_zby
    HostName 112.74.38.5
    User root
    IdentityFile ~/.ssh/id_ed25519
```

## 一键部署

```bash
git pull
make deploy
```

脚本会：

1. 构建 Linux 二进制 + 前端静态资源  
2. rsync 到 `/opt/agent-recruiting-hub`  
3. 安装/重启 systemd 服务 `agent-recruiting-hub`（端口 **8808**）  
4. 同步 `data/`、`seed/`、`docs/`

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `DEPLOY_HOST` | `mirror_zby` | SSH 主机别名 |
| `DEPLOY_DIR` | `/opt/agent-recruiting-hub` | 远端目录 |
| `DEPLOY_ADDR` | `:8808` | 监听地址 |
| `DEPLOY_SERVICE` | `agent-recruiting-hub` | systemd 单元名 |

示例：换端口部署

```bash
DEPLOY_ADDR=:8810 make deploy
```

## 访问远端

公网 **8808** 可能被安全组拦截，推荐 SSH 隧道：

```bash
ssh -L 8808:127.0.0.1:8808 mirror_zby
# 浏览器 http://localhost:8808
```

## 远端运维

```bash
ssh mirror_zby

systemctl status agent-recruiting-hub
journalctl -u agent-recruiting-hub -f

curl -s http://127.0.0.1:8808/api/health
curl -s -X POST http://127.0.0.1:8808/api/sync/questions
```

## 数据持久化

| 路径 | 说明 |
|------|------|
| `/opt/agent-recruiting-hub/data/hub.db` | SQLite 主库 |
| `/opt/agent-recruiting-hub/data/resumes/` | 简历 PDF |

**注意**：`make deploy` 会 rsync 本地 `data/`，部署前确认不会覆盖远端新数据。生产环境建议先备份 `hub.db`。

## 回滚

```bash
# 检出旧版本后重新 deploy
git checkout <commit>
make deploy
```

或仅替换二进制：

```bash
GOOS=linux GOARCH=amd64 go build -o bin/hub-linux ./backend
rsync -az bin/hub-linux mirror_zby:/opt/agent-recruiting-hub/bin/hub
ssh mirror_zby systemctl restart agent-recruiting-hub
```

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HOST="${DEPLOY_HOST:-mirror_zby}"
REMOTE_DIR="${DEPLOY_DIR:-/opt/agent-recruiting-hub}"
ADDR="${DEPLOY_ADDR:-127.0.0.1:8808}"
DOMAIN="${DEPLOY_DOMAIN:-recruit.wgdl.tech}"
SERVICE="${DEPLOY_SERVICE:-agent-recruiting-hub}"

echo "==> build frontend + linux binary"
cd "$ROOT/frontend" && npm run build
cd "$ROOT/backend"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$ROOT/bin/hub-linux" .

echo "==> sync to $HOST:$REMOTE_DIR"
ssh "$HOST" "mkdir -p '$REMOTE_DIR'/bin '$REMOTE_DIR'/frontend/dist '$REMOTE_DIR'/data/resumes '$REMOTE_DIR'/seed"
rsync -az "$ROOT/bin/hub-linux" "$HOST:$REMOTE_DIR/bin/hub"
rsync -az --delete "$ROOT/frontend/dist/" "$HOST:$REMOTE_DIR/frontend/dist/"
rsync -az "$ROOT/seed/" "$HOST:$REMOTE_DIR/seed/"
rsync -az "$ROOT/docs/" "$HOST:$REMOTE_DIR/docs/"
rsync -az "$ROOT/.env.example" "$HOST:$REMOTE_DIR/.env.example"
rsync -az "$ROOT/data/" "$HOST:$REMOTE_DIR/data/"
ssh "$HOST" "chmod +x '$REMOTE_DIR/bin/hub'"

echo "==> install systemd unit"
ssh "$HOST" "cat > /etc/systemd/system/${SERVICE}.service" <<EOF
[Unit]
Description=Agent Recruiting Hub
After=network.target

[Service]
Type=simple
WorkingDirectory=${REMOTE_DIR}
EnvironmentFile=-${REMOTE_DIR}/.env
ExecStart=${REMOTE_DIR}/bin/hub -root ${REMOTE_DIR} -addr ${ADDR}
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

ssh "$HOST" "systemctl daemon-reload && systemctl enable ${SERVICE} && systemctl restart ${SERVICE}"
sleep 1
PORT="${ADDR##*:}"
ssh "$HOST" "systemctl is-active ${SERVICE} && curl -sf http://127.0.0.1:${PORT}/api/health"

echo "==> configure nginx + TLS for ${DOMAIN}"
rsync -az "$ROOT/deploy/nginx/${DOMAIN}.conf" "$HOST:/etc/nginx/sites-available/${DOMAIN}"
ssh "$HOST" "ln -sf /etc/nginx/sites-available/${DOMAIN} /etc/nginx/sites-enabled/${DOMAIN} && nginx -t && systemctl reload nginx"
if ! ssh "$HOST" "test -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"; then
  ssh "$HOST" "certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos --register-unsafely-without-email --redirect"
else
  # 证书已存在但 nginx 可能未挂 SSL（仅 HTTP 时 HTTPS 会落到默认站点）
  ssh "$HOST" "certbot install --cert-name ${DOMAIN} --nginx 2>/dev/null || true"
  ssh "$HOST" "certbot renew --quiet 2>/dev/null || true"
fi
ssh "$HOST" "nginx -t && systemctl reload nginx"

echo ""
echo "Deployed: https://${DOMAIN}"
echo "Direct (internal): http://127.0.0.1:${PORT}"

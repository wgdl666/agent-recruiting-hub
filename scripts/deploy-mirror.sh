#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HOST="${DEPLOY_HOST:-mirror}"
REMOTE_DIR="${DEPLOY_DIR:-/opt/agent-recruiting-hub}"
ADDR="${DEPLOY_ADDR:-:8808}"
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
ExecStart=${REMOTE_DIR}/bin/hub -root ${REMOTE_DIR} -addr ${ADDR}
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

ssh "$HOST" "systemctl daemon-reload && systemctl enable ${SERVICE} && systemctl restart ${SERVICE}"
sleep 1
PORT="${ADDR#:}"
ssh "$HOST" "systemctl is-active ${SERVICE} && curl -sf http://127.0.0.1:${PORT}/api/health"

IP="$(ssh "$HOST" 'curl -sf --max-time 3 ifconfig.me 2>/dev/null || hostname -I | awk "{print \$1}"')"
echo ""
echo "Deployed: http://${IP}:${PORT}"

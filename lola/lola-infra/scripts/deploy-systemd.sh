#!/usr/bin/env bash
# Installs lola-core as a systemd service on a Linux host that isn't
# using Docker — an alternative deployment path for users who'd rather
# run the binary directly. Optional; the primary supported path is
# docker-compose.yml in this directory.
#
# Usage: sudo ./scripts/deploy-systemd.sh /path/to/built/lola/binary
set -euo pipefail

BINARY_PATH="${1:?Usage: $0 /path/to/lola/binary}"
INSTALL_PATH="/usr/local/bin/lola"
SERVICE_USER="lola"
SERVICE_FILE="/etc/systemd/system/lola-core.service"

if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root (it writes to /etc/systemd and /usr/local/bin)." >&2
  exit 1
fi

if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
  echo "==> Creating service user '$SERVICE_USER'"
  useradd --system --create-home --shell /usr/sbin/nologin "$SERVICE_USER"
fi

echo "==> Installing binary to $INSTALL_PATH"
cp "$BINARY_PATH" "$INSTALL_PATH"
chmod 755 "$INSTALL_PATH"

echo "==> Writing systemd unit"
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=LOLA OS engine
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$SERVICE_USER
ExecStart=$INSTALL_PATH serve --tcp 127.0.0.1:8899
Restart=on-failure
RestartSec=5
# LOLA_VAULT_PASSPHRASE must be supplied via an EnvironmentFile so it
# never appears in 'systemctl status' output or shell history.
EnvironmentFile=/etc/lola-core/lola-core.env

[Install]
WantedBy=multi-user.target
EOF

mkdir -p /etc/lola-core
if [ ! -f /etc/lola-core/lola-core.env ]; then
  echo "LOLA_VAULT_PASSPHRASE=changeme" > /etc/lola-core/lola-core.env
  chmod 600 /etc/lola-core/lola-core.env
  chown "$SERVICE_USER":"$SERVICE_USER" /etc/lola-core/lola-core.env
  echo "==> Wrote placeholder /etc/lola-core/lola-core.env — edit this before starting the service."
fi

systemctl daemon-reload
echo "==> Done. Review /etc/lola-core/lola-core.env, then run:"
echo "    sudo systemctl enable --now lola-core"

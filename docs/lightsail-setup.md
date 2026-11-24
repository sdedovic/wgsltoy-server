# Manual Setup / Runbook

Commands run on your machine unless marked "on the box". The box is reached as `wgsltoy-staging`, an SSH alias for
`admin@staging.wgsltoy.com`.

## Local SSH setup
Uses your own SSH key, added to the instance when creating it.
```bash
cat >> ~/.ssh/config <<'EOF'
Host wgsltoy-staging
    HostName staging.wgsltoy.com
    User admin
EOF
```

The alias resolves through DNS, so it works once the record from step 0 is in place. Check:
```bash
ssh wgsltoy-staging 'sudo sshd -T | grep -E "passwordauthentication|permitrootlogin"'
```
Both should be `no`, which is the Lightsail default.

## 0. In the Lightsail console
- Attach a static IP to the instance.
- Networking → Firewall: allow 22, 80 and 443 only.
- DNS: point an A record for `staging.wgsltoy.com` at the static IP. Do this now, because Caddy needs the record in
  place before it can get a certificate. Check with `dig +short staging.wgsltoy.com`.

## 1. Base system (on the box)
```bash
sudo apt update && sudo apt full-upgrade -y
sudo apt install -y unattended-upgrades curl gpg rsync
sudo dpkg-reconfigure -plow unattended-upgrades

# 1 GB swap, which a small box needs with Postgres running
sudo fallocate -l 1G /swapfile && sudo chmod 600 /swapfile
sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

## 2. Postgres (on the box)
```bash
sudo apt install -y postgresql
PW=$(openssl rand -hex 24)
sudo -u postgres psql -c "CREATE USER wgsltoy PASSWORD '$PW'"
sudo -u postgres createdb -O wgsltoy wgsltoy
echo "$PW"    # save it; needed for DATABASE_URL and migrations
```
The password goes into a URL, so keep it to URL-safe characters (hex is), or percent-encode it. Debian's default
configuration already listens on localhost only.

## 3. Caddy (on the box, from Caddy's official repo)
```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
  | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
  | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install -y caddy
```

## 4. Anubis (on the box)
Download the amd64 .deb from https://github.com/TecharoHQ/anubis/releases into `/tmp` (apt can't read it from your home
directory), then:
```bash
sudo apt install -y /tmp/anubis_*_amd64.deb

sudo mkdir -p /etc/anubis
sudo tee /etc/anubis/wgsltoy.env <<'EOF'
BIND=127.0.0.1:8923
METRICS_BIND=127.0.0.1:9090
TARGET=http://127.0.0.1:8080
EOF

sudo systemctl enable --now anubis@wgsltoy
```
This uses Anubis's built-in default policy. You can add a custom one later.

## 5. App directory and secrets (on the box)
```bash
sudo mkdir -p /opt/wgsltoy

sudo tee /etc/wgsltoy.env <<EOF
DATABASE_URL=postgres://wgsltoy:<password>@127.0.0.1:5432/wgsltoy
APP_SECRET=$(openssl rand -hex 32)
INVITE_CODE=<pick one, or leave empty for open signup>
TEMPLATES_DIR=/opt/wgsltoy/web/templates
STATIC_DIR=/opt/wgsltoy/web/dist
EOF

sudo chmod 600 /etc/wgsltoy.env
```

## 6. Build locally and copy up
From the repo:
```bash
just web-build
cd server && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../bin/wgsltoy-server ./cmd/wgsltoy-server && cd ..
rsync -av bin/wgsltoy-server server/application.yaml wgsltoy-staging:/tmp/wgsltoy/
rsync -av web/templates web/dist wgsltoy-staging:/tmp/wgsltoy/web/
ssh wgsltoy-staging 'sudo rsync -a --delete /tmp/wgsltoy/ /opt/wgsltoy/'
```

## 7. Run migrations (over an SSH tunnel)
```bash
# -f backgrounds ssh only once the tunnel is up
ssh -fN -o ExitOnForwardFailure=yes -L 5433:127.0.0.1:5432 wgsltoy-staging
cd server && migrate -path migrations -database 'postgres://wgsltoy:<password>@localhost:5433/wgsltoy?sslmode=disable' up
pkill -f 'L 5433:127.0.0.1:5432'
```

## 8. systemd unit for the app (on the box)
```bash
sudo tee /etc/systemd/system/wgsltoy.service <<'EOF'
[Unit]
Description=wgsltoy
After=network.target postgresql.service
Requires=postgresql.service

[Service]
WorkingDirectory=/opt/wgsltoy
ExecStart=/opt/wgsltoy/wgsltoy-server
EnvironmentFile=/etc/wgsltoy.env
DynamicUser=yes
Restart=always
RestartSec=2
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload && sudo systemctl enable --now wgsltoy
curl -s localhost:8080/health
```

## 9. Caddy configuration (on the box)
```bash
sudo tee /etc/caddy/Caddyfile <<'EOF'
staging.wgsltoy.com {
    encode zstd gzip
    reverse_proxy 127.0.0.1:8923 {
        header_up X-Real-Ip {remote_host}
    }
}
EOF

sudo systemctl reload caddy
```

## 10. Check it works
- Open https://staging.wgsltoy.com. You should see the Anubis challenge page, then the site.
- If anything fails, read the logs on the box with `journalctl -u wgsltoy -u anubis@wgsltoy -u caddy -f`.

## Later deploys
Repeat steps 6 and 7 (7 only when there are new migrations), then:
```bash
ssh wgsltoy-staging 'sudo systemctl restart wgsltoy'
```

## Still to do
- Nightly `pg_dump` to S3.
- Lightsail automatic snapshots.
- One test restore.

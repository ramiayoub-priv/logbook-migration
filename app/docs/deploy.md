# Deployment — `ayoub.fi/logbook`

**The box is shared.** Other sites and services live here and must not be disturbed (rule §0.3).
Read the tenant map before changing anything.

## Server map (surveyed 2026-08-01)

DigitalOcean droplet, **1 vCPU / 2 GB RAM / 49 GB disk (25 GB free)**, Ubuntu, kernel 5.4, uptime 844
days. **No swap** — memory pressure means the OOM killer, not slowdown.

| Port | What | Notes |
|---|---|---|
| 22 | sshd | Under constant attack: fail2ban has logged 412,570 failures / 50,264 bans. **Never risk this port.** |
| 80, 443 | Apache 2 | Terminates TLS for `ayoub.fi` (Let's Encrypt). `DocumentRoot /var/www/ayoub.fi`. |
| 8000 | Docker `kraken-predictor-python-2` | gunicorn, published `0.0.0.0:8000`, up 2 years. **Publicly exposed**, ~759 MB RSS — the box's largest consumer. Fate undecided. |
| 5432 | *(was OpenVPN)* | **Stopped + disabled 2026-08-01** at the user's request. Not Postgres — see the correction in `APP.md`. Re-enable: `sudo systemctl enable --now openvpn-server@server`. |
| 500/4500 udp | strongSwan IPsec | Left alone. |
| 9002 | **our backend** | To be added. Binds `127.0.0.1` only. |

`/var/www/ayoub.fi` is **owned by `rami`**, so static deploys need no sudo. It also holds `blog/`,
`countdown/`, `englishhouse/`, `games/`, `pdp/`, `simpleclock/`, and a now-stale `transit/`.

**Removed 2026-08-01**: transit's orphaned Quarkus (PPID 1, no systemd unit, no cron — launched by
hand from `/home/rami/transit-backend` on 2026-05-23). It held 605 MB. Transit runs on its own VM now.
The jar is untouched on disk; restart recipe is `/home/rami/transit-backend/backend-remote-start.sh`.

**`/api/` is still proxied to the dead `127.0.0.1:9001`** in the vhost. Our API therefore lives at
`/logbook/api/`, which avoids the collision entirely. Removing the stale lines is optional cleanup.

### ⚠ Docker bypasses ufw
Docker publishes ports by writing its own iptables `DOCKER` chain, evaluated **before** ufw's INPUT
rules. `ufw deny <port>` does **not** close a published container port. If `:8000` ever needs closing,
change the container's port binding to `127.0.0.1:8000`; do not expect a firewall rule to do it.

## Our footprint

| Path | Purpose |
|---|---|
| `/opt/logbook/logbook-server` | The static binary. |
| `/var/lib/logbook/logbook.db` | SQLite, mode `0600`, owned by the service user, **outside the web root**. |
| `/var/lib/logbook/backups/` | Timestamped backups. |
| `/var/www/logbook/` | Built frontend assets (static). |
| `/etc/systemd/system/logbook.service` | Unit, running as a dedicated unprivileged user. |

Nothing we install writes anywhere else.

## Build

Cross-compiled from the dev machine — **no toolchain is needed on the server** (its Go is 1.13, far
too old, and irrelevant).

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/logbook-server ./cmd/server
```

`CGO_ENABLED=0` is what makes this work: `modernc.org/sqlite` is pure Go, so the result is a static
binary with no libc dependency and no build step on the target.

Frontend: `npm run build` → static files → rsync to `/var/www/logbook/`.

## Apache

Added **additively** to the existing `ayoub.fi-le-ssl.conf` vhost. It must not touch the existing
`/api/` lines or the other sites' paths.

```apache
# --- logbook ---
ProxyPass        /logbook/api/  http://127.0.0.1:9002/api/
ProxyPassReverse /logbook/api/  http://127.0.0.1:9002/api/

Alias /logbook /var/www/logbook
<Directory /var/www/logbook>
    Require all granted
    Options -Indexes
    # SPA fallback: serve index.html for client-side routes, but never shadow /logbook/api/
    RewriteEngine On
    RewriteCond %{REQUEST_FILENAME} !-f
    RewriteCond %{REQUEST_FILENAME} !-d
    RewriteRule . /logbook/index.html [L]
</Directory>
```

`ProxyPass` is declared before `Alias` so the API path wins over the static alias.

**Always validate before reloading**: `apache2ctl configtest`, then `systemctl reload apache2` (reload,
not restart — it does not drop the other sites' connections). If configtest fails, fix it before
touching the running server.

## systemd

```ini
[Unit]
Description=Logbook
After=network.target

[Service]
User=logbook
Group=logbook
ExecStart=/opt/logbook/logbook-server
Restart=always
RestartSec=5
Environment=LOGBOOK_ADDR=127.0.0.1:9002
Environment=LOGBOOK_DB=/var/lib/logbook/logbook.db

# hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
PrivateDevices=true
ReadWritePaths=/var/lib/logbook
MemoryMax=192M

[Install]
WantedBy=multi-user.target
```

`MemoryMax` is a containment control (see `security.md`): even a runaway bug in this app cannot OOM
the other sites. 192 MB is generous — expected steady state is ~25 MB, with a transient spike during
PDF generation of the full 1295-flight logbook. **Measure the real peak and record it here.**

## Rollback

The previous binary is kept as `/opt/logbook/logbook-server.prev`. Rollback is a copy plus
`systemctl restart logbook` — no rebuild. The database is backed up before any migration, so a schema
change is reversible too.

## Verification after every deploy

1. `systemctl status logbook` — active, not restart-looping.
2. `curl -sS -o /dev/null -w '%{http_code}' https://ayoub.fi/logbook/api/health` → `200`.
3. `curl` an authenticated endpoint **without** a session → `401` (proves default-deny survived).
4. `https://ayoub.fi/` and one other sub-path still load — **prove the other sites are unharmed**.
5. `free -m` — confirm headroom did not collapse.

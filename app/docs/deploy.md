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

Frontend:
```bash
cd app/frontend
npm ci            # Node is a BUILD-time dependency only; it never runs on the server
npm run check     # tsc --noEmit + vitest -- the frontend's equivalent of `make check`
npm run build     # -> app/frontend/dist/  (~169 KB JS, 53 KB gzipped)
rsync -a --delete app/frontend/dist/ rami@ayoub.fi:/var/www/logbook/
```
`vite.config.ts` sets `base: '/logbook/'`, so every asset URL is built with that prefix. Building
without it produces a page that 404s every asset behind the Apache `Alias`.

The build also emits the PWA files — `manifest.webmanifest`, `sw.js` and `icons/` — which must land
at the web root of `/var/www/logbook/` for the home-screen install to work. Two Apache notes:

- **`sw.js` must not be served from a long-lived cache.** A stale service worker outlives a deploy
  and keeps serving the previous bundle. Serve it `Cache-Control: no-cache` so the browser
  revalidates it every time; the hashed files under `assets/` are immutable and can be cached hard.
- **`index.html` must not be cached either, and for the same reason.** It is the only file under
  `/logbook/` whose name stays the same while its bytes change on every deploy — it is what points
  at the hashed bundles. A phone holding a stale copy runs the previous frontend against the current
  API, which is exactly the mismatch that "ship the binary and the frontend together" exists to
  prevent. Reported from the field on 2026-08-01: the owner's phone would not pick up a new build.
- The worker's scope is `/logbook/`, so it can only ever intercept our own paths — never the
  owner's other sites on this box.

```apache
<Files "sw.js">
    Header set Cache-Control "no-cache"
</Files>
<Files "index.html">
    Header set Cache-Control "no-cache, no-store, must-revalidate"
    Header set Pragma "no-cache"
    Header set Expires "0"
</Files>
<Directory /var/www/logbook/assets>
    Header set Cache-Control "public, max-age=31536000, immutable"
</Directory>
```

**Three layers, deliberately, because each one covers a device the others do not:**

| Layer | Where | What it fixes |
|---|---|---|
| `Cache-Control` / `Pragma` / `Expires` on `index.html` | Apache, and as `<meta http-equiv>` in the document itself | The browser's HTTP cache. The meta tags travel with the file, so a device served by anything other than this vhost is still covered. |
| `fetch(request, {cache: 'no-store'})` on the shell | `public/sw.js` | The service worker's own "network first" was only ever as fresh as the HTTP cache underneath it. |
| `reloadWhenUpdated` | `src/swupdate.ts`, wired in `src/main.tsx` | A home-screen install has no address bar and no reload button. When a new worker claims the page, the page reloads itself onto it — once, guarded against a loop. |

Verify a deploy actually reached the device:
```bash
curl -sI https://ayoub.fi/logbook/ | grep -i cache-control     # expect no-store
curl -sI https://ayoub.fi/logbook/sw.js | grep -i cache-control # expect no-cache
```

## Apache

Added **additively** to the existing `ayoub.fi-le-ssl.conf` vhost. It must not touch the existing
`/api/` lines or the other sites' paths.

```apache
# --- logbook ---
ProxyPass        /logbook/api/  http://127.0.0.1:9002/logbook/api/
ProxyPassReverse /logbook/api/  http://127.0.0.1:9002/logbook/api/

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

⚠ **The proxy does NOT rewrite the path.** The server mounts its routes at the *full public path* —
`basePath = "/logbook/api"` in `cmd/server/server.go` — so the backend answers `/logbook/api/health`
and **404s `/api/health`**. This file said `.../9002/api/` until 2026-08-01 and it was wrong; the
health check in the install script caught it before Apache was ever reloaded. Both halves of every
`ProxyPass` line carry the same prefix.

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
# Printed on the exported PDFs. Without it the documents are unnamed, which is
# valid but not what an authority wants to receive.
Environment=LOGBOOK_HOLDER=Rami Ayoub

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

## ⚠ `-origin` must match what the browser actually sends

The CSRF check compares the `Origin` header byte for byte, so the value has to be the exact
scheme+host the page was loaded from — `https://ayoub.fi` in production, and `http://localhost:5173`
against the Vite dev server. A mismatch fails every mutating request with a 403 and nothing else,
which reads like a broken login rather than a working control. `-insecure-cookie` must never appear
in the production unit; the server already refuses to combine it with an `https://` origin.

## Verification after every deploy

1. `systemctl status logbook` — active, not restart-looping.
2. `curl -sS -o /dev/null -w '%{http_code}' https://ayoub.fi/logbook/api/health` → `200`.
3. `curl` an authenticated endpoint **without** a session → `401` (proves default-deny survived).
4. `https://ayoub.fi/` and one other sub-path still load — **prove the other sites are unharmed**.
5. `free -m` — confirm headroom did not collapse.

---

## The off-box backup (Task 14, 2026-08-02)

**Why it exists.** Until 2026-08-02 production was reconstructible from this repository: three
committed CSVs and one command. Then the data closed, the app became the only way the record grows,
and the owner logged flights from an airfield. **Those rows are in no CSV.** The pre-import backups
under `/var/lib/logbook/backups/` sit on the same disk as the database they protect — they defend
against a bad import and against nothing else. This puts a copy on another machine.

**What runs.** A systemd timer at **03:17 UTC daily**, `Persistent=true` so a missed day runs at the
next boot rather than leaving a hole exactly where the trouble was.

```
logbook-backup.timer   -> logbook-backup.service   -> /opt/logbook/backup.sh
                          User=logbook, not root      logbookctl backup -> git commit -> git push
```

**The service never stops.** `logbookctl backup` snapshots with `VACUUM INTO`, which is
transactionally consistent against a live database in WAL mode. A backup that needed downtime would
be a backup that gets skipped.

**What lands in the repository** — exactly four files, replaced every run:

| file | what it is |
|---|---|
| `logbook.db` | the whole database, sessions stripped. **This is the restore.** |
| `logbook.csv` | every flight, every stored field, in book order. Deliberate redundancy. |
| `MANIFEST.txt` | counts and checksums to verify a restore against. |
| `RESTORE.md` | how to put it back — travelling **with** the data. |

`RESTORE.md` is in the backup rather than only here on purpose: instructions that live in the
application repository are instructions you do not have on the day the server is gone.

**It commits every day, even when nothing was flown.** The database and CSV are byte-identical when
the data has not changed, so git stores one blob and the only new bytes are the manifest's timestamp
— a few hundred a day. What that buys is a **heartbeat**: a commit dated yesterday proves the backup
ran, so "nothing changed" and "this has been failing silently for a month" stop looking the same.

### Setting it up (owner, once)

Two steps cannot be scripted because both involve a secret (rule §0.3):

```bash
# 1. Generate the deploy key ON THE SERVER, so the private half never travels.
sudo -u logbook install -d -m 0700 -o logbook -g logbook /var/lib/logbook/.ssh
sudo -u logbook ssh-keygen -t ed25519 -N '' -C 'logbook backup (ayoub.fi)' \
     -f /var/lib/logbook/.ssh/backup_ed25519
sudo cat /var/lib/logbook/.ssh/backup_ed25519.pub

# 2. Create https://github.com/ramiayoub-priv/logbook-backup as a PRIVATE repository
#    with NO README and NO licence, and add that public key under
#    Settings -> Deploy keys with "Allow write access" TICKED.
```

Then, from the repo:

```bash
cd app/backend && export PATH=$HOME/.local/go/bin:$PATH
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /tmp/logbookctl ./cmd/logbookctl
cd ../.. && rsync -a /tmp/logbookctl app/deploy/ rami@ayoub.fi:/home/rami/logbook-deploy/
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/install-backup.sh'
```

`install-backup.sh` is idempotent. It pins GitHub's host keys, initialises the repository, installs
the unit and timer, **runs the backup once with a human watching**, and then does the step that
matters most.

### ⚠ The trap this cost, and the check that exists because of it

Rehearsing the script found a failure that reports success forever and is discovered only on the day
the backup is needed: **the push succeeded and a fresh clone came back EMPTY.** The remote's `HEAD`
named a branch we had never pushed to (`master` vs `main`), so `git clone` warned *"remote HEAD
refers to nonexistent ref, unable to checkout"* and produced an empty directory. Every run said
`done`. The copy was unusable.

Three things came out of it, and all three are in the scripts:

1. `backup.sh` pushes an **explicit refspec** to a branch name read from the repository's own config
   (`logbook.backupBranch`), which `install-backup.sh` discovers from the remote. The two cannot
   drift apart.
2. `install-backup.sh` **clones the repository back** and checks all four files are present and that
   `logbook.db` still hashes to what its own manifest claims. It refuses to enable the timer if not,
   and names the fix. *A push that reports success is evidence about the push; only a clone is
   evidence about the backup.*
3. It refuses if the remote already has refs and this box has never pushed — a README created with
   the repository would make the first push a non-fast-forward, failing at 03:17 where nobody looks.

Also found by rehearsing: **`git init -b main` needs git ≥ 2.28** and Ubuntu 20.04 ships 2.25. The
script uses `git init` plus `symbolic-ref`, which works everywhere.

### Checking it, and restoring

```bash
# Is it running?
systemctl list-timers logbook-backup.timer
journalctl -u logbook-backup.service -n 40

# Is the backup real? This is the only question that matters.
git clone git@github.com:ramiayoub-priv/logbook-backup.git /tmp/lb-check
cat /tmp/lb-check/MANIFEST.txt
sha256sum /tmp/lb-check/logbook.db      # must match the manifest
```

Restoring is `RESTORE.md` inside that clone. The short version: install the app as above but **do
not import the CSVs** (that rebuilds the transcribed books and discards every flight entered in the
app since), put `logbook.db` at `/var/lib/logbook/logbook.db` owned by `logbook`, check the figures
against `MANIFEST.txt`, then start the service.

**A backup nobody has ever restored from is a backup nobody should trust.** Read `RESTORE.md` once
while there is no emergency.

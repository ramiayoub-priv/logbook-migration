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
at the web root of `/var/www/logbook/` for the home-screen install to work.

### ⛔ Nothing under `/logbook` is cached (2026-08-14)

**Owner instruction, verbatim:** *"Can you make sure NOTHING is cached at all? Like the browser needs
to forget (except the cookie for the session)."* Said after a stale-looking phone, where the real
cause was a frontend that had never been uploaded — the second time this class of question had cost
an afternoon. There are now **two rules and no exceptions**:

1. **Apache serves everything under `/var/www/logbook` as `no-store`**, with `Pragma`, `Expires: 0`
   and **no `ETag`** (a 304 is the server saying "use your copy", which is the sentence being
   abolished). One `<Directory>` rule replaces what used to be three: `no-cache` on `sw.js`,
   `no-store` on `index.html`, and `immutable` for a year on `/assets`.
2. **The app registers no service worker.** `src/main.tsx` no longer calls `register()`, and
   `src/noworker.test.ts` fails if that line ever returns.

**⛔ `public/sw.js` MUST KEEP BEING DEPLOYED, and it is now a kill switch.** Deleting it from the
server does **not** remove a worker already installed on a device — the browser goes on running the
copy it has, serving the shell it cached, indefinitely. The only thing that retires an installed
worker is a **new worker at the same URL** that deletes every cache and unregisters itself, then
navigates the open page onto the network. That is what the file does now, and it must stay until
there is no device left that could still be carrying the old one. There is no way to know when that
is, so it stays. Asserted by `src/sw.test.ts`, which runs against the shipped file.

**What was given up, so it can be reconsidered honestly:** the app no longer opens without a network.
The old worker cached the shell for exactly that — an airfield with no signal — but offline *writes*
were never in scope (`app/APP.md` §2), so what it really bought was a shell that opened and then
failed every request it made. Against two stale-build incidents, that was not a good trade.

**What it costs:** ~200 KB fetched on every cold start instead of read from disk. If that ever
becomes annoying the honest middle is `no-cache` (revalidate, cheap 304s) on `/logbook/assets` only —
those filenames are content-hashed and therefore *cannot* be stale. That is a knowing trade for a
later session, not a cleanup to be made quietly.

**The session cookie is untouched.** It is `HttpOnly` and lives in the cookie store, which is not an
HTTP cache. None of this signs anyone out.

Verify a deploy actually reached the device:
```bash
curl -sI https://ayoub.fi/logbook/      | grep -i cache-control   # expect no-store
curl -sI https://ayoub.fi/logbook/sw.js | grep -i cache-control   # expect no-store
ASSET=$(curl -s https://ayoub.fi/logbook/ | grep -o 'assets/index-[A-Za-z0-9_-]*\.js' | head -1)
curl -sI "https://ayoub.fi/logbook/$ASSET" | grep -iE 'cache-control|etag'   # no-store, no ETag
```

**Forcing a phone that is still holding the old worker** (needed once, on each device that had the
app before 2026-08-14): open `https://ayoub.fi/logbook/`, wait a second for the kill switch to
activate — it reloads the page itself — then close and reopen the home-screen app. After that there
is no worker and no cache to be stale.

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

**The frontend has its own rollback, and it is not the binary's.** `rsync -a --delete` leaves nothing
behind, so take a tar of the live directory *before* the rsync — it needs no sudo and costs ~70 KB:

```bash
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
ssh rami@ayoub.fi "tar czf /home/rami/logbook-frontend.$STAMP.tar.gz -C /var/www/logbook ."
# rollback:
ssh rami@ayoub.fi "tar xzf /home/rami/logbook-frontend.<STAMP>.tar.gz -C /var/www/logbook"
```

Most recent: **`/home/rami/logbook-frontend.20260903T203332Z.tar.gz`** — the pre-Task-23 frontend
(`index-qD3NNzOE.js`).

⚠ **Check `--delete` against the live listing first.** `dist/` must produce every entry the web root
holds — `assets/`, `icons/`, `index.html`, `manifest.webmanifest` and **`sw.js`**. Deleting `sw.js`
would strip the kill switch that retires service workers still installed on devices, and no test
catches that because the file is correct in the repo.

## Answering "what is actually running?" — do this before every deploy

The repo has been wrong about this before (three weeks, 2026-08-14 → 2026-09-03; see the decision
log). **Ask the box, not `APP.md`.**

```bash
# which commit the live binary was built from -- Go stamps it in
ssh rami@ayoub.fi 'strings -a /opt/logbook/logbook-server | grep -m1 -E "vcs\.(revision|modified)"'
# which bundle the page actually asks for
curl -s https://ayoub.fi/logbook/ | grep -oE 'assets/index-[A-Za-z0-9_-]+\.js'
# what is genuinely behind
git log --oneline <revision>..HEAD -- app/backend app/frontend
```

⚠ **md5 cannot answer this.** Two builds of the same source differ only in the `vcs.revision` stamp
and build ID — **same size, different hash** — so a hash mismatch proves nothing at all. Read the
stamp. It also tells you whether the live build was **clean** (`vcs.modified=false`) or built from a
dirty tree, which is the difference between a shippable artefact and an accident.

⚠ **A 401 is not a missing route.** Under default deny nearly everything answers 401 without a
session, which looks exactly like a 404-to-be. Probe an obviously fake path
(`/logbook/api/definitely-not-a-route-xyz`) — that returns **404**, so a **401 proves the route
exists**. This is how the whole route table can be inventoried without logging in.

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

✅ **DONE on 2026-08-02.** It is installed, the timer is enabled, and the clone-back has passed.
What follows is the record of how, and what to repeat if the box is ever rebuilt.

Two steps cannot be scripted because both involve a secret (rule §0.3):

```bash
# 1. Generate the key ON THE SERVER, so the private half never travels.
sudo -u logbook install -d -m 0700 -o logbook -g logbook /var/lib/logbook/.ssh
sudo -u logbook ssh-keygen -t ed25519 -N '' -C 'logbook backup (ayoub.fi)' \
     -f /var/lib/logbook/.ssh/backup_ed25519
sudo cat /var/lib/logbook/.ssh/backup_ed25519.pub

# 2. Create https://github.com/ramiayoub-priv/logbook-backup as a PRIVATE repository
#    with NO README and NO licence, and register that public key on the
#    ramiayoub-priv ACCOUNT (Settings -> SSH and GPG keys).
```

⚠ **An account-level key, not a deploy key — OWNER RULING 2026-08-02.** This document previously
instructed a repository deploy key, on the least-privilege argument that it reaches one repository
and can be revoked without touching anything else. The owner ruled otherwise: `ramiayoub-priv` is a
**dedicated account that exists for this and holds nothing else**, so the scoping a deploy key buys
is already provided by the account boundary. **Do not "fix" this back to a deploy key.**

How to tell which one is in use, since it is otherwise invisible — GitHub's greeting names it:

```
Hi ramiayoub-priv!                   -> account-level key   (what is in use)
Hi ramiayoub-priv/logbook-backup!    -> repository deploy key
```

`install-backup.sh` step 5 prints this and reports which kind authenticated, as a fact rather than a
warning. The consequence to keep in view is that the key's reach is **every repository that account
owns** — which is why the account must keep owning nothing else.

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

# And does it hold what it claims to hold? Needs no CSVs and no sqlite3.
/opt/logbook/logbookctl check -db /tmp/lb-check/logbook.db \
                              -manifest /tmp/lb-check/MANIFEST.txt
```

### ⚠ The second trap, found on 2026-08-02 by following RESTORE.md for real

The backup was cloned and restored as a drill, with no emergency in progress — and the instructions
it carries turned out **not to be runnable**. Step 3, the mandatory verification that rule 0.2 hangs
on, said:

```bash
sudo -u logbook sqlite3 /var/lib/logbook/logbook.db 'SELECT COUNT(*), ...'
```

**`sqlite3` is not installed on this box and is not a dependency of this project** — the entire point
of `modernc.org/sqlite` is that nothing outside the Go binary has to speak SQLite. On a fresh server
that line is `command not found`, and the reader either skips the check on a legal record or
`apt install`s a database package mid-restore.

This is the same species as the `GIT_SSH_COMMAND` preflight: **a verification step that cannot do
what it claims.** The fix:

- **`logbookctl check -db <db> [-manifest <file>]`** reads the figures out of a restored database and
  compares them to a backup's `MANIFEST.txt`, printing every one and exiting non-zero on any
  disagreement. No CSVs, no `sqlite3`, no network. It hashes the file **before** opening it.
- `RESTORE.md` is generated from `internal/backup`, so it now ships that command instead, leads with
  `sha256sum` (coreutils, and on its own conclusive), and explicitly warns the reader off `sqlite3`.
- **`install-backend.sh` now installs `logbookctl`**, not just the server. `RESTORE.md` step 1 says
  "install the app as `deploy.md` describes" and step 3 then runs `/opt/logbook/logbookctl` — before
  this, only `install-backup.sh` put it there, so a restore performed exactly as documented reached
  step 3 without the tool.

**`logbookctl verify` is NOT the restore check and must not be substituted for it.** Verify compares
a database against the three transcribed CSVs and is scoped `source_book <> 0` — on a restored server
the books may not be present at all, and even where they are, verify passes happily while every
app-entered flight is missing. Those are precisely the rows that exist nowhere else and that the
whole backup exists to protect. `check` asks the only question a restore raises: **is this the data
the backup recorded?**

Restoring is `RESTORE.md` inside that clone. The short version: install the app as above but **do
not import the CSVs** (that rebuilds the transcribed books and discards every flight entered in the
app since), put `logbook.db` at `/var/lib/logbook/logbook.db` owned by `logbook`, check the figures
against `MANIFEST.txt`, then start the service.

**A backup nobody has ever restored from is a backup nobody should trust.** Read `RESTORE.md` once
while there is no emergency.

#!/bin/bash
# Deploy the logbook -- backend and frontend -- from the DEV MACHINE, as rami,
# with NO password prompt anywhere.
#
# This is the whole deploy. It runs here, not on the box: the only thing that
# happens over there is `sudo -n /opt/logbook/logbook-apply`, which is the one
# command the sudoers rule allows unattended (see install-deploy-privileges.sh).
# Everything else -- building, checking, staging, the frontend rsync -- is done
# as an ordinary user, because /var/www/logbook is owned by rami.
#
# ORDER MATTERS: BINARY FIRST, THEN FRONTEND. A new frontend against an old
# binary calls routes that do not exist; the reverse is harmless. This cost an
# afternoon on 2026-08-02 and the order is now enforced here rather than
# remembered.
set -euo pipefail

REPO="$(cd "$(dirname "$0")/../.." && pwd)"
# The deploy runs as its OWN account, not as the owner's. `rami` is in the sudo
# group; a key that can deploy unattended must not also be a key into an
# account that can become root. deploy-logbook is in no privileged group and
# holds exactly one NOPASSWD right: logbook-apply, with no arguments.
HOST=${LOGBOOK_HOST:-deploy-logbook@ayoub.fi}
STAGE=${LOGBOOK_STAGE:-/home/deploy-logbook/logbook-deploy}
KEY=${LOGBOOK_KEY:-$HOME/.ssh/logbook-deploy}
SSH="ssh -i $KEY -o IdentitiesOnly=yes -o BatchMode=yes"
WEB=${LOGBOOK_WEB:-/var/www/logbook}
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
SKIP_CHECKS=0
ALLOW_DIRTY=0

for a in "$@"; do
    case "$a" in
        --skip-checks) SKIP_CHECKS=1 ;;
        # Only for a genuine emergency, and it is recorded in the binary either
        # way: Go stamps vcs.modified=true and `strings` will say so forever.
        --allow-dirty) ALLOW_DIRTY=1 ;;
        *) echo "usage: $0 [--skip-checks] [--allow-dirty]" >&2; exit 64 ;;
    esac
done

say() { printf '\n== %s\n' "$*"; }
die() { echo "!! $*" >&2; exit 1; }

command -v go >/dev/null || export PATH="$PATH:$HOME/.local/go/bin"
command -v go >/dev/null || die "no go on PATH (it lives at ~/.local/go/bin and is not exported by .bashrc)"

say "1. the tree must be clean"
# app/APP.md, 2026-09-04: dist/ once held a DIRTY build that must never ship.
# The stamp is permanent, so this is checked before the build, not after.
if [ -n "$(cd "$REPO" && git status --porcelain)" ]; then
    if [ "$ALLOW_DIRTY" -eq 0 ]; then
        (cd "$REPO" && git status --short)
        die "uncommitted changes -- commit them, or pass --allow-dirty and own the vcs.modified=true stamp"
    fi
    echo "   !! DIRTY TREE, shipping anyway on --allow-dirty; the binary will say so forever"
fi
HEAD_SHA=$(cd "$REPO" && git rev-parse --short HEAD)
echo "   HEAD $HEAD_SHA"
if [ -n "$(cd "$REPO" && git log origin/master..HEAD --oneline)" ]; then
    echo "   !! HEAD is ahead of origin/master -- push before or after, but do not lose it (rule 0.1)"
fi

if [ "$SKIP_CHECKS" -eq 0 ]; then
    say "2. prove it (rule 0.6)"
    ( cd "$REPO/app/backend"  && make check )
    ( cd "$REPO/app/frontend" && npm run check )
else
    say "2. checks SKIPPED on request -- you are asserting they passed elsewhere"
fi

say "3. build both halves"
( cd "$REPO/app/backend"  && make build )
( cd "$REPO/app/frontend" && npm run build )

say "4. stage the backend artefacts"
# Assembled locally and rsynced as a unit so that SHA256SUMS -- which the box
# verifies before it trusts anything -- is written over the exact bytes that
# are about to be sent. logbook-apply refuses without it, which also makes a
# half-finished upload refuse rather than install.
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cp "$REPO/app/backend/dist/server"     "$TMP/logbook-server"   # cmd/server -> the name the box uses
cp "$REPO/app/backend/dist/logbookctl" "$TMP/logbookctl"
mkdir -p "$TMP/csv"
cp "$REPO"/logbook_1_final.csv "$REPO"/logbook_2_final.csv "$REPO"/logbook_3.csv "$TMP/csv/"
cp "$REPO/app/deploy/logbook-apply" "$TMP/logbook-apply"
# The CSVs are checksummed too: logbook-apply hands them to `logbookctl verify`
# as the drift check on the 1296 frozen historical rows, and a half-copied CSV
# would fail that check for a reason that has nothing to do with the database.
( cd "$TMP" && sha256sum logbook-server logbookctl csv/*.csv > SHA256SUMS )
echo "   $(cd "$TMP" && head -2 SHA256SUMS)"
echo "   stamp: $(strings -a "$TMP/logbook-server" | grep -m1 'vcs.revision' || echo none)"
rsync -a -e "$SSH" "$TMP"/ "$HOST:$STAGE/"

say "5. frontend rollback tar BEFORE anything is deleted"
# rsync --delete leaves nothing behind and needs no sudo to undo -- but only if
# the tar was taken first.
$SSH "$HOST" "tar czf /home/deploy-logbook/logbook-frontend.$STAMP.tar.gz -C $WEB ." \
    && echo "   /home/rami/logbook-frontend.$STAMP.tar.gz"

say "6. the frontend build must be complete before --delete is allowed near it"
# deploy.md: dist/ must produce every entry the web root needs. Deleting sw.js
# would strip the kill switch that retires old service workers, and no test
# catches that because the file is correct in the repo.
for f in index.html sw.js manifest.webmanifest assets icons; do
    [ -e "$REPO/app/frontend/dist/$f" ] || die "app/frontend/dist/$f is missing -- refusing to rsync --delete"
done
echo "   index.html, sw.js, manifest.webmanifest, assets/, icons/ all present"

say "7. BINARY FIRST -- the one privileged step, unattended"
$SSH "$HOST" 'sudo -n /opt/logbook/logbook-apply'

say "8. then the frontend"
rsync -a --delete -e "$SSH" "$REPO/app/frontend/dist/" "$HOST:$WEB/"
echo "   done"

say "9. verify from off-box -- ask the box, never this script"
echo "   live binary: $($SSH "$HOST" 'strings -a /opt/logbook/logbook-server | grep -m1 -E "vcs\.revision"')"
echo "   dirty?       $($SSH "$HOST" 'strings -a /opt/logbook/logbook-server | grep -m1 -E "vcs\.modified"')"
echo "   expected:    $HEAD_SHA"
echo "   live bundle: $(curl -s https://ayoub.fi/logbook/ | grep -oE 'assets/index-[A-Za-z0-9_-]+\.js' | head -1)"
echo "   repo bundle: $(basename "$(ls "$REPO"/app/frontend/dist/assets/index-*.js | head -1)")"
for p in /logbook/api/health /logbook/api/flights /logbook/; do
    printf '   %-24s HTTP %s\n' "$p" "$(curl -sS -o /dev/null -w '%{http_code}' "https://ayoub.fi$p")"
done
printf '   other sites: '
for p in / /blog/ /countdown/ /englishhouse/ /games/ /pdp/ /simpleclock/; do
    printf '%s=%s ' "$p" "$(curl -sS -o /dev/null -w '%{http_code}' "https://ayoub.fi$p")"
done
echo
echo
echo "Frontend rollback: $SSH $HOST 'tar xzf /home/deploy-logbook/logbook-frontend.$STAMP.tar.gz -C $WEB'"
echo "Backend rollback:  ssh $HOST 'sudo cp /opt/logbook/logbook-server.prev /opt/logbook/logbook-server && sudo systemctl restart logbook'"
echo "                   (that one still asks for a password -- rollback is deliberately attended)"

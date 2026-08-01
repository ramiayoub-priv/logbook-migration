#!/bin/bash
# Phase D: publish /logbook through the existing ayoub.fi vhost.
# Additive and reversible. The vhost is backed up first, the config is validated
# before anything is reloaded, and a failed configtest restores the backup and
# leaves the running server untouched.
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
VHOST=/etc/apache2/sites-enabled/ayoub.fi-le-ssl.conf
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
say() { printf '\n== %s\n' "$*"; }

say "1. mod_headers (needed for the sw.js no-cache rule)"
if [ -e /etc/apache2/mods-enabled/headers.load ]; then
    echo "   already enabled"
else
    a2enmod headers
fi

say "2. back up the vhost"
cp -a "$VHOST" "/root/ayoub.fi-le-ssl.conf.$STAMP.bak"
echo "   -> /root/ayoub.fi-le-ssl.conf.$STAMP.bak"

say "3. make the logbook block match apache-logbook.conf"
# The snippet in the staging directory is the source of truth: an existing
# block is REPLACED rather than skipped, so this script is how a change to the
# snippet (a cache header, a new path) actually reaches the vhost. Re-running
# it with an unchanged snippet is a no-op.
#
# strip_block removes our own block and nothing else. Running it over the file
# before and after the edit must produce identical output -- that is the proof
# that this touched no other site's configuration, and it is checked below
# rather than assumed (rule 0.3: additive, reversible, verified).
strip_block() {
    awk '/^# === BEGIN logbook/ { skip = 1 }
         skip { if (/^# === END logbook/) skip = 0; next }
         { print }' "$1"
}

tmp=$(mktemp)
before=$(mktemp)
strip_block "$VHOST" > "$before"
# No blank line is printed around the block. The first version of this script
# added one on every insert, which made a re-run differ from the previous run
# by one line -- enough to trip the "touched something else" check above and to
# make the script non-idempotent. Caught by rehearsing it against a copy of the
# vhost before it was ever run as root.
awk -v snip="$STAGE/apache-logbook.conf" '
    /^<\/VirtualHost>/ && !done {
        while ((getline line < snip) > 0) print line
        done = 1
    }
    { print }
' "$before" > "$tmp"

grep -q "BEGIN logbook" "$tmp" || { echo "   !! insertion failed"; rm -f "$tmp" "$before"; exit 1; }
if ! diff -q "$before" <(strip_block "$tmp") >/dev/null; then
    echo "   !! the edit changed something outside the logbook block -- refusing"
    rm -f "$tmp" "$before"
    exit 1
fi

if diff -q "$VHOST" "$tmp" >/dev/null; then
    echo "   already up to date"
else
    cat "$tmp" > "$VHOST"
    echo "   block written"
fi
rm -f "$tmp" "$before"

say "4. configtest BEFORE any reload"
if ! apache2ctl configtest; then
    echo "   !! configtest FAILED -- restoring the backup, Apache untouched"
    cp -a "/root/ayoub.fi-le-ssl.conf.$STAMP.bak" "$VHOST"
    apache2ctl configtest
    exit 1
fi

say "5. reload (not restart -- other sites keep their connections)"
systemctl reload apache2
sleep 2
systemctl --no-pager is-active apache2

say "6. verify OUR path"
for p in /logbook/api/health /logbook/ /logbook/sw.js; do
    printf '   %-22s HTTP %s\n' "$p" "$(curl -sS -o /dev/null -w '%{http_code}' "https://ayoub.fi$p")"
done
echo "   /logbook/api/flights (no session) HTTP $(curl -sS -o /dev/null -w '%{http_code}' https://ayoub.fi/logbook/api/flights)  (expect 401)"
echo "   sw.js      Cache-Control: $(curl -sSI https://ayoub.fi/logbook/sw.js | grep -i '^cache-control' || echo 'MISSING')"
echo "   index.html Cache-Control: $(curl -sSI https://ayoub.fi/logbook/ | grep -i '^cache-control' || echo 'MISSING')"

say "7. verify THE OTHER SITES are unharmed"
for p in / /blog/ /countdown/ /englishhouse/ /games/ /pdp/ /simpleclock/; do
    printf '   %-18s HTTP %s\n' "$p" "$(curl -sS -o /dev/null -w '%{http_code}' "https://ayoub.fi$p")"
done

say "8. memory headroom"
free -m

echo
echo "Apache reloaded. Backup: /root/ayoub.fi-le-ssl.conf.$STAMP.bak"
echo "Revert = restore that file, 'apache2ctl configtest', 'systemctl reload apache2'."

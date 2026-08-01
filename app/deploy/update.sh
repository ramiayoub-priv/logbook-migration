#!/bin/bash
# Update the LIVE deployment: new binary + the corrected CSV data.
#
# Two things land together on purpose. The reworked new-flight form sends
# takeoff/landing, and the binary now on the box does not accept them -- Go's
# JSON decoder ignores unknown fields, so an old binary would take the flight
# and silently drop two of its times. Shipping the pair together is what stops
# that.
#
# The data is re-imported rather than the file being replaced: your account and
# your session live in the same SQLite file, and a file swap would delete them.
# The importer only clears aircraft, discrepancies and flights WHERE
# source_book <> 0 -- users, sessions and any flight typed into the app survive
# -- inside one transaction that rolls back if a single checksum disagrees.
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
DB=/var/lib/logbook/logbook.db
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
say() { printf '\n== %s\n' "$*"; }

say "1. stop the service so the import has the file to itself"
systemctl stop logbook

say "2. back up the database before a destructive op (CLAUDE.md 0.2)"
install -d -o logbook -g logbook -m 0750 /var/lib/logbook/backups
cp -a "$DB" "/var/lib/logbook/backups/logbook.$STAMP.db"
chown logbook:logbook "/var/lib/logbook/backups/logbook.$STAMP.db"
ls -l "/var/lib/logbook/backups/logbook.$STAMP.db"

say "3. new binary, keeping the old one for rollback"
cp -a /opt/logbook/logbook-server /opt/logbook/logbook-server.prev
install -o root -g root -m 0755 "$STAGE/logbook-server" /opt/logbook/logbook-server
echo "   rollback = cp /opt/logbook/logbook-server.prev /opt/logbook/logbook-server && systemctl restart logbook"

say "4. import the corrected CSVs (1293 -> 1296 flights)"
sudo -u logbook "$STAGE/logbookctl" import -db "$DB" -csv "$STAGE/csv" 2>&1 | grep -vE '^  book' | tail -16

say "5. verify by a separate code path"
sudo -u logbook "$STAGE/logbookctl" verify -db "$DB" -csv "$STAGE/csv" 2>&1 | grep -vE '^  book'

say "6. start"
systemctl start logbook
sleep 3
systemctl --no-pager --full status logbook | head -12

say "7. checks"
curl -sS -o /dev/null -w "   local  health:  %{http_code}\n" http://127.0.0.1:9002/logbook/api/health
curl -sS -o /dev/null -w "   public health:  %{http_code}\n" https://ayoub.fi/logbook/api/health
curl -sS -o /dev/null -w "   no session ->   %{http_code}  (expect 401)\n" https://ayoub.fi/logbook/api/flights
printf '   other sites: '
for p in / /blog/ /countdown/ /englishhouse/ /games/ /pdp/ /simpleclock/; do
    printf '%s=%s ' "$p" "$(curl -sS -o /dev/null -w '%{http_code}' "https://ayoub.fi$p")"
done
echo
free -m | head -2

echo
echo "The startup log line above must read flights=1296."
echo "DB backup: /var/lib/logbook/backups/logbook.$STAMP.db"

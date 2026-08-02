#!/bin/bash
# Update the LIVE deployment: new binaries, then verify what is already there.
#
# ⛔ THIS SCRIPT NO LONGER IMPORTS. Owner ruling, 2026-08-02: "we should start
# treating the production database now as the source of truth. We don't need
# the importer anymore." The migration is finished and the three CSVs are
# frozen (CLAUDE.md 0.8), so a re-import could only ever reproduce rows that
# cannot have changed -- while running DELETE against a live legal record to do
# it. Best case a no-op, worst case the stale-CSV incident of 2026-08-02, where
# the box was one root command from writing `C192` back into production and
# reporting 61 discrepancies, with the only signal a number buried in a long
# transcript.
#
# WHAT PROTECTS THE DATA IS THE OFF-BOX BACKUP (backup.sh, daily, proven
# restorable on 2026-08-02), not the ability to rebuild from CSVs -- which
# stopped being a complete answer the moment the first flight was entered in
# the app, because those rows are in no CSV.
#
# `verify` STAYS, and is now the point of step 4: it is READ-ONLY and checks
# the database against the CSVs on nine checksums, which makes it a drift and
# tamper check on the 1296 frozen historical rows rather than a rebuild. It
# writes nothing, so it is safe to run on every deploy, forever.
#
# The binary and the frontend must land together, BINARY FIRST -- see
# app/APP.md. A new frontend against an old binary calls routes that do not
# exist; the reverse is harmless.
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
DB=/var/lib/logbook/logbook.db
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
say() { printf '\n== %s\n' "$*"; }

say "1. stop the service"
systemctl stop logbook

say "2. back up the database first, always (CLAUDE.md 0.2)"
# Kept even though this script no longer writes to the database: the new binary
# may apply an additive schema migration on its first start, and a copy taken
# before that costs a second and is the only way back.
install -d -o logbook -g logbook -m 0750 /var/lib/logbook/backups
cp -a "$DB" "/var/lib/logbook/backups/logbook.$STAMP.db"
chown logbook:logbook "/var/lib/logbook/backups/logbook.$STAMP.db"
ls -l "/var/lib/logbook/backups/logbook.$STAMP.db"

say "3. new binaries, keeping the old server for rollback"
cp -a /opt/logbook/logbook-server /opt/logbook/logbook-server.prev
install -o root -g root -m 0755 "$STAGE/logbook-server" /opt/logbook/logbook-server
# logbookctl is what takes the backup and what verifies a restore, so it must
# not fall behind the server.
if [ -f "$STAGE/logbookctl" ]; then
    install -o root -g root -m 0755 "$STAGE/logbookctl" /opt/logbook/logbookctl
fi
echo "   rollback = cp /opt/logbook/logbook-server.prev /opt/logbook/logbook-server && systemctl restart logbook"

say "4. verify the historical rows against the frozen CSVs (READ-ONLY, writes nothing)"
# A mismatch here does NOT mean "re-import". It means the 1296 transcribed rows
# in the live database no longer agree with the frozen books, which is a defect
# to investigate before the service is allowed to write again (CLAUDE.md 0.2).
sudo -u logbook /opt/logbook/logbookctl verify -db "$DB" -csv "$STAGE/csv" 2>&1 | grep -vE '^  book'

say "5. start"
systemctl start logbook
sleep 3
systemctl --no-pager --full status logbook | head -12

say "6. checks"
curl -sS -o /dev/null -w "   local  health:  %{http_code}\n" http://127.0.0.1:9002/logbook/api/health
curl -sS -o /dev/null -w "   public health:  %{http_code}\n" https://ayoub.fi/logbook/api/health
curl -sS -o /dev/null -w "   no session ->   %{http_code}  (expect 401)\n" https://ayoub.fi/logbook/api/flights
curl -sS -o /dev/null -w "   aircraft   ->   %{http_code}  (expect 401)\n" https://ayoub.fi/logbook/api/aircraft
printf '   other sites: '
for p in / /blog/ /countdown/ /englishhouse/ /games/ /pdp/ /simpleclock/; do
    printf '%s=%s ' "$p" "$(curl -sS -o /dev/null -w '%{http_code}' "https://ayoub.fi$p")"
done
echo
free -m | head -2

echo
echo "The startup line above reports the LIVE flight count: 1296 + every flight"
echo "entered in the app since. It was 1298 on 2026-08-02 and only goes up."
echo "It is NOT expected to equal 1296, and a session that 'fixes' that is wrong."
echo "DB backup: /var/lib/logbook/backups/logbook.$STAMP.db"

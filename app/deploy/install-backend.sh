#!/bin/bash
# Phase A: install the logbook backend. Touches NOTHING that is already serving
# traffic -- no Apache change happens here. If this fails, the box is unchanged
# apart from an inert user, two directories and a stopped service.
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
say() { printf '\n== %s\n' "$*"; }

say "1. service user"
if id -u logbook >/dev/null 2>&1; then
    echo "   user 'logbook' already exists -- leaving it alone"
else
    adduser --system --group --no-create-home --home /var/lib/logbook \
            --shell /usr/sbin/nologin logbook
fi

say "2. directories"
install -d -o root     -g root     -m 0755 /opt/logbook
install -d -o logbook  -g logbook  -m 0750 /var/lib/logbook
install -d -o logbook  -g logbook  -m 0750 /var/lib/logbook/backups
install -d -o rami     -g rami     -m 0755 /var/www/logbook

say "3. binary"
if [ -f /opt/logbook/logbook-server ]; then
    cp -a /opt/logbook/logbook-server /opt/logbook/logbook-server.prev
    echo "   previous binary kept as logbook-server.prev"
fi
install -o root -g root -m 0755 "$STAGE/logbook-server" /opt/logbook/logbook-server
/opt/logbook/logbook-server -h 2>&1 | head -3 || true

say "4. database"
if [ -f /var/lib/logbook/logbook.db ]; then
    echo "   !! /var/lib/logbook/logbook.db ALREADY EXISTS."
    echo "   !! Refusing to overwrite a live legal record (CLAUDE.md rule 0.2)."
    echo "   !! Back it up and remove it deliberately if a replacement is intended."
else
    install -o logbook -g logbook -m 0600 "$STAGE/logbook.db" /var/lib/logbook/logbook.db
    echo "   installed, $(stat -c '%s bytes, mode %a, owner %U' /var/lib/logbook/logbook.db)"
fi

say "5. systemd unit"
install -o root -g root -m 0644 "$STAGE/logbook.service" /etc/systemd/system/logbook.service
systemctl daemon-reload
systemctl enable logbook
systemctl restart logbook
sleep 3
systemctl --no-pager --full status logbook | head -20

say "6. health check on 127.0.0.1:9002 (not yet public)"
# The server mounts routes at the FULL public path, /logbook/api/ -- not /api/.
code=$(curl -sS -o /tmp/lb-health -w '%{http_code}' http://127.0.0.1:9002/logbook/api/health || echo FAIL)
echo "   HTTP $code  body: $(cat /tmp/lb-health 2>/dev/null)"
[ "$code" = "200" ] || { echo "   !! health check FAILED -- stopping here"; exit 1; }

say "7. default-deny check: an authenticated endpoint without a session"
echo "   /logbook/api/flights -> HTTP $(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:9002/logbook/api/flights)  (expect 401)"

say "8. memory"
systemctl show logbook -p MemoryCurrent
free -m

echo
echo "Backend installed. Apache has NOT been touched yet."

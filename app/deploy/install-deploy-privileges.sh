#!/bin/bash
# Phase E, run ONCE as root: make backend deploys unattended.
#
# Installs two things and nothing else -- /opt/logbook/logbook-apply (root-owned)
# and /etc/sudoers.d/logbook-deploy (one user, one command, no arguments). After
# this, `deploy.sh` on the dev machine ships a backend with no password prompt.
#
# ⛔ THE DANGEROUS PART IS THE SUDOERS FILE. A malformed file in
# /etc/sudoers.d/ can break sudo for every user on the box, and this box's only
# route to root is sudo. So the candidate is validated with `visudo -cf` BEFORE
# it is installed, the whole tree is re-validated AFTER, and a failure at that
# second check removes the file again rather than leaving it in place. Run this
# from a session you KEEP OPEN, and prove sudo still works from a second
# connection before you close the first (rule 0.3).
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
SUDOERS=/etc/sudoers.d/logbook-deploy
say() { printf '\n== %s\n' "$*"; }

[ "$(id -u)" -eq 0 ] || { echo "run as root: sudo $0" >&2; exit 1; }
id -u rami >/dev/null 2>&1 || { echo "no user 'rami' on this box -- refusing" >&2; exit 1; }

say "1. install the apply script, root-owned"
# Root ownership is load-bearing: the sudoers rule below names this path, so if
# rami could write to it the grant would be unrestricted root.
install -o root -g root -m 0755 "$STAGE/logbook-apply" /opt/logbook/logbook-apply
ls -l /opt/logbook/logbook-apply
bash -n /opt/logbook/logbook-apply && echo "   syntax OK"

say "2. validate the sudoers rule BEFORE installing it"
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
cat "$STAGE/sudoers-logbook-deploy" > "$tmp"
chmod 0440 "$tmp"
if ! visudo -cf "$tmp"; then
    echo "   !! the candidate rule is malformed -- nothing installed"
    exit 1
fi
echo "   candidate parses clean"

say "3. install it"
if [ -f "$SUDOERS" ] && diff -q "$SUDOERS" "$tmp" >/dev/null; then
    echo "   already up to date"
else
    install -o root -g root -m 0440 "$tmp" "$SUDOERS"
    echo "   written: $SUDOERS"
fi

say "4. re-validate the WHOLE sudoers tree"
# The candidate can be valid alone and still collide with something already in
# /etc/sudoers.d. If that happens, back the change out immediately: a box whose
# sudo is broken cannot be repaired without sudo.
if ! visudo -c; then
    echo "   !! THE TREE IS BROKEN -- removing our file and re-checking"
    rm -f "$SUDOERS"
    visudo -c && echo "   sudo is healthy again; nothing was left behind"
    exit 1
fi

say "5. prove the grant is exactly what was intended"
sudo -n -l -U rami 2>&1 | sed 's/^/   /'
echo
echo "   Expect exactly one NOPASSWD entry, for /opt/logbook/logbook-apply with no arguments."
echo "   Anything broader than that is a defect -- remove $SUDOERS and say so."

say "6. prove it actually works, as rami, without a password"
# The real proof is not that the rule parses; it is that the deploy path runs.
# logbook-apply refuses with exit 1 when nothing is staged, and that refusal is
# a SUCCESS here: it means sudo let rami in without asking for a password.
if sudo -u rami sudo -n /opt/logbook/logbook-apply >/dev/null 2>&1; then
    echo "   rami ran it with no password (and it had something staged)"
else
    rc=$?
    if [ "$rc" -eq 1 ]; then
        echo "   rami ran it with no password; it refused because nothing is staged -- correct"
    else
        echo "   !! rami could NOT run it without a password (exit $rc) -- the grant is not working"
        exit 1
    fi
fi

say "7. prove the grant did NOT widen"
# rami is in the sudo group and can still reach root WITH a password -- that was
# always true and is not what this change is about. What must NOT be true is
# passwordless root in general.
if sudo -u rami sudo -n true 2>/dev/null; then
    echo "   !! rami has passwordless sudo for EVERYTHING -- that is not what this installs."
    echo "   !! Look for another file in /etc/sudoers.d/ granting ALL, and remove this one."
    exit 1
fi
echo "   passwordless sudo is limited to logbook-apply; general sudo still asks for a password"

echo
echo "Done. Deploys from the dev machine now need no password:"
echo "    app/deploy/deploy.sh"
echo
echo "KEEP THIS SESSION OPEN and check from a second connection that ordinary"
echo "sudo still prompts and still works before you close it (rule 0.3)."
echo "Revert: rm $SUDOERS   (immediate; the apply script can stay, it is inert without the rule)"

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

# ⛔ SUPERSEDED BY install-deploy-user.sh (2026-09-06). This script grants the
# NOPASSWD right to `rami`, and the owner has since ruled that rami must be
# back to password-for-everything, with the deploy running as its own
# unprivileged account. Both scripts write the same sudoers file, so re-running
# this one AFTER the switch would silently hand the right back to a
# sudo-group account. Refuse instead.
if id -u deploy-logbook >/dev/null 2>&1; then
    echo "!! deploy-logbook exists -- this script is superseded and would re-grant rami." >&2
    echo "!! Use install-deploy-user.sh instead." >&2
    exit 1
fi

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

say "5. prove the grant is right -- BY READING IT, NEVER BY RUNNING IT"
# ⛔ THIS STEP MUST NOT EXECUTE logbook-apply. The first version did, as a
# "does sudo really let rami in?" proof -- and on 2026-09-06 that ran a REAL
# DEPLOY of the production logbook: it stopped the service, reinstalled the
# binary and started it again, from inside a script whose entire job was to
# check a permission. A verification step with a side effect on a legal record
# is not a verification step.
#
# The second fault was in the check that followed it. `sudo -u rami sudo -n
# true` was meant to prove passwordless root had NOT widened, and it reported
# that it HAD -- a false alarm, because the owner had just typed their password
# to run this very script, so rami's sudo credential cache was warm. It was
# testing the timestamp, not the grant.
#
# Both are answered by reading the policy instead of exercising it. `sudo -l -U`
# runs as root, needs no password, consults no cache, and changes nothing.
LIST=$(sudo -l -U rami 2>/dev/null || true)
printf '%s\n' "$LIST" | sed 's/^/   /'
echo

NOPASS_TOTAL=$(printf '%s\n' "$LIST" | grep -c 'NOPASSWD' || true)
NOPASS_OURS=$(printf '%s\n' "$LIST" | grep 'NOPASSWD' | grep -c 'logbook-apply' || true)
# sudo prints the restriction escaped -- \"\" , not "" -- so the backslashes
# come out before the test. Checked against the box's real output, because
# guessing at this would have produced a confident false failure.
NOARG=$(printf '%s\n' "$LIST" | grep 'NOPASSWD' | tr -d '\\' | grep -c '""' || true)

fail=0
if [ "$NOPASS_TOTAL" -ne 1 ]; then
    echo "   !! $NOPASS_TOTAL NOPASSWD entries -- expected exactly 1. Something else grants"
    echo "   !! passwordless sudo. Find it in /etc/sudoers.d/ before trusting this box."
    fail=1
fi
if [ "$NOPASS_OURS" -ne 1 ]; then
    echo "   !! the NOPASSWD entry is not for logbook-apply -- refusing to call this installed"
    fail=1
fi
if [ "$NOARG" -ne 1 ]; then
    echo "   !! the NOPASSWD entry does not carry the no-argument restriction ("")."
    echo "   !! Without it, NOPASSWD on a path permits ANY arguments."
    fail=1
fi
if [ "$fail" -eq 1 ]; then
    echo "   !! Revert with: rm $SUDOERS"
    exit 1
fi
echo "   exactly one NOPASSWD entry, for logbook-apply, restricted to no arguments"

say "6. prove rami may run it -- again without running it"
# `sudo -l -U <user> <command>` answers "is this permitted?" and prints the
# command. It does not execute it.
if sudo -l -U rami /opt/logbook/logbook-apply >/dev/null 2>&1; then
    echo "   permitted: rami may run /opt/logbook/logbook-apply"
else
    echo "   !! rami is NOT permitted to run /opt/logbook/logbook-apply -- the grant is not working"
    exit 1
fi

say "7. ordinary sudo must still ask for a password"
# Checked by reading the policy, for the reason in step 6: any execution-based
# check here is contaminated by the credential cache of the sudo that started
# this script. `(ALL : ALL) ALL` with no NOPASSWD is the sudo-group membership
# rami has always had, and it is correct for it to be there.
if printf '%s\n' "$LIST" | grep -qE 'NOPASSWD.*\bALL\b *$'; then
    echo "   !! rami has passwordless sudo for ALL commands -- that is not what this installs."
    echo "   !! Remove $SUDOERS and investigate."
    exit 1
fi
echo "   general sudo still requires a password (the (ALL : ALL) ALL line above is"
echo "   rami's long-standing sudo-group membership, and it is NOT passwordless)"

echo
echo "Done. Deploys from the dev machine now need no password:"
echo "    app/deploy/deploy.sh"
echo
echo "KEEP THIS SESSION OPEN and check from a second connection that ordinary"
echo "sudo still prompts and still works before you close it (rule 0.3)."
echo "Revert: rm $SUDOERS   (immediate; the apply script can stay, it is inert without the rule)"

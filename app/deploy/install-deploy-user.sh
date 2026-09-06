#!/bin/bash
# Phase F, run as root: give the deploy its OWN account, and take the
# unattended-root right away from the owner's.
#
# WHY. Phase E parked the NOPASSWD right on `rami` because `rami` was already
# there. But `rami` is in the `sudo` group, so an SSH key that can deploy
# unattended is also a key into an account that can become root with a
# password. Owner ruling 2026-09-06: *"we need sudo to require password for
# rami... create a deploy-logbook user for that task specifically"*. Correct,
# and the right instinct: the deploy account should be able to do the deploy
# and nothing else.
#
# TWO PHASES, ON PURPOSE, SO NOBODY GETS LOCKED OUT.
#
#   sudo ./install-deploy-user.sh                 # create + prove, rami KEEPS its right
#   sudo ./install-deploy-user.sh --revoke-rami   # only after the new account is proven
#
# Nothing here touches rami's `sudo` group membership. rami keeps full root
# with a password throughout, in both phases and afterwards. The only thing
# --revoke-rami removes is one NOPASSWD line.
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
SUDOERS=/etc/sudoers.d/logbook-deploy
USER_NAME=deploy-logbook
USER_HOME=/home/$USER_NAME
DEPLOY_STAGE=$USER_HOME/logbook-deploy
WEB=/var/www/logbook
GROUP=logbookdeploy
REVOKE=0

case "${1:-}" in
    "")             REVOKE=0 ;;
    --revoke-rami)  REVOKE=1 ;;
    *) echo "usage: $0 [--revoke-rami]" >&2; exit 64 ;;
esac

say() { printf '\n== %s\n' "$*"; }
[ "$(id -u)" -eq 0 ] || { echo "run as root: sudo $0 ${1:-}" >&2; exit 1; }

# The sudoers file is written from here rather than from a static template
# because its contents differ between the two phases, and a stale template is
# exactly how a "temporary" grant becomes permanent.
write_sudoers() {
    local body="$1" tmp
    tmp=$(mktemp); chmod 0440 "$tmp"
    printf '%s\n' "$body" > "$tmp"
    if ! visudo -cf "$tmp"; then
        echo "   !! candidate rule is malformed -- nothing changed"; rm -f "$tmp"; exit 1
    fi
    install -o root -g root -m 0440 "$tmp" "$SUDOERS"
    rm -f "$tmp"
    if ! visudo -c >/dev/null; then
        echo "   !! THE SUDOERS TREE IS BROKEN -- removing our file and re-checking"
        rm -f "$SUDOERS"; visudo -c && echo "   sudo is healthy again"; exit 1
    fi
    echo "   written and the whole tree re-validates"
}

RULE_DEPLOY="$USER_NAME ALL=(root) NOPASSWD: /opt/logbook/logbook-apply \"\""
RULE_RAMI='rami ALL=(root) NOPASSWD: /opt/logbook/logbook-apply ""'
HEADER='# Managed by app/deploy/install-deploy-user.sh -- edit there, not here (rule 0.1).
#
# One command, by absolute path, with no arguments. The trailing "" is sudo,s
# syntax for "no arguments": without it NOPASSWD on a path permits ANY of them.
# /opt/logbook/logbook-apply is root-owned 0755, so the account named here
# cannot rewrite what the grant points at.'
HEADER=${HEADER//sudo,s/sudo\'s}

if [ "$REVOKE" -eq 1 ]; then
    say "REVOKE: take the unattended-root right away from rami"
    id -u "$USER_NAME" >/dev/null 2>&1 || { echo "!! $USER_NAME does not exist -- run phase 1 first"; exit 1; }
    # Refuse to revoke until the replacement is actually usable. Removing the
    # only working deploy path and calling it done would be worse than leaving
    # the grant where it is.
    sudo -l -U "$USER_NAME" 2>/dev/null | grep -q 'logbook-apply' \
        || { echo "!! $USER_NAME has no deploy grant yet -- refusing to revoke rami's"; exit 1; }
    [ -s "$USER_HOME/.ssh/authorized_keys" ] \
        || { echo "!! $USER_NAME has no authorized_keys -- refusing to revoke rami's"; exit 1; }
    write_sudoers "$HEADER

$RULE_DEPLOY"
    say "verify: rami must have NO passwordless right left"
    if sudo -l -U rami 2>/dev/null | grep -q 'NOPASSWD'; then
        echo "   !! rami STILL has a NOPASSWD entry -- look at $SUDOERS"; exit 1
    fi
    echo "   rami has no NOPASSWD entry"
    if sudo -l -U rami 2>/dev/null | grep -qE '\(ALL : ALL\) ALL'; then
        echo "   rami still has full sudo WITH a password (sudo-group membership, untouched)"
    else
        echo "   !! rami appears to have lost sudo entirely -- THIS IS NOT WHAT THIS SCRIPT DOES."
        echo "   !! Check group membership: id rami"; exit 1
    fi
    sudo -l -U "$USER_NAME" 2>/dev/null | sed 's/^/   /'
    echo
    echo "Done. Deploys run as $USER_NAME; rami is back to password-for-everything."
    exit 0
fi

say "1. the deploy account"
if id -u "$USER_NAME" >/dev/null 2>&1; then
    echo "   $USER_NAME already exists -- leaving it alone"
else
    # A real login account (it must accept SSH) with NO password: key only.
    adduser --disabled-password --gecos "logbook deploy" --home "$USER_HOME" "$USER_NAME"
fi
passwd -l "$USER_NAME" >/dev/null && echo "   password login disabled"
# NOT in sudo, NOT in adm, NOT in anything privileged. Say so out loud.
echo "   groups: $(id -nG "$USER_NAME")"
if id -nG "$USER_NAME" | tr ' ' '\n' | grep -qxE 'sudo|admin|adm|root'; then
    echo "   !! $USER_NAME is in a privileged group -- that defeats the whole point"; exit 1
fi

say "2. shared group for the web root"
# /var/www/logbook is owned by rami. The deploy account needs to rsync --delete
# into it, so both accounts get a group in common. rami keeps ownership.
getent group "$GROUP" >/dev/null || groupadd "$GROUP"
usermod -aG "$GROUP" "$USER_NAME"
usermod -aG "$GROUP" rami
chgrp -R "$GROUP" "$WEB"
chmod -R g+w "$WEB"
find "$WEB" -type d -exec chmod g+s {} +   # new files inherit the group
ls -ld "$WEB" | sed 's/^/   /'

say "3. ssh key (deploy only, no passphrase, generated on the dev machine)"
[ -f "$STAGE/deploy-logbook.pub" ] || { echo "   !! $STAGE/deploy-logbook.pub is missing -- stage it first"; exit 1; }
install -d -o "$USER_NAME" -g "$USER_NAME" -m 0700 "$USER_HOME/.ssh"
install -o "$USER_NAME" -g "$USER_NAME" -m 0600 "$STAGE/deploy-logbook.pub" "$USER_HOME/.ssh/authorized_keys"
echo "   $(cut -d' ' -f3- "$USER_HOME/.ssh/authorized_keys")"

say "4. staging directory"
install -d -o "$USER_NAME" -g "$USER_NAME" -m 0700 "$DEPLOY_STAGE"
ls -ld "$DEPLOY_STAGE" | sed 's/^/   /'

say "5. the apply script reads from that directory now -- reinstall it"
install -o root -g root -m 0755 "$STAGE/logbook-apply" /opt/logbook/logbook-apply
grep -m1 '^STAGE=' /opt/logbook/logbook-apply | sed 's/^/   /'

say "6. grant BOTH for now -- rami's right is removed in phase 2, not here"
write_sudoers "$HEADER

$RULE_RAMI
$RULE_DEPLOY"

say "7. prove the new grant by READING the policy, never by running it"
# Reading, not running: a check that executes logbook-apply performs a real
# deploy, which is what the first version of install-deploy-privileges.sh did
# on 2026-09-06 (see the decision log).
LIST=$(sudo -l -U "$USER_NAME" 2>/dev/null || true)
printf '%s\n' "$LIST" | sed 's/^/   /'
printf '%s\n' "$LIST" | grep 'NOPASSWD' | grep -q 'logbook-apply' \
    || { echo "   !! $USER_NAME cannot run logbook-apply"; exit 1; }
printf '%s\n' "$LIST" | grep 'NOPASSWD' | tr -d '\\' | grep -q '""' \
    || { echo "   !! the no-argument restriction is missing"; exit 1; }
printf '%s\n' "$LIST" | grep -qE 'NOPASSWD.*\bALL\b *$' \
    && { echo "   !! $USER_NAME has passwordless sudo for ALL -- refusing"; exit 1; }
echo "   exactly the intended grant: logbook-apply, no arguments, nothing else"

echo
echo "Phase 1 done. rami's right is UNCHANGED so nothing can lock you out yet."
echo
echo "Next: from the dev machine, prove the new account really deploys:"
echo "    app/deploy/deploy.sh"
echo
echo "ONLY when that has worked end to end:"
echo "    sudo $STAGE/install-deploy-user.sh --revoke-rami"

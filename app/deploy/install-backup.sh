#!/bin/bash
# Install the daily off-box backup. Idempotent; safe to re-run.
#
# It touches nothing that is already serving traffic: no Apache change, no
# firewall change, no change to logbook.service. If this fails the box is left
# with an inert directory and a stopped timer.
#
# BEFORE RUNNING IT, THE OWNER MUST DO TWO THINGS BY HAND, and neither can be
# scripted from here because both involve a secret (rule 0.3, no secrets in the
# repo -- ever):
#
#   1. Create the key ON THE SERVER, so the private half never travels:
#
#        sudo -u logbook ssh-keygen -t ed25519 -N '' \
#             -C 'logbook backup (ayoub.fi)' \
#             -f /var/lib/logbook/.ssh/backup_ed25519
#        sudo cat /var/lib/logbook/.ssh/backup_ed25519.pub
#
#   2. Add that PUBLIC key to the ramiayoub-priv ACCOUNT, under GitHub
#      Settings -> SSH and GPG keys, and create
#      https://github.com/ramiayoub-priv/logbook-backup as a PRIVATE repository
#      with NO README and NO licence.
#
#      OWNER RULING 2026-08-02: an account-level key on a DEDICATED account,
#      not a repository deploy key. This header used to say the opposite, on the
#      least-privilege argument. The owner's answer is that ramiayoub-priv
#      exists for this and holds nothing else, so the account boundary already
#      is the scope. Step 5 reports which kind authenticated. Do NOT change this
#      back; see app/docs/deploy.md.
#
#      An SSH key either way, never a personal access token: the private half is
#      generated on the box and never leaves it.
#
#      THE REPOSITORY MUST BE PRIVATE. It contains the whole logbook and the
#      account's Argon2id password hash -- see app/docs/security.md.
set -euo pipefail

STAGE="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR=/var/lib/logbook/backup
SSH_DIR=/var/lib/logbook/.ssh
REMOTE=${LOGBOOK_BACKUP_REMOTE:-git@github.com:ramiayoub-priv/logbook-backup.git}
say() { printf '\n== %s\n' "$*"; }

say "0. prerequisites"
command -v git >/dev/null || { echo "   !! git is not installed"; exit 1; }
if [ ! -f "$SSH_DIR/backup_ed25519" ]; then
    echo "   !! No deploy key at $SSH_DIR/backup_ed25519."
    echo "   !! Create it and register the .pub half on GitHub first -- see the"
    echo "   !! comment at the top of this script. Stopping; nothing changed."
    exit 1
fi
echo "   deploy key present: $(stat -c 'mode %a, owner %U' "$SSH_DIR/backup_ed25519")"

say "1. directories and permissions"
install -d -o logbook -g logbook -m 0700 "$SSH_DIR"
install -d -o logbook -g logbook -m 0750 "$REPO_DIR"
chmod 0600 "$SSH_DIR/backup_ed25519"
chown logbook:logbook "$SSH_DIR/backup_ed25519"

say "2. pin GitHub's host keys"
# StrictHostKeyChecking=yes in backup.sh means the push FAILS on an unknown or
# changed host key rather than trusting it. That is only worth anything if this
# file was written when the box was known-good, which is now.
if [ -s "$SSH_DIR/known_hosts" ] && grep -q github.com "$SSH_DIR/known_hosts"; then
    echo "   known_hosts already carries github.com -- leaving it alone"
else
    ssh-keyscan -t rsa,ecdsa,ed25519 github.com > "$SSH_DIR/known_hosts" 2>/dev/null
    echo "   pinned $(grep -c github.com "$SSH_DIR/known_hosts") github.com host keys"
    echo "   VERIFY these against https://api.github.com/meta before trusting them:"
    ssh-keygen -lf "$SSH_DIR/known_hosts" | sed 's/^/     /'
fi
chown logbook:logbook "$SSH_DIR/known_hosts"
chmod 0644 "$SSH_DIR/known_hosts"

say "3. the backup repository"
# ONE definition of the ssh options, used two ways -- and they are NOT
# interchangeable, which cost a whole debugging session on 2026-08-02.
#
# GIT_SSH_COMMAND is read by GIT. It does nothing whatsoever for a bare `ssh`.
# The first version defined only asuser() and then called `asuser ssh -T
# git@github.com` in step 5, which ran plain ssh as the logbook user with no -i:
# it looked for a default identity at ~/.ssh/id_*, found none, and reported
# "Permission denied (publickey)". So step 5 -- the check whose entire job is to
# prove the deploy key works -- never once tested the deploy key, and blamed a
# GitHub setup that was correct. Every git operation around it was authenticating
# fine on the same key.
SSH_OPTS=(-i "$SSH_DIR/backup_ed25519" -o IdentitiesOnly=yes
          -o StrictHostKeyChecking=yes -o UserKnownHostsFile="$SSH_DIR/known_hosts"
          -o BatchMode=yes)
# for git: git spawns this string itself
asuser()     { sudo -u logbook env GIT_SSH_COMMAND="ssh ${SSH_OPTS[*]}" "$@"; }
# for ssh itself: the options must be real argv, not an environment variable
asuser_ssh() { sudo -u logbook ssh "${SSH_OPTS[@]}" "$@"; }

# WHICH BRANCH the remote calls its default, asked of the remote rather than
# assumed. Pushing `main` at a remote whose HEAD says `master` produces a
# repository that pushes cleanly forever and clones back EMPTY -- found by
# rehearsing, and it is the failure mode that only shows up on the day the
# backup is needed.
#
# It must also be HONEST about which of those two it did. The first version sent
# ls-remote's stderr to /dev/null and fell back to `main`, then printed "the
# remote's default branch is 'main'" either way. On 2026-08-02 that announced a
# discovered fact during a run whose every remote operation was failing on
# authentication, and sent the debugging after the wrong thing entirely.
# Reporting an assumption as a discovery is the same defect this script exists
# to catch one step later.
# THREE outcomes, not two, and the first version collapsed the last two into a
# scary message. A brand-new empty repository HAS no default branch to report:
# ls-remote succeeds and prints nothing. That is the normal first run, and
# saying "could not read the branch, the key probably cannot authenticate" about
# it sends the reader after a problem that does not exist -- which is what
# happened on 2026-08-02, one line above a step 5 that authenticated perfectly.
# So branch it on ls-remote's EXIT STATUS, not on whether the output was empty.
ls_err=$(mktemp)
if ls_out=$(asuser git ls-remote --symref "$REMOTE" HEAD 2>"$ls_err"); then ls_rc=0; else ls_rc=$?; fi
BRANCH=$(printf '%s\n' "${ls_out:-}" |
         awk '$1=="ref:" {sub("refs/heads/","",$2); print $2; exit}')
if [ -n "$BRANCH" ]; then
    echo "   the remote's default branch is '$BRANCH', read from the remote"
elif [ "$ls_rc" -eq 0 ]; then
    BRANCH=main
    echo "   the remote is reachable and EMPTY -- no default branch yet, using '$BRANCH'"
else
    BRANCH=main
    echo "   !! could NOT REACH the remote (git exit $ls_rc) -- ASSUMING '$BRANCH'."
    echo "   !! Step 5 decides whether this is an authentication problem. git said:"
    sed 's/^/     /' "$ls_err"
fi
rm -f "$ls_err"

if [ -d "$REPO_DIR/.git" ]; then
    echo "   already a git repository -- leaving its history alone"
else
    # NOT `git init -b main`: that flag arrived in git 2.28 and Ubuntu 20.04
    # ships 2.25, which fails with "unknown switch `b'". Found by rehearsing
    # this script rather than by reading it. symbolic-ref works everywhere and
    # is what -b does underneath.
    sudo -u logbook git init -q "$REPO_DIR"
    echo "   initialised"
fi
sudo -u logbook git -C "$REPO_DIR" symbolic-ref HEAD "refs/heads/$BRANCH"
# Identity is repo-local so it cannot leak into any other repository on the box.
sudo -u logbook git -C "$REPO_DIR" config user.name  "logbook backup"
sudo -u logbook git -C "$REPO_DIR" config user.email "logbook-backup@ayoub.fi"
# backup.sh reads this rather than guessing, so the two cannot drift apart.
sudo -u logbook git -C "$REPO_DIR" config logbook.backupBranch "$BRANCH"
sudo -u logbook git -C "$REPO_DIR" remote remove origin 2>/dev/null || true
sudo -u logbook git -C "$REPO_DIR" remote add origin "$REMOTE"
echo "   remote: $(sudo -u logbook git -C "$REPO_DIR" remote get-url origin)"

say "4. scripts and units"
install -o root -g root -m 0755 "$STAGE/backup.sh" /opt/logbook/backup.sh
install -o root -g root -m 0644 "$STAGE/logbook-backup.service" /etc/systemd/system/logbook-backup.service
install -o root -g root -m 0644 "$STAGE/logbook-backup.timer"   /etc/systemd/system/logbook-backup.timer
# logbookctl is what actually takes the snapshot; ship it alongside the server.
if [ -f "$STAGE/logbookctl" ]; then
    install -o root -g root -m 0755 "$STAGE/logbookctl" /opt/logbook/logbookctl
fi
[ -x /opt/logbook/logbookctl ] || { echo "   !! /opt/logbook/logbookctl is missing"; exit 1; }
systemctl daemon-reload

say "5. can we reach GitHub as the deploy key?"
# Exit status 1 with a "successfully authenticated" banner is what GitHub gives
# a working deploy key -- it never grants a shell.
auth=$(asuser_ssh -T git@github.com 2>&1 || true)
echo "   $auth"
case "$auth" in
    *successfully\ authenticated*) : ;;
    *) echo "   !! the deploy key did not authenticate. Fix that before enabling the timer."; exit 1 ;;
esac

# WHICH key authenticated, not merely that one did. GitHub greets a deploy key
# as "Hi owner/repo!" and an account-level key as "Hi owner!". Both push
# perfectly, so the kind of key in use is otherwise invisible -- and it is worth
# printing, because it is the fact that decides this box's blast radius on
# GitHub. It is reported, never acted on.
#
# OWNER RULING 2026-08-02: an ACCOUNT-LEVEL KEY ON A DEDICATED ACCOUNT, not a
# deploy key. This script's header used to instruct otherwise; the ruling wins.
# The reasoning is recorded in app/APP.md's decision log -- in short, the scoping
# a deploy key buys is already provided by the account itself, which exists for
# this and holds nothing else. Do NOT "fix" this back to a deploy key, and do
# not make it a warning: it is a decision, and it has been made.
greeting=$(printf '%s' "$auth" | sed -n 's/.*\(Hi [^!]*\)!.*/\1/p')
case "$greeting" in
    */*) echo "   OK -- '$greeting' is a deploy key, scoped to that one repository" ;;
    *)   echo "   OK -- '$greeting' is an account-level key on the dedicated backup"
         echo "        account (owner ruling 2026-08-02). Its reach is that account's"
         echo "        repositories, which is why the account holds nothing else." ;;
esac

say "6. the remote must be EMPTY the first time"
# If GitHub created the repository with a README, the first push is a
# non-fast-forward and fails -- at 03:17, into the journal, where nobody is
# looking. Better to refuse here, with a human present.
remote_refs=$(asuser git ls-remote "$REMOTE" 2>/dev/null | wc -l)
local_commits=$(sudo -u logbook git -C "$REPO_DIR" rev-list --count HEAD 2>/dev/null || echo 0)
if [ "$remote_refs" -gt 0 ] && [ "$local_commits" -eq 0 ]; then
    echo "   !! $REMOTE already has $remote_refs ref(s) but this box has never pushed to it."
    echo "   !! Create the repository with NO README and NO licence, or pull its"
    echo "   !! history here deliberately first. Stopping before the timer is enabled."
    exit 1
fi
echo "   remote refs: $remote_refs, local commits: $local_commits -- ok"

say "7. first run, now, so a failure is seen by a human rather than at 03:17"
systemctl start logbook-backup.service
# `systemctl status` EXITS 3 for a oneshot that has finished -- "inactive (dead)"
# is its success state, not a fault. Under `set -euo pipefail` that exit killed
# this script on 2026-08-02 at exactly this line: the backup had run, committed
# and PUSHED, and the run still aborted before the clone-back check and before
# the timer was enabled. The owner saw a wall of successful output and no timer.
# So: never let a status display decide control flow...
systemctl --no-pager --full status logbook-backup.service 2>&1 | head -20 || true
# ...and ask the question that actually matters separately, of a property with a
# defined value rather than of a human-readable page.
result=$(systemctl show logbook-backup.service -p Result --value)
if [ "$result" != "success" ]; then
    echo "   !! the backup run did not succeed (Result=$result). Timer NOT enabled."
    echo "   !! journalctl -u logbook-backup.service -n 40"
    exit 1
fi
echo "   the run reports Result=success"

say "8. CLONE IT BACK -- the only check that proves a backup exists"
# This is the step that would have caught the branch-name bug that made every
# push succeed while every clone came back empty. A push that reports success
# is evidence about the push. Only a clone is evidence about the BACKUP.
CHECK=$(mktemp -d)
trap 'rm -rf "$CHECK"' EXIT
# The clone runs as the logbook user (it is the only account holding the key),
# but mktemp -d here runs as ROOT and makes the directory 0700 root:root -- so
# the clone died with "could not create work tree dir ... Permission denied".
# Hand the directory to logbook. Root can still read and rm -rf it afterwards.
chown logbook:logbook "$CHECK"
asuser git clone -q "$REMOTE" "$CHECK/clone"
missing=0
for f in logbook.db logbook.csv MANIFEST.txt RESTORE.md; do
    if [ -s "$CHECK/clone/$f" ]; then
        echo "   $f  $(stat -c '%s bytes' "$CHECK/clone/$f")"
    else
        echo "   !! $f is MISSING from a fresh clone"; missing=1
    fi
done
if [ "$missing" -ne 0 ]; then
    echo "   !! THE PUSHED BACKUP IS NOT RESTORABLE. Timer NOT enabled."
    echo "   !!"
    echo "   !! If the clone warned 'remote HEAD refers to nonexistent ref', the"
    echo "   !! remote's default branch is not the branch we pushed ('$BRANCH')."
    echo "   !! Every push would succeed and every clone would come back empty."
    echo "   !! Fix it on GitHub -- Settings -> General -> Default branch -- so it"
    echo "   !! reads '$BRANCH', then run this script again."
    exit 1
fi

# And the bytes that came back must be the bytes the manifest promised.
want=$(awk '/logbook\.db$/ {print $2}' "$CHECK/clone/MANIFEST.txt")
got=$(sha256sum "$CHECK/clone/logbook.db" | cut -d' ' -f1)
if [ "$want" != "$got" ]; then
    echo "   !! the cloned database does not match its own manifest"
    echo "   !!   manifest $want"
    echo "   !!   clone    $got"
    exit 1
fi
echo "   sha256 matches the manifest: ${got:0:16}…"
grep -E '^(flights|total time|landings|users)' "$CHECK/clone/MANIFEST.txt" | sed 's/^/   /'

say "9. enable the timer"
systemctl enable --now logbook-backup.timer
systemctl list-timers logbook-backup.timer --no-pager

echo
echo "Backup installed, pushed, and cloned back verified."
echo
echo "A backup nobody has ever restored FROM is still a backup nobody should"
echo "trust. Read RESTORE.md in that clone at least once, while there is no"
echo "emergency:  git clone $REMOTE"

#!/bin/bash
# The daily off-box backup: snapshot -> commit -> push.
#
# WHY THIS RUNS AT ALL. Until 2026-08-02 the production database was
# reconstructible from the project repository -- three committed CSVs and one
# command. Then the transcription effort closed, the app became the only way the
# record grows, and the owner logged flights from an airfield. Those rows are in
# no CSV. The pre-import backups under /var/lib/logbook/backups/ sit on the same
# disk as the database they protect, which protects against a bad import and
# against nothing else. This script is what makes a copy exist on another
# machine.
#
# It runs as the `logbook` service user, not as root: it reads the database it
# already owns and writes into a directory it already owns. Nothing here needs
# privilege, so nothing here has it.
#
# THE SERVICE KEEPS RUNNING THROUGHOUT. logbookctl backup takes its snapshot
# with VACUUM INTO, which is transactionally consistent against a live database
# in WAL mode. A backup that required downtime would be a backup that gets
# skipped.
set -euo pipefail

DB=${LOGBOOK_DB:-/var/lib/logbook/logbook.db}
REPO=${LOGBOOK_BACKUP_REPO:-/var/lib/logbook/backup}
CTL=${LOGBOOK_CTL:-/opt/logbook/logbookctl}
SSH_DIR=${LOGBOOK_SSH_DIR:-/var/lib/logbook/.ssh}

say() { printf '== %s\n' "$*"; }

# The deploy key never appears in this repository or in any log (rule 0.3).
# IdentitiesOnly stops ssh from offering some other agent key first;
# StrictHostKeyChecking=yes with our own known_hosts means a substituted GitHub
# host key fails the push rather than silently trusting a new one.
export GIT_SSH_COMMAND="ssh -i $SSH_DIR/backup_ed25519 -o IdentitiesOnly=yes \
-o StrictHostKeyChecking=yes -o UserKnownHostsFile=$SSH_DIR/known_hosts \
-o PasswordAuthentication=no -o BatchMode=yes"

[ -f "$DB" ]      || { echo "backup: no database at $DB" >&2; exit 1; }
[ -d "$REPO/.git" ] || { echo "backup: $REPO is not a git repository -- run install-backup.sh" >&2; exit 1; }

say "1. snapshot and verify"
# logbookctl refuses and writes nothing if the copy disagrees with the live
# database, so a failure here leaves yesterday's good backup in place and
# `set -e` stops before anything is committed.
"$CTL" backup -db "$DB" -out "$REPO"

cd "$REPO"

say "2. commit"
# The figures go in the subject line so that `git log --oneline` in the backup
# repository reads as a history of the logbook itself -- which is the view you
# want on the day you are trying to work out what you lost and when.
FLIGHTS=$(awk '$1=="flights"          {print $2; exit}' MANIFEST.txt)
APPROWS=$(awk '$1=="hand-entered"     {print $2; exit}' MANIFEST.txt)
TOTAL=$( awk '$1=="total" && $2=="time" {print $3; exit}' MANIFEST.txt)
LDG=$(   awk '$1=="landings"          {print $2; exit}' MANIFEST.txt)

git add -A
# ALWAYS commit, even when nothing was flown. The database and the CSV are
# byte-identical when the data has not changed, so git stores one blob and the
# only new bytes are the manifest's timestamp -- a few hundred a day. What that
# buys is a heartbeat: a commit dated yesterday proves the backup ran, and
# "nothing changed" and "the job has been silently failing for a month" stop
# looking identical in the log.
git commit --allow-empty -q -m "Backup $(date -u +%Y-%m-%d): ${FLIGHTS} flights, ${TOTAL}, ${LDG} landings" \
    -m "${APPROWS} entered in the app and present in no CSV.
Verified against the live database before writing. See MANIFEST.txt and RESTORE.md."

say "3. push"
# An EXPLICIT refspec, onto the branch the REMOTE calls its default.
#
# Rehearsing this caught the worst bug in the whole task. The push succeeded and
# a fresh clone came back EMPTY -- "remote HEAD refers to nonexistent ref,
# unable to checkout" -- because the remote's HEAD named a branch we had never
# pushed to. Every run reported success, every day, and the copy was unusable at
# exactly the moment it was wanted.
#
# The branch is discovered from the remote by install-backup.sh and written to
# this repository's own config, so the two cannot drift apart. `main` is only
# the fallback for a remote that has no opinion yet.
BRANCH=$(git config --get logbook.backupBranch || echo main)
git push --quiet origin "HEAD:refs/heads/$BRANCH"
say "done: $(git log --oneline -1) -> $BRANCH"

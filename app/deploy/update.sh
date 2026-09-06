#!/bin/bash
# SUPERSEDED 2026-09-06. The privileged half of a deploy now lives in
# `logbook-apply`, installed to /opt/logbook/logbook-apply and runnable by rami
# through one scoped NOPASSWD sudoers rule -- so a deploy no longer needs a
# human typing a root password (owner ruling: "deploy should not need root").
#
# This file is kept ONLY because the box's staging directory and older notes
# still reference it. It holds no logic of its own: one implementation, so the
# two paths cannot drift.
#
#   From the dev machine (the normal way):  app/deploy/deploy.sh
#   On the box, as root (the manual way):   sudo /opt/logbook/logbook-apply
set -euo pipefail
echo "update.sh is superseded -- running /opt/logbook/logbook-apply instead."
echo "From the dev machine, use app/deploy/deploy.sh, which needs no password."
exec /opt/logbook/logbook-apply

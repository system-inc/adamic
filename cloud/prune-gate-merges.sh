#!/usr/bin/env bash
# Deletes old gate merges from origin (#6f3b76x). Every gate of a tip that doesn't hold its base pushes the merge it
# tests to refs/gate-merges/<sha> (gateMerge, cloud/fast-gate-classify.sh), and nothing else removes them: at a hundred
# gates an hour they slow every ls-remote and fetch. A merge is old when its committer date, its base's own date set at
# gate time, is over 7 days back; one main contains (push-main landed that exact merge) is kept with its history either
# way, but its ref goes too. Run from a checkout with a push credential; the watcher runs it once a day:
#
#   cloud/prune-gate-merges.sh [--days N] [--dry-run]
set -uo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
days=7 dry=""
while [ $# -gt 0 ]; do
  case $1 in
    --days) days=$2; shift 2 ;;
    --dry-run) dry=1; shift ;;
    *) echo "unknown argument $1" >&2; exit 2 ;;
  esac
done
git -C "${here}" fetch -q --prune origin '+refs/gate-merges/*:refs/prune-gate-merges/*' || exit 1
cutoff=$(( $(date -u +%s) - days * 86400 ))
old=$(git -C "${here}" for-each-ref --format='%(committerdate:unix) %(refname:strip=2)' refs/prune-gate-merges/ | awk -v cutoff="${cutoff}" '$1 < cutoff {print "refs/gate-merges/" $2}')
count=$(printf '%s\n' "${old}" | grep -c .)
echo "gate merges older than ${days} days: ${count}"
[ "${count}" -gt 0 ] || exit 0
[ -n "${dry}" ] && { printf '%s\n' "${old}"; exit 0; }
# One push per hundred refs, so a long backlog never builds a command line too long for the shell.
printf '%s\n' "${old}" | sed 's/^/:/' | xargs -n 100 git -C "${here}" push -q origin && echo "deleted ${count}"

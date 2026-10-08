#!/usr/bin/env bash
# The contribution clock, by itself: every hour it dispatches one small, real change to a fresh Codex
# worker and times it from dispatch (t0) to the worker's push (t1) to the fast gate's verdict (t2),
# appending a row to documentation/velocity/contribution-clock.csv on the branch
# records/contribution-clock. Run it from a checkout with a push credential and `ahra`:
#
#   cloud/contribution-clock.sh            # loop: one dispatch an hour
#   cloud/contribution-clock.sh --once     # one dispatch, then exit
#
# Briefs are files in cloud/contribution-clock/, used in name order, each once; the watcher
# (cloud/fast-gate-watch.sh) gates the pushed branch like any worker's, and its log gives t2.
set -uo pipefail

here=$(cd "$(dirname "$0")/.." && pwd)
state=${ADAMIC_CONTRIBUTION_CLOCK_STATE:-${HOME}/.adamic-contribution-clock}
watchLog=${ADAMIC_FAST_GATE_WATCH_LOG:-${HOME}/Projects/system/adamic-gate-logs/fast-gate-watch.log}
ahraDirectory=/Users/kirkouimet/Projects/ahra
branch=records/contribution-clock
csv=documentation/velocity/contribution-clock.csv
mkdir -p "${state}"
touch "${state}/used"

record() {
  local row=$1 records=${state}/records
  if [ ! -d "${records}" ]; then
    git -C "${here}" fetch -q origin
    if git -C "${here}" ls-remote --exit-code origin "refs/heads/${branch}" > /dev/null; then
      git -C "${here}" worktree add -q --detach "${records}" "origin/${branch}"
    else
      git -C "${here}" worktree add -q --detach "${records}" "$(git -C "${here}" commit-tree "$(git -C "${here}" hash-object -t tree /dev/null)" -m "Start the contribution clock record")"
    fi
  fi
  git -C "${records}" fetch -q origin
  git -C "${records}" ls-remote --exit-code origin "refs/heads/${branch}" > /dev/null && git -C "${records}" switch -q --detach "origin/${branch}"
  mkdir -p "$(dirname "${records}/${csv}")"
  [ -f "${records}/${csv}" ] || echo "dispatched_utc,pushed_utc,verdict_utc,brief,branch,sha,session,verdict,dispatch_to_push_seconds,push_to_verdict_seconds,total_seconds" > "${records}/${csv}"
  echo "${row}" >> "${records}/${csv}"
  git -C "${records}" add "${csv}"
  git -C "${records}" -c user.name=kirkouimet -c user.email=kirk@kirkouimet.com commit -q -m "Contribution clock: ${row%%,*}" -m "${row}" -m "Co-Authored-By: Ahra <ahra@ahra.ai>"
  git -C "${records}" push -q origin "HEAD:refs/heads/${branch}"
}

iso() { date -u -r "$1" +%FT%TZ 2>/dev/null || date -u -d "@$1" +%FT%TZ; }

dispatch() {
  local brief name number target t0 t1="" t2="" sha="" verdict="" session prompt line
  brief=$(ls "${here}"/cloud/contribution-clock/*.md 2>/dev/null | grep -vxF -f "${state}/used" | head -1)
  [ -n "${brief}" ] || { echo "$(date -u +%H:%M:%S) no unused brief left in cloud/contribution-clock/"; return 1; }
  echo "${brief}" >> "${state}/used"
  name=$(basename "${brief}" .md)
  number=$(( $(wc -l < "${state}/used") ))
  target=codex/clock-${number}-${name}
  prompt=$(mktemp)
  { cat "${brief}"; printf '\n\nHow to deliver: start from current origin/main. When the change is done and its own checks pass, push it to the branch %s (a new branch; never main, never force) and say the full sha. Keep going on ambiguity: pick the reading most consistent with this brief, write the assumption in the commit, and ask only if the answer would change what you build. Commits: author kirkouimet <kirk@kirkouimet.com>, trailer Co-Authored-By: Ahra <ahra@ahra.ai>.\n' "${target}"; } > "${prompt}"
  t0=$(date -u +%s)
  session=$(cd "${ahraDirectory}" && ahra ai start codex --directory "${here}" --label "clock-${number}" --fleet devtools-clock --prompt-file "${prompt}" 2>&1 | awk '/^[0-9a-f]{8}-/ {print $1; exit}')
  echo "$(date -u +%H:%M:%S) dispatched ${name} as ${target} (session ${session:-unknown})"
  # t1: the branch first appears on origin. Give the worker up to 90 minutes.
  while [ $(( $(date -u +%s) - t0 )) -lt 5400 ]; do
    sha=$(git -C "${here}" ls-remote origin "refs/heads/${target}" | cut -f1)
    [ -n "${sha}" ] && { t1=$(date -u +%s); break; }
    sleep 10
  done
  if [ -z "${t1}" ]; then
    record "$(iso "${t0}"),,,${name},${target},,${session},no push in 90 minutes,,,"
    return 0
  fi
  echo "$(date -u +%H:%M:%S) pushed ${target} ${sha}"
  # t2: the watcher's verdict on that branch's tip (it gates only the newest tip, so take the last line).
  while [ $(( $(date -u +%s) - t1 )) -lt 5400 ]; do
    line=$(grep -E "^[0-9:]{8} done ${target}: (green|red):" "${watchLog}" | tail -1)
    if [ -n "${line}" ]; then
      t2=$(date -u +%s)
      verdict=$(echo "${line}" | sed -E 's/.*: (green|red):.*/\1/')
      sha=$(echo "${line}" | grep -oE '[0-9a-f]{40}' | head -1)
      break
    fi
    sleep 10
  done
  record "$(iso "${t0}"),$(iso "${t1}"),${t2:+$(iso "${t2}")},${name},${target},${sha},${session},${verdict:-no verdict in 90 minutes},$(( t1 - t0 )),${t2:+$(( t2 - t1 ))},${t2:+$(( t2 - t0 ))}"
  echo "$(date -u +%H:%M:%S) ${target}: ${verdict:-no verdict}, $(( ${t2:-$t1} - t0 )) s from dispatch"
}

if [ "${1:-}" = "--once" ]; then
  dispatch
  exit
fi
while true; do
  started=$(date -u +%s)
  dispatch || true
  # One an hour, counted from each dispatch.
  sleep $(( 3600 - ( $(date -u +%s) - started ) % 3600 ))
done

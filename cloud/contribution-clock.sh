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
watchState=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
# t0 is stamped just before the clock's own start of a fresh session, never a reused session's creation.
t0Source="dispatch: fresh session via ahra ai start"
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
  local header="dispatched_utc,pushed_utc,verdict_utc,brief,branch,sha,session,verdict,dispatch_to_push_seconds,push_to_verdict_seconds,total_seconds,first_verdict,rounds,t0_source"
  # Columns are only ever added at the end, so an older row reads with its new fields empty.
  if [ -f "${records}/${csv}" ]; then
    { echo "${header}"; tail -n +2 "${records}/${csv}"; } > "${records}/${csv}.new" && mv "${records}/${csv}.new" "${records}/${csv}"
  else
    echo "${header}" > "${records}/${csv}"
  fi
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
    record "$(iso "${t0}"),,,${name},${target},,${session},no push in 90 minutes,,,,,0,${t0Source}"
    return 0
  fi
  echo "$(date -u +%H:%M:%S) pushed ${target} ${sha}"
  # t2: the watcher's verdict on the branch's newest tip. A red goes back to the worker with the gate's
  # first failure, as a reviewer would send it, and the clock keeps waiting for the next push: t2 is the
  # first green, or the last verdict after three rounds or 90 minutes. first_verdict and rounds say
  # how it got there.
  local first="" rounds=0 answered="" tip excerpt
  while [ $(( $(date -u +%s) - t1 )) -lt 5400 ]; do
    tip=$(git -C "${here}" ls-remote origin "refs/heads/${target}" | cut -f1)
    line=$(grep -E "^[0-9:]{8} done ${target}: (green|red): ${tip}" "${watchLog}" | tail -1)
    if [ -n "${tip}" ] && [ -n "${line}" ] && [ "${tip}" != "${answered}" ]; then
      answered=${tip}
      rounds=$(( rounds + 1 ))
      verdict=$(echo "${line}" | sed -E 's/.*: (green|red):.*/\1/')
      sha=${tip}
      [ -n "${first}" ] || first=${verdict}
      t2=$(date -u +%s)
      [ "${verdict}" = green ] || [ "${rounds}" -ge 3 ] && break
      excerpt=$(mktemp)
      { printf 'The fast gate on your push %s is red. Its first failure:\n\n' "${tip}"
        sed -n '/^FIRST FAILURE/,/^red:/p' "${watchState}/logs/${tip:0:12}.log" | head -40
        printf '\nFix it on the same branch %s, keep the change scoped to the brief, push, and say the new full sha.\n' "${target}"; } > "${excerpt}"
      (cd "${ahraDirectory}" && ahra ai send "${session}" --message-file "${excerpt}" > /dev/null 2>&1)
      echo "$(date -u +%H:%M:%S) ${target}: red at ${tip}, sent back to the worker (round ${rounds})"
    fi
    sleep 10
  done
  record "$(iso "${t0}"),$(iso "${t1}"),${t2:+$(iso "${t2}")},${name},${target},${sha},${session},${verdict:-no verdict in 90 minutes},$(( t1 - t0 )),${t2:+$(( t2 - t1 ))},${t2:+$(( t2 - t0 ))},${first},${rounds},${t0Source}"
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

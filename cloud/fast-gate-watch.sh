#!/usr/bin/env bash
# Every push to codex/*, area/*, devtools/* and integration's cloud/land-* gets a fast gate on the gate box, with nobody between
# the push and the gate. Run it from a checkout with a push credential (the box has none yet):
#
#   cloud/fast-gate-watch.sh
#
# It polls origin every 15 s. Branch tips it sees on its first poll are the backlog and are left
# alone; every tip that appears or moves after that is gated once (a sha already gated under another
# branch isn't gated again), one per slot, on every box in the slot table. The queue is by priority, decided
# when a gate starts: what integration is landing first (cloud/land-*, then area/* and any branch named in the state
# directory's priority file, one per line, such as a fix-forward), then devtools/*, then workers'
# codex/*, newest first within each, and only a branch's newest tip. Each tip is classed when queued:
# big (an area, a stage3/ change, which runs the stage 3 lane, or more than two touched packages) or
# small. A big gate runs in an area slot (24 CPUs), a small one in a small slot (12 each), so a
# worker's tip never waits behind an area. The slot table (${state}/slots, one "box class" per line,
# re-read every poll, with an optional third-field branch glob) says which boxes serve and how many slots of each class each has; a tip goes to
# the first box in the table with a free slot of its class, and the box picks the slot itself
# (cloud/fast-gate.sh). The gating line names the box and how long a tip waited. Each gate publishes
# gate-logs/<sha12>/<UTC stamp>/fast like any other, with the branch and, when an ai.db reply names
# the branch, the worker's session.
set -uo pipefail

here=$(cd "$(dirname "$0")/.." && pwd)
state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
mkdir -p "${state}"
# Cloud: one area slot and two small (@system_adamic, Oct 8 00:10). Workshop the same and Chonchon one
# small slot on all 16 CPUs (@system_adamic, Oct 8 04:40: 69 live tips behind two slots).
[ -s "${state}/slots" ] || printf '%s\n' "threadripper B" "threadripper S" "threadripper S" "threadripper S" > "${state}/slots"
git -C "${here}" fetch -q origin
git -C "${here}" branch -r --contains "$(git -C "${here}" rev-parse HEAD)" | grep -q . || { echo "the gate's own commit is not on origin; push it first" >&2; exit 2; }
export ADAMIC_FAST_GATE_TOOLS_ON_ORIGIN=1
# Release only this watcher's scheduling guard on an ordinary shutdown.
trap '[ "${slotTableOwned:-}" = yes ] && rm -f "${state}/slot-table.lock/holder" && rmdir "${state}/slot-table.lock" 2>/dev/null || true' EXIT

. "${here}/cloud/fast-gate-classify.sh"

# Every gate start, appended to documentation/velocity/fast-gate-waits.csv on records/fast-gate-waits
# (sha, branch, class, queued, started, waited seconds, outcome, box), so slot wait is charted from an artifact.
recordWait() {
  local row=$1 records=${state}/waits-records file=documentation/velocity/fast-gate-waits.csv
  if [ ! -d "${records}" ]; then
    if git -C "${here}" ls-remote --exit-code origin refs/heads/records/fast-gate-waits > /dev/null; then
      git -C "${here}" fetch -q origin records/fast-gate-waits && git -C "${here}" worktree add -q --detach "${records}" FETCH_HEAD
    else
      git -C "${here}" worktree add -q --detach "${records}" "$(git -C "${here}" commit-tree "$(git -C "${here}" hash-object -t tree /dev/null)" -m "Start the fast gate's slot-wait record")"
    fi
  fi
  mkdir -p "$(dirname "${records}/${file}")"
  [ -f "${records}/${file}" ] || echo "sha,branch,class,queued_utc,started_utc,waited_seconds,outcome" > "${records}/${file}"
  # outcome: "started" when a gate begins, and a second row "void:<cause>" when it ended with no real
  # verdict, so the chart never counts a void gate as served. Older rows (no column) were starts.
  head -1 "${records}/${file}" | grep -qE ',outcome(,box)?$' || sed -i '' '1s/$/,outcome/' "${records}/${file}"
  head -1 "${records}/${file}" | grep -q ',box$' || sed -i '' '1s/$/,box/' "${records}/${file}"
  echo "${row}" >> "${records}/${file}"
  git -C "${records}" add "${file}"
  git -C "${records}" -c user.name=kirkouimet -c user.email=kirk@kirkouimet.com commit -q -m "Slot wait: ${row}" -m "Co-Authored-By: Ahra <ahra@ahra.ai>"
  git -C "${records}" push -q origin "HEAD:refs/heads/records/fast-gate-waits" || true
}

# Why a finished gate has no real verdict, or nothing when it has one: no green/red line at all, or a
# red whose log shows the box failed (a missing work dir, a dropped ssh, a stale git lock, a full disk).
voidCause() {
  local log=$1
  # A gate that found its box unfit says so itself, on a line starting 'void: ': a declared tool missing
  # (cloud/fast-gate/tools.txt), a nested checkout copy in its slot.
  local said
  said=$(grep -m1 -oE '^void: .*' "${log}" 2>/dev/null | sed -E 's/^void: [0-9a-f]+ //')
  [ -n "${said}" ] && { echo "${said}"; return; }
  grep -qE '^(green|red):' "${log}" 2>/dev/null || { echo "no verdict"; return; }
  grep -qE '^red:' "${log}" || return 0
  grep -q '^stopped: ' "${log}" && return 0
  grep -oE 'creating work dir|Connection reset by|kex_exchange_identification|ssh: connect to host|Connection refused|index\.lock.: File exists|No space left on device' "${log}" | head -1
}

# Notifications are best effort, once per transition; uppercase runs are normalized because
# ahra refuses all-caps words. Preserve mixed-case paths such as /Users.
notifyStorm() {
  local text=$1 recipient
  text=$(printf '%s\n' "${text}" | awk '{
    out = ""
    while (match($0, /[A-Z][A-Z][A-Z]+/)) {
      out = out substr($0, 1, RSTART - 1) tolower(substr($0, RSTART, RLENGTH))
      $0 = substr($0, RSTART + RLENGTH)
    }
    print out $0
  }')
  for recipient in system_adamic_developer_tools system_adamic_integration; do
    (cd "${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}" &&
      ahra os send "${recipient}" "${text}" --from system_adamic_developer_tools) || true
  done
}
enterStorm() {
  [ -f "${state}/storm" ] && return
  printf '%s %s\n' "$(date -u +%s)" "$1" > "${state}/storm"
  # Permit an immediate recovery probe, then at most one start every ten minutes.
  rm -f "${state}/canary-started"
  notifyStorm "fast gate void storm; dispatch paused. first void log: $1
$(tail -20 "$1" 2>/dev/null)"
}
countVoid() {
  local now first
  now=$(date -u +%s)
  touch "${state}/void-window"
  awk -v now="${now}" '$1 > now - 600 {print}' "${state}/void-window" > "${state}/void-window.tmp"
  printf '%s %s\n' "${now}" "$1" >> "${state}/void-window.tmp"
  mv "${state}/void-window.tmp" "${state}/void-window"
  if [ "$(wc -l < "${state}/void-window")" -gt 5 ]; then
    first=$(head -1 "${state}/void-window" | cut -d ' ' -f2-)
    enterStorm "${first}"
  fi
}
# Worker and canary gates share slot selection and launch. Canary logs are separate
# even when a queued worker has main's sha, and carry the tools version they test.
dispatch() {
  local branch=$1 sha=$2 slot=$3 box=$4 class=$5 log=$6 started whole="" script=${here}/cloud/fast-gate.sh token=${canaryToken}
  started=$(date -u +%s)
  lastDispatch=${started} stallAlarmed=""
  slotReserved "${branch}" "${box}" "${slot}" && whole=--whole-box
  # Staged tools: while new tools wait for their first real green, only the canary box runs them.
  if staging && [ "${box}" != "$(cat "${state}/canary-box")" ]; then
    script=${goodTree}/cloud/fast-gate.sh token="$(cat "${state}/tools-good"):good"
  fi
  ADAMIC_FAST_GATE_BOX=${box} bash "${script}" "${sha}" --branch "${branch}" --class "${slot}" ${whole} > "${log}" 2>&1 &
  local pid=$!
  echo "${started}" > "${state}/running-started/${pid}"
  echo "${branch} ${sha} ${slot} ${box} ${class} ${token} ${log}" > "${state}/running/${pid}"
  if slotReserved "${branch}" "${box}" "${slot}"; then
    echo "${box}" > "${state}/reserved-running/${pid}"
  fi
}
# Staged rollout of the gate tools (@system_adamic, Oct 8, after three deploy incidents): with a box named
# in ${state}/canary-box, new tools run on that box only, while every other box gates with the last good
# tools (${state}/tools-good, checked out in their own worktree); the first real tip gated green on the
# canary box with the new tools promotes them everywhere. No canary-box file: one tools version, as before.
goodTree=${ADAMIC_FAST_GATE_GOOD_TREE:-${state}/tools-good-tree}
staging() {
  [ -s "${state}/canary-box" ] && [ -s "${state}/tools-good" ] && [ "$(cat "${state}/tools-good")" != "${toolsHead}" ] &&
    [ -f "${goodTree}/cloud/fast-gate.sh" ]
}
placeGoodTree() {
  local good
  good=$(cat "${state}/tools-good")
  [ -n "${ADAMIC_FAST_GATE_GOOD_TREE:-}" ] && return 0
  if [ -d "${goodTree}" ]; then
    git -C "${goodTree}" switch -q --detach "${good}"
  else
    git -C "${here}" worktree add -q --detach "${goodTree}" "${good}"
  fi
}
freeSlots() {
  cat "${state}"/running/* 2>/dev/null > "${state}/running.tmp"
  awk 'FILENAME == ARGV[1] { total[$1 " " $2]++; if (!($1 " " $2 in order)) { order[$1 " " $2] = ++n; keys[n] = $1 " " $2 }; next }
       { used[($4 == "" ? "threadripper" : $4) " " $3]++ }
       END { for (i = 1; i <= n; i++) if (total[keys[i]] > used[keys[i]]) print keys[i] }' "${state}/slots" "${state}/running.tmp"
}

# The first-step file reserves Server's B slots dynamically; explicit third-field
# globs reserve individual slots on any box. Other slots stay available when the
# reserved candidate is absent. A running reservation owns its box until reaped.
# Explicit table reservations plus the current roadmap's Server area reservation.
# Snapshot once per poll; the refresh publishes atomically in the background.
reservationGlobs() {
  awk 'NF >= 3 {print $3}' "${state}/slots"
  cat "${state}/first-step-globs.poll" 2>/dev/null || true
}
reservedBranch() {
  local branch=$1 glob
  while read -r glob; do
    [ -n "${glob}" ] && [[ ${branch} == ${glob} ]] && return 0
  done <<< "$(reservationGlobs)"
  return 1
}
slotReserved() {
  local branch=$1 box=$2 slot=$3 b c glob
  while read -r b c glob; do
    [ "${b}" = "${box}" ] && [ "${c}" = "${slot}" ] || continue
    if [ -n "${glob}" ]; then
      [[ ${branch} == ${glob} ]] && return 0
    elif [ "${b}" = server ] && [ "${c}" = B ]; then
      while read -r glob; do
        [ -n "${glob}" ] && [[ ${branch} == ${glob} ]] && return 0
      done < "${state}/first-step-globs.poll"
    fi
  done < "${state}/slots"
  return 1
}
# Both tips must match the same reservation glob on the running gate's slot.
sameReservedFamily() {
  local older=$1 newer=$2 box=$3 slot=$4 b c glob
  while read -r b c glob; do
    [ "${b}" = "${box}" ] && [ "${c}" = "${slot}" ] || continue
    if [ -n "${glob}" ]; then
      [[ ${older} == ${glob} && ${newer} == ${glob} ]] && return 0
    elif [ "${b}" = server ] && [ "${c}" = B ]; then
      while read -r glob; do
        [ -n "${glob}" ] && [[ ${older} == ${glob} && ${newer} == ${glob} ]] && return 0
      done < "${state}/first-step-globs.poll"
    fi
  done < "${state}/slots"
  return 1
}
# Stops one running gate on its box: run.py reads the reason from beside its out directory on SIGTERM,
# publishes what it had and exits (a red with its failures, or void before any).
stopGate() {
  local pid=$1 branch=$2 sha=$3 box=$4 reason=$5
  if ssh "${box}" bash -s -- "$(printf '%q ' "${sha}" "${reason}")" <<'STOP'
set -eu
sha=$1 reason=$2
for out in ~/fast-gate/out/"${sha:0:12}"-*; do
  [ -d "${out}" ] || continue
  printf '%s\n' "${reason}" > "${out}.stop-reason"
done
pkill -TERM -f "run.py .*--sha ${sha}"
STOP
  then
    touch "${state}/stopped-running/${pid}"
    return 0
  fi
  return 1
}
# Integration's skip list names tips that won't land (superseded candidates): one still running there is
# stopped, so its slot goes to work that can (integration asked twice by hand, Oct 8).
stopSkipped() {
  local file pid branch sha slot box rest
  [ -s "${state}/skip" ] || return 0
  for file in "${state}"/running/*; do
    [ -f "${file}" ] || continue
    pid=$(basename "${file}")
    [ -f "${state}/stopped-running/${pid}" ] && continue
    kill -0 "${pid}" 2> /dev/null || continue
    read -r branch sha slot box rest < "${file}"
    [ "${branch}" = canary/main ] && continue
    grep -qxF "${branch} ${sha}" "${state}/skip" || continue
    stopGate "${pid}" "${branch}" "${sha}" "${box:-threadripper}" "on the skip list" &&
      echo "$(date -u +%H:%M:%S) stopped ${branch} ${sha}: on the skip list"
  done
}
# A reserved run stops the moment a newer candidate of its family is queued, red or still clean: the newer
# one carries the older's commits and the fix, and main's full gate must never see a candidate it would
# fail (@system_adamic, Oct 8; it first stopped only runs already red). Its stop publishes what it had.
stopStaleRed() {
  local file pid branch sha slot box rest log failure started class queued newer newerSha reason
  for file in "${state}"/running/*; do
    [ -f "${file}" ] || continue
    pid=$(basename "${file}")
    [ -f "${state}/reserved-running/${pid}" ] || continue
    [ -f "${state}/stopped-running/${pid}" ] && continue
    kill -0 "${pid}" 2>/dev/null || continue
    read -r branch sha slot box rest < "${file}"
    reservedBranch "${branch}" && slotReserved "${branch}" "${box}" "${slot}" || continue
    log=${state}/logs/${sha:0:12}.log
    failure=$(sed -nE 's/^FIRST FAILURE \((.*), at ([0-9.]+) s\):.*/\1 at \2 s/p' "${log}" 2>/dev/null | head -1)
    # Older watcher versions have only the pid file; its mtime records dispatch.
    started=$(cat "${state}/running-started/${pid}" 2>/dev/null ||
      python3 -c 'import os, sys; print(int(os.stat(sys.argv[1]).st_mtime))' "${file}") || continue
    while read -r class queued newer newerSha; do
      [ "${newerSha}" != "${sha}" ] && [ "${queued}" -gt "${started}" ] || continue
      reservedBranch "${newer}" && sameReservedFamily "${branch}" "${newer}" "${box}" "${slot}" || continue
      reason="superseded by ${newer} ${newerSha}"
      if stopGate "${pid}" "${branch}" "${sha}" "${box}" "${reason}"; then
        echo "$(date -u +%H:%M:%S) stopped ${branch} ${sha}: ${reason}, $([ -n "${failure}" ] && echo "red since ${failure}" || echo "clean so far")"
      fi
      break
    done < "${state}/queue"
  done
}
# A complete run (a landing or an area) is red the moment its first failure is in its log, though it runs
# on for its whole list: that red goes to gate-logs/<sha12>/<stamp>/fast-first-failure and to its owners
# at once (the witness, Oct 8: landing reds sat on Kirk's Mac for half an hour). Once a gate.
publishEarlyRed() {
  local file pid branch sha slot box rest log stamp ref out tree commit index
  for file in "${state}"/running/*; do
    [ -f "${file}" ] || continue
    pid=$(basename "${file}")
    [ -f "${state}/early-red/${pid}" ] && continue
    read -r branch sha slot box rest < "${file}"
    [[ ${branch} == cloud/land-* || ${branch} == area/* ]] || continue
    log=${state}/logs/${sha:0:12}.log
    grep -q '^FIRST FAILURE ' "${log}" 2> /dev/null || continue
    touch "${state}/early-red/${pid}"
    (
      stamp=$(/bin/date -u +%Y%m%dT%H%M%SZ)
      ref=gate-logs/${sha:0:12}/${stamp}/fast-first-failure
      out=$(mktemp -d)
      sed -nE 's/^FIRST FAILURE \(([^,]+), at ([0-9.]+) s\):.*/red: '"${sha}"' fast gate, first failure at \1 after \2 s (the complete run goes on; its whole record publishes as fast when it ends)/p' "${log}" | head -1 > "${out}/status.txt"
      sed -n '/^FIRST FAILURE /,/^\(green\|red\|void\): /p' "${log}" | head -400 > "${out}/first-failure.txt"
      index=$(mktemp -u)
      tree=$(cd "${out}" && GIT_INDEX_FILE=${index} git --git-dir="$(git -C "${here}" rev-parse --absolute-git-dir)" --work-tree=. add -A -f . && GIT_INDEX_FILE=${index} git --git-dir="$(git -C "${here}" rev-parse --absolute-git-dir)" write-tree)
      commit=$(git -C "${here}" commit-tree "${tree}" -m "First failure of ${branch} ${sha}: $(cat "${out}/status.txt")")
      git -C "${here}" push -q origin "${commit}:refs/heads/${ref}" && echo "$(date -u +%H:%M:%S) early red ${branch} ${sha}: ${ref}"
      python3 "${here}/cloud/verdict-notify.py" "${branch}" "${sha}" "${log}" --early "${ref}" >> "${state}/logs/verdict-notify.log" 2>&1
      rm -f "${index}" "${out}/status.txt" "${out}/first-failure.txt"
      rmdir "${out}"
    ) &
  done
}
# The position of the roadmap step whose Branches: globs name this branch (0 for the first ready step that
# declares any), or 99: the queue is the waterfall, the star's work ahead of scouts and leaves
# (@system_adamic, Oct 8). Within a step, the kinds keep their order (landings, areas, tools, workers).
# Tips the dispatcher would only discard leave the queue in one pass before ranking: already gated,
# skipped by hand or by pattern, or superseded by a newer tip of their branch. Ranking reads the whole queue
# for every pick (about 40 s at 150 queued, Oct 8 19:08Z), so each discard found there cost a full pass,
# and a run of them kept the loop from reaping finished gates for minutes.
pruneQueue() {
  [ -s "${state}/queue" ] || return 0
  touch "${state}/seen" "${state}/gated" "${state}/skip"
  local verdict line class queued branch sha glob kept=${state}/queue.pruned skipped
  awk 'FILENAME == ARGV[1] { seen[$1 " " $2] = 1; next }
       FILENAME == ARGV[2] { gated[$1] = 1; next }
       FILENAME == ARGV[3] { skip[$0] = 1; next }
       { key = $3 " " $4
         if ($4 in gated) print "gated\t" $0
         else if (key in skip) print "skip\t" $0
         else if (!(key in seen)) print "superseded\t" $0
         else print "keep\t" $0 }' "${state}/seen" "${state}/gated" "${state}/skip" "${state}/queue" > "${state}/queue.verdicts"
  : > "${kept}"
  while IFS=$'\t' read -r verdict line; do
    read -r class queued branch sha <<< "${line}"
    case "${verdict}" in
      gated) continue ;;
      skip) echo "$(date -u +%H:%M:%S) skipped by hand ${branch} ${sha}"; continue ;;
      superseded) echo "$(date -u +%H:%M:%S) superseded ${branch} ${sha}"; continue ;;
    esac
    skipped=""
    while read -r glob _; do
      [ -n "${glob}" ] && [[ ${glob} != \#* ]] && [[ ${branch} == ${glob} ]] && { skipped=${glob}; break; }
    done < <(cat "${state}/skip-globs" 2>/dev/null)
    if [ -n "${skipped}" ]; then
      echo "$(date -u +%H:%M:%S) skipped by pattern ${skipped} ${branch} ${sha}"
      continue
    fi
    echo "${line}" >> "${kept}"
  done < "${state}/queue.verdicts"
  mv "${kept}" "${state}/queue"
  rm -f "${state}/queue.verdicts"
}
# ${state}/front (a glob per line, # comments) puts a tip ahead of every roadmap step, behind only a reserved
# landing: a fix the parent ruled lands first, such as Oct 8's cloud/land-gate-speed.
stepPosition() {
  local branch=$1 position glob
  while read -r glob _; do
    [ -n "${glob}" ] && [[ ${glob} != \#* ]] && [[ ${branch} == ${glob} ]] && { echo 0; return; }
  done < <(cat "${state}/front" 2>/dev/null)
  while read -r position glob; do
    [ -n "${glob}" ] && [[ ${branch} == ${glob} ]] && { echo "${position}"; return; }
  done < "${state}/step-globs.poll"
  echo 99
}
usableSlots() {
  local branch=$1 box slot b c glob allowed total used runningBranch runningSha runningSlot runningBox rest blocked reservation
  while read -r box slot; do
    blocked=no
    for reservation in "${state}"/reserved-running/*; do
      [ -f "${reservation}" ] || continue
      [ "$(cat "${reservation}")" = "${box}" ] && blocked=yes
    done
    while read -r runningBranch runningSha runningSlot runningBox rest; do
      [ "${runningBox:-threadripper}" = "${box}" ] || continue
      slotReserved "${runningBranch}" "${box}" "${runningSlot}" && blocked=yes
    done < "${state}/running.tmp"
    # A box a queued reservation waits on drains: only that tip, in its reserved slot, starts there. The
    # deploy canary is exempt: nothing dispatches until it has a verdict, so a canary kept off the only
    # free box (the draining one) froze everything, the reserved tip included (Oct 8 17:30Z, 67 queued).
    if [ "${branch}" != canary/main ] && echo "${draining:-}" | grep -qxF "${box}" && ! slotReserved "${branch}" "${box}" "${slot}"; then
      blocked=yes
    fi
    [ "${blocked}" = no ] || continue
    total=0
    while read -r b c glob; do
      [ "${b}" = "${box}" ] && [ "${c}" = "${slot}" ] || continue
      allowed=no
      if [ -n "${glob}" ]; then
        [[ ${branch} == ${glob} ]] && allowed=yes
      elif [ "${b}" = server ] && [ "${c}" = B ] && [ -s "${state}/first-step-globs.poll" ]; then
        while read -r glob; do
          [[ ${branch} == ${glob} ]] && allowed=yes
        done < "${state}/first-step-globs.poll"
      else allowed=yes; fi
      [ "${allowed}" = yes ] && total=$((total + 1))
    done < "${state}/slots"
    used=$(awk -v b="${box}" -v c="${slot}" '($4 == "" ? "threadripper" : $4) == b && $3 == c {n++} END {print n+0}' "${state}/running.tmp")
    [ "${total}" -gt "${used}" ] || continue
    # A reserved gate starts only on an empty box, so it has the whole box.
    if slotReserved "${branch}" "${box}" "${slot}"; then
      awk -v b="${box}" '($4 == "" ? "threadripper" : $4) == b {found=1} END {exit found ? 0 : 1}' "${state}/running.tmp" && continue
    fi
    echo "${box} ${slot}"
  done <<< "${free}"
}
# Boxes holding a slot reserved for a queued tip's class. Without the drain, a reserved tip waiting
# for its box to empty never gets it: each small gate that ends there hands its slot to the next.
drainingBoxes() {
  local class queued branch sha
  while read -r class queued branch sha; do
    reservedBranch "${branch}" && reservedBoxes "${branch}" "${class}"
  done < "${state}/queue" | sort -u
}
# The boxes with a slot of this class reserved for this branch. A tip that has one runs only there: it
# waits out the drain for the whole box rather than borrowing a share of another.
reservedBoxes() {
  local branch=$1 class=$2 b c glob
  while read -r b c glob; do
    [ "${c}" = "${class}" ] && slotReserved "${branch}" "${b}" "${c}" && echo "${b}"
  done < "${state}/slots"
}

# The slot-table lock is held for one scheduling pass by this watcher or a moment by auto-area-merge.py.
# One left by a watcher killed mid-pass (a launchd restart, Oct 8 17:21Z) skipped every pass after it:
# nothing dispatched for 17 minutes. A lock no live auto-area-merge process can hold is cleared at start,
# and in a pass once it is two minutes old.
clearStaleSlotLock() {
  local age holder since
  [ -d "${state}/slot-table.lock" ] || return 0
  # The watcher's own pass writes its pid and start: a dead holder's lock is taken over at once.
  if [ -f "${state}/slot-table.lock/holder" ]; then
    read -r holder since < "${state}/slot-table.lock/holder"
    kill -0 "${holder}" 2> /dev/null && return 0
    rm -f "${state}/slot-table.lock/holder"
    rmdir "${state}/slot-table.lock" 2> /dev/null && echo "$(/bin/date -u +%H:%M:%S) took over a slot-table lock whose holder ${holder} (since ${since}) is dead"
    return 0
  fi
  # No holder file: auto-area-merge.py's moment, or a lock older than holders.
  pgrep -f "auto-area-merge" > /dev/null 2>&1 && return 0
  if [ "${1:-}" != start ]; then
    age=$(python3 -c 'import os, sys, time; print(int(time.time() - os.stat(sys.argv[1]).st_mtime))' "${state}/slot-table.lock" 2> /dev/null || echo 0)
    [ "${age}" -ge 120 ] || return 0
  fi
  rmdir "${state}/slot-table.lock" 2> /dev/null && echo "$(/bin/date -u +%H:%M:%S) cleared a slot-table lock no live process held"
}
tips() {
  git -C "${here}" ls-remote origin 'refs/heads/codex/*' 'refs/heads/area/*' 'refs/heads/devtools/*' 'refs/heads/cloud/land-*' |
    awk '{sub("refs/heads/", "", $2); print $2, $1}' | sort
}

[ -s "${state}/seen" ] || tips > "${state}/seen"
clearStaleSlotLock start
watchStart=$(date -u +%s)
touch "${state}/gated" "${state}/queue"
# Running gates are pid files (macOS bash 3.2 has no associative arrays).
mkdir -p "${state}/running" "${state}/logs" "${state}/reserved-running" "${state}/running-started" "${state}/stopped-running" "${state}/early-red"
echo "$(date -u +%H:%M:%S) watching codex/*, area/*, devtools/*, cloud/land-* (tools $(git -C "${here}" rev-parse --short HEAD))"
toolsHead=$(git -C "${here}" rev-parse HEAD)
canaryToken=${toolsHead}:$$
canaryRequired=1
[ -s "${state}/tools-good" ] || echo "${toolsHead}" > "${state}/tools-good"
[ -s "${state}/canary-box" ] && placeGoodTree
# Staged, the other boxes keep gating on the good tools: no fleet-wide deploy barrier to wait out.
staging && canaryRequired=0
while true; do
  now=$(date -u +%s)
  if { [ -z "${firstStepRefresh:-}" ] || [ "$((now - firstStepRefresh))" -ge 300 ]; } &&
     { [ -z "${firstStepPid:-}" ] || ! kill -0 "${firstStepPid}" 2>/dev/null; }; then
    (bash "${here}/cloud/first-step-branches.sh" > "${state}/first-step-globs.tmp" &&
      mv "${state}/first-step-globs.tmp" "${state}/first-step-globs"
     bash "${here}/cloud/first-step-branches.sh" --all > "${state}/step-globs.tmp" &&
      mv "${state}/step-globs.tmp" "${state}/step-globs") &
    firstStepPid=$!
    firstStepRefresh=${now}
  fi
  cat "${state}/first-step-globs" > "${state}/first-step-globs.poll" 2>/dev/null || true
  cat "${state}/step-globs" > "${state}/step-globs.poll" 2>/dev/null || true
  currentHead=$(git -C "${here}" rev-parse HEAD)
  if [ "${currentHead}" != "${toolsHead}" ]; then
    toolsHead=${currentHead}
    canaryToken=${toolsHead}:$$
    canaryRequired=1
    rm -f "${state}/canary-started"
    if staging; then
      canaryRequired=0
      echo "$(date -u +%H:%M:%S) staging tools ${toolsHead:0:9} on $(cat "${state}/canary-box"); the other boxes stay on $(cut -c1-9 "${state}/tools-good")"
    fi
  fi
  if tips > "${state}/now.tmp" && [ -s "${state}/now.tmp" ]; then
    # New or moved tips, queued by kind: workers' branches first.
    comm -13 "${state}/seen" "${state}/now.tmp" | while read -r branch sha; do
      grep -qx "${sha}" "${state}/gated" && continue
      class=$(classify "${branch}" "${sha}")
      echo "${class} $(date -u +%s) ${branch} ${sha}" >> "${state}/queue"
      echo "$(date -u +%H:%M:%S) queued ${branch} ${sha} (${class})"
    done
    mv "${state}/now.tmp" "${state}/seen"
  fi
  stopStaleRed
  stopSkipped
  publishEarlyRed
  for file in "${state}"/running/*; do
    [ -e "${file}" ] || continue
    kill -0 "$(basename "${file}")" 2>/dev/null && continue
    stoppedOnPurpose=no
    [ -f "${state}/stopped-running/$(basename "${file}")" ] && stoppedOnPurpose=yes
    rm -f "${state}/reserved-running/$(basename "${file}")" "${state}/running-started/$(basename "${file}")" "${state}/stopped-running/$(basename "${file}")" "${state}/early-red/$(basename "${file}")"
    read -r branch sha class box original testedHead gateLog < "${file}"
    # Merge claims are not gate verdicts: a crashed dispatcher/SSH must never queue area-merge/*.
    if [[ ${branch} == area-merge/* ]]; then rm "${file}"; continue; fi
    entry=$(cat "${file}")
    # A void gate goes back to the queue as the class it was queued with, not the slot it borrowed.
    class=${original:-${class}}
    rm "${file}"
    gateLog=${gateLog:-${state}/logs/${sha:0:12}.log}
    cause=$(voidCause "${gateLog}")
    if [ "${branch}" = canary/main ]; then
      if [ -n "${cause}" ]; then
        countVoid "${gateLog}"
        enterStorm "${gateLog}"
        echo "$(date -u +%H:%M:%S) canary void: ${cause}"
      elif [ "${testedHead}" = "${canaryToken}" ]; then
        canaryRequired=0
        echo "$(date -u +%H:%M:%S) done canary: $(grep -E '^(green|red):' "${gateLog}" | tail -1)"
        if [ -f "${state}/storm" ]; then
          rm -f "${state}/storm" "${state}/void-window"
          notifyStorm "fast gate storm over; main canary returned a real verdict, dispatch resumed."
        fi
      fi
      continue
    fi
    if [ -z "${cause}" ]; then
      verdict=$(grep -E '^(green|red):' "${state}/logs/${sha:0:12}.log" | tail -1)
      echo "$(date -u +%H:%M:%S) done ${branch}: ${verdict}"
      # A real tip green on the canary box with the staged tools promotes them to every box.
      if staging && [[ ${testedHead} == "${toolsHead}:"* && ${verdict} == "green: ${sha} "* ]]; then
        echo "${toolsHead}" > "${state}/tools-good"
        placeGoodTree
        echo "$(date -u +%H:%M:%S) promoted tools ${toolsHead:0:9} to every box after ${branch} ${sha} green on ${box}"
      fi
      # The verdict reaches the branch's owner (and integration for landings and areas) as it exists.
      (python3 "${here}/cloud/verdict-notify.py" "${branch}" "${sha}" "${gateLog}" >> "${state}/logs/verdict-notify.log" 2>&1 &)
      if [[ ${branch} == codex/* && ${verdict} == "green: ${sha} "* ]] && [ -f "${state}/auto-area-merge" ]; then
        # Keep the completed gate available for reaping if durable enqueue fails.
        bash "${here}/cloud/auto-area-merge.sh" --enqueue "${branch}" "${sha}" || echo "${entry}" > "${file}"
      fi
      continue
    fi
    # A gate the watcher stopped (stale red, skip list) that had no verdict yet isn't a box problem: no
    # void count, no requeue. Counted as voids, two skip-list stops started a storm (Oct 8 18:21Z).
    if [ "${stoppedOnPurpose}" = yes ]; then
      echo "$(date -u +%H:%M:%S) stopped ${branch} ${sha} before a verdict, as asked: not a void"
      continue
    fi
    # A void gate (it died before a real verdict, like the 33 a WSL restart killed at 08:23Z on Oct 8)
    # isn't served: un-mark it, queue it again, at most three tries, then tell integration it's the box.
    countVoid "${gateLog}"
    (recordWait "${sha},${branch},${class},,$(date -u +%FT%TZ),,void:${cause// /_},${box:-threadripper}" > /dev/null 2>&1 &)
    if [ -f "${state}/storm" ]; then
      grep -vx "${sha}" "${state}/gated" > "${state}/gated.tmp"; mv "${state}/gated.tmp" "${state}/gated"
      echo "${class} $(date -u +%s) ${branch} ${sha}" >> "${state}/queue"
      echo "$(date -u +%H:%M:%S) void ${branch} ${sha}: ${cause}, held during storm"
      continue
    fi
    echo "${sha}" >> "${state}/void-tries"
    tries=$(grep -cx "${sha}" "${state}/void-tries")
    if [ "${tries}" -le 3 ]; then
      grep -vx "${sha}" "${state}/gated" > "${state}/gated.tmp"; mv "${state}/gated.tmp" "${state}/gated"
      echo "${class} $(date -u +%s) ${branch} ${sha}" >> "${state}/queue"
      echo "$(date -u +%H:%M:%S) void ${branch} ${sha}: ${cause} (try ${tries} of 3), queued again"
    else
      echo "$(date -u +%H:%M:%S) void ${branch} ${sha}: ${cause}, three tries, given up"
      (cd /Users/kirkouimet/Projects/ahra && ahra os send system_adamic_integration "Fast gate of ${branch} ${sha} died three times with no verdict (${cause}): a box problem, not the change. Log: ${state}/logs/${sha:0:12}.log on Kirk's Mac." > /dev/null 2>&1 || true)
    fi
  done
  # Shared with the dispatcher: checking free slots and recording a PID is one transaction.
  # If a merge is claiming a slot, leave scheduling to the next poll.
  clearStaleSlotLock
  if mkdir "${state}/slot-table.lock" 2>/dev/null; then
  slotTableOwned=yes
  echo "$$ $(/bin/date -u +%s)" > "${state}/slot-table.lock/holder"
  # A deploy barrier persists until this tools version gets a real main verdict.
  # A running probe is never duplicated, including across watcher restarts.
  # An old watcher's result cannot satisfy this startup's deploy barrier.
  if [ "${canaryRequired}" = 1 ] || [ -f "${state}/storm" ]; then
    now=$(date -u +%s)
    last=$(cat "${state}/canary-started" 2>/dev/null || echo 0)
    if ! grep -q '^canary/main ' "${state}"/running/* 2>/dev/null &&
       { [ ! -f "${state}/storm" ] || [ "$((now - last))" -ge 600 ]; }; then
      free=$(freeSlots) draining=$(drainingBoxes)
      # A small slot first, else any usable slot: with every small slot busy and Server draining for a
      # reservation, an S-only canary froze all dispatch behind it while Home's area slot sat idle
      # (Oct 8 16:42Z, 18 tips queued).
      read -r box canarySlot <<< "$(usableSlots canary/main | awk '$2 == "S" && small == "" {small = $1} big == "" {big = $1 " " $2} END {print (small != "" ? small " S" : big)}')"
      sha=$(git -C "${here}" ls-remote origin refs/heads/main | awk '$2 == "refs/heads/main" {print $1; exit}')
      if [ -n "${box}" ] && [ -n "${sha}" ]; then
        log=$(mktemp "${state}/logs/canary-${sha:0:12}-${now}.XXXXXX")
        dispatch canary/main "${sha}" "${canarySlot}" "${box}" S "${log}"
        echo "${now}" > "${state}/canary-started"
        echo "$(date -u +%H:%M:%S) gating canary/main ${sha} (${canarySlot} on ${box}, log ${log})"
      fi
    fi
  fi
  pruneQueue
  while [ "${canaryRequired}" = 0 ] && [ ! -f "${state}/storm" ] && [ -s "${state}/queue" ]; do
    touch "${state}/priority"
    # Free slots per box and class: the table's count less the gates running there (an entry from
    # before the table names no box: it was Cloud's).
    free=$(freeSlots) draining=$(drainingBoxes)
    [ -n "${free}" ] || break
    # Boxes where a big tip may borrow a small slot: a free small slot, and fewer than two big gates there
    # already (its area slot and one borrowed). Three big gates borrowing on one box stacked hundreds of
    # compiles and took Cloud and Workshop down (Oct 8 11:2xZ).
    # The limit is ${state}/big-per-box, read every poll so it moves by measurement: a bare number is the
    # default (2 without the file), and a "box N" line overrides it for that box (Cloud stays at 2 until
    # its memory is understood, @system_adamic, Oct 8 06:28).
    # A tip takes a free slot of its class; a big tip may also take a free small slot, but only when
    # no small tip could have it (rank 10 and up): 74 big tips waited on two area slots while four small
    # slots sat idle (Oct 8 10:56Z). It then runs on the small slot's CPUs (12; Chonchon's 16).
    # A small tip that has waited two minutes may take an idle area slot, ranked after every big tip that
    # could take it (the witness, Oct 8: small changes waited 240 to 300 s on Cloud with 61 Codex
    # running). The area slot has its own CPUs (cloud/fast-gate.sh's taskset), so it never runs beside a
    # big gate on the same CPUs.
    pollTime=$(date -u +%s)
    next=$(while read -r class queued branch sha; do
      eligible=$(usableSlots "${branch}")
      if reservedBranch "${branch}"; then
        own=$(reservedBoxes "${branch}" "${class}")
        [ -n "${own}" ] && eligible=$(echo "${eligible}" | grep -xF "$(echo "${own}" | sed "s/\$/ ${class}/")")
      fi
      classes=$(echo "${eligible}" | awk '{print $2}' | sort -u | tr '\n' ' ')
      borrowable=$(echo "${eligible}" | awk '$2 == "S" {print $1}' | while read -r b; do
        limit=$(awk -v b="${b}" 'NF == 1 && $1 ~ /^[0-9]+$/ { fallback = $1 } NF == 2 && $1 == b { own = $2 } END { print (own != "" ? own : (fallback != "" ? fallback : 2)) }' "${state}/big-per-box" 2>/dev/null || echo 2)
        [ "$(cat "${state}"/running/* 2>/dev/null | awk -v b="${b}" '($4 == "" ? "threadripper" : $4) == b && ($5 == "B" || $3 == "B")' | wc -l)" -lt "${limit}" ] && echo "${b}"
      done | head -1)
      slot=${class} extra=0
      if [[ " ${classes}" != *" ${class} "* ]]; then
        # A landing candidate (Kirk's five, Oct 8) borrows ahead of small worker tips; any other big tip after them.
        if [ "${class}" = B ] && [ -n "${borrowable}" ]; then slot=S extra=10; [[ ${branch} == cloud/land-* ]] && extra=0
        elif [ "${class}" = S ] && [[ " ${classes}" == *" B "* ]] && [ $(( pollTime - queued )) -ge 120 ]; then slot=B extra=10
        else continue; fi
      fi
      if reservedBranch "${branch}"; then rank=0; extra=0
      elif [[ ${branch} == cloud/land-* ]]; then rank=1
      elif [[ ${branch} == area/* ]] || grep -qxF "${branch}" "${state}/priority"; then rank=2
      elif [[ ${branch} == devtools/* ]]; then rank=3
      else rank=4; fi
      if [ "${slot}" = "${class}" ] || [ "${slot}" = B ]; then
        box=$(echo "${eligible}" | awk -v c="${slot}" '$2 == c && box == "" {box = $1} END {print box}')
      else box=${borrowable}; fi
      position=$(stepPosition "${branch}")
      [ "${rank}" = 0 ] && position=0
      echo "$(( position * 100 + rank + extra )) ${queued} ${branch} ${sha} ${class} ${slot} ${box}"
    done < "${state}/queue" | sort -k1,1n -k2,2nr | head -1)
    [ -n "${next}" ] || break
    read -r _ queued branch sha class slot box <<< "${next}"
    grep -vF " ${queued} ${branch} ${sha}" "${state}/queue" > "${state}/queue.tmp"; mv "${state}/queue.tmp" "${state}/queue"
    grep -qx "${sha}" "${state}/gated" && continue
    # The skip file ("branch sha" per line) takes a tip out by hand: a stale landing candidate, say.
    grep -qxF "${branch} ${sha}" "${state}/skip" 2>/dev/null && { echo "$(date -u +%H:%M:%S) skipped by hand ${branch} ${sha}"; continue; }
    # ${state}/skip-globs takes a whole family out (a glob per line, then why): codex/views-* while every tip
    # inherits a known red that only its area's next landing fixes (compiler, Oct 8).
    skippedBy=$(while read -r glob _; do [ -n "${glob}" ] && [[ ${glob} != \#* ]] && [[ ${branch} == ${glob} ]] && { echo "${glob}"; break; }; done < <(cat "${state}/skip-globs" 2>/dev/null))
    [ -n "${skippedBy}" ] && { echo "$(date -u +%H:%M:%S) skipped by pattern ${skippedBy} ${branch} ${sha}"; continue; }
    # A tip its branch has already moved past is superseded: gate the branch's newest only.
    grep -qx "${branch} ${sha}" "${state}/seen" || { echo "$(date -u +%H:%M:%S) superseded ${branch} ${sha}"; continue; }
    # A tip already on main has nothing left to prove: integration's landed cloud/land-* branches stay
    # on origin, and the first poll that watched them queued every one at landing rank (Oct 8 10:54Z).
    git -C "${here}" fetch -q origin "${sha}" "+refs/heads/main:refs/remotes/origin/main" 2>/dev/null
    if git -C "${here}" merge-base --is-ancestor "${sha}" refs/remotes/origin/main 2>/dev/null; then
      echo "$(date -u +%H:%M:%S) already on main ${branch} ${sha}"
      continue
    fi
    echo "${sha}" >> "${state}/gated"
    log=${state}/logs/${sha:0:12}.log
    dispatch "${branch}" "${sha}" "${slot}" "${box}" "${class}" "${log}"
    now=$(date -u +%s)
    echo "$(date -u +%H:%M:%S) gating ${branch} ${sha} (${class}$([ "${slot}" = "${class}" ] || echo " in an ${slot} slot") on ${box}, waited $(( now - queued )) s, log ${log})"
    (recordWait "${sha},${branch},${class},$(date -u -r "${queued}" +%FT%TZ),$(date -u -r "${now}" +%FT%TZ),$(( now - queued )),started,${box}" > /dev/null 2>&1 &)
  done
  rm -f "${state}/slot-table.lock/holder"
  rmdir "${state}/slot-table.lock"
  slotTableOwned=""
  fi
  # An alarm, once a stall: a queue and no dispatch anywhere for five minutes with a slot free (fifteen
  # with every slot busy, since long gates can hold them all), to developer tools and integration.
  pollEnd=$(date -u +%s)
  # A void storm pauses dispatch on purpose and sends its own alarm.
  if [ -s "${state}/queue" ] && [ ! -f "${state}/storm" ]; then
    # Slots a tip could take now: a box a reserved gate holds whole, or one draining for it, isn't free.
    free=$(freeSlots) draining=$(drainingBoxes)
    freeCount=$(usableSlots canary/main | grep -c .)
    stallLimit=900
    [ "${freeCount}" -gt 0 ] && stallLimit=300
    if [ $(( pollEnd - ${lastDispatch:-${watchStart}} )) -ge "${stallLimit}" ] && [ -z "${stallAlarmed:-}" ]; then
      stallAlarmed=1
      notifyStorm "fast gate stall: nothing dispatched for $(( pollEnd - ${lastDispatch:-${watchStart}} )) s with $(grep -c . "${state}/queue") queued and ${freeCount} slots free (tools ${toolsHead:0:8}$([ "${canaryRequired}" = 1 ] && echo ', deploy canary not yet run')$([ -d "${state}/slot-table.lock" ] && echo ', slot-table lock held')). queue head: $(head -3 "${state}/queue" | awk '{print $3}' | tr '\n' ' ')"
      echo "$(date -u +%H:%M:%S) stall alarm sent"
    fi
  fi
  # A durable queue survives restarts. The dispatcher and area-merge's own locks serialize merges;
  # a locked area stays queued, and removing the switch prevents the next attempt.
  if [ -f "${state}/auto-area-merge" ]; then
    if [ -z "${areaMergePid:-}" ] || ! kill -0 "${areaMergePid}" 2>/dev/null; then
      bash "${here}/cloud/auto-area-merge.sh" --drain >> "${state}/logs/auto-area-merge.log" 2>&1 &
      areaMergePid=$!
    fi
  fi
  sleep 15
done

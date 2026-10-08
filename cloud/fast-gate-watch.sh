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
# re-read every poll) says which boxes serve and how many slots of each class each has; a tip goes to
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
trap '[ "${slotTableOwned:-}" = yes ] && rmdir "${state}/slot-table.lock" 2>/dev/null || true' EXIT

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
  # A gate whose box lacked a declared tool says so itself (cloud/fast-gate/tools.txt).
  local lacks
  lacks=$(grep -m1 -oE '^void: .*lacks a declared tool: .*' "${log}" 2>/dev/null | sed -E 's/^void: [0-9a-f]+ //')
  [ -n "${lacks}" ] && { echo "${lacks}"; return; }
  grep -qE '^(green|red):' "${log}" 2>/dev/null || { echo "no verdict"; return; }
  grep -qE '^red:' "${log}" || return 0
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
  local branch=$1 sha=$2 slot=$3 box=$4 class=$5 log=$6
  ADAMIC_FAST_GATE_BOX=${box} bash "${here}/cloud/fast-gate.sh" "${sha}" --branch "${branch}" --class "${slot}" > "${log}" 2>&1 &
  echo "${branch} ${sha} ${slot} ${box} ${class} ${canaryToken} ${log}" > "${state}/running/$!"
}
freeSlots() {
  cat "${state}"/running/* 2>/dev/null > "${state}/running.tmp"
  awk 'FILENAME == ARGV[1] { total[$1 " " $2]++; if (!($1 " " $2 in order)) { order[$1 " " $2] = ++n; keys[n] = $1 " " $2 }; next }
       { used[($4 == "" ? "threadripper" : $4) " " $3]++ }
       END { for (i = 1; i <= n; i++) if (total[keys[i]] > used[keys[i]]) print keys[i] }' "${state}/slots" "${state}/running.tmp"
}

tips() {
  git -C "${here}" ls-remote origin 'refs/heads/codex/*' 'refs/heads/area/*' 'refs/heads/devtools/*' 'refs/heads/cloud/land-*' |
    awk '{sub("refs/heads/", "", $2); print $2, $1}' | sort
}

[ -s "${state}/seen" ] || tips > "${state}/seen"
touch "${state}/gated" "${state}/queue"
# Running gates are pid files (macOS bash 3.2 has no associative arrays).
mkdir -p "${state}/running" "${state}/logs"
echo "$(date -u +%H:%M:%S) watching codex/*, area/*, devtools/*, cloud/land-* (tools $(git -C "${here}" rev-parse --short HEAD))"
toolsHead=$(git -C "${here}" rev-parse HEAD)
canaryToken=${toolsHead}:$$
canaryRequired=1
while true; do
  currentHead=$(git -C "${here}" rev-parse HEAD)
  if [ "${currentHead}" != "${toolsHead}" ]; then
    toolsHead=${currentHead}
    canaryToken=${toolsHead}:$$
    canaryRequired=1
    rm -f "${state}/canary-started"
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
  for file in "${state}"/running/*; do
    [ -e "${file}" ] || continue
    kill -0 "$(basename "${file}")" 2>/dev/null && continue
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
      if [[ ${branch} == codex/* && ${verdict} == "green: ${sha} "* ]] && [ -f "${state}/auto-area-merge" ]; then
        # Keep the completed gate available for reaping if durable enqueue fails.
        bash "${here}/cloud/auto-area-merge.sh" --enqueue "${branch}" "${sha}" || echo "${entry}" > "${file}"
      fi
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
  if mkdir "${state}/slot-table.lock" 2>/dev/null; then
  slotTableOwned=yes
  # A deploy barrier persists until this tools version gets a real main verdict.
  # A running probe is never duplicated, including across watcher restarts.
  # An old watcher's result cannot satisfy this startup's deploy barrier.
  if [ "${canaryRequired}" = 1 ] || [ -f "${state}/storm" ]; then
    now=$(date -u +%s)
    last=$(cat "${state}/canary-started" 2>/dev/null || echo 0)
    if ! grep -q '^canary/main ' "${state}"/running/* 2>/dev/null &&
       { [ ! -f "${state}/storm" ] || [ "$((now - last))" -ge 600 ]; }; then
      box=$(freeSlots | awk '$2 == "S" {print $1; exit}')
      sha=$(git -C "${here}" ls-remote origin refs/heads/main | awk '$2 == "refs/heads/main" {print $1; exit}')
      if [ -n "${box}" ] && [ -n "${sha}" ]; then
        log=$(mktemp "${state}/logs/canary-${sha:0:12}-${now}.XXXXXX")
        dispatch canary/main "${sha}" S "${box}" S "${log}"
        echo "${now}" > "${state}/canary-started"
        echo "$(date -u +%H:%M:%S) gating canary/main ${sha} (S on ${box}, log ${log})"
      fi
    fi
  fi
  while [ "${canaryRequired}" = 0 ] && [ ! -f "${state}/storm" ] && [ -s "${state}/queue" ]; do
    touch "${state}/priority"
    # Free slots per box and class: the table's count less the gates running there (an entry from
    # before the table names no box: it was Cloud's).
    free=$(freeSlots)
    [ -n "${free}" ] || break
    classes=$(echo "${free}" | awk '{print $2}' | sort -u | tr '\n' ' ')
    # Boxes where a big tip may borrow a small slot: a free small slot, and fewer than two big gates there
    # already (its area slot and one borrowed). Three big gates borrowing on one box stacked hundreds of
    # compiles and took Cloud and Workshop down (Oct 8 11:2xZ).
    # The limit is ${state}/big-per-box, read every poll so it moves by measurement: a bare number is the
    # default (2 without the file), and a "box N" line overrides it for that box (Cloud stays at 2 until
    # its memory is understood, @system_adamic, Oct 8 06:28).
    borrowable=$(echo "${free}" | awk '$2 == "S" {print $1}' | while read -r b; do
      limit=$(awk -v b="${b}" 'NF == 1 && $1 ~ /^[0-9]+$/ { fallback = $1 } NF == 2 && $1 == b { own = $2 } END { print (own != "" ? own : (fallback != "" ? fallback : 2)) }' "${state}/big-per-box" 2>/dev/null || echo 2)
      [ "$(cat "${state}"/running/* 2>/dev/null | awk -v b="${b}" '($4 == "" ? "threadripper" : $4) == b && ($5 == "B" || $3 == "B")' | wc -l)" -lt "${limit}" ] && echo "${b}"
    done | head -1)
    # A tip takes a free slot of its class; a big tip may also take a free small slot, but only when
    # no small tip could have it (rank 10 and up): 74 big tips waited on two area slots while four small
    # slots sat idle (Oct 8 10:56Z). It then runs on the small slot's CPUs (12; Chonchon's 16).
    next=$(while read -r class queued branch sha; do
      slot=${class} extra=0
      if [[ " ${classes}" != *" ${class} "* ]]; then
        [ "${class}" = B ] && [ -n "${borrowable}" ] || continue
        slot=S extra=10
      fi
      if [[ ${branch} == cloud/land-* ]]; then rank=0
      elif [[ ${branch} == area/* ]] || grep -qxF "${branch}" "${state}/priority"; then rank=1
      elif [[ ${branch} == devtools/* ]]; then rank=2
      else rank=3; fi
      echo "$(( rank + extra )) ${queued} ${branch} ${sha} ${class} ${slot}"
    done < "${state}/queue" | sort -k1,1n -k2,2nr | head -1)
    [ -n "${next}" ] || break
    read -r _ queued branch sha class slot <<< "${next}"
    grep -vF " ${queued} ${branch} ${sha}" "${state}/queue" > "${state}/queue.tmp"; mv "${state}/queue.tmp" "${state}/queue"
    grep -qx "${sha}" "${state}/gated" && continue
    # The skip file ("branch sha" per line) takes a tip out by hand: a stale landing candidate, say.
    grep -qxF "${branch} ${sha}" "${state}/skip" 2>/dev/null && { echo "$(date -u +%H:%M:%S) skipped by hand ${branch} ${sha}"; continue; }
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
    if [ "${slot}" = "${class}" ]; then
      box=$(echo "${free}" | awk -v class="${slot}" '$2 == class {print $1; exit}')
    else
      box=${borrowable}
    fi
    log=${state}/logs/${sha:0:12}.log
    dispatch "${branch}" "${sha}" "${slot}" "${box}" "${class}" "${log}"
    now=$(date -u +%s)
    echo "$(date -u +%H:%M:%S) gating ${branch} ${sha} (${class}$([ "${slot}" = "${class}" ] || echo " in an ${slot} slot") on ${box}, waited $(( now - queued )) s, log ${log})"
    (recordWait "${sha},${branch},${class},$(date -u -r "${queued}" +%FT%TZ),$(date -u -r "${now}" +%FT%TZ),$(( now - queued )),started,${box}" > /dev/null 2>&1 &)
  done
  rmdir "${state}/slot-table.lock"
  slotTableOwned=""
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

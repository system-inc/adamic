#!/usr/bin/env bash
# Every push to codex/*, area/* and devtools/* gets a fast gate on the gate box, with nobody between
# the push and the gate. Run it from a checkout with a push credential (the box has none yet):
#
#   cloud/fast-gate-watch.sh
#
# It polls origin every 15 s. Branch tips it sees on its first poll are the backlog and are left
# alone; every tip that appears or moves after that is gated once (a sha already gated under another
# branch isn't gated again), two at a time, one per slot on the box. The queue is by priority, decided
# when a gate starts: what integration is landing first (area/*, and any branch named in the state
# directory's priority file, one per line, such as a fix-forward), then devtools/*, then workers'
# codex/*, newest first within each, and only a branch's newest tip. Each tip is classed when queued:
# big (an area, a stage3/ change, which runs the stage 3 lane, or more than two touched packages) or
# small. At most one big gate runs at a time, so one slot is always small-only and a worker's tip
# waits at most one small gate. The gating line names how long a tip waited. Each gate publishes
# gate-logs/<sha12>/<UTC stamp>/fast like any other, with the branch and, when an ai.db reply names
# the branch, the worker's session.
set -uo pipefail

here=$(cd "$(dirname "$0")/.." && pwd)
state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
mkdir -p "${state}"
slots=2
git -C "${here}" fetch -q origin
git -C "${here}" branch -r --contains "$(git -C "${here}" rev-parse HEAD)" | grep -q . || { echo "the gate's own commit is not on origin; push it first" >&2; exit 2; }
export ADAMIC_FAST_GATE_TOOLS_ON_ORIGIN=1

# A tip's class for scheduling only (coverage is the gate's business): big for an area or more than two
# touched packages (directories of changed Go files, a testdata path counting as its package).
classify() {
  local branch=$1 sha=$2 count
  [[ ${branch} == area/* ]] && { echo B; return; }
  git -C "${here}" fetch -q origin "${sha}" 2>/dev/null || { echo B; return; }
  changed=$(git -C "${here}" diff --name-only "$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)...${sha}" 2>/dev/null)
  # A stage3/ change runs the stage 3 lane (about 10 minutes): big, whatever else it touches.
  echo "${changed}" | grep -q '^stage3/' && { echo B; return; }
  count=$(echo "${changed}" | awk '/\/testdata\// {sub("/testdata/.*", ""); print; next} /\.go$/ {sub("/[^/]*$", ""); print}' | sort -u | wc -l)
  [ "${count}" -gt 2 ] && echo B || echo S
}

# Every gate start, appended to documentation/velocity/fast-gate-waits.csv on records/fast-gate-waits
# (sha, branch, class, queued, started, waited seconds), so slot wait is charted from an artifact.
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
  head -1 "${records}/${file}" | grep -q ',outcome$' || sed -i '' '1s/$/,outcome/' "${records}/${file}"
  echo "${row}" >> "${records}/${file}"
  git -C "${records}" add "${file}"
  git -C "${records}" -c user.name=kirkouimet -c user.email=kirk@kirkouimet.com commit -q -m "Slot wait: ${row}" -m "Co-Authored-By: Ahra <ahra@ahra.ai>"
  git -C "${records}" push -q origin "HEAD:refs/heads/records/fast-gate-waits" || true
}

# Why a finished gate has no real verdict, or nothing when it has one: no green/red line at all, or a
# red whose log shows the box failed (a missing work dir, a dropped ssh, a stale git lock, a full disk).
voidCause() {
  local log=$1
  grep -qE '^(green|red):' "${log}" 2>/dev/null || { echo "no verdict"; return; }
  grep -qE '^red:' "${log}" || return 0
  grep -oE 'creating work dir|Connection reset by|kex_exchange_identification|ssh: connect to host|Connection refused|index\.lock.: File exists|No space left on device' "${log}" | head -1
}

tips() {
  git -C "${here}" ls-remote origin 'refs/heads/codex/*' 'refs/heads/area/*' 'refs/heads/devtools/*' |
    awk '{sub("refs/heads/", "", $2); print $2, $1}' | sort
}

[ -s "${state}/seen" ] || tips > "${state}/seen"
touch "${state}/gated" "${state}/queue"
# Running gates are pid files (macOS bash 3.2 has no associative arrays).
mkdir -p "${state}/running" "${state}/logs"
echo "$(date -u +%H:%M:%S) watching codex/*, area/*, devtools/* (tools $(git -C "${here}" rev-parse --short HEAD))"
while true; do
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
    read -r branch sha class < "${file}"
    rm "${file}"
    cause=$(voidCause "${state}/logs/${sha:0:12}.log")
    if [ -z "${cause}" ]; then
      echo "$(date -u +%H:%M:%S) done ${branch}: $(grep -E '^(green|red):' "${state}/logs/${sha:0:12}.log" | tail -1)"
      continue
    fi
    # A void gate (it died before a real verdict, like the 33 a WSL restart killed at 08:23Z on Oct 8)
    # isn't served: un-mark it, queue it again, at most three tries, then tell integration it's the box.
    echo "${sha}" >> "${state}/void-tries"
    tries=$(grep -cx "${sha}" "${state}/void-tries")
    (recordWait "${sha},${branch},${class},,$(date -u +%FT%TZ),,void:${cause// /_}" > /dev/null 2>&1 &)
    if [ "${tries}" -le 3 ]; then
      grep -vx "${sha}" "${state}/gated" > "${state}/gated.tmp"; mv "${state}/gated.tmp" "${state}/gated"
      echo "${class} $(date -u +%s) ${branch} ${sha}" >> "${state}/queue"
      echo "$(date -u +%H:%M:%S) void ${branch} ${sha}: ${cause} (try ${tries} of 3), queued again"
    else
      echo "$(date -u +%H:%M:%S) void ${branch} ${sha}: ${cause}, three tries, given up"
      (cd /Users/kirkouimet/Projects/ahra && ahra os send system_adamic_integration "Fast gate of ${branch} ${sha} died three times with no verdict (${cause}): a box problem, not the change. Log: ${state}/logs/${sha:0:12}.log on Kirk's Mac." > /dev/null 2>&1 || true)
    fi
  done
  while [ "$(ls "${state}/running" | wc -l)" -lt "${slots}" ] && [ -s "${state}/queue" ]; do
    touch "${state}/priority"
    big=$(cat "${state}"/running/* 2>/dev/null | awk '$3 == "B"' | wc -l)
    # At most one big gate at a time, always: the other slot is small-only, so a worker's tip never
    # waits behind areas (@system_adamic, 21:42; with no small change waiting it stays free).
    only=""
    if [ "${big}" -ge 1 ]; then only=S; fi
    next=$(while read -r class queued branch sha; do
      [ -n "${only}" ] && [ "${class}" != "${only}" ] && continue
      if [[ ${branch} == area/* ]] || grep -qxF "${branch}" "${state}/priority"; then rank=1
      elif [[ ${branch} == devtools/* ]]; then rank=2
      else rank=3; fi
      echo "${rank} ${queued} ${branch} ${sha} ${class}"
    done < "${state}/queue" | sort -k1,1n -k2,2nr | head -1)
    [ -n "${next}" ] || break
    read -r _ queued branch sha class <<< "${next}"
    grep -vF " ${queued} ${branch} ${sha}" "${state}/queue" > "${state}/queue.tmp"; mv "${state}/queue.tmp" "${state}/queue"
    grep -qx "${sha}" "${state}/gated" && continue
    # A tip its branch has already moved past is superseded: gate the branch's newest only.
    grep -qx "${branch} ${sha}" "${state}/seen" || { echo "$(date -u +%H:%M:%S) superseded ${branch} ${sha}"; continue; }
    echo "${sha}" >> "${state}/gated"
    log=${state}/logs/${sha:0:12}.log
    bash "${here}/cloud/fast-gate.sh" "${sha}" --branch "${branch}" > "${log}" 2>&1 &
    echo "${branch} ${sha} ${class}" > "${state}/running/$!"
    now=$(date -u +%s)
    echo "$(date -u +%H:%M:%S) gating ${branch} ${sha} (${class}, waited $(( now - queued )) s, log ${log})"
    (recordWait "${sha},${branch},${class},$(date -u -r "${queued}" +%FT%TZ),$(date -u -r "${now}" +%FT%TZ),$(( now - queued )),started" > /dev/null 2>&1 &)
  done
  sleep 15
done

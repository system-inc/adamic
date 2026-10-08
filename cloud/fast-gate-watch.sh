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
# codex/*, newest first within each, and only a branch's newest tip. Each gate publishes gate-logs/<sha12>/<UTC stamp>/fast
# like any other, with the branch and, when an ai.db reply names the branch, the worker's session.
set -uo pipefail

here=$(cd "$(dirname "$0")/.." && pwd)
state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
mkdir -p "${state}"
slots=2
git -C "${here}" fetch -q origin
git -C "${here}" branch -r --contains "$(git -C "${here}" rev-parse HEAD)" | grep -q . || { echo "the gate's own commit is not on origin; push it first" >&2; exit 2; }
export ADAMIC_FAST_GATE_TOOLS_ON_ORIGIN=1

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
      echo "0 $(date -u +%s) ${branch} ${sha}" >> "${state}/queue"
      echo "$(date -u +%H:%M:%S) queued ${branch} ${sha}"
    done
    mv "${state}/now.tmp" "${state}/seen"
  fi
  for file in "${state}"/running/*; do
    [ -e "${file}" ] || continue
    kill -0 "$(basename "${file}")" 2>/dev/null && continue
    read -r branch sha < "${file}"
    echo "$(date -u +%H:%M:%S) done ${branch}: $(grep -E '^(green|red):' "${state}/logs/${sha:0:12}.log" | tail -1)"
    rm "${file}"
  done
  while [ "$(ls "${state}/running" | wc -l)" -lt "${slots}" ] && [ -s "${state}/queue" ]; do
    touch "${state}/priority"
    next=$(while read -r rank queued branch sha; do
      if [[ ${branch} == area/* ]] || grep -qxF "${branch}" "${state}/priority"; then rank=1
      elif [[ ${branch} == devtools/* ]]; then rank=2
      else rank=3; fi
      echo "${rank} ${queued} ${branch} ${sha}"
    done < "${state}/queue" | sort -k1,1n -k2,2nr | head -1)
    [ -n "${next}" ] || break
    read -r _ queued branch sha <<< "${next}"
    grep -v " ${queued} ${branch} ${sha}\$" "${state}/queue" > "${state}/queue.tmp"; mv "${state}/queue.tmp" "${state}/queue"
    grep -qx "${sha}" "${state}/gated" && continue
    # A tip its branch has already moved past is superseded: gate the branch's newest only.
    grep -qx "${branch} ${sha}" "${state}/seen" || { echo "$(date -u +%H:%M:%S) superseded ${branch} ${sha}"; continue; }
    echo "${sha}" >> "${state}/gated"
    log=${state}/logs/${sha:0:12}.log
    bash "${here}/cloud/fast-gate.sh" "${sha}" --branch "${branch}" > "${log}" 2>&1 &
    echo "${branch} ${sha}" > "${state}/running/$!"
    echo "$(date -u +%H:%M:%S) gating ${branch} ${sha} (log ${log})"
  done
  sleep 15
done

#!/usr/bin/env bash
# The fast gate's stall alarm, apart from the watcher, so a broken watcher can't silence it (the witness,
# Oct 8: a 24-minute stall with nothing started anywhere was found after the fact). Every minute: when a
# tip has waited in the queue over five minutes and no gate has started anywhere for five minutes,
# developer tools and integration hear it once a stall: how long, how many queued, the oldest, and
# whether the watcher is running. A start re-arms it. It reads only files: the watcher's queue and the
# slot-wait record the watcher writes at every start (documentation/velocity/fast-gate-waits.csv).
#
#   cloud/gate-stall-alarm.sh          # the loop (launchd: com.adamic.gate-stall-alarm), every 60 s
#   cloud/gate-stall-alarm.sh --once   # one check: prints the state, sends what the loop would
set -uo pipefail
state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
waits=${ADAMIC_FAST_GATE_WAITS:-${state}/waits-records/documentation/velocity/fast-gate-waits.csv}
ahraDirectory=${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}
limit=${ADAMIC_STALL_SECONDS:-300}

tell() {
  # ahra refuses all-caps words.
  local text recipient
  text=$(printf '%s' "$1" | awk '{ out = ""; while (match($0, /[A-Z][A-Z][A-Z]+/)) { out = out substr($0, 1, RSTART - 1) tolower(substr($0, RSTART, RLENGTH)); $0 = substr($0, RSTART + RLENGTH) } print out $0 }')
  for recipient in system_adamic_developer_tools system_adamic_integration; do
    (cd "${ahraDirectory}" && ahra os send "${recipient}" "${text}" --from system_adamic_developer_tools > /dev/null 2>&1) ||
      echo "$(date -u +%H:%M:%S) could not tell ${recipient}"
  done
}

# Slots a queued tip could take now: the slot table's lines on boxes not drained for a reserved run, less
# the gates running on them. A gate counts only while its watcher-recorded pid is alive, so a dead
# watcher's leftovers never make a stalled fleet look full.
freeSlots() {
  local drained="" pid box total=0 busy=0
  for pid in "${state}"/reserved-running/*; do
    [ -f "${pid}" ] && kill -0 "$(basename "${pid}")" 2>/dev/null && drained="${drained} $(cat "${pid}")"
  done
  total=$(awk -v drained="${drained} " 'NF >= 2 && index(drained, " " $1 " ") == 0' "${state}/slots" 2>/dev/null | grep -c .)
  for pid in "${state}"/running/*; do
    [ -f "${pid}" ] && kill -0 "$(basename "${pid}")" 2>/dev/null || continue
    box=$(awk '{print ($4 == "" ? "threadripper" : $4)}' "${pid}")
    case " ${drained} " in *" ${box} "*) ;; *) busy=$(( busy + 1 )) ;; esac
  done
  echo $(( total - busy ))
}

check() {
  local now oldest lastStart queued waited idle watcher
  now=$(date -u +%s)
  if [ ! -s "${state}/queue" ]; then
    rm -f "${state}/stall-alarmed"
    return 0
  fi
  oldest=$(awk '{print $2}' "${state}/queue" | sort -n | head -1)
  # The newest start the record holds: started_utc of its newest 'started' row.
  lastStart=$(python3 -c '
import csv, datetime, sys
starts = []
try:
    with open(sys.argv[1], newline="") as handle:
        for row in csv.DictReader(handle):
            if row.get("outcome", "started") in ("started", "") and row.get("started_utc"):
                starts.append(datetime.datetime.strptime(row["started_utc"], "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=datetime.timezone.utc).timestamp())
except OSError:
    pass
print(int(max(starts)) if starts else 0)' "${waits}")
  queued=$(grep -c . "${state}/queue")
  waited=$(( now - oldest ))
  idle=$(( now - lastStart ))
  pgrep -f "fast-gate-watch.sh" > /dev/null 2>&1 && watcher="running" || watcher="not running"
  free=$(freeSlots)
  echo "$(date -u +%H:%M:%S) ${queued} queued, oldest ${waited} s, last start ${idle} s ago, ${free} slots free, watcher ${watcher}"
  # Every slot busy is a full fleet, not a stall (Oct 8 19:01Z: 13 gates on 13 usable slots raised one).
  if [ "${waited}" -ge "${limit}" ] && [ "${idle}" -ge "${limit}" ] && [ "${free}" -gt 0 ]; then
    [ -f "${state}/stall-alarmed" ] && return 0
    touch "${state}/stall-alarmed"
    tell "fast gate stall (the standalone alarm): no gate has started anywhere for ${idle} s while ${queued} tips wait, the oldest for ${waited} s ($(sort -k2,2n "${state}/queue" | head -1 | awk '{print $3}')); the watcher is ${watcher}. Its log: ~/Projects/system/adamic-gate-logs/fast-gate-watch.log on Kirk's Mac."
    echo "$(date -u +%H:%M:%S) alarm sent"
  else
    rm -f "${state}/stall-alarmed"
  fi
}

if [ "${1:-}" = --once ]; then
  check
  exit 0
fi
while true; do
  check
  sleep 60
done

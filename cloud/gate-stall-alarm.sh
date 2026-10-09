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
  # A slot held for a family (a third-field glob, or Server's area slot for the first ready step's globs)
  # is usable only while a tip of that family waits or runs there: at 20:38Z the only "free" slot was
  # Server's, held for runtime's step with no runtime tip queued, and the alarm fired on a full fleet.
  python3 - "${state}" <<'PY'
import fnmatch, os, sys
state = sys.argv[1]
def read(name):
    try:
        with open(os.path.join(state, name)) as handle:
            return [line.split() for line in handle if line.strip()]
    except OSError:
        return []
def alive(pid):
    try:
        os.kill(int(pid), 0)
        return True
    except (OSError, ValueError):
        return False
drained = set()
for pid in os.listdir(os.path.join(state, 'reserved-running')) if os.path.isdir(os.path.join(state, 'reserved-running')) else []:
    if alive(pid):
        drained.update(word for line in read(os.path.join('reserved-running', pid)) for word in line)
# A box the star runs on or waits for takes no other gate (the star owns its box, Oct 8 21:47Z): its idle slots aren't
# free. At 00:30Z on Oct 9 the alarm paged a stall on the star boxes' six idle small slots.
drained.update(fields[0] for fields in read('star-boxes') if fields)
running = []
for pid in os.listdir(os.path.join(state, 'running')) if os.path.isdir(os.path.join(state, 'running')) else []:
    fields = (read(os.path.join('running', pid)) or [[]])[0]
    if alive(pid) and fields:
        running.append((fields[0], fields[2] if len(fields) > 2 else '', fields[3] if len(fields) > 3 else 'threadripper'))
queued = [fields[2] for fields in read('queue') if len(fields) >= 4]
firstStep = [fields[0] for fields in read('first-step-globs.poll')]
total = 0
for fields in read('slots'):
    if len(fields) < 2 or fields[0] in drained:
        continue
    box, slot = fields[0], fields[1]
    # Loom's pool takes side tips only, while it is switched on (#xt96xyp).
    if box == 'pool' and not (os.path.exists(os.path.join(state, 'pool-side')) and
                              any(branch.startswith(('codex/', 'devtools/')) for branch in queued)):
        continue
    globs = fields[2:3] or (firstStep if (box, slot) == ('server', 'B') else [])
    globs = [glob.rstrip(',') for glob in globs if glob.rstrip(',')]
    if globs:
        family = lambda branch: any(fnmatch.fnmatchcase(branch, glob) for glob in globs)
        if not any(family(branch) for branch in queued) and not any(family(b) and c == slot and x == box for b, c, x in running):
            continue
    total += 1
busy = sum(1 for _, _, box in running if box not in drained)
print(total - busy)
PY
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

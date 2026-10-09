#!/usr/bin/env bash
# The probe's watchdog (#bhd6n3f, for Outcome #awn479j). The probe (cloud/probe-90.sh) starts every 30 minutes, stamps
# ${state}/probe-90-started when it pushes, and posts one line on #awn479j when its run ends, up to its 35-minute cap
# later, stamping ${state}/probe-90-posted. Three times on Oct 9 its headline went stale (42 minutes the third time: a
# script that never landed, a run that skipped its cycle) and someone found it by reading. This runs every 5 minutes
# under launchd (cloud/probe-90/com.adamic.probe-90-watchdog.plist) and pages Loom Judge and Loom Operations, once per
# silence, in two cases:
#   - a run in progress (started after the last post) that has passed its cap plus 5 minutes without posting: hung;
#   - no run in progress, and nothing started or posted for 35 minutes: the loop isn't firing.
# A run inside its cap is a healthy run, never a silence (the first version paged on one at 23:24Z, Oct 9).
#
#   cloud/probe-90-watchdog.sh
#
# Before the probe's first stamp (a fresh deploy), the watchdog's own first run is the baseline, so a probe that never
# starts pages 35 minutes after the watchdog arms.
set -uo pipefail

state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
ahraDirectory=${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}
log=${ADAMIC_PROBE_LOG:-/Users/kirkouimet/Projects/system/adamic-gate-logs/probe-90.log}
stale=${ADAMIC_PROBE_STALE_SECONDS:-2100}
# The probe's own cap (cloud/probe-90.sh's limit for the leaf) and the grace past it.
cap=${ADAMIC_PROBE_SECONDS:-2100}
grace=300
recipients=${ADAMIC_PROBE_PAGE_TO:-system_adamic_loom_judge system_adamic_loom_operations}
now=${ADAMIC_PROBE_NOW:-$(date -u +%s)}

say() {
  echo "$(date -u +%H:%M:%S) $*"
}
# Tests replace both: ADAMIC_PROBE_PAGE is a command taking the recipient and the message, ADAMIC_PROBE_JOB one that
# prints the probe's launchd line.
page() {
  if [ -n "${ADAMIC_PROBE_PAGE:-}" ]; then
    ${ADAMIC_PROBE_PAGE} "$1" "$2"
  else
    (cd "${ahraDirectory}" && ahra os send "$1" "$2" > /dev/null)
  fi
}
job() {
  if [ -n "${ADAMIC_PROBE_JOB:-}" ]; then
    ${ADAMIC_PROBE_JOB}
  else
    launchctl list 2> /dev/null | awk '$3 == "com.adamic.probe-90" {print $1, $2}'
  fi
}

stamp() {
  local value
  value=$(cat "${state}/$1" 2> /dev/null)
  [[ ${value} =~ ^[0-9]+$ ]] && echo "${value}" || echo 0
}
posted=$(stamp probe-90-posted)
started=$(stamp probe-90-started)
armed=$(stamp probe-90-watchdog-armed)
if [ "${posted}" = 0 ] && [ "${started}" = 0 ] && [ "${armed}" = 0 ]; then
  armed=${now}
  echo "${armed}" > "${state}/probe-90-watchdog-armed"
  say "armed: no probe run yet, so silence is counted from now"
fi
if [ "${started}" -gt "${posted}" ]; then
  # A run in progress: healthy until its cap and grace pass without a post.
  since=${started}
  age=$((now - since))
  [ "${age}" -ge $((cap + grace)) ] || exit 0
  what="a probe run started $((age / 60)) min ago and hasn't posted, past its $((cap / 60))-minute cap"
else
  since=$((posted > started ? posted : started))
  [ "${since}" -gt 0 ] || since=${armed}
  age=$((now - since))
  [ "${age}" -ge "${stale}" ] || exit 0
  what="no probe run has started or posted for $((age / 60)) min, over its $((stale / 60))-minute line"
fi
# One page per silence: the episode is the stamp it waited on, so a fresh start or post rearms it.
[ "$(cat "${state}/probe-90-paged" 2> /dev/null)" = "${since}" ] && exit 0

read -r pid status < <(job)
if [ -z "${pid:-}" ]; then
  launchd="com.adamic.probe-90 is not loaded"
elif [ "${pid}" = - ]; then
  launchd="com.adamic.probe-90 is loaded, not running, last exit ${status}"
else
  launchd="com.adamic.probe-90 is running (pid ${pid})"
fi
message="The 90-second probe is silent: ${what}, so #awn479j's headline is stale. ${launchd}. Log: ${log}."
sent=0
for recipient in ${recipients}; do
  if page "${recipient}" "${message}"; then
    sent=1
    say "paged ${recipient}: ${message}"
  else
    say "could not page ${recipient}"
  fi
done
# A page that reached nobody is tried again on the next run.
[ "${sent}" = 1 ] && echo "${since}" > "${state}/probe-90-paged"
exit 0

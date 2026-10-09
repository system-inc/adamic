#!/usr/bin/env bash
# The probe's watchdog (#bhd6n3f, for Outcome #awn479j). The probe (cloud/probe-90.sh) posts one line on #awn479j every
# 30 minutes and stamps ${state}/probe-90-posted when it does. Three times on Oct 9 its headline went stale (42 minutes
# the third time: a script that never landed, a run that skipped its cycle) and someone found it by reading. This runs
# every 5 minutes under launchd (cloud/probe-90/com.adamic.probe-90-watchdog.plist), and when the newest post is older
# than 35 minutes it pages verdict and developer tools, once per silence: the page names the post it waited on, and a
# fresh post rearms it.
#
#   cloud/probe-90-watchdog.sh
#
# Before the probe's first stamp (a fresh deploy), the watchdog's own first run is the baseline, so a probe that never
# posts pages 35 minutes after the watchdog arms.
set -uo pipefail

state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
ahraDirectory=${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}
log=${ADAMIC_PROBE_LOG:-/Users/kirkouimet/Projects/system/adamic-gate-logs/probe-90.log}
stale=${ADAMIC_PROBE_STALE_SECONDS:-2100}
recipients=${ADAMIC_PROBE_PAGE_TO:-system_adamic_release_verdict system_adamic_developer_tools}
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

posted=$(cat "${state}/probe-90-posted" 2> /dev/null)
if [[ ${posted} =~ ^[0-9]+$ ]]; then
  since=${posted} what="its last post on #awn479j"
else
  armed=$(cat "${state}/probe-90-watchdog-armed" 2> /dev/null)
  if ! [[ ${armed} =~ ^[0-9]+$ ]]; then
    armed=${now}
    echo "${armed}" > "${state}/probe-90-watchdog-armed"
    say "armed: no probe post yet, so silence is counted from now"
  fi
  since=${armed} what="no post on #awn479j since the watchdog armed"
fi
age=$((now - since))
[ "${age}" -ge "${stale}" ] || exit 0
# One page per silence: the episode is the post it waited on, so a fresh post rearms it.
[ "$(cat "${state}/probe-90-paged" 2> /dev/null)" = "${since}" ] && exit 0

read -r pid status < <(job)
if [ -z "${pid:-}" ]; then
  launchd="com.adamic.probe-90 is not loaded"
elif [ "${pid}" = - ]; then
  launchd="com.adamic.probe-90 is loaded, not running, last exit ${status}"
else
  launchd="com.adamic.probe-90 is running (pid ${pid})"
fi
message="The 90-second probe is silent: ${what} was $((age / 60)) min ago, over its $((stale / 60))-minute line, so #awn479j's headline is stale. ${launchd}. Log: ${log}."
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

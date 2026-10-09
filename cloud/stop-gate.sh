#!/usr/bin/env bash
# Stops one fast gate on its box, by its exact commit:
#
#   cloud/stop-gate.sh <box> <40-hex sha> "<reason>"
#
# run.py reads the reason from beside its out directory on SIGTERM, publishes what it had and exits (a red
# with its failures, or void before any). The sha must be all 40 hex digits: a stop by hand on Oct 8 21:48Z
# read an empty sha, its pattern matched every gate on Workshop, and the star lost a 35-minute run. The box
# side checks again, so a caller that skips this script can't widen the match either.
set -euo pipefail
box=${1:-} sha=${2:-} reason=${3:-}
if [ -z "${box}" ] || ! [[ ${sha} =~ ^[0-9a-f]{40}$ ]] || [ -z "${reason}" ]; then
  echo "usage: cloud/stop-gate.sh <box> <40-hex sha> \"<reason>\" (refused: box '${box}', sha '${sha}')" >&2
  exit 2
fi
ssh "${box}" bash -s -- "$(printf '%q ' "${sha}" "${reason}")" <<'STOP'
set -eu
sha=$1 reason=$2
case ${#sha}:${sha} in
  40:*[!0-9a-f]*) echo "refused: sha '${sha}' is not 40 hex digits" >&2; exit 2 ;;
  40:*) ;;
  *) echo "refused: sha '${sha}' is not 40 hex digits" >&2; exit 2 ;;
esac
for out in ~/fast-gate/out/"${sha:0:12}"-*; do
  [ -d "${out}" ] || continue
  printf '%s\n' "${reason}" > "${out}.stop-reason"
done
pkill -TERM -f "run.py .*--sha ${sha}"
STOP

#!/usr/bin/env bash
# The box's side of cloud/fast-gate/tools.txt: run every check for this platform, print one line per
# declared tool the box lacks ("lacks <name>: <check>"), and exit 1 if any. Run after env.sh and the
# gate's PATH, from the gate's tools checkout: bash cloud/fast-gate/tools-check.sh
set -uo pipefail
platform=$(uname -s | tr '[:upper:]' '[:lower:]')
missing=0
while IFS=$'\t' read -r name where check; do
  case ${name} in ''|'#'*) continue ;; esac
  [ "${where}" = all ] || [ "${where}" = "${platform}" ] || continue
  if ! bash -c "${check}" > /dev/null 2>&1; then
    echo "lacks ${name}: ${check}"
    missing=1
  fi
done < "$(dirname "$0")/tools.txt"
exit ${missing}

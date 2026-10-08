#!/usr/bin/env bash
# First ready roadmap step that declares candidate branches, in roadmap order.
set -uo pipefail
cd "${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}" || exit 1
ready=$(ahra tasks ready system_adamic) || exit 1
ids=$(printf '%s\n' "${ready}" | sed -nE 's/.*#([a-z0-9]{7})([^a-z0-9].*|$)/\1/p')
while read -r id; do
  [ -n "${id}" ] || continue
  body=$(ahra tasks show "${id}") || exit 1
  line=$(printf '%s\n' "${body}" | sed -n 's/^Branches: *//p' | head -1)
  if [ -n "${line}" ]; then
    printf '%s\n' "${line}" | awk '{for (i=1; i<=NF; i++) print $i}'
    exit 0
  fi
done <<< "${ids}"

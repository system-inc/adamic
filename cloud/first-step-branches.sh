#!/usr/bin/env bash
# First ready roadmap step that declares candidate branches, in roadmap order: its globs, one a line.
# With --source, one line "<step id> <first glob> <source branch>" when that step also declares
# 'Source: <branch>' (cloud/landing-candidate.sh builds its candidates from it), else nothing.
# With --all, every ready step's globs as "<position> <glob>" lines, position 0 for the first step that
# declares any, in the order ahra tasks ready ranks them (the watcher orders its queue by them).
set -uo pipefail
mode=${1:-}
cd "${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}" || exit 1
# With --waves, "<wave> <glob>" for every open step of system_adamic's waterfall in waves 0 and 1, one line per branch
# glob its artifacts declare (git:<repository>:<glob>): the watcher stamps a pool job's tier from them (#12dg93f). A
# step that declares no git artifact has no branch to match, so its tips take the lowest tier.
if [ "${mode}" = --waves ]; then
  ahra tasks waterfall system_adamic --json | python3 -c '
import json, sys
for node in json.load(sys.stdin)["nodes"]:
    if node.get("wave") in (0, 1) and node.get("status") not in ("Done", "Cancelled", "Failed"):
        for artifact in node.get("artifacts") or []:
            if artifact.startswith("git:"):
                print(node["wave"], artifact.rsplit(":", 1)[1])
'
  exit
fi
ready=$(ahra tasks ready system_adamic) || exit 1
ids=$(printf '%s\n' "${ready}" | sed -nE 's/.*#([a-z0-9]{7})([^a-z0-9].*|$)/\1/p')
position=0
while read -r id; do
  [ -n "${id}" ] || continue
  body=$(ahra tasks show "${id}") || exit 1
  line=$(printf '%s\n' "${body}" | sed -n 's/^Branches: *//p' | head -1)
  if [ -n "${line}" ] && [ "${mode}" = --all ]; then
    printf '%s\n' "${line}" | awk -v p="${position}" '{for (i=1; i<=NF; i++) print p, $i}'
    position=$((position + 1))
    continue
  fi
  if [ -n "${line}" ]; then
    if [ "${mode}" = --source ]; then
      source=$(printf '%s\n' "${body}" | sed -n 's/^Source: *//p' | head -1 | awk '{print $1}')
      [ -n "${source}" ] && echo "${id} $(printf '%s\n' "${line}" | awk '{print $1}') ${source}"
    else
      printf '%s\n' "${line}" | awk '{for (i=1; i<=NF; i++) print $i}'
    fi
    exit 0
  fi
done <<< "${ids}"

#!/usr/bin/env bash
# First ready roadmap step that declares candidate branches, in roadmap order: its globs, one a line.
# With --source, one line "<step id> <first glob> <source branch>" when that step also declares
# 'Source: <branch>' (cloud/landing-candidate.sh builds its candidates from it), else nothing.
# With --all, every ready step's globs as "<position> <glob>" lines, position 0 for the first step that
# declares any, in the order ahra tasks ready ranks them (the watcher orders its queue by them).
set -uo pipefail
mode=${1:-}
cd "${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}" || exit 1
# With --waves, the pool tiers system_adamic's waterfall gives (#12dg93f), as "<tier> task <id>" and "<tier> branch <glob>
# <id>" lines: the star (the critical path's head) at 40 and every other open step of waves 0 and 1 at 30, by task id for
# a candidate whose commits carry a "Task: #<id>" trailer, and by the git:<repository>:<glob> artifacts it declares for
# one with no trailer (@system_adamic, Oct 9 13:33Z).
if [ "${mode}" = --waves ]; then
  ahra tasks waterfall system_adamic --json | python3 -c '
import json, sys
waterfall = json.load(sys.stdin)
star = (waterfall.get("criticalPath") or [None])[0]
for node in waterfall["nodes"]:
    if node.get("status") in ("Done", "Cancelled", "Failed"):
        continue
    tier = 40 if node["id"] == star else 30 if node.get("wave") in (0, 1) else None
    if tier is None:
        continue
    print(tier, "task", node["id"])
    for artifact in node.get("artifacts") or []:
        if artifact.startswith("git:"):
            print(tier, "branch", artifact.rsplit(":", 1)[1], node["id"])
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

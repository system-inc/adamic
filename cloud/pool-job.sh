#!/usr/bin/env bash
# One side tip's fast gate on Loom's side pool (#xt96xyp; @system_adamic, Oct 8 23:53Z: side work's fast gates go to
# the pool, so the boxes are left for the star and landings). The watcher dispatches it like cloud/fast-gate.sh, on a
# "pool P" slot, and reads its log the same way:
#
#   cloud/pool-job.sh <sha> --branch <branch> --tools <tools sha>
#
# It writes ~/.loom/jobs/fast/<sha>.json (branch, sha, base, tools, and packages "select": the pool runs
# cloud/fast-gate/run.py --select from those tools where the tree is whole, so one selection decides what a fast gate
# tests wherever it runs), waits for Loom's <sha>.verdict, and prints a fast gate's log: the header, the verdict's
# status line, and the published record. A pool that gives no verdict in time is void, and the watcher sends the tip
# to the boxes.
set -uo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
sha=${1:-}
shift || true
branch="" tools="" priority=""
while [ $# -gt 0 ]; do
  case $1 in
    --branch) branch=$2; shift 2 ;;
    --tools) tools=$2; shift 2 ;;
    --priority) priority=$2; shift 2 ;;
    *) shift ;;
  esac
done
[[ ${sha} =~ ^[0-9a-f]{40}$ ]] || { echo "void: pool job refused: sha '${sha}' is not 40 hex digits"; exit 2; }
jobs=${LOOM_FAST_JOBS:-${HOME}/.loom/jobs/fast}
limit=${ADAMIC_POOL_JOB_SECONDS:-5400}
# A job Loom hasn't started (no .work or .running) within this long is cancelled and void, and the watcher sends the
# tip to the boxes: the boxes are the fallback once the pool gates everything (@system_adamic, Oct 9 04:19Z).
take=${ADAMIC_POOL_TAKE_SECONDS:-600}
mkdir -p "${jobs}"
. "${here}/cloud/fast-gate-classify.sh"
read -r baseName base <<< "$(gateBase "${branch}" "${sha}")"
[[ ${base} =~ ^[0-9a-f]{40}$ ]] || { echo "void: ${sha} pool job: no base for ${branch} (origin unreachable?)"; exit 1; }
echo "fast gate: ${sha} against ${baseName} ${base}, tools ${tools}, on pool, class P"
# A verdict left from an earlier job of this sha belongs to other tools or another base: the pool runs it fresh.
rm -f "${jobs}/${sha}.verdict"
python3 - "${jobs}/${sha}.json" "${branch}" "${sha}" "${base}" "${baseName}" "${tools}" "${priority}" <<'PY'
import json, os, sys
path, branch, sha, base, baseName, tools, priority = sys.argv[1:]
job = {"branch": branch, "sha": sha, "base": base, "base_name": baseName, "tools": tools, "packages": "select", "env": {}}
if priority:
    job["priority"] = int(priority)
with open(path + ".partial", "w") as handle:
    json.dump(job, handle)
os.replace(path + ".partial", path)
PY
waited=0
trap 'echo "stopped: pool job of ${sha} stopped by the watcher"; exit 143' TERM
while [ ! -f "${jobs}/${sha}.verdict" ]; do
  if [ "${waited}" -ge "${take}" ] && [ ! -e "${jobs}/${sha}.work" ] && [ ! -e "${jobs}/${sha}.running" ]; then
    echo "not taken within ${take} s, sent to the boxes" > "${jobs}/${sha}.cancel"
    echo "void: ${sha} the pool hadn't started it in ${take} s"
    exit 1
  fi
  if [ "${waited}" -ge "${limit}" ]; then
    echo "void: ${sha} the pool gave no verdict in ${limit} s"
    exit 1
  fi
  sleep 10 & wait $!
  waited=$((waited + 10))
done
head -1 "${jobs}/${sha}.verdict"
record=$(sed -n 's/^record: //p' "${jobs}/${sha}.verdict" | head -1)
[ -n "${record}" ] && echo "published ${record}"
exit 0

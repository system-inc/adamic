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
# Every job carries its waterfall tier (#12dg93f): Loom ranks one without a priority at 0, behind all side work, where
# the after-chain and hidden-boundaries sat for an hour (Oct 9 13:05Z). No default here, so a caller that forgets is loud.
[[ ${priority} =~ ^[0-9]+$ ]] || { echo "void: ${sha} pool job refused: no tier (--priority), which Loom would rank 0"; exit 2; }
jobs=${LOOM_FAST_JOBS:-${HOME}/.loom/jobs/fast}
limit=${ADAMIC_POOL_JOB_SECONDS:-5400}
# A job whose units Loom hasn't all placed within this long gets a box racing it, the boxes being the fallback once the
# pool gates everything (@system_adamic, Oct 9 04:19Z and 09:21Z: two minutes). Loom writes <sha>.placed, "<placed>
# <total>" (#3tj643t); the request goes in the watcher's races directory.
take=${ADAMIC_POOL_TAKE_SECONDS:-120}
races=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}/race-wanted
mkdir -p "${jobs}"
. "${here}/cloud/fast-gate-classify.sh"
read -r baseName base <<< "$(gateBase "${branch}" "${sha}")"
[[ ${base} =~ ^[0-9a-f]{40}$ ]] || { echo "void: ${sha} pool job: no base for ${branch} (origin unreachable?)"; exit 1; }
echo "fast gate: ${sha} against ${baseName} ${base}, tools ${tools}, on pool, class P"
# The pool gates the tip merged onto its base's tip, as the boxes do (gateMerge, #11ymb02): the job names the merge
# commit as "gate", pushed where an instance can fetch it. A conflict is the tip's red now, with no pool time spent.
read -r merged gated <<< "$(gateMerge "${sha}" "${base}" || echo "error")"
case ${merged} in
  same) gated="" ;;
  merge)
    git -C "${here}" push -q origin "${gated}:refs/gate-merges/${gated}" || { echo "void: ${sha} pool job couldn't push its merge onto ${baseName} ${base}"; exit 1; }
    echo "gating ${sha} merged onto ${baseName} ${base} as ${gated}" ;;
  conflict)
    echo "red: ${sha} fast gate on Loom's side pool, first failure at merge after 0.0 s (conflicts with ${baseName} ${base:0:12} in ${gated})"
    exit 0 ;;
  *) echo "void: ${sha} pool job couldn't merge it onto ${baseName} ${base}"; exit 1 ;;
esac
# A verdict left from an earlier job of this sha belongs to other tools or another base: the pool runs it fresh. So does a
# cancel left from one, which would stop this job before it starts.
rm -f "${jobs}/${sha}.verdict" "${jobs}/${sha}.cancel"
python3 - "${jobs}/${sha}.json" "${branch}" "${sha}" "${base}" "${baseName}" "${tools}" "${priority}" "${gated}" <<'PY'
import json, os, sys
path, branch, sha, base, baseName, tools, priority, gated = sys.argv[1:]
job = {"branch": branch, "sha": sha, "base": base, "base_name": baseName, "tools": tools, "packages": "select", "env": {}}
# Landings and areas run complete, as on the boxes: every phase past a Go red, and the stage 3 lane (Loom af3ce75).
if branch.startswith(("cloud/land-", "area/")):
    job["complete"] = True
if gated:
    job["gate"] = gated
if priority:
    job["priority"] = int(priority)
# A higher tier already in the file stays: Loom or integration may have raised it by hand (a P0 at 40 outside the
# waterfall, Oct 9 13:23Z), and a requeue writing its own 10 over it sent fx5 refusals 7f746e5a to the side slots.
try:
    with open(path) as handle:
        earlier = json.load(handle).get("priority")
    if isinstance(earlier, int) and earlier > job.get("priority", 0):
        job["priority"] = earlier
except (OSError, ValueError, AttributeError):
    pass
with open(path + ".partial", "w") as handle:
    json.dump(job, handle)
os.replace(path + ".partial", path)
PY
waited=0
trap 'echo "stopped: pool job of ${sha} stopped by the watcher"; exit 143' TERM
while [ ! -f "${jobs}/${sha}.verdict" ]; do
  # Not started in time, the job stays queued at Loom and a box races it (#04gypqe; @system_adamic, Oct 9: never cancel a
  # run to move it, race it). The watcher reads the request, gates the tip on a box too, and stops whichever route is
  # still running when the other answers green or red.
  # Checked once, at the deadline: a unit that ends later doesn't start a race late.
  if [ -z "${raceChecked:-}" ] && [ "${waited}" -ge "${take}" ]; then
    raceChecked=1
    placed="" total=""
    { read -r placed total < "${jobs}/${sha}.placed"; } 2> /dev/null || true
    if ! [[ ${placed} =~ ^[0-9]+$ && ${total} =~ ^[0-9]+$ && ${placed} -ge ${total} ]] && [ ! -e "${races}/${sha}" ]; then
      mkdir -p "${races}"
      echo "${branch}" > "${races}/${sha}"
      echo "racing: ${sha} the pool placed ${placed:-0} of ${total:-its} units in ${take} s, so a box races it"
    fi
  fi
  if [ "${waited}" -ge "${limit}" ]; then
    # The job stays at Loom: its record still publishes and lands the candidate though no waiter reads it (hidden-boundaries
    # 7fddff4b was running on the pool when its cancel, written for a waiter that had given up, stopped it; Oct 9 13:35Z).
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

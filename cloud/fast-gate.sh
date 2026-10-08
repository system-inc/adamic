#!/usr/bin/env bash
# The fast gate: build, vet, the touched packages' tests, the oracle smoke set and the skip census,
# on the gate box, stopping at the first failure. Run it from any checkout with a push credential:
#
#   cloud/fast-gate.sh <full sha> [--branch <name>] [--session <ai.db session id>]
#
# The branch defaults to the origin branch whose tip is the sha; the session defaults to the newest
# ai.db session whose latest reply names that branch (a Codex worker reports the branch it pushed),
# and fast.json says which source each came from.
#
# The gate's own code comes from this script's checkout (its HEAD, which must be on origin), never
# from the candidate, so a candidate can't edit the gate that judges it. The base is origin/main at
# the moment it starts. The box keeps one tree and one tools checkout at fixed paths, so Go's build
# cache stays warm; tests always run -count=1. The result is published to
# gate-logs/<sha12>/<UTC stamp>/fast: status.txt (green: or red: on its first line), fast.json,
# test.jsonl and each step's log.
set -euo pipefail

sha=${1:?usage: cloud/fast-gate.sh <full sha> [--branch <name>] [--session <id>]}
shift
branch="" branchSource=given session="" sessionSource=given
while [ $# -gt 0 ]; do
  case $1 in
    --branch) branch=$2; shift 2 ;;
    --session) session=$2; shift 2 ;;
    *) echo "unknown argument $1" >&2; exit 2 ;;
  esac
done
[[ ${sha} =~ ^[0-9a-f]{40}$ ]] || { echo "want a full 40-character sha, got ${sha}" >&2; exit 2; }
box=${ADAMIC_FAST_GATE_BOX:-threadripper}
here=$(cd "$(dirname "$0")/.." && pwd)
tools=$(git -C "${here}" rev-parse HEAD)
# The watcher checks this once for all its gates, so two gates at once don't race one fetch.
if [ -z "${ADAMIC_FAST_GATE_TOOLS_ON_ORIGIN:-}" ]; then
  git -C "${here}" fetch -q origin
  git -C "${here}" branch -r --contains "${tools}" | grep -q . || { echo "the gate's own commit ${tools} is not on origin; push it first" >&2; exit 2; }
fi
base=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
if [ -z "${branch}" ]; then
  branch=$(git -C "${here}" ls-remote origin 'refs/heads/*' | awk -v sha="${sha}" '$1 == sha && $2 !~ /^refs\/heads\/gate-logs\// && !found {sub("refs/heads/", "", $2); print $2; found = 1}')
  branchSource="origin branch tip"
fi
database=${ADAMIC_AI_DATABASE:-/Users/kirkouimet/Projects/ahra/modules/ai/data/ai.db}
if [ -z "${session}" ] && [ -n "${branch}" ] && [ -f "${database}" ]; then
  session=$(sqlite3 "${database}" "select session_id from replies where text like '%' || '${branch//\'/}' || '%' order by fetched_at desc limit 1" 2>/dev/null || true)
  sessionSource="ai.db reply naming the branch"
fi
[ -n "${session}" ] || sessionSource=none
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out=fast-gate/out/${sha:0:12}-${stamp}

echo "fast gate: ${sha} against main ${base}, tools ${tools}, on ${box}"
set +e
# ssh joins its arguments into one remote command line, so each is quoted for the remote shell.
ssh "${box}" bash -s -- "$(printf '%q ' "${sha}" "${base}" "${tools}" "${out}" "${branch:-}" "${branchSource}" "${session:-}" "${sessionSource}")" <<'BOX'
set -euo pipefail
sha=$1 base=$2 tools=$3 out=$4 branch=$5 branchSource=$6 session=$7 sessionSource=$8
mkdir -p ~/fast-gate
# Two slots, each with its own tree and tools checkout, so a small change doesn't wait behind a
# stack's long gate; the second slot's tree starts as a copy of the first (submodules included).
# With both busy it takes whichever frees first, never waiting on one while the other is idle.
slot=""
while [ -z "${slot}" ]; do
  for candidate in 1 2; do
    exec 9> ~/fast-gate/lock$([ "${candidate}" = 1 ] && echo "" || echo "-${candidate}")
    if flock -n 9; then slot=${candidate}; break; fi
    exec 9>&-
  done
  [ -n "${slot}" ] || sleep 1
done
suffix=$([ "${slot}" = 1 ] && echo "" || echo "-${slot}")
source ~/adamic-tools/env.sh
for directory in tools tree; do
  if [ ! -d ~/fast-gate/${directory}${suffix} ]; then
    if [ -d ~/fast-gate/${directory} ]; then cp -a ~/fast-gate/${directory} ~/fast-gate/${directory}${suffix}; else git clone -q https://github.com/system-inc/adamic.git ~/fast-gate/${directory}${suffix}; fi
  fi
done
git -C ~/fast-gate/tools${suffix} fetch -q origin "${tools}"
git -C ~/fast-gate/tools${suffix} switch -q --detach "${tools}"
git -C ~/fast-gate/tree${suffix} fetch -q origin "${sha}" "${base}"
git -C ~/fast-gate/tree${suffix} switch -q --detach "${sha}"
git -C ~/fast-gate/tree${suffix} submodule update -q --init --recursive
mkdir -p ~/"${out}"
echo "slot=${slot} load_before=$(cut -d' ' -f1-3 /proc/loadavg)" > ~/"${out}"/box.txt
# The box's thread budget: each slot owns half the CPUs (Go sizes GOMAXPROCS from the affinity),
# so two gates never contend with each other; the full gate and any other long job run idle-scheduled
# (cloud/box-run.sh idle), so they only get cycles the slots leave.
cpus=$(nproc --all)
range=$([ "${slot}" = 1 ] && echo "0-$((cpus / 2 - 1))" || echo "$((cpus / 2))-$((cpus - 1))")
echo "cpus=${range}" >> ~/"${out}"/box.txt
taskset -c "${range}" python3 ~/fast-gate/tools${suffix}/cloud/fast-gate/run.py --tree ~/fast-gate/tree${suffix} --sha "${sha}" --base "${base}" --tools ~/fast-gate/tools${suffix} --out ~/"${out}" --branch "${branch}" --branch-source "${branchSource}" --session "${session}" --session-source "${sessionSource}"
BOX
code=$?
set -e

local=$(mktemp -d)
scp -q -r "${box}:${out}" "${local}/fast"
# Logs only: anything over 5 MB (a binary that strayed into the out directory) is never published.
find "${local}/fast" -type f -size +5M -print -exec mv {} "${local}" \;
logBranch=gate-logs/${sha:0:12}/${stamp}/fast
index=$(mktemp -u)
gitDirectory=$(git -C "${here}" rev-parse --absolute-git-dir)
tree=$(cd "${local}/fast" && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" --work-tree=. add -A -f . && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" write-tree)
commit=$(git -C "${here}" commit-tree "${tree}" -m "Fast gate of ${sha} against main ${base} (tools ${tools}): $(head -1 "${local}/fast/status.txt")")
git -C "${here}" push -q origin "${commit}:refs/heads/${logBranch}"
rm -f "${index}"
echo "published ${logBranch} (${commit})"
head -1 "${local}/fast/status.txt"
exit ${code}

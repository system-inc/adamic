#!/usr/bin/env bash
# The fast gate: build, vet, the touched packages' tests, the oracle smoke set and the skip census,
# on the gate box, stopping at the first failure. Run it from any checkout with a push credential:
#
#   cloud/fast-gate.sh <full sha>
#
# The gate's own code comes from this script's checkout (its HEAD, which must be on origin), never
# from the candidate, so a candidate can't edit the gate that judges it. The base is origin/main at
# the moment it starts. The box keeps one tree and one tools checkout at fixed paths, so Go's build
# cache stays warm; tests always run -count=1. The result is published to
# gate-logs/<sha12>/<UTC stamp>/fast: status.txt (green: or red: on its first line), fast.json,
# test.jsonl and each step's log.
set -euo pipefail

sha=${1:?usage: cloud/fast-gate.sh <full sha>}
[[ ${sha} =~ ^[0-9a-f]{40}$ ]] || { echo "want a full 40-character sha, got ${sha}" >&2; exit 2; }
box=${ADAMIC_FAST_GATE_BOX:-threadripper}
here=$(cd "$(dirname "$0")/.." && pwd)
tools=$(git -C "${here}" rev-parse HEAD)
git -C "${here}" fetch -q origin
git -C "${here}" branch -r --contains "${tools}" | grep -q . || { echo "the gate's own commit ${tools} is not on origin; push it first" >&2; exit 2; }
base=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out=fast-gate/out/${sha:0:12}-${stamp}

echo "fast gate: ${sha} against main ${base}, tools ${tools}, on ${box}"
set +e
ssh "${box}" bash -s -- "${sha}" "${base}" "${tools}" "${out}" <<'BOX'
set -euo pipefail
sha=$1 base=$2 tools=$3 out=$4
mkdir -p ~/fast-gate
exec 9> ~/fast-gate/lock
flock 9
source ~/adamic-tools/env.sh
for directory in tools tree; do
  [ -d ~/fast-gate/${directory} ] || git clone -q https://github.com/system-inc/adamic.git ~/fast-gate/${directory}
done
git -C ~/fast-gate/tools fetch -q origin "${tools}"
git -C ~/fast-gate/tools switch -q --detach "${tools}"
git -C ~/fast-gate/tree fetch -q origin "${sha}" "${base}"
git -C ~/fast-gate/tree switch -q --detach "${sha}"
git -C ~/fast-gate/tree submodule update -q --init --recursive
mkdir -p ~/"${out}"
python3 ~/fast-gate/tools/cloud/fast-gate/run.py --tree ~/fast-gate/tree --sha "${sha}" --base "${base}" --tools ~/fast-gate/tools --out ~/"${out}"
BOX
code=$?
set -e

local=$(mktemp -d)
scp -q -r "${box}:${out}" "${local}/fast"
branch=gate-logs/${sha:0:12}/${stamp}/fast
index=$(mktemp -u)
gitDirectory=$(git -C "${here}" rev-parse --absolute-git-dir)
tree=$(cd "${local}/fast" && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" --work-tree=. add -A -f . && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" write-tree)
commit=$(git -C "${here}" commit-tree "${tree}" -m "Fast gate of ${sha} against main ${base} (tools ${tools}): $(head -1 "${local}/fast/status.txt")")
git -C "${here}" push -q origin "${commit}:refs/heads/${branch}"
rm -f "${index}"
echo "published ${branch} (${commit})"
head -1 "${local}/fast/status.txt"
exit ${code}

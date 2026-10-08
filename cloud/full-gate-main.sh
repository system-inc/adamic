#!/usr/bin/env bash
# The whole uncached gate on every new main, continuously, on the gate box, one run at a time and
# the newest main first. Run it from a checkout with a push credential (the box has none yet):
#
#   cloud/full-gate-main.sh            # loop forever
#   cloud/full-gate-main.sh <full sha> # one run of that commit, then exit
#
# Each run publishes to gate-logs/<sha12>/<UTC stamp>/full-main, and again every time its status
# line changes: the first failure turns status.txt red at once and is sent to
# @system_adamic_integration (landings pause on it); the rest of the run goes on for triage only,
# and full.json ("finished": true) lands when it ends. The run owns the box's last quarter of CPUs at
# normal priority (the fast gate's slots own the rest): under the idle class its own Node reference
# processes starved whenever the slots were busy, and a 1 s node deadline went red on main.
set -euo pipefail

box=${ADAMIC_FULL_GATE_BOX:-home}
boxes=${ADAMIC_FULL_GATE_BOXES:-${box}}
read -r -a box_list <<< "${boxes}"
sharded=false
[ "${#box_list[@]}" -gt 1 ] && sharded=true
[ "${#box_list[@]}" -eq 1 ] && box=${box_list[0]}
# Home is the whole gate's own box, so it takes every CPU (all). On a box shared with the fast gate's
# slots it takes the last quarter (quarter).
share=${ADAMIC_FULL_GATE_SHARE:-all}
# The last main whose full gate finished green: a red names the landings since, for bisecting.
state=${ADAMIC_FULL_GATE_STATE:-${HOME}/.adamic-full-gate}
mkdir -p "${state}"
here=$(cd "$(dirname "$0")/.." && pwd)
once=${1:-}

publish() {
  local sha=$1 stamp=$2 out=$3 parent=$4
  local copy index gitDirectory tree commit branch=gate-logs/${sha:0:12}/${stamp}/full-main
  copy=$(mktemp -d)
  if ${sharded}; then
    cp -R "${out}" "${copy}/full-main"
  else
    scp -q -r "${box}:${out}" "${copy}/full-main"
  fi
  # A log over 5 MB (test.jsonl on a whole run) is published gzipped; anything else that size (a binary
  # that strayed in) never is. Names go to stderr, never stdout, which a caller may be capturing.
  find "${copy}/full-main" -type f -size +5M \( -name '*.jsonl' -o -name '*.log' -o -name '*.txt' \) -exec gzip -9 {} \;
  find "${copy}/full-main" -type f -size +5M -exec mv {} "${copy}" \; -print >&2
  index=$(mktemp -u)
  gitDirectory=$(git -C "${here}" rev-parse --absolute-git-dir)
  tree=$(cd "${copy}/full-main" && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" --work-tree=. add -A -f . && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" write-tree)
  commit=$(git -C "${here}" commit-tree "${tree}" ${parent:+-p "${parent}"} -m "Full gate of main ${sha}: $(head -1 "${copy}/full-main/status.txt")")
  git -C "${here}" push -q origin "${commit}:refs/heads/${branch}"
  echo "${commit}"
}

run() {
  local sha=$1 tools stamp out status previous="" parent="" coordinator=""
  tools=$(git -C "${here}" rev-parse HEAD)
  stamp=$(date -u +%Y%m%dT%H%M%SZ)
  out=full-gate/out/${sha:0:12}-${stamp}
  echo "$(date -u +%H:%M:%S) full gate of main ${sha} (tools ${tools}) on ${box}"
  if ${sharded}; then
    out=${state}/out/${sha:0:12}-${stamp}
    mkdir -p "${out}"
    echo "running: full gate of ${sha}, preparing shards" > "${out}/status.txt"
    python3 "${here}/cloud/fast-gate/shards.py" run --boxes "${boxes}" --sha "${sha}" --tools "${tools}" --out "${out}" > "${out}/coordinator.log" 2>&1 &
    coordinator=$!
  else
  ssh "${box}" bash -s -- "${sha}" "${tools}" "${out}" "${share}" <<'BOX'
set -euo pipefail
sha=$1 tools=$2 out=$3 share=${4:-all}
mkdir -p ~/full-gate ~/"${out}"
for directory in tools tree; do
  [ -d ~/full-gate/${directory} ] || git clone -q https://github.com/system-inc/adamic.git ~/full-gate/${directory}
done
cat > ~/"${out}"/run.sh <<RUN
set -u
exec 9> ~/full-gate/lock
flock 9
source ~/adamic-tools/env.sh
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "\${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "${tools}" && git -C ~/full-gate/tools switch -q --detach "${tools}"
git -C ~/full-gate/tree fetch -q origin "${sha}" && git -C ~/full-gate/tree switch -q --detach "${sha}"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=\$(nproc --all)
first=\$([ "${share}" = quarter ] && echo \$((cpus * 3 / 4)) || echo 0)
taskset -c "\$first-\$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "${sha}" --base "${sha}" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"${out}"
RUN
echo "running: full gate of ${sha}, waiting for the box" > ~/"${out}"/status.txt
tmux new -d -s "full-${sha:0:12}" "bash ~/${out}/run.sh > ~/${out}/driver.log 2>&1"
BOX
  fi
  while true; do
    sleep 30
    if ${sharded}; then
      if ! kill -0 "${coordinator}" 2>/dev/null && [ ! -f "${out}/full.json" ]; then
        echo "red: ${sha} full gate, first failure at coordinator (no final report)" > "${out}/status.txt"
        printf '{"finished":true,"failure":{"step":"coordinator"}}\n' > "${out}/full.json"
      fi
      status=$(head -1 "${out}/status.txt" 2>/dev/null || true)
    else
      status=$(ssh "${box}" "head -1 ~/${out}/status.txt" 2>/dev/null || true)
    fi
    if [ -n "${status}" ] && [ "${status}" != "${previous}" ]; then
      parent=$(publish "${sha}" "${stamp}" "${out}" "${parent}")
      echo "$(date -u +%H:%M:%S) ${status}"
      if [[ ${status} == red:* && ${previous} != red:* ]]; then
        green=$(cat "${state}/last-green" 2>/dev/null || true)
        if [ -n "${green}" ]; then
          range="landings since the last green full gate (${green:0:12}): $(git -C "${here}" log --first-parent --format=%h "${green}..${sha}" 2>/dev/null | tr '\n' ' ')"
        else
          range="no green full gate recorded yet, so no range to bisect"
        fi
        (cd /Users/kirkouimet/Projects/ahra && ahra os send system_adamic_integration "Full gate of main ${sha} is red: ${status}. Log: gate-logs/${sha:0:12}/${stamp}/full-main (first-failure.txt). ${range}. Bisect those on a fast slot. The rest keeps running for triage." >/dev/null 2>&1 || true)
      fi
      previous=${status}
    fi
    if { ${sharded} && [ -f "${out}/full.json" ]; } || { ! ${sharded} && ssh "${box}" "test -f ~/${out}/full.json" 2>/dev/null; }; then
      parent=$(publish "${sha}" "${stamp}" "${out}" "${parent}")
      if ${sharded}; then
        final=$(head -1 "${out}/status.txt")
        wait "${coordinator}" || true
      else
        final=$(ssh "${box}" "head -1 ~/${out}/status.txt")
      fi
      echo "$(date -u +%H:%M:%S) finished: ${final}"
      [[ ${final} == green:* ]] && echo "${sha}" > "${state}/last-green"
      return
    fi
  done
}

if [ -n "${once}" ]; then
  run "${once}"
  exit
fi
while true; do
  git -C "${here}" fetch -q origin
  main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
  if ! git -C "${here}" ls-remote origin "refs/heads/gate-logs/${main:0:12}/*" | grep -q '/full-main$'; then
    run "${main}"
  fi
  sleep 60
done

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
# Home is the whole gate's own box, so it takes every CPU (all). On a box shared with the fast gate's
# slots it takes the last quarter (quarter).
share=${ADAMIC_FULL_GATE_SHARE:-all}
# The last main whose full gate finished green: a red names the landings since, for bisecting.
state=${ADAMIC_FULL_GATE_STATE:-${HOME}/.adamic-full-gate}
mkdir -p "${state}"
here=$(cd "$(dirname "$0")/.." && pwd)
once=${1:-}

# The heartbeat (@system_adamic, Oct 8 03:36): one row when a run starts and one every 10 minutes
# while the loop lives, to documentation/velocity/full-gate-heartbeat.csv on records/full-gate-heartbeat,
# so the witness charts the age of the last started whole gate from an artifact. A dead loop is a gap.
heartbeat() {
  local event=$1 sha=${2:-} records=${state}/heartbeat-records file=documentation/velocity/full-gate-heartbeat.csv
  if [ ! -d "${records}" ]; then
    if git -C "${here}" ls-remote --exit-code origin refs/heads/records/full-gate-heartbeat > /dev/null; then
      git -C "${here}" fetch -q origin records/full-gate-heartbeat && git -C "${here}" worktree add -q --detach "${records}" FETCH_HEAD
    else
      git -C "${here}" worktree add -q --detach "${records}" "$(git -C "${here}" commit-tree "$(git -C "${here}" hash-object -t tree /dev/null)" -m "Start the whole gate's heartbeat")"
    fi
  fi
  mkdir -p "$(dirname "${records}/${file}")"
  [ -f "${records}/${file}" ] || echo "utc,event,main,box,last_started_utc,last_started_main" > "${records}/${file}"
  [ "${event}" = start ] && echo "$(date -u +%FT%TZ) ${sha}" > "${state}/last-started"
  echo "$(date -u +%FT%TZ),${event},${sha},${box},$(cut -d' ' -f1 "${state}/last-started" 2>/dev/null),$(cut -d' ' -f2 "${state}/last-started" 2>/dev/null)" >> "${records}/${file}"
  git -C "${records}" add "${file}"
  git -C "${records}" -c user.name=kirkouimet -c user.email=kirk@kirkouimet.com commit -q -m "Whole-gate heartbeat: ${event} ${sha:0:12}" -m "Co-Authored-By: Ahra <ahra@ahra.ai>"
  git -C "${records}" push -q origin "HEAD:refs/heads/records/full-gate-heartbeat" || true
}

publish() {
  local sha=$1 stamp=$2 out=$3 parent=$4
  local copy index gitDirectory tree commit branch=gate-logs/${sha:0:12}/${stamp}/full-main
  copy=$(mktemp -d)
  scp -q -r "${box}:${out}" "${copy}/full-main"
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
  local sha=$1 tools stamp out status previous="" parent=""
  tools=$(git -C "${here}" rev-parse HEAD)
  stamp=$(date -u +%Y%m%dT%H%M%SZ)
  out=full-gate/out/${sha:0:12}-${stamp}
  echo "$(date -u +%H:%M:%S) full gate of main ${sha} (tools ${tools}) on ${box}"
  (heartbeat start "${sha}" > /dev/null 2>&1 &)
  ssh "${box}" bash -s -- "${sha}" "${tools}" "${out}" "${share}" <<'BOX'
set -euo pipefail
sha=$1 tools=$2 out=$3 share=${4:-all}
mkdir -p ~/full-gate ~/"${out}"
for directory in tools tree; do
  [ -d ~/full-gate/${directory} ] || git clone -q https://github.com/system-inc/adamic.git ~/full-gate/${directory}
done
cat > ~/"${out}"/run.sh <<RUN
set -eu
exec 9> ~/full-gate/lock
flock 9
# BEGIN idle preemption (keep identical to cloud/idle-preempt.sh)
python3 - <<'IDLE_PREEMPT'
import os, signal, time

def marked():
    result = []
    for name in os.listdir('/proc'):
        if not name.isdigit():
            continue
        try:
            path = '/proc/' + name
            if os.stat(path).st_uid != os.getuid():
                continue
            if b'ADAMIC_IDLE_JOB=1' in open(path + '/environ', 'rb').read().split(b'\\0'):
                if open(path + '/stat').read().rsplit(')', 1)[1].split()[0] != 'Z':
                    result.append(int(name))
        except (OSError, ProcessLookupError):
            pass
    return result

# SIGKILL prevents a dying parent from spawning new children after the scan.
# Repeat for children born between scanning /proc and killing their parent.
deadline = time.monotonic() + 10
while True:
    pids = marked()
    if not pids:
        break
    for pid in pids:
        try:
            os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    if time.monotonic() >= deadline:
        raise SystemExit('idle preemption timed out; gate aborted')
    time.sleep(.05)
IDLE_PREEMPT
# END idle preemption
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
  local beats=0
  while true; do
    sleep 30
    beats=$((beats + 1)); [ $((beats % 20)) -eq 0 ] && (heartbeat running "${sha}" > /dev/null 2>&1 &)
    status=$(ssh "${box}" "head -1 ~/${out}/status.txt" 2>/dev/null || true)
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
    if ssh "${box}" "test -f ~/${out}/full.json" 2>/dev/null; then
      parent=$(publish "${sha}" "${stamp}" "${out}" "${parent}")
      final=$(ssh "${box}" "head -1 ~/${out}/status.txt")
      echo "$(date -u +%H:%M:%S) finished: ${final}"
      if [[ ${final} == green:* ]]; then echo "${sha}" > "${state}/last-green"; fi
      # Explicitly 0: a bare return took the status of the green test above, so after every red run
      # (06:55Z and 08:30Z on Oct 8) set -e ended the loop and no main got a whole gate.
      return 0
    fi
  done
}

if [ -n "${once}" ]; then
  run "${once}"
  exit
fi
while true; do
  # A network blip or one failed run never ends the loop: it says so and tries again next minute.
  main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1) || main=""
  if [ -n "${main}" ] && ! git -C "${here}" ls-remote origin "refs/heads/gate-logs/${main:0:12}/*" | grep -q '/full-main$'; then
    run "${main}" || echo "$(date -u +%H:%M:%S) run of ${main} failed (exit $?); trying again next minute"
  fi
  idle=$((${idle:-0} + 1)); [ $((idle % 10)) -eq 0 ] && (heartbeat idle "${main}" > /dev/null 2>&1 &)
  sleep 60
done

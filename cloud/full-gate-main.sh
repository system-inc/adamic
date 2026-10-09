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
here=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
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

# Between whole gates the box lends one area slot to the fast-gate watcher (@system_adamic, Oct 8 05:22),
# and a whole gate takes it back first: the line "<box> B" comes out of the watcher's slot table, any
# fast gate still on the box is stopped (the watcher sees it void and runs it again elsewhere), and the
# run waits for that slot's lock. Lent again when the run's verdict is published, and on every idle minute.
# A no-lend file beside the slot table keeps the box out of the watcher's hands between runs, so it can be given
# to other work (Loom's pilot, @system_adamic, Oct 8 21:47Z).
slots=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}/slots
# The box's lines are all of them, small slots too (a box shared with the fast gate, like the Threadripper, has
# both): reclaim keeps them in ${state}/lent-lines and takes them out, lend puts them back ("<box> B" if none kept).
lend() {
  [ -f "${slots}" ] || return 0
  [ -f "$(dirname "${slots}")/no-lend" ] && return 0
  grep -q "^${box} " "${slots}" && return 0
  if [ -s "${state}/lent-lines" ]; then cat "${state}/lent-lines" >> "${slots}"; else echo "${box} B" >> "${slots}"; fi
}
reclaim() {
  [ -f "${slots}" ] || return 0
  grep "^${box} " "${slots}" > "${state}/lent-lines.tmp" && mv "${state}/lent-lines.tmp" "${state}/lent-lines"
  rm -f "${state}/lent-lines.tmp"
  grep -v "^${box} " "${slots}" > "${slots}.tmp"; mv "${slots}.tmp" "${slots}"
  ssh "${box}" 'pkill -f "[f]ast-gate/run.py --tree" || true; mkdir -p ~/fast-gate; flock ~/fast-gate/lock true' || true
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

# A request's run stops the moment its candidate leaves the requests file (integration's train drops superseded
# slices, @system_adamic, Oct 8 23:13Z): a superseded candidate's verdict, red triage included, is spent, and its box
# goes to the bottom unfinished line.
requestDropped() {
  [ -f "${requests}" ] && ! grep -qx "$1" "${requests}"
}
stopRemote() {
  local sha=$1
  [[ ${sha} =~ ^[0-9a-f]{40}$ ]] || return 1
  ssh "${box}" "pkill -TERM -f 'run.py .*--sha ${sha}'" || true
}
# One whole gate per box, ever (@system_adamic, Oct 9 00:13Z: the Threadripper's requests loop took e37ea9dc onto the
# box while a hand-started rerun of 6b2c73f9 ran there, load 145 on 64 cores). Every run holds the box's claim, a
# directory per box in one place for every loop and hand-started run alike, with the holder's pid; a claim whose pid
# is gone is taken over. A loop starts nothing on a box someone else holds.
boxClaims=${ADAMIC_FULL_GATE_BOXES:-${HOME}/.adamic-full-gate/boxes}
claimBox() {
  local holder pid
  mkdir -p "${boxClaims}"
  if mkdir "${boxClaims}/${box}" 2> /dev/null; then
    echo "$$ $1" > "${boxClaims}/${box}/holder"
    return 0
  fi
  read -r pid holder < "${boxClaims}/${box}/holder" 2> /dev/null || { echo "$$ $1" > "${boxClaims}/${box}/holder"; return 0; }
  [ "${pid}" = $$ ] && { echo "$$ $1" > "${boxClaims}/${box}/holder"; return 0; }
  kill -0 "${pid}" 2> /dev/null && return 1
  echo "$$ $1" > "${boxClaims}/${box}/holder"
}
releaseBox() {
  local pid rest
  read -r pid rest < "${boxClaims}/${box}/holder" 2> /dev/null || return 0
  [ "${pid}" = $$ ] || return 0
  rm -f "${boxClaims}/${box}/holder"
  rmdir "${boxClaims}/${box}" 2> /dev/null || true
}
boxFree() {
  local pid rest
  read -r pid rest < "${boxClaims}/${box}/holder" 2> /dev/null || return 0
  [ "${pid}" = $$ ] || ! kill -0 "${pid}" 2> /dev/null
}
# ADAMIC_FULL_GATE_RUN_TO_END=1: a parity proof, run to the end after its first failure (run.py --run-to-end).
run() {
  local code
  if ! claimBox "$1"; then
    echo "$(date -u +%H:%M:%S) not starting $1: ${box} is running another whole gate ($(cat "${boxClaims}/${box}/holder" 2> /dev/null))"
    return 2
  fi
  runOnBox "$@"
  code=$?
  releaseBox
  return "${code}"
}
runOnBox() {
  local sha=$1 origin=${2:-main} tools stamp out status previous="" parent="" stopping=""
  tools=$(git -C "${here}" rev-parse HEAD)
  stamp=$(date -u +%Y%m%dT%H%M%SZ)
  out=full-gate/out/${sha:0:12}-${stamp}
  echo "$(date -u +%H:%M:%S) full gate of main ${sha} (tools ${tools}) on ${box}"
  reclaim
  (heartbeat start "${sha}" > /dev/null 2>&1 &)
  ssh "${box}" bash -s -- "${sha}" "${tools}" "${out}" "${share}" "$([ "${ADAMIC_FULL_GATE_RUN_TO_END:-}" = 1 ] && echo --run-to-end)" <<'BOX'
set -euo pipefail
sha=$1 tools=$2 out=$3 share=${4:-all} runToEnd=${5:-}
mkdir -p ~/full-gate ~/"${out}"
for directory in tools tree; do
  [ -d ~/full-gate/${directory} ] || git clone -q https://github.com/system-inc/adamic.git ~/full-gate/${directory}
done
cat > ~/"${out}"/run.sh <<RUN
set -u
exec 9> ~/full-gate/lock
flock 9
# A run that can't stop the box's idle jobs is void, naming the box: under set -e it would end with no
# full.json, and the loop waits for that file without a deadline.
preempt() {
# BEGIN idle preemption (keep identical to cloud/idle-preempt.sh)
python3 - <<'IDLE_PREEMPT'
import datetime, os, signal, time

started = time.monotonic()
deadline = started + 10
sessions, found, survivors = set(), set(), {}

def processes():
    result = {}
    for name in os.listdir('/proc'):
        if not name.isdigit():
            continue
        try:
            path = '/proc/' + name
            with open(path + '/stat') as handle:
                fields = handle.read().rsplit(')', 1)[1].split()
            # Fields after comm start at stat field 3: session is 6, starttime 22.
            if fields[0] not in ('Z', 'X'):
                result[int(name)] = (int(fields[3]), fields[19], os.stat(path).st_uid)
        except (OSError, ProcessLookupError):
            pass
    return result

def marked():
    result = {}
    for pid, info in processes().items():
        if info[2] != os.getuid():
            continue
        try:
            with open('/proc/%d/environ' % pid, 'rb') as handle:
                is_marked = b'ADAMIC_IDLE_JOB=1' in handle.read().split(b'\\0')
            if is_marked:
                result[pid] = info
                sessions.add(info[0])
                found.add((pid, info[1]))
        except (OSError, ProcessLookupError):
            pass
    return result

def in_sessions():
    # Never the preempting gate's own session: a marked process started from it must not take the gate down.
    return {pid: info for pid, info in processes().items() if info[0] in sessions and info[0] != os.getsid(0)}

def kill(pids):
    current = processes()
    for pid, info in pids.items():
        # Do not signal an unrelated process if a PID has been reused.
        if current.get(pid) != info:
            continue
        try:
            os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass

def disable(pids):
    new = {pid: info for pid, info in pids.items() if (pid, info[1]) not in survivors}
    if not new:
        return
    directory = os.path.expanduser('~/idle')
    os.makedirs(directory, exist_ok=True)
    with open(directory + '/disabled', 'a') as report:
        report.write(datetime.datetime.now(datetime.timezone.utc).isoformat() +
                     ' idle session survivors; human review required\\n')
        for pid, info in sorted(new.items()):
            try:
                with open('/proc/%d/cmdline' % pid, 'rb') as handle:
                    command = handle.read().replace(b'\\0', b' ').decode(errors='replace')
            except OSError:
                command = '<exited before command line could be read>'
            report.write('pid=%d session=%d command=%r\\n' % (pid, info[0], command))
            survivors[(pid, info[1])] = info
        report.flush()
        os.fsync(report.fileno())

try:
    # Kill marked parents before they can spawn more children; repeat for races.
    while True:
        pids = marked()
        if not pids:
            break
        kill(pids)
        if time.monotonic() >= deadline:
            break
        time.sleep(.05)
    # Check every live process in each recorded session, even without the marker.
    # A survivor disables further idle launches before we attempt to kill it.
    while True:
        pids = in_sessions()
        disable(pids)
        kill(pids)
        remaining = marked()
        remaining.update(in_sessions())
        if not remaining:
            break
        disable(remaining)
        if time.monotonic() >= deadline:
            raise SystemExit('idle preemption timed out; gate aborted')
        time.sleep(.05)
finally:
    elapsed = int((time.monotonic() - started) * 1000)
    print('idle preemption: %d marked, %d survivors in their sessions, %d ms' %
          (len(found), len(survivors), elapsed), flush=True)
IDLE_PREEMPT
# END idle preemption
}
if ! preempt; then
  echo "void: ${sha} full gate, box \$(hostname) could not stop its idle jobs in 10 s" > ~/"${out}"/status.txt
  echo '{"finished": true, "void": true}' > ~/"${out}"/full.json
  exit 0
fi
source ~/adamic-tools/env.sh
# Stock tsc is the gates' pinned TypeScript 6.0.3 (the LKG checkout setup provides), for oracles that
# take it from PATH (cmd/adamic-test262's stock-rejection check). No box had a tsc on PATH.
export PATH="\${ADAMIC_TYPESCRIPT_SOURCE}/bin:\${PATH}"
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "\${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "${tools}" && git -C ~/full-gate/tools switch -q --detach "${tools}"
# A declared tool the box lacks (cloud/fast-gate/tools.txt) makes the run void, naming the box, never red.
if ! lacks=\$(bash ~/full-gate/tools/cloud/fast-gate/tools-check.sh); then
  echo "\${lacks}" > ~/"${out}"/tools-missing.txt
  echo "void: ${sha} full gate, box \$(hostname) lacks a declared tool: \$(echo "\${lacks}" | sed -E 's/^lacks ([^:]+):.*/\1/' | tr '\n' ' ')" > ~/"${out}"/status.txt
  echo '{"finished": true, "void": true}' > ~/"${out}"/full.json
  exit 0
fi
git -C ~/full-gate/tree fetch -q origin "${sha}" && git -C ~/full-gate/tree switch -q --detach "${sha}"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=\$(nproc --all)
first=\$([ "${share}" = quarter ] && echo \$((cpus * 3 / 4)) || echo 0)
taskset -c "\$first-\$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "${sha}" --base "${sha}" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"${out}" ${runToEnd}
RUN
echo "running: full gate of ${sha}, waiting for the box" > ~/"${out}"/status.txt
tmux new -d -s "full-${sha:0:12}" "bash ~/${out}/run.sh > ~/${out}/driver.log 2>&1"
BOX
  local beats=0
  while true; do
    sleep 30
    beats=$((beats + 1)); [ $((beats % 20)) -eq 0 ] && (heartbeat running "${sha}" > /dev/null 2>&1 &)
    if [ "${origin}" = request ] && [ -z "${stopping}" ] && requestDropped "${sha}"; then
      echo "$(date -u +%H:%M:%S) stopping ${sha}: it left the requests file (superseded); ${box} goes to the next request"
      stopRemote "${sha}"
      stopping=yes
    fi
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
        if [ "${origin}" = request ]; then
          (cd /Users/kirkouimet/Projects/ahra && ahra os send system_adamic_integration "Whole gate of the star's candidate ${sha} (a request, not main) is red: ${status}. Log: gate-logs/${sha:0:12}/${stamp}/full-main (first-failure.txt). It runs on for triage while the candidate stays in the requests file." >/dev/null 2>&1 || true)
        else
          (cd /Users/kirkouimet/Projects/ahra && ahra os send system_adamic_integration "Full gate of main ${sha} is red: ${status}. Log: gate-logs/${sha:0:12}/${stamp}/full-main (first-failure.txt). ${range}. Bisect those on a fast slot. The rest keeps running for triage." >/dev/null 2>&1 || true)
        fi
        # A red that may be load (a timeout, a stall, a kill) gets Loom's same-instance A/B against the candidate's first
        # parent, and the verdict comes back to integration (cloud/ab-trigger.py; @system_adamic from the witness, Oct 9).
        firstFailure=$(mktemp)
        if ssh "${box}" "cat ~/${out}/first-failure.txt" > "${firstFailure}" 2> /dev/null; then
          git -C "${here}" fetch -q origin "${sha}" 2> /dev/null || true
          python3 "${here}/cloud/ab-trigger.py" write --candidate "${sha}" --main "$(git -C "${here}" rev-parse "${sha}^1" 2> /dev/null)" \
            --red "gate-logs/${sha:0:12}/${stamp}/full-main" --first-failure "${firstFailure}" --notify system_adamic_integration || true
        fi
        rm -f "${firstFailure}"
      fi
      previous=${status}
    fi
    if ssh "${box}" "test -f ~/${out}/full.json" 2>/dev/null; then
      parent=$(publish "${sha}" "${stamp}" "${out}" "${parent}")
      final=$(ssh "${box}" "head -1 ~/${out}/status.txt")
      echo "$(date -u +%H:%M:%S) finished: ${final}"
      if [[ ${final} == green:* ]]; then echo "${sha}" > "${state}/last-green"; fi
      if [[ ${final} == void:* ]]; then
        (cd /Users/kirkouimet/Projects/ahra && ahra os send system_adamic_developer_tools "Whole gate of main ${sha} is void, a box problem, not the change: ${final}. Fix the box and run it again (cloud/full-gate-main.sh ${sha})." >/dev/null 2>&1 || true)
      fi
      lend
      # Explicitly 0: a bare return took the status of the green test above, so after every red run
      # (06:55Z and 08:30Z on Oct 8) set -e ended the loop and no main got a whole gate.
      return 0
    fi
  done
}

# A record-only main (@system_adamic, Oct 8 22:48Z: 031a1259 took Home for 40 minutes to confirm a csv row) is
# confirmed by the green whole gate of the commit it differs from only in push-main's record paths, the same three
# cloud/integration/push-main.sh and record-paths-test.py hold: it is logged as confirmed and never run.
recordOnly() {
  local changed
  changed=$(git -C "${here}" diff --name-only "$1" "$2") || return 1
  ! printf '%s\n' "${changed}" | grep -v -e '^documentation/velocity/landings\.csv$' -e '^stage3/meter/runs/' -e '^stage3/progress\.json$' | grep -q .
}
# A commit's newest finished whole gate: green, red or void; running while one is unfinished; none without a record.
# Whose records: box (a whole gate on a box, the default), pool (Loom's pool, full.json "runner": "pool"), or any.
recordState() {
  local sha=$1 which=${2:-box} ref refs status finished runner
  refs=$(git -C "${here}" ls-remote origin "refs/heads/gate-logs/${sha:0:12}/*" | awk '$2 ~ /\/full-main$/ {print $2}' | sort -r) || true
  for ref in ${refs}; do
    git -C "${here}" fetch -q origin "${ref}" || continue
    runner=box
    git -C "${here}" show FETCH_HEAD:full.json 2> /dev/null | grep -q '"runner": "pool"' && runner=pool
    [ "${which}" = any ] || [ "${which}" = "${runner}" ] || continue
    status=$(git -C "${here}" show FETCH_HEAD:status.txt 2> /dev/null | head -1)
    finished=$(git -C "${here}" show FETCH_HEAD:full.json 2> /dev/null | grep -c '"finished": true' || true)
    break
  done
  [ -n "${status:-}" ] || { echo none; return; }
  case ${status} in
    void:*) echo void ;;
    green:*) [ "${finished}" -gt 0 ] && echo green || echo running ;;
    red:*) [ "${finished}" -gt 0 ] && echo red || echo running ;;
    *) echo running ;;
  esac
}
# The commit among main's last 20 whose green whole gate confirms main, or nothing.
confirmedBy() {
  local main=$1 candidate
  git -C "${here}" fetch -q origin "${main}" 2> /dev/null || true
  for candidate in $(git -C "${here}" rev-list -n 20 "${main}"); do
    [ "${candidate}" = "${main}" ] && continue
    recordOnly "${candidate}" "${main}" || continue
    [ "$(recordState "${candidate}" "$(poolPromoted && echo any || echo box)")" = green ] && { echo "${candidate}"; return 0; }
  done
  return 1
}
# Integration's requests (cloud/integration/star-train.py, Kirk, Oct 8): the star's train slices, bottom first. The
# loop takes the first whose whole gate hasn't finished green or red (a void runs again) between mains, and drops
# a line once its record finishes.
requests=${ADAMIC_FULL_GATE_REQUESTS:-${state}/requests}
# More than one loop can serve the requests (@system_adamic, Oct 8 22:58Z: Home takes the bottom slice, a second
# 64-core box the next). Each claims a line before running it: a directory per sha (mkdir is atomic), holding the
# claiming box and pid. A claim whose pid is gone is taken over. Only the loop whose role is all also gates mains.
claims=${ADAMIC_FULL_GATE_CLAIMS:-$(dirname "${requests}")/claims}
role=${ADAMIC_FULL_GATE_ROLE:-all}
claim() {
  local sha=$1 holder pid
  mkdir -p "${claims}"
  if mkdir "${claims}/${sha}" 2> /dev/null; then
    echo "${box} $$" > "${claims}/${sha}/holder"
    return 0
  fi
  read -r holder pid < "${claims}/${sha}/holder" 2> /dev/null || return 1
  [ "${pid}" = $$ ] && return 0
  kill -0 "${pid}" 2> /dev/null && return 1
  echo "${box} $$" > "${claims}/${sha}/holder"
}
release() {
  rm -f "${claims}/$1/holder"
  rmdir "${claims}/$1" 2> /dev/null || true
}
# Loom's pre-gate (@system_adamic, Oct 8 23:36Z): the packages that turned the star red, every unit in parallel on the
# warm pool, before a box takes the candidate. Loom writes ${pregates}/<sha>, first line green, red or running. A red
# candidate never takes a box (it goes back to its owner with the whole list) and a running one waits; with no file
# (no pre-gate for it) the candidate is eligible, so a box never idles on a pre-gate that doesn't exist.
pregates=${ADAMIC_FULL_GATE_PREGATES:-$(dirname "${requests}")/pregate}
pregateAllows() {
  local verdict
  [ -f "${pregates}/$1" ] || return 0
  verdict=$(head -1 "${pregates}/$1")
  [ "${verdict%% *}" = green ]
}
# Loom's pool as a landing gate (@system_adamic, Oct 8 23:51Z). Until the parent rules parity proven
# (${state}/pool-promoted, written by hand then), every request also runs on a box, whatever the pool did, so the two
# can be compared test for test. Once promoted, a request the pool finished takes a box only as the spot-check: about
# one in five, chosen by its sha so every loop agrees.
poolPromoted() {
  [ -f "$(dirname "${requests}")/pool-promoted" ]
}
spotCheck() {
  [ $(( 16#${1:0:8} % 5 )) -eq 0 ]
}
nextRequest() {
  local sha
  [ -s "${requests}" ] || return 1
  while read -r sha; do
    [ -n "${sha}" ] || continue
    pregateAllows "${sha}" || continue
    case $(recordState "${sha}" box) in
      none | void) ;;
      *) continue ;;
    esac
    if poolPromoted && ! spotCheck "${sha}"; then
      case $(recordState "${sha}" pool) in
        green | red) continue ;;
      esac
    fi
    claim "${sha}" && { echo "${sha}"; return 0; }
  done < "${requests}"
  return 1
}
dropRequest() {
  local sha=$1
  case $(recordState "${sha}") in
    green | red) grep -vx "${sha}" "${requests}" > "${requests}.tmp"; mv "${requests}.tmp" "${requests}" ;;
  esac
}
# A whole gate started by hand (cloud/full-gate-main.sh <sha>, on Home unless told otherwise) owns Home: Home's loop
# neither runs nor lends while one runs.
oneOffRunning() {
  [ "${box}" = home ] && pgrep -f "full-gate-main\.sh [0-9a-f]{40}" > /dev/null 2>&1
}

[ "${ADAMIC_FULL_GATE_LIBRARY:-}" = 1 ] && return 0
if [ -n "${once}" ]; then
  run "${once}"
  exit
fi
while true; do
  if oneOffRunning || ! boxFree; then
    sleep 60
    continue
  fi
  # A network blip or one failed run never ends the loop: it says so and tries again next minute.
  main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1) || main=""
  # A main with no whole gate, or only a void one, is confirmed or run; one that is running or finished is done.
  mainState=$( [ -n "${main}" ] && [ "${role}" = all ] && recordState "${main}" || echo unknown)
  if [ "${role}" = all ] && { [ "${mainState}" = none ] || [ "${mainState}" = void ]; } && ! grep -qx "${main}" "${state}/confirmed" 2> /dev/null; then
    if confirmed=$(confirmedBy "${main}"); then
      echo "${main}" >> "${state}/confirmed"
      echo "$(date -u +%H:%M:%S) full gate of main ${main} confirmed by ${confirmed}'s green whole gate: they differ only in record paths, so it isn't run"
    else
      run "${main}" || echo "$(date -u +%H:%M:%S) run of ${main} failed (exit $?); trying again next minute"
    fi
  elif request=$(nextRequest); then
    echo "$(date -u +%H:%M:%S) the star's request ${request} takes ${box} between mains"
    run "${request}" request || echo "$(date -u +%H:%M:%S) run of ${request} failed (exit $?); trying again next minute"
    dropRequest "${request}"
    release "${request}"
  fi
  lend
  idle=$((${idle:-0} + 1)); [ $((idle % 10)) -eq 0 ] && (heartbeat idle "${main}" > /dev/null 2>&1 &)
  sleep 60
done

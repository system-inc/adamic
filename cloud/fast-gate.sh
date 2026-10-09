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
# the moment it starts, or for a worker's branch cut from an area, that area's tip (gateBase in
# cloud/fast-gate-classify.sh), so it tests its own change rather than the area's. The box keeps one tree and one tools checkout at fixed paths, so Go's build
# cache stays warm; tests always run -count=1. The result is published to
# gate-logs/<sha12>/<UTC stamp>/fast: status.txt (green: or red: on its first line), fast.json,
# test.jsonl and each step's log.
set -euo pipefail

sha=${1:?usage: cloud/fast-gate.sh <full sha> [--branch <name>] [--session <id>]}
shift
branch="" branchSource=given session="" sessionSource=given cpus="" class="" wholeBox=""
while [ $# -gt 0 ]; do
  case $1 in
    --branch) branch=$2; shift 2 ;;
    --session) session=$2; shift 2 ;;
    --cpus) cpus=$2; shift 2 ;;
    --class) class=$2; shift 2 ;;
    # A reserved gate that holds its box alone runs on every CPU of it, not only its slot's share.
    --whole-box) wholeBox=1; shift ;;
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
. "${here}/cloud/fast-gate-classify.sh"
read -r baseName base <<< "$(gateBase "${branch:-}" "${sha}")"
# The watcher passes the class it queued with; a direct gate (integration's landings) gets the same judgment.
[ -n "${class}" ] || class=$(classify "${branch:-}" "${sha}")
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out=fast-gate/out/${sha:0:12}-${stamp}

echo "fast gate: ${sha} against ${baseName} ${base}, tools ${tools}, on ${box}, class ${class}"
# Landings and areas also compile on macOS (cloud/darwin-leg.sh), beside the box's gate: a Darwin-only
# compile break (area-next 517cb633, st_atimespec) got through when nothing on macOS gated (Oct 8).
darwinLog=""
if [[ ${branch} == cloud/land-* || ${branch} == area/* ]]; then
  darwinLog=$(mktemp)
  bash "${here}/cloud/darwin-leg.sh" "${sha}" "${branch}" > "${darwinLog}" 2>&1 &
  darwinPid=$!
fi
set +e
# ssh joins its arguments into one remote command line, so each is quoted for the remote shell.
ssh "${box}" bash -s -- "$(printf '%q ' "${sha}" "${base}" "${tools}" "${out}" "${branch:-}" "${branchSource}" "${session:-}" "${sessionSource}" "${cpus:-}" "${class}" "${baseName}" "${wholeBox}")" <<'BOX'
set -euo pipefail
sha=$1 base=$2 tools=$3 out=$4 branch=$5 branchSource=$6 session=$7 sessionSource=$8 width=${9:-} class=${10:-B} baseName=${11:-main} wholeBox=${12:-}
mkdir -p ~/fast-gate
# Two slots, each with its own tree and tools checkout, so a small change doesn't wait behind a
# stack's long gate; the second slot's tree starts as a copy of the first (submodules included).
# With both busy it takes whichever frees first, never waiting on one while the other is idle.
slot=""
while [ -z "${slot}" ]; do
  # Slot 1 is the area slot (landings and anything big); slots 2 and 3 take small changes only
  # (@system_adamic, Oct 8 00:10), so a worker's tip never waits behind an area.
  candidates=$([ "${class}" = S ] && echo "2 3" || echo "1")
  # The last quarter was the whole gate's until it moved to Home (Oct 8 07:51); on a 64-CPU box it is a
  # third small slot now, measured idle on Cloud and Workshop (0% busy over 10 s, nothing pinned).
  [ "${class}" = S ] && [ "$(nproc --all)" -ge 64 ] && candidates="2 3 4"
  # A box under 32 CPUs (Chonchon, 16) is one slot on all its CPUs, whatever the class.
  [ "$(nproc --all)" -lt 32 ] && candidates=1
  for candidate in ${candidates}; do
    exec 9> ~/fast-gate/lock$([ "${candidate}" = 1 ] && echo "" || echo "-${candidate}")
    if flock -n 9; then slot=${candidate}; break; fi
    exec 9>&-
  done
  [ -n "${slot}" ] || sleep 1
done
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
                is_marked = b'ADAMIC_IDLE_JOB=1' in handle.read().split(b'\0')
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
                     ' idle session survivors; human review required\n')
        for pid, info in sorted(new.items()):
            try:
                with open('/proc/%d/cmdline' % pid, 'rb') as handle:
                    command = handle.read().replace(b'\0', b' ').decode(errors='replace')
            except OSError:
                command = '<exited before command line could be read>'
            report.write('pid=%d session=%d command=%r\n' % (pid, info[0], command))
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
suffix=$([ "${slot}" = 1 ] && echo "" || echo "-${slot}")
source ~/adamic-tools/env.sh
# Stock tsc is the gates' pinned TypeScript 6.0.3 (the LKG checkout setup provides), for oracles that
# take it from PATH (cmd/adamic-test262's stock-rejection check). No box had a tsc on PATH.
export PATH="${ADAMIC_TYPESCRIPT_SOURCE}/bin:${PATH}"
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
for directory in tools tree; do
  if [ ! -d ~/fast-gate/${directory}${suffix} ]; then
    if [ -d ~/fast-gate/${directory} ]; then cp -a ~/fast-gate/${directory} ~/fast-gate/${directory}${suffix}; else git clone -q https://github.com/system-inc/adamic.git ~/fast-gate/${directory}${suffix}; fi
  fi
done
# Holding the slot's lock means nothing else uses this tree, so any git lock in it is stale: a gate
# killed mid-switch left one (Oct 7 23:02), and every later gate in the slot failed in 12 s on it.
# The deepest is the nested submodule's, .git/modules/cohere/modules/TypeScript/index.lock (depth 5):
# at depth 4 the sweep missed it, and slot 1 voided gates on it from 09:45Z (Oct 8).
find ~/fast-gate/tree${suffix}/.git -maxdepth 6 -name index.lock -print -delete 2>/dev/null | sed 's/^/removed stale /'
git -C ~/fast-gate/tools${suffix} fetch -q origin "${tools}"
git -C ~/fast-gate/tools${suffix} switch -q --detach "${tools}"
git -C ~/fast-gate/tree${suffix} fetch -q origin "${sha}" "${base}"
git -C ~/fast-gate/tree${suffix} switch -q --detach "${sha}"
git -C ~/fast-gate/tree${suffix} submodule update -q --init --recursive
mkdir -p ~/"${out}"
echo "slot=${slot} load_before=$(cut -d' ' -f1-3 /proc/loadavg)" > ~/"${out}"/box.txt
# A slot checkout holding a nested copy of a slot directory (tools-3/tools-2 and tree-3/tree-3.partial on
# Cloud, Oct 8) makes the census and every tree walk read a second tree, so the gate is void, naming box,
# slot and paths, until a person moves it out. The repository has no top-level tools* or tree* of its own;
# a tracked one would never match, since only untracked directories count.
nested=""
for checkout in ~/fast-gate/tools${suffix} ~/fast-gate/tree${suffix}; do
  for candidate in "${checkout}"/tools* "${checkout}"/tree*; do
    [ -d "${candidate}" ] || continue
    if [ -n "$(git -C "${checkout}" ls-files -- "$(basename "${candidate}")" | head -1)" ]; then continue; fi
    nested="${nested} ${candidate}"
  done
done
if [ -n "${nested}" ]; then
  echo "void: ${sha} fast gate, box $(hostname) slot ${slot} holds nested checkout copies:${nested}" > ~/"${out}"/status.txt
  exit 0
fi
# A declared tool (cloud/fast-gate/tools.txt) the box lacks makes the gate void, naming the box: it says
# nothing about the change, so the watcher gates it again elsewhere (@system_adamic, Oct 8 05:49).
if ! lacks=$(bash ~/fast-gate/tools${suffix}/cloud/fast-gate/tools-check.sh); then
  echo "${lacks}" > ~/"${out}"/tools-missing.txt
  echo "void: ${sha} fast gate, box $(hostname) lacks a declared tool: $(echo "${lacks}" | sed -E 's/^lacks ([^:]+):.*/\1/' | tr '\n' ' ')" > ~/"${out}"/status.txt
  exit 0
fi
# The box is partitioned, not time-shared: the area slot owns three eighths of the CPUs (24 of 64; Go
# sizes GOMAXPROCS from the affinity), the two small slots three sixteenths each (12: a one-function
# edit gated in 19.8 to 21.1 s there), and on a 64-CPU box a third small slot on the last quarter (16), all
# at normal priority, so no side's timing can starve another's (a shared box decided verdicts tonight).
cpus=$(nproc --all)
area=$((cpus * 3 / 8))
small=$((cpus * 3 / 16))
case ${slot} in
  1) first=0 share=${area}; [ "${cpus}" -lt 32 ] && share=${cpus} ;;
  2) first=${area} share=${small} ;;
  3) first=$((area + small)) share=${small} ;;
  *) first=$((area + 2 * small)) share=$((cpus - area - 2 * small)) ;;
esac
# The watcher holds the whole box for a reserved gate (Server for the #1 step), so it takes every CPU:
# area-next-12 ran on 0-23 of Server's 64 at load 13 while 40 CPUs and three small slots sat idle (Oct 8).
[ -n "${wholeBox}" ] && first=0 share=${cpus}
# --cpus N narrows the gate to the first N CPUs of its slot (to size slots by measurement).
range="${first}-$((first + ${width:-${share}} - 1))"
echo "cpus=${range}" >> ~/"${out}"/box.txt
# Landings and areas run every test and fixture and name every failure, failing at the end (run.py
# --complete): one pass shows all a candidate's moved outcomes, not one per gate.
complete=""
case ${branch} in cloud/land-*|area/*) complete=--complete ;; esac
taskset -c "${range}" python3 ~/fast-gate/tools${suffix}/cloud/fast-gate/run.py --tree ~/fast-gate/tree${suffix} --sha "${sha}" --base "${base}" --base-name "${baseName}" --tools ~/fast-gate/tools${suffix} --out ~/"${out}" --branch "${branch}" --branch-source "${branchSource}" --session "${session}" --session-source "${sessionSource}" ${complete}
BOX
code=$?
set -e

# A stopped runner exits nonzero with its artifacts intact: publish those just as a finished red.
local=$(mktemp -d)
scp -q -r "${box}:${out}" "${local}/fast"
if [ -n "${darwinLog}" ]; then
  # Under set -e here: capture the leg's code, never let a red or void leg end the gate before it publishes.
  darwinCode=0
  wait "${darwinPid}" || darwinCode=$?
  cp "${darwinLog}" "${local}/fast/darwin-compile.log" 2> /dev/null
  # Red if macOS can't compile it (unless the box already failed first); void if no Mac answered.
  python3 - "${local}/fast" "${darwinCode}" "${sha}" <<'DARWIN'
import json, os, sys
directory, code, sha = sys.argv[1], int(sys.argv[2]), sys.argv[3]
log = open(os.path.join(directory, "darwin-compile.log")).read()
verdict = next((line for line in log.splitlines() if line.startswith(("green:", "red:", "void:"))), "void: no verdict from the darwin leg")
path = os.path.join(directory, "fast.json")
result = json.load(open(path)) if os.path.exists(path) else {}
result.setdefault("stages_exit", {})["darwin-compile"] = code
result["darwin_compile"] = verdict
status = os.path.join(directory, "status.txt")
first = open(status).read().splitlines()[0] if os.path.exists(status) else ""
if code == 1 and first.startswith("green:"):
    result["status"] = "red"
    result["failure"] = {"step": "darwin-compile", "detail": log[-4000:]}
    open(status, "w").write("red: %s fast gate, first failure at darwin-compile (%s)\n" % (sha, verdict[:200]))
elif code == 2 and first.startswith("green:"):
    open(status, "w").write("void: %s fast gate, darwin leg %s\n" % (sha, verdict[len("void: "):]))
json.dump(result, open(path, "w"), indent=2)
DARWIN
  echo "darwin leg: $(head -1 "${local}/fast/darwin-compile.log" | cut -c1-200)"
fi
# A log over 5 MB (test.jsonl on a whole run) is published gzipped; anything else that size (a binary
# that strayed in) never is. Names go to stderr, never stdout, which a caller may be capturing.
find "${local}/fast" -type f -size +5M \( -name '*.jsonl' -o -name '*.log' -o -name '*.txt' \) -exec gzip -9 {} \;
find "${local}/fast" -type f -size +5M -exec mv {} "${local}" \; -print >&2
logBranch=gate-logs/${sha:0:12}/${stamp}/fast
index=$(mktemp -u)
gitDirectory=$(git -C "${here}" rev-parse --absolute-git-dir)
tree=$(cd "${local}/fast" && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" --work-tree=. add -A -f . && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" write-tree)
commit=$(git -C "${here}" commit-tree "${tree}" -m "Fast gate of ${sha} against ${baseName} ${base} (tools ${tools}): $(head -1 "${local}/fast/status.txt")")
git -C "${here}" push -q origin "${commit}:refs/heads/${logBranch}"
rm -f "${index}"
echo "published ${logBranch} (${commit})"
cat "${local}/fast/status.txt"
exit ${code}

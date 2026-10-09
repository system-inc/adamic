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
}
if ! preempt; then
  echo "void: cf241326dfcffa4b4f13b801a345024b604ff5da full gate, box $(hostname) could not stop its idle jobs in 10 s" > ~/"full-gate/out/cf241326dfcf-20261009T000849Z"/status.txt
  echo '{"finished": true, "void": true}' > ~/"full-gate/out/cf241326dfcf-20261009T000849Z"/full.json
  exit 0
fi
source ~/adamic-tools/env.sh
# Stock tsc is the gates' pinned TypeScript 6.0.3 (the LKG checkout setup provides), for oracles that
# take it from PATH (cmd/adamic-test262's stock-rejection check). No box had a tsc on PATH.
export PATH="${ADAMIC_TYPESCRIPT_SOURCE}/bin:${PATH}"
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "3b782dc024c4128914c227df5c96acca6c8e1530" && git -C ~/full-gate/tools switch -q --detach "3b782dc024c4128914c227df5c96acca6c8e1530"
# A declared tool the box lacks (cloud/fast-gate/tools.txt) makes the run void, naming the box, never red.
if ! lacks=$(bash ~/full-gate/tools/cloud/fast-gate/tools-check.sh); then
  echo "${lacks}" > ~/"full-gate/out/cf241326dfcf-20261009T000849Z"/tools-missing.txt
  echo "void: cf241326dfcffa4b4f13b801a345024b604ff5da full gate, box $(hostname) lacks a declared tool: $(echo "${lacks}" | sed -E 's/^lacks ([^:]+):.*/\1/' | tr '\n' ' ')" > ~/"full-gate/out/cf241326dfcf-20261009T000849Z"/status.txt
  echo '{"finished": true, "void": true}' > ~/"full-gate/out/cf241326dfcf-20261009T000849Z"/full.json
  exit 0
fi
git -C ~/full-gate/tree fetch -q origin "cf241326dfcffa4b4f13b801a345024b604ff5da" && git -C ~/full-gate/tree switch -q --detach "cf241326dfcffa4b4f13b801a345024b604ff5da"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
first=$([ "all" = quarter ] && echo $((cpus * 3 / 4)) || echo 0)
taskset -c "$first-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "cf241326dfcffa4b4f13b801a345024b604ff5da" --base "cf241326dfcffa4b4f13b801a345024b604ff5da" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/cf241326dfcf-20261009T000849Z"

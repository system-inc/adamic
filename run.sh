set -u
exec 9> ~/full-gate/lock
flock 9
# A run that can't stop the box's idle jobs is void, naming the box: under set -e it would end with no
# full.json, and the loop waits for that file without a deadline.
preempt() {
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
            if b'ADAMIC_IDLE_JOB=1' in open(path + '/environ', 'rb').read().split(b'\0'):
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
}
if ! preempt; then
  echo "void: d72728e570fe09d91cf55b37b564dbad0fefc24d full gate, box $(hostname) could not stop its idle jobs in 10 s" > ~/"full-gate/out/d72728e570fe-20261008T201805Z"/status.txt
  echo '{"finished": true, "void": true}' > ~/"full-gate/out/d72728e570fe-20261008T201805Z"/full.json
  exit 0
fi
source ~/adamic-tools/env.sh
# Stock tsc is the gates' pinned TypeScript 6.0.3 (the LKG checkout setup provides), for oracles that
# take it from PATH (cmd/adamic-test262's stock-rejection check). No box had a tsc on PATH.
export PATH="${ADAMIC_TYPESCRIPT_SOURCE}/bin:${PATH}"
# env.sh points TMPDIR into /tmp, which a WSL restart empties: every gate failed in 1 s on 'go:
# creating work dir: stat /tmp/adamic-gate' after Cloud's restart (Oct 8). Each gate makes it, world-traversable.
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
git -C ~/full-gate/tools fetch -q origin "0e1247fc366643a9f51584bb74005d363208d065" && git -C ~/full-gate/tools switch -q --detach "0e1247fc366643a9f51584bb74005d363208d065"
# A declared tool the box lacks (cloud/fast-gate/tools.txt) makes the run void, naming the box, never red.
if ! lacks=$(bash ~/full-gate/tools/cloud/fast-gate/tools-check.sh); then
  echo "${lacks}" > ~/"full-gate/out/d72728e570fe-20261008T201805Z"/tools-missing.txt
  echo "void: d72728e570fe09d91cf55b37b564dbad0fefc24d full gate, box $(hostname) lacks a declared tool: $(echo "${lacks}" | sed -E 's/^lacks ([^:]+):.*/\1/' | tr '\n' ' ')" > ~/"full-gate/out/d72728e570fe-20261008T201805Z"/status.txt
  echo '{"finished": true, "void": true}' > ~/"full-gate/out/d72728e570fe-20261008T201805Z"/full.json
  exit 0
fi
git -C ~/full-gate/tree fetch -q origin "d72728e570fe09d91cf55b37b564dbad0fefc24d" && git -C ~/full-gate/tree switch -q --detach "d72728e570fe09d91cf55b37b564dbad0fefc24d"
git -C ~/full-gate/tree submodule update -q --init --recursive
cpus=$(nproc --all)
first=$([ "all" = quarter ] && echo $((cpus * 3 / 4)) || echo 0)
taskset -c "$first-$((cpus - 1))" python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "d72728e570fe09d91cf55b37b564dbad0fefc24d" --base "d72728e570fe09d91cf55b37b564dbad0fefc24d" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/"full-gate/out/d72728e570fe-20261008T201805Z"

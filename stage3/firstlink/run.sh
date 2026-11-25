#!/usr/bin/env bash
set -u -o pipefail

failure=
trap 'status=$?; if ((status)); then printf "FAIL %s\nVERDICT FAIL\n" "${failure:-unexpected runner failure}"; fi' EXIT
stop() { failure=$1; exit 1; }

here=$(cd -- "$(dirname -- "$0")" && pwd -P) || stop "locate the harness"
repository=$(cd -- "$here/../.." && pwd -P) || stop "locate the repository"
cd "$repository" || stop "enter the repository"

# shellcheck source=/dev/null
source /opt/adamic-block/environment.sh || stop "sealed environment"
# shellcheck source=cloud/sandboxed-job.sh
source cloud/sandboxed-job.sh || stop "sandboxed setup"

work=$(mktemp -d "${LOOM_JOB_SCRATCH:-$TMPDIR}/firstlink.XXXXXX") || stop "create job scratch"
work=$(cd -- "$work" && pwd -P) || stop "resolve job scratch"
mkdir -p "$work/home" || stop "create apply home"
echo "WORK $work"

# As in the first-stop census, a Python parent reads peak child RSS from the kernel.
timed() {
	local phase=$1 log=$2
	shift 2
	python3 -I - "$phase" "$log" "$@" <<'PY'
import resource
import subprocess
import sys
import time

phase, path, *command = sys.argv[1:]
start = time.monotonic()
with open(path, "wb") as log:
    try:
        code = subprocess.call(command, stdout=log, stderr=subprocess.STDOUT)
    except OSError as error:
        print(error, file=sys.stderr)
        code = 1
    if code < 0:
        code = 128 - code
rss = resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss
print(f"PHASE {phase} wall {time.monotonic() - start:.3f} s peak {rss} KiB exit {code}")
sys.exit(code)
PY
}

(
	export -f sandboxed_npm sandboxed_typescript_source sandboxedJobFail
	export sandboxedJobScratch
	timed setup "$work/setup.log" bash -c '
		sandboxed_npm stage3/api stage3/typescript && sandboxed_typescript_source &&
			printf "%s\n" "$ADAMIC_TYPESCRIPT_SOURCE" > "$1"
	' bash "$work/typescript-source"
) || { tail -n 60 "$work/setup.log"; stop "sandboxed setup"; }
read -r ADAMIC_TYPESCRIPT_SOURCE < "$work/typescript-source" || stop "read job TypeScript source path"
export ADAMIC_TYPESCRIPT_SOURCE
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 NODE_DISABLE_COMPILE_CACHE=1 PYTHONDONTWRITEBYTECODE=1
export TSC_ADAPT_TYPESCRIPT="$repository/stage3/api/node_modules/typescript"

pin=050880ce59e30b356b686bd3144efe24f875ebc8
commit=$(git rev-parse HEAD 2>/dev/null || echo "${LOOM_JOB_COMMIT:-unknown}")
echo "ADAMIC-SHA $commit"
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pin" ] || stop "TypeScript source is not at the pin"

# Use the first-stop census checkout and apply steps.
checkout_tree() {
	local work=$1 pin=$2
	git clone -q --depth 1 --no-checkout "file://$ADAMIC_TYPESCRIPT_SOURCE" "$work/tree" && git -C "$work/tree" remote remove origin \
		&& git -C "$work/tree" sparse-checkout set --no-cone '/*' '!/tests/baselines/reference/' '/tests/baselines/reference/api/' \
		&& git -C "$work/tree" -c advice.detachedHead=false checkout -q --detach "$pin"
}
(
	export -f checkout_tree
	timed checkout "$work/checkout.log" bash -c 'checkout_tree "$@"' bash "$work" "$pin"
) || { tail -n 60 "$work/checkout.log"; stop "checkout of the pin"; }
timed adapted-tree "$work/apply.log" env HOME="$work/home" XDG_CONFIG_HOME="$work/home" \
	python3 -E -S -B stage3/apply.py --adapt-checkout "$work/tree" \
	|| { tail -n 60 "$work/apply.log"; stop "apply"; }

timed door-build "$work/build.log" go build -o "$work/adamic-firstlink" ./cmd/adamic-firstlink \
	|| { tail -n 60 "$work/build.log"; stop "build the door"; }

code=0
timed door "$work/door.log" timeout --foreground 60m "$work/adamic-firstlink" "$work/tree/src/tsc/tsc.ts" "$work/out" || code=$?
if ((code)); then
	tail -n 60 "$work/door.log"
	if ((code == 124)); then
		stop "door exceeded 60 minutes, no completed report read"
	fi
	stop "door exit $code, no completed report read"
fi
cat "$work/door.log" || stop "read door log"
[ -f "$work/out/traps.json" ] || stop "door wrote no traps.json"

timed report "$work/report.log" python3 -I -B "$here/report.py" "$work/out" \
	|| { tail -n 60 "$work/report.log"; stop "read the door report"; }
grep -Eq '^REPORT-SHA256 [0-9a-f]{64}$' "$work/report.log" || stop "report SHA256 missing"
cat "$work/report.log" || stop "print report SHA256"
echo "VERDICT PASS"

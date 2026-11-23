#!/usr/bin/env bash
# The first-stop census on a cadence (#f2kma1s, measured on #yfemexs): every census unit's first stop at the commit
# the job is laid on, both passes (entry rooted at tsc.ts, and src/compiler) side by side, first-error, CommonJSBundle,
# LATENT_TAIL=1. Every phase is timed. Prints CENSUS-SHA, the PHASE lines, each pass's peak memory, the table as
# firststops.py's FS lines, and a verdict. fsdiff.py beside it reads two such logs into one diff. Ask it sealed at F16:
#   ahra loom job 'source /opt/adamic-block/environment.sh && bash stage3/census/firststops/run.sh' --kind proof \
#       --size F16 --ceiling 60m --at adamic@<main's sha> ...
set -u -o pipefail
here=$(cd -- "$(dirname -- "$0")" && pwd -P)
repository=$(cd -- "$here/../../.." && pwd -P)
cd "$repository"
began=$SECONDS
phase() { echo "PHASE $1 $((SECONDS - began))"; }
stop() { echo "FAIL $1"; echo "VERDICT FAIL at $((SECONDS - began)) s"; exit 1; }

source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh \
	&& sandboxed_npm stage3/api stage3/typescript && sandboxed_typescript_source || stop "sandboxed setup"
phase setup
work="${LOOM_JOB_SCRATCH:-$TMPDIR}/firststops"
mkdir -p "$work/home" "$work/streams" "$work/tools"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 NODE_DISABLE_COMPILE_CACHE=1 PYTHONDONTWRITEBYTECODE=1
export TSC_ADAPT_TYPESCRIPT="$repository/stage3/api/node_modules/typescript"
# TypeScript 6.0.3, the source the adaptations apply to (the pair census holds the same pin).
pin=050880ce59e30b356b686bd3144efe24f875ebc8
commit=$(git rev-parse HEAD 2>/dev/null || echo "${LOOM_JOB_COMMIT:-unknown}")
echo "CENSUS-SHA $commit"
echo "census at adamic $commit; $(uname -n), nproc $(nproc), $(free -g | awk '/^Mem:/ {print $2 " GiB"}'), $(go version)"
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pin" ] || stop "TypeScript source is not at the pin"

# The adapted tree, the same way the pair census builds it.
git clone -q --depth 1 --no-checkout "file://$ADAMIC_TYPESCRIPT_SOURCE" "$work/tree" && git -C "$work/tree" remote remove origin \
	&& git -C "$work/tree" sparse-checkout set --no-cone '/*' '!/tests/baselines/reference/' '/tests/baselines/reference/api/' \
	&& git -C "$work/tree" -c advice.detachedHead=false checkout -q --detach "$pin" || stop "checkout of the pin"
phase checkout
(HOME="$work/home" XDG_CONFIG_HOME="$work/home" python3 -E -S -B stage3/apply.py --adapt-checkout "$work/tree") > "$work/apply.log" 2>&1 \
	|| { tail -n 60 "$work/apply.log"; stop "apply"; }
phase adapted-tree

latent=stage3/census/latent
tools="$work/tools"
python3 "$latent/make_overlay.py" "$repository" "$tools/latent-overlay" > "$tools/latent-overlay.log" 2>&1 \
	&& gofmt -w "$tools/latent-overlay/"*.go \
	&& python3 stage3/meter/entry_overlay.py "$tools/latent-overlay" "$tools/entry-overlay" > "$tools/entry-overlay.log" 2>&1 \
	&& gofmt -w "$tools/entry-overlay/"*.go || { tail -n 40 "$tools/latent-overlay.log" "$tools/entry-overlay.log"; stop "overlays"; }
phase overlays
go build -buildvcs=false -overlay="$tools/entry-overlay/overlay.json" -o "$tools/entry-census" ./$latent/tool > "$tools/entry-build.log" 2>&1 \
	|| { tail -n 60 "$tools/entry-build.log"; stop "entry tool"; }
phase entry-tool
go build -buildvcs=false -overlay="$tools/latent-overlay/overlay.json" -o "$tools/directory-census" ./$latent/tool > "$tools/directory-build.log" 2>&1 \
	|| { tail -n 60 "$tools/directory-build.log"; stop "directory tool"; }
phase directory-tool
echo "SIZE go-cache $(du -sm "$(go env GOCACHE)" | cut -f1) MiB; adapted tree $(du -sm "$work/tree" | cut -f1) MiB"

census() {
	local name=$1 tool=$2 root=$3 start=$SECONDS
	# The pass runs under a Python parent so its peak resident memory is read from the kernel (ru_maxrss, KiB).
	LATENT_TIMES="$work/streams/$name.times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 CENSUS_MODULE_FORMAT=CommonJSBundle \
		python3 -I -c 'import resource, subprocess, sys
code = subprocess.call(sys.argv[1:])
print("PEAK-RSS-KIB", resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss, file=sys.stderr)
sys.exit(code)' "$tool" "$root" "$work/streams/$name.jsonl" > "$work/streams/$name.log" 2>&1
	echo "$? $((SECONDS - start))" > "$work/streams/$name.exit"
	echo "PEAK $name $(grep -o 'PEAK-RSS-KIB [0-9]*' "$work/streams/$name.log" | awk '{printf "%.2f GiB", $2 / 1048576}')"
}
census entry "$tools/entry-census" "$work/tree/src/tsc/tsc.ts" &
census directory "$tools/directory-census" "$work/tree/src/compiler" &
wait
phase passes
failed=
for name in entry directory; do
	read -r code wall < "$work/streams/$name.exit"
	echo "PASS-WALL $name exit $code wall $wall s, $(wc -l < "$work/streams/$name.jsonl") records, $(du -k "$work/streams/$name.jsonl" | cut -f1) KiB"
	[ "$code" = 0 ] || { tail -n 30 "$work/streams/$name.log"; failed=1; }
done
python3 - "$work/streams" << 'TIMES'
import json, sys
for name in ('entry', 'directory'):
    rows = [json.loads(l) for l in open(f'{sys.argv[1]}/{name}.times.jsonl') if l.strip()]
    wall = sum(r.get('wall_seconds') or 0 for r in rows); cpu = sum(r.get('cpu_seconds') or 0 for r in rows)
    print(f'UNIT-TIMES {name}: {len(rows)} units, unit wall {wall:.1f} s, unit cpu {cpu:.1f} s')
TIMES
for name in entry directory; do
	python3 -I -B "$here/firststops.py" "$work/streams/$name" "$work/tree" "$name" > "$work/streams/$name.firststops.tsv" || failed=1
done
phase table
# The two passes stop the same units at the same sites (both #yfemexs runs, and J-avj0r6v); a difference is news.
# Only the FS rows: each table ends with its own UNITS line, and the two passes read different unit counts.
if diff <(grep '^FS' "$work/streams/entry.firststops.tsv" | cut -f3- | sort) <(grep '^FS' "$work/streams/directory.firststops.tsv" | cut -f3- | sort) > "$work/streams/passes.diff"; then
	echo "AGREE the entry and directory passes stop the same units at the same sites"
else
	echo "DISAGREE the entry and directory passes differ in $(grep -c '^[<>]' "$work/streams/passes.diff") lines"
fi
cat "$work/streams/entry.firststops.tsv" "$work/streams/directory.firststops.tsv"
(cd "$work/streams" && sha256sum ./*.jsonl ./*.tsv)
echo "CPU (user and system, this shell then its children):"
times
phase done
[ -z "$failed" ] && echo "VERDICT PASS" || { echo "VERDICT FAIL"; exit 1; }

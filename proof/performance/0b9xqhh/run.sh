#!/usr/bin/env bash
# #0b9xqhh: what rulings.md:72's checked reads cost on tsc's own run. From the adamic root, after environment.sh,
# sandboxed-job.sh and `sandboxed_npm stage3/api`: builds the adapted tree, selects the site list's members, counts
# their accesses on an instrumented tsc compiling src/compiler, checks it against a plain build, and prices the count.
# Usage: bash run.sh [sites.tsv]. Each step prints PASS or FAIL; the last line is the verdict. Never merged.
set -u
W=$(pwd)
K=$(cd "$(dirname "$0")" && pwd)
SITES=${1:-$K/sites.tsv}
O=$(mktemp -d "${TMPDIR:-/tmp}/p72.XXXXXX")
export P72_TYPESCRIPT="$W/stage3/api/node_modules/typescript"
verdict=OK
pass() { echo "PASS $1"; }
fail() { echo "FAIL $1 $2"; verdict=FAIL; }
# The log's tail, when a step fails, is what tells its reader why.
tailOf() { tail -n "${2:-15}" "$1" 2>/dev/null | sed 's/^/    /'; }
seconds() { awk -v a="$1" -v b="$2" 'BEGIN { printf "%.3f", b - a }'; }
finish() { echo "VERDICT $verdict"; exit 0; }

echo "node $(node --version 2>&1)"
echo "box: $(uname -srm), $(nproc 2>/dev/null) cores, $(grep -m1 'model name' /proc/cpuinfo 2>/dev/null | cut -d: -f2 | sed 's/^ //'), $(grep -m1 MemTotal /proc/meminfo 2>/dev/null | tr -s ' ')"
echo "load: $(cut -d' ' -f1-3 /proc/loadavg 2>/dev/null)"
echo "adamic $(git -C "$W" rev-parse --short HEAD 2>/dev/null), kit $K, scratch $O"
if [ -d "$P72_TYPESCRIPT" ] && [ -f "$SITES" ] && [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ]; then
	pass box
else
	fail box "want $P72_TYPESCRIPT (sandboxed_npm stage3/api), $SITES and ADAMIC_TYPESCRIPT_SOURCE"
	finish
fi

# The site list's coordinates are the adapted tree's, so everything reads that tree.
if bash stage3/apply.sh --build-product "$O/adapted" > "$O/apply.log" 2>&1; then
	pass adapted
	echo "  adapted tree at $(git -C "$O/adapted" rev-parse --short HEAD 2>/dev/null), $(grep -c '' "$O/adapted/patch-set.md" 2>/dev/null) lines of patch set"
else
	fail adapted "stage3/apply.sh --build-product exited nonzero"
	tailOf "$O/apply.log"
	finish
fi
T=$O/adapted
if [ -f "$T/src/compiler/diagnosticInformationMap.generated.ts" ]; then pass generated; else fail generated "the adapted tree has no diagnosticInformationMap.generated.ts"; finish; fi

# The adapted tree is a clone of the pinned checkout and keeps its lib/; the gate input's lib/ is the fallback.
LIB=$T/lib
[ -f "$LIB/lib.es5.d.ts" ] || LIB=$ADAMIC_TYPESCRIPT_SOURCE/lib
if [ -f "$LIB/lib.es5.d.ts" ]; then pass "lib $LIB"; else fail lib "neither $T/lib nor $ADAMIC_TYPESCRIPT_SOURCE/lib holds lib.es5.d.ts"; finish; fi

# src/compiler's tsconfig asks for types ["node"], and the tree has no node_modules. stage3/api's install is the
# source lockfile's @types/node (25.3.3), linked in as stage3/performance links its dependencies, so select's program
# and both tsc runs compile src/compiler as upstream does rather than through a missing-types error.
[ -e "$T/node_modules" ] || ln -s "$W/stage3/api/node_modules" "$T/node_modules"
if [ -f "$T/node_modules/@types/node/package.json" ]; then
	pass "types @types/node $(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' "$T/node_modules/@types/node/package.json" | head -1)"
else
	fail types "no @types/node under $T/node_modules"
	finish
fi

cp "$SITES" "$O/sites.tsv"
if node "$K/select.cjs" "$T" "$O/sites.tsv" "$O/out" > "$O/select.log" 2>&1; then
	pass select
	cat "$O/select.log"
	echo "--- sites.csv"
	cat "$O/out/sites.csv"
	echo "---"
else
	fail select "exited nonzero"
	tailOf "$O/select.log" 40
	finish
fi

if node "$K/instrument.cjs" "$T" "$O/out/members.json" "$O/instrumented" > "$O/instrument.log" 2>&1; then
	pass instrument
	cat "$O/instrument.log"
else
	fail instrument "exited nonzero"
	tailOf "$O/instrument.log" 40
	finish
fi

built=1
for kind in plain instrumented; do
	src=$T/src
	[ "$kind" = instrumented ] && src=$O/instrumented/src
	if node "$K/build.cjs" "$src" "$LIB" "$O/built-$kind" > "$O/build-$kind.log" 2>&1; then
		pass "build-$kind"
		tail -n 3 "$O/build-$kind.log"
	else
		fail "build-$kind" "exited nonzero"
		tailOf "$O/build-$kind.log" 40
		built=0
	fi
done
[ $built = 1 ] || finish

version=$(node "$O/built-plain/tsc/tsc.js" --version 2>&1)
case "$version" in
	"Version "*) pass "smoke $version" ;;
	*) fail smoke "the plain build's --version printed: $(echo "$version" | head -5 | tr '\n' ' ')"; finish ;;
esac

# The brief's command. Paths in diagnostics are relative to the working directory, the same for every run.
# Emitting, into a scratch directory per run: the transformers, where most of the checked members are read, run only on
# emit, so a --noEmit run would count them as zero.
run() { node "$1/tsc/tsc.js" -p "$T/src/compiler" --outDir "$2" --tsBuildInfoFile "$2/build.tsbuildinfo"; }
durations=()
codes=()
for n in 1 2 3; do
	t0=$(date +%s.%N)
	run "$O/built-plain" "$O/emit-plain-$n" > "$O/plain.$n.out" 2>&1
	codes+=($?)
	t1=$(date +%s.%N)
	durations+=("$(seconds "$t0" "$t1")")
done
median=$(printf '%s\n' "${durations[@]}" | sort -n | sed -n 2p)
echo "plain runs: ${durations[*]} s, median $median s, exit codes ${codes[*]}, $(grep -c 'error TS' "$O/plain.1.out") errors"
if [ "${codes[0]}" = "${codes[1]}" ] && [ "${codes[1]}" = "${codes[2]}" ] && cmp -s "$O/plain.1.out" "$O/plain.2.out" && cmp -s "$O/plain.1.out" "$O/plain.3.out"; then
	pass plain-runs
else
	fail plain-runs "the three plain runs disagree"
fi
head -n 20 "$O/plain.1.out" | sed 's/^/    /'

t0=$(date +%s.%N)
P72_COUNTS="$O/out/counts.json" run "$O/built-instrumented" "$O/emit-instrumented" > "$O/instrumented.out" 2>&1
code=$?
t1=$(date +%s.%N)
echo "instrumented run: $(seconds "$t0" "$t1") s, exit code $code"
if [ "$code" = "${codes[0]}" ] && cmp -s "$O/plain.1.out" "$O/instrumented.out"; then
	pass diagnostics-match
else
	fail diagnostics-match "the instrumented run's output or exit code differs from the plain run's"
	diff "$O/plain.1.out" "$O/instrumented.out" | head -n 30 | sed 's/^/    /'
fi

# Upstream's own LKG on the same tree, for the reader: adaptations may change what tsc reports, so this isn't a step.
node "$T/lib/tsc.js" -p "$T/src/compiler" --noEmit > "$O/lkg.out" 2>&1
lkg=$?
if cmp -s "$O/plain.1.out" "$O/lkg.out"; then echo "info: the LKG lib/tsc.js prints the same diagnostics (exit $lkg)"; else echo "info: the LKG lib/tsc.js differs from the plain build (exit $lkg, $(grep -c 'error TS' "$O/lkg.out") errors)"; fi

if [ -s "$O/out/counts.json" ] && node "$K/report.cjs" "$O/out/instrumented.json" "$O/out/counts.json" "$median" > "$O/report.log" 2>&1; then
	pass report
	cat "$O/report.log"
else
	fail report "no counts.json from the instrumented run, or report.cjs failed"
	tailOf "$O/report.log" 20
fi
echo "products in $O/out: sites.csv members.json instrumented.json accesses.csv counts.json"
finish

#!/bin/bash
# #pc29n2k for Kirk: the primes demo (#sp4m5ex) with `%` lowered as C's integer remainder, the code a
# proven-integer path would emit, beside main's fmod and Node, on one box. A throwaway: the patch is made
# here, in this job's own tree, and never lands. Its cast is right only where both operands are whole and
# below 2^63, which this program's are (every % is on numbers below 10,000,000); Optimizer's real fix must
# prove that, and keep -0.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
D=measure/sp4m5ex
E=internal/native/emit_expressions.go
echo "box $(hostname), $(nproc) cpus this job may use, $(uname -r), tree $(git rev-parse HEAD 2>/dev/null || echo unknown)"
[ -r /sys/fs/cgroup/cpuset.cpus.effective ] && echo "cpuset $(cat /sys/fs/cgroup/cpuset.cpus.effective)"
grep -m1 'model name' /proc/cpuinfo
echo "threads per core: $(lscpu 2>/dev/null | awk -F: '/Thread\(s\) per core/ {gsub(/ /, "", $2); print $2}')"
echo "node $(node --version)"

go build -o "$T/adamic-fmod" ./cmd/adamic > "$T/build.log" 2>&1 && echo "PASS main's compiler built (fmod)" || { cat "$T/build.log"; echo "FAIL build"; exit 1; }
python3 - "$E" <<'PATCH' || { echo "FAIL patch"; exit 1; }
import sys
path = sys.argv[1]
text = open(path).read()
old = 'return fmt.Sprintf("fmod(%s, %s)", left, right)'
new = 'return fmt.Sprintf("((double)((long long)(%s) %% (long long)(%s)))", left, right)'
assert text.count(old) == 1, "the fmod line isn't there exactly once"
open(path, "w").write(text.replace(old, new))
print("PASS patched " + path + ": % emits C's integer remainder")
PATCH
go build -o "$T/adamic-integer" ./cmd/adamic > "$T/build2.log" 2>&1 && echo "PASS the patched compiler built" || { cat "$T/build2.log"; echo "FAIL build"; exit 1; }

for compiler in fmod integer; do
	for name in parallel serial; do
		"$T/adamic-$compiler" build "$D/$name.a" -o "$T/$name-$compiler" > "$T/$name-$compiler.build" 2>&1 || { cat "$T/$name-$compiler.build"; echo "FAIL $name.a with $compiler"; exit 1; }
	done
done
node --disable-warning=ExperimentalWarning oracle/node.mjs "$D/parallel.a" > "$T/node.out" 2>&1 || { cat "$T/node.out"; echo "FAIL node"; exit 1; }
for binary in parallel-fmod serial-fmod parallel-integer serial-integer; do
	"$T/$binary" > "$T/$binary.out" 2>&1
	cmp -s "$T/node.out" "$T/$binary.out" && echo "PASS $binary prints Node's 1,001 lines byte for byte" || { diff "$T/node.out" "$T/$binary.out" | head; echo "FAIL $binary differs from Node"; exit 1; }
done
tail -1 "$T/node.out" | sed 's/^/    /'

TIMEFORMAT='%3R %3U %3S'
runners=(
	"node, Promise.all|node --disable-warning=ExperimentalWarning $D/promise_all.mts|"
	"node, the parallelMap program|node --disable-warning=ExperimentalWarning oracle/node.mjs $D/parallel.a|"
	"serial, fmod (main)|$T/serial-fmod|"
	"serial, integer remainder|$T/serial-integer|"
	"parallelMap 1 thread, integer remainder|$T/parallel-integer|ADAMIC_THREADS=1"
	"parallelMap all cores, fmod (main)|$T/parallel-fmod|"
	"parallelMap all cores, integer remainder|$T/parallel-integer|"
)
echo "load before $(cut -d' ' -f1-3 /proc/loadavg)"
for round in 1 2 3 4 5; do
	for entry in "${runners[@]}"; do
		IFS='|' read -r label command setting <<< "$entry"
		{ time env $setting $command > /dev/null 2>&1; } 2>> "$T/times.$(echo "$label" | tr -c 'a-z0-9' _)"
	done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo "runner | best wall of 5 | slowest | user | system | cores busy"
for entry in "${runners[@]}"; do
	IFS='|' read -r label command setting <<< "$entry"
	f="$T/times.$(echo "$label" | tr -c 'a-z0-9' _)"
	slowest=$(sort -n "$f" | tail -1 | cut -d' ' -f1)
	sort -n "$f" | head -1 | awk -v label="$label" -v slowest="$slowest" '{ printf "%s | %.3f s | %.3f s | %.3f s | %.3f s | %.1f\n", label, $1, slowest, $2, $3, ($2 + $3) / $1 }'
done

if command -v perf > /dev/null 2>&1; then
	echo "perf_event_paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
	for binary in serial-fmod serial-integer; do
		perf stat -x, -e instructions:u "$T/$binary" > /dev/null 2> "$T/$binary.stat" && echo "$binary instructions:u $(grep instructions "$T/$binary.stat" | cut -d, -f1)" || echo "perf stat unavailable for $binary: $(tail -1 "$T/$binary.stat")"
	done
	perf stat -x, -e instructions:u node --disable-warning=ExperimentalWarning "$D/promise_all.mts" > /dev/null 2> "$T/node.stat" && echo "node instructions:u $(grep instructions "$T/node.stat" | cut -d, -f1)"
	if perf record -q -o "$T/perf.data" "$T/serial-integer" > /dev/null 2>&1; then
		echo "top symbols, serial with the integer remainder:"
		perf report -q -i "$T/perf.data" --stdio --sort symbol 2> /dev/null | head -8 | sed 's/^/    /'
	else
		echo "perf record unavailable here"
	fi
else
	echo "no perf on this box"
fi
echo "PASS remainder"

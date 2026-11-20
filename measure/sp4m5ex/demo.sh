#!/bin/bash
# #sp4m5ex, Kirk's parallel demo: 1,000 prime-counting tasks natively through parallelMap, against the
# same work in one loop, against Node, and Promise.all on Node. Never lands. Run from the tree's root.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
D=measure/sp4m5ex
echo "box $(hostname), $(nproc) cpus this job may use, $(uname -r), tree $(git rev-parse HEAD 2>/dev/null || echo unknown)"
[ -r /sys/fs/cgroup/cpuset.cpus.effective ] && echo "cpuset $(cat /sys/fs/cgroup/cpuset.cpus.effective)"
echo "node $(node --version)"

go build -o "$T/adamic" ./cmd/adamic > "$T/build.log" 2>&1 && echo "PASS adamic built" || { cat "$T/build.log"; echo "FAIL adamic build"; exit 1; }
for name in parallel serial; do
	"$T/adamic" build "$D/$name.a" -o "$T/$name" > "$T/$name.build" 2>&1 && echo "PASS built $name.a natively" || { cat "$T/$name.build"; echo "FAIL $name.a"; exit 1; }
done
echo "Promise.all natively, expected to stop (Adamic doesn't build it yet):"
cp "$D/promise_all.mts" "$T/promise_all.ts"
"$T/adamic" build "$T/promise_all.ts" -o "$T/promise_all" 2>&1 | sed 's/^/    /' | head -5

node --disable-warning=ExperimentalWarning oracle/node.mjs "$D/parallel.a" > "$T/node.out" 2>&1 || { cat "$T/node.out"; echo "FAIL node"; exit 1; }
node --disable-warning=ExperimentalWarning "$D/promise_all.mts" > "$T/promise.out" 2>&1 || { cat "$T/promise.out"; echo "FAIL node promise_all"; exit 1; }
"$T/parallel" > "$T/parallel.out" 2>&1 || { echo "FAIL native parallel exited nonzero"; exit 1; }
"$T/serial" > "$T/serial.out" 2>&1 || { echo "FAIL native serial exited nonzero"; exit 1; }
for name in parallel serial promise; do
	cmp -s "$T/node.out" "$T/$name.out" && echo "PASS $name's output is Node's, byte for byte ($(wc -l < "$T/$name.out") lines)" || { diff "$T/node.out" "$T/$name.out" | head; echo "FAIL $name differs from Node"; exit 1; }
done
echo "first lines:"; head -2 "$T/parallel.out" | sed 's/^/    /'
echo "last lines:"; tail -2 "$T/parallel.out" | sed 's/^/    /'

# Each runner's wall, user and system time, five interleaved rounds; the best wall of the five is reported.
TIMEFORMAT='%3R %3U %3S'
cores=$(nproc)
runners=("native parallelMap, default threads|$T/parallel|" "native serial loop|$T/serial|" "node, parallelMap's sequential shim|node --disable-warning=ExperimentalWarning oracle/node.mjs $D/parallel.a|" "node, Promise.all|node --disable-warning=ExperimentalWarning $D/promise_all.mts|")
for threads in 1 2 4 8 16 32 64; do
	[ "$threads" -le "$cores" ] && runners+=("native parallelMap, $threads threads|$T/parallel|ADAMIC_THREADS=$threads")
done
echo "load before $(cut -d' ' -f1-3 /proc/loadavg)"
for round in 1 2 3 4 5; do
	for entry in "${runners[@]}"; do
		IFS='|' read -r label command setting <<< "$entry"
		{ time env $setting $command > /dev/null 2>&1; } 2>> "$T/times.$(echo "$label" | tr -c 'a-z0-9' _)"
	done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo "runner | best wall | user | system | cores busy, (user + system) / wall"
for entry in "${runners[@]}"; do
	IFS='|' read -r label command setting <<< "$entry"
	sort -n "$T/times.$(echo "$label" | tr -c 'a-z0-9' _)" | head -1 | awk -v label="$label" '{ printf "%s | %.3f s | %.3f s | %.3f s | %.1f\n", label, $1, $2, $3, ($2 + $3) / $1 }'
done
echo "PASS demo"

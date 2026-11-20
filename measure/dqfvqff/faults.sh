#!/bin/bash
# #dqfvqff, second read: where graph's 0.63 s of system time and 105,685 minor faults come from (J-h66sxf2,
# on a loaded Pupa). For graph, copy and trees, native: system time and faults per run, the THP faults the
# kernel allocated and the ones it fell back from (thp_fault_alloc, thp_fault_fallback in /proc/vmstat), and
# perf record of user page faults by the user function that took them, with callers. The programs' 400 ms
# idle wait (setTimeout, for their quiet resident set) is outside every number here except wall. Never lands.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); THP $(cat /sys/kernel/mm/transparent_hugepage/enabled); defrag $(cat /sys/kernel/mm/transparent_hugepage/defrag)"
echo "memory: $(awk '/MemTotal|MemFree|MemAvailable/ {printf "%s %s kB  ", $1, $2}' /proc/meminfo)"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
vm() { awk -v k="$1" '$1==k{print $2}' /proc/vmstat; }
for name in graph copy trees; do
	"$T/adamic" build "bench/$name.ts" -o "$T/$name" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL build $name"; exit 1; }
	for run in 1 2 3; do
		alloc=$(vm thp_fault_alloc); fallback=$(vm thp_fault_fallback)
		/usr/bin/time -f "%e %U %S %R" -o "$T/time" "$T/$name" > /dev/null 2>&1
		echo "$name run $run: wall/user/system/minor $(tail -1 "$T/time"), THP allocated $(( $(vm thp_fault_alloc) - alloc )), fell back $(( $(vm thp_fault_fallback) - fallback ))"
	done
	echo "== $name: user page faults by function, top 12"
	perf record -q -e page-faults:u -c 1 -g -o "$T/$name.data" "$T/$name" > /dev/null 2>&1
	perf report -q -i "$T/$name.data" --no-children --sort symbol --stdio 2>/dev/null | grep -E '^\s+[0-9.]+%' | head -12
	top=$(perf report -q -i "$T/$name.data" --no-children --sort symbol --stdio 2>/dev/null | grep -E '^\s+[0-9.]+%' | head -1 | awk '{print $3}')
	echo "-- callers of $top"
	perf report -q -i "$T/$name.data" --no-children --symbols "$top" -G --stdio 2>/dev/null | grep -v '^$' | head -20
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo PASS faults

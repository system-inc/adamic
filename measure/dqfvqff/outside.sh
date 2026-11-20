#!/bin/bash
# #dqfvqff: bench/copy and graph spend over half their wall outside user mode (J-vpa5442). Beside trees,
# whose huge-page region blocks took its faults from 626,554 to 124,024 (J-6sv2t7d): wall, user, system,
# minor and major faults and THP faults allocated per run, native and Node, 5 interleaved rounds, medians;
# then each native program's system calls counted by strace -c -f. Output checked against Node first.
# Never lands.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); THP enabled $(cat /sys/kernel/mm/transparent_hugepage/enabled), defrag $(cat /sys/kernel/mm/transparent_hugepage/defrag)"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
programs=(copy graph trees)
for name in "${programs[@]}"; do
	"$T/adamic" build "bench/$name.ts" -o "$T/$name" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL build $name"; exit 1; }
	want=$(node "bench/$name.ts" | sha256sum | cut -c1-16)
	got=$("$T/$name" | sha256sum | cut -c1-16)
	[ "$got" = "$want" ] && echo "PASS $name prints Node's output" || { echo "FAIL $name output $got, Node $want"; exit 1; }
done
thp() { awk '$1=="thp_fault_alloc"{print $2}' /proc/vmstat; }
for round in 1 2 3 4 5; do
	for name in "${programs[@]}"; do
		for side in native node; do
			if [ $side = native ]; then command=("$T/$name"); else command=(node "bench/$name.ts"); fi
			before=$(thp)
			/usr/bin/time -f "%e %U %S %R %F" -o "$T/time" "${command[@]}" > /dev/null
			after=$(thp)
			echo "$(tail -1 "$T/time") $((after - before))" >> "$T/$name.$side"
		done
	done
done
echo "== medians of 5: wall, user, system, outside user (wall - user), minor faults, major faults, THP faults allocated"
python3 -I - "$T" "${programs[@]}" <<'PY'
import statistics, sys
root, names = sys.argv[1], sys.argv[2:]
for name in names:
    for side in ("native", "node"):
        rows = [list(map(float, line.split())) for line in open(f"{root}/{name}.{side}")]
        wall, user, system, minor, major, thp = (statistics.median(column) for column in zip(*rows))
        print(f"{name:6} {side:6} wall {wall:.3f} s  user {user:.3f}  system {system:.3f}  outside user {wall - user:.3f}  minor {minor:,.0f}  major {major:,.0f}  THP {thp:,.0f}")
PY
echo "== system calls, native, strace -c -f (one run each, untimed)"
if command -v strace > /dev/null; then
	for name in "${programs[@]}"; do
		echo "-- $name"
		strace -c -f -o "$T/strace.$name" "$T/$name" > /dev/null 2>&1
		head -14 "$T/strace.$name"
	done
else
	echo "no strace on this box"
fi
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo PASS outside

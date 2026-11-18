#!/bin/bash
# #zbjp7m2: where parallel_files' one-thread cost goes. Builds bench/parallel_files.a and serial_files.a (the same work
# with a plain loop, no pool) as release natives, checks both print what Node prints, then:
# 1. perf stat, user-space counters, interleaved, 7 rounds: serial, pool at 1, 4 and 8 threads, and Node;
# 2. perf record at threads=1 and for serial, the top symbols by user cycles, side by side.
# Node runs the program through oracle/node.mjs. Never lands. Needs perf with kernel.perf_event_paranoid at 2 or lower (Golem, Pupa, Lunatic since Nov 17).
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), main $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
echo "perf_event_paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
go build -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
"$T/adamic" build bench/parallel_files.a -o "$T/pool" > "$T/b1.log" 2>&1 || { echo "FAIL pool build"; cat "$T/b1.log"; exit 1; }
"$T/adamic" build measure/zbjp7m2/serial_files.a -o "$T/serial" > "$T/b2.log" 2>&1 || { echo "FAIL serial build"; cat "$T/b2.log"; exit 1; }
# Node runs the .a through oracle/node.mjs, as bench/run.go's node row does.
nodeRun=(node --disable-warning=ExperimentalWarning "$PWD/oracle/node.mjs" "$PWD/bench/parallel_files.a")
want=$("${nodeRun[@]}") || { echo "FAIL node"; exit 1; }
for run in serial pool; do got=$(ADAMIC_THREADS=1 "$T/$run"); [ "$got" = "$want" ] && echo "PASS $run prints Node's output" || { echo "FAIL $run: $got, Node: $want"; exit 1; }; done
events=instructions:u,cycles:u,task-clock,context-switches,cpu-migrations,page-faults
for round in 1 2 3 4 5 6 7; do
	for variant in serial pool/1 pool/4 pool/8 node; do
		case $variant in
			serial) command=("$T/serial"); env=() ;;
			node) command=("${nodeRun[@]}"); env=() ;;
			pool/*) command=("$T/pool"); env=(ADAMIC_THREADS=${variant#pool/}) ;;
		esac
		env "${env[@]}" perf stat -x, -e "$events" -o "$T/stat.$variant.$round" "${command[@]}" > /dev/null 2>&1
		/usr/bin/time -f "%e %U %S" -o "$T/time.$variant.$round" env "${env[@]}" "${command[@]}" > /dev/null
	done
done
echo "== medians of 7 rounds: wall s, user s, sys s, then perf stat"
python3 - "$T" <<'PY'
import glob, os, statistics, sys
root = sys.argv[1]
for variant in ['serial', 'pool/1', 'pool/4', 'pool/8', 'node']:
    times = [list(map(float, open(path).read().split())) for path in glob.glob(f'{root}/time.{variant}.*')]
    stats = {}
    for path in glob.glob(f'{root}/stat.{variant}.*'):
        for line in open(path):
            parts = line.strip().split(',')
            if len(parts) > 3 and parts[0] and parts[0][0].isdigit():
                stats.setdefault(parts[2], []).append(float(parts[0]))
    row = [f'{statistics.median(column):.3f}' for column in zip(*times)]
    counted = ' '.join(f'{name}={statistics.median(values):.4g}' for name, values in sorted(stats.items()))
    print(f'{variant:8} wall/user/sys {" ".join(row)}  {counted}')
PY
echo "== perf record, user cycles, top 25 symbols"
for variant in serial pool1; do
	if [ $variant = serial ]; then command=("$T/serial"); else command=(env ADAMIC_THREADS=1 "$T/pool"); fi
	perf record -q -e cycles:u -F 2999 -o "$T/$variant.data" "${command[@]}" > /dev/null 2>&1
	echo "-- $variant"
	perf report -q -i "$T/$variant.data" --stdio --sort symbol --percent-limit 0.5 2>/dev/null | head -25 | cut -c1-140
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"

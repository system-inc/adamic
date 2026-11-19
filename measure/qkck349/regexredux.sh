#!/bin/bash
# #qkck349: native holdout/regexredux took 30 s on fasta's 5M output where Node took 3 s (J-8wdga9b).
# Output against Node first, then native and Node timed at three input sizes, then perf stat of both and
# perf record of native at 1M, user cycles by function, with the callers of the top two symbols. Never
# lands. Needs perf with kernel.perf_event_paranoid at 2 or lower (Golem, Pupa, Lunatic).
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
echo "perf_event_paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
"$T/adamic" build bench/holdout/fasta.a -o "$T/fasta" > "$T/b1.log" 2>&1 || { echo "FAIL fasta build: $(head -c 300 "$T/b1.log")"; exit 1; }
"$T/adamic" build bench/holdout/regexredux.a -o "$T/regexredux" > "$T/b2.log" 2>&1 || { echo "FAIL regexredux build: $(head -c 300 "$T/b2.log")"; exit 1; }
node=(node --disable-warning=ExperimentalWarning oracle/node.mjs bench/holdout/regexredux.a)
failed=0
for size in 500000 1000000 5000000; do
	"$T/fasta" $size > "$T/in$size.txt"
	/usr/bin/time -f "%e %U %M" -o "$T/node$size.time" timeout 300 "${node[@]}" "$T/in$size.txt" > "$T/node$size.out" 2> "$T/node$size.err"
	nodeExit=$?
	/usr/bin/time -f "%e %U %M" -o "$T/native$size.time" timeout 300 "$T/regexredux" "$T/in$size.txt" > "$T/native$size.out" 2> "$T/native$size.err"
	nativeExit=$?
	echo "size $size node exit $nodeExit wall/user/peak KB $(tail -1 "$T/node$size.time")  native exit $nativeExit wall/user/peak KB $(tail -1 "$T/native$size.time")"
	if [ $nodeExit = 0 ] && [ $nativeExit = 0 ] && cmp -s "$T/node$size.out" "$T/native$size.out"; then
		echo "PASS size $size native prints Node's output ($(wc -l < "$T/native$size.out") lines)"
	else
		echo "FAIL size $size: native stderr $(head -c 200 "$T/native$size.err"); first differing line: $(diff "$T/node$size.out" "$T/native$size.out" | head -3 | tr '\n' '|')"
		failed=1
	fi
done
echo "== perf stat at 1M, user counters"
perf stat -x, -e instructions:u,cycles:u,task-clock,page-faults,branch-misses:u,cache-misses:u "$T/regexredux" "$T/in1000000.txt" > /dev/null 2> "$T/stat.native" && grep -E '^[0-9]' "$T/stat.native" | cut -d, -f1,3 | sed 's/^/native /'
perf stat -x, -e instructions:u,cycles:u,task-clock "${node[@]}" "$T/in1000000.txt" > /dev/null 2> "$T/stat.node" && grep -E '^[0-9]' "$T/stat.node" | cut -d, -f1,3 | sed 's/^/node /'
echo "== perf record, native at 1M, user cycles, top 30 symbols"
perf record -q -e cycles:u -F 2999 -g -o "$T/r.data" "$T/regexredux" "$T/in1000000.txt" > /dev/null 2>&1
perf report -q -i "$T/r.data" --stdio --no-children --sort symbol --percent-limit 0.5 -g none 2>/dev/null | head -30 | cut -c1-160
echo "== callers of the top two symbols"
for symbol in $(perf report -q -i "$T/r.data" --stdio --no-children --sort symbol -g none 2>/dev/null | awk 'NF {print $NF}' | head -2); do
	echo "-- $symbol"
	perf report -q -i "$T/r.data" --stdio --no-children --symbols "$symbol" -g caller,0.5,callee --percent-limit 1 2>/dev/null | head -25 | cut -c1-160
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $failed

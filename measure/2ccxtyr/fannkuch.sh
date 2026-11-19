#!/bin/bash
# #2ccxtyr: native holdout/fannkuchredux didn't finish in 2 minutes at n = 12 where Node takes 27.7 s
# (J-azdag1n). Hang or slow: native and Node at n = 7 to 11, each timed, output checked against Node's
# first; then perf stat and perf record of the native run at n = 10, user cycles by function. n is fixed
# in the port, so each size is a copy with its mainThread(12) changed. Never lands. Needs perf with
# kernel.perf_event_paranoid at 2 or lower (Golem, Pupa, Lunatic).
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
echo "perf_event_paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
grep -q '^mainThread(12);$' bench/holdout/fannkuchredux.a || { echo "FAIL the port no longer calls mainThread(12) on its own line"; exit 1; }
slow=0
for n in 7 8 9 10 11; do
	source="bench/holdout/fannkuch_n$n.a"
	sed "s/^mainThread(12);\$/mainThread($n);/" bench/holdout/fannkuchredux.a > "$source"
	"$T/adamic" build "$source" -o "$T/native$n" > "$T/build$n.log" 2>&1 || { echo "FAIL build n=$n: $(head -c 300 "$T/build$n.log")"; exit 1; }
	/usr/bin/time -f "%e %U" -o "$T/node$n.time" timeout 300 node --disable-warning=ExperimentalWarning oracle/node.mjs "$source" > "$T/node$n.out" 2> "$T/node$n.err"
	nodeExit=$?
	/usr/bin/time -f "%e %U" -o "$T/native$n.time" timeout 300 "$T/native$n" > "$T/native$n.out" 2> "$T/native$n.err"
	nativeExit=$?
	echo "n=$n node exit $nodeExit wall/user $(tail -1 "$T/node$n.time")  native exit $nativeExit wall/user $(tail -1 "$T/native$n.time")"
	if [ $nodeExit = 0 ] && [ $nativeExit = 0 ]; then
		cmp -s "$T/node$n.out" "$T/native$n.out" && echo "PASS n=$n native prints Node's output: $(tr '\n' '|' < "$T/native$n.out")" || { echo "FAIL n=$n native $(tr '\n' '|' < "$T/native$n.out"), Node $(tr '\n' '|' < "$T/node$n.out")"; slow=1; }
	else
		echo "n=$n stderr native: $(head -c 200 "$T/native$n.err") node: $(head -c 200 "$T/node$n.err")"
		[ $nativeExit = 124 ] && echo "n=$n native didn't finish in 300 s" && break
	fi
done
if [ -x "$T/native10" ]; then
	echo "== perf stat, native n=10, user counters"
	perf stat -x, -e instructions:u,cycles:u,task-clock,page-faults,branch-misses:u,cache-misses:u "$T/native10" > /dev/null 2> "$T/stat10" && grep -E '^[0-9]' "$T/stat10" | cut -d, -f1,3
	echo "== perf stat, Node n=10, user counters"
	perf stat -x, -e instructions:u,cycles:u,task-clock "node" --disable-warning=ExperimentalWarning oracle/node.mjs bench/holdout/fannkuch_n10.a > /dev/null 2> "$T/nodestat10" && grep -E '^[0-9]' "$T/nodestat10" | cut -d, -f1,3
	echo "== perf record, native n=10, user cycles, top 30 symbols"
	perf record -q -e cycles:u -F 2999 -g -o "$T/n10.data" "$T/native10" > /dev/null 2>&1
	perf report -q -i "$T/n10.data" --stdio --no-children --sort symbol --percent-limit 0.5 -g none 2>/dev/null | head -30 | cut -c1-160
	echo "== callers of the top two symbols"
	for symbol in $(perf report -q -i "$T/n10.data" --stdio --no-children --sort symbol -g none 2>/dev/null | awk 'NF {print $NF}' | head -2); do
		echo "-- $symbol"
		perf report -q -i "$T/n10.data" --stdio --no-children --symbols "$symbol" -g caller,0.5,callee --percent-limit 1 2>/dev/null | head -25 | cut -c1-160
	done
fi
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $slow

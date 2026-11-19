#!/bin/bash
# #1hp4ww2: native holdout/fasta runs 2.4x Node at n = 25,000,000 (J-qvmnh81). Output against Node first,
# then native and Node timed at three sizes, then perf stat of both and perf record of native at 2.5M,
# user cycles by function, the top symbol's hottest lines and the callers of the top two. Never lands.
# Needs perf with kernel.perf_event_paranoid at 2 or lower (Golem, Pupa, Lunatic).
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
"$T/adamic" build bench/holdout/fasta.a -o "$T/fasta" > "$T/b.log" 2>&1 || { echo "FAIL fasta build: $(head -c 300 "$T/b.log")"; exit 1; }
node=(node --disable-warning=ExperimentalWarning oracle/node.mjs bench/holdout/fasta.a)
failed=0
for size in 2500000 5000000 25000000; do
	/usr/bin/time -f "%e %U %S %M" -o "$T/node$size.time" timeout 300 "${node[@]}" $size > "$T/node$size.out" 2> "$T/node$size.err"
	nodeExit=$?
	/usr/bin/time -f "%e %U %S %M" -o "$T/native$size.time" timeout 300 "$T/fasta" $size > "$T/native$size.out" 2> "$T/native$size.err"
	nativeExit=$?
	echo "size $size node exit $nodeExit wall/user/sys/peak KB $(tail -1 "$T/node$size.time")  native exit $nativeExit wall/user/sys/peak KB $(tail -1 "$T/native$size.time")"
	if [ $nodeExit = 0 ] && [ $nativeExit = 0 ] && cmp -s "$T/node$size.out" "$T/native$size.out"; then
		echo "PASS size $size native prints Node's output ($(wc -c < "$T/native$size.out") bytes, sha256 $(sha256sum < "$T/native$size.out" | cut -c1-16))"
	else
		echo "FAIL size $size: native stderr $(head -c 200 "$T/native$size.err"); cmp: $(cmp "$T/node$size.out" "$T/native$size.out" 2>&1 | head -1)"
		failed=1
	fi
	rm "$T/node$size.out" "$T/native$size.out"
done
echo "== perf stat at 2.5M, user counters"
perf stat -x, -e instructions:u,cycles:u,task-clock,page-faults,branch-misses:u,cache-misses:u "$T/fasta" 2500000 > /dev/null 2> "$T/stat.native" && grep -E '^[0-9]' "$T/stat.native" | cut -d, -f1,3 | sed 's/^/native /'
perf stat -x, -e instructions:u,cycles:u,task-clock "${node[@]}" 2500000 > /dev/null 2> "$T/stat.node" && grep -E '^[0-9]' "$T/stat.node" | cut -d, -f1,3 | sed 's/^/node /'
echo "== system calls, native at 2.5M, the five most frequent"
strace -c -f -o "$T/strace" "$T/fasta" 2500000 > /dev/null 2>&1 && sed -n 1,9p "$T/strace"
echo "== perf record, native at 2.5M, user cycles, top 30 symbols"
perf record -q -e cycles:u -F 2999 -g -o "$T/f.data" "$T/fasta" 2500000 > /dev/null 2>&1
perf report -q -i "$T/f.data" --stdio --no-children --sort symbol --percent-limit 0.5 -g none 2>/dev/null | head -30 | cut -c1-160
top=$(perf report -q -i "$T/f.data" --stdio --no-children --sort symbol -g none 2>/dev/null | awk '$2 == "[.]" {print $3}' | head -1)
echo "== perf annotate of $top, its hottest lines"
perf annotate -q -i "$T/f.data" --stdio -s "$top" 2>/dev/null | awk '$1 ~ /^[0-9.]+$/ && $1 + 0 >= 1.0' | sort -rn | head -20 | cut -c1-160
echo "== callers of the top two symbols"
for symbol in $(perf report -q -i "$T/f.data" --stdio --no-children --sort symbol -g none 2>/dev/null | awk '$2 == "[.]" {print $3}' | head -2); do
	echo "-- $symbol"
	perf report -q -i "$T/f.data" --stdio --no-children --symbols "$symbol" -g caller,0.5,callee --percent-limit 1 2>/dev/null | head -25 | cut -c1-160
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $failed

#!/bin/bash
# #dqfvqff's fix, proven and measured on Golem: copy and graph print Node's standard output byte for byte
# with ADAMIC_BENCH_QUIET unset and set, natively and on Node; the quiet line is on standard error only when
# it's set, on both; ./bench/ whole is green; then, once the one-minute load is under 0.5, the two rows,
# go run ./bench -only copy,graph -rounds 5. Never lands.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
for name in copy graph; do
	"$T/adamic" build "bench/$name.ts" -o "$T/$name" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL build $name"; exit 1; }
	for quiet in unset 1; do
		if [ $quiet = 1 ]; then export ADAMIC_BENCH_QUIET=1; else unset ADAMIC_BENCH_QUIET; fi
		node "bench/$name.ts" > "$T/node.out" 2> "$T/node.err"
		"$T/$name" > "$T/native.out" 2> "$T/native.err"
		cmp -s "$T/node.out" "$T/native.out" && echo "PASS $name, quiet $quiet: native's standard output is Node's ($(head -c 60 "$T/node.out"))" || { echo "FAIL $name, quiet $quiet: outputs differ"; diff "$T/node.out" "$T/native.out" | head -4; exit 1; }
		for side in node native; do
			if [ $quiet = 1 ]; then
				grep -q "^vmrss quiet " "$T/$side.err" && echo "PASS $name $side prints its quiet line when asked" || { cat "$T/$side.err"; echo "FAIL $name $side: no quiet line with the variable set"; exit 1; }
			else
				grep -q "^vmrss quiet " "$T/$side.err" && { cat "$T/$side.err"; echo "FAIL $name $side: a quiet line without the variable"; exit 1; } || echo "PASS $name $side prints no quiet line unasked"
			fi
		done
	done
done
unset ADAMIC_BENCH_QUIET
go test -count=1 ./bench/ > "$T/bench.log" 2>&1 && echo "PASS ./bench/ whole" || { tail -30 "$T/bench.log"; echo "FAIL ./bench/"; exit 1; }
for i in $(seq 1 120); do one=$(cut -d" " -f1 /proc/loadavg); awk -v l="$one" 'BEGIN{exit !(l < 0.5)}' && break; sleep 10; done
one=$(cut -d" " -f1 /proc/loadavg)
echo "load at start $(cut -d' ' -f1-3 /proc/loadavg), after waiting up to 20 min for under 0.5"
awk -v l="$one" 'BEGIN{exit !(l < 1.0)}' || { echo "FAIL the box never went quiet, so nothing was timed"; exit 1; }
echo "processes over 5% CPU: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go run ./bench -only copy,graph -rounds 5 -timeout 120s && echo PASS quiet || { echo FAIL; exit 1; }

#!/bin/bash
# #re2xhwa's M2: the first report over everyday and regression on main, losses included. main against
# itself through -compare: an A/A, so each row's verdict should be no change, and the columns that
# matter are each program's baseline median, its sensitivity at 15 samples, and native against Node
# with its interval, a loss or a win only when that interval excludes 1.00. The record cost harness is
# listed, not timed (suites.json). Never lands.
set -u
MAIN=$(git rev-parse HEAD)
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), main $MAIN"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go run ./bench -compare "$MAIN" "$MAIN" -suite regression -samples 15 > "$T/m2.log" 2> "$T/m2.err"
status=$?
cat "$T/m2.log"
echo "-- its stderr, without the per-sample lines:"
grep -v '^sample ' "$T/m2.err" | tail -30
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $status

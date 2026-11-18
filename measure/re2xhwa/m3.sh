#!/bin/bash
# #re2xhwa's M3, the loop's first real variant: #kha2ea0's landing (main 0b13203d, the ownership plans run
# on every program and only a suspending body is left out) against its parent 813bb775, on the regression
# suite. graph and copy carry the trailing timer #kha2ea0 was filed for, so their verdict should be a win
# whose size matches J-2m7rf3w's pairs (user time about 1.9x and 1.75x Node before, 0.83x and 0.93x
# after); every other row should read no change. A row that moves the other way is a finding. Never lands.
set -u
BASE=813bb77594c19caa8ae16f672d2a9b0e483a7db4
CAND=0b13203d7f507dc368c1d341f6772f9877e41411
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD), base $BASE, candidate $CAND"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go run ./bench -compare "$BASE" "$CAND" -suite regression -samples 15 > "$T/m3.log" 2> "$T/m3.err"
status=$?
cat "$T/m3.log"
echo "-- its stderr, without the per-sample lines:"
grep -v '^sample ' "$T/m3.err" | tail -30
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $status

#!/bin/bash
# #re2xhwa's M3 again, on the fixed -compare (eight runs a sample, every native run after Node, fresh
# pages each sample, a random environment pad each run, identical binaries as controls): #kha2ea0's
# landing 0b13203d against its parent 813bb775 on the regression suite. The first argument is straight
# (813bb775 as baseline) or swapped (0b13203d as baseline). graph and copy should read faster straight
# and slower swapped by the same size; every other row is a control and should claim nothing either way.
# A control that leans toward one named side in both runs is the side; one that leans toward a commit is
# something else. Never lands.
set -u
OLD=813bb77594c19caa8ae16f672d2a9b0e483a7db4
NEW=0b13203d7f507dc368c1d341f6772f9877e41411
case "${1:-}" in
	straight) BASE=$OLD CAND=$NEW ;;
	swapped) BASE=$NEW CAND=$OLD ;;
	*) echo "FAIL usage: m3b.sh straight|swapped"; exit 1 ;;
esac
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD), direction $1, base $BASE, candidate $CAND"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go run ./bench -compare "$BASE" "$CAND" -suite regression -samples 15 > "$T/m3.log" 2> "$T/m3.err"
status=$?
cat "$T/m3.log"
echo "-- its stderr, without the per-sample lines:"
grep -v '^sample ' "$T/m3.err" | tail -30
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $status

#!/bin/bash
# #re2xhwa's three-set report: an A/A on this tree (main plus bench's landing) over one suite, the first
# argument: regression (which includes everyday) or holdout. Both sides are the same commit, so every row
# is a control and the column that matters is native against Node, with its interval; the controls line
# says whether the tool stayed clean on this box. A holdout row whose build stops says where. Never lands.
set -u
case "${1:-}" in
	regression|holdout) ;;
	*) echo "FAIL usage: m4.sh regression|holdout"; exit 1 ;;
esac
HEAD=$(git rev-parse HEAD)
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $HEAD, suite $1"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go run ./bench -compare "$HEAD" "$HEAD" -suite "$1" -samples 15 > "$T/m4.log" 2> "$T/m4.err"
status=$?
cat "$T/m4.log"
echo "-- its stderr, without the per-sample lines:"
grep -v '^sample ' "$T/m4.err" | tail -30
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $status

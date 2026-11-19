#!/bin/bash
# #re2xhwa's three-set report: an A/A on this tree (main plus bench's landing) over one suite, the first
# argument: regression (which includes everyday) or holdout; any further arguments go to bench (-only). Both sides are the same commit, so every row
# is a control and the column that matters is native against Node, with its interval; the controls line
# says whether the tool stayed clean on this box. A holdout row whose build stops says where. Never lands.
set -u
case "${1:-}" in
	regression|holdout) ;;
	*) echo "FAIL usage: m4.sh regression|holdout"; exit 1 ;;
esac
HEAD=$(git rev-parse HEAD)
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $HEAD, suite $1, extra flags: ${*:2}"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
suite=$1
shift
# Every per-run line goes to the job's own log as it happens, stamped, so a run stopped at its ceiling
# still says how far it got and what each run cost (J-6rff648 left nothing past its header).
# The report's own lines (stdout) go into the log as they're printed too, unstamped, so a stop also
# keeps the header and every build and input line (J-dj0awxm kept neither).
go run ./bench -compare "$HEAD" "$HEAD" -suite "$suite" -samples 15 "$@" > >(tee "$T/m4.log") 2> >(while IFS= read -r line; do echo "$(date -u +%H:%M:%S) $line"; done)
status=$?
wait
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $status

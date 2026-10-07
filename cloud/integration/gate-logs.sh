#!/usr/bin/env bash
# Publishes a running gate's go test -json stream to git as it grows, so integration reads a failure's
# name and output with a git fetch the moment its package finishes, not from the worker's notes.
# A gate box starts it in the background before go test and lets it run; it publishes again every
# time the log grows and once more when <log>.done appears (write that file after the run and its
# solo reruns), then exits.
#
# Branch: gate-logs/<first 12 of sha>/<UTC start>/whole, one commit per publish, each on top of the
# last, so every push is a fast-forward of a branch only this run writes. The branch holds:
#   test.jsonl    the go test -json stream so far
#   failures.txt  one block per failed test, build or package: its name, then its last 20 output lines
#   status.txt    running or finished, the sha, packages finished, and pass, fail and skip counts
#   extra/        any further files named after the log (a stage 3 report, rerun logs)
#
# usage: cloud/integration/gate-logs.sh <full sha> <go test -json log> [extra file...] &
set -uo pipefail

if [ "$#" -lt 2 ]; then
	echo "usage: $0 <full sha> <go test -json log> [extra file...]" >&2
	exit 2
fi
sha=$1
log=$2
shift 2
extras=("$@")
directory=$(cd "$(dirname "$0")" && pwd)
ref="refs/heads/gate-logs/${sha:0:12}/$(date -u +%Y%m%dT%H%M%S)/whole"
work=$(mktemp -d)
git init -q "$work"
git -C "$work" remote add origin "$(git remote get-url origin)"
git -C "$work" config user.name "Adamic gate"
git -C "$work" config user.email "gate@adamic.invalid"
# A launcher's ignored signals pass to every child: a gate run under nohup gave its programs an ignored
# hangup, and the signal tests failed for the launcher's reason, not the code's. This script is
# started in the background, and a non-interactive shell makes a background job ignore interrupt and
# quit, so its own set isn't the run's: read the launching shell's, which go test inherits from it.
signals="inherited ignored signals: unknown (no /proc)"
if [ -r "/proc/$PPID/status" ]; then
	mask=$(awk '/^SigIgn:/ { print $2 }' "/proc/$PPID/status")
	names=$(python3 -c 'import signal, sys
mask = int(sys.argv[1], 16)
print(" ".join(signal.Signals(n).name for n in range(1, 65) if mask >> (n - 1) & 1 and n in signal.valid_signals()) or "none")' "$mask" 2>/dev/null || echo "?")
	signals="inherited ignored signals (launching shell, pid $PPID): SigIgn=$mask ($names)"
fi
export GATE_LOGS_SIGNALS=$signals
# A test whose required input is unset skips in 0 s, and main took a stage 1 test that way: name the
# inputs this run started with, so a gate is judged with its input set, and the summary marks every
# skip that names a missing one. The list is what developer tools' setup --gate-inputs provides.
requiredInputs="ADAMIC_TYPESCRIPT_SOURCE ADAMIC_TS_PRETTIER ADAMIC_JSON_PRETTIER ADAMIC_GRAPHQL_PRETTIER ADAMIC_CSS_LIBRARY ADAMIC_CSS_PRINTER_LIBRARY ADAMIC_ESTREE_LIBRARY ADAMIC_YAML_LIBRARY ADAMIC_GRAPHQL_LIBRARY ADAMIC_MEDIA_QUERY_LIBRARY ADAMIC_SELECTOR_LIBRARY ADAMIC_VALUES_LIBRARY"
missingInputs=""
for name in $requiredInputs; do
	[ -n "${!name:-}" ] || missingInputs="$missingInputs $name"
done
export GATE_LOGS_REQUIRED_INPUTS=$requiredInputs
export GATE_LOGS_INPUTS="gate inputs missing at launch:${missingInputs:- none}"
echo "gate-logs: publishing to ${ref#refs/heads/}; $signals"

publish() {
	local state=$1
	[ -f "$log" ] || return 0
	cp "$log" "$work/test.jsonl"
	python3 "$directory/gate-logs-summary.py" "$sha" "$state" "$work/test.jsonl" "$work/failures.txt" "$work/status.txt"
	if [ "${#extras[@]}" -gt 0 ]; then
		mkdir -p "$work/extra"
		for file in "${extras[@]}"; do
			[ -f "$file" ] && cp "$file" "$work/extra/"
		done
	fi
	git -C "$work" add -A
	git -C "$work" commit -q -m "Gate log for ${sha:0:12}: $(head -n 1 "$work/status.txt")" || return 0
	local attempt
	for attempt in 1 2 3; do
		git -C "$work" push -q origin "HEAD:$ref" && return 0
		sleep 10
	done
	echo "gate-logs: push to ${ref#refs/heads/} failed three times; will retry on the next change" >&2
}

lastSize=-1
while [ ! -f "$log.done" ]; do
	size=$(wc -c <"$log" 2>/dev/null || echo 0)
	if [ "$size" != "$lastSize" ]; then
		publish running
		lastSize=$size
	fi
	sleep "${GATE_LOGS_INTERVAL:-60}"
done
publish finished
echo "gate-logs: finished ${ref#refs/heads/}"

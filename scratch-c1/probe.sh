#!/usr/bin/env bash
set -u
# Scratch probe for #snypn8c c1, never landed: copies scratch-c1/probe_test.go into internal/lower,
# runs TestC1Probe once, and prints one "c1probe: <row>: ..." line per row. Run from the repository root.
source /opt/adamic-block/environment.sh
# A child bash doesn't inherit the sealed helper functions, so source them here.
source cloud/sealed-job.sh
sealed_products ./internal/lower || { echo "c1probe: products failed"; exit 1; }

probe=internal/lower/c1_probe_test.go
log=$(mktemp "${TMPDIR:-/tmp}/c1probe.XXXXXX")
cp scratch-c1/probe_test.go "$probe" || exit 1
trap 'rm -f "${probe:?}"' EXIT

gofmt -l scratch-c1/probe_test.go | sed 's/^/c1probe: gofmt wants: /'
go test -count=1 -timeout 30m -run '^TestC1Probe$' -v ./internal/lower/ > "$log" 2>&1
code=$?
echo "c1probe: go test exit=$code (log $log)"

status=0
for row in tag_expr tag_expr_condition tag_optional_expr node_call boolean_left boolean_left_chain number_left element_flags or_dead_right; do
	line=$(grep -o "c1probe: $row: .*" "$log" | head -n 1)
	if [ -n "$line" ]; then
		echo "$line"
	else
		# No verdict: the row's program didn't load (a type error in the probe) or the test panicked.
		echo "c1probe: $row: no verdict"
		grep -E -A3 -- "--- FAIL: TestC1Probe/$row( |$)" "$log" | head -n 8 | sed 's/^/c1probe:   /'
		status=1
	fi
done
grep -E '^(panic|FAIL|ok)' "$log" | sed 's/^/c1probe: /'
exit $status

#!/bin/sh
# run.sh <worktree> <name>: builds adamic from the worktree (mutated or not), then runs the branch's own
# invariance tests and every round-twelve B2 probe, writing each probe's compile verdict to <name>.verdicts
# and the tests' results to <name>.tests. Compare a mutant's verdicts with base.verdicts.
set -u
worktree=$1
name=$2
here=$(cd "$(dirname "$0")" && pwd)
probes=$here/../round12/b2
binary=$(mktemp)
trap 'rm -f "$binary"' EXIT
(cd "$worktree" && go build -o "$binary" ./cmd/adamic) || { echo "build failed" > "$here/$name.tests"; exit 1; }
{
	(cd "$worktree" && go test -count=1 ./internal/lower 2>&1 | grep -E '^(ok|FAIL|---)' )
	(cd "$worktree" && PATH=/opt/node24/bin:$PATH go test -count=1 ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/invariance' -v 2>&1 | grep -E '^(ok|FAIL|---)' )
} > "$here/$name.tests"
: > "$here/$name.verdicts"
for probe in "$probes"/*.a; do
	verdict=$("$binary" c "$probe" 2>&1 >/dev/null | head -n 1 | sed -e "s#$probes/##" | cut -c1-140)
	echo "$(basename "$probe"): ${verdict:-compiles}" >> "$here/$name.verdicts"
done

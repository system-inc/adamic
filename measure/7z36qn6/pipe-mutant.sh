#!/bin/bash
# #7z36qn6 lever (a): does TestClosedStdoutEndsAsOnNode hold SIGPIPE? With SIGPIPE added to stop_signals, and no read
# left in stop_with, the handler replaces the program's ignore, so a write to a pipe with no reader should be killed
# by the handler instead of failing with exit 70. Control first, then the mutant. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
A=internal/native_runtime/c/adamic.c
cp "$A" "$TMPDIR/adamic.c.orig"
go test -count=1 -run '^TestClosedStdoutEndsAsOnNode$' ./internal/oracle > "$TMPDIR/control.log" 2>&1 && echo "PASS control: TestClosedStdoutEndsAsOnNode green at the head" || { tail -30 "$TMPDIR/control.log"; echo "FAIL control"; exit 1; }
python3 -I - "$A" <<'PY' || { echo "FAIL seam"; exit 1; }
import sys
path = sys.argv[1]
text = open(path).read()
seam = "\tSIGTERM, SIGINT, SIGHUP, SIGQUIT, SIGUSR2,"
if text.count(seam) != 1:
    sys.exit("seam not found once")
open(path, "w").write(text.replace(seam, seam + " SIGPIPE,", 1))
PY
go test -count=1 -v -run '^TestClosedStdoutEndsAsOnNode$' ./internal/oracle > "$TMPDIR/mutant.log" 2>&1
cp "$TMPDIR/adamic.c.orig" "$A"
git diff --quiet -- "$A" || { echo "FAIL adamic.c not restored"; exit 1; }
grep -E -- '^\s*--- (PASS|FAIL)' "$TMPDIR/mutant.log"
if grep -qE -- '^--- FAIL: TestClosedStdoutEndsAsOnNode' "$TMPDIR/mutant.log"; then
	grep -m3 -E 'native: exit' "$TMPDIR/mutant.log" | cut -c1-240
	echo "PASS mutant SIGPIPE in stop_signals is red at TestClosedStdoutEndsAsOnNode"
else
	tail -20 "$TMPDIR/mutant.log"; echo "FAIL mutant SIGPIPE escaped TestClosedStdoutEndsAsOnNode"; exit 1
fi
echo VERDICT PASS

#!/bin/bash
# bench/reaper.c's proof, sealed: vet and the bench package whole at the head, TestPeakIsTheProgramsOwn
# verbose, then the mutant (buildReaper builds nothing, so run starts the program itself again), which
# must turn ThroughTheReaper red. Never lands.
set -u
go vet ./bench/ > "$TMPDIR/vet.log" 2>&1 && echo "PASS go vet ./bench/" || { cat "$TMPDIR/vet.log"; echo "FAIL vet"; exit 1; }
go test -count=1 -v -run '^TestPeakIsTheProgramsOwn$' ./bench/ > "$TMPDIR/peak.log" 2>&1; code=$?
grep -E -- '--- |reaper|directly|ok |FAIL' "$TMPDIR/peak.log" | cut -c1-300
[ $code = 0 ] && grep -q -- '--- PASS: TestPeakIsTheProgramsOwn/ThroughTheReaper' "$TMPDIR/peak.log" && echo "PASS the peak test ran and is green" || { tail -30 "$TMPDIR/peak.log"; echo "FAIL the peak test"; exit 1; }
go test -count=1 ./bench/ > "$TMPDIR/bench.log" 2>&1 && echo "PASS ./bench/ whole" || { tail -40 "$TMPDIR/bench.log"; echo "FAIL ./bench/ whole"; exit 1; }
python3 -I - <<'PY' || { echo "FAIL the mutant didn't apply"; exit 1; }
path = "bench/run.go"
old = '\tif runtime.GOOS != "linux" {\n\t\treturn "", nil\n\t}\n\tdirectory, err := benchDirectory()'
new = '\tif runtime.GOOS != "" {\n\t\treturn "", nil\n\t}\n\tdirectory, err := benchDirectory()'
text = open(path).read()
assert text.count(old) == 1, "buildReaper's guard isn't where the head put it"
open(path, "w").write(text.replace(old, new))
PY
echo "mutant applied: buildReaper builds nothing"
if go test -count=1 -v -run '^TestPeakIsTheProgramsOwn$' ./bench/ > "$TMPDIR/mutant.log" 2>&1; then grep -E -- '--- ' "$TMPDIR/mutant.log"; echo "FAIL the mutant survived"; exit 1; fi
grep -E -- '--- FAIL|peak read' "$TMPDIR/mutant.log" | cut -c1-300
grep -q -- '--- FAIL: TestPeakIsTheProgramsOwn/ThroughTheReaper' "$TMPDIR/mutant.log" && echo "PASS the mutant is caught by ThroughTheReaper" || { tail -20 "$TMPDIR/mutant.log"; echo "FAIL red for another reason"; exit 1; }
echo "PASS proof"

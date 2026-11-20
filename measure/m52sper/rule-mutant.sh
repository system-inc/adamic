#!/bin/bash
# gameCall reduced to best against best must turn TestAGameCallInsideTheNoiseIsNoWin red. Never lands.
set -u
go test -count=1 -run '^TestAGameCallInsideTheNoiseIsNoWin$' ./bench/ > "$TMPDIR/clean.log" 2>&1 && echo "PASS the test is green at the head" || { tail -20 "$TMPDIR/clean.log"; echo "FAIL at the head"; exit 1; }
python3 -I - <<'PY' || { echo "FAIL the mutant didn't apply"; exit 1; }
path = "bench/game.go"
old = "\tcase adamic.slowest.elapsed < other.best.elapsed:\n\t\treturn adamicFaster\n\tcase adamic.best.elapsed > other.slowest.elapsed:\n\t\treturn adamicSlower\n"
new = "\tcase adamic.best.elapsed < other.best.elapsed:\n\t\treturn adamicFaster\n\tcase adamic.best.elapsed >= other.best.elapsed:\n\t\treturn adamicSlower\n"
text = open(path).read()
assert text.count(old) == 1, "gameCall isn't where the head put it"
open(path, "w").write(text.replace(old, new))
PY
echo "mutant applied: gameCall is best against best"
if go test -count=1 -run '^TestAGameCallInsideTheNoiseIsNoWin$' ./bench/ > "$TMPDIR/mutant.log" 2>&1; then echo "FAIL the mutant survived"; exit 1; fi
grep -E -- '--- FAIL|is called|headline' "$TMPDIR/mutant.log" | head -6 | cut -c1-220
echo "PASS the mutant is caught by TestAGameCallInsideTheNoiseIsNoWin"

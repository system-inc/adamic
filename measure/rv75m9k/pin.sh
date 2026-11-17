#!/usr/bin/env bash
# #rv75m9k, Lower's pin for road B: TestGetterReadThroughView green, then with the mutant (noteAccessorNames records no
# name) red, then green again with the file restored. Sealed: it sources the image's environment itself.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
gofmt -l internal/oracle/getter_view_test.go | grep . && { echo "FAIL gofmt"; exit 1; }
sealed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 && echo "PASS products settled" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
run() { go test -count=1 -run '^TestGetterReadThroughView$' ./internal/oracle/ > "$TMPDIR/$1.log" 2>&1; local code=$?; grep -E -- '--- (PASS|FAIL)|disagrees|must lower|Node:|ok |FAIL' "$TMPDIR/$1.log" | head -12; return $code; }
run plain && echo "PASS pin" || { echo "FAIL pin"; tail -40 "$TMPDIR/plain.log"; exit 1; }
file=internal/lower/class_accessors.go; cp "$file" "$TMPDIR/class_accessors.go.orig"
python3 - "$file" <<'PY'
import sys
p = sys.argv[1]; s = open(p).read()
old = 'l.accessorNames[node.Name().Text()] = true'
assert s.count(old) == 1
open(p, 'w').write(s.replace(old, '_ = node // mutant: no accessor name is recorded'))
PY
grep -c 'mutant: no accessor name' "$file"
if run mutant; then echo "FAIL mutant survived"; else echo "PASS mutant caught"; fi
cp "$TMPDIR/class_accessors.go.orig" "$file"
run restored && echo "PASS restored" || echo "FAIL restored"
echo "== end"

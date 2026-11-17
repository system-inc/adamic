#!/usr/bin/env bash
# #rv75m9k, Lower's pin for road B: TestGetterReadThroughView must run and pass; with finishAccessors skipping the Property
# read (class_accessors.go, the dispatcher rewrite) it must fail; with noteAccessorNames recording nothing it is expected to
# survive, since accessorNames only feeds refusals; restored, it passes again. Sealed: it sources the image's environment.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
gofmt -l internal/oracle/getter_view_test.go | grep . && { echo "FAIL gofmt"; exit 1; }
sealed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 && echo "PASS products settled" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
run() { go test -count=1 -v -run '^TestGetterReadThroughView$' ./internal/oracle/ > "$TMPDIR/$1.log" 2>&1; grep -E -- '--- (PASS|FAIL)|disagrees|must lower|Node:' "$TMPDIR/$1.log" | head -8; grep -q -- '--- PASS: TestGetterReadThroughView' "$TMPDIR/$1.log"; }
mutate() { # mutate <file> <old> <new>
	cp "$1" "$TMPDIR/$(basename "$1").orig"
	python3 - "$1" "$2" "$3" <<'PY'
import sys
path, old, new = sys.argv[1:]
text = open(path).read()
assert text.count(old) == 1, f'{old!r} occurs {text.count(old)} times'
open(path, 'w').write(text.replace(old, new))
PY
}
restore() { cp "$TMPDIR/$(basename "$1").orig" "$1"; }
run plain && echo "PASS pin" || { echo "FAIL pin"; tail -40 "$TMPDIR/plain.log"; exit 1; }
file=internal/lower/class_accessors.go
mutate "$file" 'if names[expression.Name] && reach.owner(' 'if false && names[expression.Name] && reach.owner('
if run skip; then echo "FAIL mutant finishAccessors-skip survived"; else echo "PASS mutant finishAccessors-skip caught"; tail -15 "$TMPDIR/skip.log" | grep -vE '^\s*$' | cut -c1-200; fi
restore "$file"
mutate "$file" 'l.accessorNames[node.Name().Text()] = true' '_ = node'
if run names; then echo "note mutant accessorNames-empty survives, as expected: accessorNames only feeds refusals"; else echo "note mutant accessorNames-empty caught"; fi
restore "$file"
run restored && echo "PASS restored" || echo "FAIL restored"
echo "== end"

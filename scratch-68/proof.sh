#!/usr/bin/env bash
# Emit's review of family 68 round 5 (#bbssdcy), never to land: the head's record optional-property tests
# and internal/lower whole at pin 25f3ed7f with patch 18047f70, then the one-step receiver test's null
# arm and undefined arm each dropped, which must turn null_receiver.ts red.
set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
W="$TMPDIR/w68"; mkdir -p "$W" && echo "PASS scratch" || { echo "FAIL scratch"; exit 1; }
sandboxed_npm stage3/api stage3/typescript >"$W/npm.log" 2>&1 && echo "PASS npm" || { echo "FAIL npm"; tail -20 "$W/npm.log"; exit 1; }
sandboxed_products ./internal/oracle >"$W/products.log" 2>&1 && echo "PASS products" || { echo "FAIL products"; tail -20 "$W/products.log"; exit 1; }
pins='TestRecordOptionalProperty|TestRecordsOptionsBase|TestRecordsOptions'
run() {
	name=$1; shift
	go test -count=1 "$@" >"$W/$name.log" 2>&1; code=$?
	echo "verdict $name exit=$code"
	grep -E -e '^(ok|FAIL|--- FAIL)' "$W/$name.log" | head -30
	[ $code -ne 0 ] && grep -E -e '_test.go:[0-9]+:' "$W/$name.log" | head -12 | cut -c1-400
	return 0
}
run oracle-pins -v -run "$pins" ./internal/oracle/
grep -E -e '^\s+--- (PASS|FAIL)' "$W/oracle-pins.log" | head -30
run lower-readers -run 'TestOptionalChainReaders' ./internal/lower/
F=internal/lower/records.go
cp "$F" "$W/records.go.head"
mutate() {
	label=$1 old=$2 new=$3
	python3 - "$F" "$old" "$new" <<'PY'
import sys
path, old, new = sys.argv[1], sys.argv[2], sys.argv[3]
text = open(path).read()
assert text.count(old) == 1, (old, text.count(old))
open(path, 'w').write(text.replace(old, new))
PY
	echo "planted $label"
	run "mutant-$label" -run 'TestRecordOptionalProperty' ./internal/oracle/
	cp "$W/records.go.head" "$F"
}
mutate null-arm-dropped 'ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: saved}, Right: ir.IsNull{Value: saved}}' 'ir.IsUndefined{Value: saved}'
mutate undefined-arm-dropped 'ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: saved}, Right: ir.IsNull{Value: saved}}' 'ir.IsNull{Value: saved}'
cmp -s "$F" "$W/records.go.head" && echo "PASS restored records.go"
run lower-whole ./internal/lower/
echo DONE

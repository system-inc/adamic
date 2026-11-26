set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh && echo "PASS sandboxed-job" || echo "FAIL sandboxed-job"
mkdir -p "$TMPDIR/scratch-0q0" && echo "PASS scratch" || { echo "FAIL scratch"; exit 1; }
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #0q0cv0m on main 25f3ed7f: a ?. read that reaches dynamicProperty tests its receiver for null and undefined first
# (optionalDynamicProperty). PART=a: gofmt, vet, the oracle pins verbose, mutants M1 to M3 (each one exact line
# replaced by python, checked replaced once, the oracle pins run, put back from the head copy and checked put back).
# PART=d: lower whole. PART=b: oracle whole and native whole. PART=c: stage3/fixtures, ir, compilercensus and
# parallelcensus whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -E "_test.go:[0-9]+:" "$TMPDIR/whole.log" | head -20; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

pins='TestOptionalDynamicReceiver'
if [ "$part" = a ]; then
	files="internal/lower/nullable_objects.go internal/lower/object.go"
	for f in $files; do cp "$f" "$TMPDIR/scratch-0q0/$(basename $f).head"; done
	trap 'for f in $files; do cp "$TMPDIR/scratch-0q0/$(basename $f).head" "$f"; done' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/nullable_objects.go internal/lower/object.go internal/oracle/optional_dynamic_receiver_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/oracle
	step "oracle pins" go test -count=1 -run "$pins" ./internal/oracle/ -v
	subtests
	# Every pin's error words, so a stop's wording is read, not assumed.
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | head -20

	# mutant name file old new: replace exactly one line, run the oracle pins, report each failing subtest and its
	# words, put back.
	mutant() {
		name=$1 file=$2 old=$3 new=$4
		python3 -I - "$file" "$old" "$new" <<'PY'
import sys
path, old, new = sys.argv[1:4]
text = open(path).read()
if text.count(old) != 1:
    print(f"FAIL mutant: {old!r} occurs {text.count(old)} times"); sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
		if [ $? -ne 0 ]; then ok=0; return; fi
		go test -count=1 -run "$pins" ./internal/oracle/ -v > "$TMPDIR/mutant.log" 2>&1
		code=$?
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
		if [ $code -eq 0 ]; then echo "SURVIVED $name"; else echo "RED $name: $reds"; fi
		grep -E "^# |cannot|undefined:|_test.go:[0-9]+:" "$TMPDIR/mutant.log" | grep -v "gate cache\|go_build" | cut -c1-500 | head -14
		cp "$TMPDIR/scratch-0q0/$(basename $file).head" "$file"
		if cmp -s "$TMPDIR/scratch-0q0/$(basename $file).head" "$file"; then echo "PASS restored $file after $name"; else echo "FAIL restore $file"; ok=0; fi
	}
	mutant M1 internal/lower/nullable_objects.go \
		'	missing := ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}' \
		'	missing := ir.IsUndefined{Value: held}'
	mutant M2 internal/lower/nullable_objects.go \
		'	missing := ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}' \
		'	missing := ir.IsNull{Value: held}'
	mutant M3 internal/lower/object.go \
		'		keyRead = l.optionalDynamicProperty' \
		'		keyRead = l.dynamicProperty'
fi
if [ "$part" = d ]; then
	# The census products read the TypeScript source, which must be the job's own copy (sandboxed_typescript_source).
	step "typescript source" sandboxed_typescript_source
	whole "lower whole" go test -count=1 -timeout 40m ./internal/lower/
fi
if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle
	whole "oracle whole" go test -count=1 -timeout 40m ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi
if [ "$part" = c ]; then
	whole "stage3/fixtures whole" go test -count=1 -timeout 40m ./stage3/fixtures/
	whole "ir whole" go test -count=1 -timeout 40m ./internal/ir/
	whole "compilercensus whole" go test -count=1 -timeout 40m ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 -timeout 40m ./cmd/adamic-gate/parallelcensus/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }

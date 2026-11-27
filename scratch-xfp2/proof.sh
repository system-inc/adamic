set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh && echo "PASS sandboxed-job" || { echo "FAIL sandboxed-job"; exit 1; }
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #xfpsqtn (2) on (1) 935983d8: an enum never read of a tuple's element has the element's stored type
# (enumStoredType, tupleElementStoredType). PART=a: gofmt, vet, the oracle pins verbose, mutants M1 to M3 (each one
# exact line replaced by python, checked replaced once, the oracle pins run, put back from the head copy and checked
# put back). PART=d: lower whole. PART=b: oracle whole and native whole. PART=c: stage3/fixtures, ir, compilercensus
# and parallelcensus whole.
part=${PART:-a}
ok=1
mkdir -p "$TMPDIR/xfp2" && echo "PASS mkdir $TMPDIR/xfp2" || { echo "FAIL mkdir $TMPDIR/xfp2"; exit 1; }
scratch="$TMPDIR/xfp2"
step() { name=$1; shift; "$@" > "$scratch/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$scratch/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$scratch/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$scratch/whole.log" | sort -u | head -40; grep -E "_test.go:[0-9]+:" "$scratch/whole.log" | head -20 | cut -c1-600; grep -qE "^(--- FAIL|panic:)" "$scratch/whole.log" || tail -25 "$scratch/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$scratch/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$scratch/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

pins='TestNeverEnumTuple'
if [ "$part" = a ]; then
	files="internal/lower/enum_never.go"
	for f in $files; do cp "$f" "$scratch/$(basename $f).head" && echo "PASS copy $f" || { echo "FAIL copy $f"; ok=0; }; done
	trap 'for f in $files; do cp "$scratch/$(basename $f).head" "$f"; done' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/enum_never.go internal/oracle/never_enum_tuple_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/oracle
	step "oracle pins" go test -count=1 -run "$pins|TestNeverEnumArgument" ./internal/oracle/ -v
	subtests
	# Every pin's error words, so a stop's wording is read, not assumed.
	grep -E "_test.go:[0-9]+:" "$scratch/step.log" | head -20 | cut -c1-600

	# mutant name file old new: replace exactly one line, run the oracle pins, report each failing subtest and its
	# words, put back.
	mutant() {
		name=$1 file=$2 old=$3 new=$4
		python3 -I - "$file" "$old" "$new" <<'PY'
import sys
path, old, new = sys.argv[1:4]
text = open(path).read()
if text.count(old + "\n") != 1:
    print(f"FAIL mutant: {old!r} occurs {text.count(old + chr(10))} times"); sys.exit(1)
open(path, 'w').write(text.replace(old + "\n", new + "\n"))
print("PASS mutant line replaced once")
PY
		if [ $? -ne 0 ]; then ok=0; return; fi
		go test -count=1 -run "$pins" ./internal/oracle/ -v > "$scratch/mutant.log" 2>&1
		code=$?
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$scratch/mutant.log" | tr '\n' ' ')
		if [ $code -eq 0 ]; then echo "SURVIVED $name"; else echo "RED $name: $reds"; fi
		grep -E "^# |cannot|undefined:|_test.go:[0-9]+:" "$scratch/mutant.log" | head -14 | cut -c1-500
		cp "$scratch/$(basename $file).head" "$file"
		if cmp -s "$scratch/$(basename $file).head" "$file"; then echo "PASS restored $file after $name"; else echo "FAIL restore $file"; ok=0; fi
	}
	mutant M1 internal/lower/enum_never.go \
		'			return l.tupleElementStoredType(stored, node.AsElementAccessExpression().ArgumentExpression)' \
		'			return nil'
	mutant M2 internal/lower/enum_never.go \
		'			return elements[index]' \
		'			return elements[0]'
	mutant M3 internal/lower/enum_never.go \
		'		return elements[fixed]' \
		'		return nil'
	git diff --stat -- internal/lower | tail -1
	echo "diff after mutants: $(git diff --name-only -- internal/lower | wc -l) files"
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

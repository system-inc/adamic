set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #8nj5pn4, Emit's head on main 0e7517ac: the uniform field read takes its slot only behind a check of the object's
# shape against the literal layouts holding the name (uniformFieldGuard), and the by-name road's miss is a NotYet.
# PART=a: gofmt, vet, the narrowed_arm_read pins and the fields tests verbose, the tests that pinned the old words,
# mutant M1 (one exact line, put back from the head copy and checked), and the C identity of every oracle program at
# main's emitter files against the head's. PART=b: oracle whole, native whole with ADAMIC_TEST_WASI=1. PART=c:
# internal/native_runtime, compilercensus and parallelcensus whole. PART=m: the price, bench's object-heavy programs
# built by main's compiler and the head's and timed interleaved.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60 | cut -c1-2500; ok=0; fi; }
whole() { name=$1; shift; begun=$SECONDS; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name in $((SECONDS - begun)) s"; else echo "FAIL $name (exit $code) in $((SECONDS - begun)) s"; grep -E "^(--- FAIL|    --- FAIL|        --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -E "_test.go:[0-9]+:" "$TMPDIR/whole.log" | head -30 | cut -c1-1500; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
emitter_files="internal/native/emit_objects.go internal/native/fields.go internal/native/emit.go"
runtime_files="internal/native_runtime/c/object.c internal/native_runtime/c/adamic.h"
keep() { for file in $emitter_files $runtime_files; do cp "$file" "$TMPDIR/$(basename "$file").head"; done; }
restore() {
	for file in $emitter_files $runtime_files; do
		cp "$TMPDIR/$(basename "$file").head" "$file"
		if cmp -s "$TMPDIR/$(basename "$file").head" "$file"; then echo "PASS restored $file after $1"; else echo "FAIL restore $file after $1"; ok=0; fi
	done
}
main_files() { for file in $*; do cp "scratch-guard/main/$(basename "$file").main" "$file" || { echo "FAIL main $file"; ok=0; }; done; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part, $(nproc) cpus"
step npm sandboxed_npm stage3/api stage3/typescript

pins='^(TestNarrowedArmReadStopsOrAgrees)$'
if [ "$part" = a ]; then
	keep
	step gofmt bash -c 'difference=$(gofmt -d internal/native/emit_objects.go internal/native/fields.go internal/native/emit.go internal/native/fields_test.go internal/native/fswatch_test.go internal/lower/class.go internal/oracle/narrowed_arm_read_test.go internal/oracle/literal_optional_test.go internal/oracle/method_values_mutant_test.go internal/oracle/predicates_checked_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/native/ ./internal/oracle/
	step "products oracle" sandboxed_products ./internal/oracle
	step "narrowed arm pins" go test -count=1 -run "$pins" ./internal/oracle/ -v
	subtests
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | grep -v "gate cache" | head -20 | cut -c1-1500
	step "fields tests" go test -count=1 -run '^(TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked)$' ./internal/native/ -v
	subtests
	grep -E "runtime layouts included" "$TMPDIR/step.log" | head -2
	step "the old words' tests" go test -count=1 -run '^(TestCheckedPredicateOracle|TestLiteralOptionalOracleCatchesMutant|TestLibraryMethodOwnLoadMutants)$' ./internal/oracle/ -v
	subtests
	# The head's C for the tag row's read of a, and the JavaScript backend's.
	go build -o "$TMPDIR/adamic" ./cmd/adamic && mkdir -p "$TMPDIR/shown" && echo "PASS mkdir shown"
	cp internal/oracle/testdata/narrowed_arm_read/string_tags.ts "$TMPDIR/shown/string_tags.ts"
	"$TMPDIR/adamic" c "$TMPDIR/shown/string_tags.ts" > "$TMPDIR/shown/string_tags.c" 2>&1
	echo "string_tags C, its reads of a:"; grep -n '"a"' "$TMPDIR/shown/string_tags.c" | cut -c1-500 | head -6
	"$TMPDIR/adamic" js "$TMPDIR/shown/string_tags.ts" > "$TMPDIR/shown/string_tags.mjs" 2>&1
	echo "string_tags JavaScript, its reads of a:"; grep -n '\["a"\]\|\.a\b' "$TMPDIR/shown/string_tags.mjs" | cut -c1-300 | head -6

	# M1: the guard dropped, the uniform read bare again.
	python3 -I - internal/native/emit_objects.go <<'PY'
import sys
path = sys.argv[1]
old = '\t\treturn e.uniformFieldGuard(object, name, slot, fmt.Sprintf("adamic_object_data_field(%s, %s, &%s)", object, cString(name), e.cache()))\n'
new = '\t\treturn slot\n'
text = open(path).read()
if text.count(old) != 1:
    print(f"FAIL mutant M1: the line occurs {text.count(old)} times"); sys.exit(1)
open(path, 'w').write(text.replace(old, new))
print("mutant M1 planted")
PY
	if [ $? -eq 0 ]; then
		go test -count=1 -run "$pins" ./internal/oracle/ -v > "$TMPDIR/mutant.log" 2>&1
		code=$?
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
		greens=$(sed -nE 's/^ *--- PASS: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
		if [ $code -eq 0 ]; then echo "FAIL mutant M1 SURVIVED"; ok=0; else echo "PASS mutant M1 RED: $reds"; fi
		echo "mutant M1 green: $greens"
		grep -E "narrowed_arm_read_test.go:[0-9]+:" "$TMPDIR/mutant.log" | head -12 | cut -c1-2500
		grep -m3 -E "ERROR: AddressSanitizer|READ of size|is located" "$TMPDIR/mutant.log" | cut -c1-300
	else
		ok=0
	fi
	restore "mutant M1"

	# The C identity: every oracle program lowered from its one path in this tree, at main's emitter files and at the head's.
	find internal/oracle/testdata \( -name '*.a' -o -name '*.ts' \) -type f | LC_ALL=C sort > "$TMPDIR/identity.list"
	echo "identity probe: $(wc -l < "$TMPDIR/identity.list") programs"
	cp scratch-guard/identity_probe_test.go.txt internal/oracle/zz_guard_identity_probe_test.go
	for side in head main; do
		[ "$side" = main ] && main_files $emitter_files
		mkdir -p "$TMPDIR/c-$side" && echo "PASS mkdir c-$side"
		GUARD_C_DIR="$TMPDIR/c-$side" GUARD_IDENTITY_LIST="$TMPDIR/identity.list" go test -count=1 -timeout 60m -run '^TestZZGuardIdentityProbe$' ./internal/oracle/ -v > "$TMPDIR/identity-$side.log" 2>&1
		echo "identity probe at $side: exit $?, $(grep -c '^PROBE' "$TMPDIR/identity-$side.log") lines, $(grep -c '^PROBE .* c ' "$TMPDIR/identity-$side.log") lowered"
		grep '^PROBE' "$TMPDIR/identity-$side.log" > "$TMPDIR/identity-$side.lines"
		grep -q '^PROBE' "$TMPDIR/identity-$side.log" || { tail -30 "$TMPDIR/identity-$side.log"; ok=0; }
		restore "the identity probe at $side"
	done
	rm internal/oracle/zz_guard_identity_probe_test.go
	python3 -I scratch-guard/identity.py "$TMPDIR/identity-main.lines" "$TMPDIR/identity-head.lines" "$TMPDIR/c-main" "$TMPDIR/c-head" || ok=0
	git status --porcelain
fi
if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle ./internal/native
	whole "oracle whole" go test -count=1 -timeout 80m ./internal/oracle/
	ADAMIC_TEST_WASI=1 whole "native whole, ADAMIC_TEST_WASI=1" go test -count=1 -timeout 80m ./internal/native/
fi
if [ "$part" = c ]; then
	step "typescript source" sandboxed_typescript_source
	whole "native_runtime whole" go test -count=1 -timeout 40m ./internal/native_runtime/
	whole "compilercensus whole" go test -count=1 -timeout 40m ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 -timeout 40m ./cmd/adamic-gate/parallelcensus/
fi
if [ "$part" = m ]; then
	keep
	mkdir -p "$TMPDIR/price" && echo "PASS mkdir price"
	step "build head adamic" go build -o "$TMPDIR/price/adamic-head" ./cmd/adamic
	main_files $emitter_files $runtime_files
	step "build base adamic" go build -o "$TMPDIR/price/adamic-base" ./cmd/adamic
	restore "the base build"
	programs="nbody.ts trees.ts devirtualize.a graph.ts copy.ts"
	# step() sets name, so the program's own is held in program_name.
	program_names=""
	for program in $programs; do
		program_name=${program%.*}
		program_names="$program_names $program_name"
		for side in base head; do
			step "build $program_name at $side" "$TMPDIR/price/adamic-$side" build "bench/$program" -o "$TMPDIR/price/$program_name.$side"
			"$TMPDIR/price/adamic-$side" c "bench/$program" > "$TMPDIR/price/$program_name.$side.c" 2>&1
		done
		echo "$program_name: lines with an unguarded uniform read at base $(grep -cE '(^|[^?] )\(&ADAMIC_UNIFORM_OBJECT\(' "$TMPDIR/price/$program_name.base.c"), guarded slots at head $(grep -o '? (&ADAMIC_UNIFORM_OBJECT(' "$TMPDIR/price/$program_name.head.c" | wc -l), at base $(grep -o '? (&ADAMIC_UNIFORM_OBJECT(' "$TMPDIR/price/$program_name.base.c" | wc -l); binaries $(stat -c %s "$TMPDIR/price/$program_name.base") and $(stat -c %s "$TMPDIR/price/$program_name.head") bytes"
	done
	echo "load before: $(cat /proc/loadavg)"
	python3 -I scratch-guard/price.py 15 "$TMPDIR/price" $program_names || ok=0
	echo "load after: $(cat /proc/loadavg)"
fi
if [ "$part" = g ]; then
	# no_relation_presence.a's C at main's emitter files and at the head's, from internal/native by the path
	# TestRegExpUnservedEmissionMatchesMain loads it by; main's must be the golden, the head's differ by the guard
	# alone (identity.py's normalization), and the head's is kept as the new golden.
	keep
	mkdir -p "$TMPDIR/golden/c-main" "$TMPDIR/golden/c-head" scratch-guard/out && echo "PASS mkdir golden"
	cp scratch-guard/golden_probe_test.go.txt internal/native/zz_guard_golden_probe_test.go
	for side in head main; do
		[ "$side" = main ] && main_files $emitter_files
		step "golden probe at $side" env GUARD_GOLDEN_OUT="$TMPDIR/golden/c-$side/0.c" go test -count=1 -run '^TestZZGuardGoldenProbe$' ./internal/native/
		restore "the golden probe at $side"
	done
	rm internal/native/zz_guard_golden_probe_test.go
	if cmp -s "$TMPDIR/golden/c-main/0.c" internal/native/testdata/regexp-structural/no-relation.golden; then echo "PASS main's C is the golden, $(wc -c < "$TMPDIR/golden/c-main/0.c") bytes"; else echo "FAIL main's C isn't the golden"; ok=0; fi
	echo "head's C: $(wc -c < "$TMPDIR/golden/c-head/0.c") bytes"
	echo "PROBE 0 no_relation_presence.a c main" > "$TMPDIR/golden/main.lines"
	echo "PROBE 0 no_relation_presence.a c head" > "$TMPDIR/golden/head.lines"
	python3 -I scratch-guard/identity.py "$TMPDIR/golden/main.lines" "$TMPDIR/golden/head.lines" "$TMPDIR/golden/c-main" "$TMPDIR/golden/c-head" || ok=0
	echo "-- diff -u main head, whole:"
	diff -u "$TMPDIR/golden/c-main/0.c" "$TMPDIR/golden/c-head/0.c" | cut -c1-600
	cp "$TMPDIR/golden/c-head/0.c" scratch-guard/out/no-relation.golden && echo "PASS kept the head's C at scratch-guard/out/no-relation.golden, sha256 $(sha256sum scratch-guard/out/no-relation.golden | cut -c1-64)"
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }

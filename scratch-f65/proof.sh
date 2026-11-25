set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #2hzryxh, family 65's typed half on road (b): copied-slot presence bits for a spread, and `in` with a fixed key on
# a proven typed receiver answering from them.
# PART=0: step 0, the six programs through cloud/probe/run.sh, each stop's whole words printed. PART=i: every program.
# PART=a: gofmt, vet, the pins verbose, every wrong version run as a mutant and reversed. PART=b: internal/lower whole.
# PART=c: internal/oracle whole, internal/native whole under WASI. PART=d: internal/ir, stage3/census/latent/...,
# internal/compilercensus, cmd/adamic-gate/parallelcensus, each whole. PART=e: step 0's six, the pins' programs and
# item 5's through cloud/probe/run.sh at head and with main's four compiler files in place.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -40; ok=0; fi; }
whole() { name=$1; shift; begun=$SECONDS; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name in $((SECONDS - begun)) s"; else echo "FAIL $name (exit $code) in $((SECONDS - begun)) s"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log"; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
head_files="internal/lower/unknown.go internal/native/construction.go internal/native/emit_objects.go internal/native/reuse.go"
keep() { for file in $head_files; do cp "$file" "$TMPDIR/$(basename "$file").head"; done; }
restore() {
	for file in $head_files; do
		cp "$TMPDIR/$(basename "$file").head" "$file"
		if cmp -s "$TMPDIR/$(basename "$file").head" "$file"; then echo "PASS restored $file after $1"; else echo "FAIL restore $file after $1"; ok=0; fi
	done
}
main_files() { for file in $head_files; do cp "scratch-f65/main/$(basename "$file").main" "$file"; done; }
pins='^Test(TypedPresence|SpreadOverwrite|UnknownReflectionRefusals)'

# step0 lays the six step-0 programs in a directory: four are main's spread_overwrite fixtures under round 2's
# names, copied here from the tree, never from a copy.
step0() {
	cp internal/oracle/testdata/spread_overwrite/number_source.ts "$1/undefined-number-source.ts"
	cp internal/oracle/testdata/spread_overwrite/optional_string.ts "$1/empty-object.ts"
	cp internal/oracle/testdata/spread_overwrite/string_source.ts "$1/undefined-source.ts"
	cp internal/oracle/testdata/spread_overwrite/optional_number.ts "$1/65-leading-overwrite-empty-object.ts"
	cp scratch-f65/programs/copied-empty-object.ts scratch-f65/programs/65-leading-overwrite.ts "$1/"
}

# table runs a directory through run.sh and prints the table, then each program's Node output and every compile
# stop's whole words.
table() {
	side=$1 programs=$2
	bash cloud/probe/run.sh "$programs" --logs "$TMPDIR/logs-$side" > "$TMPDIR/run-$side.log" 2>&1
	echo "run.sh at $side: exit $?"
	grep -E "^\| |^- head|^Probes" "$TMPDIR/run-$side.log"
	grep -qE "^\| [0-9a-z]" "$TMPDIR/run-$side.log" || { tail -30 "$TMPDIR/run-$side.log"; ok=0; }
	for directory in "$TMPDIR/logs-$side"/*.ts; do
		echo "$side $(basename "$directory"): Node $(tr '\n' ' ' < "$directory/node.stdout")"
		for stage in js-compile native-compile; do
			[ -s "$directory/head.$stage.stderr" ] && echo "$side $(basename "$directory") $stage: $(tr '\n' ' ' < "$directory/head.$stage.stderr" | cut -c1-900)"
		done
		for backend in js native; do
			[ -f "$directory/head.$backend.stdout" ] && echo "$side $(basename "$directory") $backend: $(tr '\n' ' ' < "$directory/head.$backend.stdout")"
		done
	done
}

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = 0 ]; then
	mkdir -p "$TMPDIR/step0"
	step0 "$TMPDIR/step0"
	sha256sum "$TMPDIR/step0"/*.ts | sed "s|$TMPDIR/step0/||"
	table main "$TMPDIR/step0"
fi
if [ "$part" = i ]; then
	mkdir -p "$TMPDIR/all/lib"
	cp scratch-f65/programs/*.ts "$TMPDIR/all/"
	cp scratch-f65/programs/lib/*.ts "$TMPDIR/all/lib/"
	table main "$TMPDIR/all"
fi
if [ "$part" = a ]; then
	step "typescript source" sandboxed_typescript_source
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/unknown.go internal/lower/unknown_test.go internal/native/construction.go internal/native/emit_objects.go internal/native/reuse.go internal/oracle/typed_presence_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/native/ ./internal/oracle/
	step "the pins" go test -count=1 -timeout 40m -run "$pins" ./internal/oracle/ ./internal/lower/ -v
	subtests
	keep
	for wrong in type-bits no-prototype absent-spread unknown-road unpublished; do
		patch=scratch-f65/wrong-$wrong.patch
		git apply --check -p1 "$patch"
		echo "wrong $wrong: git apply --check -p1 exit $?, sha256 $(sha256sum "$patch" | cut -c1-64)"
		if git apply -p1 "$patch"; then
			go test -count=1 -timeout 40m -run "$pins" ./internal/oracle/ ./internal/lower/ -v > "$TMPDIR/mutant.log" 2>&1
			code=$?
			reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
			if [ $code -eq 0 ]; then echo "FAIL wrong $wrong SURVIVED"; ok=0; else echo "PASS wrong $wrong RED: $reds"; fi
			grep -E "(typed_presence|spread_overwrite|unknown)_test.go:[0-9]+:" "$TMPDIR/mutant.log" | grep -v "gate cache" | head -12 | cut -c1-500
			grep -qE "^(FAIL|ok)[[:space:]]" "$TMPDIR/mutant.log" || tail -20 "$TMPDIR/mutant.log"
			git apply -R -p1 "$patch" || { echo "FAIL wrong $wrong: git apply -R"; ok=0; }
		else
			echo "FAIL wrong $wrong: git apply"; ok=0
		fi
		restore "wrong $wrong"
	done
fi
if [ "$part" = b ]; then
	step "typescript source" sandboxed_typescript_source
	whole "internal/lower whole" go test -count=1 -timeout 80m ./internal/lower/
fi
if [ "$part" = c ]; then
	step products sandboxed_products ./internal/oracle
	whole "internal/oracle whole" go test -count=1 -timeout 80m ./internal/oracle/
	ADAMIC_TEST_WASI=1 whole "internal/native whole (ADAMIC_TEST_WASI=1)" go test -count=1 -timeout 80m ./internal/native/
fi
if [ "$part" = d ]; then
	step "typescript source" sandboxed_typescript_source
	whole "internal/ir whole" go test -count=1 -timeout 40m ./internal/ir/
	whole "stage3/census/latent/... whole" go test -count=1 -timeout 40m ./stage3/census/latent/...
	whole "internal/compilercensus whole" go test -count=1 -timeout 40m ./internal/compilercensus/
	whole "cmd/adamic-gate/parallelcensus whole" go test -count=1 -timeout 40m ./cmd/adamic-gate/parallelcensus/
fi
if [ "$part" = e ]; then
	programs="$TMPDIR/programs"
	mkdir -p "$programs/lib"
	step0 "$programs"
	cp scratch-f65/programs/*.ts "$programs/"
	cp scratch-f65/programs/lib/*.ts "$programs/lib/"
	sha256sum "$programs"/*.ts "$programs"/lib/*.ts | sed "s|$programs/||"
	keep
	for side in head main; do
		[ "$side" = main ] && main_files
		table "$side" "$programs"
		restore "run.sh at $side"
	done
fi
if [ "$part" = u ]; then
	# The count records the change moves, measured as the README says and printed whole, before and after.
	step products sandboxed_products ./internal/oracle
	mkdir -p "$TMPDIR/counts-before"
	cp -R internal/oracle/testdata/counts/. "$TMPDIR/counts-before/"
	go test -count=1 -timeout 80m -run '^TestCountsAreRecorded' ./internal/oracle/ -args -update-counts > "$TMPDIR/update.log" 2>&1
	echo "update-counts: exit $?"
	grep -E "^(ok|FAIL|---)" "$TMPDIR/update.log" | sort -u | head -20
	diff -r "$TMPDIR/counts-before" internal/oracle/testdata/counts
	echo "diff -r: exit $?"
	for file in $(diff -rq "$TMPDIR/counts-before" internal/oracle/testdata/counts | sed -nE 's/^Files [^ ]+ and ([^ ]+) differ$/\1/p'); do
		echo "moved $file: $(cat "$file")"
	done
fi
if [ "$part" = x ]; then
	# Why the two count records moved: the C each fixture emits at head and with main's four files, diffed.
	keep
	for side in head main; do
		[ "$side" = main ] && main_files
		step "build at $side" go build -o "$TMPDIR/adamic-$side" ./cmd/adamic
		for fixture in spread_undefined json_objects_null_variable; do
			"$TMPDIR/adamic-$side" c "internal/oracle/testdata/$fixture.a" > "$TMPDIR/$fixture.$side.c" 2> "$TMPDIR/$fixture.$side.err"
			echo "adamic c $fixture at $side: exit $?"
		done
		restore "the C at $side"
	done
	for fixture in spread_undefined json_objects_null_variable; do
		echo "== $fixture: main to head"
		diff "$TMPDIR/$fixture.main.c" "$TMPDIR/$fixture.head.c" | head -120 | cut -c1-400
	done
fi
if [ "$part" = v ]; then
	# Emit's three, verbose, whole: the JSON fixture on every backend, and the two count records it and spread_undefined hold.
	step products sandboxed_products ./internal/oracle
	echo "== internal/oracle/testdata/json_objects_null_variable.a"
	cat internal/oracle/testdata/json_objects_null_variable.a
	echo "== Node"
	node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/json_objects_null_variable.a
	echo "node exit $?"
	go test -count=1 -timeout 40m -run 'TestJSONObjectsAgreeWithNode/null_variable|TestCountsAreRecorded_000|TestCountsAreRecorded_024' ./internal/oracle/ -v > "$TMPDIR/three.log" 2>&1
	code=$?
	echo "go test: exit $code"
	grep -v "gate cache" "$TMPDIR/three.log" | grep -E "^(=== RUN|--- |    ---|ok|FAIL|PASS)|_test.go:[0-9]+:" | grep -E "null_variable|spread_undefined|TestCounts|^ok|^FAIL|^PASS|_test.go" | cut -c1-600 | head -60
	[ $code -eq 0 ] || ok=0
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }

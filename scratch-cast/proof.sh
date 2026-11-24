set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #g0qnfer, casts 1a: a field-refinement downcast is NotYet under rulings.md:18 (Adamic's item 19 on #bbssdcy).
# PART=a: gofmt, vet, TestCastStop verbose, both plants run as mutants and reversed, the branch probe at main and at
# head. PART=b: internal/lower whole. PART=c: internal/oracle whole, internal/native whole under WASI. PART=d:
# internal/ir, stage3/census/latent/..., internal/compilercensus, cmd/adamic-gate/parallelcensus, each whole.
# PART=e: the programs held to Node through cloud/probe/run.sh, at head and with main's two files in place.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -40; ok=0; fi; }
whole() { name=$1; shift; begun=$SECONDS; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name in $((SECONDS - begun)) s"; else echo "FAIL $name (exit $code) in $((SECONDS - begun)) s"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log"; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
head_files="internal/lower/cast_stop.go internal/lower/cast_proof.go"
keep() { for file in $head_files; do cp "$file" "$TMPDIR/$(basename "$file").head"; done; }
restore() {
	for file in $head_files; do
		cp "$TMPDIR/$(basename "$file").head" "$file"
		if cmp -s "$TMPDIR/$(basename "$file").head" "$file"; then echo "PASS restored $file after $1"; else echo "FAIL restore $file after $1"; ok=0; fi
	done
}
main_files() { cp scratch-cast/cast_stop.go.main internal/lower/cast_stop.go && cp scratch-cast/cast_proof.go.main internal/lower/cast_proof.go; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = a ]; then
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/cast_stop.go internal/lower/cast_proof.go internal/lower/cast_stop_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/
	step "cast stop" go test -count=1 -run 'TestCastStop|TestRecordStoragePromiseArmMeetsNoThenable' ./internal/lower/ -v
	subtests
	keep
	for plant in refinement-refused refinement-first refinement-methods; do
		patch=scratch-cast/plant-$plant.patch
		git apply --check -p1 "$patch"
		echo "plant $plant: git apply --check -p1 exit $?, sha256 $(sha256sum "$patch" | cut -c1-64)"
		if git apply -p1 "$patch"; then
			go test -count=1 -run 'TestCastStop|TestRecordStoragePromiseArmMeetsNoThenable' ./internal/lower/ -v > "$TMPDIR/mutant.log" 2>&1
			code=$?
			reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
			if [ $code -eq 0 ]; then echo "FAIL plant $plant SURVIVED"; ok=0; else echo "PASS plant $plant RED: $reds"; fi
			grep -E "(cast_stop_test|records_storage_nesting_test).go:[0-9]+:" "$TMPDIR/mutant.log" | head -8 | cut -c1-400
			git apply -R -p1 "$patch" || { echo "FAIL plant $plant: git apply -R"; ok=0; }
		else
			echo "FAIL plant $plant: git apply"; ok=0
		fi
		restore "plant $plant"
	done
	# The branch probe: temporary prints in castRoad and a test that lowers the site's shapes, at main (main's
	# cast_stop.go and cast_proof.go in place) and at head, then the head's files back, byte for byte.
	for side in main head; do
		[ "$side" = main ] && main_files
		python3 -I scratch-cast/branch-prints.py internal/lower/cast_proof.go || ok=0
		cp scratch-cast/branch_probe_test.go.txt internal/lower/zz_branch_probe_test.go
		go test -count=1 -run '^TestZZCastBranchProbe$' ./internal/lower/ -v > "$TMPDIR/probe-$side.log" 2>&1
		code=$?
		echo "branch probe at $side: exit $code"
		grep -E "BRANCH|PROBE" "$TMPDIR/probe-$side.log" | uniq | cut -c1-700
		grep -qE "PROBE" "$TMPDIR/probe-$side.log" || { tail -30 "$TMPDIR/probe-$side.log"; ok=0; }
		rm internal/lower/zz_branch_probe_test.go
		restore "the branch probe at $side"
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
	mkdir -p "$programs"
	cp scratch-cast/programs/*.ts "$programs/"
	# The control, a structural view that lowers at main: TestCheckedViewUntaggedSourceDispatch's structural/good.
	cp stage3/interface-downcasts/untagged/fixtures/structural-source-good.ts "$programs/control-structural-source-good.ts"
	keep
	for side in head main; do
		[ "$side" = main ] && main_files
		bash cloud/probe/run.sh "$programs" --logs "$TMPDIR/logs-$side" > "$TMPDIR/run-$side.log" 2>&1
		echo "run.sh at $side: exit $?"
		grep -E "^\| |^- head|^Probes" "$TMPDIR/run-$side.log"
		grep -qE "^\| [a-z]" "$TMPDIR/run-$side.log" || { tail -30 "$TMPDIR/run-$side.log"; ok=0; }
		for directory in "$TMPDIR/logs-$side"/*.ts; do
			echo "$side $(basename "$directory"): Node $(tr '\n' ' ' < "$directory/node.stdout")"
			for stage in js-compile native-compile; do
				[ -s "$directory/head.$stage.stderr" ] && echo "$side $(basename "$directory") $stage: $(head -3 "$directory/head.$stage.stderr" | tr '\n' ' ' | cut -c1-600)"
			done
			for backend in js native; do
				[ -f "$directory/head.$backend.stdout" ] && echo "$side $(basename "$directory") $backend: $(tr '\n' ' ' < "$directory/head.$backend.stdout")"
			done
		done
		restore "run.sh at $side"
	done
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }

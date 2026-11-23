set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #sbvnh96 and #yfmm1vv at their head: the cycle finder, the graph walk and the JSON reach walk relate an object
# literal by its regular widened type (relatedShape). PART=a: gofmt, vet, the pins verbose, the mutant (relatedShape
# returning the fresh type) red on the cast rows and green on the variable rows, put back, then lower, ir,
# stage3/fixtures, stage1's gaps, the latent census, compilercensus and parallelcensus whole. PART=b: oracle whole
# and native whole (ADAMIC_TEST_WASI=1).
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)|_test.go:[0-9]+:" "$TMPDIR/whole.log" | sort -u | head -50; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript
step "typescript source" sandboxed_typescript_source
pins='^(TestFreshLiteralCyclesAreCollected|TestFreshLiteralToJSONReachStops)$'
if [ "$part" = a ]; then
	target=internal/lower/fresh_relation.go
	cp "$target" "$TMPDIR/fresh_relation.go.head"
	trap 'cp "$TMPDIR/fresh_relation.go.head" internal/lower/fresh_relation.go' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/fresh_relation.go internal/lower/cycles.go internal/lower/graph_types.go internal/lower/library_json_stringify.go internal/lower/cycles_scanning_reference_test.go internal/oracle/fresh_literal_reach_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/oracle ./internal/lower
	step "pins" go test -count=1 -v -run "$pins" ./internal/oracle/
	subtests
	# The mutant: the literal related by its fresh type again.
	python3 -I - "$target" <<'PY'
import sys
path = sys.argv[1]
text = open(path).read()
old = "\treturn l.checker.GetWidenedType(regularTypeOfObjectLiteral(l.checker, shape))\n"
if text.count(old) != 1:
    print(f"FAIL mutant: the line occurs {text.count(old)} times"); sys.exit(1)
open(path, "w").write(text.replace(old, "\treturn shape\n"))
print("PASS mutant in place")
PY
	go test -count=1 -v -run "$pins" ./internal/oracle/ > "$TMPDIR/mutant.log" 2>&1 && { echo "FAIL mutant: green, a survivor"; ok=0; }
	reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
	echo "RED mutant: $reds"
	for row in cycle_cast_push.ts cycle_cast_map.ts cycle_cast_field.ts json_cast_to_json.ts; do
		case " $reds " in *"/$row "*) echo "PASS mutant red on $row" ;; *) echo "FAIL mutant survived on $row"; ok=0 ;; esac
	done
	case " $reds " in *"/cycle_variable_push.ts "*) echo "FAIL mutant reddened the variable row, which it can't reach"; ok=0 ;; esac
	cp "$TMPDIR/fresh_relation.go.head" "$target"
	if cmp -s "$TMPDIR/fresh_relation.go.head" "$target"; then echo "PASS restored $target"; else echo "FAIL restore $target"; ok=0; fi
	whole "lower whole" go test -count=1 ./internal/lower/
	whole "ir whole" go test -count=1 ./internal/ir/
	whole "stage3/fixtures whole" go test -count=1 ./stage3/fixtures/
	whole "stage1 gaps" go test -count=1 -run '^(TestParserRepresentationProbes|TestEachGapStandsWhereGapsMdSaysItDoes)$' ./stage1/cohere/markdownblocks/ ./stage1/cohere/static_single_assignment/
	whole "latent census whole" go test -count=1 ./stage3/census/latent/...
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi
if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle ./internal/native
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 ./internal/native/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }

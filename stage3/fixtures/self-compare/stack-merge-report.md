Merged the integration c2 stack into self comparison toward step 04.
Parents: 10ad5d57 and c79c75726375c937f0b18e18040e6769d916ed06.
Required validation and measured count-row changes are recorded below after completion.
The constant-folding mutant compiles and exits cleanly; sweep.a catches its incorrect NaN comparison.
This merge does not add loose comparison coercion or change the checked-write branch.

There were no content conflicts. The existing representation helpers and all integration changes are preserved. The stack's extracted test support replaces the old private unbalanced helper, so the counted test now uses internal/leakcheck.Unbalanced with the existing run adapter. Allocation balance remains checked.

Stage3 and counts exposed two inherited packed-cache presence stores, in reuse.go and emit_objects.go. Both now use adamic_slot_index on the returned slot. The two stage3 real fixture statuses for generic-optional-return and structural-method-statics now say Compiles only after recorded/current Node and sanitized native agreed. No .a source changed beyond this branch's three original witnesses.

TestSelfCompare passed against source Node in the JavaScript backend and release, sanitized, counted and wasm32-wasi native builds. LeakSanitizer and counted balance passed. TestSelfCompareConstantMutant compiled and ran without stderr or a failing exit code; its `NaN: true true` disagrees with Node's `NaN: false true`.

Commands send output directly to logs:

```sh
source /workspace/adamic-tools/env.sh
go build ./... > /tmp/stack-self-build-ready.log 2>&1
go vet ./internal/... > /tmp/stack-self-vet-ready.log 2>&1
python3 /tmp/stack-a-check.py --out /tmp/stack-self-a-check-ready > /tmp/stack-self-a-check-ready.log 2>&1
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestSelfCompare' -count=1 -v -timeout 10m > /tmp/stack-self-focus-ready.log 2>&1
go test ./stage3/fixtures -count=1 -timeout 15m > /tmp/stack-self-fixtures-ready.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -timeout 20m -args -update-counts > /tmp/stack-self-counts-ready.log 2>&1
```

The a-check selects only .a files changed against c79c7572 and uses the previously reported gate policy 3e339bbf06e0695f1b1223065ee8d4eef4813ba8. It checks the three original self-comparison fixtures.

The shared setup and WASI SDK prepared for this merge unit passed after initial merge markers were resolved. Setup timing lines: node 0.071s, Go 0.073s, markdown 0.204s, submodules 0.264s, clang 0.330s, WASI SDK 0.465s, build 68.329s, test binaries deferred 70.183s, cache warm 70.193s, total 70.669s. nproc is 5; cpu.max is 400000 100000.

Initial build/test failures from the extracted helper are retained in /tmp/stack-self-focus.log and /tmp/stack-self-counts.log. The initial stage3 run in /tmp/stack-self-delivery-fixtures.log records the cache emission error and stale statuses. Its superseded counts run was stopped before regeneration on the repaired tip. Final output is in the ready logs above; the all-mode oracle passed in 123.008s.

Final focused tests were rerun on the exact final emitter revision and passed in 1.703s. Both backends and all four native modes agree with Node, allocation balance and LeakSanitizer pass, and the NaN mutant compiles and is caught. Build and vet exit 0; changed-file a-check exit 0 (three proven .a files); stage3 PASS (111.628s).

Counts regeneration PASS (229.315s). Relative to 10ad5d57: 90 added rows, zero removed, zero existing rows moved. Relative to c79c7572: three added rows, zero removed, zero existing rows moved. The three rows are stage3/fixtures/self-compare/array-fill.a, closure.a and sweep.a. Exact added row names against both parents are in stack-counts.json.

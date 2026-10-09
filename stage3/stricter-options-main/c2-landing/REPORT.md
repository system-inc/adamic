Merged pinned c2 into stricter-options, preserving both branches' behavior.
Parents: a90264f00919e507f45c970de0cb6458df52ac8b and c79c75726375c937f0b18e18040e6769d916ed06.
Build, internal vet, 32 changed .a checks, stage3 and Linux counts regeneration pass.
All lane packages, 17 loader/library mutants, optional/catch mutants and the namespace mutant pass their intended assertions; the added parallel pending mutant is caught.
Checked optional views remain pending on compiler/views-rehearsal 2b6c032a; no full oracle, WASI oracle or new checker-entry census was run.

Six conflicts were resolved:

| File | Resolution |
|---|---|
| internal/ir/ir.go | Keep Uint16Array and Record. |
| internal/lower/lower_test.go | Keep the unawaited-async refusal; retain admitted primitive throws. |
| internal/lower/object.go | Keep numeric typed-array property boundaries and overload-result field checks. |
| internal/native/runtime/adamic.h | Keep tagged primitive/Error exceptions with thread-local carrier and independent pending flag. |
| internal/native/runtime/exceptions.c | Define the same combined thread-local exception state. |
| internal/oracle/counts.md | Regenerate the complete table on Linux. |

The merge message records each resolution and the additional interaction repairs. Normal and reused spreads and their mutants use returned slot pointers rather than the removed cache index. Optional copy probes preserve c2's atomic cache and registered shape metadata. Tests use the shared leak-check API. The parallel runtime retains earliest-index exception selection when the thrown value is undefined; its Node witness and pending-flag mutant run at one and four workers under ASan/UBSan and LeakSanitizer. The null tag remains an immutable sharing leaf. RegExp callbacks test pending state independently of their payload.

The two original Uint16 new-expression fixtures now have admitted registrations because c2 provides their storage and conversions. Both backends agree with Node under sanitizers, with source and expected output unchanged. This is covered along with existing Uint16 fixtures and four spread/reuse fixtures by the targeted native oracle.

Only stage0 changed in two real stage3 records: generic-optional-return/main.a and structural-method-statics/main.a moved from NotYet to Compiles after native agreed with recorded and current Node. Every byte outside stage0 is identical. No Compiles or CheckedStop regression was introduced. See stage0-audit.json.

Against stricter-options: 85 rows added, 0 changed, 0 removed; 1075 rows total.
Against c2: 27 rows added, 87 changed, 0 removed; 1075 rows total.
Every moved row, with old and new numbers, is listed in counts-audit.json.

Commands and raw outputs are in evidence/*.log.gz. Every test ran directly into a log file.

```sh
go build -p 1 ./...
go vet -p 2 ./internal/...
python3 /tmp/stricter-options-c2-acheck.py
# a-check uses git diff c79c75726375c937f0b18e18040e6769d916ed06 -- '*.a'
go test -p 2 ./stage3/fixtures -count=1 -timeout 30m
go test -p 1 ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go test -p 2 ./stage3/stricter-options ./stage3/stricter-records ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-indexed-all -count=1 -timeout 30m -v
go test -p 2 ./internal/oracle ./internal/native ./internal/lower -run 'Test(Caught|Optional|NumericTypedArray|ReadonlyRecord|NamespaceClosedCallGraphEdges|ParallelUndefined)' -count=1 -timeout 30m -v
python3 stage3/stricter-options/mutants.py
python3 stage3/stricter-options/production_mutants.py
python3 internal/load/testdata/project-lib/run_mutants.py /tmp/stricter-options-c2-project-lib-mutants
go test -overlay /tmp/stricter-options-namespace-mutant.json ./internal/lower -run '^TestNamespaceClosedCallGraphEdges$' -count=1
# The namespace overlay must fail with the class-expression cast panic, not a build failure.
go test -p 2 ./internal/lower -run '^TestWhatZeroOneRefusesIsRefusedWithAFix$' -count=1
go test -p 2 ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/(typed_arrays_uint16array|new_expression_uint16|new_expression_uint16_notyet|typed_arrays_uint16_stop|fallthrough_ownership|reuse_lent_global|class_instance_key_consume|class_inheritance_memory)\.a$|^TestNewExpressionUint16IsAdmitted$' -count=1 -timeout 30m -v
```

The attribution runner catches all six erasures by project-option assertions. The production runner catches four dispatcher/ownership erasures by production assertions. The library runner catches all seven by checker assertions. The lane packages rerun their indexed-presence, JSON-definedness, dictionary prototype, sparse compound read, nullable sentinel and runtime-classification mutants. Optional construction, copy, presence, storage and callback mutants and primitive/Error catch mutants are rerun by the focused tests. A build failure is not counted as a mutant kill.

Explicit skips remain: TestOptionalFieldCheckedViewPending names the views acceptance dependency; TestOptionalWideningCensus requires its corpus environment; TestNullableGuardTiming is opt-in. None is reported as acceptance passing.

Initial checks exposed the repaired runtime-null, packed-cache, leak-check API and stale Uint16 expectation interactions. A build also exhausted transient disk space while linking; the complete build then passed with one link job. No wrong production output with exit 0 was observed.

Built: Step 22 (#p9v82wa), one ordinary object with declared slots and an ordered index dictionary, reusing runtime/record.c.
Commits: approved base 8c013f1c, area-next-fixtures 4885cec50290686df487b62aac47c85d871ed40c, merge 5c7f63734; delivery SHA is reported with the push.
Commands and outputs: four acceptance witnesses plus a supplemental witness pass against Node in both backends, with native ASan, UBSan and Linux leaks; focused lowering/runtime/regression tests and vet pass after installing the pinned host dependency.
Mutants: index keys forced into fixed shape fail Node order; dictionary omitted from entries fails Node in both backends; optional dictionary miss fails Node; scalar-erasure silence fails the exact exit/diagnostic check; existing values, entries and null-slot mutants are caught.
Not covered: the full gate, all tsc Object.entries sites, other platforms, and general numeric or heterogeneous fixed-object views.

The approved base was used directly, then the requested area line was merged without conflicts. No other worker branch was merged. The scout commit 6945a8a4 has census scripts and witnesses rather than a standalone research document; those were read. No cohere implementation was copied.

Declared fields retain their ordinary object slots. The existing record dictionary stores index values and zero-valued presence/order tokens for declared fields. Tokens keep creation order across both stores, distinguish missing from present undefined, and record deletion/reinsertion. Existing ECMA integer sorting applies to the combined keys. Object.keys, values and entries, optional field access, copied views and erased reference reads all use that same object. Cleanup releases the dictionary and fixed slots separately.

Conservative assumption: preserve the lowering's proven homogeneous storage. Optional string fixed views are admitted only with compatible member contracts; scalar, heterogeneous, readonly-contract-losing, mutable-container and callback views without the reverse storage proof remain refused. Boolean-or-undefined record values use the existing union boxes. Unsupported erased scalar reads stop explicitly instead of interpreting numeric bits as a reference.

Acceptance sources are unchanged: 05_integer_order.a, 06_delete_readd.a, 07_optional_view.a and 13_strict_option.a in stage3/fixtures/records. The new .a witness additionally mixes computed index keys with fixed fields, writes absent keys, distinguishes own undefined, deletes/readds both stores, observes keys/values/entries, writes through optional views, copies them and reads an erased string view.

Only these status records change, both newly passing:
- records/07_optional_view.a: NotYet to Compiles.
- records/13_strict_option.a: NotYet to Compiles.

Every byte outside stage0 is unchanged, including every recorded Node observation. No Compiles record becomes Refused or NotYet. See evidence/status-changes.json.

Scoped commands (all output was written directly to the accompanying logs):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestMapLikeRecords'
go test ./internal/lower -count=1 -timeout 30m -v -run '^TestMapLikeRecord|^TestRecordForms$|^TestRecordRefusals$|^TestNamedRecord|^TestOptionalWidening'
go test ./internal/native -count=1 -timeout 30m -v -run '^TestRecordsAgainstNode$/(semantics|references|iteration|numeric|reads)$|^TestRecordMutants$/own-slot-null-read$|^TestRuntimeFieldLayoutsAreIncluded$'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestRecordOperationMutants$/(values|entries)$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(records_|optional_|object_|unknown|maplike_two_stores)'
npm ci --prefix stage3/api
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/unknown_narrowing_host.a$'
python3 stage3/maplike-records/run-mutants.py
go test ./internal/native -count=1 -v -run '^TestMapLikeScalarErasureGuard$'
go vet ./internal/ir ./internal/lower ./internal/native ./internal/oracle
ADAMIC_GATE_UNCACHED=1 go test ./stage3/fixtures -count=1 -timeout 30m -v -run '^TestFixtures$/records/(07_optional_view.a|13_strict_option.a)$' -args -update
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

The regression command initially failed before compilation on unknown_narrowing_host.a because @types/node 25.3.3 was absent. Installing the pinned API dependencies and rerunning that exact fixture passed. There were no observed wrong-output successes from the unmutated implementation. Source-mutant builds must run alone; an accidentally overlapping counts run was discarded and regenerated after restoration. The final acceptance also ran after restoration.

Counts regenerated on Linux: one new maplike_two_stores row (49 allocations, 49 frees, 154 retains, 122 releases, peak 15, zero regions), plus six existing record rows. Three operation/ownership/scalar fixtures lose two temporary allocations from centralized reflection; three named/optional fixtures record the additional release of zero-valued fixed-field tokens. The raw table diff is authoritative, including intentional-stop rows whose heap counts do not balance because normal cleanup was not reached.

Setup used GOPROXY=https://proxy.golang.org|direct. Reported timings: Go/Node 0.102s, clang 0.540s, markdown 1.235s, submodules 3.668s, go build 187.420s, cache 187.567s, total 187.602s. nproc=5; quota=4. Toolchain: Go 1.27.1, Node 24.19.0, clang 20.1.8.

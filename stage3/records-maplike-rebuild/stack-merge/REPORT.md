Merged c2 into records-maplike, keeping records, hidden-boundary and optional-presence behavior.
Delivery: the merge commit containing this report on compiler/records-maplike; parents b58629a2 and c79c7572.
Checks: build, vet, 38 a-check inputs, focused lower/oracle/entries/native checks, stage3 fixtures and regenerated counts pass.
Mutants: records source and runtime mutants, optional-presence/copy mutants and missing spread publication are caught.
Not covered: native scanner execution, the full oracle package, the full gate or other platforms.

## Base and conflicts

The uncommitted e3380f23 merge was aborted before merging
c79c75726375c937f0b18e18040e6769d916ed06 into pre-merge records tip
b58629a20aa22546bcb92f02c6f0e5e0ac1e9e36. The discarded c1 tip is not an
ancestor of either parent. No main or area branch was pushed or merged into.

Two text conflicts:

* internal/ir/ir.go: preserve Uint16Array and explicit typed-array classification;
  add the separate Record representation and reference classification.
* internal/lower/refusals.go: preserve c2 safety checks, hidden boundaries and
  optional presence. Replace blanket index-signature refusal with record storage
  proofs; retain detached-own and record view restrictions. Property deletes
  reach the optional-presence proof; dictionary deletes keep the record proof.

Semantic compatibility updates preserve unsupported readonly/open optional-index
boundaries with current diagnostics, use leakChecked for record ownership, and
adapt the optional-copy private runtime probe to packed slot caches and shape
registration metadata. Private symbol renaming includes the new slot-cache helper.
The checked-field alias fixture remains NotYet, with its current diagnostic pinned.

Counts initially exposed two inherited c2 emission sites using cache.index after
that member was removed. Ordinary and reused spread publication now obtain the
index from the actual slot via adamic_slot_index, matching neighboring metadata
reads and c2's runtime. Removing publication produces different Node stdout with
exit 0 and empty stderr in TestOptionalFieldConstructionCatchesAbsentPublication.
The existing layout-order mutant uses the same current slot-index API.

## Validation

All output was written to files, preserved under evidence/. Commands used
GOFLAGS=-p=2, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1.

```
go build ./... # exit 0
go build -o /tmp/records-stack-validation/adamic ./cmd/adamic # exit 0
go vet ./internal/... # exit 0
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts # exit 0
python3 /tmp/records-stack-validation/a-check.py docs/hidden-boundaries/landing/fast-gate.py.txt # exit 0
go test ./internal/lower -run 'Record|DetachedOwn|Enumeration|Entries|LibraryMethod|Hidden|Optional|TypedArray' -count=1 -timeout 10m -v # exit 0
go test ./internal/oracle -run 'TestRecord|TestPartialRecord|TestNamedRecord|TestDetachedOwn|TestOptionalField|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|optional_|object_|enum|library_method_values|library_string_raw|method_coverage_object_descriptors|entries_)' -count=1 -timeout 30m -v # exit 0
go test ./internal/oracle -run '^TestEntriesAcceptance$|^TestEntriesProvenance$|^TestEntriesRuntimeReadiness$' -count=1 -timeout 20m -v # exit 0
go test ./internal/native -run '^TestRecordsAgainstNode$|^TestRecordMutants$|^TestRecordReadMutants$' -count=1 -timeout 15m -v # exit 0
go test ./stage3/fixtures -count=1 -timeout 30m -v # exit 0
python3 stage3/records-maplike-rebuild/run-mutants.py # exit 0
python3 stage3/named-index-records/run-mutants.py # exit 0
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m # exit 0
```

The final command selectors and exits are also in checks.json. Earlier failed
attempts are preserved separately; only final corrected mutant runs count as
proof. No new oracle fixture was introduced by this merge. The counts table was
regenerated, never hand merged. The 38 changed .a inputs were selected against c2.
The inherited checked-view optional-write fixture remains explicitly pending.
Two stale c2 real-fixture statuses, generic-optional-return and
structural-method-statics, were refreshed to Compiles through guarded -update;
their Node and native checks pass and their source/provenance remain unchanged.

Setup: Node ready 0.025s, Go 0.026s, submodules 0.076s, markdown ready 0.082s
(step 0.015s), clang ready 0.163s, build ready 19.847s, deferred tests 19.981s,
cache warm 19.983s, done 20.014s. nproc=5, cgroup quota four CPUs.
Go 1.27.1, Node 24.19.0, clang 20.1.8; env /workspace/adamic-tools/env.sh.
An initial setup exhausted disposable build-cache space; rebuilding sequentially
with bounded concurrency succeeded. The original worktree was preserved after an
interrupted checkout; this merge used an isolated worktree. Remote product fetch
returned 403; the pipeline's supported local --build-product path was used.

## Counts

Against c2: 1033 -> 1066 rows, 33 additions, no removals, three changed rows.
Against records pre-tip: 976 -> 1066 rows, 90 inherited c2 additions, no changed
or removed rows. Both full comparisons are in counts-audit.json. Values below
are allocations, frees, retains, releases, peak and regions, in that order.
The three changed rows match the records pre-tip's measured counts; the merged
runtime API fixes do not change those measurements. Node checks retain the same
observable behavior while the lane's library adapters avoid the old ownership
work. Every added row belongs to the retained records lane.

| Fixture | Counts | Reason |
| --- | --- | --- |
| internal/oracle/testdata/library_string_raw.a | 45, 45, 55, 89, 11, 0 -> 45, 45, 51, 85, 11, 0 | Records-lane library dispatch and intrinsic alias typing; Node parity passed. |
| internal/oracle/testdata/method_coverage_object_descriptors.a | 47, 47, 88, 127, 7, 0 -> 46, 46, 61, 98, 6, 0 | Records-lane library dispatch and intrinsic alias typing; Node parity passed. |
| internal/oracle/testdata/library_method_values.a | 137, 137, 154, 295, 42, 0 -> 136, 136, 152, 291, 41, 0 | Records-lane library dispatch and intrinsic alias typing; Node parity passed. |
| internal/oracle/testdata/detached_own_parser.a | 4, 4, 0, 4, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/detached_own_records.a | 2, 2, 3, 4, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/detached_own_objects.a | 6, 6, 4, 8, 5, 0 | Added records-lane fixture. |
| internal/oracle/testdata/detached_own_unready.a | 0, 0, 0, 0, 0, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_named_invalidated.a | 3, 1, 11, 7, 3, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_named_options.a | 8, 8, 1, 6, 4, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_named_operations.a | 19, 19, 42, 48, 7, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_named_union.a | 14, 14, 30, 37, 5, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_named_intrinsics.a | 14, 14, 13, 28, 7, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_named_alias_unready.a | 0, 0, 0, 0, 0, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_discarded.a | 4, 4, 4, 10, 4, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_guarded_snapshot.a | 10, 10, 15, 27, 6, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_compare_missing_scalar.a | 63, 63, 75, 87, 9, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_compare_properties_left.a | 4, 0, 4, 3, 4, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_compare_properties_right.a | 4, 0, 6, 3, 4, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_environment_boundary.a | 2, 0, 1, 2, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_partial_parser.a | 5, 5, 0, 5, 5, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_partial_views.a | 20, 20, 54, 44, 6, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_partial_undefined.a | 15, 15, 13, 23, 8, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_operations.a | 64, 64, 136, 98, 21, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_ownership.a | 37, 37, 62, 60, 18, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_for_in.a | 6, 6, 28, 24, 3, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_census.a | 16, 16, 38, 45, 9, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_scalars.a | 23, 23, 28, 31, 8, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_optional.a | 12, 12, 11, 15, 6, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_narrowed_number.a | 2, 0, 4, 4, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_narrowed_reference.a | 3, 1, 7, 7, 3, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_prototype_read.a | 2, 0, 1, 1, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_prototype_in.a | 2, 0, 1, 1, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_prototype_set.a | 2, 0, 2, 1, 2, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_type_only_scanner.a | 0, 0, 0, 0, 0, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_type_only_composition.a | 0, 0, 0, 0, 0, 0 | Added records-lane fixture. |
| internal/oracle/testdata/records_type_only_value.a | 5, 5, 7, 12, 4, 0 | Added records-lane fixture. |

## Mutants

The complete records mutant-to-catcher table in ../REPORT.md applies unchanged;
all listed applicable checks were rerun here. The 12 lowering overlays and three
named-contract mutants are listed individually in static-mutants.log and
named-mutants.log, with their intended failing assertions. Nine native runtime
mutants are in native.log. Operation, ownership, readiness, narrowing, detached
own and absent-entry mutations are in oracle.log. LeakSanitizer catches dropped
releases, AddressSanitizer catches freed keys, Node stdout catches changed
operations/order, and checked-exit pins catch missing guards.

The optional-presence mutations for initial presence, insertion order, deletion,
reservation, slot storage, and static enumeration retain their catchers.
TestOptionalFieldCopyState catches independent dropped presence, readiness and
representation copying plus bitmap overlap, in release and sanitized builds.
The new publication mutation compiles and runs; its stdout differs from Node.
No compile failure is claimed as a successful semantic mutant catcher.

## Scanner

The scanner source/driver profile is the existing lane's 3ea66219 dependency,
read only in scratch, not merged. TypeScript source is 050880ce. Node's 81-file
profile exits 0 with 1,369,432 token rows, 466 errors and 108,019,935 bytes;
SHA256 9617f24e9c9f221dcf933a4624fe4ab8b389d2b1e09d6416548bd7afeeaee686.
The comparison control passes and token-end mutant differs. The rebuilt merged
compiler with ADAMIC_NATIVE_SPLIT=0 first stops at debug.ts:14:14, refusing
Debug.fail's (Error as any).captureStackTrace cast (adamic/no-unchecked-cast).
Scanner compilation exits 1 before native execution; no native scanner parity
is claimed. scanner-final.log pins this stop with the final built compiler.

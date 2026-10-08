Built records and named-index support on library e40216f2, retaining library behavior.
Commits: seven non-merge replays, 57db08b0 and d36202d5; final evidence commit follows.
Commands: uncached Node oracle, touched packages, full stage3 fixtures and counts passed; logs in evidence.
Mutants: finite and named absent entries, named guards/readiness and three source mutants caught.
Not covered: complete repository test gate, complete native scanner or non-Linux targets.

## Branch and conflict resolution

Only codex/records-maplike-library is updated. No main or area branch was merged into or pushed.
The branch starts at e40216f23caa519ad90e011dfc3a47d5c3bfeb6e (origin/area/library).
Only these non-merge commits were cherry-picked; no feature merge history was imported:

| Source on main variant | Replayed commit | Purpose |
| --- | --- | --- |
| 560ab17c | 64a0b574 | Records compiler design |
| 92f7a99a | 2954f84e | Mutable string dictionaries |
| 37001c8e | 59430e25 | Own-guarded inherited reads |
| 60d7dec5 | ef2c9b9a | Type-only MapLike fixtures |
| 2636b8e4 | a7a5fb43 | Detached own-property helper |
| 85dc76eb | c5a80e01 | Finite partial records sharing MapLike storage |
| 12d29536 | 40b73211 | Named-index dictionaries and Object alias dependencies |

The original named-index worker implementation is c20e5b56. Only its own change
and the Object alias dependencies already reconciled on the main variant were taken.
Runtime records were already present on the library base.

Conflicts were resolved hunk by hunk in JavaScript emission, expression lowering,
object-call dispatch, refusals, statements and native expression emission.
Library Buffer, filesystem, process, Error, nullable-reference, boxed unknown/object,
and full string-code spread behavior remain alongside records.
Enum and existing library index-storage paths retain their origin checks.
Boxed String and Uint8Array bypass the dictionary storage gate.
Opaque-object argument refusal applies only to the proven own-property helper ABI;
ordinary library object parameters retain tagged Union storage.
Nullable records remain NotYet because this records path has no nullable schema metadata.
The Object alias read whitelist is retained with the other library method proofs.
The records JSON path now passes the full element schema to library's scalar helper.

## Integration red and stage3 status changes

objects/27_delete_substitution.a declares numeric-index maps. Its old refusal
was at the index signature, before delete; it was never a fixed-object delete refusal.
Numeric dictionary storage is still unsupported, so the correct outcome is NotYet
at 12:23, with the exact unsupported-storage message. Commit 57db08b0 records
that decision and the library JSON schema compatibility fix. Node remains unchanged.

The full stage3 suite then exposed ten more stale statuses caused by records support.
Six now compile and agree with Node: records 01_has_property, 02_get_property,
03_own_keys, 05_integer_order, 06_delete_readd and 16_built_strings.
Four reach later NotYet checks: objects/10_build_options (unsafe unrestricted named
write), taste/06_localized_message (value/string binary operation),
records/13_strict_option (boolean|undefined slots), records/07_optional_view
(dictionary to incompatible fixed-object storage). No Node output or provenance changed.
All eleven exact before/after status entries are in evidence/status-delta.json.
The full go test ./stage3/fixtures passes in 20.069s before push.

## Counts

The complete -args -update-counts run passes in 111.797s. The generated table has
647 rows: all 614 library rows retained, plus 33 new records rows.
Every new row and the one existing change are in evidence/counts-delta.json.
Replay-only orphan method/namespace rows and duplicate context rows were removed
by the generator; none were rows from the actual library base.

The existing node_fs_directory_system.a row changes from
3160/3160/5689/4739/1893/0 to 3300/3300/5971/4949/1998/0
(allocations/frees/retains/releases/peak/regions). This input fixture enumerates
its working directory with an empty path. The tree adds 33 immediate fixture files
and two directories, so it now observes 35 more entries. The increases are balanced
and correspond to processing that larger input; the Node comparison is rerun.
The three records missing-read counts differ from incoming main-variant rows
because library's optional scalar lowering avoids extra reference temporaries;
they are new rows on this base, not changes to existing library rows.

## Validation commands

All test output is redirected to logs; final logs are compressed under evidence.

```sh
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run 'TestPartialRecord|TestNamedRecord|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|optional_|object_|enum|node_|library_error|unknown|json_stringify|from_code|process_observations)'
go test ./internal/lower ./internal/ir ./internal/flow ./internal/fresh ./internal/javascript -count=1 -timeout 30m
go test ./stage3/fixtures -count=1 -timeout 30m
go test ./internal/oracle -count=1 -timeout 30m -run '^TestCountsAreRecorded$' -args -update-counts
NAMED_INDEX_MUTANT_LOGS=/tmp/maplike-library-static-mutants python3 stage3/named-index-records/run-mutants.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestRecord|^TestDetachedOwn|TestInputAgreesWithNode/internal/oracle/testdata/node_fs_directory_system.a'
go test ./internal/native -count=1 -timeout 10m -v -run '^TestRecordsAgainstNode$|^TestRecordReadMutants$'
go vet ./...
gofmt -l cmd internal
git diff --check
```

The broad uncached oracle includes the requested 73 original and six named-index
fixtures, plus library coverage, and passes in 61.933s: native hits 0/misses 318,
Node hits 0/misses 255. Native release, sanitizer and leak comparisons use the
existing oracle; intentional checked-stop fixtures retain their pins.
Lower passes in 196.702s, IR 30.550s, flow 214.573s, fresh 164.280s;
JavaScript has no package tests. Targeted compatibility/refusal checks pass in 5.722s.
Vet and gofmt logs are empty.

Setup timings: Node 0.025s, Go 0.028s, markdown 0.078s, submodules 0.078s,
clang 0.218s, build 75.145s, cache warm 75.801s, total 75.971s.
nproc is 5, with a four-CPU quota; Go 1.27.1, Node v24.19.0, clang 20.1.8.

## MapLike replay and mutants

corePublic-MapLike.a is unchanged from scanner-native-3 503d2e0d's 01-index.a.
The final CLI is rebuilt, then source Node, native and emitted JavaScript each
exit 0, print exactly ok followed by newline, and have empty stderr.
Both byte comparisons are empty. Source and emitted JavaScript run through
oracle/node.mjs. This proves this MapLike refusal is closed, not that the full
scanner now builds or prints byte-identical token output.

Finite and named absent-entry mutants synthesize an own undefined property;
both backends disagree with Node's absent-key result. Named member-kind guard
removal misses a required checked stop; alias readiness removal misses Node's
binding-readiness stop. Three real source mutants bypass named read typing,
unrestricted-write proof and named contract invariance, each caught by its
pinned lower test. Source is restored after each. Additional mutants all pass their detection tests (record/helper run 4.456s,
runtime run 21.944s):

| Mutant | What catches it |
| --- | --- |
| Record read or write adds one | Node stdout |
| Delete is a no-op | Node stdout |
| in or has_own reverses its result | Node stdout |
| Keys swaps the first two keys | Node stdout |
| Values adds one; entries empties each key | Node stdout |
| Spread increments values; for-in returns no keys | Node stdout |
| Stringify uses Map schema | Node stdout |
| Prototype read-through | Required checked stop, while mutant runs as Node |
| Lost ownership releases | LeakSanitizer alone |
| Eager or repeated coalesce assignment fallback | Node stdout |
| Unchecked deleted scalar read | Checked exit 70 |
| Detached helper treats inherited names as own | Node stdout, records and objects |
| Detached helper readiness removed | Node initialization stop |
| Discarded/snapshot/missing-scalar own guard removed | Missing-member stop where Node finishes |
| Runtime prototype membership restored | Exact stop fixture |
| Runtime missing read silently succeeds | Exact stop fixture |
| Runtime own hit checked as missing | Own-hit fixture |

The changed filesystem input fixture also passes its uncached Node comparison.

The full repository test gate, full scanner token replay and other targets were not run.

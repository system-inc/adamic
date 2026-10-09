Rebuilt mutable records, finite partial records and named string dictionaries on main for step 12, task #ns5e098.
Commits: base 031a1259; net lane change efe9f404..fcddeb29; delivery is the commit containing this report.
Checks: 115 uncached oracle fixtures, both backend checks, sanitizers, build, vet, stage 1 gaps, stage 3, 38 a-check inputs and regenerated counts pass.
Mutants: all applicable lane runtime and lowering checks were rerun and caught, including 15 source mutations and nine record runtime mutations.
Not covered: a native scanner execution, the full oracle package, other platforms, or the complete upstream compiler.

## Rebuild and conflicts

Created compiler/records-maplike directly from origin/main
031a1259bc7973934792dc6cb1bd4074fc2204b9. Applied the binary diff from
lane base efe9f4042049234e5a52639fe77b47c311fd530c to
fcddeb299460708f76e8436b4d887a260a6b84f1, excluding counts.md for Linux
regeneration. No lane history, old base, area-next merge, or unlanded worker
branch was merged. The lane's documentation and historical evidence are
preserved; this report describes the current rebuild.

Nine paths conflicted. Each resolution is also recorded in the commit message.

| Path | Resolution |
| --- | --- |
| CohereSettings.json | Retain main's input-spread exclusions and the lane's records_buckets exclusion. |
| internal/ir/ir.go | Retain typed arrays and add the separate Record representation; both remain reference types. |
| internal/javascript/javascript.go | Retain argument-count helpers and Node Buffer support; add record runtime emission and expression dispatch. |
| internal/lower/expression.go | Retain typed arrays, Buffer, unknown and union views, enum dispatch and main's adapters; add record recognition, dispatch and recursive storage proofs. |
| internal/lower/library_method_values.go | Retain main's complete adapters; alias readiness reads use the local's actual type, including the lane's boolean intrinsic marker. |
| internal/lower/library_object.go | Route dictionary operations using the actual argument list, preserving main's descriptor, mutation and shape checks. |
| internal/lower/locals.go | Retain evolving-object inference and readiness; add detached own-property intrinsic markers and proven opaque object helper parameters. |
| internal/lower/object.go | Retain main's array, Buffer, user-method and library adapters; add dictionary literals and detached own calls. |
| internal/lower/refusals.go | Replace only blanket index-signature and delete refusals with record-specific admission. Fixed-object deletion remains refused. Retain main's predicate, assertion, cast, namespace, descriptor and enumeration checks, alongside the lane's storage and detached-method checks. |

Main already has an ObjectLiteral record emitter. The new homogeneous Record
emitter is named dictionaryLiteral so neither implementation replaces the other.

The rebuild also retains main's growing-literal entries storage when its existing
origin proof applies. Those literals use boxed Object storage for structural
views; homogeneous dictionaries use Record storage. Readonly enumeration views
keep their existing representation. Dictionary mutations bypass the older closed
const enumeration restriction and instead receive the lane's named-property and
cycle checks. Two lower tests that formerly pinned those dictionary gaps now pin
successful admission; the fixed-object spread gap remains pinned.

The small logical-record and nullable-record refusals from 8c013f1c were reused
without its merge history. Their guard-removal mutants fail. Prototype hasOwn
calls use the actual dictionary argument list; opaque own-property helper checks
apply to those helper parameters without blocking main's unrelated adapters.

## Scanner observation

Reproduced the scanner profile described by 3ea66219, with its exact driver and
corpus helper, source TypeScript 6.0.3 at 050880ce, and the same type-only input
adaptation. Main lacks adaptation 42-scanner-any, so the scratch pipeline reads
that adaptation from 3ea66219. No source or adaptation from that branch is merged
into the delivery branch. run-scanner.py records this dependency explicitly.

A Go overlay restoring main's compiler files was built on this machine. Against
the same scanner slice, it first stops at:

```
corePublic.ts:9:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added
```

The rebuilt compiler, with split disabled and enabled, gets past MapLike and
first stops at:

```
debug.ts:14:14: Adamic 0.1 refuses a cast the runtime can't check ... (adamic/no-unchecked-cast)
```

The stopping program is Debug.fail's `(Error as any).captureStackTrace` access.
The namespace stop at debug.ts:8:5 is already handled by main.

The full-tree Node driver and both sliced Node runs agree byte for byte:
509,014 tokens with trivia skipped and 860,418 with trivia retained, totaling
1,369,432 token rows across 81 files. The output contains 466 diagnostic rows and
108,020,259 bytes. Its SHA256 is
1d738e790e7e434d64cebfdf6b32218f75a1762847efcbf3fd11c5d823814d25.
The comparison control passes and each token-end mutant produces diff exit 1.
Native compilation stops before producing an executable, so there is no native
token count or native scanner agreement claim. The standalone corePublic MapLike
witness prints `ok` on Node and both backends; its native sanitized counted build
finishes with zero allocations and frees.

## Validation

Every test command wrote directly to a log. Current logs are compressed under
[evidence](evidence/). No full oracle or full repository gate was run.

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go build ./...
go vet ./internal/...
go test ./internal/lower -run 'Record|DetachedOwn|Enumeration|Entries|LibraryMethod' -count=1 -timeout 10m -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecord|TestPartialRecord|TestNamedRecord|TestDetachedOwn|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|optional_|object_|enum|library_method_values|library_string_raw|method_coverage_object_descriptors|entries_)' -count=1 -timeout 30m -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestEntriesAcceptance$|^TestEntriesProvenance$|^TestEntriesRuntimeReadiness$' -count=1 -timeout 20m -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestRecordCensusComparisonBuckets$' -count=1 -timeout 10m -v
go test ./internal/native -run '^TestRecordsAgainstNode$|^TestRecordMutants$|^TestRecordReadMutants$' -count=1 -timeout 15m -v
go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 30m
go test ./stage3/fixtures -count=1 -timeout 30m -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m
NAMED_INDEX_MUTANT_LOGS=/tmp/records-maplike-named-mutants python3 stage3/named-index-records/run-mutants.py
python3 stage3/records-maplike-rebuild/run-mutants.py
python3 stage3/records-maplike-rebuild/audit-status.py
```

All final commands pass. The focused oracle runs 115 registered fixtures with
312 native cache misses, 293 Node misses and zero hits. It includes native release,
AddressSanitizer, UndefinedBehaviorSanitizer and leak checks, plus the lane's
JavaScript comparisons and checked stops. Main's entries acceptance, provenance
and readiness checks pass separately.

The project's a-check method from fast-gate fbac28c6 was run against only the 38
added or changed .a files relative to origin/main. All pass. Three inherited
records_buckets witnesses deliberately retain the upstream boolean console calls;
they get first-line `// a-check: type error TS2345` pins. Their programs and Node
expected outputs are unchanged. Their existing Node comparison and adapted any
boundary tests were rerun after adding those headers and pass.

Setup succeeded with Go 1.27.1, Node 24.19.0 and clang 20.1.8. Timing lines:
Go 0.027s, Node 0.028s, submodules 0.071s, markdown step 0.007s / ready 0.079s,
clang 0.175s, build ready 42.555s, deferred test binaries 42.781s,
build cache warm 42.783s, done 42.819s. nproc=5; cgroup quota is four CPUs.
The inherited Go build cache was cleared when scratch capacity became limited;
logs and sources were retained. The env file is /workspace/adamic-tools/env.sh.

## Mutants observed

| Mutation | Check that catches it |
| --- | --- |
| Read, write, delete, in, hasOwn, keys, values, entries, spread, for-in, stringify: eleven emitted operation mutations | TestRecordOperationMutants compares stdout with Node. Mutants compile and finish sanitizer-clean. |
| Inherited JavaScript record read | TestRecordPrototypeMutant requires the checked exit rather than Node's inherited result. |
| Dropped record releases | TestRecordOwnershipMutant: LeakSanitizer. |
| Eager coalescing fallback | TestRecordCoalesceMutant: Node stdout. |
| Removed scalar narrowing check | TestRecordNarrowingMutant: required exit 70. |
| Removed own-read guards: discarded, guarded snapshot, scalar comparison | TestRecordObservationGuardMutants: missing-member stop where Node finishes. |
| Detached own changed to inherited membership, on records and fixed objects | TestDetachedOwnInheritedMutant: Node stdout. |
| Removed detached alias readiness | TestDetachedOwnReadinessMutant: required exit 70. |
| Partial and named absent entries inserted as undefined-present | TestPartialRecordAbsentEntryMutant and TestNamedRecordAbsentEntryMutant: stdout in both backends. |
| Removed named member-kind guard | TestNamedRecordTypeGuardMutant: required exit 70. |
| Removed named alias-copy readiness | TestNamedRecordAliasReadinessMutant: required stop in both backends. |
| Named read through index type; unrestricted named write; erased named contracts | Three named source mutations: TestNamedRecordReadTypes and TestNamedRecordRefusals fail. |
| Readonly and numeric signatures admitted; literal prototype names admitted; storage views, mutable invariance, cycles and shallow-spread invariance bypassed | Seven source overlays: TestRecordRefusals or TestRecordPrototypeLiteralNames fail on the intended witness. |
| Logical and nullable record conversions admitted | Two source overlays: TestRecordRefusals fails on the intended witness. |
| Opaque own-helper argument guard and shorthand escape guard removed | Two source overlays: TestDetachedOwnRepresentation and TestDetachedOwnRefusals fail. |
| Blanket index refusal restored | Source overlay: TestRecordForms rejects the regression. |
| Integer insertion order, uint32 maximum classification, deleted-key iteration | Three production runtime mutations: TestRecordMutants differs from Node. |
| Overwrite key leaked, stored key freed, own slot silently null | Three production runtime mutations: LeakSanitizer, AddressSanitizer and the own-hit check. |
| Inherited membership restored, missing read silently null, own read checked as missing | Three runtime mutations: TestRecordReadMutants requires exact diagnostic, exit and own-hit behavior. |

The old blanket mixed-named-signature refusal is superseded by named dictionary
support; its current contracts are tested by the three named source mutations.
The extra source runner uses Go overlays, leaving production files untouched.
The lane's three named mutations restore their source after each run. An initial
opaque-argument overlay failed to compile because its removed condition left a
local unused; that run was rejected as evidence. The corrected overlay compiles
and fails its actual admission assertion. Only final corrected logs are retained.

## Counts and stage 3 audit

counts.md was generated on this Linux machine, never hand edited. All 958 main
rows remain, with 33 additions and three changed rows: library_string_raw,
method_coverage_object_descriptors and library_method_values. The complete
before/after rows are in [counts-audit.json](evidence/counts-audit.json).
The independent TestCountsAreRecorded run passes.

[status-audit.json](evidence/status-audit.json) is the committed eleven-entry
stage 3 audit list. Deliberate regressions from Compiles to Refused or NotYet:
none. Six record fixtures move from Refused to Compiles after current Node,
recorded Node and native sanitizer/leak comparisons agree. Four move from the
blanket index refusal to their explicit NotYet boundary. objects/10_build_options
remains NotYet with the named-property write boundary replacing the closed-origin
boundary. Every Node observation and every byte outside stage0 is preserved.

The normal `-update` mechanism refreshed native successes. update-status.py then
used a test overlay to refresh only NotYet outcomes whose previous outcome was
neither Compiles nor CheckedStop, still requiring recorded Node agreement. No
gate/test source is changed. The complete unmodified stage3/fixtures test passes
afterward. audit-status.py independently checks all status files against main.

Continued checked-views integration: arrays/parser, mixed unions through 37763565, object/primitive unions, intersections and newest callables merged; existing optional, untagged and dictionaries retained.
Commits: start d718a9ffa0432245b69a3545c3ca080db865e0bb; latest green compiler merge 03aea1cbede84da46e84c78d599f287bc8585c1b; full merge ledger below; only codex/views-integration pushed.
Validation: scoped lower/native/JavaScript tests and the requested view/union/tuple/dictionary/callable/optional/array oracle filter passed after every landed merge; latest oracle 939.959s.
Mutants: executable array, callable boundary, primitive, object/primitive and intersection mutants were run and restored; redundant binding-hook survivor and partial callable payload mutant result are explicitly recorded below.
Limits: owner and tuple attempts left out; tuple retry has Map certificate failures and a JavaScript/Node disagreement; 41 global counts, fourteen optional and two lower refusal failures reproduce the starting commit.

## Merge ledger

| Item | Fetched tip | Conflict decisions | Branch SHA after item | Failures |
| --- | --- | --- | --- | --- |
| Optional boolean | a2eb65ca76816f895f846f78a210bc1717a5f7c4 | Already an ancestor | d718a9ffa0432245b69a3545c3ca080db865e0bb | None introduced |
| Untagged object unions | cd32db4234fb0a65a6ad199474b802c9e7491054 | Already an ancestor | d718a9ffa0432245b69a3545c3ca080db865e0bb | None introduced |
| Lazy admission owner | d5b3c4a9bde6ee28fe38568ef4b720aa9c2b68e3 | Aborted before resolving 18 paths; Map adapter restoration must preserve boxed dispatch and ownership | d718a9ffa0432245b69a3545c3ca080db865e0bb | Not runtime-tested; unresolved merge boundary |
| Arrays and parser | a3b0e3570fdf83fa3071b7f34ff9780d5334d5f7 | Plan: retain existing duplicate sections, whitespace and both final reports; counts: retain every distinct row and reject differing values for one fixture | e636841dd3d2996e471d97dd23712e7dbaa6aaeb | Required filter passes; inherited failures below |
| Mixed unions, first tip | d78c3f5cf19959be2bdb2ba4b72951750303cc4a | One plan hunk: append both independent reports; update the primitive helper to preserve its wrong-boolean runtime refusal | 6d9811a91da9346217f45c0c5223df9d194363eb | Stale compile-only helper expectation fixed; inherited failures unchanged |
| Mixed unions, requested newer tip | 37763565c317a80c3b362e475edab8b61ab0eafd | No conflicts; retain producer certificates and original reference obligations; update stale array-gap assertions | 93a086abbfbaa3d36ccb9bbe2516733452d2b521 | Required tests pass; same 41 inherited counts failures |
| Object and primitive unions | 8abb52a1518b9d8c3f0dbde993551d4f01ba10e2 | No conflicts; retain unread storage admission and named tuple-member refusal | b8b419d2b2e6f6a3bf62e02c749a6cd9f8465437 | Required tests pass; counts retain 41 baseline failures |
| Intersections | 4c3c3009c1ab74e4da902ffa9357b81ec0cf7d95 | Five hunks: both binding hooks, both plan reports, all count rows, superseded gap lists and both blocker histories | 167e5fdb779e8c6a64558a3c96b768b7067e776b | Required filter passes; corrected JSDoc runtime probes pass; redundant hook mutant survives |
| Callables | 433e1ec09fdba7c83fe2a3d5153cf28962a23022 | One counts hunk: all allocation rows before retained predicate table; update five closed union gap expectations | 03aea1cbede84da46e84c78d599f287bc8585c1b | Required filter passes; partial payload mutant runner result recorded |
| Tuples | 5c54e8c6583462ae061557d5fc4618deaf4b67ee | Nine paths individually reconciled, then aborted; preserve physical Map storage, callback refusals and Node agreement | 03aea1cbede84da46e84c78d599f287bc8585c1b | Nullish/optional/rest Map dependencies; emit-tuple-good and emit-helper-good disagree with Node in JavaScript |
| Dictionaries | deec3c933543c7b5e7ba0d9db0964265d03c494c | Freshly fetched newest tip is already an ancestor | 03aea1cbede84da46e84c78d599f287bc8585c1b | None introduced |

## Owner boundary

The earlier bd05075f rollback removed Map storage adapters and supporting entry
descriptors/tests. The latest owner tip adds nominal Map behavior on that base.
Its incoming native Map forEach calls closure code directly; current integration
uses viewCallableBoxedInvokeTypes. Converted key/value ownership also differs.
The merge was aborted rather than replace the current checked callback dispatch.
Restoring the general Union calling-convention refusal is not a resolution:
entry-live-mutation.a then refuses lowering and supported boxed callbacks regress.
The current unknown-producer refusals and original never-member helper refusal
remain. See docs/checked-views-blockers.md for the retained boundary.

## Commands and observations

All test output was redirected to files. Compressed exact logs are in
logs/continued, including mutant failure output and the restored never control.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Initial timings: Node 0.058s, Go 0.073s,
clang 0.492s, markdown ready 0.886s, submodules 194.800s. The initial build
encountered temporary owner-merge conflict markers. After aborting that merge,
setup succeeded: build 29.274s, cache ready 29.407s, done 29.438s; nproc=5,
cgroup cpu.max=400000 100000. Pinned stage3/api dependencies were installed with
npm ci --prefix stage3/api after the first backend run found missing Node types.

Pristine microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8 was
checked out in /tmp/checked-views-upstream. The lane2 original15 through original18
prepare.cjs adapters emitted complete declarations into /tmp/checked-views-arrayN-decls.
Shared lane4b declarations were prepared before the lane4 and lane7 manifests.
No source was copied from cohere. The isolated starting checkout references the
existing cohere submodule and Node dependencies through symlinks.

```sh
go test ./internal/lower ./internal/native ./internal/javascript -run 'Test.*View|TestLazyView|TestPrepareViewCallableRead|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional|TestMixedUnion|TestPhantomOverload' -count=1 -timeout 30m
go test ./internal/oracle -run 'Test.*View|Test.*Union|Test.*Tuple|Test.*Dictionary|Test.*Callable|Test.*Optional|Test.*Array|TestNativeAgreesWithNode/.*(view|union|tuple|dictionary|callable|optional|array)' -count=1 -timeout 30m
go test ./internal/oracle -run 'Test.*View|TestNativeAgreesWithNode/.*(view|union|tuple|dictionary|callable|optional|array)' -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCheckedViewRanked(15|16|17|18)(OriginalArrays|ArrayCounts)$' -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

The backend run passed: lower 7.601s, native 7.669s, JavaScript 0.845s.
Original-array testing supplied ADAMIC_ARRAY15_ORIGINAL_DECLS through
ADAMIC_ARRAY18_ORIGINAL_DECLS and passed all 152 probes and count rows in 361.786s.
Successful probes include leak checks; all probes compare Node and sanitized/
release native plus JavaScript output, with exact refusal diagnostics.

The broader oracle run failed in 834.349s on fourteen optional-suite subtests.
Twelve expect older scalar/literal diagnostic wording, while the checked union
read emits matches-no-member diagnostics. The nominal-slot and
nominal-subclass-slot fixtures expect behavior outside the current nominal
read/source-certificate frontier. The exact starting SHA reproduces the same
fourteen subtests in 18.675s. No refusal or test expectation was weakened.
The separate requested checked-view/fixture filter has its own exit result.

Global counts failed in 66.224s on 41 fixtures and wrote no counts table.
An isolated checkout at the exact starting SHA failed on the identical 41
fixtures in 83.139s. The full fixture list is logs/continued/count-failures.txt;
exact compiler, clang, runtime and readiness failures are in both compressed logs.
The new 152 array rows were retained, then independently remeasured successfully.

## Executed mutants

In the isolated checkout, each original15 through original17 runner removed the
native and JavaScript numeric field checks. Witnesses were function-wrong-pos,
function-parameters-wrong-pos and signature-type-parameters-wrong-pos respectively.
All six failed their oracle on execution. Original18 additionally removed native
and JavaScript numeric checks (json-diagnostics-wrong-element) and join element
checks (config-files-join-bad); all four failed on execution.

run-callable-boundary-mutants.py replaced native and JavaScript unknown-producer
refusals with ordinary invocation; TestViewCallableBoxingUnknownProducer caught
both executing mutants. Broadening the scalar exemption was caught by
TestCheckedViewCallableRetainsNeverWiderHelper, and the mutant execution control
also passed with ADAMIC_CALLABLE_NEVER_MUTANT=1. All thirteen source mutations
were restored before any integration commit.

Array filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	366.910s

## Mixed union first tip

The only merge conflict is one plan hunk: retain the array certification report
and append the independent lane4 reconciliation report. All compiler hunks
merged automatically. Unknown/untracked receiver fallback and the existing
callable-only concrete-read exemption remain intact.

The new complete-declaration conversion test passes in 0.009s. Scoped backend
tests pass: lower 5.681s, native 5.244s, JavaScript 1.430s. Complete original
inputs activate brand, primitive, array and intersection oracles. That run
finished in 1457.212s with exactly one obsolete compile-only helper expectation.
Its other cases pass, including 66 new primitive first-member/boolean-first and
member/literal-bypass mutant executions on native and JavaScript.

The original helper-unsupported.a boolean is now rejected at its helper field
read with exit 70 and the exact string | number declaration. The corrected test
also passes ordinary and viewed valid string/number objects through the same
helper, compares source Node, release/sanitized native and JavaScript, and checks
successful-run leaks. Two member-check mutants execute the original wrong value
and print ordinary/true at exit 0, proving the retained runtime refusal.
The corrected helper test passes in 1.753s. No compiler guard was changed by this
expectation update. The fresh filtered oracle includes TestPrimitiveOrdinaryProperty.

The broader lower refusal corpus additionally has two stale known-Union-callback
expectations (a union a function value takes; Array.from's undefined as a union).
The exact starting checkout reproduces both. No callback guard or expectation
was changed for these failures. Counts refresh still fails on the exact same 41
fixtures as the starting commit, in 50.651s. Callable unknown-producer and scalar
never-exemption executable mutants were rerun against the mixed-union source;
all three were caught.

User steering during validation supplied newer lane4 tip
37763565c317a80c3b362e475edab8b61ab0eafd. It was fetched and will be integrated
before moving to lane4b.

Fresh mixed filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	637.595s

## Mixed union newer tip

37763565 merged without conflicts. Primitive array reads decode producer storage
metadata before enforcing the original member contract. Dictionary selectors
retain the original reference-member obligation while permitting certified
scalar and nullish reads. No callable exemption or refusal was changed.

Scoped backend tests pass: lower 8.994s, native 30.726s, JavaScript 2.202s.
Primitive dictionary IR checks pass in 0.007s. Complete declaration-backed
primitive array and dictionary probes pass in 62.055s, including 28 executing
member-check and first-member mutant witnesses. The native runtime test catches
a forged producer storage certificate; dictionary selector tests also execute
member and null-erasure mutants. Counts refresh fails in 54.879s on the identical
41 baseline fixtures, with no additions or removals.

The first newer-tip filtered run finished in 643.345s with only six obsolete
primitive-array gap assertions. Homogeneous string/number producers now execute
checked reads and agree with Node on all three backend modes with no leaks;
heterogeneous literals retain exact producer NotYet diagnostics at line 5:46.
The corrected six-case test passes in 3.130s. No compiler guard was changed.

Fresh newer mixed filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	600.969s

## Object and primitive unions

8abb52a1 merged without conflicts. Unread tuple union storage remains admitted,
but demanded tuple members retain their named obligation. No callback guard
changed. Backend checks pass: lower 7.057s, native 8.669s, JavaScript 1.522s.
All complete original object/primitive probes pass in 244.963s, including comment
and package resolver fields, exact outer/nested refusals, successful-run leaks,
measured batch counts and executing member/nested/outer bypass mutants.
The isolated graph-size and message-size mutants execute and are caught by
AddressSanitizer heap-buffer-overflow; the tuple-member guard mutant is caught
by the exact named tuple union refusal. Every source mutation is restored.

Object/primitive filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	742.346s
Counts refresh: identical 41 baseline fixtures, 63.565s; new focused rows pass.

## Intersections

Five one-hunk conflict decisions, individually inspected:

- collections.go: keep primitive binding preparation, then retain the object
  intersection binding hook. Ordinary primitive binding and deferred-member-
  destructure exercise the separate behaviors.
- checked-views-plan.md: append both independent continuation reports.
- counts.md: retain every existing row and append the five deferred-member rows;
  duplicate fixture keys must have identical measurements.
- checked_views_brands_original_gaps_test.go: neither historical list remains
  blocked in the combined code. Explicitly supersede the compile-gap suite with
  BrandsOriginalPairs and IntersectionOriginalIdentifier runtime certificates.
- lane4/original/BLOCKERS.md: retain both histories, then mark all three former
  gaps closed by the combined runtime controls and mutants.

Backend and IR checks pass: lower 7.739s, native 29.514s, JavaScript 1.935s,
IR 0.053s. Complete original brand/intersection controls passed; the 479.740s run
reported only two obsolete lane4b JSDoc compile-frontier assertions. Those reduced
producers now retain exact runtime missing-field obligations: union parent flags,
optional parent escapedText. Independent parent-read omission mutants execute the
original Node value 80 in sanitized/release native and JavaScript and are caught
by their refusal pins. No compiler check is removed to update these expectations.

Six IR mutants execute and fail all three backend pins: deferred member selector
omission at direct, helper, callback and destructured reads; absorbed ancestor pos
omission; isolated Identifier helper omission. A seventh source mutant omitting
only the binding metadata hook survives the deferred destructuring fixture:
retained contract metadata still produces the correct refusal. That redundant
hook is not credited as an independently essential check in this integration.
The actual descendant-selection omission is killed independently. All source
mutations are restored. Counts refresh reports the identical 41 baseline failures
in 55.884s, with no additions or removals.

Intersections oracle: ok  	github.com/system-inc/adamic/internal/oracle	677.460s

Intersections frontiers: ok  	github.com/system-inc/adamic/internal/oracle	11.220s

## Callables newest tip

Fresh fetch advanced to 433e1ec09fdba7c83fe2a3d5153cf28962a23022.
Its single counts hunk retains all existing rows, appends all incoming callable
allocation rows before the existing predicate table, and keeps that table intact.
No compiler code changed. Backend checks pass: lower 6.161s, native 6.898s,
JavaScript 0.881s. Official upstream provenance verification passes for 49 ranked
fixtures, 363 later-ranked fixtures and 42 tagged aliases/discriminator fixtures.

Six isolated native/JavaScript arity, result and parameter mutants are caught on
the selected newest parameter-update, property-update, resolution-settings,
scanner-scan, lexical-start and writer-space families. Relevant guard omissions
execute valid code at exit 0, including wrong-arity output 9. Payload registration
omission catches resolution-settings via native sanitizer SEGV; it survives the
tagged parameter-update and property-update fixtures, whose independent checks
remain. The runner stops on its assumption that every selected family fails;
this partial result is retained, not counted as seven fully killed mutants.
All source mutations were restored. Global counts refresh reproduces exactly the
same 41 baseline failures in 53.734s; focused callable count checks run in the
required filtered oracle.

The first callable filtered run finished in 953.767s with five stale compile-only
untagged union expectations. The existing integrated untagged adapter now supports
string-from-node, for-update, variable-update, arrow-function and literal-type.
Their corrected controls agree with Node in sanitized/release native and
JavaScript, with successful-run leak checks, in 3.400s. Unbound-method and
destructuring refusals remain. This test update changes no compiler guard and
adds no whole-original receiver certification credit to reduced carriers.

Corrected callable filtered oracle: ok  	github.com/system-inc/adamic/internal/oracle	939.959s

## Tuple attempt left out

Fetched 5c54e8c6583462ae061557d5fc4618deaf4b67ee; nine conflicted paths
were inspected and reconciled individually. The attempt was aborted after the
remaining failures below. No tuple source or altered admission guard is landed.
The complete staged attempt is retained as logs/continued/tuples-attempt.patch.gz
for review; this is evidence of an unsuccessful attempt, not a certified patch.

- Plan: append both independent reports.
- ir/view_maps.go: retain globalOf physical-storage veto while keeping incoming
  nullish, union, tuple and callable descriptor comparisons; restore their local
  comparison helper definitions without restoring broad owner storage admission.
- javascript.go: retain tuple representation field routing and existing
  object/array untagged union selection.
- lower/object.go: route tuple reads through tuple-position metadata while
  retaining numeric/out-of-range fallback.
- lower/view_contracts.go, two hunks: retain untagged array element contracts,
  publish existing contracts before read support checks, and add tuple-union
  exclusions where object-field joining would erase position obligations.
- lower/view_lazy.go: retain certified untagged callable handling, intersection
  handling, the existing callable exemption and scoped unknown-field fallback;
  append supported tuple-position handling.
- lower/view_maps_tuples.go, modify/delete: inspect its seven functions and retain
  the tuple constructor, selected-read and optional/rest entry descriptors with
  complete proof helpers; current Map producer admission remains conservative.
- Map oracle: retain incoming nullish/recursive/nominal storage assertions.
- Count helpers: retain all array/intersection helpers and append tuple counts.

The first scoped backend/IR run passed: lower 8.675s, native 33.406s,
JavaScript 3.805s, IR 0.058s. Complete declarations were prepared from official
upstream with tuples/prepare.cjs and prepare_optional_tuple.cjs. The first tuple
original run failed in 47.726s: existing object/primitive tuple-member refusal
intercepted incoming tuple adapters, and inherited owner tests referred to twelve
missing nullish Map fixtures deleted by the earlier owner rollback.

A narrow follow-up admitted only supported tuple-scalar and tuple alternatives
through that earlier tuple-member refusal. The twelve missing fixtures were
restored from the incoming Adamic lane solely to measure the dependency. The
storage test then failed all twelve at Map admission, including:
entry-number-null.a:6:10: stage 0 can't lower a Map of number | null yet;
entry-mixed-both.a:7:10: stage 0 can't lower a Map of string | number | null | undefined yet.
Current Map physical storage/provenance and callback guards were retained.

The correctly configured follow-up ran:

```sh
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/checked-views-tuple-decls go test ./internal/oracle -run '^TestCheckedViewTupleOptionalContract$|^TestCheckedViewTupleRestContract$|^TestCheckedViewTupleOriginalOutSignature$' -count=1 -v -timeout 10m
```

It failed in 18.122s. Optional direct tuple reads and rest direct tuple reads pass,
including the wrong-tail runtime refusal. optional-map-absent,
optional-map-undefined, optional-map-present, rest-map-zero, rest-map-one,
rest-map-many and rest-map-wrong-contract refuse field values with unsupported
Map key/value certificate contract. Original emit-tuple-good and emit-helper-good
compile after the narrow routing fix but JavaScript exits 70 with:
`adamic: panic: a union value does not match its narrowed type`.
The independent source Node controls succeed; wrong-position/helper wrong
fixtures also disagree in their refusal diagnostics. Restoring broad owner
adapters would require resolving the same deferred callable/ownership boundary.
Neither removing the storage proof nor accepting the JavaScript disagreement
preserves the required behavior, so the item is left out.

The required full filtered oracle was started on the initial tuple attempt and
cancelled after the reproducible dependency and Node failures; its incomplete
log is retained and is not claimed green. Tuple-specific opt-in source mutants
were not run or credited on this unlanded attempt. The last landed compiler
state remains 03aea1cbede84da46e84c78d599f287bc8585c1b with its completed green
939.959s filtered oracle. No test repeat is needed for this documentation-only
checkpoint. After another fetch, dictionary deec3c933543c7b5e7ba0d9db0964265d03c494c
is still the newest tip and an ancestor, so no dictionary merge is necessary.


## October 8 non-null fixture ruling

Moved 436 assertion-bearing fixtures from .a to .ts, preserving every source byte and expected output.
Commit inventory: internal/oracle/views_non_null_fixtures.json names all 54 readiness/non-null and 382 checked-view fixtures; the commit message names every old and new path.
Validation: scoped backends, registered non-null source oracles, readiness mutants and declaration-bound count/copy checks pass; broader views runtime results below.
Mutants: all sixteen readiness-output mutants and four additional readiness/weak/lazy/store mutants execute and are caught by their existing independent pins.
Limits: global counts retain exactly 41 baseline failures; four additional original-pair frontier assertions reproduce at exact ffe428ab; the checked non-null area remains an unmerged dependency.

The October 8 00:27 ruling keeps non-null assertions refused in Adamic .a and
inserts checks in TypeScript .ts. The contradictory readiness registrations in
new-expression evidence aae366090e25128bd8d152b87f5fd569a40dd58f, at
stage3/notyet-new-expression/VIEWS-PROBE.md, motivated this fixture migration.
The area tip c41c0e062e99da37820f822968d4df1b48cdaee7 is read for context, not
merged. No compiler file changes in this unit. Only nine static fixture path
strings change in oracle_test.go; its harness remains intact.

The explicit fixture inventory resolves dynamic view/readiness paths. The
ordinary loader is unchanged, so independent .a refusal fixtures in the area
retain their own source contract. Bound original copies inherit the extension
of their actual source fixture; non-migrated original frontiers retain .a.
Callable and ranked-array count discovery includes both extensions and retains
all existing compile-refusal exclusions. Provenance verifiers include both
extensions. Live Node-control paths and the optional witness mutant runner
follow the relocated fixtures. Historical reports remain historical.

A byte comparison against ffe428ab passes for every relocated fixture. A fresh
TypeScript AST scan finds zero non-null expression nodes in remaining .a files
under stage3/interface-downcasts and the non_null oracle family. No fixture
body, stdout, exit, diagnostic or mutant pin is rewritten. counts.md differs
only in extension keys; every recorded numeric observation is identical.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Timing lines: Node 0.027s, Go 0.027s,
submodules 0.071s, markdown dependencies 0.093s, clang 0.186s, build 43.149s,
cache 43.297s, total 43.326s. nproc=5; cpu.max=400000 100000.

Commands, with output redirected to the corresponding views-nonnull log:

```sh
go test ./internal/lower ./internal/native ./internal/javascript -run 'Test.*View|TestLazyView|TestPrepareViewCallableRead|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional|TestMixedUnion|TestPhantomOverload|TestPrimitive' -count=1 -timeout 30m
go test ./internal/oracle -run '^TestReadinessMutants$|^TestUninitializedIsNotNullishMutant$|^TestNonNullWeakFreedNamesExpression$|^TestLazyInitializerIsNotEagerMutant$|^TestDeinitializationIsNotOrdinaryStoreMutant$|TestNativeAgreesWithNode/.*/non_null' -count=1 -v -timeout 10m
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/non_null' -count=1 -v -timeout 10m
go test ./internal/oracle -run 'Test.*View|^TestPrimitiveOrdinaryProperty$|TestNativeAgreesWithNode/.*(view|union|tuple|dictionary|callable|optional|array)' -count=1 -v -timeout 30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout 10m -args -update-counts
source /tmp/checked-views-test-env.sh
go test ./internal/oracle -run '^TestCheckedViewRanked(15|16|17|18)(OriginalArrays|ArrayCounts)$|^TestCheckedViewObjectPrimitiveOriginal' -count=1 -v -timeout 30m
go test ./internal/oracle -run '^TestCheckedViewRanked(15|16|17|18)ArrayCounts$|^TestCheckedViewObjectPrimitiveOriginalPairs$/source-file-uninitialized$' -count=1 -v -timeout 10m
# In the isolated exact ffe428ab checkout:
go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveOriginalPairs$/(bindable-expression-frontier|jsdoc-parent-probe|bindable-static-left-probe|bindable-left-probe)$' -count=1 -v -timeout 10m
node stage3/interface-downcasts/lane5/verify-ranked-callable-fixtures.cjs /tmp/checked-views-upstream 1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49
node stage3/interface-downcasts/lane5/verify-optional-aggregate-fixtures.cjs /tmp/checked-views-upstream
```

Backend checks pass: lower 19.914s, native 18.944s, JavaScript 3.400s.
Readiness mutants pass in 7.237s. The initial selector did not execute the nested
source-fixture subtests; the explicit slash-aware source selector then passes
all registered non-null controls in 10.290s. This records actual coverage rather
than crediting the earlier empty source selection. Sixteen readiness mutants
are scanner-var-capture, deinitialization-values, deinitialization-entries,
deinitialization-assign-source, deinitialization-field-alias,
deinitialization-field-method, keep-slot-proven-after-deinitialization,
deinitialization-through-capture, deinitialization-exception-path, drop-check,
erase-without-proof, initialize-to-zero, miss-captured-read,
miss-exception-path, lazy-read and weak-generic-message. Each is caught by the
existing independent stdout/exit/diagnostic pin on native execution. Additional
ordinary-nullish initializer and eager-lazy-initializer mutants fail Node
comparisons; freed-Weak diagnostic mutation fails its expression-specific native
pin; ordinary undefined-store mutation agrees with source Node at exit 0 and
fails the retained readiness pin. These tests mutate IR, not fixture source.

The declaration-bound run finishes in 607.499s. All 152 original array controls
pass: original15 15, original16 42, original17 59, original18 36. Its only four
failures are unchanged, non-migrated bindable-expression-frontier,
jsdoc-parent-probe, bindable-static-left-probe and bindable-left-probe: each expects
a compile refusal for union intersection that the integrated compiler now admits.
An isolated checkout at exact ffe428ab reproduces all four in 14.945s. No unrelated
frontier pin is changed. Final source-specific copy/count checks pass in 278.261s,
including source-file-uninitialized and all four original array count groups.
Provenance verification passes for 28 ranked and 14 optional aggregate fixtures,
with complete original declarations and source read spans intact.

The final count refresh finishes in 83.266s with exactly the recorded 41 baseline
failures, no additions or removals and no failures on migrated .ts fixtures.
It writes no complete replacement table because those failures remain. Migrated
keys were updated directly without changing numeric values and independently
remeasured by the focused controls and count groups. An earlier refresh on the
partial count-discovery plumbing finishes in 65.466s with the same baseline set;
only the final run is credited for complete renamed-fixture discovery.

Two early path-helper mistakes were caught locally: recursive absolute-path
resolution caused stack overflow, then a relative repository base missed dynamic
paths. Both were corrected before the passing readiness run. An early broader
oracle run was cancelled when superseded; its incomplete result is not credited.
The source-specific temporary-copy adjustment and final count discovery were
verified separately on the final test plumbing. Exact compressed logs are in
logs/continued. No full package, full repository gate or unlanded worker merge
was run. This completed unit receives one push under the standing rule.

Broader views oracle: passes in 1056.031s, with 1001 native observations and 3961 Node observations missed from cache. The final source-specific copy and complete count-discovery plumbing are independently covered by the 278.261s and 83.266s runs above.


## October 8 dictionary groups 9 through 11

Merged 48d166cf7ba1954de46ca2a4b30d72f6872153a9 without conflicts. Actual-key enumeration does not demand values; values and entries check selected values, preserve transitive obligations, and keep reached array-entry tuple consumers refused. Private unpublished enumeration arrays alone use DictionaryProduction. Sixty-eight executable controls are registered with measured allocation rows; six demanded tuple refusals and eighteen uncredited frontier probes are excluded from those rows.

Setup passed in 40.219s; nproc is 5. Scoped IR/lower/native/JavaScript checks passed in 0.012/7.671/69.682/0.953s. Dictionary oracle, array writes, wider-helper callable refusal, readiness and record mutants passed in 138.530s. Nine new dictionary mutants cover eager key demands, omitted element checks, wrong entry shape, omitted transitive checks on both backends, and omitted derived-origin IR roots. All were killed. The 68 regular Node, sanitized native, release and JavaScript controls and focused counts passed in 18.525s. An initial registration mistakenly marked every positive control as an expected runtime refusal; the harness caught it, and metadata now marks only demanded bad-payload controls checked. No compiler guard was changed to repair it.

The required global counts refresh reproduces exactly the inherited 41 failures, with no additions; focused measurement adds 68 rows without changing old rows. Historical incoming GROUP11-broad-gate.log retains three trailing-whitespace lines as raw evidence. No full package or full gate was run. Logs are archived under logs/continued/views-dictionaries-*.log.gz. Lane 1 0f55d664 is next; the updated tuple tip has not been supplied. This merge will share one push with the completed current integration unit.


## October 8 owner merge 0f55d664

Merged 0f55d66409d9f06e65d628f3c4d02b26a327bc99. One conflict hunk in checked_views_map_certificates_test.go combines the approved void-brand Node control with integration's fixture resolution through interfaceFixture. The former phantom-refused.a certificate frontier is now certified by the owner's descriptors; required non-void primitive brands still refuse, proved by entry-nominal-gap-required-brand-read.a and -unread.a. No compiler code conflicted. Dictionary derived-origin roots and private production append flags survive the automatic merges.

Boxed callable dispatch remains intact. All 47 callback refusal sites are pinned; entry-live-mutation.a refuses at 5:69 while Node prints number, 7, 0. Owner key and value snapshots are owned across normal and throwing calls. Removing key snapshot retention in an isolated checkout causes real AddressSanitizer heap-use-after-free in map_foreach_keys.a and map_foreach_named_keys.a; restored controls pass. Initial isolated submodule setup failure is not mutant evidence.

The new entry-nominal-array.a had two non-null assertions. Both are rewritten using an explicit undefined guard on the first element, preserving stdout. Its own count row is measured. Scoped IR/lower/native/JavaScript checks pass in 0.019/7.741/79.260/1.155s; filtered owner, dictionary and readiness oracle controls and mutants pass in 456.330s. The rewritten fixture and restored callback controls pass in 3.710s.

The global counts refresh initially reports 41 failures: 40 inherited failures, the inherited nominal-subclass-slot failure is resolved, and user_iterators.a gains a clang distinct-pointer comparison failure. Casting the optional nominal receiver to const void * preserves exactly the null sentinel comparison and fixes that compilation failure. The source Node agreement control passes after the fix. Final count refresh and compiler verification will run on the tuple combination. No full package or full gate was run; evidence is under logs/continued/views-owner-*.log.gz.


## October 8 priority owner, tuple and dictionary integration

Dictionary 48d166cf7ba1954de46ca2a4b30d72f6872153a9 was re-fetched and is already merged at 5ab46e9fc7281824b10136995dc6e10819793177. Owner 0f55d66409d9f06e65d628f3c4d02b26a327bc99 is merged at 2f7429ed199ddfa82d19ce6da34d2effc2b55078. Tuple 0804dec54050b5ac5a360b1f16cee66aa28eb352 is the second parent of this final merge. This batch receives one push, only to codex/views-integration.

The tuple merge has 27 individually resolved hunks in 13 files. No bulk strategy or whole-file selection was used. Decisions and proving fixtures follow.

| Path and hunk | Functional decision and evidence |
| --- | --- |
| plan 1 | Keep dictionary groups 9 through 11 and the tuple handoff appendices. |
| ir.go 1 | Keep Narrow.Tuple plus Undefined and its diagnostic; root-index-tuple.a and entry-convert-nan-undefined.a. |
| view_maps.go 1 | Keep optional-source presence containment; optional nominal producer/schema controls. |
| view_maps.go 2 | Use FixedTuple arity/rest domains while retaining recursive physical storage checks; optional-map-present.a and rest-map-wrong-contract.a. |
| javascript.go 1 | Combine nominal receiver checks with tuple field representation; emit-tuple-good.a and nominal widening read mutants. |
| javascript.go 2 | Keep undefined-only storage proof ahead of tuple narrowing; Map undefined storage mutants. |
| JS view_arrays.go 1 | Keep owner Map, packed-boolean and primitive union cases; permit arrays as objects only under a tuple witness. |
| JS view_arrays.go 2 | Gate primitive extraction on both primitive membership and absence of TupleUnion; keep nominal reads and add tuple selection. root-index-tuple.a and root-index-scalar.a. |
| JS view_arrays.go 3 | Keep tuple and nominal consumer guards as separate branches; signature-foreach-tuple.a and nominal array read mutants. |
| lower object.go 1 | Keep owner primitive/nominal array certificates and add only supported tuple-scalar union admission. |
| lower object.go 2 | Propagate checked tuple read errors; watch-event-narrowed-missing.a. |
| lower view_arrays.go 1 | Keep nominal contract override and validate TupleUnion separately. |
| lower view_arrays.go 2 | Keep concrete type identity and add TupleUnion metadata. |
| lower view_maps.go 1 | Keep owner phantom, nominal, nullable tuple and callable descriptor dispatch; do not discard source descriptor proofs. |
| view_maps_tuples.go 1 | Generalize the single constructor with recursive child builders, retaining Map descriptor and callable-descendant vetoes. |
| view_maps_tuples.go 2 | Accept only required-prefix, optional and terminal-rest arity plans; module-specifiers-wrong-arity.a. |
| view_maps_tuples.go 3 | Recognize explicit tuple identity, including empty tuples; tuple layout probes. |
| view_maps_tuples.go 4 | Keep checked child fields and encode optional/rest domains; rest-view-wrong-tail.a and optional absent mutant. |
| view_maps_tuples.go 5 | Keep per-position witnesses and add safe missing numeric/reference reads; watch-event-narrowed-missing.a and rest-view-zero.a. |
| native emit_slots.go 1 | Exclude tuple plans while retaining primitive membership gating, including typeof; root-index-tuple.a. |
| native view_arrays.go 1 | Exclude tuple plans while retaining primitive membership gating at evaluation. |
| native view_arrays.go 2 | Keep private read owners and producer scans; route tuple plans through owned normalization first. |
| native view_arrays.go 3 | Keep owner argument and tuple callback transfer/release; transfer-store.a and transfer-throw.a, leak and double-release mutants. |
| Map oracle 1 | Keep owner approved-brand controls; retain the tuple-entry incompatible-storage refusal in its own test. tuple-entry-read.a refuses at node.value; tuple-entry-unread.a agrees with Node. |
| Map oracle 2 | Keep owner union storage and all six Map conversion controls. Four formerly unsupported owner-storage forms now agree with Node through complete certificates; optional-object, optional-array, nested-array and key gap fixtures. |
| Map oracle 3 | Require Node agreement and leak checks for certified optional/rest tuple storage; both legacy gap fixtures. |
| Map oracle 4 | Keep nominal method/recursive gaps, required brand refusals and JSON array controls. |
| counts 1 | Keep both measured dictionary/owner rows and tuple rows; refresh only owned measurements. |

The automatic tuple admission exposed the dictionary lane's six array-entry consumers. General tuple certificates do not discharge their producer-specific ownership boundary. Keep a scoped tuple-demand refusal using the existing allocation flow and DictionaryEntryOrigins. Its mutant now uses an otherwise supported tuple descriptor: removing derived roots admits the demand, so the actual boundary is proven able to fail. An independent tuple allocation stays admitted. Scalar enumeration keys remain unaffected. All six entries-{fixed,producer}-array-{good,wrong,empty}.a controls retain their named [element] unsupported tuple refusals.

Tuple normalization uses the owner's array reader signature with a NULL private owner because normalization itself owns its output. A new private tuple extraction selector admits primitive or object storage solely for the subsequent tuple union witness. The existing primitive selector is unchanged; ordinary records still fail tuple shape selection. The explicit shape-removal mutant admits an ordinary record and fails the native oracle.

Three incoming assertion-bearing fixtures are rewritten without !: rest-map-zero.a, rest-map-one.a and rest-map-many.a. Each checks undefined explicitly and throws an Error before using a missing tuple; expected stdout is unchanged. Together with entry-nominal-array.a in the owner merge, these are the four named rewrites in this batch. TypeScript AST audit reports zero assertion-bearing .a files under interface-downcasts. The preceding 436-file TypeScript migration is retained.

Final scoped compiler checks pass: IR 0.011s, lower 5.215s, native 37.571s, JavaScript 1.452s. Final dictionary boundary checks pass in 0.008s; eager-key mutants pass in 0.268s. The complete original-declaration tuple corpus, 47 exact callback refusal pins, nominal guard mutants, dictionary values/entries, array reference writes and wider-helper never refusal pass in 177.293s. Array/union regression functions and dictionary enumeration mutants pass in 54.775s. All 68 registered dictionary controls agree across Node/native/release/JavaScript in 16.443s. Owned 153 count rows pass and refresh in 130.498s. Full original declarations remain pinned to upstream 050880ce59e30b356b686bd3144efe24f875ebc8, including the unreduced private module worker declaration.

All twelve incoming tuple mutation modes were rerun and killed by their recorded actual oracle. Presence and scalar selection, optional/rest absence, erased rest schemas, callback/module arity, narrowing and source certificate mutations fail semantic pins. Skipped transfer release leaks 24 bytes; double release triggers AddressSanitizer use-after-free. The additional tuple shape mutation fails by admitting an ordinary record. Exact environments and results are in logs/continued/views-tuples-current-mutants.json. Nine dictionary mutants, the derived-origin mutant and the owner's key snapshot ownership mutant are recorded above; none is credited for a compiler failure.

Final required global counts refresh finishes in 81.795s with 39 inherited failures and no additions relative to the 41-failure starting set. nominal-subclass-slot.a and graph_regions_entries.a no longer fail. The optional nominal pointer cast also resolves the transient user_iterators.a clang failure. The global table cannot be fully regenerated while those unrelated failures remain; owned rows are independently measured and committed. An intentionally filtered count comparison also failed its whole-table comparison after its user_iterators fixture passed; it is not credited as a counts gate pass.

Earlier tuple attempts caught a runtime reader argument mismatch, primitive interception of a valid tuple, stale owner-storage expectations, dictionary boundary bypass and unsupported literal throws. They were corrected before final green checks. An accidentally broad ordinary-oracle selector was promptly cancelled; its incomplete log is not credited. No whole package or complete repository gate passed or was claimed. Incoming review/*.patch files preserve their historical patch-context whitespace; owned edits pass diff checking. Broader validation remains with the shared fast gate.

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

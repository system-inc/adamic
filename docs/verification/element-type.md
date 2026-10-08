# elementType lowering results

Base: origin/area/compiler d577dd0d7c311bfe5ae972dea390474be3a0ee0d.
Merged census replay 9a1f14c5 and checked non-null c41c0e06 (merge 981520f8).
First storage group: f544f4a2. Reference storage is the commit containing this report.

Counts below are the table's root-site counts, not a claim of 79 individually replayed sites. One supplied example per kind was replayed. The source snapshot was rebuilt using the replay branch's stage3/apply.sh; all 81 manifest hashes matched. Replays use checker-rejected entry-root programs. A missing signature can mean an earlier stop or a generic declaration without its call-site instantiation; it does not prove whole-site lowering.

| Kind | Roots | Result after non-null merge |
| --- | ---: | --- |
| T | 42 | Skipped unresolved generic replay. core:148:27 still array of T. Concrete first/copy/append calls pass both backends. |
| U | 6 | Skipped unresolved generic replay. core:325:13 still array of U; earlier never[] at 323:18. |
| boolean or undefined | 6 | Lowered. binder:1933:17 no longer stops; next binder:1956:21 BinaryExpression as statement. |
| IncrementalBuildInfoRoot | 4 | Skipped earlier stop: builder:1408:25 value of Path. Numeric/tuple union storage is covered by the union fixture. |
| ResolvedConfigFileName | 4 | Lowered scalar brand storage. tsbuildPublic:738:29 now reading buildOrder; earlier unresolved project values remain. |
| NonNullable<T> | 3 | Skipped unresolved generic replay: core:739:31 still array of NonNullable<T>. Concrete string and number instantiations pass. |
| undefined | 3 | Lowered. binder:1940:40 no longer stops; next binder:1956:21 BinaryExpression as statement. |
| Child | 2 | Skipped unresolved generic replay: emitter:4734:86 still array of Child. |
| V | 2 | Skipped unresolved generic replay: core:107:25 still array of V. |
| Extension array union | 1 | Lowered concrete fixture. Replay selects generic flatten and stops at never[] / T or T[] at core:376/378; original instantiated signature is not reproduced. |
| Build-info array union | 1 | Lowered. builder:2407:5 array union stop gone; next builder:2407:32 value of IncrementalMultiFileEmitBuildInfoFileInfo. |
| string array union | 1 | Lowered concrete fixture. Replay selects generic flatten, same limitation as Extension. |
| CanonicalKey | 1 | Lowered scalar brand storage. commandLineParser:4131:47 stop gone; later 4140:25 value of CanonicalKey. |
| TState | 1 | Skipped unresolved generic replay: factory/utilities:1396:9 still array of TState. Earlier generic function as value at 1394:34. |
| object | 1 | Lowered. checker:15326:33 stop gone; next 15327:16 BinaryExpression with a value and a value. |
| unknown | 1 | Refused for a ruling. core:684:12 reproduced. Accepting existing Union storage collapses null to undefined. |

## Validation

All test output was redirected to logs under /tmp/element-type-*. No full repository gate was run.

- cloud/setup.sh completed: Node .039s; Go .059s; clang .438s; markdown ready 1.704s; submodules 28.916s; Go build 517.063s; total 517.492s. nproc: 5.
- go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/element_type_|TestElementType' -count=1: pass, 2.001s after merge. Four positive fixtures agree with Node in JavaScript and sanitized native; the unknown fixture is an expected lowering refusal, with Node behavior pinned separately.
- go test -overlay=/tmp/element-type-group-one-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/element_type_|TestElementType' -count=1: staged first group passes, 1.966s after merge.
- go test ./internal/lower -count=1: pass, 34.849s after merge. This package-wide run exceeded the requested focused-test scope; no full gate followed.
- go test ./internal/native -run 'TestMaybeNumbersPackIntoOneDouble|TestCEndsInNewline|TestLoopArrayHoldC' -count=1: pass after merge. An earlier native package run also passed (457.756s).
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: pass, 74.263s after merge.
- Replay: go run ./stage3/census/latent/replay -project /tmp/element-type-census-input/src/tsc/tsc.ts -where <example> -kind NotYet -reason <table signature>, for all 16 rows, before and after non-null merge.

## Mutants and rulings

Registered native C mutants all build and exit normally under sanitizers; only Node stdout comparison catches them: index zero changed to one; MaybeBoolean unpack loses presence; generic sort substituted for optional-boolean sort; pointer identity substituted for union numeric equality; generic sort substituted for union sort. The sort mutants incorrectly call the comparator for undefined.

Six local Go overlays were run after the merge. Removing scalar intersection, undefined, object, array-union, or object-intersection handling makes its positive fixture fail lowering. Accepting unknown as Union makes native print `undefined undefined` instead of Node's `object undefined`; native exits normally. This proves the unknown refusal remains necessary with current storage. A distinct null tag and all affected operations need a separate representation change before accepting unknown[].

Removing the minimal NonPrimitive hook in cycleFinder.reaches makes both object-array self-cycle and captured-closure cycle refusal tests fail (both are wrongly accepted). The hook preserves the no-cycle doctrine for newly accepted object arrays; it is not a relaxation of existing checks.

New separate runtime helpers for runtime-owner review: internal/native/runtime/array_maybe_boolean.c and .h; internal/native/runtime/array_union.c and .h. They sort only present elements and write undefined at the end, retain reference elements, preserve comparator exceptions, and tolerate comparator changes to receiver length. Existing runtime files were not edited.

Unresolved generic declarations were not erased or assigned guessed storage. The generic fixture demonstrates ordinary concrete specialization, which already works; those census stops remain pending a census/instantiation decision. No other worker's lowering function was changed except the named minimal cycleFinder.reaches safety hook.

## Runtime review follow-up

The optional-boolean and union variants now live beside the number variant in internal/native/runtime/sort_undefined.c, with declarations together in adamic.h. The four separate array_maybe_boolean / array_union .c and .h files and emitted includes were removed. This relocation changes no helper semantics.

The separately registered element_type_sort_mutation.a fixture covers a comparator that pushes and a comparator that pops all original elements, for optional numbers, optional booleans and owning unions. Node, JavaScript and sanitized native agree on all six cases, including preserving appended tails and regrowing the original range. Two new native mutants discard the appended boolean tail or preserve the comparator's shortened boolean length; both build and exit normally with empty sanitizer stderr, and fail only Node stdout comparison.

Follow-up commands (all output redirected to /tmp/element-type-consolidated-*.log):

- go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/element_type_|TestElementType' -count=1 -v: pass, 20.872s.
- go test ./internal/oracle -run 'TestElementTypeSortMutation' -count=1 -v: both new mutants caught, pass, 1.357s.
- go test ./internal/native -run 'TestCEndsInNewline|TestMaybeNumbersPackIntoOneDouble|TestLoopArrayHoldC' -count=1: pass, .396s.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: refreshed for the new fixture.

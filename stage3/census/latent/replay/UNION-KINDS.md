# enumNeverValue object-kind checks

Built represented object-kind checks for narrowed record, array, and map unions.
Earlier commits: generic values b96e6356, receiver ruling 64246ed0, replay 1e8413ac; current main merged at c4e929be.
Node differential fixtures pass; final touched-package results are recorded below.
Nine union mutants failed their fixtures; six generic mutants are recorded in ENUM-VALUES.md.
All five union reason sites disappear on replay; receivers, unproven generic census contexts, and fs remain uncovered.

The territory clarification permits the small IR/backend/freshness changes below.
The only lowering function changed for this step is enumNeverValue. Worker branch
histories were fetched and checked before editing shared files. No runtime C
file changed and no separate C runtime helper was added. The new native helper
uses existing heap tags. No code was copied from cohere.

The new Boolean IR KindIs evaluates its union operand once and proves its native
representation before narrowing. Arrays, maps (including Set's map storage),
and record objects have different heap layouts although JavaScript typeof calls
all of them object. The JavaScript emitter checks matching built-in categories;
native checks the existing heap tag with a null guard. Library map and typed-array
iterators share the Object IR representation and therefore belong to that category.
Freshness evaluates the input and gives the predicate a scalar result; existing
reflection-based flow walkers already traverse it. Unsupported object-like
representations retain the existing refusal. This does not add structural shape
checks between two object types with the same representation.

## Node fixtures and mutants

union_object_kind.a covers record/array/map transitions, MapIterator's object
representation, runtime-built strings, and an absent optional record. Node,
JavaScript, release native, and ASan/UBSan native agree on exit 0 and:

```text
Adamic
2 2
1
3
held object
true
absent
```

The existing reland_refused/narrowed_union_object_tag.a now joins the differential
fixtures and prints 1. Its prior refusal test now asserts lowering success.
union_object_kind_stale.a uses a captured writer to replace an object with an
array after checker narrowing. Node prints undefined, while both Adamic backends
raise the specified checked narrowing panic (exit 70); sanitizers stay clean.
All are registered from union_object_kind_test.go, without oracle_test.go edits.

Temporary mutant runners restore each file in finally. Final commands were:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-union-final-mutants.py > /tmp/notyet-union-final-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-union-null-mutant.py > /tmp/notyet-union-null-mutant.log 2>&1
```

Each mutant runs go test ./internal/oracle -run
'TestNativeAgreesWithNode/internal/oracle/testdata/union_object_kind' -count=1
-timeout 10m into its own /tmp/notyet-union-mutant-<name>.log. All nine
semantic mutants fail with exit 1, without build failures:

| Mutant | What catches it |
| --- | --- |
| Lowering replaces KindIs with typeof object | Stale fixture: wrong-layout read; ASan heap-buffer-overflow; release/JS miss the checked panic |
| Native array tag becomes object | Normal array read falsely panics |
| Native map tag becomes array | Normal Map read falsely panics |
| Native object tag becomes array | Record read falsely panics; stale fixture also exposes a bad cast |
| JavaScript array predicate becomes false | Normal array read falsely panics |
| JavaScript map predicate becomes false | Normal Map read falsely panics |
| JavaScript object predicate admits arrays | Stale fixture misses its checked panic |
| Native excludes map iterator from Object | Iterator next read falsely panics |
| Native null guard becomes always true | Optional record fixture: UBSan member access within null pointer |

The initial null mutant inverted the guard; it also failed, but the final version
removes its protection and specifically exposes the null dereference.

## Validation

All tests write to logs. The focused final fixture command passed (oracle 0.568s):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/union_object_kind' -count=1 -timeout 10m > /tmp/notyet-union-final-extra.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/notyet-union-final-counts.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/oracle -count=1 -timeout 30m > /tmp/notyet-final-packages.log 2>&1
```

Counts passes in 24.375s and adds exactly three fixture rows; existing rows do
not move. The earlier uncached whole-package run passed IR (28.830s), native
(419.252s), freshness (84.630s), and found no JavaScript package tests. It exposed
an obsolete generic-value refusal test (now made uncontextualized) and counts
loaded before the new rows were recorded. The corrected lower package separately
passed in 48.381s. The final package rerun follows the latest main merge.
No full gate was run.

## Replay coverage

Table pin: 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7. Filtering its
stage3/notyet-table/roots/raw.csv on the exact narrowed-union reason gives five
rows, five unique (kind, where, reason, text) signatures. All five original
reason signatures are absent in after replays. This is reason-site coverage,
not a claim that the whole compiler now lowers.

Replays use the pinned adapted project and the exact original NotYet reason:

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-this-adapted/src/tsc/tsc.ts -where /tmp/notyet-this-adapted/src/compiler/checker.ts:37210:16 -kind NotYet -reason 'a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables' > /tmp/notyet-union-site-0.log 2>&1
```

The other four use their own where and log. Before, the two requested examples
reproduce the exact refusal (exit 0). After, all five exact-signature searches
exit 1, as expected when the signature disappears:

| Site | Result of selected unit |
| --- | --- |
| checker.ts:37210:16 | Advances to core.ts:622:31, a NonNullExpression; retains an earlier dependency stop at core.ts:1750:25, a value of type unknown |
| tsbuildPublic.ts:265:50 | No lowering findings; selected unit lowers |
| tsbuildPublic.ts:1212:47 | Next statement 1213:15: a value of type ResolvedConfigFileName; further Path/map-key stops remain |
| tsbuildPublic.ts:2321:70 | Union reason gone; later 2326:9: an array of ResolvedConfigFileName; earlier rest-parameter dependency at 2287:109 remains |
| tsbuildPublic.ts:2322:93 | Same selected unit and remaining stops as 2321:70 |

After logs: /tmp/notyet-union-site-{0,1,2,3,4}.log. Before logs:
/tmp/notyet-union-{first,second}-before.log. These are checker-rejected census
entry-root observations; they do not establish that the original project is
checker-clean. Verified removed reason sites: 5/5.

## Remaining reasons

Explicit this remains the documented design refusal, 0/46; THIS-OUTSIDE.md names
the receiver/initialization ruling needed. Conservative concrete generic values
are implemented, but both requested examples stop earlier on T, so verified
census coverage is 0/17 (ENUM-VALUES.md). Neither is a territory blocker.

The single fs root is not a special enumNeverValue rule: the fallback reports
reading a non-local tracePath. On this base the exact tracing.ts:95:19 replay
selects tracingEnabled at 38:1, rather than the raw CSV startTracing unit at
59:5, and stops at 41:9 on a value of type any. Searches of production lowering
find no node:fs/openSync implementation. optional_node_host.go has only the
unhandled-argument stub, including on current main. No sound, reproducible fs
host context is present to reduce and fix; broadly accepting an unbound
identifier would violate the ordinary local-binding proof. Coverage is 0/1.
This needs the census's host-enabled context or its missing node-host bridge,
then the replay/fixture/mutant sequence; this report does not claim an fs fix.

Both requested union sites were replayed again after merging current main
749a69ad at c4e929be. /tmp/notyet-union-final-checker.log (6.011s) retains
unknown and then NonNullExpression; /tmp/notyet-union-final-tsbuild.log
(6.489s) has no findings. Both exit 1 because the requested old signature is absent.

Final touched-package rerun exits 0: IR 13.585s, lowering 43.540s, native
367.270s, freshness 67.399s, oracle 205.677s. JavaScript has no standalone
package tests; its generated backend is exercised by the oracle differential
fixtures. git diff --check is clean. The final rerun uses normal artifact
caching; the final focused fixture and semantic mutants explicitly bypass it.

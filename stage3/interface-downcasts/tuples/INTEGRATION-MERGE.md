Built: merged integration ffe428ab into codex/views-tuples, resolving all nine conflicting paths individually and preserving the single tuple path.
Commits: merge parents 5c54e8c6583462ae061557d5fc4618deaf4b67ee and ffe428ab26eefb73154adbf570ad99e9c1b4f872; final merge SHA supplied in the handoff.
Checks: tuple suite PASS 154.373s, lane 2/4 regressions PASS 40.130s, Map/callable boundaries PASS 51.713s; 84 owned counts independently PASS 81.149s with no value changes.
Mutants: 12 modes rerun, all caught by their actual oracle; transfer skip leaks 24 bytes, double release triggers AddressSanitizer use-after-free.
Not covered: full gate/global counts, deferred owner Map storage adapters and the existing direct array-to-tuple cast frontier; primary census remains 10 pairs / 15 reads.

## Hunk resolutions

| Conflicted path | Resolution |
| --- | --- |
| docs/checked-views-plan.md | Retained both independent append-only reports and added the integration dispatch hook below. |
| internal/ir/view_maps.go | Kept integration's physical Of equality veto at every recursive comparison, restored tuple arity/position/rest checks and their helper definitions. No Map read-storage conversion was restored. |
| internal/javascript/javascript.go | Combined tuple field representation routing with integration's object/array union selection; retained boxed callable dispatch. |
| internal/lower/object.go | Retained integration's primitive/dictionary handling and routed supported tuple element reads through the existing checked tuple read. |
| internal/lower/view_contracts.go, two hunks | Kept joined array element contracts and recursive untagged support publication; excluded IR-tagged tuple alternatives from object-field joining and discriminant requirements. |
| internal/lower/view_lazy.go | Kept scoped allocation fallback, certified callable exemptions and intersection checks; added exact tuple-position certificate handling. |
| internal/lower/view_maps_tuples.go, modify/delete | Kept the single constructor/read path and optional/rest descriptors. Restored only tuple entry admission and its complete recursive descriptor proof in view_maps.go. |
| internal/oracle/checked_views_map_certificates_test.go | Retained integration's conservative entry-family controls, restored six existing storage-gap witnesses; optional/rest positives agree with Node. The old tuple-entry mismatch now reaches its named runtime Map storage refusal. Nested phantom, nullable and callable unsupported families retain compile refusals. |
| internal/oracle/interface_cast_counts_test.go | Retained tuple, ranked array and deferred intersection count helpers. |

The automatically merged changes also needed three explicit primitive-array
exclusions: native evaluate, native arrayIndexSlot (including typeof), and
JavaScript emitViewArrayRead. Each excludes only reads carrying TupleUnion on
IR. Other union and array consumers keep integration's dispatch. The existing
Narrow.Tuple bit and its lowering/JavaScript use were restored so native tuple
object storage retains JavaScript array identity on narrowed tuple reads.

The first bound smoke suite passed optional/rest contracts and original
emit-tuple-good/emit-helper-good (41.581s). Full-suite attempts then caught both
primitive-array interceptions as compiler invariant panics, not runtime mutant
kills. Those failures were corrected rather than accepted. A broad initial
phantom-descriptor restoration was narrowed after boundary controls caught its
changed admission. The final compiler keeps those existing refusals. Failed
attempt logs remain in /tmp/views-tuples-integration-unit.log,
/tmp/views-tuples-integration-unit-final.log and
/tmp/views-tuples-integration-boundaries.log; no green credit is given to them.

## Final commands and artifacts

Every shell sourced /workspace/adamic-tools/env.sh. The original declaration
input was ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations,
with its existing complete original declaration hashes verified by the oracles.
The restored toolchain was reused; nproc=5. No new fixture source was invented,
no cohere implementation copied, and no protected compiler orchestration file
was manually edited. All test output was redirected to logs.

```sh
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCheckedViewTuple.*|TestCheckedViewMapStorageGaps)$' -count=1 -v -timeout 8m -args -update-counts > /tmp/views-tuples-integration-certified.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCheckedViewArrays|TestCheckedViewNativeArrays|TestCheckedViewRankedArrayContracts|TestCheckedViewRankedArrayUnionContracts|TestCheckedViewRankedParserArrayContracts|TestCheckedViewObjectPrimitiveSource|TestCheckedViewMixedSelection|TestCheckedViewObjectUnions|TestCheckedViewPrimitiveArrayPairGap|TestCheckedViewMutableArrayUnionRefusal|TestCheckedViewPrimitiveArrayPairs|TestCheckedViewPrimitiveArraySafety|TestCheckedViewUntaggedArrayUnion)$' -count=1 -v -timeout 8m > /tmp/views-tuples-integration-lanes-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCheckedViewMapCertificates|TestCheckedViewMapCertificateWrites|TestCheckedViewMapPhantomRefusal|TestCheckedViewMapEntryFamilyBoundaries|TestCheckedViewNullishMapUnread|TestCheckedViewNullishMapReads|TestCheckedViewCallableArrays|TestCheckedViewUntaggedCallableUnion)$' -count=1 -v -timeout 8m > /tmp/views-tuples-integration-boundaries-final.log 2>&1
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run '^(TestTuple.*|TestViewFallbackScopesRetainWiderHelperGuard|TestViewAggregateNullishConstantsAreNotAllocations|TestViewCallableAggregateReturnedDemand|TestViewCallableBoxingUnknownProducer|TestViewCallableProducerCertificateNative|TestViewCallableProducerCertificateNode|TestViewCallableShapeNative|TestViewCallableShapeNode)$' -count=1 -v -timeout 4m > /tmp/views-tuples-integration-backends-final.log 2>&1
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1 -v -timeout 4m > /tmp/views-tuples-integration-counts-check.log 2>&1
```

Results: tuple suite 154.373s (188 uncached Node runs); 13 lane regression
functions 40.130s (255 uncached Node runs); eight Map/callable functions 51.713s
(171 uncached Node runs); IR 0.026s, lower 0.472s, native 5.300s, JavaScript
1.315s; independent counts 81.149s (84 native misses). The scoped refresh
remeasured 82 tuple rows and two legacy optional/rest Map rows, and all 84
values are unchanged from the tuple parent. Integration's other rows were
retained. gofmt on all manually reconciled Go files and git diff --check passed.

Committed full final logs are in integration-ffe428ab/. The mutation manifest
there records exact environment, selected test, actual exit status, expected
catch evidence and the individual log for every rerun. Each mutation command
used the original declaration environment and ADAMIC_GATE_UNCACHED=1 with:

```sh
go test ./internal/oracle -run '^TEST$' -count=1 -v -timeout 3m > LOG 2>&1
```

| Mode | Catch |
| --- | --- |
| Index presence | Missing-required read no longer stops, caught in JavaScript and both native builds. |
| Index scalar as tuple | Valid scalar is wrongly refused, caught by Node-held output in both backends. |
| forEach skip release | Semantics unchanged; LeakSanitizer catches 24 bytes in one allocation. |
| forEach double release | AddressSanitizer catches heap-use-after-free on a real scalar box. |
| Optional absent as present | Invented slot differs from Node. |
| Rest absent as present | Invented tail differs from Node. |
| Rest erased schema | Wrong rest contract loses its required named Map refusal. |
| Original signature forEach arity | Malformed tuple wrongly admitted. |
| Watch arity | Malformed callback parameter tuple wrongly admitted. |
| Watch narrowed demand | Missing required event produces NaN instead of stopping. |
| Module worker arity | Six-position tuple wrongly admitted. |
| Legacy Map source certificate | Both optional/rest positive witnesses are incorrectly refused. |

All 12 modes exit 1 at the expected real oracle, representing 13 witness kills
because the last mode checks two positives. No invariant panic or compiler
warning is counted as a mutant kill. The earlier batch's 31 logical mutant
results remain historical; this merge does not claim all 31 were rerun.

## Limits

The direct mixed-array-to-tuple cast remains refused at adamic/no-unchecked-cast
and its exact frontier oracle passed. Supporting that cast still needs an
alias-preserving representation bridge. Deferred owner nullish/callable Map
storage adapters and their removed twelve general-entry witnesses were not
restored: integration's Map physical storage, callback dispatch and named
refusals remain in force. Broad phantom Map admission was not restored.
The full gate and global counts were not run. No new unrelated failure is
claimed; prior integration-wide failures remain documented in the incoming
integration report and the previous tuple batch report. Whole-tsc execution,
exact reaching-view coverage and spread/variadic rest remain unclaimed.

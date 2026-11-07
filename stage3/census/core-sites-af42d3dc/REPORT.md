# Full core.ts site census after the wave

Task b5w3ycg. Report only, not for landing.

Main/compiler/apply SHA: `af42d3dc1261a3db536064c3dfbb5eaa7e8b6405`. TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Census child exit 0; elapsed 38.046 seconds. All of core.ts was selected with LATENT_FULL=1, LATENT_ONLY_FILE=/workspace/after-wave-adapted/src/compiler/core.ts, LATENT_ASSERT_NO_OUTPUT=1. No speculative or replay flags were set. Ordinary Load and Lower cannot return output in this guarded measurement binary.

Checker diagnostics: 0. Units: 271; statuses: {'lowered': 53, 'deferred_generic_instantiation': 209, 'finding': 9}. Distinct location/kind/reason/phase sites: 24; first walls: 17.

## Every recorded site, in source order

Repeated observations at the same location, kind, exact reason and phase share one F ID. `rows.json` retains all events and recording-unit identities. Emitters are pinned compiler source lines; `emitters.json` retains their exact source text and enclosing deciding function.

| Site | Location | Unit | Kind | Exact reason | Phase | Emitting lowering line and deciding function | First wall | Diff |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| F1 | core.ts:312:5 | clear | NotYet | an array of unknown with erased element storage (retain its declared element type before reading elements) | lowering | internal/lower/object.go:396 `(*lowering).elementType` | yes | new to inventory |
| F2 | core.ts:738:9 | deduplicateSorted | NotYet | a value of type NonNullable<T> | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` | no | new to inventory |
| F3 | core.ts:739:31 | deduplicateSorted | NotYet | an array of NonNullable<T> | lowering | internal/lower/object.go:401 `(*lowering).elementType` | no | new to inventory |
| F4 | core.ts:741:15 | deduplicateSorted | NotYet | a value of type NonNullable<T> | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` | no | new to inventory |
| F5 | core.ts:742:26 | deduplicateSorted | NotYet | reading next | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` | no | new to inventory |
| F6 | core.ts:756:9 | deduplicateSorted | NotYet | an array of T | lowering | internal/lower/object.go:401 `(*lowering).elementType` | no | new to inventory |
| F7 | core.ts:759:12 | deduplicateSorted | NotYet | reading deduplicated | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` | no | new to inventory |
| F8 | core.ts:1049:63 | toSorted | NotYet | an array of T | lowering | internal/lower/object.go:401 `(*lowering).elementType` | no | new to inventory |
| F9 | core.ts:1078:5 | elementAt | NotYet | a function returning T &#124; undefined | lowering | internal/lower/functions.go:129 `(*lowering).signature` | yes | new to inventory |
| F10 | core.ts:1317:18 | getAllKeys | Refused | Object.getPrototypeOf | lowering | internal/lower/library_object.go:20 `(*lowering).objectCallArguments`; internal/lower/library_object.go:26 `(*lowering).objectCallArguments` | yes | new to inventory |
| F11 | core.ts:2116:26 | createIntlCollatorStringComparer | Refused | a method read as a value (compare would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` | yes | remains |
| F12 | core.ts:2116:26 | createIntlCollatorStringComparer | Refused | inherited library member compare read as an own field | lowering | internal/lower/prototype.go:124 `(*lowering).prototypeRead` | yes | remains |
| F13 | core.ts:2117:52 | createIntlCollatorStringComparer | NotYet | reading comparer | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` | yes | remains |
| F14 | core.ts:2475:66 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate`; internal/lower/predicates.go:40 `(*predicateFlowProof).refused` | yes | remains |
| F15 | core.ts:2475:91 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate`; internal/lower/predicates.go:40 `(*predicateFlowProof).refused` | yes | remains |
| F16 | core.ts:2477:80 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate`; internal/lower/predicates.go:40 `(*predicateFlowProof).refused` | yes | remains |
| F17 | core.ts:2477:105 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate`; internal/lower/predicates.go:40 `(*predicateFlowProof).refused` | yes | remains |
| F18 | core.ts:2477:130 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate`; internal/lower/predicates.go:40 `(*predicateFlowProof).refused` | yes | remains |
| F19 | core.ts:2609:15 | process | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop`; internal/lower/host_bindings.go:107 `(*lowering).hostBindingStatement` | yes | remains |
| F20 | core.ts:2609:80 | require | Refused | an ambient value declaration | refusal_scan | internal/lower/ambient.go:79 `(*lowering).ambientRefusal` | yes | remains |
| F21 | core.ts:2614:12 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | lowering | internal/lower/host_bindings.go:139 `hostBindingStop`; internal/lower/host_bindings.go:361 `(*lowering).hostBindingTypeOf` | yes | remains |
| F22 | core.ts:2614:19 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop`; internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` | yes | remains |
| F23 | core.ts:2615:14 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop`; internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` | yes | remains |
| F24 | core.ts:2616:13 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop`; internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` | yes | remains |

## First-wall list

The list uses the 4186581e report's criterion: every refusal/contract scan finding, plus lowering findings outside deferred generic bodies, including signature-only overload checks. This is a source-order inventory, not a proven sequence of ordinary first stops after hypothetical repairs. Refusal-scan priority and statement rollback can change which stop wins or which downstream missing-local findings survive. A deferred body without a finding is not proven to lower.

1. F1, `core.ts:312:5`, NotYet: an array of unknown with erased element storage (retain its declared element type before reading elements) (lowering).
2. F9, `core.ts:1078:5`, NotYet: a function returning T | undefined (lowering).
3. F10, `core.ts:1317:18`, Refused: Object.getPrototypeOf (lowering).
4. F11, `core.ts:2116:26`, Refused: a method read as a value (compare would lose its object, and this with it) (refusal_scan).
5. F12, `core.ts:2116:26`, Refused: inherited library member compare read as an own field (lowering).
6. F13, `core.ts:2117:52`, NotYet: reading comparer (lowering).
7. F14, `core.ts:2475:66`, Refused: a type predicate whose return is not proven (there is no body proving this parameter) (refusal_scan).
8. F15, `core.ts:2475:91`, Refused: a type predicate whose return is not proven (there is no body proving this parameter) (refusal_scan).
9. F16, `core.ts:2477:80`, Refused: a type predicate whose return is not proven (there is no body proving this parameter) (refusal_scan).
10. F17, `core.ts:2477:105`, Refused: a type predicate whose return is not proven (there is no body proving this parameter) (refusal_scan).
11. F18, `core.ts:2477:130`, Refused: a type predicate whose return is not proven (there is no body proving this parameter) (refusal_scan).
12. F19, `core.ts:2609:15`, NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, (refusal_scan).
13. F20, `core.ts:2609:80`, Refused: an ambient value declaration (refusal_scan).
14. F21, `core.ts:2614:12`, NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, (lowering).
15. F22, `core.ts:2614:19`, NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, (refusal_scan).
16. F23, `core.ts:2615:14`, NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, (refusal_scan).
17. F24, `core.ts:2616:13`, NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, (refusal_scan).

## Diff against the 4186581e report's 48 sites

Gone: 34. Remain: 14. New to this inventory: 10.

Lexical source_key, kind, exact reason and phase; repeated matching sites paired in source order. Locations are adapted coordinates in their own snapshots. New includes newly surveyed sites before getOwnValues.

A match means the same observed blocker remains; it does not assert unchanged runtime reach. Sites earlier than the former getOwnValues range are explicitly newly surveyed, rather than claimed regressions.

| Previous ID | Previous location | Unit | Kind | Exact reason | Phase | Result |
| --- | --- | --- | --- | --- | --- | --- |
| F1 | core.ts:1326:26 | getOwnValues | Refused | a cast the runtime can't check | refusal_scan | gone |
| F2 | core.ts:1339:12 | arrayOf | NotYet | a member typed any ([]) related to T at the return site | refusal_scan | gone |
| F3 | core.ts:1364:29 | assign | NotYet | a non-object representation passed through an opaque object parameter | refusal_scan | gone |
| F4 | core.ts:1484:12 | groupBy | Refused | a cast the runtime can't check | refusal_scan | gone |
| F5 | core.ts:1492:27 | clone | Refused | a cast the runtime can't check | refusal_scan | gone |
| F6 | core.ts:1506:29 | extend | Refused | a cast the runtime can't check | refusal_scan | gone |
| F7 | core.ts:1509:14 | extend | Refused | a cast the runtime can't check | refusal_scan | gone |
| F8 | core.ts:1509:36 | extend | Refused | a cast the runtime can't check | refusal_scan | gone |
| F9 | core.ts:1515:14 | extend | Refused | a cast the runtime can't check | refusal_scan | gone |
| F10 | core.ts:1515:36 | extend | Refused | a cast the runtime can't check | refusal_scan | gone |
| F11 | core.ts:1526:14 | copyProperties | Refused | a cast the runtime can't check | refusal_scan | gone |
| F12 | core.ts:1553:17 | createMultiMap | Refused | a cast the runtime can't check | refusal_scan | gone |
| F13 | core.ts:1554:5 | createMultiMap | Refused | a method read as a value (add would lose its object, and this with it) | refusal_scan | gone |
| F14 | core.ts:1555:5 | createMultiMap | Refused | a method read as a value (remove would lose its object, and this with it) | refusal_scan | gone |
| F15 | core.ts:1580:41 | createQueue | Refused | a value of type T[] seen as (T &#124; undefined)[], which can write T &#124; undefined where T is read | refusal_scan | gone |
| F16 | core.ts:1580:41 | createQueue | Refused | a value of type T[] &#124; undefined seen as (T &#124; undefined)[], which can write T &#124; undefined where T is read | refusal_scan | gone |
| F17 | core.ts:1596:24 | dequeue | Refused | a cast the runtime can't check | refusal_scan | gone |
| F18 | core.ts:2003:9 | compareComparableValues | NotYet | a BinaryExpression with a union of differently held members and a union of differently held members | lowering | gone |
| F19 | core.ts:2125:26 | createIntlCollatorStringComparer | Refused | a method read as a value (compare would lose its object, and this with it) | refusal_scan | F11 at core.ts:2116:26 |
| F20 | core.ts:2125:26 | createIntlCollatorStringComparer | Refused | inherited library member compare read as an own field | lowering | F12 at core.ts:2116:26 |
| F21 | core.ts:2126:52 | createIntlCollatorStringComparer | NotYet | reading comparer | lowering | F13 at core.ts:2117:52 |
| F22 | core.ts:2224:20 | levenshteinWithMax | NotYet | new an Identifier | lowering | gone |
| F23 | core.ts:2225:19 | levenshteinWithMax | NotYet | new an Identifier | lowering | gone |
| F24 | core.ts:2230:9 | levenshteinWithMax | NotYet | reading previous | lowering | gone |
| F25 | core.ts:2237:9 | levenshteinWithMax | NotYet | reading current | lowering | gone |
| F26 | core.ts:2241:13 | levenshteinWithMax | NotYet | reading current | lowering | gone |
| F27 | core.ts:2245:19 | levenshteinWithMax | NotYet | a value of type any | lowering | gone |
| F28 | core.ts:2248:19 | levenshteinWithMax | NotYet | a value of type any | lowering | gone |
| F29 | core.ts:2251:13 | levenshteinWithMax | NotYet | reading current | lowering | gone |
| F30 | core.ts:2252:39 | levenshteinWithMax | NotYet | reading dist | lowering | gone |
| F31 | core.ts:2255:13 | levenshteinWithMax | NotYet | reading current | lowering | gone |
| F32 | core.ts:2262:22 | levenshteinWithMax | NotYet | reading previous | lowering | gone |
| F33 | core.ts:2263:9 | levenshteinWithMax | NotYet | assigning to an Identifier | lowering | gone |
| F34 | core.ts:2264:9 | levenshteinWithMax | NotYet | assigning to an Identifier | lowering | gone |
| F35 | core.ts:2267:11 | levenshteinWithMax | NotYet | a value of type any | lowering | gone |
| F36 | core.ts:2268:12 | levenshteinWithMax | NotYet | reading res | lowering | gone |
| F37 | core.ts:2484:66 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | F14 at core.ts:2475:66 |
| F38 | core.ts:2484:91 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | F15 at core.ts:2475:91 |
| F39 | core.ts:2486:80 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | F16 at core.ts:2477:80 |
| F40 | core.ts:2486:105 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | F17 at core.ts:2477:105 |
| F41 | core.ts:2486:130 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | F18 at core.ts:2477:130 |
| F42 | core.ts:2595:16 | takeWhile | Refused | a cast the runtime can't check | refusal_scan | gone |
| F43 | core.ts:2614:15 | process | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | F19 at core.ts:2609:15 |
| F44 | core.ts:2614:80 | require | Refused | an ambient value declaration | refusal_scan | F20 at core.ts:2609:80 |
| F45 | core.ts:2619:12 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | lowering | F21 at core.ts:2614:12 |
| F46 | core.ts:2619:19 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | F22 at core.ts:2614:19 |
| F47 | core.ts:2620:14 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | F23 at core.ts:2615:14 |
| F48 | core.ts:2621:13 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | F24 at core.ts:2616:13 |

New sites: F1 at core.ts:312:5, F2 at core.ts:738:9, F3 at core.ts:739:31, F4 at core.ts:741:15, F5 at core.ts:742:26, F6 at core.ts:756:9, F7 at core.ts:759:12, F8 at core.ts:1049:63, F9 at core.ts:1078:5, F10 at core.ts:1317:18.

## Evidence and reproduction

Both runs used the same frozen full apply. `adapted-source-manifest.json`, `adapted-product.manifest.json`, `core.ts.txt` and `provenance.json` identify its exact bytes. `validation.json` checks the source stayed unchanged. Raw core census, unit timings and census log are in `raw-streams.tar.gz`, with `raw-streams.tar.gz.sha256`. The previous 48-site report and machine-readable rows are retained for the diff.

The overlay is generated by the pinned meter tools through `../next-stops-af42d3dc/build-entry.py`. This changes only physical scratch filenames to .go.txt. No production compiler source is modified. `run-census.py` retains the exact flags, process bounds and cleanup. `ast-map.cjs` and `analyze.py` retain declaration attribution and first-wall selection.

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/core-sites-af42d3dc/run-census.py > core-runner.log 2>&1
node stage3/census/core-sites-af42d3dc/ast-map.cjs > ast-map.log 2>&1
python3 stage3/census/core-sites-af42d3dc/analyze.py > analysis.log 2>&1
python3 stage3/census/core-sites-af42d3dc/map-emitters.py > emitter-map.log 2>&1
python3 stage3/census/core-sites-af42d3dc/compare-sites.py > site-comparison.log 2>&1
```

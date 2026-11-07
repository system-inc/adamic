# Core sites from getOwnValues to end of file

Task b5w3ycg. Report only, not for landing.

Box: `2c759e5913cb`. Environment: `ccarenv_b64_Y2NhcmVudl9kZjA5ZTQ0YTliMDQ4MTkxYmIzOTY3MTVlN2ZjM2U4NQ`. CPU quota: 4 cores.

Main: `4186581ec18f469c40149bf217d22b028e8374fb`. TypeScript 6.0.3 pin: `050880ce59e30b356b686bd3144efe24f875ebc8`. The adaptation stack and census overlay/tool are from this same main SHA.

Setup: 1013 seconds, exit 0. Apply: 1496 seconds, exit 0. Census: 38.004 seconds (0.633 minutes), exit 0. Timeout: false. The census process group was cleaned up.

Adapted source manifest SHA256: `69cc9ce27e1abc500e51c57dab2af97c9f68d78dde5bc33c87b55b1454a8c736` (746 files). Product manifest SHA256: `a3ad6181c4cf917e65c92eb4be5236bcc8cb8afe47d7ea41e06291b1c906d7c4`. See [provenance.json](provenance.json), [adapted-source-manifest.json](adapted-source-manifest.json), and [adapted-product.manifest.json](adapted-product.manifest.json).

[raw.jsonl](raw.jsonl) is the unchanged complete stream. [rows.json](rows.json) retains every raw finding event, selected unit rows, grouped sites, recording-unit identities and the requested-function index. [unit-times.jsonl](unit-times.jsonl) retains progress and per-unit timings. The tool is measurement-only; both no-output assertions ran before the first raw record.

Checker diagnostics: 0. The raw historical label `measured on a checker-rejected program` is retained even when the diagnostic list is empty.

Range totals: 130 units, 43 lowered, 9 finding, 78 deferred, 48 distinct finding/phase sites. Deferred generic units: **78**, including **76 instantiated=false** and **2 instantiated=true**. These counts include signature-only overload units and units deferred because of generic lexical ancestors, exactly as the meter labels them.

`createSet` at 1647 and its nested `getElementIterator` at 1651 are both deferred and instantiated=false, with no recorded site finding. There is no measured generator stop there. `arrayOf` at 1339 has a refusal-scan NotYet even though its generic body is deferred. The only deferred units with instantiated=true are `identity` at 1853 and `equateValues` at 1962.

## Reading the rows

The raw meter calls ordinary units `attempted`. This report calls such a unit `finding` if it records a finding or contains a site finding, and `lowered` otherwise. `lowered` means no recorded finding in this independent unit probe, not successful whole-program compilation. Raw deferred statuses and instantiation witnesses are unchanged. A refusal scan can report a site inside a generic whose body was never instantiated. Absence of a lowering finding in that body is not a claim that the body lowers.

Units are sorted by adapted source position, including nested named functions, overload signatures, types and variable statements. Upstream lines are matched by declaration name, lexical parent and body/signature identity. Overloads whose text changes retain their same-name signature order. Split `or` overloads map to their common upstream signature; adaptation-added declarations have no upstream line. Site findings are assigned to the smallest containing unit. Repeated observations at the same location, kind, reason and phase share one F ID; all events remain in JSON. Boundary markers are retained in JSON and are not Refused/NotYet sites.

## First walls

These are the instantiation-independent sites in source order: every refusal/contract scan finding, and lowering findings outside a deferred body, including overload checks. All 48 sites in this range satisfy that instantiation criterion. The first is marked. This is the requested source-order inventory, not a proven sequence of ordinary root stops after fixes. The refusal scan precedes lowering, so a later scan finding can precede an earlier lowering finding. In particular F2 is stored as the lowest-priority `refusal.deferred` stop by `refuseAnyRelation`; it is returned only if no stronger module refusal wins. This priority mechanism is separate from generic body deferral. The full census independently probes units and does not rerun all roots after hypothetical fixes. Later missing-local findings in `levenshteinWithMax` and `createIntlCollatorStringComparer` follow failed declarations and state rollback, and may disappear when those earlier failures are fixed.

| Order | Site | Location | Kind | Exact reason | Phase | Emitter and deciding function |
| --- | --- | --- | --- | --- | --- | --- |
| **FIRST** | F1 | core.ts:1326:26 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 2 | F2 | core.ts:1339:12 | NotYet | a member typed any ([]) related to T at the return site | refusal_scan | internal/lower/any_relations.go:43 `(*lowering).refuseAnyRelation` (stores a lowest-priority refusal.deferred stop; this is not deferred_generic_instantiation) |
| 3 | F3 | core.ts:1364:29 | NotYet | a non-object representation passed through an opaque object parameter | refusal_scan | internal/lower/detached_own.go:94 `(*lowering).detachedOwnRefusal` |
| 4 | F4 | core.ts:1484:12 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 5 | F5 | core.ts:1492:27 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 6 | F6 | core.ts:1506:29 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 7 | F7 | core.ts:1509:14 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 8 | F8 | core.ts:1509:36 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 9 | F9 | core.ts:1515:14 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 10 | F10 | core.ts:1515:36 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 11 | F11 | core.ts:1526:14 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 12 | F12 | core.ts:1553:17 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 13 | F13 | core.ts:1554:5 | Refused | a method read as a value (add would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` |
| 14 | F14 | core.ts:1555:5 | Refused | a method read as a value (remove would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` |
| 15 | F15 | core.ts:1580:41 | Refused | a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | refusal_scan | internal/lower/invariance.go:726 `(*lowering).wideningRefusal` (builds the exact reason; returns Refused at line 740, reached by refuseWidening/refuseWholeWidening) |
| 16 | F16 | core.ts:1580:41 | Refused | a value of type T[] \| undefined seen as (T \| undefined)[], which can write T \| undefined where T is read | refusal_scan | internal/lower/invariance.go:726 `(*lowering).wideningRefusal` (builds the exact reason; returns Refused at line 740, reached by refuseWidening/refuseWholeWidening) |
| 17 | F17 | core.ts:1596:24 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 18 | F18 | core.ts:2003:9 | NotYet | a BinaryExpression with a union of differently held members and a union of differently held members | lowering | internal/lower/expression.go:1132 `(*lowering).combine` (typeName at line 1295 supplies union of differently held members) |
| 19 | F19 | core.ts:2125:26 | Refused | a method read as a value (compare would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` |
| 20 | F20 | core.ts:2125:26 | Refused | inherited library member compare read as an own field | lowering | internal/lower/prototype.go:124 `(*lowering).prototypeRead` |
| 21 | F21 | core.ts:2126:52 | NotYet | reading comparer | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 22 | F22 | core.ts:2224:20 | NotYet | new an Identifier | lowering | internal/lower/object.go:1811 `(*lowering).newExpression` (new Array(length) is not handled by this fallback) |
| 23 | F23 | core.ts:2225:19 | NotYet | new an Identifier | lowering | internal/lower/object.go:1811 `(*lowering).newExpression` (new Array(length) is not handled by this fallback) |
| 24 | F24 | core.ts:2230:9 | NotYet | reading previous | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 25 | F25 | core.ts:2237:9 | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 26 | F26 | core.ts:2241:13 | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 27 | F27 | core.ts:2245:19 | NotYet | a value of type any | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` (the checker type has no supported runtime representation) |
| 28 | F28 | core.ts:2248:19 | NotYet | a value of type any | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` (the checker type has no supported runtime representation) |
| 29 | F29 | core.ts:2251:13 | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 30 | F30 | core.ts:2252:39 | NotYet | reading dist | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 31 | F31 | core.ts:2255:13 | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 32 | F32 | core.ts:2262:22 | NotYet | reading previous | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 33 | F33 | core.ts:2263:9 | NotYet | assigning to an Identifier | lowering | internal/lower/assignments.go:65 `(*lowering).assignment` (Identifier target has no represented local after rollback) |
| 34 | F34 | core.ts:2264:9 | NotYet | assigning to an Identifier | lowering | internal/lower/assignments.go:65 `(*lowering).assignment` (Identifier target has no represented local after rollback) |
| 35 | F35 | core.ts:2267:11 | NotYet | a value of type any | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` (the checker type has no supported runtime representation) |
| 36 | F36 | core.ts:2268:12 | NotYet | reading res | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| 37 | F37 | core.ts:2484:66 | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| 38 | F38 | core.ts:2484:91 | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| 39 | F39 | core.ts:2486:80 | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| 40 | F40 | core.ts:2486:105 | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| 41 | F41 | core.ts:2486:130 | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| 42 | F42 | core.ts:2595:16 | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| 43 | F43 | core.ts:2614:15 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:107 `(*lowering).hostBindingStatement` (emits NotYet at the declaration) |
| 44 | F44 | core.ts:2614:80 | Refused | an ambient value declaration | refusal_scan | internal/lower/ambient.go:79 `(*lowering).ambientRefusal` (selects the reason; constructs Refused at line 91) |
| 45 | F45 | core.ts:2619:12 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | lowering | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:361 `(*lowering).hostBindingTypeOf` (emits NotYet for typeof process) |
| 46 | F46 | core.ts:2619:19 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` (emits NotYet at the Identifier use) |
| 47 | F47 | core.ts:2620:14 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` (emits NotYet at the Identifier use) |
| 48 | F48 | core.ts:2621:13 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` (emits NotYet at the Identifier use) |

## Every unit

| Unit | Adapted line:column | Upstream line | Name | Status | Site findings |
| --- | --- | --- | --- | --- | --- |
| U142 | 1322:1 | 1312 | getOwnValues | deferred_generic_instantiation (instantiated=false) | F1 |
| U143 | 1334:1 | 1324 | arrayOf | deferred_generic_instantiation (instantiated=false) | F2 |
| U144 | 1347:1 | 1337 | arrayFrom | deferred_generic_instantiation (instantiated=false) | none |
| U145 | 1349:1 | 1339 | arrayFrom | deferred_generic_instantiation (instantiated=false) | none |
| U146 | 1351:1 | 1341 | arrayFrom | deferred_generic_instantiation (instantiated=false) | none |
| U147 | 1360:1 | 1350 | assign | deferred_generic_instantiation (instantiated=false) | F3 |
| U148 | 1380:1 | 1370 | equalOwnProperties | deferred_generic_instantiation (instantiated=false) | none |
| U149 | 1411:1 | 1401 | arrayToMap | deferred_generic_instantiation (instantiated=false) | none |
| U150 | 1413:1 | 1403 | arrayToMap | deferred_generic_instantiation (instantiated=false) | none |
| U151 | 1415:1 | 1405 | arrayToMap | deferred_generic_instantiation (instantiated=false) | none |
| U152 | 1417:1 | 1407 | arrayToMap | deferred_generic_instantiation (instantiated=false) | none |
| U153 | 1419:1 | 1409 | arrayToMap | deferred_generic_instantiation (instantiated=false) | none |
| U154 | 1430:1 | 1420 | arrayToNumericMap | deferred_generic_instantiation (instantiated=false) | none |
| U155 | 1432:1 | 1422 | arrayToNumericMap | deferred_generic_instantiation (instantiated=false) | none |
| U156 | 1434:1 | 1424 | arrayToNumericMap | deferred_generic_instantiation (instantiated=false) | none |
| U157 | 1444:1 | 1434 | arrayToMultiMap | deferred_generic_instantiation (instantiated=false) | none |
| U158 | 1446:1 | 1436 | arrayToMultiMap | deferred_generic_instantiation (instantiated=false) | none |
| U159 | 1448:1 | 1438 | arrayToMultiMap | deferred_generic_instantiation (instantiated=false) | none |
| U160 | 1458:1 | 1448 | group | deferred_generic_instantiation (instantiated=false) | none |
| U161 | 1460:1 | 1450 | group | deferred_generic_instantiation (instantiated=false) | none |
| U162 | 1462:1 | 1452 | group | deferred_generic_instantiation (instantiated=false) | none |
| U163 | 1464:1 | 1454 | group | deferred_generic_instantiation (instantiated=false) | none |
| U164 | 1466:1 | 1456 | group | deferred_generic_instantiation (instantiated=false) | none |
| U165 | 1471:1 | 1461 | groupBy | deferred_generic_instantiation (instantiated=false) | none |
| U166 | 1473:1 | 1463 | groupBy | deferred_generic_instantiation (instantiated=false) | none |
| U167 | 1474:1 | 1464 | groupBy | deferred_generic_instantiation (instantiated=false) | F4 |
| U168 | 1488:1 | 1478 | clone | deferred_generic_instantiation (instantiated=false) | F5 |
| U169 | 1505:1 | 1495 | extend | deferred_generic_instantiation (instantiated=false) | F6, F7, F8, F9, F10 |
| U170 | 1523:1 | 1513 | copyProperties | deferred_generic_instantiation (instantiated=false) | F11 |
| U171 | 1532:1 | 1522 | maybeBind | deferred_generic_instantiation (instantiated=false) | none |
| U172 | 1537:1 | 1527 | MultiMap | lowered | none |
| U173 | 1552:1 | 1542 | createMultiMap | deferred_generic_instantiation (instantiated=false) | F12, F13, F14 |
| U174 | 1558:1 | 1548 | multiMapAdd | deferred_generic_instantiation (instantiated=false) | none |
| U175 | 1568:1 | 1558 | multiMapRemove | deferred_generic_instantiation (instantiated=false) | none |
| U176 | 1579:1 | 1569 | createQueue | deferred_generic_instantiation (instantiated=false) | F15, F16 |
| U177 | 1583:5 | 1573 | isEmpty | deferred_generic_instantiation (instantiated=false) | none |
| U178 | 1587:5 | 1577 | enqueue | deferred_generic_instantiation (instantiated=false) | none |
| U179 | 1591:5 | 1581 | dequeue | deferred_generic_instantiation (instantiated=false) | F17 |
| U180 | 1621:1 | added | HashedSet | lowered | none |
| U181 | 1647:1 | 1622 | createSet | deferred_generic_instantiation (instantiated=false) | none |
| U182 | 1651:5 | 1626 | getElementIterator | deferred_generic_instantiation (instantiated=false) | none |
| U183 | 1775:1 | 1750 | isArray | lowered | none |
| U184 | 1781:1 | 1756 | toArray | deferred_generic_instantiation (instantiated=false) | none |
| U185 | 1783:1 | 1758 | toArray | deferred_generic_instantiation (instantiated=false) | none |
| U186 | 1785:1 | 1760 | toArray | deferred_generic_instantiation (instantiated=false) | none |
| U187 | 1794:1 | 1769 | isString | lowered | none |
| U188 | 1798:1 | 1773 | isNumber | lowered | none |
| U189 | 1803:1 | 1778 | tryCast | deferred_generic_instantiation (instantiated=false) | none |
| U190 | 1808:1 | 1783 | cast | deferred_generic_instantiation (instantiated=false) | none |
| U191 | 1819:1 | 1794 | noop | lowered | none |
| U192 | 1826:1 | 1801 | returnFalse | lowered | none |
| U193 | 1835:1 | 1810 | returnTrue | lowered | none |
| U194 | 1844:1 | 1819 | returnUndefined | lowered | none |
| U195 | 1853:1 | 1828 | identity | deferred_generic_instantiation (instantiated=true) | none |
| U196 | 1860:1 | 1835 | toLowerCase | lowered | none |
| U197 | 1886:1 | 1861 | fileNameLowerCaseRegExp | lowered | none |
| U198 | 1900:1 | 1875 | toFileNameLowerCase | lowered | none |
| U199 | 1911:1 | 1886 | notImplemented | lowered | none |
| U200 | 1916:1 | 1891 | memoize | deferred_generic_instantiation (instantiated=false) | none |
| U201 | 1932:1 | 1907 | memoizeOne | deferred_generic_instantiation (instantiated=false) | none |
| U202 | 1946:1 | 1921 | AssertionLevel | lowered | none |
| U203 | 1959:1 | 1934 | AnyFunction | lowered | none |
| U204 | 1962:1 | 1937 | equateValues | deferred_generic_instantiation (instantiated=true) | none |
| U205 | 1976:1 | 1951 | equateStringsCaseInsensitive | lowered | none |
| U206 | 1991:1 | 1966 | equateStringsCaseSensitive | lowered | none |
| U207 | 1996:1 | 1971 | compareComparableValues | finding | none |
| U208 | 1998:1 | 1973 | compareComparableValues | finding | none |
| U209 | 1999:1 | 1974 | compareComparableValues | finding | F18 |
| U210 | 2013:1 | 1988 | compareValues | lowered | none |
| U211 | 2022:1 | 1997 | compareTextSpans | lowered | none |
| U212 | 2027:1 | 2002 | maxBy | deferred_generic_instantiation (instantiated=false) | none |
| U213 | 2035:1 | 2010 | min | deferred_generic_instantiation (instantiated=false) | none |
| U214 | 2037:1 | 2012 | min | deferred_generic_instantiation (instantiated=false) | none |
| U215 | 2039:1 | 2014 | min | deferred_generic_instantiation (instantiated=false) | none |
| U216 | 2057:1 | 2032 | compareStringsCaseInsensitive | lowered | none |
| U217 | 2078:1 | 2053 | compareStringsCaseInsensitiveEslintCompatible | lowered | none |
| U218 | 2099:1 | 2074 | compareStringsCaseSensitive | lowered | none |
| U219 | 2104:1 | 2079 | getStringComparer | lowered | none |
| U220 | 2111:1 | 2086 | createUIStringComparer | finding | none |
| U221 | 2114:5 | 2089 | compareWithCallback | lowered | none |
| U222 | 2122:5 | 2097 | createIntlCollatorStringComparer | finding | F19, F20, F21 |
| U223 | 2130:1 | 2105 | uiComparerCaseSensitive | lowered | none |
| U224 | 2131:1 | 2106 | uiLocale | lowered | none |
| U225 | 2134:1 | 2109 | getUILocale | lowered | none |
| U226 | 2139:1 | 2114 | setUILocale | lowered | none |
| U227 | 2158:1 | 2133 | compareStringsCaseSensitiveUI | lowered | none |
| U228 | 2164:1 | 2139 | compareProperties | deferred_generic_instantiation (instantiated=false) | none |
| U229 | 2176:1 | 2151 | compareBooleans | lowered | none |
| U230 | 2194:1 | 2169 | getSpellingSuggestion | deferred_generic_instantiation (instantiated=false) | none |
| U231 | 2223:1 | 2198 | levenshteinWithMax | finding | F22, F23, F24, F25, F26, F27, F28, F29, F30, F31, F32, F33, F34, F35, F36 |
| U232 | 2272:1 | 2247 | endsWith | lowered | none |
| U233 | 2282:1 | 2257 | removeSuffix | lowered | none |
| U234 | 2287:1 | 2262 | tryRemoveSuffix | lowered | none |
| U235 | 2296:1 | 2271 | removeMinAndVersionNumbers | lowered | none |
| U236 | 2350:1 | 2325 | orderedRemoveItem | deferred_generic_instantiation (instantiated=false) | none |
| U237 | 2365:1 | 2340 | orderedRemoveItemAt | deferred_generic_instantiation (instantiated=false) | none |
| U238 | 2373:1 | 2348 | unorderedRemoveItemAt | deferred_generic_instantiation (instantiated=false) | none |
| U239 | 2384:1 | 2359 | unorderedRemoveItem | deferred_generic_instantiation (instantiated=false) | none |
| U240 | 2389:1 | 2364 | unorderedRemoveFirstItemWhere | deferred_generic_instantiation (instantiated=false) | none |
| U241 | 2400:1 | 2375 | GetCanonicalFileName | lowered | none |
| U242 | 2402:1 | 2377 | createGetCanonicalFileName | lowered | none |
| U243 | 2411:1 | 2386 | Pattern | lowered | none |
| U244 | 2417:1 | 2392 | patternText | lowered | none |
| U245 | 2427:1 | 2402 | matchedText | lowered | none |
| U246 | 2437:1 | 2412 | findBestPatternMatch | deferred_generic_instantiation (instantiated=false) | none |
| U247 | 2455:1 | 2430 | startsWith | lowered | none |
| U248 | 2462:1 | 2437 | removePrefix | lowered | none |
| U249 | 2467:1 | 2442 | tryRemovePrefix | lowered | none |
| U250 | 2472:1 | 2447 | isPatternMatch | lowered | none |
| U251 | 2479:1 | 2454 | and | deferred_generic_instantiation (instantiated=false) | none |
| U252 | 2484:1 | 2459 | or | deferred_generic_instantiation (instantiated=false) | F37, F38 |
| U253 | 2486:1 | 2461 | or | deferred_generic_instantiation (instantiated=false) | F39, F40, F41 |
| U254 | 2489:1 | 2463 | or | deferred_generic_instantiation (instantiated=false) | none |
| U255 | 2491:1 | 2463 | or | deferred_generic_instantiation (instantiated=false) | none |
| U256 | 2493:1 | 2465 | or | deferred_generic_instantiation (instantiated=false) | none |
| U257 | 2507:1 | 2479 | not | deferred_generic_instantiation (instantiated=false) | none |
| U258 | 2512:1 | 2484 | assertType | deferred_generic_instantiation (instantiated=false) | none |
| U259 | 2515:1 | 2487 | singleElementArray | deferred_generic_instantiation (instantiated=false) | none |
| U260 | 2520:1 | 2492 | enumerateInsertsAndDeletes | deferred_generic_instantiation (instantiated=false) | none |
| U261 | 2559:1 | 2531 | cartesianProduct | deferred_generic_instantiation (instantiated=false) | none |
| U262 | 2565:1 | 2537 | cartesianProductWorker | deferred_generic_instantiation (instantiated=false) | none |
| U263 | 2585:1 | 2557 | takeWhile | deferred_generic_instantiation (instantiated=false) | none |
| U264 | 2587:1 | 2559 | takeWhile | deferred_generic_instantiation (instantiated=false) | none |
| U265 | 2588:1 | 2560 | takeWhile | deferred_generic_instantiation (instantiated=false) | F42 |
| U266 | 2600:1 | 2572 | skipWhile | deferred_generic_instantiation (instantiated=false) | none |
| U267 | 2602:1 | 2574 | skipWhile | deferred_generic_instantiation (instantiated=false) | none |
| U268 | 2604:1 | 2576 | skipWhile | deferred_generic_instantiation (instantiated=false) | none |
| U269 | 2614:1 | added | process | finding | F43 |
| U270 | 2614:80 | added | require | finding | F44 |
| U271 | 2616:1 | 2588 | isNodeLikeSystem | finding | F45, F46, F47, F48 |

## Every finding by site

The reason column preserves exact words. The emitter column names production `internal/lower` locations at pinned main, not shifted scratch-overlay lines. `emitters.json` retains the source search and decision-path notes.

| Site | Location | Unit | Kind | Exact reason | Recorded phase | Emitter and deciding function |
| --- | --- | --- | --- | --- | --- | --- |
| F1 | core.ts:1326:26 | getOwnValues | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F2 | core.ts:1339:12 | arrayOf | NotYet | a member typed any ([]) related to T at the return site | refusal_scan | internal/lower/any_relations.go:43 `(*lowering).refuseAnyRelation` (stores a lowest-priority refusal.deferred stop; this is not deferred_generic_instantiation) |
| F3 | core.ts:1364:29 | assign | NotYet | a non-object representation passed through an opaque object parameter | refusal_scan | internal/lower/detached_own.go:94 `(*lowering).detachedOwnRefusal` |
| F4 | core.ts:1484:12 | groupBy | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F5 | core.ts:1492:27 | clone | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F6 | core.ts:1506:29 | extend | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F7 | core.ts:1509:14 | extend | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F8 | core.ts:1509:36 | extend | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F9 | core.ts:1515:14 | extend | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F10 | core.ts:1515:36 | extend | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F11 | core.ts:1526:14 | copyProperties | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F12 | core.ts:1553:17 | createMultiMap | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F13 | core.ts:1554:5 | createMultiMap | Refused | a method read as a value (add would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` |
| F14 | core.ts:1555:5 | createMultiMap | Refused | a method read as a value (remove would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` |
| F15 | core.ts:1580:41 | createQueue | Refused | a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | refusal_scan | internal/lower/invariance.go:726 `(*lowering).wideningRefusal` (builds the exact reason; returns Refused at line 740, reached by refuseWidening/refuseWholeWidening) |
| F16 | core.ts:1580:41 | createQueue | Refused | a value of type T[] \| undefined seen as (T \| undefined)[], which can write T \| undefined where T is read | refusal_scan | internal/lower/invariance.go:726 `(*lowering).wideningRefusal` (builds the exact reason; returns Refused at line 740, reached by refuseWidening/refuseWholeWidening) |
| F17 | core.ts:1596:24 | dequeue | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F18 | core.ts:2003:9 | compareComparableValues | NotYet | a BinaryExpression with a union of differently held members and a union of differently held members | lowering | internal/lower/expression.go:1132 `(*lowering).combine` (typeName at line 1295 supplies union of differently held members) |
| F19 | core.ts:2125:26 | createIntlCollatorStringComparer | Refused | a method read as a value (compare would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(*lowering).methodValueRefusal` |
| F20 | core.ts:2125:26 | createIntlCollatorStringComparer | Refused | inherited library member compare read as an own field | lowering | internal/lower/prototype.go:124 `(*lowering).prototypeRead` |
| F21 | core.ts:2126:52 | createIntlCollatorStringComparer | NotYet | reading comparer | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F22 | core.ts:2224:20 | levenshteinWithMax | NotYet | new an Identifier | lowering | internal/lower/object.go:1811 `(*lowering).newExpression` (new Array(length) is not handled by this fallback) |
| F23 | core.ts:2225:19 | levenshteinWithMax | NotYet | new an Identifier | lowering | internal/lower/object.go:1811 `(*lowering).newExpression` (new Array(length) is not handled by this fallback) |
| F24 | core.ts:2230:9 | levenshteinWithMax | NotYet | reading previous | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F25 | core.ts:2237:9 | levenshteinWithMax | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F26 | core.ts:2241:13 | levenshteinWithMax | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F27 | core.ts:2245:19 | levenshteinWithMax | NotYet | a value of type any | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` (the checker type has no supported runtime representation) |
| F28 | core.ts:2248:19 | levenshteinWithMax | NotYet | a value of type any | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` (the checker type has no supported runtime representation) |
| F29 | core.ts:2251:13 | levenshteinWithMax | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F30 | core.ts:2252:39 | levenshteinWithMax | NotYet | reading dist | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F31 | core.ts:2255:13 | levenshteinWithMax | NotYet | reading current | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F32 | core.ts:2262:22 | levenshteinWithMax | NotYet | reading previous | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F33 | core.ts:2263:9 | levenshteinWithMax | NotYet | assigning to an Identifier | lowering | internal/lower/assignments.go:65 `(*lowering).assignment` (Identifier target has no represented local after rollback) |
| F34 | core.ts:2264:9 | levenshteinWithMax | NotYet | assigning to an Identifier | lowering | internal/lower/assignments.go:65 `(*lowering).assignment` (Identifier target has no represented local after rollback) |
| F35 | core.ts:2267:11 | levenshteinWithMax | NotYet | a value of type any | lowering | internal/lower/expression.go:53 `(*lowering).typeOf` (the checker type has no supported runtime representation) |
| F36 | core.ts:2268:12 | levenshteinWithMax | NotYet | reading res | lowering | internal/lower/expression.go:660 `(*lowering).enumNeverValue` (Identifier branch: the local is absent after an earlier failed statement was rolled back) |
| F37 | core.ts:2484:66 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| F38 | core.ts:2484:91 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| F39 | core.ts:2486:80 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| F40 | core.ts:2486:105 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| F41 | core.ts:2486:130 | or | Refused | a type predicate whose return is not proven (there is no body proving this parameter) | refusal_scan | internal/lower/predicates.go:50 `(*lowering).proveFlowPredicate` (decides that no function body proves the parameter); internal/lower/predicates.go:40 `(*predicateFlowProof).refused` (constructs the full reason) |
| F42 | core.ts:2595:16 | takeWhile | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:188 `(*lowering).castProof` (castRefusal at line 156 invokes this decision during the refusal scan) |
| F43 | core.ts:2614:15 | process | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:107 `(*lowering).hostBindingStatement` (emits NotYet at the declaration) |
| F44 | core.ts:2614:80 | require | Refused | an ambient value declaration | refusal_scan | internal/lower/ambient.go:79 `(*lowering).ambientRefusal` (selects the reason; constructs Refused at line 91) |
| F45 | core.ts:2619:12 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | lowering | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:361 `(*lowering).hostBindingTypeOf` (emits NotYet for typeof process) |
| F46 | core.ts:2619:19 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` (emits NotYet at the Identifier use) |
| F47 | core.ts:2620:14 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` (emits NotYet at the Identifier use) |
| F48 | core.ts:2621:13 | isNodeLikeSystem | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | refusal_scan | internal/lower/host_bindings.go:139 `hostBindingStop` (constructs the exact reason from the pinned host binding table); internal/lower/host_bindings.go:269 `(*lowering).hostBindingRefusal` (emits NotYet at the Identifier use) |

## Deferred generic units without an instantiation witness

Every unit labeled deferred_generic_instantiation and instantiated=false in the requested range is below. Findings listed here may be refusal-scan findings; their presence does not establish body lowering. `none` means the census recorded no site finding in that unit.

| Unit | Adapted line | Upstream line | Name | Body present | Recorded site findings |
| --- | --- | --- | --- | --- | --- |
| U142 | 1322 | 1312 | getOwnValues | True | F1 |
| U143 | 1334 | 1324 | arrayOf | True | F2 |
| U144 | 1347 | 1337 | arrayFrom | False | none |
| U145 | 1349 | 1339 | arrayFrom | False | none |
| U146 | 1351 | 1341 | arrayFrom | True | none |
| U147 | 1360 | 1350 | assign | True | F3 |
| U148 | 1380 | 1370 | equalOwnProperties | True | none |
| U149 | 1411 | 1401 | arrayToMap | False | none |
| U150 | 1413 | 1403 | arrayToMap | False | none |
| U151 | 1415 | 1405 | arrayToMap | False | none |
| U152 | 1417 | 1407 | arrayToMap | False | none |
| U153 | 1419 | 1409 | arrayToMap | True | none |
| U154 | 1430 | 1420 | arrayToNumericMap | False | none |
| U155 | 1432 | 1422 | arrayToNumericMap | False | none |
| U156 | 1434 | 1424 | arrayToNumericMap | True | none |
| U157 | 1444 | 1434 | arrayToMultiMap | False | none |
| U158 | 1446 | 1436 | arrayToMultiMap | False | none |
| U159 | 1448 | 1438 | arrayToMultiMap | True | none |
| U160 | 1458 | 1448 | group | False | none |
| U161 | 1460 | 1450 | group | False | none |
| U162 | 1462 | 1452 | group | False | none |
| U163 | 1464 | 1454 | group | False | none |
| U164 | 1466 | 1456 | group | True | none |
| U165 | 1471 | 1461 | groupBy | False | none |
| U166 | 1473 | 1463 | groupBy | False | none |
| U167 | 1474 | 1464 | groupBy | True | F4 |
| U168 | 1488 | 1478 | clone | True | F5 |
| U169 | 1505 | 1495 | extend | True | F6, F7, F8, F9, F10 |
| U170 | 1523 | 1513 | copyProperties | True | F11 |
| U171 | 1532 | 1522 | maybeBind | True | none |
| U173 | 1552 | 1542 | createMultiMap | True | F12, F13, F14 |
| U174 | 1558 | 1548 | multiMapAdd | True | none |
| U175 | 1568 | 1558 | multiMapRemove | True | none |
| U176 | 1579 | 1569 | createQueue | True | F15, F16 |
| U177 | 1583 | 1573 | isEmpty | True | none |
| U178 | 1587 | 1577 | enqueue | True | none |
| U179 | 1591 | 1581 | dequeue | True | F17 |
| U181 | 1647 | 1622 | createSet | True | none |
| U182 | 1651 | 1626 | getElementIterator | True | none |
| U184 | 1781 | 1756 | toArray | False | none |
| U185 | 1783 | 1758 | toArray | False | none |
| U186 | 1785 | 1760 | toArray | True | none |
| U189 | 1803 | 1778 | tryCast | True | none |
| U190 | 1808 | 1783 | cast | True | none |
| U200 | 1916 | 1891 | memoize | True | none |
| U201 | 1932 | 1907 | memoizeOne | True | none |
| U212 | 2027 | 2002 | maxBy | True | none |
| U213 | 2035 | 2010 | min | False | none |
| U214 | 2037 | 2012 | min | False | none |
| U215 | 2039 | 2014 | min | True | none |
| U228 | 2164 | 2139 | compareProperties | True | none |
| U230 | 2194 | 2169 | getSpellingSuggestion | True | none |
| U236 | 2350 | 2325 | orderedRemoveItem | True | none |
| U237 | 2365 | 2340 | orderedRemoveItemAt | True | none |
| U238 | 2373 | 2348 | unorderedRemoveItemAt | True | none |
| U239 | 2384 | 2359 | unorderedRemoveItem | True | none |
| U240 | 2389 | 2364 | unorderedRemoveFirstItemWhere | True | none |
| U246 | 2437 | 2412 | findBestPatternMatch | True | none |
| U251 | 2479 | 2454 | and | True | none |
| U252 | 2484 | 2459 | or | False | F37, F38 |
| U253 | 2486 | 2461 | or | False | F39, F40, F41 |
| U254 | 2489 | 2463 | or | False | none |
| U255 | 2491 | 2463 | or | False | none |
| U256 | 2493 | 2465 | or | True | none |
| U257 | 2507 | 2479 | not | True | none |
| U258 | 2512 | 2484 | assertType | True | none |
| U259 | 2515 | 2487 | singleElementArray | True | none |
| U260 | 2520 | 2492 | enumerateInsertsAndDeletes | True | none |
| U261 | 2559 | 2531 | cartesianProduct | True | none |
| U262 | 2565 | 2537 | cartesianProductWorker | True | none |
| U263 | 2585 | 2557 | takeWhile | False | none |
| U264 | 2587 | 2559 | takeWhile | False | none |
| U265 | 2588 | 2560 | takeWhile | True | F42 |
| U266 | 2600 | 2572 | skipWhile | False | none |
| U267 | 2602 | 2574 | skipWhile | False | none |
| U268 | 2604 | 2576 | skipWhile | True | none |

## Requested functions

Every requested name is included, including all overload units. For these names, `none` explicitly means no site finding recorded, with deferred coverage interpreted as above.

| Function | Census units and statuses | Recorded site findings |
| --- | --- | --- |
| arrayOf | core.ts:1334:1 deferred_generic_instantiation (instantiated=false) | F2 |
| arrayFrom | core.ts:1347:1 deferred_generic_instantiation (instantiated=false); core.ts:1349:1 deferred_generic_instantiation (instantiated=false); core.ts:1351:1 deferred_generic_instantiation (instantiated=false) | none |
| arrayToMap | core.ts:1411:1 deferred_generic_instantiation (instantiated=false); core.ts:1413:1 deferred_generic_instantiation (instantiated=false); core.ts:1415:1 deferred_generic_instantiation (instantiated=false); core.ts:1417:1 deferred_generic_instantiation (instantiated=false); core.ts:1419:1 deferred_generic_instantiation (instantiated=false) | none |
| arrayToMultiMap | core.ts:1444:1 deferred_generic_instantiation (instantiated=false); core.ts:1446:1 deferred_generic_instantiation (instantiated=false); core.ts:1448:1 deferred_generic_instantiation (instantiated=false) | none |
| group | core.ts:1458:1 deferred_generic_instantiation (instantiated=false); core.ts:1460:1 deferred_generic_instantiation (instantiated=false); core.ts:1462:1 deferred_generic_instantiation (instantiated=false); core.ts:1464:1 deferred_generic_instantiation (instantiated=false); core.ts:1466:1 deferred_generic_instantiation (instantiated=false) | none |
| groupBy | core.ts:1471:1 deferred_generic_instantiation (instantiated=false); core.ts:1473:1 deferred_generic_instantiation (instantiated=false); core.ts:1474:1 deferred_generic_instantiation (instantiated=false) | none; F4 |
| extend | core.ts:1505:1 deferred_generic_instantiation (instantiated=false) | F6, F7, F8, F9, F10 |
| copyProperties | core.ts:1523:1 deferred_generic_instantiation (instantiated=false) | F11 |
| assign | core.ts:1360:1 deferred_generic_instantiation (instantiated=false) | F3 |
| maybeBind | core.ts:1532:1 deferred_generic_instantiation (instantiated=false) | none |
| createMultiMap | core.ts:1552:1 deferred_generic_instantiation (instantiated=false) | F12, F13, F14 |
| multiMapAdd | core.ts:1558:1 deferred_generic_instantiation (instantiated=false) | none |
| multiMapRemove | core.ts:1568:1 deferred_generic_instantiation (instantiated=false) | none |
| createQueue | core.ts:1579:1 deferred_generic_instantiation (instantiated=false) | F15, F16 |
| createSet | core.ts:1647:1 deferred_generic_instantiation (instantiated=false) | none |
| toArray | core.ts:1781:1 deferred_generic_instantiation (instantiated=false); core.ts:1783:1 deferred_generic_instantiation (instantiated=false); core.ts:1785:1 deferred_generic_instantiation (instantiated=false) | none |
| cast | core.ts:1808:1 deferred_generic_instantiation (instantiated=false) | none |
| memoizeOne | core.ts:1932:1 deferred_generic_instantiation (instantiated=false) | none |
| levenshteinWithMax | core.ts:2223:1 finding | F22, F23, F24, F25, F26, F27, F28, F29, F30, F31, F32, F33, F34, F35, F36 |
| removeMinAndVersionNumbers | core.ts:2296:1 lowered | none |
| createGetCanonicalFileName | core.ts:2402:1 lowered | none |
| orderedRemoveItem | core.ts:2350:1 deferred_generic_instantiation (instantiated=false) | none |
| or | core.ts:2484:1 deferred_generic_instantiation (instantiated=false); core.ts:2486:1 deferred_generic_instantiation (instantiated=false); core.ts:2489:1 deferred_generic_instantiation (instantiated=false); core.ts:2491:1 deferred_generic_instantiation (instantiated=false); core.ts:2493:1 deferred_generic_instantiation (instantiated=false) | F37, F38; F39, F40, F41; none |
| not | core.ts:2507:1 deferred_generic_instantiation (instantiated=false) | none |
| takeWhile | core.ts:2585:1 deferred_generic_instantiation (instantiated=false); core.ts:2587:1 deferred_generic_instantiation (instantiated=false); core.ts:2588:1 deferred_generic_instantiation (instantiated=false) | none; F42 |
| enumerateInsertsAndDeletes | core.ts:2520:1 deferred_generic_instantiation (instantiated=false) | none |
| cartesianProductWorker | core.ts:2565:1 deferred_generic_instantiation (instantiated=false) | none |
| isNodeLikeSystem | core.ts:2616:1 finding | F45, F46, F47, F48 |

## Reproduction and retained evidence

The setup follows 99w PROOF.md and the full census follows 99s CENSUS.md. Commands and process limit are retained in [run-census.py](run-census.py), [prepare.sh](prepare.sh), and the setup/apply logs. All command output went to files. The initial overlay attempt ran before the compiler submodule was available and failed; the retry after successful setup built the binary used here. Scratch Go files were renamed to `.go.txt`, with overlay replacements updated before formatting and building. No production compiler source was changed. The exact environment flags, tool hash, source hashes before/after census, and process-exit check are retained in provenance and validation.json.

The post-apply checkpoint was pushed to `typescript/core-sites-checkpoint` so the final `typescript/core-sites` branch can contain one commit on main without any force-push. Its SHA and push result are retained in provenance. Only this report directory is part of the final commit.

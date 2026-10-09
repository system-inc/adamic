# Partition 32: what each file needs to reach zero

Observed on the final wave tree with adaptations 00, 10, 30 and 32, without adaptation 20. The census retains all 78 compiler roots and unchanged checker options. Counts are observations; the proposed repairs below are inferences, not implemented or proven fixes. Each repair needs a new census and behavior proof, and resolving a diagnostic may expose another.

Four of the 26 reviewed files already have zero full-checker findings: builderPublic.ts, builderStatePublic.ts, symbolWalker.ts and performance.ts. This wave adds no new zero file. The remaining 63 findings in the five requested codes are only part of the 441 full-checker findings in this partition. Zero in those five codes alone does not count toward the 40-of-78 target.

## File-by-file completion work

Optional contracts belong to the separately landing adaptation 20, which remains absent from this proof. Iterator helpers are declared in core.ts, owned by adaptation 30; their generic inference needs coordination with that owner or a separate consumer-contract adaptation. No other worker file was edited. Enum/namespace, return-completion, fallthrough, catch, and Node host work needs separately reviewed behavior-preserving adaptations; checker options must stay fixed. Required-read assertions cannot repair these contracts.

| File | Five requested codes remaining | All checker findings | Observed code counts | Proposed work to reach zero |
| --- | ---: | ---: | --- | --- |
| corePublic.ts | 0 | 1 | TS1294: 1 | Provide a proven erasable representation of Comparison preserving its constants and public use. |
| performanceCore.ts | 0 | 3 | TS2375: 1, TS2591: 2 | Expose the existing Node require/perf_hooks host contracts to the meter; permit explicit undefined on PerformanceHooks.performance and performanceTime through the separate optional-declaration adaptation. |
| types.ts | 0 | 73 | TS1294: 73 | A separate enum adaptation must preserve values, reverse mappings, const-enum references, and public declaration shape for all 73 enums. The diagnostic-category reverse lookup must remain valid. |
| expressionToTypeNode.ts | 0 | 8 | TS2412: 3, TS7030: 5 | Optional-declaration contracts for three modifier/flag writes; preserve implicit undefined completions in five return-path sites. |
| executeCommandLine.ts | 1 | 8 | TS1294: 1, TS18048: 1, TS2375: 1, TS2379: 2, TS2488: 2, TS7030: 1 | Repair reduceLeftIterator yield inference and both destructured iterator-entry callback contracts; handle three optional-object/argument contracts, one return completion, and one enum. |
| builder.ts | 12 | 50 | TS1294: 2, TS2345: 12, TS2375: 7, TS2379: 1, TS2412: 22, TS2488: 2, TS2769: 1, TS7030: 3 | Repair Path yield inference in iterator helpers, destructured entries, and Map construction; handle 30 optional-object/argument/write contracts, three return completions, and two enums. |
| builderPublic.ts | 0 | 0 | none | Already zero; no additional change required in this census. |
| builderState.ts | 2 | 6 | TS1294: 2, TS2322: 1, TS2345: 1, TS2412: 2 | Repair Path iterator yield inference, handle two optional cached-array writes, and preserve the BuilderState namespace plus its enum with an erasable representation. |
| builderStatePublic.ts | 0 | 0 | none | Already zero; no additional change required in this census. |
| watch.ts | 2 | 12 | TS2322: 2, TS2375: 3, TS2412: 5, TS7029: 1, TS7030: 1 | Express the existing first-file diagnostic argument correlation without asserting the empty-file path; handle eight optional contracts, one return completion, and one intentional fallthrough. |
| watchPublic.ts | 0 | 19 | TS2379: 3, TS2412: 15, TS7030: 1 | Handle 18 optional host/watcher/input contracts and the scheduleInvalidateResolutionsOfFailedLookupLocations return completion at line 862. |
| watchUtilities.ts | 0 | 6 | TS1294: 2, TS2375: 1, TS2556: 1, TS2684: 2 | Correlate the selected watcher method with its argument/callback tuple for two call signatures and one spread; handle the cached-directory host optional contract and two enums. |
| resolutionCache.ts | 7 | 21 | TS18048: 2, TS2322: 2, TS2345: 3, TS2375: 2, TS2379: 3, TS2412: 8, TS7030: 1 | Repair iterator Path inference; handle 13 optional contracts and rerun to check the two watcher-result cascade findings; preserve the return completion at line 1508. |
| tsbuild.ts | 0 | 1 | TS1294: 1 | Provide an erasable representation of UpToDateStatusType preserving all values. |
| tsbuildPublic.ts | 4 | 21 | TS1294: 3, TS18048: 3, TS2345: 1, TS2379: 1, TS2412: 11, TS7029: 1, TS7030: 1 | Handle 12 optional contracts and rerun to check three emit-data cascade findings; repair forEachKey yield inference, three enums, one return completion, and one intentional fallthrough. |
| moduleNameResolver.ts | 7 | 38 | TS1294: 2, TS2322: 4, TS2345: 3, TS2375: 5, TS2412: 3, TS7030: 21 | Handle ten optional contracts and rerun to check four resolution-local cascade findings; represent an absent typesVersions entry in VersionPaths and its existing validator; preserve 21 return completions and two enums. |
| moduleSpecifiers.ts | 2 | 11 | TS1294: 2, TS2322: 1, TS2345: 1, TS2375: 1, TS2488: 1, TS7030: 5 | Make the cached-field tuple admit its existing explicit undefined values; preserve Debug.assertNever handling of an absent ending; repair iterator-entry inference, one optional object, five return completions, and two enums. |
| sys.ts | 2 | 61 | TS1294: 3, TS18046: 2, TS2304: 6, TS2322: 1, TS2345: 1, TS2375: 1, TS2591: 44, TS7006: 1, TS7029: 1, TS7031: 1 | Expose the existing Node globals/modules and their callback signatures (44 TS2591, six TS2304, one TS7006, one TS7031, plus process.execArgv inference); represent caught-error/ModuleImportResult contracts without changing catches; handle one optional object, one intentional fallthrough, and three enums. |
| program.ts | 6 | 38 | TS1294: 2, TS18046: 2, TS2345: 6, TS2375: 10, TS2379: 2, TS2412: 8, TS2488: 1, TS7029: 3, TS7030: 4 | Repair alias-based typeof narrowing and allow the existing malformed-substitution diagnostic value; prove realPath assignment; handle 20 optional contracts, two caught-error contracts, one iterator tuple, four return completions, three intentional fallthroughs, and two enums. |
| commandLineParser.ts | 8 | 25 | TS18046: 1, TS2322: 4, TS2345: 4, TS2375: 6, TS2379: 1, TS2412: 3, TS7029: 1, TS7030: 5 | Repair mapper/iterator inference, handled absent numeric CLI argument typing, Debug.fail/assert narrowing, and ProjectReference.originalPath; handle ten other optional contracts, one caught-error contract, five return completions, and one intentional fallthrough. |
| binder.ts | 0 | 18 | TS1294: 2, TS2412: 4, TS7029: 7, TS7030: 5 | Handle four optional writes, five return completions, seven intentional fallthroughs, and two enums. |
| semver.ts | 10 | 10 | TS18048: 1, TS2345: 9 | Prove mandatory major captures in both successful regexp destructurings and carry the type to their property uses; make parseComparator.operator admit its existing optional capture; repair split/for-of string element inference. |
| programDiagnostics.ts | 0 | 1 | TS2412: 1 | Handle the DiagnosticMessageChain.next explicit undefined write at line 292 through the separate optional-declaration adaptation. |
| tracing.ts | 0 | 10 | TS1294: 2, TS18046: 1, TS2379: 2, TS2412: 1, TS2591: 4 | Expose four existing Node host bindings, represent the caught-error contract, handle three optional contracts, and preserve two enum/namespace declarations. |
| symbolWalker.ts | 0 | 0 | none | Already zero; no additional change required in this census. |
| performance.ts | 0 | 0 | none | Already zero; no additional change required in this census. |

## Highest priority completion candidates

programDiagnostics.ts has one optional-property write and is the smallest contract-only candidate for adaptation 20. tsbuild.ts and corePublic.ts each have one enum blocker; a proven enum adaptation could finish them. performanceCore.ts has three blockers across optional declarations and Node host bindings. semver.ts has ten findings but no other-code blockers: capture/destructuring, optional operator, and string iteration contracts are sufficient categories to investigate. expressionToTypeNode.ts needs three optional writes and five return completions; watchPublic.ts needs 18 optional contracts and one return completion. These are candidates, not claimed zero files.

For cascade findings in moduleNameResolver.ts, resolutionCache.ts and tsbuildPublic.ts, repair the earlier optional assignment before choosing a local assertion. The existing error path may explain the apparent optional local; rerunning the meter is necessary to confirm.

## Every remaining finding in the five requested codes

Locations refer to the final adapted source. Full diagnostic chains are in declined-findings.json; every other-code finding and reason is in other-findings.json. No additional indexed assertion is justified by these findings. The file table above gives the proposed repair for each group.

### builder.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 395:85 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 395:118 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 991:17 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1246:69 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1345:64 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1347:41 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1347:97 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1493:60 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1495:26 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1496:45 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 1562:64 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 2379:45 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |

### builderState.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 493:108 | TS2345 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |
| 512:23 | TS2322 | The iterator helper callback infers a possibly undefined Path key/value; this is a callback or iterator-result contract without an element-access expression. |

### commandLineParser.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 1874:96 | TS2345 | v is a mapDefined callback parameter; this is a mapper-inference finding, not an indexed read. |
| 2077:79 | TS2345 | args[i] may be absent after the missing-argument diagnostic; parseInt(undefined) deliberately computes NaN on Node. A required-value assertion would change this handled-absence path. |
| 2677:9 | TS2322 | The mapped ProjectReference writes originalPath: undefined; this is an exact optional-property contract, not an indexed read. |
| 2689:66 | TS2345 | Set inference from optionMap.keys includes undefined under the checker iterator model; this is not an indexed expression. |
| 2952:13 | TS2322 | getNameOfCompilerOptionValue with the existing Debug.fail fallback has a narrowing/overload result finding; no indexed expression is involved. |
| 3338:9 | TS2345 | The existing Debug.assert string contract is not carried to spec by the checker; spec is a parameter, not an indexed read. |
| 3449:55 | TS2322 | arrayFrom iterator-overload inference includes undefined for extendedSourceFiles keys; no indexed expression is involved. |
| 4014:5 | TS2322 | The concatenated arrayFrom map-value results inherit iterator-overload inference; no indexed expression is involved. |

### executeCommandLine.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 1180:105 | TS18048 | The reduceLeftIterator callback receives a possibly undefined numeric value from its iterator-helper overload; this is a callback parameter contract, not an indexed read. |

### moduleNameResolver.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 286:9 | TS2322 | The resolved-module object forwards optional properties such as originalPath; this is an exact optional-property contract, not an indexed read. |
| 474:36 | TS2322 | The selected typesVersions entry is deliberately validated by readPackageJsonTypesVersionPaths with typeof; missing or malformed paths are diagnosed, so absence cannot be asserted. |
| 642:143 | TS2345 | The local resolution result remains optional after an earlier exact optional object assignment error; this is a control-flow/local finding, not an indexed read. |
| 644:144 | TS2345 | The local resolution result remains optional after an earlier exact optional object assignment error; this is a control-flow/local finding, not an indexed read. |
| 647:35 | TS2345 | The local resolution result remains optional after an earlier exact optional object assignment error; this is a control-flow/local finding, not an indexed read. |
| 648:5 | TS2322 | The local resolution result remains optional after an earlier exact optional object assignment error; this is a control-flow/local finding, not an indexed read. |
| 2772:21 | TS2322 | The returned resolution object forwards optional originalPath and packageId; this is an exact optional-property contract, not an indexed read. |

### moduleSpecifiers.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 349:5 | TS2322 | The tuple stores optional cached fields as explicit undefined; this is an exact optional tuple contract, not a required indexed read. |
| 1411:38 | TS2345 | The switch default explicitly fails through Debug.assertNever when the first ending is absent or invalid; preserving that failure handling does not permit asserting it here. |

### program.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 4157:64 | TS2345 | The substitution is deliberately validated with typeof and diagnosed if absent or malformed; these alias-narrowing and diagnostic-argument findings do not justify asserting the optional indexed substitution. |
| 4158:160 | TS2345 | The substitution is deliberately validated with typeof and diagnosed if absent or malformed; these alias-narrowing and diagnostic-argument findings do not justify asserting the optional indexed substitution. |
| 4160:69 | TS2345 | The substitution is deliberately validated with typeof and diagnosed if absent or malformed; these alias-narrowing and diagnostic-argument findings do not justify asserting the optional indexed substitution. |
| 4160:95 | TS2345 | The substitution is deliberately validated with typeof and diagnosed if absent or malformed; these alias-narrowing and diagnostic-argument findings do not justify asserting the optional indexed substitution. |
| 4165:158 | TS2345 | The substitution is deliberately validated with typeof and diagnosed if absent or malformed; these alias-narrowing and diagnostic-argument findings do not justify asserting the optional indexed substitution. |
| 4988:28 | TS2345 | realPath is a control-flow-assigned local Path, not an indexed expression; this requires a separate definite-assignment/narrowing adaptation. |

### resolutionCache.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 1282:21 | TS2322 | The watcher helper return remains optional after an earlier exact optional object-assignment error; this is a local/helper-return contract, not an indexed read. |
| 1297:17 | TS2322 | The watcher helper return remains optional after an earlier exact optional object-assignment error; this is a local/helper-return contract, not an indexed read. |
| 1616:118 | TS2345 | The firstDefinedIterator callback receives a possibly undefined Path from the iterator helper overload; this is a callback parameter contract, not an indexed read. |
| 1618:39 | TS18048 | The firstDefinedIterator callback receives a possibly undefined Path from the iterator helper overload; this is a callback parameter contract, not an indexed read. |
| 1619:46 | TS2345 | The firstDefinedIterator callback receives a possibly undefined Path from the iterator helper overload; this is a callback parameter contract, not an indexed read. |
| 1619:74 | TS2345 | The firstDefinedIterator callback receives a possibly undefined Path from the iterator helper overload; this is a callback parameter contract, not an indexed read. |
| 1619:99 | TS18048 | The firstDefinedIterator callback receives a possibly undefined Path from the iterator helper overload; this is a callback parameter contract, not an indexed read. |

### semver.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 149:25 | TS2345 | major is a destructured regexp capture, not an element-access expression; its mandatory capture invariant needs a separate adaptation. |
| 287:48 | TS18048 | simple is a split/for-of binding, not an indexed read, under the checker regex/string model. |
| 288:48 | TS2345 | rangeRegExp capture 1 is optional and parseComparator handles undefined; its string parameter declaration does not represent that contract. |
| 302:20 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| 302:42 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| 303:20 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| 304:20 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| 319:21 | TS2345 | leftResult.major inherits the destructured-capture type; this is a property read, not an indexed read. |
| 323:21 | TS2345 | rightResult.major inherits the destructured-capture type; this is a property read, not an indexed read. |
| 339:21 | TS2345 | rightResult.major inherits the destructured-capture type; this is a property read, not an indexed read. |

### sys.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 1594:180 | TS2345 | Missing Node typings leave process.execArgv callback inference as unknown; this is not an indexed expression. |
| 1616:13 | TS2322 | The require wrapper catches unknown and forwards it in ModuleImportResult; this is a catch/error-object contract, not an indexed read. |

### tsbuildPublic.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 1064:76 | TS18048 | The emit-data local remains optional after an earlier exact optional assignment error; this is a local control-flow contract, not an indexed read. |
| 1065:62 | TS18048 | The emit-data local remains optional after an earlier exact optional assignment error; this is a local control-flow contract, not an indexed read. |
| 1066:42 | TS18048 | The emit-data local remains optional after an earlier exact optional assignment error; this is a local control-flow contract, not an indexed read. |
| 1776:54 | TS2345 | The forEachKey callback receives a possibly undefined string from its iterator overload; no element-access expression exists. |

### watch.ts

| Location | Code | Reviewed reason |
| --- | --- | --- |
| 291:9 | TS2322 | The diagnostic-argument tuple contains the optional firstFileReference produced by the explicit empty-file guard; this is an optional diagnostic-argument contract. |
| 294:9 | TS2322 | The diagnostic-argument tuple contains the optional firstFileReference produced by the explicit empty-file guard; this is an optional diagnostic-argument contract. |


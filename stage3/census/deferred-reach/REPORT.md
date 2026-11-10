# Deferred generic reach (abrq61p)

The tooling is complete. Core.ts is fully classified; across all files, 501 of 504 deferred declarations fit the five requested classes. The exceptional toSearchResult call stops in signature lowering. updateCreateProgram enters its body, then stops on its callback signature. createComputedCompilerOptions enters and finishes its body while instantiated stays false because no concrete argument is inferred. These three rows remain explicitly unclassified rather than being mislabeled as calls hidden behind earlier caller stops.

## Scope and denominators

The baseline is e07f4f06, measured on main production compiler bdbc73864f82dbc0c51434decf0ffe96d4782767 and TypeScript 6.0.3 source 050880ce59e30b356b686bd3144efe24f875ebc8. The tooling commit is 8fea9f1018dfb124b0c77b4bac702e06f968680f. Inputs, adapters and source hashes are preserved in STEP1. The primary directory process is measured on a checker-rejected program; the independent tsc entry process is checker-clean. These are lowering observations, not a claim that the directory program successfully type-checks.

The primary scope is one first-error process over all 79 compiler files, followed by the daily meter's independent tsc entry pass. Reach flags are read only after all files in each process finish. Both scopes have the same 504 deferred declarations and the same flags. Entry-only additional files contribute no deferred declarations.

| Scope | All units | Signature-only | Units excluding signatures | Deferred | Deferred bodies | Raw true flags | Reached bodies |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core.ts | 264 | 84 | 180 | 207 | 123 | 61 | 61 |
| All 79 compiler files | 4555 | 137 | 4418 | 504 | 367 | 115 | 111 |

Yes: the census includes all 84 core.ts bodyless overload signatures in both its 207 deferred declarations and its 264 units. These have nothing to lower. Signature-only classification takes precedence over the old boolean. Four bodyless signatures elsewhere have true flags because censusImplementation redirects to a nongeneric implementation: factory/utilities.ts:654, utilitiesPublic.ts:786 and :826, and visitorPublic.ts:142. The boolean and its witnesses remain unchanged; they do not become reached generic bodies.

## Classes

| Class | core.ts / 207 | core.ts bodies / 123 | All / 504 | All bodies / 367 |
| --- | ---: | ---: | ---: | ---: |
| reached | 61 | 61 | 111 | 111 |
| signature only, no body | 84 | excluded | 137 | excluded |
| unreached with no call anywhere | 0 | 0 | 2 | 2 |
| unreached, called only from outside the census | 18 | 18 | 28 | 28 |
| unreached hidden behind a census caller's earlier stop | 44 | 44 | 223 | 223 |
| Unclassified: request without concrete witness | 0 | 0 | 3 | 3 |

Reached rows use the existing concrete-context flag. The body-entry exception is kept separate, so a false flag is not asserted to prove that no body was entered. The hidden class is limited to census-owned callers with an observed first stop and no recorded instantiation request for the target. Outside references may coexist with hidden census routes. A name reference alone is not treated as a direct invocation. The two declarations without any checker reference are utilities.ts:1167 insertStatementAfterStandardPrologue and watchPublic.ts:368 createWatchCompilerHost.

## Method and limits

The pinned checker's LanguageService.findReferences indexes all 712 TypeScript files under src, rather than matching text. The retained reference index contains 707 generic declarations, including nested declarations. Call expressions, import aliases, variable aliases and the parser callback dispatch table are resolved through checker symbols. Callback arguments whose checker-resolved implementation has a compiler body are traced through the passing census caller. A separate checker-reference audit verifies 19 callback parameter bindings are invoked or forwarded; callback-bindings.json retains those uses. Host callback registration is recorded as an outside invocation boundary. Generic caller chains are followed backward to independently attempted census roots. Function and method bodies not lowered as census units are outside this census even when their file is among the 79. Outside-only rows list those outside caller files. The stored route is one canonical shortest route per census root; the reference index retains the complete source references and resolved calls.

Every hidden root was replayed with diagnostic signature/body entry and exit tracing. All 112 distinct hidden roots reproduce their original first stop exactly. The caller ledger below gives the original stop, and REPORT.json retains the matching propagated error exits. Lowering prepares nested declarations before statement traversal, so a numerically later signature or body can stop an earlier source call. Numeric line ordering is not used as proof. First stop means the first lowering or deferred finding, excluding continuing refusal scans. First-error stops can prevent a later generic request; reaching a request still does not promise that inference produces a concrete witness or that lowering finishes.

The reference graph establishes syntactic and symbol reach, not runtime execution. The hidden classification identifies blocked census routes, not a guarantee that removing one stop will successfully instantiate every downstream body. Full-mode observations below test a bounded subset.

## Tooling and reader validation

Witnesses contain the independently lowered calling unit, concrete call site and checker-printed type arguments in binder order. Serialized witnesses are sorted deterministically, duplicates removed, and only the first five retained. Type printing happens after findings and relation reports are frozen. The old instantiated hook is unchanged. New state is run-owned, not lowering state, and is named in statecopy/POLICY.md.

The whole statecopy package passes, including TestLatentSpeculativePrivateIRFields. audit_deferred.py checks called, outer, nested and unused cases and the five-witness cap. The dropped-recording mutant exits 1 at the new called-witness assertion; evidence/witness-mutant.md and its log are in the tooling commit. Independent old-binary fixture runs match after stripping only the field.

For both full daily-meter scopes, audit_instantiated_by.py removes only the literal new JSON field without reserializing other bytes. The resulting latent.jsonl is byte-identical to baseline, including old flags, findings, unit ledgers and relation trailers. See evidence/steps2-3/corpus-identity.json for both hashes. Seven actual meter report artifacts are byte-identical with ordinary inputs, timings, labels and readiness probe results held fixed. See meter-identity.json. The pair reader has zero raw-finding and distinct-read deltas in both scopes. Its strict eligibility comparison needed a one-field metadata exclusion and has a regression test that still rejects a real body-span change. That repair is in this report commit so the tooling commit remains confined to stage3/census/latent. The meter reader suite ran 82 tests with 5 skips; the pair suite passed all 19 tests. Other discovered readers consume known fields or compare same-overlay records; the reader search is retained.

## Unclassified observations

These are excluded from the strict hidden count, and included explicitly in every total. No first-stop claim is fabricated.

- `executeCommandLine.ts:981:1` updateCreateProgram: request(s) caller `executeCommandLine.ts:968:1`, call `executeCommandLine.ts:974:5`.
  Caller `executeCommandLine.ts:968:1`: original first stop at `executeCommandLine.ts:983:26`, NotYet: a function returning T.
  Caller `executeCommandLine.ts:992:1`: original first stop at `executeCommandLine.ts:997:5`, NotYet: writing a possibly absent optional own field.

- `moduleNameResolver.ts:3414:1` toSearchResult: request(s) caller `moduleNameResolver.ts:2656:1`, call `moduleNameResolver.ts:2661:16`.
  Caller `moduleNameResolver.ts:1803:1`: original first stop at `moduleNameResolver.ts:1847:9`, NotYet: a value of type any.
  Caller `moduleNameResolver.ts:2619:1`: original first stop at `moduleNameResolver.ts:2620:10`, NotYet: a union or optional field read in a program with record storage.
  Caller `moduleNameResolver.ts:2656:1`: original first stop at `moduleNameResolver.ts:3414:28`, NotYet: a value of type T | undefined.
  Caller `moduleNameResolver.ts:2751:1`: original first stop at `moduleNameResolver.ts:2977:21`, NotYet: an array of any.
  Caller `moduleNameResolver.ts:3006:1`: original first stop at `moduleNameResolver.ts:3056:25`, NotYet: a call through ?. (an optional call).
  Caller `moduleNameResolver.ts:3261:1`: original first stop at `moduleNameResolver.ts:3056:25`, NotYet: a call through ?. (an optional call).
  Caller `moduleNameResolver.ts:3332:1`: original first stop at `moduleNameResolver.ts:3333:10`, NotYet: a dictionary member narrowed to an array without an element-checked view.

- `utilities.ts:9031:1` createComputedCompilerOptions: request(s) caller `utilities.ts:9042:1`, call `utilities.ts:9042:26`.
  Caller `utilities.ts:9042:1`: no first finding.

An isolated replay of utilities.ts:9042 confirms zero findings while calling createComputedCompilerOptions. Diagnostic tracing shows both lowerBody_entry and lowerBody_exit with no error for utilities.ts:9031. Its body was entered and completed despite the false flag. Binder evidence is declared=1, inferred=0, concrete=false. toSearchResult fails in signature lowering before lowerBody entry. updateCreateProgram instead enters lowerBody and fails while lowering its body at the callback signature returning T. REPORT.json retains the selected request traces. The old flag therefore does not exhaustively describe all body entry; it records only the existing concrete-context hook. Thus this case is neither outside-only nor hidden behind an earlier stop. Completing the requested taxonomy requires deciding how to treat signature-stage requests and successful or failed body entry without the old concrete-context flag; these three rows remain unclassified because none of the five definitions applies.

## Full-mode scope

Full mode over all 79 files was not attempted because the documented process exceeds 16 GB. Two bounded processes completed: core.ts alone, and core.ts plus utilities.ts. A broader 32-file attempt was stopped on its time budget and is documented separately below. Each still loads the original compiler program, but only the selected file bodies are attempted, so their flags have narrower reach scope. Full mode also inventories additional nested declarations; its totals cannot be subtracted from the first-error primary totals.

| Full-mode selected files | Deferred | Signature-only | Deferred bodies | Raw true flags | True body flags |
| --- | ---: | ---: | ---: | ---: | ---: |
| full-core: core.ts | 211 | 84 | 127 | 3 | 3 |
| full-core-utilities: core.ts | 211 | 84 | 127 | 35 | 35 |
| full-core-utilities: utilities.ts | 66 | 12 | 54 | 5 | 5 |
| full-core-utilities: total | 277 | 96 | 181 | 40 | 40 |

The two-file full run newly reaches these primary-scope bodies (quotes are full-scope witnesses):

- `core.ts:768:1`: unit `utilities.ts:6092:5`, call `utilities.ts:6099:17`, types `["string"]`; unit `utilities.ts:6092:5`, call `utilities.ts:6112:9`, types `["Diagnostic"]`.
- `core.ts:1527:1`: unit `utilities.ts:9501:1`, call `utilities.ts:9519:74`, types `["Path", "string"]`.
- `utilities.ts:11312:1`: unit `utilities.ts:11425:5`, call `utilities.ts:11439:16`, types `["string"]`.

The core-only full run adds no reach over the primary census. The two-file run exposes insertSorted and createMultiMap in core.ts and evaluatorResult in utilities.ts through independently enumerated nested caller units. It shows additional full-census reach, rather than proving their enclosing first-error roots continued successfully. It does not establish all hidden bodies as reachable.

## Per-file counts

Signature-only is a separate column in every row. Body total excludes it. `Pending` means the explicitly unclassified request cases. Files with zero deferred declarations are listed separately below.

| File | Deferred | Signature only | Bodies | Reached | No reference | Outside only | Hidden | Pending |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core.ts | 207 | 84 | 123 | 61 | 0 | 18 | 44 | 0 |
| builder.ts | 2 | 1 | 1 | 1 | 0 | 0 | 0 | 0 |
| commandLineParser.ts | 12 | 2 | 10 | 6 | 0 | 0 | 4 | 0 |
| emitter.ts | 2 | 0 | 2 | 0 | 0 | 0 | 2 | 0 |
| executeCommandLine.ts | 1 | 0 | 1 | 0 | 0 | 0 | 0 | 1 |
| factory/emitNode.ts | 23 | 0 | 23 | 3 | 0 | 2 | 18 | 0 |
| factory/nodeFactory.ts | 1 | 0 | 1 | 1 | 0 | 0 | 0 | 0 |
| factory/utilities.ts | 8 | 5 | 3 | 1 | 0 | 0 | 2 | 0 |
| factory/utilitiesPublic.ts | 1 | 0 | 1 | 1 | 0 | 0 | 0 | 0 |
| moduleNameResolver.ts | 15 | 2 | 13 | 4 | 0 | 0 | 8 | 1 |
| moduleSpecifiers.ts | 1 | 0 | 1 | 0 | 0 | 0 | 1 | 0 |
| parser.ts | 22 | 0 | 22 | 1 | 0 | 0 | 21 | 0 |
| path.ts | 4 | 2 | 2 | 2 | 0 | 0 | 0 | 0 |
| program.ts | 5 | 0 | 5 | 1 | 0 | 0 | 4 | 0 |
| scanner.ts | 10 | 4 | 6 | 4 | 0 | 0 | 2 | 0 |
| sourcemap.ts | 1 | 0 | 1 | 0 | 0 | 1 | 0 | 0 |
| sys.ts | 2 | 0 | 2 | 0 | 0 | 0 | 2 | 0 |
| transformer.ts | 2 | 0 | 2 | 1 | 0 | 0 | 1 | 0 |
| transformers/classThis.ts | 2 | 1 | 1 | 0 | 0 | 0 | 1 | 0 |
| transformers/namedEvaluation.ts | 2 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| transformers/utilities.ts | 6 | 0 | 6 | 0 | 0 | 0 | 6 | 0 |
| tsbuildPublic.ts | 72 | 2 | 70 | 2 | 0 | 2 | 66 | 0 |
| utilities.ts | 59 | 12 | 47 | 13 | 1 | 5 | 27 | 1 |
| utilitiesPublic.ts | 9 | 4 | 5 | 4 | 0 | 0 | 1 | 0 |
| visitorPublic.ts | 13 | 12 | 1 | 1 | 0 | 0 | 0 | 0 |
| watch.ts | 10 | 0 | 10 | 4 | 0 | 0 | 6 | 0 |
| watchPublic.ts | 7 | 4 | 3 | 0 | 1 | 0 | 2 | 0 |
| watchUtilities.ts | 5 | 0 | 5 | 0 | 0 | 0 | 5 | 0 |

Zero deferred declarations (all class counts and body counts are zero): `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `corePublic.ts`, `debug.ts`, `diagnosticInformationMap.generated.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `hostErrors.ts`, `performance.ts`, `performanceCore.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `semver.ts`, `symbolWalker.ts`, `tracing.ts`, `transformers/classFields.ts`, `transformers/declarations.ts`, `transformers/declarations/diagnostics.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `tsbuild.ts`, `types.ts`.

## Reached bodies and quoted witnesses

Each witness is quoted exactly apart from removing the common absolute source-root prefix. At most five per declaration are retained. REPORT.json preserves the recorded absolute witness strings.

- `core.ts:22:1` length: unit `commandLineParser.ts:2733:1`, call `commandLineParser.ts:2734:10`, types `["string"]`; unit `commandLineParser.ts:4040:1`, call `commandLineParser.ts:4048:10`, types `["string"]`; unit `program.ts:986:1`, call `program.ts:988:9`, types `["ImportAttribute"]`; unit `utilities.ts:4117:1`, call `utilities.ts:4118:12`, types `["Expression"]`; unit `utilities.ts:7615:1`, call `utilities.ts:7616:22`, types `["Declaration"]`.
- `core.ts:33:1` forEach: unit `builder.ts:813:1`, call `builder.ts:817:9`, types `["SourceFile", "boolean"]`; unit `commandLineParser.ts:4231:1`, call `commandLineParser.ts:4232:28`, types `["string[]", "string[]"]`; unit `commandLineParser.ts:4264:1`, call `commandLineParser.ts:4265:28`, types `["string[]", "string[]"]`; unit `moduleNameResolver.ts:3169:1`, call `moduleNameResolver.ts:3177:26`, types `["string", "Resolved"]`; unit `moduleSpecifiers.ts:1020:1`, call `moduleSpecifiers.ts:1105:16`, types `["unknown", "{ moduleFileToTry: string; }"]`.
- `core.ts:67:1` firstDefined: unit `moduleNameResolver.ts:538:1`, call `moduleNameResolver.ts:668:20`, types `["string", "PathAndPackageId"]`; unit `moduleSpecifiers.ts:288:1`, call `moduleSpecifiers.ts:298:12`, types `["ModulePath", "string"]`; unit `moduleSpecifiers.ts:301:1`, call `moduleSpecifiers.ts:313:12`, types `["ModulePath", "string"]`.
- `core.ts:145:1` every: unit `moduleNameResolver.ts:2573:1`, call `moduleNameResolver.ts:2574:12`, types `["string"]`; unit `transformers/utilities.ts:871:1`, call `transformers/utilities.ts:872:12`, types `["ParameterDeclaration"]`; unit `utilities.ts:10758:1`, call `utilities.ts:10759:46`, types `["Expression"]`; unit `utilities.ts:9893:1`, call `utilities.ts:9914:17`, types `["string"]`.
- `core.ts:166:1` find: unit `utilities.ts:10303:1`, call `utilities.ts:10304:12`, types `["Extension"]`; unit `utilities.ts:7624:1`, call `utilities.ts:7625:12`, types `["Extension"]`; unit `utilities.ts:8043:1`, call `utilities.ts:8045:59`, types `["Declaration"]`; unit `utilitiesPublic.ts:1106:1`, call `utilitiesPublic.ts:1280:12`, types `["JSDocTag"]`; unit `utilitiesPublic.ts:1111:1`, call `utilitiesPublic.ts:1280:12`, types `["JSDocTag"]`.
- `core.ts:198:1` findIndex: unit `watch.ts:415:1`, call `watch.ts:421:19`, types `["string"]`; unit `watch.ts:426:1`, call `watch.ts:436:19`, types `["string"]`.
- `core.ts:220:1` contains: unit `core.ts:232:1`, call `core.ts:234:13`, types `["number"]`; unit `emitter.ts:710:1`, call `emitter.ts:712:18`, types `["string"]`; unit `executeCommandLine.ts:478:1`, call `executeCommandLine.ts:492:53`, types `["CommandLineOption"]`; unit `factory/utilities.ts:697:1`, call `core.ts:1014:9`, types `["string"]`; unit `program.ts:1229:1`, call `program.ts:1289:17`, types `["ResolvedProjectReference"]`.
- `core.ts:242:1` countWhere: unit `watch.ts:228:1`, call `watch.ts:229:12`, types `["Diagnostic"]`; unit `watch.ts:303:1`, call `watch.ts:308:64`, types `["ReportFileInError | undefined"]`.
- `core.ts:277:1` filter: unit `commandLineParser.ts:2660:1`, call `commandLineParser.ts:2663:9`, types `["string"]`; unit `commandLineParser.ts:4077:1`, call `commandLineParser.ts:4085:9`, types `["string"]`; unit `emitter.ts:685:1`, call `emitter.ts:688:15`, types `["string"]`; unit `executeCommandLine.ts:176:1`, call `executeCommandLine.ts:177:25`, types `["CommandLineOption"]`; unit `executeCommandLine.ts:513:1`, call `executeCommandLine.ts:516:109`, types `["CommandLineOption"]`.
- `core.ts:299:1` filterMutate: unit `commandLineParser.ts:3377:1`, call `commandLineParser.ts:3389:9`, types `["Diagnostic"]`.
- `core.ts:320:1` map: unit `commandLineParser.ts:2660:1`, call `commandLineParser.ts:2662:19`, types `["string", "string"]`; unit `moduleSpecifiers.ts:1169:1`, call `moduleSpecifiers.ts:1177:16`, types `["string", "string"]`; unit `semver.ts:443:1`, call `semver.ts:444:12`, types `["readonly Comparator[]", "string"]`; unit `semver.ts:447:1`, call `semver.ts:448:12`, types `["Comparator", "string"]`; unit `transformer.ts:127:1`, call `transformer.ts:135:50`, types `["CustomTransformerFactory | TransformerFactory<SourceFile>", "TransformerFactory<Bundle | SourceFile>"]`.
- `core.ts:332:1` mapIterator: unit `commandLineParser.ts:142:1`, call `commandLineParser.ts:142:65`, types `["[string, JsxEmit]", "readonly [string, string]"]`.
- `core.ts:350:1` sameMap: unit `builder.ts:520:1`, call `builder.ts:525:12`, types `["Diagnostic", "Diagnostic"]`.
- `core.ts:375:1` flatten: unit `utilities.ts:9973:1`, call `utilities.ts:9973:64`, types `["Extension"]`; unit `utilities.ts:9979:1`, call `utilities.ts:9979:64`, types `["Extension"]`.
- `core.ts:399:1` flatMap: unit `moduleSpecifiers.ts:1169:1`, call `moduleSpecifiers.ts:1176:27`, types `["string", "string"]`; unit `utilities.ts:6916:1`, call `utilities.ts:6917:12`, types `["JSDocTag", "TypeParameterDeclaration"]`; unit `utilities.ts:9671:1`, call `utilities.ts:9676:12`, types `["string", "string"]`.
- `core.ts:494:1` mapDefined: unit `moduleSpecifiers.ts:1357:1`, call `moduleSpecifiers.ts:1358:12`, types `["string", "string"]`.
- `core.ts:509:1` mapDefinedIterator: unit `builder.ts:2277:1`, call `builder.ts:2403:13`, types `["Path", "[Path, never[]]"]`.
- `core.ts:529:1` tryAddToSet: unit `builder.ts:1027:1`, call `builder.ts:1035:10`, types `["string"]`.
- `core.ts:618:1` some: unit `builder.ts:1219:1`, call `builder.ts:1223:28`, types `["SourceFile"]`; unit `builder.ts:956:1`, call `builder.ts:1004:17`, types `["Declaration"]`; unit `commandLineParser.ts:2716:1`, call `commandLineParser.ts:2722:20`, types `["string"]`; unit `factory/emitNode.ts:289:1`, call `factory/emitNode.ts:292:10`, types `["EmitHelper"]`; unit `factory/utilities.ts:697:1`, call `factory/utilities.ts:718:21`, types `["string"]`.
- `core.ts:672:1` concatenate: unit `builder.ts:1058:1`, call `builder.ts:1065:12`, types `["Diagnostic"]`; unit `moduleNameResolver.ts:755:1`, call `moduleNameResolver.ts:780:12`, types `["string"]`; unit `utilities.ts:10125:1`, call `utilities.ts:10133:24`, types `["RequireOrImportCall"]`; unit `utilities.ts:2761:1`, call `utilities.ts:2769:9`, types `["CommentRange"]`; unit `utilities.ts:5227:1`, call `utilities.ts:5229:29`, types `["ExpressionWithTypeArguments"]`.
- `core.ts:735:1` deduplicateSorted: unit `program.ts:635:1`, call `core.ts:810:12`, types `["Diagnostic"]`; unit `program.ts:636:18`, call `core.ts:810:12`, types `["Diagnostic"]`; unit `program.ts:637:1`, call `core.ts:810:12`, types `["Diagnostic"]`.
- `core.ts:809:1` sortAndDeduplicate: unit `program.ts:635:1`, call `utilitiesPublic.ts:304:12`, types `["Diagnostic"]`; unit `program.ts:636:18`, call `utilitiesPublic.ts:304:12`, types `["Diagnostic"]`; unit `program.ts:637:1`, call `utilitiesPublic.ts:304:12`, types `["Diagnostic"]`.
- `core.ts:814:1` arrayIsEqualTo: unit `program.ts:1229:1`, call `program.ts:1301:18`, types `["string"]`; unit `scanner.ts:475:1`, call `scanner.ts:482:148`, types `["number"]`.
- `core.ts:936:1` append: unit `moduleNameResolver.ts:2228:1`, call `moduleNameResolver.ts:2254:19`, types `["string"]`; unit `moduleSpecifiers.ts:1169:1`, call `core.ts:409:30`, types `["string"]`; unit `transformers/destructuring.ts:244:1`, call `transformers/destructuring.ts:315:30`, types `["Expression"]`; unit `utilities.ts:9671:1`, call `core.ts:409:30`, types `["string"]`.
- `core.ts:975:1` toOffset: unit `moduleSpecifiers.ts:1169:1`, call `core.ts:998:39`, types `["string"]`; unit `program.ts:635:1`, call `core.ts:998:39`, types `["Diagnostic"]`; unit `program.ts:636:18`, call `core.ts:998:39`, types `["Diagnostic"]`; unit `program.ts:637:1`, call `core.ts:998:39`, types `["Diagnostic"]`; unit `transformer.ts:127:1`, call `core.ts:998:39`, types `["TransformerFactory<Bundle | SourceFile>"]`.
- `core.ts:995:1` addRange: unit `moduleSpecifiers.ts:1169:1`, call `core.ts:406:30`, types `["string"]`; unit `program.ts:635:1`, call `program.ts:639:19`, types `["Diagnostic"]`; unit `program.ts:636:18`, call `program.ts:639:19`, types `["Diagnostic"]`; unit `program.ts:637:1`, call `program.ts:639:19`, types `["Diagnostic"]`; unit `transformer.ts:127:1`, call `transformer.ts:135:5`, types `["TransformerFactory<Bundle | SourceFile>"]`.
- `core.ts:1013:1` pushIfUnique: unit `factory/utilities.ts:697:1`, call `factory/utilities.ts:715:25`, types `["string"]`.
- `core.ts:1048:1` toSorted: unit `utilities.ts:9827:1`, call `utilities.ts:9854:31`, types `["string"]`.
- `core.ts:1091:1` firstOrUndefined: unit `factory/utilities.ts:599:1`, call `factory/utilities.ts:600:28`, types `["Statement"]`; unit `utilities.ts:10881:1`, call `utilities.ts:10891:31`, types `["ParameterDeclaration"]`.
- `core.ts:1124:1` lastOrUndefined: unit `path.ts:449:1`, call `path.ts:452:25`, types `["string"]`; unit `utilitiesPublic.ts:2633:1`, call `utilitiesPublic.ts:2634:18`, types `["JSDocParameterTag | ParameterDeclaration"]`.
- `core.ts:1129:1` last: unit `utilities.ts:6296:1`, call `utilities.ts:6312:50`, types `["number"]`.
- `core.ts:1139:1` singleOrUndefined: unit `utilities.ts:2968:1`, call `utilities.ts:2973:16`, types `["TypeNode"]`; unit `visitorPublic.ts:1805:1`, call `visitorPublic.ts:1807:12`, types `["Node"]`.
- `core.ts:1194:1` binarySearch: unit `scanner.ts:520:1`, call `scanner.ts:521:22`, types `["number", "number"]`; unit `utilities.ts:1352:1`, call `utilities.ts:1353:12`, types `["Node", "number"]`.
- `core.ts:1210:1` binarySearchKey: unit `scanner.ts:520:1`, call `core.ts:1195:12`, types `["number", "number"]`; unit `utilities.ts:1352:1`, call `core.ts:1195:12`, types `["Node", "number"]`.
- `core.ts:1240:1` reduceLeft: unit `moduleSpecifiers.ts:1169:1`, call `core.ts:2018:12`, types `["string"]`.
- `core.ts:1274:1` hasProperty: unit `commandLineParser.ts:1717:1`, call `commandLineParser.ts:1717:118`, types `["CommandLineOption"]`; unit `commandLineParser.ts:2988:1`, call `commandLineParser.ts:2993:13`, types `["CompilerOptions"]`; unit `commandLineParser.ts:3278:1`, call `core.ts:1351:17`, types `["OptionsBase"]`; unit `commandLineParser.ts:3336:1`, call `core.ts:1351:17`, types `["MapLike<string[]>"]`; unit `commandLineParser.ts:3366:1`, call `commandLineParser.ts:3368:9`, types `["{ [key: string]: unknown; }"]`.
- `core.ts:1295:1` getOwnKeys: unit `commandLineParser.ts:3336:1`, call `commandLineParser.ts:3338:21`, types `["string[]"]`; unit `moduleNameResolver.ts:2573:1`, call `moduleNameResolver.ts:2574:18`, types `["unknown"]`; unit `moduleNameResolver.ts:2577:1`, call `moduleNameResolver.ts:2578:18`, types `["unknown"]`.
- `core.ts:1338:1` arrayFrom: unit `builder.ts:1258:1`, call `builder.ts:1421:25`, types `["Path", "IncrementalBuildInfoFileId"]`; unit `commandLineParser.ts:1873:1`, call `commandLineParser.ts:1874:25`, types `["string", "string"]`; unit `executeCommandLine.ts:244:1`, call `executeCommandLine.ts:313:15`, types `["[string, string | number]", "[string, string | number]"]`; unit `utilities.ts:1210:1`, call `utilities.ts:1223:16`, types `["[string, CommentDirective]", "[string, CommentDirective]"]`.
- `core.ts:1347:1` assign: unit `commandLineParser.ts:3278:1`, call `commandLineParser.ts:3313:21`, types `["OptionsBase"]`; unit `commandLineParser.ts:3336:1`, call `commandLineParser.ts:3343:21`, types `["MapLike<string[]>"]`.
- `core.ts:1367:1` equalOwnProperties: unit `utilities.ts:10451:1`, call `utilities.ts:10453:101`, types `["unknown"]`.
- `core.ts:1406:1` arrayToMap: unit `commandLineParser.ts:2307:1`, call `commandLineParser.ts:2308:12`, types `["string", "CommandLineOption", "CommandLineOption"]`.
- `core.ts:1534:1` multiMapAddValue: unit `sys.ts:388:1`, call `sys.ts:403:13`, types `["string", "FileWatcherCallback"]`.
- `core.ts:1546:1` multiMapRemoveValue: unit `sys.ts:388:1`, call `sys.ts:419:17`, types `["string", "FileWatcherCallback"]`.
- `core.ts:1557:1` createQueue: unit `binder.ts:3789:1`, call `binder.ts:3791:15`, types `["Expression"]`.
- `core.ts:1781:1` tryCast: unit `utilities.ts:3104:1`, call `utilities.ts:3107:16`, types `["ObjectLiteralExpression", "ArrayLiteralExpression | FalseLiteral | JsonMinusNumericLiteral | NullLiteral | NumericLiteral | ObjectLiteralExpression | StringLiteral | TrueLiteral"]`; unit `utilities.ts:4425:1`, call `utilities.ts:4428:41`, types `["NamespaceImport", "NamedImports | NamespaceImport"]`; unit `utilitiesPublic.ts:865:1`, call `utilitiesPublic.ts:867:20`, types `["KeywordSyntaxKind", "SyntaxKind.EndOfFileToken | SyntaxKind.SingleLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia | SyntaxKind.NewLineTrivia | SyntaxKind.WhitespaceTrivia | SyntaxKind.ShebangTrivia | SyntaxKind.ConflictMarkerTrivia | ... 351 more ... | SyntaxKind.Count"]`.
- `core.ts:1786:1` cast: unit `commandLineParser.ts:2207:1`, call `commandLineParser.ts:2208:12`, types `["string", "string | DiagnosticMessageChain"]`; unit `factory/parenthesizerRules.ts:675:1`, call `factory/parenthesizerRules.ts:684:48`, types `["LeftHandSideExpression", "Expression"]`; unit `transformers/utilities.ts:179:1`, call `transformers/utilities.ts:304:33`, types `["NamedExports", "NamedExports | NamespaceExport"]`.
- `core.ts:1831:1` identity: unit `core.ts:2380:1`, call `core.ts:2381:40`, types `["string"]`; unit `core.ts:2445:1`, call `core.ts:2445:107`, types `["string"]`; unit `factory/parenthesizerRules.ts:675:1`, call `factory/parenthesizerRules.ts:676:54`, types `["Expression"]`; unit `scanner.ts:520:1`, call `scanner.ts:521:57`, types `["number"]`; unit `sourcemap.ts:823:1`, call `sourcemap.ts:824:24`, types `["DocumentPosition"]`.
- `core.ts:1894:1` memoize: unit `emitter.ts:1200:1`, call `emitter.ts:1200:73`, types `["Printer"]`; unit `emitter.ts:1203:1`, call `emitter.ts:1203:79`, types `["Printer"]`; unit `emitter.ts:1206:1`, call `emitter.ts:1206:95`, types `["Printer"]`; unit `emitter.ts:1209:1`, call `emitter.ts:1209:100`, types `["Printer"]`; unit `emitter.ts:695:1`, call `emitter.ts:701:42`, types `["string"]`.
- `core.ts:1910:1` memoizeOne: unit `moduleSpecifiers.ts:133:1`, call `moduleSpecifiers.ts:133:23`, types `["string", "RegExp | undefined"]`.
- `core.ts:1940:1` equateValues: unit `core.ts:1969:1`, call `core.ts:1970:12`, types `["string"]`; unit `core.ts:232:1`, call `core.ts:220:112`, types `["number"]`; unit `emitter.ts:710:1`, call `core.ts:220:112`, types `["string"]`; unit `executeCommandLine.ts:478:1`, call `core.ts:220:112`, types `["CommandLineOption"]`; unit `factory/utilities.ts:697:1`, call `core.ts:220:112`, types `["string"]`.
- `core.ts:2005:1` maxBy: unit `watch.ts:303:1`, call `watch.ts:309:23`, types `["readonly [ReportFileInError | undefined, number]"]`.
- `core.ts:2017:1` min: unit `moduleSpecifiers.ts:1169:1`, call `moduleSpecifiers.ts:1179:22`, types `["string"]`.
- `core.ts:2328:1` orderedRemoveItem: unit `factory/emitNode.ts:271:1`, call `factory/emitNode.ts:274:16`, types `["EmitHelper"]`.
- `core.ts:2343:1` orderedRemoveItemAt: unit `factory/emitNode.ts:271:1`, call `core.ts:2331:13`, types `["EmitHelper"]`.
- `core.ts:2351:1` unorderedRemoveItemAt: unit `sys.ts:388:1`, call `core.ts:2370:13`, types `["FileWatcherCallback"]`.
- `core.ts:2362:1` unorderedRemoveItem: unit `sys.ts:388:1`, call `core.ts:1549:9`, types `["FileWatcherCallback"]`.
- `core.ts:2367:1` unorderedRemoveFirstItemWhere: unit `sys.ts:388:1`, call `core.ts:2363:12`, types `["FileWatcherCallback"]`.
- `core.ts:2415:1` findBestPatternMatch: unit `utilities.ts:10325:1`, call `utilities.ts:10336:12`, types `["Pattern"]`.
- `core.ts:2457:1` and: unit `checker.ts:1413:1`, call `checker.ts:1413:37`, types `["Declaration"]`.
- `core.ts:2474:1` or: unit `utilities.ts:10054:1`, call `utilities.ts:10054:101`, types `["[fileName: string]", "boolean"]`.
- `core.ts:2493:1` assertType: unit `utilities.ts:12031:1`, call `utilities.ts:12051:13`, types `["SyntaxKind.Unknown | SyntaxKind.EndOfFileToken | SyntaxKind.SingleLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia | SyntaxKind.NewLineTrivia | SyntaxKind.WhitespaceTrivia | SyntaxKind.ShebangTrivia | ... 345 more ... | SyntaxKind.Count"]`; unit `utilities.ts:12065:1`, call `utilities.ts:12082:13`, types `["SyntaxKind.Unknown | SyntaxKind.EndOfFileToken | SyntaxKind.SingleLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia | SyntaxKind.NewLineTrivia | SyntaxKind.WhitespaceTrivia | SyntaxKind.ShebangTrivia | ... 339 more ... | SyntaxKind.Count"]`; unit `utilities.ts:2245:1`, call `utilities.ts:2268:13`, types `["SyntaxKind.Unknown | SyntaxKind.EndOfFileToken | SyntaxKind.SingleLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia | SyntaxKind.NewLineTrivia | SyntaxKind.WhitespaceTrivia | SyntaxKind.ShebangTrivia | ... 333 more ... | SyntaxKind.Count"]`.
- `builder.ts:534:1` convertOrRepopulateDiagnosticMessageChain: unit `builder.ts:618:1`, call `builder.ts:633:13`, types `["ReusableDiagnosticMessageChain"]`.
- `commandLineParser.ts:3077:1` parseJsonConfigFileContentWorker: unit `commandLineParser.ts:3031:1`, call `commandLineParser.ts:3032:12`, types `["unknown"]`; unit `commandLineParser.ts:3042:1`, call `commandLineParser.ts:3044:20`, types `["undefined"]`.
- `commandLineParser.ts:3423:1` parseConfig: unit `commandLineParser.ts:3707:1`, call `commandLineParser.ts:3726:30`, types `["undefined"]`.
- `commandLineParser.ts:3774:1` convertCompilerOptionsFromJsonWorker: unit `commandLineParser.ts:3755:1`, call `commandLineParser.ts:3757:21`, types `["unknown"]`.
- `commandLineParser.ts:3787:1` convertTypeAcquisitionFromJsonWorker: unit `commandLineParser.ts:3761:1`, call `commandLineParser.ts:3763:21`, types `["unknown"]`.
- `commandLineParser.ts:3799:1` convertOptionsFromJson: unit `commandLineParser.ts:3755:1`, call `commandLineParser.ts:3776:5`, types `["unknown"]`; unit `commandLineParser.ts:3761:1`, call `commandLineParser.ts:3789:5`, types `["unknown"]`.
- `commandLineParser.ts:3868:1` validateJsonOptionValue: unit `commandLineParser.ts:1885:1`, call `commandLineParser.ts:1891:16`, types `["string"]`; unit `commandLineParser.ts:2040:1`, call `commandLineParser.ts:2056:37`, types `["false"]`.
- `factory/emitNode.ts:92:1` setEmitFlags: unit `factory/emitHelpers.ts:148:1`, call `factory/emitHelpers.ts:203:16`, types `["Identifier"]`; unit `transformers/es2017.ts:1059:1`, call `transformers/es2017.ts:1075:17`, types `["PropertyAccessExpression"]`; unit `transformers/es2017.ts:1059:1`, call `transformers/es2017.ts:1077:25`, types `["SuperExpression"]`; unit `visitorPublic.ts:474:1`, call `visitorPublic.ts:479:13`, types `["Block"]`.
- `factory/emitNode.ts:102:1` addEmitFlags: unit `utilities.ts:12420:1`, call `utilities.ts:12421:5`, types `["Node"]`.
- `factory/emitNode.ts:175:1` setStartsOnNewLine: unit `factory/utilities.ts:247:1`, call `factory/utilities.ts:679:12`, types `["Expression"]`; unit `factory/utilities.ts:280:1`, call `factory/utilities.ts:679:12`, types `["Expression"]`.
- `factory/nodeFactory.ts:7420:1` setOriginalNode: unit `factory/utilities.ts:411:1`, call `factory/utilities.ts:412:12`, types `["AssignmentExpression<EqualsToken>"]`; unit `factory/utilities.ts:424:1`, call `factory/utilities.ts:425:12`, types `["AssignmentExpression<EqualsToken>"]`; unit `factory/utilities.ts:437:1`, call `factory/utilities.ts:438:12`, types `["AssignmentExpression<EqualsToken>"]`.
- `factory/utilities.ts:678:1` startOnNewLine: unit `factory/utilities.ts:247:1`, call `factory/utilities.ts:260:17`, types `["Expression"]`; unit `factory/utilities.ts:280:1`, call `factory/utilities.ts:287:17`, types `["Expression"]`.
- `factory/utilitiesPublic.ts:10:1` setTextRange: unit `factory/emitHelpers.ts:148:1`, call `factory/emitHelpers.ts:242:16`, types `["CallExpression"]`; unit `transformers/destructuring.ts:446:1`, call `transformers/destructuring.ts:453:13`, types `["Expression"]`; unit `visitorPublic.ts:420:1`, call `visitorPublic.ts:431:16`, types `["NodeArray<ParameterDeclaration>"]`; unit `visitorPublic.ts:474:1`, call `visitorPublic.ts:480:17`, types `["Block"]`.
- `moduleNameResolver.ts:361:1` readPackageJsonField: unit `moduleNameResolver.ts:2421:1`, call `moduleNameResolver.ts:2422:30`, types `["\"peerDependencies\""]`; unit `moduleNameResolver.ts:397:1`, call `moduleNameResolver.ts:380:22`, types `["\"typings\""]`; unit `moduleNameResolver.ts:402:1`, call `moduleNameResolver.ts:380:22`, types `["\"tsconfig\""]`; unit `moduleNameResolver.ts:406:1`, call `moduleNameResolver.ts:380:22`, types `["\"main\""]`; unit `moduleNameResolver.ts:410:1`, call `moduleNameResolver.ts:411:27`, types `["\"typesVersions\""]`.
- `moduleNameResolver.ts:379:1` readPackageJsonPathField: unit `moduleNameResolver.ts:397:1`, call `moduleNameResolver.ts:398:12`, types `["\"typings\""]`; unit `moduleNameResolver.ts:402:1`, call `moduleNameResolver.ts:403:12`, types `["\"tsconfig\""]`; unit `moduleNameResolver.ts:406:1`, call `moduleNameResolver.ts:407:12`, types `["\"main\""]`.
- `moduleNameResolver.ts:1276:1` createModuleOrTypeReferenceResolutionCache: unit `moduleNameResolver.ts:1327:1`, call `moduleNameResolver.ts:1348:20`, types `["ResolvedModuleWithFailedLookupLocations"]`; unit `moduleNameResolver.ts:1334:1`, call `moduleNameResolver.ts:1348:20`, types `["ResolvedModuleWithFailedLookupLocations"]`; unit `moduleNameResolver.ts:1341:1`, call `moduleNameResolver.ts:1348:20`, types `["ResolvedModuleWithFailedLookupLocations"]`; unit `moduleNameResolver.ts:1360:1`, call `moduleNameResolver.ts:1381:12`, types `["ResolvedTypeReferenceDirectiveWithFailedLookupLocations"]`; unit `moduleNameResolver.ts:1367:1`, call `moduleNameResolver.ts:1381:12`, types `["ResolvedTypeReferenceDirectiveWithFailedLookupLocations"]`.
- `moduleNameResolver.ts:3051:1` forEachAncestorDirectoryStoppingAtGlobalCache: unit `moduleNameResolver.ts:2399:1`, call `moduleNameResolver.ts:2400:12`, types `["PackageJsonInfo", "string"]`; unit `moduleNameResolver.ts:3006:1`, call `moduleNameResolver.ts:3029:16`, types `["{ readonly value: Resolved | undefined; }", "string"]`; unit `moduleNameResolver.ts:3261:1`, call `moduleNameResolver.ts:3305:30`, types `["{ readonly value: Resolved | undefined; }", "string"]`; unit `moduleNameResolver.ts:787:1`, call `moduleNameResolver.ts:796:12`, types `["PackageJsonInfo", "string"]`; unit `moduleSpecifiers.ts:698:1`, call `moduleSpecifiers.ts:702:12`, types `["string", "string"]`.
- `parser.ts:1249:1` forEachChild: unit `binder.ts:350:1`, call `binder.ts:390:13`, types `["true"]`; unit `parser.ts:490:1`, call `parser.ts:491:40`, types `["Node"]`; unit `utilities.ts:12426:1`, call `utilities.ts:12427:12`, types `["Node"]`; unit `utilities.ts:8959:1`, call `utilities.ts:8961:74`, types `["Node"]`; unit `utilities.ts:960:1`, call `utilities.ts:966:13`, types `["boolean"]`.
- `path.ts:512:1` getPathFromPathComponents: unit `path.ts:1068:1`, call `path.ts:1082:12`, types `["string"]`.
- `path.ts:1096:1` forEachAncestorDirectory: unit `program.ts:329:1`, call `program.ts:330:12`, types `["string", "string"]`.
- `program.ts:1072:1` getTypeReferenceResolutionName: unit `program.ts:1076:1`, call `program.ts:1077:14`, types `["string | FileReference"]`.
- `scanner.ts:401:1` makeReverseMap: unit `scanner.ts:409:1`, call `scanner.ts:409:22`, types `["string"]`; unit `scanner.ts:423:1`, call `scanner.ts:423:29`, types `["CharacterCodes"]`.
- `scanner.ts:826:1` iterateCommentRanges: unit `scanner.ts:957:1`, call `scanner.ts:945:12`, types `["undefined", "CommentRange[] | undefined"]`; unit `scanner.ts:961:1`, call `scanner.ts:949:12`, types `["undefined", "CommentRange[] | undefined"]`.
- `scanner.ts:944:1` reduceEachLeadingCommentRange: unit `scanner.ts:957:1`, call `scanner.ts:958:12`, types `["undefined", "CommentRange[] | undefined"]`.
- `scanner.ts:948:1` reduceEachTrailingCommentRange: unit `scanner.ts:961:1`, call `scanner.ts:962:12`, types `["undefined", "CommentRange[] | undefined"]`.
- `transformer.ts:209:1` wrapCustomTransformerFactory: unit `transformer.ts:218:1`, call `transformer.ts:219:12`, types `["SourceFile"]`; unit `transformer.ts:222:1`, call `transformer.ts:223:12`, types `["Bundle | SourceFile"]`.
- `tsbuildPublic.ts:294:1` createSolutionBuilderHostBase: unit `executeCommandLine.ts:812:1`, call `tsbuildPublic.ts:324:18`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`.
- `tsbuildPublic.ts:317:1` createSolutionBuilderWithWatchHost: unit `executeCommandLine.ts:812:1`, call `executeCommandLine.ts:855:27`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`.
- `utilities.ts:747:1` forEachEntry: unit `builder.ts:956:1`, call `builder.ts:997:67`, types `["__String", "Symbol", "boolean"]`; unit `commandLineParser.ts:2774:1`, call `commandLineParser.ts:2776:12`, types `["string", "string | number", "string"]`; unit `utilities.ts:9375:1`, call `utilities.ts:9376:12`, types `["string", "string | number", "string"]`.
- `utilities.ts:763:1` forEachKey: unit `builder.ts:309:1`, call `builder.ts:314:101`, types `["string", "boolean"]`; unit `builder.ts:956:1`, call `builder.ts:1011:30`, types `["Path", "boolean"]`.
- `utilities.ts:3093:1` forEachPropertyAssignment: unit `utilities.ts:12295:1`, call `utilities.ts:12296:12`, types `["StringLiteral"]`; unit `utilities.ts:12303:1`, call `utilities.ts:12309:12`, types `["StringLiteral"]`; unit `utilities.ts:3112:1`, call `utilities.ts:3121:12`, types `["StringLiteral"]`.
- `utilities.ts:3120:1` forEachTsConfigPropArray: unit `utilities.ts:3112:1`, call `utilities.ts:3113:12`, types `["StringLiteral"]`.
- `utilities.ts:8208:1` mutateMapSkippingNewValues: unit `watchUtilities.ts:481:1`, call `utilities.ts:8251:5`, types `["Path", "FileWatcher", "string"]`.
- `utilities.ts:8249:1` mutateMap: unit `watchUtilities.ts:481:1`, call `watchUtilities.ts:487:5`, types `["Path", "FileWatcher", "string"]`.
- `utilities.ts:8314:1` addToSeen: unit `commandLineParser.ts:2716:1`, call `commandLineParser.ts:2721:13`, types `["string"]`.
- `utilities.ts:10194:1` changeExtension: unit `commandLineParser.ts:4231:1`, call `commandLineParser.ts:4243:46`, types `["string"]`; unit `commandLineParser.ts:4264:1`, call `commandLineParser.ts:4274:45`, types `["string"]`; unit `emitter.ts:573:1`, call `emitter.ts:574:12`, types `["string"]`; unit `emitter.ts:590:1`, call `emitter.ts:591:12`, types `["string"]`; unit `transformers/utilities.ts:876:1`, call `transformers/utilities.ts:884:25`, types `["string"]`.
- `utilities.ts:10655:1` setTextRangeEnd: unit `factory/emitHelpers.ts:148:1`, call `utilities.ts:10666:12`, types `["CallExpression"]`; unit `transformers/destructuring.ts:446:1`, call `utilities.ts:10666:12`, types `["Expression"]`; unit `visitorPublic.ts:420:1`, call `utilities.ts:10666:12`, types `["NodeArray<ParameterDeclaration>"]`; unit `visitorPublic.ts:474:1`, call `utilities.ts:10666:12`, types `["Block"]`.
- `utilities.ts:10665:1` setTextRangePosEnd: unit `factory/emitHelpers.ts:148:1`, call `factory/utilitiesPublic.ts:11:23`, types `["CallExpression"]`; unit `transformers/destructuring.ts:446:1`, call `factory/utilitiesPublic.ts:11:23`, types `["Expression"]`; unit `visitorPublic.ts:420:1`, call `factory/utilitiesPublic.ts:11:23`, types `["NodeArray<ParameterDeclaration>"]`; unit `visitorPublic.ts:474:1`, call `factory/utilitiesPublic.ts:11:23`, types `["Block"]`.
- `utilities.ts:10704:1` setParent: unit `binder.ts:330:1`, call `binder.ts:333:9`, types `["ModuleBody"]`; unit `factory/utilities.ts:204:1`, call `factory/utilities.ts:211:5`, types `["Identifier"]`.
- `utilities.ts:12172:1` forEachDynamicImportOrRequireCall: unit `transformers/module/esnextAnd2015.ts:67:1`, call `transformers/module/esnextAnd2015.ts:99:17`, types `["false", "false"]`.
- `utilities.ts:12308:1` forEachOptionsSyntaxByName: unit `utilities.ts:12303:1`, call `utilities.ts:12304:12`, types `["StringLiteral"]`.
- `utilitiesPublic.ts:303:1` sortAndDeduplicateDiagnostics: unit `program.ts:635:1`, call `program.ts:649:12`, types `["Diagnostic"]`; unit `program.ts:636:18`, call `program.ts:649:12`, types `["Diagnostic"]`; unit `program.ts:637:1`, call `program.ts:649:12`, types `["Diagnostic"]`.
- `utilitiesPublic.ts:766:1` getOriginalNode: unit `factory/utilities.ts:683:1`, call `factory/utilities.ts:684:23`, types `["SourceFile"]`; unit `factory/utilities.ts:690:1`, call `factory/utilities.ts:691:23`, types `["SourceFile"]`; unit `factory/utilities.ts:764:1`, call `factory/utilities.ts:775:27`, types `["SourceFile"]`; unit `transformers/utilities.ts:107:1`, call `transformers/utilities.ts:108:12`, types `["Node"]`.
- `utilitiesPublic.ts:1279:1` getFirstJSDocTag: unit `utilitiesPublic.ts:1106:1`, call `utilitiesPublic.ts:1107:14`, types `["JSDocParameterTag"]`; unit `utilitiesPublic.ts:1111:1`, call `utilitiesPublic.ts:1112:12`, types `["JSDocAugmentsTag"]`; unit `utilitiesPublic.ts:1121:1`, call `utilitiesPublic.ts:1122:12`, types `["JSDocClassTag"]`; unit `utilitiesPublic.ts:1126:1`, call `utilitiesPublic.ts:1127:12`, types `["JSDocPublicTag"]`; unit `utilitiesPublic.ts:1131:1`, call `utilitiesPublic.ts:1132:12`, types `["JSDocPublicTag"]`.
- `utilitiesPublic.ts:1284:1` getAllJSDocTags: unit `utilities.ts:4778:1`, call `utilities.ts:4779:12`, types `["JSDocOverloadTag"]`; unit `utilitiesPublic.ts:1116:1`, call `utilitiesPublic.ts:1117:12`, types `["JSDocImplementsTag"]`.
- `visitorPublic.ts:598:1` visitEachChild: unit `transformers/es2016.ts:21:1`, call `transformers/es2016.ts:34:16`, types `["SourceFile"]`; unit `transformers/es2019.ts:17:1`, call `transformers/es2019.ts:26:16`, types `["SourceFile"]`; unit `transformers/es2020.ts:42:1`, call `transformers/es2020.ts:55:16`, types `["SourceFile"]`; unit `transformers/es2021.ts:25:1`, call `transformers/es2021.ts:37:16`, types `["SourceFile"]`; unit `transformers/legacyDecorators.ts:88:1`, call `transformers/legacyDecorators.ts:114:25`, types `["SourceFile"]`.
- `watch.ts:842:1` createProgramHost: unit `executeCommandLine.ts:1011:1`, call `watch.ts:872:20`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`; unit `executeCommandLine.ts:1034:1`, call `watch.ts:872:20`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`; unit `executeCommandLine.ts:812:1`, call `tsbuildPublic.ts:295:18`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`.
- `watch.ts:870:1` createWatchCompilerHost: unit `executeCommandLine.ts:1011:1`, call `watch.ts:933:18`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`; unit `executeCommandLine.ts:1034:1`, call `watch.ts:964:18`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`.
- `watch.ts:922:1` createWatchCompilerHostOfConfigFile: unit `executeCommandLine.ts:1011:1`, call `executeCommandLine.ts:1020:31`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`.
- `watch.ts:954:1` createWatchCompilerHostOfFilesAndCompilerOptions: unit `executeCommandLine.ts:1034:1`, call `executeCommandLine.ts:1042:31`, types `["EmitAndSemanticDiagnosticsBuilderProgram"]`.

## Outside-only bodies

These have no route to an independently lowered census caller under the ownership rules above. Files are source files containing outside call sites, including host callback registration.

- `core.ts:50:1` forEachRight: `src/services/formatting/formatting.ts`.
- `core.ts:106:1` zipWith: `src/compiler/debug.ts`, `src/harness/fourslashImpl.ts`, `src/testRunner/unittests/services/textChanges.ts`.
- `core.ts:121:1` intersperse: `src/services/jsDoc.ts`, `src/services/textChanges.ts`.
- `core.ts:437:1` flatMapIterator: `src/services/codefixes/importFixes.ts`, `src/services/refactorProvider.ts`.
- `core.ts:481:1` mapAllOrFail: `src/services/codefixes/convertToEsModule.ts`.
- `core.ts:538:1` singleIterator: `src/server/session.ts`.
- `core.ts:600:1` mapEntries: `src/services/codefixes/inferFromUsage.ts`, `src/testRunner/unittests/helpers.ts`.
- `core.ts:639:1` getRangesWhere: `src/services/refactors/moveToFile.ts`.
- `core.ts:763:1` createSortedArray: `src/services/completions.ts`, `src/services/stringCompletions.ts`.
- `core.ts:963:1` combine: `src/services/codefixes/helpers.ts`, `src/services/codefixes/importFixes.ts`.
- `core.ts:1053:1` arrayReverseIterator: `src/server/session.ts`.
- `core.ts:1150:1` single: `src/services/codefixes/importFixes.ts`.
- `core.ts:1286:1` getProperty: `src/jsTyping/jsTyping.ts`, `src/testRunner/unittests/helpers/tsc.ts`, `src/typingsInstallerCore/typingsInstaller.ts`.
- `core.ts:1421:1` arrayToNumericMap: `src/services/classifier.ts`.
- `core.ts:1475:1` clone: `src/harness/fourslashImpl.ts`, `src/services/utilities.ts`, `src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts`.
- `core.ts:1519:1` maybeBind: `src/server/project.ts`, `src/services/services.ts`, `src/services/utilities.ts`.
- `core.ts:1625:1` createSet: `src/server/session.ts`, `src/testRunner/unittests/compilerCore.ts`.
- `core.ts:2142:1` compareProperties: `src/services/refactors/extractSymbol.ts`.
- `factory/emitNode.ts:326:1` setSnippetElement: `src/services/completions.ts`.
- `factory/emitNode.ts:333:1` ignoreSourceNewlines: `src/services/refactors/extractType.ts`.
- `sourcemap.ts:622:1` sameMapping: `src/harness/sourceMapRecorder.ts`.
- `tsbuildPublic.ts:2070:1` buildNextInvalidatedProject: `src/compiler/tsbuildPublic.ts`.
- `tsbuildPublic.ts:2078:1` buildNextInvalidatedProjectWorker: `src/compiler/tsbuildPublic.ts`.
- `utilities.ts:731:1` forEachAncestor: `src/services/jsDoc.ts`.
- `utilities.ts:4444:1` forEachImportClauseDeclaration: `src/services/codefixes/fixUnreferenceableDecoratorMetadata.ts`.
- `utilities.ts:8380:1` forEachNameInAccessChainWalkingLeft: `src/server/session.ts`.
- `utilities.ts:10343:1` sliceAfter: `src/services/codefixes/fixUnreachableCode.ts`.
- `utilities.ts:10675:1` setTextRangePosWidth: `src/compiler/parser.ts`.

## Hidden bodies, complete list

Core.ts comes first. Caller IDs resolve to the exact original first stops in the next section. REPORT.json gives each caller's representative symbol-call route and entry call site. Every listed body has at least one census caller; every such caller has an original earlier stop.

- `core.ts:82:1` firstDefinedIterator: C055, C058.
- `core.ts:93:1` reduceLeftIterator: C019.
- `core.ts:182:1` findLast: C006, C099, C102.
- `core.ts:209:1` findLastIndex: C006.
- `core.ts:311:1` clear: C006.
- `core.ts:418:1` flatMapToMutable: C101.
- `core.ts:458:1` sameFlatMap: C026.
- `core.ts:519:1` getOrUpdate: C006.
- `core.ts:555:1` spanMap: C074, C082.
- `core.ts:683:1` indicesOf: C006, C044, C048.
- `core.ts:687:1` deduplicateRelational: C006, C044, C048.
- `core.ts:709:1` deduplicateEquality: C006, C044, C048.
- `core.ts:725:1` deduplicate: C006, C044, C048.
- `core.ts:768:1` insertSorted: C101, C111.
- `core.ts:846:1` compact: C066, C068.
- `core.ts:881:1` relativeComplement: C006.
- `core.ts:1028:1` appendIfUnique: C004, C006, C024, C027, C044, C045, C066, C074, C076, C077, C079, C080, C081, C082, C083, C084, C085, C089.
- `core.ts:1038:1` stableSortIndices: C006, C044, C048.
- `core.ts:1060:1` rangeEquals: C006.
- `core.ts:1076:1` elementAt: C074.
- `core.ts:1096:1` firstOrUndefinedIterator: C006, C020.
- `core.ts:1106:1` first: C006, C033, C074.
- `core.ts:1112:1` firstIterator: C006.
- `core.ts:1168:1` singleOrMany: C074, C079, C083, C084, C085, C086.
- `core.ts:1175:1` replaceElement: C006.
- `core.ts:1309:1` getOwnValues: C060.
- `core.ts:1321:1` arrayOf: C006.
- `core.ts:1435:1` arrayToMultiMap: C006.
- `core.ts:1453:1` group: C006.
- `core.ts:1461:1` groupBy: C083.
- `core.ts:1492:1` extend: C008, C009.
- `core.ts:1510:1` copyProperties: C017, C018, C020.
- `core.ts:1527:1` createMultiMap: C006, C054, C056, C062, C064, C105.
- `core.ts:1763:1` toArray: C050.
- `core.ts:2172:1` getSpellingSuggestion: C006, C007, C056, C059.
- `core.ts:2462:1` orPredicate2: C006.
- `core.ts:2466:1` orPredicate3: C006.
- `core.ts:2488:1` not: C006.
- `core.ts:2496:1` singleElementArray: C006, C100.
- `core.ts:2501:1` enumerateInsertsAndDeletes: C064.
- `core.ts:2540:1` cartesianProduct: C006.
- `core.ts:2546:1` cartesianProductWorker: C006.
- `core.ts:2569:1` takeWhile: C089.
- `core.ts:2589:1` skipWhile: C089.
- `commandLineParser.ts:3747:1` convertCompileOnSaveOptionFromJson: C010.
- `commandLineParser.ts:3793:1` convertWatchOptionsFromJsonWorker: C010.
- `commandLineParser.ts:3823:1` convertJsonOption: C010, C011, C012, C013, C014.
- `commandLineParser.ts:3900:1` convertJsonOptionOfListType: C010, C011, C012, C013, C014.
- `emitter.ts:446:1` forEachEmittedFile: C016, C054.
- `emitter.ts:6374:1` getEmitListItem: C015.
- `factory/emitNode.ts:81:1` removeAllComments: C066, C068, C085, C089.
- `factory/emitNode.ts:113:1` setInternalEmitFlags: C023, C079, C089.
- `factory/emitNode.ts:123:1` addInternalEmitFlags: C041, C054, C079, C085.
- `factory/emitNode.ts:139:1` setSourceMapRange: C066, C074, C076, C077, C079, C080, C081, C083, C087, C089.
- `factory/emitNode.ts:154:1` setTokenSourceMapRange: C074.
- `factory/emitNode.ts:190:1` setCommentRange: C006, C022, C066, C068, C074, C079, C080, C081, C083, C086, C089.
- `factory/emitNode.ts:199:1` setSyntheticLeadingComments: C006, C066, C074, C089.
- `factory/emitNode.ts:204:1` addSyntheticLeadingComment: C006, C074.
- `factory/emitNode.ts:212:1` setSyntheticTrailingComments: C006, C066, C074, C081, C089.
- `factory/emitNode.ts:217:1` addSyntheticTrailingComment: C006, C081, C089.
- `factory/emitNode.ts:221:1` moveSyntheticComments: C074.
- `factory/emitNode.ts:249:1` addEmitHelper: C076, C077, C085.
- `factory/emitNode.ts:258:1` addEmitHelpers: C066, C074, C076, C077, C079, C080, C081, C082, C083, C084, C085, C089.
- `factory/emitNode.ts:339:1` setTypeNode: C089.
- `factory/emitNode.ts:346:1` getTypeNode: C015.
- `factory/emitNode.ts:351:1` setIdentifierTypeArguments: C006, C026.
- `factory/emitNode.ts:362:1` setIdentifierAutoGenerate: C026.
- `factory/emitNode.ts:373:1` setIdentifierGeneratedImportReference: C082.
- `factory/utilities.ts:1464:1` createBinaryExpressionTrampoline: C004, C006, C015.
- `factory/utilities.ts:1507:1` elideNodes: C083.
- `moduleNameResolver.ts:300:1` initializeResolutionField: C046, C047.
- `moduleNameResolver.ts:304:1` updateResolutionField: C046, C057.
- `moduleNameResolver.ts:311:1` initializeResolutionFieldForReadonlyCache: C046.
- `moduleNameResolver.ts:977:1` createCacheWithRedirects: C042, C043.
- `moduleNameResolver.ts:1068:1` getOrCreateCache: C042, C043.
- `moduleNameResolver.ts:1078:1` createPerDirectoryResolutionCache: C042, C043.
- `moduleNameResolver.ts:1119:1` createModeAwareCache: C042, C043, C054, C058.
- `moduleNameResolver.ts:1166:1` createNonRelativeNameResolutionCache: C042, C043.
- `moduleSpecifiers.ts:713:1` forEachFileNameOfModule: C049.
- `parser.ts:443:1` visitNode: C002, C004, C006, C016, C022, C051, C052, C053, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:447:1` visitNodes: C002, C004, C006, C016, C022, C051, C052, C053, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1140:1` forEachChildInCallOrConstructSignature: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1146:1` forEachChildInUnionOrIntersectionType: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1150:1` forEachChildInParenthesizedTypeOrTypeOperator: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1154:1` forEachChildInObjectOrArrayBindingPattern: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1158:1` forEachChildInCallOrNewExpression: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1166:1` forEachChildInBlock: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1170:1` forEachChildInContinueOrBreakStatement: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1174:1` forEachChildInClassDeclarationOrExpression: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1182:1` forEachChildInNamedImportsOrExports: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1186:1` forEachChildInImportOrExportSpecifier: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1191:1` forEachChildInJsxOpeningOrSelfClosingElement: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1197:1` forEachChildInOptionalRestOrJSDocParameterModifier: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1201:1` forEachChildInJSDocParameterOrPropertyTag: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1209:1` forEachChildInJSDocTypeLikeTag: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1215:1` forEachChildInJSDocLinkCodeOrPlain: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1219:1` forEachChildInJSDocTag: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1224:1` forEachChildInJSDocImportTag: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1232:1` forEachChildInPartiallyEmittedExpression: C002, C004, C006, C016, C022, C051, C052, C074, C096, C097, C098, C103, C104, C106.
- `parser.ts:1272:1` forEachChildRecursively: C001, C003, C004, C006, C016, C054, C082.
- `program.ts:1082:1` createTypeReferenceResolutionLoader: C020, C054.
- `program.ts:1105:1` loadWithModeAwareCache: C020, C054.
- `program.ts:5060:1` handleNoEmitOptions: C005, C054.
- `program.ts:5104:1` parseConfigHostFromCompilerHostLike: C017, C018, C020, C054.
- `scanner.ts:934:1` forEachLeadingCommentRange: C015.
- `scanner.ts:940:1` forEachTrailingCommentRange: C015.
- `sys.ts:192:1` pollWatchedFileQueue: C061, C063.
- `sys.ts:505:1` createSingleWatcherPerName: C065.
- `transformer.ts:248:1` transformNodes: C016, C067.
- `transformers/classThis.ts:104:1` injectClassThisAssignmentIfMissing: C079.
- `transformers/utilities.ts:377:1` multiMapSparseArrayAdd: C091, C092.
- `transformers/utilities.ts:812:1` walkUpLexicalEnvironments: C066.
- `transformers/utilities.ts:824:1` newPrivateEnvironment: C066.
- `transformers/utilities.ts:829:1` getPrivateIdentifier: C066.
- `transformers/utilities.ts:839:1` setPrivateIdentifier: C066.
- `transformers/utilities.ts:855:1` accessPrivateIdentifier: C066.
- `tsbuildPublic.ts:187:1` getOrCreateValueFromConfigFileMap: C020.
- `tsbuildPublic.ts:197:1` getOrCreateValueMapFromConfigFileMap: C020.
- `tsbuildPublic.ts:305:1` createSolutionBuilderHost: C020.
- `tsbuildPublic.ts:339:1` createSolutionBuilder: C020.
- `tsbuildPublic.ts:343:1` createSolutionBuilderWithWatch: C020.
- `tsbuildPublic.ts:428:1` createSolutionBuilderState: C020.
- `tsbuildPublic.ts:554:1` toPath: C020.
- `tsbuildPublic.ts:558:1` toResolvedConfigFilePath: C020.
- `tsbuildPublic.ts:572:1` getCachedParsedConfigFile: C020.
- `tsbuildPublic.ts:577:1` parseConfigFile: C020.
- `tsbuildPublic.ts:603:1` resolveProjectName: C020.
- `tsbuildPublic.ts:607:1` createBuildOrder: C020.
- `tsbuildPublic.ts:654:1` getBuildOrder: C020.
- `tsbuildPublic.ts:658:1` createStateBuildOrder: C020.
- `tsbuildPublic.ts:721:1` getBuildOrderFor: C020.
- `tsbuildPublic.ts:740:1` enableCache: C020.
- `tsbuildPublic.ts:777:1` disableCache: C020.
- `tsbuildPublic.ts:796:1` clearProjectStatus: C020.
- `tsbuildPublic.ts:801:1` addProjToQueue: C020.
- `tsbuildPublic.ts:811:1` setupInitialBuild: C020.
- `tsbuildPublic.ts:885:1` doneInvalidatedProject: C020.
- `tsbuildPublic.ts:895:1` createUpdateOutputFileStampsProject: C020.
- `tsbuildPublic.ts:931:1` createBuildOrUpdateInvalidedProject: C020.
- `tsbuildPublic.ts:1202:1` getNextInvalidatedProjectCreateInfo: C020.
- `tsbuildPublic.ts:1315:1` createInvalidatedProjectWithInfo: C020.
- `tsbuildPublic.ts:1340:1` getNextInvalidatedProject: C020.
- `tsbuildPublic.ts:1350:1` getOldProgram: C020.
- `tsbuildPublic.ts:1357:1` afterProgramDone: C020.
- `tsbuildPublic.ts:1374:1` getModifiedTime: C020.
- `tsbuildPublic.ts:1392:1` watchFile: C020.
- `tsbuildPublic.ts:1430:1` getOutputTimeStampMap: C020.
- `tsbuildPublic.ts:1438:1` getBuildInfoCacheEntry: C020.
- `tsbuildPublic.ts:1444:1` getBuildInfo: C020.
- `tsbuildPublic.ts:1456:1` checkConfigFileUpToDateStatus: C020.
- `tsbuildPublic.ts:1468:1` getUpToDateStatusWorker: C020.
- `tsbuildPublic.ts:1792:1` hasSameBuildInfo: C020.
- `tsbuildPublic.ts:1797:1` getUpToDateStatus: C020.
- `tsbuildPublic.ts:1815:1` updateOutputTimestampsWorker: C020.
- `tsbuildPublic.ts:1867:1` getLatestChangedDtsTime: C020.
- `tsbuildPublic.ts:1878:1` updateOutputTimestamps: C020.
- `tsbuildPublic.ts:1889:1` queueReferencingProjects: C020.
- `tsbuildPublic.ts:1949:1` build: C020.
- `tsbuildPublic.ts:1957:1` buildWorker: C020.
- `tsbuildPublic.ts:1986:1` clean: C020.
- `tsbuildPublic.ts:1994:1` cleanWorker: C020.
- `tsbuildPublic.ts:2038:1` invalidateProject: C020.
- `tsbuildPublic.ts:2053:1` invalidateProjectAndScheduleBuilds: C020.
- `tsbuildPublic.ts:2059:1` scheduleBuildInvalidatedProject: C020.
- `tsbuildPublic.ts:2111:1` watchConfigFile: C020.
- `tsbuildPublic.ts:2127:1` watchExtendedConfigFiles: C020.
- `tsbuildPublic.ts:2145:1` watchWildCardDirectories: C020.
- `tsbuildPublic.ts:2179:1` watchInputFiles: C020.
- `tsbuildPublic.ts:2200:1` watchPackageJsonFiles: C020.
- `tsbuildPublic.ts:2221:1` startWatching: C020.
- `tsbuildPublic.ts:2246:1` stopWatching: C020.
- `tsbuildPublic.ts:2260:1` createSolutionBuilderWorker: C020.
- `tsbuildPublic.ts:2282:1` relName: C020.
- `tsbuildPublic.ts:2286:1` reportStatus: C020.
- `tsbuildPublic.ts:2290:1` reportWatchStatus: C020.
- `tsbuildPublic.ts:2294:1` reportErrors: C020.
- `tsbuildPublic.ts:2298:1` reportAndStoreErrors: C020.
- `tsbuildPublic.ts:2306:1` reportParseConfigFileDiagnostic: C020.
- `tsbuildPublic.ts:2310:1` reportErrorSummary: C020.
- `tsbuildPublic.ts:2346:1` reportBuildQueue: C020.
- `tsbuildPublic.ts:2352:1` reportUpToDateStatus: C020.
- `tsbuildPublic.ts:2488:1` verboseReportProjectStatus: C020.
- `utilities.ts:626:1` getDeclarationOfKind: C006, C096.
- `utilities.ts:640:1` getDeclarationsOfKind: C006.
- `utilities.ts:779:1` copyEntries: C006.
- `utilities.ts:931:1` hasChangesInResolutions: C054.
- `utilities.ts:1112:1` isGrammarErrorElement: C094.
- `utilities.ts:1117:1` insertStatementsAfterPrologue: C074, C076, C077, C081, C084, C085, C086, C089.
- `utilities.ts:1130:1` insertStatementAfterPrologue: C074, C082.
- `utilities.ts:1152:1` insertStatementsAfterStandardPrologue: C074, C076, C077, C081, C085, C086, C089.
- `utilities.ts:1157:1` insertStatementsAfterCustomPrologue: C074, C084.
- `utilities.ts:1172:1` insertStatementAfterCustomPrologue: C074, C082.
- `utilities.ts:2891:1` forEachReturnStatement: C006, C022.
- `utilities.ts:2921:1` forEachYieldExpression: C006.
- `utilities.ts:8170:1` clearMap: C017, C018, C020, C058.
- `utilities.ts:10350:1` addRelatedInfo: C004, C006, C054, C068, C069.
- `utilities.ts:10363:1` minAndMax: C006.
- `utilities.ts:10630:1` arrayIsHomogeneous: C006.
- `utilities.ts:10645:1` setTextRangePos: C006, C015, C022, C023, C025, C026, C028, C030, C031, C032, C033, C034, C035, C036, C037, C038, C039, C040, C066, C068, C070, C071, C072, C073, C074, C075, C076, C077, C078, C079, C080, C081, C082, C083, C084, C085, C086, C087, C088, C089, C090, C093, C107, C108, C109.
- `utilities.ts:10688:1` setNodeFlags: C006.
- `utilities.ts:10723:1` setParentRecursive: C001, C003, C004, C006, C054, C082.
- `utilities.ts:11312:1` evaluatorResult: C006, C095.
- `utilities.ts:12236:1` forEachResolvedProjectReference: C054.
- `utilities.ts:12248:1` forEachProjectReference: C054, C056.
- `utilities.ts:12321:1` getSynthesizedDeepClone: C006.
- `utilities.ts:12328:1` getSynthesizedDeepCloneWithReplacements: C006.
- `utilities.ts:12345:1` getSynthesizedDeepCloneWorker: C006.
- `utilities.ts:12374:1` getSynthesizedDeepClones: C006.
- `utilities.ts:12384:1` getSynthesizedDeepClonesWithReplacements: C006.
- `utilitiesPublic.ts:1487:1` isNodeArray: C026, C029.
- `watch.ts:333:1` isBuilderProgram: C017, C018, C020, C021, C110, C112.
- `watch.ts:337:1` listFiles: C017, C018, C020, C021, C110.
- `watch.ts:563:1` emitFilesAndReportErrors: C017, C018, C020, C021, C110.
- `watch.ts:633:1` emitFilesAndReportErrorsAndGetExitStatus: C021, C110.
- `watch.ts:744:1` createWatchFactory: C017, C018, C020.
- `watch.ts:753:1` createCompilerHostFromProgramHost: C017, C018, C020.
- `watchPublic.ts:140:1` createIncrementalProgram: C110.
- `watchPublic.ts:419:1` createWatchProgram: C017, C018.
- `watchUtilities.ts:407:1` updateSharedExtendedConfigFileWatcher: C017, C018, C020.
- `watchUtilities.ts:449:1` clearSharedExtendedConfigFileWatcher: C017, C018.
- `watchUtilities.ts:514:1` updateWatchingWildcardDirectories: C017, C018, C020.
- `watchUtilities.ts:700:1` getWatchFactory: C017, C018, C020.
- `watchUtilities.ts:845:1` closeFileWatcherOf: C017, C018, C020, C058, C064, C065.

## Census caller first-stop ledger

Each ID names the independently lowered caller and its original first finding. Diagnostic replay reproduces that finding exactly.

- **C001** `binder.ts:330:1` stops at `utilities.ts:10706:10` (lowering, Refused): a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace. Propagation: binder.ts:330:1 lowerBody_exit; utilities.ts:10704:1 lowerBody_exit.
- **C002** `binder.ts:350:1` stops at `debug.ts:252:33` (lowering, NotYet): a value of type never. Propagation: binder.ts:350:1 lowerBody_exit; binder.ts:390:32 lowerBody_exit; debug.ts:252:5 latentMainSignatureRaw_exit.
- **C003** `binder.ts:422:1` stops at `binder.ts:2258:13` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: binder.ts:422:1 lowerBody_exit.
- **C004** `binder.ts:509:1` stops at `binder.ts:3201:45` (lowering, NotYet): a value of type BindableObjectDefinePropertyCall. Propagation: binder.ts:3201:5 latentMainSignatureRaw_exit; binder.ts:509:1 lowerBody_exit.
- **C005** `builder.ts:1696:1` stops at `builder.ts:1686:1` (lowering, Refused): overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram | undefined. Propagation: builder.ts:1696:1 latentMainSignatureRaw_exit.
- **C006** `checker.ts:1489:1` stops at `checker.ts:2971:55` (lowering, NotYet): a value of type ParameterPropertyDeclaration. Propagation: checker.ts:1489:1 lowerBody_exit; checker.ts:2971:5 latentMainSignatureRaw_exit.
- **C007** `commandLineParser.ts:1933:1` stops at `commandLineParser.ts:1940:25` (lowering, NotYet): a call through ?. (an optional call). Propagation: commandLineParser.ts:1933:1 lowerBody_exit.
- **C008** `commandLineParser.ts:3031:1` stops at `commandLineParser.ts:3128:33` (lowering, Refused): a generic instantiated at void whose body reads the value: getPropFromRaw's T = void (read from the signature), read at /workspace/scratch/abrq61p-adapted/src/compiler/commandLineParser.ts:3253:24 as a returned value. Propagation: commandLineParser.ts:3031:1 lowerBody_exit; commandLineParser.ts:3077:1 lowerBody_exit; commandLineParser.ts:3127:5 lowerBody_exit.
- **C009** `commandLineParser.ts:3042:1` stops at `commandLineParser.ts:3128:33` (lowering, Refused): a generic instantiated at void whose body reads the value: getPropFromRaw's T = void (read from the signature), read at /workspace/scratch/abrq61p-adapted/src/compiler/commandLineParser.ts:3253:24 as a returned value. Propagation: commandLineParser.ts:3042:1 lowerBody_exit; commandLineParser.ts:3077:1 lowerBody_exit; commandLineParser.ts:3127:5 lowerBody_exit.
- **C010** `commandLineParser.ts:3518:1` stops at `commandLineParser.ts:3519:5` (lowering, NotYet): a value of type any. Propagation: commandLineParser.ts:3518:1 latentMainSignatureRaw_exit.
- **C011** `commandLineParser.ts:3539:1` stops at `commandLineParser.ts:3563:38` (lowering, Refused): a value of type string[] | PluginImport[] | ProjectReference[] | (string | number)[] seen as unknown[], which can write unknown where string is read. Propagation: commandLineParser.ts:3539:1 lowerBody_exit.
- **C012** `commandLineParser.ts:3589:1` stops at `commandLineParser.ts:3621:9` (lowering, NotYet): a value of type any. Propagation: commandLineParser.ts:3589:1 lowerBody_exit; commandLineParser.ts:3619:5 latentMainSignatureRaw_exit.
- **C013** `commandLineParser.ts:3755:1` stops at `commandLineParser.ts:3797:1` (lowering, Refused): overload 1 of convertOptionsFromJson result WatchOptions | undefined cannot be served by implementation result CompilerOptions | TypeAcquisition | WatchOptions | undefined. Propagation: commandLineParser.ts:3755:1 lowerBody_exit; commandLineParser.ts:3774:1 lowerBody_exit; commandLineParser.ts:3799:1 latentMainSignatureRaw_exit.
- **C014** `commandLineParser.ts:3761:1` stops at `commandLineParser.ts:3797:1` (lowering, Refused): overload 1 of convertOptionsFromJson result WatchOptions | undefined cannot be served by implementation result CompilerOptions | TypeAcquisition | WatchOptions | undefined. Propagation: commandLineParser.ts:3761:1 lowerBody_exit; commandLineParser.ts:3787:1 lowerBody_exit; commandLineParser.ts:3799:1 latentMainSignatureRaw_exit.
- **C015** `emitter.ts:1211:1` stops at `emitter.ts:4949:5` (lowering, Refused): overload 1 of writeTokenText result void cannot be served by implementation result number. Propagation: emitter.ts:1211:1 lowerBody_exit; emitter.ts:4951:5 latentMainSignatureRaw_exit.
- **C016** `emitter.ts:752:1` stops at `emitter.ts:836:10` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: emitter.ts:752:1 lowerBody_exit; emitter.ts:821:5 lowerBody_exit.
- **C017** `executeCommandLine.ts:1011:1` stops at `watch.ts:861:41` (lowering, Refused): a cast the runtime can't check. Propagation: executeCommandLine.ts:1011:1 lowerBody_exit; watch.ts:842:1 lowerBody_exit; watch.ts:870:1 lowerBody_exit; watch.ts:922:1 lowerBody_exit.
- **C018** `executeCommandLine.ts:1034:1` stops at `watch.ts:861:41` (lowering, Refused): a cast the runtime can't check. Propagation: executeCommandLine.ts:1034:1 lowerBody_exit; watch.ts:842:1 lowerBody_exit; watch.ts:870:1 lowerBody_exit; watch.ts:954:1 lowerBody_exit.
- **C019** `executeCommandLine.ts:1160:1` stops at `executeCommandLine.ts:1170:28` (lowering, NotYet): a literal method through a view that erases its receiver. Propagation: executeCommandLine.ts:1160:1 lowerBody_exit.
- **C020** `executeCommandLine.ts:812:1` stops at `watch.ts:861:41` (lowering, Refused): a cast the runtime can't check. Propagation: executeCommandLine.ts:812:1 lowerBody_exit; tsbuildPublic.ts:294:1 lowerBody_exit; tsbuildPublic.ts:317:1 lowerBody_exit; watch.ts:842:1 lowerBody_exit.
- **C021** `executeCommandLine.ts:908:1` stops at `executeCommandLine.ts:916:5` (lowering, NotYet): writing a possibly absent optional own field. Propagation: executeCommandLine.ts:908:1 lowerBody_exit.
- **C022** `expressionToTypeNode.ts:174:1` stops at `expressionToTypeNode.ts:1220:39` (lowering, NotYet): a value of type PrimitiveLiteral. Propagation: expressionToTypeNode.ts:1220:5 latentMainSignatureRaw_exit; expressionToTypeNode.ts:174:1 lowerBody_exit.
- **C023** `factory/emitHelpers.ts:148:1` stops at `utilities.ts:10656:6` (lowering, Refused): a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace. Propagation: factory/emitHelpers.ts:148:1 lowerBody_exit; factory/emitHelpers.ts:240:5 lowerBody_exit; factory/utilitiesPublic.ts:10:1 lowerBody_exit; utilities.ts:10655:1 lowerBody_exit; utilities.ts:10665:1 lowerBody_exit.
- **C024** `factory/emitNode.ts:289:1` stops at `factory/emitNode.ts:300:13` (lowering, NotYet): writing a possibly absent optional own field. Propagation: factory/emitNode.ts:289:1 lowerBody_exit.
- **C025** `factory/nodeConverters.ts:41:1` stops at `factory/nodeConverters.ts:55:13` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: factory/nodeConverters.ts:41:1 lowerBody_exit; factory/nodeConverters.ts:54:5 lowerBody_exit.
- **C026** `factory/nodeFactory.ts:492:1` stops at `factory/nodeFactory.ts:2363:40` (lowering, NotYet): a rest parameter that isn't an array. Propagation: factory/nodeFactory.ts:2363:5 latentMainSignatureRaw_exit; factory/nodeFactory.ts:492:1 lowerBody_exit.
- **C027** `factory/nodeFactory.ts:7431:1` stops at `factory/nodeFactory.ts:7442:9` (lowering, NotYet): a destructured name held otherwise than its field. Propagation: factory/nodeFactory.ts:7431:1 lowerBody_exit.
- **C028** `factory/parenthesizerRules.ts:54:1` stops at `factory/parenthesizerRules.ts:461:5` (lowering, Refused): overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody. Propagation: factory/parenthesizerRules.ts:463:5 latentMainSignatureRaw_exit; factory/parenthesizerRules.ts:54:1 lowerBody_exit.
- **C029** `factory/parenthesizerRules.ts:675:1` stops at `core.ts:1789:59` (lowering, NotYet): a template interpolating an object, an array, a map, a function or undefined. Propagation: core.ts:1786:1 lowerBody_exit; factory/parenthesizerRules.ts:684:34 lowerBody_exit.
- **C030** `factory/utilities.ts:188:1` stops at `factory/utilities.ts:189:9` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: factory/utilities.ts:188:1 lowerBody_exit.
- **C031** `factory/utilities.ts:247:1` stops at `factory/emitNode.ts:176:5` (lowering, NotYet): writing a possibly absent optional own field. Propagation: factory/emitNode.ts:175:1 lowerBody_exit; factory/utilities.ts:247:1 lowerBody_exit; factory/utilities.ts:678:1 lowerBody_exit.
- **C032** `factory/utilities.ts:280:1` stops at `factory/emitNode.ts:176:5` (lowering, NotYet): writing a possibly absent optional own field. Propagation: factory/emitNode.ts:175:1 lowerBody_exit; factory/utilities.ts:280:1 lowerBody_exit; factory/utilities.ts:678:1 lowerBody_exit.
- **C033** `factory/utilities.ts:309:1` stops at `factory/utilities.ts:310:9` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: factory/utilities.ts:309:1 lowerBody_exit.
- **C034** `factory/utilities.ts:334:1` stops at `factory/utilities.ts:335:9` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: factory/utilities.ts:334:1 lowerBody_exit.
- **C035** `factory/utilities.ts:348:1` stops at `factory/utilities.ts:349:9` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: factory/utilities.ts:348:1 lowerBody_exit.
- **C036** `factory/utilities.ts:362:1` stops at `factory/utilities.ts:362:107` (lowering, NotYet): a value of type AccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; }. Propagation: factory/utilities.ts:362:1 latentMainSignatureRaw_exit.
- **C037** `factory/utilities.ts:411:1` stops at `factory/nodeFactory.ts:7422:9` (lowering, NotYet): writing a possibly absent optional own field. Propagation: factory/nodeFactory.ts:7420:1 lowerBody_exit; factory/utilities.ts:411:1 lowerBody_exit.
- **C038** `factory/utilities.ts:424:1` stops at `factory/nodeFactory.ts:7422:9` (lowering, NotYet): writing a possibly absent optional own field. Propagation: factory/nodeFactory.ts:7420:1 lowerBody_exit; factory/utilities.ts:424:1 lowerBody_exit.
- **C039** `factory/utilities.ts:437:1` stops at `factory/nodeFactory.ts:7422:9` (lowering, NotYet): writing a possibly absent optional own field. Propagation: factory/nodeFactory.ts:7420:1 lowerBody_exit; factory/utilities.ts:437:1 lowerBody_exit.
- **C040** `factory/utilities.ts:516:1` stops at `factory/utilities.ts:527:18` (lowering, NotYet): a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation. Propagation: factory/utilities.ts:516:1 lowerBody_exit.
- **C041** `factory/utilities.ts:697:1` stops at `factory/utilities.ts:719:38` (lowering, NotYet): a comparator that doesn't take two elements and return a number. Propagation: factory/utilities.ts:697:1 lowerBody_exit.
- **C042** `moduleNameResolver.ts:1341:1` stops at `moduleNameResolver.ts:1284:31` (lowering, NotYet): a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions. Propagation: moduleNameResolver.ts:1276:1 lowerBody_exit; moduleNameResolver.ts:1341:1 lowerBody_exit.
- **C043** `moduleNameResolver.ts:1374:1` stops at `moduleNameResolver.ts:1284:31` (lowering, NotYet): a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions. Propagation: moduleNameResolver.ts:1276:1 lowerBody_exit; moduleNameResolver.ts:1374:1 lowerBody_exit.
- **C044** `moduleNameResolver.ts:2228:1` stops at `moduleNameResolver.ts:2256:54` (lowering, NotYet): a union or optional field read in a program with record storage. Propagation: moduleNameResolver.ts:2228:1 lowerBody_exit.
- **C045** `moduleNameResolver.ts:2280:1` stops at `moduleNameResolver.ts:2283:5` (lowering, NotYet): a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }. Propagation: moduleNameResolver.ts:2280:1 latentMainSignatureRaw_exit.
- **C046** `moduleNameResolver.ts:259:1` stops at `moduleNameResolver.ts:271:13` (lowering, NotYet): writing a possibly absent optional own field. Propagation: moduleNameResolver.ts:259:1 lowerBody_exit.
- **C047** `moduleNameResolver.ts:538:1` stops at `moduleNameResolver.ts:674:21` (lowering, NotYet): a dictionary member narrowed to an array without an element-checked view. Propagation: moduleNameResolver.ts:538:1 lowerBody_exit; moduleNameResolver.ts:662:5 lowerBody_exit; moduleNameResolver.ts:668:44 lowerBody_exit.
- **C048** `moduleNameResolver.ts:813:1` stops at `utilities.ts:9026:1` (lowering, NotYet): a checked predicate target whose base has no required field to check (CompilerOptions & { types: string[]; }). Propagation: moduleNameResolver.ts:813:1 lowerBody_exit.
- **C049** `moduleSpecifiers.ts:809:1` stops at `debug.ts:344:5` (lowering, NotYet): checked predicate overload target ModuleResolutionHost. Propagation: moduleSpecifiers.ts:809:1 lowerBody_exit.
- **C050** `parser.ts:10619:1` stops at `parser.ts:10620:5` (lowering, NotYet): writing a possibly absent optional own field. Propagation: parser.ts:10619:1 lowerBody_exit.
- **C051** `parser.ts:1311:1` stops at `parser.ts:1249:17` (lowering, NotYet): a function returning T | undefined. Propagation: parser.ts:1249:1 latentMainSignatureRaw_exit; parser.ts:1311:1 lowerBody_exit.
- **C052** `parser.ts:490:1` stops at `parser.ts:1253:17` (deferred, NotYet): a member typed any (167(0)) related to QualifiedName at the assertion site. Propagation: parser.ts:490:1 lowerBody_exit.
- **C053** `parser.ts:505:1` stops at `parser.ts:506:5` (lowering, NotYet): a computed field name.
- **C054** `program.ts:1516:1` stops at `program.ts:1972:16` (lowering, NotYet): an optional Map get through a structural receiver. Propagation: program.ts:1516:1 lowerBody_exit; program.ts:1971:5 lowerBody_exit.
- **C055** `program.ts:4892:1` stops at `program.ts:4903:9` (lowering, NotYet): a value of type any. Propagation: program.ts:4892:1 lowerBody_exit.
- **C056** `programDiagnostics.ts:89:1` stops at `programDiagnostics.ts:191:36` (lowering, Refused): a cast the runtime can't check. Propagation: programDiagnostics.ts:190:5 lowerBody_exit; programDiagnostics.ts:89:1 lowerBody_exit.
- **C057** `resolutionCache.ts:532:1` stops at `debug.ts:227:5` (lowering, NotYet): a checked predicate target whose base has no required field to check ((externalModuleName: string) => string). Propagation: debug.ts:234:5 lowerBody_exit; resolutionCache.ts:532:1 lowerBody_exit.
- **C058** `resolutionCache.ts:579:1` stops at `resolutionCache.ts:632:11` (lowering, NotYet): a value of type PathPathComponents. Propagation: resolutionCache.ts:579:1 lowerBody_exit.
- **C059** `scanner.ts:1023:1` stops at `scanner.ts:2362:62` (lowering, NotYet): a PostfixUnaryExpression. Propagation: scanner.ts:1023:1 lowerBody_exit; scanner.ts:1892:5 lowerBody_exit.
- **C060** `symbolWalker.ts:30:1` stops at `checker.ts:20082:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: symbolWalker.ts:30:1 lowerBody_exit; symbolWalker.ts:44:5 lowerBody_exit; symbolWalker.ts:71:9 lowerBody_exit.
- **C061** `sys.ts:251:1` stops at `sys.ts:263:11` (lowering, NotYet): a value of type PollingIntervalQueue. Propagation: sys.ts:251:1 lowerBody_exit.
- **C062** `sys.ts:388:1` stops at `core.ts:2363:60` (lowering, NotYet): function identity comparison in a program with specialized generic function values. Propagation: core.ts:1546:1 lowerBody_exit; core.ts:2362:1 lowerBody_exit; core.ts:2363:49 lowerBody_exit; sys.ts:388:1 lowerBody_exit; sys.ts:401:5 lowerBody_exit; sys.ts:411:20 lowerBody_exit.
- **C063** `sys.ts:464:1` stops at `sys.ts:470:9` (lowering, NotYet): a value of type any. Propagation: sys.ts:464:1 lowerBody_exit.
- **C064** `sys.ts:599:1` stops at `sys.ts:603:5` (lowering, Refused): a method in object destructuring. Propagation: sys.ts:599:1 lowerBody_exit.
- **C065** `sys.ts:997:1` stops at `sys.ts:1007:5` (lowering, Refused): a method in object destructuring. Propagation: sys.ts:997:1 lowerBody_exit.
- **C066** `transformers/classFields.ts:354:1` stops at `transformers/classFields.ts:671:55` (lowering, NotYet): a value of type PrivateIdentifierInExpression. Propagation: transformers/classFields.ts:354:1 lowerBody_exit; transformers/classFields.ts:671:5 latentMainSignatureRaw_exit.
- **C067** `transformers/declarations.ts:221:1` stops at `core.ts:261:1` (lowering, Refused): overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined. Propagation: core.ts:277:1 latentMainSignatureRaw_exit; transformers/declarations.ts:221:1 lowerBody_exit.
- **C068** `transformers/declarations.ts:259:1` stops at `binder.ts:1109:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: transformers/declarations.ts:259:1 lowerBody_exit; transformers/declarations.ts:313:5 lowerBody_exit; transformers/declarations.ts:314:65 lowerBody_exit.
- **C069** `transformers/declarations/diagnostics.ts:615:1` stops at `binder.ts:1109:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: transformers/declarations/diagnostics.ts:615:1 lowerBody_exit; transformers/declarations/diagnostics.ts:673:5 lowerBody_exit.
- **C070** `transformers/destructuring.ts:244:1` stops at `debug.ts:273:82` (lowering, Refused): overload 1 of assertNode parameter test cannot be served by implementation parameter test. Propagation: debug.ts:275:5 latentMainSignatureRaw_exit; transformers/destructuring.ts:244:1 lowerBody_exit; transformers/destructuring.ts:318:5 lowerBody_exit.
- **C071** `transformers/destructuring.ts:446:1` stops at `utilities.ts:10656:6` (lowering, Refused): a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace. Propagation: factory/utilitiesPublic.ts:10:1 lowerBody_exit; transformers/destructuring.ts:446:1 lowerBody_exit; utilities.ts:10655:1 lowerBody_exit; utilities.ts:10665:1 lowerBody_exit.
- **C072** `transformers/destructuring.ts:583:1` stops at `binder.ts:1109:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: transformers/destructuring.ts:583:1 lowerBody_exit.
- **C073** `transformers/destructuring.ts:99:1` stops at `transformers/destructuring.ts:171:14` (lowering, NotYet): a branded supertype inside a checked predicate's array target. Propagation: transformers/destructuring.ts:99:1 lowerBody_exit.
- **C074** `transformers/es2015.ts:488:1` stops at `transformers/es2015.ts:825:14` (lowering, NotYet): a function returning CapturedThis. Propagation: transformers/es2015.ts:488:1 lowerBody_exit; transformers/es2015.ts:825:5 latentMainSignatureRaw_exit.
- **C075** `transformers/es2016.ts:21:1` stops at `visitorPublic.ts:587:119` (lowering, Refused): overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor. Propagation: transformers/es2016.ts:21:1 lowerBody_exit; transformers/es2016.ts:29:5 lowerBody_exit; visitorPublic.ts:598:1 latentMainSignatureRaw_exit.
- **C076** `transformers/es2017.ts:119:1` stops at `transformers/es2017.ts:736:5` (lowering, Refused): overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody. Propagation: transformers/es2017.ts:119:1 lowerBody_exit; transformers/es2017.ts:738:5 latentMainSignatureRaw_exit.
- **C077** `transformers/es2018.ts:154:1` stops at `transformers/es2018.ts:1220:5` (lowering, Refused): overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody. Propagation: transformers/es2018.ts:1222:5 latentMainSignatureRaw_exit; transformers/es2018.ts:154:1 lowerBody_exit.
- **C078** `transformers/es2020.ts:42:1` stops at `visitorPublic.ts:587:119` (lowering, Refused): overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor. Propagation: transformers/es2020.ts:42:1 lowerBody_exit; transformers/es2020.ts:50:5 lowerBody_exit; visitorPublic.ts:598:1 latentMainSignatureRaw_exit.
- **C079** `transformers/esDecorators.ts:295:1` stops at `transformers/esDecorators.ts:670:14` (lowering, NotYet): a function returning ImmediatelyInvokedArrowFunction. Propagation: transformers/esDecorators.ts:295:1 lowerBody_exit; transformers/esDecorators.ts:670:5 latentMainSignatureRaw_exit.
- **C080** `transformers/esnext.ts:71:1` stops at `visitorPublic.ts:123:5` (lowering, Refused): overload 1 of visitNode parameter node cannot be served by implementation parameter node. Propagation: transformers/esnext.ts:71:1 lowerBody_exit; transformers/esnext.ts:92:5 lowerBody_exit; visitorPublic.ts:148:1 latentMainSignatureRaw_exit.
- **C081** `transformers/generators.ts:336:1` stops at `transformers/generators.ts:1387:43` (lowering, NotYet): a value of type InitializedVariableDeclaration. Propagation: transformers/generators.ts:1387:5 latentMainSignatureRaw_exit; transformers/generators.ts:336:1 lowerBody_exit.
- **C082** `transformers/jsx.ts:92:1` stops at `transformers/jsx.ts:110:13` (lowering, NotYet): a value of type (VariableDeclaration & { name: Identifier; }) | undefined. Propagation: transformers/jsx.ts:109:5 lowerBody_exit; transformers/jsx.ts:92:1 lowerBody_exit.
- **C083** `transformers/legacyDecorators.ts:88:1` stops at `visitorPublic.ts:587:119` (lowering, Refused): overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor. Propagation: transformers/legacyDecorators.ts:113:5 lowerBody_exit; transformers/legacyDecorators.ts:88:1 lowerBody_exit; visitorPublic.ts:598:1 latentMainSignatureRaw_exit.
- **C084** `transformers/module/esnextAnd2015.ts:67:1` stops at `utilities.ts:12183:22` (lowering, Refused): a cast the runtime can't check. Propagation: transformers/module/esnextAnd2015.ts:67:1 lowerBody_exit; transformers/module/esnextAnd2015.ts:90:5 lowerBody_exit; utilities.ts:12172:1 lowerBody_exit.
- **C085** `transformers/module/module.ts:174:1` stops at `transformers/module/module.ts:1926:43` (lowering, NotYet): a value of type InitializedVariableDeclaration. Propagation: transformers/module/module.ts:174:1 lowerBody_exit; transformers/module/module.ts:1926:5 latentMainSignatureRaw_exit.
- **C086** `transformers/module/system.ts:137:1` stops at `transformers/module/system.ts:206:22` (lowering, NotYet): an assignment value to a member. Propagation: transformers/module/system.ts:137:1 lowerBody_exit; transformers/module/system.ts:183:5 lowerBody_exit.
- **C087** `transformers/namedEvaluation.ts:179:1` stops at `transformers/namedEvaluation.ts:173:1` (lowering, Refused): overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration. Propagation: transformers/namedEvaluation.ts:179:1 latentMainSignatureRaw_exit.
- **C088** `transformers/taggedTemplate.ts:110:1` stops at `factory/nodeFactory.ts:6111:21` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: transformers/taggedTemplate.ts:110:1 lowerBody_exit.
- **C089** `transformers/ts.ts:235:1` stops at `transformers/ts.ts:1430:55` (lowering, NotYet): a value of type ParameterPropertyDeclaration. Propagation: transformers/ts.ts:1430:5 latentMainSignatureRaw_exit; transformers/ts.ts:235:1 lowerBody_exit.
- **C090** `transformers/typeSerializer.ts:132:1` stops at `utilities.ts:6757:17` (lowering, NotYet): a function returning (ConstructorDeclaration & { body: Block; }) | undefined. Propagation: transformers/typeSerializer.ts:132:1 lowerBody_exit; transformers/typeSerializer.ts:200:5 lowerBody_exit; utilities.ts:6757:1 latentMainSignatureRaw_exit.
- **C091** `transformers/utilities.ts:179:1` stops at `core.ts:1789:59` (lowering, NotYet): a template interpolating an object, an array, a map, a function or undefined. Propagation: core.ts:1786:1 lowerBody_exit; transformers/utilities.ts:179:1 lowerBody_exit; transformers/utilities.ts:303:5 lowerBody_exit.
- **C092** `transformers/utilities.ts:355:1` stops at `transformers/utilities.ts:356:9` (lowering, NotYet): checked predicate membership over an open tag domain. Propagation: transformers/utilities.ts:355:1 lowerBody_exit.
- **C093** `transformers/utilities.ts:880:1` stops at `utilities.ts:10195:12` (lowering, Refused): an unproven relation from string to T: the source is not assignable to the target. Propagation: transformers/utilities.ts:880:1 lowerBody_exit; utilities.ts:10194:1 lowerBody_exit.
- **C094** `utilities.ts:1097:1` stops at `parser.ts:3983:9` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: utilities.ts:1097:1 lowerBody_exit.
- **C095** `utilities.ts:11317:1` stops at `utilities.ts:11317:35` (lowering, Refused): a method in object destructuring. Propagation: utilities.ts:11317:1 lowerBody_exit.
- **C096** `utilities.ts:11492:1` stops at `utilities.ts:11500:5` (lowering, NotYet): a destructured name that isn't plain. Propagation: utilities.ts:11492:1 lowerBody_exit.
- **C097** `utilities.ts:12201:1` stops at `binder.ts:1034:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: utilities.ts:12201:1 lowerBody_exit.
- **C098** `utilities.ts:12426:1` stops at `parser.ts:1253:17` (deferred, NotYet): a member typed any (167(0)) related to QualifiedName at the assertion site. Propagation: utilities.ts:12426:1 lowerBody_exit.
- **C099** `utilities.ts:1277:1` stops at `binder.ts:1034:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: utilities.ts:1277:1 lowerBody_exit.
- **C100** `utilities.ts:5227:1` stops at `core.ts:656:1` (lowering, Refused): overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] | undefined. Propagation: core.ts:672:1 latentMainSignatureRaw_exit; utilities.ts:5227:1 lowerBody_exit.
- **C101** `utilities.ts:6058:1` stops at `utilities.ts:6120:5` (lowering, Refused): overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[]. Propagation: utilities.ts:6058:1 lowerBody_exit; utilities.ts:6122:5 latentMainSignatureRaw_exit.
- **C102** `utilities.ts:7877:1` stops at `binder.ts:1034:17` (lowering, NotYet): writing a checked object field without its source-slot type certificate. Propagation: utilities.ts:7877:1 lowerBody_exit.
- **C103** `utilities.ts:8293:1` stops at `parser.ts:1249:17` (lowering, NotYet): a function returning T | undefined. Propagation: parser.ts:1249:1 latentMainSignatureRaw_exit; utilities.ts:8293:1 lowerBody_exit.
- **C104** `utilities.ts:8959:1` stops at `parser.ts:1253:17` (deferred, NotYet): a member typed any (167(0)) related to QualifiedName at the assertion site. Propagation: utilities.ts:8959:1 lowerBody_exit.
- **C105** `utilities.ts:9501:1` stops at `utilities.ts:9539:18` (lowering, NotYet): optional chaining to .size on a value. Propagation: utilities.ts:9501:1 lowerBody_exit; utilities.ts:9538:5 lowerBody_exit.
- **C106** `utilities.ts:960:1` stops at `utilities.ts:970:14` (lowering, Refused): a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace. Propagation: utilities.ts:960:1 lowerBody_exit.
- **C107** `visitorPublic.ts:223:1` stops at `visitorPublic.ts:196:5` (lowering, Refused): overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor. Propagation: visitorPublic.ts:223:1 latentMainSignatureRaw_exit.
- **C108** `visitorPublic.ts:420:1` stops at `utilities.ts:10656:6` (lowering, Refused): a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace. Propagation: factory/utilitiesPublic.ts:10:1 lowerBody_exit; utilities.ts:10655:1 lowerBody_exit; utilities.ts:10665:1 lowerBody_exit; visitorPublic.ts:420:1 lowerBody_exit.
- **C109** `visitorPublic.ts:474:1` stops at `utilities.ts:10656:6` (lowering, Refused): a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace. Propagation: factory/utilitiesPublic.ts:10:1 lowerBody_exit; utilities.ts:10655:1 lowerBody_exit; utilities.ts:10665:1 lowerBody_exit; visitorPublic.ts:474:1 lowerBody_exit.
- **C110** `watch.ts:986:1` stops at `watch.ts:988:33` (lowering, NotYet): an assignment value to a member. Propagation: watch.ts:986:1 lowerBody_exit.
- **C111** `watchUtilities.ts:118:1` stops at `watchUtilities.ts:156:13` (lowering, NotYet): writing a possibly absent optional own field. Propagation: watchUtilities.ts:118:1 lowerBody_exit; watchUtilities.ts:148:5 lowerBody_exit.
- **C112** `watchUtilities.ts:572:1` stops at `watchUtilities.ts:578:5` (lowering, NotYet): a destructured name held otherwise than its field. Propagation: watchUtilities.ts:572:1 lowerBody_exit.

## Broader full-mode attempt

A broader bounded full-mode attempt selected core.ts plus all 31 files containing hidden census callers of core.ts (32 files in one process). The selected list and guard script are retained. It ran for 1201.0 seconds and peaked at 7.08 GiB of cgroup memory, including page cache. Actual compiler child exit: -15. Stop reason: 20 minute bounded full-caller experiment budget.

This attempt is incomplete. It completed 1000 unit timing records and last completed `checker.ts:17857:5`. Records are buffered until the final census-wide flags can be attached; its output has no complete census or relation trailer. No reach number in this report comes from this partial run. The retained guard policy reserves an auditable publication checkpoint within the session. Its measured peak does not establish that the selected full census would finish within 16 GB.

## Reproduction and evidence

STEP1 preserves the baseline streams and setup. evidence/steps2-3 retains the instrumented streams, request logs, checker reference index, full-mode streams, caller traces, reader comparisons and their producing scripts. Go source evidence uses .go.txt. Actual child exit files are retained. All completed runs exited 0 except the intentionally failing witness mutant (1). No full79 process was started.

Run `python stage3/census/latent/audit_instantiated_by.py BASE HEAD OUT` for byte identity, and `python stage3/census/deferred-reach/audit_report.py` for the report ledger and denominators.

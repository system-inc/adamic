# 99n census and next stops

Compiler: `ab8a85bebf2a14a06fa8c3c4ca8cc93201a2309f`. All six independent measurements use the same 79 roots, selected `debug.ts`, `LATENT_FULL=1`, and `LATENT_ASSERT_NO_OUTPUT=1`. Every header has zero diagnostics and `checker_rejected: false`. The overlay retains the generic label "measured on a checker-rejected program"; that text does not override the header.

## Kept finding

The real pair tool reports 1 removed finding/read, 0 moved reasons, 0 new findings, 0 new sites, 0 first-stop movements, 0 regressions and 0 panics. Deduplicated Refused findings are 52 before and 51 after; NotYet is 63 on both sides. Raw recorded instances are 210 before and 209 after. CW157 maps from the sort input's debug.ts:1120:30 to this tree's debug.ts:1099:30. Only the existing helper signature changes.

The actual measured files differ by a 20-character type expansion on line 949. `pair.py` retains every event, proves all finding/unit columns are unchanged, maps unit body offsets through that expansion, and substitutes the absolute tree prefix. It then invokes the unmodified `stage3/census/pair/pair.py` against main coordinates. `projection.json` records both actual source hashes and both original stream hashes. The raw streams are retained separately; the projection is not a claim that their source bytes are identical.

## H07 exact condition

The coordinated H07 trial removes 18 refusal-scan invariance findings and 17 lowering invariance findings, 35 recorded instances at 18 view sites. The expected enum-object-to-record-storage NotYet count is **0**. The separate namespace-value scan stop remains at all **18** sites. Seventeen formatting declarations acquire `NotYet reading ts`; the worker remains blocked earlier. Two new storage NotYet scan findings occur at debug.ts:405:46 and :419:29, where readonly enumObject reaches the mutable cache key type. These two cache findings are not the requested enum-view storage stops. H07 is dropped, so its original invariance findings remain in the shipped tree. Lower owns the remaining namespace/storage work.

## All site decisions

| Site | Hand | Decision | Input owner |
| --- | --- | --- | --- |
| CW093 | H06 | dropped | assertEachNode |
| CW095 | H06 | dropped | assertNode |
| CW096 | H06 | dropped | assertNotNode |
| CW097 | H06 | dropped | assertOptionalNode |
| CW102 | H07 | dropped | formatSyntaxKind |
| CW104 | H07 | dropped | formatSnippetKind |
| CW106 | H07 | dropped | formatScriptKind |
| CW108 | H07 | dropped | formatNodeFlags |
| CW110 | H07 | dropped | formatNodeCheckFlags |
| CW112 | H07 | dropped | formatModifierFlags |
| CW114 | H07 | dropped | formatTransformFlags |
| CW116 | H07 | dropped | formatEmitFlags |
| CW118 | H07 | dropped | formatSymbolFlags |
| CW120 | H07 | dropped | formatTypeFlags |
| CW122 | H07 | dropped | formatSignatureFlags |
| CW124 | H07 | dropped | formatObjectFlags |
| CW126 | H07 | dropped | formatFlowFlags |
| CW128 | H07 | dropped | formatRelationComparisonResult |
| CW130 | H07 | dropped | formatCheckMode |
| CW132 | H07 | dropped | formatSignatureCheckMode |
| CW134 | H07 | dropped | formatTypeFacts |
| CW136 | H07 | dropped | attachFlowNodeDebugInfoWorker.get |
| CW155 | H08 | dropped | formatControlFlowGraph.buildGraphNode |
| CW156 | H09 | dropped | formatControlFlowGraph.computeColumnWidths |
| CW157 | H10 | kept | formatControlFlowGraph.renderFlowNode |
| CW158 | H11 | dropped | formatControlFlowGraph.renderGraph |

Reasons are recorded per site in `sites.json`, and every independent added/removed event is in `trial-deltas.json`. Dropped sites are unchanged in the shipped tree. H08 was not edited or represented as a successful lowering experiment.

## Every measured debug.ts declaration

These are all 105 census units, including overloads and deferred generic declarations. A blank finding means no lowering stop was recorded for that attempted unit, not proof of a complete native build. Full mode continues after failed statements; this table reports the first non-Boundary lowering finding per unit. These first stops are unchanged by H10.

| Declaration | Measured location | Status | First recorded lowering stop |
| --- | --- | --- | --- |
| KindImportDeclaration | compiler/debug.ts:1:1 | attempted | No stop recorded |
| KindImportDeclaration | compiler/debug.ts:2:1 | attempted | No stop recorded |
| LogLevel | compiler/debug.ts:97:1 | attempted | No stop recorded |
| LoggingHost | compiler/debug.ts:106:1 | attempted | No stop recorded |
| Debug | compiler/debug.ts:111:1 | attempted | NotYet compiler/debug.ts:399:29: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions |
| shouldLog | compiler/debug.ts:121:5 | attempted | No stop recorded |
| logMessage | compiler/debug.ts:125:5 | attempted | No stop recorded |
| log | compiler/debug.ts:131:5 | attempted | No stop recorded |
| error | compiler/debug.ts:136:9 | attempted | No stop recorded |
| warn | compiler/debug.ts:140:9 | attempted | No stop recorded |
| log | compiler/debug.ts:144:9 | attempted | No stop recorded |
| trace | compiler/debug.ts:148:9 | attempted | No stop recorded |
| getAssertionLevel | compiler/debug.ts:153:5 | attempted | No stop recorded |
| setAssertionLevel | compiler/debug.ts:157:5 | attempted | No stop recorded |
| shouldAssert | compiler/debug.ts:161:5 | attempted | No stop recorded |
| shouldAssertFunction | compiler/debug.ts:171:5 | deferred_generic_instantiation | Deferred; instantiated=true |
| fail | compiler/debug.ts:175:5 | attempted | No stop recorded |
| failBadSyntaxKind | compiler/debug.ts:185:5 | attempted | No stop recorded |
| assert | compiler/debug.ts:192:5 | attempted | No stop recorded |
| assertEqual | compiler/debug.ts:202:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertLessThan | compiler/debug.ts:209:5 | attempted | No stop recorded |
| assertLessThanOrEqual | compiler/debug.ts:215:5 | attempted | No stop recorded |
| assertGreaterThanOrEqual | compiler/debug.ts:221:5 | attempted | No stop recorded |
| assertIsDefined | compiler/debug.ts:227:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| checkDefined | compiler/debug.ts:234:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachIsDefined | compiler/debug.ts:239:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachIsDefined | compiler/debug.ts:240:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachIsDefined | compiler/debug.ts:241:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| checkEachDefined | compiler/debug.ts:247:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertNever | compiler/debug.ts:252:5 | attempted | NotYet compiler/debug.ts:252:33: a value of type never |
| assertEachNode | compiler/debug.ts:257:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachNode | compiler/debug.ts:258:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachNode | compiler/debug.ts:259:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachNode | compiler/debug.ts:260:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertEachNode | compiler/debug.ts:261:5 | attempted | Refused compiler/debug.ts:257:86: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test |
| assertEachNode | compiler/debug.ts:262:5 | attempted | Refused compiler/debug.ts:257:86: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test |
| assertNode | compiler/debug.ts:273:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertNode | compiler/debug.ts:274:5 | attempted | Refused compiler/debug.ts:273:82: overload 1 of assertNode parameter test cannot be served by implementation parameter test |
| assertNode | compiler/debug.ts:275:5 | attempted | Refused compiler/debug.ts:273:82: overload 1 of assertNode parameter test cannot be served by implementation parameter test |
| assertNotNode | compiler/debug.ts:286:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertNotNode | compiler/debug.ts:287:5 | attempted | NotYet compiler/debug.ts:294:35: an overloaded function as a value |
| assertNotNode | compiler/debug.ts:288:5 | attempted | NotYet compiler/debug.ts:294:35: an overloaded function as a value |
| assertOptionalNode | compiler/debug.ts:299:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertOptionalNode | compiler/debug.ts:300:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertOptionalNode | compiler/debug.ts:301:5 | attempted | Refused compiler/debug.ts:299:78: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test |
| assertOptionalNode | compiler/debug.ts:302:5 | attempted | Refused compiler/debug.ts:299:78: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test |
| assertOptionalToken | compiler/debug.ts:313:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertOptionalToken | compiler/debug.ts:314:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| assertOptionalToken | compiler/debug.ts:315:5 | attempted | NotYet compiler/debug.ts:322:35: an overloaded function as a value |
| assertOptionalToken | compiler/debug.ts:316:5 | attempted | NotYet compiler/debug.ts:322:35: an overloaded function as a value |
| assertMissingNode | compiler/debug.ts:327:5 | attempted | NotYet compiler/debug.ts:334:35: an overloaded function as a value |
| assertMissingNode | compiler/debug.ts:328:5 | attempted | NotYet compiler/debug.ts:334:35: an overloaded function as a value |
| type | compiler/debug.ts:344:5 | deferred_generic_instantiation | Deferred; instantiated=false |
| type | compiler/debug.ts:345:5 | attempted | No stop recorded |
| getFunctionName | compiler/debug.ts:347:5 | attempted | NotYet compiler/debug.ts:351:30: a function viewed as unknown or object (dynamic function descriptors) |
| formatSymbol | compiler/debug.ts:361:5 | attempted | NotYet compiler/debug.ts:362:135: a template interpolating an object, an array, a map, a function or undefined |
| formatEnum | compiler/debug.ts:368:5 | attempted | NotYet compiler/debug.ts:369:15: a value of type SortedReadonlyArray<[number, string]> |
| getEnumMembers | compiler/debug.ts:401:5 | attempted | NotYet compiler/debug.ts:401:14: a function returning SortedReadonlyArray<[number, string]> |
| formatSyntaxKind | compiler/debug.ts:423:5 | attempted | Refused compiler/debug.ts:424:34: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read |
| formatSnippetKind | compiler/debug.ts:427:5 | attempted | Refused compiler/debug.ts:428:34: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read |
| formatScriptKind | compiler/debug.ts:431:5 | attempted | Refused compiler/debug.ts:432:34: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read |
| formatNodeFlags | compiler/debug.ts:435:5 | attempted | Refused compiler/debug.ts:436:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read |
| formatNodeCheckFlags | compiler/debug.ts:439:5 | attempted | Refused compiler/debug.ts:440:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read |
| formatModifierFlags | compiler/debug.ts:443:5 | attempted | Refused compiler/debug.ts:444:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read |
| formatTransformFlags | compiler/debug.ts:447:5 | attempted | Refused compiler/debug.ts:448:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read |
| formatEmitFlags | compiler/debug.ts:451:5 | attempted | Refused compiler/debug.ts:452:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read |
| formatSymbolFlags | compiler/debug.ts:455:5 | attempted | Refused compiler/debug.ts:456:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read |
| formatTypeFlags | compiler/debug.ts:459:5 | attempted | Refused compiler/debug.ts:460:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read |
| formatSignatureFlags | compiler/debug.ts:463:5 | attempted | Refused compiler/debug.ts:464:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read |
| formatObjectFlags | compiler/debug.ts:467:5 | attempted | Refused compiler/debug.ts:468:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read |
| formatFlowFlags | compiler/debug.ts:471:5 | attempted | Refused compiler/debug.ts:472:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read |
| formatRelationComparisonResult | compiler/debug.ts:475:5 | attempted | Refused compiler/debug.ts:476:36: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read |
| formatCheckMode | compiler/debug.ts:479:5 | attempted | Refused compiler/debug.ts:480:34: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read |
| formatSignatureCheckMode | compiler/debug.ts:483:5 | attempted | Refused compiler/debug.ts:484:34: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read |
| formatTypeFacts | compiler/debug.ts:487:5 | attempted | Refused compiler/debug.ts:488:35: a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read |
| attachFlowNodeDebugInfoWorker | compiler/debug.ts:495:5 | attempted | Refused compiler/debug.ts:496:15: in that narrows a union of object types (any object with the key passes) |
| attachFlowNodeDebugInfo | compiler/debug.ts:531:5 | attempted | NotYet compiler/debug.ts:533:24: a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow |
| attachNodeArrayDebugInfoWorker | compiler/debug.ts:552:5 | attempted | NotYet compiler/debug.ts:553:15: in when a possibly absent spread has no property presence descriptors |
| attachNodeArrayDebugInfo | compiler/debug.ts:571:5 | attempted | NotYet compiler/debug.ts:573:24: a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow |
| enableDebugInfo | compiler/debug.ts:592:5 | attempted | NotYet compiler/debug.ts:596:33: new an Identifier |
| formatVariance | compiler/debug.ts:788:5 | attempted | No stop recorded |
| attachDebugPrototypeIfDebug | compiler/debug.ts:836:5 | attempted | Refused compiler/debug.ts:838:20: Object.setPrototypeOf |
| printControlFlowGraph | compiler/debug.ts:843:5 | attempted | No stop recorded |
| formatControlFlowGraph | compiler/debug.ts:847:5 | attempted | NotYet compiler/debug.ts:923:15: a value of type Record<number, FlowGraphNode> |
| getDebugFlowNodeId | compiler/debug.ts:850:9 | attempted | No stop recorded |
| isFlowSwitchClause | compiler/debug.ts:937:9 | attempted | No stop recorded |
| hasAntecedents | compiler/debug.ts:941:9 | attempted | Refused compiler/debug.ts:942:56: a cast the runtime can't check |
| hasAntecedent | compiler/debug.ts:945:9 | attempted | No stop recorded |
| hasNode | compiler/debug.ts:949:9 | attempted | No stop recorded |
| getChildren | compiler/debug.ts:953:9 | attempted | No stop recorded |
| getParents | compiler/debug.ts:963:9 | attempted | No stop recorded |
| buildGraphNode | compiler/debug.ts:973:9 | attempted | NotYet compiler/debug.ts:975:29: reading links |
| buildGraphEdge | compiler/debug.ts:1008:9 | attempted | No stop recorded |
| computeLevel | compiler/debug.ts:1016:9 | attempted | NotYet compiler/debug.ts:1024:20: returning a property or element assignment |
| computeHeight | compiler/debug.ts:1027:9 | attempted | No stop recorded |
| computeColumnWidths | compiler/debug.ts:1035:9 | attempted | NotYet compiler/debug.ts:1210:38: a value of type T |
| computeLanes | compiler/debug.ts:1043:9 | attempted | No stop recorded |
| getHeader | compiler/debug.ts:1060:9 | attempted | No stop recorded |
| getNodeText | compiler/debug.ts:1075:9 | attempted | No stop recorded |
| renderFlowNode | compiler/debug.ts:1080:9 | attempted | NotYet compiler/debug.ts:937:9: a checked predicate target whose check would read past the depth bound (FlowSwitchClause) |
| renderGraph | compiler/debug.ts:1107:9 | attempted | NotYet compiler/debug.ts:1178:17: assigning to a NonNullExpression |
| writeLane | compiler/debug.ts:1177:13 | attempted | NotYet compiler/debug.ts:1178:17: assigning to a NonNullExpression |
| getBoxCharacter | compiler/debug.ts:1182:9 | attempted | No stop recorded |
| fill | compiler/debug.ts:1210:9 | deferred_generic_instantiation | Deferred; instantiated=true |
| repeat | compiler/debug.ts:1222:9 | attempted | Refused compiler/debug.ts:1223:17: inherited library member repeat read as an own field |

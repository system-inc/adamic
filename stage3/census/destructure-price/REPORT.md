| Site (binding start) | Type | Public or internal | Members (`*` = method signature) | Calls per lane run |
| --- | --- | --- | --- | --- |
| src/compiler/sys.ts:590 | `RecursiveDirectoryWatcherHost` | Internal/private | `watchDirectory`, `useCaseSensitiveFileNames`, `getCurrentDirectory`, `getAccessibleSortedChildDirectories*`, `fileSystemEntryExists`, `realpath*`, `setTimeout`, `clearTimeout` | runs=30; watchDirectory=238; useCaseSensitiveFileNames=0; getCurrentDirectory=20; getAccessibleSortedChildDirectories=167; fileSystemEntryExists=316; realpath=379; setTimeout=83; clearTimeout=54 |
| src/compiler/sys.ts:993 | `CreateSystemWatchFunctions` | Internal/private | `pollingWatchFileWorker`, `getModifiedTime`, `setTimeout`, `clearTimeout`, `fsWatchWorker`, `fileSystemEntryExists`, `useCaseSensitiveFileNames`, `getCurrentDirectory`, `fsSupportsRecursiveFsWatch`, `getAccessibleSortedChildDirectories*`, `realpath*`, `tscWatchFile`, `useNonPollingWatchers`, `tscWatchDirectory`, `inodeWatching`, `fsWatchWithTimestamp`, `sysLog` | runs=4,785; pollingWatchFileWorker=4,786; getModifiedTime=1,180; setTimeout=358; clearTimeout=54; fsWatchWorker=13,330; fileSystemEntryExists=18,252; useCaseSensitiveFileNames=0; getCurrentDirectory=66; fsSupportsRecursiveFsWatch=0; getAccessibleSortedChildDirectories=167; realpath=379; tscWatchFile=0; useNonPollingWatchers=0; tscWatchDirectory=0; inodeWatching=0; fsWatchWithTimestamp=0; sysLog=168 |
| src/compiler/utilities.ts:8213 | `MutateMapSkippingNewValuesOptions<K, T, U>` | Internal/private | `onDeleteValue*`, `onExistingValue*` | runs=33,708; onDeleteValue=76; onExistingValue=326 |
| src/compiler/utilities.ts:8253 | `MutateMapOptions<K, T, U>` | Internal/private | `createNewValue*` | runs=8,913; createNewValue=2,855 |
| src/compiler/utilities.ts:11317 | `EvaluationResolver` | Internal/private | `evaluateElementAccessExpression*`, `evaluateEntityNameExpression*` | runs=86,185; evaluateElementAccessExpression=51; evaluateEntityNameExpression=4,493 |
| src/compiler/parser.ts:1472 | `NodeFactory` | Public | `createNodeArray*`, `createNumericLiteral*`, `createStringLiteral*`, `createLiteralLikeNode*`, `createIdentifier*`, `createPrivateIdentifier*`, `createToken*`, `createArrayLiteralExpression*`, `createObjectLiteralExpression*`, `createPropertyAccessExpression*`, `createPropertyAccessChain*`, `createElementAccessExpression*`, `createElementAccessChain*`, `createCallExpression*`, `createCallChain*`, `createNewExpression*`, `createParenthesizedExpression*`, `createBlock*`, `createVariableStatement*`, `createExpressionStatement*`, `createIfStatement*`, `createWhileStatement*`, `createForStatement*`, `createForOfStatement*`, `createVariableDeclaration*`, `createVariableDeclarationList*` | runs=9; createNodeArray=4,068,227; createNumericLiteral=377,494; createStringLiteral=375,394; createLiteralLikeNode=2,671; createIdentifier=8,254,453; createPrivateIdentifier=4,176; createToken=3,399,650; createArrayLiteralExpression=81,897; createObjectLiteralExpression=67,014; createPropertyAccessExpression=151,152; createPropertyAccessChain=812; createElementAccessExpression=40,053; createElementAccessChain=217; createCallExpression=106,406; createCallChain=394; createNewExpression=16,513; createParenthesizedExpression=44,643; createBlock=127,614; createVariableStatement=227,665; createExpressionStatement=184,328; createIfStatement=12,746; createWhileStatement=1,250; createForStatement=3,007; createForOfStatement=2,908; createVariableDeclaration=239,240; createVariableDeclarationList=232,784 |
| src/compiler/moduleSpecifiers.ts:583 | `ModuleSpecifierPreferences` | Internal/private | `getAllowedEndingsInPreferredOrder*`, `relativePreference`, `excludeRegexes` | runs=3,769; getAllowedEndingsInPreferredOrder=3,141; relativePreference=0; excludeRegexes=0 |
| src/compiler/sourcemap.ts:40 | `Timer` | Internal/private | `enter*`, `exit*` | runs=9,494; enter=3,659,380; exit=3,659,380 |
| src/compiler/transformers/ts.ts:236 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `startLexicalEnvironment*`, `resumeLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=49,199; factory=0; getEmitHelperFactory=666; startLexicalEnvironment=6,788; resumeLexicalEnvironment=1,214; endLexicalEnvironment=8,002; hoistVariableDeclaration=124 |
| src/compiler/transformers/classFields.ts:355 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `hoistVariableDeclaration*`, `endLexicalEnvironment*`, `startLexicalEnvironment*`, `resumeLexicalEnvironment*`, `addBlockScopedVariable*` | runs=49,199; factory=0; getEmitHelperFactory=1,891; hoistVariableDeclaration=3,402; endLexicalEnvironment=3,555; startLexicalEnvironment=974; resumeLexicalEnvironment=2,581; addBlockScopedVariable=20 |
| src/compiler/transformers/typeSerializer.ts:133 | `TransformationContext` | Public | `factory`, `hoistVariableDeclaration*` | runs=402; factory=0; hoistVariableDeclaration=38 |
| src/compiler/transformers/legacyDecorators.ts:89 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `hoistVariableDeclaration*` | runs=1,027; factory=0; getEmitHelperFactory=1,583; hoistVariableDeclaration=44 |
| src/compiler/transformers/esDecorators.ts:296 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `startLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=46,653; factory=0; getEmitHelperFactory=3,162; startLexicalEnvironment=1,503; endLexicalEnvironment=1,503; hoistVariableDeclaration=301 |
| src/compiler/transformers/es2017.ts:120 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `resumeLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=36,281; factory=0; getEmitHelperFactory=2,513; resumeLexicalEnvironment=2,513; endLexicalEnvironment=2,513; hoistVariableDeclaration=106 |
| src/compiler/transformers/es2018.ts:155 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `resumeLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=36,628; factory=0; getEmitHelperFactory=2,215; resumeLexicalEnvironment=26,640; endLexicalEnvironment=26,640; hoistVariableDeclaration=356 |
| src/compiler/transformers/es2020.ts:43 | `TransformationContext` | Public | `factory`, `hoistVariableDeclaration*` | runs=36,871; factory=0; hoistVariableDeclaration=432 |
| src/compiler/transformers/es2021.ts:26 | `TransformationContext` | Public | `hoistVariableDeclaration*`, `factory` | runs=37,105; hoistVariableDeclaration=36; factory=0 |
| src/compiler/transformers/esnext.ts:77 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `hoistVariableDeclaration*`, `startLexicalEnvironment*`, `endLexicalEnvironment*` | runs=47,449; factory=0; getEmitHelperFactory=2,081; hoistVariableDeclaration=626; startLexicalEnvironment=340; endLexicalEnvironment=340 |
| src/compiler/transformers/jsx.ts:99 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*` | runs=590; factory=0; getEmitHelperFactory=148 |
| src/compiler/transformers/es2016.ts:22 | `TransformationContext` | Public | `factory`, `hoistVariableDeclaration*` | runs=36,212; factory=0; hoistVariableDeclaration=228 |
| src/compiler/transformers/es2015.ts:489 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `startLexicalEnvironment*`, `resumeLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=3,754; factory=0; getEmitHelperFactory=779; startLexicalEnvironment=7,183; resumeLexicalEnvironment=10,368; endLexicalEnvironment=17,551; hoistVariableDeclaration=431 |
| src/compiler/transformers/generators.ts:337 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `resumeLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistFunctionDeclaration*`, `hoistVariableDeclaration*` | runs=3,754; factory=0; getEmitHelperFactory=1,029; resumeLexicalEnvironment=1,011; endLexicalEnvironment=1,011; hoistFunctionDeclaration=6; hoistVariableDeclaration=1,085 |
| src/compiler/transformers/module/module.ts:192 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*`, `startLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=48,396; factory=0; getEmitHelperFactory=2,845; startLexicalEnvironment=11,206; endLexicalEnvironment=11,206; hoistVariableDeclaration=24 |
| src/compiler/transformers/module/system.ts:143 | `TransformationContext` | Public | `factory`, `startLexicalEnvironment*`, `endLexicalEnvironment*`, `hoistVariableDeclaration*` | runs=618; factory=0; startLexicalEnvironment=632; endLexicalEnvironment=632; hoistVariableDeclaration=1,229 |
| src/compiler/transformers/module/esnextAnd2015.ts:68 | `TransformationContext` | Public | `factory`, `getEmitHelperFactory*` | runs=45,988; factory=0; getEmitHelperFactory=15,905 |
| src/compiler/emitter.ts:771 | `Timer` | Internal/private | `enter*`, `exit*` | runs=54,954; enter=54,954; exit=54,950 |
| src/compiler/emitter.ts:1215 | `PrintHandlers` | Public | `hasGlobalName*`, `onEmitNode*`, `isEmitNotificationEnabled*`, `substituteNode*`, `onBeforeEmitNode`, `onAfterEmitNode`, `onBeforeEmitNodeArray`, `onAfterEmitNodeArray`, `onBeforeEmitToken`, `onAfterEmitToken` | runs=73,916; hasGlobalName=49,960; onEmitNode=4,628,553; isEmitNotificationEnabled=3,975,026; substituteNode=3,986,302; onBeforeEmitNode=451,527; onAfterEmitNode=451,527; onBeforeEmitNodeArray=110,298; onAfterEmitNodeArray=110,298; onBeforeEmitToken=6,180; onAfterEmitToken=6,180 |
| src/compiler/emitter.ts:1275 | `Timer` | Internal/private | `enter*`, `exit*` | runs=73,916; enter=10,552,068; exit=10,552,068 |
| src/compiler/expressionToTypeNode.ts:199 | `Recovery boundary (anonymous)` | Internal/private | `finalizeBoundary*`, `startRecoveryScope*`, `hadError*`, `markError*` | runs=184,002; finalizeBoundary=184,001; startRecoveryScope=332,198; hadError=664,647; markError=1,093 |

Certified for the single complete oracle run below. `runs` means enclosing-body entries; parser runs are **namespace initializations (9), not parsed-file counts**. Non-callable values have zero calls. Existing function properties are counted as well as method signatures; only `*` members on public receiver types incur the proposed extra calls. Alias-local names, exact declarations and positions are in [heat.json](evidence/heat.json).

## Price

| Road | Public sites / wrappers | Extra calls per measured lane | Stock strictFunctionTypes diagnostics (79 roots) | Declaration result |
| --- | --- | ---: | ---: | --- |
| Baseline | — | 0 | 0 | Adapted baseline |
| All 11 receiver types: taken methods become properties | — | 0; JavaScript fingerprint unchanged for all 5 edited files | 1 (TS2322) | 34 public members change; 3 internal overloads exposed; normal `hereby dts` fails |
| Hybrid: internal/private types as properties, public sites as receiver arrows | 19 / 82 | **30,851,760** | **0** | Normal `hereby dts` passes; both public files byte-identical |

Public-call breakdown: parser **18,022,708**, printer handlers **12,639,841**, the 17 transformer bindings **189,211**. Total counted local-member calls across all 29 sites, including existing function properties, were **61,753,448**. This is one observed run with the pinned workload and batching, not a universal constant or a wall-time forecast.

The extra-call total is the measured number of calls through public method bindings multiplied by the proposal’s one forwarding call per use. Counter wall time is not arrow-overhead timing. Source recipes are private measurement experiments, not an Adamic adaptation or evidence of native compilation. No compiler implementation or rule changed.

## Workload, controls and counter validity

The unfiltered oracle used all runners and eight workers, with its unchanged **40,000 ms** hook timeout. Results were **106,366 passing, one failing, zero pending**. The sole failure was `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`; the sole baseline diff was `api/typescript.d.ts`. Install/build exited zero and tests exited one as expected. All eight worker snapshots were present; controls and discarded attempts are excluded. See [oracle-summary.json](evidence/oracle-summary.json).

The counter wraps each local **after destructuring**, including defaults; it never proxies the source object. Strict forwarding functions preserve the receiver and argument list and retain name/length metadata. Original values are retained for six symbol-resolved identity comparisons: three printer sentinel checks and three non-callable preference comparisons. This preserves the default-sentinel shortcuts instead of creating artificial calls. Every function-valued taken member is counted, not just methods.

Before the accepted lane, the default control recorded **1/1**, the supplied callback control **1/1**, and the parser/printer printed `const answer = 42;`. Sentinel controls confirmed default notification/substitution calls stayed zero; bare and explicit receivers were preserved. [controls.cjs](controls.cjs) and [controls-result.json](evidence/controls-result.json) retain the assertions and observations.

All **80,520** test input/reference files were verified byte-identical to the baseline. Thirty nested node_modules fixture files and two nested built reference files omitted by an early copy were restored. Scratch was moved from tmpfs to disk; the private `.parallelperf.json` history was cleared for fresh-checkout batching. Incomplete, fixture-deficient, timed-out and OOM attempts were discarded. [test-input-verification.json](evidence/test-input-verification.json) records the accepted input inventory hash.

## Full latent census and AST comparison

All **24 files** holding the 29 AST sites were measured separately with `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1`, one exact `LATENT_ONLY_FILE` per run, **CommonJSBundle**, a **3,600 second** limit per file and no overlapping measurement jobs. Every run exited zero, emitted exactly one selected-file record and ended with a relation trailer; none timed out.

The census emitted **18 raw events** for `a method in object destructuring`, at **16 distinct locations / binding sites**. The extra two events repeat the module-specifier binding. No other reason is folded into that count. There are **13 AST-only sites** and **zero census-only sites** for the broader scan. The counter and hybrid price cover all 29 AST sites, not merely the 16 observed compiler refusals.

| Selected file | AST sites | Distinct method refusals | Raw events | Seconds | Peak summed RSS (KiB) |
| --- | ---: | ---: | ---: | ---: | ---: |
| src/compiler/emitter.ts | 3 | 0 | 0 | 693.389 | 1,681,732 |
| src/compiler/expressionToTypeNode.ts | 1 | 1 | 1 | 90.364 | 1,100,724 |
| src/compiler/moduleSpecifiers.ts | 1 | 1 | 3 | 54.005 | 1,025,156 |
| src/compiler/parser.ts | 1 | 1 | 1 | 652.683 | 2,161,384 |
| src/compiler/sourcemap.ts | 1 | 0 | 0 | 47.663 | 907,804 |
| src/compiler/sys.ts | 2 | 2 | 2 | 96.875 | 995,512 |
| src/compiler/transformers/classFields.ts | 1 | 0 | 0 | 138.695 | 1,150,156 |
| src/compiler/transformers/es2015.ts | 1 | 0 | 0 | 205.727 | 1,368,484 |
| src/compiler/transformers/es2016.ts | 1 | 1 | 1 | 27.720 | 920,864 |
| src/compiler/transformers/es2017.ts | 1 | 1 | 1 | 103.318 | 1,131,284 |
| src/compiler/transformers/es2018.ts | 1 | 0 | 0 | 109.306 | 1,210,852 |
| src/compiler/transformers/es2020.ts | 1 | 1 | 1 | 48.062 | 997,680 |
| src/compiler/transformers/es2021.ts | 1 | 1 | 1 | 23.954 | 930,352 |
| src/compiler/transformers/esDecorators.ts | 1 | 0 | 0 | 104.988 | 1,087,996 |
| src/compiler/transformers/esnext.ts | 1 | 1 | 1 | 79.550 | 1,050,816 |
| src/compiler/transformers/generators.ts | 1 | 0 | 0 | 109.026 | 1,147,668 |
| src/compiler/transformers/jsx.ts | 1 | 1 | 1 | 74.555 | 1,040,548 |
| src/compiler/transformers/legacyDecorators.ts | 1 | 1 | 1 | 84.288 | 1,042,532 |
| src/compiler/transformers/module/esnextAnd2015.ts | 1 | 1 | 1 | 62.214 | 1,021,900 |
| src/compiler/transformers/module/module.ts | 1 | 0 | 0 | 89.470 | 1,139,448 |
| src/compiler/transformers/module/system.ts | 1 | 1 | 1 | 136.579 | 1,288,056 |
| src/compiler/transformers/ts.ts | 1 | 0 | 0 | 138.699 | 1,159,592 |
| src/compiler/transformers/typeSerializer.ts | 1 | 1 | 1 | 49.038 | 1,002,956 |
| src/compiler/utilities.ts | 3 | 1 | 1 | 494.161 | 1,851,384 |

The full typed scan covers all **79 compiler roots** and **54 object-binding files**. A literal initializer-only scan has **25 sites** and misses the four destructured parameters at `sys.ts:590`, `sys.ts:993`, `moduleSpecifiers.ts:583` and `utilities.ts:11317`. Those four are observed by the census. Including parameter types yields the 29-site inventory. A fresh scan reproduced all original positions.

| AST criterion | Sites | Census matches | AST-only | Census-only |
| --- | ---: | ---: | ---: | ---: |
| Initializer type only | 25 | 12 | 13 | 4 parameters |
| Initializer or parameter type | 29 | 16 | 13 | 0 |

The following are measured coverage gaps, not additional method refusals:

| AST-only site | Why this census does not reach the method check |
| --- | --- |
| src/compiler/utilities.ts:8213 | deferred_generic_instantiation (no concrete instantiation) |
| src/compiler/utilities.ts:8253 | deferred_generic_instantiation (no concrete instantiation) |
| src/compiler/sourcemap.ts:40 | src/compiler/sourcemap.ts:40:5: Adamic 0.1 refuses var; use const or let |
| src/compiler/transformers/ts.ts:236 | src/compiler/transformers/ts.ts:1430:55: stage 0 can't lower a value of type ParameterPropertyDeclaration yet |
| src/compiler/transformers/classFields.ts:355 | src/compiler/transformers/classFields.ts:671:55: stage 0 can't lower a value of type PrivateIdentifierInExpression yet |
| src/compiler/transformers/esDecorators.ts:296 | src/compiler/transformers/esDecorators.ts:670:14: stage 0 can't lower a function returning ImmediatelyInvokedArrowFunction yet |
| src/compiler/transformers/es2018.ts:155 | src/compiler/transformers/es2018.ts:1233:36: stage 0 can't lower a branded supertype inside a checked predicate's array target yet |
| src/compiler/transformers/es2015.ts:489 | src/compiler/transformers/es2015.ts:825:14: stage 0 can't lower a function returning CapturedThis yet |
| src/compiler/transformers/generators.ts:337 | src/compiler/transformers/generators.ts:1387:43: stage 0 can't lower a value of type InitializedVariableDeclaration yet |
| src/compiler/transformers/module/module.ts:192 | src/compiler/transformers/module/module.ts:1926:43: stage 0 can't lower a value of type InitializedVariableDeclaration yet |
| src/compiler/emitter.ts:771 | src/compiler/emitter.ts:771:5: Adamic 0.1 refuses var; use const or let |
| src/compiler/emitter.ts:1215 | src/compiler/emitter.ts:1215:5: Adamic 0.1 refuses var; use const or let |
| src/compiler/emitter.ts:1275 | src/compiler/emitter.ts:1275:5: Adamic 0.1 refuses var; use const or let |

The gaps comprise **4 var gates, 7 function signature/prologue failures and 2 deferred generic bodies**. [reconciliation.json](evidence/reconciliation.json) records each span, enclosing declaration, observed refusal or precise boundary. Headers have empty diagnostic arrays and `checker_rejected: false`; the overlay’s fixed historical label “measured on a checker-rejected program” does not establish rejection here. The no-output guards still run.

## Receiver declarations and public status

`typescript.d.ts` was built in the adapted checkout after `npm ci` and `node_modules/.bin/hereby dts`. `tsserverlibrary.d.ts` literally contains only the re-export below: public API presence follows that re-export, not duplicate interface text. The public/private decision is by receiver type as proposed, including internal members of public types.

```typescript
import ts = require("./typescript.js");
export = ts;
```

| Receiver type | Source declaration | Source visibility | typescript.d.ts | tsserverlibrary.d.ts API |
| --- | --- | --- | --- | --- |
| `RecursiveDirectoryWatcherHost` | src/compiler/sys.ts:574:1 | Private, non-exported; no @internal tag | Absent | Absent |
| `CreateSystemWatchFunctions` | src/compiler/sys.ts:968:1 | @internal on CreateSystemWatchFunctions | Absent | Absent |
| `MutateMapSkippingNewValuesOptions<K, T, U>` | src/compiler/utilities.ts:8182:1 | @internal on MutateMapSkippingNewValuesOptions | Absent | Absent |
| `MutateMapOptions<K, T, U>` | src/compiler/utilities.ts:8238:1 | @internal on MutateMapOptions | Absent | Absent |
| `EvaluationResolver` | src/compiler/types.ts:10618:1 | @internal on EvaluationResolver | Absent | Absent |
| `NodeFactory` | src/compiler/types.ts:8748:1 | Public | Present | Re-exported |
| `ModuleSpecifierPreferences` | src/compiler/moduleSpecifiers.ts:172:1 | @internal on ModuleSpecifierPreferences | Absent | Absent |
| `Timer` | src/compiler/performance.ts:23:1 | @internal on Timer | Absent | Absent |
| `TransformationContext` | src/compiler/types.ts:9616:1 | Public | Present | Re-exported |
| `PrintHandlers` | src/compiler/types.ts:9820:1 | Public | Present | Re-exported |
| `Recovery boundary (anonymous)` | src/compiler/types.ts:10669:71 | @internal on SyntacticTypeNodeBuilderResolver | Absent | Absent |

The anonymous boundary belongs to `SyntacticTypeNodeBuilderResolver.createRecoveryBoundary`. RecursiveDirectoryWatcherHost is private rather than itself tagged @internal; its absence from the public bundles places it on the internal/private road.

## Enclosing functions and implementation sources

Parser binding calls follow `createSourceFile → Parser.parseSourceFile → parseSourceFileWorker → parse… → factoryCreate*`. Its binding lives in namespace initialization. The named transformer paths and all other enclosing functions are listed here; counts remain in the leading table.

| Site | Enclosing function / scope |
| --- | --- |
| src/compiler/sys.ts:590 | `createDirectoryWatcherSupportingRecursive` |
| src/compiler/sys.ts:993 | `createSystemWatchFunctions` |
| src/compiler/utilities.ts:8213 | `mutateMapSkippingNewValues` |
| src/compiler/utilities.ts:8253 | `mutateMap` |
| src/compiler/utilities.ts:11317 | `createEvaluator` |
| src/compiler/parser.ts:1472 | `Parser namespace initialization` |
| src/compiler/moduleSpecifiers.ts:583 | `getLocalModuleSpecifier` |
| src/compiler/sourcemap.ts:40 | `createSourceMapGenerator` |
| src/compiler/transformers/ts.ts:236 | `transformTypeScript` |
| src/compiler/transformers/classFields.ts:355 | `transformClassFields` |
| src/compiler/transformers/typeSerializer.ts:133 | `createRuntimeTypeSerializer` |
| src/compiler/transformers/legacyDecorators.ts:89 | `transformLegacyDecorators` |
| src/compiler/transformers/esDecorators.ts:296 | `transformESDecorators` |
| src/compiler/transformers/es2017.ts:120 | `transformES2017` |
| src/compiler/transformers/es2018.ts:155 | `transformES2018` |
| src/compiler/transformers/es2020.ts:43 | `transformES2020` |
| src/compiler/transformers/es2021.ts:26 | `transformES2021` |
| src/compiler/transformers/esnext.ts:77 | `transformESNext` |
| src/compiler/transformers/jsx.ts:99 | `transformJsx` |
| src/compiler/transformers/es2016.ts:22 | `transformES2016` |
| src/compiler/transformers/es2015.ts:489 | `transformES2015` |
| src/compiler/transformers/generators.ts:337 | `transformGenerators` |
| src/compiler/transformers/module/module.ts:192 | `transformModule` |
| src/compiler/transformers/module/system.ts:143 | `transformSystemModule` |
| src/compiler/transformers/module/esnextAnd2015.ts:68 | `transformECMAScriptModule` |
| src/compiler/emitter.ts:771 | `emitFiles` |
| src/compiler/emitter.ts:1215 | `createPrinter` |
| src/compiler/emitter.ts:1275 | `createPrinter` |
| src/compiler/expressionToTypeNode.ts:199 | `tryReuseExistingTypeNode` |

The source supplier audit spans **604 src roots**, including compiler, server, harness and unit-test adapters. It follows contextual object construction, class heritage, direct arguments, direct assignments, Object.assign overrides and the recovery-boundary return object. It has **zero unresolved indexed callable edges**. [implementation-index.json](evidence/implementation-index.json) and the following tables enumerate the resolved source definitions per member. This is a conservative source index, not a closed-world proof for future external public API callers or arbitrary dynamic reflection. Conditional factory variants are identified explicitly: real createTimer sites exclude noop; parser factory flags exclude the NoIndentationOnFreshPropertyAccess wrappers. Other type-boundary candidates are retained.

`this` means the implementation’s own receiver use; nested normal-function/class scopes are excluded and nested arrows are included. Arrow `this` is lexical. Bound entries retain their target and fixed receiver. No resolved plain-function method supplier reads its own `this`; the this-reading method suppliers are lexical arrows or a bound class target. Non-callable members have no function implementation / this observation.

| Site | Taken callable member → implementation definitions |
| --- | --- |
| src/compiler/sys.ts:590 | watchDirectory → F001; getCurrentDirectory → F002, F003; getAccessibleSortedChildDirectories → F005, F006; fileSystemEntryExists → F007, F008; realpath → F010, F011; setTimeout → F013, F014; clearTimeout → F016, F017 |
| src/compiler/sys.ts:993 | pollingWatchFileWorker → F019, F020; getModifiedTime → F022, F023; setTimeout → F013, F014; clearTimeout → F016, F017; fsWatchWorker → F025, F026; fileSystemEntryExists → F007, F008; getCurrentDirectory → F002, F003; getAccessibleSortedChildDirectories → F005, F006; realpath → F010, F011; sysLog → F028, F029 |
| src/compiler/utilities.ts:8213 | onDeleteValue → F030, F031, F032, F033, F034; onExistingValue → F035 |
| src/compiler/utilities.ts:8253 | createNewValue → F036, F037, F038, F039, F040, F041 |
| src/compiler/utilities.ts:11317 | evaluateElementAccessExpression → F042; evaluateEntityNameExpression → F043 |
| src/compiler/parser.ts:1472 | createNodeArray → F044; createNumericLiteral → F045; createStringLiteral → F046; createLiteralLikeNode → F047; createIdentifier → F048; createPrivateIdentifier → F049; createToken → F050; createArrayLiteralExpression → F051; createObjectLiteralExpression → F052; createPropertyAccessExpression → F053; createPropertyAccessChain → F054; createElementAccessExpression → F055; createElementAccessChain → F056; createCallExpression → F057; createCallChain → F058; createNewExpression → F059; createParenthesizedExpression → F060; createBlock → F061; createVariableStatement → F062; createExpressionStatement → F063; createIfStatement → F064; createWhileStatement → F065; createForStatement → F066; createForOfStatement → F067; createVariableDeclaration → F068; createVariableDeclarationList → F069 |
| src/compiler/moduleSpecifiers.ts:583 | getAllowedEndingsInPreferredOrder → F070 |
| src/compiler/sourcemap.ts:40 | enter → F071; exit → F072 |
| src/compiler/transformers/ts.ts:236 | getEmitHelperFactory → F002, F073; startLexicalEnvironment → F074, F075; resumeLexicalEnvironment → F076, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/classFields.ts:355 | getEmitHelperFactory → F002, F073; hoistVariableDeclaration → F079, F075; endLexicalEnvironment → F077, F078; startLexicalEnvironment → F074, F075; resumeLexicalEnvironment → F076, F075; addBlockScopedVariable → F080, F075 |
| src/compiler/transformers/typeSerializer.ts:133 | hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/legacyDecorators.ts:89 | getEmitHelperFactory → F002, F073; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/esDecorators.ts:296 | getEmitHelperFactory → F002, F073; startLexicalEnvironment → F074, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/es2017.ts:120 | getEmitHelperFactory → F002, F073; resumeLexicalEnvironment → F076, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/es2018.ts:155 | getEmitHelperFactory → F002, F073; resumeLexicalEnvironment → F076, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/es2020.ts:43 | hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/es2021.ts:26 | hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/esnext.ts:77 | getEmitHelperFactory → F002, F073; hoistVariableDeclaration → F079, F075; startLexicalEnvironment → F074, F075; endLexicalEnvironment → F077, F078 |
| src/compiler/transformers/jsx.ts:99 | getEmitHelperFactory → F002, F073 |
| src/compiler/transformers/es2016.ts:22 | hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/es2015.ts:489 | getEmitHelperFactory → F002, F073; startLexicalEnvironment → F074, F075; resumeLexicalEnvironment → F076, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/generators.ts:337 | getEmitHelperFactory → F002, F073; resumeLexicalEnvironment → F076, F075; endLexicalEnvironment → F077, F078; hoistFunctionDeclaration → F081, F075; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/module/module.ts:192 | getEmitHelperFactory → F002, F073; startLexicalEnvironment → F074, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/module/system.ts:143 | startLexicalEnvironment → F074, F075; endLexicalEnvironment → F077, F078; hoistVariableDeclaration → F079, F075 |
| src/compiler/transformers/module/esnextAnd2015.ts:68 | getEmitHelperFactory → F002, F073 |
| src/compiler/emitter.ts:771 | enter → F071; exit → F072 |
| src/compiler/emitter.ts:1215 | hasGlobalName → F082, F073; onEmitNode → F083, F084, F085; isEmitNotificationEnabled → F086; substituteNode → F087, F088; onBeforeEmitNode → F089; onAfterEmitNode → F090; onBeforeEmitNodeArray → F091; onAfterEmitNodeArray → F092; onBeforeEmitToken → F093; onAfterEmitToken → F094 |
| src/compiler/emitter.ts:1275 | enter → F071, F075; exit → F072, F075 |
| src/compiler/expressionToTypeNode.ts:199 | finalizeBoundary → F095; startRecoveryScope → F096; hadError → F097; markError → F098 |

| ID | Source location | Kind | Reads this | Target / receiver or name |
| --- | --- | --- | --- | --- |
| F001 | src/compiler/sys.ts:1132:5 | plain function | no | nonRecursiveWatchDirectory |
| F002 | src/compiler/core.ts:1896:12 | arrow | no |  |
| F003 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:473:34 | bound | fixed binding; inspect target | targets F004; receiver this |
| F004 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1282:36 | arrow | lexical yes |  |
| F005 | src/compiler/sys.ts:1508:50 | arrow | no |  |
| F006 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:475:50 | arrow | lexical yes |  |
| F007 | src/compiler/sys.ts:1880:9 | plain function | no | fileSystemEntryExists |
| F008 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:471:36 | bound | fixed binding; inspect target | targets F009; receiver this |
| F009 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1022:5 | class method | yes | fileSystemEntryExists |
| F010 | src/compiler/sys.ts:1911:9 | plain function | no | realpath |
| F011 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:476:23 | bound | fixed binding; inspect target | targets F012; receiver this |
| F012 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1261:5 | class method | yes | realpath |
| F013 | src/compiler/sys.ts:1499:25 | arrow | no |  |
| F014 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:25 | bound | fixed binding; inspect target | targets F015; receiver this |
| F015 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103:5 | class method | yes | setTimeout |
| F016 | src/compiler/sys.ts:1500:27 | arrow | no |  |
| F017 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:27 | bound | fixed binding; inspect target | targets F018; receiver this |
| F018 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111:5 | class method | yes | clearTimeout |
| F019 | src/compiler/sys.ts:1733:9 | plain function | no | fsWatchFileWorker |
| F020 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:466:37 | bound | fixed binding; inspect target | targets F021; receiver this |
| F021 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:866:5 | class method | yes | watchFileWorker |
| F022 | src/compiler/sys.ts:1920:9 | plain function | no | getModifiedTime |
| F023 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:467:30 | bound | fixed binding; inspect target | targets F024; receiver this |
| F024 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1031:5 | class method | yes | getModifiedTime |
| F025 | src/compiler/sys.ts:1766:9 | plain function | no | fsWatchWorker |
| F026 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:470:28 | bound | fixed binding; inspect target | targets F027; receiver this |
| F027 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:873:5 | class method | yes | fsWatchWorker |
| F028 | src/compiler/sys.ts:565:1 | plain function | no | sysLog |
| F029 | src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:481:21 | arrow | lexical yes |  |
| F030 | src/compiler/utilities.ts:8033:1 | plain function | no | closeFileWatcher |
| F031 | src/compiler/watchUtilities.ts:845:1 | plain function | no | closeFileWatcherOf |
| F032 | src/compiler/tsbuildPublic.ts:703:30 | arrow | no |  |
| F033 | src/compiler/tsbuildPublic.ts:709:30 | arrow | no |  |
| F034 | src/compiler/tsbuildPublic.ts:715:30 | arrow | no |  |
| F035 | src/compiler/watchUtilities.ts:545:5 | plain function | no | updateWildcardDirectoryWatcher |
| F036 | src/compiler/watchPublic.ts:1086:5 | plain function | no | watchMissingFilePath |
| F037 | src/server/project.ts:1704:17 | arrow | lexical yes |  |
| F038 | src/compiler/watchUtilities.ts:537:5 | plain function | no | createWildcardDirectoryWatcher |
| F039 | src/compiler/resolutionCache.ts:1632:5 | plain function | no | createTypeRootsWatch |
| F040 | src/compiler/tsbuildPublic.ts:2193:29 | arrow | no |  |
| F041 | src/compiler/tsbuildPublic.ts:2214:29 | arrow | no |  |
| F042 | src/compiler/checker.ts:48013:5 | plain function | no | evaluateElementAccessExpression |
| F043 | src/compiler/checker.ts:47978:5 | plain function | no | evaluateEntityNameExpression |
| F044 | src/compiler/factory/nodeFactory.ts:1169:5 | plain function | no | createNodeArray |
| F045 | src/compiler/factory/nodeFactory.ts:1233:5 | plain function | no | createNumericLiteral |
| F046 | src/compiler/factory/nodeFactory.ts:1259:5 | plain function | no | createStringLiteral |
| F047 | src/compiler/factory/nodeFactory.ts:1281:5 | plain function | no | createLiteralLikeNode |
| F048 | src/compiler/factory/nodeFactory.ts:1326:5 | plain function | no | createIdentifier |
| F049 | src/compiler/factory/nodeFactory.ts:1396:5 | plain function | no | createPrivateIdentifier |
| F050 | src/compiler/factory/nodeFactory.ts:1453:5 | plain function | no | createToken |
| F051 | src/compiler/factory/nodeFactory.ts:2848:5 | plain function | no | createArrayLiteralExpression |
| F052 | src/compiler/factory/nodeFactory.ts:2869:5 | plain function | no | createObjectLiteralExpression |
| F053 | src/compiler/factory/nodeFactory.ts:2903:5 | plain function | no | createPropertyAccessExpression |
| F054 | src/compiler/factory/nodeFactory.ts:2930:5 | plain function | no | createPropertyAccessChain |
| F055 | src/compiler/factory/nodeFactory.ts:2968:5 | plain function | no | createElementAccessExpression |
| F056 | src/compiler/factory/nodeFactory.ts:2995:5 | plain function | no | createElementAccessChain |
| F057 | src/compiler/factory/nodeFactory.ts:3038:5 | plain function | no | createCallExpression |
| F058 | src/compiler/factory/nodeFactory.ts:3064:5 | plain function | no | createCallChain |
| F059 | src/compiler/factory/nodeFactory.ts:3088:5 | plain function | no | createNewExpression |
| F060 | src/compiler/factory/nodeFactory.ts:3160:5 | plain function | no | createParenthesizedExpression |
| F061 | src/compiler/factory/nodeFactory.ts:3856:5 | plain function | no | createBlock |
| F062 | src/compiler/factory/nodeFactory.ts:3876:5 | plain function | no | createVariableStatement |
| F063 | src/compiler/factory/nodeFactory.ts:3907:5 | plain function | no | createExpressionStatement |
| F064 | src/compiler/factory/nodeFactory.ts:3925:5 | plain function | no | createIfStatement |
| F065 | src/compiler/factory/nodeFactory.ts:3970:5 | plain function | no | createWhileStatement |
| F066 | src/compiler/factory/nodeFactory.ts:3991:5 | plain function | no | createForStatement |
| F067 | src/compiler/factory/nodeFactory.ts:4046:5 | plain function | no | createForOfStatement |
| F068 | src/compiler/factory/nodeFactory.ts:4251:5 | plain function | no | createVariableDeclaration |
| F069 | src/compiler/factory/nodeFactory.ts:4276:5 | plain function | no | createVariableDeclarationList |
| F070 | src/compiler/moduleSpecifiers.ts:199:44 | arrow | no |  |
| F071 | src/compiler/performance.ts:41:5 | plain function | no | enter |
| F072 | src/compiler/performance.ts:47:5 | plain function | no | exit |
| F073 | src/compiler/core.ts:1889:1 | plain function | no | notImplemented |
| F074 | src/compiler/transformer.ts:480:5 | plain function | no | startLexicalEnvironment |
| F075 | src/compiler/core.ts:1797:1 | plain function | no | noop |
| F076 | src/compiler/transformer.ts:509:5 | plain function | no | resumeLexicalEnvironment |
| F077 | src/compiler/transformer.ts:520:5 | plain function | no | endLexicalEnvironment |
| F078 | src/compiler/core.ts:1822:1 | plain function | no | returnUndefined |
| F079 | src/compiler/transformer.ts:431:5 | plain function | no | hoistVariableDeclaration |
| F080 | src/compiler/transformer.ts:621:5 | plain function | no | addBlockScopedVariable |
| F081 | src/compiler/transformer.ts:449:5 | plain function | no | hoistFunctionDeclaration |
| F082 | src/compiler/checker.ts:51235:5 | plain function | no | hasGlobalName |
| F083 | src/compiler/transformer.ts:414:5 | plain function | no | emitNodeWithNotification |
| F084 | src/harness/typeWriter.ts:54:21 | arrow | no |  |
| F085 | src/compiler/transformer.ts:232:1 | plain function | no | noEmitNotification |
| F086 | src/compiler/transformer.ts:402:5 | plain function | no | isEmitNotificationEnabled |
| F087 | src/compiler/transformer.ts:385:5 | plain function | no | substituteNode |
| F088 | src/compiler/transformer.ts:227:1 | plain function | no | noEmitSubstitution |
| F089 | src/services/textChanges.ts:1422:65 | arrow | no |  |
| F090 | src/services/textChanges.ts:1427:63 | arrow | no |  |
| F091 | src/services/textChanges.ts:1432:75 | arrow | no |  |
| F092 | src/services/textChanges.ts:1437:73 | arrow | no |  |
| F093 | src/services/textChanges.ts:1442:67 | arrow | no |  |
| F094 | src/services/textChanges.ts:1447:65 | arrow | no |  |
| F095 | src/compiler/checker.ts:8140:13 | plain function | no | finalizeBoundary |
| F096 | src/compiler/checker.ts:8125:13 | plain function | no | startRecoveryScope |
| F097 | src/compiler/checker.ts:8116:27 | arrow | no |  |
| F098 | src/compiler/checker.ts:8118:13 | plain function | no | markError |

## Stock strictFunctionTypes and declaration experiments

All ledgers use the actual **stock tsc 6.0.3 CLI**, the 79-root compiler tsconfig and explicit `--strictFunctionTypes true --noEmit --incremental false --composite false --pretty false --noErrorTruncation`. The checkout’s node types resolve normally. [root-manifest.json](evidence/root-manifest.json) records the roots/options; [baseline](evidence/strict-baseline.json), [property](evidence/strict-properties.json) and [hybrid](evidence/strict-hybrid.json) retain the complete diagnostics. Earlier missing-node prototype diagnostics are discarded.

The property road converts **53 unique source member groups across 5 files** for the 11 receiver types, including inherited declarations; overloads become callable property type literals. Transpiling every changed file with stock TypeScript, ES2020/ESNext and comments removed produced byte-identical JavaScript. The hybrid converts **16 internal/private groups**, and replaces the 19 public patterns with 82 receiver arrows. Original callable types are restored locally with type assertions to preserve generics and overloads; optional undefined values, defaults and printer identity guards are preserved. These choices are part of the measured recipe, not a claim that Adamic supports every introduced syntax form.

The exact private recipes are [properties-source.patch](properties-source.patch), [hybrid-source.patch](hybrid-source.patch) and [variants.cjs](variants.cjs). A fresh typed scan found **baseline 29, property road 0, hybrid 0** method-signature object-binding sites. [road-audit.json](evidence/road-audit.json) also records source internal-tag ownership.

The property road’s sole diagnostic is:

```text
src/compiler/factory/nodeFactory.ts:639:9 TS2322
Type '(expression: Expression, questionDotToken: QuestionDotToken | undefined, name: string) => Mutable<PropertyAccessChain>' is not assignable to type '(expression: Expression, questionDotToken: QuestionDotToken | undefined, name: string | MemberName) => PropertyAccessChain'.
  Types of parameters 'name' and 'name' are incompatible.
    Type 'string | MemberName' is not assignable to type 'string'.
      Type 'Identifier' is not assignable to type 'string'.
```

No annotation was fixed. Normal property `hereby dts` failed in build-services on this same error (task exit 1; tsc exit 2). It emitted fresh declaration intermediates, including the typescript entrypoint and compiler types. To obtain the actual API projection without weakening checking or modifying source, the unchanged official dts-services bundler command was run directly on those intermediates:

```sh
node scripts/dtsBundler.mjs --entrypoint ./built/local/typescript/typescript.d.ts --output ./built/local/typescript.d.ts
```

That projection succeeded. It is **not a successful property-road hereby dts build**. The property tsserverlibrary shim was not emitted because the full task stopped before dts-lssl. The exact **19,174-byte** [typescript.d.ts diff](property-typescript.d.ts.diff) is from the emitted intermediates and official bundler, not a predicted/manual declaration edit. It changes **34 public members** and retains previously internal call signatures inside the new property type literals: createStringLiteral **1→2**, createIdentifier **1→2**, createToken **11→12**. The @internal comments remain visible in that public projection; the bundler did not strip those nested signatures.

The hybrid’s unmodified `hereby dts` passed. Both files compare byte for byte with baseline:

| Bundle / road | Bytes | SHA-256 |
| --- | ---: | --- |
| baseline/typescript.d.ts | 593536 | `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516` |
| baseline/tsserverlibrary.d.ts | 866 | `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd` |
| properties/typescript.d.ts | 594004 | `762dfec7eb3d43ba06d8b4b8bcaca884b732d27d2bf23769c6ac5ffaacad4aef` |
| hybrid/typescript.d.ts | 593536 | `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516` |
| hybrid/tsserverlibrary.d.ts | 866 | `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd` |

## Sequential execution and peak memory

No accepted counting, census, declaration or type-check experiment overlaps another. [timeline.json](evidence/timeline.json) records the serial intervals and measurements. Endpoints derive from metadata-write times and measured wall durations. Peak sum RSS is sampled at 0.2 s across descendants even when the oracle starts new process sessions; shared pages count once per process. Largest-descendant high-water RSS is recorded separately. These are RSS observations, not cgroup physical-memory totals. Census peaks and limits are in the per-file table above.

| Job | Exit | Seconds | Peak summed RSS (KiB) | Largest descendant HWM (KiB) |
| --- | ---: | ---: | ---: | ---: |
| clean-lane.json | 1 | 498.719 | 9,049,844 | 2,737,656 |
| reconciliation-memory.json | 0 | 4.720 | 527,664 | 527,444 |
| strict-baseline-memory.json | 0 | 14.323 | 701,536 | 701,296 |
| variants-preparation.json | 0 | 4.929 | 517,532 | 542,080 |
| strict-properties-memory.json | 2 | 13.925 | 699,208 | 699,560 |
| strict-hybrid-memory.json | 0 | 16.377 | 701,700 | 701,908 |
| dts-properties.json | 1 | 23.846 | 1,097,628 | 1,052,236 |
| property-official-bundler.json | 0 | 4.297 | 483,320 | 494,924 |
| dts-hybrid.json | 0 | 25.040 | 1,057,964 | 1,018,120 |
| providers-memory-final.json | 0 | 21.671 | 1,122,368 | 1,123,192 |
| road-audit-memory.json | 0 | 8.609 | 1,086,536 | 1,086,276 |

Discarded attempts also ran serially; their peaks are retained for audit, not included in heat. See [discarded-attempts.json](evidence/discarded-attempts.json).

| Attempt | Why discarded | Peak summed RSS (KiB) |
| --- | --- | ---: |
| counting-lane | stopped: sentinel identity affected | 8,872,196 |
| certified-lane | discarded: omitted node_modules fixtures | 8,802,516 |
| final-lane | discarded: OOM while scratch occupied tmpfs | 10,063,928 |
| accepted-lane | discarded: two omitted built references and one hook timeout | 12,268,576 |

## Public declaration quotations

The public receiver declarations below are quoted from the baseline bundle; CoreTransformationContext is included because TransformationContext inherits the taken lexical/hoist methods.

### PrintHandlers

```typescript
interface PrintHandlers {
        /**
         * A hook used by the Printer when generating unique names to avoid collisions with
         * globally defined names that exist outside of the current source file.
         */
        hasGlobalName?(name: string): boolean;
        /**
         * A hook used by the Printer to provide notifications prior to emitting a node. A
         * compatible implementation **must** invoke `emitCallback` with the provided `hint` and
         * `node` values.
         * @param hint A hint indicating the intended purpose of the node.
         * @param node The node to emit.
         * @param emitCallback A callback that, when invoked, will emit the node.
         * @example
         * ```ts
         * var printer = createPrinter(printerOptions, {
         *   onEmitNode(hint, node, emitCallback) {
         *     // set up or track state prior to emitting the node...
         *     emitCallback(hint, node);
         *     // restore state after emitting the node...
         *   }
         * });
         * ```
         */
        onEmitNode?(hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void): void;
        /**
         * A hook used to check if an emit notification is required for a node.
         * @param node The node to emit.
         */
        isEmitNotificationEnabled?(node: Node): boolean;
        /**
         * A hook used by the Printer to perform just-in-time substitution of a node. This is
         * primarily used by node transformations that need to substitute one node for another,
         * such as replacing `myExportedVar` with `exports.myExportedVar`.
         * @param hint A hint indicating the intended purpose of the node.
         * @param node The node to emit.
         * @example
         * ```ts
         * var printer = createPrinter(printerOptions, {
         *   substituteNode(hint, node) {
         *     // perform substitution if necessary...
         *     return node;
         *   }
         * });
         * ```
         */
        substituteNode?(hint: EmitHint, node: Node): Node;
    }
```

### NodeFactory

```typescript
interface NodeFactory {
        createNodeArray<T extends Node>(elements?: readonly T[], hasTrailingComma?: boolean): NodeArray<T>;
        createNumericLiteral(value: string | number, numericLiteralFlags?: TokenFlags): NumericLiteral;
        createBigIntLiteral(value: string | PseudoBigInt): BigIntLiteral;
        createStringLiteral(text: string, isSingleQuote?: boolean): StringLiteral;
        createStringLiteralFromNode(sourceNode: PropertyNameLiteral | PrivateIdentifier, isSingleQuote?: boolean): StringLiteral;
        createRegularExpressionLiteral(text: string): RegularExpressionLiteral;
        createIdentifier(text: string): Identifier;
        /**
         * Create a unique temporary variable.
         * @param recordTempVariable An optional callback used to record the temporary variable name. This
         * should usually be a reference to `hoistVariableDeclaration` from a `TransformationContext`, but
         * can be `undefined` if you plan to record the temporary variable manually.
         * @param reservedInNestedScopes When `true`, reserves the temporary variable name in all nested scopes
         * during emit so that the variable can be referenced in a nested function body. This is an alternative to
         * setting `EmitFlags.ReuseTempVariableScope` on the nested function itself.
         */
        createTempVariable(recordTempVariable: ((node: Identifier) => void) | undefined, reservedInNestedScopes?: boolean): Identifier;
        /**
         * Create a unique temporary variable for use in a loop.
         * @param reservedInNestedScopes When `true`, reserves the temporary variable name in all nested scopes
         * during emit so that the variable can be referenced in a nested function body. This is an alternative to
         * setting `EmitFlags.ReuseTempVariableScope` on the nested function itself.
         */
        createLoopVariable(reservedInNestedScopes?: boolean): Identifier;
        /** Create a unique name based on the supplied text. */
        createUniqueName(text: string, flags?: GeneratedIdentifierFlags): Identifier;
        /** Create a unique name generated for a node. */
        getGeneratedNameForNode(node: Node | undefined, flags?: GeneratedIdentifierFlags): Identifier;
        createPrivateIdentifier(text: string): PrivateIdentifier;
        createUniquePrivateName(text?: string): PrivateIdentifier;
        getGeneratedPrivateNameForNode(node: Node): PrivateIdentifier;
        createToken(token: SyntaxKind.SuperKeyword): SuperExpression;
        createToken(token: SyntaxKind.ThisKeyword): ThisExpression;
        createToken(token: SyntaxKind.NullKeyword): NullLiteral;
        createToken(token: SyntaxKind.TrueKeyword): TrueLiteral;
        createToken(token: SyntaxKind.FalseKeyword): FalseLiteral;
        createToken(token: SyntaxKind.EndOfFileToken): EndOfFileToken;
        createToken(token: SyntaxKind.Unknown): Token<SyntaxKind.Unknown>;
        createToken<TKind extends PunctuationSyntaxKind>(token: TKind): PunctuationToken<TKind>;
        createToken<TKind extends KeywordTypeSyntaxKind>(token: TKind): KeywordTypeNode<TKind>;
        createToken<TKind extends ModifierSyntaxKind>(token: TKind): ModifierToken<TKind>;
        createToken<TKind extends KeywordSyntaxKind>(token: TKind): KeywordToken<TKind>;
        createSuper(): SuperExpression;
        createThis(): ThisExpression;
        createNull(): NullLiteral;
        createTrue(): TrueLiteral;
        createFalse(): FalseLiteral;
        createModifier<T extends ModifierSyntaxKind>(kind: T): ModifierToken<T>;
        createModifiersFromModifierFlags(flags: ModifierFlags): Modifier[] | undefined;
        createQualifiedName(left: EntityName, right: string | Identifier): QualifiedName;
        updateQualifiedName(node: QualifiedName, left: EntityName, right: Identifier): QualifiedName;
        createComputedPropertyName(expression: Expression): ComputedPropertyName;
        updateComputedPropertyName(node: ComputedPropertyName, expression: Expression): ComputedPropertyName;
        createTypeParameterDeclaration(modifiers: readonly Modifier[] | undefined, name: string | Identifier, constraint?: TypeNode, defaultType?: TypeNode): TypeParameterDeclaration;
        updateTypeParameterDeclaration(node: TypeParameterDeclaration, modifiers: readonly Modifier[] | undefined, name: Identifier, constraint: TypeNode | undefined, defaultType: TypeNode | undefined): TypeParameterDeclaration;
        createParameterDeclaration(modifiers: readonly ModifierLike[] | undefined, dotDotDotToken: DotDotDotToken | undefined, name: string | BindingName, questionToken?: QuestionToken, type?: TypeNode, initializer?: Expression): ParameterDeclaration;
        updateParameterDeclaration(node: ParameterDeclaration, modifiers: readonly ModifierLike[] | undefined, dotDotDotToken: DotDotDotToken | undefined, name: string | BindingName, questionToken: QuestionToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): ParameterDeclaration;
        createDecorator(expression: Expression): Decorator;
        updateDecorator(node: Decorator, expression: Expression): Decorator;
        createPropertySignature(modifiers: readonly Modifier[] | undefined, name: PropertyName | string, questionToken: QuestionToken | undefined, type: TypeNode | undefined): PropertySignature;
        updatePropertySignature(node: PropertySignature, modifiers: readonly Modifier[] | undefined, name: PropertyName, questionToken: QuestionToken | undefined, type: TypeNode | undefined): PropertySignature;
        createPropertyDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, questionOrExclamationToken: QuestionToken | ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): PropertyDeclaration;
        updatePropertyDeclaration(node: PropertyDeclaration, modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, questionOrExclamationToken: QuestionToken | ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): PropertyDeclaration;
        createMethodSignature(modifiers: readonly Modifier[] | undefined, name: string | PropertyName, questionToken: QuestionToken | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): MethodSignature;
        updateMethodSignature(node: MethodSignature, modifiers: readonly Modifier[] | undefined, name: PropertyName, questionToken: QuestionToken | undefined, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode | undefined): MethodSignature;
        createMethodDeclaration(modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | PropertyName, questionToken: QuestionToken | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): MethodDeclaration;
        updateMethodDeclaration(node: MethodDeclaration, modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: PropertyName, questionToken: QuestionToken | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): MethodDeclaration;
        createConstructorDeclaration(modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], body: Block | undefined): ConstructorDeclaration;
        updateConstructorDeclaration(node: ConstructorDeclaration, modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], body: Block | undefined): ConstructorDeclaration;
        createGetAccessorDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): GetAccessorDeclaration;
        updateGetAccessorDeclaration(node: GetAccessorDeclaration, modifiers: readonly ModifierLike[] | undefined, name: PropertyName, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): GetAccessorDeclaration;
        createSetAccessorDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, parameters: readonly ParameterDeclaration[], body: Block | undefined): SetAccessorDeclaration;
        updateSetAccessorDeclaration(node: SetAccessorDeclaration, modifiers: readonly ModifierLike[] | undefined, name: PropertyName, parameters: readonly ParameterDeclaration[], body: Block | undefined): SetAccessorDeclaration;
        createCallSignature(typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): CallSignatureDeclaration;
        updateCallSignature(node: CallSignatureDeclaration, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode | undefined): CallSignatureDeclaration;
        createConstructSignature(typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): ConstructSignatureDeclaration;
        updateConstructSignature(node: ConstructSignatureDeclaration, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode | undefined): ConstructSignatureDeclaration;
        createIndexSignature(modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): IndexSignatureDeclaration;
        updateIndexSignature(node: IndexSignatureDeclaration, modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): IndexSignatureDeclaration;
        createTemplateLiteralTypeSpan(type: TypeNode, literal: TemplateMiddle | TemplateTail): TemplateLiteralTypeSpan;
        updateTemplateLiteralTypeSpan(node: TemplateLiteralTypeSpan, type: TypeNode, literal: TemplateMiddle | TemplateTail): TemplateLiteralTypeSpan;
        createClassStaticBlockDeclaration(body: Block): ClassStaticBlockDeclaration;
        updateClassStaticBlockDeclaration(node: ClassStaticBlockDeclaration, body: Block): ClassStaticBlockDeclaration;
        createKeywordTypeNode<TKind extends KeywordTypeSyntaxKind>(kind: TKind): KeywordTypeNode<TKind>;
        createTypePredicateNode(assertsModifier: AssertsKeyword | undefined, parameterName: Identifier | ThisTypeNode | string, type: TypeNode | undefined): TypePredicateNode;
        updateTypePredicateNode(node: TypePredicateNode, assertsModifier: AssertsKeyword | undefined, parameterName: Identifier | ThisTypeNode, type: TypeNode | undefined): TypePredicateNode;
        createTypeReferenceNode(typeName: string | EntityName, typeArguments?: readonly TypeNode[]): TypeReferenceNode;
        updateTypeReferenceNode(node: TypeReferenceNode, typeName: EntityName, typeArguments: NodeArray<TypeNode> | undefined): TypeReferenceNode;
        createFunctionTypeNode(typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): FunctionTypeNode;
        updateFunctionTypeNode(node: FunctionTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode): FunctionTypeNode;
        createConstructorTypeNode(modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): ConstructorTypeNode;
        updateConstructorTypeNode(node: ConstructorTypeNode, modifiers: readonly Modifier[] | undefined, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode): ConstructorTypeNode;
        createTypeQueryNode(exprName: EntityName, typeArguments?: readonly TypeNode[]): TypeQueryNode;
        updateTypeQueryNode(node: TypeQueryNode, exprName: EntityName, typeArguments?: readonly TypeNode[]): TypeQueryNode;
        createTypeLiteralNode(members: readonly TypeElement[] | undefined): TypeLiteralNode;
        updateTypeLiteralNode(node: TypeLiteralNode, members: NodeArray<TypeElement>): TypeLiteralNode;
        createArrayTypeNode(elementType: TypeNode): ArrayTypeNode;
        updateArrayTypeNode(node: ArrayTypeNode, elementType: TypeNode): ArrayTypeNode;
        createTupleTypeNode(elements: readonly (TypeNode | NamedTupleMember)[]): TupleTypeNode;
        updateTupleTypeNode(node: TupleTypeNode, elements: readonly (TypeNode | NamedTupleMember)[]): TupleTypeNode;
        createNamedTupleMember(dotDotDotToken: DotDotDotToken | undefined, name: Identifier, questionToken: QuestionToken | undefined, type: TypeNode): NamedTupleMember;
        updateNamedTupleMember(node: NamedTupleMember, dotDotDotToken: DotDotDotToken | undefined, name: Identifier, questionToken: QuestionToken | undefined, type: TypeNode): NamedTupleMember;
        createOptionalTypeNode(type: TypeNode): OptionalTypeNode;
        updateOptionalTypeNode(node: OptionalTypeNode, type: TypeNode): OptionalTypeNode;
        createRestTypeNode(type: TypeNode): RestTypeNode;
        updateRestTypeNode(node: RestTypeNode, type: TypeNode): RestTypeNode;
        createUnionTypeNode(types: readonly TypeNode[]): UnionTypeNode;
        updateUnionTypeNode(node: UnionTypeNode, types: NodeArray<TypeNode>): UnionTypeNode;
        createIntersectionTypeNode(types: readonly TypeNode[]): IntersectionTypeNode;
        updateIntersectionTypeNode(node: IntersectionTypeNode, types: NodeArray<TypeNode>): IntersectionTypeNode;
        createConditionalTypeNode(checkType: TypeNode, extendsType: TypeNode, trueType: TypeNode, falseType: TypeNode): ConditionalTypeNode;
        updateConditionalTypeNode(node: ConditionalTypeNode, checkType: TypeNode, extendsType: TypeNode, trueType: TypeNode, falseType: TypeNode): ConditionalTypeNode;
        createInferTypeNode(typeParameter: TypeParameterDeclaration): InferTypeNode;
        updateInferTypeNode(node: InferTypeNode, typeParameter: TypeParameterDeclaration): InferTypeNode;
        createImportTypeNode(argument: TypeNode, attributes?: ImportAttributes, qualifier?: EntityName, typeArguments?: readonly TypeNode[], isTypeOf?: boolean): ImportTypeNode;
        updateImportTypeNode(node: ImportTypeNode, argument: TypeNode, attributes: ImportAttributes | undefined, qualifier: EntityName | undefined, typeArguments: readonly TypeNode[] | undefined, isTypeOf?: boolean): ImportTypeNode;
        createParenthesizedType(type: TypeNode): ParenthesizedTypeNode;
        updateParenthesizedType(node: ParenthesizedTypeNode, type: TypeNode): ParenthesizedTypeNode;
        createThisTypeNode(): ThisTypeNode;
        createTypeOperatorNode(operator: SyntaxKind.KeyOfKeyword | SyntaxKind.UniqueKeyword | SyntaxKind.ReadonlyKeyword, type: TypeNode): TypeOperatorNode;
        updateTypeOperatorNode(node: TypeOperatorNode, type: TypeNode): TypeOperatorNode;
        createIndexedAccessTypeNode(objectType: TypeNode, indexType: TypeNode): IndexedAccessTypeNode;
        updateIndexedAccessTypeNode(node: IndexedAccessTypeNode, objectType: TypeNode, indexType: TypeNode): IndexedAccessTypeNode;
        createMappedTypeNode(readonlyToken: ReadonlyKeyword | PlusToken | MinusToken | undefined, typeParameter: TypeParameterDeclaration, nameType: TypeNode | undefined, questionToken: QuestionToken | PlusToken | MinusToken | undefined, type: TypeNode | undefined, members: NodeArray<TypeElement> | undefined): MappedTypeNode;
        updateMappedTypeNode(node: MappedTypeNode, readonlyToken: ReadonlyKeyword | PlusToken | MinusToken | undefined, typeParameter: TypeParameterDeclaration, nameType: TypeNode | undefined, questionToken: QuestionToken | PlusToken | MinusToken | undefined, type: TypeNode | undefined, members: NodeArray<TypeElement> | undefined): MappedTypeNode;
        createLiteralTypeNode(literal: LiteralTypeNode["literal"]): LiteralTypeNode;
        updateLiteralTypeNode(node: LiteralTypeNode, literal: LiteralTypeNode["literal"]): LiteralTypeNode;
        createTemplateLiteralType(head: TemplateHead, templateSpans: readonly TemplateLiteralTypeSpan[]): TemplateLiteralTypeNode;
        updateTemplateLiteralType(node: TemplateLiteralTypeNode, head: TemplateHead, templateSpans: readonly TemplateLiteralTypeSpan[]): TemplateLiteralTypeNode;
        createObjectBindingPattern(elements: readonly BindingElement[]): ObjectBindingPattern;
        updateObjectBindingPattern(node: ObjectBindingPattern, elements: readonly BindingElement[]): ObjectBindingPattern;
        createArrayBindingPattern(elements: readonly ArrayBindingElement[]): ArrayBindingPattern;
        updateArrayBindingPattern(node: ArrayBindingPattern, elements: readonly ArrayBindingElement[]): ArrayBindingPattern;
        createBindingElement(dotDotDotToken: DotDotDotToken | undefined, propertyName: string | PropertyName | undefined, name: string | BindingName, initializer?: Expression): BindingElement;
        updateBindingElement(node: BindingElement, dotDotDotToken: DotDotDotToken | undefined, propertyName: PropertyName | undefined, name: BindingName, initializer: Expression | undefined): BindingElement;
        createArrayLiteralExpression(elements?: readonly Expression[], multiLine?: boolean): ArrayLiteralExpression;
        updateArrayLiteralExpression(node: ArrayLiteralExpression, elements: readonly Expression[]): ArrayLiteralExpression;
        createObjectLiteralExpression(properties?: readonly ObjectLiteralElementLike[], multiLine?: boolean): ObjectLiteralExpression;
        updateObjectLiteralExpression(node: ObjectLiteralExpression, properties: readonly ObjectLiteralElementLike[]): ObjectLiteralExpression;
        createPropertyAccessExpression(expression: Expression, name: string | MemberName): PropertyAccessExpression;
        updatePropertyAccessExpression(node: PropertyAccessExpression, expression: Expression, name: MemberName): PropertyAccessExpression;
        createPropertyAccessChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, name: string | MemberName): PropertyAccessChain;
        updatePropertyAccessChain(node: PropertyAccessChain, expression: Expression, questionDotToken: QuestionDotToken | undefined, name: MemberName): PropertyAccessChain;
        createElementAccessExpression(expression: Expression, index: number | Expression): ElementAccessExpression;
        updateElementAccessExpression(node: ElementAccessExpression, expression: Expression, argumentExpression: Expression): ElementAccessExpression;
        createElementAccessChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, index: number | Expression): ElementAccessChain;
        updateElementAccessChain(node: ElementAccessChain, expression: Expression, questionDotToken: QuestionDotToken | undefined, argumentExpression: Expression): ElementAccessChain;
        createCallExpression(expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): CallExpression;
        updateCallExpression(node: CallExpression, expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[]): CallExpression;
        createCallChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): CallChain;
        updateCallChain(node: CallChain, expression: Expression, questionDotToken: QuestionDotToken | undefined, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[]): CallChain;
        createNewExpression(expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): NewExpression;
        updateNewExpression(node: NewExpression, expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): NewExpression;
        createTaggedTemplateExpression(tag: Expression, typeArguments: readonly TypeNode[] | undefined, template: TemplateLiteral): TaggedTemplateExpression;
        updateTaggedTemplateExpression(node: TaggedTemplateExpression, tag: Expression, typeArguments: readonly TypeNode[] | undefined, template: TemplateLiteral): TaggedTemplateExpression;
        createTypeAssertion(type: TypeNode, expression: Expression): TypeAssertion;
        updateTypeAssertion(node: TypeAssertion, type: TypeNode, expression: Expression): TypeAssertion;
        createParenthesizedExpression(expression: Expression): ParenthesizedExpression;
        updateParenthesizedExpression(node: ParenthesizedExpression, expression: Expression): ParenthesizedExpression;
        createFunctionExpression(modifiers: readonly Modifier[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[] | undefined, type: TypeNode | undefined, body: Block): FunctionExpression;
        updateFunctionExpression(node: FunctionExpression, modifiers: readonly Modifier[] | undefined, asteriskToken: AsteriskToken | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block): FunctionExpression;
        createArrowFunction(modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, equalsGreaterThanToken: EqualsGreaterThanToken | undefined, body: ConciseBody): ArrowFunction;
        updateArrowFunction(node: ArrowFunction, modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, equalsGreaterThanToken: EqualsGreaterThanToken, body: ConciseBody): ArrowFunction;
        createDeleteExpression(expression: Expression): DeleteExpression;
        updateDeleteExpression(node: DeleteExpression, expression: Expression): DeleteExpression;
        createTypeOfExpression(expression: Expression): TypeOfExpression;
        updateTypeOfExpression(node: TypeOfExpression, expression: Expression): TypeOfExpression;
        createVoidExpression(expression: Expression): VoidExpression;
        updateVoidExpression(node: VoidExpression, expression: Expression): VoidExpression;
        createAwaitExpression(expression: Expression): AwaitExpression;
        updateAwaitExpression(node: AwaitExpression, expression: Expression): AwaitExpression;
        createPrefixUnaryExpression(operator: PrefixUnaryOperator, operand: Expression): PrefixUnaryExpression;
        updatePrefixUnaryExpression(node: PrefixUnaryExpression, operand: Expression): PrefixUnaryExpression;
        createPostfixUnaryExpression(operand: Expression, operator: PostfixUnaryOperator): PostfixUnaryExpression;
        updatePostfixUnaryExpression(node: PostfixUnaryExpression, operand: Expression): PostfixUnaryExpression;
        createBinaryExpression(left: Expression, operator: BinaryOperator | BinaryOperatorToken, right: Expression): BinaryExpression;
        updateBinaryExpression(node: BinaryExpression, left: Expression, operator: BinaryOperator | BinaryOperatorToken, right: Expression): BinaryExpression;
        createConditionalExpression(condition: Expression, questionToken: QuestionToken | undefined, whenTrue: Expression, colonToken: ColonToken | undefined, whenFalse: Expression): ConditionalExpression;
        updateConditionalExpression(node: ConditionalExpression, condition: Expression, questionToken: QuestionToken, whenTrue: Expression, colonToken: ColonToken, whenFalse: Expression): ConditionalExpression;
        createTemplateExpression(head: TemplateHead, templateSpans: readonly TemplateSpan[]): TemplateExpression;
        updateTemplateExpression(node: TemplateExpression, head: TemplateHead, templateSpans: readonly TemplateSpan[]): TemplateExpression;
        createTemplateHead(text: string, rawText?: string, templateFlags?: TokenFlags): TemplateHead;
        createTemplateHead(text: string | undefined, rawText: string, templateFlags?: TokenFlags): TemplateHead;
        createTemplateMiddle(text: string, rawText?: string, templateFlags?: TokenFlags): TemplateMiddle;
        createTemplateMiddle(text: string | undefined, rawText: string, templateFlags?: TokenFlags): TemplateMiddle;
        createTemplateTail(text: string, rawText?: string, templateFlags?: TokenFlags): TemplateTail;
        createTemplateTail(text: string | undefined, rawText: string, templateFlags?: TokenFlags): TemplateTail;
        createNoSubstitutionTemplateLiteral(text: string, rawText?: string): NoSubstitutionTemplateLiteral;
        createNoSubstitutionTemplateLiteral(text: string | undefined, rawText: string): NoSubstitutionTemplateLiteral;
        createYieldExpression(asteriskToken: AsteriskToken, expression: Expression): YieldExpression;
        createYieldExpression(asteriskToken: undefined, expression: Expression | undefined): YieldExpression;
        updateYieldExpression(node: YieldExpression, asteriskToken: AsteriskToken | undefined, expression: Expression | undefined): YieldExpression;
        createSpreadElement(expression: Expression): SpreadElement;
        updateSpreadElement(node: SpreadElement, expression: Expression): SpreadElement;
        createClassExpression(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassExpression;
        updateClassExpression(node: ClassExpression, modifiers: readonly ModifierLike[] | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassExpression;
        createOmittedExpression(): OmittedExpression;
        createExpressionWithTypeArguments(expression: Expression, typeArguments: readonly TypeNode[] | undefined): ExpressionWithTypeArguments;
        updateExpressionWithTypeArguments(node: ExpressionWithTypeArguments, expression: Expression, typeArguments: readonly TypeNode[] | undefined): ExpressionWithTypeArguments;
        createAsExpression(expression: Expression, type: TypeNode): AsExpression;
        updateAsExpression(node: AsExpression, expression: Expression, type: TypeNode): AsExpression;
        createNonNullExpression(expression: Expression): NonNullExpression;
        updateNonNullExpression(node: NonNullExpression, expression: Expression): NonNullExpression;
        createNonNullChain(expression: Expression): NonNullChain;
        updateNonNullChain(node: NonNullChain, expression: Expression): NonNullChain;
        createMetaProperty(keywordToken: MetaProperty["keywordToken"], name: Identifier): MetaProperty;
        updateMetaProperty(node: MetaProperty, name: Identifier): MetaProperty;
        createSatisfiesExpression(expression: Expression, type: TypeNode): SatisfiesExpression;
        updateSatisfiesExpression(node: SatisfiesExpression, expression: Expression, type: TypeNode): SatisfiesExpression;
        createTemplateSpan(expression: Expression, literal: TemplateMiddle | TemplateTail): TemplateSpan;
        updateTemplateSpan(node: TemplateSpan, expression: Expression, literal: TemplateMiddle | TemplateTail): TemplateSpan;
        createSemicolonClassElement(): SemicolonClassElement;
        createBlock(statements: readonly Statement[], multiLine?: boolean): Block;
        updateBlock(node: Block, statements: readonly Statement[]): Block;
        createVariableStatement(modifiers: readonly ModifierLike[] | undefined, declarationList: VariableDeclarationList | readonly VariableDeclaration[]): VariableStatement;
        updateVariableStatement(node: VariableStatement, modifiers: readonly ModifierLike[] | undefined, declarationList: VariableDeclarationList): VariableStatement;
        createEmptyStatement(): EmptyStatement;
        createExpressionStatement(expression: Expression): ExpressionStatement;
        updateExpressionStatement(node: ExpressionStatement, expression: Expression): ExpressionStatement;
        createIfStatement(expression: Expression, thenStatement: Statement, elseStatement?: Statement): IfStatement;
        updateIfStatement(node: IfStatement, expression: Expression, thenStatement: Statement, elseStatement: Statement | undefined): IfStatement;
        createDoStatement(statement: Statement, expression: Expression): DoStatement;
        updateDoStatement(node: DoStatement, statement: Statement, expression: Expression): DoStatement;
        createWhileStatement(expression: Expression, statement: Statement): WhileStatement;
        updateWhileStatement(node: WhileStatement, expression: Expression, statement: Statement): WhileStatement;
        createForStatement(initializer: ForInitializer | undefined, condition: Expression | undefined, incrementor: Expression | undefined, statement: Statement): ForStatement;
        updateForStatement(node: ForStatement, initializer: ForInitializer | undefined, condition: Expression | undefined, incrementor: Expression | undefined, statement: Statement): ForStatement;
        createForInStatement(initializer: ForInitializer, expression: Expression, statement: Statement): ForInStatement;
        updateForInStatement(node: ForInStatement, initializer: ForInitializer, expression: Expression, statement: Statement): ForInStatement;
        createForOfStatement(awaitModifier: AwaitKeyword | undefined, initializer: ForInitializer, expression: Expression, statement: Statement): ForOfStatement;
        updateForOfStatement(node: ForOfStatement, awaitModifier: AwaitKeyword | undefined, initializer: ForInitializer, expression: Expression, statement: Statement): ForOfStatement;
        createContinueStatement(label?: string | Identifier): ContinueStatement;
        updateContinueStatement(node: ContinueStatement, label: Identifier | undefined): ContinueStatement;
        createBreakStatement(label?: string | Identifier): BreakStatement;
        updateBreakStatement(node: BreakStatement, label: Identifier | undefined): BreakStatement;
        createReturnStatement(expression?: Expression): ReturnStatement;
        updateReturnStatement(node: ReturnStatement, expression: Expression | undefined): ReturnStatement;
        createWithStatement(expression: Expression, statement: Statement): WithStatement;
        updateWithStatement(node: WithStatement, expression: Expression, statement: Statement): WithStatement;
        createSwitchStatement(expression: Expression, caseBlock: CaseBlock): SwitchStatement;
        updateSwitchStatement(node: SwitchStatement, expression: Expression, caseBlock: CaseBlock): SwitchStatement;
        createLabeledStatement(label: string | Identifier, statement: Statement): LabeledStatement;
        updateLabeledStatement(node: LabeledStatement, label: Identifier, statement: Statement): LabeledStatement;
        createThrowStatement(expression: Expression): ThrowStatement;
        updateThrowStatement(node: ThrowStatement, expression: Expression): ThrowStatement;
        createTryStatement(tryBlock: Block, catchClause: CatchClause | undefined, finallyBlock: Block | undefined): TryStatement;
        updateTryStatement(node: TryStatement, tryBlock: Block, catchClause: CatchClause | undefined, finallyBlock: Block | undefined): TryStatement;
        createDebuggerStatement(): DebuggerStatement;
        createVariableDeclaration(name: string | BindingName, exclamationToken?: ExclamationToken, type?: TypeNode, initializer?: Expression): VariableDeclaration;
        updateVariableDeclaration(node: VariableDeclaration, name: BindingName, exclamationToken: ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): VariableDeclaration;
        createVariableDeclarationList(declarations: readonly VariableDeclaration[], flags?: NodeFlags): VariableDeclarationList;
        updateVariableDeclarationList(node: VariableDeclarationList, declarations: readonly VariableDeclaration[]): VariableDeclarationList;
        createFunctionDeclaration(modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): FunctionDeclaration;
        updateFunctionDeclaration(node: FunctionDeclaration, modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): FunctionDeclaration;
        createClassDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassDeclaration;
        updateClassDeclaration(node: ClassDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassDeclaration;
        createInterfaceDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly TypeElement[]): InterfaceDeclaration;
        updateInterfaceDeclaration(node: InterfaceDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly TypeElement[]): InterfaceDeclaration;
        createTypeAliasDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, type: TypeNode): TypeAliasDeclaration;
        updateTypeAliasDeclaration(node: TypeAliasDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, type: TypeNode): TypeAliasDeclaration;
        createEnumDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, members: readonly EnumMember[]): EnumDeclaration;
        updateEnumDeclaration(node: EnumDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier, members: readonly EnumMember[]): EnumDeclaration;
        createModuleDeclaration(modifiers: readonly ModifierLike[] | undefined, name: ModuleName, body: ModuleBody | undefined, flags?: NodeFlags): ModuleDeclaration;
        updateModuleDeclaration(node: ModuleDeclaration, modifiers: readonly ModifierLike[] | undefined, name: ModuleName, body: ModuleBody | undefined): ModuleDeclaration;
        createModuleBlock(statements: readonly Statement[]): ModuleBlock;
        updateModuleBlock(node: ModuleBlock, statements: readonly Statement[]): ModuleBlock;
        createCaseBlock(clauses: readonly CaseOrDefaultClause[]): CaseBlock;
        updateCaseBlock(node: CaseBlock, clauses: readonly CaseOrDefaultClause[]): CaseBlock;
        createNamespaceExportDeclaration(name: string | Identifier): NamespaceExportDeclaration;
        updateNamespaceExportDeclaration(node: NamespaceExportDeclaration, name: Identifier): NamespaceExportDeclaration;
        createImportEqualsDeclaration(modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, name: string | Identifier, moduleReference: ModuleReference): ImportEqualsDeclaration;
        updateImportEqualsDeclaration(node: ImportEqualsDeclaration, modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, name: Identifier, moduleReference: ModuleReference): ImportEqualsDeclaration;
        createImportDeclaration(modifiers: readonly ModifierLike[] | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes?: ImportAttributes): ImportDeclaration;
        updateImportDeclaration(node: ImportDeclaration, modifiers: readonly ModifierLike[] | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes: ImportAttributes | undefined): ImportDeclaration;
        createImportClause(phaseModifier: ImportPhaseModifierSyntaxKind | undefined, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        /** @deprecated */ createImportClause(isTypeOnly: boolean, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        updateImportClause(node: ImportClause, phaseModifier: ImportPhaseModifierSyntaxKind | undefined, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        /** @deprecated */ updateImportClause(node: ImportClause, isTypeOnly: boolean, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        /** @deprecated */ createAssertClause(elements: NodeArray<AssertEntry>, multiLine?: boolean): AssertClause;
        /** @deprecated */ updateAssertClause(node: AssertClause, elements: NodeArray<AssertEntry>, multiLine?: boolean): AssertClause;
        /** @deprecated */ createAssertEntry(name: AssertionKey, value: Expression): AssertEntry;
        /** @deprecated */ updateAssertEntry(node: AssertEntry, name: AssertionKey, value: Expression): AssertEntry;
        /** @deprecated */ createImportTypeAssertionContainer(clause: AssertClause, multiLine?: boolean): ImportTypeAssertionContainer;
        /** @deprecated */ updateImportTypeAssertionContainer(node: ImportTypeAssertionContainer, clause: AssertClause, multiLine?: boolean): ImportTypeAssertionContainer;
        createImportAttributes(elements: NodeArray<ImportAttribute>, multiLine?: boolean): ImportAttributes;
        updateImportAttributes(node: ImportAttributes, elements: NodeArray<ImportAttribute>, multiLine?: boolean): ImportAttributes;
        createImportAttribute(name: ImportAttributeName, value: Expression): ImportAttribute;
        updateImportAttribute(node: ImportAttribute, name: ImportAttributeName, value: Expression): ImportAttribute;
        createNamespaceImport(name: Identifier): NamespaceImport;
        updateNamespaceImport(node: NamespaceImport, name: Identifier): NamespaceImport;
        createNamespaceExport(name: ModuleExportName): NamespaceExport;
        updateNamespaceExport(node: NamespaceExport, name: ModuleExportName): NamespaceExport;
        createNamedImports(elements: readonly ImportSpecifier[]): NamedImports;
        updateNamedImports(node: NamedImports, elements: readonly ImportSpecifier[]): NamedImports;
        createImportSpecifier(isTypeOnly: boolean, propertyName: ModuleExportName | undefined, name: Identifier): ImportSpecifier;
        updateImportSpecifier(node: ImportSpecifier, isTypeOnly: boolean, propertyName: ModuleExportName | undefined, name: Identifier): ImportSpecifier;
        createExportAssignment(modifiers: readonly ModifierLike[] | undefined, isExportEquals: boolean | undefined, expression: Expression): ExportAssignment;
        updateExportAssignment(node: ExportAssignment, modifiers: readonly ModifierLike[] | undefined, expression: Expression): ExportAssignment;
        createExportDeclaration(modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, exportClause: NamedExportBindings | undefined, moduleSpecifier?: Expression, attributes?: ImportAttributes): ExportDeclaration;
        updateExportDeclaration(node: ExportDeclaration, modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, exportClause: NamedExportBindings | undefined, moduleSpecifier: Expression | undefined, attributes: ImportAttributes | undefined): ExportDeclaration;
        createNamedExports(elements: readonly ExportSpecifier[]): NamedExports;
        updateNamedExports(node: NamedExports, elements: readonly ExportSpecifier[]): NamedExports;
        createExportSpecifier(isTypeOnly: boolean, propertyName: string | ModuleExportName | undefined, name: string | ModuleExportName): ExportSpecifier;
        updateExportSpecifier(node: ExportSpecifier, isTypeOnly: boolean, propertyName: ModuleExportName | undefined, name: ModuleExportName): ExportSpecifier;
        createExternalModuleReference(expression: Expression): ExternalModuleReference;
        updateExternalModuleReference(node: ExternalModuleReference, expression: Expression): ExternalModuleReference;
        createJSDocAllType(): JSDocAllType;
        createJSDocUnknownType(): JSDocUnknownType;
        createJSDocNonNullableType(type: TypeNode, postfix?: boolean): JSDocNonNullableType;
        updateJSDocNonNullableType(node: JSDocNonNullableType, type: TypeNode): JSDocNonNullableType;
        createJSDocNullableType(type: TypeNode, postfix?: boolean): JSDocNullableType;
        updateJSDocNullableType(node: JSDocNullableType, type: TypeNode): JSDocNullableType;
        createJSDocOptionalType(type: TypeNode): JSDocOptionalType;
        updateJSDocOptionalType(node: JSDocOptionalType, type: TypeNode): JSDocOptionalType;
        createJSDocFunctionType(parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): JSDocFunctionType;
        updateJSDocFunctionType(node: JSDocFunctionType, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): JSDocFunctionType;
        createJSDocVariadicType(type: TypeNode): JSDocVariadicType;
        updateJSDocVariadicType(node: JSDocVariadicType, type: TypeNode): JSDocVariadicType;
        createJSDocNamepathType(type: TypeNode): JSDocNamepathType;
        updateJSDocNamepathType(node: JSDocNamepathType, type: TypeNode): JSDocNamepathType;
        createJSDocTypeExpression(type: TypeNode): JSDocTypeExpression;
        updateJSDocTypeExpression(node: JSDocTypeExpression, type: TypeNode): JSDocTypeExpression;
        createJSDocNameReference(name: EntityName | JSDocMemberName): JSDocNameReference;
        updateJSDocNameReference(node: JSDocNameReference, name: EntityName | JSDocMemberName): JSDocNameReference;
        createJSDocMemberName(left: EntityName | JSDocMemberName, right: Identifier): JSDocMemberName;
        updateJSDocMemberName(node: JSDocMemberName, left: EntityName | JSDocMemberName, right: Identifier): JSDocMemberName;
        createJSDocLink(name: EntityName | JSDocMemberName | undefined, text: string): JSDocLink;
        updateJSDocLink(node: JSDocLink, name: EntityName | JSDocMemberName | undefined, text: string): JSDocLink;
        createJSDocLinkCode(name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkCode;
        updateJSDocLinkCode(node: JSDocLinkCode, name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkCode;
        createJSDocLinkPlain(name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkPlain;
        updateJSDocLinkPlain(node: JSDocLinkPlain, name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkPlain;
        createJSDocTypeLiteral(jsDocPropertyTags?: readonly JSDocPropertyLikeTag[], isArrayType?: boolean): JSDocTypeLiteral;
        updateJSDocTypeLiteral(node: JSDocTypeLiteral, jsDocPropertyTags: readonly JSDocPropertyLikeTag[] | undefined, isArrayType: boolean | undefined): JSDocTypeLiteral;
        createJSDocSignature(typeParameters: readonly JSDocTemplateTag[] | undefined, parameters: readonly JSDocParameterTag[], type?: JSDocReturnTag): JSDocSignature;
        updateJSDocSignature(node: JSDocSignature, typeParameters: readonly JSDocTemplateTag[] | undefined, parameters: readonly JSDocParameterTag[], type: JSDocReturnTag | undefined): JSDocSignature;
        createJSDocTemplateTag(tagName: Identifier | undefined, constraint: JSDocTypeExpression | undefined, typeParameters: readonly TypeParameterDeclaration[], comment?: string | NodeArray<JSDocComment>): JSDocTemplateTag;
        updateJSDocTemplateTag(node: JSDocTemplateTag, tagName: Identifier | undefined, constraint: JSDocTypeExpression | undefined, typeParameters: readonly TypeParameterDeclaration[], comment: string | NodeArray<JSDocComment> | undefined): JSDocTemplateTag;
        createJSDocTypedefTag(tagName: Identifier | undefined, typeExpression?: JSDocTypeExpression | JSDocTypeLiteral, fullName?: Identifier | JSDocNamespaceDeclaration, comment?: string | NodeArray<JSDocComment>): JSDocTypedefTag;
        updateJSDocTypedefTag(node: JSDocTypedefTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | JSDocTypeLiteral | undefined, fullName: Identifier | JSDocNamespaceDeclaration | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocTypedefTag;
        createJSDocParameterTag(tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression?: JSDocTypeExpression, isNameFirst?: boolean, comment?: string | NodeArray<JSDocComment>): JSDocParameterTag;
        updateJSDocParameterTag(node: JSDocParameterTag, tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression: JSDocTypeExpression | undefined, isNameFirst: boolean, comment: string | NodeArray<JSDocComment> | undefined): JSDocParameterTag;
        createJSDocPropertyTag(tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression?: JSDocTypeExpression, isNameFirst?: boolean, comment?: string | NodeArray<JSDocComment>): JSDocPropertyTag;
        updateJSDocPropertyTag(node: JSDocPropertyTag, tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression: JSDocTypeExpression | undefined, isNameFirst: boolean, comment: string | NodeArray<JSDocComment> | undefined): JSDocPropertyTag;
        createJSDocTypeTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocTypeTag;
        updateJSDocTypeTag(node: JSDocTypeTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment: string | NodeArray<JSDocComment> | undefined): JSDocTypeTag;
        createJSDocSeeTag(tagName: Identifier | undefined, nameExpression: JSDocNameReference | undefined, comment?: string | NodeArray<JSDocComment>): JSDocSeeTag;
        updateJSDocSeeTag(node: JSDocSeeTag, tagName: Identifier | undefined, nameExpression: JSDocNameReference | undefined, comment?: string | NodeArray<JSDocComment>): JSDocSeeTag;
        createJSDocReturnTag(tagName: Identifier | undefined, typeExpression?: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocReturnTag;
        updateJSDocReturnTag(node: JSDocReturnTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocReturnTag;
        createJSDocThisTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocThisTag;
        updateJSDocThisTag(node: JSDocThisTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocThisTag;
        createJSDocEnumTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocEnumTag;
        updateJSDocEnumTag(node: JSDocEnumTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment: string | NodeArray<JSDocComment> | undefined): JSDocEnumTag;
        createJSDocCallbackTag(tagName: Identifier | undefined, typeExpression: JSDocSignature, fullName?: Identifier | JSDocNamespaceDeclaration, comment?: string | NodeArray<JSDocComment>): JSDocCallbackTag;
        updateJSDocCallbackTag(node: JSDocCallbackTag, tagName: Identifier | undefined, typeExpression: JSDocSignature, fullName: Identifier | JSDocNamespaceDeclaration | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocCallbackTag;
        createJSDocOverloadTag(tagName: Identifier | undefined, typeExpression: JSDocSignature, comment?: string | NodeArray<JSDocComment>): JSDocOverloadTag;
        updateJSDocOverloadTag(node: JSDocOverloadTag, tagName: Identifier | undefined, typeExpression: JSDocSignature, comment: string | NodeArray<JSDocComment> | undefined): JSDocOverloadTag;
        createJSDocAugmentsTag(tagName: Identifier | undefined, className: JSDocAugmentsTag["class"], comment?: string | NodeArray<JSDocComment>): JSDocAugmentsTag;
        updateJSDocAugmentsTag(node: JSDocAugmentsTag, tagName: Identifier | undefined, className: JSDocAugmentsTag["class"], comment: string | NodeArray<JSDocComment> | undefined): JSDocAugmentsTag;
        createJSDocImplementsTag(tagName: Identifier | undefined, className: JSDocImplementsTag["class"], comment?: string | NodeArray<JSDocComment>): JSDocImplementsTag;
        updateJSDocImplementsTag(node: JSDocImplementsTag, tagName: Identifier | undefined, className: JSDocImplementsTag["class"], comment: string | NodeArray<JSDocComment> | undefined): JSDocImplementsTag;
        createJSDocAuthorTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocAuthorTag;
        updateJSDocAuthorTag(node: JSDocAuthorTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocAuthorTag;
        createJSDocClassTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocClassTag;
        updateJSDocClassTag(node: JSDocClassTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocClassTag;
        createJSDocPublicTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocPublicTag;
        updateJSDocPublicTag(node: JSDocPublicTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocPublicTag;
        createJSDocPrivateTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocPrivateTag;
        updateJSDocPrivateTag(node: JSDocPrivateTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocPrivateTag;
        createJSDocProtectedTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocProtectedTag;
        updateJSDocProtectedTag(node: JSDocProtectedTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocProtectedTag;
        createJSDocReadonlyTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocReadonlyTag;
        updateJSDocReadonlyTag(node: JSDocReadonlyTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocReadonlyTag;
        createJSDocUnknownTag(tagName: Identifier, comment?: string | NodeArray<JSDocComment>): JSDocUnknownTag;
        updateJSDocUnknownTag(node: JSDocUnknownTag, tagName: Identifier, comment: string | NodeArray<JSDocComment> | undefined): JSDocUnknownTag;
        createJSDocDeprecatedTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocDeprecatedTag;
        updateJSDocDeprecatedTag(node: JSDocDeprecatedTag, tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocDeprecatedTag;
        createJSDocOverrideTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocOverrideTag;
        updateJSDocOverrideTag(node: JSDocOverrideTag, tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocOverrideTag;
        createJSDocThrowsTag(tagName: Identifier, typeExpression: JSDocTypeExpression | undefined, comment?: string | NodeArray<JSDocComment>): JSDocThrowsTag;
        updateJSDocThrowsTag(node: JSDocThrowsTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | undefined, comment?: string | NodeArray<JSDocComment> | undefined): JSDocThrowsTag;
        createJSDocSatisfiesTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocSatisfiesTag;
        updateJSDocSatisfiesTag(node: JSDocSatisfiesTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment: string | NodeArray<JSDocComment> | undefined): JSDocSatisfiesTag;
        createJSDocImportTag(tagName: Identifier | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes?: ImportAttributes, comment?: string | NodeArray<JSDocComment>): JSDocImportTag;
        updateJSDocImportTag(node: JSDocImportTag, tagName: Identifier | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes: ImportAttributes | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocImportTag;
        createJSDocText(text: string): JSDocText;
        updateJSDocText(node: JSDocText, text: string): JSDocText;
        createJSDocComment(comment?: string | NodeArray<JSDocComment> | undefined, tags?: readonly JSDocTag[] | undefined): JSDoc;
        updateJSDocComment(node: JSDoc, comment: string | NodeArray<JSDocComment> | undefined, tags: readonly JSDocTag[] | undefined): JSDoc;
        createJsxElement(openingElement: JsxOpeningElement, children: readonly JsxChild[], closingElement: JsxClosingElement): JsxElement;
        updateJsxElement(node: JsxElement, openingElement: JsxOpeningElement, children: readonly JsxChild[], closingElement: JsxClosingElement): JsxElement;
        createJsxSelfClosingElement(tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxSelfClosingElement;
        updateJsxSelfClosingElement(node: JsxSelfClosingElement, tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxSelfClosingElement;
        createJsxOpeningElement(tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxOpeningElement;
        updateJsxOpeningElement(node: JsxOpeningElement, tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxOpeningElement;
        createJsxClosingElement(tagName: JsxTagNameExpression): JsxClosingElement;
        updateJsxClosingElement(node: JsxClosingElement, tagName: JsxTagNameExpression): JsxClosingElement;
        createJsxFragment(openingFragment: JsxOpeningFragment, children: readonly JsxChild[], closingFragment: JsxClosingFragment): JsxFragment;
        createJsxText(text: string, containsOnlyTriviaWhiteSpaces?: boolean): JsxText;
        updateJsxText(node: JsxText, text: string, containsOnlyTriviaWhiteSpaces?: boolean): JsxText;
        createJsxOpeningFragment(): JsxOpeningFragment;
        createJsxJsxClosingFragment(): JsxClosingFragment;
        updateJsxFragment(node: JsxFragment, openingFragment: JsxOpeningFragment, children: readonly JsxChild[], closingFragment: JsxClosingFragment): JsxFragment;
        createJsxAttribute(name: JsxAttributeName, initializer: JsxAttributeValue | undefined): JsxAttribute;
        updateJsxAttribute(node: JsxAttribute, name: JsxAttributeName, initializer: JsxAttributeValue | undefined): JsxAttribute;
        createJsxAttributes(properties: readonly JsxAttributeLike[]): JsxAttributes;
        updateJsxAttributes(node: JsxAttributes, properties: readonly JsxAttributeLike[]): JsxAttributes;
        createJsxSpreadAttribute(expression: Expression): JsxSpreadAttribute;
        updateJsxSpreadAttribute(node: JsxSpreadAttribute, expression: Expression): JsxSpreadAttribute;
        createJsxExpression(dotDotDotToken: DotDotDotToken | undefined, expression: Expression | undefined): JsxExpression;
        updateJsxExpression(node: JsxExpression, expression: Expression | undefined): JsxExpression;
        createJsxNamespacedName(namespace: Identifier, name: Identifier): JsxNamespacedName;
        updateJsxNamespacedName(node: JsxNamespacedName, namespace: Identifier, name: Identifier): JsxNamespacedName;
        createCaseClause(expression: Expression, statements: readonly Statement[]): CaseClause;
        updateCaseClause(node: CaseClause, expression: Expression, statements: readonly Statement[]): CaseClause;
        createDefaultClause(statements: readonly Statement[]): DefaultClause;
        updateDefaultClause(node: DefaultClause, statements: readonly Statement[]): DefaultClause;
        createHeritageClause(token: HeritageClause["token"], types: readonly ExpressionWithTypeArguments[]): HeritageClause;
        updateHeritageClause(node: HeritageClause, types: readonly ExpressionWithTypeArguments[]): HeritageClause;
        createCatchClause(variableDeclaration: string | BindingName | VariableDeclaration | undefined, block: Block): CatchClause;
        updateCatchClause(node: CatchClause, variableDeclaration: VariableDeclaration | undefined, block: Block): CatchClause;
        createPropertyAssignment(name: string | PropertyName, initializer: Expression): PropertyAssignment;
        updatePropertyAssignment(node: PropertyAssignment, name: PropertyName, initializer: Expression): PropertyAssignment;
        createShorthandPropertyAssignment(name: string | Identifier, objectAssignmentInitializer?: Expression): ShorthandPropertyAssignment;
        updateShorthandPropertyAssignment(node: ShorthandPropertyAssignment, name: Identifier, objectAssignmentInitializer: Expression | undefined): ShorthandPropertyAssignment;
        createSpreadAssignment(expression: Expression): SpreadAssignment;
        updateSpreadAssignment(node: SpreadAssignment, expression: Expression): SpreadAssignment;
        createEnumMember(name: string | PropertyName, initializer?: Expression): EnumMember;
        updateEnumMember(node: EnumMember, name: PropertyName, initializer: Expression | undefined): EnumMember;
        createSourceFile(statements: readonly Statement[], endOfFileToken: EndOfFileToken, flags: NodeFlags): SourceFile;
        updateSourceFile(node: SourceFile, statements: readonly Statement[], isDeclarationFile?: boolean, referencedFiles?: readonly FileReference[], typeReferences?: readonly FileReference[], hasNoDefaultLib?: boolean, libReferences?: readonly FileReference[]): SourceFile;
        createNotEmittedStatement(original: Node): NotEmittedStatement;
        createNotEmittedTypeElement(): NotEmittedTypeElement;
        createPartiallyEmittedExpression(expression: Expression, original?: Node): PartiallyEmittedExpression;
        updatePartiallyEmittedExpression(node: PartiallyEmittedExpression, expression: Expression): PartiallyEmittedExpression;
        createCommaListExpression(elements: readonly Expression[]): CommaListExpression;
        updateCommaListExpression(node: CommaListExpression, elements: readonly Expression[]): CommaListExpression;
        createBundle(sourceFiles: readonly SourceFile[]): Bundle;
        updateBundle(node: Bundle, sourceFiles: readonly SourceFile[]): Bundle;
        createComma(left: Expression, right: Expression): BinaryExpression;
        createAssignment(left: ObjectLiteralExpression | ArrayLiteralExpression, right: Expression): DestructuringAssignment;
        createAssignment(left: Expression, right: Expression): AssignmentExpression<EqualsToken>;
        createLogicalOr(left: Expression, right: Expression): BinaryExpression;
        createLogicalAnd(left: Expression, right: Expression): BinaryExpression;
        createBitwiseOr(left: Expression, right: Expression): BinaryExpression;
        createBitwiseXor(left: Expression, right: Expression): BinaryExpression;
        createBitwiseAnd(left: Expression, right: Expression): BinaryExpression;
        createStrictEquality(left: Expression, right: Expression): BinaryExpression;
        createStrictInequality(left: Expression, right: Expression): BinaryExpression;
        createEquality(left: Expression, right: Expression): BinaryExpression;
        createInequality(left: Expression, right: Expression): BinaryExpression;
        createLessThan(left: Expression, right: Expression): BinaryExpression;
        createLessThanEquals(left: Expression, right: Expression): BinaryExpression;
        createGreaterThan(left: Expression, right: Expression): BinaryExpression;
        createGreaterThanEquals(left: Expression, right: Expression): BinaryExpression;
        createLeftShift(left: Expression, right: Expression): BinaryExpression;
        createRightShift(left: Expression, right: Expression): BinaryExpression;
        createUnsignedRightShift(left: Expression, right: Expression): BinaryExpression;
        createAdd(left: Expression, right: Expression): BinaryExpression;
        createSubtract(left: Expression, right: Expression): BinaryExpression;
        createMultiply(left: Expression, right: Expression): BinaryExpression;
        createDivide(left: Expression, right: Expression): BinaryExpression;
        createModulo(left: Expression, right: Expression): BinaryExpression;
        createExponent(left: Expression, right: Expression): BinaryExpression;
        createPrefixPlus(operand: Expression): PrefixUnaryExpression;
        createPrefixMinus(operand: Expression): PrefixUnaryExpression;
        createPrefixIncrement(operand: Expression): PrefixUnaryExpression;
        createPrefixDecrement(operand: Expression): PrefixUnaryExpression;
        createBitwiseNot(operand: Expression): PrefixUnaryExpression;
        createLogicalNot(operand: Expression): PrefixUnaryExpression;
        createPostfixIncrement(operand: Expression): PostfixUnaryExpression;
        createPostfixDecrement(operand: Expression): PostfixUnaryExpression;
        createImmediatelyInvokedFunctionExpression(statements: readonly Statement[]): CallExpression;
        createImmediatelyInvokedFunctionExpression(statements: readonly Statement[], param: ParameterDeclaration, paramValue: Expression): CallExpression;
        createImmediatelyInvokedArrowFunction(statements: readonly Statement[]): ImmediatelyInvokedArrowFunction;
        createImmediatelyInvokedArrowFunction(statements: readonly Statement[], param: ParameterDeclaration, paramValue: Expression): ImmediatelyInvokedArrowFunction;
        createVoidZero(): VoidExpression;
        createExportDefault(expression: Expression): ExportAssignment;
        createExternalModuleExport(exportName: Identifier): ExportDeclaration;
        restoreOuterExpressions(outerExpression: Expression | undefined, innerExpression: Expression, kinds?: OuterExpressionKinds): Expression;
        /**
         * Updates a node that may contain modifiers, replacing only the modifiers of the node.
         */
        replaceModifiers<T extends HasModifiers>(node: T, modifiers: readonly Modifier[] | ModifierFlags | undefined): T;
        /**
         * Updates a node that may contain decorators or modifiers, replacing only the decorators and modifiers of the node.
         */
        replaceDecoratorsAndModifiers<T extends HasModifiers & HasDecorators>(node: T, modifiers: readonly ModifierLike[] | undefined): T;
        /**
         * Updates a node that contains a property name, replacing only the name of the node.
         */
        replacePropertyName<T extends AccessorDeclaration | MethodDeclaration | MethodSignature | PropertyDeclaration | PropertySignature | PropertyAssignment>(node: T, name: T["name"]): T;
    }
```

### TransformationContext

```typescript
interface TransformationContext extends CoreTransformationContext {
        /** Records a request for a non-scoped emit helper in the current context. */
        requestEmitHelper(helper: EmitHelper): void;
        /** Gets and resets the requested non-scoped emit helpers. */
        readEmitHelpers(): EmitHelper[] | undefined;
        /** Enables expression substitutions in the pretty printer for the provided SyntaxKind. */
        enableSubstitution(kind: SyntaxKind): void;
        /** Determines whether expression substitutions are enabled for the provided node. */
        isSubstitutionEnabled(node: Node): boolean;
        /**
         * Hook used by transformers to substitute expressions just before they
         * are emitted by the pretty printer.
         *
         * NOTE: Transformation hooks should only be modified during `Transformer` initialization,
         * before returning the `NodeTransformer` callback.
         */
        onSubstituteNode: (hint: EmitHint, node: Node) => Node;
        /**
         * Enables before/after emit notifications in the pretty printer for the provided
         * SyntaxKind.
         */
        enableEmitNotification(kind: SyntaxKind): void;
        /**
         * Determines whether before/after emit notifications should be raised in the pretty
         * printer when it emits a node.
         */
        isEmitNotificationEnabled(node: Node): boolean;
        /**
         * Hook used to allow transformers to capture state before or after
         * the printer emits a node.
         *
         * NOTE: Transformation hooks should only be modified during `Transformer` initialization,
         * before returning the `NodeTransformer` callback.
         */
        onEmitNode: (hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void) => void;
    }
```

### CoreTransformationContext (inherited owner)

```typescript
interface CoreTransformationContext {
        readonly factory: NodeFactory;
        /** Gets the compiler options supplied to the transformer. */
        getCompilerOptions(): CompilerOptions;
        /** Starts a new lexical environment. */
        startLexicalEnvironment(): void;
        /** Suspends the current lexical environment, usually after visiting a parameter list. */
        suspendLexicalEnvironment(): void;
        /** Resumes a suspended lexical environment, usually before visiting a function body. */
        resumeLexicalEnvironment(): void;
        /** Ends a lexical environment, returning any declarations. */
        endLexicalEnvironment(): Statement[] | undefined;
        /** Hoists a function declaration to the containing scope. */
        hoistFunctionDeclaration(node: FunctionDeclaration): void;
        /** Hoists a variable declaration to the containing scope. */
        hoistVariableDeclaration(node: Identifier): void;
    }
```

## Raw evidence hashes

[raw-manifest.json](evidence/raw-manifest.json) records byte sizes and SHA-256 for every raw trace/log and deterministic gzip archive. Large raw artifacts (at least 100,000 bytes) are listed here. Archives are retained in private scratch; if permitted by automatic review they are published separately under evidence/raw. If an archive is absent from the branch, its hash and raw file remain available in scratch; no blocked payload is forced.

| Raw file | Bytes | SHA-256 |
| --- | ---: | --- |
| census/emitter-times.jsonl | 117611 | `d5aba47bb368955ed34ba01d0518ee26181356d2de6f9c4c02396113a8ba2387` |
| census/emitter.jsonl | 2487598 | `9869e70f6239a9a9a059df4f51b4170f6c90d4f7d11ebc247abd09b626dc7b63` |
| census/expressionToTypeNode.jsonl | 478907 | `d2f84ffbf07aa1bdf72a5e39813f4f258413c675777fce94b1eb7d96863a91f5` |
| census/moduleSpecifiers.jsonl | 394631 | `b702fb716c946385a1cd87921bf7872daac500667ec85f48c26997927389cf1a` |
| census/parser-times.jsonl | 150517 | `b205d83c95965b0a09de7a8567bf4f514d9abd0c364f6c365f11b6193e429353` |
| census/parser.jsonl | 5146635 | `cabb35d792fff5b68de7972ab7b95cd86f207310245e31727aeece5256723bb4` |
| census/sourcemap.jsonl | 294202 | `8fc714535a4929763d06743115fef5d0f4dae79cbdad2492dc9aaeb3880b3906` |
| census/sys.jsonl | 637089 | `1a2aa3ce929203007df9c045e2ac8f15cf2d5ce3e9f90f1ac8bfd43a23dfaccc` |
| census/transformers__classFields.jsonl | 989267 | `6c81eeaa5e01b242ef9cb382245efa57b062a24ba2c3e54ec495aa45559de319` |
| census/transformers__es2015.jsonl | 1595577 | `67e585bc53bf93337758f0c81e151be094698ac73b56af59e65aa02ec24b789a` |
| census/transformers__es2017.jsonl | 488973 | `02a4dc272067118ece3eee3a5c5780fd2d4ee2a58c098d312824905c4c3dcbcf` |
| census/transformers__es2018.jsonl | 619360 | `7b44ea4a22f484c934e3ec6564390d144efa292227135c9663e45d6e5452be7d` |
| census/transformers__es2020.jsonl | 158740 | `83cf9682bf2249aa3598d090afa41b69267af4fbb7d6006d5a2855205ef927d5` |
| census/transformers__esDecorators.jsonl | 806780 | `953bcf3c4433c3a60f98a23b5d54eae1774bca69d6e3572fdb8f72740e8ca8c9` |
| census/transformers__esnext.jsonl | 301247 | `e60529a5ac7308acdabc1d1f7e3df15717307dc679a824ee3c7501153ee1f537` |
| census/transformers__generators.jsonl | 956100 | `6ed8871348cc58a314229714b7c0c3de45123f9e24cf044126416ff2456d02c7` |
| census/transformers__jsx.jsonl | 304907 | `52d9ee6379a819db56fb2d122f4b851df835bb7ec60a562fb9717ff52e72e8ab` |
| census/transformers__legacyDecorators.jsonl | 371921 | `8f9141e08c8c35412b7ba56069dfa376ca4d23dfe3c9f4048a241f3d5e582023` |
| census/transformers__module__esnextAnd2015.jsonl | 157897 | `ab9fe798acb81f39d3d8cfa926ca630c555f2f168d89b75834f0d06b7751a3c0` |
| census/transformers__module__module.jsonl | 640167 | `06909bad51fd45df7cc21dd57dfe9e4b4ae4eff8cc5f32de08aa987ea1cb18e1` |
| census/transformers__module__system.jsonl | 601877 | `18b207190df514bcf88209e58053f21334bf1efe4477aeb76fdba0cce3b1eb21` |
| census/transformers__ts.jsonl | 944412 | `6ebb205cbe94e7ef664eba0e8a94bf3eb23b41aacef63f908e27a72d62aade07` |
| census/transformers__typeSerializer.jsonl | 171044 | `24bdc684aac49ce04bdacfff01a7dfcd18c852e3234b83098e3b6ca2d652af9d` |
| census/utilities-times.jsonl | 244804 | `828ad3e29592f4a5111fdb09f0f122e6cdaac56b9f9ad1bed48cafb52da0660c` |
| census/utilities.jsonl | 3942150 | `f4ee02e9c9cc39fdae37b7f887c5af9a69f26d0c6b6432165d017b7bca6eec35` |
| clean-oracle/report.json | 169524 | `6cb5d258cfe73f5783edb9132bb8c6db947f540b8a49e6bbb5136eeb97ca26ff` |

## Provenance and scope

Source pin: Adamic `8e3a784aa5a601f57342c07b68f145fe56252c21`, fresh private detached worktree before this report branch. Recursive shallow submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, compiler fork `d92d9bfee114c80be2c375d72edae966176e3a4f`. Upstream v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Node v24.19.0; Go go1.27.1. Adapter npm ci, upstream npm ci and stage3/apply.sh completed. ADAMIC_TYPESCRIPT_SOURCE, temporary directories, Go build/module caches and npm caches were task scratch. The original /tmp task paths now point to disk-backed scratch through a symlink.

The census tool is the repository’s scratch Go overlay, built with CommonJSBundle and no-output guards. The counting copy and two source recipes are separate from the real adapted baseline. The shared checkout, compiler and method rule are unchanged. Only this report, measurement scripts and evidence are committed; no Go files are added.

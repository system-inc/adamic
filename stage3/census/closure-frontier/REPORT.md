# Function-value result frontier census

5091 source sites satisfy the recovered unknown-target, non-scalar result rule: 4460 non-scalar-only results and 631 mixed scalar/non-scalar unions. 359 additional sites are unsure and are excluded from both stop and scalar totals. 2088 unknown-target sites have scalar results.

This is a static source-site census, conditional on each site reaching the relevant IR node. It is not a count of executed calls, generic instantiations, generated adapter nodes, or successive compiler failures. Unrelated earlier lowering refusals and reachability are not simulated. A helper body is counted once per syntactic call expression, irrespective of how many callers use it.

## Inputs and rule confirmation

- Main at fetch time: `0816e18511b5e1ef4383543ee81c526fc36acfaf`. Only this directory is changed.
- TypeScript source: `050880ce59e30b356b686bd3144efe24f875ebc8`, v6.0.3, adapted by main's `stage3/apply.sh`. Patch table SHA-256: `7f7db01dc5170559132ff704cc117c67dc19424abbc1240d9f4953a32bd71cd2`.
- Checker: TypeScript 6.0.3 from `stage3/api`; compiler tsconfig roots plus `src/tsc/tsc.ts`, with strict, exactOptionalPropertyTypes and noUncheckedIndexedAccess enabled. Effective options and all source hashes are in [counts.json](counts.json).
- Probe: [39a3cebf versus 266e385b](https://github.com/system-inc/adamic/compare/266e385b...39a3cebf). The changed `graphFlowRules` entries stop exactly `CallClosure` and `ArrayReduce` when `ClosureTargets(...).Unknown && !graphFlowScalar(expression.Type())`. They do not add the same stop to ArrayMap, ArrayVisit, ArrayFrom, or MapForEach.

The proposed source reading needs these corrections:

1. [ClosureTargets](https://github.com/system-inc/adamic/blob/39a3cebf/internal/ir/call_targets.go) also bounds `CallClosure.Direct`, `BoundClosure`, and `Effects` whose result unwraps to a bounded value. Plain const aliases are not transitively followed. [locals.go](https://github.com/system-inc/adamic/blob/39a3cebf/internal/lower/locals.go) sets ConstantClosure only for an initializer that lowers directly to MakeClosure. Indirect consts have their own extra row because the six requested kinds are not exhaustive.
2. [callOrMethod](https://github.com/system-inc/adamic/blob/39a3cebf/internal/lower/class.go) routes a class MethodDeclaration to a direct or virtual Call, but structural MethodSignature declarations fall back to [call/callClosure](https://github.com/system-inc/adamic/blob/39a3cebf/internal/lower/expression.go). Consequently interface methods, as well as function-valued properties, belong in the property row. Namespace references resolving to declared functions are excluded. External host dispatch and unresolved element/property dispatch are unsure.
3. [graphFlowScalar](https://github.com/system-inc/adamic/blob/39a3cebf/internal/lower/graph_flow_classes.go) accepts IR void, number, boolean, string, MaybeNumber, MaybeBoolean, and typed arrays. The [representation function](https://github.com/system-inc/adamic/blob/39a3cebf/internal/lower/expression.go) boxes differently represented scalar unions such as string | number as ir.Union, so those stop too. Standalone undefined is ir.Object; void and never call results are separately lowered to zero. The suggested scalar shorthand disagrees at 8 classified sites, listed in counts.json's scalarShorthandDifferences. Mixed means the checker union contains both a source scalar member and a non-scalar member; object | undefined is mixed.
4. Only `.reduce(callback, initial)` is recognized by [arrayReduce](https://github.com/system-inc/adamic/blob/39a3cebf/internal/lower/object.go); reduceRight is absent from the intrinsic set. Unknown-callback reduceRight and reduce without the required initial value are unsure. Closure literals and direct const closure callbacks are bounded. Intrinsic callback representation adapters can introduce a bounded outer closure plus an unknown inner CallClosure; generated sites have no separate source call and are outside this source-site census.
5. `optionalCall` generally refuses `f?.()` before CallClosure. `optionalIntrinsic` can rewrite chains and `object?.method()` has exceptions. Every optional-chain candidate is therefore kept in unsure rather than presumed to hit the probe. Generic result parameters, any, conditional/indexed-access results, asserted callable expressions/initializers, and unresolved callable expressions are also unsure. Arrays or objects containing a generic element still have a known non-scalar outer representation.

## Counts as JSON

```json
{
  "totals": {
    "nonScalar": 4460,
    "mixed": 631,
    "scalar": 2088,
    "unsure": 359,
    "stops": 5091
  },
  "byKind": {
    "parameter": {
      "nonScalar": 104,
      "mixed": 316,
      "scalar": 381,
      "unsure": 124,
      "stops": 420
    },
    "property": {
      "nonScalar": 4034,
      "mixed": 242,
      "scalar": 1244,
      "unsure": 192,
      "stops": 4276
    },
    "let or var": {
      "nonScalar": 61,
      "mixed": 40,
      "scalar": 83,
      "unsure": 19,
      "stops": 101
    },
    "returned value": {
      "nonScalar": 5,
      "mixed": 1,
      "scalar": 1,
      "unsure": 0,
      "stops": 6
    },
    "join": {
      "nonScalar": 5,
      "mixed": 1,
      "scalar": 8,
      "unsure": 1,
      "stops": 6
    },
    "reduce callback": {
      "nonScalar": 0,
      "mixed": 0,
      "scalar": 0,
      "unsure": 0,
      "stops": 0
    },
    "const indirect": {
      "nonScalar": 251,
      "mixed": 31,
      "scalar": 371,
      "unsure": 23,
      "stops": 282
    },
    "other": {
      "nonScalar": 0,
      "mixed": 0,
      "scalar": 0,
      "unsure": 0,
      "stops": 0
    }
  },
  "compilerFiles": 79,
  "compilerFilesOutsideEntry": [],
  "scalarShorthandDifferences": 8
}
```

Stops = nonScalar + mixed. Unsure is disjoint from every other column. The other row holds cases that cannot safely be assigned a requested kind. The const indirect row includes immutable aliases, destructured const function values and for-of const function bindings. Destructured parameter function values are parameters.

## By callee kind

| Kind | Stops | Non-scalar only | Mixed | Scalar | Unsure |
|---|---|---|---|---|---|
| parameter | 420 | 104 | 316 | 381 | 124 |
| property | 4276 | 4034 | 242 | 1244 | 192 |
| let or var | 101 | 61 | 40 | 83 | 19 |
| returned value | 6 | 5 | 1 | 1 | 0 |
| join | 6 | 5 | 1 | 8 | 1 |
| reduce callback | 0 | 0 | 0 | 0 | 0 |
| const indirect | 282 | 251 | 31 | 371 | 23 |
| other | 0 | 0 | 0 | 0 | 0 |

## By file, descending stop count

Load order is a 1-based, dependency-first completion index from src/tsc/tsc.ts. The walk follows static imports and re-exports in source order, skips explicit type-only edges, and marks modules on entry to break cycles. Namespace barrel and supporting modules occupy indices too. This is module initialization order, not the time a function body executes. Ties sort by unsure count then filename. All 79 compiler files are listed, including zeros; an absent entry-program index is shown as outside entry. The full order is in counts.json's evaluationOrder.

| File | Load index | Stops | Non-scalar only | Mixed | Scalar | Unsure |
|---|---|---|---|---|---|---|
| src/compiler/checker.ts | 34 | 594 | 498 | 96 | 300 | 30 |
| src/compiler/visitorPublic.ts | 35 | 484 | 207 | 277 | 13 | 1 |
| src/compiler/transformers/es2015.ts | 55 | 416 | 408 | 8 | 37 | 0 |
| src/compiler/transformers/module/module.ts | 57 | 314 | 305 | 9 | 21 | 0 |
| src/compiler/transformers/esDecorators.ts | 46 | 276 | 273 | 3 | 5 | 0 |
| src/compiler/parser.ts | 27 | 266 | 263 | 3 | 184 | 18 |
| src/compiler/factory/nodeFactory.ts | 20 | 223 | 219 | 4 | 17 | 1 |
| src/compiler/transformers/classFields.ts | 43 | 202 | 197 | 5 | 31 | 3 |
| src/compiler/transformers/ts.ts | 42 | 187 | 179 | 8 | 23 | 1 |
| src/compiler/transformers/module/system.ts | 58 | 178 | 169 | 9 | 15 | 0 |
| src/compiler/transformers/es2018.ts | 48 | 174 | 171 | 3 | 24 | 2 |
| src/compiler/transformers/declarations.ts | 62 | 152 | 138 | 14 | 34 | 0 |
| src/compiler/transformers/generators.ts | 56 | 131 | 128 | 3 | 13 | 0 |
| src/compiler/factory/emitHelpers.ts | 22 | 128 | 128 | 0 | 35 | 0 |
| src/compiler/expressionToTypeNode.ts | 78 | 124 | 110 | 14 | 46 | 4 |
| src/compiler/program.ts | 66 | 119 | 81 | 38 | 206 | 23 |
| src/compiler/transformers/esnext.ts | 52 | 89 | 86 | 3 | 3 | 0 |
| src/compiler/factory/utilities.ts | 25 | 84 | 81 | 3 | 5 | 4 |
| src/compiler/transformers/es2017.ts | 47 | 80 | 78 | 2 | 29 | 2 |
| src/compiler/transformers/jsx.ts | 53 | 66 | 62 | 4 | 1 | 0 |
| src/compiler/transformers/typeSerializer.ts | 44 | 60 | 60 | 0 | 2 | 2 |
| src/compiler/transformers/module/esnextAnd2015.ts | 59 | 59 | 58 | 1 | 12 | 0 |
| src/compiler/builder.ts | 70 | 58 | 38 | 20 | 27 | 3 |
| src/compiler/transformers/legacyDecorators.ts | 45 | 55 | 53 | 2 | 5 | 0 |
| src/compiler/utilities.ts | 16 | 43 | 28 | 15 | 158 | 24 |
| src/compiler/transformers/destructuring.ts | 38 | 40 | 40 | 0 | 10 | 0 |
| src/compiler/emitter.ts | 64 | 39 | 35 | 4 | 144 | 19 |
| src/compiler/transformers/es2020.ts | 50 | 39 | 39 | 0 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 40 | 37 | 37 | 0 | 1 | 0 |
| src/compiler/resolutionCache.ts | 72 | 35 | 24 | 11 | 69 | 19 |
| src/compiler/factory/parenthesizerRules.ts | 18 | 31 | 31 | 0 | 0 | 0 |
| src/compiler/sys.ts | 11 | 25 | 15 | 10 | 60 | 51 |
| src/compiler/watch.ts | 73 | 24 | 23 | 1 | 55 | 2 |
| src/compiler/moduleNameResolver.ts | 29 | 23 | 10 | 13 | 37 | 18 |
| src/compiler/builderState.ts | 69 | 23 | 15 | 8 | 5 | 0 |
| src/compiler/watchPublic.ts | 74 | 20 | 14 | 6 | 43 | 7 |
| src/compiler/tsbuildPublic.ts | 76 | 18 | 12 | 6 | 47 | 20 |
| src/compiler/watchUtilities.ts | 65 | 15 | 13 | 2 | 44 | 8 |
| src/compiler/programDiagnostics.ts | 67 | 13 | 4 | 9 | 9 | 1 |
| src/compiler/transformers/es2016.ts | 54 | 13 | 13 | 0 | 0 | 0 |
| src/compiler/transformer.ts | 63 | 12 | 12 | 0 | 6 | 2 |
| src/compiler/factory/nodeConverters.ts | 19 | 12 | 12 | 0 | 0 | 0 |
| src/compiler/transformers/es2021.ts | 51 | 11 | 11 | 0 | 0 | 0 |
| src/compiler/transformers/taggedTemplate.ts | 41 | 11 | 11 | 0 | 1 | 0 |
| src/compiler/transformers/utilities.ts | 37 | 10 | 8 | 2 | 0 | 1 |
| src/compiler/symbolWalker.ts | 31 | 10 | 8 | 2 | 1 | 0 |
| src/compiler/moduleSpecifiers.ts | 32 | 9 | 5 | 4 | 38 | 15 |
| src/compiler/debug.ts | 3 | 8 | 8 | 0 | 6 | 12 |
| src/compiler/executeCommandLine.ts | 77 | 8 | 8 | 0 | 95 | 5 |
| src/compiler/transformers/classThis.ts | 39 | 8 | 8 | 0 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 60 | 7 | 7 | 0 | 6 | 0 |
| src/compiler/commandLineParser.ts | 28 | 6 | 3 | 3 | 20 | 3 |
| src/compiler/core.ts | 2 | 5 | 4 | 1 | 58 | 36 |
| src/compiler/factory/baseNodeFactory.ts | 17 | 5 | 5 | 0 | 0 | 0 |
| src/compiler/binder.ts | 30 | 4 | 2 | 2 | 11 | 0 |
| src/compiler/transformers/es2019.ts | 49 | 3 | 3 | 0 | 0 | 0 |
| src/compiler/sourcemap.ts | 36 | 2 | 0 | 2 | 18 | 0 |
| src/compiler/tracing.ts | 9 | 1 | 0 | 1 | 22 | 2 |
| src/compiler/builderPublic.ts | 71 | 1 | 1 | 0 | 0 | 0 |
| src/compiler/utilitiesPublic.ts | 15 | 1 | 1 | 0 | 10 | 0 |
| src/compiler/factory/nodeChildren.ts | 24 | 0 | 0 | 0 | 0 | 10 |
| src/compiler/performance.ts | 6 | 0 | 0 | 0 | 6 | 5 |
| src/compiler/scanner.ts | 14 | 0 | 0 | 0 | 7 | 4 |
| src/compiler/path.ts | 12 | 0 | 0 | 0 | 10 | 1 |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 33 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 7 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 79 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderStatePublic.ts | 68 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/corePublic.ts | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/emitNode.ts | 21 | 0 | 0 | 0 | 1 | 0 |
| src/compiler/factory/nodeTests.ts | 23 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilitiesPublic.ts | 26 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/hostErrors.ts | 8 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/performanceCore.ts | 5 | 0 | 0 | 0 | 1 | 0 |
| src/compiler/semver.ts | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | 61 | 0 | 0 | 0 | 1 | 0 |
| src/compiler/tsbuild.ts | 75 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/types.ts | 10 | 0 | 0 | 0 | 0 | 0 |

## Twenty sample sites

Scalar and unsure core helper controls are included deliberately. Parameter callbacks behind map, filter and forEach-style helpers are individual source sites, even when many compiler callers pass through them. Exact unabridged call text and checker type strings for every candidate are in counts.json.

1. **src/compiler/core.ts:36:28**; parameter; unsure; helper/owner ` forEach `.

   Call: ` callback(array[i]!, i) `

   Checker result: ` U | undefined `

2. **src/compiler/core.ts:53:28**; parameter; unsure; helper/owner ` forEachRight `.

   Call: ` callback(array[i]!, i) `

   Checker result: ` U | undefined `

3. **src/compiler/core.ts:325:25**; parameter; unsure; helper/owner ` map `.

   Call: ` f(array[i]!, i) `

   Checker result: ` U `

4. **src/compiler/core.ts:281:27**; parameter; scalar; helper/owner ` filter `.

   Call: ` f(array[i]!) `

   Checker result: ` boolean `

5. **src/compiler/core.ts:354:28**; parameter; unsure; helper/owner ` sameMap `.

   Call: ` f(item, i) `

   Checker result: ` U `

6. **src/compiler/core.ts:403:23**; parameter; unsure; helper/owner ` flatMap `.

   Call: ` mapfn(array[i]!, i) `

   Checker result: ` U | readonly U[] | undefined `

7. **src/compiler/core.ts:439:23**; parameter; mixed; helper/owner ` flatMapIterator `.

   Call: ` mapfn(x) `

   Checker result: ` readonly U[] | Iterable<U> | undefined `

8. **src/compiler/debug.ts:588:33**; property; nonScalar; helper/owner ` enableDebugInfo `.

   Call: ` objectAllocator.getSymbolConstructor() `

   Checker result: ` new (flags: SymbolFlags, name: __String) => Symbol `

9. **src/compiler/sys.ts:1054:24**; let or var; nonScalar; helper/owner ` watchFile `.

   Call: ` nonPollingWatchFile(fileName, callback, pollingInterval, getFallbackOptions(options)) `

   Checker result: ` FileWatcher `

10. **src/compiler/sys.ts:1038:24**; returned value; nonScalar; helper/owner ` watchFile `.

   Call: ` ensureDynamicPollingWatchFile()(fileName, callback, pollingInterval, /*options*/ undefined) `

   Checker result: ` FileWatcher `

11. **src/compiler/utilities.ts:4670:39**; join; nonScalar; helper/owner ` getJSDocCommentsAndTags `.

   Call: ` (noCache ? getJSDocParameterTagsNoCache : getJSDocParameterTags)(node as ParameterDeclaration) `

   Checker result: ` readonly JSDocParameterTag[] `

12. **src/compiler/core.ts:2137:33**; const indirect; nonScalar; helper/owner ` compareStringsCaseSensitiveUI `.

   Call: ` createUIStringComparer(uiLocale) `

   Checker result: ` Comparer<string> `

13. **src/compiler/debug.ts:180:13**; property; unsure; helper/owner ` fail `.

   Call: ` Error.captureStackTrace(e, stackCrawlMark || fail) `

   Checker result: ` void `

14. **src/compiler/tracing.ts:307:39**; property; mixed; helper/owner ` dumpTypes `.

   Call: ` type.checker.getRecursionIdentity(type) `

   Checker result: ` object | undefined `

15. **src/compiler/debug.ts:605:33**; property; nonScalar; helper/owner ` enableDebugInfo `.

   Call: ` objectAllocator.getTypeConstructor() `

   Checker result: ` new (checker: TypeChecker, flags: TypeFlags) => Type `

16. **src/compiler/debug.ts:836:44**; parameter; nonScalar; helper/owner ` __debugToString `.

   Call: ` t() `

   Checker result: ` Type `

17. **src/compiler/sys.ts:118:12**; property; mixed; helper/owner ` getModifiedTime `.

   Call: ` host.getModifiedTime(fileName) `

   Checker result: ` Date | undefined `

18. **src/compiler/sys.ts:404:42**; parameter; mixed; helper/owner ` nonPollingWatchFile `.

   Call: ` getModifiedTime(fileName) `

   Checker result: ` Date | undefined `

19. **src/compiler/utilitiesPublic.ts:790:24**; parameter; nonScalar; helper/owner ` findAncestor `.

   Call: ` callback(node) `

   Checker result: ` boolean | "quit" `

20. **src/compiler/utilities.ts:764:22**; property; nonScalar; helper/owner ` forEachKey `.

   Call: ` map.keys() `

   Checker result: ` IterableIterator<K> `

## Unsure and validation

| Reason | Sites |
|---|---|
| result representation depends on generic instantiation, erased type, or scalar-union boxing | 149 |
| optional chain: optionalCall/optionalIntrinsic can refuse or rewrite before CallClosure | 128 |
| standard library receiver dispatch not proven to be an intrinsic | 31 |
| external library or host intrinsic dispatch | 30 |
| asserted const initializer may wrap or adapt closure identity | 10 |
| property or element dispatch not settled by declaration | 9 |
| asserted callable expression may wrap or adapt closure identity | 2 |

The compiler program reports 0 diagnostics. Their codes, locations and messages are retained in counts.json; checker errors are not silently suppressed. The census does not claim successful Adamic lowering of the whole program.

Validation used a separate small program covering parameter/destructured parameter calls, mixed and scalar returns, unknown returns, typed arrays, structural method signatures, function properties, mutable bindings, indirect consts, returned functions, joins, reduce callbacks, optional calls and generic results. Fifteen classification assertions and bounded/direct/virtual-call exclusion assertions passed.

Reproduce from the main base above:

```bash
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/closure-setup.log 2>&1
rc=$?; printf 'setup exit=%s\n' "$rc"; [ "$rc" -eq 0 ] || exit "$rc"
npm ci --prefix stage3/api > /tmp/closure-npm.log 2>&1
rc=$?; printf 'npm exit=%s\n' "$rc"; [ "$rc" -eq 0 ] || exit "$rc"
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript-at-050880ce59e30b356b686bd3144efe24f875ebc8
bash stage3/apply.sh /tmp/closure-adapted > /tmp/closure-apply.log 2>&1
rc=$?; printf 'apply exit=%s\n' "$rc"; [ "$rc" -eq 0 ] || exit "$rc"
NODE_PATH=stage3/api/node_modules node stage3/census/closure-frontier/census.cjs /tmp/closure-adapted stage3/census/closure-frontier > /tmp/closure-census.log 2>&1
rc=$?; printf 'census exit=%s\n' "$rc"; exit "$rc"
```

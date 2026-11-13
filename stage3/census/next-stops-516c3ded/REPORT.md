# Next stops on 516c3ded

Measured commit: `516c3dedf370293d4e4f8f78ab4628d391aed5e4` (diagnostics/assignment-value-1). Parent: `751879544ea6e627e67c91d6ab3c4ba595be6356`. Baseline: `2cda42f68690691c4ed080293684e1ff911adcfa` on typescript/next-stops-f2546f1f. Additional per-root comparison with 464e5daa is included.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

All 73 former optional-Low roots move to one shared new stop in both formats: **sys.ts:272:23**, Refused: **a cast the runtime can't check**, in the refusal scan (owner 01a11410). The other six roots are unchanged. Files lowering whole in both formats: corePublic.ts and hostErrors.ts (2/79).

Adapted source at the new stop:

```typescript
        const queue = [] as WatchedFileWithUnchangedPolls[] as PollingIntervalQueue;
```

Ordinary roots stop in the refusal scan before lowering can reach the element assignment at sys.ts:160. Full mode finds the NotYet at sys.ts:160:17 on both parent and head; that lowering gap is unchanged. The complete sys.ts delta is one removed refusal-scan finding at sys.ts:160:35, no new or changed findings, and 278 identical findings (including all 260 lowering findings). Both full runs exited 0; no scan-only fallback was needed.

## CommonJSBundle


Measured compiler/apply commit: `516c3dedf370293d4e4f8f78ab4628d391aed5e4`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:272:23 | Refused | a cast the runtime can't check | refusal_scan | 01a11410 |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:272:23 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:469:12 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with f2546f1f

Baseline: typescript/next-stops-f2546f1f at `2cda42f68690691c4ed080293684e1ff911adcfa`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:160:35 roots now go to:

- 73 to sys.ts:272:23: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| core.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| path.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| program.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| tracing.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |

### Comparison with 464e5daa

Table at `c0d27049a616498d5619a9487536c83542c775ec`.

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| core.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| path.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| program.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| tracing.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |

### Adapted source at stops

#### sys.ts:272:23

```typescript
269:     }
270: 
271:     function createPollingIntervalQueue(pollingInterval: PollingInterval): PollingIntervalQueue {
272:         const queue = [] as WatchedFileWithUnchangedPolls[] as PollingIntervalQueue;
273:         queue.pollingInterval = pollingInterval;
274:         queue.pollIndex = 0;
275:         queue.pollScheduled = false;
```

#### commandLineParser.ts:293:35

```typescript
290:     {
291:         name: "watchFile",
292:         type: new Map(Object.entries({
293:             fixedpollinginterval: WatchFileKind.FixedPollingInterval,
294:             prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
295:             dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
296:             fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

#### diagnosticInformationMap.generated.ts:13:45

```typescript
10: /* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
11: /** @internal */
12: export const Diagnostics = {
13:     Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
14:     Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
15:     _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
16:     A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

#### scanner.ts:469:12

```typescript
466: /** @internal */
467: export function getPositionOfLineAndCharacter(sourceFile: SourceFileLike, line: number, character: number, allowEdits?: true): number; // eslint-disable-line @typescript-eslint/unified-signatures
468: export function getPositionOfLineAndCharacter(sourceFile: SourceFileLike, line: number, character: number, allowEdits?: true): number {
469:     return sourceFile.getPositionOfLineAndCharacter ?
470:         sourceFile.getPositionOfLineAndCharacter(line, character, allowEdits) :
471:         computePositionOfLineAndCharacter(getLineStarts(sourceFile), line, character, sourceFile.text, allowEdits);
472: }
```

#### semver.ts:71:9

```typescript
68:         const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;
69: 
70:         Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
71:         Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
72: 
73:         this.major = major;
74:         this.minor = minor;
```

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
## EcmaScript


Measured compiler/apply commit: `516c3dedf370293d4e4f8f78ab4628d391aed5e4`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:272:23 | Refused | a cast the runtime can't check | refusal_scan | 01a11410 |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:272:23 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:469:12 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with f2546f1f

Baseline: typescript/next-stops-f2546f1f at `2cda42f68690691c4ed080293684e1ff911adcfa`.

6 unchanged exact tuples; 73 changed.

The former 73 sys.ts:160:35 roots now go to:

- 73 to sys.ts:272:23: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| core.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| path.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| program.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| tracing.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:160:35; Refused; optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields | sys.ts:272:23; Refused; a cast the runtime can't check |

### Comparison with 464e5daa

Table at `c0d27049a616498d5619a9487536c83542c775ec`.

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| core.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| path.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| program.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| tracing.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:272:23; Refused; a cast the runtime can't check |

### Adapted source at stops

#### sys.ts:272:23

```typescript
269:     }
270: 
271:     function createPollingIntervalQueue(pollingInterval: PollingInterval): PollingIntervalQueue {
272:         const queue = [] as WatchedFileWithUnchangedPolls[] as PollingIntervalQueue;
273:         queue.pollingInterval = pollingInterval;
274:         queue.pollIndex = 0;
275:         queue.pollScheduled = false;
```

#### commandLineParser.ts:293:35

```typescript
290:     {
291:         name: "watchFile",
292:         type: new Map(Object.entries({
293:             fixedpollinginterval: WatchFileKind.FixedPollingInterval,
294:             prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
295:             dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
296:             fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

#### diagnosticInformationMap.generated.ts:13:45

```typescript
10: /* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
11: /** @internal */
12: export const Diagnostics = {
13:     Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
14:     Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
15:     _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
16:     A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

#### scanner.ts:469:12

```typescript
466: /** @internal */
467: export function getPositionOfLineAndCharacter(sourceFile: SourceFileLike, line: number, character: number, allowEdits?: true): number; // eslint-disable-line @typescript-eslint/unified-signatures
468: export function getPositionOfLineAndCharacter(sourceFile: SourceFileLike, line: number, character: number, allowEdits?: true): number {
469:     return sourceFile.getPositionOfLineAndCharacter ?
470:         sourceFile.getPositionOfLineAndCharacter(line, character, allowEdits) :
471:         computePositionOfLineAndCharacter(getLineStarts(sourceFile), line, character, sourceFile.text, allowEdits);
472: }
```

#### semver.ts:71:9

```typescript
68:         const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;
69: 
70:         Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
71:         Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
72: 
73:         this.major = major;
74:         this.minor = minor;
```

### Method

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.
## sys.ts full census

Both full runs completed with exit 0 under the 1500-second bound; no scan-only fallback was used. Each used the sys.ts file as the load root and LATENT_ONLY_FILE, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, CommonJSBundle, and the same adapted input. The input is checker-clean in both runs; the tool uses its standard measurement-only label. No usable compiler output was produced.

Findings: {"parent": 279, "head": 278, "unchanged": 278, "gone": 1, "new": 0, "changed": 0}. The unchanged count compares complete finding objects, including unit, phase, kind, location, reason and text. Repeated locations in different nested units are retained. Unambiguous single old/new findings with the same phase/unit/kind are shown as changed; other differences are new or gone.

| Commit | Refusal scan | Lowering | Units | Seconds |
| --- | ---: | ---: | ---: | ---: |
| parent | 19 | 260 | 136 | 55.586 |
| head | 18 | 260 | 136 | 69.62 |

Every new, gone, or changed finding follows. Text is exact except that the common adapted compiler directory prefix is omitted. Raw absolute text is retained in JSON and raw streams.

| Change | Phase | Kind | Unit | Before text | After text |
| --- | --- | --- | --- | --- | --- |
| gone | refusal_scan | Refused | sys.ts:157:9 | sys.ts:160:35: Adamic 0.1 refuses optional property Low in Partial&lt;Levels&gt; absent from structural source {}, which can hide fields; declare Low on the source type, or build a fresh object with known fields (adamic/no-optional-widening) | — |

All 260 lowering findings are unchanged. Both commits record the NotYet at sys.ts:160:17: `stage 0 can't lower assigning an element of a value yet`, including its boundary records in three nested units. This gap is not newly introduced by the head. The ordinary census never reaches it because the refusal scan stops first at sys.ts:272:23.

Adapted line 160, where full mode reaches the lowering gap:

```typescript
                (customLevels || (customLevels = {}))[level] = Number(customLevel);
```


## Interrupted attempt

The environment reconnect removed the outer runner during CommonJSBundle roots 11–20. Its orphaned children were terminated, and incomplete evidence is preserved under interrupted-CommonJSBundle-11-20 without a fabricated group exit. The full group was repeated; primary observations use the completed bounded run. Subsequent outer runners were detached from tool sessions while retaining their process-group timeouts and actual-exit recording.

## Setup and limits

Fresh detached scratch worktrees used the full measured and parent SHAs, with pinned cohere and TypeScript submodules initialized. The upstream clone confirms v6.0.3 is 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. /tmp/adamic-gate was created before build. The initial parent local-submodule copy failed because file transport was disabled; the local-only retry enabled it and retained the same pinned SHAs. The first parent latent build exited 1 because it was invoked from the head workspace, which did not include the parent module; running the unchanged build from the parent worktree exited 0. Both statuses are retained. No compiler or adaptation changes were made. One completed apply and one ordinary build served both formats. The parent full census uses the same adapted source: the measured commits differ only in compiler code and tests.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and sha256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by sha256; no file over 20 MB is committed. Raw paired observations, actual exit statuses, compressed logs with hashes, source excerpts, both baseline comparisons, and [sys-comparison.json](sys-comparison.json) are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | 202ed5a6e725555679071dfa99c851ddfac5da13 |
| baobab: CommonJSBundle checkpoint 20/79 roots | 32e55c8169b55cfc9f817bd97821d897a9113e92 |
| baobab: CommonJSBundle checkpoint 30/79 roots | 5aa48a9d3b85272b36147c636097e8afd60ae76d |
| baobab: CommonJSBundle checkpoint 40/79 roots | 27dc8f50131e9663d2c8bba097c672ec612aa89a |
| baobab: CommonJSBundle checkpoint 50/79 roots | 6317d7f8989b1b24bf670abd495e885390da9bab |
| baobab: CommonJSBundle checkpoint 60/79 roots | 41f508a7fa7d1122ba50c80d5e8c51d4cff8f184 |
| baobab: CommonJSBundle checkpoint 70/79 roots | 255b0e48313ce8be6a6d62023cbe49948438641b |
| baobab: CommonJSBundle checkpoint 79/79 roots | 28b64fe24e87d3456c26b1387cf9d9dc2d2cdd56 |
| baobab: EcmaScript checkpoint 10/79 roots | 2e639d3ef415b32f86ffe62032f566dd82b4857e |
| baobab: EcmaScript checkpoint 20/79 roots | aba597767c35b39432746b9b6e782679e24ccf03 |
| baobab: EcmaScript checkpoint 30/79 roots | 5d1f0408ea3472fc6311ee8d58eeb1388c107223 |
| baobab: EcmaScript checkpoint 40/79 roots | 68d3c4a3cd9971b77d7cd997622e36e660dce03e |
| baobab: EcmaScript checkpoint 50/79 roots | 4d0fb975189a6be51f06a8d5e2624f9eb9607d8e |
| baobab: EcmaScript checkpoint 60/79 roots | be7c2e85b68cf233f8b7e0bd2d3cff8124b6d660 |
| baobab: EcmaScript checkpoint 70/79 roots | b9865f4c47232b530f41579a1fa2dc150149abf5 |
| baobab: EcmaScript checkpoint 79/79 roots | 5c4070f88bc369332f687d0889824a7fec2fd361 |

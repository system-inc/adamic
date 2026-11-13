# Next stops on f2546f1f

Measured commit: f2546f1fcb7c76f8dd549094401246296b6998a7. Baseline: completed 464e5daa table at c0d27049a616498d5619a9487536c83542c775ec.

Both formats independently measured all 79 compiler roots, twice each. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

All 73 former PollingInterval roots move to one shared new stop in each format: sys.ts:160:35, Refused: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields. The other six roots are unchanged. Files lowering whole in both formats: corePublic.ts and hostErrors.ts (2/79).

## CommonJSBundle


Measured compiler/apply commit: `f2546f1fcb7c76f8dd549094401246296b6998a7`. Locations and roots are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:160:35 | Refused | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | refusal_scan | internal/lower/optional_widening.go: (*lowering).optionalWideningRefusal (via refuseOptionalWidening) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:160:35 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:469:12 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison

Baseline: typescript/next-stops-464e5daa at c0d27049a616498d5619a9487536c83542c775ec.

6 unchanged exact tuples; 73 changed.

The former 73 PollingInterval roots now go to:

- 73 to sys.ts:160:35: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| _namespaces/ts.performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| _namespaces/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| binder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builderPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builderState.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builderStatePublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| checker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| commandLineParser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| core.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| executeCommandLine.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| expressionToTypeNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/baseNodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/emitHelpers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/emitNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeChildren.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeConverters.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeTests.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/parenthesizerRules.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| parser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| path.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| performanceCore.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| program.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| programDiagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| resolutionCache.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| scanner.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| semver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| sourcemap.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| symbolWalker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| tracing.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/classFields.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/classThis.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/declarations/diagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/declarations.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/destructuring.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2016.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2017.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2018.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2019.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2020.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2021.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/esDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/esnext.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/generators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/jsx.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/legacyDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/esnextAnd2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/module.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/system.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/namedEvaluation.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/taggedTemplate.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/typeSerializer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| tsbuild.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| tsbuildPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| visitorPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| watch.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| watchPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| watchUtilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |

### Adapted source at stops

#### sys.ts:160:35

```typescript
157:         function setCustomLevel(level: keyof Levels) {
158:             const customLevel = getLevel(baseVariable, level);
159:             if (customLevel) {
160:                 (customLevels || (customLevels = {}))[level] = Number(customLevel);
161:             }
162:         }
163:     }
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

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from the measured owners.json or the emitting method.
## EcmaScript


Measured compiler/apply commit: `f2546f1fcb7c76f8dd549094401246296b6998a7`. Locations and roots are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:160:35 | Refused | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | refusal_scan | internal/lower/optional_widening.go: (*lowering).optionalWideningRefusal (via refuseOptionalWidening) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:160:35 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:469:12 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison

Baseline: typescript/next-stops-464e5daa at c0d27049a616498d5619a9487536c83542c775ec.

6 unchanged exact tuples; 73 changed.

The former 73 PollingInterval roots now go to:

- 73 to sys.ts:160:35: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| _namespaces/ts.performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| _namespaces/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| binder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builder.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builderPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builderState.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| builderStatePublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| checker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| commandLineParser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| core.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| executeCommandLine.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| expressionToTypeNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/baseNodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/emitHelpers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/emitNode.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeChildren.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeConverters.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeFactory.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/nodeTests.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/parenthesizerRules.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| factory/utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| moduleSpecifiers.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| parser.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| path.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| performance.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| performanceCore.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| program.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| programDiagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| resolutionCache.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| scanner.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| semver.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| sourcemap.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| symbolWalker.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| tracing.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/classFields.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/classThis.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/declarations/diagnostics.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/declarations.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/destructuring.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2016.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2017.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2018.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2019.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2020.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/es2021.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/esDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/esnext.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/generators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/jsx.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/legacyDecorators.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/esnextAnd2015.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/module.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/module/system.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/namedEvaluation.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/taggedTemplate.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/ts.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/typeSerializer.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| transformers/utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| tsbuild.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| tsbuildPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| utilitiesPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| visitorPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| watch.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| watchPublic.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| watchUtilities.ts | sys.ts:142:81; Refused; a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | sys.ts:160:35; Refused; optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |

### Adapted source at stops

#### sys.ts:160:35

```typescript
157:         function setCustomLevel(level: keyof Levels) {
158:             const customLevel = getLevel(baseVariable, level);
159:             if (customLevel) {
160:                 (customLevels || (customLevels = {}))[level] = Number(customLevel);
161:             }
162:         }
163:     }
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

One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from the measured owners.json or the emitting method.
## Setup and limits

A fresh detached worktree at the measured commit initialized both pinned submodules. The dedicated upstream clone confirms v6.0.3 is 050880ce59e30b356b686bd3144efe24f875ebc8; stage3/api npm ci exited 0. The first build exited 1 because the tools environment referenced missing /tmp/adamic-gate. Creating that directory and retrying the unchanged build exited 0. Apply ran once and exited 0. Failed and successful command statuses are retained. The largest recorded completed command duration is below 30 minutes. No compiler or adaptation change was made.

## Interrupted attempt

The environment/session reset removed the outer runner during CommonJSBundle roots 51–60. Its orphaned processes were terminated; the incomplete attempt is preserved under interrupted-CommonJSBundle-51-60 with no fabricated group exit. The full group was repeated under the bounded runner; the primary table uses that completed pair per root. Defunct process entries are excluded from the live-child check.

## Evidence

The copied recipe runs each child and group with a 1780-second bound, at most ten roots per group and three children at a time. Status files retain actual exits. Source/binary hashes, duplicate checks, per-root before/after comparisons, emitter evidence, and checkpoint history are retained beside this report. Only this census directory differs from the measured compiler.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | 4ee3c677f41d85249edefe55e6d601175fe52a14 |
| baobab: CommonJSBundle checkpoint 20/79 roots | f82e42026694c1e45a6ae7287be63c7749ba166e |
| baobab: CommonJSBundle checkpoint 30/79 roots | 6f4f6d367beb78ec374fb3278e3654ca2dd9906a |
| baobab: CommonJSBundle checkpoint 40/79 roots | aee2623602f7dbc3ad890f92817a43891ee631c2 |
| baobab: CommonJSBundle checkpoint 50/79 roots | 8e9d1f9f524819e781df8569eb654e49f52641db |
| baobab: CommonJSBundle checkpoint 60/79 roots | 1f84bd5a5d31531374062073e0578afdc140abde |
| baobab: CommonJSBundle checkpoint 70/79 roots | fb0bd4719c8121e7aec83a1adb459e8d3948b291 |
| baobab: CommonJSBundle checkpoint 79/79 roots | 261c0c19ecc1c0bac6f7757030701562d22781ac |
| baobab: EcmaScript checkpoint 10/79 roots | d13785c61408070e8954b432d3e11f6ba79197bb |
| baobab: EcmaScript checkpoint 20/79 roots | 84bbdae739c5476dd1f039820a49638606efbd74 |
| baobab: EcmaScript checkpoint 30/79 roots | 05e72cd32cf90b687df8272f60830919353f7f0e |
| baobab: EcmaScript checkpoint 40/79 roots | 36c10d7ea9d8e9e278ed41e5d9bcaf38ff508269 |
| baobab: EcmaScript checkpoint 50/79 roots | d914a272811f40c6907c0baf3934482c2790e23e |
| baobab: EcmaScript checkpoint 60/79 roots | 0ecc4fc3aa9df82cd54ad0076bc791e15b190f26 |
| baobab: EcmaScript checkpoint 70/79 roots | 5b01baf6557c7c12d96b1ad7be0a516b23132c1e |
| baobab: EcmaScript checkpoint 79/79 roots | 01bccab83d28e714691551ccd42419f4b48aa981 |

# Baobab next stops: CommonJSBundle

Measured compiler/apply commit: `464e5daa23a60588f5aaa656e95d93af05d582c3`. Locations and roots are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:142:81 | Refused | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | refusal_scan | internal/lower/invariance.go: (*lowering).wideningRefusal (via refuseWidening) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

## Root membership

- sys.ts:142:81 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:469:12 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

## Comparison

Bundle baseline: meter README Nov 13 method-change section, with exact raw head records from 3dd24184b9150872269ff035bd0f3814397ac437.

5 unchanged exact tuples; 74 changed. Full before/after tuples are in comparison.json.

The former 73 process roots now go to:

- 72 to sys.ts:142:81: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace
- 1 to scanner.ts:469:12: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it)

sys.ts root: scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it).

## Method and limits

One adapted tree built once with stage3/apply.sh; unchanged ordinary stage3/census/tool built at the measured commit; one root per bounded child (first bundle group sequential, subsequent groups up to three concurrent children with unique output paths); groups of ten (last nine), pushed to typescript/next-stops-464e5daa-checkpoint. The child environment explicitly sets CENSUS_MODULE_FORMAT=CommonJSBundle. Status JSON records every actual child exit and elapsed bound. The initial build failed due to absent checker submodules; pinned submodule initialization and build retry are retained. No compiler overlay or implementation change.

Fresh-run adaptations: copied the requested six Python scripts from 3dc2a59; repinned worktree, binary, tree and evidence paths; split format output directories; analyzer uses exact requested baselines; report and validation cover each format separately; checkpoint.py packages only completed logs and preserves hashes for oversized raw artifacts. Historical tooling notes are retained. The named source README does not exist; REPORT.md and tooling-adaptations.md supply its recipe.

Owner and phase evidence: source-evidence.json and retained internal/lower excerpts. owners.json is copied from the measured commit; a code emitter is named where there is no exact mapping. Source and binary SHA256 identify measured bytes. Validation covers all roots, duplicate agreement, exit codes, time limits and source integrity.

# Next stops on 9c322233

Measured commit: `9c3222331605a6ac129012439169d13e179da78a` (main, "Judge assignment values through their declared write handles"). Parent: `d0527d1aebbb74e9c85317272c5f87abdb968872`. Baseline: `a578b3ab0f65a5db2b651d11e1cc492fcca8f924` on typescript/next-stops-516c3ded.

Both formats independently measured all 79 compiler roots twice. 79/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.

The prediction is confirmed for all 73 roots in both formats: **sys.ts:272:23**, Refused: **a cast the runtime can't check**, in the refusal scan (owner 01a11410). Of the other six roots, five are unchanged; **sys.ts** now stops at **scanner.ts:2934:41** instead of scanner.ts:469:12. Thus 78/79 exact tuples are unchanged from the 516c3ded table in each format. Files lowering whole in both formats: corePublic.ts and hostErrors.ts (2/79).

Adapted source at the predicted shared stop:

```typescript
        const queue = [] as WatchedFileWithUnchangedPolls[] as PollingIntervalQueue;
```


## Adapted source at new stops

### scanner.ts:2934:41

```typescript
                decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
```

Exact reason: a value of type (TextRange & { value: number; })[] | undefined seen as {}[] | undefined, which can write {} where TextRange & { value: number; } is read.

Phase: refusal_scan. Owner: adaptation 70, stage 3. Emitter: internal/lower/invariance.go: (*lowering).wideningRefusal (via refuseWidening and stringWideningRefusal, rule 3100).

## CommonJSBundle


Measured compiler/apply commit: `9c3222331605a6ac129012439169d13e179da78a`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:272:23 | Refused | a cast the runtime can't check | refusal_scan | 01a11410 |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:2934:41 | Refused | a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | refusal_scan | adaptation 70, stage 3 |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:272:23 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:2934:41 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 516c3ded

Baseline: typescript/next-stops-516c3ded at `a578b3ab0f65a5db2b651d11e1cc492fcca8f924`.

78 unchanged exact tuples; 1 changed.

The former 73 sys.ts:272:23 roots now go to:

- 73 to sys.ts:272:23: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| core.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| path.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| program.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read |
| tracing.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |

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

#### scanner.ts:2934:41

```typescript
2931:             if (ch >= CharacterCodes._1 && ch <= CharacterCodes._9) {
2932:                 const start = pos;
2933:                 scanDigits();
2934:                 decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
2935:                 return true;
2936:             }
2937:             return false;
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


Measured compiler/apply commit: `9c3222331605a6ac129012439169d13e179da78a`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:272:23 | Refused | a cast the runtime can't check | refusal_scan | 01a11410 |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:2934:41 | Refused | a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | refusal_scan | adaptation 70, stage 3 |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### Root membership

- sys.ts:272:23 Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`
- - accepted: `corePublic.ts`, `hostErrors.ts`
- commandLineParser.ts:293:35 NotYet: `diagnosticInformationMap.generated.ts`
- diagnosticInformationMap.generated.ts:13:45 NotYet: `types.ts`
- scanner.ts:2934:41 Refused: `sys.ts`
- semver.ts:71:9 NotYet: `debug.ts`

Files lowering whole: 2/79: corePublic.ts, hostErrors.ts. No native-output success is claimed.

### Comparison with 516c3ded

Baseline: typescript/next-stops-516c3ded at `a578b3ab0f65a5db2b651d11e1cc492fcca8f924`.

78 unchanged exact tuples; 1 changed.

The former 73 sys.ts:272:23 roots now go to:

- 73 to sys.ts:272:23: a cast the runtime can't check

| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| _namespaces/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| binder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builder.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderState.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| builderStatePublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| checker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| commandLineParser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| core.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| corePublic.ts | -; accepted; - | -; accepted; - |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| emitter.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| executeCommandLine.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| expressionToTypeNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/baseNodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitHelpers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/emitNode.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeChildren.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeConverters.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeFactory.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/nodeTests.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/parenthesizerRules.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| factory/utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| hostErrors.ts | -; accepted; - | -; accepted; - |
| moduleNameResolver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| moduleSpecifiers.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| parser.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| path.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| performance.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| performanceCore.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| program.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| programDiagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| resolutionCache.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| scanner.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| semver.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| sourcemap.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| symbolWalker.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| sys.ts | scanner.ts:469:12; Refused; a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | scanner.ts:2934:41; Refused; a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read |
| tracing.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classFields.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/classThis.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations/diagnostics.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/declarations.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/destructuring.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2016.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2017.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2018.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2019.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2020.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/es2021.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/esnext.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/generators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/jsx.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/legacyDecorators.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/esnextAnd2015.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/module.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/module/system.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/namedEvaluation.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/taggedTemplate.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/ts.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/typeSerializer.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| transformers/utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuild.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| tsbuildPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration |
| utilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| utilitiesPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| visitorPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| watch.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchPublic.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |
| watchUtilities.ts | sys.ts:272:23; Refused; a cast the runtime can't check | sys.ts:272:23; Refused; a cast the runtime can't check |

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

#### scanner.ts:2934:41

```typescript
2931:             if (ch >= CharacterCodes._1 && ch <= CharacterCodes._9) {
2932:                 const start = pos;
2933:                 scanDigits();
2934:                 decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
2935:                 return true;
2936:             }
2937:             return false;
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
## Setup and limits

A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. Setup deviation: the initial npm ci used the default /home/agent/.npm cache outside scratch; its node_modules install was inside the scratch worktree. No shared checkout writes were made. No compiler or adaptation changes were made. One completed apply and one ordinary build served both formats. The separate sys.ts single-file census was skipped.

Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.

## Evidence

[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | 6d8bee88c7dd4ec10028c9b0b869ef38ce0a089a |
| baobab: CommonJSBundle checkpoint 20/79 roots | 4280248c52bbe21ca39f7956185f2ad3e9acb2e8 |
| baobab: CommonJSBundle checkpoint 30/79 roots | 9b2d2344fc2b27324b345428218ba78db9e6db7d |
| baobab: CommonJSBundle checkpoint 40/79 roots | 7eaf0bc14adc8db84f3f0784e39a24665ee2d6a1 |
| baobab: CommonJSBundle checkpoint 50/79 roots | 475de5422c4eda86a2bebab87cdf15a4f392b958 |
| baobab: CommonJSBundle checkpoint 60/79 roots | f1c0cd339729403fd4ac51e9dd3882901b46dc7d |
| baobab: CommonJSBundle checkpoint 70/79 roots | 4b74f1045ea15c5e581c68a748e0bbd2ac1affbd |
| baobab: CommonJSBundle checkpoint 79/79 roots | d0748e64a2e2067c1cc99d97d163b867d39a0ead |
| baobab: EcmaScript checkpoint 10/79 roots | 00a9c60e75b54ae0f6838939b4d582a1f1c14fef |
| baobab: EcmaScript checkpoint 20/79 roots | 71c12b9530724885c2bf246a33dd1b0c21662427 |
| baobab: EcmaScript checkpoint 30/79 roots | 54dd22b393ca9b1070b1f8da308e401929187657 |
| baobab: EcmaScript checkpoint 40/79 roots | 58e227921eb72c5201116ea85c0cc93d71e861d9 |
| baobab: EcmaScript checkpoint 50/79 roots | c9adfcf7d2400f788b1f7e2bb36ce7b31b5b174d |
| baobab: EcmaScript checkpoint 60/79 roots | 21db29baf37d7b56ba705f9db59ad991247c8adf |
| baobab: EcmaScript checkpoint 70/79 roots | 396e7cc42a1ed996224cb2a71ae4086a0a7138ea |
| baobab: EcmaScript checkpoint 79/79 roots | 0944ff50c3918d51919ea4ec48f195e9e13bd9a5 |

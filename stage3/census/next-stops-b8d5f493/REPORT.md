# Next stops on main b8d5f493

Measured per-root next stops. Report only.

Compiler and apply SHA: `b8d5f493df0227efcb33c6040ba40c906eec73c9`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `185e124556ba`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-12T03:41:59.808937+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Compared with 93e8aff5 | Root names |
| ---: | --- | --- | --- | --- | --- |
| 73 | `tracing.ts:381:1` | Refused | an ambient value declaration | 73 new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 unchanged | `debug.ts` (artifact classification incomplete) |
| 1 | `sys.ts:1654:13` | NotYet | node:inspector.Session.post | 1 unchanged | `tracing.ts` |

The 93e8aff5 baseline had 73 roots stopped at tracing.ts:340:14. This run measures their actual next stops on the fetched main tip. The measured groups and exact per-root changes below show their next stops. The forecast named tracing.ts:380:1; in the measured adapted tree line 380 is blank and the ambient declare const require is at tracing.ts:381:1.

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with 93e8aff5

Baseline: 93e8aff5467b195ec93baf4b875f71bd0d86f378, from typescript/next-stops-11f69c09 (69f58ebd). Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `tracing.ts:381:1` Refused | 73 | 0 | 0 | 73 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |
| `sys.ts:1654:13` NotYet | 1 | 1 | 0 | 0 |

Total: 6 unchanged, 0 moved, 73 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `_namespaces/ts.performance.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `_namespaces/ts.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `binder.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `builder.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `builderPublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `builderState.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `builderStatePublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `checker.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `commandLineParser.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `core.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `executeCommandLine.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `expressionToTypeNode.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/baseNodeFactory.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/emitHelpers.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/emitNode.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/nodeChildren.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/nodeConverters.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/nodeFactory.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/nodeTests.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/parenthesizerRules.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/utilities.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `factory/utilitiesPublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `moduleSpecifiers.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `parser.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `path.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `performance.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `performanceCore.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `program.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `programDiagnostics.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `resolutionCache.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `scanner.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `semver.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `sourcemap.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `symbolWalker.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `sys.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `tracing.ts` | unchanged | `sys.ts:1654:13` NotYet | `sys.ts:1654:13` NotYet |
| `transformer.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/classFields.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/classThis.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/declarations/diagnostics.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/declarations.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/destructuring.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2015.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2016.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2017.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2018.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2019.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2020.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/es2021.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/esDecorators.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/esnext.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/generators.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/jsx.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/legacyDecorators.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/module/esnextAnd2015.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/module/impliedNodeFormatDependent.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/module/module.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/module/system.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/namedEvaluation.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/taggedTemplate.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/ts.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/typeSerializer.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `transformers/utilities.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `tsbuild.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `tsbuildPublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `utilitiesPublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `visitorPublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `watch.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `watchPublic.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |
| `watchUtilities.ts` | new | `tracing.ts:340:14` Refused | `tracing.ts:381:1` Refused |

## Stop source and lowering emitters

### tracing.ts:381:1

Refused: `an ambient value declaration`. internal/lower/ambient.go:79 constructs the reason; :91 returns the Refused error.

```ts
    tracingPath?: Path;
}

declare const require: (specifier: "fs") => any;
declare const process: { readonly pid: number };
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:287 and :338 emit this NotYet reason.

```ts
    {
        name: "watchFile",
        type: new Map(Object.entries({
            fixedpollinginterval: WatchFileKind.FixedPollingInterval,
            prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
            dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
            fixedchunksizepolling: WatchFileKind.FixedChunkSizePolling,
```

### diagnosticInformationMap.generated.ts:13:45

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:287 and :338 emit this NotYet reason.

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### semver.ts:71:9

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. internal/lower/namespaces.go:289 and :341 emit this NotYet reason.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

### sys.ts:1654:13

NotYet: `node:inspector.Session.post`. internal/lower/library_node.go:639 emits this NotYet reason; merged_fields.go:134 invokes the host refusal for this merged method.

```ts
            const session = new inspector.Session();
            session.connect();

            session.post("Profiler.enable", () => {
                session.post("Profiler.start", () => {
                    activeSession = session;
                    profilePath = path;
```

## tracing.ts single-file full-mode census

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/tracing.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 45.484 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
38 records: 36 lowering, 2 refusal_scan. Kinds: 19 NotYet, 18 Boundary, 1 Refused. Refusal-scan findings: 2.

| Phase | Kind | Location | Exact reason |
| --- | --- | --- | --- |
| lowering | NotYet | `checker.ts:12423:13` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:15216:13` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:15816:64` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:17250:9` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:19176:9` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:26503:9` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:28646:9` | writing a checked object field without its source-slot type certificate |
| lowering | Boundary | `tracing.ts:66:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:66:22` | a call returning any |
| lowering | Boundary | `tracing.ts:98:9` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `tracing.ts:98:19` | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| lowering | NotYet | `tracing.ts:98:19` | reading tracingEnabled |
| lowering | Boundary | `tracing.ts:186:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:186:44` | a spread into record storage from a fixed object |
| lowering | Boundary | `tracing.ts:197:19` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:197:67` | JSON.stringify recursive array types |
| lowering | Boundary | `tracing.ts:251:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:252:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:253:46` | reading indexedAccessType |
| lowering | Boundary | `tracing.ts:260:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:261:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:262:39` | reading referenceType |
| lowering | Boundary | `tracing.ts:270:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:271:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:272:43` | reading conditionalType |
| lowering | Boundary | `tracing.ts:281:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:282:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:283:43` | reading substitutionType |
| lowering | Boundary | `tracing.ts:290:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:291:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:292:46` | reading reverseMappedType |
| lowering | Boundary | `tracing.ts:300:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:301:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:302:47` | reading evolvingArrayType |
| lowering | Boundary | `tracing.ts:319:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:341:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:341:50` | reading descriptor |
| refusal_scan | Refused | `tracing.ts:381:1` | an ambient value declaration |

## sys.ts single-file full-mode census

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/sys.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 146.487 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
396 records: 354 lowering, 41 refusal_scan, 1 deferred. Kinds: 233 Boundary, 130 NotYet, 33 Refused. Refusal-scan findings: 41.

| Phase | Kind | Location | Exact reason |
| --- | --- | --- | --- |
| lowering | Boundary | `core.ts:221:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:221:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:221:30` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:221:30` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:222:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:222:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:222:48` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:222:48` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:223:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:223:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:223:53` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:223:53` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:224:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:224:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:228:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:228:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:619:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:619:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:619:30` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:619:30` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:620:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:620:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:620:38` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:620:38` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:621:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:621:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:621:52` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:621:52` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:622:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:622:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:622:43` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:622:43` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:623:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:623:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:627:14` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:627:14` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:628:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:628:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:631:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:631:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:1941:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:1941:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:2330:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:2330:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `core.ts:2330:13` | function identity comparison in a program with specialized generic function values |
| lowering | NotYet | `core.ts:2330:13` | function identity comparison in a program with specialized generic function values |
| lowering | Boundary | `core.ts:2363:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `core.ts:2363:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `core.ts:2363:60` | function identity comparison in a program with specialized generic function values |
| lowering | NotYet | `core.ts:2363:60` | function identity comparison in a program with specialized generic function values |
| lowering | NotYet | `debug.ts:251:33` | a value of type never |
| lowering | NotYet | `debug.ts:251:33` | a value of type never |
| refusal_scan | Refused | `sys.ts:57:1` | an ambient function declaration |
| lowering | NotYet | `sys.ts:57:1` | a function without a body |
| lowering | Boundary | `sys.ts:57:1` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:57:1` | a function without a body |
| lowering | NotYet | `sys.ts:57:1` | a function without a body |
| refusal_scan | Refused | `sys.ts:58:1` | an ambient function declaration |
| lowering | NotYet | `sys.ts:58:1` | a function without a body |
| lowering | Boundary | `sys.ts:58:1` | function signature or prologue failed; continued at next declaration |
| lowering | Boundary | `sys.ts:128:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:129:9` | a computed field name |
| lowering | Boundary | `sys.ts:141:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:141:10` | a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it |
| refusal_scan | Refused | `sys.ts:144:81` | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace |
| lowering | NotYet | `sys.ts:162:17` | assigning an element of a value |
| lowering | Boundary | `sys.ts:162:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:162:17` | assigning an element of a value |
| lowering | Boundary | `sys.ts:162:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:162:17` | assigning an element of a value |
| lowering | Boundary | `sys.ts:162:17` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | Refused | `sys.ts:162:35` | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| lowering | NotYet | `sys.ts:178:13` | assigning an element of a value |
| lowering | Boundary | `sys.ts:178:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:178:13` | assigning an element of a value |
| lowering | Boundary | `sys.ts:178:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:178:13` | assigning an element of a value |
| lowering | Boundary | `sys.ts:178:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:184:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:184:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `sys.ts:185:81` | a spread after the first field |
| lowering | Refused | `sys.ts:185:81` | a spread after the first field |
| lowering | Boundary | `sys.ts:201:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:201:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:201:79` | a PostfixUnaryExpression |
| lowering | NotYet | `sys.ts:201:79` | a PostfixUnaryExpression |
| lowering | NotYet | `sys.ts:240:17` | assigning a field of a value |
| lowering | Boundary | `sys.ts:240:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:240:17` | assigning a field of a value |
| lowering | Boundary | `sys.ts:240:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:251:1` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:263:11` | a value of type PollingIntervalQueue |
| lowering | NotYet | `sys.ts:280:17` | writing a possibly absent optional own field |
| lowering | Boundary | `sys.ts:280:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:288:5` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:288:14` | a function returning PollingIntervalQueue |
| refusal_scan | Refused | `sys.ts:289:23` | a cast the runtime can't check |
| lowering | Boundary | `sys.ts:296:5` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:296:61` | a value of type PollingIntervalQueue |
| lowering | Boundary | `sys.ts:308:5` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:308:64` | a value of type PollingIntervalQueue |
| refusal_scan | Refused | `sys.ts:310:19` | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| refusal_scan | Refused | `sys.ts:324:13` | a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read |
| lowering | Boundary | `sys.ts:356:5` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:356:14` | a function returning PollingIntervalQueue |
| lowering | NotYet | `sys.ts:356:14` | a function returning PollingIntervalQueue |
| lowering | Refused | `sys.ts:368:9` | inherited library member push read as an own field |
| lowering | Boundary | `sys.ts:368:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:378:9` | failed statement; state rolled back; continued at next statement or declaration |
| deferred | NotYet | `sys.ts:384:79` | a member typed any ((0)) related to string at the argument site |
| lowering | Boundary | `sys.ts:436:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:436:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:436:25` | a value of type any |
| lowering | NotYet | `sys.ts:436:25` | a value of type any |
| lowering | NotYet | `sys.ts:441:29` | assigning to an Identifier |
| lowering | Boundary | `sys.ts:441:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:441:29` | assigning to an Identifier |
| lowering | Boundary | `sys.ts:441:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:442:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:442:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:442:33` | reading currentModifiedTime |
| lowering | NotYet | `sys.ts:442:33` | reading currentModifiedTime |
| lowering | NotYet | `sys.ts:444:25` | a value of type any |
| lowering | Boundary | `sys.ts:444:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:444:25` | a value of type any |
| lowering | Boundary | `sys.ts:444:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:445:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:445:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:445:54` | reading currentModifiedTime |
| lowering | NotYet | `sys.ts:445:54` | reading currentModifiedTime |
| lowering | Boundary | `sys.ts:447:30` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:447:30` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:447:34` | reading currentModifiedTime |
| lowering | NotYet | `sys.ts:447:34` | reading currentModifiedTime |
| lowering | Boundary | `sys.ts:450:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:450:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:450:59` | reading currentModifiedTime |
| lowering | NotYet | `sys.ts:450:59` | reading currentModifiedTime |
| lowering | Boundary | `sys.ts:464:1` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:470:9` | a value of type any |
| lowering | NotYet | `sys.ts:483:17` | writing a possibly absent optional own field |
| lowering | Boundary | `sys.ts:483:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:490:9` | assigning to an Identifier |
| lowering | Boundary | `sys.ts:490:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:496:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:496:37` | reading pollScheduled |
| lowering | NotYet | `sys.ts:497:9` | assigning to an Identifier |
| lowering | Boundary | `sys.ts:497:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:519:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:519:9` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | Refused | `sys.ts:521:17` | a cast the runtime can't check |
| lowering | Refused | `sys.ts:521:17` | a cast the runtime can't check |
| lowering | Refused | `sys.ts:521:17` | a cast the runtime can't check |
| lowering | Boundary | `sys.ts:599:1` | function signature or prologue failed; continued at next declaration |
| lowering | Refused | `sys.ts:603:5` | a method in object destructuring |
| refusal_scan | Refused | `sys.ts:627:33` | a cast the runtime can't check |
| lowering | Boundary | `sys.ts:655:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:655:29` | optional record property access |
| lowering | Boundary | `sys.ts:773:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:773:13` | reading timerToUpdateChildWatches |
| lowering | NotYet | `sys.ts:777:9` | assigning to an Identifier |
| lowering | Boundary | `sys.ts:777:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:781:9` | assigning to an Identifier |
| lowering | Boundary | `sys.ts:781:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:786:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:786:17` | reading timerToUpdateChildWatches |
| refusal_scan | Refused | `sys.ts:794:87` | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| lowering | Boundary | `sys.ts:813:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:813:117` | reading timerToUpdateChildWatches |
| lowering | Boundary | `sys.ts:833:5` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `sys.ts:838:13` | a value of type any |
| refusal_scan | Refused | `sys.ts:881:34` | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |
| lowering | Boundary | `sys.ts:941:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:941:13` | optional record property access |
| lowering | Boundary | `sys.ts:997:1` | function signature or prologue failed; continued at next declaration |
| lowering | Refused | `sys.ts:1007:5` | a method in object destructuring |
| lowering | Boundary | `sys.ts:1030:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1030:54` | a boolean \| undefined variable a function value captures |
| lowering | Boundary | `sys.ts:1052:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1052:140` | a boolean \| undefined variable a function value captures |
| lowering | Boundary | `sys.ts:1056:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1061:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1061:84` | reading getModifiedTime |
| lowering | Boundary | `sys.ts:1065:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1065:90` | reading getModifiedTime |
| lowering | Boundary | `sys.ts:1100:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1100:40` | optional record property access |
| lowering | Boundary | `sys.ts:1101:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1103:30` | reading defaultFallbackPolling |
| lowering | Boundary | `sys.ts:1122:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1125:17` | reading fileSystemEntryExists |
| lowering | Boundary | `sys.ts:1172:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1186:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1186:48` | optional record property access |
| refusal_scan | Refused | `sys.ts:1198:13` | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| refusal_scan | Refused | `sys.ts:1214:25` | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| lowering | Boundary | `sys.ts:1253:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1253:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1253:77` | function identity comparison in a program with specialized generic function values |
| lowering | NotYet | `sys.ts:1253:77` | function identity comparison in a program with specialized generic function values |
| lowering | Boundary | `sys.ts:1269:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1269:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1269:89` | a boolean \| undefined variable a function value captures |
| lowering | NotYet | `sys.ts:1269:89` | a boolean \| undefined variable a function value captures |
| lowering | NotYet | `sys.ts:1277:17` | reading presentWatcher |
| lowering | Boundary | `sys.ts:1277:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1277:17` | reading presentWatcher |
| lowering | Boundary | `sys.ts:1277:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1281:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1281:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1281:24` | reading presentWatcher |
| lowering | NotYet | `sys.ts:1281:24` | reading presentWatcher |
| refusal_scan | Refused | `sys.ts:1387:31` | a method read as a value (writeFile would lose its object, and this with it) |
| refusal_scan | Refused | `sys.ts:1388:5` | a method read as a value (writeFile would lose its object, and this with it) |
| lowering | NotYet | `sys.ts:1388:5` | replacing a represented method at runtime |
| lowering | Boundary | `sys.ts:1388:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1494:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1494:9` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1494:35` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | NotYet | `sys.ts:1494:35` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | NotYet | `sys.ts:1494:35` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| refusal_scan | NotYet | `sys.ts:1494:92` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| refusal_scan | NotYet | `sys.ts:1494:121` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | Boundary | `sys.ts:1498:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1498:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1519:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1519:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1519:15` | a function value that captures the variable its own initializer declares |
| lowering | NotYet | `sys.ts:1519:15` | a function value that captures the variable its own initializer declares |
| refusal_scan | Refused | `sys.ts:1534:13` | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| refusal_scan | Refused | `sys.ts:1535:13` | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| lowering | Boundary | `sys.ts:1542:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1542:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1542:65` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1542:65` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1545:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1545:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1546:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1546:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1549:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1549:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1549:56` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1549:56` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1551:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1551:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1588:42` | a union of differently held members variable a function value captures |
| lowering | NotYet | `sys.ts:1588:42` | a union of differently held members variable a function value captures |
| lowering | NotYet | `sys.ts:1590:191` | process.recordreplay as a value (use its supported process operation directly) |
| lowering | NotYet | `sys.ts:1590:191` | process.recordreplay as a value (use its supported process operation directly) |
| refusal_scan | NotYet | `sys.ts:1590:192` | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| refusal_scan | Refused | `sys.ts:1612:22` | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| refusal_scan | Refused | `sys.ts:1614:77` | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean |
| lowering | Boundary | `sys.ts:1642:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1642:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1642:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1642:17` | a union of differently held members variable a function value captures |
| lowering | NotYet | `sys.ts:1642:17` | a union of differently held members variable a function value captures |
| lowering | NotYet | `sys.ts:1642:17` | a union of differently held members variable a function value captures |
| lowering | Boundary | `sys.ts:1642:32` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1642:32` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1642:32` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1643:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1643:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1643:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1644:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1644:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1644:17` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1647:32` | node:inspector.Session |
| refusal_scan | NotYet | `sys.ts:1647:42` | node:inspector.Session |
| lowering | Boundary | `sys.ts:1651:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1651:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1651:13` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1651:29` | node:inspector.Session |
| lowering | NotYet | `sys.ts:1651:29` | new a PropertyAccessExpression |
| lowering | NotYet | `sys.ts:1651:29` | new a PropertyAccessExpression |
| lowering | NotYet | `sys.ts:1651:29` | new a PropertyAccessExpression |
| refusal_scan | NotYet | `sys.ts:1651:33` | node:inspector.Session |
| refusal_scan | NotYet | `sys.ts:1651:43` | node:inspector.Session |
| refusal_scan | NotYet | `sys.ts:1652:13` | node:inspector.Session.connect |
| lowering | NotYet | `sys.ts:1652:13` | a method of a class stage 0 doesn't have |
| lowering | Boundary | `sys.ts:1652:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1652:13` | a method of a class stage 0 doesn't have |
| lowering | Boundary | `sys.ts:1652:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1652:13` | a method of a class stage 0 doesn't have |
| lowering | Boundary | `sys.ts:1652:13` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1652:21` | node:inspector.Session.connect |
| refusal_scan | NotYet | `sys.ts:1654:13` | node:inspector.Session.post |
| refusal_scan | NotYet | `sys.ts:1654:13` | node:inspector.Session.post |
| lowering | NotYet | `sys.ts:1654:13` | reading session |
| lowering | Boundary | `sys.ts:1654:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1654:13` | reading session |
| lowering | Boundary | `sys.ts:1654:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1654:13` | reading session |
| lowering | Boundary | `sys.ts:1654:13` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1654:21` | node:inspector.Session.post |
| refusal_scan | NotYet | `sys.ts:1655:17` | node:inspector.Session.post |
| refusal_scan | NotYet | `sys.ts:1655:25` | node:inspector.Session.post |
| lowering | Boundary | `sys.ts:1690:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1690:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1690:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1690:17` | a union of differently held members variable a function value captures |
| lowering | NotYet | `sys.ts:1690:17` | a union of differently held members variable a function value captures |
| lowering | NotYet | `sys.ts:1690:17` | a union of differently held members variable a function value captures |
| lowering | Boundary | `sys.ts:1690:64` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1690:64` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1690:64` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1691:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1691:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1691:17` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1692:17` | node:inspector.Session.post |
| lowering | Boundary | `sys.ts:1692:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1692:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1692:17` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1692:31` | node:inspector.Session.post |
| lowering | Boundary | `sys.ts:1693:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1693:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1693:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1693:31` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1693:31` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1693:31` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1694:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1694:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1694:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1694:67` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1694:67` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1694:67` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1695:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1695:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1695:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1695:70` | toISOString: Date methods other than getTime/valueOf in the fs file host |
| lowering | NotYet | `sys.ts:1695:70` | toISOString: Date methods other than getTime/valueOf in the fs file host |
| lowering | NotYet | `sys.ts:1695:70` | toISOString: Date methods other than getTime/valueOf in the fs file host |
| lowering | Boundary | `sys.ts:1697:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1697:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1697:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1698:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1698:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1698:29` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1703:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1703:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1703:25` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1705:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1705:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1705:21` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1706:21` | node:inspector.Session.disconnect |
| lowering | NotYet | `sys.ts:1706:21` | a method of a class stage 0 doesn't have |
| lowering | Boundary | `sys.ts:1706:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1706:21` | a method of a class stage 0 doesn't have |
| lowering | Boundary | `sys.ts:1706:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1706:21` | a method of a class stage 0 doesn't have |
| lowering | Boundary | `sys.ts:1706:21` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1706:23` | node:inspector.Session.disconnect |
| lowering | Boundary | `sys.ts:1707:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1707:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1707:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1709:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1709:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1709:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1710:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1710:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1710:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1712:18` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1712:18` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1712:18` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1713:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1713:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1713:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1714:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1714:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1714:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1724:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1724:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1724:13` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `sys.ts:1724:41` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | NotYet | `sys.ts:1724:41` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | NotYet | `sys.ts:1724:41` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | NotYet | `sys.ts:1724:41` | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| lowering | Boundary | `sys.ts:1818:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1818:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `sys.ts:1818:24` | reading byteOrderMarkIndicator |
| lowering | NotYet | `sys.ts:1818:24` | reading byteOrderMarkIndicator |
| lowering | Refused | `sys.ts:1869:17` | sort without a comparator |
| lowering | Boundary | `sys.ts:1869:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `sys.ts:1869:17` | sort without a comparator |
| lowering | Boundary | `sys.ts:1869:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `sys.ts:1869:17` | sort without a comparator |
| lowering | Boundary | `sys.ts:1869:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `sys.ts:1870:17` | sort without a comparator |
| lowering | Boundary | `sys.ts:1870:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `sys.ts:1870:17` | sort without a comparator |
| lowering | Boundary | `sys.ts:1870:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `sys.ts:1870:17` | sort without a comparator |
| lowering | Boundary | `sys.ts:1870:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `sys.ts:1967:1` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | Refused | `sys.ts:1967:12` | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| lowering | NotYet | `sys.ts:1967:12` | a literal method through a view that erases its receiver |

Additional raw phases are retained exactly. The sys finding with phase deferred is included at its source location with the raw phase unchanged.

## Root-order evidence

Classification: incomplete. The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.

The baseline report.py criteria require attempted-source and readiness-finding evidence from the entry run plus a Node throw. An incomplete entry ledger cannot establish an artifact. Independent stops remain counted. Node probes retain the existing 60-second TypeScript emit bound. Their actual exits and first errors are listed below; tool errors and timeouts do not establish Node throws.

| Root | Independent stop | Node exit | First error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Latent totals

Guarded entry: LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, bound 1780 seconds. Child exit -15; elapsed 1780.218 seconds; stopping reason elapsed bound. Emitted findings: NotYet 0, Refused 0, panics 0. Completed sources 0/81; stream records 1; checker diagnostics 0.

Did not finish inside its bound. Totals cover emitted findings only. Unemitted findings are unknown, so these are not complete-program latent totals. Per-unit progress is recorded separately below.

Ordinary census, 79 counted observations: NotYet 4, Refused 73, panics 0, accepted 2. These are first-stop totals. The ordinary stream carries no latent counters, so ordinary full latent totals are unavailable.

## Method, checkpoints and cleanup

Read the 69f58ebd method and scripts first and reused its recipe. Fetched origin/main was b8d5f493df0227efcb33c6040ba40c906eec73c9. One adapted tree and an ordinary binary built from this SHA without an overlay. Child bounds are 1780 seconds to leave time for process-group cleanup within the 30-minute process cap. Sorted roots run sequentially in groups of 20, 20, 20 and 19. Each emits two observations (158 raw); only the first counts. run.py preserves the per-root isolation and kills and reaps every child process group. Groups run directly in the foreground without the prior outer group timeout.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The source checkout HEAD is the pin above. apply-status.json and build-status.json identify the single apply and ordinary binary build.

Guarded entry uses this SHA's stage3/census/latent/make_overlay.py and stage3/meter/entry_overlay.py, then go build with the resulting overlay. LATENT_ASSERT_NO_OUTPUT guards against usable IR output. The binaries follow stage3/meter/twice-daily.sh build_census: ordinary build, latent overlay, gofmt, entry overlay, gofmt, then separate entry and latent builds. The timing ledger uses LATENT_TIMES, as in the 93e8aff5 run and twice-daily.sh. Binary hashes, Go build metadata, adapted-source hashes and child status files identify measured bytes.

The final branch has one report-only commit directly on b8d5f493df0227efcb33c6040ba40c906eec73c9. No checkpoint branches were pushed. Each root group was invoked separately. Child logs and status files retain actual exits.

Every census child process group was killed and reaped, including after normal exits. validation.json records observation counts, duplicate agreement, source cleanliness and no surviving census processes. Go evidence files, including archive members, use .go.txt names. Files over 1 MB are gzip-compressed.

### Guarded entry timing progress

2823 completed unit timing records across 21 source files. The source ledger count above is distinct from unit progress. Timing rows do not carry findings and cannot establish complete latent totals.

The previous 93e8aff5 guarded entry pass reached 2590 completed unit timings within the earlier 1800-second bound. These are observed progress counts, not forecasts.

Last emitted timing record: `factory/nodeFactory.ts:3394:5`.

| Source | Unit timing records |
| --- | ---: |
| `/workspace/scratch/next-stops-adapted/src/tsc/tsc.ts` | 6 |
| `corePublic.ts` | 9 |
| `core.ts` | 59 |
| `debug.ts` | 62 |
| `semver.ts` | 28 |
| `performanceCore.ts` | 12 |
| `performance.ts` | 26 |
| `_namespaces/ts.performance.ts` | 1 |
| `hostErrors.ts` | 3 |
| `tracing.ts` | 23 |
| `types.ts` | 876 |
| `sys.ts` | 137 |
| `path.ts` | 77 |
| `diagnosticInformationMap.generated.ts` | 3 |
| `scanner.ts` | 152 |
| `utilitiesPublic.ts` | 243 |
| `utilities.ts` | 867 |
| `factory/baseNodeFactory.ts` | 8 |
| `factory/parenthesizerRules.ts` | 42 |
| `factory/nodeConverters.ts` | 12 |
| `factory/nodeFactory.ts` | 177 |

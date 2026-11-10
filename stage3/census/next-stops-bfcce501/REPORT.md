# Next stops on main bfcce501

Measured for tasks baobab and b5w3ycg. Report only.

Compiler and apply SHA: `bfcce501cc552d9a739e84b19bc4651f548dd3ce`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `cd8904e91268`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-10T20:07:29.053217+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Compared with f1c19c6a | Root names |
| ---: | --- | --- | --- | --- | --- |
| 73 | `tracing.ts:110:22` | NotYet | a namespace value seen as an argument to a callee whose body can't be named | 73 new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 unchanged | `debug.ts` (artifact classification incomplete) |
| 1 | `sys.ts:1654:13` | Refused | a checked view with a callable contract | 1 new | `tracing.ts` |

The baseline had 73 roots stopped at tracing.ts:96:19 and tracing.ts stopped at the primitive brand refusal. The measured groups and exact per-root changes below show their next stops.

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with f1c19c6a

Baseline: f1c19c6a37c7e26299461519c9d19b92015fbbe5, from typescript/next-stops-f1c19c6a. Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `tracing.ts:110:22` NotYet | 73 | 0 | 0 | 73 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |
| `sys.ts:1654:13` Refused | 1 | 0 | 0 | 1 |

Total: 5 unchanged, 0 moved, 74 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `_namespaces/ts.performance.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `_namespaces/ts.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `binder.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `builder.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `builderPublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `builderState.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `builderStatePublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `checker.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `commandLineParser.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `core.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `executeCommandLine.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `expressionToTypeNode.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/baseNodeFactory.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/emitHelpers.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/emitNode.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/nodeChildren.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/nodeConverters.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/nodeFactory.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/nodeTests.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/parenthesizerRules.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/utilities.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `factory/utilitiesPublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `moduleSpecifiers.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `parser.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `path.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `performance.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `performanceCore.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `program.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `programDiagnostics.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `resolutionCache.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `scanner.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `semver.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `sourcemap.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `symbolWalker.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `sys.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `tracing.ts` | new | `types.ts:4560:38` Refused | `sys.ts:1654:13` Refused |
| `transformer.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/classFields.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/classThis.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/declarations/diagnostics.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/declarations.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/destructuring.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2015.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2016.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2017.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2018.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2019.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2020.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/es2021.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/esDecorators.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/esnext.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/generators.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/jsx.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/legacyDecorators.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/module/esnextAnd2015.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/module/impliedNodeFormatDependent.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/module/module.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/module/system.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/namedEvaluation.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/taggedTemplate.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/ts.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/typeSerializer.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `transformers/utilities.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `tsbuild.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `tsbuildPublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `utilitiesPublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `visitorPublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `watch.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `watchPublic.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |
| `watchUtilities.ts` | new | `tracing.ts:96:19` NotYet | `tracing.ts:110:22` NotYet |

## Stop source and lowering emitters

### tracing.ts:110:22

NotYet: `a namespace value seen as an argument to a callee whose body can't be named`. internal/lower/namespace_value.go:362 constructs this reason and :380 selects it for an unnamed callee.

```ts

    /** Stops tracing for the in-progress project and dumps the type catalog. */
    export function stopTracing(): void {
        Debug.assert(tracing, "Tracing is not in progress");
        Debug.assert(!!typeCatalog.length === (mode !== "server")); // Have a type catalog iff not in server mode

        fs.writeSync(traceFd, `\n]\n`);
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:284 and :335 emit this NotYet reason.

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

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:284 and :335 emit this NotYet reason.

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

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. internal/lower/namespaces.go:286 and :338 emit this NotYet reason.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

### sys.ts:1654:13

Refused: `a checked view with a callable contract`. internal/lower/interface_cast.go:272 and internal/lower/view_contracts.go:57 emit this Refused reason.

```ts
            const session = new inspector.Session();
            session.connect();

            session.post("Profiler.enable", () => {
                session.post("Profiler.start", () => {
                    activeSession = session;
                    profilePath = path;
```

## debug.ts single-file full-mode census

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/debug.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 48.861 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
22 records: 22 lowering. Kinds: 11 Boundary, 5 NotYet, 6 Refused. Refusal-scan findings: 0.

| Phase | Kind | Location | Exact reason |
| --- | --- | --- | --- |
| lowering | Boundary | `core.ts:1049:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `core.ts:1049:82` | a comparator that isn't a function |
| lowering | Boundary | `debug.ts:252:5` | function signature or prologue failed; continued at next declaration |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | Refused | `debug.ts:257:86` | overload 1 of assertEachNode parameter test cannot be served by implementation parameter test |
| lowering | Refused | `debug.ts:257:86` | overload 1 of assertEachNode parameter test cannot be served by implementation parameter test |
| lowering | Boundary | `debug.ts:261:5` | function signature or prologue failed; continued at next declaration |
| lowering | Boundary | `debug.ts:262:5` | function signature or prologue failed; continued at next declaration |
| lowering | Refused | `debug.ts:273:82` | overload 1 of assertNode parameter test cannot be served by implementation parameter test |
| lowering | Refused | `debug.ts:273:82` | overload 1 of assertNode parameter test cannot be served by implementation parameter test |
| lowering | Boundary | `debug.ts:274:5` | function signature or prologue failed; continued at next declaration |
| lowering | Boundary | `debug.ts:275:5` | function signature or prologue failed; continued at next declaration |
| lowering | Refused | `debug.ts:299:78` | overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test |
| lowering | Refused | `debug.ts:299:78` | overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test |
| lowering | Boundary | `debug.ts:301:5` | function signature or prologue failed; continued at next declaration |
| lowering | Boundary | `debug.ts:302:5` | function signature or prologue failed; continued at next declaration |
| lowering | Boundary | `debug.ts:350:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `debug.ts:350:135` | a template interpolating an object, an array, a map, a function or undefined |
| lowering | Boundary | `debug.ts:400:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `debug.ts:400:27` | a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls) |
| lowering | Boundary | `debug.ts:401:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `debug.ts:401:24` | reading value |

## semver.ts single-file full-mode census

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/semver.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 44.987 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
57 records: 57 lowering. Kinds: 27 NotYet, 29 Boundary, 1 Refused. Refusal-scan findings: 0.

| Phase | Kind | Location | Exact reason |
| --- | --- | --- | --- |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target Comparator[][] |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| lowering | NotYet | `debug.ts:227:5` | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `debug.ts:235:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | NotYet | `debug.ts:252:33` | a value of type never |
| lowering | Boundary | `semver.ts:46:1` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `semver.ts:47:41` | a constructor object escaping before static fields are initialized |
| lowering | Boundary | `semver.ts:118:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:118:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:118:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:118:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:118:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:118:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:123:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:123:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:123:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:123:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:123:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:123:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `semver.ts:124:13` | a destructured name that isn't plain |
| lowering | NotYet | `semver.ts:124:13` | a destructured name that isn't plain |
| lowering | NotYet | `semver.ts:124:13` | a destructured name that isn't plain |
| lowering | NotYet | `semver.ts:124:13` | a destructured name that isn't plain |
| lowering | NotYet | `semver.ts:124:13` | a destructured name that isn't plain |
| lowering | NotYet | `semver.ts:124:13` | a destructured name that isn't plain |
| lowering | Boundary | `semver.ts:130:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:130:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:130:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:130:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:130:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:130:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `semver.ts:130:28` | reading major |
| lowering | NotYet | `semver.ts:130:28` | reading major |
| lowering | NotYet | `semver.ts:130:28` | reading major |
| lowering | NotYet | `semver.ts:130:28` | reading major |
| lowering | NotYet | `semver.ts:130:28` | reading major |
| lowering | NotYet | `semver.ts:130:28` | reading major |
| lowering | Boundary | `semver.ts:342:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `semver.ts:380:13` | a case whose type differs from the switch's |
| lowering | Boundary | `semver.ts:439:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `semver.ts:452:5` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `semver.ts:452:37` | a template interpolating an object, an array, a map, a function or undefined |

## tracing.ts single-file full-mode census

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/tracing.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 42.502 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
61 records: 52 lowering, 9 refusal_scan. Kinds: 23 NotYet, 26 Boundary, 12 Refused. Refusal-scan findings: 9.

| Phase | Kind | Location | Exact reason |
| --- | --- | --- | --- |
| lowering | NotYet | `checker.ts:15216:13` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:17250:9` | writing a checked object field without its source-slot type certificate |
| lowering | NotYet | `checker.ts:19176:9` | writing a checked object field without its source-slot type certificate |
| lowering | Boundary | `tracing.ts:64:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:64:22` | a call returning any |
| lowering | NotYet | `tracing.ts:72:9` | assigning a field of a value |
| lowering | Boundary | `tracing.ts:72:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:96:9` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | NotYet | `tracing.ts:96:19` | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| lowering | NotYet | `tracing.ts:96:19` | reading tracingEnabled |
| lowering | Boundary | `tracing.ts:100:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Refused | `tracing.ts:103:67` | a spread after the first field |
| lowering | NotYet | `tracing.ts:164:9` | assigning a field of a value |
| lowering | Boundary | `tracing.ts:164:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:171:9` | assigning a field of a value |
| lowering | Boundary | `tracing.ts:171:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:183:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:183:44` | a spread into record storage from a fixed object |
| lowering | Boundary | `tracing.ts:194:19` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:194:67` | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| refusal_scan | Refused | `tracing.ts:206:67` | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| refusal_scan | Refused | `tracing.ts:207:65` | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| lowering | Boundary | `tracing.ts:224:9` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:224:38` | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions |
| lowering | Boundary | `tracing.ts:232:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:232:19` | a value of type any |
| refusal_scan | Refused | `tracing.ts:232:34` | a cast the runtime can't check |
| lowering | Boundary | `tracing.ts:237:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:237:18` | reading objectFlags |
| lowering | Boundary | `tracing.ts:248:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:249:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:250:46` | reading indexedAccessType |
| lowering | Boundary | `tracing.ts:256:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:256:17` | reading objectFlags |
| lowering | Boundary | `tracing.ts:267:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:268:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:269:43` | reading conditionalType |
| lowering | Boundary | `tracing.ts:278:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:279:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:280:43` | reading substitutionType |
| lowering | Boundary | `tracing.ts:286:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:286:17` | reading objectFlags |
| lowering | Boundary | `tracing.ts:296:13` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:296:17` | reading objectFlags |
| lowering | Boundary | `tracing.ts:309:17` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:309:34` | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions |
| lowering | Boundary | `tracing.ts:311:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | NotYet | `tracing.ts:311:38` | reading recursionIdentityMap |
| lowering | NotYet | `tracing.ts:312:21` | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions |
| lowering | Boundary | `tracing.ts:312:21` | failed statement; state rolled back; continued at next statement or declaration |
| lowering | Boundary | `tracing.ts:316:13` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | Refused | `tracing.ts:318:33` | a cast the runtime can't check |
| lowering | Refused | `tracing.ts:318:33` | a cast the runtime can't check |
| lowering | Boundary | `tracing.ts:338:13` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | Refused | `tracing.ts:338:14` | a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| lowering | Refused | `tracing.ts:338:14` | a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| lowering | Boundary | `tracing.ts:357:9` | failed statement; state rolled back; continued at next statement or declaration |
| refusal_scan | Refused | `tracing.ts:357:10` | a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| lowering | Refused | `tracing.ts:357:10` | a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| refusal_scan | Refused | `tracing.ts:378:1` | an ambient value declaration |
| refusal_scan | Refused | `tracing.ts:379:1` | an ambient value declaration |

## Root-order evidence

Classification: incomplete. The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.

The baseline report.py criteria require attempted-source and readiness-finding evidence from the entry run plus a Node throw. An incomplete entry ledger cannot establish an artifact. Independent stops remain counted. Node probes retain the existing 60-second TypeScript emit bound. Their actual exits and first errors are listed below; tool errors and timeouts do not establish Node throws.

| Root | Independent stop | Node exit | First error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | 1 | TypeError: Cannot read properties of undefined (reading 'assert') | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | 1 | TypeError: Cannot read properties of undefined (reading 'Error') | artifact classification incomplete |

## Latent totals

Guarded entry: LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, bound 1800 seconds. Child exit -15; elapsed 1800.537 seconds; stopping reason elapsed bound. Emitted findings: NotYet 0, Refused 0, panics 0. Completed sources 0/81; stream records 1; checker diagnostics 0.

Did not finish inside its bound. Totals cover emitted findings only. Unemitted findings are unknown, so these are not complete-program latent totals. Per-unit progress is recorded separately below.

Ordinary census, 79 counted observations: NotYet 76, Refused 1, panics 0, accepted 2. These are first-stop totals. The ordinary stream carries no latent counters, so ordinary full latent totals are unavailable.

## Method, checkpoints and cleanup

Read the f1c19c6a method and scripts first and reused its recipe at bfcce501. One adapted tree and an ordinary binary built from this SHA without an overlay. Sorted roots run sequentially in groups of 20, 20, 20 and 19. Each emits two observations (158 raw); only the first counts. run.py retains the prior per-root bounds and kills and reaps every child process group. Groups run directly in the foreground without the prior outer group timeout.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The source checkout HEAD is the pin above. The first setup attempt exited 1 after an npm 503; setup-retry-status.json records the successful retry. apply-status.json and build-status.json identify the single apply and ordinary binary build.

Guarded entry uses this SHA's stage3/census/latent/make_overlay.py and stage3/meter/entry_overlay.py, then go build with the resulting overlay. LATENT_ASSERT_NO_OUTPUT guards against usable IR output. The timing ledger uses LATENT_TIMES, as in the f1c19c6a run and twice-daily.sh. Binary hashes, Go build metadata, adapted-source hashes and child status files identify measured bytes.

A plain push to typescript/next-stops-bfcce501-checkpoint followed apply immediately, then each root group. checkpoint-push-*-status.json records each child exit. The final branch has one report-only commit directly on bfcce501. No force-push and no push to main.

Every census child process group was killed and reaped, including after normal exits. validation.json records observation counts, duplicate agreement, source cleanliness and no surviving census processes. Files over 1 MB are gzip-compressed.

### Guarded entry timing progress

2920 completed unit timing records across 21 source files. The source ledger count above is distinct from unit progress. Timing rows do not carry findings and cannot establish complete latent totals.

Last emitted timing record: `factory/nodeFactory.ts:4649:5`.

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
| `types.ts` | 875 |
| `sys.ts` | 137 |
| `path.ts` | 77 |
| `diagnosticInformationMap.generated.ts` | 3 |
| `scanner.ts` | 150 |
| `utilitiesPublic.ts` | 243 |
| `utilities.ts` | 867 |
| `factory/baseNodeFactory.ts` | 8 |
| `factory/parenthesizerRules.ts` | 42 |
| `factory/nodeConverters.ts` | 12 |
| `factory/nodeFactory.ts` | 277 |

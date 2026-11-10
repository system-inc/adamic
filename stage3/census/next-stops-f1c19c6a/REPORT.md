# Next stops on main f1c19c6a

Measured for tasks baobab and b5w3ycg. Report only.

Compiler and apply SHA: `f1c19c6a37c7e26299461519c9d19b92015fbbe5`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `34d72651924b`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-10T11:15:35.948256+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Compared with 7910ab94 | Root names |
| ---: | --- | --- | --- | --- | --- |
| 73 | `tracing.ts:96:19` | NotYet | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 73 new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | 1 unchanged | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 1 unchanged | `debug.ts` (artifact classification incomplete) |
| 1 | `types.ts:4560:38` | Refused | a primitive brand member _isResolvedConfigFileName whose type is not void | 1 new | `tracing.ts` |

Of the 74 roots formerly stopped at debug.ts:525:37, 73 now stop at tracing.ts:96:19. The tracing.ts root instead stops at types.ts:4560:38. None first stops at semver.ts:124:13; that destructuring finding appears in the separate semver full-mode list below. Exact per-root changes follow.

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with 7910ab94

Baseline: 7910ab94a42b0c7f62a79329d7f250b1b2045d05, from typescript/next-stops-7910ab94. Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `tracing.ts:96:19` NotYet | 73 | 0 | 0 | 73 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |
| `types.ts:4560:38` Refused | 1 | 0 | 0 | 1 |

Total: 5 unchanged, 0 moved, 74 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `_namespaces/ts.performance.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `_namespaces/ts.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `binder.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `builder.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `builderPublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `builderState.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `builderStatePublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `checker.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `commandLineParser.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `core.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `executeCommandLine.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `expressionToTypeNode.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/baseNodeFactory.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/emitHelpers.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/emitNode.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/nodeChildren.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/nodeConverters.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/nodeFactory.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/nodeTests.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/parenthesizerRules.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/utilities.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `factory/utilitiesPublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `moduleSpecifiers.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `parser.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `path.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `performance.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `performanceCore.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `program.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `programDiagnostics.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `resolutionCache.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `scanner.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `semver.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `sourcemap.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `symbolWalker.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `sys.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `tracing.ts` | new | `debug.ts:525:37` Refused | `types.ts:4560:38` Refused |
| `transformer.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/classFields.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/classThis.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/declarations/diagnostics.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/declarations.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/destructuring.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2015.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2016.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2017.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2018.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2019.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2020.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/es2021.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/esDecorators.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/esnext.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/generators.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/jsx.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/legacyDecorators.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/module/esnextAnd2015.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/module/impliedNodeFormatDependent.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/module/module.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/module/system.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/namedEvaluation.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/taggedTemplate.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/ts.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/typeSerializer.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `transformers/utilities.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `tsbuild.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `tsbuildPublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `utilitiesPublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `visitorPublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `watch.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `watchPublic.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |
| `watchUtilities.ts` | new | `debug.ts:525:37` Refused | `tracing.ts:96:19` NotYet |

## Stop source and lowering emitters

### tracing.ts:96:19

NotYet: `a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports`. internal/lower/namespaces.go:233 emits this NotYet reason.

```ts
        });

        traceFd = fs.openSync(tracePath, "w");
        tracing = tracingEnabled; // only when traceFd is properly set

        // Start with a prefix that contains some metadata that the devtools profiler expects (also avoids a warning on import)
        const meta = { cat: "__metadata", ph: "M", ts: 1000 * timestamp(), pid: 1, tid: 1 };
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:267 and :318 emit this NotYet reason.

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

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. internal/lower/namespaces.go:267 and :318 emit this NotYet reason.

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

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. internal/lower/namespaces.go:269 and :321 emit this NotYet reason.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

### types.ts:4560:38

Refused: `a primitive brand member _isResolvedConfigFileName whose type is not void`. internal/lower/phantom_brands.go:40 emits this Refused reason.

```ts
 * specified like "./blah" to an absolute path to an actual
 * tsconfig file, e.g. "/root/blah/tsconfig.json"
 */
export type ResolvedConfigFileName = string & { _isResolvedConfigFileName: never; };

/** @internal */
export interface ResolvedRefAndOutputDts {
```

## debug.ts single-file full-mode census

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/debug.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 50.416 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
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

LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the exact adapted src/compiler/semver.ts path. The meter latent binary loaded the complete compiler directory. Child exit 0; elapsed 57.018 seconds. Findings below are ordered by source filename, line and column; all kinds, phases, repeated findings and boundary records are retained. Locations in reachable callees are kept, even when outside the selected file.
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

## Root-order evidence

Classification: incomplete. The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.

The baseline report.py criteria require attempted-source and readiness-finding evidence from the entry run plus a Node throw. An incomplete entry ledger cannot establish an artifact. Independent stops remain counted. The Node probes used their existing 60-second TypeScript emit bound and timed out there; no Node-throw evidence was established.

| Root | Independent stop | Node exit | First error | Label |
| --- | --- | ---: | --- | --- |
| `debug.ts` | `semver.ts:71:9` | None | TypeScript emit failed or timed out | artifact classification incomplete |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | None | TypeScript emit failed or timed out | artifact classification incomplete |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | None | TypeScript emit failed or timed out | artifact classification incomplete |

## Latent totals

Guarded entry: LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, bound 1800 seconds. Child exit -15; elapsed 1800.274 seconds; stopping reason elapsed bound. Emitted findings: NotYet 0, Refused 0, panics 0. Completed sources 0/81; stream records 1; checker diagnostics 0.

Did not finish inside its bound. Totals cover emitted findings only. Unemitted findings are unknown, so these are not complete-program latent totals. Per-unit progress is recorded separately below.

Ordinary census, 79 counted observations: NotYet 76, Refused 1, panics 0, accepted 2. These are first-stop totals. The ordinary stream carries no latent counters, so ordinary full latent totals are unavailable.

## Method, checkpoints and cleanup

Read the 7910ab94 method and scripts first and reused its recipe at f1c19c6a. One adapted tree and an ordinary binary built from this SHA without an overlay. Sorted roots run sequentially in groups of 20, 20, 20 and 19. Each emits two observations (158 raw); only the first counts. run.py retains the prior per-root bounds and kills and reaps every child process group. Groups run directly in the foreground without the prior outer group timeout.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The source checkout HEAD is the pin above. apply-status.json and build-status.json identify the single apply and ordinary binary build.

Guarded entry uses this SHA's stage3/census/latent/make_overlay.py and stage3/meter/entry_overlay.py, then go build with the resulting overlay. LATENT_ASSERT_NO_OUTPUT guards against usable IR output. The timing ledger uses LATENT_TIMES, the variable read by this SHA and set by twice-daily.sh; the prior entry script set ADAMIC_LATENT_TIMES, which emitted no ledger. This changes timing instrumentation only. Binary hashes, Go build metadata, adapted-source hashes and child status files identify measured bytes.

A plain push to typescript/next-stops-f1c19c6a-checkpoint followed apply immediately, then each root group. checkpoint-push-*-status.json records each child exit. The final branch has one report-only commit directly on f1c19c6a. No force-push and no push to main.

Every census child process group was killed and reaped, including after normal exits. validation.json records observation counts, duplicate agreement, source cleanliness and no surviving census processes. Files over 1 MB are gzip-compressed.

### Guarded entry timing progress

2896 completed unit timing records across 21 source files. The source ledger count above is distinct from unit progress. Timing rows do not carry findings and cannot establish complete latent totals.

Last emitted timing record: `factory/nodeFactory.ts:4251:5`.

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
| `factory/nodeFactory.ts` | 253 |

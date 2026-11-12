# Next stops on main fb1c8d3b

Measured per-root next stops. Report only.

Compiler and apply SHA: `fb1c8d3b72de2686c486cf97685079589308dafd`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `54429f0b8487`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-12T06:52:30.767601+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Phase | Compared with b8d5f493 | Root names |
| ---: | --- | --- | --- | --- | --- | --- |
| 73 | `sys.ts:1654:13` | NotYet | node:inspector.Session.post | refusal_scan | 1 unchanged, 72 moved | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | accepted | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | 1 unchanged | `diagnosticInformationMap.generated.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | 1 unchanged | `types.ts` |
| 1 | `path.ts:513:45` | Refused | a cast the runtime can't check | refusal_scan | 1 new | `sys.ts` |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | 1 unchanged | `debug.ts` |

The baseline had 73 roots stopped at tracing.ts:381:1, the ambient declare const require. Their measured destinations are:

| Former roots | Location | Kind | Exact reason | Phase | Root names |
| ---: | --- | --- | --- | --- | --- |
| 72 | `sys.ts:1654:13` | NotYet | node:inspector.Session.post | refusal_scan | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 1 | `path.ts:513:45` | Refused | a cast the runtime can't check | refusal_scan | `sys.ts` |

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with b8d5f493

Baseline: b8d5f493df0227efcb33c6040ba40c906eec73c9, from typescript/next-stops-b8d5f493 (09998fc3d1fb2b1d18cd88242506bc12b2d105f4). Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `sys.ts:1654:13` NotYet | 73 | 1 | 72 | 0 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `path.ts:513:45` Refused | 1 | 0 | 0 | 1 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |

Total: 6 unchanged, 72 moved, 1 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `_namespaces/ts.performance.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `_namespaces/ts.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `binder.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `builder.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `builderPublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `builderState.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `builderStatePublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `checker.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `commandLineParser.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `core.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `executeCommandLine.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `expressionToTypeNode.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/baseNodeFactory.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/emitHelpers.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/emitNode.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/nodeChildren.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/nodeConverters.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/nodeFactory.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/nodeTests.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/parenthesizerRules.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/utilities.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `factory/utilitiesPublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `moduleSpecifiers.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `parser.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `path.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `performance.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `performanceCore.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `program.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `programDiagnostics.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `resolutionCache.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `scanner.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `semver.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `sourcemap.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `symbolWalker.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `sys.ts` | new | `tracing.ts:381:1` Refused | `path.ts:513:45` Refused |
| `tracing.ts` | unchanged | `sys.ts:1654:13` NotYet | `sys.ts:1654:13` NotYet |
| `transformer.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/classFields.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/classThis.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/declarations/diagnostics.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/declarations.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/destructuring.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2015.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2016.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2017.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2018.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2019.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2020.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/es2021.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/esDecorators.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/esnext.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/generators.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/jsx.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/legacyDecorators.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/module/esnextAnd2015.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/module/impliedNodeFormatDependent.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/module/module.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/module/system.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/namedEvaluation.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/taggedTemplate.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/ts.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/typeSerializer.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `transformers/utilities.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `tsbuild.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `tsbuildPublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `utilitiesPublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `visitorPublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `watch.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `watchPublic.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |
| `watchUtilities.ts` | moved | `tracing.ts:381:1` Refused | `sys.ts:1654:13` NotYet |

## Stop source and lowering emitters

### sys.ts:1654:13

NotYet: `node:inspector.Session.post`. Phase: refusal_scan. lower.go:137-141 calls refuse before declareModule. refusals.go:63 calls mergedFieldRulings; merged_fields.go:84-103 walks merged reads, and :134 calls nodeLibraryRefusal. library_node.go:638-640 emits the Node member NotYet. This stop occurs in the refusal scan despite its NotYet kind.

```ts
            const session = new inspector.Session();
            session.connect();

            session.post("Profiler.enable", () => {
                session.post("Profiler.start", () => {
                    activeSession = session;
                    profilePath = path;
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. Phase: lowering (initialization preflight). lower.go:128 invokes enumInitialization before the module refusal scan. namespaces.go emits the enum-initialization NotYet.

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

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. Phase: lowering (initialization preflight). lower.go:128 invokes enumInitialization before the module refusal scan. namespaces.go emits the enum-initialization NotYet.

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### path.ts:513:45

Refused: `a cast the runtime can't check`. Phase: refusal_scan. lower.go:137-141 calls refuse before declareModule and moduleStatements. refusals.go:103-110 runs the registered per-node refusal rules. cast_proof.go:155-174 registers castRefusal as rule 1900; :161 calls castProof, whose :195 constructs this Refused reason for the empty-string cast to T. This stop occurs in the refusal scan.

```ts
 * @internal
 */
export function getPathFromPathComponents<T extends string>(pathComponents: readonly T[], length?: number): T {
    if (pathComponents.length === 0) return "" as T;

    const root = pathComponents[0] && ensureTrailingDirectorySeparator(pathComponents[0]);
    return root + pathComponents.slice(1, length).join(directorySeparator) as T;
```

### semver.ts:71:9

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. Phase: lowering (initialization preflight). lower.go:133 invokes namespaceInitialization before the module refusal scan. namespaces.go emits the namespace-initialization NotYet.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

## Method, checkpoints and cleanup

Read and reused the per-root first-stop census from 09998fc3. The ordinary census tool, source pin, sorted 79-root set, adapted-tree path, and duplicate-observation agreement check follow that recipe. Only compiler/apply revision and the comparison baseline changed. This report measures ordinary first stops; it does not claim latent findings are ordinary stops.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The single apply command is bash stage3/apply.sh /workspace/scratch/next-stops-adapted. The ordinary binary is built without an overlay using go build -o /workspace/scratch/next-stops-census ./stage3/census/tool. Exact commands, actual child exit codes, elapsed times, and process-group cleanup are retained in status JSON files.

Each root is measured in its own child, with a 1780-second bound and cleanup allowance. Root scheduling is split into bounded pieces to keep every process below 30 minutes. Each root emits two observations (158 raw); only the first counts. Binary hashes, Go build metadata (ordinary-buildinfo.txt.gz), adapted-source hashes, source excerpts and Go evidence (.go.txt) identify the measured bytes. Raw logs are preserved losslessly as .log.gz; compression.json records their original and compressed hashes.

One report-only commit is based directly on the measured SHA; only this directory is committed. Validation checks all 79 roots, duplicate agreement, exact baseline classification, clean tracked source, and no surviving census children.

# Next stops on main 3007955b

Measured per-root next stops. Report only.

Compiler and apply SHA: `3007955b6db96ae5f65cedb2e680d77ad1b88664`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `5b1ba3b21e83`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-12T18:29:17.355959+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Phase | Owner (method) | Compared with fb1c8d3b | Root names |
| ---: | --- | --- | --- | --- | --- | --- | --- |
| 73 | `sys.ts:1692:17` | NotYet | node:inspector.Session.post | refusal_scan | internal/lower/library_node.go: nodeLibraryRefusal (code) | 73 new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | accepted | — (unchanged) | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | — (unchanged) | 1 unchanged | `diagnosticInformationMap.generated.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | — (unchanged) | 1 unchanged | `types.ts` |
| 1 | `path.ts:1105:28` | Refused | a cast the runtime can't check | refusal_scan | 01a11410 (owners.json exact reason) | 1 new | `sys.ts` |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | — (unchanged) | 1 unchanged | `debug.ts` |

The baseline had 73 roots stopped at sys.ts:1654:13, node:inspector.Session.post. Their measured destinations are:

| Former roots | Location | Kind | Exact reason | Phase | Owner (method) | Root names |
| ---: | --- | --- | --- | --- | --- | --- |
| 73 | `sys.ts:1692:17` | NotYet | node:inspector.Session.post | refusal_scan | internal/lower/library_node.go: nodeLibraryRefusal (code) | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with fb1c8d3b

Baseline: fb1c8d3b72de2686c486cf97685079589308dafd, from typescript/next-stops-fb1c8d3b (4a94d305617d2edfaaac0385fa9fa6d5b4cbd0b2). Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `sys.ts:1692:17` NotYet | 73 | 0 | 0 | 73 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `path.ts:1105:28` Refused | 1 | 0 | 0 | 1 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |

Total: 5 unchanged, 0 moved, 74 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `_namespaces/ts.performance.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `_namespaces/ts.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `binder.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `builder.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `builderPublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `builderState.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `builderStatePublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `checker.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `commandLineParser.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `core.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `executeCommandLine.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `expressionToTypeNode.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/baseNodeFactory.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/emitHelpers.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/emitNode.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/nodeChildren.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/nodeConverters.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/nodeFactory.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/nodeTests.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/parenthesizerRules.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/utilities.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `factory/utilitiesPublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `moduleSpecifiers.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `parser.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `path.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `performance.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `performanceCore.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `program.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `programDiagnostics.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `resolutionCache.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `scanner.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `semver.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `sourcemap.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `symbolWalker.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `sys.ts` | new | `path.ts:513:45` Refused | `path.ts:1105:28` Refused |
| `tracing.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformer.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/classFields.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/classThis.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/declarations/diagnostics.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/declarations.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/destructuring.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2015.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2016.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2017.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2018.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2019.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2020.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/es2021.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/esDecorators.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/esnext.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/generators.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/jsx.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/legacyDecorators.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/module/esnextAnd2015.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/module/impliedNodeFormatDependent.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/module/module.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/module/system.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/namedEvaluation.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/taggedTemplate.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/ts.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/typeSerializer.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `transformers/utilities.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `tsbuild.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `tsbuildPublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `utilitiesPublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `visitorPublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `watch.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `watchPublic.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |
| `watchUtilities.ts` | new | `sys.ts:1654:13` NotYet | `sys.ts:1692:17` NotYet |

## Stop source and lowering emitters

### sys.ts:1692:17

NotYet: `node:inspector.Session.post`. Phase: refusal_scan. lower.go:139-143 calls refuse before declareModule and moduleStatements. refusals.go:63 calls mergedFieldRulings; merged_fields.go:84-103 walks merged reads and :134 calls nodeLibraryRefusal. library_node.go:642-643 emits this NotYet; :648 also registers nodeLibraryRefusal as refusal rule 800. owners.json has no matching reason; owner is the emitting file and function (code method).

```ts
        function disableCPUProfiler(cb: () => void) {
            if (activeSession && activeSession !== "stopping") {
                const s = activeSession;
                activeSession.post("Profiler.stop", (err, { profile }) => {
                    if (!err) {
                        if (statSync(profilePath)?.isDirectory()) {
                            profilePath = _path.join(profilePath, `${(new Date()).toISOString().replace(/:/g, "-")}+P${process.pid}.cpuprofile`);
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. Phase: lowering (initialization preflight). lower.go:131 invokes enumInitialization before the module refusal scan. namespaces.go emits the enum-initialization NotYet.

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

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. Phase: lowering (initialization preflight). lower.go:131 invokes enumInitialization before the module refusal scan. namespaces.go emits the enum-initialization NotYet.

```ts
/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion*/
/** @internal */
export const Diagnostics = {
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
    Identifier_expected: diag(1003, DiagnosticCategory.Error, "Identifier_expected_1003", "Identifier expected.") as DiagnosticMessage,
    _0_expected: diag(1005, DiagnosticCategory.Error, "_0_expected_1005", "'{0}' expected.") as DiagnosticMessage,
    A_file_cannot_have_a_reference_to_itself: diag(1006, DiagnosticCategory.Error, "A_file_cannot_have_a_reference_to_itself_1006", "A file cannot have a reference to itself.") as DiagnosticMessage,
```

### path.ts:1105:28

Refused: `a cast the runtime can't check`. Phase: refusal_scan. lower.go:139-143 calls refuse before declareModule and moduleStatements. refusals.go:103-108 runs registered per-node refusal rules. cast_proof.go:155 castRefusal calls castProof at :161; castProof at :195 constructs this Refused reason. castRefusal is registered at :174 as rule 1900. Owner 01a11410 comes from stage3/meter/owners.json by the exact reason (owners.json method).

```ts
            return result;
        }

        const parentPath = getDirectoryPath(directory) as P;
        if (parentPath === directory) {
            return undefined;
        }
```

### semver.ts:71:9

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. Phase: lowering (initialization preflight). lower.go:136 invokes namespaceInitialization before the module refusal scan. namespaces.go:279 namespaceInitialization emits the namespace-initialization NotYet.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

## Method, checkpoints and cleanup

Read and reused the per-root first-stop census from 4a94d305. The ordinary census tool, source pin, sorted 79-root set, adapted-tree path, and duplicate-observation agreement check follow that recipe. Only compiler/apply revision and the comparison baseline changed. This report measures ordinary first stops; it does not claim latent findings are ordinary stops.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The single apply command is bash stage3/apply.sh /workspace/scratch/next-stops-adapted. The ordinary binary is built without an overlay using go build -o /workspace/scratch/next-stops-census ./stage3/census/tool. Exact commands, actual child exit codes, elapsed times, and process-group cleanup are retained in status JSON files.

Each root is measured in its own child, with a 1780-second bound and cleanup allowance. Root scheduling is split into bounded pieces to keep every process below 30 minutes. Each root emits two observations (158 raw); only the first counts. Binary hashes, Go build metadata (ordinary-buildinfo.txt.gz), adapted-source hashes, source excerpts and Go evidence (.go.txt) identify the measured bytes. Raw logs are preserved losslessly as .log.gz; compression.json records their original and compressed hashes.

One report-only commit is based directly on the measured SHA; only this directory is committed. The first build attempt preceded setup completion and exited 1 because cohere/TypeScript/tsc/go.mod was absent; build-before-setup-status.json retains it. The subsequent build supplies the measured binary. Validation checks all 79 roots, duplicate agreement, exact baseline classification, clean tracked source, and no surviving census children.

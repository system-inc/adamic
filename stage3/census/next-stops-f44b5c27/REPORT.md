# Next stops on main f44b5c27

Measured per-root next stops. Report only.

Compiler and apply SHA: `f44b5c272a39bcbd7cd2fb278939ac1c1832ce2e`. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `ae0a66d56fe7`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. Started 2026-11-13T04:43:44.610037+00:00.

## Grouped first stops

Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.

| Roots | Location | Kind | Exact reason | Phase | Owner (method) | Compared with 3007955b | Root names |
| ---: | --- | --- | --- | --- | --- | --- | --- |
| 73 | `sys.ts:57:1` | Refused | an ambient function declaration | refusal_scan | internal/lower/ambient.go: (*lowering).ambientRefusal (code method) | 73 new | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | accepted | lower.Lower (accepted; no stop) | 2 unchanged | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization (code method) | 1 unchanged | `diagnosticInformationMap.generated.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization (code method) | 1 unchanged | `types.ts` |
| 1 | `path.ts:1105:28` | Refused | a cast the runtime can't check | refusal_scan | 01a11410 (owners.json exact reason) | 1 unchanged | `sys.ts` |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization (code method) | 1 unchanged | `debug.ts` |

The baseline had 73 roots stopped at sys.ts:1692:17, node:inspector.Session.post. Their measured destinations are:

| Former roots | Location | Kind | Exact reason | Phase | Owner (method) | Root names |
| ---: | --- | --- | --- | --- | --- | --- |
| 73 | `sys.ts:57:1` | Refused | an ambient function declaration | refusal_scan | internal/lower/ambient.go: (*lowering).ambientRefusal (code method) | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |

All 73 former inspector roots now stop at the ambient setTimeout declaration at adapted sys.ts:57:1. The expected toISOString / __filename progression is not the ordinary first stop observed here: refusal scanning ends before body lowering reaches either. The compiler-source evidence explains this ordering; no claim about a measured latent getNodeSystem stop is made.

## Roots lowering whole

2/79: `corePublic.ts`, `hostErrors.ts`. Lower acceptance covers each complete resolved root program. Native output is not measured.

## Comparison with 3007955b

Baseline: 3007955b6db96ae5f65cedb2e680d77ad1b88664, from typescript/next-stops-3007955b (933c8981edfa0e046893b1a1f1570c166a9e8e2c). Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.

| Current group | Roots | Unchanged | Moved | New |
| --- | ---: | ---: | ---: | ---: |
| `sys.ts:57:1` Refused | 73 | 0 | 0 | 73 |
| `-` accepted | 2 | 2 | 0 | 0 |
| `commandLineParser.ts:293:35` NotYet | 1 | 1 | 0 | 0 |
| `diagnosticInformationMap.generated.ts:13:45` NotYet | 1 | 1 | 0 | 0 |
| `path.ts:1105:28` Refused | 1 | 1 | 0 | 0 |
| `semver.ts:71:9` NotYet | 1 | 1 | 0 | 0 |

Total: 6 unchanged, 0 moved, 73 new. Exact before and after reasons are in comparison.json.

| Root | Classification | Before location / kind | After location / kind |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `_namespaces/ts.performance.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `_namespaces/ts.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `binder.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `builder.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `builderPublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `builderState.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `builderStatePublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `checker.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `commandLineParser.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `core.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `corePublic.ts` | unchanged | `-` accepted | `-` accepted |
| `debug.ts` | unchanged | `semver.ts:71:9` NotYet | `semver.ts:71:9` NotYet |
| `diagnosticInformationMap.generated.ts` | unchanged | `commandLineParser.ts:293:35` NotYet | `commandLineParser.ts:293:35` NotYet |
| `emitter.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `executeCommandLine.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `expressionToTypeNode.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/baseNodeFactory.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/emitHelpers.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/emitNode.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/nodeChildren.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/nodeConverters.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/nodeFactory.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/nodeTests.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/parenthesizerRules.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/utilities.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `factory/utilitiesPublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `hostErrors.ts` | unchanged | `-` accepted | `-` accepted |
| `moduleNameResolver.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `moduleSpecifiers.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `parser.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `path.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `performance.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `performanceCore.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `program.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `programDiagnostics.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `resolutionCache.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `scanner.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `semver.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `sourcemap.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `symbolWalker.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `sys.ts` | unchanged | `path.ts:1105:28` Refused | `path.ts:1105:28` Refused |
| `tracing.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformer.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/classFields.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/classThis.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/declarations/diagnostics.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/declarations.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/destructuring.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2015.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2016.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2017.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2018.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2019.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2020.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/es2021.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/esDecorators.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/esnext.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/generators.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/jsx.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/legacyDecorators.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/module/esnextAnd2015.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/module/impliedNodeFormatDependent.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/module/module.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/module/system.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/namedEvaluation.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/taggedTemplate.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/ts.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/typeSerializer.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `transformers/utilities.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `tsbuild.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `tsbuildPublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `types.ts` | unchanged | `diagnosticInformationMap.generated.ts:13:45` NotYet | `diagnosticInformationMap.generated.ts:13:45` NotYet |
| `utilities.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `utilitiesPublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `visitorPublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `watch.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `watchPublic.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |
| `watchUtilities.ts` | new | `sys.ts:1692:17` NotYet | `sys.ts:57:1` Refused |

## Stop source and lowering emitters

### sys.ts:57:1

Refused: `an ambient function declaration`. Phase: refusal_scan. No exact reason in owners.json. lower.go:139-143 runs refuse for each module before declareModule and moduleStatements. refusals.go:63 runs mergedFieldRulings first; once the former inspector stop no longer returns there, refusals.go:103-108 visits registered per-node rules. ambient.go:62 ambientRefusal identifies an ambient function at :80-81, returns Refused at :91 and registers rule 2200 at :94. This source is the ambient setTimeout declaration, so ordinary refusal scanning stops before getNodeSystem body lowering. This does not measure its latent next lowering stop.

```ts
import * as _os from "node:os";
import * as _crypto from "node:crypto";

declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;
declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;

/**
```

### commandLineParser.ts:293:35

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. Phase: lowering (initialization preflight). No exact reason in owners.json. lower.go:131 validates enum declarations using enumInitialization; :136 invokes namespaceInitialization before refusal scanning. namespaces.go:279 (*lowering).namespaceInitialization traces reachable namespace/enum reads and emits the enum reason at :287 or namespace reason at :289 (also direct-read checks at :338/:341). modules.go:14-17 documents per-entry depth-first dependency order, skipping cycle back edges. The result is retained for this exact root order.

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

NotYet: `reading an enum before its runtime initialization; move the call after the enum declaration`. Phase: lowering (initialization preflight). No exact reason in owners.json. lower.go:131 validates enum declarations using enumInitialization; :136 invokes namespaceInitialization before refusal scanning. namespaces.go:279 (*lowering).namespaceInitialization traces reachable namespace/enum reads and emits the enum reason at :287 or namespace reason at :289 (also direct-read checks at :338/:341). modules.go:14-17 documents per-entry depth-first dependency order, skipping cycle back edges. The result is retained for this exact root order.

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

Refused: `a cast the runtime can't check`. Phase: refusal_scan. Exact reason maps to 01a11410 in stage3/meter/owners.json. lower.go:139-143 runs refuse before declarations and statements. refusals.go:103-108 dispatches registered per-node rules; cast_proof.go:155 (*lowering).castRefusal invokes castProof at :161, constructs the Refused reason in (*lowering).castProof at :195; registered rule 1900 at :174.

```ts
            return result;
        }

        const parentPath = getDirectoryPath(directory) as P;
        if (parentPath === directory) {
            return undefined;
        }
```

### semver.ts:71:9

NotYet: `a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration`. Phase: lowering (initialization preflight). No exact reason in owners.json. lower.go:131 validates enum declarations using enumInitialization; :136 invokes namespaceInitialization before refusal scanning. namespaces.go:279 (*lowering).namespaceInitialization traces reachable namespace/enum reads and emits the enum reason at :287 or namespace reason at :289 (also direct-read checks at :338/:341). modules.go:14-17 documents per-entry depth-first dependency order, skipping cycle back edges. The result is retained for this exact root order.

```ts
        const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

        Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

        this.major = major;
        this.minor = minor;
```

## Method, checkpoints and cleanup

Read and reused the per-root first-stop census from baseline branch commit 933c8981. The ordinary census tool, source pin, sorted 79-root set, adapted-tree path, and duplicate-observation agreement check follow that recipe. Compiler/apply revision, comparison baseline and fresh-run tooling adaptations are recorded below. This report measures ordinary first stops; it does not claim latent findings are ordinary stops.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The single apply command is bash stage3/apply.sh /workspace/scratch/next-stops-f44b5c27-adapted. The ordinary binary is built without an overlay using go build -o /workspace/scratch/next-stops-f44b5c27-census ./stage3/census/tool. Exact commands, actual child exit codes, elapsed times, and process-group cleanup are retained in status JSON files.

Each root is measured in its own child, with a 1780-second bound and cleanup allowance. Root scheduling is split into bounded pieces to keep every process below 30 minutes. Each root child emits two observations (158 raw); only the first counts. Each observation repeats the full resolved root program, and the two must agree in all non-timing fields. Binary hashes, Go build metadata (ordinary-buildinfo.txt.gz), adapted-source hashes, source excerpts and Go evidence (.go.txt) identify the measured bytes. Raw logs are preserved losslessly as .log.gz; compression.json records their original and compressed hashes.

One report-only commit is based directly on the measured SHA; only this directory is committed. Build and setup history for this run is retained in its own status files; no previous-run failed-build status was imported. Validation checks all 79 roots, duplicate agreement, exact baseline classification, clean tracked source, and no surviving census children.

## sys.ts and root-order artifact roots

These are ordinary per-root observations. A root changes module traversal/initialization order; these three initialization-preflight stops are retained separately from the broad shared sys.ts stop. No alternate-order run or latent overlay was used.

| Root | Location | Kind | Exact reason | Phase | Owner (method) |
| --- | --- | --- | --- | --- | --- |
| `sys.ts` | `path.ts:1105:28` | Refused | a cast the runtime can't check | refusal_scan | 01a11410 (owners.json exact reason) |
| `debug.ts` | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization (code method) |
| `diagnosticInformationMap.generated.ts` | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization (code method) |
| `types.ts` | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization (code method) |

All line/column locations are measured in the completed adapted tree. Upstream line numbers mentioned in the request can differ after adaptations; the retained source excerpts identify the actual expressions.

## Tooling adaptations for this run

- `run.py`, `run-one.py`, `analyze.py`, `source-excerpts.py`, `measure.py`, and `validate.py`: use the fresh worktree and unique adapted-tree, binary, output and report paths so the stop record and prior run remain intact.
- `run.py`: remove the unused all-groups scheduler; invoke one bounded `run-one.py` process per root to keep scheduler processes under 30 minutes as well as census children. Child exit codes and process-group cleanup remain recorded.
- `validate.py`: remove the old failed-build artifact requirement and historical result field. Record every status file's actual integer child exit code; require all root children to exit zero, 79 exact sorted roots, two observations per root agreeing in every non-timing field, verified comparison tuples, binary/source hashes, process cleanup and duration bounds. Require phase and owner evidence for every stopped group, including unchanged groups.
- `write-report.py`: update target and baseline identifiers, the prior inspector location, fresh-run build narrative and measurement paths; add the sys.ts/root-order artifact discussion, this change list and actual checkpoint history.
- `package-evidence.py`: accumulate compression metadata across checkpoint batches and retain SHA256 rather than commit a compressed artifact over 20 MB. Raw logs are packaged only after their child exits.
- `assemble.py` is unchanged. `analyze.py` retains the baseline's exact tuple comparison definitions: unchanged, moved to an existing baseline tuple, or new tuple absent from baseline. `source-excerpts.py` retains its source extraction algorithm. The ordinary Go census source and compiler are unchanged, with no overlays.
- Fresh `provenance.json`, `owners.json`, `stop-phases.json`, binary metadata and source evidence describe this run; previous status files and emitter excerpts were not imported as current evidence.

## Pushed checkpoints

| Group | Roots | Commit |
| --- | --- | --- |
| 1 | 1-20 | `9af8d7ec7686c3ee51a22006f21d5b84c7af6079` |
| 2 | 21-40 | `84820137b7ff325664317b1d9d9c7aad3022f8ac` |
| 3 | 41-60 | `f25787fa9ee872794446ced7d86448ce864489e1` |
| 4 | 61-79 | `f4e41a31e97cbc733347b78764fe9185654cf7da` |

Validation passed: all 79 sorted roots are covered, all 158 raw observations agree in their 79 non-timing pairs, all root child exit codes are zero and every recorded child exit is retained, all child runtimes are below 1800 seconds, and no census child remains. See validation.json and the per-command status files.

The final branch is typescript/next-stops-f44b5c27-2, one new commit directly on the measured SHA. The earlier typescript/next-stops-f44b5c27 branch retains the stop record and was not rewritten.

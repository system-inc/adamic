# CommonJSBundle next stops at ba619e28

## Roots first stopped by the new bind reason

None of the 79 roots first stops with the new bind reason.

## Every root whose first stop differs from the comparison

Comparison: `typescript/next-stops-55c375d9` at `e61f6a1b7a6149fb49fab05a64e4c57c394d47f5`. Complete validated CommonJSBundle results are available for main compiler 55c375d9d245b580c010c3e0b5659f2849f24fb6; no fallback baseline was used.

None. All 79 first-stop tuples are unchanged.

## Grouped first-stop table

| Roots | Site | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:1385:31 | Refused | a method read as a value (writeFile would lose its object, and this with it) | refusal_scan (unbound method) | internal/lower/refusals.go: (*lowering).methodValueRefusal / methodValueStop (patchWriteFileEnsuringDirectory) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | utilitiesPublic.ts:1280:52 | Refused | an unproven predicate argument for parameter predicate (argument "predicate") | refusal_scan (predicate contracts) | internal/lower/predicates.go: (*lowering).predicateArgumentProof (getFirstJSDocTag) |

## Group membership

- `sys.ts:1385:31`, Refused: `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `core.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts`.
- `-`, accepted: `corePublic.ts`, `hostErrors.ts`.
- `commandLineParser.ts:293:35`, NotYet: `diagnosticInformationMap.generated.ts`.
- `diagnosticInformationMap.generated.ts:13:45`, NotYet: `types.ts`.
- `semver.ts:71:9`, NotYet: `debug.ts`.
- `utilitiesPublic.ts:1280:52`, Refused: `sys.ts`.

## All 79 root comparisons

| Root | Before: site; kind; exact reason | After: site; kind; exact reason | Changed |
| --- | --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| _namespaces/ts.performance.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| _namespaces/ts.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| binder.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| builder.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| builderPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| builderState.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| builderStatePublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| checker.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| commandLineParser.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| core.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| corePublic.ts | -; accepted; - | -; accepted; - | no |
| debug.ts | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | no |
| diagnosticInformationMap.generated.ts | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | no |
| emitter.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| executeCommandLine.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| expressionToTypeNode.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/baseNodeFactory.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/emitHelpers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/emitNode.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/nodeChildren.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/nodeConverters.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/nodeFactory.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/nodeTests.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/parenthesizerRules.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| factory/utilitiesPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| hostErrors.ts | -; accepted; - | -; accepted; - | no |
| moduleNameResolver.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| moduleSpecifiers.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| parser.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| path.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| performance.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| performanceCore.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| program.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| programDiagnostics.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| resolutionCache.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| scanner.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| semver.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| sourcemap.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| symbolWalker.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| sys.ts | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | utilitiesPublic.ts:1280:52; Refused; an unproven predicate argument for parameter predicate (argument "predicate") | no |
| tracing.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformer.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/classFields.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/classThis.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/declarations/diagnostics.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/declarations.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/destructuring.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2015.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2016.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2017.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2018.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2019.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2020.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/es2021.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/esDecorators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/esnext.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/generators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/jsx.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/legacyDecorators.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/module/esnextAnd2015.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/module/impliedNodeFormatDependent.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/module/module.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/module/system.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/namedEvaluation.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/taggedTemplate.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/ts.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/typeSerializer.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| transformers/utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| tsbuild.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| tsbuildPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| types.ts | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration | no |
| utilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| utilitiesPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| visitorPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| watch.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| watchPublic.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |
| watchUtilities.ts | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | sys.ts:1385:31; Refused; a method read as a value (writeFile would lose its object, and this with it) | no |

## Source at measured stops

### sys.ts:1385:31

```typescript
    const originalWriteFile = sys.writeFile;
```

### commandLineParser.ts:293:35

```typescript
            fixedpollinginterval: WatchFileKind.FixedPollingInterval,
```

### diagnosticInformationMap.generated.ts:13:45

```typescript
    Unterminated_string_literal: diag(1002, DiagnosticCategory.Error, "Unterminated_string_literal_1002", "Unterminated string literal.") as DiagnosticMessage,
```

### semver.ts:71:9

```typescript
        Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");
```

### utilitiesPublic.ts:1280:52

```typescript
    return find(getJSDocTagsWorker(node, noCache), predicate);
```


## Reach of the new bind check

The new reason is emitted by `checkBoundArguments` in `internal/lower/function_bind.go:501`, called from `internal/lower/lower.go:218`. The dominant unbound-method stop runs in the earlier refusal scan (`internal/lower/lower.go:141-145`). An unchanged first-stop table does not establish whether those roots would encounter the new bind check after the earlier stop is resolved. No adaptations were made to bypass stops.

The adapted compiler input hashes match main for all 82 files under `src/compiler/`; see `main-input-comparison.json`.

## Method and limits

Measured compiler and existing adaptation inputs: `ba619e28774130e24331c3026a938cd6f3cd6a34` (task #04f5s0s). A fresh detached worktree initialized recursive submodules. TypeScript source was cloned at `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3); stage3/api `npm ci` passed. Existing `stage3/apply.sh` prepared the ordinary input tree without changing adaptations. One ordinary census binary was built and reused. CommonJSBundle only; no separate sys.ts single-file census.

All 79 sorted compiler roots ran in separate child processes, with two observations per root equal except timing. Groups contain at most ten roots; three groups run concurrently, each preserving the copied three-root-child concurrency. Each root and group has a 1780-second bound with exit and cleanup recorded. Source hashes and the binary hash were checked. No compiler or adaptation changes were made. This measures first stops and lowering acceptance, not successful native execution.

Paths and sites are relative to `src/compiler/`. Comparisons use exact (site, kind, reason) tuples. Raw paired observations, statuses, compressed logs, source evidence, and the full root comparison are retained.

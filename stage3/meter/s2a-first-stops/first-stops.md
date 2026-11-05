# First stops: main plus S2a

All counts and findings below are measured on **main plus S2a**, not main alone. This is a measurement only. No source changes or latent passes were made.

Main input: `fd30fac46130d994f0cd470652d90dffa84f9518`. S2a input: `d231af0d34a4a145b14139dbc76d74f87e478b98`. Merge result: `d231af0d34a4a145b14139dbc76d74f87e478b98`.

git merge --no-edit origin/main reported Already up to date; main is already an ancestor of S2a, so no new merge commit was created.

Roots: **79**. Lower whole: **2**. Stop at core.ts mapIterator: **0**. Stop at another generator site: **0**. Stop for a reason naming generators, yield, iterators or for-of: **0**.

Census peak RSS: **935,060 KiB** (913.14 MiB). Elapsed: 557.959 seconds. Method: Linux getrusage(RUSAGE_CHILDREN).ru_maxrss for the census process.

Locations are in the adapted tree and relative to `src/compiler`, with one-based line and column. Root names are also relative to `src/compiler`. Generator-site classification uses the pinned TypeScript parser to identify an enclosing generator at the stop location. The reason-filtered list uses generator, yield, iterator or for-of in the reason text.

The unchanged meter census command also emits non-source records and a final aggregate call. Those are retained in census.jsonl and first-stops.json but excluded from the per-root totals.

The next dominant stop is `sameMap` at `core.ts:356:37`: 73 roots stop there. The remaining unchecked-cast refusal is the `core.ts` root stopping at `debug.ts:167:31`. No source root panicked.

The final aggregate row reports `lower: stage 0 compiles a program from one entry file, got 79`, as expected for this unchanged census tool. It does not affect individual-root results.

## Reasons, biggest group first

| Reason | Kind | Root count |
| --- | --- | ---: |
| a cast the runtime can't check | Refused | 74 |
| reading an enum before its runtime initialization; move the call after the enum declaration | NotYet | 2 |
| a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | NotYet | 1 |

### a cast the runtime can't check (74 roots)

| Root | Kind | First stop | Enclosing function |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `_namespaces/ts.performance.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `_namespaces/ts.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `binder.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `builder.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `builderPublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `builderState.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `builderStatePublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `checker.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `commandLineParser.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `core.ts` | Refused | `debug.ts:167:31` | `setAssertionLevel` |
| `emitter.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `executeCommandLine.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `expressionToTypeNode.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/baseNodeFactory.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/emitHelpers.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/emitNode.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/nodeChildren.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/nodeConverters.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/nodeFactory.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/nodeTests.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/parenthesizerRules.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/utilities.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `factory/utilitiesPublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `moduleNameResolver.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `moduleSpecifiers.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `parser.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `path.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `performance.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `performanceCore.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `program.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `programDiagnostics.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `resolutionCache.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `scanner.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `semver.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `sourcemap.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `symbolWalker.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `sys.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `tracing.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformer.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/classFields.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/classThis.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/declarations.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/declarations/diagnostics.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/destructuring.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2015.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2016.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2017.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2018.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2019.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2020.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/es2021.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/esDecorators.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/esnext.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/generators.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/jsx.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/legacyDecorators.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/module/esnextAnd2015.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/module/impliedNodeFormatDependent.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/module/module.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/module/system.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/namedEvaluation.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/taggedTemplate.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/ts.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/typeSerializer.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `transformers/utilities.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `tsbuild.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `tsbuildPublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `utilities.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `utilitiesPublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `visitorPublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `watch.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `watchPublic.ts` | Refused | `core.ts:356:37` | `sameMap` |
| `watchUtilities.ts` | Refused | `core.ts:356:37` | `sameMap` |

### reading an enum before its runtime initialization; move the call after the enum declaration (2 roots)

| Root | Kind | First stop | Enclosing function |
| --- | --- | --- | --- |
| `diagnosticInformationMap.generated.ts` | NotYet | `commandLineParser.ts:293:35` | `` |
| `types.ts` | NotYet | `diagnosticInformationMap.generated.ts:13:45` | `` |

### a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration (1 root)

| Root | Kind | First stop | Enclosing function |
| --- | --- | --- | --- |
| `debug.ts` | NotYet | `semver.ts:71:9` | `<anonymous>` |

## Roots lowering whole

- `corePublic.ts`
- `hostErrors.ts`

## Stops still naming generator or iterator forms

None.

## Stops at generator sites

None.

## Method

Compiler built from the merge result using `go build -o /workspace/scratch/s2a-bin/census ./stage3/census/tool`. Census binary SHA-256: `b287c69758f61965d24612a3fff5fa9dd1b28e20b2a6d04628e6aecd84f6a771`.

Census command: `/workspace/scratch/s2a-bin/census /workspace/scratch/s2a-adapted/src/compiler /workspace/adamic/stage3/meter/s2a-first-stops/census.jsonl`. Every source file is its own root. Ordinary lowering returns its first error. All latent passes and the tsc-entry pass were skipped.

# Promise-discard sites in TypeScript’s compiler

**0 new-reason sites observed in the 79-file census so far: 21/79 files attempted (20 full, 1 scan only). Both controls passed; coverage is incomplete.**

No compiler finding’s reason or text contains `an unawaited async task`. Control findings are excluded from this count.

## Controls

**3210 refusePromiseIntoVoid: passed**, scan-only binary; exit 0, 0.268 seconds.

```ts
async function fails(): Promise<void> {}
const run: () => unknown = fails;
run();
```

Location: `control.ts:2:28`. Named type: `Promise<void>`.

```text
an unawaited async task: a Promise<void> at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task
/workspace/scratch/promise-discard-sites-run/control.ts:2:28: Adamic 0.1 refuses an unawaited async task: a Promise<void> at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task; await the task inside a function that returns void, or keep the Promise in the type it's seen as
```

**3220 refusePromiseDiscard: passed**, scan-only binary; exit 0, 0.116 seconds.

```ts
async function fails(): Promise<void> {}
[0].map(fails);
```

Location: `control-3220.ts:2:1`. Named type: `Promise<void>[]`.

```text
an unawaited async task: a Promise<void>[] thrown away, whose tasks nothing awaits
/workspace/scratch/promise-discard-sites-run/control-3220.ts:2:1: Adamic 0.1 refuses an unawaited async task: a Promise<void>[] thrown away, whose tasks nothing awaits; await the tasks (Promise.all for many), or keep them and return them
```

The 3210 control also passed the original full census. Its scan-only finding matched the full control exactly. Both scan-only controls emitted zero lowering units. Full raw control output remains in sites.json.

## Per-file coverage

Compiler head `e4b65199b13da9dcce678ffe93e0f06fe11bb2e5`. Upstream TypeScript HEAD and peeled v6.0.3 tag both confirmed as `050880ce59e30b356b686bd3144efe24f875ebc8`. apply.sh completed with exactly 79 TypeScript files. Every measurement loads the adapted compiler directory, selects one absolute LATENT_ONLY_FILE, uses CommonJSBundle and output guards, and has a 1500-second process-group kill bound.

The measurement-only wrapper in scan-only.go.txt invokes the existing latentScan helper (stage3/census/latent/lower.go.txt:519), which full mode calls before registration and body lowering (units.go.txt:141). It runs the registered refusal rules without lowering bodies. Both 3210 and 3220 execute in that scan. checker.ts’s original full-mode timeout remains preserved; the table uses its successful scan-only follow-up.

| File | Measurement | Status | Exit | Seconds | Timeout | Total findings | New-reason sites |
|---|---|---|---:|---:|---|---:|---:|
| _namespaces/ts.moduleSpecifiers.ts | full | ran | 0 | 37.439 | False | 0 | 0 |
| _namespaces/ts.performance.ts | full | ran | 0 | 36.445 | False | 0 | 0 |
| _namespaces/ts.ts | full | ran | 0 | 62.66 | False | 0 | 0 |
| binder.ts | full | ran | 0 | 393.778 | False | 504 | 0 |
| builder.ts | full | ran | 0 | 180.526 | False | 399 | 0 |
| builderPublic.ts | full | ran | 0 | 50.686 | False | 3 | 0 |
| builderState.ts | full | ran | 0 | 127.299 | False | 37 | 0 |
| builderStatePublic.ts | full | ran | 0 | 38.806 | False | 0 | 0 |
| checker.ts | scan only | ran | 0 | 206.199 | False | 1473 | 0 |
| commandLineParser.ts | full | ran | 0 | 272.421 | False | 367 | 0 |
| core.ts | full | ran | 0 | 77.278 | False | 14 | 0 |
| corePublic.ts | full | ran | 0 | 32.716 | False | 0 | 0 |
| debug.ts | full | ran | 0 | 84.415 | False | 10 | 0 |
| diagnosticInformationMap.generated.ts | full | ran | 0 | 33.865 | False | 0 | 0 |
| emitter.ts | full | ran | 0 | 1357.596 | False | 767 | 0 |
| executeCommandLine.ts | full | ran | 0 | 192.473 | False | 365 | 0 |
| expressionToTypeNode.ts | full | ran | 0 | 198.716 | False | 183 | 0 |
| factory/baseNodeFactory.ts | full | ran | 0 | 37.64 | False | 10 | 0 |
| factory/emitHelpers.ts | full | ran | 0 | 73.803 | False | 47 | 0 |
| factory/emitNode.ts | full | ran | 0 | 37.801 | False | 26 | 0 |
| factory/nodeChildren.ts | full | ran | 0 | 72.538 | False | 16 | 0 |
| factory/nodeConverters.ts | not measured | not run | | | | unknown | unknown |
| factory/nodeFactory.ts | not measured | not run | | | | unknown | unknown |
| factory/nodeTests.ts | not measured | not run | | | | unknown | unknown |
| factory/parenthesizerRules.ts | not measured | not run | | | | unknown | unknown |
| factory/utilities.ts | not measured | not run | | | | unknown | unknown |
| factory/utilitiesPublic.ts | not measured | not run | | | | unknown | unknown |
| hostErrors.ts | not measured | not run | | | | unknown | unknown |
| moduleNameResolver.ts | not measured | not run | | | | unknown | unknown |
| moduleSpecifiers.ts | not measured | not run | | | | unknown | unknown |
| parser.ts | not measured | not run | | | | unknown | unknown |
| path.ts | not measured | not run | | | | unknown | unknown |
| performance.ts | not measured | not run | | | | unknown | unknown |
| performanceCore.ts | not measured | not run | | | | unknown | unknown |
| program.ts | not measured | not run | | | | unknown | unknown |
| programDiagnostics.ts | not measured | not run | | | | unknown | unknown |
| resolutionCache.ts | not measured | not run | | | | unknown | unknown |
| scanner.ts | not measured | not run | | | | unknown | unknown |
| semver.ts | not measured | not run | | | | unknown | unknown |
| sourcemap.ts | not measured | not run | | | | unknown | unknown |
| symbolWalker.ts | not measured | not run | | | | unknown | unknown |
| sys.ts | not measured | not run | | | | unknown | unknown |
| tracing.ts | not measured | not run | | | | unknown | unknown |
| transformer.ts | not measured | not run | | | | unknown | unknown |
| transformers/classFields.ts | not measured | not run | | | | unknown | unknown |
| transformers/classThis.ts | not measured | not run | | | | unknown | unknown |
| transformers/declarations/diagnostics.ts | not measured | not run | | | | unknown | unknown |
| transformers/declarations.ts | not measured | not run | | | | unknown | unknown |
| transformers/destructuring.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2015.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2016.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2017.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2018.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2019.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2020.ts | not measured | not run | | | | unknown | unknown |
| transformers/es2021.ts | not measured | not run | | | | unknown | unknown |
| transformers/esDecorators.ts | not measured | not run | | | | unknown | unknown |
| transformers/esnext.ts | not measured | not run | | | | unknown | unknown |
| transformers/generators.ts | not measured | not run | | | | unknown | unknown |
| transformers/jsx.ts | not measured | not run | | | | unknown | unknown |
| transformers/legacyDecorators.ts | not measured | not run | | | | unknown | unknown |
| transformers/module/esnextAnd2015.ts | not measured | not run | | | | unknown | unknown |
| transformers/module/impliedNodeFormatDependent.ts | not measured | not run | | | | unknown | unknown |
| transformers/module/module.ts | not measured | not run | | | | unknown | unknown |
| transformers/module/system.ts | not measured | not run | | | | unknown | unknown |
| transformers/namedEvaluation.ts | not measured | not run | | | | unknown | unknown |
| transformers/taggedTemplate.ts | not measured | not run | | | | unknown | unknown |
| transformers/ts.ts | not measured | not run | | | | unknown | unknown |
| transformers/typeSerializer.ts | not measured | not run | | | | unknown | unknown |
| transformers/utilities.ts | not measured | not run | | | | unknown | unknown |
| tsbuild.ts | not measured | not run | | | | unknown | unknown |
| tsbuildPublic.ts | not measured | not run | | | | unknown | unknown |
| types.ts | not measured | not run | | | | unknown | unknown |
| utilities.ts | not measured | not run | | | | unknown | unknown |
| utilitiesPublic.ts | not measured | not run | | | | unknown | unknown |
| visitorPublic.ts | not measured | not run | | | | unknown | unknown |
| watch.ts | not measured | not run | | | | unknown | unknown |
| watchPublic.ts | not measured | not run | | | | unknown | unknown |
| watchUtilities.ts | not measured | not run | | | | unknown | unknown |

## What this cannot see

Scan-only observes refusal rules; it does not attempt body lowering, generic instantiation, or prove successful compilation. Full mode may skip directly diagnosed bodies and stop at failed construct boundaries. A rule can be suppressed by an earlier rule or a pruned syntax branch, as in the normal refusal walk. The report counts reasons actually emitted by this census, rather than asserting how an isolated predicate would answer at every AST node. Timed-out or failed measurements have unknown totals.

Findings are deduplicated by (kind, where, reason, text). Full-mode totals can include non-compiler boundary events; scan-only totals cover the scan and relation trailer. The original checker timeout, controls, all previously pushed per-file observations, source/raw hashes, and interruption history remain in sites.json. No adaptation or production compiler source was changed. Only stage3/census/promise-discard-sites/ is committed.

## Continuation memory

| Batch | Files | Workers | Free bytes | Available bytes | Cgroup current bytes | Limit bytes |
|---|---:|---:|---:|---:|---:|---:|

The continuation records memory immediately before each batch. Four workers are used when at least 8 GiB of effective available memory remains. Checkpoints are committed and pushed after every 20 new results, plus the final partial batch. The earlier interruption and sandbox bookkeeping stop were preserved; this continuation performs writes only in the writable scratch worktree.

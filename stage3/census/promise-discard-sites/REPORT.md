# Promise-discard sites in TypeScript’s compiler

**0 new-reason sites across all 79 compiler files: 20 measured in full mode and 59 scan only. Both controls passed.**

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
| factory/nodeConverters.ts | scan only | ran | 0 | 23.876 | False | 19 | 0 |
| factory/nodeFactory.ts | scan only | ran | 0 | 28.785 | False | 421 | 0 |
| factory/nodeTests.ts | scan only | ran | 0 | 56.724 | False | 4 | 0 |
| factory/parenthesizerRules.ts | scan only | ran | 0 | 24.829 | False | 17 | 0 |
| factory/utilities.ts | scan only | ran | 0 | 26.433 | False | 83 | 0 |
| factory/utilitiesPublic.ts | scan only | ran | 0 | 18.847 | False | 0 | 0 |
| hostErrors.ts | scan only | ran | 0 | 18.947 | False | 0 | 0 |
| moduleNameResolver.ts | scan only | ran | 0 | 20.438 | False | 43 | 0 |
| moduleSpecifiers.ts | scan only | ran | 0 | 24.795 | False | 45 | 0 |
| parser.ts | scan only | ran | 0 | 43.282 | False | 168 | 0 |
| path.ts | scan only | ran | 0 | 18.591 | False | 0 | 0 |
| performance.ts | scan only | ran | 0 | 18.648 | False | 0 | 0 |
| performanceCore.ts | scan only | ran | 0 | 17.457 | False | 0 | 0 |
| program.ts | scan only | ran | 0 | 30.859 | False | 87 | 0 |
| programDiagnostics.ts | scan only | ran | 0 | 26.118 | False | 13 | 0 |
| resolutionCache.ts | scan only | ran | 0 | 20.689 | False | 29 | 0 |
| scanner.ts | scan only | ran | 0 | 22.737 | False | 4 | 0 |
| semver.ts | scan only | ran | 0 | 18.535 | False | 0 | 0 |
| sourcemap.ts | scan only | ran | 0 | 18.008 | False | 18 | 0 |
| symbolWalker.ts | scan only | ran | 0 | 16.768 | False | 7 | 0 |
| sys.ts | scan only | ran | 0 | 21.74 | False | 18 | 0 |
| tracing.ts | scan only | ran | 0 | 20.239 | False | 1 | 0 |
| transformer.ts | scan only | ran | 0 | 20.661 | False | 7 | 0 |
| transformers/classFields.ts | scan only | ran | 0 | 36.186 | False | 75 | 0 |
| transformers/classThis.ts | scan only | ran | 0 | 18.609 | False | 3 | 0 |
| transformers/declarations/diagnostics.ts | scan only | ran | 0 | 30.921 | False | 54 | 0 |
| transformers/declarations.ts | scan only | ran | 0 | 32.788 | False | 58 | 0 |
| transformers/destructuring.ts | scan only | ran | 0 | 25.881 | False | 35 | 0 |
| transformers/es2015.ts | scan only | ran | 0 | 35.435 | False | 154 | 0 |
| transformers/es2016.ts | scan only | ran | 0 | 25.143 | False | 10 | 0 |
| transformers/es2017.ts | scan only | ran | 0 | 27.408 | False | 26 | 0 |
| transformers/es2018.ts | scan only | ran | 0 | 29.153 | False | 53 | 0 |
| transformers/es2019.ts | scan only | ran | 0 | 23.235 | False | 1 | 0 |
| transformers/es2020.ts | scan only | ran | 0 | 26.788 | False | 6 | 0 |
| transformers/es2021.ts | scan only | ran | 0 | 17.534 | False | 0 | 0 |
| transformers/esDecorators.ts | scan only | ran | 0 | 35.621 | False | 101 | 0 |
| transformers/esnext.ts | scan only | ran | 0 | 25.503 | False | 28 | 0 |
| transformers/generators.ts | scan only | ran | 0 | 30.021 | False | 83 | 0 |
| transformers/jsx.ts | scan only | ran | 0 | 26.46 | False | 14 | 0 |
| transformers/legacyDecorators.ts | scan only | ran | 0 | 24.959 | False | 18 | 0 |
| transformers/module/esnextAnd2015.ts | scan only | ran | 0 | 27.186 | False | 13 | 0 |
| transformers/module/impliedNodeFormatDependent.ts | scan only | ran | 0 | 20.867 | False | 1 | 0 |
| transformers/module/module.ts | scan only | ran | 0 | 32.996 | False | 98 | 0 |
| transformers/module/system.ts | scan only | ran | 0 | 32.842 | False | 51 | 0 |
| transformers/namedEvaluation.ts | scan only | ran | 0 | 25.64 | False | 28 | 0 |
| transformers/taggedTemplate.ts | scan only | ran | 0 | 17.158 | False | 1 | 0 |
| transformers/ts.ts | scan only | ran | 0 | 40.851 | False | 76 | 0 |
| transformers/typeSerializer.ts | scan only | ran | 0 | 27.494 | False | 9 | 0 |
| transformers/utilities.ts | scan only | ran | 0 | 31.701 | False | 23 | 0 |
| tsbuild.ts | scan only | ran | 0 | 17.899 | False | 0 | 0 |
| tsbuildPublic.ts | scan only | ran | 0 | 55.021 | False | 59 | 0 |
| types.ts | scan only | ran | 0 | 18.718 | False | 0 | 0 |
| utilities.ts | scan only | ran | 0 | 58.151 | False | 243 | 0 |
| utilitiesPublic.ts | scan only | ran | 0 | 44.864 | False | 34 | 0 |
| visitorPublic.ts | scan only | ran | 0 | 27.415 | False | 42 | 0 |
| watch.ts | scan only | ran | 0 | 17.901 | False | 17 | 0 |
| watchPublic.ts | scan only | ran | 0 | 16.977 | False | 33 | 0 |
| watchUtilities.ts | scan only | ran | 0 | 16.022 | False | 14 | 0 |

## What this cannot see

Scan-only observes refusal rules; it does not attempt body lowering, generic instantiation, or prove successful compilation. Full mode may skip directly diagnosed bodies and stop at failed construct boundaries. A rule can be suppressed by an earlier rule or a pruned syntax branch, as in the normal refusal walk. The report counts reasons actually emitted by this census, rather than asserting how an isolated predicate would answer at every AST node. Timed-out or failed measurements have unknown totals.

Findings are deduplicated by (kind, where, reason, text). Full-mode totals can include non-compiler boundary events; scan-only totals cover the scan and relation trailer. The original checker timeout, controls, all previously pushed per-file observations, source/raw hashes, and interruption history remain in sites.json. No adaptation or production compiler source was changed. Only stage3/census/promise-discard-sites/ is committed.

## Continuation memory

| Batch | Files | Workers | Free bytes | Available bytes | Cgroup current bytes | Limit bytes |
|---|---:|---:|---:|---:|---:|---:|
| 1-20 of 58 | 20 | 4 | 17549721600 | 17802444800 | 654712832 | 17179869184 |
| 21-40 of 58 | 20 | 4 | 17354866688 | 17622642688 | 681111552 | 17179869184 |
| 41-58 of 58 | 18 | 4 | 17516654592 | 17786789888 | 683450368 | 17179869184 |

The continuation records memory immediately before each batch. Four workers are used when at least 8 GiB of effective available memory remains. Checkpoints are committed and pushed after every 20 new results, plus the final partial batch. The earlier interruption and sandbox bookkeeping stop were preserved; this continuation performs writes only in the writable scratch worktree.

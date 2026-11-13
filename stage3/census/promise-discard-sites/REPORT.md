# Promise-discard sites in TypeScript’s compiler

**0 new-reason sites observed in the 79-file census: 21 files measured, 58 unmeasured. The census is incomplete.**

20 files completed full mode; checker.ts completed scan only. No site matching `an unawaited async task` was observed in these files. This does not establish zero sites across all 79.

## Controls

Rule 3210 control — passed in both the original full binary and scan-only binary:

```ts
async function fails(): Promise<void> {}
const run: () => unknown = fails;
run();
```

Location: `control.ts:2:28`; rule: `3210 refusePromiseIntoVoid`; named type: `Promise<void>`.

```text
/workspace/scratch/promise-discard-sites-run/control.ts:2:28: Adamic 0.1 refuses an unawaited async task: a Promise<void> at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task; await the task inside a function that returns void, or keep the Promise in the type it's seen as
```

Full control: exit 0, 0.173 seconds. Scan-only control: exit 0, 0.268 seconds. Exact kind/location/reason/text match; scan-only control had zero lowering units. Full raw output is preserved in sites.json.

Rule 3220 control — **not run** before the stop:

```ts
async function fails(): Promise<void> {}
[0].map(fails);
```

## Per-file coverage

Compiler head: `e4b65199b13da9dcce678ffe93e0f06fe11bb2e5`. Upstream TypeScript HEAD and peeled v6.0.3 tag both confirmed as `050880ce59e30b356b686bd3144efe24f875ebc8`. Adaptation completed successfully with 79 .ts files. Full runs use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, one absolute LATENT_ONLY_FILE, CommonJSBundle, and a 1500-second process-group kill bound.

checker.ts’s original full run timed out at 1500.312 seconds, exit -9. Its existing refusal scan was then invoked through the measurement-only scan-only.go.txt wrapper: exit 0, 206.199 seconds, 1473 distinct findings, zero new-reason sites, and zero lowering units. The original timeout remains unchanged in sites.json.

| File | Measurement | Status | Exit | Seconds | Total findings |
|---|---|---|---:|---:|---:|
| _namespaces/ts.moduleSpecifiers.ts | full | ran | 0 | 37.439 | 0 |
| _namespaces/ts.performance.ts | full | ran | 0 | 36.445 | 0 |
| _namespaces/ts.ts | full | ran | 0 | 62.66 | 0 |
| binder.ts | full | ran | 0 | 393.778 | 504 |
| builder.ts | full | ran | 0 | 180.526 | 399 |
| builderPublic.ts | full | ran | 0 | 50.686 | 3 |
| builderState.ts | full | ran | 0 | 127.299 | 37 |
| builderStatePublic.ts | full | ran | 0 | 38.806 | 0 |
| checker.ts | scan only | ran | 0 | 206.199 | 1473 |
| commandLineParser.ts | full | ran | 0 | 272.421 | 367 |
| core.ts | full | ran | 0 | 77.278 | 14 |
| corePublic.ts | full | ran | 0 | 32.716 | 0 |
| debug.ts | full | ran | 0 | 84.415 | 10 |
| diagnosticInformationMap.generated.ts | full | ran | 0 | 33.865 | 0 |
| emitter.ts | full | ran | 0 | 1357.596 | 767 |
| executeCommandLine.ts | full | ran | 0 | 192.473 | 365 |
| expressionToTypeNode.ts | full | ran | 0 | 198.716 | 183 |
| factory/baseNodeFactory.ts | full | ran | 0 | 37.64 | 10 |
| factory/emitHelpers.ts | full | ran | 0 | 73.803 | 47 |
| factory/emitNode.ts | full | ran | 0 | 37.801 | 26 |
| factory/nodeChildren.ts | full | ran | 0 | 72.538 | 16 |
| factory/nodeConverters.ts | not measured | not run | | | unknown |
| factory/nodeFactory.ts | not measured | not run | | | unknown |
| factory/nodeTests.ts | not measured | not run | | | unknown |
| factory/parenthesizerRules.ts | not measured | not run | | | unknown |
| factory/utilities.ts | not measured | not run | | | unknown |
| factory/utilitiesPublic.ts | not measured | not run | | | unknown |
| hostErrors.ts | not measured | not run | | | unknown |
| moduleNameResolver.ts | not measured | not run | | | unknown |
| moduleSpecifiers.ts | not measured | not run | | | unknown |
| parser.ts | not measured | not run | | | unknown |
| path.ts | not measured | not run | | | unknown |
| performance.ts | not measured | not run | | | unknown |
| performanceCore.ts | not measured | not run | | | unknown |
| program.ts | not measured | not run | | | unknown |
| programDiagnostics.ts | not measured | not run | | | unknown |
| resolutionCache.ts | not measured | not run | | | unknown |
| scanner.ts | not measured | not run | | | unknown |
| semver.ts | not measured | not run | | | unknown |
| sourcemap.ts | not measured | not run | | | unknown |
| symbolWalker.ts | not measured | not run | | | unknown |
| sys.ts | not measured | not run | | | unknown |
| tracing.ts | not measured | not run | | | unknown |
| transformer.ts | not measured | not run | | | unknown |
| transformers/classFields.ts | not measured | not run | | | unknown |
| transformers/classThis.ts | not measured | not run | | | unknown |
| transformers/declarations/diagnostics.ts | not measured | not run | | | unknown |
| transformers/declarations.ts | not measured | not run | | | unknown |
| transformers/destructuring.ts | not measured | not run | | | unknown |
| transformers/es2015.ts | not measured | not run | | | unknown |
| transformers/es2016.ts | not measured | not run | | | unknown |
| transformers/es2017.ts | not measured | not run | | | unknown |
| transformers/es2018.ts | not measured | not run | | | unknown |
| transformers/es2019.ts | not measured | not run | | | unknown |
| transformers/es2020.ts | not measured | not run | | | unknown |
| transformers/es2021.ts | not measured | not run | | | unknown |
| transformers/esDecorators.ts | not measured | not run | | | unknown |
| transformers/esnext.ts | not measured | not run | | | unknown |
| transformers/generators.ts | not measured | not run | | | unknown |
| transformers/jsx.ts | not measured | not run | | | unknown |
| transformers/legacyDecorators.ts | not measured | not run | | | unknown |
| transformers/module/esnextAnd2015.ts | not measured | not run | | | unknown |
| transformers/module/impliedNodeFormatDependent.ts | not measured | not run | | | unknown |
| transformers/module/module.ts | not measured | not run | | | unknown |
| transformers/module/system.ts | not measured | not run | | | unknown |
| transformers/namedEvaluation.ts | not measured | not run | | | unknown |
| transformers/taggedTemplate.ts | not measured | not run | | | unknown |
| transformers/ts.ts | not measured | not run | | | unknown |
| transformers/typeSerializer.ts | not measured | not run | | | unknown |
| transformers/utilities.ts | not measured | not run | | | unknown |
| tsbuild.ts | not measured | not run | | | unknown |
| tsbuildPublic.ts | not measured | not run | | | unknown |
| types.ts | not measured | not run | | | unknown |
| utilities.ts | not measured | not run | | | unknown |
| utilitiesPublic.ts | not measured | not run | | | unknown |
| visitorPublic.ts | not measured | not run | | | unknown |
| watch.ts | not measured | not run | | | unknown |
| watchPublic.ts | not measured | not run | | | unknown |
| watchUtilities.ts | not measured | not run | | | unknown |

## What this cannot see

The latest request’s 59-file scan-only continuation did not start. One additional full run, factory/nodeChildren.ts, completed during the handoff and is preserved. Other in-flight full runs were intentionally cancelled to honor the requested switch. The 3220 control remains unverified. Unmeasured files prevent an all-79 conclusion.

The existing latentScan helper (stage3/census/latent/lower.go.txt:519) invokes the continuing refusal scan; full mode calls it before registration and body lowering (units.go.txt:141). Both registered rules run there. The wrapper invokes this existing helper without body lowering and changes no production compiler source. Findings are deduplicated by (kind, where, reason, text); total findings also include boundary events. Full mode can skip diagnosed bodies and stop at failed construct boundaries. This census does not demonstrate successful compilation.

## Stop and preservation

The handoff bookkeeping command was mistakenly run in the read-only sandbox and failed with:

```text
OSError: [Errno 30] Read-only file system: '/workspace/scratch/promise-discard-sites-run/adamic/stage3/census/promise-discard-sites/sites.json'
```

Measurements stopped under the user’s instruction to stop on anything unexpected. This status write and push only preserve the evidence. Every previously pushed file result and the checker follow-up were checked for exact equality before writing this report.

Initial ten-file and subsequent recovery checkpoints were pushed without force. All repository changes are confined to stage3/census/promise-discard-sites/. The environment interruption and reruns remain recorded in sites.json.

## Memory

Before scan-wrapper setup and invocation, free -h reported 9.6 GiB free and 15 GiB available. Immediately before the four-worker full batch, MemFree was 11,345,326,080 bytes and MemAvailable was 17,755,111,424 bytes; cgroup current was 6,776,045,568 bytes under a 17,179,869,184-byte limit. The batch file list and measurements are preserved in sites.json. No new scan-only batch was started after the latest request.

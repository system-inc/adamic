# Promise-discard sites in TypeScript's compiler

**0 new-reason sites observed** across 20/79 attempted files. Full coverage requires successful full runs; the checker scan-only follow-up measures its refusal scan separately.


## Control

```ts
async function fails(): Promise<void> {}
const run: () => unknown = fails;
run();
```

Status: ran; exit 0; 0.173 seconds.

```json
{"checker_rejected":false,"diagnostic_sites":[],"diagnostics":[],"latent_mode":"full","measurement":"measured on a checker-rejected program","module_format":"CommonJSBundle","status":"measurement"}
{"measurement":"measured on a checker-rejected program","file":"/workspace/scratch/promise-discard-sites-run/control.ts","declarations":[{"measurement":"measured on a checker-rejected program","where":"/workspace/scratch/promise-discard-sites-run/control.ts:1:1","name":" fails","kind":"KindFunctionDeclaration","start":0,"end":40},{"measurement":"measured on a checker-rejected program","where":"/workspace/scratch/promise-discard-sites-run/control.ts:2:7","name":" run","kind":"KindVariableDeclaration","start":46,"end":73}],"units":[{"pass_ledger":[{"name":"checkOverloadResults","status":"not_run_scoped"},{"name":"forwarderValues","status":"not_run_scoped"},{"name":"finishClassCalls","status":"not_run_scoped"},{"name":"finishAccessors","status":"not_run_scoped"},{"name":"finishToPrimitive","status":"not_run_scoped"},{"name":"checkIterableViews","status":"not_run_scoped"},{"name":"exceptions","status":"not_run_scoped"},{"name":"checkLibrarySpreads","status":"not_run_scoped"},{"name":"checkAccessorSpreads","status":"not_run_scoped"},{"name":"checkRecordGuardedReads","status":"not_run_scoped"},{"name":"checkDynamicKeyedGetters","status":"not_run_scoped"},{"name":"checkGeneratorRoads","status":"not_run_scoped"},{"name":"flow.NormalizeAsync","status":"not_run_scoped"},{"name":"asyncEntrySource","status":"not_run_scoped"},{"name":"findCycles","status":"not_run_whole_program"},{"name":"checkBoundArguments","status":"not_run_scoped"},{"name":"checkViewOptionalWrites","status":"not_run_scoped"},{"name":"checkViewCopies","status":"not_run_scoped"},{"name":"checkViewWidenedReads","status":"not_run_scoped"},{"name":"checkHostMemberWrites","status":"not_run_scoped"},{"name":"asyncPayloadFailures","status":"not_run_scoped"},{"name":"checkPromiseConsumption","status":"not_run_scoped"},{"name":"checkViewLogicalWrites","status":"not_run_scoped"},{"name":"programRegion","status":"not_run_whole_program"},{"name":"applyProgramMetadata","status":"not_run_whole_program"},{"name":"ownershipQuery","status":"not_run_whole_program"},{"name":"readiness","status":"not_run_scoped"},{"name":"borrow","status":"not_run_scoped"},{"name":"counters","status":"not_run_scoped"},{"name":"voidDiscards","status":"not_run_scoped"},{"name":"checkInspectorDoomedContracts","status":"not_run_scoped"},{"name":"checkFlowLiveness","status":"not_run_scoped"}],"measurement":"measured on a checker-rejected program","where":"/workspace/scratch/promise-discard-sites-run/control.ts:1:1","kind":"KindFunctionDeclaration","body_start":37,"body_end":40,"instantiated":false,"status":"attempted","checker_diagnostics":[],"name":"fails","top":"/workspace/scratch/promise-discard-sites-run/control.ts:1:1"},{"pass_ledger":[{"name":"checkOverloadResults","status":"not_run_scoped"},{"name":"forwarderValues","status":"not_run_scoped"},{"name":"finishClassCalls","status":"not_run_scoped"},{"name":"finishAccessors","status":"not_run_scoped"},{"name":"finishToPrimitive","status":"not_run_scoped"},{"name":"checkIterableViews","status":"not_run_scoped"},{"name":"exceptions","status":"not_run_scoped"},{"name":"checkLibrarySpreads","status":"not_run_scoped"},{"name":"checkAccessorSpreads","status":"not_run_scoped"},{"name":"checkRecordGuardedReads","status":"not_run_scoped"},{"name":"checkDynamicKeyedGetters","status":"not_run_scoped"},{"name":"checkGeneratorRoads","status":"not_run_scoped"},{"name":"flow.NormalizeAsync","status":"not_run_scoped"},{"name":"asyncEntrySource","status":"not_run_scoped"},{"name":"findCycles","status":"not_run_whole_program"},{"name":"checkBoundArguments","status":"not_run_scoped"},{"name":"checkViewOptionalWrites","status":"not_run_scoped"},{"name":"checkViewCopies","status":"not_run_scoped"},{"name":"checkViewWidenedReads","status":"not_run_scoped"},{"name":"checkHostMemberWrites","status":"not_run_scoped"},{"name":"asyncPayloadFailures","status":"not_run_scoped"},{"name":"checkPromiseConsumption","status":"not_run_scoped"},{"name":"checkViewLogicalWrites","status":"not_run_scoped"},{"name":"programRegion","status":"not_run_whole_program"},{"name":"applyProgramMetadata","status":"not_run_whole_program"},{"name":"ownershipQuery","status":"not_run_whole_program"},{"name":"readiness","status":"not_run_scoped"},{"name":"borrow","status":"not_run_scoped"},{"name":"counters","status":"not_run_scoped"},{"name":"voidDiscards","status":"not_run_scoped"},{"name":"checkInspectorDoomedContracts","status":"not_run_scoped"},{"name":"checkFlowLiveness","status":"not_run_scoped"}],"measurement":"measured on a checker-rejected program","where":"/workspace/scratch/promise-discard-sites-run/control.ts:2:1","kind":"KindVariableStatement","instantiated":false,"status":"attempted","checker_diagnostics":[],"top":"/workspace/scratch/promise-discard-sites-run/control.ts:2:1"},{"pass_ledger":[{"name":"checkOverloadResults","status":"not_run_scoped"},{"name":"forwarderValues","status":"not_run_scoped"},{"name":"finishClassCalls","status":"not_run_scoped"},{"name":"finishAccessors","status":"not_run_scoped"},{"name":"finishToPrimitive","status":"not_run_scoped"},{"name":"checkIterableViews","status":"not_run_scoped"},{"name":"exceptions","status":"not_run_scoped"},{"name":"checkLibrarySpreads","status":"not_run_scoped"},{"name":"checkAccessorSpreads","status":"not_run_scoped"},{"name":"checkRecordGuardedReads","status":"not_run_scoped"},{"name":"checkDynamicKeyedGetters","status":"not_run_scoped"},{"name":"checkGeneratorRoads","status":"not_run_scoped"},{"name":"flow.NormalizeAsync","status":"not_run_scoped"},{"name":"asyncEntrySource","status":"not_run_scoped"},{"name":"findCycles","status":"not_run_whole_program"},{"name":"checkBoundArguments","status":"not_run_scoped"},{"name":"checkViewOptionalWrites","status":"not_run_scoped"},{"name":"checkViewCopies","status":"not_run_scoped"},{"name":"checkViewWidenedReads","status":"not_run_scoped"},{"name":"checkHostMemberWrites","status":"not_run_scoped"},{"name":"asyncPayloadFailures","status":"not_run_scoped"},{"name":"checkPromiseConsumption","status":"not_run_scoped"},{"name":"checkViewLogicalWrites","status":"not_run_scoped"},{"name":"programRegion","status":"not_run_whole_program"},{"name":"applyProgramMetadata","status":"not_run_whole_program"},{"name":"ownershipQuery","status":"not_run_whole_program"},{"name":"readiness","status":"not_run_scoped"},{"name":"borrow","status":"not_run_scoped"},{"name":"counters","status":"not_run_scoped"},{"name":"voidDiscards","status":"not_run_scoped"},{"name":"checkInspectorDoomedContracts","status":"not_run_scoped"},{"name":"checkFlowLiveness","status":"not_run_scoped"}],"measurement":"measured on a checker-rejected program","where":"/workspace/scratch/promise-discard-sites-run/control.ts:3:1","kind":"KindExpressionStatement","instantiated":false,"status":"attempted","checker_diagnostics":[],"top":"/workspace/scratch/promise-discard-sites-run/control.ts:3:1"}],"findings":[{"root_kind":"","depth":0,"measurement":"measured on a checker-rejected program","unit":"/workspace/scratch/promise-discard-sites-run/control.ts:2:1","phase":"refusal_scan","kind":"Refused","where":"/workspace/scratch/promise-discard-sites-run/control.ts:2:28","reason":"an unawaited async task: a Promise\u003cvoid\u003e at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task","text":"/workspace/scratch/promise-discard-sites-run/control.ts:2:28: Adamic 0.1 refuses an unawaited async task: a Promise\u003cvoid\u003e at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task; await the task inside a function that returns void, or keep the Promise in the type it's seen as"}]}
{"findings":[],"measurement":"measured on a checker-rejected program","module_format":"CommonJSBundle","relation_overflow":{"per_unit":false,"after_load":"","end":"","attribution":"endpoints only; any overflow conservatively marks every attempted unit because its onset is unknown","cache_sizes":{"assignable":240,"comparable":0,"enum":0,"flaggedPairs":0,"identity":0,"strictSubtype":4,"subtype":0,"ts2859":0}},"status":"relation_report"}
```

## Per-file coverage

Head `e4b65199b13da9dcce678ffe93e0f06fe11bb2e5`; TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`. Full mode, output guards enabled, CommonJSBundle. Each run loads the adapted compiler directory, selects one absolute file with LATENT_ONLY_FILE, and has a 1500-second process-group kill limit.

| File | Status | Exit | Seconds | Timeout | Total findings | Compiler findings |
|---|---|---:|---:|---|---:|---:|
| _namespaces/ts.moduleSpecifiers.ts | ran | 0 | 37.439 | False | 0 | 0 |
| _namespaces/ts.performance.ts | ran | 0 | 36.445 | False | 0 | 0 |
| _namespaces/ts.ts | ran | 0 | 62.66 | False | 0 | 0 |
| binder.ts | ran | 0 | 393.778 | False | 504 | 270 |
| builder.ts | ran | 0 | 180.526 | False | 399 | 240 |
| builderPublic.ts | ran | 0 | 50.686 | False | 3 | 1 |
| builderState.ts | ran | 0 | 127.299 | False | 37 | 18 |
| builderStatePublic.ts | ran | 0 | 38.806 | False | 0 | 0 |
| checker.ts | scan only (ran) | 0 | 206.199 | False | 1473 | 1473 |
| commandLineParser.ts | ran | 0 | 272.421 | False | 367 | 215 |
| core.ts | ran | 0 | 77.278 | False | 14 | 7 |
| corePublic.ts | ran | 0 | 32.716 | False | 0 | 0 |
| debug.ts | ran | 0 | 84.415 | False | 10 | 5 |
| diagnosticInformationMap.generated.ts | ran | 0 | 33.865 | False | 0 | 0 |
| emitter.ts | ran | 0 | 1357.596 | False | 767 | 335 |
| executeCommandLine.ts | ran | 0 | 192.473 | False | 365 | 187 |
| expressionToTypeNode.ts | ran | 0 | 198.716 | False | 183 | 97 |
| factory/baseNodeFactory.ts | ran | 0 | 37.64 | False | 10 | 5 |
| factory/emitHelpers.ts | ran | 0 | 73.803 | False | 47 | 26 |
| factory/emitNode.ts | ran | 0 | 37.801 | False | 26 | 17 |
| factory/nodeChildren.ts | not run | | | | unknown | unknown |
| factory/nodeConverters.ts | not run | | | | unknown | unknown |
| factory/nodeFactory.ts | not run | | | | unknown | unknown |
| factory/nodeTests.ts | not run | | | | unknown | unknown |
| factory/parenthesizerRules.ts | not run | | | | unknown | unknown |
| factory/utilities.ts | not run | | | | unknown | unknown |
| factory/utilitiesPublic.ts | not run | | | | unknown | unknown |
| hostErrors.ts | not run | | | | unknown | unknown |
| moduleNameResolver.ts | not run | | | | unknown | unknown |
| moduleSpecifiers.ts | not run | | | | unknown | unknown |
| parser.ts | not run | | | | unknown | unknown |
| path.ts | not run | | | | unknown | unknown |
| performance.ts | not run | | | | unknown | unknown |
| performanceCore.ts | not run | | | | unknown | unknown |
| program.ts | not run | | | | unknown | unknown |
| programDiagnostics.ts | not run | | | | unknown | unknown |
| resolutionCache.ts | not run | | | | unknown | unknown |
| scanner.ts | not run | | | | unknown | unknown |
| semver.ts | not run | | | | unknown | unknown |
| sourcemap.ts | not run | | | | unknown | unknown |
| symbolWalker.ts | not run | | | | unknown | unknown |
| sys.ts | not run | | | | unknown | unknown |
| tracing.ts | not run | | | | unknown | unknown |
| transformer.ts | not run | | | | unknown | unknown |
| transformers/classFields.ts | not run | | | | unknown | unknown |
| transformers/classThis.ts | not run | | | | unknown | unknown |
| transformers/declarations/diagnostics.ts | not run | | | | unknown | unknown |
| transformers/declarations.ts | not run | | | | unknown | unknown |
| transformers/destructuring.ts | not run | | | | unknown | unknown |
| transformers/es2015.ts | not run | | | | unknown | unknown |
| transformers/es2016.ts | not run | | | | unknown | unknown |
| transformers/es2017.ts | not run | | | | unknown | unknown |
| transformers/es2018.ts | not run | | | | unknown | unknown |
| transformers/es2019.ts | not run | | | | unknown | unknown |
| transformers/es2020.ts | not run | | | | unknown | unknown |
| transformers/es2021.ts | not run | | | | unknown | unknown |
| transformers/esDecorators.ts | not run | | | | unknown | unknown |
| transformers/esnext.ts | not run | | | | unknown | unknown |
| transformers/generators.ts | not run | | | | unknown | unknown |
| transformers/jsx.ts | not run | | | | unknown | unknown |
| transformers/legacyDecorators.ts | not run | | | | unknown | unknown |
| transformers/module/esnextAnd2015.ts | not run | | | | unknown | unknown |
| transformers/module/impliedNodeFormatDependent.ts | not run | | | | unknown | unknown |
| transformers/module/module.ts | not run | | | | unknown | unknown |
| transformers/module/system.ts | not run | | | | unknown | unknown |
| transformers/namedEvaluation.ts | not run | | | | unknown | unknown |
| transformers/taggedTemplate.ts | not run | | | | unknown | unknown |
| transformers/ts.ts | not run | | | | unknown | unknown |
| transformers/typeSerializer.ts | not run | | | | unknown | unknown |
| transformers/utilities.ts | not run | | | | unknown | unknown |
| tsbuild.ts | not run | | | | unknown | unknown |
| tsbuildPublic.ts | not run | | | | unknown | unknown |
| types.ts | not run | | | | unknown | unknown |
| utilities.ts | not run | | | | unknown | unknown |
| utilitiesPublic.ts | not run | | | | unknown | unknown |
| visitorPublic.ts | not run | | | | unknown | unknown |
| watch.ts | not run | | | | unknown | unknown |
| watchPublic.ts | not run | | | | unknown | unknown |
| watchUtilities.ts | not run | | | | unknown | unknown |

## What this cannot see

This is measurement on a checker-rejected program, not successful compilation. Full mode skips directly diagnosed bodies, independently attempts eligible nested declarations, and may stop at failed construct boundaries. It cannot establish absence inside skipped bodies or unexamined children. Timed-out or failed files have unknown totals. Findings deduplicate (kind, where, reason, text) across attempts; total findings include non-compiler events, while compiler findings count Refused and NotYet. No adaptation or production compiler source is changed.

The JSON preserves full matching findings, control output, exit status, timing, coverage, source hashes, and raw stream hashes. Raw logs remain in scratch.

## Checker refusal scan only

Existing `latentScan` in `stage3/census/latent/lower.go.txt:519`, called before registration/lowering by `units.go.txt:141`, is invoked by the measurement-only wrapper `scan-only.go.txt`. It performs no body lowering. Status: ran; exit 0; 206.199 seconds; findings 1473; new-reason findings 0. The original 25-minute full timeout remains preserved in sites.json.

Control for the scan-only binary:
```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/promise-discard-sites-run/control.ts:2:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/workspace/scratch/promise-discard-sites-run/control.ts:2:28",
    "reason": "an unawaited async task: a Promise<void> at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task",
    "text": "/workspace/scratch/promise-discard-sites-run/control.ts:2:28: Adamic 0.1 refuses an unawaited async task: a Promise<void> at the initializer site through () seen as unknown, a type that names no Promise, so whoever holds it drops the task; await the task inside a function that returns void, or keep the Promise in the type it's seen as",
    "file": "/workspace/scratch/promise-discard-sites-run/control.ts",
    "line": 2,
    "column": 28,
    "rule": "3210 refusePromiseIntoVoid",
    "named_type": "Promise<void>"
  }
]
```

Interruption: the environment restart terminated the original supervisor. The orphan emitter process group was killed, and results lacking durable exit/timing records were rerun. Pushed results were retained.

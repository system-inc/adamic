# Census CommonJS bundle evidence — 995spev

Main: `fb2b54da12a02e1550b9f5bf046d3314c77bf657`. Change: `2934a22d8341476aaf07925518187422de89e138`.
TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Both probes read the same adapted tree; all source hashes are in `adapted-source-manifest.json`.

Main uses its implicit EcmaScript default. Head uses its declared CommonJSBundle default. No compiler feature changes or source adaptation changes are applied between the two measurements. No native-output claim is made.

Box: 5 visible CPUs; cpu.max `400000 100000`; memory.max `17179869184` bytes.

## Method

The recipe follows `typescript/next-stops-f44b5c27-2`, `stage3/census/next-stops-f44b5c27/{run.py,run-one.py,measure.py}`. The ordinary driver runs once per child for each of the sorted 79 compiler roots, retaining both per-file and aggregate rows and checking exact agreement except timing. Checkpoints contain at most 20 roots. `roots.py` and `run.py` are retained here. Each child and checkpoint has a 1,780-second bound, with exit code and stopping reason in its status JSON. The checkpoint process group includes its nested children.

Full sys.ts pair: `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=/workspace/scratch/995spev/adapted/src/compiler/sys.ts <main-or-head-latent> /workspace/scratch/995spev/adapted/src/compiler <output>`. Both complete streams retain measurement headers and relation trailers. Sites compare unique `(kind, where, reason, text)` tuples across attempts, with only the identical adapted source prefix removed.

`apply.sh` ran on the frozen main source pin; its final source has `getNodeSystem` at 1475 and the three CommonJS module-value reads at 1494. The complete source manifest was captured before measurement and verified afterwards. The two builds share exactly pinned cohere/TypeScript dependencies; the baseline uses a symlink to them and `-buildvcs=false`. All four probe binary hashes are recorded in provenance.

The ordinary census children were paused for at most five minutes to give the warm whole-package test the CPU quota. measurement-pause.json records the exact interval and automatic resumption; measurement-pause-processes.json names the affected commands. Their raw elapsed times include the pause. The evidence compares stops, never performance.

## Ordinary first-stop delta

74/79 roots moved. Every moved root is listed below; `next-stops.json` contains all 79 before/after rows, and raw per-root JSONL retains the duplicate observation.

| Root | Main first stop | Head first stop |
| --- | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| _namespaces/ts.performance.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| _namespaces/ts.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| binder.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| builder.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| builderPublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| builderState.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| builderStatePublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| checker.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| commandLineParser.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| core.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet performanceCore.ts:29:15 the host binding performance in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| emitter.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| executeCommandLine.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| expressionToTypeNode.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/baseNodeFactory.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/emitHelpers.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/emitNode.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/nodeChildren.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/nodeConverters.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/nodeFactory.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/nodeTests.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/parenthesizerRules.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/utilities.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| factory/utilitiesPublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| moduleNameResolver.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| moduleSpecifiers.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| parser.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| path.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| performance.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| performanceCore.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| program.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| programDiagnostics.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| resolutionCache.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| scanner.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| semver.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| sourcemap.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| symbolWalker.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| sys.ts | Refused scanner.ts:469:12 a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| tracing.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformer.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/classFields.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/classThis.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/declarations/diagnostics.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/declarations.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/destructuring.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2015.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2016.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2017.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2018.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2019.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2020.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/es2021.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/esDecorators.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/esnext.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/generators.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/jsx.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/legacyDecorators.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/module/esnextAnd2015.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/module/impliedNodeFormatDependent.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/module/module.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/module/system.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/namedEvaluation.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/taggedTemplate.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/ts.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/typeSerializer.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| transformers/utilities.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| tsbuild.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| tsbuildPublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| utilities.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| utilitiesPublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| visitorPublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| watch.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| watchPublic.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |
| watchUtilities.ts | Refused sys.ts:57:1 an ambient function declaration | NotYet core.ts:2599:15 the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), |

## Full-mode sys.ts site delta

Unique sites: main 157; head 151. Gone: 6; new: 0.

| Kind | Main unique sites | Head unique sites |
| --- | --- | --- |
| NotYet | 59 | 55 |
| Refused | 26 | 26 |
| Boundary | 72 | 70 |

Four primary NotYet sites disappear; the two additional gone records are recovery boundaries. No primary or boundary site is new.

### Gone

| Kind | Site | Exact reason | Exact text |
| --- | --- | --- | --- |
| Boundary | sys.ts:1494:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| Boundary | sys.ts:1724:13 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1724:41: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1494:121 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:121: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1494:35 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1494:92 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:92: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1724:41 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1724:41: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |

### New

| Kind | Site | Exact reason | Exact text |
| --- | --- | --- | --- |

### The three reads at 1494

These are the first `__filename`, the `__dirname` argument, and the final `__filename`. The rows below retain every finding at that line, including exact refusals.

**main**

| Kind | Site | Exact reason | Exact text |
| --- | --- | --- | --- |
| Boundary | sys.ts:1494:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1494:121 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:121: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1494:35 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| NotYet | sys.ts:1494:92 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:92: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |

**head**

| Kind | Site | Exact reason | Exact text |
| --- | --- | --- | --- |

### getNodeSystem, sys.ts:1475

**main**

```json
[
  {
    "pass_ledger": [
      {
        "name": "checkOverloadResults",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "forwarderValues",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "finishClassCalls",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "finishAccessors",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "finishToPrimitive",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkIterableViews",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "exceptions",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkLibrarySpreads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkAccessorSpreads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkRecordGuardedReads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkDynamicKeyedGetters",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkGeneratorRoads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "flow.NormalizeAsync",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "asyncEntrySource",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "findCycles",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkBoundArguments",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewOptionalWrites",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewCopies",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewWidenedReads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkHostMemberWrites",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "asyncPayloadFailures",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkPromiseConsumption",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewLogicalWrites",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "programRegion",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "applyProgramMetadata",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "ownershipQuery",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "readiness",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "borrow",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "counters",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "voidDiscards",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkInspectorDoomedContracts",
        "status": "not_reached_lowering_stop"
      }
    ],
    "measurement": "measured on a checker-rejected program",
    "where": "sys.ts:1475:5",
    "kind": "KindFunctionDeclaration",
    "body_start": 65107,
    "body_end": 85438,
    "instantiated": false,
    "status": "attempted",
    "checker_diagnostics": [],
    "name": "getNodeSystem",
    "parent": "sys.ts:1469:1",
    "top": "sys.ts:1469:1"
  }
]
```

**head**

```json
[
  {
    "pass_ledger": [
      {
        "name": "checkOverloadResults",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "forwarderValues",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "finishClassCalls",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "finishAccessors",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "finishToPrimitive",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkIterableViews",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "exceptions",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkLibrarySpreads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkAccessorSpreads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkRecordGuardedReads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkDynamicKeyedGetters",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkGeneratorRoads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "flow.NormalizeAsync",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "asyncEntrySource",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "findCycles",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkBoundArguments",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewOptionalWrites",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewCopies",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewWidenedReads",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkHostMemberWrites",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "asyncPayloadFailures",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkPromiseConsumption",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkViewLogicalWrites",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "programRegion",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "applyProgramMetadata",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "ownershipQuery",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "readiness",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "borrow",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "counters",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "voidDiscards",
        "status": "not_reached_lowering_stop"
      },
      {
        "name": "checkInspectorDoomedContracts",
        "status": "not_reached_lowering_stop"
      }
    ],
    "measurement": "measured on a checker-rejected program",
    "where": "sys.ts:1475:5",
    "kind": "KindFunctionDeclaration",
    "body_start": 65107,
    "body_end": 85438,
    "instantiated": false,
    "status": "attempted",
    "checker_diagnostics": [],
    "name": "getNodeSystem",
    "parent": "sys.ts:1469:1",
    "top": "sys.ts:1469:1"
  }
]
```

**main getNodeSystem findings**

| Phase | Kind | Site | Exact reason | Exact text |
| --- | --- | --- | --- | --- |
| lowering | Boundary | sys.ts:1494:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| lowering | Boundary | sys.ts:1498:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:57:1: stage 0 can't lower a function without a body yet |
| lowering | Boundary | sys.ts:1519:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1590:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet |
| lowering | Boundary | sys.ts:1542:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1542:65 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1545:21 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1546:25 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1549:25 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1549:56 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1551:29 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1724:13 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1724:41: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| lowering | Boundary | sys.ts:1818:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1818:24: stage 0 can't lower reading byteOrderMarkIndicator yet |
| lowering | Boundary | sys.ts:1869:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1869:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| lowering | Boundary | sys.ts:1870:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1870:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| lowering | NotYet | sys.ts:1494:35 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| lowering | NotYet | sys.ts:1519:15 | a function value that captures the variable its own initializer declares | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | NotYet | sys.ts:1590:191 | process.recordreplay as a value (use its supported process operation directly) | sys.ts:1590:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet |
| lowering | NotYet | sys.ts:1724:41 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1724:41: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| lowering | NotYet | sys.ts:1818:24 | reading byteOrderMarkIndicator | sys.ts:1818:24: stage 0 can't lower reading byteOrderMarkIndicator yet |
| lowering | NotYet | sys.ts:57:1 | a function without a body | sys.ts:57:1: stage 0 can't lower a function without a body yet |
| lowering | Refused | sys.ts:1869:17 | sort without a comparator | sys.ts:1869:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| lowering | Refused | sys.ts:1870:17 | sort without a comparator | sys.ts:1870:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| refusal_scan | NotYet | sys.ts:1494:121 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:121: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| refusal_scan | NotYet | sys.ts:1494:35 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:35: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| refusal_scan | NotYet | sys.ts:1494:92 | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | sys.ts:1494:92: stage 0 can't lower node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) yet |
| refusal_scan | NotYet | sys.ts:1590:192 | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | sys.ts:1590:192: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet |
| refusal_scan | Refused | sys.ts:1534:13 | a function taking PollingInterval seen as one taking number &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | sys.ts:1534:13: Adamic 0.1 refuses a function taking PollingInterval seen as one taking number &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) |
| refusal_scan | Refused | sys.ts:1535:13 | a function taking boolean seen as one taking boolean &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | sys.ts:1535:13: Adamic 0.1 refuses a function taking boolean seen as one taking boolean &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) |
| refusal_scan | Refused | sys.ts:1612:22 | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string &#124; undefined where string is read | sys.ts:1612:22: Adamic 0.1 refuses a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string &#124; undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| refusal_scan | Refused | sys.ts:1614:77 | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) &#124; undefined would replace a method of type (path: string) => boolean | sys.ts:1614:77: Adamic 0.1 refuses a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) &#124; undefined would replace a method of type (path: string) => boolean; make directoryExists readonly in the type it's seen as (readonly directoryExists: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable) |

**head getNodeSystem findings**

| Phase | Kind | Site | Exact reason | Exact text |
| --- | --- | --- | --- | --- |
| lowering | Boundary | sys.ts:1498:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:57:1: stage 0 can't lower a function without a body yet |
| lowering | Boundary | sys.ts:1519:9 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1590:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet |
| lowering | Boundary | sys.ts:1542:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1542:65 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1545:21 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1546:25 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1549:25 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1549:56 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1551:29 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | Boundary | sys.ts:1818:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1818:24: stage 0 can't lower reading byteOrderMarkIndicator yet |
| lowering | Boundary | sys.ts:1869:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1869:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| lowering | Boundary | sys.ts:1870:17 | failed statement; state rolled back; continued at next statement or declaration | sys.ts:1870:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| lowering | NotYet | sys.ts:1519:15 | a function value that captures the variable its own initializer declares | sys.ts:1519:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| lowering | NotYet | sys.ts:1590:191 | process.recordreplay as a value (use its supported process operation directly) | sys.ts:1590:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet |
| lowering | NotYet | sys.ts:1818:24 | reading byteOrderMarkIndicator | sys.ts:1818:24: stage 0 can't lower reading byteOrderMarkIndicator yet |
| lowering | NotYet | sys.ts:57:1 | a function without a body | sys.ts:57:1: stage 0 can't lower a function without a body yet |
| lowering | Refused | sys.ts:1869:17 | sort without a comparator | sys.ts:1869:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| lowering | Refused | sys.ts:1870:17 | sort without a comparator | sys.ts:1870:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| refusal_scan | NotYet | sys.ts:1590:192 | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | sys.ts:1590:192: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet |
| refusal_scan | Refused | sys.ts:1534:13 | a function taking PollingInterval seen as one taking number &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | sys.ts:1534:13: Adamic 0.1 refuses a function taking PollingInterval seen as one taking number &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) |
| refusal_scan | Refused | sys.ts:1535:13 | a function taking boolean seen as one taking boolean &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | sys.ts:1535:13: Adamic 0.1 refuses a function taking boolean seen as one taking boolean &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) |
| refusal_scan | Refused | sys.ts:1612:22 | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string &#124; undefined where string is read | sys.ts:1612:22: Adamic 0.1 refuses a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string &#124; undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| refusal_scan | Refused | sys.ts:1614:77 | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) &#124; undefined would replace a method of type (path: string) => boolean | sys.ts:1614:77: Adamic 0.1 refuses a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) &#124; undefined would replace a method of type (path: string) => boolean; make directoryExists readonly in the type it's seen as (readonly directoryExists: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable) |

The full unit is attempted on both sides. Its first recorded refusal_scan finding moves from sys.ts:1494:35 to the existing sys.ts:1534:13 PollingInterval parameter-variance refusal. In the lowering-phase ledger, its first recorded non-boundary finding moves from sys.ts:1724:41 to sys.ts:1818:24 (reading byteOrderMarkIndicator). Full mode continues rather than treating either ledger as an ordinary next-stop result. Its 1494 and nested 1724 module-value stops are removed. Remaining lowering findings are at sys.ts:57:1 (function without a body), 1519:15 (self-capturing initializer), 1590:191 (process.recordreplay as a value), 1818:24 (byteOrderMarkIndicator), and the two comparator-free sorts at 1869:17 and 1870:17. No whole-unit lowering or native-output success is claimed.

## Validation

Final audit: all 158 root children exit 0, all 316 raw observations agree in duplicate pairs, all eight checkpoints exit 0, and the 747 source-file hashes are unchanged. Every recorded wrapper is below its 1,780-second bound. The initial complete Go command reached a 10-minute Go test timeout in internal/lower (exit 1); all census package results in that command passed. A warm whole-package retry uses -timeout=24m inside the same 1,780-second outer bound. It completed after 1234.7 test seconds and reported only five failures: the ProgramRegion/CycleScanning corpus-product tests required ADAMIC_TYPESCRIPT_SOURCE. That input had been set for apply but was omitted from the whole-package test command. The five named failures were rerun with the pinned checkout in lower-source-tests and all passed (exit 0, 26.6 test seconds). The whole-package exit remains 1 in the record; every other test in that completed run passed. No code was changed between these runs. The complete command and child exit code for every setup, build, test and measurement are in `*-status.json`; logs are compressed losslessly. The two full sys.ts streams are also gzip-compressed; the small raw per-root JSONL files remain plain text. artifact-manifest.json hashes every packaged artifact. To rerun analysis on these raw files, restore them under the recorded scratch path and decompress the two sys.ts streams first.

| Run | Child exit | Seconds | Stopping reason |
| --- | --- | --- | --- |
| validation | 0 | 0.616 | normal completion |
| remote-code | 0 | 0.566 | normal completion |
| go-tests-2 | 1 | 1051.742 | Go test default 10-minute timeout |
| lower-tests-warm | 1 | 1241.582 | five missing-source-input failures; corrected retry passed |
| lower-source-tests | 0 | 32.736 | normal completion |
| overlay-tests | 0 | 352.654 | normal completion |
| guards-final | 0 | 0.841 | normal completion |
| guard-mutant | 1 | 2.257 | expected omission mutant rejected |
| latent-python-tests | 0 | 0.719 | normal completion |
| meter-tests-full | 0 | 38.147 | normal completion |
| head-build | 0 | 385.885 | normal completion |
| latent-build | 0 | 25.398 | normal completion |
| entry-build | 0 | 106.541 | normal completion |
| main-build | 0 | 424.384 | normal completion |
| main-latent-build | 0 | 100.154 | normal completion |
| predicate-build | 1 | 23.598 | pre-existing harness compile failure |
| main-predicate-build | 1 | 20.43 | pre-existing harness compile failure |
| void-build | 1 | 60.629 | pre-existing harness compile failure |
| main-void-build | 1 | 25.168 | pre-existing harness compile failure |
| template-overlay-tests | 1 | 4.055 | pre-existing harness compile failure |
| main-sys | 0 | 286.14 | normal completion |
| head-sys | 0 | 268.809 | normal completion |

`go-tests-2` runs whole `stage3/census/latent/...`, `stage3/census/moduleformat`, `stage3/census/tool`, and `internal/lower` packages. `overlay-tests` runs the whole latent subtree with the actual overlay and `census_overlay` tag, including statecopy and overlaytext. All 94 meter tests ran with real ordinary/latent/entry binaries and no skips. New meter cases are `ModuleFormatTests.test_explicit_formats_and_historical_default`, `ModuleFormatTests.test_entry_and_ordinary_record_declared_format`, `ModuleFormatTests.test_declared_format_reaches_real_lowering`, `ModuleFormatTests.test_entry_template_changes_probe_key_and_kills_omission_mutant`, `ModuleFormatTests.test_format_beside_numbers`, and `ModuleFormatTests.test_erased_report_format_mutant`.

The dropped-format guard mutant intentionally exits 1, with `moduleFormat disagrees: production "moduleFormat", latent ""`. The meter erasure and cache-input omission mutants are rejected inside their named tests. The format guard's sole frozen setting exception is elementKey. Per-attempt program/checker/entry/plan/cycleTypes context is outside that build-setting comparison.

Extra legacy-overlay builds reveal pre-existing failures reproduced on untouched main: predicates' text-refusal rewriting generates syntax errors at lines 479 and 556; the void-only scanner does not return the relation report now required by the stream driver. The additional direct internal/lower overlay-test invocation cannot compile because the measurement overlay removes production helpers referenced by that package's tests; the supported whole latent overlay suite passes. These failures were retained, and those historical harness repairs were left outside this change.

`construction-sites.txt` retains the full stage3 construction-site search. All active tsc lowerings in census/meter were updated, including first-error, full, speculative, replay, predicates, disposable module-order probes and the void-only scanner. The remaining census literals are fixture-only tests or copy-state tests. Other stage3 locations are separate historical probes/evidence outside census/meter. Entry reuses the latent lowering literals. build-probes keys the shared Go helper, all census templates, and now the entry driver template. New package moduleformat imports only the standard library, so compiler-dependencies.json's compiler edges remain true.

Initial setup failures (missing submodules, missing tsc) and report-layout test failures were resolved before the final passing runs; their exact logs/statuses are retained. The first audit attempted a remote-tracking ref absent under this checkout’s narrow fetch mapping. remote-code verifies the actual server ref with git ls-remote, and the corrected validation checks that receipt. The original failed audit receipt is retained. No force push is used.

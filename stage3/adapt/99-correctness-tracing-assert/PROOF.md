# baobab proof

Base: e24f2600a5d8a8521c5128827cad395d47c81c81 (origin/main at task setup).
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8, version 6.0.3.
Node: v24.19.0, Linux x64.

## Owner and exact scope

32, 47, 62, 63 and 65 adapt tracing.ts, but none pins stopTracing whole or the
exact first assert expression. This adaptation therefore owns the complete
composed stopTracing implementation. Both pins preserve CRLF bytes. The source
audit compares 747 files: only src/compiler/tracing.ts differs, exactly by the
two exclamation marks. The message and all other bytes match. Three owner-drift
mutations are refused, and the second application is identical.

Debug.assert receives a boolean where upstream passed the namespace object or
undefined. The asserts-expression narrowing is lost. The next occurrence of
tracing assigns undefined; there is no intervening or subsequent value read in
stopTracing. Both all-79-root adamic types commands exit 0 with zero errors.

## Completed normal checks

Both full bash stage3/apply.sh commands exit 0. Both normal npm run build
commands exit 0. The paired Node witness returns successfully after startTracing
and throws when stopTracing is called without starting. The paired serialized
outcomes and messages are byte-identical. The failure message is:

    Debug Failure. False expression: Tracing is not in progress

The baseline and candidate positive counter controls each record two calls,
one passing and one failing assertion, with zero conversion or outcome differences.
The controls use separate directories from the whole lane measurement.

## Core root and restoring mutant

The real stage3/census/tool gate is used with a scratch selection-only overlay.
Its loader, checker and lowering are unchanged. core.ts alone is the selected root.
The baseline and restoring mutant both stop NotYet at tracing.ts:110:22:

    a namespace value seen as an argument to a callee whose body can't be named

The mutant copies the candidate source and restores only the bare argument.
Every one of its 747 source files then matches the baseline byte for byte.
The candidate's new first stop is Refused at tracing.ts:206:67:

    optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

The proposed fix at that next wall is to declare getPositionOfLineAndCharacter
on the source type or build a fresh object with known fields (adamic/no-optional-widening).
That wall is outside this adaptation's scope.

The historical census loader-overlay generator failed a stale assertion against
current main. Its failed log is retained; the successful real-gate measurements
use only the root-selection overlay described above. An obsolete declaration
census build was explicitly cancelled and is not used as evidence.

## Every emitted JavaScript hunk

All 10 normal emitted JavaScript files are compared. Exactly three hunks differ:

| File | Hunk | Classification |
| --- | --- | --- |
| _tsc.js | `@@ -2221,7 +2221,7 @@` | stopTracing argument: tracing becomes !!tracing |
| run.js | `@@ -2611,7 +2611,7 @@` | stopTracing argument: tracing becomes !!tracing |
| typescript.js | `@@ -4817,7 +4817,7 @@` | stopTracing argument: tracing becomes !!tracing |

There are no additional hunks and no esbuild rename hunks. A direct byte check
also proves that each complete JavaScript file equals its baseline after only
the approved literal argument replacement. A stock TypeScript
scanner also compares all tokens after the one approved argument edit,
normalizing numeric identifier suffixes to admit pure esbuild renames; it passes.
The seven remaining JavaScript files are byte-identical. The runtime ts object's
2,256 sorted own export keys are identical; all are enumerable and there are no symbol keys.

## Whole lane

prepare_shards.py succeeds. All 398 real TestShard_ Go units pass
with exit 0. Each Go unit checks its observed oracle counts against its manifest
expectation: 106366 passing, 1 expected failing and
0 skipped tests in total. The counter preload records
797 completed Node processes, and every manifest unit has
counter-loaded process coverage.

stopTracing's first assert executes **0 times** in the entire
lane. Boolean(tracing) versus !!tracing differences: **0**;
assertion-outcome differences: **0**. The oracle never calls stopTracing. This is a zero count, not a coverage claim for the changed assertion. The separate positive controls supply direct execution evidence.

The counted tree is a temporary copy of the prepared artifact. The source
adaptation is never instrumented. The prepared artifact's 747 source files and
10 normal JavaScript files match the validated candidate byte for byte. Only
private emitted bundle copies receive a counter before the first assert.
The VM-aware preload persists zero rows as well as actual call rows. A
measurement-only Go overlay selects that copied product and carries the preload
through the gate's filtered tool environment; production Go files remain intact.

The first attempt is invalid: all 398 units failed before running their oracle
because the shared artifact's node_modules collided with the product-path
link. It is retained separately as attempt1 and excluded from all counts.
Removing that directory only from the temporary copy restores the gate's product
layout. The successful full rerun uses parallel=2, without a test filter beyond
all TestShard_ units. No failed subset is spliced into the successful run.

Prepared artifact key: 992a2f82aa271647ada0e23b446f5cf9c161ee2fa63c47629d3e51a1825d2a11.
The lane ran against frozen inputs before this final README and proof summary
were added. Adapter code, pins and all runtime source bytes remain unchanged.


## Reproduction and evidence

Raw evidence, scripts, counter rows and complete logs: [typescript/tracing-assert-evidence](https://github.com/system-inc/adamic/tree/typescript/tracing-assert-evidence/stage3/tracing-assert-evidence).

Setup uses source /workspace/adamic-tools/env.sh, bash cloud/setup.sh, npm ci
for stage3/api and stage3/typescript, and ADAMIC_TYPESCRIPT_SOURCE at the pinned
source commit. The evidence run wrappers record every completed child's exit
and preserve logs to files. Source adapters and normal build outputs contain
no counter instrumentation. No production Go file is changed.

## Operational deviations retained in evidence

An early log preview accidentally used tail -c; the complete primary log was
retained. A checkpoint script used an absent remote-tracking ref and accidentally
deleted its checkpoint branch. The retained original commit was restored as the
parent of the next checkpoint, using a normal push. No force push or main push
was used. checkpoint-history.json and the push logs retain the incident.

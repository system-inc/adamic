# Lowering time on tsc's 79 files, three compilers

Asked: per-file seconds for `adamic c` on each file of the adapted tsc tree's `src/compiler`, under main
and the two trees before #vdt9f1q's predicate changes, and a profile of where lowering spends its time,
because the stage 3 meter rose 22% (3,558 s to 4,324 s, Oct 26 to Oct 27).

## Answer

`adamic c` never reaches lowering on tsc's files, on any of the three compilers, so this measurement
cannot see the predicate passes and does not explain the meter's rise.

- 77 of 79 files stop at the checker on all three trees, each printing the same 269 lines of checker
  diagnostics (first: `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.`). Every
  file's import closure reaches `checker.ts`, so every one fails there in about 3.3 s.
- The other two, `corePublic.ts` and `hostErrors.ts`, are checker-clean, lower and emit C in 0.05 s.
- Totals: main 259.4 s, A 261.9 s, B 259.1 s (main/A 0.990, main/B 1.001). No file differs by more
  than the run-to-run noise (largest round-to-round difference on any file and tree: 2.11 s).
- CPU profiles of main's 5 slowest files have no samples in `internal/lower` at all: zero in
  `predicateVerifier`, zero in `GetTypeAtLocation`, zero in `predicateFalseFlow` (the false-direction
  walk, `predicate_false_flow.go`) and zero in its escape scan (`predicateFalseFlow.escapes` and
  `escapingReference`). All of the time is typescript-go's checker under `load.Load`.
- Decisions: the stdout and stderr of `adamic c` are byte-identical between main and A and between
  main and B on all 79 files (316 comparisons, 0 differ). The two rounds were also identical.

Inferred, not measured here: the 310 s whole-directory wall with `predicateVerifier.verify` on its
stack has to come from a path that lowers bodies in checker-rejected programs. That is the meter's
latent census (`stage3/census/latent`, full mode, which "measures eligible nested bodies even on a
rejected entry-root program"). Ordinary `adamic c`, as in measure.py, stops before lowering. To find
the costly pass, time and profile the meter's latent overlay per file on the same three trees. I did not
build that overlay here.

## The box

- Linux 6.18.44, 4 CPUs (`nproc` 4), `Intel(R) Xeon(R) Processor @ 2.10GHz`, 15 GB RAM, no swap.
- `/proc/loadavg` before the builds: `0.26 0.06 0.02`. At the timing run's start: `1.29 1.33 1.39`
  (the tail of the three compiler builds and the apply). At its end: `1.31 2.50 1.97`
  ([loadavg-start.txt](lowering-time-tsc/loadavg-start.txt), [loadavg-end.txt](lowering-time-tsc/loadavg-end.txt)).
  Nothing else ran during the timings, which were one process at a time.
- Go 1.27.1, as go.mod asks. Node 22.22.0, not 24: apply accepted it. clang was not needed, since
  `adamic c` emits C without compiling it and apply doesn't call clang.
- GNU time from Ubuntu noble's `time` package (`/usr/bin/time` was missing and was installed).

## The trees

| label | commit | note |
|---|---|---|
| main | `8dfcdf7e1eec44ef94263b3f2b373091bc12b33c` | |
| A | `76ea8b17ea0ed467ecb86f355c06f738fd262983` | parent of 4ac3f04d (false-flow NotYet) |
| B | `22026c7392f55556e721b6f425dd369ceafdf14f` | parent of 5e382ce4 (checked predicate values) |

Each is a `git worktree add --detach` with `git submodule update --init --recursive --depth 1`. All three
pin cohere `06af2e39` and cohere/TypeScript `d92d9bfe`. Each built with
`go build -o <tree>/adamic ./cmd/adamic`.

The adapted tree was built once from main's `stage3/apply.sh` (TypeScript v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`, `stage3/api` installed with `npm ci` from its lockfile)
and used by all three compilers:

```
STAGE3_CACHE=/tmp/lt/stage3cache ADAMIC_TYPESCRIPT_SOURCE=/tmp/lt/ts stage3/apply.sh /tmp/lt/adapted
```

Apply's inputs differ between main and A or B only in `stage3/adapt/ACHECK.md` (new docs), four `.a`
files under `stage3/adapt/*/language-questions` and `evidence`, and one probe renamed from `.a` to `.ts`
in `75-optional-widening`. None of them is written into `src/compiler`, so A's and B's own apply would
not change these inputs.

## Commands

[run.sh](lowering-time-tsc/run.sh.txt) lists the 79 `.ts` files under the adapted `src/compiler`
([files.txt](lowering-time-tsc/files.txt)). For each file, from the file's own directory (as
measure.py does), it runs:

```
/usr/bin/time -v -o <out>.time timeout 600 <tree>/adamic c <file> > <out>.stdout 2> <out>.stderr
```

for main, A and B in turn. It goes through all 79 files once, then again. The table takes the lower wall of
the two rounds per file and tree. [table.py](lowering-time-tsc/table.py.txt) renders it. The raw `time -v`
output for every run is under [time/](lowering-time-tsc/time).

## Per file

Seconds are wall time, the lower of two runs. Exit is main/A/B. The first line is main's first line of
stderr; A's and B's are identical.

| file | main s | A s | B s | main/A | main/B | main user s | main RSS MB | exit m/A/B | first line (main) |
|---|--:|--:|--:|--:|--:|--:|--:|:-:|---|
| watchUtilities.ts | 3.92 | 3.55 | 3.28 | 1.10 | 1.20 | 4.43 | 374 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| performanceCore.ts | 3.83 | 3.45 | 3.46 | 1.11 | 1.11 | 4.28 | 367 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| program.ts | 3.75 | 3.43 | 3.64 | 1.09 | 1.03 | 4.21 | 347 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformer.ts | 3.75 | 3.47 | 3.72 | 1.08 | 1.01 | 3.95 | 367 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| performance.ts | 3.72 | 3.56 | 3.33 | 1.04 | 1.12 | 4.01 | 370 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/emitHelpers.ts | 3.66 | 3.37 | 3.19 | 1.09 | 1.15 | 4.01 | 382 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/nodeConverters.ts | 3.62 | 3.29 | 3.68 | 1.10 | 0.98 | 4.03 | 355 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| scanner.ts | 3.62 | 3.41 | 3.69 | 1.06 | 0.98 | 4.03 | 354 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| moduleSpecifiers.ts | 3.61 | 3.73 | 3.43 | 0.97 | 1.05 | 3.89 | 370 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| parser.ts | 3.60 | 3.44 | 3.42 | 1.05 | 1.05 | 4.12 | 361 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| semver.ts | 3.59 | 3.70 | 3.19 | 0.97 | 1.13 | 3.99 | 349 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| resolutionCache.ts | 3.56 | 3.16 | 3.40 | 1.13 | 1.05 | 4.06 | 366 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| types.ts | 3.56 | 3.51 | 3.27 | 1.01 | 1.09 | 3.95 | 381 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| programDiagnostics.ts | 3.54 | 3.62 | 3.49 | 0.98 | 1.01 | 3.70 | 383 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/module/module.ts | 3.54 | 3.40 | 3.31 | 1.04 | 1.07 | 3.79 | 386 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| path.ts | 3.52 | 3.65 | 3.38 | 0.96 | 1.04 | 3.75 | 369 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| sourcemap.ts | 3.51 | 3.66 | 3.40 | 0.96 | 1.03 | 3.70 | 360 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/destructuring.ts | 3.50 | 3.29 | 3.46 | 1.06 | 1.01 | 3.91 | 363 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/parenthesizerRules.ts | 3.47 | 3.47 | 3.59 | 1.00 | 0.97 | 3.98 | 347 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/declarations.ts | 3.47 | 3.73 | 3.43 | 0.93 | 1.01 | 3.94 | 356 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2019.ts | 3.47 | 3.38 | 3.25 | 1.03 | 1.07 | 3.86 | 350 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| tracing.ts | 3.46 | 3.44 | 3.84 | 1.01 | 0.90 | 3.86 | 361 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/nodeChildren.ts | 3.44 | 3.53 | 3.37 | 0.97 | 1.02 | 3.94 | 373 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2016.ts | 3.44 | 3.35 | 3.22 | 1.03 | 1.07 | 3.88 | 363 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| tsbuildPublic.ts | 3.44 | 3.25 | 3.12 | 1.06 | 1.10 | 3.93 | 371 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| watch.ts | 3.44 | 3.31 | 3.33 | 1.04 | 1.03 | 3.73 | 373 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/emitNode.ts | 3.42 | 3.37 | 3.41 | 1.01 | 1.00 | 3.82 | 365 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/taggedTemplate.ts | 3.42 | 3.28 | 3.24 | 1.04 | 1.06 | 3.77 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| emitter.ts | 3.41 | 3.37 | 3.61 | 1.01 | 0.94 | 3.86 | 364 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/generators.ts | 3.40 | 3.33 | 3.14 | 1.02 | 1.08 | 3.75 | 350 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/nodeTests.ts | 3.39 | 3.17 | 3.44 | 1.07 | 0.99 | 3.76 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| moduleNameResolver.ts | 3.39 | 3.28 | 3.30 | 1.03 | 1.03 | 3.75 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| binder.ts | 3.37 | 3.32 | 3.26 | 1.02 | 1.03 | 3.71 | 357 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/module/system.ts | 3.37 | 3.39 | 3.29 | 0.99 | 1.02 | 3.49 | 372 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| executeCommandLine.ts | 3.36 | 4.34 | 3.45 | 0.77 | 0.97 | 3.67 | 368 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/nodeFactory.ts | 3.36 | 3.33 | 3.31 | 1.01 | 1.02 | 3.61 | 370 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/utilities.ts | 3.36 | 3.46 | 3.54 | 0.97 | 0.95 | 3.55 | 391 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| builderState.ts | 3.34 | 3.21 | 3.21 | 1.04 | 1.04 | 3.62 | 347 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/classFields.ts | 3.32 | 3.93 | 3.30 | 0.84 | 1.01 | 3.58 | 362 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| core.ts | 3.31 | 3.28 | 3.38 | 1.01 | 0.98 | 3.53 | 358 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| symbolWalker.ts | 3.30 | 3.34 | 3.56 | 0.99 | 0.93 | 3.69 | 349 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/classThis.ts | 3.30 | 3.31 | 3.58 | 1.00 | 0.92 | 3.69 | 378 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/esDecorators.ts | 3.30 | 3.31 | 3.17 | 1.00 | 1.04 | 3.69 | 395 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/module/impliedNodeFormatDependent.ts | 3.30 | 3.60 | 3.41 | 0.92 | 0.97 | 3.64 | 351 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| sys.ts | 3.29 | 3.32 | 3.30 | 0.99 | 1.00 | 3.67 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| visitorPublic.ts | 3.29 | 3.20 | 3.55 | 1.03 | 0.93 | 3.59 | 366 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2020.ts | 3.28 | 3.17 | 3.05 | 1.03 | 1.08 | 3.73 | 378 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| debug.ts | 3.27 | 3.31 | 3.46 | 0.99 | 0.95 | 3.48 | 373 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| checker.ts | 3.26 | 3.24 | 3.31 | 1.01 | 0.98 | 3.73 | 353 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/esnext.ts | 3.26 | 3.27 | 3.33 | 1.00 | 0.98 | 3.68 | 352 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| watchPublic.ts | 3.26 | 3.39 | 3.34 | 0.96 | 0.98 | 3.66 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| _namespaces/ts.moduleSpecifiers.ts | 3.25 | 3.31 | 3.23 | 0.98 | 1.01 | 3.58 | 374 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/declarations/diagnostics.ts | 3.25 | 3.58 | 3.36 | 0.91 | 0.97 | 3.43 | 392 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2017.ts | 3.25 | 3.31 | 3.43 | 0.98 | 0.95 | 3.46 | 372 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| commandLineParser.ts | 3.24 | 3.56 | 3.27 | 0.91 | 0.99 | 3.38 | 376 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2015.ts | 3.24 | 3.30 | 3.43 | 0.98 | 0.94 | 3.63 | 362 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| expressionToTypeNode.ts | 3.22 | 3.40 | 3.39 | 0.95 | 0.95 | 3.44 | 370 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/utilities.ts | 3.22 | 3.25 | 3.31 | 0.99 | 0.97 | 3.37 | 386 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| _namespaces/ts.ts | 3.21 | 3.15 | 3.21 | 1.02 | 1.00 | 3.34 | 373 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/utilitiesPublic.ts | 3.21 | 3.64 | 3.46 | 0.88 | 0.93 | 3.58 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/legacyDecorators.ts | 3.21 | 3.15 | 3.20 | 1.02 | 1.00 | 3.57 | 350 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| builder.ts | 3.20 | 3.33 | 3.38 | 0.96 | 0.95 | 3.42 | 372 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| builderPublic.ts | 3.20 | 3.33 | 3.22 | 0.96 | 0.99 | 3.48 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| factory/baseNodeFactory.ts | 3.19 | 3.50 | 3.42 | 0.91 | 0.93 | 3.56 | 367 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2018.ts | 3.19 | 3.28 | 3.47 | 0.97 | 0.92 | 3.60 | 354 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/es2021.ts | 3.19 | 3.25 | 3.10 | 0.98 | 1.03 | 3.57 | 392 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/ts.ts | 3.19 | 3.29 | 3.21 | 0.97 | 0.99 | 3.44 | 382 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/typeSerializer.ts | 3.19 | 3.21 | 3.17 | 0.99 | 1.01 | 3.51 | 349 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| diagnosticInformationMap.generated.ts | 3.18 | 3.37 | 3.41 | 0.94 | 0.93 | 3.56 | 351 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| builderStatePublic.ts | 3.17 | 3.34 | 3.18 | 0.95 | 1.00 | 3.58 | 348 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| utilities.ts | 3.17 | 3.30 | 3.29 | 0.96 | 0.96 | 3.56 | 357 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| _namespaces/ts.performance.ts | 3.16 | 3.14 | 3.31 | 1.01 | 0.95 | 3.46 | 358 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| utilitiesPublic.ts | 3.16 | 3.46 | 3.67 | 0.91 | 0.86 | 3.57 | 353 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/jsx.ts | 3.13 | 3.29 | 3.32 | 0.95 | 0.94 | 3.55 | 347 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/namedEvaluation.ts | 3.13 | 3.36 | 3.17 | 0.93 | 0.99 | 3.46 | 349 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| tsbuild.ts | 3.10 | 3.59 | 3.35 | 0.86 | 0.93 | 3.44 | 353 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| transformers/module/esnextAnd2015.ts | 3.09 | 3.44 | 3.16 | 0.90 | 0.98 | 3.29 | 368 | 1/1/1 | `checker.ts:12034:30: error TS18048: 'file' is possibly 'undefined'.` |
| corePublic.ts | 0.05 | 0.06 | 0.06 | 0.83 | 0.83 | 0.05 | 32 | 0/0/0 | `(none)` |
| hostErrors.ts | 0.05 | 0.05 | 0.05 | 1.00 | 1.00 | 0.05 | 32 | 0/0/0 | `(none)` |
| **total** | **259.4** | **261.9** | **259.1** | **0.990** | **1.001** | | | | |

## Profiles

There is no CPU-profile hook in `cmd/adamic`, `internal/load` or `internal/lower` (no `pprof` anywhere
in them). A throwaway test in main's `internal/lower` (never committed, deleted after) ran
`load.Load([]string{path})` and then `lower.Lower` when loading succeeded. It was compiled once with
`go test -c` and run as `lower.test -test.run '^TestLoweringProfileThrowaway$' -test.cpuprofile cpu.out`,
then `go tool pprof -top -cum`.

| file | load | lower |
|---|--:|---|
| watchUtilities.ts | 3.40 s, checker error | not reached |
| performanceCore.ts | 3.66 s, checker error | not reached |
| program.ts | 3.94 s, checker error | not reached |
| transformer.ts | 3.98 s, checker error | not reached |
| performance.ts | 3.14 s, checker error | not reached |

Shares, from `go tool pprof -top -nodefraction=0` (every node kept), on all five profiles:

| pass | where | share |
|---|---|--:|
| `predicateVerifier` | `internal/lower/predicates_proof.go` | 0 samples |
| `GetTypeAtLocation` (any caller) | checker | 0 samples |
| false-direction walk, `predicateFalseFlow` | `internal/lower/predicate_false_flow.go` | 0 samples |
| escape scan, `predicateFalseFlow.escapes`, `escapingReference` | same | 0 samples |
| anything in `internal/lower` but the test itself | | 0 samples |

The top 25 by cumulative time for program.ts (the others are in [profiles/](lowering-time-tsc/profiles)):

```
File: lower.test
Build ID: 0f91fd98cffd6d244c4c62207d8ecb2ca39fd5aa
Type: cpu
Time: 2026-10-27 23:38:06 UTC
Duration: 3.95s, Total samples = 4.56s (115.58%)
Showing nodes accounting for 2.95s, 64.69% of 4.56s total
Dropped 611 nodes (cum <= 0.02s)
      flat  flat%   sum%        cum   cum%
     0.01s  0.22%  0.22%      2.72s 59.65%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkSourceElementWorker
     0.01s  0.22%  0.44%      2.71s 59.43%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkSourceElement
     0.02s  0.44%  0.88%      2.68s 58.77%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkSourceElements
         0     0%  0.88%      2.44s 53.51%  github.com/microsoft/TypeScript/tsc/internal/core.(*singleThreadedWorkGroup).RunAndWait
         0     0%  0.88%      2.31s 50.66%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkBlock
         0     0%  0.88%      2.17s 47.59%  github.com/system-inc/adamic/internal/load.load
     0.02s  0.44%  1.32%      2.15s 47.15%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkExpressionWorker
     0.01s  0.22%  1.54%      2.14s 46.93%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkExpressionEx
         0     0%  1.54%      2.09s 45.83%  github.com/system-inc/adamic/internal/load.Load (inline)
         0     0%  1.54%      2.09s 45.83%  github.com/system-inc/adamic/internal/lower.TestLoweringProfileThrowaway
         0     0%  1.54%      2.08s 45.61%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkSourceFile
         0     0%  1.54%      2.03s 44.52%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkFunctionOrMethodDeclaration
         0     0%  1.54%      2.03s 44.52%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getDiagnostics
         0     0%  1.54%      1.96s 42.98%  testing.tRunner
         0     0%  1.54%      1.95s 42.76%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).checkFunctionDeclaration
         0     0%  1.54%      1.93s 42.32%  github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).GetDiagnostics (inline)
         0     0%  1.54%      1.93s 42.32%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*Program).getBindAndCheckDiagnosticsWithChecker
         0     0%  1.54%      1.90s 41.67%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*Program).getSemanticDiagnosticsWithChecker
         0     0%  1.54%      1.84s 40.35%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*Program).collectCheckerDiagnosticsFromFiles.func1
     0.01s  0.22%  1.75%      1.79s 39.25%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*checkerPool).forEachCheckerGroupDo.func1
         0     0%  1.75%      1.72s 37.72%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*checkerPool).forEachCheckerGroupDo
         0     0%  1.75%      1.69s 37.06%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*Program).collectCheckerDiagnosticsFromFiles
         0     0%  1.75%      1.66s 36.40%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*Program).collectCheckerDiagnostics
         0     0%  1.75%      1.66s 36.40%  github.com/system-inc/adamic/internal/load.(*Program).diagnostics
         0     0%  1.75%      1.63s 35.75%  github.com/microsoft/TypeScript/tsc/internal/compiler.(*Program).GetSemanticDiagnostics
```

## What this doesn't cover

- `adamic c` doesn't lower the 77 checker-rejected files, so the predicate passes' cost on tsc stays
  unmeasured. It needs the latent census overlay, which lowers past checker errors.
- The two checker-clean files are too small (0.05 s) to show a lowering cost.
- Timings are from one 4-core box, with a load average of about 1.3 during the run. The noise between
  rounds (up to 2.1 s on a 3.3 s run) is larger than any difference between the trees.

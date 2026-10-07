Built: a scratch-only per-unit census prototype; adapted TypeScript lowering remains blocked.
Base: ef3d907ecdc4c771b016f7d9c52372def057a340; TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.
Commands/results: apply.sh with adaptations 10 and 20; branch builds, checker runs and probe audit recorded below.
Mutants: planted overlay-only NotYet added exactly one finding to one.a and zero to two.a; misattribution and six report-recount mutants were caught.
Not completed: exhaustive latent corpus counts and feature-lowering deltas, two conflicted feature merges, and native/oracle correctness validation.

# Observed result

This unit is incomplete. The adapted corpus fails the unchanged Adamic checker,
so the tool does not pass it to lowering. Corpus NotYet/Refused counts and all
feature-lowering deltas are **unknown**, represented by JSON null. They are not
zero. Per-file and per-reason tables below count observed checker diagnostics,
not inferred latent blockers. REPORT.json includes all source hashes, every
file's blocked status, and null lowering counts for every configuration.

Apply ran successfully in a scratch main-based worktree with the base pipeline,
original census, and both adaptation branches merged. Adaptation 10 changed 72
files, replacing 3,719 lines; adaptation 20 changed 26 files, replacing 406 lines.
The total is 73 files and 4,125 replaced lines. The source population is 77
original compiler sources plus diagnosticInformationMap.generated.ts, for 78.
The same adapted bytes were used for every comparison. The two JSON inputs are
not source roots. Full apply output and the patch-set ledger are under data/.

The request names main and four feature branches but calls them four scratch
branches. With no clarification received, I attempted five individual
configurations so each named feature could have its own comparison. An extra
scratch preparation branch built the adapted tree. None was pushed.

| Configuration | Scratch result | Whole-project checker diagnostics | Latent counts | Feature lowering delta |
|---|---|---:|---|---|
| main | complete | 2165 | unknown | unknown |
| taste | complete | 2165 | unknown | unknown |
| flags | merge | unknown | unknown | unknown |
| namespaces | merge | unknown | unknown | unknown |
| nested | complete | 2165 | unknown | unknown |

Main, taste, and nested configurations check all 78 roots together using their
production `load.Load`. The overlay does not edit `internal/load`, change its
options, suppress diagnostics, replace standard library types, or reuse the
older census's upstream-config counterfactual. Per-file checker counts assign
each diagnostic to its reported source location in that whole-project run;
they are not repeated per-entry runs. A file with no own checker diagnostic
still remains blocked by the rejected project.

The flag-enum and namespace octopus merges failed with conflicts in
internal/lower/class_inheritance.go, internal/lower/lower.go,
internal/lower/refusals.go, and internal/oracle/counts.md. The exact failed merge
logs are committed under data/. I did not resolve compiler semantic conflicts
for this measurement unit. Those configurations have no compiler binary and no
observed checker or lowering delta.

# Branch provenance

All configurations start from the base SHA above. Scratch merge SHAs import
existing feature work; this unit makes no committed edits to internal/.

| Configuration | Feature SHA | Scratch merge SHA | Never-pushed branch |
|---|---|---|---|
| main | baseline | fc3482f2605590e1aa8bc49b7b4fd356f3c25717 | scratch/latent-compare-main |
| taste | aa896b5d5ccc82210184fd01b8fe4d0ce0730a50 | f6fca9973fb2236f7b1ec3c0da5c23ae143a194f | scratch/latent-compare-taste |
| flags | f7d62772fa8e52fcfae754e047ceb66dba88b782 | merge failed | scratch/latent-compare-flags |
| namespaces | ce8a2acf14e420a9c82345236845a377cd4c7a50 | merge failed | scratch/latent-compare-namespaces |
| nested | b15216dabf65ffaa7152f6e64709b7b062ea01a9 | b17f58168a8af23001b8e30bb0eb082cbcfc251b | scratch/latent-compare-nested |

Dependency checkout: cohere 715ba94f3608a6500086b1076ce5cb7e51b836db;
typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. The conflict-free
scratch merges retain this dependency pin. Scratch builds share that initialized
checkout through a symlink. Initial Go builds failed on VCS stamping because Git
could not identify the symlink as an initialized submodule; retrying with
`-buildvcs=false` fixed the build without changing compiler semantics.

# Per reason and delta

These are observed TS-code buckets, including diagnostics outside compiler/
when the project imports them. Negative deltas would mean fewer checker
findings; they would still not establish a reduction in latent lowering errors.

| Reason | Main | Taste | Nested | Taste minus main | Nested minus main |
|---|---:|---:|---:|---:|---:|
| TS1294 | 180 | 180 | 180 | 0 | 0 |
| TS1484 | 1 | 1 | 1 | 0 | 0 |
| TS18046 | 8 | 8 | 8 | 0 | 0 |
| TS18048 | 380 | 380 | 380 | 0 | 0 |
| TS2304 | 6 | 6 | 6 | 0 | 0 |
| TS2307 | 1 | 1 | 1 | 0 | 0 |
| TS2320 | 6 | 6 | 6 | 0 | 0 |
| TS2322 | 119 | 119 | 119 | 0 | 0 |
| TS2339 | 28 | 28 | 28 | 0 | 0 |
| TS2345 | 738 | 738 | 738 | 0 | 0 |
| TS2375 | 15 | 15 | 15 | 0 | 0 |
| TS2379 | 15 | 15 | 15 | 0 | 0 |
| TS2412 | 6 | 6 | 6 | 0 | 0 |
| TS2420 | 1 | 1 | 1 | 0 | 0 |
| TS2430 | 10 | 10 | 10 | 0 | 0 |
| TS2488 | 11 | 11 | 11 | 0 | 0 |
| TS2532 | 225 | 225 | 225 | 0 | 0 |
| TS2538 | 7 | 7 | 7 | 0 | 0 |
| TS2556 | 2 | 2 | 2 | 0 | 0 |
| TS2591 | 54 | 54 | 54 | 0 | 0 |
| TS2684 | 2 | 2 | 2 | 0 | 0 |
| TS2722 | 2 | 2 | 2 | 0 | 0 |
| TS2740 | 1 | 1 | 1 | 0 | 0 |
| TS2769 | 10 | 10 | 10 | 0 | 0 |
| TS7006 | 1 | 1 | 1 | 0 | 0 |
| TS7029 | 83 | 83 | 83 | 0 | 0 |
| TS7030 | 252 | 252 | 252 | 0 | 0 |
| TS7031 | 1 | 1 | 1 | 0 | 0 |

# Per file

Each cell counts diagnostic locations in the whole-project checker run.
All 78 files have unknown NotYet/Refused counts in all five configurations.
Flag-enum and namespace per-file observations are unavailable after merge failure.
The generated file is included explicitly rather than silently omitted.

| File under src/compiler | Main checker | Taste checker | Nested checker |
|---|---:|---:|---:|
| _namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 |
| _namespaces/ts.performance.ts | 0 | 0 | 0 |
| _namespaces/ts.ts | 0 | 0 | 0 |
| binder.ts | 18 | 18 | 18 |
| builder.ts | 45 | 45 | 45 |
| builderPublic.ts | 0 | 0 | 0 |
| builderState.ts | 4 | 4 | 4 |
| builderStatePublic.ts | 0 | 0 | 0 |
| checker.ts | 990 | 990 | 990 |
| commandLineParser.ts | 31 | 31 | 31 |
| core.ts | 89 | 89 | 89 |
| corePublic.ts | 1 | 1 | 1 |
| debug.ts | 42 | 42 | 42 |
| diagnosticInformationMap.generated.ts | 1 | 1 | 1 |
| emitter.ts | 47 | 47 | 47 |
| executeCommandLine.ts | 8 | 8 | 8 |
| expressionToTypeNode.ts | 6 | 6 | 6 |
| factory/baseNodeFactory.ts | 0 | 0 | 0 |
| factory/emitHelpers.ts | 6 | 6 | 6 |
| factory/emitNode.ts | 5 | 5 | 5 |
| factory/nodeChildren.ts | 0 | 0 | 0 |
| factory/nodeConverters.ts | 0 | 0 | 0 |
| factory/nodeFactory.ts | 9 | 9 | 9 |
| factory/nodeTests.ts | 0 | 0 | 0 |
| factory/parenthesizerRules.ts | 1 | 1 | 1 |
| factory/utilities.ts | 29 | 29 | 29 |
| factory/utilitiesPublic.ts | 0 | 0 | 0 |
| moduleNameResolver.ts | 39 | 39 | 39 |
| moduleSpecifiers.ts | 11 | 11 | 11 |
| parser.ts | 58 | 58 | 58 |
| path.ts | 8 | 8 | 8 |
| performance.ts | 0 | 0 | 0 |
| performanceCore.ts | 2 | 2 | 2 |
| program.ts | 71 | 71 | 71 |
| programDiagnostics.ts | 6 | 6 | 6 |
| resolutionCache.ts | 22 | 22 | 22 |
| scanner.ts | 26 | 26 | 26 |
| semver.ts | 15 | 15 | 15 |
| sourcemap.ts | 8 | 8 | 8 |
| symbolWalker.ts | 0 | 0 | 0 |
| sys.ts | 61 | 61 | 61 |
| tracing.ts | 32 | 32 | 32 |
| transformer.ts | 11 | 11 | 11 |
| transformers/classFields.ts | 16 | 16 | 16 |
| transformers/classThis.ts | 1 | 1 | 1 |
| transformers/declarations/diagnostics.ts | 1 | 1 | 1 |
| transformers/declarations.ts | 4 | 4 | 4 |
| transformers/destructuring.ts | 14 | 14 | 14 |
| transformers/es2015.ts | 39 | 39 | 39 |
| transformers/es2016.ts | 0 | 0 | 0 |
| transformers/es2017.ts | 10 | 10 | 10 |
| transformers/es2018.ts | 6 | 6 | 6 |
| transformers/es2019.ts | 0 | 0 | 0 |
| transformers/es2020.ts | 11 | 11 | 11 |
| transformers/es2021.ts | 0 | 0 | 0 |
| transformers/esDecorators.ts | 6 | 6 | 6 |
| transformers/esnext.ts | 7 | 7 | 7 |
| transformers/generators.ts | 32 | 32 | 32 |
| transformers/jsx.ts | 5 | 5 | 5 |
| transformers/legacyDecorators.ts | 1 | 1 | 1 |
| transformers/module/esnextAnd2015.ts | 6 | 6 | 6 |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 |
| transformers/module/module.ts | 5 | 5 | 5 |
| transformers/module/system.ts | 6 | 6 | 6 |
| transformers/namedEvaluation.ts | 1 | 1 | 1 |
| transformers/taggedTemplate.ts | 1 | 1 | 1 |
| transformers/ts.ts | 8 | 8 | 8 |
| transformers/typeSerializer.ts | 5 | 5 | 5 |
| transformers/utilities.ts | 9 | 9 | 9 |
| tsbuild.ts | 1 | 1 | 1 |
| tsbuildPublic.ts | 29 | 29 | 29 |
| types.ts | 83 | 83 | 83 |
| utilities.ts | 112 | 112 | 112 |
| utilitiesPublic.ts | 23 | 23 | 23 |
| visitorPublic.ts | 1 | 1 | 1 |
| watch.ts | 8 | 8 | 8 |
| watchPublic.ts | 6 | 6 | 6 |
| watchUtilities.ts | 6 | 6 | 6 |

# Prototype and mutant evidence

The scratch overlay disables `lower.Lower`: it always returns nil IR with
`latent census: measurement only; no IR output`. There is no emitter import or
backend invocation. The overlay-only `Latent` API walks refusal nodes without
stopping, then attempts each top-level function/statement with fresh lowering
state, recording returned errors and proceeding to the next unit. Each finding
has its actual kind, location, reason, and complete diagnostic text.

`audit.py` generates two scratch .a files. It observes both failing named-nested
functions, including the later function after the earlier failure, and both
non-null assertion refusal sites in another function. The injected NotYet is at
one.a:2:1 in target, in the overlay only: the selected file's finding count rises
by exactly one, while two.a's entire record remains identical. A second mutant
places the extra finding in the other file and the equality assertion rejects
it. The same audit also calls disabled Lower and checks its exact failure and
nil IR. These probes test bookkeeping; they are not TypeScript-derived runtime
fixtures, native programs or Node comparisons. data/audit.log preserves results.
The mutant was not tested on the TypeScript corpus because the checker blocks it.

`audit_report.py` independently checks source hashes against the adapted bytes,
raw diagnostic reasons against the JSON totals, per-file location attribution,
file coverage, unknown lowering results, and checker deltas. Six separate
artifact mutants were run and caught: inflated checker total, changed source
hash, dropped file row, fabricated zero lowering delta, changed per-file count,
and changed checker delta. data/report-audit.log records every catch.

The prototype still stops at the first lowering error within each unit, and its
fresh state registers only that unit, not sibling globals/functions. Thus an
isolation error can reflect missing sibling registration rather than a genuine
production blocker. Generic declarations are attempted without invented
specializations. Final module-order and ownership passes are omitted. The
prototype is not an exhaustive lowerer and should not be used to rank production
blockers until these limitations and the checker prerequisite are addressed.

# Commands and validation

All command output was redirected to files and subsequently read.

```sh
bash cloud/setup.sh > /tmp/latent-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/tsc-latent-adapted > /tmp/latent-apply.log 2>&1
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent-adapted /tmp/latent-comparisons > /tmp/latent-comparisons.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/latent-final-overlay > /tmp/latent-final-overlay.log 2>&1
gofmt -w /tmp/latent-final-overlay/*.go
go build -overlay=/tmp/latent-final-overlay/overlay.json -o /tmp/latent-final-census ./stage3/census/latent/tool > /tmp/latent-final-build.log 2>&1
python3 stage3/census/latent/audit.py /tmp/latent-final-census > /tmp/latent-audit.log 2>&1
go vet -overlay=/tmp/latent-final-overlay/overlay.json ./stage3/census/latent/tool > /tmp/latent-vet.log 2>&1
python3 stage3/census/latent/summarize.py /tmp/latent-comparisons /tmp/tsc-latent-adapted > /tmp/latent-summary.log 2>&1
python3 stage3/census/latent/write_report.py > /tmp/latent-report.log 2>&1
python3 stage3/census/latent/audit_report.py /tmp/tsc-latent-adapted > /tmp/latent-report-audit.log 2>&1
```

The three conflict-free configurations were rebuilt with the final overlay
and `-buildvcs=false` after the initial VCS-stamping failure; their raw outputs
are data/main.jsonl.gz, data/taste.jsonl.gz and data/nested.jsonl.gz.
Scratch builds and runs return exit 0, but the raw checker status is rejected.
Probe audit, report audit, and overlay vet return exit 0. Python syntax compilation also returns 0.
Setup: Go 1.27.1 ready in 0s; clang 20.1.8, Node 24.19.0 and submodules ready
by 1s; cache warming 119s; total 119s. `nproc` is 5; cgroup quota is four CPUs.

I did not produce the requested exhaustive latent TypeScript counts or numeric
feature-lowering deltas: the adapted project remains checker-rejected, and two
feature merges conflict. I did not run TypeScript's full suite, Node/native
output comparisons, ownership validation, or the full uncached Adamic gate.
This work provides a tested partial measurement prototype and explicit blockers,
not evidence that native tsc is closer to correctness.

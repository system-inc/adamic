# Ambient declarations refusal, task #3bddhs8

Code: `2559de762981983afde1af4d1a03e4fc56788801` on `compiler/lower-ambient-declarations`.
Base: `141fec7702d4aed21429724b2afd7efcd980f277`, newest origin/main when the work was based and rechecked before commit.
Evidence is separate, on `evidence/compiler/lower-ambient-declarations`.
Code pushed green at 2026-10-19 07:08:11 UTC.

Program modules must define runtime values rather than have the checker believe an ambient declaration. Lower now refuses ambient values, functions, classes, enums, value-bearing namespaces, augmentations, and parsed triple-slash references. Type-only declarations remain legal. The existing program-module walk keeps the prelude, bundled library, and generated platform bindings trusted.

## Files and lines

- `internal/lower/ambient.go:12`: parsed reference directives and source locations; `:40`: ambient flags; `:46`: nested namespace value detection; `:61`: declaration and augmentation refusal.
- `internal/lower/refusals.go:37`: reference refusal beside the directive checks; `:168`: declaration refusal immediately after namespaceRefusal.
- `internal/lower/enum_initialization.go:12`, `internal/lower/enums.go:393`, `internal/lower/namespaces.go:140`: let ambient declarations reach their refusal instead of an earlier NotYet.
- `internal/lower/ambient_test.go:15`: defined/type-only control; `:34`: identifier-module AST; `:46`: defined imported module control.
- `internal/lower/enums_test.go:126`, `internal/lower/namespaces_test.go:10`: update former NotYet expectations.
- `internal/oracle/ambient_refused_test.go:13`: directory test, source Node behavior, complete diagnostics, and a-check headers.
- `internal/oracle/testdata/ambient_refused/`: 13 programs, each with a pinned `.refused`, plus the reference-only `.d.ts` helper.

## Reproduction and controls

On the same `declare let limit: number` arithmetic program, base main's native binary prints `1` and exits 0. Source Node through the oracle reports `ReferenceError: limit is not defined` and exits 70. The new compiler exits 1 with the pinned ambient-value refusal. See `main-let-*`, `source-let.*`, and `new-let.*`.

All 13 refused fixtures throw ReferenceError on source Node. The lib-reference and module-augmentation witnesses include a subsequent ambient value read so Node also exercises the missing-value hole while the directive or augmentation remains the first compiler refusal.

The lower controls accept `interface`, `type`, a type-only ambient namespace, and a relative imported module whose function and values are defined. No accepted oracle fixture was added.

The mutant prepends `return nil` in both `ambientRefusal` and `referenceRefusal`. All 13 fixture subtests fail their refusal or whole-diagnostic assertion. `declare var` reaches the old var refusal, which the whole diagnostic pin correctly rejects. No build failure or clang warning killed the mutant. The runner restores `ambient.go` in `finally`; the final unmutated oracle gate ran afterward, and the required package gates also passed unmutated. See `mutant.py`, `mutant-summary.log.gz`, and `mutant.jsonl.gz`.

## Validation

Toolchain: Linux, Go 1.27.1, clang 20.1.8, Node 24.19.0. Every test command wrote its complete output to a file, without head/tail pipelines.

```
source /workspace/adamic-tools/env.sh
gofmt -l cmd internal
  exit 0, no output (gofmt.log.gz)
go vet ./internal/lower ./internal/oracle
  exit 0, no output (vet.log.gz)
go test -count=1 -timeout 30m ./internal/lower ./internal/load ./internal/fresh ./internal/flow
ok  	github.com/system-inc/adamic/internal/lower	142.187s
ok  	github.com/system-inc/adamic/internal/load	8.172s
ok  	github.com/system-inc/adamic/internal/fresh	61.513s
ok  	github.com/system-inc/adamic/internal/flow	145.157s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle
ok  	github.com/system-inc/adamic/internal/oracle	576.152s
```

Counts changed: none. The complete oracle checked recorded counts. Counts were not regenerated because the accepted controls are lower tests, and there is no new accepted oracle row.

## Sweep

Both compilers were built from the recorded revisions: base main and the code commit. `adamic c` ran on all 3,694 git-tracked `.a` and `.ts` programs, excluding `cohere/`, every `node_modules/`, and `.d.ts`. This includes review inputs, but no file under review was edited. Per-command timeout: 120 seconds. Timeouts: 0. Full command results, exit codes and stderr: `sweep.jsonl.gz`. Manifest: `sweep-paths.txt`. Reproduction: `sweep.py main`, `sweep.py new`. The five strengthened witness inputs were refreshed on base main with `refresh-main-witnesses.py` before the branch sweep; their final observations replace the earlier rows.

30 programs acquire this refusal as their first stop: 17 pre-existing programs and 13 new fixtures. The complete list follows.

```
internal/oracle/testdata/ambient_refused/class.a
internal/oracle/testdata/ambient_refused/const.a
internal/oracle/testdata/ambient_refused/enum.a
internal/oracle/testdata/ambient_refused/function.a
internal/oracle/testdata/ambient_refused/global.a
internal/oracle/testdata/ambient_refused/let.a
internal/oracle/testdata/ambient_refused/let_copy.ts
internal/oracle/testdata/ambient_refused/module.a
internal/oracle/testdata/ambient_refused/namespace.a
internal/oracle/testdata/ambient_refused/reference_lib.a
internal/oracle/testdata/ambient_refused/reference_path.a
internal/oracle/testdata/ambient_refused/reference_types.a
internal/oracle/testdata/ambient_refused/var.a
stage3/adapt/65-temporary-node-builtins/evidence/previous-process-host-types.a
stage3/adapt/71-writable-views/language-questions/generic-storage.a
stage3/adapt/71-writable-views/language-questions/node-flags.a
stage3/adapt/71-writable-views/language-questions/text-range.a
stage3/adapt/75-optional-widening/coverage/reduced-union.a
stage3/drivers/parser/native-process-next-tick-read.a
stage3/drivers/tsc/corpus/002_conditionalEqualityTestingNullability/conditionalEqualityTestingNullability.ts
stage3/drivers/tsc/corpus/005_ambientEnumElementInitializer3/ambientEnumElementInitializer3.ts
stage3/drivers/tsc/corpus/010_nestedTypeVariableInfersLiteral/nestedTypeVariableInfersLiteral.ts
stage3/drivers/tsc/corpus/016_yieldStarContextualType/yieldStarContextualType.ts
stage3/drivers/tsc/corpus/030_asyncYieldStarContextualType/asyncYieldStarContextualType.ts
stage3/drivers/tsc/corpus/032_templateLiteralsAndDecoratorMetadata/templateLiteralsAndDecoratorMetadata.ts
stage3/drivers/tsc/corpus/039_narrowByClauseExpressionInSwitchTrue2/narrowByClauseExpressionInSwitchTrue2.ts
stage3/drivers/tsc/corpus/045_discriminantUsingEvaluatableTemplateExpression/discriminantUsingEvaluatableTemplateExpression.ts
stage3/drivers/tsc/corpus/048_uniqueSymbolAssignmentOnGlobalAugmentationSuceeds/uniqueSymbolAssignmentOnGlobalAugmentationSuceeds.ts
stage3/drivers/tsc/corpus/053_contextualTypeBasedOnIntersectionWithAnyInTheMix5/contextualTypeBasedOnIntersectionWithAnyInTheMix5.ts
stage3/fixtures/host/22_createHash_fallback.a
```

Old and new diagnostics for every changed program are in `sweep-new-refusals.json`.


The sweep adds no new acceptances. Nine review probes panic natively on both revisions with the same first diagnostic and exit 2; their differing stack addresses and checkout paths account for the other nine raw-stderr differences. They were left outside this refusal task. Their full paths are in `sweep-summary.json`, and both stack traces are retained in `sweep.jsonl.gz`.

The changed pre-existing consumers are probes, corpus programs, or fixtures, not Adamic's own tooling. In particular, `stage3/fixtures/host/22_createHash_fallback.a:23` is now refused for its ambient createSHA256Hash function. It was not edited. Runtime Library must re-measure host status.json after landing.

## Pinned fixture diagnostics

```text
internal/oracle/testdata/ambient_refused/class.a:2:1: Adamic 0.1 refuses an ambient class declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/const.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/enum.a:2:1: Adamic 0.1 refuses an ambient enum declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/function.a:2:1: Adamic 0.1 refuses an ambient function declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/global.a:2:1: Adamic 0.1 refuses a global augmentation; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/let.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/let_copy.ts:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/module.a:2:1: Adamic 0.1 refuses a module augmentation or ambient external module; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/namespace.a:2:1: Adamic 0.1 refuses an ambient namespace or module holding values; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/reference_lib.a:2:21: Adamic 0.1 refuses a triple-slash reference lib directive; remove the directive and import a module that defines its values; use Adamic's library or generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/reference_path.a:2:22: Adamic 0.1 refuses a triple-slash reference path directive; remove the directive and import a module that defines its values; use Adamic's library or generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/reference_types.a:2:23: Adamic 0.1 refuses a triple-slash reference types directive; remove the directive and import a module that defines its values; use Adamic's library or generated platform bindings for host values
```

```text
internal/oracle/testdata/ambient_refused/var.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values
```

## Scope choices and limitations

The pinned checker already rejects identifier `declare module Name` with TS1540, before Lower can run. A lower test supplies that AST keyword on a checked namespace to verify the same outermost ambient refusal; no loader or checker exception was added. The string-module augmentation fixture uses the existing 'adamic' module because unresolved literal 'x' is rejected first with TS2664. These independent checker stops are recorded in `identifier-module.stderr` and `unresolved-module.stderr`.

No accepted oracle fixture or counts regeneration was needed. The whole-repository ./... gate was not run; the requested packages, vet targets, formatting check, and full uncached oracle were run. No host fixture or host status.json was changed.

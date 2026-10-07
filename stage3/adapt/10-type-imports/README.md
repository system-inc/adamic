# 10: Inline type imports

`adapt.cjs` marks named imports diagnosed as TS1484 and named re-exports
from a module diagnosed as TS1205 with inline `type` specifiers. Adamic uses
`verbatimModuleSyntax`, so a name with no runtime value must explicitly say
it is a type. This uses the same diagnostic-driven decision as
`cmd/adamic-meter`'s `importAdaptations`, through stock npm TypeScript 6.0.3.
The AST verifies the declaration and specifier before each insertion.

Run on the current tree, including any earlier adaptations:

```sh
npm install --prefix /workspace/type-imports/npm --no-audit --no-fund typescript@6.0.3
export CENSUS_TYPESCRIPT=/workspace/type-imports/npm/node_modules/typescript/lib/typescript.js
node stage3/adapt/10-type-imports/adapt.cjs <tree>
```

Without `CENSUS_TYPESCRIPT`, the script resolves `typescript` from the target
checkout or this directory. It rejects every version except 6.0.3. The pinned
upstream build dependencies use 5.9.3, so they cannot supply this dependency.
The adaptation does not install packages, generate diagnostics, or build tsc.

The roots are the current `src/compiler/**/*.ts` files, including the generated
diagnostics source if present. The options match Adamic's strictness settings:
strict, exact optional properties, unchecked indexed access, implicit returns,
switch fallthrough, verbatim modules, erasable syntax, ES2024, ESNext modules,
bundler resolution, forced modules, no emit, no ambient type packages.
Other checker errors neither authorize edits nor stop this specific adaptation.
Parse errors stop it before any edits.

Only `type ` is inserted, at AST specifier starts, in descending offset order.
No source regex, printer, formatter, sorting of declarations, or line-ending
conversion is used. Even an import containing only types stays an ordinary
import with inline modifiers; it retains its module dependency under verbatim
emit. Side-effect imports, value specifiers, module names, declaration order,
comments, and all other source bytes stay intact. Aliased imports and re-exports
receive the modifier before the imported/exported name, not before its alias.

## Declined sites

There were **no declined sites** on the pinned compiler source. There were no
named re-export TS1205 findings there: its barrels use `export *`, which this
adaptation leaves intact. A separate probe exercises named type re-exports.

Diagnosed default/namespace import forms, local exports without a module source,
or findings outside `src/compiler` are reported with a reason and cause exit 1;
the script does not invent a declaration-level conversion for them. Existing
type specifiers and declarations, value imports, and star exports need no edit.
A value-bearing class/enum imported only for type uses is not a TS1484 decision;
this adaptation retains that value import, following the meter's decisions.

## Pinned source and diff

Input: https://github.com/microsoft/TypeScript.git, tag `v6.0.3`, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base is `ef3d907`, with census
`429c117` merged first on `codex/stage3-type-imports`. No TypeScript source is
committed here. `stage3/source.json`, `stage3/apply.sh`, and the shared baseline
oracle belong to the separate base worker and were not implemented here.

On the original 77-source population: **3,718 imports in 71 files**, zero exports,
zero declined sites. `git diff --shortstat` reports **71 files changed, 3,718
insertions(+), 3,718 deletions(-)**: 3,718 existing lines changed, each by five
inserted characters. An independent AST audit removes exactly the newly marked
specifiers' modifiers and recovers every original source byte.

Upstream's diagnostic generator adds one source with one more type-only import:
`DiagnosticMessage` in `diagnosticInformationMap.generated.ts`. On a prepared
checkout it is also marked; the original-source scoreboard remains 71 files and
3,718 lines. Build-generated files are not included in that scoreboard. Generate
before adapting when integrating with the shared stage-3 pipeline. Upstream's
ordinary build may regenerate this file; its upstream build configuration does
not require verbatim import annotations.

## Census proof

The merged census driver was built without a loader overlay or weakened options.
It attempted each original source and then all 77 roots together for both trees.
All 77 pristine source SHA256s match `stage3/census/data/files.json`. Its fresh
whole-program diagnostic counts exactly match `data/stock.jsonl.gz` at `429c117`.
Neither tree reaches lowering. Whole-program total: **6,569 -> 2,851**.

| Reason | Pristine | Adapted |
| --- | ---: | ---: |
| TS1294 | 180 | 180 |
| TS1484 | 3,718 | 0 |
| TS2304 | 6 | 6 |
| TS2307 | 2 | 2 |
| TS2322 | 120 | 120 |
| TS2339 | 28 | 28 |
| TS2345 | 735 | 735 |
| TS2366 | 1 | 1 |
| TS2375 | 76 | 76 |
| TS2379 | 31 | 31 |
| TS2412 | 617 | 617 |
| TS2420 | 1 | 1 |
| TS2488 | 11 | 11 |
| TS2532 | 224 | 224 |
| TS2538 | 7 | 7 |
| TS2556 | 2 | 2 |
| TS2591 | 54 | 54 |
| TS2684 | 2 | 2 |
| TS2722 | 2 | 2 |
| TS2724 | 17 | 17 |
| TS2740 | 1 | 1 |
| TS2769 | 10 | 10 |
| TS7006 | 1 | 1 |
| TS7029 | 83 | 83 |
| TS7030 | 251 | 251 |
| TS7031 | 1 | 1 |
| TS18046 | 8 | 8 |
| TS18048 | 380 | 380 |

Every other diagnostic bucket is unchanged. For each tree, the walk also
attempted three JSON files and returned one extension error each for
`diagnosticMessages.json`, `tsconfig.json`, and `diagnosticMessages.generated.json`.
The last is a build-generated JSON leftover in both snapshots, outside the
77-source population; none is a source root or affects the table above. The
generated TypeScript file was excluded from both census snapshots to reproduce
the original census population. Summed census durations were 181.650s pristine
and 182.241s adapted on this worker, not predictions of native compile time.

```sh
source /workspace/adamic-tools/env.sh
go build -o /workspace/type-imports/census ./stage3/census/tool > /tmp/type-imports-census-build.log 2>&1
/workspace/type-imports/census /workspace/type-imports/before/src/compiler /workspace/type-imports/before.jsonl > /tmp/type-imports-before-census.log 2>&1
/workspace/type-imports/census /workspace/type-imports/after/src/compiler /workspace/type-imports/after.jsonl > /tmp/type-imports-after-census.log 2>&1
```

`before/src` is the pristine source snapshot; `after/src` is the adapted snapshot.
The JSONL records preserve every per-entry diagnostic, plus the whole-program
record. `/tmp/type-imports-count-proof.log` records the exact count assertions.

## Upstream oracle

Separate pristine and adapted worktrees were built with upstream's own commands,
using upstream's lockfile dependencies. Each `npm run build` passed, including
both compiler build/typechecking and test harness build/typechecking. Their
ordinary build invoked `scripts/processDiagnosticMessages.mjs` itself. Nothing
was formatted or baseline-accepted. Then the same subset ran on each tree:

```sh
npm ci --prefix /workspace/type-imports/pristine --no-audit --no-fund > /tmp/type-imports-upstream-npm.log 2>&1
npm run build --prefix /workspace/type-imports/pristine > /tmp/type-imports-pristine-build.log 2>&1
npm run build --prefix /workspace/type-imports/adapted > /tmp/type-imports-adapted-build.log 2>&1
# Run this from each tree, redirecting to its respective log:
node node_modules/hereby/bin/hereby.js runtests --tests='(verbatimModuleSyntax|typeOnly|importType|exportType|importsNotUsedAsValues|preserveValueImports)' --light=false > /tmp/type-imports-adapted-tests.log 2>&1
```

The adapted tree shares the installed upstream dependencies through a symlink,
but builds its own compiler and test harness. Both runs printed **1,085 passing**
and exited 0, with no changed reference baselines or local mismatch artifacts.
The subset includes compiler/conformance, fourslash, and unit tests selected by
that exact expression, including tsserver type-only import chains. Pristine suite
log: `/tmp/type-imports-pristine-tests.log`; adapted suite log:
`/tmp/type-imports-adapted-tests.log`.

Upstream `src/harness/harnessIO.ts` compares actual encoded baseline text against
the checked-in reference exactly and throws on any difference. Passing baselines
are compared in memory; only mismatches are written into local directories.
Empty local directories are therefore expected and are not independent evidence
of how many baseline comparisons ran. The 1,085 figure is passing test checks,
not a claim that all 1,085 each write a baseline file.

## Idempotence and mutants

Running the adaptation a second time on `after` produced zero files/imports/
exports changed; SHA256 maps of all **744 files** were identical. A prepared tree,
after marking its additional generated import, likewise changed zero files on
its second run, with **82,833 files** identical, including built/test artifacts.
Hashes excluded Git metadata and the shared dependency directory. Adding the
generated file between adaptation runs changes the input; that preliminary run
correctly reported one new modifier and was not treated as idempotence evidence.
Logs: `/tmp/type-imports-raw-idempotence.log`,
`/tmp/type-imports-prepared-idempotence.log`,
`/tmp/type-imports-idempotence-proof.log`.

| Mutant | Named check and observed failure |
| --- | --- |
| Mark the value import `Debug` in `src/compiler/scanner.ts` as inline `type` | Upstream `npm run build:compiler` exited 1; its typechecking reported TS1361 at scanner's actual `Debug` value uses. Log: `/tmp/type-imports-value-mutant.log`. |
| Skip `CommentDirective` in the same file, removing its inline modifier | The census whole-program TS1484 count became **1**, violating the required zero. Every other bucket stayed unchanged. Logs: `/tmp/type-imports-skipped-census.log`, `/tmp/type-imports-skip-mutant-proof.log`; raw records: `/workspace/type-imports/skipped.jsonl`. |

Both mutants ran in separate external copies; neither contaminated the adapted
tree. The value mutant was caught by the **build**, not claimed as a suite kill.
The skipped-import mutant ran the same full census driver as the primary check.

A small external TypeScript probe verified aliased imports, an all-type ordinary
import, a mixed named type re-export, an astral Unicode comment before the import,
retained comments, and CRLF bytes against an exact expected result. It printed
2 imports, 1 export, 1 file, no declines. `/tmp/type-imports-probe.log` records the
adaptation result. `/tmp/type-imports-source-proof.log` records the independent
71-file byte-preservation audit.

Relevant Go checks passed:

```sh
go test ./cmd/adamic-meter ./stage3/census/tool -count=1 > /tmp/type-imports-packages.log 2>&1
```

Meter: `ok`, 4.095s; census package has no test files. `git diff --check` passed.
Toolchain setup log `/tmp/adamic-setup.log`: Go 0s; clang, Node, submodules 1s;
build cache warm 109s; total 109s; `nproc` 5; CPU quota 4. Go 1.27.1,
clang 20.1.8, Node 24.19.0. Setup succeeded.

## Limits

The complete TypeScript suite and Adamic uncached integration gate were not run.
The named re-export path has a direct probe, since this pinned corpus has no such
findings. Default/namespace type-import conversion is intentionally declined.
No native tsc, latent lowering blockers, future earlier-adaptation combinations,
shared apply pipeline, or shared baseline oracle is claimed by this unit. The
adapter reparses current text and never assumes pinned line numbers. All
verification caches, upstream sources, logs, and mutant copies are outside the
Adamic repository under `/workspace/type-imports` and `/tmp`.

# 10: Inline type imports

`adapt.cjs` marks named imports diagnosed as TS1484 and named re-exports
from a module diagnosed as TS1205 with inline `type` specifiers. Adamic uses
`verbatimModuleSyntax`, so a name with no runtime value must explicitly say
it is a type. This uses the same diagnostic-driven decision as
`cmd/adamic-meter`'s `importAdaptations`, through stock npm TypeScript 6.0.3.
The AST verifies the declaration and specifier before each insertion. The
generated diagnostics import is fixed at its owner: the output string in
`scripts/processDiagnosticMessages.mjs`, not the generated artifact.

Run on the current tree, including any earlier adaptations:

```sh
npm install --prefix /workspace/type-imports/npm --no-audit --no-fund typescript@6.0.3
export CENSUS_TYPESCRIPT=/workspace/type-imports/npm/node_modules/typescript/lib/typescript.js
node stage3/adapt/10-type-imports/adapt.cjs <tree>
```

Without `CENSUS_TYPESCRIPT`, the script uses `require("typescript")`, honoring
the stock compiler supplied through `NODE_PATH` by `stage3/apply.sh`. It rejects
every version except 6.0.3. The target checkout's upstream build dependencies
use 5.9.3 and are not searched for the adaptation dependency.
The adaptation does not install packages or build tsc. After changing the output
template, it runs the same upstream diagnostic generator as 00-setup, with the
same relative input path. It also regenerates a missing artifact or one still
diagnosed as TS1484/TS1205 when the template was already fixed. A clean second
run skips regeneration.

The roots are the current `src/compiler/**/*.ts` files, including the generated
diagnostics source if present for checking. That generated file is excluded from
direct edits; its generator template is adapted separately. The options match
Adamic's strictness settings:
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
oracle were implemented by the separate base worker. The follow-up below merges
and exercises that pipeline without changing its scripts.

On the original 77-source population: **3,718 imports in 71 files**, zero exports,
zero declined sites. `git diff --shortstat` reports **71 files changed, 3,718
insertions(+), 3,718 deletions(-)**: 3,718 existing lines changed, each by five
inserted characters. An independent AST audit removes exactly the newly marked
specifiers' modifiers and recovers every original source byte.

The current adaptation also changes **one line in one generator script**:
`DiagnosticMessage` in the output import template of
`scripts/processDiagnosticMessages.mjs`. The generated artifact is never directly
edited. The apply scoreboard is **73 files, 3,720 lines added, 3,720 lines
removed**: 71 compiler sources, their generator owner, and its regenerated
artifact. The artifact is correct before a build and on subsequent upstream
regeneration.

The generator is parsed as JavaScript. The adapter locates the `result` array in
`buildInfoFileOutput`, parses its single `./types.js` import string as TypeScript,
and resolves each imported name against the real exports of `compiler/types.ts`
using stock 6.0.3's checker. It marks only symbols with a type meaning and no value
meaning: `DiagnosticMessage`, while retaining the value `DiagnosticCategory`.
It requires exactly one matching template. Literal content must equal its raw
source interior, so escaped string encodings are refused rather than edited with
incorrect decoded offsets. The edit is one insertion at the AST specifier offset
inside that literal, preserving its original quoting, CRLF, and other bytes.
Neither script source nor template source is matched with a regular expression.
Missing, duplicated, malformed, or unresolved templates fail before any writes.
No diagnostic definitions, codes, messages, or generation logic change.

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

## Initial standalone idempotence and mutants

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

The initial proof above used a filtered TypeScript suite. The follow-up below
runs the complete default oracle suites. Adamic's uncached integration gate
was not run.
The named re-export path has a direct probe, since this pinned corpus has no such
findings. Default/namespace type-import conversion is intentionally declined.
No native tsc, latent lowering blockers, or future earlier-adaptation combinations
are claimed by this unit. The
adapter reparses current text and never assumes pinned line numbers. All
verification caches, upstream sources, logs, and mutant copies are outside the
Adamic repository under `/workspace/type-imports`, `/tmp`, and the base pipeline
cache at `~/.cache/adamic-stage3`.

## Initial base pipeline integration (superseded by the owner fix below)

Merged base pipeline commit `8728405` through merge `3fa6403`. Changed only the
adapter's dependency lookup to `require("typescript")` by default, honoring the
stock 6.0.3 installed and exposed through `NODE_PATH` by apply. The explicit
`CENSUS_TYPESCRIPT` override remains available for the earlier standalone proof.
`stage3/apply.sh`, `apply.py`, and the oracle scripts are unchanged.

The following commands ran strictly sequentially, with fresh apply and oracle
output directories and no oracle runner or test filters:

```sh
source /workspace/adamic-tools/env.sh
stage3/apply.sh /workspace/type-imports/pipeline-adapted > /tmp/type-imports-pipeline-apply.log 2>&1
stage3/oracle/run.sh /workspace/type-imports/pipeline-adapted /workspace/type-imports/pipeline-oracle > /tmp/type-imports-pipeline-oracle.log 2>&1
/workspace/type-imports/census /workspace/type-imports/pipeline-adapted/src/compiler /workspace/type-imports/pipeline-census.jsonl > /tmp/type-imports-pipeline-census.log 2>&1
```

Apply exited 0 and reported 72 files, 3,719 imports, zero exports, zero declined
sites. This includes the generated diagnostics import. The regenerated
`stage3/patch-set.md` row is:

```text
| 10-type-imports | 72 | 3719 | 3719 |
```

The total row is also 72 files, 3,719 lines added, 3,719 lines removed; 00-setup
remains zero. Those rows describe the tree immediately after apply.

`/workspace/type-imports/pipeline-oracle/report.json` reports:

```json
{
  "status": "pass",
  "runners": "all",
  "tests": null,
  "workers": 4,
  "counts": { "passing": 106367, "failing": 0, "pending": 0 },
  "baseline_diffs": [],
  "wall_seconds": 331.761
}
```

All phase exits were 0: install 2.274s, build 22.156s, tests 307.278s. The actual
npm test command was `npm test -- --workers=4 --lint=false --no-colors`;
package.json supplies `--light=false`. `baseline.diff` in that results directory
is **empty, 0 bytes**. No baseline was accepted or updated. No container OOM
or OOM-kill event was recorded during the oracle run.

The census on the **post-oracle tree** exited 0: 81 per-file attempts, 78 source
roots in the whole program, 2,834 whole-program diagnostics, **TS1484 count 1**.
Its single TS1484 is:

```text
src/compiler/diagnosticInformationMap.generated.ts:4:30: TS1484: 'DiagnosticMessage' is a type and must be imported using a type-only import when 'verbatimModuleSyntax' is enabled.
```

This was a pipeline regeneration finding in the previous version, not one of the
original 3,718 imports:
upstream's ordinary build reruns `scripts/processDiagnosticMessages.mjs`, whose
output includes `import { DiagnosticCategory, DiagnosticMessage } ...`. That
replaces the inline modifier added by apply to the generated source. The other
3,718 modifiers remain in place. The prepared population also includes the
generated source and installed upstream dependencies, so its total is not the
original unprepared 77-root population's 2,851.

No adaptation was reapplied between oracle and census, and the generator and
pipeline were not patched to hide this result. Thus the default oracle is green,
but **that previous post-build tree did not have zero TS1484 findings**. Keeping that tree
at zero requires preserving or reapplying the generated-file annotation after
upstream regeneration, or a separate generator adaptation. The earlier mutants
remain the independently caught checks documented above; this follow-up did not
rerun them. Lint, browser integration, ESLint rule tests, Adamic's uncached gate,
and native tsc remain outside this default oracle's coverage.

## Generator-owner fix: previous post-build proof

This earlier run fixed upstream build regeneration at its owner, but left a
stale artifact immediately after apply (resolved in the follow-up below). It edited 71 compiler sources plus `scripts/processDiagnosticMessages.mjs` and
**does not edit `diagnosticInformationMap.generated.ts` directly**. At apply's
completion, the existing artifact still contains the old import; upstream's
ordinary oracle build regenerates it from the newly annotated template.
The scoreboard counts the owner's one-line edit, not a derived artifact edit.

Ran these three steps in order, with fresh paths, one at a time:

```sh
source /workspace/adamic-tools/env.sh
stage3/apply.sh /workspace/type-imports/generator-adapted > /tmp/type-imports-generator-apply.log 2>&1
stage3/oracle/run.sh /workspace/type-imports/generator-adapted /workspace/type-imports/generator-oracle > /tmp/type-imports-generator-oracle.log 2>&1
/workspace/type-imports/census /workspace/type-imports/generator-adapted/src/compiler /workspace/type-imports/generator-census.jsonl > /tmp/type-imports-generator-census.log 2>&1
```

All three exited 0. Apply reported 72 files, 3,718 source imports, 1 generator
import, zero exports, and zero declined sites. The regenerated scoreboard row
is unchanged numerically from the earlier version, but the extra file is now
the generator owner:

```text
| 10-type-imports | 72 | 3719 | 3719 |
```

`/workspace/type-imports/generator-oracle/report.json` records `status: "pass"`,
`runners: "all"`, `tests: null`, 4 workers, **106,367 passing, 0 failing,
0 pending**, and `baseline_diffs: []`. `baseline.diff` in that directory is
**empty, 0 bytes**. Install/build/tests all exited 0, taking 2.207s, 21.329s,
and 306.494s respectively; total wall time was **330.085s**. The default suites
and light=false verification were used, with no test filter. The generated
import after build is `import { DiagnosticCategory, type DiagnosticMessage } ...`.

The census on that exact post-build tree recorded 81 per-file attempts and
78 whole-program roots: **2,833 diagnostics, TS1484 = 0**. Compared with the
previous post-build run, TS1484 fell 1 -> 0 and every other diagnostic bucket
was unchanged. No adaptation was reapplied between oracle and census.

After those three steps completed, reran adaptation through `NODE_PATH`.
It reported zero files/imports/exports/generator imports changed, no declines;
SHA256 maps of **82,834 files** in the post-build tree were identical. Git
metadata and dependency files were excluded. A byte audit proved the generator
script differs from the pinned script only by inserting `type ` once before
`DiagnosticMessage`. Logs: `/tmp/type-imports-generator-idempotence.log` and
`/tmp/type-imports-generator-idempotence-proof.log`.

A dedicated template-ownership mutant renamed both `buildInfoFileOutput` and
its call in an isolated copy. Adaptation exited 1 with `want one generated types
import template, got 0`, and all input hashes stayed unchanged. The renamed
script still ran and produced byte-identical diagnostic outputs using upstream's
relative input path: only the ownership-location guard caught this mutant.
An initial comparison passed an absolute input path and differed in the generated
provenance comment; the corrected relative-path comparison passed. Logs:
`/tmp/type-imports-generator-guard.log`,
`/tmp/type-imports-generator-guard-proof.log`, and
`/tmp/type-imports-generator-guard-runtime.log`. The earlier value-import and
skipped-import mutants remain documented above; they were not rerun here.

## Apply ordering fix: pre-build proof

Merged `origin/area/stage3` with `git merge --no-edit FETCH_HEAD` (fast-forward
to `634ef06`, no rebase). Setup generates diagnostics before adaptation 10 runs,
so changing only the template had left apply's output stale. Adaptation 10 now
runs `node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json`
from the checkout after editing the template. It uses the upstream owner rather
than editing generated source. An already-fixed template with a stale or missing
artifact also triggers generation; a clean second run does not. Child process
errors and nonzero exits fail the adaptation.

Ran sequentially on a fresh checkout, measuring before any build:

```sh
source /workspace/adamic-tools/env.sh
stage3/apply.sh /workspace/type-imports/ordering-adapted > /tmp/type-imports-ordering-apply.log 2>&1
go build -o /workspace/type-imports/ordering-census ./stage3/census/tool > /tmp/type-imports-ordering-census-build.log 2>&1
/workspace/type-imports/ordering-census /workspace/type-imports/ordering-adapted/src/compiler /workspace/type-imports/ordering-census.jsonl > /tmp/type-imports-ordering-census.log 2>&1
stage3/oracle/run.sh /workspace/type-imports/ordering-adapted /workspace/type-imports/ordering-oracle > /tmp/type-imports-ordering-oracle.log 2>&1
```

Apply exited 0: 3,718 source imports, one generator import, zero exports/declines,
`regenerated: true`. Direct writes remain 72 files; the scoreboard also measures
the owner's regenerated artifact: `| 10-type-imports | 73 | 3720 | 3720 |`.
The pre-build census exited 0: 81 individual attempts plus the 78-root whole
program, with **TS1484 = 0 in every attempt**, including the generated file's
individual entry. The whole-program total is 896 diagnostics. Other adaptations
merged from area/stage3 account for the reduction from the previous 2,833 total;
this run does not attribute their changes to adaptation 10.

The default oracle exited 1: **106,366 passing, 1 failing, 0 pending**; install
and build exited 0. Its only mismatch is `api/typescript.d.ts`, with a
**48,691-byte nonempty baseline.diff**. All 189 changed lines are exactly
` | undefined` additions to optional types, matching adaptation 20's documented
189 public snapshot edits. This integration run therefore does **not** meet
the requested empty-baseline gate. Adaptation 20 documents accepting a proved
API snapshot separately, which the shared apply command does not do. No
reference baselines were accepted by this follow-up.

The read-only adaptation 20 snapshot proof was also attempted with its emitted
373-owner ledger. It fails on inherited adaptation 40 brand changes (`any` to
`undefined`), so it cannot independently certify this combined pipeline without
accounting for adaptation 40. Log: `/tmp/type-imports-ordering-api-proof.log`.
The oracle's existing reference already contains those 28 brand changes; the
remaining observed oracle mismatch consists only of the 189 optional additions.

After the oracle finished, a second adaptation 10 run reported no edits and
`regenerated: false`; SHA256 maps of **82,836 files** (excluding Git metadata
and dependencies) were byte-identical. An isolated probe removed only the
artifact's `type DiagnosticMessage` marker while keeping the owner patched. The
previous adapter from merged HEAD exited 0 but failed the generated-artifact
byte-equality check. The new adapter reported `regenerated: true` and restored
the exact expected artifact bytes using upstream generation. Log:
`/tmp/type-imports-ordering-idempotence.log`. This regression check proves the
already-patched-owner recovery path, in addition to fresh apply's owner edit.

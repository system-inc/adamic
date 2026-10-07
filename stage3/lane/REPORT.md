# Landing lane, October 7

## Owner-proof follow-up on today's area tip

Current verdict: **FAIL, exactly ten unapproved method-to-property conversions**.
Merged `origin/area/stage3` at `06f89a045c877808bcce0d72687cc59c893645ed`
into the existing lane branch, without rebasing. Merge receipt:
`2a17522c6c5b0510319d5e9974f830b6d9b76439`. The upstream adaptation, apply,
oracle, API dependency and source-pin files match that area tip.

The generator now reads the owner proofs for 20, 30, 32, 33, 40, 70 and 75.
Adaptation 32 contributes three existing property-form undefined unions:
`AmdDependency.name`, `CommentRange.hasTrailingNewLine`, and
`BuilderProgramHost.createHash`. Its handoff and host proof data supply the
owner paths and exact before/after contracts. The pristine AST identifies the
last as an existing property despite the host ledger's public-method label.
The same AST classification excludes all ten original methods. Adaptation 30's
remaining proof and every adaptation 33 API projection record no additional
public owner; their internal handoffs contribute no API line. Proof hashes
are recorded in the regenerated manifest. Its totals are 189/0/3/0/28/1/1,
for **222 normalized declaration changes**.

Fresh full command:
`stage3/lane/run.sh /workspace/stage3-lane-area-06f89a04-handoffs`, logged to
`/tmp/stage3-lane-area-handoffs.log`. Lane exit 1; apply/install/build exit 0,
tests/oracle exit 1. Linux x64, Node v24.19.0, four workers, all runners:
**106,366 passing, one failing, zero pending**. The exact expected API failure
name and `baseline_diffs == ["api/typescript.d.ts"]` match. The diff is
**244 added, 203 removed**. All platform/count expectations are untouched.
Oracle wall 395.283s; tests 369.581s; lane wall 498.402s including fresh apply.

The ten remaining owner errors are exactly:

- `CompilerHost.createHash`, `getDefaultLibLocation`, `readDirectory`
- `ModuleResolutionHost.directoryExists`, `getDirectories`, `realpath`, `trace`
- `ProgramHost.createHash`, `getEnvironmentVariable`, `realpath`

`python3 -m unittest discover -s stage3/lane -p 'test_*.py' -v`:
**29 passed**, 14.228s. The new tests verify the three proved property owners
are generated, original methods are excluded, and a planted method-to-property
conversion fails the same CLI guard. All prior diff, count, wrong-failure,
platform and printer-form mutants still pass their kill assertions.
Manifest regeneration followed by cmp, shell/Node syntax checks and
`git diff --check` pass.

A separate **checker-only control passes** after restoring exactly ten methods
from the pinned pristine API in an isolated snapshot. It has 222 composed and
222 sanctioned changes, 194 added and 193 removed lines. Counts and oracle logs
were held fixed; no upstream tests were rerun for this control. This verifies
the API guard can turn green after those precise reversions, without claiming
the forthcoming adaptation revert has already passed its full oracle.
No adaptation was edited by this unit. The real revert and its landing remain
with integration. This worker did not run macOS or the native Adamic gate.

New evidence: `handoffs-area-report.json`, `handoffs-oracle-report.json`,
`handoffs-verdict.txt`, `handoffs-baseline.diff`, `handoffs-tests.log`, compressed
full apply/oracle logs, and `method-revert-control.json`. Historical evidence
below predates the owner's ruling and the generator's reading of 32's proofs.

## Initial run before the owner-proof follow-up

Built the platform-keyed lane and mechanically generated API guard from
`origin/area/stage3` at `c1938019cbbc3b531f5ca2dd60cc4adb0fcdb190`.
The measured area-tip verdict is **FAIL**: 13 declaration changes are outside
adaptations 20/40/70/75's recorded public API sanction.
The counts, exact single failure title, phase exits and baseline paths all match.
No additional API change was approved to turn this run green.

## Observation

Command: `stage3/lane/run.sh /workspace/stage3-lane-area-c1938019-final`,
with output in `/tmp/stage3-lane-area-final.log`. Lane exit 1.
Current checker replay: `python3 stage3/lane/check.py RESULTS`, also exit 1,
with exactly the same 13 API errors.

Linux x64, Node v24.19.0, four workers, all runners, no test filter:

| Measurement | Observed |
|---|---:|
| Passing | 106,366 |
| Failing | 1 |
| Pending | 0 |
| Baseline added lines | 244 |
| Baseline removed lines | 203 |
| Already proved reference declarations | 28 |
| Composed changed declarations | 232 |
| Sanctioned changed declarations | 219 |

Apply exit 0; install and build exit 0; tests and oracle exit 1.
The exact failure title is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
The only baseline path is `api/typescript.d.ts`.
Install took 2.846s, build 24.861s, tests 363.923s, oracle wall 391.713s.
Whole lane wall was 494.923s, including fresh apply.

The unsanctioned owners are:

- `AmdDependency.name`
- `BuilderProgramHost.createHash`
- `CommentRange.hasTrailingNewLine`
- `CompilerHost.createHash`, `getDefaultLibLocation`, `readDirectory`
- `ModuleResolutionHost.directoryExists`, `getDirectories`, `realpath`, `trace`
- `ProgramHost.createHash`, `getEnvironmentVariable`, `realpath`

Ten are optional method-to-property/indexed-access union conversions; three
add undefined to optional properties outside the recorded 189-owner proof.
Normalization removes printer layout but retains these AST changes. Their
presence is observed; approval would require additional owner proof and an
intentional sanction update, or the source adaptations must retain the prior
public contracts. That integration decision is outside this unit.

## Guard evidence

`python3 -m unittest discover -s stage3/lane -p 'test_*.py' -v`, logged to
`/tmp/stage3-lane-tests.log`: **27 tests passed**, 7.631s, exit 0.
The required mutants change one observation each:

| Planted mutant | Catcher |
|---|---|
| Extra unsanctioned API line | `test_unsanctioned_line`: unsanctioned API declaration |
| Missing sanctioned API line | `test_missing_sanctioned_line`: missing sanctioned API declaration |
| Passing count 106,367 on Linux | `test_wrong_passed_count`: test counts and reporter counts |
| A different single failure | `test_different_single_failure`: failed test names; counts and diff still match |

The first two and the wrong-title mutant produce exactly one guard error.
Additional mutants verify changed types, reference edits, documentation,
source pin, diff tampering, phase exits, pending counts, filters and Node version.
An unknown-platform runner probe exits 1 before creating apply.log or a tree.
Positive controls prove multiline `| { ... }['trace']` formatting, quote style
and redundant parentheses normalize equally, and already-proved reference
changes remain checked as part of the full composition.

Manifest regeneration through `generate-api.py` followed by `cmp` against the
checked-in JSON succeeds byte for byte. `bash -n stage3/lane/run.sh` and
`git diff --check` succeed. The 219 changes derive from the recorded mechanical
owner data and pinned source, independently of this failing full-run diff.
README.md describes normalization and proof provenance in detail.

Raw evidence is in `evidence/`: lane and oracle JSON, verdict, baseline diff,
checker replay, test output, setup output and compressed apply/oracle logs.
The adapted upstream checkout stays outside the repository.

## Setup and boundaries

`ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh` succeeded.
Timing lines: Go ready 0.022s, Node ready 0.022s, clang ready 0.179s,
markdown dependencies 1.102s, submodules 30.266s, Go build ready 194.014s,
cache warm 194.546s, total 194.582s. `nproc` 5, CPU quota four cores,
Go 1.27.1, clang 20.1.8, Node 24.19.0. Test binaries were deferred.
Every subsequent command sourced `/workspace/adamic-tools/env.sh`.

This worker did not run macOS; its expected count comes from the user's
measurement on this area tip. The full native Adamic gate was not run because
this unit changes Python, shell, Node tooling and JSON only; setup built the
Go packages and the complete upstream oracle ran. Integration owns workflow
wiring. No adaptation, shared oracle, compiler or bucket files changed.
Apply's generated patch-set.md was restored. Only codex/stage3-lane is pushed;
there is no main push, force-push, rebase or pull request.

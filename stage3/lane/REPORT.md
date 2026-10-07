# Landing lane, October 7

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

New work for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): a deletion-set replay. The audit and its defender wave found tests in __PACKAGE__ that never showed a catch of their own. Before any of them goes to its owner for deletion, you find out what the package loses if they all go at once. Nothing gets deleted here; you only report.

Before anything else: your workspace is warm and keeps earlier units' build caches. Run `df -h /tmp /workspace`; if either has under 15 GB free, delete earlier units' scratch and per-mutant cache directories under /tmp (never /workspace/adamic or your tools).

## The set

Package `__PACKAGE__`, at current origin/main (fetch it, detached, clean, submodules at their pins; skip setup if /workspace/adamic-tools/env.sh works). The candidates, all removed together:

__CANDIDATES__

Their evidence is on these branches (fetch each; the diffs are standalone, one mutant per file, under review/):
__BRANCHES__

## What to do

1. Gather the mutants. On those branches, take every mutant diff (`M*.diff` from the audit, `D*.diff` from the defenders) whose recorded failing rows include at least one candidate. The rows each mutant failed are in the branch's matrix or results file next to it. Write the list down before running anything: mutant, file, branch, the candidates it failed, the other rows it failed. A diff that doesn't apply to current main gets "stale" and is skipped, never rebased by hand.
2. Baseline: run the package once on clean main with the candidates skipped, `go test -json -count=1 -timeout 30m ./__PACKAGE_DIR__/ -skip '__SKIP__'`, with `ADAMIC_GATE_UNCACHED=1` and a fresh `ADAMIC_BUILD_CACHE_DIR`. It must be green. If it isn't because the code is red, stop and report. If it isn't because a test needs something your environment lacks (an external checkout like `ADAMIC_TYPESCRIPT_SOURCE`, a prepared file, a toolchain), don't stop: narrow every run, baseline and mutants alike, with `-run` to the tests that caught the gathered mutants in the evidence (the other rows each mutant failed), and say in the report which tests you couldn't run and why. A mutant is still caught if one of those narrowed tests fails.
3. Each mutant: apply it, run the same command (its own fresh build cache), record every failing top-level test, restore. Stop a mutant's run at its first failure if the package is slow, but always run it to completion when it passes.

Three things that are not a catch. A pin: a test that fails on any change to its input, right or wrong (it compares a hash, a count or a census of the inputs, e.g. "scan inputs changed: <hash>"), can't tell a bug from a fix; list it but never as the only catcher. A witness (a test whose job is to plant its own mutants and prove another check can fail: names ending in Mutants, MutantKilled, PlantedFailure, and the like) failing under a production mutant doesn't count; list it but never as the only catcher. And a run that panics stops the whole package at that test, so the tests after it were never asked: rerun that mutant with the panicking test added to the skip list, again until either a non-witness test fails cleanly or the package passes to completion. Report the panicking tests separately, and remember that a non-witness test that panics because of the mutant has caught it: list it in still_caught_by. You skip it only to see who else catches the mutant, never to discount it.

4. A mutant that the full package caught (per the evidence) but nothing fails once the candidates are skipped has lost its last catcher. The candidates it failed are the ones that guard it.

## Report

One short paragraph: what you ran, wall times, anything stale or broken. Then exactly one fenced json block:

```json
{"package": "__PACKAGE__", "main": "<sha>", "skipped": ["..."],
 "mutants": [{"mutant": "D2", "file_line": "x.go:12", "branch": "test-defend/...", "candidates_failed": ["TestA"], "still_caught_by": ["TestB"], "stale": false}],
 "keep": [{"test": "TestA", "because": "D2 loses its last catcher without it"}],
 "deletable": ["TestC"]}
```

`deletable` holds only candidates for which every mutant they failed is still caught by something outside the set. A candidate with no gathered mutant at all (it never failed one) goes in `deletable` with a note, since it guards nothing that was shown. Restore the tree when done. Push the matrix, the mutant list and the logs to `test-defend/deletion-set/__SLUG__` under review/test-defend/deletion-set/__SLUG__/, and nothing else.

# Apply table and default lane results

`codex/stage3-apply-table` merges area tip
`832a759439540977cd210ce5f0972186aa0113e8`, including the ten method reverts.
The implementation tested by the full lane is
`ecf08643d946a7f73fad98f7c8b2cbfb85f9c52f`.
The area did not contain the lane when this unit started, so the existing
`codex/stage3-lane` branch was merged before updating its runner.

Apply always writes `<out>/patch-set.md`. Only `--write-table` also writes the
checked-in `stage3/patch-set.md`. The lane copies the output table beside
`report.json` before starting the oracle and fails clearly if that copy fails.
No-argument lane invocation uses Python's `tempfile.mkdtemp`, prints the fresh
absolute path first, and prints its verdict last. Explicit existing results
paths remain refused. Lane exit codes and expected oracle failures are unchanged.

## Commands and observations

All test and pipeline output was redirected to log files. Complete compressed
logs, reports, the baseline diff and table are in `evidence/apply-table/`.

- `ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh`: exit 0.
  Submodules 28.765 seconds, Go build 200.387 seconds, deferred test binaries
  200.606 seconds, warm build cache 200.608 seconds, total 200.642 seconds.
  `nproc`: 5, CPU quota 4, Node 24.19.0, Go 1.27.1, clang 20.1.8.
- `python3 -m unittest discover -s stage3 -p test_apply.py -v`: exit 0,
  one integration test passed in 126.977 seconds. It commits a deliberately
  stale table in a fresh detached worktree, runs real apply without a flag,
  and asserts the entire Git status remains empty and the table unchanged.
- `bash stage3/apply.sh --write-table /workspace/stage3-apply-table-generated`:
  exit 0. Checked-in and output tables match byte for byte. Adaptation 75 is
  1 file, 17 added, 6 removed. Total is 78 files, 5120 added, 5094 removed.
  This is ten fewer added and removed lines than the earlier 78/5130/5104
  request because the newer area tip restores the ten methods.
- `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s stage3/lane -p 'test*.py' -v`:
  exit 0, 32 tests passed in 8.081 seconds. This includes a no-argument success
  path through the runner with command doubles and a real no-argument shell
  invocation checking unknown-platform failure, report creation and output order.
- `PYTHONDONTWRITEBYTECODE=1 bash stage3/lane/run.sh`, without arguments:
  exit 0, `PASS stage3 landing lane`, 474.972 seconds on Linux x64/Node 24.19.0.
  First output line `/tmp/adamic-gate/stage3-lane-hi6_f74a`, last output line
  `PASS stage3 landing lane`. Apply/install/build exit 0; oracle/tests exit 1
  for exactly `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
  Counts: 106366 passing, 1 failing, 0 pending. Baseline paths:
  `["api/typescript.d.ts"]`; all 222 composed declarations match sanctions.
  Git status after the complete lane is empty; all three table copies match.
- `bash -n stage3/apply.sh stage3/lane/run.sh` and `git diff --check`: pass.

## Mutants

- `STAGE3_APPLY_MUTANT=1` with the same real apply integration test commits
  `if True:` in place of `if args.write_table:` in its scratch checkout.
  Exit 1, killed by `test_ordinary_apply_leaves_git_clean`: observed
  ` M stage3/patch-set.md` and “ordinary apply dirtied the source checkout”.
- A disposable lane copy restores the required results argument. Exit 1:
  `test_no_arguments_creates_fresh_results` catches argparse's `SystemExit: 2`;
  `test_shell_entry_point_with_no_arguments` catches exit 2 instead of the
  expected reported platform failure's exit 1.
- A disposable lane copy reads the table instead of copying it. Exit 1:
  `test_table_is_preserved_beside_report` and
  `test_no_arguments_creates_fresh_results` catch the missing copied table.
- The remaining lane guards' planted-diff/count/failure mutants are also
  exercised by the 32-test run; their named catches remain in its log.

No macOS run or Adamic native/compiler gate was performed for this Python and
shell pipeline change. The full unfiltered TypeScript suite was run on Linux.
The `area-merge.sh` caller itself was not edited: its no-argument lane call is
now supported by the entry point.

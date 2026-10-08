Built: Linux setup restores root and recursive submodule workspace sums from an EXIT trap.
Base: 7394b7cbdccc0ce679084c13c274804e7e5f3874; proof checkout main ce0750f28ef3943057f1f852b3ae5d93e6c5d644.
Proof: full setup injected failure after all module preparation, returned 73, and left main and every submodule clean; normal setup returned 0 and was clean too.
Mutants: missing EXIT trap leaves an untracked sum and fails the clean-tree assertion; Git failure preserves files and original exit status, including status 0.
Coverage: 57 cloud setup tests passed with six opt-in integration skips; macOS's separate Python installer and the full compiler gate were not exercised.

Assumption: keep the existing restoration policy: tracked sums are restored from HEAD, only nonignored sums Git positively identifies as untracked are removed. Ignored files and newly staged tracked files are retained. A Git diff error is not permission to overwrite. HEAD is copied completely to a unique temporary file before replacement. Every cleanup operation checks its status explicitly, because an EXIT handler called inside a conditional cannot rely on Bash errexit. Restoration failures print a diagnostic to stderr and preserve setup's original exit code.

The test-only switch is `ADAMIC_SETUP_TEST_FAIL_AFTER_MODULES=1`. It fires after all parallel preparation children finish, prints `setup: injected test failure after module download`, and exits 73. No worker uses this switch normally.

Full cloud proof used an existing main worktree at `/workspace/main`, with an untracked root `go.work.sum` and a newline appended to the tracked `cohere/go.work.sum` to model the reported partial failure. Both trees started clean before this controlled mutation. The actual setup command was:

```bash
source /workspace/css-gate-tools/env.sh
ADAMIC_TOOLS=/workspace/css-gate-tools GOMODCACHE=/workspace/warm-modules-cache \
GOCACHE=/home/agent/.cache/go-build GOFLAGS=-trimpath \
ADAMIC_SETUP_REPOSITORY=/workspace/main ADAMIC_SETUP_TEST_FAIL_AFTER_MODULES=1 \
bash /workspace/adamic/cloud/setup.sh > /tmp/adamic-gate/setup-restore-failure.log 2>&1
```

It printed verification of all 22 module directories, then the exact injected failure message, and returned 73. `git -C /workspace/main status --porcelain --ignore-submodules=none` was empty. `git -C /workspace/main submodule foreach --recursive 'git status --porcelain'` printed only directory-entry banners, no status entries. The same setup invocation without the switch returned 0 and the same clean-tree results.

Tests:

```bash
python3 -m unittest discover -s cloud -p 'test_*.py' > /tmp/adamic-gate/setup-restore-all-tests.log 2>&1
bash -n cloud/setup.sh
git diff --check
```

Six focused tests use actual Git worktrees and a real submodule: success/failure untracked removal, tracked submodule restoration, ignored-file retention, Git-error diagnostics/status preservation, staged-file retention, and the missing-trap mutant. Full setup logs are retained at `/tmp/adamic-gate/setup-restore-failure.log` and `/tmp/adamic-gate/setup-restore-success.log`.

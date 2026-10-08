Built: merged current lint area into the existing port; no rebase and no shared source changes.
Commit: tested merge and upstream pins are in evidence/checker-landing/metadata.json; final branch SHA is reported once.
Checks: full lint package 29 top-level passes, 5 failures, 1 named skip; owned witnesses, all 76 rule mutants, registry, gofmt and vet pass.
Mutant: object-alias-message-changed caught by independent Go bytes on source Node, emitted JavaScript and native.
Stopped: full rule certification still requires shared fixed() to reject no-progress proposals instead of panicking.

The area merge retains the one-program-per-run checker, typed descriptors,
checker replay guards and every check from both histories. This rule is syntax
only and has no private checker or recorded answers to replace. Its messages,
options adapter, all upstream test prefixes, four owned witnesses and multiple
individual automatic edits are unchanged.

## Current evidence

Run source /workspace/adamic-tools/env.sh, then evidence/checker-landing/run.py
with the checkout root and a new output directory as its two arguments. The
receipt records every input variable, exact command, all load samples, wall time,
CPU count, top-level and inclusive event counts and the sole skip name. Complete
Go test JSON output and stderr are archived as events.jsonl.gz and stderr.log.gz.
Registry, formatting and vet logs are alongside them. Setup took 220.504 seconds,
with nproc 5, CPU quota four cores and memory 17.6 GB; its full timing log is saved.

The full enabled-input run completed in 1080.186 seconds: 29 pass / 5 fail / 1
skip at top level, or 109 pass / 5 fail / 1 skip including subtests. Load at start
was 5.136 / 5.002 / 3.925 and at the last sample 4.340 / 5.428 / 5.394.
TestOwnedWitnesses passed across unchanged Go, source Node, emitted JavaScript
and sanitized native. All 76 registered rule mutants passed their required
comparison detectors, including this rule's object-alias-message-changed.
No selected assertion or input was skipped to obtain these results.

## Shared failures and mandatory skip

All five failing tests end at the same shared panic: TestRulesAgree,
TestCompilerAndStage1Agree, TestNodeTableIsLinkOnly, TestShardsAgree and
TestProfileSnapshotsAgree. stage1/cohere/lint/lint.ts:175 fixed() executes
panic('nonprogressing fix'). Unchanged Go cohere internal/edit/apply.go:207
applyToText rejects the individual proposal with ReasonNoProgress and keeps
applying the other proposals. MULTIEDIT_REPORT.md and its archived minimal
reproducer give the exact no-semicolon Shape alias, manifest and outputs.
This branch preserves the zero-width edit in its findings; removing it would
make its diagnostic bytes wrong. No shared fix engine was edited or bypassed.

TestCheckerBridgeRefusalPending is the sole skip. Its shared precondition is
TSGoError in internal/load/prelude.d.ts; it awaits codex/tsgo-errors-as-values.
Its exact reason is in the event log. All benchmark, compiler-source and
profile-input checks ran with their inputs supplied. The full repository gate
was not run and complete upstream rule parity is not claimed.

The initial full attempt exhausted the rebuildable Go cache filesystem and
cached that build error within its process. Its separate initial-disk-failure
logs and receipt are retained as infrastructure evidence. Only old rebuildable
cache entries were pruned; the fresh final run above has the real source
verdict. Neither attempt's failures were hidden or reclassified as passes.

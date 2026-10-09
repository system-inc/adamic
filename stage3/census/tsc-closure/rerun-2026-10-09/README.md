Rerun in progress: latent census complete; full census processing checker.ts.
Base: 89ac4a8c1de0b02d95965be72f7f8bf1c92433d2; measured main: 5e33a17b186a8a2218d27b69b21e2de5acc5b750.
Before: 5,430,761 / 10,010,532 bytes hidden (54.250473%); after hidden share pending full census.
After latent: 810 NotYet, 4,070 Refused, 2 panic, 5 SkippedDependency; 81 source files, 10,013,332 bytes.
Known-answer fixture and both actual root-drop mutants pass; new-result byte-mask audit remains pending.

This partial checkpoint is pushed at the dispatcher's explicit request. It is not a completed measurement. No partial ledger is published as RESULT.json. The full run has a 3,600-second limit and is checked every minute. The remaining census, arithmetic, byte-mask audit and final report are estimated at 30 to 90 minutes, with uncertainty concentrated in checker.ts.

Measurement tools remain pinned to f1502d130bc0b4440f29b93b76b7d18bff3f6a60; upstream TypeScript remains 050880ce59e30b356b686bd3144efe24f875ebc8. The source-only denominator convention is unchanged. Main adaptations changed 25 source hashes and added 2,800 bytes, so this compares compiler plus adaptation changes. Stock TypeScript independently confirms the same 81-file source closure.

What broke and how it was adapted:

- The shallow checkout lacked the measurement pin; fetching that exact SHA restored preparation.
- Cold setup exceeded a 240-second bound. A retry and an adapter clone exhausted /tmp. The final setup passed with Go build concurrency two; the final unchanged adapter run used workspace disk.
- Main introduced the private derived IR cache Program.argumentFacts. The pinned snapshot copier rejected it. compatibility.py leaves that cache nil in copied snapshots, retaining all other copy guards. No production compiler file changes. Restoring the rejecting copier is caught by the known-answer fixture.
- Main added a-check metadata headers to both existing fixture files. The original runner's fixed byte answer no longer matched. The scratch runner strips those headers and restores byte-for-byte the original authored fixture, verified against commit 002aaaab.
- The original independent byte-mask audit assumed a compiler-only root. audit.py reads the full source root from the closure result while preserving the byte-mask arithmetic. A root-regression mutant is caught.
- Ten-minute census trials reached only 28 latent and 8 full files. They are retained as trials, not results. The unchanged runs were restarted with one-hour hard limits.

Final setup timing: node 0.021s, Go 0.022s, submodules 0.067s, markdown dependencies 0.074s, clang 0.159s, shared cache 0.452s, build 45.728s, warm cache 45.820s, done 45.844s. nproc: 5; CPU quota: 4. Full timing output is archived under evidence/.

Checks already passed: the two-source fixture's 54/183 hidden bytes, latent and full root-drop mutants, eight pinned hidden arithmetic tests, full continuation audit and its first-error-only and failed-state-retention mutants, cache compatibility mutant, byte-mask fixture audit and root mutant, exclusive reason-credit fixture and one-byte credit mutant, and baseline credits summing to 5,430,761. All test output was written to logs.

See reproduce.sh, compatibility.py, audit.py, compare.py and evidence/ for commands and completed evidence. No native execution, full gate, or completed after hidden share is claimed at this checkpoint.

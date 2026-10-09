Evidence for u067 at origin/main 6c60da091afddc9c2fe88b3a1067845b6dc79cb3.

Read REPORT.md for the deliverable, rows.json for exact commands and failure lines, matrix.json for all observations, and reached-functions.txt for the 489 reached lowering functions recorded before mutation.

Each diffs/M*.diff is a standalone production mutant without a selector. P01 and P02 are empty-answer probes; W01 is permitted witness comparison weakening. All 17 diffs applied to the starting commit and passed go vet, recorded in standalone-validation.json. Apply exactly one diff to the starting commit, source /workspace/adamic-tools/env.sh, unset PREDICATE_MISCOMPILE_BASELINE, set ADAMIC_GATE_UNCACHED=1 and a distinct ADAMIC_BUILD_CACHE_DIR, then replay commands from rows.json or matrix-commands.json with test output redirected to a log.

plan.py and switched-source.diff preserve the actual switched construction. matrix.py preserves the script actually run; matrix-replay.py corrects its false Go-panic detector. No production source changes remain on this branch.

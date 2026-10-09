Starting source commit: 12e77e8972a2e606cab6db05d84f428246a85339.

Production mutants: diffs/M1.diff through M4.diff. Each is independent, has no selector, applies to this starting commit and compiled with adamic build --sanitize. P1.diff is the independent empty-answer probe. W1.diff and W2.diff are the brief-authorized Go harness weakenings, not production mutants.

For standalone replay, apply one diff, set a fresh ADAMIC_BUILD_CACHE_DIR, and run the package or reached rows with go test -json -count=1 -timeout 90s. Whole-package and combined requested-slice baselines cook at 90 seconds; matrix.json records which rows were actually run per mutant. Do not infer results for other rows.

For the exact matrix selector, run-audit.py installs the scratch sources, reads /tmp/u085/selector at runtime, and restores production source in a finally block. It also performs witness overlays and standalone builds. Scratch source copies and the selector helper are stored as .txt, so they do not participate in ordinary builds. Replaying the full script requires the /tmp/u085/weak overlays constructed from W1/W2 diffs and the pinned original-library dependencies.

The input-family membership is in families.json. Code-under-test/oracle declarations and raw pre-mutation V8 coverage are included. Logs are intentionally committed despite the repository log ignore rule.

ImportedInputsRunFresh defended; ParallelCachedMatchesSerial not defended after three aimed attempts.
Main b8bcadb2c493173855f19d7e5c508b34f5eeb5b6. Full 44-row matrix, every run completed, no panic, skips, timeout, or bounded subset.
Source restored; no tests or oracles changed. Every standalone diff applies and passed go vet ./cmd/adamic-test262/.

Read code-and-oracle.md for code, oracle and coverage reasoning. Coverage profiles and differences are stored beside it. matrix.json records all passing top-level rows and exact commands. D1 uniquely fails ImportedInputsRunFresh, with all 43 other rows passing. D2 fails ParallelCachedMatchesSerial, CacheAtomicBytes and CompilerWorkerMatchesSubprocess. D3 and D4 pass all 44.

The three parity attempts:
D2 changes recorded stderr to nil, losing diagnostic information across the cache and worker boundary. The target detects report divergence, but both named other catchers detect this independently.
D3 drops long-program-first sorting, so parallel work is queued in corpus order instead. Parity still holds. No claim that removing this scheduling optimization produces wrong answers.
D4 changes jobs := e.jobs to jobs := 1, so requested parallel work runs on one worker. Every answer remains right; the target does not detect lost parallelism. Its name promises a comparison of parallel execution with serial execution, but its assertions do not independently verify that parallel execution occurred. It also has no cost or concurrency threshold. This finite defense does not justify deletion.

Imported input defense:
D1 drops only engine.attempt's dependent-program cache bypass. dependentProgram still recognizes imports/references, so CompilerCacheProgram passes. Source remains unchanged while input moves from before to after; the reused observation hides the change and reasons become identical. This is a real freshness regression, uniquely caught by the target.

Friction, timing, limits:
Warm env.sh worked, setup skipped. npm ci log retained. Initial disk free: /tmp 5.1GB, /workspace 13GB. Removed earlier u156-typescript, u060-oracle.test and adamic-gate scratch products. Some deliberately locked test files resisted deletion; the remainder is tiny. Recheck /tmp 8.8GB free, /workspace 13GB. The total /tmp filesystem is only 8.8GB, so its 15GB threshold cannot be met. No disk errors occurred during this defense; repository and tools were not deleted.
Audit uses REPORT.md rather than report.md; first attempted filename was absent, then the actual report, rows and matrix were read. Initial baseline passed in 42.912 binary seconds but skipped the opt-in measurement. Enabled clean baseline passed in 56.197 seconds; all mutation runs enabled it too. Coverage commands and wall times are in coverage-runs.json.
D1/D2/D3/D4 wall seconds, including vet: 60.32 / 54.15 / 67.85 / 57.30, total 239.62 seconds. All binary runs below 90 seconds. Fresh per-mutant ADAMIC_BUILD_CACHE_DIR paths were used. No compiler implementation or port rebuild was mutated.
Go coverage does not instrument separate subprocesses automatically; profiles represent the runner process and include executed block ranges, not precise machine-instruction reachability. Seven more top-level tests than the audit's 37 were included. No repo-wide uniqueness, race detector, or exhaustive mutant search was attempted. All undefended-name findings are above.

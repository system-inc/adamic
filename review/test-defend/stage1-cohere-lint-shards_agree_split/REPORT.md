Both target families catch production breaks their former subsumer does not catch in the completed matrix.
Package-wide uniqueness remains unproven: 6,748 current tests were not run under each mutant.
Keep both pending broader replay; evidence does not support the old mutual-subsumption relationship.

Starting origin/main: 157a43552015f41a79331949c2e82b6f8c7caaab. Read the fetched audit report, rows and replay notes before defense. Current discovery contains 6,760 top-level tests. Source and tests restored before publication.

## Code and oracle

See code-and-oracle.md, written before mutants. Production Linter.fixed and Parser are mutated. The unchanged Go cohere adapters decide exact output. No test, fixture, oracle or harness was changed.

## Coverage and differences

Per-family go test -coverpkg=./internal/lower -coverprofile profiles are supplementary compiler coverage. Go cannot cover the TypeScript port. Actual Node V8 coverage records port execution. Corrected coverage-source-lines.json maps transformed JavaScript offsets back to original TypeScript through Node source maps, with nested zero-count ranges removed. The two older exclusive-line mappings are explicitly superseded and support no claim. Suggestion-only mapped leads include main.ts:75-82 and suggestion serialization; kind-only leads include JSX parser execution. These leads are not package uniqueness evidence.

The decisive semantic difference is shared fixer code receiving different findings: the suggestion fixture carries a semicolon automatic replacement alongside suggestions; normal no-debugger findings carry empty deletion replacements. S1 and S2 break only the former in the completed matrix. The script-kind family includes genuine JSX and reparses fixed JSX using the original path. K1 and K2 break that behavior; the scalar mixed-fix fixture survives.

## Matrix and compilation

matrix.json enumerates every current top-level test, observed pass/fail and unknown. Each of four standalone diffs was applied to the starting source, built by the native and emitted-JavaScript products before checks, then restored. Each has its own ADAMIC_BUILD_CACHE_DIR. All standalone diffs pass git apply --check against the restored starting commit. build-times.json records each actual cold product build line. runner.py.txt contains the execution procedure; matrix-runs.json gives commands, exit status and wall seconds.

Completed matrix: suggestion family three executors, script-kind two shards plus corpus union, complete-serialization six executors. For S1/S2 all three suggestion executors fail and nine other top-level tests pass. For K1/K2 script-kind shard 000 fails and eleven other tests pass. All twelve selected tests completed for every mutant. K1/K2 fail while source Node parses JSX after successful native product compilation; no native execution failure is claimed. Pass lists appear per mutant in matrix.json. Empty-answer probes were not requested in this defense task.

## Limits and verdict

Both rows are cannot-judge for package uniqueness, not not-defended. Two aimed attempts each already separate them from the previous subsumer. No deletion recommendation is supported. The remaining 6,748 tests are unknown under each mutant. A full clean baseline times out at 90.569s without an assertion failure. A clean dedicated JSX-tree extension times out at 90.033s; a clean general rule-agreement extension times out at 90.043s. These timeout runs are not semantic reds and are never recorded as passes or mutant kills. Wider uniqueness needs replay with sufficient build/run capacity.

## Costs, ambiguity and owner findings

Warm toolchain setup skipped, nproc 5. npm ci ran before baseline. The pinned TypeScript corpus was absent and had to be fetched at 050880ce59e30b356b686bd3144efe24f875ebc8. The first suggestion coverage baseline cooked at 90.173s during cold Go-oracle preparation; its retry passed in 6.854s. Script-kind clean coverage passed in 66.119s. Complete-serialization clean baseline passed. Twelve completed mutant commands took 444.265 wall seconds total, including compilation; per-command timings and product builds are saved. Corpus fetch and npm install were not separately timed. Total session roughly 27 minutes.

The brief asks Go coverage of production code that is TypeScript, so V8 coverage and source maps were needed. Initial V8-to-source mapping treated transformed offsets as original offsets; corrected artifacts supersede it. Thousands of existing tests and cold registry products make whole-package 90-second baselines impractical; narrowing cannot establish all-package uniqueness. The requested verdict definition requires no other package row to fail, which bounded results cannot establish. Families mix execution checks and corpus-union construction checks; union passing does not prove parser correctness. The twin verdict is absent from the proposed JSON enum, but is not needed here: executors are inside each family, and the two target families are semantically different.

No name/assertion gap found for either family. Mixed automatic fix and suggestions are checked through exact output, and script-kind coverage includes actual JSX. The mixed-fix deadline also asserts a work budget; no answer-preserving performance mutation was attempted, and no negative cost verdict is claimed. Optional release/throughput test skipped in the full baseline; all twelve selected tests ran without skips. No tests deleted, rewritten or weakened. No unobserved package-wide pass or unique kill is claimed.

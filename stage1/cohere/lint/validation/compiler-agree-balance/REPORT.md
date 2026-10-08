# Balanced compiler agreement validation

The baseline uses `39856c192`'s exact lint_test.go through a Go overlay in the same physical checkout. `git diff 39856c192 --name-only` confirms that every other executable source is unchanged; other additions are this unit's evidence. This controls absolute corpus paths and warmed toolchain caches. Both runs use `go test ./stage1/cohere/lint -count=1 -v -json -timeout=90m`, a clean TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`, wasi SDK 27, and separate fresh directories each shared by the two profile variables. No concurrent benchmark ran. Package process wall uses the final Go JSON event minus the prelaunch metadata timestamp; the separate polling upper bound is retained in wall JSON. Go-reported package time is also recorded.

Five CPUs are visible (`nproc=5`); quota is `400000 100000` (four CPUs). Each wall JSON records loads before/after; load snapshots were taken every 15 seconds. Tables report active wall per top-level test, including its children and excluding time paused by `t.Parallel`; Go's own reported duration is beside it because it excludes parallel children in TestMutants.

| Run | Package process wall (s) | Compiler agreement (s) | Load before | Load after |
|---|---:|---:|---|---|
| before | 2271.377 | 681.156 | [0.02392578125, 0.47607421875, 1.27978515625] | [1.94677734375, 1.86669921875, 1.8603515625] |
| after | 2061.313 | 363.691 | [1.94677734375, 1.86669921875, 1.8603515625] | [1.93798828125, 1.78515625, 2.0380859375] |

Observed Go-reported package time fell from 2252.683s to 2048.682s; CLI wall fell by 210.064s. Compiler agreement fell by 317.465s. Increases elsewhere include profile snapshots (+64.783s), factory hooks (+38.150s), and mutants (+30.087s); these tests' code is unchanged. Parallel test durations overlap and must not be summed.

before: sampled one-minute load min/mean/max 0.024 / 1.674 / 6.228 (152 samples). CPU visibility and quota are unchanged before and after: 5 CPUs, 400000 100000.

after: sampled one-minute load min/mean/max 1.005 / 2.099 / 5.646 (137 samples). CPU visibility and quota are unchanged before and after: 5 CPUs, 400000 100000.

The seat's 1,452s and the earlier 2,918.94s are not controlled before/after measurements on this run. The earlier run's sampled one-minute load mean was 2.062 (max 6.64); that alone does not prove a load-only explanation. Its unchanged mutant suite took 451.80s and snapshot check 419.18s. These measurements establish an improvement against the same-box baseline, not the reason this box differs from the seat.

## Sanitized native shards

| Shard | Seconds |
|---|---:|
| 0/5 | 283.762 |
| 1/5 | 87.611 |
| 2/5 | 117.717 |
| 3/5 | 156.302 |
| 4/5 | 158.849 |

Longest/mean: 1.764; headroom below 600 seconds: 316.238s. Shard 0 contains only TypeScript src/compiler/checker.ts (3,151,774 bytes), so its time measures that file including process startup. The other shards carry 2,360,783–2,360,791 bytes each. Scheduling is descending file size, ties in original manifest order, assigned to lowest-byte shard with lowest index on ties. Local case headers map back to the original manifest IDs; missing/extra/duplicate refusal and byte comparison remain in shards.Run/Merge. The Go oracle still receives the original manifest unchanged.

The 1.3 times mean target is not reached: checker.ts alone is the longest shard. Moving files among the other four shards cannot split that file's work. This is an observed whole-file limit, not a claim that bytes perfectly predict cost: the equal-byte other shards have different running times. The longest command is nevertheless substantially below both the previous 439.40-second shard and the 600-second cap.

## Planted failures (scratch Go overlays, never committed)

- byte: `lint_test.go:454: sanitized native: /workspace/scratch/typescript-6.0.3/src/compiler/checker.ts: line 3172: port ".workspace/scratch/typescript-6.0.3/src/compiler/checker.ts:1:1", Go "/workspace/scratch/typescript-6.0.3/src/compiler/checker.ts:1:1"`
- drop: `lint_test.go:443: Node: /workspace/scratch/typescript-6.0.3/src/compiler/checker.ts: missing from every shard`

The scratch byte proof's native times were 351.687, 186.268, 101.141, 197.856, 196.969 seconds; its longest was 351.687s (248.313s below the cap). This fresh-build run was slower than the package run, so both measurements are retained. The line number in the byte failure is the comparison output line; the quoted finding also carries source location checker.ts:1:1.

## Before: every top-level test, largest first

| Test | Own active wall (s) | Go reported (s) |
|---|---:|---:|
| TestCompilerAndStage1Agree | 681.156 | 681.160 |
| TestShardsAgree | 286.590 | 286.590 |
| TestProfileSnapshotsAgree | 260.407 | 260.410 |
| TestMutants | 227.758 | 0.150 |
| TestDotARename | 148.547 | 148.550 |
| TestCompleteSuggestionSerialization | 106.008 | 106.010 |
| TestJsxLintTrees | 74.867 | 74.870 |
| TestEmittedJavaScriptMismatch | 70.938 | 70.940 |
| TestWitnessScriptKind | 70.442 | 70.440 |
| TestSuggestionAlongsideAutomaticFix | 65.755 | 65.750 |
| TestFactoryHooks | 59.279 | 59.280 |
| TestProfileCompilation | 55.335 | 55.340 |
| TestProfileArtifacts | 55.162 | 55.160 |
| TestNodeTableIsLinkOnly | 29.187 | 29.180 |
| TestRulesAgree | 28.721 | 28.720 |
| TestCountGuardMutant | 12.921 | 12.920 |
| TestDecorationOptionMutant | 12.915 | 9.770 |
| TestPositionIndexMutant | 9.768 | 9.660 |
| TestCommentFoldMutant | 9.657 | 9.430 |
| TestOptionAndComparatorGaps | 9.403 | 0.960 |
| TestLegacyMutants | 5.209 | 5.210 |
| TestDecodedOptionsAndMutant | 4.855 | 4.850 |
| TestOwnedWitnesses | 4.325 | 4.330 |
| TestRegistrationMutant | 3.849 | 3.850 |
| TestNestedOutsideModuleCopy | 0.795 | 0.800 |
| TestNestedConstructorGap | 0.258 | 0.260 |
| TestExecuteFailsOnStderrOtherThanModuleDownloads | 0.128 | 0.130 |
| TestJsxLintReleaseAndThroughput | 0.001 | 0.000 |
| TestCommandDiagnosticsDropOnlyModuleDownloads | 0.001 | 0.000 |
| TestThroughput | 0.000 | 0.000 |

## After: every top-level test, largest first

| Test | Own active wall (s) | Go reported (s) |
|---|---:|---:|
| TestCompilerAndStage1Agree | 363.691 | 363.690 |
| TestProfileSnapshotsAgree | 325.190 | 325.190 |
| TestShardsAgree | 283.064 | 283.060 |
| TestMutants | 257.845 | 0.050 |
| TestDotARename | 141.155 | 141.160 |
| TestFactoryHooks | 97.428 | 97.430 |
| TestCompleteSuggestionSerialization | 78.367 | 78.370 |
| TestEmittedJavaScriptMismatch | 69.771 | 69.770 |
| TestSuggestionAlongsideAutomaticFix | 67.890 | 67.890 |
| TestWitnessScriptKind | 67.199 | 67.200 |
| TestProfileCompilation | 62.571 | 62.580 |
| TestProfileArtifacts | 59.251 | 59.240 |
| TestJsxLintTrees | 56.583 | 56.580 |
| TestCountGuardMutant | 37.325 | 37.330 |
| TestDecorationOptionMutant | 37.316 | 31.740 |
| TestPositionIndexMutant | 31.699 | 31.520 |
| TestCommentFoldMutant | 31.481 | 30.970 |
| TestOptionAndComparatorGaps | 30.879 | 1.620 |
| TestRulesAgree | 27.886 | 27.890 |
| TestNodeTableIsLinkOnly | 22.360 | 22.360 |
| TestDecodedOptionsAndMutant | 9.009 | 9.010 |
| TestRegistrationMutant | 6.718 | 6.710 |
| TestOwnedWitnesses | 6.318 | 6.320 |
| TestLegacyMutants | 5.296 | 5.300 |
| TestNestedOutsideModuleCopy | 3.015 | 3.020 |
| TestNestedConstructorGap | 0.287 | 0.290 |
| TestExecuteFailsOnStderrOtherThanModuleDownloads | 0.107 | 0.110 |
| TestCommandDiagnosticsDropOnlyModuleDownloads | 0.001 | 0.000 |
| TestJsxLintReleaseAndThroughput | 0.000 | 0.000 |
| TestThroughput | 0.000 | 0.000 |

Only throughput benchmarks may skip; all correctness checks passed. Raw JSON events, output, wall measurements, and sampled loads are alongside this report.

Tested lint_test.go SHA-256: `4c8a539d6d7ef650b808748f4bb83063fe511117c0ddf5c1f157b56b61ce57c9`. `go vet ./stage1/cohere/lint` and the compile smoke check passed; their logs are retained.

All backend shard times, seconds:

| Shard | Node | Emitted JavaScript | Sanitized native |
|---|---:|---:|---:|
| 0/5 | 21.925 | 38.774 | 283.762 |
| 1/5 | 15.486 | 21.075 | 87.611 |
| 2/5 | 16.072 | 20.279 | 117.717 |
| 3/5 | 13.889 | 21.747 | 156.302 |
| 4/5 | 14.634 | 21.622 | 158.849 |

To reproduce each negative run, copy lint_test.go to a scratch file and supply it with a Go overlay mapping the original absolute path to that file. At the launcher's final `process.stdout.write(Buffer.from(merged, 'latin1'));`, replace the write in the byte proof with `const bytes = Buffer.from(merged, 'latin1'); if (configuration.Command !== 'node' && shardIndex === 0) bytes[bytes.indexOf(10) + 1] ^= 1; process.stdout.write(bytes);`. In the dropped-case proof, only when `configuration.Command === 'node' && shardIndex === 0`, set `const next = merged.indexOf('\ncase ', 1)` and write `next < 0 ? '' : merged.slice(next + 1)` as a Latin-1 buffer; otherwise write the original merged buffer. This removes exactly one case (on this five-CPU run, checker.ts is that shard's only case). Run `go test -overlay=<overlay.json> ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=30m` with the same pinned inputs and a fresh shared profile directory. Both negatives return status 1 for the recorded comparison/merge failure; both verify that the original source remains unchanged. No planted source is committed.

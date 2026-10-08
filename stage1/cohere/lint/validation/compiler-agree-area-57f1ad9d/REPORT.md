Merged `57f1ad9d6ab3d23974cfc29b6043fdedc84bcbe9` into `lint-rules/compiler-agree-shards` with a merge, not a rebase. Merge commit: `116e8d8b9b78848ee306cddc23ae61f71eb39ef9`; parents: `6644c7cc6fb2619892851858a04b815077927265` and `57f1ad9d6ab3d23974cfc29b6043fdedc84bcbe9`.

There were no conflicts, including in lint_test.go. The merged file is byte-identical to the area version outside TestCompilerAndStage1Agree, its sharding helpers, and the added sort import. All 33 area functions were audited: only TestCompilerAndStage1Agree differs. The checker compilation, typed-project setup, live native recording, Node and emitted JavaScript replay comparisons, expanded native mutant checks, bounded builds, and JSX canary equality check are retained.

Inputs: clean TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`; installed wasi SDK at `/workspace/adamic-tools/wasi-sdk`; WASI_SYSROOT at its share/wasi-sysroot; ADAMIC_LINT_BENCH=1; both profile variables point to the same fresh empty directory for each invocation. Source SHA-256 for lint_test.go: `ce83b45a0b5425d50eddb81e6d87e41d12a05d55fc1f2d2148a26188b503f874`.

The first full invocation used `go test ./stage1/cohere/lint -count=1 -v -json -timeout=90m`. It reached its outer package deadline in TestProfileSnapshotsAgree: Go package 5401.146s, CLI wall 5547.525872s. It had 23 top-level passes, zero individual test failures, one skip, and unfinished tests; the package failed from the deadline. All 75 registered mutants passed. This failed invocation is retained as gate.jsonl.gz, gate-wall.json, gate-load.json and gate-tests.md. The full rerun uses a 180-minute outer allowance; the backend ten-minute limits and every comparison remain unchanged.

The original native byte probe failed before its output alteration ran: sanitized native shard 0/5 hit spawnSync ETIMEDOUT at the 600s cap. Its sole input is `/workspace/scratch/typescript-6.0.3/src/compiler/checker.ts`. This is a native performance failure, not a proven byte-mutant catch. Go package 819.607s, CLI wall 879.894361s. Its full log and timing/load metadata are retained as byte-native.*. The requested byte and dropped-case probes are therefore run on a Node shard through the same shared comparator and completeness merge, using scratch Go overlays. No planted source is committed.

The initial successful compiler comparison held 872 files and 30,504,305 output bytes. Its native shard times were 487.250110, 249.117453, 164.355587, 273.583626 and 236.636804 seconds. The rerun also passed that comparison; native times were 436.008343, 225.947134, 262.146032, 228.660676 and 258.459117 seconds, leaving 163.991657s of headroom in that invocation. The separate native timeout shows that this headroom is not reliable across invocations. No cause is inferred from load alone.

Completed retry: **34 top-level passes, zero failures, one skip**; including subtests, **113 passes, zero failures, one skip**. The named skip is TestCheckerBridgeRefusalPending: checker_pending_test.go:49: awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer. This is an existing pending checker API check, not a missing gate input. Go package elapsed 3751.486s; command wall 3764.7883919690066s. Go vet passed.

nproc=5; cpu.max=400000 100000 (four CPU quota). Retry load averages (1/5/15 minutes): before [0.873046875, 1.9052734375, 2.4765625], after [2.40234375, 1.66015625, 2.08984375]. Full sampled load is retry-load.json. Every top-level test timing is in retry-tests.md.

Compiler comparison wall: 583.75s. Per-shard seconds:

| Backend | 0 | 1 | 2 | 3 | 4 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Node | 32.698 | 22.390 | 23.698 | 23.230 | 23.622 |
| emitted JavaScript | 97.257 | 36.540 | 72.864 | 80.195 | 43.427 |
| sanitized native | 436.008 | 225.947 | 262.146 | 228.661 | 258.459 |

Scratch Node output probes both failed as required:

- byte: `lint_test.go:503: Node: /workspace/scratch/typescript-6.0.3/src/compiler/checker.ts: line 3180: port "rkipped @typescript-eslint/no-unnecessary-boolean-literal-compare no program", Go "skipped @typescript-eslint/no-unnecessary-boolean-literal-compare no program"`
- drop: `lint_test.go:492: Node: /workspace/scratch/typescript-6.0.3/src/compiler/checker.ts: missing from every shard`

Both probes used Go overlays, fresh identical profile/snapshot directories, and all required inputs. Both source hashes and clean TypeScript checkout checks remained unchanged after execution. Procedures are node-proofs.py.txt, retry.py.txt and run.py.txt; raw logs and start/wall/load metadata cover successful and failed attempts. “skipped” inside a lint finding is protocol output, distinct from Go test skip events counted above.

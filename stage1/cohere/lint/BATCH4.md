# Syntax-only lint batch 4

Implementation commit: `7bd94a2f92cc17d12bb69bbd78bfe2c4ac4ca63d` on `codex/stage1-lint-batch4`.
Base: completed batch 2, `c4373c05259c7235cf1202d5cd4138c16d498caa`.

Twenty additional syntax-only families, selected by unqualified rule names from a through m. Each family has its own source file; each registration occupies one line in batch4_registry.ts. No compiler, parser, runtime or submodule source changes.

Scanner tip was fetched before each selection: `0090256e607c3f2de7d5b67cef680ec010f95c1d`. Batch 3 was initially unpublished, so all twenty VOLUME.md frequency selections were reserved. Final fetch found batch 3 at `fa9781c2c0911c52312002ba97ffd4b560d178ae`; its source registry and report were inspected and neither branch contains any selected family. Per-rule checks are recorded in [batch4_tip_checks.json](batch4_tip_checks.json).

## Families and mutants

| Rule | Captured upstream cases | Compiled mutant result |
| --- | ---: | --- |
| `default-case-last` | 38 | Node and native caught omission |
| `default-param-last` | 128 | Node and native caught omission |
| `for-direction` | 46 | Node and native caught omission |
| `guard-for-in` | 20 | Node and native caught omission |
| `max-classes-per-file` | 29 | Node and native caught omission |
| `max-depth` | 38 | Node and native caught omission |
| `max-lines` | 15 | Node and native caught omission |
| `max-nested-callbacks` | 36 | Node and native caught omission |
| `grouped-accessor-pairs` | 160 | Node and native caught omission |
| `@typescript-eslint/default-param-last` | 117 | Node and native caught omission |
| `@typescript-eslint/ban-tslint-comment` | 42 | Node and native caught omission |
| `@typescript-eslint/init-declarations` | 92 | Node and native caught omission |
| `base/boundary-no-global-container` | 13 | Node and native caught omission |
| `nexus/consistency-no-stuttering-name` | 17 | Node and native caught omission |
| `base/consistency-no-hand-built-declared-error` | 14 | Node and native caught omission |
| `base/consistency-require-pagination-argument-name` | 29 | Node and native caught omission |
| `base/correctness-require-orm-column-declare` | 22 | Node and native caught omission |
| `nexus/consistency-no-for-in` | 18 | Node and native caught omission |
| `nexus/consistency-no-screaming-snake-case` | 23 | Node and native caught omission |
| `nexus/consistency-no-utils-folder` | 8 | Node and native caught omission |

Total: 905 captured cases. Each mutant changes its own activation to `false && ctx.enabled(...)`, disabling that family for both selected and all-rule execution. Every mutant compiled and executed successfully on Node and ASan/UBSan native, then disagreed with the independent Go answer. Compilation failures, crashes and sanitizer errors were not accepted as catches. [mutants.log](batch4_evidence/mutants.log) records all twenty results.

The independent oracle runs pinned, unmodified Go Cohere rules, report formatting and iterative source fixing. Cohere commit: `715ba94f3608a6500086b1076ce5cb7e51b836db`. Node executes the same TypeScript port; it is a second execution platform, not an independent semantic oracle. Findings, IDs, messages, UTF-8 ranges, displayed positions, edit payloads and iterative fixed source are compared byte for byte. Filename preservation tests path-gated rules. Numeric negative options, default Go settings during decoded-option projection, complete parent indexing before source-file visitors, and same-position listener order are covered by controls.

## Commands and observed outputs

Commands run from repository root after `source /workspace/adamic-tools/env.sh`. Test output was redirected directly to each named log, never piped.

| Command | Result | Evidence |
| --- | --- | --- |
| `go test ./stage1/cohere/lint -run '^TestBatch4(Controls\|UpstreamAgree\|RecoveryCensus)$' -count=1 -v -timeout 10m` | PASS, 138.186s; 25 controls, 16,103 identical bytes; 905 upstream cases, 765,692 identical bytes | [validation.log](batch4_evidence/validation.log) |
| `go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout 10m` | PASS, 213.20s; 2,793 unique captured combinations plus generated controls; 1,705,812 identical bytes | [rules.log](batch4_evidence/rules.log) |
| `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3 go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout 15m` | PASS, 209.71s; 77 compiler and 133 stage1 files; 19,482,979 identical bytes | [compiler.log](batch4_evidence/compiler.log) |
| `go test ./stage1/cohere/lint -run '^TestBatch4Mutants$' -count=1 -parallel 3 -v -timeout 20m` | PASS, 419.818s; all twenty omissions caught on both platforms | [mutants.log](batch4_evidence/mutants.log) |
| `go vet ./...` | exit 0 | [vet.log](batch4_evidence/vet.log) |
| `go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v` | PASS, 20.924s | [external-oracle.log](batch4_evidence/external-oracle.log) |
| `gofmt -l stage1/cohere/lint` | empty output, exit 0 | [gofmt.log](batch4_evidence/gofmt.log) |
| `git diff --cached --check` | empty output, exit 0 | [diffcheck.log](batch4_evidence/diffcheck.log) |

TypeScript corpus: 6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Native semantic comparisons use ASan/UBSan with leak checking. An initial combined run exposed zeroed Go max-lines defaults when unrelated decoded settings were applied; default initialization was corrected, then both independent final corpus commands passed.

## Findings per second

`ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3 go test ./stage1/cohere/lint -run '^TestThroughput$' -count=1 -v -timeout 15m` passed in 145.949s. [throughput.log](batch4_evidence/throughput.log) contains five fresh-process rounds per platform and the release full-output comparison prerequisite.

All rounds counted 15,385 findings across 77 compiler files. Best elapsed times:

| Platform | Seconds | Findings/s |
| --- | ---: | ---: |
| Go Cohere | 1.486772 | 10,347.92 |
| Native release O2 | 9.143800 | 1,682.56 |
| Node | 3.858283 | 3,987.52 |

Native is slower than both Go and Node on this workload. Timing includes process startup, parsing, linting and counting, excluding display/fixing. A separate kind-predicate prototype still counted 15,385 but took 10.71 to 12.41 seconds; it did not improve performance and was discarded. Its [count-only probe log](batch4_evidence/kind-probe.log) is exploratory evidence, not semantic parity evidence. Production source remained unchanged.

## Toolchain and limits

`bash cloud/setup.sh` succeeded. [setup.log](batch4_evidence/setup.log) prints Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build-cache warm 33s, total 33s; `nproc` returned 5, with cgroup cpu.max `400000 100000` and 17.6 GB. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

901 new upstream cases have full output and fix comparison. Four unterminated-comment ban-tslint-comment cases are findings-only; the Go recovery census records their diagnostics. There are no new parser refusal cases in the final twenty. Nine inherited unsupported-recovery cases remain explicit refusal checks in the combined test. General JSX, top-level-await parser coverage, binder/type-checker rules, configuration validation, suppression and non-UTF-8 input are outside this unit. Full repository `go test ./...` and inherited mutant rebuild suites were not run; touched lint tests, the new twenty mutants, repository vet and the filtered external oracle were run.

The final pinned Cohere `--no-fix --no-cache` check over the twenty files, shared helpers/messages/registry, settings.ts and lint.ts exits 1 with 52 policy findings: 47 ambiguous-identifier and five caller-data-mutation findings for mutable context/output arrays. It is not a clean policy gate. [cohere.log](batch4_evidence/cohere.log) records the exact files and findings. No unchecked casts, unused-variable or shadow findings remain in that check.

# Batch 32 landing refresh

Implementation a8bc0fa77 and claim 59c6b8d82 were published before the final remote check observed a new integration tip. The worker then took origin/area/stage1-lint 2fbfe42155387f67f297a41c89d318e3b6cde310 through merge c19a971a710572d1865e55e83a21d0e8cbc7e3f6. Current origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b remains an ancestor. History was preserved and no main or area branch was pushed.

The new base adds four rule directories and their evidence: no-underscore-dangle, no-unsafe-negation, no-unsafe-optional-chaining and typescript-no-this-alias. All 315 changed upstream paths are retained in landing/identity.json. Compiler, runtime, command and oracle trees and pinned cohere are identical to the tested implementation; no owned helper source or common dependency changed. Retained proofs for prior helpers remain applicable.

Fresh uncached validation after taking the new base:

- ADAMIC_SLOT05_BATCH32_EVIDENCE=<landing directory> ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch32 -count=1 -v -timeout=20m: PASS, 32.157s. All 15573 Go/source Node/emitted JavaScript/sanitized native queries and all 16 compiling semantic mutants passed again. Every raw factual case, Go verdict and coverage file reproduced the initial proof byte-for-byte. Prior evidence remains intact in evidence/; new proof is in landing/.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v: PASS, 0.937s; seven actual probes, zero hits and seven misses.
- go vet ./... and gofmt -l cmd internal stage1/cohere/lint/helpers/slot05: exit zero, empty logs.
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint05-typescript-6.0.3 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestShardsAgree$' -count=1 -v -timeout=20m: PASS, 175.973s. 3777 unique upstream source/rule/options combinations; 3952 total rows with the required pinned TypeScript compiler files. Unsharded native output and counted output agree at 1, 2 and 5 shards, 33779511 output bytes. This proves shard agreement, not independent whole-rule findings parity. The pinned TypeScript v6.0.3 input is the retained checkout at 050880ce59e30b356b686bd3144efe24f875ebc8.

The first shared command used '^TestShardParity$', which selected no tests and printed [no tests to run]. It receives no validation credit. Its log is retained as empty-filter-uncredited.log.gz; the correct TestShardsAgree was then run with its required real compiler input. No required check was skipped, relaxed or deleted to obtain a pass.

All test output went directly to logs. The setup timing and nproc from REPORT.md remain applicable because the toolchain and compiler inputs did not change: 25.024s and five processors. Readiness remains 92 helpers, 437 occurrences, 77 consumers and 51 helper-ready candidates; these three helpers remove nine prerequisites across three Structure rules with none newly ready.

The full gate and its 17 required external correctness checks, full shared suite, independent Adamic AST integration, whole-rule diagnostics/fixes/suggestions and arbitrary invalid/cyclic adapter inputs remain outside this worker validation. See REPORT.md for exact contracts, every mutant and the initial corrected compiler refusal. These limits are unchanged.

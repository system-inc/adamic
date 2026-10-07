Built the complexity rule with classic/modified branching, nested function frames, defaults, optional chains, class fields and static blocks.
Claim: b293c569; implementation commit contains this report.
Comparison: 201 upstream cases plus 242 compiler/stage1 sources and one witness; all three runtimes identical to Go, 12,656,078 bytes.
Mutant: invert modified-switch counting; Node, emitted JavaScript and sanitized native execute successfully and byte comparison catches each.
Uncovered: shared default gate, strict raw-option decoding and names from computed expressions beyond direct literals.

Run `source /workspace/adamic-tools/env.sh` then `python3 stage1/cohere/lint/rules/complexity/validation/validate.py > /tmp/w05-fourth-complexity.log 2>&1`. The upstream `^TestComplexity` suite passes; every captured case is compared, including naming probes run through the actual Go complexity rule with Maximum 0. TypeScript compiler corpus is pinned at 050880ce in `/tmp/lint-wave1-05-typescript`; corpus.json records exact input hashes. Decoded settings are consumed, matching the shared harness convention.

The independent driver calls the actual rule, compares ordered findings, byte ranges, ids, messages and final source. Go rule bodies and fix behavior remain unchanged; scratch overlays only expose the adapters and capture successful tests. Build uses the native sanitizer mode; all successful comparison executions exit zero with empty stderr. Complexity has no fixes.

Throughput on 77 compiler files plus a 1,000-function stress source, three interleaved runs, best elapsed including process startup: 1,325 findings; native 601.18 findings/s (2.203993s), Node 628.89 (2.106883s), Go 2338.92 (0.566501s). Release native is timed separately from sanitized correctness. These are observations for this corpus, not an extrapolation.

Toolchain: `bash cloud/setup.sh` logged Go, clang, Node and submodules ready at 0s each; cache warming failed at `stage1/cohere/lint/profile_test.go:32` (cannot range over the `portFiles` function). Continued with `source /workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5.

Shared checks: `go test ./stage1/cohere/lint/helpers -count=1` passed, including helper mutants. Filtered `go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1` passed. Registry tests fail because the current generator requires owned TS mutant modules, rejecting the `.a` mutants in this batch. The lint package fails to compile at the profile test above. No shared files changed; no full gate claimed. Logs, corpus hashes and output hashes are under `validation/`. Source-only `.ts` registration wrappers remain for the local discovery protocol, as Ahra authorized. Helpers and independent drivers are `.a`.

Claim b293c569 was pushed before implementation. Selection explicitly fetched all origin branches (341 refs, 52 unique claim blobs); the helper-ready pool was exhausted. These were the first three available syntax inventory entries. No next claim was taken while finishing this batch.

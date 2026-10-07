Built: for-direction, registered by directory, with exact Go messages and spans; no repairs.
Commits: claim 45b8b656 was pushed before code; implementation 8334acce; evidence commit in git history.
Commands and outputs: independent Go parity on Node, emitted JavaScript and sanitized native; see the unit report.
Mutant: mutant.json changes a semantic judgment, with clean compilation and execution required before comparison credit.
Not covered: unmodified shared-harness integration, pinned CLI self-lint of .a, the full repository gate or arbitrary malformed inputs.

This directory owns the rule, message, registration descriptor, independent Go oracle adapter, raw source witness and semantic mutant. No shared source is edited. None of these three rules offers a fix or suggestion, so their complete output is represented by the existing finding model.

This continuation ports default-case-last, for-direction and guard-for-in. Prior claimed candidates and evidence were pushed through df9f0451 before selecting these. After fetching 330 origin refs, checking 52 distinct claim Markdown blobs recursively and checking origin/main ef3d907ecdc4c771b016f7d9c52372def057a340, all 46 helper-ready names were covered. These are the first remaining rules in the inventory's syntax-ready list. Claim 45b8b656 was committed and pushed before writing code.

Default-case-last examines the first default clause and reports its complete span when it is not last. Guard-for-in reproduces the five shallow structural exemptions without inspecting the guard's meaning. For-direction examines both comparison operands, counts every relevant assignment in comma expressions, requires exactly one modification, and reads literal/unary signs exactly as Go does. Optional for-loop clauses have no placeholder in the Adamic parser arena; the rule reads separator tokens to recover the condition and update slots. Static constant-binding resolution is absent from Go's rule, so the port deliberately stays silent on that shape too.

The three rules use the ordinary shared linter frontend and directory-generated dispatch for comparison. No custom finding renderer is needed here. Scratch overlays provide .a registry support from published codex/lint-harness-dot-a 2650ad59, the compatible profiling snapshot, and bounded comparison tests. Shared repository source stays untouched. The default branch still requires its shared profiling compile fix and .a loading changes before ordinary package tests can run. The previous three repair-rule candidates retain their documented complete-record frontend integration gap; they were not changed or claimed again.

Setup first failed because the published profiling source called copyPort and emittedNode, which the earlier bounded comparison source does not define. setup.log and parity-1.log retain the exact undefined-function failures. Selecting the compatible scratch profiling snapshot resolved that pairing issue. `GOFLAGS=-overlay=/tmp/lint-wave1-13-loops/overlay.json bash cloud/setup.sh` then succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 14s, total 14s. nproc=5, quota 400000/100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands wrote directly to logs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -overlay=/tmp/lint-wave1-13-loops/overlay.json ./stage1/cohere/lint -run '^TestWave13Loops' -count=1 -v -timeout 20m > /tmp/lint-wave1-13-loops/parity-2.log 2>&1
go test -overlay=/tmp/lint-wave1-13-loops/overlay.json ./stage1/cohere/lint -run '^TestWave13LoopsOwnedWitnesses$' -count=1 -v -timeout 10m > /tmp/lint-wave1-13-loops/owned-final.log 2>&1
go test -overlay=/tmp/lint-wave1-13-loops/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-13-loops/registry-final.log 2>&1
go vet -overlay=/tmp/lint-wave1-13-loops/overlay.json ./... > /tmp/lint-wave1-13-loops/vet-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint-wave1-13-loops/oracle-final.log 2>&1
```

Replay with `python3 evidence/reproduce.py --compiler /path/to/typescript-6.0.3`. The compiler pin is 050880ce59e30b356b686bd3144efe24f875ebc8. Captured fixtures execute the actual cohere Go test assertions before harvesting source/rule/options records. The independent Go oracle executes unmodified rule bodies. All compared backend runs must finish cleanly with no stderr; native parity and mutants use ASan and UBSan, while timings use a separately built release binary.

The dedicated owned-witness check confirms a positive Go finding for every witness, then selected and all-rule modes match on all three backends: six rows, 4,069 identical bytes, PASS in 18.003s. Registration passes in 0.050s, vet has an empty successful log, and the uncached input oracle passes all six fixtures in 1.140s with six probe misses and no cache hits.

Pinned cohere CLI command: `/tmp/lint-wave1-13/cohere --no-fix --lint --no-cache` followed by all six new .a modules. It exits 1 with `nothing to check` because .a is not accepted as a TypeScript or JavaScript input. self-lint.log retains that refusal. Loader/typechecker and both compiler backends do accept the modules, as shown by the native and emitted JavaScript builds. No full repository gate or clean pinned self-lint is claimed.

The corpus suite passes in 206.73s: 800 source/rule inputs, with zero parser gaps for these three rules. It includes 77 pinned compiler files and 154 stage1 .ts/.a files per rule, plus 38/46/20 unique upstream cases and one owned witness per rule. The capture process additionally logs an inherited return-void JSX gap outside this unit; it is not counted as covered. Identical serialized output totals 12,390,348 / 12,394,578 / 12,393,670 bytes for the three rules, 37,178,596 bytes overall.

Throughput is best of three interleaved whole-process runs, including startup, I/O and parsing. Go, release native and Node counts agree. The small number of findings means these values primarily measure processing a largely clean corpus rather than an isolated rule algorithm.

| Rule | Inputs | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: | ---: |
| default-case-last | 270 | 16 | 12.34 | 18.10 | 77.86 |
| for-direction | 278 | 22 | 16.64 | 27.51 | 102.93 |
| guard-for-in | 252 | 17 | 12.57 | 18.77 | 76.00 |

Mutants: default-case-last reverses the last-position judgment; for-direction treats decrement as increment; guard-for-in reverses the structural guard exemption. Each mutant must compile and exit zero with empty stderr on Node, sanitized native and emitted JavaScript before the Go output mismatch can be credited. The witness sources themselves are parsed rather than executed, including infinite-loop examples.

The unmodified package check remains blocked before tests: `go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1` exits 1 at `profile_test.go:32:23`, because portFiles is a function rather than a rangeable value. default-harness.log retains the exact compiler error. No shared file was changed to resolve this; the scratch overlay and published integration branch provide the bounded verification path.

Final corpus and mutant command: PASS in 251.269s. All three mutants are rejected only by output comparison on each of Node, sanitized native and emitted JavaScript; mutant suite PASS in 44.53s. The six-row selected/all owned-witness command passes separately. All final production .a files were unchanged between these runs and their implementation commits. Logs named setup.log and parity-1.log are the failed scratch profiling pairing; setup-final.log and parity-2.log are the corrected successful runs.

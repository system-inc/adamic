Built: wave 08 rebased onto current main b8fb957aa and re-green; no additional rules claimed.
Commits: tested rebased tip b8a0e5e588f585762e7fc98a869ec223b6300710; this report accompanies its evidence commit.
Checks: nine completed owned rule profiles, supported React subset/kernels, bridge ownership and filtered Node oracle pass.
Mutants: all nine rule mutants, Globals, two partial kernels, raw facts and bridge ownership checks catch their witnesses again.
Uncovered: shared production factory/dispatcher integration and full JSX/HIR/SSA React parity; full repository gate not run.

The previous published tip was 2a181f6c334f65bd684bdd267d4d37881e1ce421.
The clean rebase takes main's inherited static-field emission fix without changing
owned rule sources. A fresh compiler was built. Existing checker archives were
reused because main changed no bridge, parser, checker, or owned rule source.
Only codex/typeaware-wave-08 is pushed, with the previous published tip as the
explicit force-with-lease expectation. No main or area branch is pushed.

The default fetch refspec covered main only. An explicit all-heads fetch then
covered 580 origin refs and 33 distinct claim Markdown blobs. All 197 ranked
checker-dependent rules resolve as 25 already ported plus 172 named in claims.
All 25 base port paths were verified on the freshly fetched main/bridge refs.
There are zero unclaimed entries, including the zero-volume tail. No additional
claim was made. Full ref tips, claim texts and selection are committed in
validation-b8fb/claims.json.gz and selection.json.gz. The current harness remote
is lint-rules/harness at 8a537d93c; it has not been imported into this branch.
Existing rule.json declarations already use AST kind names. No new rules or
regular-expression translations were introduced by this landing unit.

Fresh compiler: go build -o /workspace/wave08-b8fb/adamic ./cmd/adamic.
Original gate: go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v
-timeout 30m with both frozen corpus manifests and ADAMIC_TYPESCRIPT_SOURCE set.
PASS 256.105 s. Normal/sanitized controls have 31 findings; compiler 77 has
28 findings / 20,884 bytes; repository 287 has two / 19,558 bytes. Compiling
loop, require-import and cycle mutants exit zero with empty stderr and differ
only at bytes 51, 7383 and 12293. The retained released registry is caught by
the expected exit-70 refusal. Declaration flags, type-node classification and
resolved-module mutations are caught by direct checker comparisons.

Core compare_core.py passes 264/264 cases normally and under ASan/UBSan/leak
checks, including all finding, fix and suggestion bytes. Native executables were
rebuilt from suite.a with the fresh compiler and normal/sanitized archives.
validate_mutants.py catches outdated-set inversion in 72 cases, suggestion
semicolon inversion in 99, and Symbol name-filter inversion in 20; first
changed bytes are 119, 715 and 147. All compile, exit zero and have empty stderr.
Four new question modes reject released handles before output with exit 70.

Process validate_process.py passes all 100 upstream programs, both corpora,
normal and sanitizers. Callee and blocking-order mutants are caught in six and
nineteen programs at bytes 164 and 1690. Race validate.py passes 22 controls /
13 findings and both corpora, normal/sanitized; its compiling lost-timeout
mutant is caught at byte 52. All exercised raw question modes reject released
handles with exit 70. Globals validate.py passes 35 controls / 34 findings and
both corpora, normal/sanitized; its compiling ancestry mutant differs at 9162.
The JSX witness still refuses before native findings while Go reports.
validate_cores.py matches 52 sanitized partial-kernel results against Go,
catches join and dependency-count mutants, and verifies both blocked production
entries refuse with exit 70. These two are kernel mutants, not full-rule mutants.

GOFLAGS=-overlay=/workspace/wave08-core/bridge/overlay.json go test ./bridge/tsgo
./bridge/tsgo/checker ./stage1/cohere/typeaware/wave08-core-next -count=1 -v
-timeout 30m passes in 144.641 s, 0.356 s and 0.066 s. Ownership checks include
100 ABI queries, retained outputs, stale/zero rejection and the 162-position
independent Go oracle. Length mutants trigger ASan, omitted frees/heap-region
ownership trigger LeakSanitizer, retained handles fail the stale assertion,
wrong-position output differs, and removed link opt-in fails refusal.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v passes
in 0.458 s with native and Node cache misses. Every test's output went to a file.

Quiet three-round timing used time_core.py with --baseline
/workspace/wave08-b8fb/compare, --native /workspace/wave08-b8fb/core-native and
--oracle /workspace/wave08-b8fb/compare/oracle. Every alternating fresh-process
round also matches complete Go output. Medians include loading and teardown.
Native remains slower. These observations do not establish a speed improvement.

| Corpus | Rule | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | atomic | 9.572223 | 0.441246 | 21.694x |
| compiler | await | 9.484028 | 0.496887 | 19.087x |
| compiler | symbol | 3.490604 | 0.416915 | 8.372x |
| repository | atomic | 1.215545 | 0.189480 | 6.415x |
| repository | await | 1.127090 | 0.179615 | 6.275x |
| repository | symbol | 0.492887 | 0.182193 | 2.705x |

Complete fresh stdout/stderr streams and logs are compressed in
validation-b8fb/outputs.json.gz; comparison, mutants and timing JSON are beside
it. Scratch artifacts are /workspace/wave08-b8fb. Historical complete rule scope
and setup evidence remain in REPORT.md and the earlier owned reports. Setup was
already completed for this unit: 132 seconds, nproc 5, cgroup quota 4 cores;
Go 1.27.1, clang 20.1.8, Node 24.19.0. This turn sourced the existing toolchain
environment. No full repository gate or unimplemented React parity is claimed.

Built: wave 08 rebased onto merged stage1-lint; Globals now has complete captured upstream parity in its owned profile.
Commits: tested rebased tip 21a79262a1c1e585379e7af7dedf018b6ad1866b; this report accompanies the landing evidence commit.
Checks: ten completed profiles, two parked kernels, core 264/264 normal/sanitized, merged .a/suggestion harness and Node oracle pass.
Mutants: all ten completed-rule mutants caught by Go bytes; parked kernel mutants and released-handle refusals pass separately.
Uncovered: shared type-aware factories/checker-program registration, two React HIR/SSA/capture rules and full repository gate.

The branch was cleanly rebased onto origin/area/stage1-lint at
7481e0324e34a2537aafa9db7eeacda50405611b. It contains the requested 50a5f105
harness landing, 41eb6eab2, and current main
39638d9e278d38bb5aeae887f46d55a70e47aaad. Publication targets only
codex/typeaware-wave-08, with previous published 7110fcf78 as the exact lease.
No main or area branch is pushed. A fresh compiler was built; unchanged checker
archives were reused through the owned overlay, with raw checker tests rerun.
Shared source changes from integration were retained without modification.

The merged JSX parser closes the prior Globals blocker. The old validator's
expected exit 70 fails because native now exits zero; that failed expectation
is preserved in react-old-refusal.log. Its replacement requires complete Go
byte parity normally and sanitized. The expanded capture wraps every upstream
RunTyped/Run call, leaving production calls and all original assertions intact.
112 programs were captured, 111 typed and one deliberately checker-free guard
witness. All upstream Go assertions pass. The 111 typed programs have 58
findings and identical normal/sanitized native bytes, including JSX, destructuring,
shadowing, compilation-root exclusions and precise reported spans. The existing
35 controls have 34 findings; compiler 77 and repository 287 remain silent and
match Go. The compiling ancestry mutant is caught at byte 9162 in controls and
in seven upstream programs. It exits zero with empty stderr. A sanitized symbol
question after program release refuses with exit 70 before any output.

Commands from the repository root, with the existing cloud environment sourced:

```sh
go build -o /workspace/wave08-area/adamic ./cmd/adamic
python3 stage1/cohere/typeaware/wave08-react/capture_globals.py /workspace/wave08-area/globals-upstream
python3 stage1/cohere/typeaware/wave08-react/validate.py --artifacts /workspace/wave08-area/react --stage0 /workspace/wave08-area/adamic --archive /workspace/wave08-core/bridge/checker.a --asan-archive /workspace/wave08-core/bridge/checker-asan.a --compiler-root /workspace/typescript-wave08-corpus --fixtures /workspace/wave08-area/globals-upstream
python3 stage1/cohere/typeaware/wave08-react/time_globals.py --baseline /workspace/wave08-area/react --artifacts /workspace/wave08-area/globals-timing
```

Quiet alternating three-round fresh-process Globals medians, including loading,
serialization and teardown. Every round also compares full Go bytes:

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 3.159701 | 0.490314 | 6.444x |
| Repository | 0.510625 | 0.306692 | 1.665x |

Native remains slower; no performance improvement is inferred. Historical core
measurements remain in the prior landing report, rather than being presented as
measurements of this parser rebase.

Earlier claims were re-green with fresh native executables. Original
`go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m`
passes in 383.609 s with both frozen manifests and compiler root set. Controls
have 31 findings; compiler has 28 / 20,884 bytes; repository two / 19,558 bytes,
normal and sanitized. Compiling loop, require-import and cycle mutants differ
at bytes 51, 7383 and 12145. Released-registry and declaration/type-node/module
fact mutants pass their refusal and direct-comparison expectations.

Core compare_core.py passes 264/264 programs normally and under sanitizers.
validate_mutants.py catches atomic outdated-set inversion in 72 cases, await
suggestion semicolon inversion in 99 and Symbol name inversion in 20; first
changed bytes are 119, 715 and 147. All compile, exit zero and have empty stderr.
All four core raw-question modes reject released handles with exit 70.
Process validate_process.py passes 100 upstream programs and both corpora,
normal/sanitized; callee and blocking-order mutants differ in six and nineteen
programs, at bytes 164 and 1690. Race validate.py passes 22 controls / 13 findings
and both corpora, normal/sanitized; its compiling timeout mutant differs at byte
52. Exercised raw questions reject released handles with exit 70.

validate_cores.py matches 52 sanitized partial React kernel results against Go,
catches compiling join and dependency-count mutants, and checks both unsupported
production entries refuse with exit 70. These are kernel mutants, not completed
production-rule mutants. Immutability and deriving-state remain parked for native
HIR/SSA/capture analysis on #dnv6f2c.

The owned raw checker overlay passes `go test ./bridge/tsgo/checker
./stage1/cohere/typeaware/wave08-core-next -count=1 -v`. The shared registry tests
pass in 0.122 s. Shared `go test ./stage1/cohere/lint -run
'^(TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix)$'
-count=1 -v -timeout 30m` passes in 333.995 s. These compare unchanged Go,
source Node, emitted JavaScript and sanitized native, including complete
suggestions alongside an applied automatic fix. Filtered
`TestTheOracleCatchesOneByte` passes in 0.102 s with existing native/Node cache
hits. Every test's output went to a log file. The full repository gate and full
bridge foundation mutant suite were not rerun; bridge source is unchanged, and
this turn checks sanitizer ownership and released handles through the profiles.

The new pre-push checker was attempted with the existing claim and prints:
`lint-wave-check: FAIL claim: --claim must name a reservation under stage1/cohere/lint/claims/`.
It certifies the registered syntax-lint driver, not these owned type-aware
profiles. No receipt or shared-driver certification is claimed. The shared
RuleContext has no checker-program handle/configured root list; installing typed
factories and raw-question dispatcher hooks remains an integration prerequisite.
No shared harness, registry, context, dispatcher or compiler source was edited
as a workaround. The user explicitly authorized publication of the owned work
and parking its integration gaps.

A fresh all-heads audit covers 592 origin refs and 33 distinct type-aware claim
Markdown blobs. All 197 volume-ranked entries are already ported or claimed;
no additional rule was reserved. The claim now records Globals as complete in
its owned profile and the two HIR-dependent rules as parked. New native sources
would be .a; this unit changes tests and evidence only and introduces no regex.
Complete fresh logs and output streams, all 112 captured inputs, byte comparison
results, mutants and timing samples are committed in validation-area. Scratch
artifacts remain in /workspace/wave08-area.

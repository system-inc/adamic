Built: wave 08 rebased onto current stage1-lint runtime tip d65a8f931; ten completed profiles and two parked kernels re-green.
Commits: tested rebased tip 66088a4f88a4e1549b7f61ec163b146789196858; this report accompanies its landing evidence commit.
Checks: core 264/264 normal/sanitized, Globals 111 typed upstream programs, prior rules, runtime ownership/search and Node oracle pass.
Mutants: all ten completed-rule mutants caught by Go bytes; parked kernel mutants and released-handle refusals pass separately.
Uncovered: shared type-aware registration, two HIR/SSA/capture-dependent React rules, six fresh quiet profile timings and full repository gate.

Previous published tip was aa6916c4a8c42ffc39a41c328aefd1d4e3e20563.
The clean rebase onto origin/area/stage1-lint at
 d65a8f931c98655936ae04c6899f38f14862b73e retains its native heap release,
string equality and backward-search changes. It contains current main
39638d9e278d38bb5aeae887f46d55a70e47aaad and the requested harness landing.
A fresh compiler and every native profile/mutant executable were built. Existing
checker archives were reused because their Go dependencies are unchanged.
Only codex/typeaware-wave-08 is published with the prior tip as the explicit
lease; no main or area branch is pushed. No shared source was edited.

The all-heads fetch and audit covers 606 origin refs and 33 distinct type-aware
claim Markdown blobs. Of 197 ranked rules, 25 verified main/bridge port paths
and 172 claims cover the entire ranking, with no unclaimed entry. No additional
rule was reserved. The claim retains the two HIR-dependent parked reservations.

All test output went to log files under /workspace/wave08-d65. Fresh gate:
`go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m`
with ADAMIC_WAVE08_ARTIFACTS, ADAMIC_WAVE08_FACT_ARTIFACTS, compiler/repository
manifest variables and ADAMIC_TYPESCRIPT_SOURCE set, passes in 300.434 s.
Controls have 31 findings / 17,486 bytes, normal/sanitized; compiler 77 has
28 / 20,884 bytes; repository 287 has two / 19,558 bytes. Compiling loop,
require-import and cycle mutants exit zero with empty stderr and differ from Go
at bytes 50, 7373 and 12176. Released-registry and declaration/type-node/module
fact mutants pass their refusal and independent checker-comparison expectations.
Different scratch-header lengths explain changed control byte totals/offsets;
Go and native compare complete bytes within each run.

Core compare_core.py passes 264/264 cases normally and under ASan/UBSan/leak
checks, with complete finding, fix and suggestion bytes. validate_mutants.py
catches atomic outdated-set inversion in 72 cases, await suggestion semicolon
inversion in 99, and Symbol name inversion in 20. First differences are 119,
715 and 147. All mutants compile, exit zero and have empty stderr. Four core
raw-question modes reject released handles before output with exit 70.

Process validate_process.py passes all 100 upstream programs and both corpora,
normal/sanitized; callee and blocking-order mutants differ in six and nineteen
programs at bytes 164 and 1690. Race validate.py passes 22 controls / 13 findings
and both corpora, normal/sanitized; its compiling timeout mutant differs at byte
45. Its exercised question modes all refuse released handles with exit 70.

Globals validate.py passes all 111 typed captured upstream programs / 58
findings, 35 controls / 34 findings, JSX witness and both corpora, normal and
sanitized. The unchanged ancestry mutant differs at byte 9171 in controls and
in seven upstream programs. All runs exit zero with empty stderr; only Go byte
comparison catches it. The sanitized standard symbol question refuses a released
program with exit 70 before output. The 112 captured inputs and original passing
Go assertions remain in wave08-react/validation-area; one checker-free upstream
witness is excluded from the typed native comparison, rather than inventing a
native untyped mode.

validate_cores.py again matches 52 sanitized partial React kernel results
against Go, catches compiling join and dependency-count mutants, and verifies
both unsupported production entry points refuse with exit 70. These are kernel
mutants, not completed production-rule mutants. Native React HIR/SSA/capture
support remains pending on #dnv6f2c.

Related runtime checks: `go test ./internal/native -run
'^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$' -count=1 -v` passes in
15.703 s. `go test ./internal/oracle -run
'^(TestRuntimeLastIndexOfMatchesNode|TestTheOracleCatchesOneByte)$' -count=1 -v
-timeout 30m` passes in 0.798 s: Node, emitted JavaScript, release native and
sanitized native agree on 758 backward-search output bytes, with leak checks.
The one-byte oracle mutant is caught. The log reports native hits 1/misses 3,
Node hits 0/misses 3. Full repository and bridge foundation gates were not rerun.

Quiet timing ran after all gates finished. Commands:

```sh
python3 stage1/cohere/typeaware/wave08-core-next/time_core.py --artifacts /workspace/wave08-d65/core-timing --baseline /workspace/wave08-d65/compare --native /workspace/wave08-d65/core-native --oracle /workspace/wave08-d65/compare/oracle
python3 stage1/cohere/typeaware/wave08-react/time_globals.py --baseline /workspace/wave08-d65/globals --artifacts /workspace/wave08-d65/globals-timing
```

Alternating three-round fresh-process medians include loading, rule execution,
serialization and teardown. Every round retains complete Go byte equality.
Native remains slower; these observations do not establish a speed improvement.
Process/race validators also record their timing rounds, but those overlapped
other gates and are not presented as quiet performance evidence. Fresh quiet
medians were not run for the other six profiles in this landing unit.

| Corpus | Profile | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | atomic | 9.038948 | 0.440759 | 20.508x |
| compiler | await | 8.901748 | 0.427582 | 20.819x |
| compiler | symbol | 3.503097 | 0.416852 | 8.404x |
| repository | atomic | 1.359368 | 0.248208 | 5.477x |
| repository | await | 1.712671 | 0.272926 | 6.275x |
| repository | symbol | 0.539796 | 0.197813 | 2.729x |
| compiler | globals | 3.341152 | 0.511223 | 6.536x |
| repository | globals | 0.473490 | 0.180638 | 2.621x |

Shared type-aware factories, checker-program context and raw-question dispatcher
integration remain pending. The registered syntax-lint checker cannot certify
these owned profiles, as documented in the prior area landing report; no shared
certification receipt is claimed. The merged .a/suggestion harness itself was
validated in that report and its source is unchanged by this runtime update.
This unit changes only owned claims/evidence, introduces no new native sources
or regexes, and does not alter shared generator/harness/compiler files. The
native runtime updates are inherited intact through the requested rebase.

Complete fresh log/manifest/stdout/stderr streams are compressed in
validation-d65/outputs.json.gz, with comparison, mutant, audit and timing JSON
beside it. Historical complete inputs and per-rule commands remain in the
preceding owned reports. Toolchain environment is the previously configured
/workspace/adamic-tools/env.sh; unit setup was 132 s, nproc 5 with four-core quota,
Go 1.27.1, clang 20.1.8 and Node 24.19.0.

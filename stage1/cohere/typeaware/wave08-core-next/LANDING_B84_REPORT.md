Built: wave 08 rebased onto current main via lint area b84a9d931; ten completed profiles and two parked kernels re-green.
Commits: tested rebased tip 220cc318f8817e6e162688555329b23a6e4727e1; this report accompanies its landing evidence commit.
Checks: core 264/264 normal/sanitized, Globals 111 typed programs, prior rules, new proof/refusal lowering and Node oracle pass.
Mutants: all ten completed-rule mutants caught by Go bytes; parked kernel mutants and released-handle refusals pass separately.
Uncovered: shared typed registration, two HIR rules, six fresh quiet profile timings and the full required-input gate.

Previous published tip was 13e3c849294bbabf81e229cbbae86eef5c9f1fbd. The clean
rebase takes origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da,
which includes main c7991b900362796aefd111474e65eb5398e91953. New lowering and
record-runtime changes are inherited intact. A fresh compiler and all profile/
mutant native executables were built. Unchanged checker archives were reused.
Only codex/typeaware-wave-08 is published, with the prior tip as the explicit
lease. No shared implementation files, main or area branch are changed by this
worker. The all-origin audit covers 631 refs and 33 distinct claim Markdown
blobs: 25 verified port paths and 172 claims cover all 197 ranked rules. Nothing
unclaimed remains, so no additional rules were reserved.

The filesystem had 2.3 GB free before validation. Only reproducible ELF binaries
and ar archives larger than one megabyte were removed from the named older
scratch roots /workspace/wave08-b8fb and /workspace/wave08-d65: 58 artifacts,
2,334,375,980 bytes. Their committed compressed streams and inputs remain;
source files and the reusable /workspace/wave08-core/bridge archives were kept.

All output went to files under /workspace/wave08-b84. With both frozen manifests,
compiler corpus root and original artifact variables provided,
`go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m`
passes in 280.051 s. Controls have 31 findings / 17,486 bytes, compiler 77 has
28 / 20,884 bytes, repository 287 has two / 19,558 bytes, normal and sanitized.
Compiling loop, require-import and cycle mutants exit zero with empty stderr
and differ from unmodified Go at bytes 50, 7373 and 12133. Released-registry and
declaration/type-node/module fact mutants pass their refusal/direct-comparison
expectations. Scratch path headers explain differences from prior byte offsets;
every within-run comparison retains the complete stream.

Fresh core executables pass compare_core.py 264/264 normally and under ASan,
UBSan and leak checks, including findings, fixes and suggestions. The compiling
atomic outdated-set, await semicolon and Symbol name-filter mutants differ in
72, 99 and 20 cases, first bytes 119, 715 and 147. All exit zero with empty
stderr, and only Go byte comparison catches them. Four new raw-question modes
reject released handles before output with exit 70.

Process validate_process.py passes 100 upstream programs and both frozen
corpora normally and sanitized. Callee and blocking-order mutants compile and
exit zero with empty stderr, differing in six and nineteen programs at bytes
164 and 1690. Race validate.py passes 22 controls / 13 findings and both corpora,
normal/sanitized; its compiling timeout mutant differs at byte 45. All exercised
raw question modes reject released handles before output, exit 70.

Globals validate.py passes all 111 typed upstream programs / 58 findings,
35 controls / 34 findings, JSX witness and both corpora, normal/sanitized. Its
compiling ancestry mutant differs at byte 9171 in controls and in seven upstream
programs, with exit zero and empty stderr. A sanitized standard-symbol question
refuses a released program with exit 70 and no output. The complete 112 captured
inputs and passing original Go assertions remain in wave08-react/validation-area;
one checker-free guard witness is retained there but excluded from the typed
native comparison. The two parked React kernels match 52 sanitized Go results;
join and dependency-count mutants differ, and both HIR-dependent production
entries refuse with exit 70. Kernel mutants are not complete production-rule
mutants. Their native HIR/SSA/capture blocker remains #dnv6f2c.

Focused related checks:
`go test ./internal/lower -run
'^(TestUnprovenPredicateReturnsAreRefused|TestPredicateBodiesAreProven|TestProvenRelationsRefuse|TestProvenRelationsErase)$'
-count=1 -v` passes in 2.803 s. Filtered Node oracle
`go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v`
passes in 5.117 s with native and Node cache misses. A scan of selected Go logs
finds no skipped test. No check was weakened, relaxed, or deleted. The full gate,
including all 17 externally provisioned correctness checks, was not run or
claimed green; this is the authorized touched-profile and filtered-oracle gate.
Shared typed factory/checker-program/dispatcher integration still has no syntax
registration certification receipt, as detailed in the preceding area report.

Quiet alternating three-round fresh-process measurements ran after all gates:

```sh
python3 stage1/cohere/typeaware/wave08-core-next/time_core.py --artifacts /workspace/wave08-b84/core-timing --baseline /workspace/wave08-b84/compare --native /workspace/wave08-b84/core-native --oracle /workspace/wave08-b84/compare/oracle
python3 stage1/cohere/typeaware/wave08-react/time_globals.py --baseline /workspace/wave08-b84/globals --artifacts /workspace/wave08-b84/globals-timing
```

Every round also compares full Go bytes. Medians include loading, execution,
serialization and teardown. Native remains slower; no speed improvement is
inferred. Process/race timings overlapped other validation and are recorded but
not presented as quiet performance evidence. Six other profiles did not receive
new quiet medians in this landing unit.

| Corpus | Profile | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | atomic | 9.569203 | 0.470478 | 20.339x |
| compiler | await | 10.614610 | 0.503672 | 21.074x |
| compiler | symbol | 3.227348 | 0.378790 | 8.520x |
| repository | atomic | 1.472391 | 0.274197 | 5.370x |
| repository | await | 1.044791 | 0.174802 | 5.977x |
| repository | symbol | 0.493719 | 0.172296 | 2.866x |
| compiler | globals | 2.583068 | 0.393775 | 6.560x |
| repository | globals | 0.377155 | 0.171279 | 2.202x |

Complete fresh logs/manifests/stdout/stderr streams and the comparison, mutant,
audit and timing JSON are committed in validation-b84. This unit changes only
owned claim/evidence files, introduces no native implementation or regex, and
uses the already configured cloud toolchain environment. Unit setup was 132 s;
nproc 5, quota four cores, Go 1.27.1, clang 20.1.8, Node 24.19.0.

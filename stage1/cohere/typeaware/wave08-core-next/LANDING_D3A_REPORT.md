Built: wave 08 rebased onto lint area d3a37422c and current main b6b1538b0; fresh native profiles are green.
Commits: tested rebased tip 6926fb7d86f6d3e071e3797c3342e6184a31881b; this report accompanies its evidence commit.
Checks: ten completed profiles, two parked kernels, 264 core comparisons and 111 typed Globals programs pass with sanitizers.
Mutants: all ten completed-rule mutants caught by Go bytes; released-handle refusals and new typeof oracle mutants pass.
Uncovered: shared typed registration, two HIR rules, six fresh quiet profile timings and the full required-input gate.

The clean rebase takes origin/area/stage1-lint
d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing main
b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Previous published wave-08 tip was
970fdeabde14d0260fd76970be28e142d794d2b0. This base delta changes native typeof
classification, null meaning and slot lookup presence. The new compiler was
built from the rebased tree and all native profiles and mutants were rebuilt.
Unchanged normal and sanitized checker archives were reused. No shared source
was edited, and no main or area branch is a push target.

To retain disk space, only reproducible ELF/ar scratch binaries larger than one
megabyte under /workspace/wave08-b469 were removed: 26 files, 1,053,609,853 bytes.
Their input sources and committed output evidence remain. All new validation
output went to /workspace/wave08-d3a, without piping any running test.

With /workspace/adamic-tools/env.sh sourced, commands and observations:

- go build -o /workspace/wave08-d3a/adamic ./cmd/adamic: PASS.
- go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m,
  with ADAMIC_WAVE08_ARTIFACTS and ADAMIC_WAVE08_FACT_ARTIFACTS under the new
  scratch root, ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus,
  and both frozen /workspace/wave08-corpora manifests: PASS, 295.267 seconds.
  Controls 31 findings / 17486 bytes, compiler 77 roots 28 / 20884 and repository
  287 roots two / 19558 agree normally and sanitized. Loop, require-import and
  cycle mutants compile, exit zero with empty stderr, and differ under Go byte
  comparison at bytes 50, 7373 and 12179. Released handle rejects with exit 70;
  retaining-registry mutant violates that expectation. Declaration flags,
  type-syntax and resolved-module mutants fail direct checker comparisons.
- Fresh adamic build of wave08-core-next/suite.a, normally with checker.a and
  with --sanitize and checker-asan.a: PASS. compare_core.py --fixtures
  /workspace/wave08-core/upstream --compiler-root /workspace/typescript-wave08-corpus
  using these two binaries and separate artifact roots: 264/264 full streams
  match in each run, including fixes and suggestions. validate_mutants.py with
  the current stage0 and sanitized archive catches atomic outdated-set, await
  semicolon and Symbol name-filter mutants in 72, 99 and 20 programs, first
  differing bytes 119, 715 and 147. All compile, exit zero and emit no stderr.
  Four raw checker question modes reject released handles before output, exit 70.
- wave08-next/validate_process.py with 100 captured upstream programs, compiler
  root and both archives: all programs and both frozen corpora match normally
  and sanitized. Callee and blocking-order mutants compile and exit zero with
  empty stderr, differing in six and nineteen programs at first bytes 164 and
  1690. Three released question modes refuse with exit 70.
- wave08-next/validate.py using the fresh stage0, both archives and compiler
  root: 22 controls / 13 findings / 8152 bytes and both corpora match normally
  and sanitized. Timeout mutant compiles and exits zero with empty stderr,
  differing at byte 45. All exercised question modes reject released handles
  before output, exit 70.
- wave08-react/validate.py with /workspace/wave08-area/globals-upstream,
  current stage0, both archives and compiler root: 111 typed upstream programs
  / 58 findings, 35 controls / 34 findings, JSX and both corpora agree normally
  and sanitized. Compiling ancestry mutant exits zero with empty stderr and
  differs at byte 9171 in controls and in seven upstream programs. The original
  checker-free witness remains in captured Go assertions, excluded from typed
  native comparison. Fresh sanitized node-symbol-details probe rejects a
  released handle before output, exit 70.
- wave08-react/validate_cores.py with current stage0: 52 parked kernel results
  agree under sanitizers. Join and dependency-count mutants compile and differ
  under Go byte comparison. Both HIR-dependent production entries refuse,
  exit 70. Kernel checks are not complete production-rule mutants.
- go test ./internal/oracle -run
  '^(TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTheOracleCatchesOneByte)$'
  -count=1 -v -timeout 30m: PASS, 8.007 seconds. All new typeof and slot-presence
  mutant checks pass; native cache five hits/six misses, Node one hit/eight
  misses. These hold the newly inherited runtime behavior to Node.

All four validation command chains exited zero. No selected Go check skipped.
The full gate, including all 17 externally provisioned correctness checks, was
not run or claimed green. No test was skipped, relaxed or deleted. Shared
registry/.a/suggestion serialization checks from LANDING_B469_REPORT.md are not
rerun because their implementation is unchanged in this base delta.

Quiet timings ran sequentially after every gate, using three alternating fresh
process runs and complete byte comparison on every round:

```sh
python3 stage1/cohere/typeaware/wave08-core-next/time_core.py --artifacts /workspace/wave08-d3a/core-timing --baseline /workspace/wave08-d3a/compare --native /workspace/wave08-d3a/core-native --oracle /workspace/wave08-d3a/compare/oracle
python3 stage1/cohere/typeaware/wave08-react/time_globals.py --baseline /workspace/wave08-d3a/globals --artifacts /workspace/wave08-d3a/globals-timing
```

| Corpus | Profile | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | atomic | 8.843927 | 0.375758 | 23.536x |
| compiler | await | 9.801527 | 0.372939 | 26.282x |
| compiler | symbol | 3.323325 | 0.408360 | 8.138x |
| repository | atomic | 1.535654 | 0.472571 | 3.250x |
| repository | await | 1.137527 | 0.193251 | 5.886x |
| repository | symbol | 0.499598 | 0.186324 | 2.681x |
| compiler | globals | 3.237371 | 0.452720 | 7.151x |
| repository | globals | 0.464217 | 0.231106 | 2.009x |

Native remains slower. Medians include loading, execution, serialization and
teardown; these observations do not prove a performance change from the prior
base. Process/race timing observations overlap gates and are retained but not
presented as quiet comparisons. Six other profiles have no new quiet medians.

The audit covers 654 origin refs and 33 distinct claim Markdown blobs. All 197
ranked rules are covered by 25 verified ports and 172 claims; eligible list is
empty, so no new claim is made. Complete log/stdout/stderr/manifest streams and
nine summary JSON files are retained in validation-d3a. Existing setup timing
is 132 seconds, nproc 5, quota four cores; the configured toolchain is reused.

Shared typed factory/driver integration remains pending: RuleContext lacks a
checker-program handle and configured root list. Owned profiles and scratch
bridge overlays do not certify shared registration. React immutability and
no-deriving-state-in-effects remain parked for native HIR/SSA/capture analysis
on #dnv6f2c. These limitations are unchanged by this evidence-only unit.

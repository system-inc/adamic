Built: wave 08 rebased onto main 4e0bfda50, then merged lint area bb2ece564; all owned native profiles re-green.
Commits: tested combined tip 0f422894357a29808300644e293def4d1e08021f; this report accompanies its evidence commit.
Checks: fresh compiler, ten completed profiles, parked kernels, registry and option-bearing witnesses pass with sanitizers.
Mutants: all ten completed-rule mutants caught by Go bytes; released handles, kernel mutants and filtered Node checks pass.
Uncovered: shared typed registration, two HIR rules, six fresh quiet profile timings and the full required-input gate.

Previous published tip was fe8c2376e8d4bd67f08b8a39fac5cbf99f7291d8. Current main
4e0bfda50a19c705a1aac0d9932e08483806d61c contains newly integrated runtime and
developer-tools work. Lint area bb2ece564842c4b2f909b9f75c27e74c2efa4f29 adds
option-bearing owned witnesses and recovery-row classification, but does not yet
contain current main. The owned branch was therefore rebased onto main and the
area was merged into the owned branch. Both operations were clean. No shared
implementation was edited and no main or area branch is a push target.

Runtime changes include borrowing, string construction/indexing/normalization,
Map/Set behavior and native compilation caching. All owned native profiles and
mutants were rebuilt with a fresh compiler. The checker bridge and owned profile
sources are unchanged from the previous published tree, so existing normal and
sanitized Go checker archives were reused. Only reproducible ELF/ar binaries
larger than one megabyte under /workspace/wave08-b84 were removed for disk space:
29 artifacts, 1,169,038,320 bytes. Inputs and committed compressed streams remain.
All new test output went to /workspace/wave08-4e0, never through a running pipe.

Commands and observed results, with the configured toolchain environment sourced:

- go build -o /workspace/wave08-4e0/adamic ./cmd/adamic: PASS.
- go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m,
  with ADAMIC_WAVE08_ARTIFACTS and ADAMIC_WAVE08_FACT_ARTIFACTS under the new
  scratch root, ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus and
  both frozen /workspace/wave08-corpora manifests: PASS, 355.030 seconds.
  Controls 31 findings / 17486 bytes, compiler 77 roots 28 / 20884 and repository
  287 roots two / 19558 agree normally and sanitized. Compiling loop,
  require-import and cycle mutants exit zero with empty stderr and differ only
  under Go byte comparison at bytes 50, 7373 and 12133. Released handle rejects
  with exit 70; retaining-registry mutant exits zero and violates the expected
  refusal. Declaration flags, type-syntax and resolved-module mutants fail
  direct checker comparisons.
- Fresh adamic build of wave08-core-next/suite.a, normal and --sanitize with
  matching checker archives: PASS. compare_core.py --fixtures
  /workspace/wave08-core/upstream --compiler-root /workspace/typescript-wave08-corpus
  using these binaries: 264/264 complete finding/fix/suggestion streams match
  normally and sanitized. validate_mutants.py with current stage0 and sanitized
  archive catches atomic outdated-set, await semicolon and Symbol name-filter
  mutants in 72, 99 and 20 programs, first bytes 119, 715 and 147. All compile,
  exit zero and emit no stderr. Four raw-question modes reject released handles
  before output with exit 70.
- wave08-next/validate_process.py with current stage0, both checker archives,
  /workspace/wave08-process-full and compiler root: all 100 upstream programs
  and both corpora agree normally and sanitized. Callee and blocking-order
  mutants compile, exit zero with empty stderr, differing in six and nineteen
  programs at first bytes 164 and 1690. Released question modes reject with
  exit 70 before output.
- wave08-next/validate.py with current stage0, both archives and compiler root:
  22 controls / 13 findings / 8152 bytes and both corpora agree normally and
  sanitized. Timeout mutant compiles and exits zero with empty stderr, differing
  at byte 45. All exercised released-question modes reject with exit 70.
- wave08-react/validate.py with current stage0, both archives, compiler root
  and /workspace/wave08-area/globals-upstream: 111 typed programs / 58 findings,
  35 controls / 34 findings, JSX and both corpora agree normally and sanitized.
  Ancestry mutant compiles and exits zero with empty stderr, differing at byte
  9171 in controls and in seven upstream programs. The original checker-free
  witness remains captured with its Go assertion, outside typed native
  comparison. A freshly rebuilt sanitized node-symbol-details probe rejects
  a released handle before output, exit 70.
- wave08-react/validate_cores.py with current stage0: 52 parked kernel results
  agree under sanitizers. Join and dependency-count mutants compile and differ
  under Go byte comparison. Both HIR-dependent production entries refuse with
  exit 70; kernel mutants are not complete production-rule mutants.
- go test ./internal/native -run
  '^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$' -count=1 -v:
  PASS, 11.963 seconds.
- go test ./internal/oracle -run
  '^(TestRuntimeLastIndexOfMatchesNode|TestTheOracleCatchesOneByte)$'
  -count=1 -v -timeout 30m: PASS, 1.008 seconds. Native observation cache one
  hit/three misses, Node zero hits/three misses.
- go test ./stage1/cohere/lint/registry -count=1 -v: PASS, 0.371 seconds,
  including descriptor rejection mutants and .a support.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v
  -timeout 30m: PASS, 113.572 seconds. Shared owned witnesses, including the
  newly option-bearing no-bitwise witness, report findings and match the real
  Go oracle across source Node, emitted JavaScript and sanitized native. No
  nil-options adapter panic occurred; no guard was weakened or bypassed.

All validation command chains exited zero. No selected Go test skipped. The
full gate, including all 17 externally provisioned correctness checks, was not
run or claimed green. No test was skipped, relaxed or deleted. Inherited shared
allocator/leak-check and developer-tools changes were retained.

Quiet timing commands ran sequentially after every validation job, with three
alternating fresh processes and full byte comparison on every round:

```sh
python3 stage1/cohere/typeaware/wave08-core-next/time_core.py --artifacts /workspace/wave08-4e0/core-timing --baseline /workspace/wave08-4e0/compare --native /workspace/wave08-4e0/core-native --oracle /workspace/wave08-4e0/compare/oracle
python3 stage1/cohere/typeaware/wave08-react/time_globals.py --baseline /workspace/wave08-4e0/globals --artifacts /workspace/wave08-4e0/globals-timing
```

| Corpus | Profile | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| compiler | atomic | 9.065854 | 0.418204 | 21.678x |
| compiler | await | 10.111018 | 0.896136 | 11.283x |
| compiler | symbol | 2.866555 | 0.493608 | 5.807x |
| repository | atomic | 1.051856 | 0.204995 | 5.131x |
| repository | await | 1.009563 | 0.217329 | 4.645x |
| repository | symbol | 0.429043 | 0.184666 | 2.323x |
| compiler | globals | 2.552635 | 0.425831 | 5.994x |
| repository | globals | 0.447679 | 0.207011 | 2.163x |

Native remains slower. Medians include program loading, execution, serialization
and teardown. These observations do not prove a performance improvement between
bases. Process/race validator timings overlap gates and are retained separately;
six other profiles received no fresh quiet medians.

The audit covers 720 origin refs and 42 distinct claim Markdown blobs. All 197
ranked rules are covered by 25 verified ports and 172 claims; none is eligible.
No new reservation is made. Complete stdout/stderr/log/manifest streams and
nine summary JSON files are committed in validation-4e0. Unit setup remains the
previous 132-second setup, nproc 5 and quota four cores; configured tools reused.

Shared typed factory/driver integration still needs a checker-program handle
and configured root list in RuleContext; owned overlays do not certify that
integration. React immutability and no-deriving-state-in-effects remain parked
for native HIR/SSA/capture analysis on #dnv6f2c. Only owned claim/evidence files
are added, and only codex/typeaware-wave-08 is published with an explicit lease
on the previous published SHA.

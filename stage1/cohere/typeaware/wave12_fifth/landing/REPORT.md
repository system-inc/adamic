Built: rebased the published wave 12 branch onto fetched main e8ba3d5d; no new rules claimed or ported.
Commits: old published tip dc360b91, definitive tested tip 4e96ec41; full rewritten commit mapping is preserved.
Checks: twelve owned rule oracles, 1,126 controls, 77 compiler roots and 287 repository roots agree byte for byte, normal and sanitized; checker, uncached Node oracle and vet pass.
Mutants: twelve rule mutants, five raw-question mutants, four registry-retention mutants, JSX guard reversal and Node one-byte mutation were caught.
Uncovered: three native-HIR/JSX-blocked claims, seventeen JSX dependency controls, nondefault options, full repository gate, inherited 26-rule regression suite and emitted-JavaScript lint execution.

## Landing scope and base

This worker has one published branch, codex/typeaware-wave-12. It was not an ancestor of main. The first rebase onto e011f8f6 succeeded without conflicts and all twelve owned ports passed their oracles. While that gate ran, main advanced with compiler call-target/devirtualization changes. A second clean rebase onto fetched main e8ba3d5d81de4d3773c723914fccd4c76248b965 was followed by the complete owned-rule gate again. That second run is the definitive evidence below.

The tested code tip is 4e96ec4188b33db8358f3d8ae36109215406ba9b. Main is an ancestor of that tip. No protected compiler file differs from fetched main: internal/native/emit.go, internal/lower/lower.go, internal/native/native.go and internal/oracle/oracle_test.go. No rule implementation needed a change. The only maintenance edit compresses the owned historical compound-name-failure.log and updates its report reference. Its decompressed bytes were asserted identical; this removes a preexisting trailing-whitespace failure from the complete branch patch check.

The rebase rewrites commit identities. [rebase-map.json](evidence/rebase-map.json) maps all eighteen previously published commits to their new identities and identifies the additional archive-cleanup commit. The original remote tip is dc360b9199008f6d73831d14dda92cee54a0079a. The authorized rebased push uses an explicit force-with-lease for that exact tip, guarding against overwriting a concurrent branch update. No PR is opened.

## Definitive byte agreement

All four suites use independent pinned production Go rule Run methods, their own Go program and AST walk, and complete findings/fixes/suggestions serialization. No bridge predicate supplies a lint verdict. All normal and sanitized native comparison stderr is empty.

| Owned batch | Control files | Findings | Identical bytes | Normal + ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Redeclare/global-regex/write-only collection | 187 | 89 | 65,212 | PASS |
| Process/race-timeout/blocking-streams | 142 | 138 | 87,753 | PASS |
| Three global constructor rules | 94 | 56 | 26,497 | PASS |
| Regex literals/rest params/exhaustive deps | 703 | 596 | 327,920 | PASS |

Each suite also ran the same frozen 77 compiler and 287 repository roots (212 .a, 75 .ts) in normal and sanitized builds. First batch: compiler 5 findings/8,014 bytes, repository 1 finding/18,903 bytes. Other three batches: compiler 0 findings/5,010 bytes, repository 0 findings/18,485 bytes. Entire canonical streams agree, including proposed fixes and suggestions. Byte counts differ from older reports where generated control paths have different lengths; findings did not change.

## Commands and observed results

bash cloud/setup.sh passed: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 159s, total 159s. nproc: 5, cgroup quota four cores, memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. All commands source /workspace/adamic-tools/env.sh, and test stdout/stderr go directly to log files.

For each batch the corresponding ADAMIC_WAVE12[_NEXT|_THIRD|_FOURTH]_COMPILER_MANIFEST and _REPOSITORY_MANIFEST variables point to /workspace/wave-12/compiler.manifest and /workspace/wave-12/repository.manifest. Corresponding _ARTIFACTS variables point to /workspace/wave-12/landing-current/{first,next,third,fourth}. ADAMIC_TYPESCRIPT_SOURCE is /workspace/wave-12/corpus.

```sh
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware \
  -run '^TestWave12AgreementAndMutants$' \
  > /tmp/wave-12-landing-current-first.log 2>&1
go test -p 1 -v -count=1 -timeout 30m \
  ./stage1/cohere/typeaware/wave12_next \
  ./stage1/cohere/typeaware/wave12_third \
  ./stage1/cohere/typeaware/wave12_fourth \
  > /tmp/wave-12-landing-current-other.log 2>&1
go test -v -count=1 ./bridge/tsgo/checker \
  > /tmp/wave-12-landing-current-checker.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  > /tmp/wave-12-landing-current-node-uncached.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware \
  ./stage1/cohere/typeaware/wave12_next ./stage1/cohere/typeaware/wave12_third \
  ./stage1/cohere/typeaware/wave12_fourth \
  > /tmp/wave-12-landing-current-vet.log 2>&1
```

PASS: first 126.057s, next 160.178s, third 110.663s, fourth 144.616s; checker 0.155s; vet empty output. The initial filtered Node run passed in 1.518s but reported six application-cache hits. The definitive repeat with ADAMIC_GATE_UNCACHED=1 passed in 1.437s with zero cache hits. Six selected native/Node/emitted-JavaScript fixtures actually ran, including regexp_cycle_closures matched by the filter, with sanitizers/leak checks and the one-byte mutant. This is the compiler oracle, not emitted JavaScript of the lint suites.

The three blocked claims' pinned Go reference tests also passed, 0.179s with 38 top-level tests. That reference-only run is not native agreement or a completed port.

## Every mutant and catch

Every rule and raw-fact byte mutant below exits 0 with empty stderr; only the independent full-byte comparison rejects it.

| Mutation | First differing byte |
| --- | ---: |
| Redeclare skips one surviving declaration | 60 |
| Global regex g flag becomes y | 5,592 |
| Write-only collection diagnostic end +1 | 9,378 |
| Process rule loses one-level callee writes | 3,987 |
| Timeout reverses lost-handle decision | 4,664 |
| Blocking-streams reverses shebang reason | 725 |
| Function constructor reverses global classification | 1,418 |
| Native nonconstructor reverses global classification | 68 |
| Wrapper constructor reverses global classification | 584 |
| Regex reverses comment suggestion gate | 528 |
| Rest reverses implicit-binding declaration check | 5,652 |
| Exhaustive deps reverses optional suggestion access | 7,044 |
| Ancestry declaration-file bits false | 75 |
| Resolved declaration loses body flag | 3,987 |
| Import resolution loses target paths | 13,420 |
| Constructor declaration-file bits false | 68 |
| Identifier-binding declaration-file bits false | 69 |

All four suites reject released handles with panic 70 and the exact invalid or released checker handle message. Four test-only overlays retain the registry entry; each probe then exits 0 and is caught by the release assertion. Reversing raw JSX presence makes valid controls fail, caught by their success requirement. TestTheOracleCatchesOneByte catches its changed reference byte. The two first-batch raw-question mutants recorded in the original report were not repeated; their direct checker assertions ran normally here.

## Fresh native against Go timings

Three alternating whole-process rounds for each suite/corpus, after all builds and initial gates finished. The existing benchmark_wave_12.py driver was reused unchanged; every round compared complete output bytes. No build or test competed during these runs. Medians include loading, parsing, rules, serialization, sorting and process overhead. Native remains slower.

| Suite / corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| first / compiler | 4.139s | 0.415s | 9.98x |
| first / repository | 0.500s | 0.113s | 4.41x |
| next / compiler | 2.632s | 0.370s | 7.11x |
| next / repository | 0.415s | 0.114s | 3.63x |
| third / compiler | 2.566s | 0.360s | 7.12x |
| third / repository | 0.430s | 0.095s | 4.54x |
| fourth / compiler | 3.313s | 0.435s | 7.61x |
| fourth / repository | 0.422s | 0.109s | 3.88x |

Per-suite measurement JSON and compressed timing streams are in evidence/. Compiler medians range from 2.566 to 4.139s native and 0.360 to 0.435s Go; repository medians range from 0.415 to 0.500s native and 0.095 to 0.114s Go. These are observations, not a claim that this rebase improved performance.

## Evidence and remaining scope

[Stream hashes](evidence/stream-hashes.json) accompany 372 compressed oracle/mutant/timing streams. [Control hashes](evidence/control-hashes.json) record all 1,126 regenerated controls; [source hashes](evidence/source-hashes.json) pin the frozen corpus and compiler/bridge/rule source state. Initial and definitive gate/rebase/setup logs are preserved compressed.

No new rules were claimed. Existing set-state-in-effect, set-state-in-render and static-components claims remain unfinished because native React HIR lowering, SSA/capture translation, memoization preparation and post-dominator/control-dependence analyses are absent. Static-components additionally needs native JSX parsing. Current fetched main and the shared harness branch still have no native HIR paths. The owned runner explicitly refuses JSX; seventeen dependency-rule JSX controls remain blocked. See the parent REPORT.md for the exact substrate evidence.

This is the landing gate for the twelve owned completed ports. It is not the full repository gate or a repeat of the inherited 26-rule suite. Nondefault options, all upstream multi-file fixtures, arbitrary regex patterns/JSX projects, suppression/edit application and emitted-JavaScript lint-suite execution remain outside the measured coverage. No Go-only reference result is counted as a port, and no finding-free placeholder was added for the blocked claims.

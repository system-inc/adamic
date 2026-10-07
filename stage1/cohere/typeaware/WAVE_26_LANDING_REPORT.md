Built: rebased the wave 26 landing unit onto current main; fifteen completed ports are green again, three React claims remain unfinished.
Commits: validated code e507ef697284f1d96f70e59b4a1daf9146d0b131; base e8ba3d5d81de4d3773c723914fccd4c76248b965; previous pushed tip 2b64539b1416cba60fd2f407e01cdc8e76948251.
Checks: five rule suites, 218 controls and 225 findings; both 77-file compiler and 287-file repository corpora; normal and sanitizer bytes match; bridge, Node, vet and formatting pass.
Mutants: fifteen rule mutants, thirteen raw-question mutant runs, five released-registry mutants, seven bridge mutants and Node's one-byte mutant were caught.
Not covered: the three React HIR analyses, the full repository gate, nondefault rule options or a combined fifteen-rule throughput benchmark.

The user's landing-first instruction makes this turn a rebase and validation unit.
Only codex/typeaware-wave-26 was pushed by this worker. No new claims were taken.
The first rebase onto e011f8f6 passed all five suites in 826.659 seconds. Main
advanced during that run. After it finished, the branch was rebased again onto
e8ba3d5d without conflicts and all five suites passed again. The second run
used independent processes and separate artifact directories. Its concurrent
per-test timings are validation observations, not performance benchmarks.

Each suite ran this command with its exact test name from results.json:

```sh
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace/wave-26-bridge-scratch \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-26-typescript \
ADAMIC_WAVE26_REPOSITORY_MANIFEST=/workspace/wave-26-repository.manifest \
ADAMIC_WAVE26_COMPILER_MANIFEST=/workspace/wave-26-compiler.manifest \
go test ./stage1/cohere/typeaware -run '^TEST_NAME$' -count=1 -v -timeout 30m
```

The five test names are TestWave26AgreementAndMutants,
TestWave26NextAgreementAndMutants, TestWave26ThirdAgreementAndMutants,
TestWave26FourthAgreementAndMutants and TestWave26FifthAgreementAndMutants.
All test stdout and stderr went directly to log files. Five artifact environment
variables ADAMIC_WAVE26_ARTIFACTS, ADAMIC_WAVE26_NEXT_ARTIFACTS,
ADAMIC_WAVE26_THIRD_ARTIFACTS, ADAMIC_WAVE26_FOURTH_ARTIFACTS and
ADAMIC_WAVE26_FIFTH_ARTIFACTS selected separate /workspace/wave-26-rebase-*
directories.

The control counts by batch are 30, 43, 44, 36 and 65. Finding counts are
27, 26, 45, 33 and 94. Each compares the complete canonical diagnostic stream,
including proposed fixes and suggestions, to unchanged production Go rules.
The two frozen corpora have zero findings for these fifteen rules; positive
controls and diagnostic-only mutants supply the nonzero evidence.
Every ordinary rule/question mutant compiles, exits 0 and has empty stderr;
only the diagnostic byte comparison catches it. Exact names and differing
byte offsets are retained in mutants.json and the complete suite logs.
Released handles panic with exit 70; five registry-retention mutants exit 0
and fail that required refusal contract.

Additional checks on the final base:

```sh
go test ./bridge/tsgo/... -count=1 -v -timeout 15m
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)\.a$' -count=1 -v -timeout 10m
go vet ./...
gofmt -l cmd internal
git diff --check
```

These commands passed. Vet and formatting output are empty. The filtered
Node oracle passes fifteen fixtures, including the newly landed devirtualization
fixtures, and proves its one-byte mutant fails. The bridge covers 100 C queries,
outputs surviving release, invalid handles and distinct programs; 162 positions
match in 3261 bytes under ASan/UBSan/LSan. Its seven mutants are input length,
output length, retained released handle, wrong source position, removed link
opt-in guard, omitted C output free and region allocation on the heap.
ASan, LeakSanitizer, the byte oracle and refusal assertions catch them as
recorded in bridge.log.gz.

After all builds and validation processes ended, three alternating quiet rounds
measured the rebased preference suite against its independent Go oracle:

| Corpus | Native process median | Go process median | Findings |
| --- | ---: | ---: | ---: |
| compiler, 77 files | 5.679325 s | 0.566245 s | 0 |
| repository, 287 files | 0.763188 s | 0.192700 s | 0 |

Native remains slower. The benchmark uses volume_bench.py with
/workspace/wave-26-rebase-fifth/wave26 and wave26-oracle. It compares output
on each run. Raw measurements, phases and stderr are preserved; this is a
three-rule preference-suite benchmark, not a benchmark of all fifteen ports.
The pinned sources and portable manifests remain in validation-wave-26-fifth.
Toolchain setup was reused: Go ready 0s, clang ready 1s, Node ready 1s,
submodules 1s, warm 135s, total 135s; nproc is 5.

The three React claims remain blocked on native/raw HIR, SSA, captures,
memoization rewriting and control/post-dominance support. Rebase green does
not imply those three are implemented. Their previous blocker and overlapping
origin claims remain documented in claims/wave-26.md. No shared harness,
registration generator or protected compiler file was edited for this landing
unit. Compiler changes came only from the upstream rebase.

Validation artifacts are under validation-wave-26-landing. Canonical Go streams
are compressed with deterministic gzip; diagnostic-hashes.json proves all four
normal/sanitized Go/native streams match for every batch and population.
The rebase push uses an exact lease against the previously verified remote tip,
as required to publish the user-requested rewritten branch without replacing
an unexpected concurrent remote update.

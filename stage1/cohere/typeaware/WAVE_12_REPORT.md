# Wave 12 report

Built: no-redeclare, no-test-on-global-regex, and no-write-only-collection as native Adamic passes.
Commits: claim `adea8d38` was pushed before implementation; implementation `6f852293`, based on `0d540f4`.
Checks: 187 controls, 77 compiler roots, and 287 repository roots agree byte for byte, including sanitizer runs.
Mutants: three rule mutants, two raw-question mutants, and one released-handle registry mutant were caught.
Uncovered: full repository gate, nondefault rule options and JSX, and the pinned cohere CLI's unsupported `.a` lint gate.

## Scope and ownership

Positions 34–36 after excluding the 26 existing ports are:

| Rule | Compiler findings | Repository findings |
| --- | ---: | ---: |
| @typescript-eslint/no-redeclare | 2 | 0 |
| nexus/correctness-no-test-on-global-regex | 2 | 0 |
| nexus/correctness-no-write-only-collection | 1 | 1 |

All fetched origin branches were checked for claims and ports before the claim commit. A subsequent deduplicated scan of 576 Adamic/TypeScript source blobs across 268 origin refs found only configuration references on `origin/codex/realpath-node-walk`, not implementations. No rule was skipped.

Each rule has its own `.a` file. The suite and support decoders are new files; no existing Adamic rule was edited. Raw `scope-metadata` and `global-symbol-details` questions have separate Go and Adamic files. Each Go question registers in one line in its own file. The only existing shared source edit is one dispatcher line in `bridge/tsgo/checker/facts.go`; the registry itself is new. Protected compiler files and pinned submodules are unchanged.

Go returns raw scope locations and global symbol declarations. Adamic makes lint decisions. The independent oracle invokes unchanged production cohere implementations and serializes full diagnostic records, including fixes and suggestions. These three rules produce neither fixes nor suggestions; their empty fields are still compared.

## Inputs and observations

Cohere pin: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
TypeScript-go pin: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
TypeScript compiler pin: `050880ce59e30b356b686bd3144efe24f875ebc8`, downloaded from GitHub codeload after the full clone was too slow.

The portable compiler and repository manifests are copied unchanged from validation-coverage. Compiler: 77 roots. Repository: 287 frozen roots (212 `.a`, 75 `.ts`). Controls: 27 targeted cases plus 160 source rows extracted from the pinned Go reference tests (86 + 29 + 22 + 23). Source rows are evaluated with production defaults; fixture-specific option settings are not reproduced.

| Input | Findings | Identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | --- |
| Controls | 89 | 62,220 | PASS |
| Compiler | 5 | 8,014 | PASS |
| Repository | 1 | 18,903 | PASS |

Sanitized native runs produced no sanitizer diagnostics. Compressed streams, SHA-256 hashes, process stderr and test logs are in [validation-wave-12](validation-wave-12/diagnostic-hashes.json).

## Commands and outputs

`bash cloud/setup.sh` passed: Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s, build-cache warm 115s, total 115s. Go 1.27.1; clang 20.1.8; Node 24.19.0; `nproc` = 5 (CPU quota 4). Environment: `source /workspace/adamic-tools/env.sh`.

```sh
ADAMIC_WAVE12_ARTIFACTS=/workspace/wave-12/final \
ADAMIC_WAVE12_REPOSITORY_MANIFEST=/workspace/wave-12/repository.manifest \
ADAMIC_WAVE12_COMPILER_MANIFEST=/workspace/wave-12/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave12AgreementAndMutants$'
```

PASS, 101.359s. The harness builds stage0, the C checker archive, native normal/sanitized binaries, and the independent Go oracle. It writes each subprocess's stdout and stderr directly to files.

`go test -v -count=1 ./bridge/tsgo/checker`: PASS, 0.152s.

```sh
go test -v -count=1 -timeout 10m ./internal/oracle \
-run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|maps_and_text)\.a$'
```

PASS, 22.749s; functions, generic_functions, maps_and_text, and the one-byte oracle mutant ran.

`go vet ./bridge/tsgo/... ./stage1/cohere/typeaware`, `gofmt -l bridge/tsgo/checker stage1/cohere/typeaware`, and `git diff --check`: clean. All test output was saved directly to logs, never piped.

Initial reuse of existing `Caller.ts` pulled in `Unused.ts` and failed native compilation at its constructor's early `this` escape. A new minimal raw-symbol decoder avoids that dependency and passes native compilation; no compiler or shared rule workaround was made.

## Mutant evidence

| Mutant | Observation | Catch |
| --- | --- | --- |
| no-redeclare `.slice(1)` to `.slice(2)` | Native exits 0, empty stderr | Full-byte oracle differs at byte 44 |
| global-regex flag `g` to `y` | Native exits 0, empty stderr | Full-byte oracle differs at byte 5,496 |
| write-only collection diagnostic end + 1 | Native exits 0, empty stderr | Full-byte oracle differs at byte 9,106 |
| Scope-container count + 1 | Compiles; raw question test fails | Scope header assertion |
| Global lookup name replaced with `undefined` | Compiles; raw question test fails | Global symbol header assertion |
| Remove released-program registry deletion | Probe exits 0 instead of panic 70 | Released-handle assertion |

The three rule mutants are caught only by comparison, not by crashes or sanitizer errors. A normal released-program query to the new global-symbol question panics 70 for an invalid/released handle. Raw checker mutants were run using Go overlays; they failed the intended assertions, not compilation.

## Native versus Go timing

Three interleaved whole-process runs per corpus, with full bytes compared every round. Reproduce with `testdata/benchmark_wave_12.py`; raw measurements are [measurements.json](validation-wave-12/measurements.json).

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 3.522s | 0.450s | 7.82x |
| Repository | 0.442s | 0.108s | 4.08x |

Compiler phase medians: native load 0.324s, run 3.185s; Go load 0.320s, run 0.113s. Repository: native load 0.104s, run 0.347s; Go load 0.082s, run 0.016s. These are observations from this five-CPU worker, not a speedup claim. Native is slower on both corpora.

## Limits

The full repository gate and the prior 26-rule suite were not rerun. The touched checker package, new rule gate, vet, and filtered external Node oracle passed. Default options are covered; configurable no-redeclare options, JSX, and arbitrary projects outside these corpora are not established by this evidence.

The additional pinned cohere CLI lint check over the seven new `.a` files exits 1: “nothing to check,” because that CLI considers `.a` neither TypeScript nor JavaScript despite the repository source-extension configuration. This lint gate is uncovered. The native compiler and the independent production-rule oracle both consumed `.a` successfully. No authored `.ts` workaround or submodule upgrade was introduced.

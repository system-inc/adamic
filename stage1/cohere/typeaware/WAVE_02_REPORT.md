Built: logical-assignment-operators, no-unsafe-return, and correctness-no-nullish-stripping-assertion in separate .a files.
Commits: claim e3fdad37 pushed before code; implementation ec89dc1858cdbda70d31e72ab226835496afac7f.
Checks: wave byte oracle PASS (78.303s), bridge packages PASS, filtered Node oracle PASS, lint/type/format PASS.
Mutants: three running rule mutants caught by Go bytes; two checker mutants caught by identity; released-registry mutant caught by panic 70; foundation mutants listed below.
Limits: default rule options, frozen corpora, filtered runtime gate; native is slower than Go; full repository test suite not run.

# Wave 02

The branch is `codex/typeaware-wave-02`, from the requested checker-library base
`0d540f413625f016f20fea39761c7b184f335de6`. The unit-specific base took precedence
over the generic origin/main instruction. CLAUDE.md and all three named type-aware
reports were read before editing. All origin heads were fetched and searched for
claims and named ports. No selected rule was already claimed or ported; none was
skipped. The claim was committed and pushed before implementation.

## Observed agreement

| Corpus | Files | Logical assignment | Unsafe return | Nullish assertion | Total | Identical bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Controls | 23 | 14 | 9 | 5 | 28 | 10,323 |
| Frozen repository | 287 | 3 | 0 | 0 | 3 | 19,425 |
| TypeScript compiler | 77 | 140 | 126 | 97 | 363 | 142,372 |

Every serialized finding includes rule, message ID and text, UTF-8 byte range,
ordered fixes and ordered suggestions with their edits. Both normal and
ASan/UBSan/LSan native builds matched the independent production Go oracle.
The volume inventory had compiler/repository logical counts 143/0; this run
observed 140/3 with the frozen populations and their respective tsconfigs. The
combined logical count remains 143. The observed streams, not an inferred count,
are the agreement evidence. Hashes and compressed complete streams are saved in
[validation-wave-02](validation-wave-02/README.md).

The controls cover repairs versus suggestions, repeated references, computed
static properties, optional chains, grouping and comment guards, React hook
exclusion, call/array/object precedence, any and error returns, any arrays,
contextual signatures, generic nested assignments, sync and async promises,
unknown and explicit annotation exemptions, nullish union removal, never,
generic constraints and indexed-access targets, assignment targets, type
assertion syntax, Unicode and CRLF byte offsets.

## Changes and checker boundary

Each rule is isolated in its own `.a` file, using the existing parser, Rules,
type-graph and diagnostic machinery. `wave_02_suite.a` is a separate runnable
entry point; it does not change another worker's suite. The independent oracle
only selects the three upstream production rules and serializes their results.

`awaited-shape` and `constraint-shape` expose raw checker operations, each in a
new named Go file and a new named `.a` adapter. The only existing tracked source
changed is `bridge/tsgo/checker/facts.go`, with two minimal case/return dispatcher
registrations (four physical Go lines after gofmt). No shared checker algorithms,
submodule pins or protected compiler files changed. See [WAVE_02_FACTS.md](WAVE_02_FACTS.md).

## Mutants actually run

| Mutant | Outcome and detector |
| --- | --- |
| Logical safe assignment changed to suggestion | Compiles, exit 0, 28 findings; Go diagnostic byte mismatch at byte 305 |
| Unsafe sync Promise<any> exemption inverted | Compiles, exit 0, 28 findings; Go byte mismatch at byte 4,775 |
| Nullish identity membership shifted by one | Compiles, exit 0, 31 findings; Go byte mismatch at byte 6,368 |
| Awaited operation returns its input type | Go compilation succeeds; direct checker identity test fails |
| Constraint operation returns its input type | Go compilation succeeds; direct checker identity test fails |
| Released registry entry kept live | Compiles, exit 0; required panic 70 is absent, so released-handle check kills it |
| Input C length off by one | Foundation ASan heap-buffer-overflow |
| Output string length off by one | Foundation ASan heap-buffer-overflow |
| Foundation registry retains released handle | Foundation stale-handle assertion |
| Query type at source-file position | Foundation independent oracle mismatch at byte 6 |
| Link opt-in guard removed | Foundation refusal test |
| C output free removed | Foundation LeakSanitizer leak report |
| Region entry allocates on heap | Foundation LeakSanitizer unowned-result report |
| Node oracle one-byte mutation | TestTheOracleCatchesOneByte passes by rejecting changed stdout |

An initial nullish source mutant was rejected by type checking because it narrowed
an ID to literal 0; that did not count. The final mutant above compiles and is
caught exclusively by comparison. During implementation the byte oracle also
caught an incorrect Never flag on the positive/negative controls; the flag is
corrected to 262144 and pinned by the direct checker test.

## Measured timing

Three alternating count-only rounds, builds excluded, no concurrent validation.
Medians in seconds:

| Corpus | Native process | Go process | Native/Go | Native load | Go load | Native run | Go run |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 5.122489 | 1.286359 | 3.98x | 0.263589 | 0.246980 | 4.848747 | 1.009421 |
| Repository | 0.341273 | 0.143879 | 2.37x | 0.064184 | 0.057239 | 0.270314 | 0.076153 |

These are observed complete-process and reported phase times, not a claim that
native is faster. Native makes 164,709 checker queries on the compiler and
11,873 on the repository. Saved measurements preserve every round, query time,
first-query time, load time and run time. Go run time and native run time have
different instrumentation boundaries; whole-process time is the direct comparison.
No performance optimization is claimed in this unit.

## Commands and environment

`bash cloud/setup.sh` passed: Go 1.27.1 ready (0s), clang 20.1.8 ready (0s), Node
v24.19.0 ready (0s), submodules (0s), build cache (76s), done (76s).
Environment: `source /workspace/adamic-tools/env.sh`; `nproc` printed 5; cgroup
quota is four CPUs. Every test's output was saved to a log, never piped.

* `go test ./stage1/cohere/typeaware -run '^TestWave02AgreementAndMutants$' -count=1 -v -timeout 30m`: PASS, 78.303s, with all three ADAMIC_WAVE02 artifact/manifest variables and ADAMIC_TYPESCRIPT_SOURCE set.
* `TMPDIR=/workspace/wave-02/scratch ADAMIC_TSGO_CORPUS=/workspace/wave-02/typescript go test -v -count=1 -timeout 15m ./bridge/tsgo/...`: PASS; TestBridge 97.55s, checker package 0.280s. 1,600 query positions on four compiler files match 54,982 bytes under sanitizers; C ABI checks 100 queries.
* `go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$'`: PASS, 21.466s. Go's component regex also selects generic_functions and method_closures; the saved log names every fixture.
* `python3 validation-wave-02/checker_mutants.py /workspace/wave-02` (from the typeaware directory): both compiling operation mutants killed.
* Pinned cohere `--no-cache --no-fix` on temporary extension/import adapters for the six modules: exit 0, 276 rules, six files checked, 100% Adamic-ready. The older pinned CLI ignores `.a` directly; that no-op was not counted as a gate.
* `python3 bridge/tsgo/profile/volume_bench.py ... --rounds 3 --corpus compiler ... --corpus repository ...`: twelve count-agreeing runs; medians above.
* `go vet ./...`: exit 0; touched-package vet: exit 0; `gofmt -l cmd internal bridge/tsgo/checker stage1/cohere/typeaware/*.go`: empty; `git diff --check`: exit 0.

## Coverage limits

The port follows production default options. Logical-assignment-operators' custom
`never` and `ifStatements` modes are not exposed by this native runner. There is
no exhaustive fixture/configuration claim beyond these corpora and controls.
The full `go test ./...` gate and all 26 prior rule suites were not rerun; the
new rule suite, all bridge packages and the filtered runtime oracle were run.
The claim inventory is not regenerated or rewritten. Frozen manifests exclude
new wave files. No PR was opened; the branch is intended for Ahra's merge.

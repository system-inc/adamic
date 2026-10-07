Built: three base-family type-aware rules in separate .a files, plus one raw checker question.
Commits: claim 61c406e6, implementation e973a3c326337ea607c7970995444daf27b36ab8; branch codex/typeaware-wave-26.
Commands and outputs: wave comparison PASS 110.576s; bridge PASS 107.130s; filtered Node oracle PASS 99.730s; vet and required gofmt output empty.
Mutants: three rule mutants and the raw-question mutant caught by Go finding bytes; released registry caught by panic expectation; foundation mutants caught by ASan, LSan, byte oracle and link refusal; Node one-byte mutant caught.
Not covered: full repository gate, all previous 26-rule suites, every upstream fixture, suppression/edit application, Go-heap sanitization and pinned cohere CLI lint of .a files.

## Selection and isolation

The base is origin/codex/tsgo-c-library at
0d540f413625f016f20fea39761c7b184f335de6. Cohere remains pinned at
715ba94f3608a6500086b1076ce5cb7e51b836db, and typescript-go at
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. No submodule pin changed.

VOLUME_REPORT.md links the all-family compiler and repository count tables.
Their combined counts are sorted descending, with lexical ties, after excluding
rules ported on the base. The assigned remaining positions are:

| Position | Rule | Recorded combined findings |
| --- | --- | ---: |
| 76 | base/correctness-require-matching-operation-context | 0 |
| 77 | base/correctness-require-matching-provider-return | 0 |
| 78 | base/correctness-require-optional-relation | 0 |

All origin branch references were fetched and their claim files and named ports
checked. A further content search of origin Adamic/TypeScript sources found only
configuration-set entries for these names on three older branches, not rule
implementations. No selected rule was already claimed or ported, so none was
skipped. The claim commit was pushed before implementation code was written.
The unnecessary recursive historical submodule fetch was stopped after the parent
branch references had been fetched; setup independently obtained the pinned
submodules successfully.

The native runner is `wave_26.a`. Existing suites are untouched. Adamic owns
parent selection, decorator scoping, wrapper unwrapping, assignability decisions,
undefined membership, names, messages and diagnostic spans. The independent
`testdata/oracle_wave_26.go` loads and walks its own program and invokes the three
unchanged production registry rules. It imports no bridge implementation.
Complete canonical findings include fix/suggestion counts and edit fields; these
three production rules offer no fixes or suggestions.

The new `type-arguments` question exposes type/alias symbol names and argument
identities, followed by the existing owned graph frame. Its implementations are
`bridge/tsgo/checker/type_arguments.go` and `type_arguments.a`; its separate
contract is in `bridge/tsgo/type_arguments.md`. The only shared source edit is the
one-line registration in `checker/facts.go`. No protected compiler file changed.
The existing property, return-signature, reference and assignability questions
supply the other facts. The new question contains no lint decision.

## Agreement and controls

TypeScript v6.0.3 is pinned at
050880ce59e30b356b686bd3144efe24f875ebc8. The populations are exactly the prior
77 compiler roots and frozen 287 repository roots, including 212 .a files and
75 .ts files. Configured declaration roots are retained. New ports are not added
to the selection population. Portable manifests and input hashes are preserved
in `validation-wave-26`.

| Population | Roots | Findings | Identical bytes, normal and sanitizer |
| --- | ---: | ---: | ---: |
| Compiler | 77 | 0 | 5,318 |
| Repository | 287 | 0 | 18,485 |
| Generated controls | 30 | 27 | 7,023 |

Every normal and sanitizer native comparison has empty stderr. ASan, UBSan and
LeakSanitizer check native memory and the C adapter; they do not instrument the Go
heap. `diagnostic-hashes.json` records each Go, native and sanitizer stream;
compressed canonical outputs and the generated source strings are preserved.
Headers contain this run's absolute paths, so stream hashes describe this run,
not relocated byte streams.

The controls cover all three operation decorators, wrong base types, both
subtype directions, real unions, nested Promise/array/Readonly/nullable wrappers,
inferred and explicit method returns, bare and namespaced branded decorators,
optional brands, any/unknown exemptions, non-method owners, local-class guards,
all three ORM relation decorators, null versus undefined, any/unknown/void/never,
optional fields, computed/literal/private names, Unicode and CRLF positions.
These populations' zero findings and zero timed checker queries mean the corpus
measurements alone do not exercise checker judgments. The positive controls and
mutants do, and no broader finding-volume claim is inferred.

## Mutants and refusals

The final judgment mutants all compile and exit 0 with empty stderr. Only the
independent Go finding-byte comparison catches them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| Operation context | Drop the reverse assignability requirement | 526 |
| Provider return | Reverse source and expected assignability | 2,013 |
| Optional relation | Accept null alongside undefined | 2,768 |
| Raw type arguments | Omit reference argument identities | 55 |

The first provider reversal survived the original controls: unrelated types fail
in both directions, and identical types pass in both. Adding a valid narrower
string return against the optional string brand made the reversal observable.
The final complete run kills it as recorded above. The surviving attempt is
preserved and is not counted as a successful mutant.

A released program queried with `type-arguments` panics with exit 70 and exactly
`adamic: panic: invalid or released checker handle`. Keeping that same program in
the registry makes the same valid question finish with exit 0; the required-panic
expectation catches this mutant. The direct-checker test separately compares
reference and Readonly alias arguments with the checker's own operations and
rejects missing, zero, noncanonical, oversized and extra-field identities.

The bridge regression independently compares 162 positions and 3,261 bytes under
ASan/UBSan/LSan, and checks 100 C ABI queries, output strings surviving release,
zero/stale handles and distinct later handles. Its mutants are all caught:

| Foundation mutant | Catch |
| --- | --- |
| Input byte length +1 | ASan heap-buffer-overflow |
| Output byte length +1 | ASan heap-buffer-overflow |
| Released handle kept live | Stale-handle assertion |
| Query at the source-file position | Independent byte oracle at byte 6 |
| Link opt-in guard removed | Refusal expectation |
| C output frees omitted | LeakSanitizer |
| Region output allocated on the heap | LeakSanitizer |

The filtered Node oracle also catches its one-byte mutant and passes eight native,
JavaScript and Node fixtures with sanitizer/leak checks. The subtest filter
additionally selects `method_closures` and `generic_functions`.

## Toolchain, attempts and commands

`bash cloud/setup.sh` passed. Go 1.27.1 ready at 0s; clang 20.1.8 ready at 1s;
Node v24.19.0 ready at 1s; submodules ready at 1s; build cache warm and setup done
at 135s. `nproc` is 5, cgroup cpu.max is `400000 100000`, and memory is 17.6 GB.
Each build/test shell sourced `/workspace/adamic-tools/env.sh`.

The first native build hit the documented nested-constructor gap. Constructing
`DecoratorScope` in the runner before handing it to the rules avoids the gap;
no compiler source workaround was used. The first control config discovered no
.a inputs and the Go loader refused it. An explicit .a file root fixed that
configuration error. The initial direct-checker assertion compared an internal
invalid-UTF-8 symbol name with the wire's replacement text; the assertion now
uses the documented replacement conversion. Attempt logs are preserved.

The final comparison command was:

```sh
ADAMIC_WAVE26_ARTIFACTS=/workspace/wave-26-validation-final \
ADAMIC_WAVE26_REPOSITORY_MANIFEST=/workspace/wave-26-repository.manifest \
ADAMIC_WAVE26_COMPILER_MANIFEST=/workspace/wave-26-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-26-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave26AgreementAndMutants$' \
  -count=1 -v -timeout 30m > /tmp/wave-26-test-complete.log 2>&1
# PASS, 110.576s

TMPDIR=/workspace/wave-26-bridge-scratch \
go test ./bridge/tsgo/... -count=1 -timeout 15m -v \
  > /tmp/wave-26-bridge.log 2>&1
# PASS, bridge 107.130s, checker 0.156s

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout 10m -v > /tmp/wave-26-node-oracle.log 2>&1
# PASS, 99.730s, eight selected fixtures and the one-byte mutant

go vet ./... > /tmp/wave-26-vet-final.log 2>&1
# exit 0, empty log

gofmt -l cmd internal > /tmp/wave-26-gofmt.txt
# empty output; the four new Go files also produce empty gofmt output
```

All test output goes to files. The standalone checker run also passes in 0.222s.
The pinned cohere CLI's attempted `.a` lint run refuses all six paths because
that CLI cannot place the extension in its program. The attempted format-only
run checks no types or lint and is not claimed as successful source validation.
Native compilation does check and compile the .a sources. Extending the pinned
CLI is outside this unit; no shared CLI, parser, emitter or lowerer was edited.

## Quiet cost observation

After builds and tests completed, three alternating count-only rounds measured
whole-process load/parse/lint throughput with no competing build or test. Counts
remain zero on both sides. Complete diagnostic streams had already matched in
the preceding normal and sanitizer runs. Medians, seconds:

| Corpus | Implementation | Load | Run | Whole process |
| --- | --- | ---: | ---: | ---: |
| Compiler | Native | 0.328 | 1.415 | 1.762 |
| Compiler | Go | 0.313 | 0.042 | 0.397 |
| Repository | Native | 0.090 | 0.175 | 0.272 |
| Repository | Go | 0.083 | 0.058 | 0.163 |

Native whole process is 4.44x Go on compiler and 1.67x Go on repository. These
are observations of the three-rule runner, not a speedup claim or a measurement
of the new question's latency. Both corpora produce zero native checker queries.
Separate medians need not add to the whole-process median. Raw nanoseconds,
round order, phases and outputs are in `validation-wave-26`.

```sh
python3 bridge/tsgo/profile/volume_bench.py \
  /workspace/wave-26-validation-final/wave26 \
  /workspace/wave-26-validation-final/wave26-oracle /workspace/wave-26-bench \
  --corpus compiler /workspace/wave-26-typescript/src/compiler/tsconfig.json /workspace/wave-26-compiler.manifest \
  --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-26-repository.manifest \
  > /tmp/wave-26-bench.log 2>&1
```

The full repository test gate and previous 26-rule agreement suites were not
rerun. The new suite, complete bridge/checker packages, filtered Node oracle and
vet are the actual gate. The runner emits proposed diagnostics under production
defaults; it does not apply edits or implement suppression/CLI behavior. Every
upstream fixture and nondefault project configuration was not covered. No pull
request is opened.

# Type-aware wave 08

Built: no-loop-func, nexus/correctness-no-import-cycle-load-time-read, and @typescript-eslint/no-require-imports in separate .a files, with three fact-only bridge questions.
Commits: claim ddc55032 pushed before implementation; port a015e718 based on 0d540f413625f016f20fea39761c7b184f335de6; evidence and this report follow in a separate commit.
Checks: wave tests PASS 236.294s; bridge gate PASS 136.296s; checker PASS 0.273s; filtered Node oracle PASS 22.943s; go vet, gofmt and formatting idempotence passed.
Mutants: all three rule mutants compiled, exited zero with empty stderr, and failed only byte comparison; three fact mutants and the released-handle mutant were caught; foundation ownership, length, position and linkage mutants passed their negative checks.
Not covered: the full repository gate, arbitrary rule options, JSX populations, or exhaustive TypeScript syntax; the pinned cohere CLI cannot lint .a paths.

## Claim and implementation

The claim records positions 22, 23 and 24, their original compiler/repository
volumes (9/2, 10/0 and 9/0), and the search across fetched origin stage1 trees.
None was already ported or claimed in those fetched refs, so none was skipped.
The requested bridge branch overrides the general unit instruction to use main.
The claim was pushed before any implementation was written.

`wave08_suite.a` creates one checker program for the manifest and releases it
once. The rules use the existing native parser and diagnostic representation.
`no_loop_func.a` handles loop ancestry, writes, bindings, IIFEs and escaping
closures. `no_import_cycle_load_time_read.a` builds the runtime module graph and
checks unsafe initialization reads. `no_require_imports.a` distinguishes ambient
require declarations from local bindings and handles import-equals declarations.
The oracle invokes the three production cohere rule implementations unchanged.
No production lint verdict runs through the bridge.

Three new Go/Adamic file pairs implement raw symbol metadata, syntax metadata and
resolved module links. The only changes to existing shared bridge files are one
registration line in facts.go and one dispatcher return line in
declaration_facts.go. The question contract is in bridge/tsgo/wave08-facts.md.
No protected compiler file was edited. Every new Adamic source is .a.

## Observed agreement and sanitizer results

The frozen compiler population is 77 src/compiler files from TypeScript v6.0.3,
commit 050880ce59e30b356b686bd3144efe24f875ebc8. The repository population is the
existing frozen 287-file manifest (212 .a, 75 .ts). Portable original manifests,
source hashes, compressed complete finding streams and finding hashes are in
validation-wave08. Absolute paths in streams identify this execution workspace;
portable manifests preserve the population for another checkout.

| Population | Findings | Complete bytes | Normal | ASan/UBSan/LSan |
| --- | ---: | ---: | --- | --- |
| 30 controls | 31 | 17,426 | identical | identical |
| compiler 77 | 28 | 20,884 | identical | identical |
| repository 287 | 2 | 19,558 | identical | identical |

The serialized comparison includes diagnostic namespace, rule, message ID,
message text, byte bounds, fixes and suggestions. These three rules emit no
fixes or suggestions under their default configuration; their empty payloads
also compare. Controls exercise unsafe/safe captures, binding writes,
destructuring, nested closures, named IIFEs, async/generator syntax, ambient and
shadowed require, Unicode/CRLF, runtime/type-only/elided imports, cycles,
namespace reads, reexports, class heritage/static initialization and .server
exclusions. Orphan speculative parser nodes and zero-width compiler syntax are
excluded from exact-node requests.

The bridge gate separately compared 1,600 positions in four compiler files:
54,982 identical bytes under ASan/UBSan/LSan. C ABI checks cover 100 queries,
output survival after release, zero/stale handles, and distinct new handles.
The native wave-specific released-handle probe panicked with exit 70 and exactly
`adamic: panic: invalid or released checker handle`.

## Mutants actually run

| Mutation | What caught it |
| --- | --- |
| no-loop-func requires more than one unsafe name | Go finding byte oracle, first difference 48 |
| no-require-imports reverses ambient declaration test | Go finding byte oracle, first difference 7,353 |
| cycle rule reverses uninitialized predicate | Go finding byte oracle, first difference 12,238 |
| raw symbol declaration Ambient flag flipped | direct checker declaration comparison |
| syntax IsTypeNode predicate inverted | direct checker syntax comparison |
| resolved module target emptied | direct checker resolved-module comparison |
| released program retained in registry | expected native panic; mutant instead exits 0 |
| foundation input length off by one | ASan heap-buffer-overflow |
| foundation output length off by one | ASan heap-buffer-overflow |
| foundation released handle kept live | C stale-handle assertion |
| foundation type queried at source-file position | oracle mismatch at byte 6 |
| foundation link guard removed | unlinked-call refusal test |
| foundation C output free removed | LeakSanitizer |
| foundation region entry allocates on heap | LeakSanitizer |
| Node oracle one byte changed | TestTheOracleCatchesOneByte |

Rule mutants have successful compilation, exit zero and empty stderr. Their
failure comes from comparing findings, not a compiler error or sanitizer crash.
Fact mutants compile too; each fails its independent checker comparison.
Mutant binaries and archives were removed; tests create fresh overlays on rerun.
Control module order comes from a Go map, so cycle mutant's first differing byte
may vary across runs while the complete comparison still catches it.

## Timing observations

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build-cache warm 132s, done 132s. nproc is 5; CPU quota is four cores.
Go 1.27.1, clang 20.1.8 and Node 24.19.0 were used. Setup succeeded.

Three rounds, alternating native/Go order, after validation jobs completed;
complete outputs compared on every run. These are process wall times including
checker loading, native parsing, rule execution, formatting and output, excluding
builds and sanitizers. Raw timings and stderr phase counters are preserved.

| Population | Go median | Native median | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 0.851077s | 15.302608s | 17.9803 |
| repository | 0.187816s | 0.844608s | 4.4970 |

Observation: native is slower on both populations. One compiler round issued
1,099,131 bridge queries, with 3.571s measured query time and 14.430s native run
time. Inference: reducing repeated metadata queries could improve performance;
that optimization is not implemented or measured here.

## Commands and evidence

All test output was redirected to log files, never piped. After sourcing
/workspace/adamic-tools/env.sh:

```sh
bash cloud/setup.sh > /workspace/typeaware-wave-08-setup.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus \
ADAMIC_WAVE08_COMPILER_MANIFEST=/workspace/wave08-corpora/compiler.manifest \
ADAMIC_WAVE08_REPOSITORY_MANIFEST=/workspace/wave08-corpora/repository.manifest \
ADAMIC_WAVE08_ARTIFACTS=/workspace/wave08-validation \
ADAMIC_WAVE08_FACT_ARTIFACTS=/workspace/wave08-fact-mutants \
go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m \
  > /workspace/wave08-final-test.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/typescript-wave08-corpus \
go test ./bridge/tsgo/... -count=1 -v -timeout 30m \
  > /workspace/wave08-bridge-test.log 2>&1
go test ./internal/oracle -count=1 -v \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  > /workspace/wave08-node-oracle.log 2>&1
go vet ./... > /workspace/wave08-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware \
  > /workspace/wave08-gofmt.log
python3 stage1/cohere/typeaware/testdata/time_wave08.py \
  --artifacts /workspace/wave08-validation \
  --compiler-config /workspace/typescript-wave08-corpus/src/compiler/tsconfig.json \
  --repository-config /workspace/adamic/tsconfig.json \
  --compiler-manifest /workspace/wave08-corpora/compiler.manifest \
  --repository-manifest /workspace/wave08-corpora/repository.manifest \
  --output /workspace/wave08-timings.json
```

The portable timing script uses the same three-round procedure as the measured
workspace script. It validates equality across every timed run. To relocate the
corpora, prefix compiler manifest paths with the TypeScript checkout and repository
manifest paths with the Adamic checkout.

The pinned cohere CLI refuses explicit .a lint paths as outside its TypeScript
program, and format-only does not discover them. A small Go overlay wrapper calls
its production JavaScript formatter directly on .a paths; all seven sources were
formatted and a second pass produced no changes. No CLI lint success is claimed.
The old bridge also does not resolve .a-to-.a module imports; module controls use
.ts files, while .a roots are supported and covered. Custom no-require-imports
allow patterns/options are not implemented; the corpus uses the default rule.
The prior 26-rule suites and full repository gate were not rerun. Existing checker
question regression tests, the complete bridge gate, touched-package tests and
five filtered Node oracle fixtures were run instead.

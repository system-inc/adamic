Built: three continuation rules in native Adamic, with isolated checker questions and no shared harness edits.
Commits: claim `940df421`; implementation `99c9e8d3`; original three completed through `67b15e35`.
Commands and outputs: byte oracle PASS on both corpora and controls; sanitizers PASS; native/Go medians compiler 1.815s/0.335s, repository 0.266s/0.118s.
Mutants: timer end+1 byte 82, process end+1 byte 79, blocking end+1 byte 9245; all exit 0, caught only by comparison; retained released handle caught by expected panic.
Not covered: full repository gate, emitted-JavaScript bridge execution, options/suppression matrices; shared registration awaits integration.

## Delivered rules

- `nexus/correctness-no-process-exit-after-output`
- `nexus/correctness-no-uncleared-race-timeout`
- `nexus/correctness-require-blocking-standard-streams`

All native implementations and decoders are `.a`. Rule decisions, declaration
classification, the import closure and control-flow walk are native. Go supplies
raw symbol/declaration ancestry, selected signature bodies, resolved module
syntax, and top-level parse context. It supplies no lint verdict or edits.

The private adapter regenerates a copy of the production C archive entry point
inside ignored `_adapter/`, replacing its one checker dispatch call with
`InspectWave04Next`. The production C boundary and handle registry stay identical.
Four new questions dispatch in the owned Go file. No shared generator, harness,
archive entry point or protected compiler file was edited for this continuation.
A merger can register these four questions in the shared dispatcher and wire the
three rule classes into its combined suite. The private suite is runnable now.

## Observed byte comparisons

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Timer controls | 26 | 12 | 8588 |
| Process controls and helpers | 59 | 39 | 25214 |
| Blocking-stream fixtures | 45 programs | 117 | See individual hashes |
| TypeScript src/compiler | 77 | 0 | 6088 |
| Frozen repository | 287 | 0 | 18485 |

Every row passes both normal and ASan/UBSan builds, with empty native stderr.
LeakSanitizer is included in the sanitized native run. Go's own runtime is not
instrumented. All 168 positive-control findings include complete diagnostic,
fix and suggestion fields. These three production rules have no fixes or
suggestions; the corresponding zero counts match.

The independent Go oracle invokes the unmodified pinned production rules with
its own program loader and AST walk, importing no bridge implementation. The
blocking fixtures export the production test tables and all assembled real-site
programs through an isolated Go overlay, preserving complete multi-file inputs.
An extra recursive call-order case checks the cache's source-order behavior.
The process controls include all 48 extracted source arrays, eight dynamically
assembled real-site examples and an optional-chain example. Timer controls are
17 extracted examples plus nine adversarial examples.

Evidence includes every run's stderr, comparison hashes, complete compressed
large oracle outputs, smaller raw outputs, control inputs, source hashes, corpus
input hashes, and build/test logs in [evidence-final](evidence-final).

## Mutants and released handles

Each rule span mutant changes only its reported end by one byte. Every mutant
build succeeds with empty build stderr, then exits 0 with empty run stderr.
Only comparison against Go kills it, at bytes 82, 79 and 9245 respectively.
The released-handle control exits 70 with no stdout and the exact 50-byte
`adamic: panic: invalid or released checker handle` line. Its registry mutant
retains the handle, builds and exits 0 with 41 stdout bytes and empty stderr;
the expected-panic assertion catches it. Mutation changes only generated owned
adapter source and restores it before returning.

## Reproduction and checks

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_next/extract_process.py /workspace/typeaware-wave-04-next/process-controls-ts > /workspace/typeaware-wave-04-next/process-ts-extract.log 2>&1
python3 stage1/cohere/typeaware/wave_04_next/extract_blocking.py /workspace/typeaware-wave-04-next/blocking-controls-final > /workspace/typeaware-wave-04-next/blocking-final-extract.log 2>&1
python3 stage1/cohere/typeaware/wave_04_next/validate.py /workspace/typeaware-wave-04-next/release-final --compiler-manifest /workspace/typeaware-wave-04/compiler.manifest --compiler-config /workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json --repository-manifest /workspace/typeaware-wave-04/repository.manifest --blocking-controls /workspace/typeaware-wave-04-next/blocking-controls-final --process-controls /workspace/typeaware-wave-04-next/process-controls-ts > /workspace/typeaware-wave-04-next/release-final-validation.log 2>&1
python3 stage1/cohere/typeaware/wave_04_next/benchmark.py /workspace/typeaware-wave-04-next/release-final /workspace/typeaware-wave-04-next/benchmark-final --compiler-manifest /workspace/typeaware-wave-04/compiler.manifest --compiler-config /workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json --repository-manifest /workspace/typeaware-wave-04/repository.manifest > /workspace/typeaware-wave-04-next/benchmark-final.log 2>&1
go vet ./... > /workspace/typeaware-wave-04-next/go-vet-final.log 2>&1
go test -count=1 -timeout 30m ./bridge/tsgo/checker ./bridge/tsgo/archive > /workspace/typeaware-wave-04-next/go-test-bridge-final.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(array_methods|map|template|try).*|TestTheOracleCatchesOneByte' > /workspace/typeaware-wave-04-next/go-test-oracle-filtered.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run '^TestTheOracleCatchesOneByte$' > /workspace/typeaware-wave-04-next/go-test-oracle-mutant.log 2>&1
gofmt -l cmd internal bridge/tsgo/checker/wave04_next*.go > /workspace/typeaware-wave-04-next/gofmt.log
```

Validation ends `PASS three claimed rules`. Checker tests PASS in 0.135s;
archive has no tests. Filtered native/Node/emitted-JavaScript oracle PASS in
15.015s; its one-byte mutant PASS in 3.364s. Vet and formatting logs are empty.
The full gate was not run. New lint-suite emitted-JavaScript execution is not
covered; these results are the native checker bridge compared with Go.

Setup already passed for this branch: Go ready 0s, clang ready 0s, Node ready
0s, submodules 0s, cache warm 77s, done 77s. `nproc=5`; the CPU quota is four
cores. Go 1.27.1, clang 20.1.8 and Node 24.19.0. Every toolchain command sources
the printed environment. Benchmarks are three quiet wall-clock samples after
validation, including loading and parsing; native is slower on both populations.

## Resolved blockers and remaining integration

The historical dispatch authorization blocker is superseded by the owned
adapter. Git credentials also recovered; prior blocker commits were pushed.
Adamic's ownership-cycle analysis stalled on a recursive foreign-file cache;
removing the new cache made the build succeed without compiler edits.
Top-level `await import(...)` initially failed the shared parser. The owned
`source_context.a` adapter applies raw compiler context and top-level await
presence to the parser's public context switch, preserving source text and
positions. A destructuring ancestor initially made Go's `Node.Text` panic;
raw names now serialize only textual name kinds.

The pinned consumer checker does not resolve `.a` module specifiers. An explicit
`.a` target control exposed this; using `.ts` fixture helpers recovered the
production helper findings. The original Go fixture source/module names are
therefore preserved for those helper modules outside the repository. All new
Adamic implementation files remain `.a`. Supporting `.a` consumer modules and
shared suite registration remains the integration worker's territory. Claims
here refer to the shipped private suite, not new registrations in the shared CLI.

## Final private-dispatch rejection check

The private adapter now rejects unknown `wave04-next-` questions directly,
rather than entering the production dispatcher while still holding its checker
lease. The added unknown-question control passes with all checker tests in
0.090s. A nil-error-return mutant compiles and fails only the assertion
`unknown private question must reject: <nil>`; both logs are preserved.
This last change affects the unsupported-question path. Corpus artifacts above
record the supported-question suite before this rejection-only adjustment;
the supported four dispatch cases were not altered.

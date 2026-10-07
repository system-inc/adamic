# Wave 27, second batch

The original three parity rules were completed and pushed in `5da2444b` before
this batch was selected. The next claim was committed and pushed in
`d095b4f795a8057c500f2ac62d76321307823669`, before implementation. This batch adds:

- `nexus/correctness-no-process-exit-after-output`
- `nexus/correctness-no-uncleared-race-timeout`
- `nexus/correctness-require-blocking-standard-streams`

After `git fetch --no-recurse-submodules origin
'+refs/heads/*:refs/remotes/origin/*'`, all 325 fetched origin refs were inspected.
The 197 checker-dependent names from VOLUME_REPORT's linked compiler/repository
counts were ranked by descending sum and lexical ties. Port files and bundled
rule names on main and codex/tsgo-c-library and every Markdown claim under
stage1/cohere/typeaware/claims on every origin ref were excluded. These were the
first three available names, all with combined volume zero. Unqualified candidate
names were also checked in claims. The claim file records this selection.

## Implementation and boundaries

Each rule is in its own `.a` file in this owned directory. Helpers and the standalone
`suite.a`, independent `oracle.go.txt`, control generator and `validate.py` are
local to this batch. No shared harness, registration generator, compiler emitter,
lowerer or submodule pin was changed. The only shared bridge edit is four physical
one-line dispatch additions in `bridge/tsgo/checker/facts.go`, one per question.

The new question files on both sides are named for their wave-specific protocols:

| Question | Raw facts |
| --- | --- |
| wave-27-declaration-ancestry | Own/aliased symbol identities, declaration file/default-library/module facts, ancestor kinds, name kinds/text, flags and global-augmentation marker |
| wave-27-call-declaration | Resolved signature declaration location, function flags, body/module presence and return-type flags |
| wave-27-module-sources | Program source texts and declaration/module markers; import kind/type-only/literal flags and resolved target filenames |
| wave-27-syntax-flow | Generic syntax graph blocks, reachability, ordered expression/statement locations and successor indices |

The raw graph implementation is copied from the pinned cohere generic
control_flow_graph shelf, with its rslint/MIT provenance retained and local
symbols prefixed to avoid collisions. Graph construction runs in the Go bridge;
platform membership, callee-write judgments, entry selection, import closure,
blocking judgments, path states, findings and diagnostic rendering run in native
Adamic. No Go cohere rule, policy or lint-verdict helper is called by the bridge.
The production-rule oracle imports no bridge implementation.

Wave 23 already had a `declaration_ancestry.go` on origin. This batch's protocols,
Go methods, Go files and Adamic decoders use a `wave-27` namespace to avoid that
collision. The initial namespace transition interrupted an in-progress preliminary
mutant run; that log is preserved and is not passing evidence.

The race port preserves default-library Promise/race checks, global timer
identity, const/await-using resolution, nested-function exclusion, lost-handle
contexts and symbol-based reads including shorthand values. The stream ports
preserve NodeJS.Process declaration ancestry, ambient console identity, quoted
and parenthesized console keys, one-level synchronous/awaited callee following,
try/catch write-state rollback, all graph paths, import closure, load-time blocking,
started local functions/callbacks/generators, await barriers and first-exit/count
messages. Duplicates, byte spans and the production message text are preserved.
These three production rules emit no fixes or suggestions; their zero fields are
compared, rather than omitted from the protocol.

## Independent verification

The oracle has an independent loader and Go AST walk and invokes the unchanged
production registry Run methods with the correct program views and file cache.
It preserves configured declaration roots and accepts explicit `.a` roots. All
canonical fields are compared: file headers, UTF-8 ranges, full rule names,
message IDs/text, fixes, suggestions and their edits, duplicates and counts.

There are 93 generated control files with 101 findings: process exits after
writes, branch/loop/switch/label paths, try/catch/finally, exits in arguments,
callbacks, awaited and generator callees, one-level external helpers, ambient
aliases and lookalikes, readonly const functions, imports and blocking order,
load-time blockers, shebang entries, all console method names, Unicode/CRLF,
inline/named promises, globalThis/window timers, dropped/void/assigned/chained
handles, read and shadowed bindings, shorthand reads, duplicate race elements,
destructuring ancestors, quoted namespaces and await-using declarations.

All authored implementation and control bodies are `.a`. Three scratch-only
`.ts` symlinks point to `.a` fixture bodies so the pinned stock module resolver
can exercise `.js` imports and the production StandardStreams filename suffix.
No `.ts` Adamic implementation or fixture body was authored or committed. This
is an isolated fixture adaptation; it does not change the shared module loader.
The native rule modules themselves compile directly as `.a`.

Frozen populations: all 77 TypeScript src/compiler files and the existing
287-file repository manifest. Pins remain cohere
`715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, and TypeScript
`050880ce59e30b356b686bd3144efe24f875ebc8`.

An actual compiler-corpus failure exposed a raw metadata defect: ancestor names
can be destructuring binding patterns, and Node.Text() panics on those. The
question now sends the name kind and only reads Text() for supported name kinds.
The final controls include that shape, and the complete compiler comparison
must pass. The failure log is preserved, not reported as passing evidence.

An extra name-kind mutant initially left a Go variable unused and failed to
compile. That is not a killed semantic mutant. The corrected mutant uses the
variable while always returning the wrong name kind, so it must compile, exit
0 with empty stderr, and fail only the finding-byte comparison.

## Commands and limits

Toolchain setup from the first batch remains in use: Go 1.27.1, clang 20.1.8,
Node v24.19.0; setup go/clang/node/submodules 0s, cache warm 116s, total 116s;
`nproc` 5, cgroup cpu.max `400000 100000`, 17.6 GB. Every command sources
`/workspace/adamic-tools/env.sh`. All test output goes directly to files.

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE27_NEXT_ARTIFACTS=/workspace/wave-27-scratch/next-verified \
ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest \
ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript \
python3 stage1/cohere/typeaware/wave_27_next/validate.py > wave-27-next-verified.log 2>&1
```

The validator builds the native compiler, normal and Go-ASan C archives, optimized
and sanitized native runners, and independent production oracle. It runs all
three populations under normal and ASan/UBSan/LeakSanitizer builds, three native
rule mutants, four question mutants and an additional raw-name-kind mutant. Every
semantic mutant must compile, exit 0, have empty stderr and differ from Go bytes.
It separately queries every new question after releasing its program (required
panic 70), and repeats all four with registry deletion removed (required-panic
check catches exit 0). Three alternating complete Go/native runs per corpus
verify every output stream and record whole-process nanoseconds and phase stderr.

The owned validator adds no shared profile/harness support. The full repository
gate, all upstream fixture/options matrices and emitted-JavaScript rule comparison
were not run. Source modules were formatted by the pinned house formatter via a
TypeScript parser hint, writing the original `.a` files. Stock-cohere discovery
and `.a` import resolution remain the shared harness worker's task; a clean stock
source-lint result is not claimed. Native compilation, production Go diagnostic
comparison and sanitizer runs are the validation of these actual `.a` ports.

Evidence lives in this directory's `evidence/`: full logs, normal/Go/sanitized
canonical streams, normally exiting mutant streams, release probes, timing JSON,
relative manifests, selection audit and source/evidence hashes. Output headers
contain scratch paths, so hashes and byte lengths describe this run, not relocated
bytes. Binaries, archives, fixture symlinks and downloaded corpora are not committed.

## Final gate and measurements

`validate.py`: **PASS**. Normal and sanitized complete streams:

| Population | Files | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Controls | 93 | 101 (52 process, 36 blocking, 13 race) | 62,573 |
| Repository | 287 | 0 | 18,485 |
| Compiler | 77 | 0 | 5,934 |

All semantic mutants compiled, exited 0 and produced empty stderr. Only the
independent full-byte comparison caught them:

| Mutant | Changed fact or decision | First differing byte |
| --- | --- | ---: |
| Process exit | Written-state test inverted | 84 |
| Race timeout | Lost-handle test inverted | 52,486 |
| Blocking streams | Suppress every entry finding | 585 |
| Declaration ancestry | Every declaration marked non-declaration-file | 60 |
| Syntax flow | Every block marked unreachable | 1,330 |
| Resolved call | Every return type marked never | 12,520 |
| Module sources | Every import marked type-only | 42,269 |
| Ancestry name kind | Every declaration name marked Identifier | 61,952 |

Each of the four new questions after release exits 70 with exactly
`adamic: panic: invalid or released checker handle`. With the released handle
retained in the registry, each query exits 0 and fails the required-panic check.

Three alternating complete-process runs were isolated from builds and other
tests. Every timed output was compared again; medians include loading, native
parsing/walks, facts, rendering and process startup/teardown:

| Corpus | Native median | Go median | Native / Go | Native queries/run |
| --- | ---: | ---: | ---: | ---: |
| repository | 445.766 ms | 148.092 ms | 3.010x | 1000 |
| compiler | 2901.653 ms | 391.767 ms | 7.407x | 33 |

Native is slower. Zero corpus findings do not mean zero questions: symbol
lookalikes still need checking. These whole-process results are not bare C-call
latency measurements. Raw load/run/query timing lines, all samples and output
hashes are preserved. No performance-parity claim is made.

## Focused regression results

After quiet timings completed, the touched checker/bridge packages and the
original three ports were tested together:

```bash
ADAMIC_WAVE27_ARTIFACTS=/workspace/wave-27-scratch/first-batch-regression \
ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest \
ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript \
go test -v -count=1 -timeout 20m ./bridge/tsgo ./bridge/tsgo/checker \
  ./stage1/cohere/typeaware \
  -run '^(TestBridge|TestTSGoRequiresLink|TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestWave27AgreementAndMutants)$' \
  > wave-27-next-regression-corrected.log 2>&1

go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > wave-27-next-vet.log 2>&1
gofmt -l bridge/tsgo/checker/wave_27_*.go > wave-27-next-gofmt.log
```

PASS: bridge 95.635s, checker 0.203s, original three-rule gate 89.136s.
Vet and formatting checks produced empty logs. The shared dispatch additions
retain one physical line per question, as requested; formatting checks target
the new Go files. An earlier mistyped timeout command did not execute tests;
its failure is preserved and is not passing evidence.

The bridge gate additionally kills input/output length +1 with ASan,
retained handles with the stale-handle assertion, a source-file type with the
independent byte oracle (byte 6), removed link opt-in with expected refusal,
and missing C output frees/heap-backed region allocation with LeakSanitizer.
The first batch's ORM, Serializable and Verify-array mutants and released-registry
mutant all passed their required failure expectations again. These regression
process timings overlap package builds and are not benchmark figures.
The filtered Node oracle already passed in the first batch (87.821s; eight fixtures
and its one-byte mutant); no compiler/runtime source changed in this batch.

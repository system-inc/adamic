Built prefer-const, radix and non-nullable-type-assertion-style as separate native Adamic .a ports.
Claim commit: 7534c1b8; implementation is the subsequent commit on codex/typeaware-wave-09.
Final wave gate PASS 112.982s; checker PASS 0.300s; bridge PASS 131.146s; filtered Node oracle PASS 19.482s.
All three rule mutants, the type-parameter mutant, direct-fact mutant, released-handle mutant and seven bridge mutants were caught.
Not covered: the full repository gate, old 26-rule corpus rerun, nondefault options, full upstream fixtures or JSX populations; pinned cohere CLI rejects named .a lint inputs.

## Selection and ownership

Base: `origin/codex/tsgo-c-library`,
`0d540f413625f016f20fea39761c7b184f335de6`. The requested base takes precedence
over the generic main-branch instruction. Cohere remains pinned at
`715ba94f3608a6500086b1076ce5cb7e51b836db`, and typescript-go at
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. No submodule pin changed.

VOLUME_REPORT.md links the all-family checker-dependent compiler/repository
counts in validation-volume. Summing those counts, removing the base's 26 ports,
and sorting descending with lexical ties gives:

| Remaining position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 25 | prefer-const | 6 | 1 | 7 |
| 26 | radix | 4 | 3 | 7 |
| 27 | @typescript-eslint/non-nullable-type-assertion-style | 4 | 0 | 4 |

Every fetched origin branch's stage1/cohere source, registration and claim files
was inspected, deduplicating blobs. Inventory/config names were distinguished
from implementations. None of these rules was ported or claimed; none was
skipped. The claim was committed and pushed before implementation began.

The new driver is `wave_09_suite.a`. It loads one checker program, parses each
file once with Adamic's parser, and runs the three new rule classes. Their
judgments, messages, ranges, fixes and suggestions run in native Adamic. The
independent `testdata/oracle_wave_09.go` calls unchanged production Go cohere
rules with its own loader, AST walk, file cache and program views. It imports no
bridge code. Complete canonical diagnostic lines sort per file, preserving
fix/suggestion order and duplicates. Proposed edits are compared, not applied.

The only shared-file change is one registration line in checker/facts.go.
The new `declaration-file-flags` question has its own Go implementation, Adamic
decoder and protocol document. It exposes checker declaration-file flags in
order, including absent and zero-declaration symbols. It supplies no lint
judgment. Existing raw type shapes already expose interned identities and type
parameter constraints. No protected compiler file was edited. Every new Adamic
source, including generated controls and final mutant copies, has extension .a.

## Byte agreement and failures found

Final comparisons use TypeScript v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`, obtained from that commit's codeload
archive, and the branch's frozen 287-file repository manifest. Configured
declaration roots are retained on both sides. The compiler manifest has 77 roots;
the repository manifest has 212 .a and 75 .ts roots. New port files do not alter
these populations. Portable manifests and source hashes are committed.

| Population | Findings | Identical full-stream bytes |
| --- | ---: | ---: |
| Generated controls, 18 files | 44 | 14,510 |
| Compiler, 77 files | 14 | 12,783 |
| Repository, 287 files | 4 | 19,865 |

Every row passes for both ordinary native and ASan/UBSan/LeakSanitizer builds,
with empty native stderr. Equality includes file headings, UTF-8 ranges, rule and
message IDs, message text, fixes, suggestion descriptions and every edit. The
compressed Go/native/sanitizer streams and matching hashes are in
[validation-wave-09](validation-wave-09). Absolute file headings affect the
hashes and stream sizes when artifacts are relocated.

Controls exercise initialized and later-assigned lets, shadowing, nested
function writes, mixed destructuring, omitted slots, assignment patterns and
rest, loop heads, reads before assignment, radix spellings/shadowing/spreads,
signed and hexadecimal numbers, trailing commas, type identity, any/unknown,
generic constraints, angle/as assertions, parentheses, Unicode and CRLF.

The first corpus run exposed an omitted destructuring slot passed to name lookup;
it panicked rather than completing with silently missing findings. The walker
now skips absent names, with an explicit holes control. The first compiler
comparison then found an extra assertion-style finding on createQueue's
`elements[headIndex] as T`: an incorrect hardcoded type-parameter flag treated
unconstrained T as non-nullish. The flag is now pinned to the actual checker
value, 524288, with unconstrained/constrained controls and a caught runtime
mutant. Initial disagreement logs are preserved.

Stage 0 also refused an optional call and an inferred empty never array during
the first builds. Explicit checked lookups and typed empty arrays resolve those
refusals without changing the compiler. A direct-checker harness initially had
no config roots because its input was .a; adding a declaration root made the
config valid while preserving the explicit .a program root.

## Mutants

Each rule mutant compiles, exits 0 with empty stderr, and is caught only by
comparison with independent Go diagnostic bytes. The final controls catch:

| Mutant | Mutation | First differing byte |
| --- | --- | ---: |
| prefer-const | Allow initialized bindings even when written | 44 |
| radix | Treat base 36 as invalid | 9,817 |
| non-nullable-type-assertion-style | Add one to the fix end | 12,054 |
| type-parameter flag | Replace 524288 with incorrect 262144 | 12,634 |

The declaration-file fact mutant inverts `IsDeclarationFile`; its direct-checker
wire comparison fails on global parseInt. It compiles and runs: a build error
is not counted as a kill. Querying the new question after program release must
panic 70 with `invalid or released checker handle`. The registry mutant keeps
that released program and exits 0, caught by the required panic expectation.

The unchanged foundation bridge tests additionally catch every following mutant:

| Bridge mutant | What catches it |
| --- | --- |
| Input length plus one | ASan heap-buffer-overflow |
| Output string length plus one | ASan heap-buffer-overflow |
| Released handle retained | Stale-handle assertion |
| Type queried at source-file position | Independent Go byte comparison, byte 6 |
| Link opt-in guard removed | Refusal expectation |
| C output free omitted | LeakSanitizer |
| Region result allocated on heap | LeakSanitizer |

The C ABI checks 100 queries, Unicode, buffers surviving release, zero/stale
handle rejection and distinct handles. The independent foundation sample checks
162 positions and 3,261 identical bytes under ASan/UBSan/LSan. Sanitizers cover
native/C memory, not the Go heap.

## Native time against Go

Three alternating rounds ran after builds and tests finished, with full
finding output to files and byte equality checked on every round. Native bridge
phase timing was enabled; Go's oracle always records loader and rule timing.
These whole-process measurements include loading, parsing, linting, sorting,
rendering and teardown, excluding builds and sanitizer runs.

| Population | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 5.188 s | 0.973 s | 5.33 |
| Repository | 0.709 s | 0.257 s | 2.75 |

Compiler native rounds range from 2.481 to 5.593 seconds; Go rounds from 0.759
to 1.266 seconds. That variation limits precise speed conclusions. Native remains
slower; this unit makes no performance improvement claim. Native makes 15,678
compiler and 1,938 repository queries per run. Raw nanoseconds, phase lines,
stream hashes and all rounds are in measurements.json and adjacent stderr logs.
The benchmark script is preserved in the evidence directory.

## Commands, environment and scope

`bash cloud/setup.sh` passed: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s,
Node v24.19.0 ready 0s, submodules ready 0s, build cache warm 88s, done 88s.
`nproc` is 5; cgroup cpu.max is 400000 100000. Commands source
`/workspace/adamic-tools/env.sh`. Every test/build writes stdout and stderr to
logs, without piping a test process.

The final gate command, from the repository root:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE09_ARTIFACTS=/workspace/wave-09-final \
ADAMIC_WAVE09_REPOSITORY_MANIFEST=/workspace/wave-09-validation/repository.manifest \
ADAMIC_WAVE09_COMPILER_MANIFEST=/workspace/wave-09-validation/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8 \
go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v \
  > /workspace/wave-09-final-test.log 2>&1
# PASS 112.982s

go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-09-final-checker-test.log 2>&1
# PASS 0.300s
go test ./bridge/tsgo -count=1 -timeout=15m -v > /workspace/wave-09-bridge-test.log 2>&1
# PASS 131.146s
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /workspace/wave-09-node-oracle.log 2>&1
# PASS 19.482s, eight fixtures and one-byte mutant
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave-09-final-vet.log 2>&1
# Exit 0, empty log
gofmt -l cmd internal > /workspace/wave-09-gofmt.log
# Empty log
```

The new Go files are gofmt formatted. The existing facts.go dispatch deliberately
uses one physical registration line, respecting this wave's shared-file limit.
The pinned production CLI reports that named .a lint inputs are unsupported;
that failed check is recorded, not interpreted as a clean lint result. Its own
formatter API was run on the five new .a sources with resolved project options,
using a virtual TypeScript parser filename and no .ts source files. A second
pass reports all five unchanged. Final native agreement uses those formatted
sources. The formatter overlay tool is preserved beside the logs.

The full repository gate and an additional old 26-rule corpus sweep were not
run. The touched packages, bridge foundation and filtered external Node oracle
were run instead. The suite exposes production defaults only and has no lint
configuration, suppression, fix application or complete cohere CLI surface.
Every nondefault option, full upstream fixture matrix, JSX/JavaScript population
and ambient variant beyond these controls needs additional oracle coverage.

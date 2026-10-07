# Wave 12 continuation

Built: three more native rules; wave 12 now owns six completed ports.
Commits: continuation claim `047ec355` pushed before code; implementation `75fb78e8`.
Checks: 142 controls, 77 compiler roots, and 287 repository roots agree byte for byte, normal and sanitized.
Mutants: all three rule mutants, three raw-question mutants, and the registry-retention mutant were caught.
Uncovered: full repository gate, emitted-JavaScript comparison, and the complete upstream multi-file fixture matrix.

## Selection and territory

The original three rules were already tested and pushed in `6f852293`, with evidence in `1a1be27f`, before this continuation. A fresh fetch of all origin heads inspected 324 origin refs, claims across all branches, and native source names on main and `codex/tsgo-c-library`. The combined VOLUME_REPORT count tables contained 197 checker-dependent rules. The audit found 98 names claimed and 25 matching checker-dependent port names; method-signature-style is the additional non-checker inventory port. The first three available rules, with lexical ties, were:

- `nexus/correctness-no-process-exit-after-output`: 0 compiler, 0 repository.
- `nexus/correctness-no-uncleared-race-timeout`: 0 compiler, 0 repository.
- `nexus/correctness-require-blocking-standard-streams`: 0 compiler, 0 repository.

Their claim update was committed and pushed before implementation. No further rules were claimed. Implementation, runner, independent oracle, tests, and evidence are inside this worker's `wave12_next` directory. The only additions elsewhere are the three separately named Go question files and their direct-checker test. No existing shared harness, registration generator, bridge dispatcher, compiler source, or submodule pin changed in this continuation.

All nine new Adamic modules use `.a`. The two runtime-generated Nexus `.ts` files are TypeScript reference input fixtures: the production rule deliberately identifies the literal `/source/system/StandardStreams.ts` path. They are parsed as lint input, not compiled Adamic implementation modules.

## Decisions and raw facts

Each rule has its own file. Adamic decides declaration ownership, Promise and timer identity, handle reads, write provenance, path completion, catch provenance rollback, loop recurrence, one-level callee writes, entry classification, import closure, blocking order, callback/function reachability, report spans, counts, and messages.

The new raw questions are `symbol-ancestry`, `call-declaration`, and `program-imports`, each in its own Go and Adamic file. Each Go file registers with one initializer line using the registry already on this branch. They expose declaration/file/ancestor metadata, resolved signature declarations and flags, and import syntax/resolved paths. No Go cohere predicate or diagnostic is called by the bridge.

The owned source loader initializes top-level await parsing from the existing raw `scope-metadata` external-module fact. That fixes caller context without editing the shared parser. The independent oracle has its own loader and AST walk and calls all three pinned production `Run` implementations unchanged. It imports no bridge code and serializes findings, fixes, and suggestions completely. These three rules have no fixes or suggestions; their zero fields are still compared.

## Inputs and byte agreement

Pins remain:

- Cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- TypeScript-go: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
- TypeScript compiler: `050880ce59e30b356b686bd3144efe24f875ebc8`.

The frozen validation-coverage manifests are unchanged: 77 compiler roots, and 287 repository roots (212 `.a`, 75 `.ts`). Their portable manifests and source SHA-256 values are in evidence.

Controls are 39 targeted sources plus 103 literal reference source rows: 48 process-exit, 18 timeout, and 37 blocking-streams rows extracted with Go's AST parser from the pinned reference tests. Each source is re-evaluated against production Go, with shared ambient Node-like declarations and Nexus reference files. Controls include positive findings, shadowing and imports, dropped/read/assigned timer handles, generic Promise constructors, callbacks, helpers and never returns, loops and jumps, catch/finally provenance, literal loop tests, short-circuit branches and assignments, default parameters, class initializer refusal, Unicode, CRLF, shebang/count messages, and blocking before/after an exit.

| Population | Findings | Identical bytes | Normal + ASan/UBSan/LSan |
| --- | ---: | ---: | --- |
| 142 controls | 138 | 86,475 | PASS |
| 77 compiler roots | 0 | 5,010 | PASS |
| 287 repository roots | 0 | 18,485 | PASS |

The zero corpus counts agree with selection. Positive controls establish that none of the three implementations is a no-op. Compressed canonical Go/native/sanitized/mutant streams, hashes and stderr are in [evidence](evidence/diagnostic-hashes.json). Native sanitizer runs had empty stderr.

## Commands and observed outputs

The toolchain setup from the first batch remains in use: `bash cloud/setup.sh` passed, Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s, build cache warm 115s, total 115s. Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` again printed 5. Commands source `/workspace/adamic-tools/env.sh`.

```sh
ADAMIC_WAVE12_NEXT_ARTIFACTS=/workspace/wave-12/next/final3 \
ADAMIC_WAVE12_NEXT_REPOSITORY_MANIFEST=/workspace/wave-12/repository.manifest \
ADAMIC_WAVE12_NEXT_COMPILER_MANIFEST=/workspace/wave-12/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware/wave12_next
```

PASS, **135.611s**. Builds stage0, normal/sanitized C checker archives and native binaries, the independent oracle, and every mutant. Each subprocess writes stdout and stderr directly to files; test output is [agreement.log](evidence/agreement.log), never piped.

`go test -v -count=1 ./bridge/tsgo/checker`: PASS, **0.197s**, including direct alias/compound-name ancestry, resolved signature flags, import-resolution facts, and invalid-request refusals. `go vet ./bridge/tsgo/... ./stage1/cohere/typeaware/wave12_next`, gofmt over touched files, and `git diff --check`: clean. The first batch's filtered external Node oracle remains recorded in its report; it was not repeated in this continuation.

Failures were used to strengthen the implementation:

- Native build initially refused `??=`, spread arguments, Array.from(Set), untyped empty array fallbacks, a default sort comparator, and Number conversion. Owned code uses supported explicit initialization, loops, typed arrays, comparison and parser token values; no shared compiler changes were made.
- Top-level `await import(...)` initially lacked await context in the owned loader. The external-module fact now initializes that context.
- The byte oracle caught an incorrect pinned `never` bit and an extra blocking-streams finding for a class initializer the production CFG does not record. Both are corrected.
- The first full compiler run panicked because raw ancestry serialization called Node.Text on a destructuring binding pattern. Compound names now use their raw source span, and the direct checker regression verifies this case. The failed run is preserved as `compound-name-failure.log.gz`. The complete corrected run passed above.
- Early test-only overlays had a quoting/path error; those build failures were not counted as caught mutants. The final run records all mutants compiling and exiting normally.

## Mutants

| Change | Normal execution | Catch |
| --- | --- | --- |
| Process rule uses direct writes instead of following one callee | Exit 0, empty stderr | Go bytes differ at 3,924 |
| Timeout rule reverses handle-is-lost decision | Exit 0, empty stderr | Go bytes differ at 4,583 |
| Blocking rule reverses shebang reason selection | Exit 0, empty stderr | Go bytes differ at 716 |
| Raw ancestry reports declarations as non-declaration files | Exit 0, empty stderr | Go bytes differ at 66 |
| Raw signature reports no declaration body | Exit 0, empty stderr | Go bytes differ at 3,924 |
| Raw module resolution drops target paths | Exit 0, empty stderr | Go bytes differ at 13,231 |
| Released-program registry retains the handle | Exit 0 rather than required panic 70 | Released-handle assertion |

Every rule and raw-question mutant is caught only by full diagnostic bytes, not by a crash or sanitizer error. All three new questions reject released handles with panic 70 and `invalid or released checker handle`. Keeping the program in the registry makes the valid program-imports probe succeed, proving that assertion can fail.

## Quiet timing

Three interleaved native/Go whole-process runs per corpus, with full diagnostic bytes compared in every round. The existing benchmark driver was reused without editing it. Raw rounds and phases are [measurements.json](evidence/measurements.json).

| Corpus | Native process median | Go process median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.574s | 0.390s | 6.60x |
| Repository | 0.365s | 0.095s | 3.84x |

Compiler phase medians: native load 0.358s, run 2.088s; Go load 0.311s, run 0.061s. Repository: native load 0.081s, run 0.259s; Go load 0.075s, run 0.008s. Separate medians need not sum. Native made 108 compiler queries and 348 repository queries. These are observations on this worker, with no speed-parity claim; native is slower.

## Limits

The complete repository gate, the prior 26-rule suite, and emitted-JavaScript comparison were not run in this continuation. Shared harness work is left to its owner; no shared generator or harness file was edited. The pinned cohere CLI still cannot lint `.a`, as documented in the first batch.

Reference source rows are replayed, not the complete upstream multi-file fixture harness: helper maps, configuration/setup variants, and every original fixture expectation are not reproduced. The complete default behavior on the two frozen corpora and the positive controls above is established; arbitrary projects, every CFG corner, JSX, and JavaScript inputs are not established by these runs. Neither lint suppression nor edit application is implemented by this isolated comparison runner.

Built: the three claimed continuation rules, isolated checker questions, dedicated runner and independent byte oracle.
Commits: claim 463014f1; draft 29e0aa1b; completed implementation is the commit containing this report.
Commands and outputs: agreement PASS 285.713s; bridge PASS 168.060s; checker PASS 0.296s; filtered Node PASS 63.365s; vet PASS.
Mutants: all three native rule mutants caught only by full byte comparison; both question guards, registry retention, bridge ownership and Node byte mutants caught.
Not covered: full repository gate, every option combination, JSX controls rejected by Go, CLI lint for .a, and emitted-JavaScript execution of these FFI rules.

## Completed continuation

This report supersedes the blocked status in WAVE_22_NEXT_REPORT.md. Original wave 22 was already complete in 63298efb. The premature continuation claim 463014f1 reserved @typescript-eslint/no-misused-promises, @typescript-eslint/no-misused-spread and nexus/concurrency-no-lost-update before code; these three are now implemented and validated. No further claims were made before finishing them.

Promise and spread share a native PromiseTypes classifier and TypeOperations adapter. Go supplies generic type operations, signature/parameter/return identities, property lookup and declaration origins. Adamic makes every lint judgment, repair, suggestion and exemption. Lost-update builds an acyclic numeric-link control-flow graph, then tracks shared targets and cached dependencies across suspension, branching, loops and abrupt completion. Its BindingState adapter supplies raw binding, shorthand, alias and declaration facts.

The shared Reassign import caused an existing constructor refusal. The rule now contains its own write traversal, leaving that shared helper and compiler untouched. The direct metadata test difference was solely ObjectFlagsMembersResolved, a cache bit set by serialization; the independent comparison excludes only that cache bit. See wave_22_next_questions.md for the complete dedicated protocols. No shared registration generator or existing test harness was modified, and no protected emitter/lowerer files were edited. New Adamic files use .a.

## Byte agreement

Production Go rules at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db are the independent authority. The oracle imports no native bridge. TypeScript compiler source is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8; typescript-go is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Defaults and frozen root manifests match the base runner.

| Population | Roots | Findings | Identical bytes, Go/native/sanitizers |
| --- | ---: | ---: | ---: |
| Independently parsed controls | 546 | 346 | 101,550 |
| Frozen repository | 287 | 0 | 18,485 |
| TypeScript src/compiler | 77 | 0 | 5,857 |

The controls come from production promise/spread test strings, lost-update helpers and corpus fixtures, plus targeted Unicode/CRLF, inherited methods, Promise unions and variadic void callbacks. Each rule has positive controls. Sixteen extracted sources are rejected by the independent Go parser under .a, including JSX; rejected-controls.json records all sixteen. Comparison checks complete finding spans, IDs/messages, fix triples and suggestion IDs/descriptions/edits, preserving ordering and duplicates. Sanitizer stderr is empty for all comparisons.

Debugging comparisons caught binary token selection, function return-annotation selection and two ternary branch child indexes. Those are actual corrected defects, separately from deliberate mutants.

## Mutants and handle checks

All three rule mutants build, exit 0 and have empty stderr. Only the production byte oracle catches them:

| Mutant | First differing byte |
| --- | ---: |
| Promise conditional message ID changed | 14,024 |
| Spread await suggestion insertion shifted one byte | 18,782 |
| Lost-update write diagnostic end shifted one byte | 56,328 |

The binding-state suffix guard mutant initially survived because the invalid-suffix test also supplied an invalid node. Adding a valid Identifier with a suffix made the baseline pass and mutant fail with accepted binding-state suffix on valid Identifier. The type identity guard mutant accepts raw identity 00 and fails TestTypeOperations. Both failures are runtime test assertions, not build errors.

Both new question modes after release panic with exit 70 and exact invalid-or-released-checker-handle text. A retained-registry mutant queried through binding-state exits 0 and fails the required refusal. The full bridge gate independently validates 100 ownership queries and 1,600 compiler positions, 54,982 identical bytes under ASan/UBSan/LSan, plus input/output-length, stale-registry, wrong-position, missing-link and leak mutants. The filtered Node gate checks eight fixtures across native and JavaScript and proves its one-byte oracle mutant. It does not execute these three FFI rules through JavaScript.

## Native time against Go

Three quiet alternating full-output runs, with byte comparison each round, exclude builds and sanitizer runs. Native is slower on both populations.

| Population | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 1.134065s | 0.354046s | 3.203 |
| Compiler | 13.328035s | 2.448317s | 5.444 |

Exact runs and native query counters are in validation-wave-22-next/timings.json. These are observations on this worker, not speed guarantees.

## Commands and evidence

Source /workspace/adamic-tools/env.sh first. All test output went directly to logs, never through a pipe.

```sh
ADAMIC_WAVE22_NEXT_ARTIFACTS=/workspace/wave-22-next-final ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22NextAgreement$' -count=1 -timeout=30m -v > /workspace/wave-22-next-final.log 2>&1
TMPDIR=/workspace/wave-22-scratch ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-22-next-bridge.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-22-next-checker-final.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /workspace/wave-22-next-node.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... > /workspace/wave-22-next-vet-final.log 2>&1
python3 stage1/cohere/typeaware/validation-wave-22-next/benchmark.py /workspace/wave-22-next-final /workspace/adamic /workspace/wave-22-typescript-pinned /workspace/wave-22-next-benchmark > /workspace/wave-22-next-benchmark.log 2>&1
```

Guard mutants use scratch Go overlays; their baseline, survivor and final caught-mutant logs are committed. Reproducible native rule and released-registry mutants are part of TestWave22NextAgreement. Exact gzip diagnostic streams, hashes, accepted/rejected sources, manifests, logs and timing counters are committed in validation-wave-22-next. Absolute stream headers reflect the artifact directory above.

Initial toolchain setup remains the original unit's run: go/clang/node/submodules ready 0s each, build cache 118s, total 118s, nproc 5 with four-core quota. Setup was not repeated. The existing CLI .a lint limitation remains reported in WAVE_22_REPORT.md; no CLI or whole-repository gate success is claimed.

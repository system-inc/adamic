Built: rebased the nine completed wave-07 ports onto current main; no new rules claimed.
Commits: rebased source 305128993767e0b42edec69b5739871d392fe4aa; landing runner 1e6803b1b9044fea19daaf90c43e928d129f1699, resumed runner 5c468d29ab9cbbd087069b26ba4cdf48b68cb4c5; main f8013f0baac41ddc340d76f83bddde38536a8f07.
Commands and outputs: all nine Go rule gates, sanitizers, released handles, bridge/decoder checks, 30 supported Node fixtures, 12 iterator refusals, vet, formatting and 39-file source lint pass; timings below.
Mutants: nine per-rule mutants, nine numeric declarations and all existing graph, fact, handle and decoder guards were caught again; individual mechanisms and observations below.
Not covered: three React ports, four production bridge routes, numeric shared-driver/body migration, full repository gate, exhaustive options and emitted-JavaScript rule comparisons.

## Landing scope

The one owned pushed branch is `codex/typeaware-wave-07`. Its previous remote tip was 0dca2669f9ee126a9532e14d68117e511cd01e8d. All 19 commits were rebased without conflicts onto current main f8013f0b. Main added narrowed-expression checks, override checks, checking-pragmas refusals and collection iterator/runtime fixes. Native rule and bridge question implementations are unchanged; the worker-owned landing runner now includes 30 supported Node fixtures and the 12 iterator-copy refusal controls. No main or area branch is a push destination. The final publication uses an exact lease against the previous owned tip.

Numeric listener declarations remain verified against production Go's actual registrations. This does not establish shared driver integration: current ParseNode still exposes string kinds, and existing rule bodies retain their previous traversal. No new rules were implemented or declared in `rule.json`; no shared API was invented. See [SPEED_REPORT.md](SPEED_REPORT.md) for the listener map and exact API gap.

The three React claims still depend on JSX parser integration. The published dependency at a8a62d62 parses the controls in isolation; current main's parser still rejects them. The React rules require native value-flow implementations after that parser dependency lands. Four bridge dispatch routes also remain private overlays. Shared registration generator, harness, parser and protected compiler files are not edited in this refresh.

## Reproduction

Source `/workspace/adamic-tools/env.sh`, then run:

```sh
ADAMIC_WAVE07_LANDING=/workspace/wave-07-latest-landing python3 stage1/cohere/typeaware/wave07_react/landing_gate.py > /workspace/wave-07-artifacts/latest-landing-gates.log 2>&1
python3 stage1/cohere/typeaware/wave07_react/validate_listeners.py > /workspace/wave-07-artifacts/latest-listeners.log 2>&1
ADAMIC_WAVE07_LANDING=/workspace/wave-07-latest-landing python3 stage1/cohere/typeaware/wave07_react/landing_bench.py > /workspace/wave-07-latest-landing/bench.log 2>&1
```

Setup passed: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 90s, total 90s. nproc 5, CPU quota four cores, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Production cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript-Go is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a; TypeScript compiler corpus is 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Frozen corpus manifests cover 287 repository and 77 compiler sources.

## Mutants and expected check

Each per-rule mutant must compile, exit 0 with empty stderr and differ from the independent production Go output. The changed behaviors are: no-undef-init removal starts one byte later; prefer-for-of inverts its body predicate; consistent-indexed-object-style changes message punctuation; each of the three correctness rules, rest parameters, promise rejection and regex literals changes its message ID.

The native type-reference graph mutant inverts the anchor comparison. Five Go fact mutants zero graph symbol identity, zero continuation symbol identity, drop control-flow edges, erase module targets and drop regex character spans. These also require normal exits and independent byte mismatches. Each numeric-listener mutant advances one rule's first declared kind by one; the production listener-map probe must catch all nine.

The original, continuation and regex registry-retention mutants are caught by required released-handle panic-70 checks. Bridge tests additionally catch wrong input/output lengths with ASan, omitted output frees and an unowned region allocation with LSan, retained stale handles with a stale-handle assertion, wrong source position with independent oracle bytes and removed link opt-in with refusal. The decoder catches 19 wire mutants with its expected panic: empty, length-negative, length-short, trailing, version, question, not-integer, rounded-integer, noncanonical-integer, negative-natural, bad-boolean, zero-id, negative-tuple, missing-root, missing-constraint, missing-element, duplicate-id, missing-link-17 and missing-link-18. Wrong-kind and unknown-question guard mutants violate their required refusals. The external Node one-byte mutant must fail its comparison. All individual observations, including offsets and sanitizer diagnostics, are retained in this run's evidence.

## Observed results

All 16 landing steps pass. The first run completed the nine rule gates, both question validators and bridge package tests. The decoder step then failed because the 32 GB filesystem had no space for a runtime-cache directory; it did not report a decoder mismatch. The retained failure log states `no space left on device`. Removing 42 explicitly named obsolete ELF binaries and checker archives from this worker's older scratch directories freed 1,734,252,908 bytes. No sources, logs or committed evidence were removed, and no shared cache was cleared.

The owned runner gained `ADAMIC_WAVE07_RESUME`: before skipping a prior step it verifies that step has a recorded successful exit. The resumed run starts at `fact-guards` and completes decoder guards, Node oracle, vet, formatting and both React/dependency probes. The command ledger preserves the failed decoder attempt and its successful retry. Native rules, Go questions and compiler source are identical between the initial and resumed tested revisions; only this owned runner changed. There was no reason to repeat already passing rule gates after disk recovery.

```sh
ADAMIC_WAVE07_LANDING=/workspace/wave-07-latest-landing ADAMIC_WAVE07_RESUME=fact-guards python3 stage1/cohere/typeaware/wave07_react/landing_gate.py > /workspace/wave-07-artifacts/latest-resumed-gates.log 2>&1
```

The original trio's 283 valid controls agree on 47,635 bytes and 173 findings: no-undef-init 26 findings / 11 fix edits; prefer-for-of 78 findings; indexed-object-style 69 findings / 58 fix edits / seven suggestions. Its frozen repository corpus agrees on 22,753 bytes / 16 findings, compiler corpus on 9,282 bytes / 24 findings. Normal and ASan/UBSan/LSan observations match.

| Continuation controls | Findings | Bytes | Suggestions |
| --- | ---: | ---: | ---: |
| Timeout | 11 | 6942 | 0 |
| Process exit | 16 | 9638 | 0 |
| Blocking streams | 16 | 11990 | 0 |
| Rest parameters | 9 | 3431 | 0 |
| Promise rejection | 22 | 8175 | 0 |
| Regex literals | 41 | 24341 | 30 |

All six continuation rules match their control, DOM and sanitizer comparisons. Each emits zero findings on the frozen corpora with 18,485 repository bytes and 5,318 compiler bytes. Findings, fixes and suggestions are compared as complete serialized streams, including spans, text, order and duplicates. Numeric listener declarations match 358 production-Go bytes under ordinary and sanitized execution; all nine numeric-key mutants compile, exit 0 and differ only in oracle bytes. All rule and fact mutants described above are caught again; live questions succeed and released questions panic 70.

Bridge package tests and decoder/refusal/type-flag tests pass. The filtered external oracle passes 30 supported native/JavaScript/source-Node fixtures, the one-byte mutant and 12 iterator-copy refusal controls. `go vet ./...` and formatting checks pass. Source lint checks all 39 owned `.a` files with 276 rules and zero findings. The full `go test ./...` gate was not run, and Go-managed heap internals are outside C sanitizer instrumentation.

## Fresh native versus Go time

Three alternating rounds per corpus compare complete output bytes after all builds and tests finish. All 84 timed executions exit 0 and match their paired output. Whole-process medians in seconds:

| Unit | Repository native / Go | Compiler native / Go |
| --- | --- | --- |
| original-trio | 0.327103 / 0.162773 | 2.193105 / 0.328886 |
| timer | 0.265519 / 0.129650 | 1.759839 / 0.314650 |
| process | 0.276196 / 0.145297 | 1.958776 / 0.322890 |
| streams | 0.368249 / 0.153337 | 2.004535 / 0.353150 |
| rest | 0.322226 / 0.160384 | 2.119131 / 0.378037 |
| promise | 0.272530 / 0.129353 | 1.816946 / 0.292556 |
| regex | 0.264108 / 0.127590 | 1.995707 / 0.308855 |

Native remains slower in every measured group. Original-trio timing covers three rules together; continuation rows cover individual rules. These observations establish no speed gain from unused listener declarations. Exact command records, raw streams, fixtures, hashes, disk-recovery records, source pins and medians are in [current-landing-evidence](current-landing-evidence). Earlier landing and speed evidence are preserved separately.

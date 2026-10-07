Built: approved lowering inventory, per-emitter cached immortal-field proof, and old-slot release elision, with graph dispatch preserved after the immortal test.
Commits: integration 4616e91b; main merge f856da5c (origin/main 71d7e491); graph compatibility f094f5f0, e0c85c96 and 1825500b; expectation refreshes 236b5f6e and d56e3af8; final tip recorded in the landing response.
Commands and outputs: refreshed three-variant profiles and full batch 8 output comparison; final gate and timing results below.
Mutants: ignored heap write and unconditional field proof each leak 425 bytes/five allocations; treating the heap-string parameter as immortal causes ASan use-after-free.
Not covered: dominant kind-field release elision, hardware counters, or queued bitwise folding.

## Approved hooks

* internal/native/emit.go:200: one per-emitter immortalFields map; internal/native/branch_shape.go:73 populates it once from the inventory, requiring a complete, nonempty list and proving every expression. Iteration order never affects emitted output.
* internal/lower/lower.go:75: one call fills Program.FieldWrites after lowering; internal/ir/ir.go:21 holds it; new internal/ir/field_writes.go:5 defines the reusable Expressions/Complete record. internal/lower/field_writes.go:18 collects writes across all functions and conservative shared field names. Unknown/runtime writers retain releases.
* internal/native/emit_statements.go:169: the tested integration skips the old-slot read/release only for proven immortal reference fields, preserving evaluation and store checks.

The small named immortal test remains internal/native/emit_ownership.go:75, called by retained at line 66. heldReferenceIn and dropIn test it first in internal/native/graph_regions.go:27 and :48; subsequent graph-header dispatch is unchanged. No runtime C implementation was modified by this unit.

## Landing compatibility

Merged current main 71d7e491b3c9724f7a0e2ee754592149e7f9790b. The first full-gate attempt exposed main's new CallTargetReaders guard against the older graph fallback's direct target reads. Imported the graph worker's existing f409709 fix as f094f5f0, plus its shared leak-check package from 4305edef as e0c85c96. Added two narrow oracle test adapters required by its graph review probes, preserving main's existing oracle leak path. These are repository-owned changes, with no cohere code copied. Graph compatibility emits byte-identical parse C to the final profiled candidate.

The updated count inventory differs from the previous static-only branch by two fewer releases for library_for_in_live, with other columns unchanged. Subsequent refresh including graph classification_override made no changes. The guard passes (0.545s) and complete count refresh passes (35.162s).

The full repository run also exposed stale parser/stage 3 gap catalogs inherited from graph/closure integration. Both reproduce with the matched before overlay, disabling both immortal optimizations. Commit 236b5f6e changes the parent-cycle probe to require Node agreement in release and ASan/UBSan builds, and updates its gap documentation. It refreshes seven stage 3 entries to Compiles only after the existing native/sanitizer/leak oracle hook agrees with recorded and current Node. Two remaining unsupported fixtures receive the exact current Refused/NotYet diagnostics; Node answers and provenance fields are preserved. Full stage 3 passes (13.601s), and all six parser gap probes pass (0.897s). No fixture source or compiler behavior was changed by this compatibility refresh.

## Fresh per-fix measurements

Main changed imported stage 1 source since the earlier staged report. Therefore none of that report's times or profiles are reused here. Before, static-only and cached-field candidates were regenerated from the same new source tree and runtime. Baseline overlays undo only the immortal ownership changes and the field-store integration. All share numeric-switch lowering and graph runtime. Corpus, compiler flags and cache model are the same as the earlier report.

Fix 1: static retain elision.

| Variant | Simulated mispredictions | L1 instruction misses | All L1 misses | Instructions | Best of ten user s |
| --- | ---: | ---: | ---: | ---: | ---: |
| Before | 54,475,354 | 80,505,510 | 101,966,869 | 6,845,562,628 | 0.864725 |
| Static addresses | 56,056,839 | 79,302,712 | 100,764,291 | 6,770,724,604 | 0.841627 |

Fix 2: cached complete-field proof.

| Variant | Simulated mispredictions | L1 instruction misses | All L1 misses | Instructions | Best of ten user s |
| --- | ---: | ---: | ---: | ---: | ---: |
| Static addresses | 56,056,839 | 79,302,712 | 100,764,291 | 6,770,724,604 | 0.841627 |
| Plus field proof | 57,459,621 | 78,020,238 | 99,481,817 | 6,770,724,604 | 0.851550 |

Static elision cuts instructions 1.0932% and I1 misses 1.4941%, while increasing simulated mispredictions 2.9031%. Field proof leaves executed instructions unchanged, cuts I1 misses 1.6172%, and increases mispredictions 2.5024% relative to static-only. Combined: instructions -1.0932%, I1 -3.0871%, mispredictions +5.4782%. Best user time changes: static -2.6711%, fields versus static +1.1790%, combined -1.5236%. The field-only time regression is retained in the table; no field-proof speedup is claimed. Field effects are code-layout effects, not evidence of hot releases removed. The analysis still rejects kind; recordKind is literal-only but outside hot parse. The original eligibility lower bound of 8,867,204 immortal outcomes (54.1%) clears the measurement gate but does not prove those values statically.

Bcm/Bim: before 49,954,595/4,520,759; static 51,540,807/4,516,032; fields 52,954,637/4,504,984. C bytes: 2,393,110 / 2,384,305 / 2,383,446. C size is not an instruction-cache footprint. No hardware attribution or repeated-series stability claim is made.

## Commands and environment

bash cloud/setup.sh: ready 0s, cache warm/done 124s; nproc=5, CPU quota four cores. Each shell sources /workspace/adamic-tools/env.sh. The first post-merge full run lacked the new Markdown width dependencies because its environment script predated main. Re-running setup installs the locked oracle dependencies and adds ADAMIC_MARKDOWNWIDTH_DEPS: Markdown dependencies ready 0.895s, Go build ready 37.934s, cache warm 38.151s, done 38.183s; nproc remains 5. No module-fetch 403 occurred. Go 1.27.1, clang 20.1.8, Node 24.19.0, AMD EPYC 9V74. TypeScript corpus 050880ce59e30b356b686bd3144efe24f875ebc8, 77-file manifest; cohere submodule 715ba94f3608a6500086b1076ce5cb7e51b836db. Parse-only batch 8 harness and Go-driver reconstruction are preserved with the earlier evidence.

For VARIANT = landing-before, landing-static, landing-fields:

```sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/immortal-sites/VARIANT.c internal/native/runtime/*.c -lm -o /workspace/scratch/immortal-sites/VARIANT
VALGRIND_LIB=/workspace/scratch/branch-shape/valgrind/usr/libexec/valgrind /workspace/scratch/branch-shape/valgrind/usr/bin/valgrind --tool=callgrind --cache-sim=yes --branch-sim=yes --dump-instr=yes --collect-jumps=yes --callgrind-out-file=/workspace/scratch/immortal-sites/VARIANT.callgrind /workspace/scratch/immortal-sites/VARIANT --manifest /workspace/scratch/branch-shape/compiler.txt --count
python3 /workspace/scratch/branch-shape/profile.py /workspace/scratch/immortal-sites/VARIANT.callgrind --output /workspace/scratch/immortal-sites/VARIANT-profile.json
python3 /workspace/scratch/immortal-sites/landing-timing.py /workspace/scratch/immortal-sites --manifest /workspace/scratch/branch-shape/compiler.txt --cpu 3
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 90m ./stage1/cohere/lint
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 90m ./stage1/cohere/markdownblocks
go test -count=1 -timeout 30m ./stage3/fixtures
go test -count=1 -timeout 10m ./stage1/typescript/parser -run '^Test(StrongAstParentSupported|PushSpreadGap|TypeImportCycleGap|ClassMethodInterfaceGap|OptionalFunctionValueGap|ConditionalEmptyArrayGap)$'
go test -count=1 -timeout 10m ./stage1/cohere/markdownblocks -run '^TestParserRepresentationProbes$'
go vet ./...
```

No LTO, PGO, architecture specialization, sanitizer, counting or slab flags in measurements; runtime C files compiled as separate translation units. Callgrind 3.24 simulates 32KiB 8-way I1/D1 with 64-byte lines and 256MiB direct-mapped LL. Self event totals reconcile to footer; inclusive edges are excluded. I1 is I1mr; total L1 is I1mr+D1mr+D1mw; mispredictions are Bcm+Bim. Profiles exit zero and print exactly 0 newline. Nonfatal Valgrind brk-limit warnings remain in logs.

Timing warms each variant once, rotates/reverses order for ten rounds, pins CPU 3 before exec, measures direct-child user time with wait4, and verifies output/exit for each run. It runs after all tests/builds/profilers finish. Raw samples, load observations and commands are retained; best wall time is never substituted for user time. Every test and tool execution writes a log; no test is piped.

## Correctness

The uncached full ./... run passes lower (42.543s), native (662.668s), oracle (373.192s), IR, flow, fresh, fuzz, regexp, exhaustive Unicode, stage 1 CSS/JSON/Markdown inline/type-aware and the other packages. Its four failing packages are preserved in landing-green-gate.log.gz: lint exhausted the 30-minute package budget during TestMutants; Markdown lacked the newly locked width dependencies and later exhausted the same budget during active layout comparisons; parser retained the old strong-parent refusal expectation; stage 3 retained nine obsolete closure diagnostics. These are not hidden or called a successful full invocation.

The two expectation packages are rechecked after the Node-verified refresh above. Lint and Markdown are rerun uncached with 90-minute timeouts after refreshing setup. Lint passes uncached in 1702.034s. Markdown completes in 2450.163s with only three stale TestParserRepresentationProbes expectations failing; these also reproduce with both immortal changes disabled. Commit d56e3af8 refreshes the remaining first-class nested-function diagnostic and replaces obsolete graph-cycle refusals with positive checks against Node, release C, sanitized C, JavaScript backend and the leak checker. The entire changed probe passes uncached in 0.907s. All other Markdown tests completed in the full retry. Thus every package test is covered by the full runs and affected-test reruns, without claiming a single successful ./... invocation. No compiler source changes occurred after the full run; the subsequent changes refresh tests and gap documentation only. Final go vet ./... and diff check pass; formatting has empty output. The final response distinguishes these combined suite checks from a single uninterrupted full-gate pass.

Full batch 8 release output matches the independent Go harness byte for byte: 11,444,034 bytes, 77 files, zero exits and empty stderr. Final parse C is byte-identical after graph compatibility and again at source tip d56e3af8 (landing-tip-parse-cmp.log.gz).

Fresh mutants on the integrated cached implementation:

| Mutation | Test and semantic result |
| --- | --- |
| Omit SetProperty writes from collectFieldWrites | TestImmortalFieldsMatchesNode fails LSan: 425 bytes/five heap-string allocations leaked |
| Accept every write expression in the immortal-field proof, dropping heap old values | Same pinned test and five-allocation LSan leak |
| Treat heap-string parameter adamic_local_N_text as immortal | Same pinned test fails ASan heap-use-after-free in heap.c:393 |

All three mutants compile and reach execution; none is killed by -Werror. The last mutant exercises missing heap retain; the second exercises missing real heap release. The good fixture overwrites heap strings and literals under Node, release and ASan/UBSan lanes. Prior direct unconditional-release omission mutant and its sanitizer failure are preserved in the historical implementation evidence.

Exact mutant command for NAME = heap-write, field-release, heap-string: go test -overlay /workspace/scratch/immortal-sites/landing-mutant-NAME-overlay.json -count=1 ./internal/native -run '^TestImmortalFieldsMatchesNode$'. Logs and exact overlay sources are preserved alongside the profile evidence.

Bitwise folding stays queued after this landing; no bitwise conversion code was changed.

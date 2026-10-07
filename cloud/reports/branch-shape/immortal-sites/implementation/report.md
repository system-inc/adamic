Built: immortal static-address retain elision and reusable field-write inventory; field-store elision is a tested, measured integration patch awaiting scope approval.
Commits: fdd5706 is the production change, on codex/branch-shape; base 0db045a includes graph-regions 410cb1c and current main b6b1538.
Commands and outputs: three reconciled cache/branch profiles, pinned interleaved best-of-ten timing, all staged semantic oracle fixtures, and byte-identical 11,444,034-byte batch 8 output against Go.
Mutants: heap-string retain omission causes ASan use-after-free; omitted heap write, unsound field proof, and real heap release omission each cause LSan's 425-byte/five-allocation leak.
Not covered: dominant kind-field release elision, hardware counters, staged-proof counting-table refresh, full repository gate, or production integration of the restricted lowering hook.

## Result and limits

The eligibility measurement remains valid: at least 8,867,204 immortal outcomes (54.1% of 16,386,567) trace to generated kind-field stores and getter results. That does not imply all those values are statically provable. This deliberately conservative inventory rejects kind. It is shared with runtime-produced objects, has 36 write expressions on parse, and scanner writes include map results and restored values. The inventory proves recordKind only, with six write expressions. It does not remove the dominant measured references-zero release branch.

Static retain elision reduces parse instructions by 1.10% and best user time by 0.67%, but raises simulated mispredictions by 6.00% and L1 instruction misses by 3.63%. The staged field proof lowers the latter two relative to static-only, but instructions differ by only +212 across 6.76 billion. Its proven field is outside hot parse. The miss/time changes are layout-sensitive and must not be claimed as evidence of hot release traffic removed. There is no hardware causal attribution or repeatability claim beyond this one ten-sample series.

## Table per fix

Every row uses the same corpus, merged runtime and release policy below. Mispredictions are Bcm+Bim; all L1 misses are I1mr+D1mr+D1mw. Report I1 separately, since it dominates the cache cost.

Fix 1: static-address retain elision, production commit fdd5706.

| Variant | Simulated mispredictions | L1 instruction misses | All L1 misses | Instructions | Best of ten user s |
| --- | ---: | ---: | ---: | ---: | ---: |
| Before | 56,549,287 | 77,277,738 | 98,599,720 | 6,835,600,973 | 0.830478 |
| Static addresses | 59,939,994 | 80,086,689 | 101,407,715 | 6,760,762,949 | 0.824922 |

Fix 2: complete-field proof, staged overlay only; restricted integration is in field-integration.patch.

| Variant | Simulated mispredictions | L1 instruction misses | All L1 misses | Instructions | Best of ten user s |
| --- | ---: | ---: | ---: | ---: | ---: |
| Static addresses | 59,939,994 | 80,086,689 | 101,407,715 | 6,760,762,949 | 0.824922 |
| Plus field proof | 56,300,738 | 77,372,688 | 98,729,775 | 6,760,763,161 | 0.816559 |

Raw Bcm/Bim respectively: before 52,007,579/4,541,708; static 55,272,579/4,667,415; fields 51,712,895/4,587,843. Data L1 misses respectively: 21,321,982; 21,321,026; 21,357,087. Generated C sizes: 2,234,555; 2,225,780; 2,224,922 bytes. Emitted C byte size is not an instruction-cache footprint. Benchmark C from the final expanded static whitelist is byte-identical to the profiled static C; the comparison is recorded.

## Implementation

internal/native/emit_ownership.go:65 retained calls staticallyImmortal before the existing NULL special case. The small named function is at line 75, its anchored whitelist at line 79. It recognizes generated string statics, the runtime's empty/true/false strings, boolean boxes, typeof strings and their emitted pointer casts. Arbitrary expressions, heap temporaries and unrecognized names retain their existing behavior.

internal/native/graph_regions.go:27 heldReferenceIn checks immortality first; line 48 dropIn does the same. The remaining holder/header-based graph dispatch is unchanged. An immortal static never belongs to a graph region. No runtime implementation was edited.

internal/lower/field_writes.go:21 collectFieldWrites inventories every ObjectLiteral field and SetProperty in main and every function, including nested loops, catch/finally and closures through the existing IR walker. FieldWrites contains Expressions and Complete. The exported CollectFieldWrites permits reuse by other compiler properties without mutating the program. Names share writes conservatively across classes, inheritance and structural aliases. Runtime fields and accessors are incomplete. Regex named fields, unknown spreads and dynamic Object writers make the inventory incomplete. No write is ignored because it occurs in a rarely executed or unreachable function. An empty inventory proves nothing.

The staged branch_shape.go helper requires completeness and proves every expression immortal: literal strings, NULL/undefined, boolean-to-string statics, static reference boxes and conditional arms. One parameter, allocation or unproven expression keeps the old release. SetProperty retains its existing value/object evaluation order and write checks, then omits reading and dropping the old slot only when this proof succeeds. No shape lookup, store layout, stack check or naming policy was changed.

Native cannot directly import lower: existing lower tests import native, creating an import cycle, caught by the first attempted test run. The concrete integration patch adds one collectFieldWrites call after lowering finishes, stores the immutable inventory on ir.Program, and consumes it in native. This avoids duplicate analysis, global caches and repeated whole-program scans. The original unit rule explicitly excludes internal/lower/lower.go unless the unit authorizes it; approval for that single hook and IR storage was requested, so those repository files remain untouched. The full candidate is built and tested through a Go overlay. The additional test-only parse-inventory logger is preserved in the overlay snapshot but excluded from the proposed patch.

## Pins, flags and measurement commands

Area/runtime fetched as 94a9c83; graph-regions 410cb1c is not its ancestor. Used the explicitly supplied fallback 410cb1c7b989b679fa73d70c4dd605265f212426, merged into our existing branch without conflict. A final fetch of main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 confirms it is already an ancestor. Baseline program/runtime base: 0db045ae0b05542a1f7783f83d0b1e11f4ace8be. Baseline Go overlay restores only emit_ownership.go and graph_regions.go from that base. Static and field variants share the same numeric-switch lowering and runtime. No historical 6.46G or 9.26G counts are substituted for this new base.

bash cloud/setup.sh: Go/clang/Node/submodules ready at 0s; build cache warm 124s; done 124s. nproc=5, cpu.max=400000 100000, memory 17.6 GB. Each shell sources /workspace/adamic-tools/env.sh. Go 1.27.1; clang 20.1.8 (87f0227cb60147a26a1eeb4fb06e3b505e9c7261); Node 24.19.0; AMD EPYC 9V74; Linux 6.18.44. Cohere submodule 715ba94f3608a6500086b1076ce5cb7e51b836db. No cohere code was copied or edited.

Use the previously accepted batch 8 parse-only reconstruction in /workspace/scratch/branch-shape/batch8/parse.a, rule traversal removed, on 77 compiler files at TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8. Corpus and Go-driver reconstruction are preserved in cloud/reports/branch-shape/evidence. Manifest: /workspace/scratch/branch-shape/compiler.txt. Full batch8 main.a is used only for the independent Go output comparison.

Exact timed/profiled native build, substituting VARIANT with before, static or fields:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/immortal-sites/VARIANT.c internal/native/runtime/*.c -lm -o /workspace/scratch/immortal-sites/VARIANT
VALGRIND_LIB=/workspace/scratch/branch-shape/valgrind/usr/libexec/valgrind /workspace/scratch/branch-shape/valgrind/usr/bin/valgrind --tool=callgrind --cache-sim=yes --branch-sim=yes --dump-instr=yes --collect-jumps=yes --callgrind-out-file=/workspace/scratch/immortal-sites/VARIANT.callgrind /workspace/scratch/immortal-sites/VARIANT --manifest /workspace/scratch/branch-shape/compiler.txt --count
python3 /workspace/scratch/branch-shape/profile.py /workspace/scratch/immortal-sites/VARIANT.callgrind --output /workspace/scratch/immortal-sites/VARIANT-profile.json
python3 /workspace/scratch/immortal-sites/timing.py /workspace/scratch/immortal-sites --manifest /workspace/scratch/branch-shape/compiler.txt --cpu 3
```

All actual compiler/test/profile/timing output was redirected to separate logs or captures. No test execution was piped. No LTO, PGO, sanitizers, ADAMIC_COUNT, slab flags, or architecture specialization in measurements. Runtime C files are separate translation units. Callgrind models 32KiB 8-way I1/D1, 64-byte lines; 256MiB direct-mapped LL. Each profile exits zero with stdout exactly 0 newline. Nonfatal Valgrind brk-limit warnings are preserved. Self costs reconcile to footer; only the previously validated zero/two-instruction startup summary difference is tolerated. Inclusive call-edge costs are excluded.

Timing warms each variant once, then rotates and reverses the three variants for ten rounds, CPU 3 pinned before exec. wait4 supplies direct child user time. All 30 measured captures pass output/exit checks. No profiler, test or build ran concurrently with timing; environment services and host scheduling remain. All samples, commands and load observations are in timing.json. Best wall time is selected separately and is not substituted for user time.

## Correctness and mutants

Pinned production checks: TestCollectFieldWrites, TestImmortalStaticDispatch, TestImmortalFieldsMatchesNode, TestFieldWriteInventoryIncludesRuntimeLayouts and the independent TestRuntimeFieldLayoutsAreIncluded pass. Positive proof and negative heap-writing field are tested in the staged TestImmortalFieldProof. The semantic fixture compares release and ASan/UBSan builds to source on Node, repeatedly overwrites a real heap string and then a literal, and requires clean execution. Graph dispatch assertions check static constants bypass it while heap values still use it.

Initial uncached package gate: lower PASS 20.861s, native PASS 382.643s. Oracle's only failing tests are TestCountsAreRecorded and TestGraphRegionsCountsAndFree, both expected retain-count rows. Counts were refreshed with go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts: PASS 31.146s. All 245 changed rows differ only in the retain column, with allocations/frees/releases/peaks/regions unchanged. Graph counting test then PASS 3.373s.

Staged full semantic gate:

```sh
ADAMIC_GATE_UNCACHED=1 go test -overlay /workspace/scratch/immortal-sites/integration-overlay.json -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/oracle -skip '^Test(CountsAreRecorded|GraphRegionsCountsAndFree)$'
```

PASS: lower 18.817s, native 322.177s, oracle 173.105s. Every fixture's semantic and sanitized lane is covered; only the two named table-count tests are excluded from the staged candidate. Both static-only and staged full batch8 release binaries compare byte for byte with Go, all 77 files, 11,444,034 bytes, empty stderr, zero exits. Final ordinary counts PASS 23.673s; lower pinned PASS 0.007s; native pinned PASS 0.342s; go vet ./... PASS; gofmt and git diff --check have empty output. Full ./... gate is not claimed.

| Mutant | Lane and observation |
| --- | --- |
| Treat the fixture's heap string parameter as immortal | TestImmortalFieldsMatchesNode reaches ASan heap-use-after-free in adamic_release |
| Ignore SetProperty write expressions in the shared inventory | Staged semantic lane reaches LSan: 425 bytes leaked in five allocations |
| Accept every write expression as immortal, dropping real heap old values | Staged semantic lane reaches the same five-allocation leak |
| Drop old-value releases directly on all reference stores | Production semantic lane reaches the same five-allocation leak |
| Treat every C value as static | Additional broad control reaches ASan use-after-free; not used as the specific heap-string proof |

Every ownership mutant compiles, reaches the semantic lane, exits nonzero, and is killed by execution/sanitizers rather than -Werror. Raw logs and exact overlay sources are included. None changes the measured good binaries.

## Pending integration and next item

field-integration.patch is ready for the requested scope decision. Its lowering hook and IR inventory storage need approval under the original territory restriction; no emit.go cache change is needed. With approval, apply the patch, refresh counting rows for the field candidate, recheck them, and push only codex/branch-shape. The largest kind-field release traffic remains unoptimized by this conservative proof and requires a stronger allocation/flow proof, not an assertion that all observed immortals are literals.

Queued after this item: measure named-enum/literal constant bitwise conversion on a checker-shaped SymbolFlags/TypeFlags/NodeFlags workload, using origin/codex/bitwise-inline 5b24522 and integer-fast-paths a183e50 as references. Build only if that measurement pays. Convert constants with exact JavaScript ToInt32/ToUint32 behavior, compare edge cases to Node, and kill the above-2^31 missing-wrap mutant. No bitwise code was changed here.

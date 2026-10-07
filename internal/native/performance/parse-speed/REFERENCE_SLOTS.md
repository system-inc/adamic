Built: measured non-reference destruction-loop visits on all 77 compiler files; stop condition met, production unchanged.
Commits: measured tree runtime/seat16-runtime 1e88d1674c360b57d570acacc521cb7dbd452fa7; report recorded separately on codex/parse-speed after a709d575.
Commands and outputs: release -O2 parse 6,254,332,929 Ir; scalar visits 7,412,064 and 51,852,148 Ir, 0.829%; pinned best-of-ten native user 0.814669s.
Mutants: one extra scalar visit fails counter/branch reconciliation; one extra footer instruction fails self/footer reconciliation.
Not covered: no compact reference list, emitter hook, dropped-reference leak mutant, new parity run, or before/after speedup; the explicit under-2% stop applies.

# Reference-bearing destruction slots, October 7, 2026

**Stop: non-reference destruction-loop visits account for 0.829% of parse instructions, below the requested 2%.** No production runtime, shape, compiler, scanner or parser source is edited. In particular emit_objects.go and shapeWith remain untouched; there are no hook lines to clear with the compiler worker. This is a measured stop, not an optimization claim.

## Measured tree and workload

The seat 16 train was completed, gated and pushed before starting this measurement. Build inputs come from its landed tree `1e88d1674c360b57d570acacc521cb7dbd452fa7` in /workspace/adamic, including the area/runtime string views, release fixes, inline fields and other current area behavior. The separate codex/parse-speed report worktree does not substitute its older runtime for the measured tree.

The same accepted batch 8 parse-only .a driver reconstructed from 4189abd runs the same 77 TypeScript compiler files at corpus commit `050880ce59e30b356b686bd3144efe24f875ebc8`. It retains file reading, parsing, eager line table and parent map, and deletes only rule traversal from the scratch driver. Its count-mode output is `0\n`. Both instrumented and ordinary runs finish with that output. The compiler regenerated C on the measured tree rather than reusing old generated C. Evidence records source, manifest, generated-C and binary hashes.

## Scalar visits and cost

| Object path | Destroyed objects | All slot visits | Scalar visits | Reference visits | Instructions per scalar visit | Scalar instructions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Class objects | 1,083,014 | 13,679,516 | 7,379,764 | 6,299,752 | 7 | 51,658,348 |
| Plain objects | 7,274 | 82,558 | 32,300 | 50,258 | 6 | 193,800 |
| Total | 1,090,288 | 13,762,074 | 7,412,064 | 6,350,010 | | **51,852,148** |

889,146 ParseNodes are destroyed. Their seven non-reference fields are pos, end, optional, literalFlags, list, trailing and multiLine, giving 6,224,022 scalar visits. A scratch counter identifies the 13-slot kind/children/multiLine shape and counts these actual destructor entries. It changes only a copied class_inheritance.c; it is not used for performance instruction or time measurements. No element or child is freed earlier.

The uninstrumented production assembly's field-type branches at 0x6e7a8 and 0x6e7e1 take the scalar skip exactly 7,379,764 and 32,300 times in Callgrind. Those counts independently equal the instrumented counters. Total branch executions also equal scalar plus reference visits on both paths. The emitter's reference flags, not guessed source types, identify what returns immediately without examining a child header.

For a class scalar visit, the seven executed instructions are loop cmp and jbe, two metadata loads, the reference-flag cmpb, index lea and scalar jne. Plain visits execute the reference-table load, cmpb and jne plus the loop inc, cmp and jae: six. This allocates the common per-iteration loop work to each actual scalar visit, while keeping fixed prologue, class traversal and final loop-exit overhead separate. Multiplication gives the measured-path attribution above; it is not an instrumented binary's inflated instruction count.

All adamic_object_free_children self work is 158,850,468 Ir, including reference-field checks and callback transfer instructions as well as scalar visits and fixed overhead. That full function is not the requested non-reference bucket. Calls' inclusive child destruction costs are excluded from the scalar attribution.

## Unchanged baseline measurements

All native numbers in this section use the exact release clang flags below: -O2 -g, sanitizers off, -ffp-contract=off, no LTO or CPU override, no ADAMIC_COUNT.

| Observation | Measured value |
| --- | ---: |
| Whole parse instructions, collected/footer/self | 6,254,332,929 Ir |
| Scalar visit share | 0.8290596% |
| Whole parse I1 instruction-cache misses | 85,749,110 I1mr |
| Whole parse conditional plus indirect branch mispredictions | 49,269,030 |
| Whole parse data L1 misses | 20,265,120 D1mr + D1mw |
| Object child-walk self instruction-cache misses | 309,711 |
| Object child-walk branch mispredictions | 6,947,260 Bcm + Bim |
| Class field-type branch mispredictions | 3,976,020 |
| Plain field-type branch mispredictions | 5,310 |
| Native pinned best-of-ten user time | 0.814669 seconds |
| Go pinned best-of-ten user time | 0.149510 seconds |

The field-type branch miss counts combine scalar and reference outcomes; Callgrind does not attribute Bcm to a specific outcome. They cannot all be labeled scalar misses. Cache and branch results are simulator observations, not hardware perf counters or a causal prediction of user-time savings.

Valgrind 3.24.0 simulates I1 and D1 as 32 KiB, 64-byte lines, 8-way; detected LL geometry is rounded by Valgrind to 256 MiB, 64-byte lines, direct-mapped. Its stderr records that warning and a brk-segment overflow fallback. The run exits 0 and completes all files. No cache override was hidden in the command.

Every self event sums exactly to the totals footer and stderr Collected vector. The profile header summary has **two additional Ir**, with all other events identical. These remain an explicit **unassigned Callgrind header/footer remainder of 2 Ir**, not attributed to any parser or destruction mechanism. The report uses collected/footer instructions; using header instructions would not affect the stop condition. The parser retains header_delta in evidence and rejects unexpected discrepancies.

Timing warms each native/Go command once, then alternates their order over ten interleaved rounds. CPU affinity is {3}; GOMAXPROCS=1. os.wait4 captures each child's user time, and every exit/stdout/stderr is checked (exit 0, stdout `0\n`, stderr empty). Go is the existing optimized Go 1.27.1 parser driver through cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`, with normal gc compiler, GOAMD64=v1, no race/sanitizer build and no GODEBUG override for timing; its build metadata is retained. No new Go instruction claim or native/Go whole-run comparison is made.

## Exact commands and flags

Run from the measured /workspace/adamic checkout:

```sh
source /workspace/adamic-tools/env.sh
go build -o /workspace/scratch/reference-slots/adamic ./cmd/adamic > /workspace/scratch/reference-slots/compiler-build.log 2>&1
/workspace/scratch/reference-slots/adamic c /workspace/scratch/parse-speed/batch8/parse.a > /workspace/scratch/reference-slots/parse.c 2> /workspace/scratch/reference-slots/emission.log
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/reference-slots/parse.c internal/native/runtime/*.c -lm -o /workspace/scratch/reference-slots/baseline > /workspace/scratch/reference-slots/baseline-build.log 2>&1
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --cache-sim=yes --branch-sim=yes --dump-instr=yes --collect-jumps=yes --callgrind-out-file=/workspace/scratch/reference-slots/baseline.callgrind /workspace/scratch/reference-slots/baseline --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/reference-slots/baseline.stdout 2> /workspace/scratch/reference-slots/baseline.stderr
python3 /workspace/scratch/reference-slots/reference_slots.py /workspace/scratch/reference-slots > /workspace/scratch/reference-slots/result.log 2>&1
python3 /workspace/scratch/reference-slots/timing.py /workspace/scratch/reference-slots --go /workspace/scratch/parse-speed/go-parse --manifest /workspace/scratch/parse-speed/compiler.txt --cpu 3 > /workspace/scratch/reference-slots/timing.log 2>&1
```

The counter uses the same clang line, replacing include and runtime inputs with counter-runtime and output with counter. Its only instrumentation is the retained class_inheritance.c copy. It is used for event counts, never Ir or timing. Normal baseline includes reading and cleanup; no Callgrind collection toggle omits destruction.

## Measurement mutants and checks

```sh
python3 /workspace/scratch/reference-slots/reference_slots.py /workspace/scratch/reference-slots --mutate-count > /workspace/scratch/reference-slots/count-mutant.log 2>&1
python3 /workspace/scratch/reference-slots/read_profile.py /workspace/scratch/reference-slots/footer-mutant.callgrind --output /workspace/scratch/reference-slots/footer-mutant.json > /workspace/scratch/reference-slots/footer-mutant.log 2>&1
```

Both exit 1. Adding one scalar-class visit fails the independent branch/counter comparison. Adding one instruction to the totals footer fails self/footer reconciliation. These are measurement mutants, not production leak-check mutants. The latter file differs only in that single totals entry.

The measured runtime tree already passed the seat 16 build/package gates, full uncached oracle (all leak and sanitized slab lanes), vet and gofmt before this unit. Those results are in cloud/reports/seat16-runtime/report.md on runtime/seat16-runtime, not retroactively claimed as new tests of an optimization. No compact slot list or dropped-reference mutant was implemented, and no new all-77 parity run or post-change gates were required after the explicit stop condition. Existing accepted source parity is not substituted for a fresh native parity claim on this tree.

The October 8 12:00 MDT limit of 2.53G and October 9 12:00 limit of 1.685G remain unmet by this 6.254G baseline. Evidence is under reference_slot_evidence; no speculative after number or speedup is reported.

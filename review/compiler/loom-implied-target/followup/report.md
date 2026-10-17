Built: share identical plain scalar literal shapes in the cycle finder; keep the implied-target cache from 8a963329.
Commits: start 8a963329 on f93a65fb; current main advanced to 062ba8b7; submission SHAs are in the final delivery.
Validation: focused lowering, widening, cycle and refusal checks across 52 reader packages; reader guard, vet, Node and both backends.
Mutants: all five new shape mutants and all four original implied-target mutants caught by assertions, not timeouts.
Not covered: mutable parallelMap admission, non-scalar shape coalescing, full package suites or the full gate; no runtime C changed.

The initial profile wraps only lower.Lower, after load.Load. At 8a963329, the sequential control took 5.718119417 seconds wall time. Of 6.19 sampled CPU seconds, cumulative frames were:

| Frame | Cumulative CPU seconds |
|---|---:|
| lower.Lower | 5.66 |
| lowering.findCycles | 5.50 |
| cycleFinder.reaches / slotsOf | 5.48 |
| Checker.IsTypeAssignableTo | 5.13 |
| Checker.checkTypeRelatedToEx | 3.79 |
| Relater.isRelatedToEx | 3.23 |
| Relater.recursiveTypeRelatedTo | 2.66 |

These are overlapping cumulative times, not additive costs. The complete profile tables and compressed CPU profiles accompany this report.

Observation: reaches scans the program's complete shapes slice for each object it visits. The 2,048 literal types have distinct checker identities but identical number/boolean fields, so this expands a dense all-pairs graph. This is another quadratic path, not an irreducible linear checker cost.

The fix shares a shape only for a fresh object literal made of directly written identifier properties, with no index or call/construct signatures and scalar field types. The key retains property order, exact checker field-type identities (including distinct literal values), field flags and readonly status. All members of a shared shape therefore have the same complete structural type and the same field edges. Reference-bearing shapes, spreads, shorthand properties, methods, accessors and computed keys retain their individual identities. The cache lives only on cycleFinder; no IR map or runtime feature changes.

The 3,072-element generated regression asserts one scalar literal shape after every registration, then runs the complete cycle analysis. It fails at the second literal when sharing is reverted, before the slow graph walk. The other regression distinguishes field names, scalar types, literal values and readonly fields, and checks that reference and non-plain shapes remain excluded.

Assumptions: step 2 authorizes fixing the cycle-finder scan found by the requested profile. Sharing is deliberately restricted to the fully described scalar shapes above; this is not a general type interning change.

The original cycle-corpus objects.a remains absent from main. Its recovered source and provenance are in ../report.md. The mutable parallelMap program still refuses in preflight before impliedTarget; the sequential control preserves its literal and worker body but uses items.map. The timings below are for that explicit control, not a claim that mutable task moves now compile. Node, generated JavaScript, native release and sanitized native all print `2048 moved objects, ordered digest 661084526`, exit zero and have empty stderr (backend-results.json).

No registered fixture is added. Generated test programs and review controls do not add counts.md rows; counts.md is unchanged. Existing fixture count validation from 8a963329 remains in the parent report. No C file, runtime line, IR program map, cohere source, or other package's test was edited.

All exact commands are in commands.txt; all test output was captured to files and is stored compressed. The reader package closure is ../reader-packages.txt. The focused selection includes invariance/widening tests, proven relations, enum and closure refusal fixtures, checked-view dispatch, cycle tests, the fresh-write refusal corpus, and stage3 cycle fixtures. Packages with no matching test were compiled only. No complete package suite or full gate ran.

The setup completed successfully; nproc=5 with a four-CPU cgroup quota. Timing lines:

```
submodule cache: off
go version go1.27.1 linux/amd64
v24.19.0
setup: go ready (0.025s)
setup: node ready (0.024s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: submodules ready (0.082s)
setup: markdown dependencies ready (0.082s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.192s)
setup: shared cache ready (0.932s)
setup: go build ready (21.116s)
setup: test binaries deferred (use --warm-tests) (21.193s)
setup: build cache warm (21.194s)
setup: build-flags commit=8a963329d4a8f27b1a146c19c4c966b0c714363c nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.01 1.38 1.90 1/170 18335 load-after=2.73 1.92 2.07 1/175 18983
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (21.218s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.VnMvMn
```

The 32 GB root filesystem filled when the guard generated its 1.8 GB export cache. Only obsolete, explicitly named regenerable export cache directories were removed; no source or evidence was removed. No validation failed from this condition.

Branch-history constraint: 8a963329 is already pushed and is not on main. Replacing it with a squashed commit on the same remote branch would require a forbidden force push. Preserve that branch with a fast-forward follow-up and also provide compiler/loom-implied-target-squashed as a single commit on current main. Do not open a PR or submit to Loom.

Final isolated lower.Lower wall times, with load.Load excluded:

| Elements | 8a963329 before shape sharing (s) | Final shape sharing (s) |
|---:|---:|---:|
| 512 | 0.347849 | 0.045032 |
| 1024 | 1.359086 | 0.092603 |
| 2048 | 5.822567 | 0.150931 |
| 4096 | 34.030937 | 0.332471 |

The unchanged 2,048-element sequential control took 0.153709186 seconds in the final separate profile run. The requested before/after table across both steps is:

| State | Lowering wall seconds | impliedTarget array-literal queries |
|---|---:|---:|
| Original main control (prior unit) | 32.723537201 | 4096 |
| 8a963329 control (prior unit) | 5.422958125 | 1 |
| Final control (this follow-up) | 0.153709186 | 1 |

The query count is specifically the implied-target checker query, not every checker query elsewhere in lowering. Timings are single isolated Linux runs on the four-CPU quota; the count assertions are the stable regression bar.

New test leaf times: TestCycleScalarLiteralShapesOnce 0.29 s; TestCycleScalarShapeKeysPreserveDistinctFields 0.05 s. Both include setup and are under 60 s. The reader selection also passed all six implied-target leaves (0.36 / 1.24 / 2.47 / 0.60 / 4.68 / 15.62 s), and its two new leaves were 0.83 / 0.06 s under shared package load. The mandatory existing TestCallTargetReaders passed in 20.75 s; go vet ./internal/lower exited zero with empty output.

| Mutant | Assertion that caught it | Mutant leaf seconds |
|---|---|---:|
| shape-per-literal | shape count became 2 at literal 2 | 0.23 |
| shape-ignores-name | distinct keys collided | 0.04 |
| shape-ignores-type | distinct keys collided | 0.03 |
| shape-ignores-readonly | distinct keys collided | 0.03 |
| shape-includes-reference | keyed/unkeyed count changed | 0.03 |
| array-per-element-query | second parent query | 0.09 |
| conditional-per-branch-query | second parent query | 0.04 |
| logical-per-operand-query | second parent query, all three operators | 0.06 each |
| nil-cache-miss | second tuple parent query | 0.03 |

All nine mutants exited nonzero at the expected assertions and were restored before green validation. The drivers, exact patches, logs and result JSON are included.

# Cached compiler and parallel lowering proof

The five slow C-emission gap programs all classify accepted-by-both using lowering only. The revision is ced32bf9b1f410b8be07f078adb78365a20a9f5b; four workers; classification deadline 45 s, runtime command deadline 10 s. Identical base/head binaries reuse their classification observation. No classification-result cache is used.

| Phase | Cold seconds | Warm seconds |
| --- | ---: | ---: |
| base_build | 170.094 | 0.000 |
| checkout_and_provision | 2.660 | 5.062 |
| classification | 5.261 | 5.130 |
| head_build | 0.000 | 0.000 |
| manifest_and_verification | 0.015 | 0.016 |
| runtime | 0.000 | 0.000 |
| total | 178.035 | 10.213 |

Compiler products are reused through internal/buildcache. The key includes the revision SHA (which pins source and submodules), the lowering adapter content, build settings and toolchain. The adapter calls the revision's existing compile function, without C or JavaScript emission. Classification runs one Go CPU per process, across available CPUs. Only selected newly accepted inputs are built natively and run.

Sampling alone cannot make the old C-emission classifier a 60 s gate: all inputs must be classified before sampling, and five gap programs alone cost roughly 125 s per revision sequentially. Lowering-only classification removes that measured cost, so those five compile observations do not need caching across runs. Compiler products do need caching to keep cold revision builds outside the 60 s gate unit. This five-input proof is not yet a full-corpus timing claim. Mandatory diff and witness runtime checks can also exceed the budget; --budget is a sampling reservation, not a total wall deadline.

Validation: the command tests and TestCallTargetReaders passed. Source mutants omitting SHA, entering C emission and serializing workers each failed its intended test. New TestCompilerCacheIncludesSHA and TestParallelLoweringOnly took 0.00 and 0.01 s.

The explicit --manifest-generator-revision option retains honest provenance for the approved generator not yet present at historical head. Generator verification still defaults to head.

Remaining: full generated-corpus main versus ced32bf9 run and its findings.

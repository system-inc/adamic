Built mandatory diff coverage and SHA-cached compilers with parallel lowering-only classification.
Implementation commits: c52692db (diff) and 40462ce8 (speed); this commit preserves the real run evidence.
Full generated-manifest runs with --workers 4 --budget 60 exited 0; verdict pass; admitted 1; sampled 1; omitted 0.
Mutants caught: omit outside paths, sample away diff, omit SHA, enter C emission, serialize workers; original timeout-as-refusal mutant also failed.
Not covered: runtime equivalence for accepted-by-both or refused-by-both inputs; a general 60 s guarantee for larger mandatory deltas or cold compiler products.

# Main versus fixes

Base: 2b1be38362046455e0e5454d8b0e674a8630e95d
Head: ced32bf9b1f410b8be07f078adb78365a20a9f5b
Fetched fixes tip: ced32bf9b1f410b8be07f078adb78365a20a9f5b

The exact invocation is recorded in repaired-fixes.inputs.json. The original generated manifest is preserved without clearing metadata. Generator revision 0fd9b945e15525b1dc25a840a40324bdbf0c31d0 and blob 7967672ee99c72c8dda83a82b526f1ea6e4a9141 are explicitly verified; this generator has not landed in historical head.

Generated corpora: witnesses 2, fixtures 912, gaps 107, review 10, fuzz 0. The two approved lower-testdata patterns include 111 inputs at fixes head (96 at the main SHA used to validate the generator change). Diff contains 21 paths, including all 15 array_narrowing programs, and is fully listed with blobs in JSON. There are 1,032 unique classified programs: 892 accepted-by-both, 139 refused-by-both, one newly accepted, zero newly refused and zero compiler errors/timeouts/crashes.

Newly accepted: internal/lower/testdata/array_narrowing/find.a

| Observation | Exit | stdout |
| --- | ---: | --- |
| base lowering | 1 | `""` |
| head lowering | 0 | `""` |
| Node | 0 | `"found 2 / double 4\n"` |
| JavaScript | 0 | `"found 2 / double 4\n"` |
| native release | 0 | `"found 2 / double 4\n"` |

All three execution outputs are exactly "found 2 / double 4\n", exits zero; agree true. JavaScript and native compilation also exited zero. Disagreements: none. The other 14 array_narrowing programs were already accepted by base, so they correctly classify accepted-by-both and are not runtime-checked by this admission-only gate.

| Phase | First run seconds | Compiler-cache-hit run seconds |
| --- | ---: | ---: |
| base_build | 30.210 | 0.000 |
| checkout_and_provision | 5.102 | 5.261 |
| classification | 42.343 | 42.556 |
| head_build | 0.000 | 0.000 |
| manifest_and_verification | 2.210 | 2.221 |
| runtime | 0.327 | 0.317 |
| total | 80.198 | 50.359 |

The first run missed the main compiler product and hit the fixes product. The second hit both. No classification or runtime-result cache was used. Sampling alone cannot reduce classification: the whole corpus must be classified before newly admitted runtime checks can be sampled. The old five C-emission compiles consumed about 125 s per revision sequentially. Removing C emission and classifying across four CPUs brings the full classification to about 42 s, without caching those observations. Compiler-product caching makes the measured full run fit 60 s; the cold run does not. Mandatory diff and witness runtime work may exceed --budget for other changes, so the budget is not a universal total-wall deadline.

All five original gap inputs now classify accepted-by-both, with no hangs. Timings of each lowering observation are in both result JSON files.

Implementation validation: go test ./cmd/adamic-admission-delta -count=1 -timeout 60s -v and go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout 60s passed. New test leaf times: TestDiffCoverage 0.09 s, TestDiffSamplingIsMandatory 0.00 s, TestCompilerCacheIncludesSHA 0.00 s, TestParallelLoweringOnly 0.01 s. Diff lane checks passed in 3.4 s; speed lane checks passed in 3.4 s. Source mutants and their failure logs are preserved under repair/.

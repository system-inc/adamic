Merged V3 b8dfa388 into the metadata fix without conflicts.
Merge commit: adbb495b; metadata fix: 483cf7cb.
The regression and all requested focused suites passed; no tests skipped.
The restored-store mutant failed the zero-store bound with 2,885 stores.
No additional source changes, language admissions or fixture changes were made.

The emitted JSON translation unit still has zero direct metadata stores and 2,885 runtime helper calls. TestPortElementMetadataEmission passed in 2.05 seconds with GOMAXPROCS=4. Its restored-store overlay mutant failed in 2.22 seconds at the intended bound, not at compilation. No added or touched test leaf in this merge needs a new budget; the fix's added leaf remains below 60 seconds.

The uncached oracle sweep ran TestCheckedViewArrays, TestArrayHolesMilestone, TestCheckedViewObjectPrimitiveSource and every Test function in checked_views_v2_source_test.go. The 26 V3 fixtures and formerly skipped cases passed their Node or pinned-check comparisons in both backends, including native sanitizers. No skip or failure was reported. The focused package run completed in 51.175 seconds. TestViewMixedUnionUnknownAndUnavailable in internal/native passed in 18.254 seconds; all six MixedUnion tests in internal/lower passed, package 0.430 seconds.

Commands used GOMAXPROCS=4, go test -count=1 -v and explicit -run selectors, with ADAMIC_GATE_UNCACHED=1 and -parallel=4 for the oracle sweep. The bound's run selector is ^TestPortElementMetadataEmission$. The mutant uses /tmp/v3-json-measure/merged-mutant.json as a temporary Go overlay. The native selector is ^TestViewMixedUnionUnknownAndUnavailable$, and the lower selector is MixedUnion. No whole package suite or full gate was run. All logs are under evidence/merged-*.log.gz.

The exact repository-root lane command was run after the merge commit:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The initial lane run exceeded its vet budget; a separate go vet completed successfully for internal/flow, internal/fresh, internal/ir, internal/javascript, internal/lower, internal/oracle and stage1/cohere/json. The lane rerun output was:

```text
lane checks 1.4 s: gofmt and tools on 57 Go files, t.Parallel on 7 test packages; no t.Parallel analyzer on this tree; vet 7 packages
```

Only review evidence is added after the merge. The final committed tip receives the same lane checks before push. Both requested branches are pushed with ordinary fast-forward-only Git pushes; b8dfa388 is an ancestor of the result. No force, rebase or protected branch update is used. The owner explicitly authorized updating compiler/views-v3 for this merge.

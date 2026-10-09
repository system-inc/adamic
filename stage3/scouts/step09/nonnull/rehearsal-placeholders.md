# Placeholder rehearsal

This nonlanding rehearsal refreshes the previously published 3d0620c6 with delivery c8f858b979e7461762b1742d0982dd73567b9a74. It retains train 9f16421c7 and the main merge included in delivery. Step 24 can build ahead against the observable unset storage and checked typed-use boundaries. It is not a delivery branch and does not land.

Compiler files match delivery. The rehearsal additionally runs the dominating-write readiness cases in both .a and .ts modes, retaining the earlier rehearsal's coverage while counting the new typed-use checks. Ordinary assertions remain checked in .ts and refused in .a; literal placeholder slots retain distinct null and undefined values.

The refresh conflicts were in internal/lower/non_null_impossible_test.go, readiness_test.go, refusals.go and internal/oracle/counts.md. Unsupported escaping reset results and accessor storage retain delivery's NotYet boundaries rather than the obsolete literal-read rule. The refusal predicate keeps enumeration-index admission, the main source-mode assertion policy and the literal-slot exception. The counts were regenerated. The full lower run also exposed a standalone refusal-pass caller without an IR result; the corrected delivery initializes placeholder bookkeeping for that entry path, and the final full lower run passes. An automatic merge retained an obsolete restriction on resets; it was reconciled with delivery. Both backends retain NamespaceState initialization alongside placeholder readiness resets.

## Checks

Every command writes output to a log. The compiler build, full internal/lower package, uncached source Node/native/JavaScript placeholder group including checked negatives and mutants, both counts runs and vet pass. Native fixtures use ASan/UBSan and counted/leak checks. Regenerate and verify counts, then vet lower and both backends.

```
go build -o /tmp/rehearsal-placeholders-adamic ./cmd/adamic
go test ./internal/lower -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/placeholder_nonnull_|TestPlaceholder' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
go test ./internal/oracle -run TestCountsAreRecorded -count=1
go vet ./internal/lower ./internal/native ./internal/javascript
```

The full gate, macOS and the complete 301-project native tsc corpus are not run here. Delivery's pre-push a-check, stage3 host records, stage1 comparison, changed-count explanations and independent compiler mutants are recorded in [placeholder-lowering.md](placeholder-lowering.md). Rehearsal's compiler has no additional feature or refusal changes.

## Counts

The regenerated Linux counts match delivery exactly. Delivery's report explains all 40 differences from main, comprising 19 new fixture rows and 21 existing rows whose operations now reach completion or a checked typed use. No rehearsal-only count differs.

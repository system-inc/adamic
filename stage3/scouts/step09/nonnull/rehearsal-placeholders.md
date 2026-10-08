# Placeholder rehearsal

This rehearsal merges delivery 191939326cd21c900cb80d2e4eaa58a271df71d5 into train 9f16421c7. It is not a delivery branch and does not land. The amended observable-unset read lowering remains uncommitted on the delivery worktree; this rehearsal contains the published initialization/reset machinery and its three positive fixtures.

Conflict resolutions retain assignment readiness resets alongside NamespaceState initialization in both backends. The refusal predicate retains enumeration-index admission and the train's checked-source policy, adding the literal-slot exception. Literal placeholders use readiness in both .a and .ts. Ordinary non-literal assertions retain their existing policy. Reset expressions whose results are consumed retain their refusal; an unsupported container reset stays NotYet. The old blanket literal-refusal tests now test these distinctions and dominating-write readiness in both source modes.

## Checks

All output was written to logs. `go build -o /tmp/rehearsal-adamic ./cmd/adamic` passed. `go test ./internal/lower -count=1` passed. `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/placeholder_nonnull_' -count=1 -v` passed all three fixtures: scanner, local and field. These compare source Node, native with sanitizers and counted leak checks, and the JavaScript backend. `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` passed. `go vet ./internal/lower ./internal/native ./internal/javascript` and `git diff --check` passed.

The new worktree initially lacked @types/node; it now references the existing pinned installation. The first full lower run exposed stale blanket-refusal assertions; the final full run passes after reconciling them. No new mutant was requested or run for this rehearsal. The full gate and the amended observable-unset negatives are not covered by this rehearsal.

## Counts

All 21 changed or added rows exactly match the published delivery's counts. The three additions are its completing-write fixtures. The other 18 rows change because a literal initializer reserves storage instead of stopping eagerly: the existing fixture proceeds to its write, or to its checked read, so the operations before that boundary are counted. Ordinary non-literal eager assertions retain the train's counts. No unrelated row changes.

The numeric columns below are the six columns in internal/oracle/counts.md, in the same order.

| Fixture | Before | After | Reason |
| --- | --- | --- | --- |
| internal/oracle/testdata/placeholder_nonnull_scanner.a | new | 4 / 4 / 6 / 5 / 3 / 0 | Added completing-write scanner fixture. |
| internal/oracle/testdata/placeholder_nonnull_local.a | new | 10 / 10 / 6 / 16 / 3 / 0 | Added completing-write local fixture. |
| internal/oracle/testdata/placeholder_nonnull_field.a | new | 7 / 7 / 7 / 14 / 4 / 0 | Added completing-write field fixture. |
| internal/oracle/testdata/non_null_initialized.ts | 0 / 0 / 0 / 0 / 0 / 0 | 42 / 42 / 23 / 67 / 12 / 0 | Completing writes and subsequent reads run after placeholder initialization. |
| internal/oracle/testdata/non_null_uninitialized_field.ts | 1 / 0 / 0 / 0 / 1 / 0 | 1 / 0 / 1 / 2 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_capture.ts | 0 / 0 / 0 / 0 / 0 / 0 | 2 / 0 / 1 / 0 / 2 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_exception.ts | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 3 / 2 / 1 / 0 | The exceptional path runs before the checked placeholder read. |
| internal/oracle/testdata/non_null_static_initialized.ts | 1 / 0 / 0 / 1 / 1 / 0 | 8 / 8 / 6 / 16 / 5 / 0 | Completing writes and subsequent reads run after placeholder initialization. |
| internal/oracle/testdata/non_null_static_uninitialized.ts | 1 / 0 / 0 / 1 / 1 / 0 | 1 / 0 / 1 / 2 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_optional.ts | 1 / 0 / 0 / 0 / 1 / 0 | 1 / 0 / 1 / 2 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_spread.ts | 1 / 0 / 0 / 0 / 1 / 0 | 2 / 0 / 3 / 2 / 2 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_iteration.ts | 1 / 0 / 0 / 1 / 1 / 0 | 6 / 1 / 7 / 8 / 6 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_interface.ts | 1 / 0 / 1 / 0 / 1 / 0 | 1 / 0 / 1 / 1 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| internal/oracle/testdata/non_null_uninitialized_append.ts | 0 / 0 / 0 / 0 / 0 / 0 | 0 / 0 / 1 / 0 / 0 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| stage3/interface-downcasts/default-staged.ts | 0 / 0 / 0 / 0 / 0 / 0 | 2 / 2 / 4 / 5 / 2 / 0 | Staged writes complete before the view read. |
| stage3/interface-downcasts/default-boxed-write.ts | 0 / 0 / 0 / 0 / 0 / 0 | 5 / 5 / 5 / 12 / 4 / 0 | Staged writes complete before the view read. |
| stage3/interface-downcasts/default-read-before-set.ts | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 0 / 3 / 2 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| stage3/interface-downcasts/readiness-identifier.ts | 0 / 0 / 0 / 0 / 0 / 0 | 2 / 2 / 4 / 6 / 2 / 0 | The completing write and subsequent view read run. |
| stage3/interface-downcasts/readiness-identifier-uninitialized.ts | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 0 / 3 / 3 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |
| stage3/interface-downcasts/readiness-number.ts | 0 / 0 / 0 / 0 / 0 / 0 | 3 / 3 / 4 / 8 / 3 / 0 | The completing write and subsequent view read run. |
| stage3/interface-downcasts/readiness-number-uninitialized.ts | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 0 / 3 / 3 / 1 / 0 | Receiver, capture, collection or view operations run before the checked unset read. |

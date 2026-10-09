Built: V1 checked-view optional scalar reads and writes, including the exit-70 misfit stop.
Commits: runtime descriptor correction 163e39f8; checked-write implementation follows on this branch.
Validation: Node comparisons, both backends, native sanitizers, targeted regressions, counts and CallTargetReaders pass.
Mutants: dropped store, presence publication, write guard, absence decoding, boolean decoding, union boxing, literal guard, boolean reservation, string retain and checked-view origin were caught.
Remaining: runtime review of the new checked-write hunks; full gate was not run.

The V1 pending fixture is enabled. Fitting number, boolean and string writes preserve physical storage and publish presence; aliases, deletion, reinsertion, enumeration and spread agree with Node. A mismatched write stops before storing with exit 70 and names slot, expected string and found number. The finite optional literal read rejects a hidden incompatible value. Nullable and aggregate optional checked views retain their existing refusal boundary.

The inherited CallTargetReaders failure was corrected by using CallTargets in the class-static guard and updating the renamed lower reader in its allowlist. The final reader check passed in 33.29 seconds. The earlier runtime push preceded that correction; its failing log remains recorded.

Validation logs: step2-focus.log (4.334 s), step2-regression.log, step2-source-mutants.log, step2-lower.log, step2-vet.log, step2-counts.log (21.768 s). Tests used hard timeouts and uncached oracle comparisons. Counts refreshed with go test ./internal/oracle -run ^TestCountsAreRecorded$/^fixtures$/internal/oracle/testdata/optional_field_ -count=1 -timeout=90s -args -update-counts. Five new fixture rows were unioned into the complete table; no existing optional row moved. See step2-counts-diff.md.

New test leaf timings:

--- PASS: TestOptionalFieldCheckedViewEnumeration (1.34s)
--- PASS: TestOptionalFieldCheckedViewWrites (1.74s)
--- PASS: TestOptionalFieldCheckedViewCatchesBooleanRead (2.59s)
--- PASS: TestOptionalFieldCheckedViewCatchesMissingBoxing (2.59s)
--- PASS: TestOptionalFieldCheckedViewCatchesAbsentReadAsZero (2.74s)
--- PASS: TestOptionalFieldCheckedViewLiteral (1.38s)
--- PASS: TestOptionalFieldCheckedViewCatchesMissingPublication (1.29s)
--- PASS: TestOptionalFieldCheckedViewMisfit (0.57s)
--- PASS: TestOptionalFieldCheckedViewCatchesDroppedStore (0.79s)
--- PASS: TestOptionalFieldCheckedViewCatchesMissingWriteGuard (1.41s)

Source mutants and their catches are recorded in step2-source-mutants.log and run-mutants.py; runtime and generated-JavaScript mutants run in optional_field_view_test.go. The original static-key-list, copy-state and original source mutants also passed their detection checks. An experimental redundant nullable guard mutant survived and was discarded with that redundant guard; no detection is claimed for it.

Setup: 46.470 s total, build 46.309 s; nproc 5, CPU quota 4. No full packages, full gate, WASI, Darwin or fuzz sweep was run. New runtime review is the remaining external dependency; implementation and targeted tests are green.

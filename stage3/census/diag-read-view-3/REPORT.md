# Diagnostic read-view v3: record fixes

Implementation: 2f0aa72ce0583e175028daab11ef5b6c5988d3b1. Parent: 8e3a784aa5a601f57342c07b68f145fe56252c21. One commit on the requested base. No Loom submission.

Exactly two files differ from v2:

- LEDGER.md retains every original byte and appends one type-only #ty5myjt entry after the final main entry. It covers the three readonly fields, helper constraint, public-preserving internal sorter overload, all eight reader parameters, two accepted renames and both owner-pin amendments.
- Binding-views README uses the requested Correctness wording exactly.

## Required reruns

- Zero-write adapter replay: passed, 0 writes, all five checked source files unchanged.
- Fresh normal apply: passed, 448.585 seconds. All 747 source files are byte-identical to proven v2.
- Whole lane: PASS, unfiltered and with unchanged settings; 106,366 passing, exactly the sanctioned public-API row, 0 pending.
- The lane rerun is required because artifact.py is in shard_plan.RUNTIME and its actual broad input set includes LEDGER.md. Runtime file bytes, the runtime hash and all 29 sanctioned proof inputs are unchanged; no generated files were refreshed.

The prior v2 full-census, mutant, type, declaration and JavaScript proofs apply to identical source bytes. See ../diag-read-view-2/REPORT.md. This revision reruns the checks requested for these two record edits.

See FINAL.json, record-only.patch, record-edits.json, hash-scope.json, replay.json, source-match.json, phase result records and complete lane logs.

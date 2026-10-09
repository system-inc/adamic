# Condition assertions

Built Q1 on compiler/step09-predicates-ahead after Q5 delivery 5661342e. This serves roadmap step 09. No main or area branch was pushed or merged into.

The resolved condition assertion no longer requires a target T. The existing independent truthiness verifier partitions inputs into truthy and falsy cases. Every surviving normal return must establish the truthy case. Direct failure helpers must independently prove non-return, including their bodies; a never annotation alone supplies no proof. A runtime assertion-level switch cannot establish a proof. No proof verifier was weakened to obtain admission.

For a proved condition assertion, its normal call is retained and the checker supplies the caller's condition narrowing. No predicate wrapper or check is inserted. For an unproved .ts assertion, the existing argument wrapper saves each already-evaluated argument in order, calls the implementation once, and tests the saved argument with the shared IR truthiness operation immediately after normal return. It never re-evaluates the caller's expression. A failure exits 70 with the assertion name, call coordinate, asserts branch, incoming type and truthy-condition target. The call counts as checked. Unproved .a bodies remain refused.

## The pinned 445 calls

| Assertion function | Original bodies | Proven calls | Checked calls | Pending calls |
| --- | ---: | ---: | ---: | ---: |
| Debug.assert, debug.ts:213:142 | 1 | 445 | 0 | 0 |

This is the original condition-only subset of 8a7ab17e's 580 bodies. All 79 source hashes, all 320 baseline checker diagnostics, and the exact 445 baseline caller coordinates match. Debug.assert's falsy branch reaches its independently verified failure helper unconditionally. All 445 production call constructors record one proven asserts direction. The body already had a logical proof; the previous target-T-only call gate blocked those calls. No checked site is invented for a proven body.

Measurement uses HATCH_CONDITION_ONLY=1 with the existing output-disabled checker/proof overlay. It measures body admission and call construction, not executable lowering of every compiler function or a clean adapted-corpus build. Q5's 289 complete views remain pending at their optional array/object reads. This result neither relabels them nor depends on using their incomplete field contracts.

## Fixtures and independent mutants

Six .a sources are Node witnesses. Temporary .ts copies opt into checks. Two independently proven fixtures also execute as .a. Both backends, release native, ASan/UBSan native and successful-run leak checks agree with their specified observations. The proven nullable-string fixture uses the checker's condition narrowing to read length and prints 5.

| Rule or mutant | Catcher and observation |
| --- | --- |
| Return normally on a falsy argument | The normal-return proof mutant is Refused in .a and gets one checked site in .ts |
| Throw behind assertion-level switch | The gated proof mutant is Refused in .a and gets one checked site in .ts; with the switch off and a false argument, both compiled backends exit 70 while source Node continues |
| Override throw with finally return | A third proof mutant stays Refused in .a and checked in .ts |
| Drop inserted check | Independently removed on lying, gated and false_saved fixtures; both backends reproduce Node's normally returning falsy assertion and are caught against the exit-70 contract |
| Re-evaluate instead of reuse saved condition | Saved fixture's condition returns true on its first evaluation and false on its second. Source and correct backends print `1` then `continued`. The mutant prints `1` then `2`, exits 70, and is caught in both backends |
| Lose or invent corpus evidence | Dropping one original caller, changing a proven direction to checked, and erasing its body proof each fail audit-condition-calls.py |

Initial literal-false diagnostic pins incorrectly expected boolean; they were corrected to the checker's literal false. This was a fixture pin correction, not a compiler option or contract change. Only the final green runs count below.

## Commands and counts

All output went directly to logs retained compressed in evidence/condition-*:

- `go test ./internal/lower -run 'Predicate|ConditionAssertion|TestEveryNeedsCallbackEffects' -count=1 -v`: PASS, 8.259s; includes all three body-proof mutants.
- `go test ./internal/oracle -run '^TestConditionAssertionsOracle$|^TestConditionAssertionReevaluationMutant$|^TestCheckedPredicateOracle$|^TestPredicateKindProofOracle$|^TestPredicateOptionalAliasPresenceMutant$' -count=1 -v`: PASS, 15.508s; old typed-predicate and Q5 controls stay green.
- Added the nullable-string proof witness, then `go test ./internal/oracle -run '^TestConditionAssertionCountsAreRecorded$|^TestConditionAssertionsOracle$|^TestConditionAssertionReevaluationMutant$' -count=1 -v`: PASS, 4.962s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts`: PASS, 41.266s.
- `go test ./cmd/adamic -run '^TestExplainChecksOutput$' -count=1`: PASS, 1.413s.
- Condition-only overlay probe and audit-condition-calls.py: PASS; 445 proven, zero checked or pending; no executable IR output path.

Counts add eight allocation rows: both .a and checked-mode rows for proven and proven_narrow, and checked-mode rows for lying, gated, saved and false_saved. Six assertion-direction rows are added: two proven and four checked. No existing allocation or direction row changes. Assertions that intentionally exit 70 are recorded as checked outcomes, not source Node semantic passes.

Toolchain setup is the same successful setup recorded in optional-report.md: nproc 5, Go 1.27.1, Node 24.19.0, clang 20.1.8. No full gate or whole-package test ran. Q1 is resolved; Q2–Q4 and Q6 stay recorded in corpus-report.md and await the owning views slices. Optional array/object alias conversion and general accessor effects remain pending. Indirect or unnamed assertion calls retain the existing conservative boundary.

# Flow comparison at proof 57fd72ae

Same command on each isolated unmodified checkout, identical pinned cohere 7945d102, same toolchain:
`go test ./internal/flow -count=1 -timeout 10m`

- Main 48c05d09: PASS, 117.159s.
- Library a5d5dc9: PASS, 150.472s.
- Combined proof 57fd72ae: FAIL, 109.401s.

The four failing top-level proof tests and first lines:

```
--- FAIL: TestEveryFunctionIsInSingleAssignment
flow_test.go:95: Lower: non_null_deinitialize_alias.a:5:15: Adamic 0.1 refuses a non-null assertion whose operand is exactly null; declare the variable optional and assign undefined

--- FAIL: TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/unknown_narrowing.a
trace_test.go:36: Lower: unknown_narrowing.a:2:16: stage 0 can't lower an unknown value observed outside Array.isArray or its narrowed array length yet

--- FAIL: TestEveryMutationIsInItsRange/programs/../oracle/testdata/unknown_narrowing.a
ranges_test.go:28: Lower: unknown_narrowing.a:2:16: stage 0 can't lower an unknown value observed outside Array.isArray or its narrowed array length yet

--- FAIL: TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/unknown_narrowing.a
trace_test.go:387: Lower: unknown_narrowing.a:2:16: stage 0 can't lower an unknown value observed outside Array.isArray or its narrowed array length yet
```

Each of the latter three has the same 42 failing source-fixture subtests. None of those 42 fixture files exists on either baseline, verified by git cat-file, so the identical subtest cannot be described as passing there: it is absent. Both complete baseline gates do pass. The proof's failures are lowering incompatibilities before SSA/range/trace checks. Unknown observations correspond to a85a9cb1 interacting with the existing unknown-narrowing proof; exact-nullish assertions correspond to ed6e30e2 and its fixtures; six non_null_write_* cases stop at a logical assignment to this target. That last group is a compiler-branch interaction, not observed on either baseline. The full first-line listing, including every subtest and wrapper, is in flow-failure-first-lines.json. Raw logs are fs_predicate_nonnull_flow.log, fs_flow_main_baseline.log and fs_flow_library_baseline.log beside this report.

This comparison measures the cited pre-regex proof checkpoint, not a fresh full flow gate after the regex merge. No baseline code or compiler source was changed, and no main or area branch was pushed.

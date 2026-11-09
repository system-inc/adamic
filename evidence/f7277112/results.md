# Rule G-2 pin at f7277112

Exact revision: f727711256b624ab4fb025e2facdedffc858e7ef (compiler/rule-g-2). Fetch exit 0. Detached test worktree /workspace/adamic-rule-g-2594029; exact revision, pinned submodules and clean source recorded. Go 1.27.1, clang 20.1.8, Node 24.19.0.

## 1. Formatting and focused suite

`gofmt -l internal/lower/`: exit 0, empty output.

`go test -count=1 -run 'TestAGenericTableCast|TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod' -v ./internal/lower/`: exit 0, all tests/subtests pass. No failures to quote.

Full focused stdout:

```text
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== RUN   TestAHostMethodAViewMakesGenericStops
=== PAUSE TestAHostMethodAViewMakesGenericStops
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== RUN   TestAGenericTableCastThroughAWrittenAnyStaysNotYet
=== PAUSE TestAGenericTableCastThroughAWrittenAnyStaysNotYet
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== CONT  TestAGenericTableCastThroughAWrittenAnyStaysNotYet
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== RUN   TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
=== PAUSE TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== RUN   TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts
=== PAUSE TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== CONT  TestAHostMethodAViewMakesGenericStops
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
--- PASS: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.20s)
=== CONT  TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts
--- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers (0.00s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts (0.20s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a (0.21s)
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
--- PASS: TestAGenericTableCastThroughAWrittenAnyStaysNotYet (0.00s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts (0.28s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a (0.29s)
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
--- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers (0.00s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a (0.26s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a (0.57s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts (0.60s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts (0.69s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
--- PASS: TestAGenericMethodSeenThroughANarrowerViewStops (0.00s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a (0.25s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a (0.16s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts (0.26s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts (0.17s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a (0.27s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts (0.26s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a (0.27s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts (0.16s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a (0.14s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a (0.13s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts (0.13s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts (0.13s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a (0.32s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts (0.32s)
--- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (2.21s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (2.25s)
--- PASS: TestAHostMethodAViewMakesGenericStops (0.00s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.a (0.75s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts (0.45s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	2.736s
```

Full focused stderr:

```text
```

## 2. refuseAnyRelation registration mutant

Exact line `func init() { refusalRules.register(3270, "refuseAnyRelation", (*lowering).refuseAnyRelation) }` replaced by `func init() {}` in internal/lower/any_relations.go. Patch retained as 20-mutant-patch.stdout.

`go test -count=1 -run '^TestAGenericTableCastThroughAWrittenAnyStaysNotYet$' -v ./internal/lower/`: exit 1. Both extensions newly go red: without the written-any refusal, lowering reaches the computed field name stop at 9:2 instead of the required any-refusal at 13:14. It still does not lower the full program. This test is sensitive to removing the refusal rule.

Full mutant stdout, quoting every failure in full:

```text
=== RUN   TestAGenericTableCastThroughAWrittenAnyStaysNotYet
=== PAUSE TestAGenericTableCastThroughAWrittenAnyStaysNotYet
=== CONT  TestAGenericTableCastThroughAWrittenAnyStaysNotYet
=== RUN   TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
=== PAUSE TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
=== RUN   TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts
=== PAUSE TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts
=== CONT  TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
=== CONT  TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts
    rule_g_test.go:111: got /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyStaysNotYet.ts1155923692/001/main.ts:9:2: stage 0 can't lower a computed field name yet, want the cast at 13:14 stopped for its written any
=== NAME  TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a
    rule_g_test.go:111: got /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyStaysNotYet.a2741024954/001/main.a:9:2: stage 0 can't lower a computed field name yet, want the cast at 13:14 stopped for its written any
--- FAIL: TestAGenericTableCastThroughAWrittenAnyStaysNotYet (0.00s)
    --- FAIL: TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.ts (0.12s)
    --- FAIL: TestAGenericTableCastThroughAWrittenAnyStaysNotYet/.a (0.12s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.137s
FAIL
```

Full mutant stderr:

```text
```

## Restoration

The original any_relations.go bytes were restored exactly. Final diff against f7277112 exits 0; final test worktree status is empty. No compiler changes retained. No other tests or whole-suite run were requested or performed. All commands, stdout, stderr and numeric exit codes are saved separately and in results.json; driver run.py and exact test file rule_g_test.go.txt are retained.

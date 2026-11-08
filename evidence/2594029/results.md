# Rule G rerun at 2594029

Exact tested revision: 2594029da13920577f63b2a4798325d7d76d8f4f, in detached worktree /workspace/adamic-rule-g-2594029. Evidence is appended to the existing evidence/compiler/rule-g branch; its compiler source was not advanced or changed. The exact driver, commands, temporary logging patch, mutant patches, provenance, complete output and exit codes are retained beside this report.

Scope: steps 2–4 only. Step 5 skipped as requested. Go 1.27.1, clang 20.1.8, Node 24.19.0; pinned submodules initialized recursively. The first attempt to reuse shallow submodule objects failed (02, exit 1); normal recursive initialization succeeded (06, exit 0).

## Step 2

`go test -count=1 -run 'TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod' -v ./internal/lower/`

Exit 1. All selected tests/subtests pass except TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet: lowering returns nil. The timer passes for both extensions; both stop later at the unbound-method refusal. The exact temporary t.Log lines are:

```text
    rule_g_test.go:24: stops at: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.a1293995590/001/main.a:6:13: Adamic 0.1 refuses a method read as a value (setTimeout would lose its object, and this with it); call it in an arrow that keeps the object: (value) => timers.setTimeout(value) (unbound-method)
    rule_g_test.go:24: stops at: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.ts1887164406/001/main.ts:6:13: Adamic 0.1 refuses a method read as a value (setTimeout would lose its object, and this with it); call it in an arrow that keeps the object: (value) => timers.setTimeout(value) (unbound-method)
```

The temporary log was removed immediately after this run. Full focused output, including the sole complete failure:

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
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestAHostMethodAViewMakesGenericStops
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== NAME  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:73: got <nil>, want the unfixed type parameter named as NotYet
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.22s)
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
--- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers (0.00s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a (0.22s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts (0.35s)
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
--- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers (0.00s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a (0.25s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts (0.10s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a (0.10s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts (0.22s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
    rule_g_test.go:24: stops at: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.a1293995590/001/main.a:6:13: Adamic 0.1 refuses a method read as a value (setTimeout would lose its object, and this with it); call it in an arrow that keeps the object: (value) => timers.setTimeout(value) (unbound-method)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
--- PASS: TestAHostMethodAViewMakesGenericStops (0.00s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts (1.66s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.a (1.76s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
    rule_g_test.go:24: stops at: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.ts1887164406/001/main.ts:6:13: Adamic 0.1 refuses a method read as a value (setTimeout would lose its object, and this with it); call it in an arrow that keeps the object: (value) => timers.setTimeout(value) (unbound-method)
--- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (1.70s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (1.59s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
--- PASS: TestAGenericMethodSeenThroughANarrowerViewStops (0.00s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a (0.26s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts (0.06s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a (0.14s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts (0.08s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a (0.08s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts (0.12s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a (0.15s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts (0.14s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a (0.16s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts (0.12s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts (0.21s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a (0.14s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts (0.06s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2.088s
FAIL

```

## Step 3: mutant A, rule G off

Exact requested replacement; patch retained in 20-mutant-a.patch. Exit 1. Timer .a and .ts newly go red at method-signature-style, proving sensitivity of the imported timer pin to rule G. The narrower-constraint-unused test remains green for both extensions.

Full output, including every complete failure:

```text
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
--- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers (0.00s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a (0.14s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts (0.14s)
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
    rule_g_test.go:29: the provider's relation to the slot stopped: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.ts930501549/001/main.ts:5:26: Adamic 0.1 refuses a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
    rule_g_test.go:29: the provider's relation to the slot stopped: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.a2984888062/001/main.a:5:26: Adamic 0.1 refuses a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
--- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (0.62s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (0.63s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.673s
FAIL

```

## Step 4: mutant B, condition 3 off

Exact requested replacement with `if false {`; patch retained in 30-mutant-b.patch. Exit 1. L2 reports `got <nil>` both at baseline and with this mutant. It is already red and does not demonstrate sensitivity to condition 3.

Full output, including the complete failure:

```text
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:72: got <nil>, want the unfixed type parameter named as NotYet
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.074s
FAIL

```

## Restoration and result

Both restoration comparisons exit 0; the final worktree status is empty. Compiler source and tests exactly match 2594029 after all temporary edits are removed. Step 2 still fails and condition-3 coverage is unproven. The corrected imported timer pin now demonstrates rule G sensitivity.

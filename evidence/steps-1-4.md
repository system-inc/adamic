# Steps 1–4 results

Step 1 at dda69fa1: gofmt exit 0 and empty output; vet retry exit 0 (initial attempt exit 1: configured TMPDIR missing).

Steps 2–4 at d643b42469bd229d461ca025652bf0dc360ee675.

Step 2 exit 1. Timer .a/.ts fail during Load with TS18004. New narrower-constraint-read .a/.ts fail during Load with TS2322. The condition-3 test fails because it sees an earlier generic-function-value stop. All remaining focused tests/subtests pass. Node handler.ts exit 0, stdout exactly `a\n`.

No `stops at:` t.Log line exists for either extension of the new test: lowerSourceAs calls t.Fatalf on the Load error, before returning to its t.Logf. Likewise the timer's temporary t.Logf is never reached. Their complete loader diagnostics are quoted below. This is a failure of the requested test gate, not a passing refusal test.

Step 3 mutant A exit 1: timer .a/.ts remain red at the same loader error; narrower-constraint-unused .a/.ts remain green. No selected test newly goes red, so this mutation does not demonstrate rule G coverage.

Step 4 mutant B exit 1: condition-3 test is red with the same earlier stop as baseline; it does not newly go red and therefore does not demonstrate condition-3 coverage.

Restoration: git diff against d643b424 for internal/lower is empty, exit 0. No compiler changes retained.

## Full step-2 output (including every complete failure)

```text
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== RUN   TestAHostMethodAViewMakesGenericStops
=== PAUSE TestAHostMethodAViewMakesGenericStops
=== RUN   TestNodeWatchFilePollingForms
=== PAUSE TestNodeWatchFilePollingForms
=== RUN   TestNodeWatchFileOtherFormsRefused
=== PAUSE TestNodeWatchFileOtherFormsRefused
=== RUN   TestNodeWatchFileTscCloseShape
=== PAUSE TestNodeWatchFileTscCloseShape
=== RUN   TestNodeWatchFileVoidViewCannotHideOwnedResult
=== PAUSE TestNodeWatchFileVoidViewCannotHideOwnedResult
=== RUN   TestNodeWatchFileUsesLibrarySymbols
=== PAUSE TestNodeWatchFileUsesLibrarySymbols
=== RUN   TestNodeWatchFileCallbackViews
=== PAUSE TestNodeWatchFileCallbackViews
=== RUN   TestNodeWatchFileMtimeUnaryPlus
=== PAUSE TestNodeWatchFileMtimeUnaryPlus
=== RUN   TestNodeWatchFileOptionTypeRefusals
=== PAUSE TestNodeWatchFileOptionTypeRefusals
=== RUN   TestNodeWatchTscShape
=== PAUSE TestNodeWatchTscShape
=== RUN   TestNodeWatchDirectForms
=== PAUSE TestNodeWatchDirectForms
=== RUN   TestNodeWatchOtherFormsRefused
=== PAUSE TestNodeWatchOtherFormsRefused
=== RUN   TestNodeWatchUsesLibrarySymbols
=== PAUSE TestNodeWatchUsesLibrarySymbols
=== RUN   TestNodeWatchInheritedOnStaysNotYet
=== PAUSE TestNodeWatchInheritedOnStaysNotYet
=== RUN   TestNodeWatchInheritedOnMutant
    library_node_watch_test.go:161: TestNodeWatchInheritedOnStaysNotYet goes red with the inherited declaration registered
--- PASS: TestNodeWatchInheritedOnMutant (0.83s)
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused
=== PAUSE TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers
=== PAUSE TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops
=== CONT  TestNodeWatchTscShape
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestNodeWatchFileVoidViewCannotHideOwnedResult
=== CONT  TestNodeWatchFileOtherFormsRefused
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestNodeWatchFileTscCloseShape
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
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
=== CONT  TestAHostMethodAViewMakesGenericStops
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== CONT  TestNodeWatchFilePollingForms
--- PASS: TestNodeWatchFileVoidViewCannotHideOwnedResult (1.82s)
=== CONT  TestNodeWatchFileMtimeUnaryPlus
--- PASS: TestNodeWatchFilePollingForms (1.90s)
=== CONT  TestNodeWatchFileOptionTypeRefusals
=== RUN   TestNodeWatchFileOptionTypeRefusals/persistent
--- PASS: TestNodeWatchFileTscCloseShape (1.94s)
=== CONT  TestNodeWatchFileCallbackViews
=== RUN   TestNodeWatchFileCallbackViews/h11
--- PASS: TestNodeWatchTscShape (2.01s)
=== CONT  TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/map
=== RUN   TestNodeWatchFileCallbackViews/any
--- PASS: TestNodeWatchFileMtimeUnaryPlus (1.27s)
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== RUN   TestNodeWatchFileOptionTypeRefusals/interval
=== NAME  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:91: got /tmp/adamic-gate/TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet3574840114/001/main.ts:5:27: stage 0 can't lower a generic function value without a concrete contextual signature yet, want the unfixed type parameter named as NotYet
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.33s)
=== CONT  TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers
=== RUN   TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.a
=== PAUSE TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.a
=== RUN   TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.ts
=== PAUSE TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== CONT  TestNodeWatchFileUsesLibrarySymbols
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/class
=== RUN   TestNodeWatchFileCallbackViews/unknown
--- PASS: TestNodeWatchFileOptionTypeRefusals (2.48s)
    --- PASS: TestNodeWatchFileOptionTypeRefusals/persistent (1.25s)
    --- PASS: TestNodeWatchFileOptionTypeRefusals/interval (1.20s)
=== CONT  TestNodeWatchUsesLibrarySymbols
--- PASS: TestNodeWatchUsesLibrarySymbols (0.13s)
=== CONT  TestNodeWatchInheritedOnStaysNotYet
=== RUN   TestNodeWatchInheritedOnStaysNotYet/direct
--- PASS: TestNodeWatchFileUsesLibrarySymbols (1.09s)
=== CONT  TestNodeWatchOtherFormsRefused
=== RUN   TestNodeWatchOtherFormsRefused/arguments
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/primitive_union
--- PASS: TestNodeWatchFileCallbackViews (3.20s)
    --- PASS: TestNodeWatchFileCallbackViews/h11 (1.11s)
    --- PASS: TestNodeWatchFileCallbackViews/any (1.22s)
    --- PASS: TestNodeWatchFileCallbackViews/unknown (0.86s)
=== CONT  TestNodeWatchDirectForms
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/string_union
=== RUN   TestNodeWatchInheritedOnStaysNotYet/view
=== RUN   TestNodeWatchOtherFormsRefused/option
--- PASS: TestNodeWatchDirectForms (0.95s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
--- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused (4.18s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/map (1.56s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/class (0.99s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/primitive_union (0.73s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/string_union (0.89s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts
=== RUN   TestNodeWatchOtherFormsRefused/persistent_missing
--- PASS: TestNodeWatchInheritedOnStaysNotYet (1.98s)
    --- PASS: TestNodeWatchInheritedOnStaysNotYet/direct (0.86s)
    --- PASS: TestNodeWatchInheritedOnStaysNotYet/view (1.12s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
--- PASS: TestAGenericMethodSeenThroughANarrowerViewStops (0.00s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a (0.15s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts (0.14s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts (0.24s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a (0.17s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts (0.09s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts (0.07s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a (0.08s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a (0.06s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a (0.06s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a (0.06s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts (0.09s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a (0.09s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts (0.08s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts (0.13s)
=== CONT  TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.a
    rule_g_test.go:68: Load: /tmp/adamic-gate/TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers.a539354390/001/main.a:6:7: error TS2322: Type 'StringHandler' is not assignable to type 'Handler'.
          Types of property 'handle' are incompatible.
            Type '<T extends string>(x: Animal, t: T) => string' is not assignable to type '<T>(x: Animal, t: T) => string'.
              Types of parameters 't' and 't' are incompatible.
                Type 'T' is not assignable to type 'string'.
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== CONT  TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.ts
    rule_g_test.go:68: Load: /tmp/adamic-gate/TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers.ts469165570/001/main.ts:6:7: error TS2322: Type 'StringHandler' is not assignable to type 'Handler'.
          Types of property 'handle' are incompatible.
            Type '<T extends string>(x: Animal, t: T) => string' is not assignable to type '<T>(x: Animal, t: T) => string'.
              Types of parameters 't' and 't' are incompatible.
                Type 'T' is not assignable to type 'string'.
--- FAIL: TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers (0.00s)
    --- FAIL: TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.a (0.24s)
    --- FAIL: TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.ts (0.05s)
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
--- PASS: TestNodeWatchFileOtherFormsRefused (7.28s)
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
    rule_g_test.go:22: Load: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.a2812326191/001/main.a:4:26: error TS18004: No value exists in scope for the shorthand property 'setTimeout'. Either declare one or provide an initializer.
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
--- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers (0.00s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a (0.06s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts (0.11s)
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== RUN   TestNodeWatchOtherFormsRefused/persistent_run_time
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
    rule_g_test.go:22: Load: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.ts397409887/001/main.ts:4:26: error TS18004: No value exists in scope for the shorthand property 'setTimeout'. Either declare one or provide an initializer.
--- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (0.10s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (0.08s)
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
--- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers (0.00s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a (0.20s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts (0.04s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a (0.08s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts (0.11s)
--- PASS: TestAHostMethodAViewMakesGenericStops (0.00s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts (1.11s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.a (1.12s)
=== RUN   TestNodeWatchOtherFormsRefused/arms_differ
=== RUN   TestNodeWatchOtherFormsRefused/rest_listener
=== RUN   TestNodeWatchOtherFormsRefused/numeric_slot
=== RUN   TestNodeWatchOtherFormsRefused/other_event
=== RUN   TestNodeWatchOtherFormsRefused/detached_close
=== RUN   TestNodeWatchOtherFormsRefused/detached_on
--- PASS: TestNodeWatchOtherFormsRefused (5.75s)
    --- PASS: TestNodeWatchOtherFormsRefused/arguments (1.07s)
    --- PASS: TestNodeWatchOtherFormsRefused/option (0.88s)
    --- PASS: TestNodeWatchOtherFormsRefused/persistent_missing (0.96s)
    --- PASS: TestNodeWatchOtherFormsRefused/persistent_run_time (0.58s)
    --- PASS: TestNodeWatchOtherFormsRefused/arms_differ (0.38s)
    --- PASS: TestNodeWatchOtherFormsRefused/rest_listener (0.34s)
    --- PASS: TestNodeWatchOtherFormsRefused/numeric_slot (0.36s)
    --- PASS: TestNodeWatchOtherFormsRefused/other_event (0.34s)
    --- PASS: TestNodeWatchOtherFormsRefused/detached_close (0.51s)
    --- PASS: TestNodeWatchOtherFormsRefused/detached_on (0.33s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	11.184s
FAIL
```

## Full mutant A output

```text
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
    rule_g_test.go:22: Load: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.a1995793088/001/main.a:4:26: error TS18004: No value exists in scope for the shorthand property 'setTimeout'. Either declare one or provide an initializer.
=== NAME  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
    rule_g_test.go:22: Load: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.ts2590798619/001/main.ts:4:26: error TS18004: No value exists in scope for the shorthand property 'setTimeout'. Either declare one or provide an initializer.
--- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (0.13s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (0.14s)
--- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers (0.00s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts (0.12s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a (0.14s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.175s
FAIL
```

## Full mutant B output

```text
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:90: got /tmp/adamic-gate/TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet3007918554/001/main.ts:5:27: stage 0 can't lower a generic function value without a concrete contextual signature yet, want the unfixed type parameter named as NotYet
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.10s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.131s
FAIL
```

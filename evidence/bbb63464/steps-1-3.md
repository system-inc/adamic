# bbb63464: steps 1–3

Exact head: bbb63464cd039548dacaaa7ccbde06915c291b8a. Comparison pins: main e3aa0bb7642163ebd004a71ba40a4657d27eab71 and unguarded rule G e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1. Fetched compiler/rule-g successfully, exit 0. Go 1.27.1, clang 20.1.8, Node 24.19.0. No retained compiler edits. Entire rule_g_test.go from bbb63464 copied for each comparison; exact copy retained as rule_g_test.go.txt and per-revision copies.

## Step 1

- gofmt -l internal/lower/: exit 0, empty output.
- go vet ./internal/lower/: exit 0, empty output.
- Requested focused suite: exit 0, all tests/subtests pass; no failures to quote.

The new test's exact stops at bbb63464, both extensions:

```text
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.a3887345954/001/main.a:13:14: Adamic 0.1 refuses a function taking (node: Node) => T | undefined seen as one taking (node: Node) => T | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style); use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.ts1836708990/001/main.ts:13:14: Adamic 0.1 refuses a function taking (node: Node) => T | undefined seen as one taking (node: Node) => T | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
```

## Step 2

Main's copied test exits 0: both extensions refuse the cast at 13:14 with method-signature-style wrapped in adamic/no-unchecked-cast.

Unguarded e5531c8's copied test ALSO exits 0: both extensions refuse at 13:14, with a different diagnostic: a member typed any related to NodeA at the assertion site. It does not lower, and the location is the same rather than a later site.

## Step 3

Exact requested guard replaced with `if true {` (31-mutant-patch.stdout). Both selected tests and all four extension subtests pass, exit 0. Timer remains green as requested. The new cast pin does NOT go red: both extensions stop at the same 13:14 member-typed-any refusal as unguarded e5531c8. The test accepts that refusal because it checks only the location after checking err is non-nil. This mutation therefore does not establish guard sensitivity. No test failures occurred to quote.

The exact source was restored (33-final-restoration exit 0; 34-final-status empty). Working test tree is at bbb63464 for step 4.

## Complete outputs

### Focused suite

```sh
go test -count=1 -run 'TestAGenericTableCast|TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod|TestNodeWatch' -v ./internal/lower/
```

Exit 0. Full stdout:

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
--- PASS: TestNodeWatchInheritedOnMutant (0.76s)
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused
=== PAUSE TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== PAUSE TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops
=== CONT  TestNodeWatchTscShape
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a
=== RUN   TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== PAUSE TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts
=== CONT  TestNodeWatchFileOptionTypeRefusals
=== RUN   TestNodeWatchFileOptionTypeRefusals/persistent
=== CONT  TestNodeWatchFileMtimeUnaryPlus
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/map
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
=== CONT  TestNodeWatchFileCallbackViews
=== RUN   TestNodeWatchFileCallbackViews/h11
=== RUN   TestNodeWatchFileOptionTypeRefusals/interval
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/class
=== RUN   TestNodeWatchFileCallbackViews/any
--- PASS: TestNodeWatchFileMtimeUnaryPlus (2.02s)
=== CONT  TestNodeWatchFileUsesLibrarySymbols
--- PASS: TestNodeWatchTscShape (2.05s)
=== CONT  TestNodeWatchFileVoidViewCannotHideOwnedResult
--- PASS: TestNodeWatchFileVoidViewCannotHideOwnedResult (1.97s)
=== CONT  TestNodeWatchFileTscCloseShape
=== RUN   TestNodeWatchFileCallbackViews/unknown
--- PASS: TestNodeWatchFileOptionTypeRefusals (4.04s)
    --- PASS: TestNodeWatchFileOptionTypeRefusals/persistent (1.87s)
    --- PASS: TestNodeWatchFileOptionTypeRefusals/interval (2.17s)
=== CONT  TestNodeWatchFileOtherFormsRefused
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/primitive_union
--- PASS: TestNodeWatchFileUsesLibrarySymbols (2.11s)
=== CONT  TestNodeWatchFilePollingForms
=== RUN   TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/string_union
--- PASS: TestNodeWatchFileTscCloseShape (0.93s)
=== CONT  TestAHostMethodAViewMakesGenericStops
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.a
=== RUN   TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== PAUSE TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== RUN   TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== PAUSE TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
--- PASS: TestNodeWatchFileCallbackViews (5.06s)
    --- PASS: TestNodeWatchFileCallbackViews/h11 (1.95s)
    --- PASS: TestNodeWatchFileCallbackViews/any (2.08s)
    --- PASS: TestNodeWatchFileCallbackViews/unknown (1.02s)
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
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
=== CONT  TestNodeWatchUsesLibrarySymbols
--- PASS: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.20s)
=== CONT  TestNodeWatchInheritedOnStaysNotYet
=== RUN   TestNodeWatchInheritedOnStaysNotYet/direct
--- PASS: TestNodeWatchFilePollingForms (1.04s)
=== CONT  TestNodeWatchOtherFormsRefused
=== RUN   TestNodeWatchOtherFormsRefused/arguments
--- PASS: TestNodeWatchUsesLibrarySymbols (0.11s)
=== CONT  TestNodeWatchDirectForms
--- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused (6.45s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/map (1.94s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/class (2.12s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/primitive_union (0.79s)
    --- PASS: TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused/string_union (1.60s)
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a
=== CONT  TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts
=== RUN   TestNodeWatchInheritedOnStaysNotYet/view
=== RUN   TestNodeWatchOtherFormsRefused/option
--- PASS: TestNodeWatchDirectForms (1.46s)
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
--- PASS: TestAGenericMethodSeenThroughANarrowerViewStops (0.00s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.a (0.11s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.ts (0.12s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/one_class's_generic_method_under_two_type_arguments.a (0.13s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.ts (0.20s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/o2,_an_overloaded_generic_view.a (0.23s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.ts (0.17s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g7,_an_interface_seen_as_another.a (0.08s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.a (0.09s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g6,_generic_on_the_target_side_only.ts (0.11s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.ts (0.44s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g5,_generic_on_the_source_side_only.a (0.46s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.a (0.14s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g4,_T_only_in_a_callback's_type.ts (0.33s)
    --- PASS: TestAGenericMethodSeenThroughANarrowerViewStops/g1,_T_unused_on_both_sides.ts (0.17s)
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a
=== CONT  TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts
--- PASS: TestNodeWatchInheritedOnStaysNotYet (3.01s)
    --- PASS: TestNodeWatchInheritedOnStaysNotYet/direct (1.46s)
    --- PASS: TestNodeWatchInheritedOnStaysNotYet/view (1.56s)
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts
=== RUN   TestNodeWatchOtherFormsRefused/persistent_missing
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a
=== CONT  TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts
--- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers (0.00s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).a (0.07s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.ts (0.26s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/a_class_seen_through_the_interface_it_implements.a (0.10s)
    --- PASS: TestAGenericMethodSeenAsItsIdenticalTypeLowers/g2,_an_honest_handle<T>(x:_Animal).ts (0.22s)
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.a3887345954/001/main.a:13:14: Adamic 0.1 refuses a function taking (node: Node) => T | undefined seen as one taking (node: Node) => T | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style); use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== NAME  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.ts1836708990/001/main.ts:13:14: Adamic 0.1 refuses a function taking (node: Node) => T | undefined seen as one taking (node: Node) => T | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
--- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal (0.00s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a (0.18s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts (0.36s)
=== CONT  TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts
--- PASS: TestAHostMethodAViewMakesGenericStops (0.01s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.a (1.17s)
    --- PASS: TestAHostMethodAViewMakesGenericStops/setMaxListeners.ts (1.59s)
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
--- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers (0.00s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.a (0.27s)
    --- PASS: TestAGenericMethodWithANarrowerConstraintAndTheViewsParametersLowers/.ts (0.14s)
=== RUN   TestNodeWatchOtherFormsRefused/persistent_run_time
--- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (1.64s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (1.33s)
=== RUN   TestNodeWatchOtherFormsRefused/arms_differ
--- PASS: TestNodeWatchFileOtherFormsRefused (7.43s)
=== RUN   TestNodeWatchOtherFormsRefused/rest_listener
=== RUN   TestNodeWatchOtherFormsRefused/numeric_slot
=== RUN   TestNodeWatchOtherFormsRefused/other_event
=== RUN   TestNodeWatchOtherFormsRefused/detached_close
=== RUN   TestNodeWatchOtherFormsRefused/detached_on
--- PASS: TestNodeWatchOtherFormsRefused (8.74s)
    --- PASS: TestNodeWatchOtherFormsRefused/arguments (1.46s)
    --- PASS: TestNodeWatchOtherFormsRefused/option (1.66s)
    --- PASS: TestNodeWatchOtherFormsRefused/persistent_missing (1.35s)
    --- PASS: TestNodeWatchOtherFormsRefused/persistent_run_time (1.49s)
    --- PASS: TestNodeWatchOtherFormsRefused/arms_differ (0.49s)
    --- PASS: TestNodeWatchOtherFormsRefused/rest_listener (0.48s)
    --- PASS: TestNodeWatchOtherFormsRefused/numeric_slot (0.58s)
    --- PASS: TestNodeWatchOtherFormsRefused/other_event (0.44s)
    --- PASS: TestNodeWatchOtherFormsRefused/detached_close (0.40s)
    --- PASS: TestNodeWatchOtherFormsRefused/detached_on (0.40s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	14.774s
```

Full stderr:

```text
```

### Main with copied test

```sh
go test -count=1 -run '^TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal$' -v ./internal/lower/
```

Exit 0. Full stdout:

```text
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.ts4073819133/001/main.ts:13:14: Adamic 0.1 refuses a function taking (node: Node) => T | undefined seen as one taking (node: Node) => T | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
=== NAME  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.a3388361756/001/main.a:13:14: Adamic 0.1 refuses a function taking (node: Node) => T | undefined seen as one taking (node: Node) => T | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style); use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)
--- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal (0.00s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts (0.21s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a (0.22s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.261s
```

Full stderr:

```text
```

### Unguarded rule G with copied test

```sh
go test -count=1 -run '^TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal$' -v ./internal/lower/
```

Exit 0. Full stdout:

```text
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== NAME  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.a2561040789/001/main.a:13:14: stage 0 can't lower a member typed any (1(0)) related to NodeA at the assertion site yet
=== NAME  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.ts1775879698/001/main.ts:13:14: stage 0 can't lower a member typed any (1(0)) related to NodeA at the assertion site yet
--- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal (0.00s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a (0.31s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts (0.32s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.336s
```

Full stderr:

```text
```

### Guard mutant and timer control

```sh
go test -count=1 -run 'TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal|TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot' -v ./internal/lower/
```

Exit 0. Full stdout:

```text
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== RUN   TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== PAUSE TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
=== CONT  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== RUN   TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== PAUSE TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a
=== CONT  TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts
=== NAME  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.ts2586525414/001/main.ts:13:14: stage 0 can't lower a member typed any (1(0)) related to NodeA at the assertion site yet
=== NAME  TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a
    rule_g_test.go:109: stops at: /tmp/adamic-gate/TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal.a151384287/001/main.a:13:14: stage 0 can't lower a member typed any (1(0)) related to NodeA at the assertion site yet
--- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal (0.00s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.ts (0.19s)
    --- PASS: TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal/.a (0.22s)
--- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (1.00s)
    --- PASS: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (1.05s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	1.102s
```

Full stderr:

```text
```


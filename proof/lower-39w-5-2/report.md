# Adamic #39wtqdj agreement proof

HEAD: `5ff2d7f5701a2fd32e4d9a1406bf3058ea0b2617`. Base: `3340212a6f684607e6b4fa7f1a246f466856ed3a`. Date: 2026-11-12 UTC.

## Decisive results

| Check | Result |
|---|---|
| Four focused HEAD pins | PASS, every selected subtest |
| Six mutants | All detected; every observed red passes at HEAD |
| gofmt / go vet | PASS |
| Full lower / IR guards | PASS / PASS |
| Full oracle | Exit 1, 13 environment failures listed below |
| Complete admission delta | VOID, gate exit 3: one native compile timeout |
| Admission changes | 0 newly admitted, 0 newly refused; 0 classification-stderr differences |
| Timeout reruns alone at both SHAs | Both builds PASS under the same 10-second deadline |

## Focused pins and mutants

All four focused pin commands passed at clean HEAD. All six mutants were detected. No mutant-red subtest was red at HEAD. The table explicitly names a red only for mutant 4 (shorthand), which was observed; the other five rows request measured red sets, recorded below.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` before work. Created its missing `/tmp/adamic-gate` outside the repository. Fetched `origin compiler/predicate-false-read-5`, checked out HEAD detached, and initialized recursive submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. HEAD has the specified base as its sole parent.

Provisioned `https://github.com/microsoft/TypeScript.git` at `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/adamic-proof-39w/typescript`; exported `ADAMIC_TYPESCRIPT_SOURCE` to it. Copied stage3/api package.json and package-lock.json to `/tmp/adamic-proof-39w/api`, ran `npm ci --ignore-scripts` there, and put its node_modules/.bin on PATH: `tsc --version` = `Version 6.0.3`. Go: `go1.27.1 linux/amd64`.

Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true` for npm and all subsequent commands, including the admission gate and its children. Heavy commands ran sequentially; the admission delta ran alone.

`gofmt -l internal/lower internal/oracle` exited 0 and printed nothing.

Command: `go vet ./internal/lower ./internal/oracle`

Exit **0**; wall **218.748 s**.

## head-2a

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **0**; wall **63.717 s**.

```text
PASS TestPredicateOverloadLocalWrittenAfterTheCall
PASS TestPredicateOverloadWithoutPredicate
PASS TestPredicateOverloadReturnsAnotherRead
PASS TestIndirectPredicateOverloadIsPending
PASS TestPredicateOverloadInPlaceCallback
PASS TestPredicateOverloadNamedCallback
PASS TestPredicateOverloadStaysOutOfA
PASS TestPredicateOverloadCallbackFromParameter
PASS TestPredicateOverloadCallbackPushes
PASS TestBrandedSupertypeIdentityNarrowingLowers
PASS TestKindPredicateThroughModules
PASS TestKindPredicateThroughModules/import_as
PASS TestKindPredicateThroughModules/namespace_import
PASS TestKindPredicateThroughModules/re-export_as
PASS TestKindPredicateCallIsItsComparison
PASS TestKindPredicateCallIsItsComparison/true_and_false
PASS TestKindPredicateCallIsItsComparison/!==_true
PASS TestKindPredicateCallIsItsComparison/const_arrow
PASS TestKindPredicateCallIsItsComparison/walked_call_beside_filter
PASS TestKindPredicateCallIsItsComparison/early_return
PASS TestKindPredicateCallIsItsComparison/===_false
PASS TestKindPredicateCallIsItsComparison/negated
PASS TestKindPredicateCallIsItsComparison/conditional
PASS TestCheckedPredicateFalseReadDirections
PASS TestCheckedPredicateFalseReadDirections/empty_branches
PASS TestCheckedPredicateFalseReadDirections/renarrowed
PASS TestCheckedPredicateFalseReadDirections/const_alias
PASS TestCheckedPredicateFalseReadDirections/&&_operand
PASS TestCheckedPredicateFalseReadDirections/||_operand
PASS TestCheckedPredicateFalseReadDirections/conditional
PASS TestCheckedPredicateFalseReadDirections/the_same_predicate_again
PASS TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
PASS TestCheckedPredicateFalseReadDirections/getter
PASS TestCheckedPredicateFalseReadDirections/spread
PASS TestCheckedPredicateFalseReadDirections/false_branch
PASS TestCheckedPredicateFalseReadDirections/else
PASS TestCheckedPredicateFalseReadDirections/after_an_early_exit
PASS TestCheckedPredicateFalseReadDirections/capture
PASS TestCheckedPredicateFalseReadDirections/semver's_toString
PASS TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
PASS TestCheckedPredicateFalseReadDirections/replaced_before_the_read
PASS TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
PASS TestPredicateOverloadLanguageLies
PASS TestPredicateOverloadLanguageLies/field_reassigned
PASS TestPredicateOverloadLanguageLies/or_beside_the_call
PASS TestPredicateOverloadLanguageLies/lying_named_predicate
PASS TestKindPredicateWalkRefusals
PASS TestKindPredicateWalkRefusals/comparison_control
PASS TestKindPredicateWalkRefusals/this_parameter
PASS TestKindPredicateWalkRefusals/inferred_predicate
PASS TestKindPredicateWalkRefusals/generic,_reduced
PASS TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
PASS TestKindPredicateWalkRefusals/twin_layouts
PASS TestKindPredicateWalkRefusals/lying_declared_type
PASS TestKindPredicateWalkRefusals/shared_value_comparison
PASS TestKindPredicateWalkRefusals/shared_value_call
PASS TestKindPredicateWalkRefusals/optional_parameter
PASS TestKindPredicateWalkRefusals/wide_kind_call
PASS TestKindPredicateWalkRefusals/overloaded
PASS TestKindPredicateWalkRefusals/parameter_with_a_default
PASS TestKindPredicateWalkRefusals/let_arrow
PASS TestKindPredicateWalkRefusals/value_alias
PASS TestKindPredicateWalkRefusals/method
PASS TestKindPredicateWalkRefusals/asserts
PASS TestKindPredicateWalkRefusals/switch_(true)
PASS TestKindPredicateWalkRefusals/two_statements
PASS TestKindPredicateMatchesTheWrittenComparison
PASS TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
PASS TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
PASS TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
PASS TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
PASS TestPredicateOverloadRuntime
PASS TestPredicateOverloadRuntime/parser_every_result
PASS TestPredicateOverloadRuntime/parser_some_result
PASS TestPredicateOverloadRuntime/overload_some
PASS TestPredicateOverloadRuntime/overload_some_empty
PASS TestPredicateOverloadRuntime/overload_some_false_valid
PASS TestPredicateOverloadRuntime/overload_some_false_read
PASS TestPredicateOverloadRuntime/overload_some_objects
PASS TestPredicateOverloadRuntime/overload_some_true_only
PASS TestPredicateOverloadRuntime/overload_callback
PASS TestPredicateOverloadRuntime/overload_erased
PASS TestPredicateOverloadRuntime/overload_checked
PASS TestPredicateOverloadRuntime/overload_false
PASS TestPredicateOverloadRuntime/overload_once
PASS TestPredicateOverloadRuntime/overload_every
PASS TestPredicateOverloadRuntime/overload_array_alias
PASS TestPredicateOverloadRuntime/overload_assertion
PASS TestPredicateOverloadRuntime/overload_nominal
```

## head-2b

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **0**; wall **355.590 s**.

```text
PASS TestCountsAreRecordedTable
PASS TestCheckedPredicateCountsAreRecorded
PASS TestPredicateFalseRead
PASS TestPredicateFalseRead/getter
PASS TestPredicateFalseRead/getter_capture
PASS TestPredicateFalseRead/renarrowed
PASS TestPredicateFalseRead/capture
PASS TestPredicateFalseRead/no_read
PASS TestPredicateFalseRead/logical_assignment
PASS TestPredicateFalseRead/again
PASS TestPredicateFalseRead/spread
PASS TestPredicateFalseRead/shorthand
PASS TestPredicateFalseRead/early_exit
PASS TestPredicateFalseRead/else_read
PASS TestPredicateFalseRead/false_read
PASS TestCheckedPredicateOracle
PASS TestCheckedPredicateOracle/true
PASS TestCheckedPredicateOracle/true_unused
PASS TestCheckedPredicateOracle/false
PASS TestCheckedPredicateOracle/false_alias
PASS TestCheckedPredicateOracle/false_compound_unchanged
PASS TestCheckedPredicateOracle/asserts
PASS TestCheckedPredicateOracle/valid
PASS TestCheckedPredicateOracle/asserts_valid
PASS TestCheckedPredicateOracle/condition_valid
PASS TestCheckedPredicateOracle/condition_lie
PASS TestCheckedPredicateOracle/condition_once
PASS TestCheckedPredicateOracle/condition_once_lie
PASS TestCheckedPredicateOracle/false_unchanged
PASS TestCheckedPredicateOracle/assert_not_node
PASS TestCheckedPredicateOracle/false_switch_true
PASS TestCheckedPredicateOracle/false_eq_false
PASS TestCheckedPredicateOracle/false_ne_true
PASS TestCheckedPredicateOracle/false_false_eq
PASS TestCheckedPredicateOracle/false_comma
```

## head-2c

Command: `go test ./cmd/adamic -run Explain -count=1 -v -timeout 30m`

Exit **0**; wall **29.639 s**.

```text
PASS TestExplainChecksDriver
PASS TestExplainChecksListsBrandMints
PASS TestExplainDriversReadTheTestsBuildCache
PASS TestPlaceholderExplainChecksOutputJavaScript
PASS TestPlaceholderExplainChecksOutputC
PASS TestPlaceholderExplainChecksOutputBuild
PASS TestExplainChecksOutput
PASS TestExplainChecksOutput/overload_some_empty
PASS TestExplainChecksOutput/overload_some_empty/c
PASS TestExplainChecksOutput/overload_some_empty/js
PASS TestExplainChecksOutput/overload_some_empty/build
PASS TestExplainChecksOutput/overload_some_false_read
PASS TestExplainChecksOutput/overload_some_false_read/c
PASS TestExplainChecksOutput/overload_some_false_read/js
PASS TestExplainChecksOutput/overload_some_false_read/build
PASS TestExplainChecksOutput/overload_erased
PASS TestExplainChecksOutput/overload_erased/c
PASS TestExplainChecksOutput/overload_erased/js
PASS TestExplainChecksOutput/overload_erased/build
PASS TestExplainChecksOutput/overload_assertion
PASS TestExplainChecksOutput/overload_assertion/c
PASS TestExplainChecksOutput/overload_assertion/js
PASS TestExplainChecksOutput/overload_assertion/build
PASS TestExplainChecksOutput/overload_some_true_only
PASS TestExplainChecksOutput/overload_some_true_only/c
PASS TestExplainChecksOutput/overload_some_true_only/js
PASS TestExplainChecksOutput/overload_some_true_only/build
PASS TestNonNullExplainChecks
PASS TestNonNullExplainChecks/narrowed
PASS TestNonNullExplainChecks/narrowed/c
PASS TestNonNullExplainChecks/narrowed/js
PASS TestNonNullExplainChecks/narrowed/build
PASS TestNonNullExplainChecks/initializer
PASS TestNonNullExplainChecks/initializer/c
PASS TestNonNullExplainChecks/initializer/js
PASS TestNonNullExplainChecks/initializer/build
PASS TestNonNullExplainChecks/refuse_null
PASS TestNonNullExplainChecks/refuse_null/c
PASS TestNonNullExplainChecks/refuse_null/js
PASS TestNonNullExplainChecks/refuse_null/build
PASS TestNonNullExplainChecks/refuse_undefined
PASS TestNonNullExplainChecks/refuse_undefined/c
PASS TestNonNullExplainChecks/refuse_undefined/js
PASS TestNonNullExplainChecks/refuse_undefined/build
PASS TestNonNullExplainChecks/let_initializer
PASS TestNonNullExplainChecks/let_initializer/c
PASS TestNonNullExplainChecks/let_initializer/js
PASS TestNonNullExplainChecks/let_initializer/build
PASS TestNonNullExplainChecks/undefined
PASS TestNonNullExplainChecks/undefined/c
PASS TestNonNullExplainChecks/undefined/js
PASS TestNonNullExplainChecks/undefined/build
PASS TestNonNullExplainChecks/null
PASS TestNonNullExplainChecks/null/c
PASS TestNonNullExplainChecks/null/js
PASS TestNonNullExplainChecks/null/build
PASS TestNonNullExplainChecks/present
PASS TestNonNullExplainChecks/present/c
PASS TestNonNullExplainChecks/present/js
PASS TestNonNullExplainChecks/present/build
```

## head-2d

Command: `go test ./internal/lower -run '^TestBrandedCallRuleTwoIdentity$' -count=1 -v -timeout 20m`

Exit **0**; wall **4.296 s**.

```text
PASS TestBrandedCallRuleTwoIdentity
```

## Main comparison

Both conditional triggers passed at clean HEAD. Mutant 5 made overload_every red, so a supplementary exact-subtest run was performed at pinned main, with recursive submodules updated. Complete output follows. Returned to HEAD, updated submodules, and verified `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run '^TestPredicateOverloadRuntime$/^overload_every$' -count=1 -v -timeout 40m`

Exit **0**; wall **28.006 s**.

```text
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/overload_every
--- PASS: TestPredicateOverloadRuntime (2.66s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (2.66s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	2.676s
```

## Mutant 1: The condition-only arm restored

Started from clean HEAD. Inserted the exact Go block from the table before checkedPredicateDirections’ last `return directions`. Restored afterward; `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **1**; wall **36.708 s**.

```text
TestCheckedPredicateFalseReadDirections/empty_branches — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestCheckedPredicateFalseReadDirections/replaced_before_the_read — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestCheckedPredicateFalseReadDirections/semver's_toString — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestPredicateOverloadRuntime/overload_some_true_only — also red at HEAD: no
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_true_only2252690003/001/main.ts:7:9 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly number[] | undefined -> readonly number[]} {Direction:false Status:checked Reason:readonly number[] | undefined -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
TestPredicateOverloadRuntime/overload_checked — also red at HEAD: no
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_checked980726712/001/main.ts:3:50 Function:lie Overload:1 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}, want [0 1 1] (proven, checked, unobservable)
```

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **1**; wall **93.783 s**.

```text
TestCheckedPredicateCountsAreRecorded — also red at HEAD: no
    predicates_checked_test.go:280: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
    predicates_checked_test.go:280: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
TestPredicateFalseRead/no_read — also red at HEAD: no
    predicate_false_read_test.go:85: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
TestCheckedPredicateOracle/true — also red at HEAD: no
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue951951080/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
TestCheckedPredicateOracle/false — also red at HEAD: no
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse3607991767/001/false.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
TestCheckedPredicateOracle/false_alias — also red at HEAD: no
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse_alias289122181/001/false_alias.ts:2:61 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
```

For TestCheckedPredicateOracle/{true,false,false_alias}, want Proven=0, Checked=1, Unobservable=1, one site (fixture table and assertion at predicates_checked_test.go:99–101). Their full got is printed above.

## Mutant 2: Types decide alone

Started from clean HEAD. Asserted the old text occurs exactly once in the named file (observed count 1), then applied the exact replacement; table `\|` means `|`. Restored afterward; `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **1**; wall **46.702 s**.

```text
TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked false, want true
```

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **1**; wall **91.650 s**.

```text
TestPredicateFalseRead/logical_assignment — also red at HEAD: no
    predicate_false_read_test.go:85: some's false direction at 23:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/logical_assignment.ts:23:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
TestPredicateFalseRead/getter_capture — also red at HEAD: no
    predicate_false_read_test.go:85: some's false direction at 38:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/getter_capture.ts:38:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
```

## Mutant 3: A merge counts

Started from clean HEAD. Asserted the old text occurs exactly once in the named file (observed count 1), then applied the exact replacement; table `\|` means `|`. Restored afterward; `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **1**; wall **39.853 s**.

```text
TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
```

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **0**; wall **94.308 s**.

No red subtests.

## Mutant 4: The shorthand read dropped

Started from clean HEAD. Asserted the old text occurs exactly once in the named file (observed count 1), then applied the exact replacement; table `\|` means `|`. Restored afterward; `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **0**; wall **43.064 s**.

No red subtests.

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **1**; wall **88.585 s**.

```text
TestPredicateFalseRead/shorthand — also red at HEAD: no
    predicate_false_read_test.go:85: some's false direction at 23:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/shorthand.ts:23:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
```

## Mutant 5: Holders read

Started from clean HEAD. Asserted the old text occurs exactly once in the named file (observed count 1), then applied the exact replacement; table `\|` means `|`. Restored afterward; `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **1**; wall **43.873 s**.

```text
TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestCheckedPredicateFalseReadDirections/replaced_before_the_read — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
TestPredicateOverloadRuntime/overload_some_empty — also red at HEAD: no
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_empty2893479517/001/main.ts:7:20 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:number[] -> readonly number[]} {Direction:false Status:checked Reason:number[] -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
TestPredicateOverloadRuntime/overload_every — also red at HEAD: no
    predicates_overload_test.go:76: /tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_every17921789/001/main.ts:18:16: stage 0 can't lower checked predicate membership over an open tag domain yet
```

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **1**; wall **88.180 s**.

```text
TestPredicateFalseRead/no_read — also red at HEAD: no
    predicate_false_read_test.go:85: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
```

## Mutant 6: The name after a dot read

Started from clean HEAD. Asserted the old text occurs exactly once in the named file (observed count 1), then applied the exact replacement; table `\|` means `|`. Restored afterward; `git diff --exit-code` = 0.

Command: `go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

Exit **1**; wall **41.852 s**.

```text
TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference — also red at HEAD: no
    predicates_false_read_test.go:91: false side checked true, want false
```

Command: `go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

Exit **0**; wall **83.032 s**.

No red subtests.
## Step 4: full lower

Command: `bash -c 'go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1'`

Exit **0**; wall **1458.838 s**.

No failing tests; no environment failures.

```text
ok  	github.com/system-inc/adamic/internal/lower	1455.899s
```

## Step 4b: IR guards

Command: `go test ./internal/ir -count=1 -timeout 30m`

Exit **0**; wall **128.715 s**.

No failing tests; no environment failures.

```text
ok  	github.com/system-inc/adamic/internal/ir	98.596s
```
## Step 5: full oracle

Command: `bash -c 'go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1'`

Exit **1**; wall **3106.943 s**.

The first full oracle attempt started 2026-11-12 05:40:16 UTC. The tool wait was interrupted by an environment transition after approximately 11 minutes. On recovery both go (PID 47773) and oracle.test (PID 47816) were defunct; /proc/47773/stat recorded signal 9. /tmp/oracle.log was empty. No test verdict or failure list was produced. This attempt is interrupted, not a PASS or a test failure. The exact full command was rerun after confirming there was no surviving heavy process.


The completed rerun had **13 environment failures** and no other failing tests. Twelve assignment-proof tests explicitly set NODE_PATH to the repository’s absent `stage3/api/node_modules`; the requested pinned installation lives outside the repository and tsc is on PATH, but that does not satisfy these tests’ hardcoded Node module lookup. The WASI test stops at missing WASI_SYSROOT before the expected typed refusal. No compiler or test source was changed to hide these failures.

| Failing test | First error line | Environment reason |
|---|---|---|
| `TestAssignmentProofWiderTargetMutant` | `assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofWiderTarget` | `assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofCompoundMutant` | `assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofCompound` | `assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofChainedMutant` | `assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofChained` | `assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofReturnDefinedMutant` | `assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofReturnDefined` | `assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofIfNarrowing` | `assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofIfNarrowingMutant` | `assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofScannerKeywordMutant` | `assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestAssignmentProofScannerKeyword` | `assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: Node cannot find module 'typescript' through the hardcoded repository path |
| `TestStdoutColumnsRefusesWASI` | `library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT` | ENV: Missing WASI_SYSROOT |

Common assignment-proof diagnostic:
```text
Error: Cannot find module 'typescript'
Require stack:
- /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
Node.js v24.19.0
```

```text
FAIL
gate cache: native hits=3082 misses=6600
gate cache: node hits=1746 misses=6640
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	3097.453s
FAIL
```

## Step 6: admission delta

Command: `bash -c 'go run ./cmd/adamic-admission-delta --base 3340212a6f684607e6b4fa7f1a246f466856ed3a --head 5ff2d7f5701a2fd32e4d9a1406bf3058ea0b2617 --manifest /tmp/adamic-proof-39w/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json > /tmp/adamic-proof-39w/admission.json 2> /tmp/adamic-proof-39w/admission.stderr'`

Exit **1**; wall **1384.138 s**.

The complete gate ran alone, with no concurrent tests/builds, --workers 2, no corpus filter, no shard, no sampling budget, and the generated head manifest. The `go run` wrapper exited **1**; the gate itself exited **3**, as its stderr records. The gate’s observed verdict is **void**, not pass. Subsequent solo builds below resolved the one observed native-build timeout at both SHAs, but do not rewrite the full gate’s verdict.

| Field | Observed value |
|---|---|
| verdict | `"void"` |
| void | `true` |
| void_reason | `"internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout"` |
| workers | `2` |
| complete_corpus | `true` |
| classification | `"lowering-only"` |
| diff_count | `13` |
| admitted | `0` |
| omitted | `0` |
| timed_out_runs | `1` |
| retried_runs | `1` |
| still_timed_out_runs | `1` |
| negative_witness_count | `50` |
| accepted_witnesses | `49` |
| node_agreements | `13` |
| ruled_runtime_stops | `52` |
| ruled_backend_divergences | `0` |
| compile_timeout_seconds | `45` |
| runtime_timeout_seconds | `10` |
| generator_revision | `"5ff2d7f5701a2fd32e4d9a1406bf3058ea0b2617"` |
| generator_blob | `"09d061a03b1e016387e975d25ee0cd5b75ac6c1f"` |
| manifest_blob | `"4f4759f4c9d44063f63996571ba5db561bc0dd1c"` |

Distinct classified programs: **5314**. Accepted by both: **3330**. Refused by both: **1984**. Newly refused: **0**. Newly admitted: **0**.

**Classification stderr differences: none (0 programs).** Compared exact base_compile.stderr against compile.stderr for every program; therefore there are no differing base/head texts to list.

Gate stderr:
```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 1 runs timed out, 1 retried alone, 1 timed out again
changed programs: 13
admission delta is void: internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout
exit status 3
```

Diff corpus (path and pinned blob):
```text
internal/lower/testdata/predicates_checked/false_read.a a9df503eff15df40a0f3ca4ef5e7a0704f534b22
internal/oracle/testdata/predicate_false_read/again.ts 9f6f8070d0b6347b7ceeb08c9159b78b160e172f
internal/oracle/testdata/predicate_false_read/capture.ts 1689498e103b3729216dbeaeaf3e1a8c6f635bda
internal/oracle/testdata/predicate_false_read/early_exit.ts c89883d7b0206eeb6678fce44716a1d5844459ff
internal/oracle/testdata/predicate_false_read/else_read.ts 16aad0696c46b00f5029e0bc9d3fea92bda804b9
internal/oracle/testdata/predicate_false_read/false_read.ts 5954ab19e76cecd29dd4619c61c2774ba530bbfa
internal/oracle/testdata/predicate_false_read/getter.ts 6e488b9c05361e22d3f0328ac98b45513b2953dc
internal/oracle/testdata/predicate_false_read/getter_capture.ts 540af883f64cd9e7a975524748489ae53a320518
internal/oracle/testdata/predicate_false_read/logical_assignment.ts 9f3a5719548789c4c9853c7a6f6f52ac5dbafdce
internal/oracle/testdata/predicate_false_read/no_read.ts 81a574f9ee239a4df14fa3bfc3eda7c596d73b78
internal/oracle/testdata/predicate_false_read/renarrowed.ts abedbe6418d601b79db8200bb51e4b2de8928ff7
internal/oracle/testdata/predicate_false_read/shorthand.ts fbf863f2065f06abf7715587acf756952b1bcbe0
internal/oracle/testdata/predicate_false_read/spread.ts a8bcb7a42d2ac99d056faa0fe578c491a3ce6c7b
```

All void/timeout records:
```text
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
  "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.0732908,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.116983187,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.142602456,
    "exit": 0,
    "stdout": "other\n4 false\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.150568749,
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.087679149,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.004815441,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "timeout"
  },
  "agree": false,
  "negative_related": true,
  "negative_witness": false,
  "notyet_related": false,
  "ruled_runtime_stop": false,
  "ruled_backend_divergence": false,
  "sampled": true,
  "void_reason": "native compile timeout",
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```

### Solo timeout reruns at both revisions

Only `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts` timed out. Its bytes are identical at base and HEAD. Verified the gate’s SHA-keyed compiler product keys by reproducing compilerInputs and buildcache v5 hashing: base c10af343222cf30dcf73ee53248cd384c938c47d09f7d7f7cb7e9c51f07e0764; HEAD f85037269e0ca3d59bc041787046a49eaf8c147b5aae20ce3866d4aa26bb6bf9. Checked out each SHA detached, updated recursive submodules, and ran each build alone. Both passed the original 10-second deadline. Additional builds with a diagnostic 60-second ceiling also passed. Both resulting native programs exited 70 with the same checked-field witness stop. Returned to exact HEAD, updated submodules, and verified clean diff.

Command: `timeout --kill-after=2s 10s /home/agent/.cache/adamic-build/local/c10af343222cf30dcf73ee53248cd384c938c47d09f7d7f7cb7e9c51f07e0764/adamic build internal/oracle/testdata/intersection_views_stops/nested_boolean.ts -o /tmp/adamic-proof-39w/nested-base`

Exit **0**; wall **2.836 s**.

```text
(no stdout or stderr)
```

Command: `timeout --kill-after=2s 60s /home/agent/.cache/adamic-build/local/c10af343222cf30dcf73ee53248cd384c938c47d09f7d7f7cb7e9c51f07e0764/adamic build internal/oracle/testdata/intersection_views_stops/nested_boolean.ts -o /tmp/adamic-proof-39w/nested-base`

Exit **0**; wall **2.522 s**.

```text
(no stdout or stderr)
```

Command: `/tmp/adamic-proof-39w/nested-base`

Exit **70**; wall **0.003 s**.

```text
other
adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false
```

Command: `timeout --kill-after=2s 10s /home/agent/.cache/adamic-build/local/f85037269e0ca3d59bc041787046a49eaf8c147b5aae20ce3866d4aa26bb6bf9/adamic build internal/oracle/testdata/intersection_views_stops/nested_boolean.ts -o /tmp/adamic-proof-39w/nested-head`

Exit **0**; wall **2.924 s**.

```text
(no stdout or stderr)
```

Command: `timeout --kill-after=2s 60s /home/agent/.cache/adamic-build/local/f85037269e0ca3d59bc041787046a49eaf8c147b5aae20ce3866d4aa26bb6bf9/adamic build internal/oracle/testdata/intersection_views_stops/nested_boolean.ts -o /tmp/adamic-proof-39w/nested-head`

Exit **0**; wall **2.620 s**.

```text
(no stdout or stderr)
```

Command: `/tmp/adamic-proof-39w/nested-head`

Exit **70**; wall **0.003 s**.

```text
other
adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false
```

## Report publication

Publication is limited to `proof/lower-39w-5-2/report.md` on new branch `proof/lower-39w-5-2-report`, based directly on the requested HEAD. All proof logs, package installation, upstream checkout and helpers remain outside the repository. No compiler/source change is included.

# Adamic compiler proof — #39wtqdj

HEAD: `89629dc4400e56669730e3ad048e114639c56ff9` (`compiler/predicate-false-read-6`). Base: `0c7052759671347775615af5b4a8028dd8688af8`. Report branch: `proof/lower-39w-6-2-report`.

## Decisive results

All requested HEAD pins passed. All seven single-edit mutants were detected by the requested 2a/2b commands. Every mutant-red test/subtest passed at clean HEAD. The full lower suite and IR guards passed. Admission passed over 5,314 programs: no newly admitted/refused programs and no classification stderr differences. One native build timeout cleared the gate’s isolated retry and separate isolated builds at both SHAs. Step 6 was not triggered.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` before work. Created its missing `TMPDIR=/tmp/adamic-gate` outside the repository. Fetched `origin compiler/predicate-false-read-6`, checked out the exact HEAD detached, and completed `git submodule update --init --recursive` (cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`). No AGENTS.md files were found.

Provisioned upstream Microsoft TypeScript outside the repository at `/tmp/adamic-typescript-39w`, exact commit `050880ce59e30b356b686bd3144efe24f875ebc8`, exported as `ADAMIC_TYPESCRIPT_SOURCE`. Installed stage3/api’s package.json/package-lock.json with `npm ci --prefix /tmp/adamic-proof-39w`; `tsc --version` reported `Version 6.0.3`. That external node_modules/.bin was first on PATH.

For npm, every test, admission gate, and their children: `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, `npm_config_strict_ssl=true`. Go toolchain: go1.27.1.

Scheduling deviation: initial npm provisioning and the external upstream clone overlapped recursive submodule setup. All verification commands (vet, pins, individual mutant reruns, full suites, admission) ran one at a time; admission ran alone. No compiler edits were committed or pushed.

`gofmt -l internal/lower internal/oracle`: exit 0, no output. `go vet ./internal/lower ./internal/oracle`: exit=0 wall_seconds=354, no output.

## Step 2 — HEAD pins

### 2a

`go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m`

exit=0 wall_seconds=70. No failures.

| Test/subtest | Result |
|---|---|
| `TestPredicateOverloadReturnsAnotherRead` | PASS |
| `TestPredicateOverloadWithoutPredicate` | PASS |
| `TestPredicateOverloadLocalWrittenAfterTheCall` | PASS |
| `TestIndirectPredicateOverloadIsPending` | PASS |
| `TestPredicateOverloadInPlaceCallback` | PASS |
| `TestPredicateOverloadNamedCallback` | PASS |
| `TestPredicateOverloadStaysOutOfA` | PASS |
| `TestPredicateOverloadCallbackFromParameter` | PASS |
| `TestPredicateOverloadCallbackPushes` | PASS |
| `TestBrandedSupertypeIdentityNarrowingLowers` | PASS |
| `TestKindPredicateThroughModules` | PASS |
| `TestKindPredicateThroughModules/import_as` | PASS |
| `TestKindPredicateThroughModules/namespace_import` | PASS |
| `TestKindPredicateThroughModules/re-export_as` | PASS |
| `TestKindPredicateCallIsItsComparison` | PASS |
| `TestKindPredicateCallIsItsComparison/!==_true` | PASS |
| `TestKindPredicateCallIsItsComparison/const_arrow` | PASS |
| `TestKindPredicateCallIsItsComparison/true_and_false` | PASS |
| `TestKindPredicateCallIsItsComparison/walked_call_beside_filter` | PASS |
| `TestKindPredicateCallIsItsComparison/===_false` | PASS |
| `TestKindPredicateCallIsItsComparison/early_return` | PASS |
| `TestKindPredicateCallIsItsComparison/conditional` | PASS |
| `TestKindPredicateCallIsItsComparison/negated` | PASS |
| `TestKindPredicateWalkRefusals` | PASS |
| `TestKindPredicateWalkRefusals/overloaded` | PASS |
| `TestKindPredicateWalkRefusals/comparison_control` | PASS |
| `TestKindPredicateWalkRefusals/switch_(true)` | PASS |
| `TestKindPredicateWalkRefusals/asserts` | PASS |
| `TestKindPredicateWalkRefusals/two_statements` | PASS |
| `TestKindPredicateWalkRefusals/parameter_with_a_default` | PASS |
| `TestKindPredicateWalkRefusals/optional_parameter` | PASS |
| `TestKindPredicateWalkRefusals/let_arrow` | PASS |
| `TestKindPredicateWalkRefusals/value_alias` | PASS |
| `TestKindPredicateWalkRefusals/method` | PASS |
| `TestKindPredicateWalkRefusals/twin_layouts` | PASS |
| `TestKindPredicateWalkRefusals/inferred_predicate` | PASS |
| `TestKindPredicateWalkRefusals/generic,_reduced` | PASS |
| `TestKindPredicateWalkRefusals/this_parameter` | PASS |
| `TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow` | PASS |
| `TestKindPredicateWalkRefusals/shared_value_comparison` | PASS |
| `TestKindPredicateWalkRefusals/lying_declared_type` | PASS |
| `TestKindPredicateWalkRefusals/shared_value_call` | PASS |
| `TestKindPredicateWalkRefusals/wide_kind_call` | PASS |
| `TestPredicateOverloadLanguageLies` | PASS |
| `TestPredicateOverloadLanguageLies/or_beside_the_call` | PASS |
| `TestPredicateOverloadLanguageLies/field_reassigned` | PASS |
| `TestPredicateOverloadLanguageLies/lying_named_predicate` | PASS |
| `TestKindPredicateMatchesTheWrittenComparison` | PASS |
| `TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag` | PASS |
| `TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body` | PASS |
| `TestKindPredicateMatchesTheWrittenComparison/const_enum_tag` | PASS |
| `TestCheckedPredicateFalseReadDirections` | PASS |
| `TestCheckedPredicateFalseReadDirections/renarrowed` | PASS |
| `TestCheckedPredicateFalseReadDirections/empty_branches` | PASS |
| `TestCheckedPredicateFalseReadDirections/const_alias` | PASS |
| `TestCheckedPredicateFalseReadDirections/&&_operand` | PASS |
| `TestCheckedPredicateFalseReadDirections/\|\|_operand` | PASS |
| `TestCheckedPredicateFalseReadDirections/conditional` | PASS |
| `TestCheckedPredicateFalseReadDirections/the_same_predicate_again` | PASS |
| `TestCheckedPredicateFalseReadDirections/getter` | PASS |
| `TestCheckedPredicateFalseReadDirections/spread` | PASS |
| `TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference` | PASS |
| `TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later` | PASS |
| `TestCheckedPredicateFalseReadDirections/capture` | PASS |
| `TestCheckedPredicateFalseReadDirections/after_an_early_exit` | PASS |
| `TestCheckedPredicateFalseReadDirections/else` | PASS |
| `TestCheckedPredicateFalseReadDirections/false_branch` | PASS |
| `TestCheckedPredicateFalseReadDirections/replaced_before_the_read` | PASS |
| `TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge` | PASS |
| `TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference` | PASS |
| `TestCheckedPredicateFalseReadDirections/semver's_toString` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance` | PASS |
| `TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim` | PASS |
| `TestPredicateOverloadRuntime` | PASS |
| `TestPredicateOverloadRuntime/parser_every_result` | PASS |
| `TestPredicateOverloadRuntime/parser_some_result` | PASS |
| `TestPredicateOverloadRuntime/overload_some` | PASS |
| `TestPredicateOverloadRuntime/overload_some_empty` | PASS |
| `TestPredicateOverloadRuntime/overload_some_false_valid` | PASS |
| `TestPredicateOverloadRuntime/overload_some_false_read` | PASS |
| `TestPredicateOverloadRuntime/overload_some_objects` | PASS |
| `TestPredicateOverloadRuntime/overload_some_true_only` | PASS |
| `TestPredicateOverloadRuntime/overload_callback` | PASS |
| `TestPredicateOverloadRuntime/overload_erased` | PASS |
| `TestPredicateOverloadRuntime/overload_checked` | PASS |
| `TestPredicateOverloadRuntime/overload_false` | PASS |
| `TestPredicateOverloadRuntime/overload_once` | PASS |
| `TestPredicateOverloadRuntime/overload_every` | PASS |
| `TestPredicateOverloadRuntime/overload_array_alias` | PASS |
| `TestPredicateOverloadRuntime/overload_assertion` | PASS |
| `TestPredicateOverloadRuntime/overload_nominal` | PASS |

<details><summary>Complete output: head-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== CONT  TestPredicateOverloadNamedCallback
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestPredicateOverloadWithoutPredicate
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestPredicateOverloadInPlaceCallback
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.19s)
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
--- PASS: TestPredicateOverloadWithoutPredicate (0.21s)
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.19s)
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestIndirectPredicateOverloadIsPending (0.12s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadInPlaceCallback (0.72s)
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestPredicateOverloadNamedCallback (0.73s)
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadStaysOutOfA (0.07s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadCallbackFromParameter (0.11s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestKindPredicateThroughModules/import_as
--- PASS: TestPredicateOverloadCallbackPushes (0.55s)
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.73s)
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateCallIsItsComparison/conditional
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.12s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.15s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.30s)
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (20.28s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (20.80s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (21.32s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (21.21s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (2.09s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.96s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.64s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.72s)
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.19s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.17s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.18s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.19s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.20s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.30s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.30s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.19s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.23s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.19s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.20s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.17s)
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.08s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.14s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.90s)
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.15s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.10s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.05s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
--- PASS: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.16s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.15s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.16s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.09s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.10s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (36.75s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (22.91s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.50s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.47s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.87s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.69s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.86s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.79s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.69s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	37.026s
```

</details>

### 2b

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m`

exit=0 wall_seconds=446. No failures.

| Test/subtest | Result |
|---|---|
| `TestCountsAreRecordedTable` | PASS |
| `TestCheckedPredicateCountsAreRecorded` | PASS |
| `TestPredicateFalseRead` | PASS |
| `TestPredicateFalseRead/shorthand` | PASS |
| `TestPredicateFalseRead/again` | PASS |
| `TestPredicateFalseRead/logical_assignment` | PASS |
| `TestPredicateFalseRead/no_read` | PASS |
| `TestPredicateFalseRead/getter` | PASS |
| `TestPredicateFalseRead/spread` | PASS |
| `TestPredicateFalseRead/capture` | PASS |
| `TestPredicateFalseRead/renarrowed` | PASS |
| `TestPredicateFalseRead/getter_capture` | PASS |
| `TestPredicateFalseRead/else_read` | PASS |
| `TestPredicateFalseRead/false_read` | PASS |
| `TestPredicateFalseRead/early_exit` | PASS |
| `TestCheckedPredicateOracle` | PASS |
| `TestCheckedPredicateOracle/true` | PASS |
| `TestCheckedPredicateOracle/true_unused` | PASS |
| `TestCheckedPredicateOracle/false` | PASS |
| `TestCheckedPredicateOracle/false_alias` | PASS |
| `TestCheckedPredicateOracle/false_compound_unchanged` | PASS |
| `TestCheckedPredicateOracle/asserts` | PASS |
| `TestCheckedPredicateOracle/valid` | PASS |
| `TestCheckedPredicateOracle/asserts_valid` | PASS |
| `TestCheckedPredicateOracle/condition_valid` | PASS |
| `TestCheckedPredicateOracle/condition_lie` | PASS |
| `TestCheckedPredicateOracle/condition_once` | PASS |
| `TestCheckedPredicateOracle/condition_once_lie` | PASS |
| `TestCheckedPredicateOracle/false_unchanged` | PASS |
| `TestCheckedPredicateOracle/assert_not_node` | PASS |
| `TestCheckedPredicateOracle/false_switch_true` | PASS |
| `TestCheckedPredicateOracle/false_eq_false` | PASS |
| `TestCheckedPredicateOracle/false_ne_true` | PASS |
| `TestCheckedPredicateOracle/false_false_eq` | PASS |
| `TestCheckedPredicateOracle/false_comma` | PASS |

<details><summary>Complete output: head-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== CONT  TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== CONT  TestCheckedPredicateCountsAreRecorded
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/logical_assignment
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/shorthand
=== CONT  TestPredicateFalseRead/again
--- PASS: TestCheckedPredicateCountsAreRecorded (58.14s)
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 355.17
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 347.04
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 355.17
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 miss 355.17
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 355.17
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.03
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.02
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.02
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.03
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.03
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/shorthand (406.42s)
    --- PASS: TestPredicateFalseRead/again (348.29s)
    --- PASS: TestPredicateFalseRead/logical_assignment (406.44s)
    --- PASS: TestPredicateFalseRead/no_read (406.46s)
    --- PASS: TestPredicateFalseRead/getter (0.42s)
    --- PASS: TestPredicateFalseRead/spread (1.83s)
    --- PASS: TestPredicateFalseRead/capture (1.40s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.80s)
    --- PASS: TestPredicateFalseRead/getter_capture (2.08s)
    --- PASS: TestPredicateFalseRead/else_read (1.10s)
    --- PASS: TestPredicateFalseRead/false_read (1.12s)
    --- PASS: TestPredicateFalseRead/early_exit (1.22s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.02
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 27f5e7446388 hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (429.30s)
    --- PASS: TestCheckedPredicateOracle/true (407.51s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.71s)
    --- PASS: TestCheckedPredicateOracle/false (1.10s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.07s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.04s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.31s)
    --- PASS: TestCheckedPredicateOracle/valid (1.15s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.98s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.15s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.20s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.17s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.22s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.03s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.43s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.20s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.27s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.17s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.35s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.23s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=72
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	429.428s
```

</details>

### 2c

`go test ./cmd/adamic -run 'Explain' -count=1 -v -timeout 30m`

exit=0 wall_seconds=45. No failures.

| Test/subtest | Result |
|---|---|
| `TestExplainChecksDriver` | PASS |
| `TestExplainChecksListsBrandMints` | PASS |
| `TestExplainDriversReadTheTestsBuildCache` | PASS |
| `TestPlaceholderExplainChecksOutputC` | PASS |
| `TestPlaceholderExplainChecksOutputJavaScript` | PASS |
| `TestPlaceholderExplainChecksOutputBuild` | PASS |
| `TestExplainChecksOutput` | PASS |
| `TestExplainChecksOutput/overload_some_empty` | PASS |
| `TestExplainChecksOutput/overload_some_empty/c` | PASS |
| `TestExplainChecksOutput/overload_some_empty/js` | PASS |
| `TestExplainChecksOutput/overload_some_empty/build` | PASS |
| `TestExplainChecksOutput/overload_some_true_only` | PASS |
| `TestExplainChecksOutput/overload_some_true_only/c` | PASS |
| `TestExplainChecksOutput/overload_some_true_only/js` | PASS |
| `TestExplainChecksOutput/overload_some_true_only/build` | PASS |
| `TestExplainChecksOutput/overload_erased` | PASS |
| `TestExplainChecksOutput/overload_erased/c` | PASS |
| `TestExplainChecksOutput/overload_erased/js` | PASS |
| `TestExplainChecksOutput/overload_erased/build` | PASS |
| `TestExplainChecksOutput/overload_some_false_read` | PASS |
| `TestExplainChecksOutput/overload_some_false_read/c` | PASS |
| `TestExplainChecksOutput/overload_some_false_read/js` | PASS |
| `TestExplainChecksOutput/overload_some_false_read/build` | PASS |
| `TestExplainChecksOutput/overload_assertion` | PASS |
| `TestExplainChecksOutput/overload_assertion/c` | PASS |
| `TestExplainChecksOutput/overload_assertion/js` | PASS |
| `TestExplainChecksOutput/overload_assertion/build` | PASS |
| `TestNonNullExplainChecks` | PASS |
| `TestNonNullExplainChecks/narrowed` | PASS |
| `TestNonNullExplainChecks/narrowed/c` | PASS |
| `TestNonNullExplainChecks/narrowed/js` | PASS |
| `TestNonNullExplainChecks/narrowed/build` | PASS |
| `TestNonNullExplainChecks/let_initializer` | PASS |
| `TestNonNullExplainChecks/let_initializer/c` | PASS |
| `TestNonNullExplainChecks/let_initializer/js` | PASS |
| `TestNonNullExplainChecks/let_initializer/build` | PASS |
| `TestNonNullExplainChecks/refuse_null` | PASS |
| `TestNonNullExplainChecks/refuse_null/c` | PASS |
| `TestNonNullExplainChecks/refuse_null/js` | PASS |
| `TestNonNullExplainChecks/refuse_null/build` | PASS |
| `TestNonNullExplainChecks/refuse_undefined` | PASS |
| `TestNonNullExplainChecks/refuse_undefined/c` | PASS |
| `TestNonNullExplainChecks/refuse_undefined/js` | PASS |
| `TestNonNullExplainChecks/refuse_undefined/build` | PASS |
| `TestNonNullExplainChecks/initializer` | PASS |
| `TestNonNullExplainChecks/initializer/c` | PASS |
| `TestNonNullExplainChecks/initializer/js` | PASS |
| `TestNonNullExplainChecks/initializer/build` | PASS |
| `TestNonNullExplainChecks/null` | PASS |
| `TestNonNullExplainChecks/null/c` | PASS |
| `TestNonNullExplainChecks/null/js` | PASS |
| `TestNonNullExplainChecks/null/build` | PASS |
| `TestNonNullExplainChecks/undefined` | PASS |
| `TestNonNullExplainChecks/undefined/c` | PASS |
| `TestNonNullExplainChecks/undefined/js` | PASS |
| `TestNonNullExplainChecks/undefined/build` | PASS |
| `TestNonNullExplainChecks/present` | PASS |
| `TestNonNullExplainChecks/present/c` | PASS |
| `TestNonNullExplainChecks/present/js` | PASS |
| `TestNonNullExplainChecks/present/build` | PASS |

<details><summary>Complete output: head-2c</summary>

```text
=== RUN   TestExplainChecksListsBrandMints
=== PAUSE TestExplainChecksListsBrandMints
=== RUN   TestExplainChecksOutput
=== PAUSE TestExplainChecksOutput
=== RUN   TestExplainDriversReadTheTestsBuildCache
=== PAUSE TestExplainDriversReadTheTestsBuildCache
=== RUN   TestExplainChecksDriver
--- PASS: TestExplainChecksDriver (0.00s)
=== RUN   TestPlaceholderExplainChecksOutputC
=== PAUSE TestPlaceholderExplainChecksOutputC
=== RUN   TestPlaceholderExplainChecksOutputJavaScript
=== PAUSE TestPlaceholderExplainChecksOutputJavaScript
=== RUN   TestPlaceholderExplainChecksOutputBuild
=== PAUSE TestPlaceholderExplainChecksOutputBuild
=== RUN   TestNonNullExplainChecks
=== PAUSE TestNonNullExplainChecks
=== CONT  TestExplainChecksListsBrandMints
=== CONT  TestPlaceholderExplainChecksOutputJavaScript
--- PASS: TestExplainChecksListsBrandMints (0.00s)
=== CONT  TestExplainChecksOutput
=== CONT  TestExplainDriversReadTheTestsBuildCache
=== RUN   TestExplainChecksOutput/overload_some_empty
=== PAUSE TestExplainChecksOutput/overload_some_empty
=== RUN   TestExplainChecksOutput/overload_some_false_read
=== PAUSE TestExplainChecksOutput/overload_some_false_read
=== RUN   TestExplainChecksOutput/overload_some_true_only
=== PAUSE TestExplainChecksOutput/overload_some_true_only
=== RUN   TestExplainChecksOutput/overload_erased
=== PAUSE TestExplainChecksOutput/overload_erased
=== RUN   TestExplainChecksOutput/overload_assertion
=== PAUSE TestExplainChecksOutput/overload_assertion
=== CONT  TestPlaceholderExplainChecksOutputC
=== CONT  TestPlaceholderExplainChecksOutputBuild
--- PASS: TestExplainDriversReadTheTestsBuildCache (0.00s)
=== CONT  TestExplainChecksOutput/overload_some_empty
=== CONT  TestNonNullExplainChecks
=== RUN   TestExplainChecksOutput/overload_some_empty/c
=== RUN   TestNonNullExplainChecks/narrowed
=== PAUSE TestExplainChecksOutput/overload_some_empty/c
=== RUN   TestExplainChecksOutput/overload_some_empty/js
=== PAUSE TestExplainChecksOutput/overload_some_empty/js
=== PAUSE TestNonNullExplainChecks/narrowed
=== RUN   TestNonNullExplainChecks/present
=== RUN   TestExplainChecksOutput/overload_some_empty/build
=== PAUSE TestNonNullExplainChecks/present
=== PAUSE TestExplainChecksOutput/overload_some_empty/build
=== RUN   TestNonNullExplainChecks/undefined
=== CONT  TestExplainChecksOutput/overload_erased
=== PAUSE TestNonNullExplainChecks/undefined
=== RUN   TestNonNullExplainChecks/null
=== PAUSE TestNonNullExplainChecks/null
=== RUN   TestNonNullExplainChecks/initializer
=== RUN   TestExplainChecksOutput/overload_erased/c
=== PAUSE TestNonNullExplainChecks/initializer
=== PAUSE TestExplainChecksOutput/overload_erased/c
=== RUN   TestNonNullExplainChecks/let_initializer
=== PAUSE TestNonNullExplainChecks/let_initializer
=== RUN   TestExplainChecksOutput/overload_erased/js
=== PAUSE TestExplainChecksOutput/overload_erased/js
=== RUN   TestNonNullExplainChecks/refuse_undefined
=== RUN   TestExplainChecksOutput/overload_erased/build
=== PAUSE TestNonNullExplainChecks/refuse_undefined
=== RUN   TestNonNullExplainChecks/refuse_null
=== PAUSE TestExplainChecksOutput/overload_erased/build
=== PAUSE TestNonNullExplainChecks/refuse_null
=== CONT  TestExplainChecksOutput/overload_some_true_only
=== RUN   TestExplainChecksOutput/overload_some_true_only/c
=== PAUSE TestExplainChecksOutput/overload_some_true_only/c
=== RUN   TestExplainChecksOutput/overload_some_true_only/js
=== PAUSE TestExplainChecksOutput/overload_some_true_only/js
=== RUN   TestExplainChecksOutput/overload_some_true_only/build
=== PAUSE TestExplainChecksOutput/overload_some_true_only/build
=== CONT  TestExplainChecksOutput/overload_some_false_read
=== RUN   TestExplainChecksOutput/overload_some_false_read/c
=== PAUSE TestExplainChecksOutput/overload_some_false_read/c
=== RUN   TestExplainChecksOutput/overload_some_false_read/js
=== PAUSE TestExplainChecksOutput/overload_some_false_read/js
=== RUN   TestExplainChecksOutput/overload_some_false_read/build
=== PAUSE TestExplainChecksOutput/overload_some_false_read/build
=== CONT  TestExplainChecksOutput/overload_assertion
=== RUN   TestExplainChecksOutput/overload_assertion/c
=== PAUSE TestExplainChecksOutput/overload_assertion/c
=== RUN   TestExplainChecksOutput/overload_assertion/js
=== PAUSE TestExplainChecksOutput/overload_assertion/js
=== RUN   TestExplainChecksOutput/overload_assertion/build
=== PAUSE TestExplainChecksOutput/overload_assertion/build
=== CONT  TestExplainChecksOutput/overload_some_empty/c
=== CONT  TestExplainChecksOutput/overload_some_empty/build
--- PASS: TestPlaceholderExplainChecksOutputC (0.24s)
=== CONT  TestExplainChecksOutput/overload_some_empty/js
=== CONT  TestExplainChecksOutput/overload_erased/c
--- PASS: TestPlaceholderExplainChecksOutputJavaScript (0.30s)
=== CONT  TestNonNullExplainChecks/narrowed
=== RUN   TestNonNullExplainChecks/narrowed/c
=== PAUSE TestNonNullExplainChecks/narrowed/c
=== RUN   TestNonNullExplainChecks/narrowed/js
=== PAUSE TestNonNullExplainChecks/narrowed/js
=== RUN   TestNonNullExplainChecks/narrowed/build
=== PAUSE TestNonNullExplainChecks/narrowed/build
=== CONT  TestExplainChecksOutput/overload_some_true_only/c
=== CONT  TestExplainChecksOutput/overload_erased/build
=== CONT  TestExplainChecksOutput/overload_erased/js
=== CONT  TestExplainChecksOutput/overload_some_false_read/c
=== CONT  TestNonNullExplainChecks/refuse_null
=== RUN   TestNonNullExplainChecks/refuse_null/c
=== PAUSE TestNonNullExplainChecks/refuse_null/c
=== RUN   TestNonNullExplainChecks/refuse_null/js
=== PAUSE TestNonNullExplainChecks/refuse_null/js
=== RUN   TestNonNullExplainChecks/refuse_null/build
=== PAUSE TestNonNullExplainChecks/refuse_null/build
=== CONT  TestNonNullExplainChecks/refuse_undefined
=== RUN   TestNonNullExplainChecks/refuse_undefined/c
=== PAUSE TestNonNullExplainChecks/refuse_undefined/c
=== RUN   TestNonNullExplainChecks/refuse_undefined/js
=== PAUSE TestNonNullExplainChecks/refuse_undefined/js
=== RUN   TestNonNullExplainChecks/refuse_undefined/build
=== PAUSE TestNonNullExplainChecks/refuse_undefined/build
=== CONT  TestNonNullExplainChecks/let_initializer
=== RUN   TestNonNullExplainChecks/let_initializer/c
=== PAUSE TestNonNullExplainChecks/let_initializer/c
=== RUN   TestNonNullExplainChecks/let_initializer/js
=== PAUSE TestNonNullExplainChecks/let_initializer/js
=== RUN   TestNonNullExplainChecks/let_initializer/build
=== PAUSE TestNonNullExplainChecks/let_initializer/build
=== CONT  TestNonNullExplainChecks/initializer
=== RUN   TestNonNullExplainChecks/initializer/c
=== PAUSE TestNonNullExplainChecks/initializer/c
=== RUN   TestNonNullExplainChecks/initializer/js
=== PAUSE TestNonNullExplainChecks/initializer/js
=== RUN   TestNonNullExplainChecks/initializer/build
=== PAUSE TestNonNullExplainChecks/initializer/build
=== CONT  TestNonNullExplainChecks/null
=== RUN   TestNonNullExplainChecks/null/c
=== PAUSE TestNonNullExplainChecks/null/c
=== RUN   TestNonNullExplainChecks/null/js
=== PAUSE TestNonNullExplainChecks/null/js
=== RUN   TestNonNullExplainChecks/null/build
=== PAUSE TestNonNullExplainChecks/null/build
=== CONT  TestNonNullExplainChecks/undefined
=== RUN   TestNonNullExplainChecks/undefined/c
=== PAUSE TestNonNullExplainChecks/undefined/c
=== RUN   TestNonNullExplainChecks/undefined/js
=== PAUSE TestNonNullExplainChecks/undefined/js
=== RUN   TestNonNullExplainChecks/undefined/build
=== PAUSE TestNonNullExplainChecks/undefined/build
=== CONT  TestNonNullExplainChecks/present
=== RUN   TestNonNullExplainChecks/present/c
=== PAUSE TestNonNullExplainChecks/present/c
=== RUN   TestNonNullExplainChecks/present/js
=== PAUSE TestNonNullExplainChecks/present/js
=== RUN   TestNonNullExplainChecks/present/build
=== PAUSE TestNonNullExplainChecks/present/build
=== CONT  TestExplainChecksOutput/overload_assertion/c
=== CONT  TestExplainChecksOutput/overload_some_true_only/build
=== CONT  TestExplainChecksOutput/overload_some_true_only/js
=== CONT  TestExplainChecksOutput/overload_some_false_read/js
=== CONT  TestExplainChecksOutput/overload_some_false_read/build
=== CONT  TestExplainChecksOutput/overload_assertion/js
=== CONT  TestExplainChecksOutput/overload_assertion/build
=== CONT  TestNonNullExplainChecks/narrowed/c
--- PASS: TestPlaceholderExplainChecksOutputBuild (7.01s)
=== CONT  TestNonNullExplainChecks/narrowed/build
=== CONT  TestNonNullExplainChecks/narrowed/js
=== CONT  TestNonNullExplainChecks/refuse_null/c
=== CONT  TestNonNullExplainChecks/refuse_null/build
=== CONT  TestNonNullExplainChecks/refuse_null/js
=== CONT  TestNonNullExplainChecks/refuse_undefined/c
=== CONT  TestNonNullExplainChecks/refuse_undefined/build
=== CONT  TestNonNullExplainChecks/refuse_undefined/js
=== CONT  TestNonNullExplainChecks/let_initializer/c
=== CONT  TestNonNullExplainChecks/let_initializer/build
=== CONT  TestNonNullExplainChecks/let_initializer/js
=== CONT  TestNonNullExplainChecks/initializer/c
=== CONT  TestNonNullExplainChecks/null/c
=== CONT  TestNonNullExplainChecks/initializer/build
=== CONT  TestNonNullExplainChecks/initializer/js
=== CONT  TestNonNullExplainChecks/undefined/c
--- PASS: TestExplainChecksOutput (0.00s)
    --- PASS: TestExplainChecksOutput/overload_some_empty (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/c (0.27s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/js (0.38s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/build (6.78s)
    --- PASS: TestExplainChecksOutput/overload_some_true_only (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/c (0.46s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/js (0.25s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/build (5.93s)
    --- PASS: TestExplainChecksOutput/overload_erased (0.00s)
        --- PASS: TestExplainChecksOutput/overload_erased/c (0.21s)
        --- PASS: TestExplainChecksOutput/overload_erased/js (0.24s)
        --- PASS: TestExplainChecksOutput/overload_erased/build (7.01s)
    --- PASS: TestExplainChecksOutput/overload_some_false_read (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/c (0.26s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/js (0.17s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/build (6.46s)
    --- PASS: TestExplainChecksOutput/overload_assertion (0.00s)
        --- PASS: TestExplainChecksOutput/overload_assertion/c (0.12s)
        --- PASS: TestExplainChecksOutput/overload_assertion/js (0.18s)
        --- PASS: TestExplainChecksOutput/overload_assertion/build (5.64s)
=== CONT  TestNonNullExplainChecks/null/build
=== CONT  TestNonNullExplainChecks/null/js
=== CONT  TestNonNullExplainChecks/present/c
=== CONT  TestNonNullExplainChecks/undefined/build
=== CONT  TestNonNullExplainChecks/undefined/js
=== CONT  TestNonNullExplainChecks/present/build
=== CONT  TestNonNullExplainChecks/present/js
--- PASS: TestNonNullExplainChecks (0.00s)
    --- PASS: TestNonNullExplainChecks/narrowed (0.00s)
        --- PASS: TestNonNullExplainChecks/narrowed/c (0.15s)
        --- PASS: TestNonNullExplainChecks/narrowed/js (0.17s)
        --- PASS: TestNonNullExplainChecks/narrowed/build (4.91s)
    --- PASS: TestNonNullExplainChecks/let_initializer (0.00s)
        --- PASS: TestNonNullExplainChecks/let_initializer/c (0.30s)
        --- PASS: TestNonNullExplainChecks/let_initializer/js (0.15s)
        --- PASS: TestNonNullExplainChecks/let_initializer/build (3.92s)
    --- PASS: TestNonNullExplainChecks/refuse_null (0.00s)
        --- PASS: TestNonNullExplainChecks/refuse_null/c (0.13s)
        --- PASS: TestNonNullExplainChecks/refuse_null/js (0.19s)
        --- PASS: TestNonNullExplainChecks/refuse_null/build (5.36s)
    --- PASS: TestNonNullExplainChecks/refuse_undefined (0.00s)
        --- PASS: TestNonNullExplainChecks/refuse_undefined/c (0.17s)
        --- PASS: TestNonNullExplainChecks/refuse_undefined/js (0.27s)
        --- PASS: TestNonNullExplainChecks/refuse_undefined/build (5.35s)
    --- PASS: TestNonNullExplainChecks/initializer (0.00s)
        --- PASS: TestNonNullExplainChecks/initializer/c (0.16s)
        --- PASS: TestNonNullExplainChecks/initializer/js (0.17s)
        --- PASS: TestNonNullExplainChecks/initializer/build (4.18s)
    --- PASS: TestNonNullExplainChecks/null (0.00s)
        --- PASS: TestNonNullExplainChecks/null/c (0.14s)
        --- PASS: TestNonNullExplainChecks/null/js (0.10s)
        --- PASS: TestNonNullExplainChecks/null/build (4.19s)
    --- PASS: TestNonNullExplainChecks/undefined (0.00s)
        --- PASS: TestNonNullExplainChecks/undefined/c (0.14s)
        --- PASS: TestNonNullExplainChecks/undefined/js (0.12s)
        --- PASS: TestNonNullExplainChecks/undefined/build (4.08s)
    --- PASS: TestNonNullExplainChecks/present (0.00s)
        --- PASS: TestNonNullExplainChecks/present/c (0.13s)
        --- PASS: TestNonNullExplainChecks/present/js (0.13s)
        --- PASS: TestNonNullExplainChecks/present/build (3.89s)
PASS
ok  	github.com/system-inc/adamic/cmd/adamic	16.948s
```

</details>

### 2d

`go test ./internal/lower -run '^TestBrandedCallRuleTwoIdentity$' -count=1 -v -timeout 20m`

exit=0 wall_seconds=4. No failures.

| Test/subtest | Result |
|---|---|
| `TestBrandedCallRuleTwoIdentity` | PASS |

<details><summary>Complete output: head-2d</summary>

```text
=== RUN   TestBrandedCallRuleTwoIdentity
=== PAUSE TestBrandedCallRuleTwoIdentity
=== CONT  TestBrandedCallRuleTwoIdentity
--- PASS: TestBrandedCallRuleTwoIdentity (0.82s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.840s
```

</details>

Both conditional baseline triggers passed at HEAD: `TestPredicateOverloadRuntime/overload_every` and `TestBrandedCallRuleTwoIdentity`. Therefore step 2 required no main checkout/rerun. The Go pattern is unanchored and did select all of `TestCheckedPredicateFalseReadDirections`.

## Step 3 — seven mutants

| Mutant | 2a exit / seconds | 2b exit / seconds | Red leaf tests, 2a / 2b | Also red at HEAD |
|---|---|---|---|---|
| 1 | 1 / 41 | 1 / 96 | 8 / 5 | No |
| 2 | 1 / 53 | 1 / 119 | 1 / 2 | No |
| 3 | 1 / 50 | 0 / 101 | 1 / 0 | No |
| 4 | 0 / 49 | 1 / 101 | 0 / 1 | No |
| 5 | 1 / 55 | 1 / 116 | 5 / 1 | No |
| 6 | 1 / 58 | 0 / 98 | 1 / 0 | No |
| 7 | 1 / 57 | 0 / 107 | 1 / 0 | No |

Explicit got/want interpretation (the full diagnostics below retain entire structures and paths):

- Every `false side checked` failure states its complete got/want boolean. For mutant 1 the six red direction probes each got true/want false; mutant 2’s capture got false/want true; mutants 3, 5, 6, 7’s red direction probes each got true/want false.
- Mutant 1’s `overload_some_true_only` and `overload_checked`, and mutant 5’s `overload_some_empty`: got `Proven=0, Checked=2, Unobservable=0`, want `Proven=0, Checked=1, Unobservable=1`. Full site structures are below.
- Mutant 1’s `TestCheckedPredicateOracle/{true,false,false_alias}` each got `Proven=0, Checked=2, Unobservable=0`, one site with true and false both checked; want `Proven=0, Checked=1, Unobservable=1`, one site with true checked and false unobservable. The test’s fatal line prints only got; the want here comes from `checkedPredicateFixtures`.
- Mutant 1’s recorded-count pin got measured `| 1 | 0 | 0 | 0 | 1 | 0 |` for both `false.a` and `false_alias.a` in .ts checked mode, want recorded `| 1 | 1 | 0 | 1 | 1 | 0 |`. Entire rows are below.
- Mutants 1 and 5’s `no_read`: got false direction `"checked"`, want `"unobservable"`. Mutant 2’s `logical_assignment` and `getter_capture`, and mutant 4’s `shorthand`: got false direction `"unobservable"`, want `"checked"`. Entire sites are below.
- Mutant 5’s `overload_every`: got a lowering NotYet at `main.ts:18:16`: `stage 0 can't lower checked predicate membership over an open tag domain yet`; want successful lowering (`err=nil`), then exit 0, stdout `"name\ntrue\n"`, stderr `""` on both native and JavaScript. At HEAD this entire subtest passed. This mutant-only failure does not trigger step 2’s clean-HEAD baseline condition.


Each mutant started from clean HEAD; mutants 2–7 asserted exactly one occurrence of the table’s old text in the named file before replacement. `\|` in the table was treated as literal `|`. Mutant 1 inserted the documented Go block immediately before checkedPredicateDirections’ final return. Each was restored, and `git diff --exit-code` passed after every restore. No simultaneous mutants.

### Mutant 1 — The condition-only arm restored

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=1 wall_seconds=41 |
| 2b | exit=1 wall_seconds=96 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** `TestCheckedPredicateFalseReadDirections/empty_branches`, `TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference`, `TestCheckedPredicateFalseReadDirections/replaced_before_the_read`, `TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference`, `TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge`, `TestCheckedPredicateFalseReadDirections/semver's_toString`, `TestPredicateOverloadRuntime/overload_some_true_only`, `TestPredicateOverloadRuntime/overload_checked`.

<details><summary>Complete output: mutant-1-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestPredicateOverloadWithoutPredicate
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestPredicateOverloadReturnsAnotherRead
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.18s)
=== CONT  TestPredicateOverloadNamedCallback
--- PASS: TestPredicateOverloadWithoutPredicate (0.25s)
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.25s)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestPredicateOverloadInPlaceCallback
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadCallbackPushes (0.53s)
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestIndirectPredicateOverloadIsPending (0.11s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadInPlaceCallback (0.43s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestPredicateOverloadNamedCallback (0.54s)
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadStaysOutOfA (0.09s)
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
--- PASS: TestPredicateOverloadCallbackFromParameter (0.13s)
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/early_return
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.65s)
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestKindPredicateThroughModules/import_as
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateWalkRefusals/comparison_control
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.10s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.11s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.11s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (2.12s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (2.53s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (2.80s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (2.76s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (1.15s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.32s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (1.33s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (1.44s)
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
--- PASS: TestKindPredicateWalkRefusals (0.01s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.20s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.18s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.24s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.20s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.22s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.21s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.22s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.15s)
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== NAME  TestCheckedPredicateFalseReadDirections/empty_branches
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== NAME  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.13s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.13s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.85s)
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== NAME  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== NAME  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.16s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.04s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== NAME  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== NAME  TestCheckedPredicateFalseReadDirections/semver's_toString
    predicates_false_read_test.go:93: false side checked true, want false
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/empty_branches (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.11s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.12s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.12s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.10s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.15s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/semver's_toString (0.07s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.04s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.09s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.06s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.07s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.06s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_true_only1468687416/001/main.ts:7:9 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly number[] | undefined -> readonly number[]} {Direction:false Status:checked Reason:readonly number[] | undefined -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_checked2021949861/001/main.ts:3:50 Function:lie Overload:1 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}, want [0 1 1] (proven, checked, unobservable)
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- FAIL: TestPredicateOverloadRuntime (15.72s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (3.93s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.36s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.24s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.69s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.63s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (1.01s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_true_only (0.06s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.66s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.69s)
    --- FAIL: TestPredicateOverloadRuntime/overload_checked (0.06s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.66s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.73s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.87s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.65s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.89s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	16.047s
FAIL
```

</details>

**2b red tests/subtests:** `TestCheckedPredicateCountsAreRecorded`, `TestPredicateFalseRead/no_read`, `TestCheckedPredicateOracle/true`, `TestCheckedPredicateOracle/false`, `TestCheckedPredicateOracle/false_alias`.

<details><summary>Complete output: mutant-1-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== CONT  TestPredicateFalseRead
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
--- PASS: TestCountsAreRecordedTable (0.00s)
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/logical_assignment
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:85: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/shorthand
=== NAME  TestCheckedPredicateOracle/true
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue3272239673/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/true_unused
=== CONT  TestPredicateFalseRead/again
=== NAME  TestCheckedPredicateCountsAreRecorded
    predicates_checked_test.go:280: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
    predicates_checked_test.go:280: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
--- FAIL: TestCheckedPredicateCountsAreRecorded (14.36s)
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c miss 31.06
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 31.06
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 31.06
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 26.35
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 31.07
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.03
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse386056917/001/false.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/false_alias
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.02
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.03
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.03
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestCheckedPredicateOracle/false_alias
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse_alias3357505534/001/false_alias.ts:2:61 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.02
=== NAME  TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/no_read (8.58s)
    --- PASS: TestPredicateFalseRead/getter (8.61s)
    --- PASS: TestPredicateFalseRead/spread (27.28s)
    --- PASS: TestPredicateFalseRead/logical_assignment (41.68s)
    --- PASS: TestPredicateFalseRead/again (33.15s)
    --- PASS: TestPredicateFalseRead/shorthand (33.20s)
    --- PASS: TestPredicateFalseRead/early_exit (1.51s)
    --- PASS: TestPredicateFalseRead/capture (1.44s)
    --- PASS: TestPredicateFalseRead/getter_capture (1.69s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.60s)
    --- PASS: TestPredicateFalseRead/false_read (1.68s)
    --- PASS: TestPredicateFalseRead/else_read (1.81s)
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 1a9b60be8a0c hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- FAIL: TestCheckedPredicateOracle (62.16s)
    --- FAIL: TestCheckedPredicateOracle/true (8.59s)
    --- PASS: TestCheckedPredicateOracle/true_unused (33.99s)
    --- FAIL: TestCheckedPredicateOracle/false (0.44s)
    --- FAIL: TestCheckedPredicateOracle/false_alias (0.48s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.72s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.30s)
    --- PASS: TestCheckedPredicateOracle/valid (0.99s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (1.44s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (0.97s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.36s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.95s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.18s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.93s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.47s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.24s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.39s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.32s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=67
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	62.281s
FAIL
```

</details>

```text
mutant 1 restored: clean
```

### Mutant 2 — Types decide alone

```text
old text count = 1: /workspace/adamic/internal/lower/predicates_checked.go '|| l.predicateFalseRead(call, argument)'
applied mutant 2
```

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=1 wall_seconds=53 |
| 2b | exit=1 wall_seconds=119 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** `TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later`.

<details><summary>Complete output: mutant-2-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestPredicateOverloadWithoutPredicate
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
=== CONT  TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestPredicateOverloadCallbackFromParameter
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadWithoutPredicate (0.18s)
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadCallbackFromParameter (0.20s)
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.20s)
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestPredicateOverloadStaysOutOfA (0.24s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestIndirectPredicateOverloadIsPending (0.15s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.15s)
=== CONT  TestPredicateOverloadNamedCallback
--- PASS: TestPredicateOverloadNamedCallback (0.53s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
--- PASS: TestPredicateOverloadCallbackPushes (0.58s)
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
--- PASS: TestPredicateOverloadInPlaceCallback (0.75s)
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.96s)
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.10s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.18s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (2.63s)
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== NAME  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
    predicates_false_read_test.go:93: false side checked false, want true
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateThroughModules/import_as
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.21s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.22s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.25s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.13s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.18s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.11s)
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (2.54s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (2.17s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (2.72s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (1.33s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (1.48s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.51s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.62s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (2.04s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.17s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.06s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.20s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.30s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.10s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.42s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.50s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.66s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.39s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.52s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.34s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.26s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.17s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.44s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.67s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.48s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (20.93s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (3.63s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.43s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.67s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (2.23s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (1.11s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (2.05s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.77s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.92s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.81s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.60s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.67s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.69s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (1.08s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.81s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.89s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.72s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	21.201s
FAIL
```

</details>

**2b red tests/subtests:** `TestPredicateFalseRead/logical_assignment`, `TestPredicateFalseRead/getter_capture`.

<details><summary>Complete output: mutant-2-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/shorthand
=== NAME  TestPredicateFalseRead/logical_assignment
    predicate_false_read_test.go:85: some's false direction at 23:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/logical_assignment.ts:23:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/again
--- PASS: TestCheckedPredicateCountsAreRecorded (20.96s)
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 32.56
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d miss 32.56
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 32.56
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 32.57
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 25.57
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestPredicateFalseRead/getter_capture
    predicate_false_read_test.go:85: some's false direction at 38:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/getter_capture.ts:38:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.05
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.04
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.02
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.02
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.03
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter (11.96s)
    --- FAIL: TestPredicateFalseRead/logical_assignment (12.69s)
    --- PASS: TestPredicateFalseRead/shorthand (35.73s)
    --- PASS: TestPredicateFalseRead/again (35.00s)
    --- PASS: TestPredicateFalseRead/spread (26.79s)
    --- PASS: TestPredicateFalseRead/no_read (47.76s)
    --- FAIL: TestPredicateFalseRead/getter_capture (0.50s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.35s)
    --- PASS: TestPredicateFalseRead/early_exit (1.42s)
    --- PASS: TestPredicateFalseRead/capture (1.43s)
    --- PASS: TestPredicateFalseRead/false_read (1.25s)
    --- PASS: TestPredicateFalseRead/else_read (0.97s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.02
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.02
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 80d4c1a89d5d hit 0.02
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (74.42s)
    --- PASS: TestCheckedPredicateOracle/true (48.37s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.51s)
    --- PASS: TestCheckedPredicateOracle/false (0.99s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.32s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.21s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.35s)
    --- PASS: TestCheckedPredicateOracle/valid (3.03s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (2.62s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.10s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.34s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.99s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.54s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.08s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.12s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.52s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.54s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.23s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.22s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.34s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=70
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	74.557s
FAIL
```

</details>

```text
mutant 2 restored: clean
```

### Mutant 3 — A merge counts

```text
old text count = 1: /workspace/adamic/internal/lower/predicates_checked.go 'return falseMet && (!trueMet || reduced)'
applied mutant 3
```

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=1 wall_seconds=50 |
| 2b | exit=0 wall_seconds=101 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** `TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge`.

<details><summary>Complete output: mutant-3-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== CONT  TestPredicateOverloadWithoutPredicate
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestIndirectPredicateOverloadIsPending
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadWithoutPredicate (0.19s)
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestIndirectPredicateOverloadIsPending (0.20s)
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadStaysOutOfA (0.13s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestPredicateOverloadNamedCallback
--- PASS: TestPredicateOverloadCallbackFromParameter (0.12s)
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.13s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.13s)
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.61s)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestPredicateOverloadNamedCallback (0.46s)
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/===_false
--- PASS: TestPredicateOverloadInPlaceCallback (0.50s)
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
--- PASS: TestPredicateOverloadCallbackPushes (0.56s)
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (1.95s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (2.29s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (2.51s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (2.28s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.63s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.73s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (1.93s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.77s)
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestKindPredicateThroughModules/import_as
=== NAME  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.12s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.17s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.94s)
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/overloaded
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.13s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.12s)
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.09s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.09s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.12s)
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.17s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.06s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.13s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.98s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.09s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.16s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.07s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.07s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (0.93s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (21.04s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (3.71s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.50s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.27s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (1.05s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.79s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.62s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.94s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (1.18s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (1.04s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.77s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (1.05s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (1.26s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (1.23s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (1.30s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (1.33s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (1.16s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	21.104s
FAIL
```

</details>

**2b red tests/subtests:** none.

<details><summary>Complete output: mutant-3-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== CONT  TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle/true
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/getter
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/early_exit
--- PASS: TestCheckedPredicateCountsAreRecorded (12.46s)
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de miss 31.47
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 31.47
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 31.47
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 31.47
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 26.76
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/again
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
=== CONT  TestPredicateFalseRead/shorthand
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
=== CONT  TestPredicateFalseRead/getter_capture
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter (7.44s)
    --- PASS: TestPredicateFalseRead/capture (32.93s)
    --- PASS: TestPredicateFalseRead/renarrowed (40.41s)
    --- PASS: TestPredicateFalseRead/early_exit (27.97s)
    --- PASS: TestPredicateFalseRead/no_read (40.46s)
    --- PASS: TestPredicateFalseRead/false_read (1.30s)
    --- PASS: TestPredicateFalseRead/again (1.30s)
    --- PASS: TestPredicateFalseRead/logical_assignment (1.41s)
    --- PASS: TestPredicateFalseRead/else_read (1.70s)
    --- PASS: TestPredicateFalseRead/spread (1.51s)
    --- PASS: TestPredicateFalseRead/getter_capture (1.46s)
    --- PASS: TestPredicateFalseRead/shorthand (1.70s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.03
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.04
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.03
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 25ac399920de hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (69.05s)
    --- PASS: TestCheckedPredicateOracle/true (41.06s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.98s)
    --- PASS: TestCheckedPredicateOracle/false (1.12s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.88s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.45s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.90s)
    --- PASS: TestCheckedPredicateOracle/valid (1.13s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (1.03s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.46s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (2.00s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.05s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.30s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.82s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.39s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.30s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (2.14s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.66s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (2.32s)
    --- PASS: TestCheckedPredicateOracle/false_comma (2.05s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=72
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	69.175s
```

</details>

```text
mutant 3 restored: clean
```

### Mutant 4 — The shorthand read dropped

```text
old text count = 1: /workspace/adamic/internal/lower/predicates_proof.go 'if parent == nil || parent.Kind != ast.KindShorthandPropertyAssignment || argument.Kind != ast.KindIdentifier {'
applied mutant 4
```

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=0 wall_seconds=49 |
| 2b | exit=1 wall_seconds=101 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** none.

<details><summary>Complete output: mutant-4-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestPredicateOverloadWithoutPredicate
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== CONT  TestIndirectPredicateOverloadIsPending
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadWithoutPredicate (0.28s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestIndirectPredicateOverloadIsPending (0.40s)
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadCallbackFromParameter (0.20s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestPredicateOverloadStaysOutOfA (0.33s)
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.15s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.94s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.20s)
=== CONT  TestPredicateOverloadNamedCallback
--- PASS: TestPredicateOverloadNamedCallback (0.39s)
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/===_false
--- PASS: TestPredicateOverloadInPlaceCallback (0.78s)
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
--- PASS: TestPredicateOverloadCallbackPushes (0.52s)
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (2.44s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (2.47s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (2.66s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (2.72s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (1.32s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.29s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.66s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.70s)
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.10s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.11s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.79s)
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateThroughModules/import_as
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
--- PASS: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.16s)
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.06s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.07s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.06s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.21s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.19s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.21s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.23s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.10s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.11s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.15s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.38s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.09s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.06s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.13s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.09s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (18.33s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (4.73s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.06s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.40s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.98s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (1.04s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.93s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.63s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.61s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.93s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.95s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.65s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	18.411s
```

</details>

**2b red tests/subtests:** `TestPredicateFalseRead/shorthand`.

<details><summary>Complete output: mutant-4-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== CONT  TestCheckedPredicateCountsAreRecorded
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/shorthand
    predicate_false_read_test.go:85: some's false direction at 23:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/shorthand.ts:23:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/again
--- PASS: TestCheckedPredicateCountsAreRecorded (14.44s)
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 31.21
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 31.21
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e miss 31.21
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 31.21
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 26.56
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.03
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.02
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.05
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.02
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/shorthand (9.21s)
    --- PASS: TestPredicateFalseRead/spread (27.50s)
    --- PASS: TestPredicateFalseRead/again (32.75s)
    --- PASS: TestPredicateFalseRead/logical_assignment (41.99s)
    --- PASS: TestPredicateFalseRead/no_read (42.04s)
    --- PASS: TestPredicateFalseRead/getter (0.50s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.66s)
    --- PASS: TestPredicateFalseRead/capture (1.85s)
    --- PASS: TestPredicateFalseRead/getter_capture (2.08s)
    --- PASS: TestPredicateFalseRead/early_exit (2.05s)
    --- PASS: TestPredicateFalseRead/else_read (1.31s)
    --- PASS: TestPredicateFalseRead/false_read (1.19s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.05
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.04
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 7a8c09103e3e hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (65.62s)
    --- PASS: TestCheckedPredicateOracle/true (43.05s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.90s)
    --- PASS: TestCheckedPredicateOracle/false (0.92s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.17s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.90s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.18s)
    --- PASS: TestCheckedPredicateOracle/valid (0.92s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (1.08s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.87s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.45s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.06s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.38s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.85s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.45s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.33s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.37s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.30s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.11s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.31s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=71
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	65.730s
FAIL
```

</details>

```text
mutant 4 restored: clean
```

### Mutant 5 — Holders read

```text
old text count = 1: /workspace/adamic/internal/lower/predicates_proof.go 'return predicateReadPosition(node) && l.predicateSameReference(node, argument)'
applied mutant 5
```

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=1 wall_seconds=55 |
| 2b | exit=1 wall_seconds=116 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** `TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference`, `TestCheckedPredicateFalseReadDirections/replaced_before_the_read`, `TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference`, `TestPredicateOverloadRuntime/overload_some_empty`, `TestPredicateOverloadRuntime/overload_every`.

<details><summary>Complete output: mutant-5-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestPredicateOverloadWithoutPredicate
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== CONT  TestPredicateOverloadReturnsAnotherRead
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestPredicateOverloadNamedCallback
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.21s)
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
--- PASS: TestPredicateOverloadWithoutPredicate (0.22s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.18s)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestPredicateOverloadCallbackPushes (0.54s)
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadNamedCallback (0.70s)
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestPredicateOverloadInPlaceCallback (0.38s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.58s)
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestIndirectPredicateOverloadIsPending (0.11s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
--- PASS: TestPredicateOverloadStaysOutOfA (0.11s)
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
--- PASS: TestPredicateOverloadCallbackFromParameter (0.10s)
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestKindPredicateThroughModules/import_as
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.11s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.10s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.12s)
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (3.25s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (3.37s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (3.14s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (3.51s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (1.36s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.23s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.49s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.83s)
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.11s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.14s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.85s)
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== NAME  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== NAME  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== NAME  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.21s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.18s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.15s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.09s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.14s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.12s)
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
--- PASS: TestKindPredicateWalkRefusals (0.01s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.33s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.43s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.51s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.61s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.30s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.37s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.06s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.11s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== NAME  TestPredicateOverloadRuntime/overload_some_empty
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_empty2462090962/001/main.ts:7:20 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:number[] -> readonly number[]} {Direction:false Status:checked Reason:number[] -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.06s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.13s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.07s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.12s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.08s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (2.44s)
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
    predicates_overload_test.go:76: /tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_every782016158/001/main.ts:18:16: stage 0 can't lower checked predicate membership over an open tag domain yet
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- FAIL: TestPredicateOverloadRuntime (22.18s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (4.44s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.27s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.72s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_empty (0.12s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (1.28s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (1.15s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (2.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (1.02s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (1.02s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (1.47s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (1.49s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.97s)
    --- FAIL: TestPredicateOverloadRuntime/overload_every (0.07s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.79s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (1.04s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	22.792s
FAIL
```

</details>

**2b red tests/subtests:** `TestPredicateFalseRead/no_read`.

<details><summary>Complete output: mutant-5-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== CONT  TestCheckedPredicateOracle
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/logical_assignment
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/shorthand
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:85: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/again
--- PASS: TestCheckedPredicateCountsAreRecorded (13.63s)
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 miss 30.18
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 30.18
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 30.18
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 30.18
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 24.41
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/renarrowed
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.02
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.03
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.03
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.02
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter (7.28s)
    --- FAIL: TestPredicateFalseRead/no_read (7.30s)
    --- PASS: TestPredicateFalseRead/again (32.32s)
    --- PASS: TestPredicateFalseRead/logical_assignment (39.63s)
    --- PASS: TestPredicateFalseRead/shorthand (32.46s)
    --- PASS: TestPredicateFalseRead/spread (26.13s)
    --- PASS: TestPredicateFalseRead/early_exit (1.30s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.21s)
    --- PASS: TestPredicateFalseRead/getter_capture (1.59s)
    --- PASS: TestPredicateFalseRead/else_read (1.81s)
    --- PASS: TestPredicateFalseRead/false_read (1.08s)
    --- PASS: TestPredicateFalseRead/capture (1.30s)
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.02
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.02
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.03
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.02
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ef6be177d251 hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (64.59s)
    --- PASS: TestCheckedPredicateOracle/true (40.60s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.79s)
    --- PASS: TestCheckedPredicateOracle/false (1.38s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.38s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.40s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.35s)
    --- PASS: TestCheckedPredicateOracle/valid (0.98s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (1.12s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.00s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.42s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.03s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.66s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.91s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.40s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.73s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.69s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.23s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.26s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.24s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=71
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	64.715s
FAIL
```

</details>

```text
mutant 5 restored: clean
```

### Mutant 6 — The name after a dot read

```text
old text count = 1: /workspace/adamic/internal/lower/predicates_proof.go 'return ast.IsExpressionNode(node) && !ast.IsRightSideOfPropertyAccess(node)'
applied mutant 6
```

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=1 wall_seconds=58 |
| 2b | exit=0 wall_seconds=98 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** `TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference`.

<details><summary>Complete output: mutant-6-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestIndirectPredicateOverloadIsPending
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
=== CONT  TestPredicateOverloadStaysOutOfA
=== CONT  TestPredicateOverloadWithoutPredicate
--- PASS: TestIndirectPredicateOverloadIsPending (0.62s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
--- PASS: TestPredicateOverloadWithoutPredicate (0.62s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestPredicateOverloadStaysOutOfA (0.64s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.15s)
=== CONT  TestPredicateOverloadNamedCallback
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.20s)
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (1.14s)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
--- PASS: TestPredicateOverloadCallbackPushes (0.50s)
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadNamedCallback (0.55s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
--- PASS: TestPredicateOverloadInPlaceCallback (0.56s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestPredicateOverloadCallbackFromParameter
=== CONT  TestKindPredicateCallIsItsComparison/negated
--- PASS: TestPredicateOverloadCallbackFromParameter (0.14s)
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestKindPredicateThroughModules/re-export_as
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (2.21s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (3.45s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (4.60s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (3.40s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.93s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (1.73s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.99s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (2.08s)
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateThroughModules/import_as
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.15s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.18s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.15s)
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.17s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.14s)
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.17s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.14s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.12s)
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== NAME  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.05s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.12s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.66s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.21s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.23s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.26s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.14s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.11s)
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.12s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.07s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.70s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (20.90s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (4.24s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.13s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.21s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (1.22s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.71s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.96s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.95s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.92s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (1.30s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (1.34s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (1.47s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (1.47s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	22.262s
FAIL
```

</details>

**2b red tests/subtests:** none.

<details><summary>Complete output: mutant-6-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== CONT  TestPredicateFalseRead
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/logical_assignment
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/shorthand
--- PASS: TestCheckedPredicateCountsAreRecorded (11.57s)
=== CONT  TestPredicateFalseRead/again
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 28.59
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 28.59
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 miss 28.58
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 28.59
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 23.80
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.03
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.04
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/shorthand (36.49s)
    --- PASS: TestPredicateFalseRead/no_read (36.49s)
    --- PASS: TestPredicateFalseRead/logical_assignment (36.64s)
    --- PASS: TestPredicateFalseRead/again (25.08s)
    --- PASS: TestPredicateFalseRead/getter (0.67s)
    --- PASS: TestPredicateFalseRead/spread (2.14s)
    --- PASS: TestPredicateFalseRead/renarrowed (2.17s)
    --- PASS: TestPredicateFalseRead/getter_capture (2.48s)
    --- PASS: TestPredicateFalseRead/capture (1.91s)
    --- PASS: TestPredicateFalseRead/else_read (1.07s)
    --- PASS: TestPredicateFalseRead/early_exit (1.31s)
    --- PASS: TestPredicateFalseRead/false_read (1.03s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.07
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.04
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.03
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.03
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.02
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f3e695527b39 hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (62.75s)
    --- PASS: TestCheckedPredicateOracle/true (37.84s)
    --- PASS: TestCheckedPredicateOracle/true_unused (2.03s)
    --- PASS: TestCheckedPredicateOracle/false (0.87s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.28s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.85s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.33s)
    --- PASS: TestCheckedPredicateOracle/valid (1.06s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (1.11s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.17s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.27s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.21s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (2.00s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.89s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.44s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.57s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.89s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.25s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.36s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=72
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	62.872s
```

</details>

```text
mutant 6 restored: clean
```

### Mutant 7 — Labels read

```text
old text count = 1: /workspace/adamic/internal/lower/predicates_proof.go 'return ast.IsExpressionNode(node) && !ast.IsRightSideOfPropertyAccess(node)'
applied mutant 7
```

| Rerun | Exit and wall time |
|---|---|
| 2a | exit=1 wall_seconds=57 |
| 2b | exit=0 wall_seconds=107 |

Every red leaf test/subtest below is **not red at HEAD** (it passed in step 2). Parent FAIL lines are propagated suite failures; complete output below includes them.

**2a red tests/subtests:** `TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference`.

<details><summary>Complete output: mutant-7-2a</summary>

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestPredicateOverloadWithoutPredicate
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== CONT  TestPredicateOverloadNamedCallback
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.35s)
--- PASS: TestPredicateOverloadWithoutPredicate (0.35s)
=== CONT  TestPredicateOverloadStaysOutOfA
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadStaysOutOfA (0.17s)
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadCallbackFromParameter (0.17s)
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestIndirectPredicateOverloadIsPending (0.15s)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestPredicateOverloadNamedCallback (0.80s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.85s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.12s)
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
--- PASS: TestPredicateOverloadInPlaceCallback (0.44s)
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
--- PASS: TestPredicateOverloadCallbackPushes (0.54s)
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestKindPredicateThroughModules/import_as
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (2.89s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (2.58s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (2.67s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (2.84s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (1.60s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.44s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.51s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.69s)
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/overloaded
--- PASS: TestKindPredicateThroughModules (0.02s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.15s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.12s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.06s)
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.17s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.24s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.79s)
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.17s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.19s)
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== NAME  TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference
    predicates_false_read_test.go:93: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.10s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/a_label_named_as_the_reference (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.12s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.30s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.07s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.11s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.22s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.21s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.10s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (18.12s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (3.51s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.42s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.51s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (1.02s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.94s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.90s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.67s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.73s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (1.01s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.95s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.75s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.75s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	18.705s
FAIL
```

</details>

**2b red tests/subtests:** none.

<details><summary>Complete output: mutant-7-2b</summary>

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/shorthand
--- PASS: TestCheckedPredicateCountsAreRecorded (13.14s)
=== CONT  TestPredicateFalseRead/again
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 miss 31.00
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 31.00
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 26.05
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 31.00
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 31.00
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.08
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.03
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.03
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.05
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.04
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.04
=== NAME  TestCheckedPredicateOracle/true_unused
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.03
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/logical_assignment (40.26s)
    --- PASS: TestPredicateFalseRead/shorthand (40.26s)
    --- PASS: TestPredicateFalseRead/again (27.14s)
    --- PASS: TestPredicateFalseRead/no_read (40.28s)
    --- PASS: TestPredicateFalseRead/getter (0.50s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.96s)
    --- PASS: TestPredicateFalseRead/spread (1.99s)
    --- PASS: TestPredicateFalseRead/getter_capture (2.17s)
    --- PASS: TestPredicateFalseRead/capture (2.09s)
    --- PASS: TestPredicateFalseRead/false_read (1.54s)
    --- PASS: TestPredicateFalseRead/early_exit (2.00s)
    --- PASS: TestPredicateFalseRead/else_read (2.17s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.03
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.03
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.03
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.05
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types a1c6b847be80 hit 0.02
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (68.35s)
    --- PASS: TestCheckedPredicateOracle/true (41.39s)
    --- PASS: TestCheckedPredicateOracle/true_unused (2.44s)
    --- PASS: TestCheckedPredicateOracle/false (1.38s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.14s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.49s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.70s)
    --- PASS: TestCheckedPredicateOracle/valid (1.37s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.98s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.21s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.24s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.33s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.85s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.01s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (2.16s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.63s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.60s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.32s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.37s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.73s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=72
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	68.502s
```

</details>

```text
mutant 7 restored: clean
```

The table’s only explicit named Red cell, mutant 4’s `TestPredicateFalseRead/shorthand`, failed as stated. The other six cells say “measured on the head’s proof”; the measured red sets are enumerated above. Source-comment detectors also red: mutant 1 unread rows; mutant 2 capture of a let assigned later; mutant 3 merged read; mutant 5 replaced-before-read and no_read; mutant 6 dot-name; mutant 7 label.

## Step 4 — full lower suite

`go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1`

exit=0 wall_seconds=1911

```text
ok  	github.com/system-inc/adamic/internal/lower	1904.994s
```

No failing tests; no environment failures. Full lower wall time: 31m 51s.

## Step 4b — IR source and field guards

`go test ./internal/ir -count=1 -timeout 30m`

exit=0 wall_seconds=158

<details><summary>Complete output: ir</summary>

```text
ok  	github.com/system-inc/adamic/internal/ir	122.435s
```

</details>

No IR failures (wall time: 2m 38s), including `TestCallTargetReadersLower` and the IR field guards.

## Step 5 — admission delta

Built `/tmp/adamic-proof-39w/admission-delta` from HEAD, generated the complete manifest from HEAD, then ran alone:

```sh
/tmp/adamic-proof-39w/admission-delta --base 0c7052759671347775615af5b4a8028dd8688af8 --head 89629dc4400e56669730e3ad048e114639c56ff9 --workers 2 --manifest /tmp/adamic-proof-39w/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```

exit=0 wall_seconds=1521


Verdict **pass**; `void=false`; void reasons: none (no report-level or program-level void reason). `diff_count=13`; `admitted=0`; newly refused=0. All 5,314 unique classified programs: 3,330 accepted by both; 1,984 refused by both. Every program whose classification stderr differs: **none** (all 5,314 base/head stderr strings compared exactly, including refused-by-both programs). Thus there are no differing stderr text pairs to list.

```json
{
  "negative_witness_count": 50,
  "accepted_witnesses": 50,
  "node_agreements": 13,
  "runtime_notyet_count": 52,
  "ruled_runtime_stops": 52,
  "stale_runtime_stops": [],
  "pending_runtime_stops": [],
  "backend_divergence_count": 12,
  "ruled_backend_divergences": 0,
  "ruled_backend_divergence_paths": [],
  "corpus_filter": "",
  "shard": "",
  "complete_corpus": true,
  "phase_seconds": {
    "base_build": 85.225282037,
    "checkout": 25.314342245,
    "checkout_and_provision": 27.08059187,
    "classification": 775.625376274,
    "head_build": 45.336728276,
    "manifest_and_verification": 29.328964691,
    "provision": 1.766048565,
    "runtime": 557.405500403,
    "total": 1520.015584399
  },
  "workers": 2,
  "classification": "lowering-only",
  "diff": [
    {
      "path": "internal/lower/testdata/predicates_checked/false_read.a",
      "blob": "a9df503eff15df40a0f3ca4ef5e7a0704f534b22"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/again.ts",
      "blob": "9f6f8070d0b6347b7ceeb08c9159b78b160e172f"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/capture.ts",
      "blob": "1689498e103b3729216dbeaeaf3e1a8c6f635bda"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/early_exit.ts",
      "blob": "c89883d7b0206eeb6678fce44716a1d5844459ff"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/else_read.ts",
      "blob": "16aad0696c46b00f5029e0bc9d3fea92bda804b9"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/false_read.ts",
      "blob": "5954ab19e76cecd29dd4619c61c2774ba530bbfa"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/getter.ts",
      "blob": "6e488b9c05361e22d3f0328ac98b45513b2953dc"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/getter_capture.ts",
      "blob": "540af883f64cd9e7a975524748489ae53a320518"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/logical_assignment.ts",
      "blob": "9f3a5719548789c4c9853c7a6f6f52ac5dbafdce"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/no_read.ts",
      "blob": "81a574f9ee239a4df14fa3bfc3eda7c596d73b78"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/renarrowed.ts",
      "blob": "abedbe6418d601b79db8200bb51e4b2de8928ff7"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/shorthand.ts",
      "blob": "fbf863f2065f06abf7715587acf756952b1bcbe0"
    },
    {
      "path": "internal/oracle/testdata/predicate_false_read/spread.ts",
      "blob": "a8bcb7a42d2ac99d056faa0fe578c491a3ce6c7b"
    }
  ],
  "diff_count": 13,
  "diff_ts_unmatched": [],
  "compile_timeout_seconds": 45,
  "runtime_timeout_seconds": 10,
  "base": "0c7052759671347775615af5b4a8028dd8688af8",
  "head": "89629dc4400e56669730e3ad048e114639c56ff9",
  "generator_revision": "89629dc4400e56669730e3ad048e114639c56ff9",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "c983750584e46e90be44e1e2e62499406c426f92",
  "admitted": 0,
  "verdict": "pass",
  "void": false,
  "sampling_seed": "417372f6ee1ac53c9e536e2fec94650e5d2a4d2eb5209f92c151fadc7a79490e",
  "seed_given": false,
  "sampling_size": 115,
  "omitted": 0,
  "budget_seconds": 0,
  "budget_used_seconds": 557.405500403,
  "timed_out_runs": 1,
  "retried_runs": 1,
  "still_timed_out_runs": 0
}
```

Gate stderr (entire):

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 1 runs timed out, 1 retried alone, 0 timed out again
changed programs: 13
```

### Timeout and isolated reruns at both revisions

The sole timeout was `native compile` for `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts`. The gate retried alone and succeeded: 1 timeout, 1 isolated retry, 0 still timed out. Full gate entry:

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
  "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.067859323,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.105156579,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.135398823,
    "exit": 0,
    "stdout": "other\n4 false\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.07422074,
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "native": {
    "wall_seconds": 0.002691676,
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "javascript_compile": {
    "wall_seconds": 0.105512519,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 2.63260524,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "agree": false,
  "negative_related": true,
  "negative_witness": true,
  "notyet_related": false,
  "ruled_runtime_stop": false,
  "ruled_backend_divergence": false,
  "repair_agree": true,
  "sampled": true,
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```

Separately reran that exact native compile alone with the gate’s cached compiler for each SHA, using the original 10-second limit. Checked out main detached and updated submodules before the main run; returned to HEAD detached and updated submodules before the HEAD run. Both source revisions hold the same fixture blob. The two reruns did not overlap. Complete command/outcome records:

```json
{
  "sha": "0c7052759671347775615af5b4a8028dd8688af8",
  "command": [
    "/home/agent/.cache/adamic-build/local/d8059468d288f99d5970aa6966970b06675dee42beac78afe9c906b1a3878b92/adamic",
    "build",
    "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "-o",
    "/tmp/adamic-proof-39w/timeout-main-program"
  ],
  "exit": 0,
  "stdout": "",
  "stderr": "",
  "wall_seconds": 3.5645867760003966,
  "timed_out": false
}
```

```json
{
  "sha": "89629dc4400e56669730e3ad048e114639c56ff9",
  "command": [
    "/home/agent/.cache/adamic-build/local/48f981120279e8c848fb71552a862e6ff631d93baed088261d3e00ffa2cbc7d1/adamic",
    "build",
    "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "-o",
    "/tmp/adamic-proof-39w/timeout-head-program"
  ],
  "exit": 0,
  "stdout": "",
  "stderr": "",
  "wall_seconds": 2.8358922100005657,
  "timed_out": false
}
```

The initial timeout is an environmental timing event, not a persistent compiler failure: it passed the gate’s isolated retry and the separate native compile at each revision.

## Step 6 — conditional full oracle suite

**Not run.** Step 5 had zero newly admitted programs, zero newly refused programs, and zero classification stderr differences anywhere, including outside the two allowed fixture locations. Therefore the stated condition for the full oracle suite was false. The selected uncached oracle pins did run and passed at clean HEAD (step 2b).

## Restoration and report delivery

After all mutants and both timeout reruns, HEAD was restored to `89629dc4400e56669730e3ad048e114639c56ff9`; submodules were updated; `git diff --exit-code` passed. The only repository file created for delivery is `proof/lower-39w-6-2/report.md`. The report branch is created directly from that checked-out HEAD. Only that file is staged and committed; only `proof/lower-39w-6-2-report` is pushed to origin.

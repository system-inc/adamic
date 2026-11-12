# lower-39w-4 proof report


Completed 2026-11-12T04:16:36+00:00.


## Result


The proof is not green. Formatting and vet passed; targeted lower tests, the targeted oracle suite, the full lower suite, and Explain tests failed. The IR suite passed. Mutant 4 did **not** make `TestPredicateFalseRead/shorthand` red. Mutants 1 and 3 produced only the baseline `no_read` red, so their nonzero exits do not establish an additional detection. The full admission delta was **void** because a native compilation timed out twice; subsequent isolated reruns compiled successfully at both SHAs under the same 10-second limit. The original void verdict is retained.


## Revisions and setup


- Repository: `https://github.com/system-inc/adamic.git`.
- Tested HEAD: `63eb56cd701c744a603f3c694295d749ba9d46dc`, checked out detached after `git fetch origin compiler/predicate-false-read-4`.
- Base/main: `730e7c26e590af0c57c545ba0cb905b09e4bee54`; HEAD is one commit above this base.
- Ran `git submodule update --init --recursive`: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`, nested `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- No AGENTS.md files were found under `/workspace`.
- Loaded `/workspace/adamic-tools/env.sh` before repository work and for all proof commands. It sets `TMPDIR=/tmp/adamic-gate`; created that directory outside the repository before invoking Go.
- Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/tmp/adamic-proof/typescript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`; exported `ADAMIC_TYPESCRIPT_SOURCE` accordingly.
- Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/adamic-proof/api` and ran `npm ci --ignore-scripts --prefix /tmp/adamic-proof/api`. Packages are outside the repository; PATH begins with `/tmp/adamic-proof/api/node_modules/.bin`. `tsc --version`: `Version 6.0.3`. Pinned direct packages: TypeScript 6.0.3 and @types/node 25.3.3.
- Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. TLS verification stayed enabled.
- Toolchain: Go 1.27.1 linux/amd64; Node v24.19.0; Clang 20.1.8.
- One heavy top-level command ran at a time. No test source was fixed. Temporary mutation edits were restored after each run; all four `git diff --exit-code` checks returned 0. The tracked tree was clean again after all proof commands.
- All raw execution evidence was kept under `/tmp/adamic-proof`, with the requested full lower output at `/tmp/lower.log`. The requested results are reproduced in this report.


## 1. Formatting and vet


```sh
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```


Both exited 0. `gofmt -l` printed nothing. Vet printed nothing.


## 2. Targeted lower proof


```sh
go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m
```


Exit **1**; package test time 29.488s. Three failing leaf subtests:

- `TestCheckedPredicateFalseReadDirections/replaced_before_the_read`: false side checked true, want false.
- `TestPredicateOverloadRuntime/overload_some_empty`: counts proven 0, checked 2, unobservable 0; wanted 0, 1, 1.
- `TestPredicateOverloadRuntime/overload_every`: checked predicate membership over an open tag domain is not lowered.

All other selected leaf subtests passed. The complete verbose output below records every parent/subtest PASS or FAIL and every failure’s whole message, including temporary paths.


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
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestIndirectPredicateOverloadIsPending
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
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
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadWithoutPredicate (0.17s)
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestIndirectPredicateOverloadIsPending (0.18s)
=== CONT  TestPredicateOverloadNamedCallback
--- PASS: TestPredicateOverloadStaysOutOfA (0.13s)
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.08s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.13s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadNamedCallback (0.47s)
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.74s)
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
--- PASS: TestPredicateOverloadCallbackFromParameter (0.10s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
--- PASS: TestPredicateOverloadCallbackPushes (0.46s)
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
--- PASS: TestPredicateOverloadInPlaceCallback (0.38s)
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (17.51s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (17.39s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (17.45s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (17.46s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.13s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (1.27s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.27s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.27s)
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== NAME  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestKindPredicateThroughModules/import_as
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.16s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.21s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.12s)
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/overloaded
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.08s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.13s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.07s)
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
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.17s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.11s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.74s)
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.18s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.18s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.07s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.10s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== NAME  TestPredicateOverloadRuntime/overload_some_empty
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_empty3619706650/001/main.ts:7:20 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:number[] -> readonly number[]} {Direction:false Status:checked Reason:number[] -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.02s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.16s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.15s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.14s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.09s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.10s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (0.70s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
    predicates_overload_test.go:76: /tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_every4268608108/001/main.ts:18:16: stage 0 can't lower checked predicate membership over an open tag domain yet
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- FAIL: TestPredicateOverloadRuntime (29.43s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (18.87s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.04s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.40s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_empty (0.10s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.66s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.63s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.70s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.57s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.60s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.62s)
    --- FAIL: TestPredicateOverloadRuntime/overload_every (0.05s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.64s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.60s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	29.488s
FAIL
```


## 3. Uncached targeted oracle proof


```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 60m
```


Exit **1**; package test time 290.980s. Only failing leaf subtest: `TestPredicateFalseRead/no_read`, whose `some` false direction is checked rather than unobservable. All other 11 false-read subtests, all 19 checked-predicate oracle subtests, `TestCheckedPredicateCountsAreRecorded`, and `TestCountsAreRecordedTable` passed. Complete per-subtest output and messages:


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
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/renarrowed
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:84: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/capture
--- PASS: TestCheckedPredicateCountsAreRecorded (43.24s)
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 230.81
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 230.82
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 230.82
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 226.16
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 miss 230.83
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== CONT  TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/shorthand
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.10
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.02
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.02
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter (38.10s)
    --- FAIL: TestPredicateFalseRead/no_read (38.12s)
    --- PASS: TestPredicateFalseRead/renarrowed (232.58s)
    --- PASS: TestPredicateFalseRead/capture (232.56s)
    --- PASS: TestPredicateFalseRead/logical_assignment (270.67s)
    --- PASS: TestPredicateFalseRead/early_exit (227.47s)
    --- PASS: TestPredicateFalseRead/false_read (1.90s)
    --- PASS: TestPredicateFalseRead/shorthand (1.87s)
    --- PASS: TestPredicateFalseRead/again (1.92s)
    --- PASS: TestPredicateFalseRead/else_read (1.94s)
    --- PASS: TestPredicateFalseRead/spread (0.78s)
    --- PASS: TestPredicateFalseRead/getter_capture (0.94s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.02
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 375455d5a8e5 hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (290.86s)
    --- PASS: TestCheckedPredicateOracle/true (271.97s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.39s)
    --- PASS: TestCheckedPredicateOracle/false (0.95s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.74s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.88s)
    --- PASS: TestCheckedPredicateOracle/asserts (0.98s)
    --- PASS: TestCheckedPredicateOracle/valid (0.87s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (1.10s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (0.81s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.38s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.87s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.79s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.21s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.01s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.29s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.11s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.24s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.09s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=71
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	290.980s
FAIL
```


## 3b. Mutants


Each mutant used the unmodified HEAD as its starting point. Before applying it, the exact old text below was asserted to occur **exactly once**. Mutants ran sequentially with the step 3 command restricted to `TestPredicateFalseRead`:


```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead' -count=1 -v -timeout 60m
```


| Mutant | Exit | Red leaf subtests | Additional red versus HEAD | Restored diff exit |
|---|---:|---|---|---:|
| 1. Condition-only arm restored | 1 | `no_read` | None | 0 |
| 2. Types decide alone | 1 | `logical_assignment`, `getter_capture` | Both; baseline `no_read` passes here | 0 |
| 3. A merge counts | 1 | `no_read` | None | 0 |
| 4. Shorthand read dropped | 1 | `no_read` | None; **shorthand passed** | 0 |


**Required fourth-mutant result: FAIL.** `TestPredicateFalseRead/shorthand` stayed green despite the documented edit. A nonzero mutant exit caused solely by the existing `no_read` failure is not counted as evidence that the mutant was caught.


### Mutant 1 edit and complete result


File: `internal/lower/predicates_checked.go`. Asserted old-text occurrence count: 1.

Old text:


```go
	return directions
}

// predicateFalseRead
```


Replacement text (empty block means deletion):


```go
	conditionNarrows := func(condition *ast.Node, truth bool) bool {
		if condition == nil || l.predicateConditionDirection(condition, call.AsNode(), truth, 0)&predicateFalse == 0 {
			return false
		}
		flags := ast.FlowFlagsFalseCondition
		if truth {
			flags = ast.FlowFlagsTrueCondition
		}
		branch := &ast.FlowNode{Flags: flags, Node: condition, Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
		narrowed := predicateFlowType(l.checker, argument, source, source, nil, branch)
		return !checker.Checker_isTypeIdenticalTo(l.checker, source, narrowed)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var condition *ast.Node
		switch node.Kind {
		case ast.KindIfStatement:
			condition = node.AsIfStatement().Expression
		case ast.KindConditionalExpression:
			condition = node.AsConditionalExpression().Condition
		case ast.KindWhileStatement:
			condition = node.AsWhileStatement().Expression
		case ast.KindDoStatement:
			condition = node.AsDoStatement().Expression
		case ast.KindForStatement:
			condition = node.AsForStatement().Condition
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindBarBarToken && conditionNarrows(binary.Left, false) {
				directions |= predicateFalse
			}
		case ast.KindSwitchStatement:
			statement := node.AsSwitchStatement()
			if ast.SkipParentheses(statement.Expression).Kind == ast.KindTrueKeyword {
				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
					if clause.Kind == ast.KindCaseClause && conditionNarrows(clause.Expression(), false) {
						directions |= predicateFalse
					}
				}
			}
		}
		if conditionNarrows(condition, true) || conditionNarrows(condition, false) {
			directions |= predicateFalse
		}
		node.ForEachChild(visit)
		return false
	}
	ast.GetSourceFileOfNode(call.AsNode()).AsNode().ForEachChild(visit)
	return directions
}

// predicateFalseRead
```


After the test, restored the edited file with `git restore -- <file>` and ran `git diff --exit-code`; exit 0, no diff. Full test log:


```text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
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
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/shorthand
=== CONT  TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:84: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/getter_capture
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec miss 30.06
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 30.08
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 30.08
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 30.08
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 30.08
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 0.03
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 0.01
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types e921885397ec hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter (7.81s)
    --- FAIL: TestPredicateFalseRead/no_read (7.82s)
    --- PASS: TestPredicateFalseRead/shorthand (39.24s)
    --- PASS: TestPredicateFalseRead/getter_capture (31.45s)
    --- PASS: TestPredicateFalseRead/spread (31.49s)
    --- PASS: TestPredicateFalseRead/logical_assignment (39.31s)
    --- PASS: TestPredicateFalseRead/again (39.32s)
    --- PASS: TestPredicateFalseRead/early_exit (1.38s)
    --- PASS: TestPredicateFalseRead/capture (1.34s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.39s)
    --- PASS: TestPredicateFalseRead/else_read (1.45s)
    --- PASS: TestPredicateFalseRead/false_read (1.45s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=22
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	40.812s
FAIL
```


### Mutant 2 edit and complete result


File: `internal/lower/predicates_checked.go`. Asserted old-text occurrence count: 1.

Old text:


```go
 || l.predicateFalseRead(call, argument)
```


Replacement text (empty block means deletion):


```go

```


After the test, restored the edited file with `git restore -- <file>` and ran `git diff --exit-code`; exit 0, no diff. Full test log:


```text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
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
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/shorthand
=== CONT  TestPredicateFalseRead/again
=== NAME  TestPredicateFalseRead/logical_assignment
    predicate_false_read_test.go:84: some's false direction at 23:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/logical_assignment.ts:23:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
    predicate_false_read_test.go:84: some's false direction at 38:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/getter_capture.ts:38:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 miss 25.78
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 25.78
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 25.78
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 25.29
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 25.78
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/false_read
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 0.01
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 625071ae87f5 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/logical_assignment (6.34s)
    --- PASS: TestPredicateFalseRead/getter (6.38s)
    --- FAIL: TestPredicateFalseRead/getter_capture (0.45s)
    --- PASS: TestPredicateFalseRead/again (33.50s)
    --- PASS: TestPredicateFalseRead/shorthand (33.50s)
    --- PASS: TestPredicateFalseRead/else_read (26.69s)
    --- PASS: TestPredicateFalseRead/spread (27.20s)
    --- PASS: TestPredicateFalseRead/no_read (33.54s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.18s)
    --- PASS: TestPredicateFalseRead/false_read (1.17s)
    --- PASS: TestPredicateFalseRead/capture (1.22s)
    --- PASS: TestPredicateFalseRead/early_exit (1.19s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=21
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	34.750s
FAIL
```


### Mutant 3 edit and complete result


File: `internal/lower/predicates_checked.go`. Asserted old-text occurrence count: 1.

Old text:


```go
return falseMet && (!trueMet || reduced)
```


Replacement text (empty block means deletion):


```go
return falseMet
```


After the test, restored the edited file with `git restore -- <file>` and ran `git diff --exit-code`; exit 0, no diff. Full test log:


```text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
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
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/again
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:84: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 26.72
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 26.72
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 26.72
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 miss 26.72
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 26.72
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== CONT  TestPredicateFalseRead/shorthand
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 0.03
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 0.02
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 0.02
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 5e191b011030 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/no_read (8.77s)
    --- PASS: TestPredicateFalseRead/getter (8.81s)
    --- PASS: TestPredicateFalseRead/renarrowed (28.83s)
    --- PASS: TestPredicateFalseRead/early_exit (28.87s)
    --- PASS: TestPredicateFalseRead/spread (37.64s)
    --- PASS: TestPredicateFalseRead/getter_capture (37.65s)
    --- PASS: TestPredicateFalseRead/again (37.65s)
    --- PASS: TestPredicateFalseRead/false_read (1.57s)
    --- PASS: TestPredicateFalseRead/logical_assignment (1.59s)
    --- PASS: TestPredicateFalseRead/else_read (1.59s)
    --- PASS: TestPredicateFalseRead/capture (1.62s)
    --- PASS: TestPredicateFalseRead/shorthand (1.60s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=22
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	39.284s
FAIL
```


### Mutant 4 edit and complete result


File: `internal/lower/predicates_proof.go`. Asserted old-text occurrence count: 1.

Old text:


```go
if parent == nil || parent.Kind != ast.KindShorthandPropertyAssignment || argument.Kind != ast.KindIdentifier {
```


Replacement text (empty block means deletion):


```go
if true {
```


After the test, restored the edited file with `git restore -- <file>` and ran `git diff --exit-code`; exit 0, no diff. Full test log:


```text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
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
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/shorthand
=== CONT  TestPredicateFalseRead/again
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:84: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 27.54
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 27.54
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 27.54
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 27.54
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 miss 27.55
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 0.02
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 0.07
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 0.09
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 0.02
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 22d5d4246141 hit 0.03
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/no_read (6.35s)
    --- PASS: TestPredicateFalseRead/getter (6.37s)
    --- PASS: TestPredicateFalseRead/again (35.42s)
    --- PASS: TestPredicateFalseRead/logical_assignment (35.42s)
    --- PASS: TestPredicateFalseRead/spread (29.09s)
    --- PASS: TestPredicateFalseRead/getter_capture (29.08s)
    --- PASS: TestPredicateFalseRead/shorthand (35.45s)
    --- PASS: TestPredicateFalseRead/early_exit (1.76s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.77s)
    --- PASS: TestPredicateFalseRead/capture (1.96s)
    --- PASS: TestPredicateFalseRead/false_read (2.14s)
    --- PASS: TestPredicateFalseRead/else_read (2.74s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=22
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	38.210s
FAIL
```


## 4. Full IR and lower suites


```sh
go test ./internal/ir -count=1 -timeout 30m
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```


IR exit **0**, 127.200s:


```text
ok  	github.com/system-inc/adamic/internal/ir	127.200s
```


Lower exit **1**, 1550.710s. Every failing leaf test and its first error line follows. Parent failure rows are retained in the complete log below. **Environment failures: none** among these failures. The refusal/count/direction messages are compiler behavior or expectation failures.


| Failing test | First error line | Environment-related |
|---|---|---|
| `TestBrandedCallRuleTwoIdentity` | `predicates_branded_rules_test.go:87: Lower refused acceptance row: /tmp/adamic-gate/TestBrandedCallRuleTwoIdentity3986115382/005/main.ts:20:84: stage 0 can't lower checked predicate membership of a branded supertype on its false branch yet` | No |
| `TestPredicateOverloadRuntime/overload_some_empty` | `predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 ...}, want [0 1 1] (proven, checked, unobservable)` (full first line below) | No |
| `TestPredicateOverloadRuntime/overload_every` | `predicates_overload_test.go:76: /tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_every2666857236/001/main.ts:18:16: stage 0 can't lower checked predicate membership over an open tag domain yet` | No |
| `TestCheckedPredicateFalseReadDirections/replaced_before_the_read` | `predicates_false_read_test.go:89: false side checked true, want false` | No |


```text
--- FAIL: TestBrandedCallRuleTwoIdentity (1.11s)
    predicates_branded_rules_test.go:87: Lower refused acceptance row: /tmp/adamic-gate/TestBrandedCallRuleTwoIdentity3986115382/005/main.ts:20:84: stage 0 can't lower checked predicate membership of a branded supertype on its false branch yet
--- FAIL: TestPredicateOverloadRuntime (21.44s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_empty (0.14s)
        predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_empty2397776581/001/main.ts:7:20 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:number[] -> readonly number[]} {Direction:false Status:checked Reason:number[] -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
    --- FAIL: TestPredicateOverloadRuntime/overload_every (0.14s)
        predicates_overload_test.go:76: /tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_every2666857236/001/main.ts:18:16: stage 0 can't lower checked predicate membership over an open tag domain yet
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.14s)
        predicates_false_read_test.go:89: false side checked true, want false
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1550.710s
FAIL
```


## 5. Explain tests


```sh
go test ./cmd/adamic -run 'Explain' -count=1 -timeout 30m
```


Exit **1**, 12.113s. Failing leaves: `TestExplainChecksOutput/overload_some_empty/c`, `/js`, and `/build`; their parents also fail. Each output says false checked and counts 0/2/0, expected false unobservable and 0/1/1. Complete failure output:


```text
--- FAIL: TestExplainChecksOutput (0.00s)
    --- FAIL: TestExplainChecksOutput/overload_some_empty (0.00s)
        --- FAIL: TestExplainChecksOutput/overload_some_empty/c (0.13s)
            checks_test.go:54: c explanation:
                /tmp/adamic-gate/TestExplainChecksOutputoverload_some_empty1564464627/001/main.ts:7:20: predicate overload 1 of some
                  true: checked; number[] -> readonly number[]
                  false: checked; number[] -> readonly number[]
                adamic: predicate checks: proven 0 checked 2 unobservable 0
                
                want:
                internal/lower/testdata/predicates/overload_some_empty.a:7:20: predicate overload 1 of some
                  true: checked; number[] -> readonly number[]
                  false: unobservable; number[] -> readonly number[]
                adamic: predicate checks: proven 0 checked 1 unobservable 1
        --- FAIL: TestExplainChecksOutput/overload_some_empty/js (0.11s)
            checks_test.go:54: js explanation:
                /tmp/adamic-gate/TestExplainChecksOutputoverload_some_empty1564464627/001/main.ts:7:20: predicate overload 1 of some
                  true: checked; number[] -> readonly number[]
                  false: checked; number[] -> readonly number[]
                adamic: predicate checks: proven 0 checked 2 unobservable 0
                
                want:
                internal/lower/testdata/predicates/overload_some_empty.a:7:20: predicate overload 1 of some
                  true: checked; number[] -> readonly number[]
                  false: unobservable; number[] -> readonly number[]
                adamic: predicate checks: proven 0 checked 1 unobservable 1
        --- FAIL: TestExplainChecksOutput/overload_some_empty/build (3.93s)
            checks_test.go:54: build explanation:
                /tmp/adamic-gate/TestExplainChecksOutputoverload_some_empty1564464627/001/main.ts:7:20: predicate overload 1 of some
                  true: checked; number[] -> readonly number[]
                  false: checked; number[] -> readonly number[]
                adamic: predicate checks: proven 0 checked 2 unobservable 0
                
                want:
                internal/lower/testdata/predicates/overload_some_empty.a:7:20: predicate overload 1 of some
                  true: checked; number[] -> readonly number[]
                  false: unobservable; number[] -> readonly number[]
                adamic: predicate checks: proven 0 checked 1 unobservable 1
FAIL
FAIL	github.com/system-inc/adamic/cmd/adamic	12.113s
FAIL
```


## 6. Admission delta


```sh
python3 cloud/admission-corpus/manifest.py --sha 63eb56cd701c744a603f3c694295d749ba9d46dc > /tmp/adamic-proof/manifest.json
ADAMIC_BUILD_LOG=/tmp/adamic-proof/admission-build.log go run ./cmd/adamic-admission-delta \
  --base 730e7c26e590af0c57c545ba0cb905b09e4bee54 \
  --head 63eb56cd701c744a603f3c694295d749ba9d46dc \
  --manifest /tmp/adamic-proof/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json \
  > /tmp/adamic-proof/admission.json 2> /tmp/adamic-proof/admission.stderr
```


This was the complete, unsampled corpus, with no filter or shard, budget 0, and workers 2. It ran alone. Classification timeout: 45 seconds; runtime command timeout: 10 seconds. The gate provisioned its own external head worktree.


| Metric | Result |
|---|---|
| Gate exit | **3** (`go run` wrapper exit **1**, stderr `exit status 3`) |
| Verdict | **void** |
| Void | true |
| Void reason | `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout` |
| diff_count | **13** |
| Newly admitted (`admitted`) | **0** |
| Newly refused | **1** |
| Accepted by both | 3329 |
| Refused by both | 1984 |
| Unique programs classified | 5314 |
| Node agreements | 13 |
| Accepted negative witnesses | 49 of 50 listed |
| Ruled runtime stops | 52 of 52 listed |
| Stale/pending runtime stops | 0 / 0 |
| Runtime runs timed out / retried alone / still timed out | 1 / 1 / 1 |
| Classification time | 725.350s |
| Runtime time | 555.552s |
| Total time | 1395.551s |


Manifest counts: 2 witnesses; 4366 fixtures; 105 gaps; 20 review; 0 fuzz; 821 stage3. Generator revision is tested HEAD; generator blob `09d061a03b1e016387e975d25ee0cd5b75ac6c1f`; manifest blob `049ed5b09da459778fe34ead488e425260d82c36`. Raw JSON SHA-256: `5df13ea572c29d7ee484fcd69b4d9a0b4f555115d7b3af61e720f73ef128a762`.


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


### Every classification-stderr difference


Exactly **1** program differs. Stderr was compared byte-for-byte across all 5314 classifications.


Program: `internal/lower/testdata/predicates/overload_every.ts`. Classification: `newly-refused`. Base exit 0; HEAD exit 1.

Base stderr (empty string):


```text

```


HEAD stderr (exact text, including the original gate worktree path; ends with a newline):


```text
adamic: /tmp/adamic-gate/admission-delta-1884244102/head/internal/lower/testdata/predicates/overload_every.ts:18:16: stage 0 can't lower checked predicate membership over an open tag domain yet
```


### Changed-program coverage


- `internal/lower/testdata/predicates_checked/false_read.a` (`a9df503eff15df40a0f3ca4ef5e7a0704f534b22`)

- `internal/oracle/testdata/predicate_false_read/again.ts` (`9f6f8070d0b6347b7ceeb08c9159b78b160e172f`)

- `internal/oracle/testdata/predicate_false_read/capture.ts` (`1689498e103b3729216dbeaeaf3e1a8c6f635bda`)

- `internal/oracle/testdata/predicate_false_read/early_exit.ts` (`c89883d7b0206eeb6678fce44716a1d5844459ff`)

- `internal/oracle/testdata/predicate_false_read/else_read.ts` (`16aad0696c46b00f5029e0bc9d3fea92bda804b9`)

- `internal/oracle/testdata/predicate_false_read/false_read.ts` (`5954ab19e76cecd29dd4619c61c2774ba530bbfa`)

- `internal/oracle/testdata/predicate_false_read/getter.ts` (`6e488b9c05361e22d3f0328ac98b45513b2953dc`)

- `internal/oracle/testdata/predicate_false_read/getter_capture.ts` (`540af883f64cd9e7a975524748489ae53a320518`)

- `internal/oracle/testdata/predicate_false_read/logical_assignment.ts` (`9f3a5719548789c4c9853c7a6f6f52ac5dbafdce`)

- `internal/oracle/testdata/predicate_false_read/no_read.ts` (`81a574f9ee239a4df14fa3bfc3eda7c596d73b78`)

- `internal/oracle/testdata/predicate_false_read/renarrowed.ts` (`abedbe6418d601b79db8200bb51e4b2de8928ff7`)

- `internal/oracle/testdata/predicate_false_read/shorthand.ts` (`fbf863f2065f06abf7715587acf756952b1bcbe0`)

- `internal/oracle/testdata/predicate_false_read/spread.ts` (`a8bcb7a42d2ac99d056faa0fe578c491a3ce6c7b`)


### Timeout evidence and isolated reruns at both SHAs


There were no classification timeouts. Exactly one program had a runtime command timeout: `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts`, native compile. The gate retried it alone once and still recorded a native compile timeout (10.011731157s on the retained observation). Complete retained observation:


```json
[
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.080579663,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.072043248,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.211927374,
      "exit": 0,
      "stdout": "other\n4 false\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.087374582,
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
      "wall_seconds": 0.122242762,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 10.011731157,
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
]
```


After the gate exited, reran the same native build command alone, sequentially, using each revision’s cached admission compiler. Cache input records were checked against the full respective SHAs. No other heavy command ran. Both used the unchanged fixture from the checked-out HEAD; its Git blob is `367fcf1d9a384a462f420a41d793dd4c9c88c252` at **both** SHAs. Working directory was `/workspace/adamic`; outputs were outside the repository. The original 10-second deadline was enforced for each build using a separate process group killed on timeout.


| Revision | Native build exit | Seconds | Timed out | Native execution exit |
|---|---:|---:|---|---:|

| `730e7c26e590af0c57c545ba0cb905b09e4bee54` | 0 | 4.488 | False | 70 |

| `63eb56cd701c744a603f3c694295d749ba9d46dc` | 0 | 3.410 | False | 70 |


Both builds had empty stdout/stderr. Both executables printed `other\n` and exited 70 with exactly:


```text
adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false
```


That matches the JavaScript guard from the gate. Node exits 0 and prints `other\n4 false\n`; this fixture is a listed checked-boundary negative witness, so the guard is the expected outcome rather than a new source/native agreement. These successful isolated reruns resolve the specific transient timing observation but do **not** rewrite the original gate verdict to pass. No fresh complete admission run was performed.


Exact rerun commands and outcomes:


```json
[
  {
    "revision": "base",
    "sha": "730e7c26e590af0c57c545ba0cb905b09e4bee54",
    "command": [
      "/home/agent/.cache/adamic-build/local/027fcb9152937f26f93d97114259158a9a7989c7d0c18338f16520761f4d9981/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
      "-o",
      "/tmp/adamic-proof/nested-boolean-base"
    ],
    "timeout_seconds": 10,
    "wall_seconds": 4.4881525700002385,
    "exit": 0,
    "timed_out": false,
    "stdout": "",
    "stderr": "",
    "native": {
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
    }
  },
  {
    "revision": "head",
    "sha": "63eb56cd701c744a603f3c694295d749ba9d46dc",
    "command": [
      "/home/agent/.cache/adamic-build/local/7feaf65705ee3ea7f6ca5e7385eb0b5a68bb595f1893005b4103254a955560fb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
      "-o",
      "/tmp/adamic-proof/nested-boolean-head"
    ],
    "timeout_seconds": 10,
    "wall_seconds": 3.4100306000000273,
    "exit": 0,
    "timed_out": false,
    "stdout": "",
    "stderr": "",
    "native": {
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
    }
  }
]
```


## Report delivery scope


This report is the only file to be committed, on new branch `proof/lower-39w-4-report` made directly from tested HEAD. The only authorized push destination is `origin` branch `proof/lower-39w-4-report`. No compiler changes, dependency installations, mutation edits, or external evidence files are included in the commit.


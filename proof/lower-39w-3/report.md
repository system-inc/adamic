# lower-39w-3 proof report

Completed 2026-11-12T04:37:59.788552+00:00.

All requested formatting, vet, focused lower, uncached oracle, full IR, full lower, and Explain checks passed. The complete admission delta exited 3 with verdict `void` because seven native compiles exceeded its 10-second limit, including its built-in serial retries. All seven subsequently completed without timeouts in isolated serial reruns at both exact revisions, under the same limits. The original complete run remains recorded as void; the targeted reruns are partial coverage, not a replacement full-gate pass.

## Revisions and environment

- Repository: `https://github.com/system-inc/adamic.git`.
- Tested HEAD: `159f072e9faac3f02a7d26a88a60b41bfb7074c4` (detached).
- Main/base: `730e7c26e590af0c57c545ba0cb905b09e4bee54`; HEAD is its direct child.
- Fetched `origin compiler/predicate-false-read-3`, checked out the exact HEAD detached, and successfully ran `git submodule update --init --recursive`.
- Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; nested `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` first. Created its `TMPDIR=/tmp/adamic-gate` outside the repository before running Go.
- Separate upstream checkout: `/tmp/lower-39w-3-proof/typescript`, fetched from `https://github.com/microsoft/TypeScript.git`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`; exported as `ADAMIC_TYPESCRIPT_SOURCE`.
- Copied only `stage3/api/package.json` and `package-lock.json` to `/tmp/lower-39w-3-proof/api`, then installed with `npm ci --prefix /tmp/lower-39w-3-proof/api`. Packages remain outside the repository. Pinned TypeScript 6.0.3 and @types/node 25.3.3 installed successfully.
- `tsc --version`: `Version 6.0.3`; `go version`: `go version go1.27.1 linux/amd64`; Node: `v24.19.0`.
- One heavy command at a time throughout; no concurrent proof commands. No compiler source edits or fixes.

Every proof command sourced this external environment file:
```bash
if [ -f /workspace/adamic-tools/env.sh ]; then source /workspace/adamic-tools/env.sh; fi
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lower-39w-3-proof/typescript
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export PATH=/tmp/lower-39w-3-proof/api/node_modules/.bin:$PATH
export NODE_PATH=/tmp/lower-39w-3-proof/api/node_modules
```

The admission tool additionally provisioned its own temporary external revision checkouts from their pinned package locks. The main repository working tree remained clean throughout testing.

## 1. Formatting and vet

```bash
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

`gofmt` exit 0; stdout empty (required). `go vet` exit 0; stdout/stderr empty. PASS.

## 2. Focused lower tests

```bash
go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m
```

Exit 0. All tests/subtests passed; no failures or skips. Complete output follows, including every subtest result.

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
=== CONT  TestPredicateOverloadCallbackPushes
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
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.40s)
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestPredicateOverloadWithoutPredicate (0.41s)
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.49s)
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
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestPredicateOverloadNamedCallback
=== CONT  TestPredicateOverloadInPlaceCallback
--- PASS: TestIndirectPredicateOverloadIsPending (0.27s)
--- PASS: TestPredicateOverloadInPlaceCallback (0.59s)
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadCallbackPushes (1.33s)
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestPredicateOverloadNamedCallback (1.05s)
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
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadStaysOutOfA (0.26s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
--- PASS: TestPredicateOverloadCallbackFromParameter (0.25s)
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.59s)
=== CONT  TestKindPredicateCallIsItsComparison/conditional
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
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestKindPredicateThroughModules/import_as
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
--- PASS: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.21s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.09s)
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.15s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.10s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.06s)
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestPredicateOverloadRuntime/overload_some
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (24.81s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (25.16s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (25.56s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (25.40s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.40s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.53s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (1.94s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (2.79s)
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.05s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.25s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.25s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.23s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.23s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.21s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.24s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.25s)
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (2.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.23s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.07s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.20s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.09s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.08s)
--- PASS: TestPredicateOverloadLanguageLies (0.00s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.14s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.16s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.88s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.18s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.19s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (1.24s)
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
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
--- PASS: TestPredicateOverloadRuntime (43.19s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (27.15s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (2.42s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.75s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (1.08s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.81s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.87s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.89s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.73s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.70s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.77s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.88s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.70s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.84s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.89s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.97s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.62s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (1.11s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	43.703s
```

## 3. Uncached oracle tests

```bash
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 60m
```

Exit 0. All tests/subtests passed; no failures or skips. Native gate cache: 0 hits/19 misses; Node gate cache: 0 hits/68 misses. Complete output follows.

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
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/renarrowed
--- PASS: TestCountsAreRecordedTable (0.02s)
=== CONT  TestPredicateFalseRead/capture
--- PASS: TestCheckedPredicateCountsAreRecorded (53.07s)
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba miss 350.99
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 350.99
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 350.99
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 350.99
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 342.01
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/again
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.06
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.02
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.02
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.02
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/capture (397.71s)
    --- PASS: TestPredicateFalseRead/early_exit (344.73s)
    --- PASS: TestPredicateFalseRead/no_read (397.80s)
    --- PASS: TestPredicateFalseRead/renarrowed (397.86s)
    --- PASS: TestPredicateFalseRead/else_read (1.78s)
    --- PASS: TestPredicateFalseRead/false_read (1.72s)
    --- PASS: TestPredicateFalseRead/again (1.70s)
    --- PASS: TestPredicateFalseRead/getter_capture (1.81s)
    --- PASS: TestPredicateFalseRead/getter (0.26s)
    --- PASS: TestPredicateFalseRead/spread (0.81s)
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.02
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.04
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.02
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types ead767b53bba hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (421.53s)
    --- PASS: TestCheckedPredicateOracle/true (399.07s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.30s)
    --- PASS: TestCheckedPredicateOracle/false (0.93s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.69s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.05s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.03s)
    --- PASS: TestCheckedPredicateOracle/valid (1.49s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.99s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.06s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.16s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.09s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.63s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.47s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.63s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.62s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.30s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.35s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.42s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.25s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=68
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	421.676s
```

## 4. Full IR and lower suites

```bash
go test ./internal/ir -count=1 -timeout 30m
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

IR exit 0. Failures: none. Environment failures: none.

```text
ok  	github.com/system-inc/adamic/internal/ir	142.340s
```

Full lower exit 0. Failing tests/first error lines: none. Environment failures: none. Complete `/tmp/lower.log`:
```text
ok  	github.com/system-inc/adamic/internal/lower	2366.991s
```

## 5. Explain tests

```bash
go test ./cmd/adamic -run 'Explain' -count=1 -timeout 30m
```

Exit 0. Failures: none.

```text
ok  	github.com/system-inc/adamic/cmd/adamic	20.303s
```

## 6. Complete admission delta

```bash
python3 cloud/admission-corpus/manifest.py --sha 159f072e9faac3f02a7d26a88a60b41bfb7074c4 > /tmp/lower-39w-3-proof/manifest.json
go build -o /tmp/lower-39w-3-proof/admission-delta ./cmd/adamic-admission-delta
/tmp/lower-39w-3-proof/admission-delta --base 730e7c26e590af0c57c545ba0cb905b09e4bee54 --head 159f072e9faac3f02a7d26a88a60b41bfb7074c4 --manifest /tmp/lower-39w-3-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json > /tmp/lower-39w-3-proof/delta.json 2> /tmp/lower-39w-3-proof/delta.stderr
```

Run alone, no budget/filter/shard, full generated HEAD corpus. Compiler classification timeout 45 seconds; each runtime build/execution timeout 10 seconds.

| Measurement | Result |
| --- | --- |
| Process exit | 3 |
| Verdict | void |
| void | True |
| diff_count | 11 |
| admitted | 0 |
| newly refused | 0 |
| Classified programs | 5312 |
| Accepted by both | 3328 |
| Refused by both | 1984 |
| Classification timeouts | 0 |
| Programs with classification stderr differences | 0 |
| Runtime selections | 115 |
| Omitted | 0 |
| Node agreements | 10 |
| Accepted negative witnesses | 47 |
| Accepted runtime stops | 51 |
| Timed-out runs / retried alone / still timed out | 7 / 7 / 7 |

Classification stderr comparison was byte-for-byte across all programs. There are no differing programs, so there are no paired differing stderr texts to list.

Manifest corpus counts:
```text
witnesses=2 fixtures=4364 gaps=105 review=20 fuzz=0 stage3=821
```

Manifest provenance:
```json
{
  "generator_revision": "159f072e9faac3f02a7d26a88a60b41bfb7074c4",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "a556ab646b95fd56adbf03233e4584cf79f20c2a",
  "complete_corpus": true,
  "workers": 2,
  "classification": "lowering-only",
  "phase_seconds": {
    "base_build": 68.639337108,
    "checkout": 23.305908838,
    "checkout_and_provision": 25.063291651,
    "classification": 1137.098479738,
    "head_build": 37.009931843,
    "manifest_and_verification": 33.177800013,
    "provision": 1.7573781080000002,
    "runtime": 752.449368017,
    "total": 2053.452073153
  }
}
```

Changed programs (each accepted by both):

- `internal/lower/testdata/predicates_checked/false_read.a` — blob `a9df503eff15df40a0f3ca4ef5e7a0704f534b22`.
- `internal/oracle/testdata/predicate_false_read/again.ts` — blob `9f6f8070d0b6347b7ceeb08c9159b78b160e172f`.
- `internal/oracle/testdata/predicate_false_read/capture.ts` — blob `1689498e103b3729216dbeaeaf3e1a8c6f635bda`.
- `internal/oracle/testdata/predicate_false_read/early_exit.ts` — blob `c89883d7b0206eeb6678fce44716a1d5844459ff`.
- `internal/oracle/testdata/predicate_false_read/else_read.ts` — blob `16aad0696c46b00f5029e0bc9d3fea92bda804b9`.
- `internal/oracle/testdata/predicate_false_read/false_read.ts` — blob `5954ab19e76cecd29dd4619c61c2774ba530bbfa`.
- `internal/oracle/testdata/predicate_false_read/getter.ts` — blob `6e488b9c05361e22d3f0328ac98b45513b2953dc`.
- `internal/oracle/testdata/predicate_false_read/getter_capture.ts` — blob `540af883f64cd9e7a975524748489ae53a320518`.
- `internal/oracle/testdata/predicate_false_read/no_read.ts` — blob `81a574f9ee239a4df14fa3bfc3eda7c596d73b78`.
- `internal/oracle/testdata/predicate_false_read/renarrowed.ts` — blob `abedbe6418d601b79db8200bb51e4b2de8928ff7`.
- `internal/oracle/testdata/predicate_false_read/spread.ts` — blob `a8bcb7a42d2ac99d056faa0fe578c491a3ce6c7b`.

### Void reasons

Every void reason is a native compile timeout. All seven programs lowered successfully at both revisions with empty classification stderr; none is a classification timeout. The complete gate stderr is:
```text
negative witnesses: 50
ruled run-time stops: 51
ruled backend divergences: 0
runtime timeouts: 7 runs timed out, 7 retried alone, 7 timed out again
changed programs: 11
admission delta is void: internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts: native compile timeout; internal/oracle/testdata/reflect_get/stop_literal_getter.a: native compile timeout
```

Original per-program timeout observations, including the built-in serial retries:
#### `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts`

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
  "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.198875476,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.148101015,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.158403206,
    "exit": 0,
    "stdout": "other\n4 false\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.090772357,
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
    "wall_seconds": 0.087414613,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.014129808,
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

#### `internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts`

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
  "blob": "292abbff3e18d3b3b150620a8e8a66814324117f",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.141082255,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.199984371,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.155412162,
    "exit": 0,
    "stdout": "other\n4 true\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.09764017,
    "exit": 0,
    "stdout": "other\n4 true\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.103351016,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.004265571,
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

#### `internal/oracle/testdata/intersection_views_stops/nested_literal.ts`

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
  "blob": "46d3a819e8b786ab0d97d7edb5fbf6aa75b1d4ec",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.193044097,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.090596559,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.224876864,
    "exit": 0,
    "stdout": "other\nnode:fs 12\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.06949709,
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n"
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.123017143,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.005966775,
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

#### `internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts`

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
  "blob": "79bd0de0dd576b63f6672c1acb9af631b79133c0",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.157032705,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.147611905,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.202394199,
    "exit": 0,
    "stdout": "other\nnode:fs 11\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.104083609,
    "exit": 0,
    "stdout": "other\nnode:fs 11\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.148388909,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.0064493,
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

#### `internal/oracle/testdata/intersection_views_stops/nested_token.ts`

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
  "blob": "f5773367111892f88cc1b3203da59ec2491d3515",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.156253022,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.128656316,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.13654382,
    "exit": 0,
    "stdout": "other\nnode:fs undefined\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.225114925,
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n"
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.263722363,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.012298,
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

#### `internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts`

```json
{
  "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
  "blob": "80ddd7b8f8cba29eb3fb0e5aa08aa2ee2974e5a6",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.133337369,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.268006771,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.350769463,
    "exit": 0,
    "stdout": "other\nnode:fs 2\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.135491356,
    "exit": 0,
    "stdout": "other\nnode:fs 2\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.233573777,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.009042322,
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

#### `internal/oracle/testdata/reflect_get/stop_literal_getter.a`

```json
{
  "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
  "blob": "ee5d428ca5813c54046541e9d0a6d83ede986d55",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.226066809,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.154216353,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.458519218,
    "exit": 0,
    "stdout": "got\nstring\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.085564368,
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.135438033,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.001495137,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "timeout"
  },
  "agree": false,
  "negative_related": false,
  "negative_witness": false,
  "notyet_related": true,
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

### Timeout reruns alone at both SHAs

After the complete gate exited, each exact revision was rerun separately. Each rerun used a manifest containing only the seven timeout programs and `--workers 1`, so cases and all their runtime commands ran serially. `--base` and `--head` were set to the same SHA to evaluate that revision on its own source tree. The original 45-second classification and 10-second runtime limits were retained; no limit was increased. `--corpus-filter fixtures` explicitly marks these targeted runs as partial.

#### HEAD `159f072e9faac3f02a7d26a88a60b41bfb7074c4`

```bash
/tmp/lower-39w-3-proof/admission-delta --base 159f072e9faac3f02a7d26a88a60b41bfb7074c4 --head 159f072e9faac3f02a7d26a88a60b41bfb7074c4 --manifest /tmp/lower-39w-3-proof/retry-head-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --corpus-filter fixtures --workers 1 --json
```

Exit 0; verdict `partial`; void `false`; runtime timeouts `0`. All seven completed: three Node agreements, three accepted negative witnesses (with passing repairs), and one ruled runtime stop.

```text
negative witnesses: 50
ruled run-time stops: 1
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 0
```

| Program | Native build exit | Native build seconds | Native run exit | Validation |
| --- | --- | --- | --- | --- |
| `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts` | 0 | 2.044293 | 70 | accepted negative witness; repair agrees |
| `internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts` | 0 | 2.016951 | 0 | Node agreement |
| `internal/oracle/testdata/intersection_views_stops/nested_literal.ts` | 0 | 2.058333 | 70 | accepted negative witness; repair agrees |
| `internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts` | 0 | 2.088382 | 0 | Node agreement |
| `internal/oracle/testdata/intersection_views_stops/nested_token.ts` | 0 | 2.141112 | 70 | accepted negative witness; repair agrees |
| `internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts` | 0 | 2.097624 | 0 | Node agreement |
| `internal/oracle/testdata/reflect_get/stop_literal_getter.a` | 0 | 8.273499 | 70 | ruled runtime stop |

Exact targeted manifest:
```json
{
  "sha": "159f072e9faac3f02a7d26a88a60b41bfb7074c4",
  "generator": "cloud/admission-corpus/manifest.py",
  "corpora": [
    {
      "name": "fixtures",
      "programs": [
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
          "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
          "blob": "292abbff3e18d3b3b150620a8e8a66814324117f"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
          "blob": "46d3a819e8b786ab0d97d7edb5fbf6aa75b1d4ec"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
          "blob": "79bd0de0dd576b63f6672c1acb9af631b79133c0"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
          "blob": "f5773367111892f88cc1b3203da59ec2491d3515"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
          "blob": "80ddd7b8f8cba29eb3fb0e5aa08aa2ee2974e5a6"
        },
        {
          "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
          "blob": "ee5d428ca5813c54046541e9d0a6d83ede986d55"
        }
      ]
    }
  ]
}
```

Full per-program observations, preserving both backend outputs and diagnostic texts:
```json
[
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.066978476,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.066978476,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.132507756,
      "exit": 0,
      "stdout": "other\n4 false\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.072006454,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
    },
    "native": {
      "wall_seconds": 0.00501492,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.084808601,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.044293428,
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
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
    "blob": "292abbff3e18d3b3b150620a8e8a66814324117f",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.107409292,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.107409292,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.175229106,
      "exit": 0,
      "stdout": "other\n4 true\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.056593945,
      "exit": 0,
      "stdout": "other\n4 true\n",
      "stderr": ""
    },
    "native": {
      "wall_seconds": 0.002485968,
      "exit": 0,
      "stdout": "other\n4 true\n",
      "stderr": ""
    },
    "javascript_compile": {
      "wall_seconds": 0.064145286,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.016951395,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": true,
    "negative_related": true,
    "negative_witness": false,
    "notyet_related": false,
    "ruled_runtime_stop": false,
    "ruled_backend_divergence": false,
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
    "blob": "46d3a819e8b786ab0d97d7edb5fbf6aa75b1d4ec",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.11026799,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.11026799,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.117540878,
      "exit": 0,
      "stdout": "other\nnode:fs 12\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.066183269,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n"
    },
    "native": {
      "wall_seconds": 0.002414369,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.06482113,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.058332536,
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
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
    "blob": "79bd0de0dd576b63f6672c1acb9af631b79133c0",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.068250688,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.068250688,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.131780187,
      "exit": 0,
      "stdout": "other\nnode:fs 11\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.062394414,
      "exit": 0,
      "stdout": "other\nnode:fs 11\n",
      "stderr": ""
    },
    "native": {
      "wall_seconds": 0.006665627,
      "exit": 0,
      "stdout": "other\nnode:fs 11\n",
      "stderr": ""
    },
    "javascript_compile": {
      "wall_seconds": 0.106751801,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.088381925,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": true,
    "negative_related": true,
    "negative_witness": false,
    "notyet_related": false,
    "ruled_runtime_stop": false,
    "ruled_backend_divergence": false,
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
    "blob": "f5773367111892f88cc1b3203da59ec2491d3515",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.066131934,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.066131934,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.138523517,
      "exit": 0,
      "stdout": "other\nnode:fs undefined\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.050117719,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n"
    },
    "native": {
      "wall_seconds": 0.002539015,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.06352239,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.141111804,
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
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
    "blob": "80ddd7b8f8cba29eb3fb0e5aa08aa2ee2974e5a6",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.075583842,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.075583842,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.113828599,
      "exit": 0,
      "stdout": "other\nnode:fs 2\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.082250499,
      "exit": 0,
      "stdout": "other\nnode:fs 2\n",
      "stderr": ""
    },
    "native": {
      "wall_seconds": 0.002144762,
      "exit": 0,
      "stdout": "other\nnode:fs 2\n",
      "stderr": ""
    },
    "javascript_compile": {
      "wall_seconds": 0.073649878,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.097624265,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": true,
    "negative_related": true,
    "negative_witness": false,
    "notyet_related": false,
    "ruled_runtime_stop": false,
    "ruled_backend_divergence": false,
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "blob": "ee5d428ca5813c54046541e9d0a6d83ede986d55",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.054099428,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.054099428,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.118943953,
      "exit": 0,
      "stdout": "got\nstring\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.072675038,
      "exit": 70,
      "stdout": "got\n",
      "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
    },
    "native": {
      "wall_seconds": 0.002512193,
      "exit": 70,
      "stdout": "got\n",
      "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.062429116,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 8.273499406,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": false,
    "negative_related": false,
    "negative_witness": false,
    "notyet_related": true,
    "ruled_runtime_stop": true,
    "ruled_backend_divergence": false,
    "sampled": true
  }
]
```

#### BASE `730e7c26e590af0c57c545ba0cb905b09e4bee54`

```bash
/tmp/lower-39w-3-proof/admission-delta --base 730e7c26e590af0c57c545ba0cb905b09e4bee54 --head 730e7c26e590af0c57c545ba0cb905b09e4bee54 --manifest /tmp/lower-39w-3-proof/retry-base-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --corpus-filter fixtures --workers 1 --json
```

Exit 0; verdict `partial`; void `false`; runtime timeouts `0`. All seven completed: three Node agreements, three accepted negative witnesses (with passing repairs), and one ruled runtime stop.

```text
negative witnesses: 50
ruled run-time stops: 1
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 0
```

| Program | Native build exit | Native build seconds | Native run exit | Validation |
| --- | --- | --- | --- | --- |
| `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts` | 0 | 3.812320 | 70 | accepted negative witness; repair agrees |
| `internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts` | 0 | 2.008017 | 0 | Node agreement |
| `internal/oracle/testdata/intersection_views_stops/nested_literal.ts` | 0 | 2.082625 | 70 | accepted negative witness; repair agrees |
| `internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts` | 0 | 2.083508 | 0 | Node agreement |
| `internal/oracle/testdata/intersection_views_stops/nested_token.ts` | 0 | 2.256500 | 70 | accepted negative witness; repair agrees |
| `internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts` | 0 | 2.188319 | 0 | Node agreement |
| `internal/oracle/testdata/reflect_get/stop_literal_getter.a` | 0 | 2.083952 | 70 | ruled runtime stop |

Exact targeted manifest:
```json
{
  "sha": "730e7c26e590af0c57c545ba0cb905b09e4bee54",
  "generator": "cloud/admission-corpus/manifest.py",
  "corpora": [
    {
      "name": "fixtures",
      "programs": [
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
          "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
          "blob": "292abbff3e18d3b3b150620a8e8a66814324117f"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
          "blob": "46d3a819e8b786ab0d97d7edb5fbf6aa75b1d4ec"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
          "blob": "79bd0de0dd576b63f6672c1acb9af631b79133c0"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
          "blob": "f5773367111892f88cc1b3203da59ec2491d3515"
        },
        {
          "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
          "blob": "80ddd7b8f8cba29eb3fb0e5aa08aa2ee2974e5a6"
        },
        {
          "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
          "blob": "ee5d428ca5813c54046541e9d0a6d83ede986d55"
        }
      ]
    }
  ]
}
```

Full per-program observations, preserving both backend outputs and diagnostic texts:
```json
[
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "blob": "367fcf1d9a384a462f420a41d793dd4c9c88c252",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.148699763,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.148699763,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.164076469,
      "exit": 0,
      "stdout": "other\n4 false\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.100545683,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
    },
    "native": {
      "wall_seconds": 0.004057041,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.104973558,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 3.812320229,
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
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
    "blob": "292abbff3e18d3b3b150620a8e8a66814324117f",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.108483047,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.108483047,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.135947821,
      "exit": 0,
      "stdout": "other\n4 true\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.051608125,
      "exit": 0,
      "stdout": "other\n4 true\n",
      "stderr": ""
    },
    "native": {
      "wall_seconds": 0.002731766,
      "exit": 0,
      "stdout": "other\n4 true\n",
      "stderr": ""
    },
    "javascript_compile": {
      "wall_seconds": 0.05781783,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.008016734,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": true,
    "negative_related": true,
    "negative_witness": false,
    "notyet_related": false,
    "ruled_runtime_stop": false,
    "ruled_backend_divergence": false,
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
    "blob": "46d3a819e8b786ab0d97d7edb5fbf6aa75b1d4ec",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.09768593,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.09768593,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.111107163,
      "exit": 0,
      "stdout": "other\nnode:fs 12\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.061461641,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n"
    },
    "native": {
      "wall_seconds": 0.002983967,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.062194908,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.082625133,
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
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
    "blob": "79bd0de0dd576b63f6672c1acb9af631b79133c0",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.057852304,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.057852304,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.236506554,
      "exit": 0,
      "stdout": "other\nnode:fs 11\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.064342694,
      "exit": 0,
      "stdout": "other\nnode:fs 11\n",
      "stderr": ""
    },
    "native": {
      "wall_seconds": 0.002453345,
      "exit": 0,
      "stdout": "other\nnode:fs 11\n",
      "stderr": ""
    },
    "javascript_compile": {
      "wall_seconds": 0.060653023,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.083507636,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": true,
    "negative_related": true,
    "negative_witness": false,
    "notyet_related": false,
    "ruled_runtime_stop": false,
    "ruled_backend_divergence": false,
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
    "blob": "f5773367111892f88cc1b3203da59ec2491d3515",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.111775769,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.111775769,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.101980233,
      "exit": 0,
      "stdout": "other\nnode:fs undefined\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.052936888,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n"
    },
    "native": {
      "wall_seconds": 0.00247575,
      "exit": 70,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.071236025,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.256500046,
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
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
    "blob": "80ddd7b8f8cba29eb3fb0e5aa08aa2ee2974e5a6",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.179728385,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.179728385,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.114346617,
      "exit": 0,
      "stdout": "other\nnode:fs 2\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.050477595,
      "exit": 0,
      "stdout": "other\nnode:fs 2\n",
      "stderr": ""
    },
    "native": {
      "wall_seconds": 0.002349152,
      "exit": 0,
      "stdout": "other\nnode:fs 2\n",
      "stderr": ""
    },
    "javascript_compile": {
      "wall_seconds": 0.065160153,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.188318947,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": true,
    "negative_related": true,
    "negative_witness": false,
    "notyet_related": false,
    "ruled_runtime_stop": false,
    "ruled_backend_divergence": false,
    "sampled": true
  },
  {
    "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "blob": "ee5d428ca5813c54046541e9d0a6d83ede986d55",
    "corpus": "fixtures",
    "class": "accepted-by-both",
    "base_compile": {
      "wall_seconds": 0.221701371,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "compile": {
      "wall_seconds": 0.221701371,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "node": {
      "wall_seconds": 0.141132385,
      "exit": 0,
      "stdout": "got\nstring\n",
      "stderr": ""
    },
    "javascript": {
      "wall_seconds": 0.074173845,
      "exit": 70,
      "stdout": "got\n",
      "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
    },
    "native": {
      "wall_seconds": 0.002508165,
      "exit": 70,
      "stdout": "got\n",
      "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
    },
    "javascript_compile": {
      "wall_seconds": 0.063121025,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "native_compile": {
      "wall_seconds": 2.083952295,
      "exit": 0,
      "stdout": "",
      "stderr": ""
    },
    "agree": false,
    "negative_related": false,
    "negative_witness": false,
    "notyet_related": true,
    "ruled_runtime_stop": true,
    "ruled_backend_divergence": false,
    "sampled": true
  }
]
```

## Result and report delivery

All requested Go checks passed. The full-corpus classification found no admission change, no newly refused programs, and no changed classification stderr. The complete runtime gate was void on seven native build deadlines; isolated reruns cleared every one at both SHAs under the same limits. These observations establish that the seven timeout cases pass separately, while preserving the full run’s original void verdict.

This report is the only repository file created for delivery. The requested delivery branch is `proof/lower-39w-3-report`, created from tested HEAD `159f072e9faac3f02a7d26a88a60b41bfb7074c4`. Only `proof/lower-39w-3/report.md` is to be committed and that branch alone pushed to origin. Raw working logs remain outside the repository under `/tmp/lower-39w-3-proof/` and `/tmp/lower.log`.

# Flow corpus refusal classification

Merged current main `cf735d9fb` in merge commit `b918aaf76`; no conflicts. The fix changes only two flow test files and review evidence. No fixtures move, and no compiler admission changes.

`programs` in `internal/flow/flow_test.go` already removes deliberate refusal controls with named predicates and exact filenames. The new `refusedHiddenBoundaryFixture` uses that same mechanism for `hidden_boundary_alias_cast_refused.a` and `hidden_boundary_optional_array_mutation_refused.a`. Neither has lowered IR to trace. Their explicit oracle tests remain responsible for the refusals and verify source Node stdout `3\n` and `text|7\n`, respectively. The new flow classification test requires both absent and seven runnable hidden-boundary controls present.

## Fixture audit

Compared added oracle fixtures against the merged main. Grepped hidden-boundary and overload oracle tests for Refused, NotYet and explicit refusal assertions. All 65 added fixtures are classified below and in `new-fixtures.json`. There are two top-level refusal controls, 18 runnable top-level fixtures, 18 nested refusal controls, two nested NotYet controls and 25 runnable nested fixtures. The nested fixtures are already outside the flow corpus's `../oracle/testdata/*.a` glob. No additional top-level deliberate refusal was found.

| Fixture | Flow classification |
| --- | --- |
| `internal/oracle/testdata/census_overload_binder_result_checked.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_boundary_alias_cast_refused.a` | Refused, exact flow exclusion |
| `internal/oracle/testdata/hidden_boundary_generic_optional_array.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_boundary_generic_tnode.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_boundary_never_array.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_boundary_never_array_observations.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_boundary_optional_array_mutation_refused.a` | Refused, exact flow exclusion |
| `internal/oracle/testdata/hidden_boundary_optional_array_order.a` | runnable, flow remainder |
| `internal/oracle/testdata/hidden_wave2_method_destructuring.a` | runnable, flow remainder |
| `internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_mutation.a` | NotYet, nested outside flow glob |
| `internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_value.a` | NotYet, nested outside flow glob |
| `internal/oracle/testdata/overload_callback_served.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_field_hatch/kind-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/kind-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/kind-valid.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/kind-valid.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/value-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/value-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/value-undefined.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/value-undefined.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/value-valid.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_field_hatch/value-valid.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_results_binding.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_results_evaluate.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_results_parameter.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_results_scalar.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_results_transform.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_structural_block.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_structural_evaluator.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_structural_factory.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_structural_fields.a` | runnable, flow remainder |
| `internal/oracle/testdata/overload_values/module-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/module-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/module-valid.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/module-valid.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/narrow-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/narrow-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/narrow-valid.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/narrow-valid.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/order-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/order-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/order-valid.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/order-valid.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/proven.a` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/returned-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/returned-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_values/returned-valid.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_values/returned-valid.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/default-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/default-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/helper-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/helper-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/helper.a` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/helper.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/instantiations.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/instantiations.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/original-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/original-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/original.a` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/original.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/overloaded-helper-liar.a` | Refused, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/overloaded-helper-liar.ts` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/overloaded-helper.a` | runnable, nested outside flow glob |
| `internal/oracle/testdata/overload_visitors/overloaded-helper.ts` | runnable, nested outside flow glob |

## Validation

All commands had an outer 90-second limit. Every Go test command also used `-timeout 90s`, and test output went directly to files.

- `go test ./internal/flow -run '^TestFlowCorpus' -count=1 -timeout 90s -json`: passed, including the remainder's 18 new runnable programs. The final restored run sets `ADAMIC_UNIT_BUDGET=1` on the four-CPU box.
- Listed all selectable corpus leaves with `go test ./internal/flow -list '^TestFlow(Program|SingleAssignment|MutationRanges|GraphPaths|Liveness)' -timeout 90s`. Ran every listed leaf with exact anchored selections, `-count=1 -timeout 90s -json`, in batches of 80, with `ADAMIC_GATE_UNCACHED=1`. All **831 leaves pass exactly once**, asserted against the selection list in `corpus-results.json`. The slowest existing leaf is **15.710 seconds**. Coverage reports **846 programs, 3,384 analysis pieces and 831 selectable units**; the remainder handles the 18 new programs separately.
- One grouped batch command hit its outer 90-second limit during the last batch. Its interrupted log is retained and not counted as a pass. The final 31-leaf batch was rerun alone and passed in 18.694 wall seconds. Every batch has a passing final status.
- `go test ./internal/oracle -run '^TestHiddenBoundary04(AliasCast|Mutation)Refused$' -count=1 -timeout 90s -json`: passed in 0.245 package seconds; both leaves 0.230 seconds. Node runs successfully; the compiler retains the specific unchecked-cast and mutable-generic refusals.
- `go test ./internal/oracle -run '^TestOverload(FieldHatch|Values|Visitors|VisitorInstantiations)|^TestHiddenTNodeValueExactNotYet$' -count=1 -timeout 90s -json`: passed in 8.702 package seconds, checking the nested refusal controls and the exact generic-value NotYet control.
- No fixture changed or was added, so no counts regeneration was needed. `git diff --check` passes.

### Restored flow timings

| Test | Seconds |
| --- | ---: |
| TestFlowCorpusHiddenBoundaryRefusalClassification | 0.000 |
| TestFlowCorpusUnitsCoverEveryProgram | 0.010 |
| TestFlowCorpusSetupIsShared | 0.050 |
| TestFlowCorpusRemainder | 2.650 |
| flow package | 2.658 |

### Mutants

Each mutation is a Go overlay of the test helper, never a product source edit. Evidence sources use `.go.txt`.

| Mutant | Catcher |
| --- | --- |
| Remove alias-cast refusal classification | `TestFlowCorpusHiddenBoundaryRefusalClassification`: alias fixture entered runnable corpus; exit 1 |
| Remove optional-array mutation refusal classification | Same test: mutation fixture entered runnable corpus; exit 1 |
| Exclude every hidden fixture by prefix | Same test: all seven runnable controls left corpus; exit 1 |

All three failed their intended assertions, not compilation or time limits. The unmodified restored corpus tests pass afterward. The new leaf completes below 0.010 seconds at the JSON timer's precision.

Setup: Go ready 0.024s, Node ready 0.031s, Markdown dependencies ready 0.090s, submodules ready 0.109s, clang ready 0.216s, Go build ready 39.973s, build cache warm 40.110s, done 40.136s. `nproc=5`, cgroup CPU quota four. Sourced `/workspace/adamic-tools/env.sh` for each Go command. No whole package or repository gate was run; all flow corpus leaves and metadata tests were selected explicitly.

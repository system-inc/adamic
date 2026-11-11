# Captured-union compiler proof — #n7sx9gz

HEAD: `5eeac42cf8c0037c627f32b975dd254cbfd20c6b`. Main: `16429a30452d027809f0ea85bc395479423e493e`. Measurements performed on 2026-11-11 UTC.

HEAD passes all 64 captured-union oracle backend checks, including both held cycle programs. Main stops all 14 ordinary captured-union pins. M1 and M2 are detected by lowering refusals; M3 leaks on by_value, released, and alternating in both sanitizer and counted builds. The full lowerer suite has one baseline-reproduced path-sensitive assertion failure. The full oracle run has thirteen baseline-reproduced setup failures and two internal probe timeouts that pass in isolation at both SHAs. The full admission delta passes, admits eight programs, and newly refuses none.

## 1. Checkout and environment

Loaded `/workspace/adamic-tools/env.sh` before setup. Created its missing `TMPDIR=/tmp/adamic-gate` outside the repository. Ran `git fetch origin compiler/captured-union-2`, detached checkout of the exact HEAD, and `git submodule update --init --recursive`: all succeeded. Submodule pins: cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. `gofmt -l internal/lower internal/oracle` printed nothing.

Command: `go vet ./internal/lower ./internal/oracle`

Exit **0**; wall **213.19 seconds**.

```text

```

Provisioned upstream Microsoft TypeScript separately at `/tmp/n7s-proof/TypeScript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path. Copied stage3/api package.json and package-lock.json to `/tmp/n7s-proof/api`, ran npm ci there, and prepended its node_modules/.bin to PATH; `tsc --version` printed `Version 6.0.3`. All subsequent commands inherited these settings and the exports below, including admission-gate children. One heavy top-level command ran at a time.

```sh
source /workspace/adamic-tools/env.sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/n7s-proof/TypeScript
export PATH=/tmp/n7s-proof/api/node_modules/.bin:$PATH
export NODE_PATH=/tmp/n7s-proof/api/node_modules
```

NODE_PATH was added only for the explicitly described supplemental checks and admission run; it was absent from the initial pin/full-suite measurements.

The managed cloud-environment runtime skill was used for environment/dependency setup. A transient exec transport disconnect occurred before vet started; retry succeeded. A shallow cohere clone could not be used as a Git reference for main; normal submodule cloning succeeded. Neither event is a compiler test outcome.

## 2. HEAD pins
### Lowerer

Command: `go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat' -count=1 -v -timeout 30m`

Exit **1**; wall **28.45 seconds**.

| Test / subtest | Result |
|---|---|
| `TestCapturedMaybeBooleanStillStops` | FAIL |
| `TestCapturedUnionLowersToOneCell` | PASS |
| `TestCapturedUnionLowersToOneCell/alternating.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/class.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/graph.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/nested.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/released.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/by_value.a` | PASS |
| `TestCapturedUnionLowersToOneCell/alternating.a` | PASS |
| `TestCapturedUnionLowersToOneCell/by_value.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/released.a` | PASS |
| `TestCapturedUnionLowersToOneCell/nested.a` | PASS |
| `TestCapturedUnionLowersToOneCell/object.a` | PASS |
| `TestCapturedUnionLowersToOneCell/object.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/class.a` | PASS |
| `TestCapturedUnionLowersToOneCell/graph.a` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions` | PASS |

The only failure is a path-sensitive test assertion: got includes the absolute temporary path, while want is exactly `main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet`. The expected refusal wording and location suffix are present. No source/test changes were made to normalize this.

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== CONT  TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedMaybeBooleanStillStops
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops131462813/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
--- FAIL: TestCapturedMaybeBooleanStillStops (0.27s)
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
--- PASS: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.26s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.27s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (0.28s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.30s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.12s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.16s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (0.16s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.18s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.09s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.10s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.11s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.15s)
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
--- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat (0.01s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object (0.05s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only (0.13s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only (0.15s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only (0.14s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean (0.15s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function (0.05s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number (0.13s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?. (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union (0.05s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.335s
FAIL
```

### Oracle

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees|TestCapturedUnionCycleHeldOrRefused' -count=1 -v -timeout 40m`

Exit **0**; wall **317.87 seconds**.

| Pin | Overall | Native | Sanitized | Counted | JavaScript | Logged counts |
|---|---|---|---|---|---|---|
| class.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| class.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| object.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| object.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| by_value.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 35, frees 35, in regions 0 |
| by_value.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 35, frees 35, in regions 0 |
| released.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 96, frees 96, in regions 0 |
| released.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 96, frees 96, in regions 0 |
| nested.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0 |
| nested.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0 |
| alternating.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 54, frees 54, in regions 0 |
| alternating.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 54, frees 54, in regions 0 |
| graph.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0 |
| graph.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0 |
| cycle.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0 |
| cycle.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0 |

`TestCapturedUnionCycleHeldOrRefused/cycle.a`:
```text
    captured_union_test.go:67: held: graph cell true
```

`TestCapturedUnionCycleHeldOrRefused/cycle.ts`:
```text
    captured_union_test.go:67: held: graph cell true
```

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== RUN   TestCapturedUnionCycleHeldOrRefused
=== PAUSE TestCapturedUnionCycleHeldOrRefused
=== CONT  TestCapturedUnionAgrees
=== CONT  TestCapturedUnionCycleHeldOrRefused
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.a
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.a
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a
    captured_union_test.go:67: held: graph cell true
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts
    captured_union_test.go:67: held: graph cell true
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/native
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/sanitized
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/counted
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/counted
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:141: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts/counted
    captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/javascript
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:141: counts: allocations 96, frees 96, in regions 0
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a/counted
    captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== NAME  TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f miss 217.04
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 217.04
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 217.04
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 217.05
=== NAME  TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 217.05
=== CONT  TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/alternating.ts
--- PASS: TestCapturedUnionCycleHeldOrRefused (0.00s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a (259.02s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/native (0.37s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/sanitized (0.81s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/counted (0.40s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/javascript (217.47s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts (259.02s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/native (0.38s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/sanitized (0.74s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/counted (0.40s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/javascript (217.54s)
=== CONT  TestCapturedUnionAgrees/nested.a
=== CONT  TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:141: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:141: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:141: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.04
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:141: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.03
=== CONT  TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.03
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
--- PASS: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/class.ts (258.95s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.37s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (0.93s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.41s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (217.23s)
    --- PASS: TestCapturedUnionAgrees/released.a (258.99s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.40s)
        --- PASS: TestCapturedUnionAgrees/released.a/sanitized (0.85s)
        --- PASS: TestCapturedUnionAgrees/released.a/counted (0.39s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (217.45s)
    --- PASS: TestCapturedUnionAgrees/alternating.a (259.03s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.34s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/sanitized (0.58s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/counted (0.29s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (217.83s)
    --- PASS: TestCapturedUnionAgrees/nested.a (1.79s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.25s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.65s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.32s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (1.81s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.26s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (0.73s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.30s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/alternating.ts (1.84s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.32s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/sanitized (0.73s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/counted (0.29s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/by_value.a (1.60s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.27s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/sanitized (0.49s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/counted (0.31s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.16s)
    --- PASS: TestCapturedUnionAgrees/by_value.ts (1.96s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.36s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/sanitized (0.50s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (0.27s)
    --- PASS: TestCapturedUnionAgrees/released.ts (2.03s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.36s)
        --- PASS: TestCapturedUnionAgrees/released.ts/sanitized (0.50s)
        --- PASS: TestCapturedUnionAgrees/released.ts/counted (0.67s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/object.ts (1.80s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.33s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (0.61s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/object.a (1.95s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.47s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (0.63s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.40s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.16s)
    --- PASS: TestCapturedUnionAgrees/class.a (1.85s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.43s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (0.54s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.42s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.21s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (44.65s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (12.42s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (18.48s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (13.32s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/graph.a (44.72s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (12.34s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (18.59s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (13.23s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.07s)
PASS
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	303.742s
```

## 3. Main measured with copied HEAD programs and tests

Created `/tmp/n7s-main` as a detached Git worktree at the exact main SHA, initialized its submodules, and copied only internal/lower/captured_union_test.go, internal/oracle/captured_union_test.go, and internal/oracle/testdata/captured_union/ from HEAD. Removed this worktree after the two measurements. Later recreated the same external worktree solely for full-suite failure reruns.

All ordinary pins stop during lowering, so native/sanitized/counted/JavaScript subtests are not reached. The cycle tests report PASS/refused, but their diagnostics are ordinary captured-union refusals: `strings.Contains(err.Error(), "cycle")` also matches the temporary path/filename. These passes do not establish cycle-specific refusal.

Command: `cd /tmp/n7s-main && go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops' -count=1 -v -timeout 30m`

Exit **1**; wall **218.62 seconds**.

| Test / subtest | Result |
|---|---|
| `TestCapturedMaybeBooleanStillStops` | FAIL |
| `TestCapturedUnionLowersToOneCell` | FAIL |
| `TestCapturedUnionLowersToOneCell/class.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/graph.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/alternating.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/graph.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/nested.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/nested.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/alternating.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/released.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/released.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/by_value.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/by_value.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/object.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/object.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/class.a` | FAIL |

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedMaybeBooleanStillStops
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops40632028/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
--- FAIL: TestCapturedMaybeBooleanStillStops (0.17s)
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts695314039/001/main.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts498915242/001/main.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a2136879483/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a672375377/001/main.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts118281188/001/main.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a357958274/001/main.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts644987174/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a2499742629/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts2397913529/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts317039858/001/main.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a921134405/001/main.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1071771235/001/main.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts1277228065/001/main.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a1884384893/001/main.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.17s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.19s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.20s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.26s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.21s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.24s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.24s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.21s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.16s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.08s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.10s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (0.10s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.09s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.09s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.569s
FAIL
```

Command: `cd /tmp/n7s-main && ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees|TestCapturedUnionCycleHeldOrRefused' -count=1 -v -timeout 40m`

Exit **1**; wall **32.74 seconds**.

| Pin | Overall | Native | Sanitized | Counted | JavaScript | Logged counts |
|---|---|---|---|---|---|---|
| class.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| class.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| object.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| object.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| by_value.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| by_value.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| released.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| released.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| nested.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| nested.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| alternating.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| alternating.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| graph.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| graph.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| cycle.ts | PASS | not reached | not reached | not reached | not reached | not reached |
| cycle.a | PASS | not reached | not reached | not reached | not reached | not reached |

`TestCapturedUnionCycleHeldOrRefused/cycle.ts`:
```text
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.ts3621076482/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

`TestCapturedUnionCycleHeldOrRefused/cycle.a`:
```text
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.a1968545644/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== RUN   TestCapturedUnionCycleHeldOrRefused
=== PAUSE TestCapturedUnionCycleHeldOrRefused
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== CONT  TestCapturedUnionCycleHeldOrRefused
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.a
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts1652147777/001/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.ts3621076482/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a1597717643/001/released.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.a1968545644/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
--- PASS: TestCapturedUnionCycleHeldOrRefused (0.00s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts (0.35s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a (0.39s)
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a42203199/001/graph.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts1881353140/001/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a530017961/001/object.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts1219716/001/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/graph.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts3683699167/001/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a3388657434/001/by_value.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/nested.ts
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a1291287912/001/class.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts151721920/001/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts869239816/001/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a1118372776/001/alternating.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a1014535085/001/nested.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts3253339898/001/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.30s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.39s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.44s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.31s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.31s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.28s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.33s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.43s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.19s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.14s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.15s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.14s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.12s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.14s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	1.066s
FAIL
```

## 4. Mutants

Each edit asserted that old text occurred exactly once, ran the two step-2 commands sequentially, restored the source, and passed `git diff --exit-code`. No mutant failed to compile. The maybe-boolean path assertion remained red in each lowerer run. Every red subtest and its complete diagnostic is preserved below; the expected stdout and test assertions are reproduced in the expectations appendix.

### M1 — The guard restored

```diff
M1 old text occurs exactly once
diff --git a/internal/lower/locals.go b/internal/lower/locals.go
index a6a057992..b1452f72d 100644
--- a/internal/lower/locals.go
+++ b/internal/lower/locals.go
@@ -201,7 +201,7 @@ func (l *lowering) localOf(identifier *ast.Node, symbol *ast.Symbol) (int, bool)
 		if declared := l.result.Locals[local]; declared.Placeholder != "" && declared.Captured && declared.Preallocated && declared.Ready == 0 && l.unlowerable == nil {
 			l.unlowerable = l.notYet(identifier, "an unset-capable capture before its declaration without a separate declaration-readiness cell")
 		}
-		if declared := l.result.Locals[local]; declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil {
+		if declared := l.result.Locals[local]; declared.Captured && slotless(declared.Type) && !(declared.Placeholder != "" && declared.Type == ir.Union) && l.unlowerable == nil {
 			// A Union is one counted heap reference whatever initializes it, a placeholder or a
 			// value (#n7sx9gz), and a cell holds one: each write fits its member into the Union, the
 			// cell releases what it held, and a read is the Union every closure shares. boolean |
```

Command: `go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat' -count=1 -v -timeout 30m`

Exit **1**; wall **32.24 seconds**.

| Test / subtest | Result |
|---|---|
| `TestCapturedMaybeBooleanStillStops` | FAIL |
| `TestCapturedUnionLowersToOneCell` | FAIL |
| `TestCapturedUnionLowersToOneCell/class.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/released.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/released.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/graph.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/by_value.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/object.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/by_value.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/class.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/object.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/alternating.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/alternating.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/nested.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/nested.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/graph.ts` | FAIL |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays` | PASS |

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== CONT  TestCapturedUnionLowersToOneCell
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedMaybeBooleanStillStops
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops1874793713/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
--- FAIL: TestCapturedMaybeBooleanStillStops (0.17s)
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts2537387842/001/main.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts2288572571/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a2436068689/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a2685670303/001/main.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts2373933999/001/main.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts2252197066/001/main.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a1071958003/001/main.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a3143989393/001/main.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a989811507/001/main.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3343693126/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts371543591/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts975938607/001/main.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a207445194/001/main.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts3159944966/001/main.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.01s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.17s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.20s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.23s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.26s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.18s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.14s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.20s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.13s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (0.18s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.10s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.11s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.09s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.11s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.13s)
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
--- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat (0.01s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only (0.17s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only (0.17s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function (0.05s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?. (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never (0.05s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.286s
FAIL
```

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees|TestCapturedUnionCycleHeldOrRefused' -count=1 -v -timeout 40m`

Exit **1**; wall **33.31 seconds**.

| Pin | Overall | Native | Sanitized | Counted | JavaScript | Logged counts |
|---|---|---|---|---|---|---|
| class.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| class.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| object.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| object.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| by_value.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| by_value.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| released.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| released.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| nested.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| nested.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| alternating.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| alternating.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| graph.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| graph.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| cycle.ts | PASS | not reached | not reached | not reached | not reached | not reached |
| cycle.a | PASS | not reached | not reached | not reached | not reached | not reached |

`TestCapturedUnionCycleHeldOrRefused/cycle.ts`:
```text
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.ts1437448532/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

`TestCapturedUnionCycleHeldOrRefused/cycle.a`:
```text
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.a3285124712/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== RUN   TestCapturedUnionCycleHeldOrRefused
=== PAUSE TestCapturedUnionCycleHeldOrRefused
=== CONT  TestCapturedUnionAgrees
=== CONT  TestCapturedUnionCycleHeldOrRefused
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.a
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.ts1437448532/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts1364981257/001/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts235550699/001/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.a3285124712/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
--- PASS: TestCapturedUnionCycleHeldOrRefused (0.00s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts (0.23s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a (0.27s)
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a2152535073/001/released.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/graph.ts
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts1702642365/001/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a1265850782/001/by_value.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts1556671035/001/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a2902307756/001/graph.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts4171739376/001/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/nested.ts
=== NAME  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a2722960833/001/alternating.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts69525463/001/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a1133297395/001/class.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a241025974/001/object.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts791079166/001/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a2349552362/001/nested.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.25s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.26s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.27s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.32s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.31s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.33s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.34s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.37s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.10s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.16s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.12s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.13s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.14s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.12s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.840s
FAIL
```

Restoration check: exit 0, no diff output.

### M2 — The cell held as a plain reference

```diff
M2 old text occurs exactly once
diff --git a/internal/lower/locals.go b/internal/lower/locals.go
index a6a057992..c6c466fdc 100644
--- a/internal/lower/locals.go
+++ b/internal/lower/locals.go
@@ -201,6 +201,9 @@ func (l *lowering) localOf(identifier *ast.Node, symbol *ast.Symbol) (int, bool)
 		if declared := l.result.Locals[local]; declared.Placeholder != "" && declared.Captured && declared.Preallocated && declared.Ready == 0 && l.unlowerable == nil {
 			l.unlowerable = l.notYet(identifier, "an unset-capable capture before its declaration without a separate declaration-readiness cell")
 		}
+		if declared := l.result.Locals[local]; declared.Captured && declared.Type == ir.Union && declared.Placeholder == "" {
+			l.result.Locals[local].Type = ir.Object
+		}
 		if declared := l.result.Locals[local]; declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil {
 			// A Union is one counted heap reference whatever initializes it, a placeholder or a
 			// value (#n7sx9gz), and a cell holds one: each write fits its member into the Union, the
```

| Oracle program | Lowered | Compiled | Ran | Printed against Node |
|---|---|---|---|---|
| class.ts | No | No | No | No output; lowering stopped before Node comparison |
| class.a | No | No | No | No output; lowering stopped before Node comparison |
| object.ts | No | No | No | No output; lowering stopped before Node comparison |
| object.a | No | No | No | No output; lowering stopped before Node comparison |
| by_value.ts | No | No | No | No output; lowering stopped before Node comparison |
| by_value.a | No | No | No | No output; lowering stopped before Node comparison |
| released.ts | No | No | No | No output; lowering stopped before Node comparison |
| released.a | No | No | No | No output; lowering stopped before Node comparison |
| nested.ts | No | No | No | No output; lowering stopped before Node comparison |
| nested.a | No | No | No | No output; lowering stopped before Node comparison |
| alternating.ts | No | No | No | No output; lowering stopped before Node comparison |
| alternating.a | No | No | No | No output; lowering stopped before Node comparison |
| graph.ts | No | No | No | No output; lowering stopped before Node comparison |
| graph.a | No | No | No | No output; lowering stopped before Node comparison |
| cycle.ts | No | No | No | No output; lowering stopped before Node comparison |
| cycle.a | No | No | No | No output; lowering stopped before Node comparison |

Cycle subtests still report PASS/refused through filename matching; all ordinary oracle subtests fail.

Command: `go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat' -count=1 -v -timeout 30m`

Exit **1**; wall **30.95 seconds**.

| Test / subtest | Result |
|---|---|
| `TestCapturedMaybeBooleanStillStops` | FAIL |
| `TestCapturedUnionLowersToOneCell` | FAIL |
| `TestCapturedUnionLowersToOneCell/class.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/released.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/alternating.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/alternating.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/nested.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/graph.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/nested.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/graph.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/object.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/class.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/by_value.a` | FAIL |
| `TestCapturedUnionLowersToOneCell/released.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/by_value.ts` | FAIL |
| `TestCapturedUnionLowersToOneCell/object.ts` | FAIL |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays` | PASS |

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedMaybeBooleanStillStops
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops923556422/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
--- FAIL: TestCapturedMaybeBooleanStillStops (0.17s)
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts2093907275/001/main.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a1101925778/001/main.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a2410292226/001/main.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts3068832888/001/main.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a2291359430/001/main.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts2229688489/001/main.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts1979929572/001/main.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a1954124790/001/main.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1843429908/001/main.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a3219438557/001/main.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a1174279188/001/main.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts3579923231/001/main.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts4189717080/001/main.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:36: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts2365039694/001/main.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.17s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.19s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.20s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.26s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.15s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.15s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.22s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.21s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (0.15s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.10s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.15s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.19s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.12s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.10s)
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
--- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat (0.01s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object (0.16s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only (0.16s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null (0.22s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal (0.21s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string (0.20s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number (0.23s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error (0.19s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?. (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set (0.15s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak (0.13s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays (0.08s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.414s
FAIL
```

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees|TestCapturedUnionCycleHeldOrRefused' -count=1 -v -timeout 40m`

Exit **1**; wall **29.07 seconds**.

| Pin | Overall | Native | Sanitized | Counted | JavaScript | Logged counts |
|---|---|---|---|---|---|---|
| class.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| class.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| object.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| object.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| by_value.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| by_value.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| released.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| released.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| nested.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| nested.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| alternating.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| alternating.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| graph.ts | FAIL | not reached | not reached | not reached | not reached | not reached |
| graph.a | FAIL | not reached | not reached | not reached | not reached | not reached |
| cycle.ts | PASS | not reached | not reached | not reached | not reached | not reached |
| cycle.a | PASS | not reached | not reached | not reached | not reached | not reached |

`TestCapturedUnionCycleHeldOrRefused/cycle.ts`:
```text
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.ts262994108/001/cycle.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
```

`TestCapturedUnionCycleHeldOrRefused/cycle.a`:
```text
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.a3797049178/001/cycle.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
```

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== RUN   TestCapturedUnionCycleHeldOrRefused
=== PAUSE TestCapturedUnionCycleHeldOrRefused
=== CONT  TestCapturedUnionAgrees
=== CONT  TestCapturedUnionCycleHeldOrRefused
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.a
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== CONT  TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a1452480319/001/alternating.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts3728800879/001/graph.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts4132139224/001/class.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/nested.ts
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a737510196/001/graph.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/released.a
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.ts262994108/001/cycle.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a1678518035/001/nested.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts646974086/001/nested.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts3432827658/001/alternating.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts1367018374/001/released.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a2940646513/001/released.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a3749201416/001/object.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.a
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts268279987/001/by_value.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a3382867827/001/by_value.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a3939755926/001/class.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:42: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts143551322/001/object.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.14s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.21s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.21s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.22s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.16s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.18s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.27s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.19s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.22s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.06s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.09s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.11s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.10s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.10s)
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a
    captured_union_test.go:70: refused: /tmp/adamic-gate/TestCapturedUnionCycleHeldOrRefusedcycle.a3797049178/001/cycle.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
--- PASS: TestCapturedUnionCycleHeldOrRefused (0.00s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts (0.24s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.608s
FAIL
```

Restoration check: exit 0, no diff output.

### M3 — The overwritten value kept

```diff
M3 old text occurs exactly once
diff --git a/internal/native/emit_locals.go b/internal/native/emit_locals.go
index 775b738bd..e9c8e39b2 100644
--- a/internal/native/emit_locals.go
+++ b/internal/native/emit_locals.go
@@ -60,7 +60,7 @@ func (e *emitter) store(local int, value string, owned bool) {
 		if _, framed := e.asyncSlots[local]; framed {
 			e.line("if (%s != NULL) adamic_release(%s);", old, old)
 		} else {
-			e.line("adamic_release(%s);", old)
+			e.line("(void)%s;", old)
 		}
 		return
 	}
```

| Pin (each .ts and .a separately has the same result) | Sanitized leak check | Counted build | Leaked heap values |
|---|---|---|---|
| class.ts, class.a | PASS | PASS | 0 |
| object.ts, object.a | PASS | PASS | 0 |
| by_value.ts, by_value.a | FAIL | FAIL | 3 |
| released.ts, released.a | FAIL | FAIL | 30 |
| nested.ts, nested.a | PASS | PASS | 0 |
| alternating.ts, alternating.a | FAIL | FAIL | 18 |
| graph.ts, graph.a | PASS | PASS | 0 |
| cycle.ts, cycle.a | PASS | PASS | 0 |

For every failing sanitized check, want is no leak diagnostic and successful comparable execution; got is the full LeakSanitizer report below. For every failing counted check, want is allocations equal frees and zero leaked heap values; got is logged below. Native and JavaScript stdout checks pass for every pin.

Command: `go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat' -count=1 -v -timeout 30m`

Exit **1**; wall **23.94 seconds**.

| Test / subtest | Result |
|---|---|
| `TestCapturedMaybeBooleanStillStops` | FAIL |
| `TestCapturedUnionLowersToOneCell` | PASS |
| `TestCapturedUnionLowersToOneCell/by_value.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/class.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/released.a` | PASS |
| `TestCapturedUnionLowersToOneCell/released.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/alternating.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/graph.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/alternating.a` | PASS |
| `TestCapturedUnionLowersToOneCell/graph.a` | PASS |
| `TestCapturedUnionLowersToOneCell/by_value.a` | PASS |
| `TestCapturedUnionLowersToOneCell/nested.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/object.a` | PASS |
| `TestCapturedUnionLowersToOneCell/object.ts` | PASS |
| `TestCapturedUnionLowersToOneCell/class.a` | PASS |
| `TestCapturedUnionLowersToOneCell/nested.a` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_\|_undefined` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions` | PASS |
| `TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array` | PASS |

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== CONT  TestCapturedUnionLowersToOneCell
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
=== RUN   TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== PAUSE TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== NAME  TestCapturedMaybeBooleanStillStops
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops4108828608/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
--- FAIL: TestCapturedMaybeBooleanStillStops (0.19s)
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only
--- PASS: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (0.15s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.16s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.22s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.24s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.13s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.25s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.19s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.09s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.09s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.10s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.10s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.14s)
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?.
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array
=== CONT  TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions
--- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat (0.01s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_inside_a_function (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_spread_seen_with_a_field_weak_in_one_view_only (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_result_weak_in_one_view_only (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_interpolating_an_object (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_callback_taking_arrays_given_tuples (0.12s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_seen_with_a_parameter_weak_in_one_view_only (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_boolean (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_object_seen_with_a_field_weak_in_one_view_only (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_function (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string_or_a_number (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/rethrowing_a_mutable_stored_Error (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_a_call_that_reaches_a_typed_array's_set (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_try_around_repeat (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_an_object_literal (0.04s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_generic_function_value_without_context (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_string (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_null (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_called_through_?. (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number_literal (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/throwing_a_number (0.11s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_function_value_capturing_what_it_initializes (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_given_to_a_function_value_taking_an_array (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/tuples_where_arrays_of_arrays_go (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_tuple_passed_where_an_array_goes (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_boolean_|_undefined (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/assigning_Array.from's_undefined (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_template_of_a_union_with_an_object (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_with_a_named_callback (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_array_of_targets_seen_as_an_array_of_Weak (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_without_a_callback (0.08s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_class_field_of_a_union (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/a_captured_boolean_|_undefined (0.07s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Math.random,_not_called (0.09s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/concat_on_an_array_of_arrays (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/an_arrow_returning_never (0.10s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/join_on_an_array_of_functions (0.06s)
    --- PASS: TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat/Array.from_of_an_array (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.224s
FAIL
```

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees|TestCapturedUnionCycleHeldOrRefused' -count=1 -v -timeout 40m`

Exit **1**; wall **34.03 seconds**.

| Pin | Overall | Native | Sanitized | Counted | JavaScript | Logged counts |
|---|---|---|---|---|---|---|
| class.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| class.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| object.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| object.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0 |
| by_value.ts | FAIL | PASS | FAIL | FAIL | PASS | captured_union_test.go:141: counts: allocations 35, frees 32, in regions 0 |
| by_value.a | FAIL | PASS | FAIL | FAIL | PASS | captured_union_test.go:141: counts: allocations 35, frees 32, in regions 0 |
| released.ts | FAIL | PASS | FAIL | FAIL | PASS | captured_union_test.go:141: counts: allocations 96, frees 66, in regions 0 |
| released.a | FAIL | PASS | FAIL | FAIL | PASS | captured_union_test.go:141: counts: allocations 96, frees 66, in regions 0 |
| nested.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0 |
| nested.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0 |
| alternating.ts | FAIL | PASS | FAIL | FAIL | PASS | captured_union_test.go:141: counts: allocations 54, frees 36, in regions 0 |
| alternating.a | FAIL | PASS | FAIL | FAIL | PASS | captured_union_test.go:141: counts: allocations 54, frees 36, in regions 0 |
| graph.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0 |
| graph.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0 |
| cycle.ts | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0 |
| cycle.a | PASS | PASS | PASS | PASS | PASS | captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0 |

`TestCapturedUnionCycleHeldOrRefused/cycle.ts`:
```text
    captured_union_test.go:67: held: graph cell true
```

`TestCapturedUnionCycleHeldOrRefused/cycle.a`:
```text
    captured_union_test.go:67: held: graph cell true
```

Complete command output, verbatim:

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== RUN   TestCapturedUnionCycleHeldOrRefused
=== PAUSE TestCapturedUnionCycleHeldOrRefused
=== CONT  TestCapturedUnionAgrees
=== CONT  TestCapturedUnionCycleHeldOrRefused
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a
=== PAUSE TestCapturedUnionCycleHeldOrRefused/cycle.a
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionCycleHeldOrRefused/cycle.a
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts
    captured_union_test.go:67: held: graph cell true
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a
    captured_union_test.go:67: held: graph cell true
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/native
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/counted
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/counted
=== NAME  TestCapturedUnionAgrees/alternating.a/sanitized
    captured_union_test.go:119: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12384==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x56007f4513a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56007f4bddcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56007f4bddcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x56007f494e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x56007f494ac0 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x56007f4513a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56007f4bddcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56007f4bddcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x56007f494e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x56007f494a84 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x56007f4513a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56007f4bddcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56007f4bddcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x56007f4513a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56007f49977f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== NAME  TestCapturedUnionAgrees/released.a/sanitized
    captured_union_test.go:119: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12383==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x559abcab23a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559abcb1f3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559abcb1f3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x559abcaf68cc in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:197:85
            #4 0x559abcaf5d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x559abcaf5d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x559abcaf5a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x559abcab23a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559abcb1f3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559abcb1f3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x559abcaf66a6 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:178:84
            #4 0x559abcaf5d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x559abcaf5d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x559abcaf5a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x559abcab23a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559abcb1f3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559abcb1f3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x559abcaf6431 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:145:80
            #4 0x559abcaf5d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x559abcaf5d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x559abcaf5a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x559abcab23a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559abcb1f3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559abcb1f3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x559abcab23a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559abcb1f3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559abcb1f3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x559abcab23a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559abcafadaf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a/counted
    captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.a/javascript
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts/counted
    captured_union_test.go:141: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionCycleHeldOrRefused/cycle.ts/javascript
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:135: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:141: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:135: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:141: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.60
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.62
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionCycleHeldOrRefused/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.74
=== NAME  TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.76
=== NAME  TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.78
--- PASS: TestCapturedUnionCycleHeldOrRefused (0.00s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a (10.88s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/native (0.44s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/sanitized (1.35s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/counted (0.52s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.a/javascript (1.18s)
    --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts (10.99s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/native (0.45s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/sanitized (1.31s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/counted (0.66s)
        --- PASS: TestCapturedUnionCycleHeldOrRefused/cycle.ts/javascript (1.13s)
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== NAME  TestCapturedUnionAgrees/alternating.ts/sanitized
    captured_union_test.go:119: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12810==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5615442003a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56154426cdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56154426cdcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x561544243e1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x561544243ac0 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5615442003a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56154426cdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56154426cdcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x561544243e1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x561544243a84 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x5615442003a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56154426cdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56154426cdcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x5615442003a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56154424877f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:135: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:141: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:141: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== CONT  TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== NAME  TestCapturedUnionAgrees/by_value.ts/sanitized
    captured_union_test.go:119: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12957==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x5581f97923a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5581f97fed2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5581f97fed2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5581f97923a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5581f97fed2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5581f97fed2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5581f97d602d in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:136:2
            #4 0x5581f97d602d in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x5581f97d5a6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5581f97923a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5581f97fed2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5581f97fed2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5581f97d5def in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:91:6
            #4 0x5581f97d5def in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x5581f97d5a6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== NAME  TestCapturedUnionAgrees/released.ts/sanitized
    captured_union_test.go:119: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12969==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55cee5e313a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55cee5e9e3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55cee5e9e3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55cee5e758cc in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:197:85
            #4 0x55cee5e74d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55cee5e74d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55cee5e74a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55cee5e313a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55cee5e9e3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55cee5e9e3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55cee5e756a6 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:178:84
            #4 0x55cee5e74d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55cee5e74d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55cee5e74a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55cee5e313a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55cee5e9e3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55cee5e9e3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55cee5e75431 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:145:80
            #4 0x55cee5e74d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55cee5e74d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55cee5e74a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x55cee5e313a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55cee5e9e3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55cee5e9e3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x55cee5e313a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55cee5e9e3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55cee5e9e3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x55cee5e313a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55cee5e79daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/sanitized
    captured_union_test.go:119: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12979==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x55d776c353a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d776ca1d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d776ca1d2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55d776c353a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d776ca1d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d776ca1d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55d776c7902d in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:136:2
            #4 0x55d776c7902d in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x55d776c78a6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55d776c353a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d776ca1d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d776ca1d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55d776c78def in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:91:6
            #4 0x55d776c78def in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x55d776c78a6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:135: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:141: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:135: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:141: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:135: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:141: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== CONT  TestCapturedUnionAgrees/object.a
=== CONT  TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.03
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:141: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:141: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/class.ts (10.91s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.50s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (1.33s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.80s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (0.95s)
    --- FAIL: TestCapturedUnionAgrees/released.a (11.07s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.47s)
        --- FAIL: TestCapturedUnionAgrees/released.a/sanitized (1.54s)
        --- FAIL: TestCapturedUnionAgrees/released.a/counted (0.49s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (1.14s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (11.08s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.46s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/sanitized (1.52s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/counted (0.49s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (1.18s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (1.55s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.31s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/sanitized (0.46s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/counted (0.26s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/nested.a (1.69s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.31s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.45s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (1.76s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.37s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (0.49s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.32s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (0.16s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (1.71s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.33s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/sanitized (0.69s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/counted (0.30s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (1.70s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.35s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/sanitized (0.63s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/counted (0.32s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (1.59s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.41s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/sanitized (0.53s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/counted (0.27s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/object.ts (1.86s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.48s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (0.53s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.46s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/object.a (1.90s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.40s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (0.73s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.37s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/class.a (1.91s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.37s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (0.74s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.40s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.07s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (8.39s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (2.95s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (2.75s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (2.09s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.10s)
    --- PASS: TestCapturedUnionAgrees/graph.a (8.42s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (2.95s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (2.79s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (2.11s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.08s)
FAIL
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	19.583s
FAIL
```

Restoration check: exit 0, no diff output.

## 4b. Full IR suite

Ran alone at detached HEAD `5eeac42cf8c0037c627f32b975dd254cbfd20c6b`, with all step-4 mutants restored and no other test or build running. This is a supplemental measurement requested after the original steps 5–7 had completed; it is placed here in the requested report order, rather than represented as having run before the original step 5.

Command: `go test ./internal/ir -count=1 -timeout 30m > /tmp/ir.log 2>&1`

Exit **0**; wall **127.97 seconds**.

Failing tests and first error lines: **none**. No main reruns were needed because the suite passed.

Complete `/tmp/ir.log`, verbatim:

```text
ok  	github.com/system-inc/adamic/internal/ir	88.819s
```

## 5. Full lowerer suite

Command: `go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1`

Exit **1**; wall **1562.91 seconds**.

```text
--- FAIL: TestCapturedMaybeBooleanStillStops (0.11s)
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops2417082272/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1558.402s
FAIL
```

| Failing test | First error line | Attribution | Also fails on main? |
|---|---|---|---|
| TestCapturedMaybeBooleanStillStops | captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops2417082272/001/main.a:4:3: stage 0 can't lower a boolean \| undefined variable a function value captures yet, want main's stop at 4:3 | Environment/path-sensitive assertion | Yes, isolated rerun |

Command: `cd /tmp/n7s-main && go test ./internal/lower -run '^TestCapturedMaybeBooleanStillStops$' -count=1 -v -timeout 30m`

Exit **1**; wall **7.32 seconds**.

Complete command output, verbatim:

```text
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedMaybeBooleanStillStops
    captured_union_test.go:75: got /tmp/adamic-gate/TestCapturedMaybeBooleanStillStops4026591422/001/main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet, want main's stop at 4:3
--- FAIL: TestCapturedMaybeBooleanStillStops (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.079s
FAIL
```

## 6. Full oracle suite

Command: `go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1`

Exit **1**; wall **3888.54 seconds**.

```text
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestAssignmentProofWiderTargetMutant (0.21s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.15s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.22s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.14s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.17s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.20s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.24s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.10s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.09s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.06s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.09s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.09s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestStdoutColumnsRefusesWASI (0.87s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestFamiliesCatchAPlantedCase_CountsAreRecorded (90.02s)
    families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s:
        === RUN   TestCountsAreRecorded_000
        === PAUSE TestCountsAreRecorded_000
        === RUN   TestCountsAreRecorded_001
        === PAUSE TestCountsAreRecorded_001
        === RUN   TestCountsAreRecorded_002
        === PAUSE TestCountsAreRecorded_002
        === RUN   TestCountsAreRecorded_003
        === PAUSE TestCountsAreRecorded_003
        === RUN   TestCountsAreRecorded_004
        === PAUSE TestCountsAreRecorded_004
        === RUN   TestCountsAreRecorded_005
        === PAUSE TestCountsAreRecorded_005
        === RUN   TestCountsAreRecorded_006
        === PAUSE TestCountsAreRecorded_006
        === RUN   TestCountsAreRecorded_007
        === PAUSE TestCountsAreRecorded_007
        === RUN   TestCountsAreRecorded_008
        === PAUSE TestCountsAreRecorded_008
        === RUN   TestCountsAreRecorded_009
        === PAUSE TestCountsAreRecorded_009
        === RUN   TestCountsAreRecorded_010
        === PAUSE TestCountsAreRecorded_010
        === RUN   TestCountsAreRecorded_011
        === PAUSE TestCountsAreRecorded_011
        === RUN   TestCountsAreRecorded_012
        === PAUSE TestCountsAreRecorded_012
        === RUN   TestCountsAreRecorded_013
        === PAUSE TestCountsAreRecorded_013
        === RUN   TestCountsAreRecorded_014
        === PAUSE TestCountsAreRecorded_014
        === RUN   TestCountsAreRecorded_015
        === PAUSE TestCountsAreRecorded_015
        === RUN   TestCountsAreRecorded_016
        === PAUSE TestCountsAreRecorded_016
        === RUN   TestCountsAreRecorded_017
        === PAUSE TestCountsAreRecorded_017
        === RUN   TestCountsAreRecorded_018
        === PAUSE TestCountsAreRecorded_018
        === RUN   TestCountsAreRecorded_019
        === PAUSE TestCountsAreRecorded_019
        === RUN   TestCountsAreRecorded_020
        === PAUSE TestCountsAreRecorded_020
        === RUN   TestCountsAreRecorded_021
        === PAUSE TestCountsAreRecorded_021
        === RUN   TestCountsAreRecorded_022
        === PAUSE TestCountsAreRecorded_022
        === RUN   TestCountsAreRecorded_023
        === PAUSE TestCountsAreRecorded_023
        === RUN   TestCountsAreRecorded_024
        === PAUSE TestCountsAreRecorded_024
        === RUN   TestCountsAreRecorded_025
        === PAUSE TestCountsAreRecorded_025
        === RUN   TestCountsAreRecorded_026
        === PAUSE TestCountsAreRecorded_026
        === RUN   TestCountsAreRecorded_027
        === PAUSE TestCountsAreRecorded_027
        === RUN   TestCountsAreRecorded_028
        === PAUSE TestCountsAreRecorded_028
        === RUN   TestCountsAreRecorded_029
        === PAUSE TestCountsAreRecorded_029
        === RUN   TestCountsAreRecorded_030
        === PAUSE TestCountsAreRecorded_030
        === RUN   TestCountsAreRecorded_031
        === PAUSE TestCountsAreRecorded_031
        === RUN   TestCountsAreRecorded_little_streams
        === PAUSE TestCountsAreRecorded_little_streams
        === RUN   TestCountsAreRecorded_lots_streams
        === PAUSE TestCountsAreRecorded_lots_streams
        === RUN   TestCountsAreRecorded_stderr_only
        === PAUSE TestCountsAreRecorded_stderr_only
        === RUN   TestCountsAreRecorded_panic_edges
        === PAUSE TestCountsAreRecorded_panic_edges
        === RUN   TestCountsAreRecorded_read_after_print
        === PAUSE TestCountsAreRecorded_read_after_print
        === RUN   TestCountsAreRecorded_async_counts_payload
        === PAUSE TestCountsAreRecorded_async_counts_payload
        === RUN   TestCountsAreRecorded_async_counts_stable
        === PAUSE TestCountsAreRecorded_async_counts_stable
        === RUN   TestCountsAreRecorded_async_counts_reassigned
        === PAUSE TestCountsAreRecorded_async_counts_reassigned
        === RUN   TestCountsAreRecorded_async_counts_captured
        === PAUSE TestCountsAreRecorded_async_counts_captured
        === RUN   TestCountsAreRecorded_async_counts_expression_write
        === PAUSE TestCountsAreRecorded_async_counts_expression_write
        === RUN   TestCountsAreRecorded_async_counts_parameter
        === PAUSE TestCountsAreRecorded_async_counts_parameter
        === RUN   TestCountsAreRecorded_async_counts_thrown
        === PAUSE TestCountsAreRecorded_async_counts_thrown
        === RUN   TestCountsAreRecorded_async_counts_root_error
        === PAUSE TestCountsAreRecorded_async_counts_root_error
        === RUN   TestCountsAreRecorded_async_catch_return
        === PAUSE TestCountsAreRecorded_async_catch_return
        === RUN   TestCountsAreRecorded_async_catches
        === PAUSE TestCountsAreRecorded_async_catches
        === RUN   TestCountsAreRecorded_async_closures
        === PAUSE TestCountsAreRecorded_async_closures
        === RUN   TestCountsAreRecorded_async_conditions
        === PAUSE TestCountsAreRecorded_async_conditions
        === RUN   TestCountsAreRecorded_async_control
        === PAUSE TestCountsAreRecorded_async_control
        === RUN   TestCountsAreRecorded_async_escaped_capture
        === PAUSE TestCountsAreRecorded_async_escaped_capture
        === RUN   TestCountsAreRecorded_async_frame_capture_safe
        === PAUSE TestCountsAreRecorded_async_frame_capture_safe
        === RUN   TestCountsAreRecorded_async_fresh_holder
        === PAUSE TestCountsAreRecorded_async_fresh_holder
        === RUN   TestCountsAreRecorded_async_live_values
        === PAUSE TestCountsAreRecorded_async_live_values
        === RUN   TestCountsAreRecorded_async_methods
        === PAUSE TestCountsAreRecorded_async_methods
        === RUN   TestCountsAreRecorded_async_mutation
        === PAUSE TestCountsAreRecorded_async_mutation
        === RUN   TestCountsAreRecorded_async_operand_order
        === PAUSE TestCountsAreRecorded_async_operand_order
        === RUN   TestCountsAreRecorded_async_promise_expression
        === PAUSE TestCountsAreRecorded_async_promise_expression
        === RUN   TestCountsAreRecorded_async_receiver
        === PAUSE TestCountsAreRecorded_async_receiver
        === RUN   TestCountsAreRecorded_async_snapshots
        === PAUSE TestCountsAreRecorded_async_snapshots
        === RUN   TestCountsAreRecorded_async_void_value
        === PAUSE TestCountsAreRecorded_async_void_value
        === RUN   TestCountsAreRecorded_typed_arrays_integer
        === PAUSE TestCountsAreRecorded_typed_arrays_integer
        === RUN   TestCountsAreRecorded_typed_arrays_integer_stop
        === PAUSE TestCountsAreRecorded_typed_arrays_integer_stop
        === RUN   TestCountsAreRecorded_typed_arrays_sort_uint8array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_uint8array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_uint16array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_uint16array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_int32array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_int32array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_float64array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_float64array
        === RUN   TestCountsAreRecorded_counters_shapes
        === PAUSE TestCountsAreRecorded_counters_shapes
        === RUN   TestCountsAreRecorded_counters_square_jump
        === PAUSE TestCountsAreRecorded_counters_square_jump
        === RUN   TestCountsAreRecorded_release_fast_graph
        === PAUSE TestCountsAreRecorded_release_fast_graph
        === RUN   TestCountsAreRecorded_release_fast_shared
        === PAUSE TestCountsAreRecorded_release_fast_shared
        === RUN   TestCountsAreRecorded_flat
        === PAUSE TestCountsAreRecorded_flat
        === RUN   TestCountsAreRecorded_array
        === PAUSE TestCountsAreRecorded_array
        === RUN   TestCountsAreRecorded_async_loop_body_capture
        === PAUSE TestCountsAreRecorded_async_loop_body_capture
        === RUN   TestCountsAreRecorded_async_finally_completion
        === PAUSE TestCountsAreRecorded_async_finally_completion
        === RUN   TestCountsAreRecorded_async_labels
        === PAUSE TestCountsAreRecorded_async_labels
        === RUN   TestCountsAreRecorded_async_labels_finally
        === PAUSE TestCountsAreRecorded_async_labels_finally
        === RUN   TestCountsAreRecorded_async_labels_block_finally
        === PAUSE TestCountsAreRecorded_async_labels_block_finally
        === RUN   TestCountsAreRecorded_label_block_edges
        === PAUSE TestCountsAreRecorded_label_block_edges
        === RUN   TestCountsAreRecorded_async_iterables_cells_labels
        === PAUSE TestCountsAreRecorded_async_iterables_cells_labels
        === RUN   TestCountsAreRecorded_label_block_break_sync
        === PAUSE TestCountsAreRecorded_label_block_break_sync
        === RUN   TestCountsAreRecorded_label_block_continue_sync
        === PAUSE TestCountsAreRecorded_label_block_continue_sync
        === RUN   TestCountsAreRecorded_label_block_switch_break_sync
        === PAUSE TestCountsAreRecorded_label_block_switch_break_sync
        === RUN   TestCountsAreRecorded_label_block_outer_continue_sync
        === PAUSE TestCountsAreRecorded_label_block_outer_continue_sync
        === RUN   TestCountsAreRecorded_label_block_break_async
        === PAUSE TestCountsAreRecorded_label_block_break_async
        === RUN   TestCountsAreRecorded_label_block_continue_async
        === PAUSE TestCountsAreRecorded_label_block_continue_async
        === RUN   TestCountsAreRecorded_label_block_switch_break_async
        === PAUSE TestCountsAreRecorded_label_block_switch_break_async
        === RUN   TestCountsAreRecorded_label_block_outer_continue_async
        === PAUSE TestCountsAreRecorded_label_block_outer_continue_async
        === RUN   TestCountsAreRecorded_async_finally_return
        === PAUSE TestCountsAreRecorded_async_finally_return
        === RUN   TestCountsAreRecorded_async_finally_throw
        === PAUSE TestCountsAreRecorded_async_finally_throw
        === RUN   TestCountsAreRecorded_async_finally_break
        === PAUSE TestCountsAreRecorded_async_finally_break
        === RUN   TestCountsAreRecorded_async_finally_continue
        === PAUSE TestCountsAreRecorded_async_finally_continue
        === RUN   TestCountsAreRecorded_async_finally_void_return
        === PAUSE TestCountsAreRecorded_async_finally_void_return
        === RUN   TestCountsAreRecorded_async_finally_snapshot
        === PAUSE TestCountsAreRecorded_async_finally_snapshot
        === RUN   TestCountsAreRecorded_async_gap_label_admission
        === PAUSE TestCountsAreRecorded_async_gap_label_admission
        === RUN   TestCountsAreRecorded_async_gap_catch_admission
        === PAUSE TestCountsAreRecorded_async_gap_catch_admission
        === RUN   TestCountsAreRecorded_async_gap_finally_admission
        === PAUSE TestCountsAreRecorded_async_gap_finally_admission
        === RUN   TestCountsAreRecorded_async_gap_finally_return_admission
        === PAUSE TestCountsAreRecorded_async_gap_finally_return_admission
        === RUN   TestCountsAreRecorded_async_iterable_string_ascii_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_ascii_function
        === RUN   TestCountsAreRecorded_async_for_of_string
        === PAUSE TestCountsAreRecorded_async_for_of_string
        === RUN   TestCountsAreRecorded_async_for_of_map
        === PAUSE TestCountsAreRecorded_async_for_of_map
        === RUN   TestCountsAreRecorded_async_for_of_collections_mutation
        === PAUSE TestCountsAreRecorded_async_for_of_collections_mutation
        === RUN   TestCountsAreRecorded_async_iterable_string_ascii_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_ascii_module
        === RUN   TestCountsAreRecorded_async_iterable_string_emoji_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_emoji_function
        === RUN   TestCountsAreRecorded_async_iterable_string_emoji_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_emoji_module
        === RUN   TestCountsAreRecorded_async_iterable_string_surrogates_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_surrogates_function
        === RUN   TestCountsAreRecorded_async_iterable_string_surrogates_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_surrogates_module
        === RUN   TestCountsAreRecorded_async_iterable_map_entries_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_entries_function
        === RUN   TestCountsAreRecorded_async_iterable_map_entries_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_entries_module
        === RUN   TestCountsAreRecorded_async_iterable_map_keys_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_keys_function
        === RUN   TestCountsAreRecorded_async_iterable_map_keys_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_keys_module
        === RUN   TestCountsAreRecorded_async_iterable_map_values_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_values_function
        === RUN   TestCountsAreRecorded_async_iterable_map_values_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_values_module
        === RUN   TestCountsAreRecorded_async_iterable_set_values_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_values_function
        === RUN   TestCountsAreRecorded_async_iterable_set_values_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_values_module
        === RUN   TestCountsAreRecorded_async_iterable_set_entries_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_entries_function
        === RUN   TestCountsAreRecorded_async_iterable_set_entries_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_entries_module
        === RUN   TestCountsAreRecorded_async_iterable_map_clear_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_clear_function
        === RUN   TestCountsAreRecorded_async_iterable_map_clear_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_clear_module
        === RUN   TestCountsAreRecorded_async_iterable_map_readd_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_readd_function
        === RUN   TestCountsAreRecorded_async_iterable_map_readd_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_readd_module
        === RUN   TestCountsAreRecorded_async_iterable_map_rehash_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_rehash_function
        === RUN   TestCountsAreRecorded_async_iterable_map_rehash_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_rehash_module
        === RUN   TestCountsAreRecorded_async_iterable_nested_maps_function
        === PAUSE TestCountsAreRecorded_async_iterable_nested_maps_function
        === RUN   TestCountsAreRecorded_async_iterable_nested_maps_module
        === PAUSE TestCountsAreRecorded_async_iterable_nested_maps_module
        === RUN   TestCountsAreRecorded_async_iterable_map_break_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_break_function
        === RUN   TestCountsAreRecorded_async_iterable_map_break_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_break_module
        === RUN   TestCountsAreRecorded_async_iterable_map_return_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_return_function
        === RUN   TestCountsAreRecorded_async_iterable_map_return_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_return_module
        === RUN   TestCountsAreRecorded_async_iterable_string_break_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_break_function
        === RUN   TestCountsAreRecorded_async_iterable_string_break_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_break_module
        === RUN   TestCountsAreRecorded_async_iterable_set_return_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_return_function
        === RUN   TestCountsAreRecorded_async_iterable_set_return_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_return_module
        === RUN   TestCountsAreRecorded_async_for_of_string_admission
        === PAUSE TestCountsAreRecorded_async_for_of_string_admission
        === RUN   TestCountsAreRecorded_async_top_catch_binding
        === PAUSE TestCountsAreRecorded_async_top_catch_binding
        === RUN   TestCountsAreRecorded_async_top_catch_after_await
        === PAUSE TestCountsAreRecorded_async_top_catch_after_await
        === RUN   TestCountsAreRecorded_async_top_loop
        === PAUSE TestCountsAreRecorded_async_top_loop
        === RUN   TestCountsAreRecorded_async_for_of_body
        === PAUSE TestCountsAreRecorded_async_for_of_body
        === RUN   TestCountsAreRecorded_async_for_of_mutation
        === PAUSE TestCountsAreRecorded_async_for_of_mutation
        === RUN   TestCountsAreRecorded_async_for_of_return
        === PAUSE TestCountsAreRecorded_async_for_of_return
        === RUN   TestCountsAreRecorded_async_switch_body
        === PAUSE TestCountsAreRecorded_async_switch_body
        === RUN   TestCountsAreRecorded_async_loop_cells
        === PAUSE TestCountsAreRecorded_async_loop_cells
        === RUN   TestCountsAreRecorded_async_for_of_capture
        === PAUSE TestCountsAreRecorded_async_for_of_capture
        === RUN   TestCountsAreRecorded_async_for_of_cells
        === PAUSE TestCountsAreRecorded_async_for_of_cells
        === RUN   TestCountsAreRecorded_async_for_of_body_cells
        === PAUSE TestCountsAreRecorded_async_for_of_body_cells
        === RUN   TestCountsAreRecorded_async_for_of_break_cells
        === PAUSE TestCountsAreRecorded_async_for_of_break_cells
        === RUN   TestCountsAreRecorded_async_coverage_discard
        === PAUSE TestCountsAreRecorded_async_coverage_discard
        === RUN   TestCountsAreRecorded_async_coverage_parameters
        === PAUSE TestCountsAreRecorded_async_coverage_parameters
        === RUN   TestCountsAreRecorded_async_coverage_reject_eager
        === PAUSE TestCountsAreRecorded_async_coverage_reject_eager
        === RUN   TestCountsAreRecorded_async_coverage_reject_empty
        === PAUSE TestCountsAreRecorded_async_coverage_reject_empty
        === RUN   TestCountsAreRecorded_async_coverage_reject_nested
        === PAUSE TestCountsAreRecorded_async_coverage_reject_nested
        === RUN   TestCountsAreRecorded_async_coverage_typeof
        === PAUSE TestCountsAreRecorded_async_coverage_typeof
        === RUN   TestCountsAreRecorded_async_coverage_unions
        === PAUSE TestCountsAreRecorded_async_coverage_unions
        === RUN   TestCountsAreRecorded_async_coverage_values
        === PAUSE TestCountsAreRecorded_async_coverage_values
        === RUN   TestCountsAreRecorded_async_nested
        === PAUSE TestCountsAreRecorded_async_nested
        === RUN   TestCountsAreRecorded_async_plain
        === PAUSE TestCountsAreRecorded_async_plain
        === RUN   TestCountsAreRecorded_async_three
        === PAUSE TestCountsAreRecorded_async_three
        === RUN   TestCountsAreRecorded_async_throw
        === PAUSE TestCountsAreRecorded_async_throw
        === RUN   TestCountsAreRecorded_async_typeof
        === PAUSE TestCountsAreRecorded_async_typeof
        === RUN   TestCountsAreRecorded_slab_sizes
        === PAUSE TestCountsAreRecorded_slab_sizes
        === RUN   TestCountsAreRecorded_reuse_handover_fresh
        === PAUSE TestCountsAreRecorded_reuse_handover_fresh
        === RUN   TestCountsAreRecorded_async_iterable_map_boolean_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_boolean_function
        === RUN   TestCountsAreRecorded_async_iterable_set_boolean_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_boolean_function
        === CONT  TestCountsAreRecorded_000
        --- PASS: TestCountsAreRecorded_000 (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_boolean_function
        === CONT  TestCountsAreRecorded_label_block_continue_sync
        === CONT  TestCountsAreRecorded_async_iterable_map_rehash_module
        === CONT  TestCountsAreRecorded_async_counts_captured
        --- PASS: TestCountsAreRecorded_async_iterable_set_boolean_function (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_reassigned
        --- PASS: TestCountsAreRecorded_label_block_continue_sync (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_stable
        --- PASS: TestCountsAreRecorded_async_counts_captured (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_payload
        --- PASS: TestCountsAreRecorded_async_counts_reassigned (0.01s)
        === CONT  TestCountsAreRecorded_read_after_print
        --- PASS: TestCountsAreRecorded_async_counts_payload (0.02s)
        === CONT  TestCountsAreRecorded_stderr_only
        --- PASS: TestCountsAreRecorded_read_after_print (0.01s)
        === CONT  TestCountsAreRecorded_lots_streams
        --- PASS: TestCountsAreRecorded_stderr_only (0.00s)
        === CONT  TestCountsAreRecorded_little_streams
        --- PASS: TestCountsAreRecorded_async_counts_stable (0.01s)
        === CONT  TestCountsAreRecorded_panic_edges
        --- PASS: TestCountsAreRecorded_panic_edges (0.00s)
        === CONT  TestCountsAreRecorded_031
        --- PASS: TestCountsAreRecorded_lots_streams (0.01s)
        === CONT  TestCountsAreRecorded_030
        --- PASS: TestCountsAreRecorded_030 (0.00s)
        === CONT  TestCountsAreRecorded_029
        --- PASS: TestCountsAreRecorded_little_streams (0.01s)
        === CONT  TestCountsAreRecorded_028
        --- PASS: TestCountsAreRecorded_029 (0.00s)
        === CONT  TestCountsAreRecorded_027
        --- PASS: TestCountsAreRecorded_async_iterable_map_rehash_module (0.05s)
        === CONT  TestCountsAreRecorded_026
        --- PASS: TestCountsAreRecorded_027 (0.00s)
        === CONT  TestCountsAreRecorded_025
        --- PASS: TestCountsAreRecorded_026 (0.00s)
        === CONT  TestCountsAreRecorded_024
        --- PASS: TestCountsAreRecorded_028 (0.01s)
        === CONT  TestCountsAreRecorded_023
        --- PASS: TestCountsAreRecorded_025 (0.02s)
        === CONT  TestCountsAreRecorded_022
        --- PASS: TestCountsAreRecorded_022 (0.00s)
        --- PASS: TestCountsAreRecorded_024 (0.03s)
        === CONT  TestCountsAreRecorded_020
        === CONT  TestCountsAreRecorded_021
        --- PASS: TestCountsAreRecorded_031 (0.05s)
        === CONT  TestCountsAreRecorded_019
        --- PASS: TestCountsAreRecorded_019 (0.01s)
        === CONT  TestCountsAreRecorded_017
        --- PASS: TestCountsAreRecorded_023 (0.05s)
        === CONT  TestCountsAreRecorded_016
        --- PASS: TestCountsAreRecorded_016 (0.00s)
        === CONT  TestCountsAreRecorded_015
        --- PASS: TestCountsAreRecorded_020 (0.01s)
        === CONT  TestCountsAreRecorded_018
        --- PASS: TestCountsAreRecorded_017 (0.02s)
        === CONT  TestCountsAreRecorded_014
        --- PASS: TestCountsAreRecorded_021 (0.03s)
        === CONT  TestCountsAreRecorded_013
        --- PASS: TestCountsAreRecorded_015 (0.01s)
        === CONT  TestCountsAreRecorded_012
        --- PASS: TestCountsAreRecorded_014 (0.01s)
        --- PASS: TestCountsAreRecorded_013 (0.01s)
        === CONT  TestCountsAreRecorded_010
        --- PASS: TestCountsAreRecorded_010 (0.00s)
        === CONT  TestCountsAreRecorded_009
        === CONT  TestCountsAreRecorded_011
        --- PASS: TestCountsAreRecorded_018 (0.01s)
        === CONT  TestCountsAreRecorded_008
        --- PASS: TestCountsAreRecorded_009 (0.00s)
        === CONT  TestCountsAreRecorded_007
        --- PASS: TestCountsAreRecorded_012 (0.01s)
        === CONT  TestCountsAreRecorded_006
        --- PASS: TestCountsAreRecorded_007 (0.01s)
        === CONT  TestCountsAreRecorded_005
        --- PASS: TestCountsAreRecorded_008 (0.02s)
        === CONT  TestCountsAreRecorded_004
        --- PASS: TestCountsAreRecorded_006 (0.01s)
        === CONT  TestCountsAreRecorded_003
        --- PASS: TestCountsAreRecorded_011 (0.02s)
        --- PASS: TestCountsAreRecorded_005 (0.01s)
        --- PASS: TestCountsAreRecorded_004 (0.01s)
        === CONT  TestCountsAreRecorded_002
        === CONT  TestCountsAreRecorded_001
        === CONT  TestCountsAreRecorded_async_for_of_capture
        --- PASS: TestCountsAreRecorded_async_for_of_capture (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_boolean_function
        --- PASS: TestCountsAreRecorded_001 (0.01s)
        --- PASS: TestCountsAreRecorded_async_iterable_map_boolean_function (0.00s)
        === CONT  TestCountsAreRecorded_slab_sizes
        --- PASS: TestCountsAreRecorded_002 (0.01s)
        --- PASS: TestCountsAreRecorded_slab_sizes (0.00s)
        === CONT  TestCountsAreRecorded_async_throw
        === CONT  TestCountsAreRecorded_reuse_handover_fresh
        === CONT  TestCountsAreRecorded_async_typeof
        --- PASS: TestCountsAreRecorded_async_throw (0.00s)
        === CONT  TestCountsAreRecorded_async_three
        --- PASS: TestCountsAreRecorded_reuse_handover_fresh (0.01s)
        --- PASS: TestCountsAreRecorded_async_three (0.01s)
        === CONT  TestCountsAreRecorded_async_plain
        === CONT  TestCountsAreRecorded_async_nested
        --- PASS: TestCountsAreRecorded_async_typeof (0.01s)
        --- PASS: TestCountsAreRecorded_async_nested (0.00s)
        === CONT  TestCountsAreRecorded_async_coverage_unions
        --- PASS: TestCountsAreRecorded_async_plain (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_typeof
        === CONT  TestCountsAreRecorded_async_coverage_values
        --- PASS: TestCountsAreRecorded_async_coverage_unions (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_nested
        --- PASS: TestCountsAreRecorded_async_coverage_typeof (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_empty
        --- PASS: TestCountsAreRecorded_async_coverage_reject_nested (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_parameters
        === CONT  TestCountsAreRecorded_async_coverage_reject_eager
        --- PASS: TestCountsAreRecorded_async_coverage_parameters (0.00s)
        === CONT  TestCountsAreRecorded_async_coverage_discard
        --- PASS: TestCountsAreRecorded_async_coverage_reject_eager (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_break_cells
        --- PASS: TestCountsAreRecorded_async_coverage_reject_empty (0.00s)
        --- PASS: TestCountsAreRecorded_async_coverage_values (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_body_cells
        --- PASS: TestCountsAreRecorded_async_coverage_discard (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_cells
        --- PASS: TestCountsAreRecorded_async_for_of_break_cells (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_string_ascii_module
        --- PASS: TestCountsAreRecorded_async_for_of_cells (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_rehash_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_ascii_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_readd_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_rehash_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_readd_function
        --- PASS: TestCountsAreRecorded_async_for_of_body_cells (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_clear_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_readd_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_clear_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_clear_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_set_entries_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_clear_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_entries_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_entries_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_values_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_readd_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_set_values_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_values_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_values_module
        --- PASS: TestCountsAreRecorded_async_iterable_set_entries_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_values_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_values_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_keys_module
        --- PASS: TestCountsAreRecorded_async_iterable_set_values_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_keys_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_keys_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_entries_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_values_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_entries_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_entries_function (0.00s)
        --- PASS: TestCountsAreRecorded_async_iterable_map_entries_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_surrogates_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_keys_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_emoji_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_surrogates_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_emoji_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_emoji_function (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_void_return
        --- PASS: TestCountsAreRecorded_async_iterable_string_emoji_module (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_collections_mutation
        === CONT  TestCountsAreRecorded_async_iterable_string_surrogates_module
        === CONT  TestCountsAreRecorded_async_for_of_map
        --- PASS: TestCountsAreRecorded_async_iterable_string_surrogates_module (0.00s)
        --- PASS: TestCountsAreRecorded_async_finally_void_return (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_string
        --- PASS: TestCountsAreRecorded_async_for_of_string (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_ascii_function
        --- PASS: TestCountsAreRecorded_async_for_of_map (0.01s)
        === CONT  TestCountsAreRecorded_async_gap_finally_return_admission
        --- PASS: TestCountsAreRecorded_async_iterable_string_ascii_function (0.00s)
        === CONT  TestCountsAreRecorded_async_gap_finally_admission
        --- PASS: TestCountsAreRecorded_async_for_of_collections_mutation (0.01s)
        === CONT  TestCountsAreRecorded_async_gap_catch_admission
        --- PASS: TestCountsAreRecorded_async_gap_finally_admission (0.01s)
        === CONT  TestCountsAreRecorded_async_gap_label_admission
        --- PASS: TestCountsAreRecorded_async_gap_finally_return_admission (0.01s)
        === CONT  TestCountsAreRecorded_async_finally_snapshot
        --- PASS: TestCountsAreRecorded_async_gap_label_admission (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_set_return_module
        --- PASS: TestCountsAreRecorded_async_finally_snapshot (0.01s)
        === CONT  TestCountsAreRecorded_async_loop_cells
        --- PASS: TestCountsAreRecorded_async_gap_catch_admission (0.02s)
        --- PASS: TestCountsAreRecorded_async_iterable_set_return_module (0.01s)
        === CONT  TestCountsAreRecorded_label_block_break_sync
        --- PASS: TestCountsAreRecorded_label_block_break_sync (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_return
        === CONT  TestCountsAreRecorded_async_switch_body
        --- PASS: TestCountsAreRecorded_async_switch_body (0.00s)
        === CONT  TestCountsAreRecorded_async_iterables_cells_labels
        --- PASS: TestCountsAreRecorded_async_loop_cells (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_mutation
        --- PASS: TestCountsAreRecorded_async_iterables_cells_labels (0.00s)
        === CONT  TestCountsAreRecorded_label_block_edges
        --- PASS: TestCountsAreRecorded_async_for_of_mutation (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_body
        --- PASS: TestCountsAreRecorded_label_block_edges (0.00s)
        === CONT  TestCountsAreRecorded_async_labels_block_finally
        --- PASS: TestCountsAreRecorded_async_labels_block_finally (0.00s)
        === CONT  TestCountsAreRecorded_async_top_loop
        --- PASS: TestCountsAreRecorded_async_for_of_return (0.01s)
        === CONT  TestCountsAreRecorded_async_top_catch_after_await
        --- PASS: TestCountsAreRecorded_async_for_of_body (0.00s)
        === CONT  TestCountsAreRecorded_async_labels_finally
        --- PASS: TestCountsAreRecorded_async_top_catch_after_await (0.00s)
        === CONT  TestCountsAreRecorded_async_top_catch_binding
        --- PASS: TestCountsAreRecorded_async_labels_finally (0.00s)
        === CONT  TestCountsAreRecorded_async_labels
        --- PASS: TestCountsAreRecorded_async_top_catch_binding (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_string_admission
        --- PASS: TestCountsAreRecorded_async_top_loop (0.01s)
        === CONT  TestCountsAreRecorded_async_finally_completion
        --- PASS: TestCountsAreRecorded_async_finally_completion (0.00s)
        === CONT  TestCountsAreRecorded_async_promise_expression
        --- PASS: TestCountsAreRecorded_async_labels (0.01s)
        === CONT  TestCountsAreRecorded_async_operand_order
        --- PASS: TestCountsAreRecorded_async_for_of_string_admission (0.01s)
        === CONT  TestCountsAreRecorded_async_loop_body_capture
        --- PASS: TestCountsAreRecorded_async_loop_body_capture (0.00s)
        === CONT  TestCountsAreRecorded_async_mutation
        --- PASS: TestCountsAreRecorded_async_operand_order (0.01s)
        === CONT  TestCountsAreRecorded_array
        --- PASS: TestCountsAreRecorded_async_promise_expression (0.01s)
        === CONT  TestCountsAreRecorded_async_methods
        --- PASS: TestCountsAreRecorded_array (0.01s)
        === CONT  TestCountsAreRecorded_flat
        --- PASS: TestCountsAreRecorded_async_methods (0.01s)
        === CONT  TestCountsAreRecorded_async_live_values
        --- PASS: TestCountsAreRecorded_flat (0.00s)
        === CONT  TestCountsAreRecorded_release_fast_shared
        --- PASS: TestCountsAreRecorded_async_live_values (0.01s)
        === CONT  TestCountsAreRecorded_async_fresh_holder
        --- PASS: TestCountsAreRecorded_async_fresh_holder (0.02s)
        === CONT  TestCountsAreRecorded_release_fast_graph
        --- PASS: TestCountsAreRecorded_async_mutation (0.04s)
        === CONT  TestCountsAreRecorded_counters_square_jump
        --- PASS: TestCountsAreRecorded_release_fast_graph (0.00s)
        === CONT  TestCountsAreRecorded_counters_shapes
        --- PASS: TestCountsAreRecorded_counters_shapes (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_float64array
        --- PASS: TestCountsAreRecorded_counters_square_jump (0.00s)
        === CONT  TestCountsAreRecorded_async_frame_capture_safe
        --- PASS: TestCountsAreRecorded_release_fast_shared (0.03s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_int32array
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_int32array (0.00s)
        === CONT  TestCountsAreRecorded_async_escaped_capture
        --- PASS: TestCountsAreRecorded_async_escaped_capture (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_uint16array
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_float64array (0.01s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_uint8array
        --- PASS: TestCountsAreRecorded_async_frame_capture_safe (0.01s)
        === CONT  TestCountsAreRecorded_async_control
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_uint8array (0.00s)
        === CONT  TestCountsAreRecorded_async_conditions
        === CONT  TestCountsAreRecorded_async_closures
        --- PASS: TestCountsAreRecorded_async_conditions (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_integer_stop
        --- PASS: TestCountsAreRecorded_async_closures (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_integer
        --- PASS: TestCountsAreRecorded_typed_arrays_integer (0.00s)
        === CONT  TestCountsAreRecorded_async_void_value
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_uint16array (0.01s)
        === CONT  TestCountsAreRecorded_async_snapshots
        --- PASS: TestCountsAreRecorded_async_control (0.00s)
        --- PASS: TestCountsAreRecorded_async_snapshots (0.00s)
        === CONT  TestCountsAreRecorded_async_receiver
        --- PASS: TestCountsAreRecorded_typed_arrays_integer_stop (0.00s)
        === CONT  TestCountsAreRecorded_async_counts_thrown
        --- PASS: TestCountsAreRecorded_async_counts_thrown (0.01s)
        === CONT  TestCountsAreRecorded_async_catch_return
        --- PASS: TestCountsAreRecorded_async_void_value (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_root_error
        --- PASS: TestCountsAreRecorded_async_counts_root_error (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_return_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_return_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_return_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_return_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_string_break_module
        --- PASS: TestCountsAreRecorded_async_receiver (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_break_function
        --- PASS: TestCountsAreRecorded_async_catch_return (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_return_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_break_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_break_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_break_function (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_break_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_break_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_nested_maps_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_break_module (0.00s)
        === CONT  TestCountsAreRecorded_label_block_outer_continue_async
        --- PASS: TestCountsAreRecorded_async_iterable_nested_maps_module (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_continue
        === CONT  TestCountsAreRecorded_async_finally_break
        --- PASS: TestCountsAreRecorded_async_finally_continue (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_throw
        --- PASS: TestCountsAreRecorded_async_finally_break (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_return
        --- PASS: TestCountsAreRecorded_async_finally_throw (0.00s)
        === CONT  TestCountsAreRecorded_label_block_break_async
        --- PASS: TestCountsAreRecorded_async_finally_return (0.00s)
        === CONT  TestCountsAreRecorded_label_block_switch_break_async
        --- PASS: TestCountsAreRecorded_label_block_break_async (0.00s)
        === CONT  TestCountsAreRecorded_label_block_continue_async
        --- PASS: TestCountsAreRecorded_label_block_switch_break_async (0.00s)
        === CONT  TestCountsAreRecorded_async_counts_parameter
        --- PASS: TestCountsAreRecorded_label_block_outer_continue_async (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_expression_write
        --- PASS: TestCountsAreRecorded_async_counts_expression_write (0.00s)
        --- PASS: TestCountsAreRecorded_async_iterable_map_return_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_nested_maps_function
        --- PASS: TestCountsAreRecorded_label_block_continue_async (0.01s)
        === CONT  TestCountsAreRecorded_label_block_outer_continue_sync
        === CONT  TestCountsAreRecorded_label_block_switch_break_sync
        --- PASS: TestCountsAreRecorded_async_iterable_nested_maps_function (0.00s)
        === CONT  TestCountsAreRecorded_async_catches
        --- PASS: TestCountsAreRecorded_label_block_outer_continue_sync (0.00s)
        --- PASS: TestCountsAreRecorded_label_block_switch_break_sync (0.00s)
        --- PASS: TestCountsAreRecorded_async_counts_parameter (0.01s)
        --- PASS: TestCountsAreRecorded_async_catches (0.00s)
        === NAME  TestCountsAreRecorded_003
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
--- FAIL: TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically (90.28s)
    families_generated_test.go:5425: the probe of TestStage0EmitsDeterministically was killed at 90 s:
        === RUN   TestStage0EmitsDeterministically_000
        === PAUSE TestStage0EmitsDeterministically_000
        === RUN   TestStage0EmitsDeterministically_001
        === PAUSE TestStage0EmitsDeterministically_001
        === RUN   TestStage0EmitsDeterministically_002
        === PAUSE TestStage0EmitsDeterministically_002
        === RUN   TestStage0EmitsDeterministically_003
        === PAUSE TestStage0EmitsDeterministically_003
        === RUN   TestStage0EmitsDeterministically_004
        === PAUSE TestStage0EmitsDeterministically_004
        === RUN   TestStage0EmitsDeterministically_005
        === PAUSE TestStage0EmitsDeterministically_005
        === RUN   TestStage0EmitsDeterministically_006
        === PAUSE TestStage0EmitsDeterministically_006
        === RUN   TestStage0EmitsDeterministically_007
        === PAUSE TestStage0EmitsDeterministically_007
        === RUN   TestStage0EmitsDeterministically_008
        === PAUSE TestStage0EmitsDeterministically_008
        === RUN   TestStage0EmitsDeterministically_009
        === PAUSE TestStage0EmitsDeterministically_009
        === RUN   TestStage0EmitsDeterministically_010
        === PAUSE TestStage0EmitsDeterministically_010
        === RUN   TestStage0EmitsDeterministically_011
        === PAUSE TestStage0EmitsDeterministically_011
        === RUN   TestStage0EmitsDeterministically_012
        === PAUSE TestStage0EmitsDeterministically_012
        === RUN   TestStage0EmitsDeterministically_013
        === PAUSE TestStage0EmitsDeterministically_013
        === RUN   TestStage0EmitsDeterministically_014
        === PAUSE TestStage0EmitsDeterministically_014
        === RUN   TestStage0EmitsDeterministically_015
        === PAUSE TestStage0EmitsDeterministically_015
        === RUN   TestStage0EmitsDeterministically_016
        === PAUSE TestStage0EmitsDeterministically_016
        === RUN   TestStage0EmitsDeterministically_017
        === PAUSE TestStage0EmitsDeterministically_017
        === RUN   TestStage0EmitsDeterministically_018
        === PAUSE TestStage0EmitsDeterministically_018
        === RUN   TestStage0EmitsDeterministically_019
        === PAUSE TestStage0EmitsDeterministically_019
        === RUN   TestStage0EmitsDeterministically_020
        === PAUSE TestStage0EmitsDeterministically_020
        === RUN   TestStage0EmitsDeterministically_021
        === PAUSE TestStage0EmitsDeterministically_021
        === RUN   TestStage0EmitsDeterministically_022
        === PAUSE TestStage0EmitsDeterministically_022
        === RUN   TestStage0EmitsDeterministically_023
        === PAUSE TestStage0EmitsDeterministically_023
        === RUN   TestStage0EmitsDeterministically_024
        === PAUSE TestStage0EmitsDeterministically_024
        === RUN   TestStage0EmitsDeterministically_025
        === PAUSE TestStage0EmitsDeterministically_025
        === RUN   TestStage0EmitsDeterministically_026
        === PAUSE TestStage0EmitsDeterministically_026
        === RUN   TestStage0EmitsDeterministically_027
        === PAUSE TestStage0EmitsDeterministically_027
        === RUN   TestStage0EmitsDeterministically_028
        === PAUSE TestStage0EmitsDeterministically_028
        === RUN   TestStage0EmitsDeterministically_029
        === PAUSE TestStage0EmitsDeterministically_029
        === RUN   TestStage0EmitsDeterministically_030
        === PAUSE TestStage0EmitsDeterministically_030
        === RUN   TestStage0EmitsDeterministically_031
        === PAUSE TestStage0EmitsDeterministically_031
        === CONT  TestStage0EmitsDeterministically_000
        === CONT  TestStage0EmitsDeterministically_016
        === CONT  TestStage0EmitsDeterministically_031
        === CONT  TestStage0EmitsDeterministically_030
        === NAME  TestStage0EmitsDeterministically_000
            families_generated_test.go:5265: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 2.03
        === NAME  TestStage0EmitsDeterministically_016
            families_generated_test.go:5345: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 2.11
        === NAME  TestStage0EmitsDeterministically_030
            families_generated_test.go:5415: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 2.23
        === NAME  TestStage0EmitsDeterministically_031
            families_generated_test.go:5420: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 2.40
            families_generated_test.go:5420: build go_build_./cmd/adamic_adamic e092ceb3d7ba miss 81.94
        --- PASS: TestStage0EmitsDeterministically_031 (88.23s)
        === CONT  TestStage0EmitsDeterministically_029
        === NAME  TestStage0EmitsDeterministically_000
            families_generated_test.go:5265: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 81.94
        --- PASS: TestStage0EmitsDeterministically_000 (88.24s)
        === CONT  TestStage0EmitsDeterministically_028
        === NAME  TestStage0EmitsDeterministically_016
            families_generated_test.go:5345: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 81.94
        --- PASS: TestStage0EmitsDeterministically_016 (88.24s)
        === CONT  TestStage0EmitsDeterministically_027
        === NAME  TestStage0EmitsDeterministically_030
            families_generated_test.go:5415: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 81.95
        --- PASS: TestStage0EmitsDeterministically_030 (88.24s)
        === CONT  TestStage0EmitsDeterministically_026
            families_generated_test.go:5395: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.05
        === NAME  TestStage0EmitsDeterministically_027
            families_generated_test.go:5400: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.06
        === NAME  TestStage0EmitsDeterministically_028
            families_generated_test.go:5405: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.07
        === NAME  TestStage0EmitsDeterministically_029
            families_generated_test.go:5410: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.07
        === NAME  TestStage0EmitsDeterministically_028
            families_generated_test.go:5405: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_028 (0.12s)
        === CONT  TestStage0EmitsDeterministically_025
        === NAME  TestStage0EmitsDeterministically_026
            families_generated_test.go:5395: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.06
        --- PASS: TestStage0EmitsDeterministically_026 (0.11s)
        === CONT  TestStage0EmitsDeterministically_024
        === NAME  TestStage0EmitsDeterministically_029
            families_generated_test.go:5410: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.05
        --- PASS: TestStage0EmitsDeterministically_029 (0.13s)
        === CONT  TestStage0EmitsDeterministically_023
        === NAME  TestStage0EmitsDeterministically_027
            families_generated_test.go:5400: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_027 (0.12s)
        === CONT  TestStage0EmitsDeterministically_022
        === NAME  TestStage0EmitsDeterministically_025
            families_generated_test.go:5390: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.02
        === NAME  TestStage0EmitsDeterministically_023
            families_generated_test.go:5380: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.03
        === NAME  TestStage0EmitsDeterministically_024
            families_generated_test.go:5385: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_022
            families_generated_test.go:5375: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_025
            families_generated_test.go:5390: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.04
        --- PASS: TestStage0EmitsDeterministically_025 (0.05s)
        === CONT  TestStage0EmitsDeterministically_021
        === NAME  TestStage0EmitsDeterministically_022
            families_generated_test.go:5375: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.02
        --- PASS: TestStage0EmitsDeterministically_022 (0.06s)
        === CONT  TestStage0EmitsDeterministically_020
        === NAME  TestStage0EmitsDeterministically_023
            families_generated_test.go:5380: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_023 (0.07s)
        === CONT  TestStage0EmitsDeterministically_019
        === NAME  TestStage0EmitsDeterministically_024
            families_generated_test.go:5385: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_024 (0.07s)
        === CONT  TestStage0EmitsDeterministically_018
        === NAME  TestStage0EmitsDeterministically_021
            families_generated_test.go:5370: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.05
        === NAME  TestStage0EmitsDeterministically_020
            families_generated_test.go:5365: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_018
            families_generated_test.go:5355: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_019
            families_generated_test.go:5360: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.05
        === NAME  TestStage0EmitsDeterministically_021
            families_generated_test.go:5370: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.02
        --- PASS: TestStage0EmitsDeterministically_021 (0.07s)
        === CONT  TestStage0EmitsDeterministically_017
        === NAME  TestStage0EmitsDeterministically_020
            families_generated_test.go:5365: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        === RUN   TestStage0EmitsDeterministically_020/internal/oracle/testdata/determinism_dispatch.a
        === PAUSE TestStage0EmitsDeterministically_020/internal/oracle/testdata/determinism_dispatch.a
        === CONT  TestStage0EmitsDeterministically_012
        === NAME  TestStage0EmitsDeterministically_019
            families_generated_test.go:5360: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_019 (0.08s)
        === CONT  TestStage0EmitsDeterministically_015
        === NAME  TestStage0EmitsDeterministically_018
            families_generated_test.go:5355: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.04
        --- PASS: TestStage0EmitsDeterministically_018 (0.08s)
        === CONT  TestStage0EmitsDeterministically_014
        === NAME  TestStage0EmitsDeterministically_017
            families_generated_test.go:5350: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.03
        === NAME  TestStage0EmitsDeterministically_012
            families_generated_test.go:5325: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.02
        === NAME  TestStage0EmitsDeterministically_017
            families_generated_test.go:5350: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.02
        --- PASS: TestStage0EmitsDeterministically_017 (0.05s)
        === CONT  TestStage0EmitsDeterministically_013
        === NAME  TestStage0EmitsDeterministically_015
            families_generated_test.go:5340: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.03
        === NAME  TestStage0EmitsDeterministically_014
            families_generated_test.go:5335: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.05
        === NAME  TestStage0EmitsDeterministically_013
            families_generated_test.go:5330: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_012
            families_generated_test.go:5325: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.05
        --- PASS: TestStage0EmitsDeterministically_012 (0.07s)
        === CONT  TestStage0EmitsDeterministically_010
        === NAME  TestStage0EmitsDeterministically_015
            families_generated_test.go:5340: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_015 (0.09s)
        === CONT  TestStage0EmitsDeterministically_011
        === NAME  TestStage0EmitsDeterministically_013
            families_generated_test.go:5330: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_013 (0.07s)
        === CONT  TestStage0EmitsDeterministically_004
        === NAME  TestStage0EmitsDeterministically_014
            families_generated_test.go:5335: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.04
        --- PASS: TestStage0EmitsDeterministically_014 (0.09s)
        === CONT  TestStage0EmitsDeterministically_007
        === NAME  TestStage0EmitsDeterministically_010
            families_generated_test.go:5315: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_004
            families_generated_test.go:5285: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.03
        === NAME  TestStage0EmitsDeterministically_011
            families_generated_test.go:5320: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_010
            families_generated_test.go:5315: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.05
        --- PASS: TestStage0EmitsDeterministically_010 (0.09s)
        === CONT  TestStage0EmitsDeterministically_006
        === NAME  TestStage0EmitsDeterministically_007
            families_generated_test.go:5300: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.07
        === NAME  TestStage0EmitsDeterministically_004
            families_generated_test.go:5285: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.07
        --- PASS: TestStage0EmitsDeterministically_004 (0.10s)
        === CONT  TestStage0EmitsDeterministically_005
        === NAME  TestStage0EmitsDeterministically_006
            families_generated_test.go:5295: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.03
        === NAME  TestStage0EmitsDeterministically_007
            families_generated_test.go:5300: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        === NAME  TestStage0EmitsDeterministically_011
            families_generated_test.go:5320: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.09
        --- PASS: TestStage0EmitsDeterministically_007 (0.11s)
        === CONT  TestStage0EmitsDeterministically_002
        --- PASS: TestStage0EmitsDeterministically_011 (0.13s)
        === CONT  TestStage0EmitsDeterministically_003
        === NAME  TestStage0EmitsDeterministically_005
            families_generated_test.go:5290: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.06
        === NAME  TestStage0EmitsDeterministically_002
            families_generated_test.go:5275: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.05
        === NAME  TestStage0EmitsDeterministically_006
            families_generated_test.go:5295: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.06
        --- PASS: TestStage0EmitsDeterministically_006 (0.09s)
        === CONT  TestStage0EmitsDeterministically_009
        === NAME  TestStage0EmitsDeterministically_003
            families_generated_test.go:5280: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.06
        === NAME  TestStage0EmitsDeterministically_002
            families_generated_test.go:5275: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.02
        --- PASS: TestStage0EmitsDeterministically_002 (0.08s)
        === CONT  TestStage0EmitsDeterministically_001
        === NAME  TestStage0EmitsDeterministically_005
            families_generated_test.go:5290: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_005 (0.09s)
        === CONT  TestStage0EmitsDeterministically_008
        === NAME  TestStage0EmitsDeterministically_003
            families_generated_test.go:5280: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_003 (0.08s)
        === CONT  TestStage0EmitsDeterministically_020/internal/oracle/testdata/determinism_dispatch.a
        === NAME  TestStage0EmitsDeterministically_009
            families_generated_test.go:5310: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.04
        === NAME  TestStage0EmitsDeterministically_001
            families_generated_test.go:5270: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.02
        === NAME  TestStage0EmitsDeterministically_008
            families_generated_test.go:5305: build go_build_./cmd/adamic_adamic ebf2716d63c3 hit 0.02
        === NAME  TestStage0EmitsDeterministically_001
            families_generated_test.go:5270: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.02
        --- PASS: TestStage0EmitsDeterministically_001 (0.04s)
        === NAME  TestStage0EmitsDeterministically_009
            families_generated_test.go:5310: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.03
        --- PASS: TestStage0EmitsDeterministically_009 (0.07s)
        === NAME  TestStage0EmitsDeterministically_008
            families_generated_test.go:5305: build go_build_./cmd/adamic_adamic e092ceb3d7ba hit 0.02
        --- PASS: TestStage0EmitsDeterministically_008 (0.04s)
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
FAIL
gate cache: native hits=3088 misses=6600
gate cache: node hits=1745 misses=6543
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	3881.080s
FAIL
```

### Failure attribution and individual main reruns

The original full oracle run did not set NODE_PATH or provide stage3/api/node_modules. The pinned packages and tsc were installed outside the repository as requested. The twelve assignment-proof tests explicitly override NODE_PATH with the checkout-relative stage3/api/node_modules, so PATH alone does not satisfy their require. WASI_SYSROOT was absent. The two family tests have their own internal 90-second probe budget.

| Failing test | First error line (verbatim) | Environment / timing attribution | Main isolated result |
|---|---|---|---|
| `TestAssignmentProofWiderTargetMutant` | assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 8.86s |
| `TestAssignmentProofWiderTarget` | assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 4.35s |
| `TestAssignmentProofCompoundMutant` | assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.38s |
| `TestAssignmentProofCompound` | assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.52s |
| `TestAssignmentProofChainedMutant` | assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.11s |
| `TestAssignmentProofChained` | assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 2.95s |
| `TestAssignmentProofReturnDefinedMutant` | assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.35s |
| `TestAssignmentProofReturnDefined` | assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.00s |
| `TestAssignmentProofIfNarrowingMutant` | assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.10s |
| `TestAssignmentProofIfNarrowing` | assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 2.74s |
| `TestAssignmentProofScannerKeywordMutant` | assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 3.15s |
| `TestAssignmentProofScannerKeyword` | assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Environment: Node module resolution | FAIL (also fails on main), exit 1, 4.63s |
| `TestStdoutColumnsRefusesWASI` | library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT | Environment: missing WASI_SYSROOT | FAIL (also fails on main), exit 1, 4.00s |
| `TestFamiliesCatchAPlantedCase_CountsAreRecorded` | families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s: | Timing/resource-sensitive internal probe timeout; HEAD passes alone | PASS (does not fail alone on main), exit 0, 24.48s |
| `TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically` | families_generated_test.go:5425: the probe of TestStage0EmitsDeterministically was killed at 90 s: | Timing/resource-sensitive internal probe timeout; HEAD passes alone | PASS (does not fail alone on main), exit 0, 73.30s |

Each main test was rerun by its own exact anchored name, one at a time. Full individual outputs follow.

#### Main: TestAssignmentProofWiderTargetMutant

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofWiderTargetMutant$' -count=1 -v -timeout 120m`

Exit **1**; wall **8.86 seconds**.

```text
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.082s
FAIL
```

#### Main: TestAssignmentProofWiderTarget

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofWiderTarget$' -count=1 -v -timeout 120m`

Exit **1**; wall **4.35 seconds**.

```text
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== CONT  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.21s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.252s
FAIL
```

#### Main: TestAssignmentProofCompoundMutant

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofCompoundMutant$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.38 seconds**.

```text
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.075s
FAIL
```

#### Main: TestAssignmentProofCompound

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofCompound$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.52 seconds**.

```text
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== CONT  TestAssignmentProofCompound
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.09s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.098s
FAIL
```

#### Main: TestAssignmentProofChainedMutant

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofChainedMutant$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.11 seconds**.

```text
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== CONT  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.066s
FAIL
```

#### Main: TestAssignmentProofChained

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofChained$' -count=1 -v -timeout 120m`

Exit **1**; wall **2.95 seconds**.

```text
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== CONT  TestAssignmentProofChained
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.081s
FAIL
```

#### Main: TestAssignmentProofReturnDefinedMutant

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofReturnDefinedMutant$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.35 seconds**.

```text
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== CONT  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.067s
FAIL
```

#### Main: TestAssignmentProofReturnDefined

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofReturnDefined$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.00 seconds**.

```text
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== CONT  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.064s
FAIL
```

#### Main: TestAssignmentProofIfNarrowingMutant

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofIfNarrowingMutant$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.10 seconds**.

```text
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== CONT  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.072s
FAIL
```

#### Main: TestAssignmentProofIfNarrowing

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofIfNarrowing$' -count=1 -v -timeout 120m`

Exit **1**; wall **2.74 seconds**.

```text
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== CONT  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.066s
FAIL
```

#### Main: TestAssignmentProofScannerKeywordMutant

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofScannerKeywordMutant$' -count=1 -v -timeout 120m`

Exit **1**; wall **3.15 seconds**.

```text
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== CONT  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.08s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.096s
FAIL
```

#### Main: TestAssignmentProofScannerKeyword

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProofScannerKeyword$' -count=1 -v -timeout 120m`

Exit **1**; wall **4.63 seconds**.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.100s
FAIL
```

#### Main: TestStdoutColumnsRefusesWASI

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestStdoutColumnsRefusesWASI$' -count=1 -v -timeout 120m`

Exit **1**; wall **4.00 seconds**.

```text
=== RUN   TestStdoutColumnsRefusesWASI
=== PAUSE TestStdoutColumnsRefusesWASI
=== CONT  TestStdoutColumnsRefusesWASI
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestStdoutColumnsRefusesWASI (0.66s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.686s
FAIL
```

#### Main: TestFamiliesCatchAPlantedCase_CountsAreRecorded

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_CountsAreRecorded$' -count=1 -v -timeout 120m`

Exit **0**; wall **24.48 seconds**.

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (21.22s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	21.230s
```

#### Main: TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically$' -count=1 -v -timeout 120m`

Exit **0**; wall **73.30 seconds**.

```text
=== RUN   TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== PAUSE TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== CONT  TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
    families_generated_test.go:5425: the mutant planted in internal/oracle/testdata/determinism_dispatch.a failed exactly TestStage0EmitsDeterministically_020, with "stage 0 emitted different C"
--- PASS: TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically (67.35s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	67.372s
```

### Supplemental environment and timing checks

Setting external NODE_PATH alone did not resolve the twelve assignment checks, because assignmentProofTypecheck appends its own NODE_PATH. The first supplemental commands below preserve that result. Then temporarily linked each checkout's stage3/api/node_modules to /tmp/n7s-proof/api/node_modules, with all package bytes remaining outside the repositories. All twelve tests passed on each SHA. Removed both links afterward. This is a focused supplemental result; the original full-suite exit stays 1.

#### head-assignment-resolved

Command: `go test ./internal/oracle -run '^TestAssignmentProof' -count=1 -v -timeout 30m`

Exit **1**; wall **13.21 seconds**.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofIfNarrowing
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.10s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.10s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.12s)
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.16s)
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.05s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.15s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
=== CONT  TestAssignmentProofScannerKeywordMutant
--- FAIL: TestAssignmentProofWiderTarget (0.12s)
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.14s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.06s)
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.24s)
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.10s)
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.44s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.462s
FAIL
```

#### main-assignment-resolved

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProof' -count=1 -v -timeout 30m`

Exit **1**; wall **4.52 seconds**.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofCompound
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.07s)
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.07s)
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.08s)
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.08s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.08s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.06s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.07s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.06s)
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.07s)
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.08s)
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.05s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [ '/tmp/n7s-main/stage3/fixtures/assignment-proofs/check-types.cjs' ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.04s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.198s
FAIL
```

#### head-assignment-linked

Command: `go test ./internal/oracle -run '^TestAssignmentProof' -count=1 -v -timeout 30m`

Exit **0**; wall **22.01 seconds**.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.67
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.67
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.68
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.69
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.82
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:202: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128 1\n", mutant "128 0 1\n"
--- PASS: TestAssignmentProofWiderTargetMutant (10.55s)
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:192: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "12 12 1\n14 14 2\n", mutant "2 2 1\n2 2 2\n"
--- PASS: TestAssignmentProofCompoundMutant (10.55s)
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofCompound
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofCompound (10.56s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofScannerKeyword (10.65s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofChained
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofChained (10.70s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofReturnDefined (3.57s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofWiderTarget (3.53s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:172: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 1\nvalue2 value2 2\n", mutant "value1 undefined 1\nvalue2 undefined 2\n"
--- PASS: TestAssignmentProofReturnDefinedMutant (3.66s)
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:162: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "4 1\n-1 2\n", mutant "4 1\n7 2\n"
--- PASS: TestAssignmentProofIfNarrowingMutant (3.80s)
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofIfNarrowing (3.76s)
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 028a47ca193f hit 0.02
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:182: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 value1 1\n", mutant "value1 value1 undefined 1\n"
--- PASS: TestAssignmentProofChainedMutant (2.77s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:152: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128\n83 83\n80 80\n80 80\n80 80\n", mutant "80 80\n80 80\n80 80\n80 80\n80 80\n"
--- PASS: TestAssignmentProofScannerKeywordMutant (2.80s)
PASS
gate cache: native hits=12 misses=24
gate cache: node hits=0 misses=24
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	17.028s
```

#### main-assignment-linked

Command: `cd /tmp/n7s-main && go test ./internal/oracle -run '^TestAssignmentProof' -count=1 -v -timeout 30m`

Exit **0**; wall **35.55 seconds**.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofIfNarrowingMutant
=== CONT  TestAssignmentProofCompound
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e miss 14.13
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 14.13
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 14.13
=== NAME  TestAssignmentProofCompound
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 14.13
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 14.13
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofScannerKeyword (24.02s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofChained
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofChained (24.08s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:162: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "4 1\n-1 2\n", mutant "4 1\n7 2\n"
--- PASS: TestAssignmentProofIfNarrowingMutant (24.11s)
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofCompound
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofCompound (24.12s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:172: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 1\nvalue2 value2 2\n", mutant "value1 undefined 1\nvalue2 undefined 2\n"
--- PASS: TestAssignmentProofReturnDefinedMutant (24.12s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.04
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.01
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.01
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.01
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.01
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofReturnDefined (3.32s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:182: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 value1 1\n", mutant "value1 value1 undefined 1\n"
--- PASS: TestAssignmentProofChainedMutant (3.60s)
=== CONT  TestAssignmentProofCompoundMutant
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:202: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128 1\n", mutant "128 0 1\n"
--- PASS: TestAssignmentProofWiderTargetMutant (3.66s)
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofWiderTarget (3.72s)
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofIfNarrowing (3.85s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
    assignment_proofs_test.go:152: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.01
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:152: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128\n83 83\n80 80\n80 80\n80 80\n", mutant "80 80\n80 80\n80 80\n80 80\n80 80\n"
--- PASS: TestAssignmentProofScannerKeywordMutant (2.16s)
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:192: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "12 12 1\n14 14 2\n", mutant "2 2 1\n2 2 2\n"
--- PASS: TestAssignmentProofCompoundMutant (1.95s)
PASS
gate cache: native hits=12 misses=24
gate cache: node hits=0 misses=24
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	29.650s
```

#### head-family-counts-alone

Command: `go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_CountsAreRecorded$' -count=1 -v -timeout 120m`

Exit **0**; wall **22.23 seconds**.

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (19.21s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	19.228s
```

#### head-family-determinism-alone

Command: `go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically$' -count=1 -v -timeout 120m`

Exit **0**; wall **11.35 seconds**.

```text
=== RUN   TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== PAUSE TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== CONT  TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
    families_generated_test.go:5425: the mutant planted in internal/oracle/testdata/determinism_dispatch.a failed exactly TestStage0EmitsDeterministically_020, with "stage 0 emitted different C"
--- PASS: TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically (4.97s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	5.001s
```

The family probes pass in isolation on both SHAs. Their full-suite timeouts are not reproduced by the isolated HEAD run, so they are recorded as timing/resource-sensitive failures rather than a demonstrated change regression. The WASI setup failure remains unresolved and baseline-reproduced.

## 7. Admission delta

Exit **0**; wall **1309.59 seconds**; verdict **pass**; void **false**; void reasons **none**; diff_count **8**; newly admitted **8**; newly refused **0**.

Generated HEAD's full manifest with cloud/admission-corpus/manifest.py and built cmd/adamic-admission-delta outside the repository. Ran the gate alone with no other heavy command running; its two classification workers were the requested internal concurrency. No corpus filter, shard, sampling omission, or runtime budget was used.

```sh
/tmp/n7s-proof/admission --base 16429a30452d027809f0ea85bc395479423e493e --head 5eeac42cf8c0037c627f32b975dd254cbfd20c6b --workers 2 --manifest /tmp/n7s-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json > /tmp/n7s-proof/admission-result.json 2> /tmp/n7s-proof/admission.stderr
```

```text
witnesses=2 fixtures=4274 gaps=105 review=20 fuzz=0 stage3=821
```

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 1 runs timed out, 1 retried alone, 0 timed out again
changed programs: 8
```

Aggregate JSON (corpus program lists summarized by count):

```json
{
  "negative_witness_count": 50,
  "accepted_witnesses": 50,
  "node_agreements": 21,
  "runtime_notyet_count": 52,
  "ruled_runtime_stops": 52,
  "ruled_runtime_stop_paths": [
    "internal/oracle/testdata/intl_collator/in_compare.a",
    "internal/oracle/testdata/intl_collator/locale_en.a",
    "internal/oracle/testdata/intl_collator/locale_malformed.a",
    "internal/oracle/testdata/intl_collator/locale_variable.a",
    "internal/oracle/testdata/intl_collator/own_keys.a",
    "internal/oracle/testdata/intl_collator/tsc_shape.a",
    "internal/oracle/testdata/intl_collator/user_and_host.a",
    "internal/oracle/testdata/intl_collator_compare_value/p01_tsc_shape.a",
    "internal/oracle/testdata/intl_collator_compare_value/p01_tsc_shape.ts",
    "internal/oracle/testdata/reflect_get/stop_error_stack.a",
    "internal/oracle/testdata/reflect_get/stop_function_name.a",
    "internal/oracle/testdata/reflect_get/stop_hasownproperty_host.a",
    "internal/oracle/testdata/reflect_get/stop_host_collator.a",
    "internal/oracle/testdata/reflect_get/stop_host_dirent.a",
    "internal/oracle/testdata/reflect_get/stop_host_error_prototype.a",
    "internal/oracle/testdata/reflect_get/stop_host_fs_error_prototype.a",
    "internal/oracle/testdata/reflect_get/stop_host_groups.a",
    "internal/oracle/testdata/reflect_get/stop_host_handle.a",
    "internal/oracle/testdata/reflect_get/stop_host_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_host_iterator_next.a",
    "internal/oracle/testdata/reflect_get/stop_host_memory_field.a",
    "internal/oracle/testdata/reflect_get/stop_host_memory_heap.a",
    "internal/oracle/testdata/reflect_get/stop_host_performance_mark.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_ctime.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_date.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_field.a",
    "internal/oracle/testdata/reflect_get/stop_host_timeout.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_error_code.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_internal.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_literal.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_method.a",
    "internal/oracle/testdata/reflect_get/stop_keys_host.a",
    "internal/oracle/testdata/reflect_get/stop_keys_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_keys_memory.a",
    "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "internal/oracle/testdata/reflect_get/stop_method.a",
    "internal/oracle/testdata/reflect_get/stop_primitive.a",
    "internal/oracle/testdata/reflect_get/stop_private.a",
    "internal/oracle/testdata/reflect_get/stop_proto.a",
    "internal/oracle/testdata/reflect_get/stop_prototype_class.a",
    "internal/oracle/testdata/reflect_get/stop_string_host.a",
    "internal/oracle/testdata/timers_source/stop_foreign_generic.a",
    "internal/oracle/testdata/timers_source/stop_foreign_handle.a",
    "internal/oracle/testdata/timers_source/stop_foreign_intersection.a",
    "internal/oracle/testdata/timers_source/stop_foreign_record.a",
    "internal/oracle/testdata/to_primitive/stop_boolean_array.a",
    "internal/oracle/testdata/to_primitive/stop_record_function.a",
    "internal/oracle/testdata/to_primitive/stop_scalar_array.a",
    "internal/oracle/testdata/to_primitive/stop_static_class.a",
    "internal/oracle/testdata/to_primitive/stop_undefined_array.a",
    "internal/oracle/testdata/to_primitive/stop_union_numbers.a"
  ],
  "stale_runtime_stops": [],
  "pending_runtime_stops": [],
  "backend_divergence_count": 12,
  "ruled_backend_divergences": 0,
  "ruled_backend_divergence_paths": [],
  "corpus_filter": "",
  "shard": "",
  "complete_corpus": true,
  "phase_seconds": {
    "base_build": 31.936587701,
    "checkout": 18.796723526,
    "checkout_and_provision": 19.831478216,
    "classification": 689.164539283,
    "head_build": 28.776359319,
    "manifest_and_verification": 22.067056524,
    "provision": 1.034750338,
    "runtime": 516.819569123,
    "total": 1308.608849414
  },
  "workers": 2,
  "classification": "lowering-only",
  "diff": [
    {
      "path": "internal/oracle/testdata/captured_union/alternating.ts",
      "blob": "4b580eb148914fd952757e4b14deb1f083910dfc"
    },
    {
      "path": "internal/oracle/testdata/captured_union/by_value.ts",
      "blob": "b24ed53fa8a4d070d6c48e453b01ee04c5eadc56"
    },
    {
      "path": "internal/oracle/testdata/captured_union/class.ts",
      "blob": "17a9fa67adb8887621d653d8940cb63219b2105d"
    },
    {
      "path": "internal/oracle/testdata/captured_union/cycle.ts",
      "blob": "926410728889a65f9a244c923418b198a7782112"
    },
    {
      "path": "internal/oracle/testdata/captured_union/graph.ts",
      "blob": "cdd254124dd70aed8f3820833d50d3fbaae903e7"
    },
    {
      "path": "internal/oracle/testdata/captured_union/nested.ts",
      "blob": "f2e563f8f214272e0756d4d18f4708878ce2bc7e"
    },
    {
      "path": "internal/oracle/testdata/captured_union/object.ts",
      "blob": "994b85cabefa5c8ec5a2bd214bfc6e6ae9edeb2f"
    },
    {
      "path": "internal/oracle/testdata/captured_union/released.ts",
      "blob": "eeed38791e78e927fe7788868ac09f0c3f9ef76c"
    }
  ],
  "diff_count": 8,
  "diff_ts_unmatched": [],
  "compile_timeout_seconds": 45,
  "runtime_timeout_seconds": 10,
  "base": "16429a30452d027809f0ea85bc395479423e493e",
  "head": "5eeac42cf8c0037c627f32b975dd254cbfd20c6b",
  "generator_revision": "5eeac42cf8c0037c627f32b975dd254cbfd20c6b",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "6c809214df325dc2e8a6755808ed2fc70da85f87",
  "admitted": 8,
  "verdict": "pass",
  "void": false,
  "sampling_seed": "53ca2c8ff25a8d0f4141c7108dbb5d80f126aa3e42e1f697ce06e7a2658d3f4e",
  "seed_given": false,
  "sampling_size": 123,
  "omitted": 0,
  "budget_seconds": 0,
  "budget_used_seconds": 516.819569123,
  "timed_out_runs": 1,
  "retried_runs": 1,
  "still_timed_out_runs": 0,
  "corpora": [
    {
      "name": "diff",
      "program_count": 8
    },
    {
      "name": "witnesses",
      "program_count": 2
    },
    {
      "name": "fixtures",
      "program_count": 4274
    },
    {
      "name": "gaps",
      "program_count": 105
    },
    {
      "name": "review",
      "program_count": 20
    },
    {
      "name": "fuzz",
      "program_count": 0
    },
    {
      "name": "stage3",
      "program_count": 821
    }
  ]
}
```

### Every classification stderr difference (8)

The following are all programs for which base_compile.stderr differs from compile.stderr, comparing exact strings without normalizing paths. Empty stderr is shown as an empty code block and as the JSON string `""`. No other program has a classification stderr difference.

#### internal/oracle/testdata/captured_union/alternating.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/by_value.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/class.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/cycle.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/graph.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/nested.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/object.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

#### internal/oracle/testdata/captured_union/released.ts

Classification: newly-accepted. Main exit 1; HEAD exit 0; Node/backend agreement True.

Main stderr, verbatim:

```text
adamic: /tmp/adamic-gate/admission-delta-3701263834/head/internal/oracle/testdata/captured_union/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr, verbatim (zero bytes, `""`):

```text

```

### Timeout and isolated reruns at both SHAs

```json
{
  "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
  "blob": "ee5d428ca5813c54046541e9d0a6d83ede986d55",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.081911565,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.108525261,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.2960363,
    "exit": 0,
    "stdout": "got\nstring\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.103199655,
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "native": {
    "wall_seconds": 0.002207825,
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "javascript_compile": {
    "wall_seconds": 0.127890631,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 9.174329106,
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
  "sampled": true,
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```

Only internal/oracle/testdata/reflect_get/stop_literal_getter.a recorded a timeout: native compile, original 10-second limit. The gate retried it alone successfully (9.174329106 seconds), leaving zero still-timed-out runs. It is accepted by both SHAs and has an already ruled runtime stop.

For the requested separate replays, built a compiler from each exact SHA, using a fresh detached external main worktree and the checked-out HEAD. The program blob is unchanged between the SHAs (git diff for that path printed nothing). Ran each native build alone, sequentially, under timeout 10s. Also executed each resulting program alone. Both produce the same ruled getter stop; Node's source output from the gate is `got\nstring\n`, while native prints `got\n` and exits 70 with the diagnostic shown below.

#### timeout-head-native-build

Command: `timeout 10s /tmp/n7s-proof/timeout-head-adamic build internal/oracle/testdata/reflect_get/stop_literal_getter.a -o /tmp/n7s-proof/timeout-head-native`

Exit **0**; wall **2.60 seconds**.

```text

```

#### timeout-head-native-run

Command: `timeout 10s /tmp/n7s-proof/timeout-head-native`

Exit **70**; wall **0.01 seconds**.

```text
got
adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')
```

#### timeout-main-native-build

Command: `cd /tmp/n7s-timeout-main && timeout 10s /tmp/n7s-proof/timeout-main-adamic build internal/oracle/testdata/reflect_get/stop_literal_getter.a -o /tmp/n7s-proof/timeout-main-native`

Exit **0**; wall **2.27 seconds**.

```text

```

#### timeout-main-native-run

Command: `timeout 10s /tmp/n7s-proof/timeout-main-native`

Exit **70**; wall **0.01 seconds**.

```text
got
adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')
```

The comparison and timeout worktrees were removed. All three mutants were restored and each clean-diff check passed. Only this report is intended for the report branch.

## Expectations appendix

Verbatim HEAD test sources defining all wants (including exact stdout strings, zero-leak/count checks, and lowerer assertions):

```go
package lower

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A captured let of differently held members lowers to one cell holding the Union, the same payload
// an uncaptured one has and a placeholder one already had: no placeholder, and every closure that
// reads or writes it carries that one cell (#n7sx9gz). The oracle's TestCapturedUnionAgrees runs them.
func TestCapturedUnionLowersToOneCell(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ fixture, name string }{
		{"class", "active"},
		{"object", "active"},
		{"by_value", "active"},
		{"released", "active"},
		{"nested", "extended"},
		{"alternating", "active"},
		{"graph", "active"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(probe.fixture+extension, func(t *testing.T) {
				t.Parallel()
				source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "captured_union", probe.fixture+".ts"))
				if err != nil {
					t.Fatal(err)
				}
				program, err := lowerSourceAs(t, string(source), extension)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for index, local := range program.Locals {
					if local.Name != probe.name {
						continue
					}
					found = true
					if local.Type != ir.Union || !local.Captured || local.Placeholder != "" {
						t.Fatalf("%s: Type %v Captured %t Placeholder %q, want a captured Union that isn't a placeholder", local.Name, local.Type, local.Captured, local.Placeholder)
					}
					if probe.fixture == "graph" && !local.GraphCell {
						t.Errorf("%s: not a graph cell, though the methods that capture it are on a cycle", local.Name)
					}
					closures := 0
					for _, function := range program.Functions {
						if slices.Contains(function.Environment, index) {
							closures++
						}
					}
					if closures < 2 {
						t.Errorf("%s: %d closures carry its cell, want every one that reads or writes it", local.Name, closures)
					}
				}
				if !found {
					t.Fatalf("no local named %s", probe.name)
				}
			})
		}
	}
}

// boolean | undefined has no word of its own in a cell yet, so a captured one still stops, with
// main's words.
func TestCapturedMaybeBooleanStillStops(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "function run(): boolean {\n\tlet seen: boolean | undefined;\n\tconst mark = (): void => {\n\t\tseen = true;\n\t};\n\tmark();\n\treturn seen ?? false;\n}\nconsole.log(`${run()}`);\n")
	var notYet *NotYet
	if !errors.As(err, &notYet) || err.Error() != "main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet" {
		t.Fatalf("got %v, want main's stop at 4:3", err)
	}
}
```

```go
package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// capturedUnionPins are #n7sx9gz's programs: a let whose members are held differently (an object, a
// string literal, undefined), captured by the closures that write and read it. Each is held to Node,
// and to the pinned output too, so an agreement on the wrong text can't pass.
var capturedUnionPins = []struct{ name, want string }{
	{"class", "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"},
	{"object", "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"},
	{"by_value", "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"},
	{"released", "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"},
	{"nested", "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"},
	{"alternating", "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n"},
	{"graph", "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"},
}

// Each pin is committed as .ts and run as .a too, from the same text, so both dialects are held to
// one program. Every build is asked whether it let go of everything: the sanitized one by the shared
// leak check, and the counted one by its own counts, where an overwritten member never released reads
// as a value leaked.
func TestCapturedUnionAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range capturedUnionPins {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(pin.name+extension, func(t *testing.T) {
				t.Parallel()
				path := capturedUnionSource(t, pin.name, extension)
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				if pin.name == "graph" && !capturedUnionGraphCell(program, "active") {
					t.Fatal("active's cell isn't a graph cell, so the graph pin doesn't pin one")
				}
				capturedUnionRuns(t, path, program, pin.want)
			})
		}
	}
}

// Memory's m3: the Session the cell holds keeps a closure that keeps the cell, a cycle through the
// Union. The rule holds it, agreeing with Node and letting go of everything, or refuses it as a cycle;
// either is an answer, and anything else fails.
func TestCapturedUnionCycleHeldOrRefused(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".ts", ".a"} {
		t.Run("cycle"+extension, func(t *testing.T) {
			t.Parallel()
			path := capturedUnionSource(t, "cycle", extension)
			program, err := lowered(t, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			switch {
			case err == nil:
				t.Logf("held: graph cell %t", capturedUnionGraphCell(program, "active"))
				capturedUnionRuns(t, path, program, "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n")
			case (errors.As(err, &refused) || errors.As(err, &notYet)) && strings.Contains(err.Error(), "cycle"):
				t.Logf("refused: %v", err)
			default:
				t.Fatalf("neither held nor refused as a cycle: %v", err)
			}
		})
	}
}

// capturedUnionSource writes a pin's committed .ts text to a file of the given extension.
func capturedUnionSource(t *testing.T, name, extension string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/captured_union", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+extension)
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// capturedUnionGraphCell is whether the captured local of that name lives in a graph cell.
func capturedUnionGraphCell(program *ir.Program, name string) bool {
	for _, local := range program.Locals {
		if local.Name == name && local.Captured && local.GraphCell {
			return true
		}
	}
	return false
}

// capturedUnionRuns holds a lowered pin to Node and to want, natively, sanitized with the leak check,
// on the counted build with its counts, and on JavaScript.
func capturedUnionRuns(t *testing.T, path string, program *ir.Program, want string) {
	t.Helper()
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != want {
		t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, want, truth.stderr)
	}
	t.Run("native", func(t *testing.T) {
		got := releasedUncached(t, program)
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
		}
	})
	t.Run("sanitized", func(t *testing.T) {
		got, leaked, comparable := nativelyCheckingLeaks(t, program, true)
		if leaked != "" {
			t.Errorf("leaks: %s", leaked)
		}
		if !comparable {
			return
		}
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
		}
	})
	t.Run("counted", func(t *testing.T) {
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
			t.Errorf("counts: %s", report)
		}
		if string(got.stdout) != want {
			t.Errorf("got stdout %q", got.stdout)
		}
		if match := leakcheck.CountsLine.FindSubmatch(got.stderr); match != nil {
			t.Logf("counts: allocations %s, frees %s, in regions %s", match[1], match[2], match[6])
		}
	})
	t.Run("javascript", func(t *testing.T) {
		got := onJavaScriptBackend(t, program)
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
		}
	})
}
```


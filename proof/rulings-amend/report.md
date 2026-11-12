# Rulings S3c amendment test report

Base commit: `9a6df75a7756eb6c70baf8e3ab7d5f47355a27c1` (checked out detached).

Loaded `/workspace/adamic-tools/env.sh` before commands, created `/tmp/adamic-gate` outside the repository, and exported `GOFLAGS=-buildvcs=false` for both test runs.

## Preparation

- `git fetch origin emit/rulings-s3c-amend`: exit 0.
- `git checkout --detach 9a6df75a7756eb6c70baf8e3ab7d5f47355a27c1`: exit 0.
- `git submodule update --init --recursive`: exit 0. `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; nested `TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- `gofmt -l internal/lower`: exit 0; printed nothing.

## Baseline

Command: `go test ./internal/lower -run '^TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad$' -count=1 -v -timeout 30m`

Exit: **0**. Parent test and all **27 subtests PASS**; no failure messages.

Full output (includes PASS per subtest):

```text
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2()
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2()
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string>
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string>
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string>
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2()
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad (0.00s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string> (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter (0.10s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called (0.10s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on (0.10s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2() (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop (0.05s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2) (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2) (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2) (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2) (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it (0.05s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2) (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop (0.05s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.363s
```

## Mutant

In `internal/lower/generators.go`, asserted that this exact text occurred once:

```go
case of == ir.Number, of == ir.Boolean, of == ir.String, of == ir.Object, of == ir.Array, of == ir.Map, of == ir.Record:
```

Replaced that one occurrence with:

```go
case of == ir.Number, of == ir.Boolean, of == ir.MaybeBoolean, of == ir.String, of == ir.Object, of == ir.Array, of == ir.Map, of == ir.Record:
```

Reran the identical command above. Exit: **1**. **26 subtests PASS, 1 FAIL**.

The subtest **"two generator functions both yielding boolean | undefined (S3c-2)" is red**. Its whole failure message is preserved in the full output below, along with every subtest result:

```text
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2()
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2()
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string>
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string>
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2)
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2)
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it
=== RUN   TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it
=== PAUSE TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string>
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2)
=== NAME  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2)
    iterable_generators_test.go:141: got /tmp/adamic-gate/TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoadtwo_ge4294150969/001/main.a:3:98: stage 0 can't lower for...of over an object yet, want a NotYet saying "yielding boolean | undefined (each yielded value must be one present slot)"
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2()
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2)
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it
=== CONT  TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break
--- FAIL: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad (0.00s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string> (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it (0.07s)
    --- FAIL: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_both_yielding_boolean_|_undefined_(S3c-2) (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it (0.11s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2) (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2) (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it (0.08s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2) (0.09s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2() (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop (0.09s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2) (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it (0.05s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other (0.03s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop (0.07s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on (0.06s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break (0.04s)
    --- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.355s
FAIL
```

## Restoration

Restored `internal/lower/generators.go` byte-for-byte from its pre-mutation copy. `git diff --exit-code` then exited **0** and printed nothing, before creating this report. No source changes remain.

This report is the only file to be committed, on branch `proof/rulings-amend-report`, created from the base commit above.

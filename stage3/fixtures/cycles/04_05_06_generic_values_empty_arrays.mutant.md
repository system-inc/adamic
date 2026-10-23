Mutant evidence for cycles fixtures 04, 05 and 06's stage-0 records in status.json.

What moved: this change lowers a generic function used as a value and an empty array of never. Fixture 04 (tsc's path.ts startsWithDirectory, passed the generic `identity` as its callback) and fixtures 05 and 06 (tsc's core.ts `emptyArray: never[] = [] as never[]`, read across an import cycle) move from not yet to Compiles, and the test's native run agrees with Node on each, including 06's ReferenceError exit 70. The fixtures are unchanged.

Mutant: the change's lowering undone. Its 18 non-test Go files under internal/lower and internal/native go back to main e50e502c21: the 14 main has are restored to main's bytes, and the 4 it adds (internal/lower/empty_arrays.go, internal/lower/generic_function_value.go, internal/lower/phantom_brands.go, internal/native/phantom_brands.go) are removed; the change's status.json stays. The files: internal/lower/callback_widening.go internal/lower/cast_proof.go internal/lower/census_small.go internal/lower/class_inheritance.go internal/lower/empty_arrays.go internal/lower/exceptions.go internal/lower/expression.go internal/lower/generic.go internal/lower/generic_function_value.go internal/lower/invariance.go internal/lower/library_string.go internal/lower/object.go internal/lower/phantom_brands.go internal/lower/proven_relations.go internal/lower/refusals.go internal/lower/statements.go internal/native/emit_expressions.go internal/native/phantom_brands.go.
Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesCycles$' -v
Result with the mutant: exit 1, red on exactly 04_directory_callback/main.a, 05_safe_load_read/main.a and 06_import_order_mutant/main.a. Restored, the same command passes, each of the three with its native subtest (darwin/arm64, Node v24.14.1).

    fixtures_test.go:474: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/cycles/04_directory_callback/main.a:9:71: stage 0 can't lower a generic function as a value yet Stdout:}; recorded {Outcome:Compiles What: Stdout:}
    fixtures_test.go:474: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/cycles/05_safe_load_read/core.a:4:36: stage 0 can't lower an array of never yet Stdout:}; recorded {Outcome:Compiles What: Stdout:}
    fixtures_test.go:474: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/cycles/06_import_order_mutant/core.a:4:36: stage 0 can't lower an array of never yet Stdout:}; recorded {Outcome:Compiles What: Stdout:}
--- FAIL: TestFixturesCycles (1.26s)
    --- FAIL: TestFixturesCycles/cycles (0.00s)
        --- FAIL: TestFixturesCycles/cycles/04_directory_callback/main.a (0.20s)
            --- FAIL: TestFixturesCycles/cycles/04_directory_callback/main.a/stage0 (0.00s)
        --- FAIL: TestFixturesCycles/cycles/05_safe_load_read/main.a (0.25s)
            --- FAIL: TestFixturesCycles/cycles/05_safe_load_read/main.a/stage0 (0.00s)
        --- FAIL: TestFixturesCycles/cycles/06_import_order_mutant/main.a (0.33s)
            --- FAIL: TestFixturesCycles/cycles/06_import_order_mutant/main.a/stage0 (0.00s)

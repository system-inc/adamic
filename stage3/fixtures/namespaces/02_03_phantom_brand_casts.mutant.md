Mutant evidence for namespaces fixtures 02 and 03's stage-0 records in status.json.

What moved: this change proves phantom primitive brands, so a cast from a primitive to a brand whose members are all void is a proven upcast. Fixtures 02 and 03 (tsc's checker.ts JsxNames and ReactNames, each `"..." as __String` with `__String = string & { __escapedIdentifier: void }`) move from the unchecked cast refusal to Compiles, and the test's native run agrees with Node on each. The fixtures are unchanged.

Mutant: the change's lowering undone. Its 18 non-test Go files under internal/lower and internal/native go back to main e50e502c21: the 14 main has are restored to main's bytes, and the 4 it adds (internal/lower/empty_arrays.go, internal/lower/generic_function_value.go, internal/lower/phantom_brands.go, internal/native/phantom_brands.go) are removed; the change's status.json stays. The files: internal/lower/callback_widening.go internal/lower/cast_proof.go internal/lower/census_small.go internal/lower/class_inheritance.go internal/lower/empty_arrays.go internal/lower/exceptions.go internal/lower/expression.go internal/lower/generic.go internal/lower/generic_function_value.go internal/lower/invariance.go internal/lower/library_string.go internal/lower/object.go internal/lower/phantom_brands.go internal/lower/proven_relations.go internal/lower/refusals.go internal/lower/statements.go internal/native/emit_expressions.go internal/native/phantom_brands.go.
Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesNamespaces$' -v
Result with the mutant: exit 1, red on exactly 02_jsx_names.a and 03_react_names.a. Restored, the same command passes, each of the two with its native subtest (darwin/arm64, Node v24.14.1).

    fixtures_test.go:474: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/namespaces/03_react_names.a:7:29: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) Stdout:}; recorded {Outcome:Compiles What: Stdout:}
    fixtures_test.go:474: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/namespaces/02_jsx_names.a:7:24: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) Stdout:}; recorded {Outcome:Compiles What: Stdout:}
--- FAIL: TestFixturesNamespaces (1.10s)
    --- FAIL: TestFixturesNamespaces/namespaces (0.00s)
        --- FAIL: TestFixturesNamespaces/namespaces/03_react_names.a (0.13s)
            --- FAIL: TestFixturesNamespaces/namespaces/03_react_names.a/stage0 (0.00s)
        --- FAIL: TestFixturesNamespaces/namespaces/02_jsx_names.a (0.23s)
            --- FAIL: TestFixturesNamespaces/namespaces/02_jsx_names.a/stage0 (0.00s)

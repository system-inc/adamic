All twelve host overlays exited 1 with a failed test, without compiler build failure; runner exited 0.

- non-null-check: ^TestCheckedNonNullTypeScript/undefined$
- non-null-narrowing: ^TestCheckedNonNullTypeScript/narrowed$
- nested-empty: ^TestNestedEmptyArrayElementKinds$
- string-concatenation: ^TestNativeAgreesWithNode/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a$
- phantom-primitive: ^TestPhantomCastsAreErased$
- proven-predicates: ^TestCensusComputedRelation$
- generic-empty: ^TestNativeAgreesWithNode/internal/oracle/testdata/generic_empty_array.a$
- debugger: ^TestDebuggerNativeEmitsNothing$
- generic-values: ^TestGenericFunctionValuesHaveSeparateInstances$
- generic-namespace-interaction: ^TestGenericFunctionValueNamespaceDirectCallsKeepIdentity$
- phantom-overload-interaction: ^TestPhantomOverloadOnlyBrandsKeepMainProof$
- proven-cleanup: ^TestCensusGenericRefusalInsideClosure$

Both arrow mutants fail on undeclared C identifiers. Seven host IR mutants are numeric concatenation spelling, boolean spelling, null spelling, undefined spelling, empty array falsiness, fresh array sharing and wrong generic result. All are caught by independent Node output in both backends. Existing non-null mutants also catch bypassed checks and representation look-through.

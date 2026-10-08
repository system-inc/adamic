Mutants and independent catchers

Each group lists the actual mutations and catcher. Linked logs contain all named subtests and observed outputs. Full gates also execute existing permanent mutants; selected new families are listed explicitly. Intentional wrong-output controls are not unmodified compiler findings.

Excluded attempts: item09 interrupted first source run; item10 initial scratch dependency/harness setup; item11 interrupted contextual run; item12 first numeric fixture did not exercise the changed path; item12a first closure overlay failed Go compilation before correction; item13 first phantom-array regex ran no tests; initial closure direct Narrow mutant survived, and one later shell chose the wrong Go executable. These are not successful proof claims.

### 01

wrong carrier, overload adapter, omitted argument, omitted method and captured-cycle ownership; Node comparisons and LSan

[01-mutants](/tmp/stage3-front-3-01-mutants.log)

```text
=== RUN   TestNestedSiblingCycleMutantIsCaught
=== PAUSE TestNestedSiblingCycleMutantIsCaught
=== RUN   TestNestedCallbackCarrierMutantIsCaught
    nested_functions_test.go:177: wrong carrier caught by Node: stdout differs
--- PASS: TestNestedCallbackCarrierMutantIsCaught (17.41s)
=== RUN   TestOmittedArgumentZeroMutantIsCaught
=== PAUSE TestOmittedArgumentZeroMutantIsCaught
=== RUN   TestOmittedReaderZeroMutantIsCaught
=== PAUSE TestOmittedReaderZeroMutantIsCaught
=== RUN   TestScannerNestedOverloadImplementationMutantIsCaught
    scanner_nested_overload_test.go:58: implementation mutant caught by Node: native "\n", Node "x\n"
--- PASS: TestScannerNestedOverloadImplementationMutantIsCaught (0.63s)
=== CONT  TestNestedSiblingCycleMutantIsCaught
=== CONT  TestOmittedReaderZeroMutantIsCaught
=== CONT  TestOmittedArgumentZeroMutantIsCaught
=== NAME  TestOmittedReaderZeroMutantIsCaught
    omitted_arguments_test.go:136: method zero padding caught by Node: native "false 0\n"
--- PASS: TestOmittedReaderZeroMutantIsCaught (0.97s)
=== NAME  TestOmittedArgumentZeroMutantIsCaught
    omitted_arguments_test.go:69: zero padding caught by Node: native "0\n", Node 11
--- PASS: TestOmittedArgumentZeroMutantIsCaught (1.08s)
=== NAME  TestNestedSiblingCycleMutantIsCaught
    nested_functions_test.go:92: sibling capture mutant caught only by the leak check:
        exit 1

        =================================================================
        ==6554==ERROR: LeakSanitizer: detected memory leaks

        Indirect leak of 176 byte(s) in 3 object(s) allocated from:
            #0 0x55fb43c90244 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55fb43ce6b5f in adamic_allocate /home/agent/.cache/adamic/runtime/.build-2617284180/heap.c:196:10
            #2 0x55fb43cd38fa in main /tmp/adamic-gate/adamic-build-3019233852/main.c:129:40

        Indirect leak of 160 byte(s) in 4 object(s) allocated from:
            #0 0x55fb43c90244 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55fb43ce6b5f in adamic_allocate /home/agent/.cache/adamic/runtime/.build-2617284180/heap.c:196:10

        Indirect leak of 112 byte(s) in 2 object(s) allocated from:
            #0 0x55fb43c90244 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55fb43ce6b5f in adamic_allocate /home/agent/.cache/adamic/runtime/.build-2617284180/heap.c:196:10
            #2 0x55fb43cd391f in main /tmp/adamic-gate/adamic-build-3019233852/main.c:132:40

        SUMMARY: AddressSanitizer: 448 byte(s) leaked in 9 allocation(s).
--- PASS: TestNestedSiblingCycleMutantIsCaught (1.20s)
PASS
gate cache: native hits=0 misses=4
gate cache: node hits=0 misses=5
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	19.260s
```

### 02

sibling, generic sibling, ancestor and stable function identity; source Node in both backends

[02-mutants](/tmp/stage3-front-3-02-mutants.log)

```text
=== RUN   TestScannerNestedReferenceMutants
=== RUN   TestScannerNestedReferenceMutants/nested-sibling-callback
    scanner_nested_references_test.go:70: native mutant caught: output "" next to Node "x\n"
    scanner_nested_references_test.go:70: JavaScript mutant caught: output "" next to Node "x\n"
=== RUN   TestScannerNestedReferenceMutants/nested-generic-sibling-call
    scanner_nested_references_test.go:70: JavaScript mutant caught: output "" next to Node "x\n"
    scanner_nested_references_test.go:70: native mutant caught: output "" next to Node "x\n"
=== RUN   TestScannerNestedReferenceMutants/nested-ancestor-call
    scanner_nested_references_test.go:70: native mutant caught: output "0\n" next to Node "1\n"
    scanner_nested_references_test.go:70: JavaScript mutant caught: output "0\n" next to Node "1\n"
--- PASS: TestScannerNestedReferenceMutants (2.35s)
    --- PASS: TestScannerNestedReferenceMutants/nested-sibling-callback (0.80s)
    --- PASS: TestScannerNestedReferenceMutants/nested-generic-sibling-call (0.82s)
    --- PASS: TestScannerNestedReferenceMutants/nested-ancestor-call (0.73s)
=== RUN   TestNestedReferenceIdentityMutant
    scanner_nested_references_test.go:140: native identity mutant caught: output "false false false false\n11 12 21\nfalse 7\n32 33\n" next to Node "true true true false\n11 12 21\ntrue 7\n32 33\n"
    scanner_nested_references_test.go:140: JavaScript identity mutant caught: output "false false false false\n11 12 21\nfalse 7\n32 33\n" next to Node "true true true false\n11 12 21\ntrue 7\n32 33\n"
--- PASS: TestNestedReferenceIdentityMutant (0.75s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=8
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	3.124s
```

### 03

unready read, write, fallthrough and function hoisting; source Node in both backends

[03-mutants](/tmp/stage3-front-3-03-mutants.log)

```text
=== RUN   TestSwitchCaseDeclarationMutants
=== PAUSE TestSwitchCaseDeclarationMutants
=== CONT  TestSwitchCaseDeclarationMutants
=== RUN   TestSwitchCaseDeclarationMutants/unready_read
=== PAUSE TestSwitchCaseDeclarationMutants/unready_read
=== RUN   TestSwitchCaseDeclarationMutants/unready_write
=== PAUSE TestSwitchCaseDeclarationMutants/unready_write
=== RUN   TestSwitchCaseDeclarationMutants/no_fallthrough
=== PAUSE TestSwitchCaseDeclarationMutants/no_fallthrough
=== RUN   TestSwitchCaseDeclarationMutants/function_not_hoisted
=== PAUSE TestSwitchCaseDeclarationMutants/function_not_hoisted
=== CONT  TestSwitchCaseDeclarationMutants/unready_read
=== CONT  TestSwitchCaseDeclarationMutants/no_fallthrough
=== CONT  TestSwitchCaseDeclarationMutants/unready_write
=== CONT  TestSwitchCaseDeclarationMutants/function_not_hoisted
=== NAME  TestSwitchCaseDeclarationMutants/unready_write
    switch_case_declaration_test.go:122: Node stdout "before\nright side\n" exit 70; native stdout "before\nright side\nunreachable\n" exit 0; JavaScript stdout "before\nright side\nunreachable\n" exit 0
=== NAME  TestSwitchCaseDeclarationMutants/unready_read
    switch_case_declaration_test.go:122: Node stdout "value0\nbefore\n" exit 70; native stdout "value0\nbefore\n\nunreachable\n" exit 0; JavaScript stdout "value0\nbefore\n\nunreachable\n" exit 0
=== NAME  TestSwitchCaseDeclarationMutants/function_not_hoisted
    switch_case_declaration_test.go:122: Node stdout "value1!1 value0!0 other\n2 6\n" exit 0; native stdout "value1!1 value0!0 other\n" exit 70; JavaScript stdout "value1!1 value0!0 other\n" exit 70
=== NAME  TestSwitchCaseDeclarationMutants/no_fallthrough
    switch_case_declaration_test.go:122: Node stdout "zero box0:1 default tail0\nbox1:2 default tail1\ntail2\ndefault tail3\nzero box0:1 default tail0\n" exit 0; native stdout "zero \nbox1:2 \ntail2\ndefault \nzero \n" exit 0; JavaScript stdout "zero \nbox1:2 \ntail2\ndefault \nzero \n" exit 0
--- PASS: TestSwitchCaseDeclarationMutants (0.00s)
    --- PASS: TestSwitchCaseDeclarationMutants/unready_write (1.28s)
    --- PASS: TestSwitchCaseDeclarationMutants/unready_read (1.34s)
    --- PASS: TestSwitchCaseDeclarationMutants/function_not_hoisted (1.44s)
    --- PASS: TestSwitchCaseDeclarationMutants/no_fallthrough (1.46s)
PASS
gate cache: native hits=0 misses=7
gate cache: node hits=0 misses=8
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	1.489s
```

### 05

undefined converted to zero: Node; disabled resolved mapper: named lowering diagnostic

[05-mutants](/tmp/stage3-front-3-05-mutants.log)

```text
=== RUN   TestGenericOptionalUndefinedMutant
=== PAUSE TestGenericOptionalUndefinedMutant
=== CONT  TestGenericOptionalUndefinedMutant
    generic_optional_results_test.go:46: Node caught zero instead of undefined: "0 7 0\n|wordword|undefined\nitemitem missing\n0 0\nuu|undefined\nitemitem missing\n"
--- PASS: TestGenericOptionalUndefinedMutant (0.69s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=1
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	0.703s
```

[05-mapper-mutant](/tmp/stage3-front-3-05-mapper-mutant.log)

```text
=== RUN   TestNativeAgreesWithNode
=== PAUSE TestNativeAgreesWithNode
=== CONT  TestNativeAgreesWithNode
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/generic-optional-callback-result.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/generic-optional-callback-result.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/generic-optional-callback-result.a
    oracle_test.go:727: Lower: /workspace/adamic/internal/oracle/testdata/generic-optional-callback-result.a:2:10: stage 0 can't lower a function returning U | undefined yet
--- FAIL: TestNativeAgreesWithNode (0.02s)
    --- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/generic-optional-callback-result.a (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.083s
FAIL
```

### 06

number, boolean, undefined and null spelling; source Node in both backends

[06-mutants](/tmp/stage3-front-3-06-mutants.log)

```text
=== RUN   TestScannerConcatenationMutant
=== PAUSE TestScannerConcatenationMutant
=== RUN   TestScannerPrimitiveConcatenationMutants
=== RUN   TestScannerPrimitiveConcatenationMutants/boolean
=== PAUSE TestScannerPrimitiveConcatenationMutants/boolean
=== RUN   TestScannerPrimitiveConcatenationMutants/undefined
=== PAUSE TestScannerPrimitiveConcatenationMutants/undefined
=== RUN   TestScannerPrimitiveConcatenationMutants/null
=== PAUSE TestScannerPrimitiveConcatenationMutants/null
=== CONT  TestScannerPrimitiveConcatenationMutants/boolean
=== CONT  TestScannerPrimitiveConcatenationMutants/null
=== CONT  TestScannerPrimitiveConcatenationMutants/undefined
=== NAME  TestScannerPrimitiveConcatenationMutants/null
    scanner_expressions_test.go:110: null spelling mutant caught only by stdout in both backends
=== NAME  TestScannerPrimitiveConcatenationMutants/undefined
    scanner_expressions_test.go:110: undefined spelling mutant caught only by stdout in both backends
=== NAME  TestScannerPrimitiveConcatenationMutants/boolean
    scanner_expressions_test.go:110: boolean spelling mutant caught only by stdout in both backends
--- PASS: TestScannerPrimitiveConcatenationMutants (0.00s)
    --- PASS: TestScannerPrimitiveConcatenationMutants/null (1.25s)
    --- PASS: TestScannerPrimitiveConcatenationMutants/undefined (1.25s)
    --- PASS: TestScannerPrimitiveConcatenationMutants/boolean (1.36s)
=== CONT  TestScannerConcatenationMutant
--- PASS: TestScannerConcatenationMutant (0.80s)
PASS
gate cache: native hits=0 misses=8
gate cache: node hits=0 misses=8
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	2.180s
```

### 07

UTF-16 length off by one; Node output

[07-mutants](/tmp/stage3-front-3-07-mutants.log)

```text
=== RUN   TestScannerStringDestructuringMutant
=== PAUSE TestScannerStringDestructuringMutant
=== CONT  TestScannerStringDestructuringMutant
    scanner_string_destructuring_test.go:51: Node prints 10; length + 1 mutant prints 11; caught by stdout comparison
--- PASS: TestScannerStringDestructuringMutant (0.98s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=1
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	0.999s
```

### 08

postfix returns new value, field target twice, array target twice; source Node stdout in both backends

[08-mutants](/tmp/stage3-front-3-08-mutants.log)

```text
postfix-return-new: caught by Node versus native and JavaScript stdout
field-target-twice: caught by Node versus native and JavaScript stdout
array-target-twice: caught by Node versus native and JavaScript stdout
```

### 09

thirteen source overlays for instantiation, conversion, unseeded/outer generics, mixed unions, optional calls, hoisting, specialization key, spread, captures, property observation, erased cast and typeof; lowering/IR pins or Node; separate wrong-result Node mutant

[09-source-mutants-final](/tmp/stage3-front-3-09-source-mutants-final.log)

```text
instantiation caught: Lower:
conversion caught: want a diagnostic, got <nil>
unseeded caught: want a diagnostic, got <nil>
outer-instantiation caught: want a diagnostic, got <nil>
mixed-union caught: want a diagnostic, got <nil>
optional-call caught: want a diagnostic, got <nil>
hoisting caught: Lower:
slot-key caught: want number/string/object bodies
spread caught: want a diagnostic, got <nil>
captured-locals caught: Lower:
property-observation caught: want a diagnostic, got <nil>
erased-cast caught: want a diagnostic, got <nil>
typeof caught: stdout differs
```

[09-result-mutant](/tmp/stage3-front-3-09-result-mutant.log)

```text
=== RUN   TestGenericFunctionPropertyMutant
=== PAUSE TestGenericFunctionPropertyMutant
=== CONT  TestGenericFunctionPropertyMutant
    generic_function_properties_test.go:68: native mutant caught only by Node stdout
    generic_function_properties_test.go:68: JavaScript mutant caught only by Node stdout
--- PASS: TestGenericFunctionPropertyMutant (1.57s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	1.622s
```

### 10

sixteen readiness variants, four non-null storage variants, logical RHS eagerness, read-half check, index twice, ordinary nullish initializer, eager lazy initializer, and numeric-NULL artifact invariant; pinned stop/Node/IR; twelve typed-runtime mutants: UBSan, ASan, LSan, Node and exact runtime panic; three Uint16 wrapping/bounds/copy mutants: Node/ASan

[10-readiness-mutants](/tmp/stage3-front-3-10-readiness-mutants.log)

```text
=== RUN   TestNonNullLogicalAlwaysEvaluateRightMutant
=== PAUSE TestNonNullLogicalAlwaysEvaluateRightMutant
=== RUN   TestNonNullNarrowedPointerTestMutant
=== PAUSE TestNonNullNarrowedPointerTestMutant
=== RUN   TestNonNullStorageReadinessMutants
=== PAUSE TestNonNullStorageReadinessMutants
=== RUN   TestNonNullWriteReadHalfMutant
=== PAUSE TestNonNullWriteReadHalfMutant
=== RUN   TestNonNullWriteIndexOnceMutant
=== PAUSE TestNonNullWriteIndexOnceMutant
=== RUN   TestReadinessMutants
=== PAUSE TestReadinessMutants
=== RUN   TestDefiniteAssignmentIsNotNullishMutant
=== PAUSE TestDefiniteAssignmentIsNotNullishMutant
=== RUN   TestLazyInitializerIsNotEagerMutant
=== PAUSE TestLazyInitializerIsNotEagerMutant
=== CONT  TestNonNullLogicalAlwaysEvaluateRightMutant
=== CONT  TestNonNullWriteIndexOnceMutant
=== CONT  TestNonNullStorageReadinessMutants
=== RUN   TestNonNullStorageReadinessMutants/local
=== PAUSE TestNonNullStorageReadinessMutants/local
=== RUN   TestNonNullStorageReadinessMutants/field
=== PAUSE TestNonNullStorageReadinessMutants/field
=== RUN   TestNonNullStorageReadinessMutants/static
=== PAUSE TestNonNullStorageReadinessMutants/static
=== RUN   TestNonNullStorageReadinessMutants/identifier
=== PAUSE TestNonNullStorageReadinessMutants/identifier
=== CONT  TestNonNullNarrowedPointerTestMutant
=== CONT  TestDefiniteAssignmentIsNotNullishMutant
=== CONT  TestNonNullWriteReadHalfMutant
=== NAME  TestNonNullNarrowedPointerTestMutant
    non_null_narrowed_test.go:132: restored pointer test caught before clang: adamic_temporary_1 != NULL
--- PASS: TestNonNullNarrowedPointerTestMutant (1.48s)
=== CONT  TestLazyInitializerIsNotEagerMutant
=== NAME  TestNonNullWriteReadHalfMutant
    non_null_write_test.go:53: read-half check removed: exit 0 stdout "before\nright\n2\n", caught by pinned panic
--- PASS: TestNonNullWriteReadHalfMutant (1.86s)
=== CONT  TestReadinessMutants
=== RUN   TestReadinessMutants/deinitialization-values
=== PAUSE TestReadinessMutants/deinitialization-values
=== RUN   TestReadinessMutants/deinitialization-entries
=== PAUSE TestReadinessMutants/deinitialization-entries
=== RUN   TestReadinessMutants/deinitialization-assign-source
=== PAUSE TestReadinessMutants/deinitialization-assign-source
=== RUN   TestReadinessMutants/deinitialization-field-alias
=== PAUSE TestReadinessMutants/deinitialization-field-alias
=== RUN   TestReadinessMutants/deinitialization-field-method
=== PAUSE TestReadinessMutants/deinitialization-field-method
=== RUN   TestReadinessMutants/keep-slot-proven-after-deinitialization
=== PAUSE TestReadinessMutants/keep-slot-proven-after-deinitialization
=== RUN   TestReadinessMutants/deinitialization-through-capture
=== PAUSE TestReadinessMutants/deinitialization-through-capture
=== RUN   TestReadinessMutants/deinitialization-exception-path
=== PAUSE TestReadinessMutants/deinitialization-exception-path
=== RUN   TestReadinessMutants/scanner-var-capture
=== PAUSE TestReadinessMutants/scanner-var-capture
=== RUN   TestReadinessMutants/drop-check
=== PAUSE TestReadinessMutants/drop-check
=== RUN   TestReadinessMutants/erase-without-proof
=== PAUSE TestReadinessMutants/erase-without-proof
=== RUN   TestReadinessMutants/initialize-to-zero
=== PAUSE TestReadinessMutants/initialize-to-zero
=== RUN   TestReadinessMutants/miss-captured-read
=== PAUSE TestReadinessMutants/miss-captured-read
=== RUN   TestReadinessMutants/miss-exception-path
=== PAUSE TestReadinessMutants/miss-exception-path
=== RUN   TestReadinessMutants/lazy-read
=== PAUSE TestReadinessMutants/lazy-read
=== RUN   TestReadinessMutants/weak-generic-message
=== PAUSE TestReadinessMutants/weak-generic-message
=== CONT  TestNonNullStorageReadinessMutants/local
=== NAME  TestNonNullWriteIndexOnceMutant
    non_null_write_test.go:96: index evaluated twice: exit 0 stdout "2 4 5\n" versus Node "1 5 8\n"
--- PASS: TestNonNullWriteIndexOnceMutant (2.33s)
=== CONT  TestNonNullStorageReadinessMutants/identifier
=== NAME  TestNonNullLogicalAlwaysEvaluateRightMutant
    non_null_logical_test.go:102: always-evaluate RHS caught by Node: stdout "3 0\n" versus "0 0\n", exit 0
--- PASS: TestNonNullLogicalAlwaysEvaluateRightMutant (2.59s)
=== CONT  TestNonNullStorageReadinessMutants/field
=== NAME  TestDefiniteAssignmentIsNotNullishMutant
    readiness_test.go:212: ordinary-nullish initializer mutant caught by Node output: adamic: panic: non-null assertion failed: number is null or undefined
--- PASS: TestDefiniteAssignmentIsNotNullishMutant (3.04s)
=== CONT  TestNonNullStorageReadinessMutants/static
=== NAME  TestLazyInitializerIsNotEagerMutant
    readiness_test.go:335: eager initializer caught by Node output: adamic: panic: non-null assertion failed: textInitial! is null or undefined
--- PASS: TestLazyInitializerIsNotEagerMutant (2.31s)
=== CONT  TestReadinessMutants/deinitialization-values
=== NAME  TestNonNullStorageReadinessMutants/local
    non_null_storage_test.go:85: drop local readiness caught: stdout "before\n0\n", exit 0
=== CONT  TestReadinessMutants/scanner-var-capture
=== NAME  TestNonNullStorageReadinessMutants/identifier
    non_null_storage_test.go:85: drop identifier readiness caught: stdout "before\ntrue\n", exit 0
=== CONT  TestReadinessMutants/weak-generic-message
=== NAME  TestNonNullStorageReadinessMutants/field
    non_null_storage_test.go:85: drop field readiness caught: stdout "before\n0\n", exit 0
=== CONT  TestReadinessMutants/lazy-read
=== NAME  TestNonNullStorageReadinessMutants/static
    non_null_storage_test.go:85: drop static readiness caught: stdout "before\n0\n", exit 0
--- PASS: TestNonNullStorageReadinessMutants (0.00s)
    --- PASS: TestNonNullStorageReadinessMutants/local (2.62s)
    --- PASS: TestNonNullStorageReadinessMutants/identifier (2.65s)
    --- PASS: TestNonNullStorageReadinessMutants/field (2.40s)
    --- PASS: TestNonNullStorageReadinessMutants/static (2.49s)
=== CONT  TestReadinessMutants/miss-exception-path
=== NAME  TestReadinessMutants/deinitialization-values
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n1\n" stderr ""
=== CONT  TestReadinessMutants/miss-captured-read
=== NAME  TestReadinessMutants/scanner-var-capture
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\nfalse\n" stderr ""
=== CONT  TestReadinessMutants/initialize-to-zero
=== NAME  TestReadinessMutants/weak-generic-message
    readiness_test.go:171: caught by pinned output: exit 70 stdout "before\n" stderr "adamic: panic: a weak reference was read after what it pointed to was freed\n"
=== CONT  TestReadinessMutants/erase-without-proof
=== NAME  TestReadinessMutants/lazy-read
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n\n" stderr ""
=== CONT  TestReadinessMutants/drop-check
=== NAME  TestReadinessMutants/miss-exception-path
    readiness_test.go:171: caught by pinned output: exit 0 stdout "caught\n0\n" stderr ""
=== CONT  TestReadinessMutants/deinitialization-field-method
=== NAME  TestReadinessMutants/miss-captured-read
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== CONT  TestReadinessMutants/deinitialization-exception-path
=== NAME  TestReadinessMutants/initialize-to-zero
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== CONT  TestReadinessMutants/deinitialization-through-capture
=== NAME  TestReadinessMutants/erase-without-proof
    readiness_test.go:171: caught by pinned output: exit 0 stdout "0\n" stderr ""
=== CONT  TestReadinessMutants/keep-slot-proven-after-deinitialization
=== NAME  TestReadinessMutants/drop-check
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== CONT  TestReadinessMutants/deinitialization-assign-source
=== NAME  TestReadinessMutants/deinitialization-field-method
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== CONT  TestReadinessMutants/deinitialization-field-alias
=== NAME  TestReadinessMutants/deinitialization-exception-path
    readiness_test.go:171: caught by pinned output: exit 0 stdout "caught\n0\n" stderr ""
=== CONT  TestReadinessMutants/deinitialization-entries
=== NAME  TestReadinessMutants/deinitialization-through-capture
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== NAME  TestReadinessMutants/keep-slot-proven-after-deinitialization
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== NAME  TestReadinessMutants/deinitialization-assign-source
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n" stderr ""
=== NAME  TestReadinessMutants/deinitialization-field-alias
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n0\n" stderr ""
=== NAME  TestReadinessMutants/deinitialization-entries
    readiness_test.go:171: caught by pinned output: exit 0 stdout "before\n1\n" stderr ""
--- PASS: TestReadinessMutants (0.00s)
    --- PASS: TestReadinessMutants/deinitialization-values (2.19s)
    --- PASS: TestReadinessMutants/scanner-var-capture (2.14s)
    --- PASS: TestReadinessMutants/weak-generic-message (2.02s)
    --- PASS: TestReadinessMutants/lazy-read (2.12s)
    --- PASS: TestReadinessMutants/miss-exception-path (1.87s)
    --- PASS: TestReadinessMutants/miss-captured-read (1.65s)
    --- PASS: TestReadinessMutants/initialize-to-zero (1.69s)
    --- PASS: TestReadinessMutants/erase-without-proof (1.78s)
    --- PASS: TestReadinessMutants/drop-check (1.89s)
    --- PASS: TestReadinessMutants/deinitialization-field-method (1.81s)
    --- PASS: TestReadinessMutants/deinitialization-exception-path (1.99s)
    --- PASS: TestReadinessMutants/deinitialization-through-capture (1.43s)
    --- PASS: TestReadinessMutants/keep-slot-proven-after-deinitialization (1.62s)
    --- PASS: TestReadinessMutants/deinitialization-assign-source (1.50s)
    --- PASS: TestReadinessMutants/deinitialization-field-alias (1.68s)
    --- PASS: TestReadinessMutants/deinitialization-entries (1.41s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	11.070s
```

[10-typed-runtime-mutants](/tmp/stage3-front-3-10-typed-runtime-mutants.log)

```text
missing-wrap: caught by UBSan
missing-write-bounds: caught by ASan
copying-subarray: caught by Node comparison
missing-view-retain: caught by ASan
missing-iterator-retain: caught by ASan
missing-buffer-free: caught by LeakSanitizer
overlap-memcpy: caught by ASan
missing-kind-check: caught by panic assertion
missing-offset-check: caught by UBSan
missing-set-capacity-check: caught by ASan
missing-length-check: caught by UBSan
fractional-index-accepted: caught by runtime assertion
all 12 caught; logs: /tmp/adamic-typed-array-mutants
```

[10-uint16-mutants](/tmp/stage3-front-3-10-uint16-mutants.log)

```text
wrap C: caught; see /tmp/adamic-uint16-mutants/wrap-C.log
wrap backends: caught; see /tmp/adamic-uint16-mutants/wrap-backends.log
bounds C: caught; see /tmp/adamic-uint16-mutants/bounds-C.log
bounds backends: caught; see /tmp/adamic-uint16-mutants/bounds-backends.log
copying-subarray C: caught; see /tmp/adamic-uint16-mutants/copying-subarray-C.log
copying-subarray backends: caught; see /tmp/adamic-uint16-mutants/copying-subarray-backends.log
all 3 caught by C tests and backend oracles
```

### 11

six source overlays: bottom view, readonly destination, object layout, mutable alias, scalar fallback and contextual scope; refusal/IR pins. Empty array falsiness and shared literal identity: Node

[11-source-mutants](/tmp/stage3-front-3-11-source-mutants.log)

```text
bottom-view: caught by Lower:
readonly-destination: caught by Lower:
object-layout: caught by want number and object instantiations
mutable-bottom: caught by want refusal, got <nil>
scalar-guard: caught by want scalar fallback NotYet, got <nil>
```

[11-contextual-scope-resumed](/tmp/stage3-front-3-11-contextual-scope-resumed.log)

```text
--- FAIL: TestNativeAgreesWithNode (0.06s)
    --- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/gaps.a (0.10s)
        oracle_test.go:736: Lower: /workspace/adamic/internal/oracle/testdata/gaps.a:38:30: Adamic 0.1 refuses a value of type Point | undefined seen as { label: any; x: any; }, whose readonly field label becomes writable: a readonly field may hold something narrower than string, which a write of any would replace; keep label readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.180s
FAIL
```

[11-runtime-mutants](/tmp/stage3-front-3-11-runtime-mutants.log)

```text
=== RUN   TestEmptyArrayRuntimeMutants
=== PAUSE TestEmptyArrayRuntimeMutants
=== CONT  TestEmptyArrayRuntimeMutants
=== RUN   TestEmptyArrayRuntimeMutants/empty_array_is_falsy
=== PAUSE TestEmptyArrayRuntimeMutants/empty_array_is_falsy
=== RUN   TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared
=== PAUSE TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared
=== CONT  TestEmptyArrayRuntimeMutants/empty_array_is_falsy
=== CONT  TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared
    empty_arrays_test.go:89: Node caught fresh number array is shared: "0 0 true true\ntrue true\ntrue false\n9 freshfresh\n1 0 true\ntrue true\ntrue 0 2 1\n1\n"
=== NAME  TestEmptyArrayRuntimeMutants/empty_array_is_falsy
    empty_arrays_test.go:89: Node caught empty array is falsy: "0 0 true true\ntrue true\nfalse true\n9 freshfresh\n0 0 false\ntrue true\ntrue 0 2 1\n0\n"
--- PASS: TestEmptyArrayRuntimeMutants (0.00s)
    --- PASS: TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared (1.38s)
    --- PASS: TestEmptyArrayRuntimeMutants/empty_array_is_falsy (1.39s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	1.420s
```

### 12

proof-report erasure and opaque callback: source pins; numeric truthiness inversion: Node; checked-cast skip/wrong tag, twice operand, numeric/bool check erasure: Node/pinned stop; actual inherited slot metadata index and unchecked union storage refusal: runtime/refusal assertions

[12-source-mutants](/tmp/stage3-front-3-12-source-mutants.log)

```text
drop-false-report: caught by explanation:
trust-inferred-callback: caught by want pinned callback argument refusal
Traceback (most recent call last):
  File "/tmp/stage3-front-3-12-mutants.py", line 12, in <module>
    s=log.read_text();assert r.returncode==1 and expect in s and '[build failed]' not in s and 'clang failed' not in s,(name,r.returncode,s);assert p.read_text()==original;print(name+': caught by '+expect,flush=True)
                             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
AssertionError: ('number-truthiness', 0, 'ok  \tgithub.com/system-inc/adamic/internal/oracle\t1.540s\n')
```

[12-numeric-mutant](/tmp/stage3-front-3-12-numeric-mutant.log)

```text
=== RUN   TestNativeAgreesWithNode
=== PAUSE TestNativeAgreesWithNode
=== CONT  TestNativeAgreesWithNode
=== RUN   TestNativeAgreesWithNode/dedication/dedication.a
=== PAUSE TestNativeAgreesWithNode/dedication/dedication.a
=== RUN   TestNativeAgreesWithNode/../../tmp/stage3-front-3-12-numeric-condition.a
=== PAUSE TestNativeAgreesWithNode/../../tmp/stage3-front-3-12-numeric-condition.a
=== CONT  TestNativeAgreesWithNode/dedication/dedication.a
=== CONT  TestNativeAgreesWithNode/../../tmp/stage3-front-3-12-numeric-condition.a
    oracle_test.go:759: stdout differs
        node:   exit 0, stdout "no\nno\nyes\nyes\nno\nyes\n", stderr ""
        native: exit 0, stdout "yes\nyes\nno\nno\nno\nno\n", stderr ""
--- FAIL: TestNativeAgreesWithNode (0.05s)
    --- PASS: TestNativeAgreesWithNode/dedication/dedication.a (1.55s)
    --- FAIL: TestNativeAgreesWithNode/../../tmp/stage3-front-3-12-numeric-condition.a (1.87s)
FAIL
gate cache: native hits=0 misses=6
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	1.973s
FAIL
```

[12-runtime-mutants](/tmp/stage3-front-3-12-runtime-mutants.log)

```text
=== RUN   TestInterfaceCastRuntimeMutants
=== RUN   TestInterfaceCastRuntimeMutants/skip_tag
    interface_cast_test.go:130: caught skip tag: exit codes differ; expected exit 70 stdout "casting\n", mutant exit 0 stdout "casting\notherother\n" stderr ""
=== RUN   TestInterfaceCastRuntimeMutants/wrong_tag
    interface_cast_test.go:130: caught wrong tag: exit codes differ; expected exit 0 stdout "idid\n42\ntexttext\n5\ntrue\nonce2\ncalls 2\n", mutant exit 70 stdout "idid\n42\ntexttext\n5\n" stderr "adamic: panic: cast failed: this Node is not a Identifier\n"
=== RUN   TestInterfaceCastRuntimeMutants/twice_operand
    interface_cast_test.go:130: caught twice operand: stdout differs; expected exit 0 stdout "idid\n42\ntexttext\n5\ntrue\nonce2\ncalls 2\n", mutant exit 0 stdout "idid\n42\ntexttext\n5\ntrue\nonce3\ncalls 3\n" stderr ""
--- PASS: TestInterfaceCastRuntimeMutants (1.54s)
    --- PASS: TestInterfaceCastRuntimeMutants/skip_tag (0.61s)
    --- PASS: TestInterfaceCastRuntimeMutants/wrong_tag (0.48s)
    --- PASS: TestInterfaceCastRuntimeMutants/twice_operand (0.44s)
=== RUN   TestInterfaceCastScalarTags
=== RUN   TestInterfaceCastScalarTags/number_passes
=== RUN   TestInterfaceCastScalarTags/number_fails
    interface_cast_test.go:231: caught skip 4 | 9 tag: exit codes differ; checked exit 70, mutant exit 0 stdout "casting\npayloadpayload\n"
=== RUN   TestInterfaceCastScalarTags/boolean_passes
=== RUN   TestInterfaceCastScalarTags/boolean_fails
    interface_cast_test.go:231: caught skip boolean tag: exit codes differ; checked exit 70, mutant exit 0 stdout "casting\npayloadpayload\n"
--- PASS: TestInterfaceCastScalarTags (2.73s)
    --- PASS: TestInterfaceCastScalarTags/number_passes (0.43s)
    --- PASS: TestInterfaceCastScalarTags/number_fails (0.62s)
    --- PASS: TestInterfaceCastScalarTags/boolean_passes (0.50s)
    --- PASS: TestInterfaceCastScalarTags/boolean_fails (1.18s)
PASS
gate cache: native hits=0 misses=15
gate cache: node hits=0 misses=12
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	4.303s
```

[12-fix-mutants](/tmp/stage3-front-3-12-fix-mutants.log)

```text
stale-slot-index: caught by inherited field owner
unchecked-union-storage: caught by want unsupported structural union slot refusal
```

### 12a

qualified cast, widening, computed key and early closure context restoration; three census panic regression assertions plus corrected compiling closure overlay

[12a-mutants](/tmp/stage3-front-3-12a-mutants.log)

```text
qualified-cast: caught by TestCensusQualifiedCast panic regression
qualified-widening: caught by TestCensusQualifiedCast panic regression
computed-key: caught by TestCensusComputedRelation panic regression
Traceback (most recent call last):
  File "/tmp/stage3-front-3-12a-mutants.py", line 8, in <module>
    out=log.read_text();assert r.returncode==1 and 'panic:' in out and '[build failed]' not in out,(name,r.returncode,out);assert p.read_text()==s;print(name+': caught by '+test+' panic regression',flush=True)
                               ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
AssertionError: ('closure-state', 1, '# github.com/system-inc/adamic/internal/lower [github.com/system-inc/adamic/internal/lower.test]\n/tmp/stage3-front-3-12a-closure-state.go:113:49: declared and not used: outerTypeMapper\nFAIL\tgithub.com/system-inc/adamic/internal/lower [build failed]\nFAIL\n')
```

[12a-closure-mutant-rerun](/tmp/stage3-front-3-12a-closure-mutant-rerun.log)

```text
closure-state: caught by TestCensusGenericRefusalInsideClosure panic regression
```

### 12b

callback proof, assertion proof, failure helper, null/undefined specialization and scalar evaluation; source proof/Node. Ordinary and assertion predicate report erasure, mixed union admission, optional widening and uncalled closed-caller proof: report/refusal pins

[12b-mutants](/tmp/stage3-front-3-12b-mutants.log)

```text
callback: caught by TestParserCallbackLie
assertion: caught by TestParserAssertionProofErasure
failure-helper: caught by TestParserAssertionProofErasure
specialized-null: caught by TestParserAssertionBackends
specialized-undefined: caught by TestParserAssertionBackends
scalar-evaluation: caught by TestParserAssertionBackends
```

[12b-report-mutants](/tmp/stage3-front-3-12b-report-mutants.log)

```text
ordinary-report: caught by TestExplainChecksOutput/ordinary_regions
assertion-report: caught by TestExplainChecksOutput/assert_defined_read
unsafe-union-admission: caught by TestMixedUnionCallbackABIIsPending
```

[12b-optional-mutant](/tmp/stage3-front-3-12b-optional-mutant.log)

```text
optional-widening: caught by TestOptionalWideningRefused
```

[12b-uncalled-contract](/tmp/stage3-front-3-12b-uncalled-contract.log)

```text
--- FAIL: TestPredicateOverloadRuntime (0.37s)
    --- FAIL: TestPredicateOverloadRuntime/parser_callback_parameter (0.37s)
        predicates_overload_test.go:86: want the closed-caller predicate contract refusal, got <nil>
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.385s
FAIL
```

### 14

eight source overlays: mixed named signature, readonly/numeric dictionary, prototype names, partial storage, cycle, mutable invariance and spread; refusal/proof assertions. Permanent full-gate mutants: indices insertion, UINT32 maximum, deleted iteration (Node); overwrite leak (LSan), freed stored key (ASan), null own-slot (UBSan); read/write/delete/in/hasOwn/keys/values/entries/spread/for-in/stringify (Node); missing/prototype reads (checked stops); prototype and narrowing guards (checked stops), ownership (LSan), eager coalescing (Node), absent partial entry in both backends (Node), discarded/own/scalar observations (Node)

[14-static-mutants](/tmp/stage3-front-3-14-static-mutants.log)

```text
mixed-named: caught by TestRecordRefusals
readonly: caught by TestRecordRefusals
numeric: caught by TestRecordRefusals
prototype: caught by TestRecordPrototypeLiteralNames
storage: caught by TestPartialRecordStorageRefusals
cycle: caught by TestRecordRefusals
invariance: caught by TestRecordRefusals
spread: caught by TestRecordRefusals
```

### 15

six source overlays: no memo, partial cycle union, early enum, unreaching known call, const-enum and function runtime readiness. Two merge-fix overlays: parameter AST kind and namespace repeated var. Named work/refusal/panic/Node assertions. Permanent full-gate eight namespace binding/order semantic mutants (Node), lose assignment/hoist (Node), skip readiness (checked JS stop), missing/late parameter-property store (Node), callback-field retain (ASan)

[15-mutants](/tmp/stage3-front-3-15-mutants.log)

```text
no-memo: caught by ^TestNamespaceCallGraphLinearWork$/12$ (nonlinear function walks: got 8191, want 13)
partial-cycle: caught by ^TestNamespaceCallGraphCycleUnion$ (partial cycle reach)
early-enum: caught by ^TestNamespaceEnumInitializationIndependentOfModuleAnalysis$ (namespace enum lost its independent refusal)
known-call-unreaching: caught by ^TestNamespaceInitializationReachability$ (reaching call lost its initialization refusal)
const-enum-readiness: caught by ^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_unknown_enum_before.a$ (exit codes differ)
runtime-readiness: caught by ^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_unknown_function_before.a$ (exit codes differ)
```

[15-fix-mutants](/tmp/stage3-front-3-15-fix-mutants.log)

```text
parameter-ast-kind: caught by ^TestParameterPropertyCheckerContracts$
namespace-repeated-var: caught by ^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_parser_state.a$
```

### 15a

erased enum reach (five refusal pins), no memo (8191 versus13 expansions), erased readiness (exact stop/UBSan), wrong callable branch (Node stdout). Landed in 282874ebcf1d11a19daa78f06a53ef2be37095a2 after fixing generic registration for the three source-Node premature-read stops; all dedicated observations and exact backend pins remain.

[15a-mutants](/tmp/stage3-front-3-15a-mutants.log)

```text
unreaching: caught by ^TestEnumInitializationReach/reaching- (pending enum read on line 2 must stay refused: <nil>)
no-memo: caught by ^TestEnumInitializationGraphMemo/12$ (nonlinear enum function walks: got 8191, want 13)
no-readiness: caught by ^TestEnumInitializationUnknownPinned$ (--- FAIL: TestEnumInitializationUnknownPinned)
wrong-callable-branch: caught by ^TestEnumInitializationNode/callable-selection$ (stdout differs)
```

### 13 SKIPPED CANDIDATE

ten checked-view admission/representation overlays, six runtime checked-view/sort mutants and scoped phantom-array optional normalization; assertion/Node/stop catchers. Candidate was aborted; none is claimed as landed item13 coverage.

[13-mutants](/tmp/stage3-front-3-13-mutants.log)

```text
array-kind: caught by TestViewArraysNode/kind
element: caught by TestViewArraysNode/second
callable-kind: caught by TestViewCallablesNode/kind
signature: caught by TestViewCallableSignature
views-preflight: caught by TestCheckedViewGenericArrayCast
phantom-preflight: caught by TestPhantomSortedArrayProbe
nullable-dispatch: caught by TestCheckedViewParserReturns/native-same-map-return-cast
readonly-consumer: caught by TestCheckedViewParserReturns/native-sorted-empty-conditional
mutable-admission: caught by TestViewArrayConsumerRefusals
undefined-reference-sort: caught by TestViewArrayOptionalComparatorRefusal
```

[13-runtime-mutants](/tmp/stage3-front-3-13-runtime-mutants.log)

```text
array-kind: caught by TestCheckedViewArrays/array-length-kind
element: caught by TestCheckedViewArrays/array-second
signature: caught by TestCheckedViewOpaqueSignature
string-element-check: caught by TestCheckedViewArrays/nullable-array-string
string-literal-check: caught by TestCheckedViewArrays/array-string-literal
numeric-default-sort: caught by TestCheckedViewParserReturns/optional-sort
```

[13-phantom-optional-mutant](/tmp/stage3-front-3-13-phantom-optional-mutant.log)

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
    --- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/phantom_array_brands.a (0.06s)
        oracle_test.go:743: Lower: /workspace/adamic/internal/oracle/testdata/phantom_array_brands.a:11:18: Adamic 0.1 refuses optional property marker in BrandedReadonly absent from structural source readonly number[], which can hide fields; declare marker on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.093s
FAIL
```

### 04 REPLACEMENT SKIPPED CANDIDATE

checked-call erasure is caught structurally. Initial direct Narrow guard erasure SURVIVED because source cases used helpers; not a proof. Subsequent direct IR tests and two compiling overlays catch representation erasure and retained same-reference wrapper; paths below. Focused typed arity and optional-method permanent mutants passed; full candidate is not green and is aborted.

[closure-rule-mutants](/tmp/stage3-front-3-closure-rule-mutants.log)

```text
representation-changing-narrow exit 0 log /tmp/stage3-front-3-closure-rule-mutants/representation-changing-narrow.log
SURVIVED: this fixture does not exercise the changed path
checked-call-erasure exit 1 log /tmp/stage3-front-3-closure-rule-mutants/checked-call-erasure.log
e_number.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_number_minimal.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_reference.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_maybe_number.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_nullable_reference.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_boolean.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_string.a
=== CONT  TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_number.a
--- PASS: TestOracleCDoesNotCompareNumbersWithNull (0.21s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_number_minimal.a (0.93s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_nullable_reference.a (0.88s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_maybe_number.a (0.95s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_reference.a (1.02s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_boolean.a (0.25s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_number.a (0.30s)
    --- PASS: TestOracleCDoesNotCompareNumbersWithNull/internal/oracle/testdata/non_null_narrowed_string.a (0.34s)
=== NAME  TestNonNullNarrowedPointerTestMutant
    non_null_narrowed_test.go:126: pointer-test mutant changed 0 returns
--- FAIL: TestNonNullNarrowedPointerTestMutant (2.24s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	2.309s
FAIL
```

[direct-representation-erasure](/tmp/stage3-front-3-closure-rule-mutants/direct-representation-erasure.log)

```text
=== RUN   TestNonNullNarrowingRepresentationAndChecks
=== RUN   TestNonNullNarrowingRepresentationAndChecks/union-number
=== RUN   TestNonNullNarrowingRepresentationAndChecks/maybe-number
    non_null_narrowing_test.go:64: non-null representation or check changed: got ir.Coalesce{Value:ir.Read{Local:0, Of:7, Checked:false, Readiness:""}, Fallback:ir.Expression(nil), Panic:ir.StringConstant{Index:0}, Of:1}, want ir.Narrow{Value:ir.Read{Local:0, Of:7, Checked:false, Readiness:""}, To:1}
=== RUN   TestNonNullNarrowingRepresentationAndChecks/union-boolean
=== RUN   TestNonNullNarrowingRepresentationAndChecks/same-reference
=== RUN   TestNonNullNarrowingRepresentationAndChecks/checked-call
--- FAIL: TestNonNullNarrowingRepresentationAndChecks (0.24s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/union-number (0.06s)
    --- FAIL: TestNonNullNarrowingRepresentationAndChecks/maybe-number (0.05s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/union-boolean (0.04s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/same-reference (0.05s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/checked-call (0.04s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.258s
FAIL
```

[same-reference-wrapper](/tmp/stage3-front-3-closure-rule-mutants/same-reference-wrapper.log)

```text
=== RUN   TestNonNullNarrowingRepresentationAndChecks
=== RUN   TestNonNullNarrowingRepresentationAndChecks/union-number
=== RUN   TestNonNullNarrowingRepresentationAndChecks/maybe-number
=== RUN   TestNonNullNarrowingRepresentationAndChecks/union-boolean
=== RUN   TestNonNullNarrowingRepresentationAndChecks/same-reference
    non_null_narrowing_test.go:64: non-null representation or check changed: got ir.Narrow{Value:ir.Read{Local:0, Of:4, Checked:false, Readiness:""}, To:4}, want ir.Read{Local:0, Of:4, Checked:false, Readiness:""}
=== RUN   TestNonNullNarrowingRepresentationAndChecks/checked-call
--- FAIL: TestNonNullNarrowingRepresentationAndChecks (0.27s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/union-number (0.06s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/maybe-number (0.05s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/union-boolean (0.07s)
    --- FAIL: TestNonNullNarrowingRepresentationAndChecks/same-reference (0.05s)
    --- PASS: TestNonNullNarrowingRepresentationAndChecks/checked-call (0.04s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.291s
FAIL
```


## Item 16 phantom brands (candidate until its gate and push)

- non-brand-result-accepted: caught by assertion; TestPhantomOverloadLiteralResultRefused; /tmp/stage3-front-3-16-mutants/non-brand-result-accepted.log
- ordinary-result-proof-skipped: caught by assertion; TestPhantomOverloadLiteralResultRefused; /tmp/stage3-front-3-16-mutants/ordinary-result-proof-skipped.log
- branded-literal-accepted: caught by assertion; TestPhantomOverloadBrandLiteralConstraintRefused; /tmp/stage3-front-3-16-mutants/branded-literal-accepted.log
- runtime-brand-result: caught by assertion; TestPhantomOverloadResultCastsAreErased; /tmp/stage3-front-3-16-mutants/runtime-brand-result.log
- overload-parameter-proof-skipped: caught by assertion; TestPhantomOverloadParameterProof; /tmp/stage3-front-3-16-mutants/overload-parameter-proof-skipped.log
- overload-value-accepted: SURVIVED; TestPhantomOverloadFunctionValueNotYet; /tmp/stage3-front-3-16-mutants/overload-value-accepted.log
- overload-header-kept: needle needs adaptation; no test executed; needle did not match integrated source
- any-brand-accepted: caught by assertion; TestPhantomOverloadAnyBrandRefused; /tmp/stage3-front-3-16-mutants/any-brand-accepted.log
- in-refusal-dropped: caught by assertion; TestPhantomArrayPresenceNarrowing; /tmp/stage3-front-3-16-mutants/in-refusal-dropped.log
- own-method-refusal-dropped: caught by assertion; TestPhantomArrayPresenceRefusals; /tmp/stage3-front-3-16-mutants/own-method-refusal-dropped.log
- object-reflection-refusal-dropped: caught by assertion; TestPhantomArrayPresenceRefusals; /tmp/stage3-front-3-16-mutants/object-reflection-refusal-dropped.log
- spread-refusal-dropped: caught by assertion; TestPhantomArrayPresenceRefusals; /tmp/stage3-front-3-16-mutants/spread-refusal-dropped.log
- assign-source-refusal-dropped: caught by assertion; TestPhantomArrayPresenceRefusals; /tmp/stage3-front-3-16-mutants/assign-source-refusal-dropped.log
- json-refusal-dropped: caught by assertion; TestPhantomArrayPresenceRefusals; /tmp/stage3-front-3-16-mutants/json-refusal-dropped.log
- write-refusal-dropped: caught by assertion; TestPhantomArrayPresenceRefusals; /tmp/stage3-front-3-16-mutants/write-refusal-dropped.log
- generic-constraint-forgotten: caught by assertion; TestPhantomArrayPresenceViews; /tmp/stage3-front-3-16-mutants/generic-constraint-forgotten.log
- union-arm-forgotten: caught by assertion; TestPhantomArrayPresenceNarrowing; /tmp/stage3-front-3-16-mutants/union-arm-forgotten.log
- member-name-forgotten: caught by assertion; TestPhantomArrayWritesNameTheMember; /tmp/stage3-front-3-16-mutants/member-name-forgotten.log
- required-undefined-rejected: needle needs adaptation; no test executed; needle did not match integrated source
- required-cast-not-erased: caught by assertion; TestPhantomArrayRequiredCastsAreErased; /tmp/stage3-front-3-16-mutants/required-cast-not-erased.log

The single-helper overloaded-value mutant survived the independent census guard; the adapted mutant removes both guards and is caught by TestPhantomOverloadFunctionValueNotYet. The adapted overload-header mutant is caught by TestPhantomOverloadResultCastsAreErased. Logs: /tmp/stage3-front-3-16-adapted-mutants/.

Rest-array-as-fixed-parameter and phantom-optional-data-confusion are caught by the unchanged Node-backed census_overload_contracts.a and phantom_array_brands.a fixtures, respectively. Logs: /tmp/stage3-front-3-16-fix-mutants/. Some redundant refusal mutants are diagnostic-only kills and do not establish unsafe runtime admission. Required-undefined rerun pending.

Item 16 required-undefined-rejected adapted rerun caught by TestPhantomSortedArrayProbe (valid source overlay, no build failure); log /tmp/stage3-front-3-16-required-undefined.log.

Item 16 additional core source-overlay mutants:

- primitive-data-accepted: caught by ^TestPhantomRefusals$; /tmp/stage3-front-3-16-core-mutants/primitive-data-accepted.log
- primitive-prototype-accepted: caught by ^TestPhantomPrimitiveNames$; /tmp/stage3-front-3-16-core-mutants/primitive-prototype-accepted.log
- array-prototype-accepted: caught by ^TestPhantomArrayNames$; /tmp/stage3-front-3-16-core-mutants/array-prototype-accepted.log
- primitive-cast-not-erased: caught by ^TestPhantomCastsAreErased$; /tmp/stage3-front-3-16-core-mutants/primitive-cast-not-erased.log
- array-element-proof-skipped: caught by ^TestPhantomArrayProofs$; /tmp/stage3-front-3-16-core-mutants/array-element-proof-skipped.log
- undefined-brand-check-erased: caught by ^TestPhantomUndefinedReview$; /tmp/stage3-front-3-16-core-mutants/undefined-brand-check-erased.log

Initial full gate also caught the incoming direct Call.Function read at phantom_overload_results.go:overloadedCall via TestCallTargetReaders. Fixed through the shared CallTargets API, retaining the original guard. The first gate log remains /tmp/stage3-front-3-16-first-gate.log. The corrected gate is pending at the time of this entry.

## Item 16 landed and item 17 excluded candidate

Item 16 passed its corrected uncached gate and was pushed at 58270ec655f020bc7e3d20e37469fc3b942af56b. Item 17 Array-hole focused tests passed lower .411s, flow .024s, oracle 43.816s; all ten accepted and 31 refusal sources are covered. All five requested mutants compiled and executed cleanly and were caught by Node stdout: forEach visiting holes (adapted one-loop anchor), map filling holes, join spelling holes undefined, maximum-length off by one, missing hole-count decrement. Raw logs /tmp/stage3-front-3-17-mutants/. Original forEach anchor had two matches and was excluded; rerun restricts to arrayVisit before arrayReduce.

Item 17 full six-package gate was not attempted after counts failed in 94.129s on 22 carried host fixtures: 17 optional-errno cast refusals and five namespace-readiness stops. Initial six static closure initializer clang failures were repaired with designated initializers, retaining count-first ABI, before focused checks and the counts retry. Candidate patch /tmp/stage3-front-3-17-skipped-resolution.patch. No candidate-only mutant is claimed as landed coverage. No unmodified wrong-output exit0 was seen.


### 18, excluded candidate only

No item18 compiler code was landed. The unmodified frozen-target capture probe is the exit-zero disagreement described in item18/REPORT.md, not a mutant. The full candidate gate is not green. Actual completed source controls, exact test filters, catchers and outcomes follow; raw logs are in item18/mutants/. These do not extend coverage of the landed compiler tree.

The original fake-error control's stale diagnostic also failed on the unmodified candidate, so that result is excluded. The corrected control survives the existing optional-field refusal. The stack-read control survives another stack guard. Original missing anchors, build failures, the wrong environment invocation, and interrupted controls are excluded. The restarted freshness baseline and full gate were interrupted at the real disagreement and are not claimed passing. Earlier first counts passed in184.935s and lowering rerun in269.558s before later candidate changes.

| Suite | Mutant | Outcome | Test | Exact catcher | Evidence |
| --- | --- | --- | --- | --- | --- |
| precision | callback-method-receiver-unbound | caught | TestCallbackTypedClassMethodCallBindsReceiver$ | must bind its receiver | item18/mutants/precision/callback-method-receiver-unbound.log |
| precision | loop-presence-unproved | caught | TestLoopPresenceSurvivesRuntimeMutation$ | loop condition proves cursor present | item18/mutants/precision/loop-presence-unproved.log |
| precision | unknown-store-ignored | caught | TestGeneratedGuardFactsIncludeUnknownStores$ | must remain throwing | item18/mutants/precision/unknown-store-ignored.log |
| precision | counter-body-write-ignored | caught | TestGeneratedGuardCounterWritesRemainThrowing$ | must remain throwing | item18/mutants/precision/counter-body-write-ignored.log |
| precision | presence-call-mutation-ignored | caught | TestNativeAgreesWithNode/internal/oracle/testdata/error_checks.a$ | exit codes differ | item18/mutants/precision/presence-call-mutation-ignored.log |
| precision | all-calls-pure | caught | TestNativeAgreesWithNode/internal/oracle/testdata/lent_reads.a$ | heap-use-after-free | item18/mutants/precision/all-calls-pure.log |
| precision | all-calls-throw | caught | TestMayThrowPrecision$ | cannot throw | item18/mutants/precision/all-calls-throw.log |
| precision | all-functions-throw | caught | TestMayThrowPrecision$ | cannot throw | item18/mutants/precision/all-functions-throw.log |
| precision | readiness-unproved | caught | TestMayThrowPrecision$ | cannot throw | item18/mutants/precision/readiness-unproved.log |
| library | capture-feature-false | caught | TestNativeAgreesWithNode$/internal/oracle/testdata/library_error_probe.a$ | stdout differs | item18/mutants/library/capture-feature-false.log |
| library | capture-throws | invalid or unexpected failure | TestNativeAgreesWithNode$/internal/oracle/testdata/library_error_debug_fail.a$ | stdout differs | item18/mutants/library/capture-throws.log |
| library | capture-stack-hook-removed | caught | TestLibraryErrorRefusals$/plain_target_stack$ | got <nil> | item18/mutants/library/capture-stack-hook-removed.log |
| library | capture-destructuring-refusal-removed | caught | TestLibraryErrorRefusals$/destructured_stack$ | got <nil> | item18/mutants/library/capture-destructuring-refusal-removed.log |
| library | capture-typeof-object | caught | TestNativeAgreesWithNode$/internal/oracle/testdata/library_error_typeof.a$ | stdout differs | item18/mutants/library/capture-typeof-object.log |
| library | capture-identity-fresh | caught | TestNativeAgreesWithNode$/internal/oracle/testdata/library_error_probe.a$ | stdout differs | item18/mutants/library/capture-identity-fresh.log |
| library | capture-intrinsic-cast-not-recognized | caught | TestNativeAgreesWithNode$/internal/oracle/testdata/library_error_original_probe.a$ | no-unchecked-cast | item18/mutants/library/capture-intrinsic-cast-not-recognized.log |
| library | accessor-targets-cleared | needs adaptation |  |  | anchor not applied |
| library | node-named-refusal-delayed | caught | TestNodeLibraryNamesUnimplementedMembers$/node:fs.readFile | want named NotYet | item18/mutants/library/node-named-refusal-delayed.log |
| adapted | capture-throws | caught | TestNativeAgreesWithNode$/internal/oracle/testdata/library_error_debug_fail.a$ | exit codes differ | item18/mutants/adapted/capture-throws.log |
| adapted | accessor-targets-cleared | caught | TestNativeAgreesWithNode$/internal/oracle/testdata/class_features_accessors.a$ | virtual call has no target set | item18/mutants/adapted/accessor-targets-cleared.log |
| runtime | fresh-error-initializer-proof-removed | caught | TestNativeAgreesWithNode/internal/oracle/testdata/57f2d04_with_frozen.a$ | a write to a potentially frozen object | item18/mutants/runtime/fresh-error-initializer-proof-removed.log |
| runtime | concat-bound-ignored | caught | TestNativeAgreesWithNode/internal/oracle/testdata/catchability-limits/d96d304_try_finally_concat.a$ | stdout differs | item18/mutants/runtime/concat-bound-ignored.log |
| runtime | unprotected-finally-guarded | caught | TestProtectedStringGuardPrecision$ | cannot throw on this bounded protected path | item18/mutants/runtime/unprotected-finally-guarded.log |
| runtime | unknown-object-bound-ignored | caught | TestStringLengthBoundsIncludeUnknownObjects$ | unknown object must retain the full string bound | item18/mutants/runtime/unknown-object-bound-ignored.log |
| runtime | repeat-constant-length-ignored | caught | TestRepeatRefusalChecksResultLength$ | oversized constant repeat must remain a named refusal | item18/mutants/runtime/repeat-constant-length-ignored.log |
| runtime | protected-catch-root-omitted | caught | TestRuntimeStackCatchFinally$ | stdout differs | item18/mutants/runtime/protected-catch-root-omitted.log |
| runtime | javascript-stack-error-identity-omitted | caught | TestRuntimeStackCleanup$ | stdout differs | item18/mutants/runtime/javascript-stack-error-identity-omitted.log |
| fix | closed-input-builtin-initializer-not-exempt | invalid or unexpected failure | TestClosedFrameInputRejectsMutation$ | closed literal input was not proven | item18/mutants/fix/closed-input-builtin-initializer-not-exempt.log |
| fix | closed-input-ordinary-write-admitted | caught | TestClosedFrameInputRejectsMutation$ | mutable graph accepted as closed | item18/mutants/fix/closed-input-ordinary-write-admitted.log |
| fix | code-point-list-omitted | caught | TestCodePointCatchabilityBoundary/9984394_lib_codepoint.a$ | want String.fromCodePoint catchability refusal | item18/mutants/fix/code-point-list-omitted.log |
| fix | interface-method-targets-omitted | caught | TestCodePointCatchabilityBoundary/9984394_lib_dispatch_codepoint.a$ | want String.fromCodePoint catchability refusal | item18/mutants/fix/interface-method-targets-omitted.log |
| closed-adapted | closed-input-builtin-initializer-not-exempt | caught | TestClosedFrameInputRejectsMutation$ | adamic/cycle-capable | item18/mutants/closed-adapted/closed-input-builtin-initializer-not-exempt.log |
| boundary | mayThrow-overwritten | caught | TestNativeAgreesWithNode/internal/oracle/testdata/error_checks.a$ | differs | item18/mutants/boundary/mayThrow-overwritten.log |
| boundary | derived-error-descriptors | caught | TestErrorBoundariesAreExplicit/derived_error_(hasOwn\|enumerable\|locale\|static_own)$ | want refusal | item18/mutants/boundary/derived-error-descriptors.log |
| boundary | ready-read-flow-edge | caught | TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/error_checks.a$ | the call ended | item18/mutants/boundary/ready-read-flow-edge.log |
| boundary | ready-write-flow-edge | caught | TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/error_checks.a$ | the call ended | item18/mutants/boundary/ready-write-flow-edge.log |
| boundary | inherited-constructor-field | caught | TestErrorBoundariesAreExplicit/inherited_constructor_value$ | want refusal | item18/mutants/boundary/inherited-constructor-field.log |
| boundary | nullable-generic-receiver | caught | TestErrorBoundariesAreExplicit/nullable_generic_receiver$ | want refusal | item18/mutants/boundary/nullable-generic-receiver.log |
| boundary | stack-read | SURVIVED | TestErrorBoundariesAreExplicit/stack_read$ | want refusal | item18/mutants/boundary/stack-read.log |
| boundary | stack-write | caught | TestErrorBoundariesAreExplicit/stack_write$ | want refusal | item18/mutants/boundary/stack-write.log |
| boundary | cause-cycle | caught | TestErrorBoundariesAreExplicit/cause_closes_a_cycle$ | want refusal | item18/mutants/boundary/cause-cycle.log |
| boundary | custom-toString | caught | TestErrorBoundariesAreExplicit/override$ | want refusal | item18/mutants/boundary/custom-toString.log |
| boundary | optional-generic-field | caught | TestErrorBoundariesAreExplicit/optional_generic_field$ | want refusal | item18/mutants/boundary/optional-generic-field.log |
| boundary | erased-cause | caught | TestErrorBoundariesAreExplicit/erased_cause$ | want refusal | item18/mutants/boundary/erased-cause.log |
| boundary | nonliteral-options | caught | TestErrorBoundariesAreExplicit/nonliteral_options$ | want refusal | item18/mutants/boundary/nonliteral-options.log |
| boundary | normalization-expansion | caught | TestErrorBoundariesAreExplicit/normalization_expansion$ | want refusal | item18/mutants/boundary/normalization-expansion.log |
| boundary | structural-view | caught | TestErrorBoundariesAreExplicit/(structural_view\|nested_structural_view)$ | want refusal | item18/mutants/boundary/structural-view.log |
| boundary | fake-error | excluded: stale diagnostic also fails without the mutant; adapted control survives | TestErrorBoundariesAreExplicit/fake_error$ | want refusal | item18/mutants/boundary/fake-error.log |
| boundary | spread-error | caught | TestErrorBoundariesAreExplicit/spread$ | want refusal | item18/mutants/boundary/spread-error.log |
| boundary | inherited-own-field | needs adaptation |  | && !isRandom && !l.errorPrototypeRead(node) { | anchor not applied |
| boundary | mutable-unknown-alias | caught | TestErrorBoundariesAreExplicit/mutable_unknown_alias$ | want refusal | item18/mutants/boundary/mutable-unknown-alias.log |
| boundary | unrecorded-initializer | caught | TestEveryWriteIsRecordedAndKnown$ | a write lowering didn't record | item18/mutants/boundary/unrecorded-initializer.log |
| boundary | mutable-cause | caught | TestErrorBoundariesAreExplicit/mutable_cause$ | want refusal | item18/mutants/boundary/mutable-cause.log |
| fresh-fix | phantom-freshness-receiver-effects-erased | caught | TestPhantomMemberFreshness$/receiver_effects_remain$ | phantom receiver writes | item18/mutants/fresh-fix/phantom-freshness-receiver-effects-erased.log |
| fresh-fix | phantom-freshness-retains-receiver | caught | TestPhantomMemberFreshness$/absent_member_holds_no_receiver$ | hold no receiver | item18/mutants/fresh-fix/phantom-freshness-retains-receiver.log |
| fresh-fix | phantom-freshness-node-unrecognized | caught | TestPhantomMemberFreshness$ | phantom receiver writes | item18/mutants/fresh-fix/phantom-freshness-node-unrecognized.log |
| fake-control | fake-error-nominal-guard-removed | SURVIVED | TestErrorBoundariesAreExplicit/fake_error$ | want refusal | item18/mutants/fake-control/fake-error-nominal-guard-removed.log |

The later inherited-method adapted invocation did not use the configured Go environment and printed Go: Unknown option: test; it is excluded. A corrected invocation was queued after the new freshness controls but did not complete before the stop. The flow fixture-classification control was also queued but not completed. Freshness controls that erased receiver effects, returned the receiver as the phantom member, and removed PhantomMember recognition were caught by the targeted fresh tests; no final full write-site baseline is claimed. Existing permanent reader/error mutants were part of the interrupted gate, so no new passing permanent-mutant claim is made for item18.

## Continued items 19 and 20

These controls run against the continued front-3 compiler through test-only Go overlays. No control is an unmodified compiler finding. Every listed control compiled and failed its intended assertion; raw logs are committed under item19 and item20.

| Item | Mutant | Catcher | Evidence |
| --- | --- | --- | --- |
| 19 | restore-whole-module-guard | before enum initialization | item19/restore-whole-module-guard.log |
| 19 | remove-direct-enum-proof | proven enum read retained a readiness check | item19/remove-direct-enum-proof.log |
| 19 | erase-enum-readiness | exit codes differ (UBSan null-object access instead of readiness panic) | item19/erase-enum-readiness.log |
| 20 optional | present-undefined-false | stdout differs | item20/optional/present-undefined-false.log |
| 20 optional | present-null-false | stdout differs | item20/optional/present-null-false.log |
| 20 optional | required-allowed | got <nil> | item20/optional/required-allowed.log |
| 20 optional | nullable-refused | Lower: | item20/optional/nullable-refused.log |
| 20 optional | nullable-object-allowed | got <nil> | item20/optional/nullable-object-allowed.log |
| 20 optional | mixed-absence-allowed | got <nil> | item20/optional/mixed-absence-allowed.log |
| 20 method | shadowed-undefined | got <nil>, want unbound-method | item20/method/shadowed-undefined.log |
| 20 method | stored | got <nil>, want unbound-method | item20/method/stored.log |
| 20 method | prototype-absent | stdout | item20/method/prototype-absent.log |
| 20 method | absent-present | stdout | item20/method/absent-present.log |
| 20 method | wide-name-omitted | stdout | item20/method/wide-name-omitted.log |

The added method_presence_wide_union.a probe makes wide-name-omitted fail on the current unsupported union ABI: Node prints wide union true, the mutant native build prints wide union false, both exit0. The original optional-boolean Wide method now has a supported slot, so it alone would not exercise this descriptor branch. The unmodified added probe passes Node/native/JavaScript/release/leak checks. The two memoize probes retain their original source Node output and cycle-capable refusal boundary.

## Continued item 22

All six debugger controls were rerun through Go source overlays with ADAMIC_GATE_UNCACHED=1. Raw logs and results are in item22/. The first attempt lacked the toolchain environment and is not evidence; the sourced-environment rerun below is the evidence.

| Mutant | Catcher | Result |
| --- | --- | --- |
| trap-artifact | debugger emitted a trap or breakpoint call | caught |
| trap-oracle | exit codes differ | caught |
| javascript-dropped | JavaScript kept 0 debugger statements | caught |
| lowering-dropped | JavaScript kept 0 debugger statements | caught |
| debugger-refused | refuses debugger | caught |
| flow-instruction | debugger changed the flow: 2 instructions | caught |

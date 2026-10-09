Added a Node-backed parameter-property behavior guard and a checker-contract positive-case floor for task #0k2fe8d.
Base: origin/main c0a7667baadaf161d6bb9066e0838b32774caa6c; branch: compiler/param-prop-guards.
Focused lowering tests pass after restoration; existing parameter_properties.a oracle checked uncached.
M03 prints 0 instead of Node's 7 in both backends; P1's empty answer fails the positive-case floor.
No production changes, new oracle fixtures, counts.md changes or whole-package test runs.

TestParameterPropertyValueMatchesNode uses the audit witness: class C has constructor(public value:number), constructed with 7; console.log prints its value. Source and generated JavaScript run through oracle/node.mjs with bounded Node processes. Native is built with ASan/UBSan. All successful runs require exit 0, empty stderr and matching stdout; source stdout is pinned to "7\n".

compiler/lower-agree was absent from origin's branch list when checked. A minimal local copy of the Node comparison used in internal/lower/class_static_guard_test.go was used, with the witness changed and native comparison retained. No unlanded helper branch merged.

Mutants are exact audit diffs from test-audit/internal-lower-parameter_properties 2e99cb59:
- M03 inserts an early nil,nil return in parameterPropertyStores. New behavior test fails with JavaScript stdout "0\n" and native stdout "0\n" versus source Node "7\n"; leaf 0.30 seconds. This is a changed-output catch, not a build error.
- Existing TestNativeAgreesWithNode/internal/oracle/testdata/parameter_properties.a also catches M03, 0.55 seconds: JavaScript starts with plain body 0 instead of 7 and later exits 70; sanitized native reports UBSan null string access and exit 1. Source Node finishes exit 0.
- P1 inserts a nil,nil early return in Lower. TestParameterPropertyCheckerContracts fails its new class-and-executable-statements floor, 0.10 seconds. Checker-negative subcases continue to pass; only the vacuous success case is changed.

Mutants were applied independently and reversed before restored checks. Production files are unchanged.

Commands source /workspace/adamic-tools/env.sh. Each go test uses -count=1 -v -timeout 90s and outer timeout 120:
- ./internal/lower -run '^TestParameterProperty(ValueMatchesNode|CheckerContracts)$' baseline.
- ./internal/lower -run '^TestParameterPropertyValueMatchesNode$' under M03.
- ADAMIC_GATE_UNCACHED=1 ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/parameter_properties\.a$' under M03 and restored.
- ./internal/lower -run '^TestParameterPropertyCheckerContracts$' under P1.
- ./internal/lower -run '^TestParameter(Property(ValueMatchesNode|CheckerContracts|CallbackReceiver)|PropertiesSoundness)$' restored.

Restored lower leaves: new behavior test 0.36 seconds; touched checker-contract test 0.20 seconds. Baseline new leaf 0.31 seconds, checker-contract 0.14 seconds. All are below 60 seconds, including the focused package command.

Setup: node 0.022s; Go 0.022s; submodules 0.073s; markdown dependencies 0.073s; clang 0.180s; Go build 35.934s; cache warm 36.173s; total 36.204s. nproc=5, cpu.max=400000 100000 (4 CPUs). Full setup and test outputs are saved here as logs.

Restored existing oracle passed uncached in 0.65 seconds (package 0.670 seconds), including source Node, JavaScript, sanitized native, release native and leaks.
Test commit: 5555f605e.
Lane checks passed: "lane checks 1.4 s: gofmt and tools on 2 Go files, t.Parallel on 1 test packages; vet 1 packages".

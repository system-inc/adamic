# Iteration acceptance conversion (#41bkfdw, wave 1)

Base: compiler/lower-agree at 96e5c1dc. The explicit unit base takes precedence over the generic main instruction. No production changes are delivered.

The 12 original lowerSource call sites expand to 31 source rows: 2 accept and 29 refuse. No IR-only rows. The direct lowering safeguard is an additional refusal test without a lowerSource call.

| Test name | Class | What changed | Why |
|---|---|---|---|
| TestIteratorGapsAreExplicit (7 sources) | refuse | Left unchanged | Each expects explicit iterator NotYet. |
| TestIteratorViewsCannotHideReturn | refuse | Left unchanged | Checks hidden-return diagnostic. |
| TestIteratorViewsCannotEraseReceivers | refuse | Left unchanged | Checks erased receiver convention diagnostic. |
| TestIteratorDestructuringDoesNotLieAboutExhaustion | refuse | Left unchanged | Checks Refused and default fix. |
| TestGeneratorsAreRefusedEvenWithoutYield (3 sources) | refuse | Left unchanged | Checks suspended-frame ownership refusal. |
| TestLiteralMethodCapturesCannotMakeCycles | refuse | Left unchanged | Checks captured-method cycle refusal. |
| TestLiteralMethodViewsDoNotLoseThis (2 sources) | accept | Both call lowersAndAgreesWithNode; method-signature source split into TestLiteralMethodSignatureViewsDoNotLoseThis | Receiver preservation is visible in printed value 1 and exit 0. Existing prints retained, no extra IR assertion. |
| TestIteratorDescriptorReasons (7 sources) | refuse | Left unchanged | Checks specific NotYet reasons. |
| TestDestructuredMethodsCannotLoadOwnSlots (5 sources) | refuse | Left unchanged | Expects Refused or NotYet before own-slot load. |
| TestGenericIteratorViewsPreserveNativeArguments | refuse | Left unchanged | Checks nominal ancestry refusal. |
| TestIteratorMapperIndexHasNumberRepresentation | refuse | Left unchanged | Checks index representation NotYet. |
| TestIteratorSymbolKeysAreNotStringKeys | refuse | Left unchanged | Checks symbol-key storage NotYet. |

TestRepresentedMethodReplacementIsNotYet directly exercises setProperty and remains a refusal test. No Node run is appropriate for any refusal row. No rows skipped and no new oracle fixture, so counts.md needs no refresh.

Brief guidance: count call sites separately from source rows. A name suggesting representation or preservation does not make a row IR-only: inspect its expected result. This file has only literal-method acceptance, despite its iteration name. Split acceptance loops into top-level tests for separate grain and mutant evidence. The helper needed no change and both sources already had observable output.

## Evidence

`run-mutants.py` independently plants each one-line mutant and restores the original lowering and tests in a finally block. The old control is the entire original TestLiteralMethodViewsDoNotLoseThis loop, so both original sources are exercised.

| Mutant | Old acceptance | Converted acceptance | Observation |
|---|---|---|---|
| unbound-method-call.diff: callClosure sets Property.Method false | PASS, exit 0 | TestLiteralMethodViewsDoNotLoseThis FAIL, exit 1 | Backend stdout empty versus source Node `1\n`. |
| wrong-method-result.diff: objectMethod replaces its body with return 2 | PASS, exit 0 | TestLiteralMethodSignatureViewsDoNotLoseThis FAIL, exit 1 | Backend stdout `2\n` versus source Node `1\n`. |
| erased-receiver.diff: object_method Receiver false | PASS, exit 0 | PASS, exit 0 | Survived. Not counted as proof; replaced by the binding mutant above. Metadata alone does not change this JavaScript observation. |

All temporary production edits restored. Final filtered run covers every top-level test in iteration_test.go: PASS, package 0.519s. Converted leaves: 0.25s and 0.26s (isolated acceptance run: 0.17s each). Existing refusal tests and subtests pass. No whole-package or full gate run. No native ownership, retain/release or region claims.

Setup passed using GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh. Timing lines: Go ready 0.093s; Node ready 0.113s; clang ready 0.580s; markdown dependencies ready 1.170s; submodules ready 17.748s; go build ready 238.421s; test binaries deferred 238.537s; cache warm 238.539s; done 238.573s. nproc=5, cgroup quota=4 CPUs. First gofmt invocation lacked the sourced PATH; rerun after sourcing passed. Setup included a cold shared compiler build; per-test load/lower and Node execution remain inside the measured leaves.

The checkout fetches only main by default. Explicit remote-tracking refspecs were needed for the named base and integration refs after the prescribed fetch did not populate them. The base helper matches landed main byte for byte. Delivery stays on the explicitly requested base, with only this unit's own commit.

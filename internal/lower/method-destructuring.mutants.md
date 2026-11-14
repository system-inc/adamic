# Method destructuring diagnostic mutants

All four mutants were run through Go overlays and restored by running without
an overlay. No mutant changes remain in the production files.

| Mutant | Mutated site | Red test and assertion |
| --- | --- | --- |
| Restore the old Fix, `call it on its receiver or wrap that call in an arrow; destructuring would lose this`. | `internal/lower/collections.go:597` | `TestMethodDestructuringFix` (variable and parameter, `.a` and `.ts`), `method_destructuring_test.go:24`; `TestEveryCaseKeepsTheSpecsPromise/classes/method_destructuring.a`, `internal/conformance/conformance_test.go:206`. Both reject the old words. |
| Also refuse a function-typed property by testing its call signatures beside SymbolFlagsMethod. | `internal/lower/collections.go:596` | `TestEveryCaseKeepsTheSpecsPromise/classes/method_destructuring_property_fix.a`, `internal/conformance/conformance_test.go:206`: the promised lowering instead refuses at `method_destructuring_property_fix.a:5:9`. |
| Clear the class-method-origin stop (`found = nil`), making erasedMethodField return nil. | `internal/lower/iteration_origin.go:102` | `TestThisReadingMethodInFunctionPropertyDoesNotLower` (declaration, parameter pattern across a call, and plain read across a call, each in `.a` and `.ts`), `method_destructuring_test.go:64` (current assertion line): `got <nil>`. |
| Skip the literalMethod arm instead of recording its stop. | `internal/lower/iteration_origin.go:90` | `TestThisReadingMethodInFunctionPropertyDoesNotLower/literal.a` and `/literal.ts`, `method_destructuring_test.go:64`: `this-reading method behind a function-typed property must stop, got <nil>`. |

The old-Fix logs are `mutant-old-unit.log` and `mutant-old-conformance.log`;
the other logs are `mutant-property.log` and `mutant-origin.log`, all under
`/workspace/scratch/c3kwkwk/`. Each command exited 1 for its asserted failure.
The unmutated pin command passed before and after these runs.

The class-method-origin mutant was rerun with both call-crossing cases. All six
subtests went red with `got <nil>`; see `call-mutant.log` in the same scratch
directory. Without the overlay, both new cases stopped before lowering, and the
full pin file and conformance package passed again.

The literal-method mutant used `return candidate.ForEachChild(visit)` in place
of its stop and went red in both extensions; see `literal-first-mutant.log`
in the same scratch directory. The unmutated literal-first probe stopped in
both extensions. The pin file and conformance package passed after restoration.

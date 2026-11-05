# Mutants for a library function read as a value (#4q1b2er head 2)

Each mutant is one exact line edit, run on a Linux box against the test it names, then put back from
a saved copy. The oracle rows run with `ADAMIC_GATE_UNCACHED=1`. The lower tests are in
internal/lower/library_function_values_test.go, and the oracle's is
internal/oracle/library_function_values_test.go.

| Mutant | Edit | Red |
|---|---|---|
| The allowlist dropped to its global | lower/library_function_values.go, `libraryFunctionMember`: `if node.Name().Text() == member.name && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, member.global) {` made `if l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, member.global) {` | TestLibraryFunctionValueReadsElsewhereKeepTheirStop, another Date member: Date.parse, read into a `(text: string) => number` slot that is exactly its own, is admitted, and would run Date.now's body. The other rows stay green: their slots aren't their members' own, or their globals aren't Date |
| The stable holding dropped | `libraryFunctionValueOf`: `if function.Source.Role == role {` made `if false && function.Source.Role == role {` | TestLibraryFunctionValuesAgreeWithNode, identity.a, native, sanitized and javascript: each read makes a function value of its own, so `first === second` prints false where Node prints true |
| The body made a constant | `libraryFunctionValueOf`: `call := member.call()` made `call := ir.Expression(ir.NumberConstant{Value: 1})` | TestLibraryFunctionValuesAgreeWithNode, all three, on every backend: timestamp.a and callback.a are no longer close to the clock, and identity.a's value no longer reads it. Positive and monotone stay true, which is why each fixture also compares the value with Date.now() called directly |

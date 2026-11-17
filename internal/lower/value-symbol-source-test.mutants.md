# Mutants for TestReferenceWalksResolveShorthandNames, the source test holding reference walks to valueSymbol (#3sdynvn)

One edit to entries_provenance.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| enumerationBindingWritten's write walk back on l.symbol | `if ast.IsIdentifier(part) && l.valueSymbol(part) == symbol {` | `if ast.IsIdentifier(part) && l.symbol(part) == symbol {` | TestReferenceWalksResolveShorthandNames, naming `entries_provenance.go:88 in enumerationBindingWritten`, a function symbolComparers doesn't list |

The test reads source, so any fails-open site step 1 moved to valueSymbol, put back, turns it red the same way, naming its function. TestShorthandValueSymbol's enumeration_binding row goes red under this mutant too, at run time; the source test is the one that names the function without a program reaching it.

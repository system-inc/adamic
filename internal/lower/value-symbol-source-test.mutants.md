# Mutants for TestReferenceWalksResolveShorthandNames, the source test holding reference walks to valueSymbol (#3sdynvn)

Each is one edit, run on a house box alone, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| enumerationBindingWritten's write walk back on l.symbol | `if ast.IsIdentifier(part) && l.valueSymbol(part) == symbol {` (entries_provenance.go) | `if ast.IsIdentifier(part) && l.symbol(part) == symbol {` | TestReferenceWalksResolveShorthandNames, naming `entries_provenance.go:88 in enumerationBindingWritten`, a function symbolComparers doesn't list |
| assignedSymbol's record back on l.symbol | `l.assignedSymbols[l.valueSymbol(n)] = true` (predicates.go) | `l.assignedSymbols[l.symbol(n)] = true` | TestReferenceWalksResolveShorthandNames, naming `predicates.go:892 in assignedSymbol`: a set recorded under a taken symbol is a key like a lookup's, and assignedSymbols isn't one of loweringTables. TestShorthandValueSymbol is printed beside it; no row of it is known to reach assignedSymbol, so its red isn't asked |
| a second comparison inside callsParameter, a listed function | `if ast.IsIdentifier(callee) && l.symbol(callee) == parameter {` (class_static.go) | the same line, then `_ = l.symbol(callee) == parameter` as the block's first statement | TestReferenceWalksResolveShorthandNames, naming callsParameter listed with 1 and making 2, class_static.go:640 and :641: a listed function is held to its count, so a comparison added to it has to be read and counted |

The test reads source, so any fails-open site step 1 moved to valueSymbol, put back, turns it red the same way, naming its function. TestShorthandValueSymbol's enumeration_binding row goes red under the first mutant too, at run time; the source test is the one that names the function without a program reaching it.

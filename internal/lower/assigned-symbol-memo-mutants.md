# The assigned symbol memo: its mutants (#pv935yw)

predicateArguments asks `assignedSymbol`, which collects every symbol an identifier in the program assigns to into
`lowering.assignedSymbols` the first time it's asked. Three mutants in predicates.go's `assignedSymbol`, each applied
alone at b65d7026 (one commit on main 5f0b2152) and run with
`go test ./internal/lower -run '^(TestAssignedSymbolAnswersAsAFreshLowering|TestPredicateCallbackContracts)$' -count=1 -v`
on a Linux cloud worker (go1.27.1, clang 20.1.8; evidence/compiler/assigned-symbol-memo 0e9e349a). Restored, both pass.

| Mutant | Red | Failing line |
|---|---|---|
| a. the walk reads `l.program.Files()[:1]` | TestAssignedSymbolAnswersAsAFreshLowering | `assignedThere alone: assigned=false, want true` |
| b. `if l.assignedSymbols != nil { return false }` at the top | TestAssignedSymbolAnswersAsAFreshLowering | `assignedThere from the memo: assigned=false, want true` |
| c. `return len(l.assignedSymbols) > 0` | TestAssignedSymbolAnswersAsAFreshLowering | `readHere alone: assigned=true, want false` |

TestPredicateCallbackContracts passed under all three, `reassigned` included: its program is one file with one
assignment, which each mutant still answers right.

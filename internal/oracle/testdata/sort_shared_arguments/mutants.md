# Mutants for the sort's shared argument array (#367nnz2)

The test is internal/oracle/sort_shared_arguments_test.go, `TestSortSharedArguments`. Each mutant is one
exact edit, run alone with the unmutated head beside it, then put back.

| Mutant | Edit | Red |
|---|---|---|
| M1: the pair written after the call | internal/javascript/javascript.go, the adamicSort helper: `{ values[0] = left; values[1] = right; return code(callback, values); }` made `{ const result = code(callback, values); values[0] = left; values[1] = right; return result; }` | Run on Loom's J-wxds489 (sealed, azure-f110-22, at 5465cb25, whose tree this head's files match). Unmutated, all six rows pass. Mutated, direct.a, captured.a, throwing.a, not_a_number.a and reentrant.a go red on the JavaScript backend; arguments_kept.ts stays green, since it never reaches a backend. Each comparison reads the pair before it, and the first reads `undefined, undefined`, so the order comes out wrong. The JavaScript backend printed `5,3,1,9` where Node prints `1,3,5,9`, `100 121 770 670` for `100 2 468 999`, `30,10,20,50,40` for `10,20,30,40,50`, `3,2,5,4,1` for `1,3,2,4,5`, `4,3,8,1,6` for `8,6,4,3,1`, and `3,4,2` for `2,3,4` |

A closure reading `values[0]` lazily instead of at its entry was the first mutant considered, and set
aside: swapping the closure prologue's entry copy for a later read is too invasive an edit for one line.
M1 breaks the same promise from the helper's side, a comparison that doesn't see its own pair.

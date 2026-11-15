# Mutants for the sort's shared argument array (#367nnz2)

The test is internal/oracle/sort_shared_arguments_test.go, `TestSortSharedArguments`. Planned, not run
yet: written on a machine that doesn't build Go, so "red" below is what the code says, not what a run
printed. Each mutant is one exact edit, run alone, then put back.

| Mutant | Edit | Red |
|---|---|---|
| M1: the pair written after the call | internal/javascript/javascript.go, the adamicSort helper: `{ values[0] = left; values[1] = right; return code(callback, values); }` made `{ const result = code(callback, values); values[0] = left; values[1] = right; return result; }` | direct.a and reentrant.a on the JavaScript backend: the first comparison reads `undefined, undefined` (NaN, so equal) and every later one reads the pair before it, so the order comes out wrong. A hand simulation on Node, the helper with closures shaped as the backend emits them, printed `5,3,1,9` for `1,3,5,9`, `121 770 670` for `2 468 999`, and `30,10,20,50,40` for `10,20,30,40,50`. Release and sanitized native stay green, since native never runs adamicSort |

A closure reading `values[0]` lazily instead of at its entry was the first mutant considered, and set
aside: swapping the closure prologue's entry copy for a later read is too invasive an edit for one line.
M1 breaks the same promise from the helper's side, a comparison that doesn't see its own pair.

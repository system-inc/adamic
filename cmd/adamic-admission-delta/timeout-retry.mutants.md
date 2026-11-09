# Mutants for runtime timeouts under load (#pajbaba)

Each mutant is one edit, run on a Linux box against cmd/adamic-admission-delta, then put back.

| Mutant | Edit | Red |
|---|---|---|
| (a) No retry alone | main.go, `observeAlone`: after `result.RetriedRuns++`, `still = first; result.StillTimedOutRuns++; return` in place of the second observe | PENDING |
| (b) Only the first timeout named | main.go, `timedOut`: `names = append(names, command.name)` followed by `return names` inside the loop | PENDING |
| (c) A repair's timeout dropped | main.go: `for _, command := range repairStill {` made `for _, command := range repairStill[:0] {` | PENDING |

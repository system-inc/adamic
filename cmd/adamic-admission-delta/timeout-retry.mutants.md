# Mutants for runtime timeouts under load (#pajbaba)

Each mutant is one edit, run on a Linux box against cmd/adamic-admission-delta, then put back.

| Mutant | Edit | Red |
|---|---|---|
| (a) No retry alone | main.go, `observeAlone`: after `result.RetriedRuns++`, `still = first; result.StillTimedOutRuns++; return` in place of the second observe | Pupa job 01M713ZF07GH97NBXFJ1B5CKPH, at ea859b19 (Nov 9, 2026): TestRuntimeTimeoutRetriedAloneIsNotCounted, "exit 3, void true, verdict void, timed out 1, retried 1, still 1": a load timeout voids the clean run |
| (b) Only the first timeout named | main.go, `timedOut`: `names = append(names, command.name)` followed by `return names` inside the loop | Pupa job 01M713ZF07GH97NBXFJ1B5CKPH, at ea859b19 (Nov 9, 2026): TestEveryTimeoutIsItsReason, the reason reads "corpus/input.a: javascript compile timeout" without native compile |
| (c) A repair's timeout dropped | main.go: `for _, command := range repairStill {` made `for _, command := range repairStill[:0] {` | Pupa job 01M713ZF07GH97NBXFJ1B5CKPH, at ea859b19 (Nov 9, 2026): TestRepairTimeoutVoidsNotFails, "exit 1, void false, verdict fail": the repair's timeout reads as a red, as on #ny0c3gb's proof |

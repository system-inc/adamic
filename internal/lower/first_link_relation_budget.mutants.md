# Mutants for a first-link pass past a spent relation budget (#nxb3xh1, Compiler's ruling (1))

Each is one edit, run on a house box through Loom, then put back.

| Mutant | File | Old text | New text | Red |
|---|---|---|---|---|
| 1. The budget read after the run again, not before it (Compiler's) | first_link_passes.go | `if l.firstLinkBudgetSpent(name, rewrites) {`<br>`return nil`<br>`}`<br>`err = run()` | `err = run()`<br>`if l.firstLinkBudgetSpent(name, rewrites) {`<br>`return nil`<br>`}` | TestFirstLinkPassNeverRunsPastASpentRelationBudget: the rerun is made past the budget planted after the first stop and hangs, and the pin's 10 s deadline fails it, naming the rerun |
| 2. One site hands firstLinkPass only a stop | lower.go | `if err := lowering.checkSlotWrites(); err != nil \|\| lowering.firstLinkTraps {` | `if err := lowering.checkSlotWrites(); err != nil {` | TestFirstLinkBudgetSpentAtTheBaselineAbandonsEveryPass: one row fewer than the passes lower.go hands firstLinkPass, checkSlotWrites' row and abandonment missing |

In the old and new text, each `<br>` is a line break, with the file's own indentation, and `\|` is `|`.

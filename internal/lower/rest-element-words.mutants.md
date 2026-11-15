# Rest element wording mutants

Mutant run and restored through a Go overlay. Remove valueSlotKeeps' call to
restElementHeldOtherwise and its NotYet branch, leaving the keeping rule intact.
The overlay also removes the now-unused strconv import so the mutant compiles.

| Mutant | File and line | Red test and assertion |
| --- | --- | --- |
| Remove the rest-element diagnostic branch. | `internal/lower/expression.go:2188` | `TestRestElementWords/r12.a`, `/r12.ts`, `/r13.a`, and `/r13.ts`, at `internal/lower/rest_element_words_test.go:25`. Each receives the old weak sentence instead of the complete rest-element NotYet. |
| Print the rest element plus `[]` instead of the rest parameter's array type. | `internal/lower/expression.go:2193` | `TestRestElementPrintedFixLowers/.a` and `/.ts`, at `internal/lower/rest_element_words_test.go:73`. Both reject the extracted fix with TS2370 and TS2322. |

The compiled mutant exited 1 for those four assertion failures. Its log is
`/workspace/scratch/2tnz278/mutant.log`. The first attempt only found the unused
import and did not count as the proof. Production source was never overlaid on
disk. The unmutated pin run passed, and the whole lower gate includes the pins.

After the wording correction, the removed-branch mutant was rerun in
`final-mutant.log`: all four message rows failed at line 25, and both printed-fix
rows failed at line 70 because the hint was absent. The array-spelling mutant
was run independently against the printed-fix check in `final-fix-mutant.log`.
Both compiled mutants exited 1 for these assertions. The pin file passed again
without either overlay. Gofmt and vet passed. The whole lower gate was not rerun
for the wording correction; its earlier run passed in 2211.755s.

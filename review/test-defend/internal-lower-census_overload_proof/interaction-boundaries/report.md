All three requested rows are defended. Keep them.
D1, D2 and D5 each have exactly one failing top-level row in completed full-package runs.
Evidence starts from origin/main 7b9d4272c28f59530ab13daa5c49067e47933b06; production source and tests are unchanged.

Read rows.json for the requested verdict objects, matrix.json for every passed/failed/skipped row, code-and-oracle.md for the checked behavior and assertions, and coverage-differences.json for exclusive blocks. The full census contains 276 rows, including 37 added since the audit and no vanished audit rows. Each unique matrix has 273 other passing rows and two skipped opt-ins. All runs finished below the 90-second binary budget. No narrowing, panic recovery or unknown completed-row outcome was needed.

| Mutant | Starting-main file:line | Aim | Rows failed |
|---|---|---|---|
| D1 | internal/lower/predicates.go:37 | Lose the actionable rule label while preserving the refusal and its reason. | TestCensusPredicateInteractionBoundary |
| D2 | internal/lower/predicates.go:442 | Treat exported callback owners with local direct callers as closed. | TestCensusPredicateInteractionEscapes |
| D3 | internal/lower/census_predicate_marker.go:60 | Admit a number-rest slot as erased, despite its observable parameter contract. | survivor |
| D4 | internal/lower/census_small.go:429 | Classify number-rest slots as erased in both marker admission and function assignment, avoiding the independent refusal that masked D3. | TestCensusPredicateMarkerKeepsProofBoundaries, TestNestedRestIsSupported |
| D5 | internal/lower/census_small.go:429 | Widen erasure to number-rest explicit callable types, while keeping actual function declarations subject to ordinary rest-call rules. | TestCensusPredicateMarkerKeepsProofBoundaries |

D3 is a survivor with proven changed behavior, not an equivalent candidate: the fourth marker source loses its predicate proof refusal but receives a different function-parameter relation Refused instead. The class-only assertion still passes. See marker.clean.log, marker.D3.log and marker.a.txt. D4 catches Marker plus TestNestedRestIsSupported; D5 confines the false erasure classification to explicit function types and leaves function declarations alone. D5 accepts the marker source, so Marker fails while the other enabled rows pass.

Brief issues and actual costs:
1. The mandatory 15 GB free threshold is impossible on the 8.8 GB /tmp mount. Earlier scratch directories were removed; /tmp became nearly empty and /workspace had 18 GB free. No disk-full failure occurred.
2. Coverage-exclusive blocks alone miss Boundary's strongest assertion. Its unique mutation changes a Fix string reached by other rows, whose assertions do not require the rule label. This is a semantic defense on shared lines.
3. One production admission fault can be masked by another valid refusal. D3 and its CLI witness show this directly. Marker only checks Refused, so its name's proof-boundary promise is broader than the diagnostic specificity it asserts. D5 nevertheless proves a unique guard on observable rest slots. All rows are defended; none is recommended for deletion or weakening.
4. Current main differs from the audit commit and adds 37 rows. The full current census was used; file:line references in this report and plans are against the new starting main.
5. Native rebuild cost is nested in tests. Each matrix used its own build cache; timings.json gives observed binary times and wall totals including vet and compilation. No separate per-product rebuild timing was inferred.
6. The two opt-in census rows require project inputs and remained skipped. Their unknown catches are excluded from package-uniqueness claims. The mixed-union subcase also skipped.
7. A preliminary evidence-directory command used the Node directory as its working directory and failed before npm ran. It was corrected before the successful npm install and clean baseline. This was an execution error, not a red code baseline.

Warm setup was skipped; nproc=5; npm reported 0.773 seconds. The baseline binary took 66.800 seconds. Every standalone diff passed apply-check and go vet, then compiled and ran the full package. No other package was tested and no source or test changes are being published. Five production mutants were used, with three honest attempts for Marker and one unique defense each for Boundary and Escapes.

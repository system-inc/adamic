# Mutants for the element read's checked boundary witnesses (#6wfvfhm)

Adamic ruled (b) on #6wfvfhm (Nov 6): an element read through a checked view is the same hatch as a field read through one, so its stop joins the one checked_boundary form, and V3's lie fixtures in internal/oracle/testdata/views_v3 are listed as its witnesses. Each mutant is one edit, run on a house box, then put back. The delta runs are the admission delta against main 2c3f3229 with the head's manifest.

| Mutant | Edit | Red |
|---|---|---|
| The pattern without element | negative.go, `checkedBoundaryStop`: `(?:field\|element) read failed` back to `field read failed` | TestNegativeCheckedBoundaryElementStop, both ruled element stops (index and handout) refused; TestNegativeCheckedBoundaryArrayDeclaration, every array row refused as an unsupported lie; TestNegativeWitnessListFieldStopsUnchanged, the list as it lands no longer reads |
| A lie fixture's entry removed | cloud/admission-corpus/negative-witnesses.json: index-lie.ts's entry deleted | The delta reads fail: index-lie.ts is newly accepted, agree false and witness false, while the other listed fixtures stay witnesses |
| The element stop without its expected clause | internal/native/runtime/view_unions_untagged.c and internal/javascript/view_arrays.go: `; expected <T>` dropped from the element stop, built into the head compiler the delta is handed | The delta reads fail: no listed fixture is a witness, since neither backend's stop is its entry's words any more. Words without the expected clause can't be listed in their place, since the pattern refuses them (TestNegativeCheckedBoundaryElementStop's no-expected row) |

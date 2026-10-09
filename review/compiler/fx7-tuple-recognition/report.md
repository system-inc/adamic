Tuple recognition for 138

JavaScript now marks emitted object-backed tuples with hidden WeakSet state. Checked object and union reads accept those tuples through the selected contract, whose fixed-tuple membership still checks marker, length, and field types. Ordinary array literals remain unmarked and fail the tuple view in both backends. No native runtime change.

The workspace and stash survived the outage. No mutant was active on recovery; all three mutants were run here with finally-based source restoration. The stash remains intact.

Fixtures p68 and the extended lower p69 print `first:1` in source Node and both backends. Integration's original p69 remains an equality control printing `true`. New leaves: p68 0.84s, p69 0.87s, ordinary-array control 0.85s, rejected-array view 0.38s.

Commands (each test writes its own log):
- `go test ./internal/lower -run '^TestTupleRecognition' -v -count=1 -timeout 90s`: PASS, 0.887s.
- Three sequential mutant runs with the relevant tuple leaf regex, `-v -count=1 -timeout 90s`: each FAIL after successful compilation. Rejecting tuple objects and omitting their marker produce empty JavaScript output versus Node `first:1`. Admitting arbitrary arrays produces JavaScript exit 0 and `object`, where the required checked view rejects with exit 70.
- `go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s`: PASS, 51.869s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 540s -args -update-counts`: PASS, 215.291s. Three new rows; no existing rows changed. Native IR counts are independent of the JavaScript mutants.
- `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/review/agree/fxspptb_oct9_views_p(68|69)' -v -count=1 -timeout 180s`: PASS, 1.811s.
- `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/lower/testdata/fx7_tuple_recognition/p69.a' -v -count=1 -timeout 180s`: PASS, 1.836s.
- Full internal/lower and committed lane checks: see their logs.

Added counts: p68 4/4/8/14/4/0; integration p69 3/3/8/13/3/0; extended p69 4/4/9/15/4/0. No previous rows moved.

Scope: tuple membership and subsequent reads for these shapes. The callable retention unit was already pushed at 56280072. No changes to arbitrary-array object compatibility or native tuple representation.

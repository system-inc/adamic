# Stored scalar field tags

Rebuilt `44619d37`. Number and boolean reads narrowed from ordinary stored unions now load their declared slot once and check its tag before unboxing. Strict equality and typeof observe the stored value. Existing view metadata remains attached by readObjectField. The taste refusal witness is now an admission test.

Commands after sourcing `/workspace/adamic-tools/env.sh`; logs are adjacent:

- `timeout 180 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(scout_boxed_scalar|taste_stage3_representations)' -count=1 -v -timeout 90s`: PASS, 1.433s. New fixture leaves 1.27s and 1.36s; existing taste witness 1.31s. Node, both backends, sanitizer, release and leak checks.
- `timeout 240 go test ./internal/lower -count=1 -timeout 180s` reached its aggregate deadline while other work ran. Active leaves at expiry were at most one second; no individual failure was reported before timeout. After the other commands completed, `timeout 600 go test ./internal/lower -count=1 -timeout 570s`: PASS, 119.249s.
- `timeout 150 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s`: PASS, 73.524s.
- `timeout 150 go test ./internal/lower -run '^TestTasteBoxedScalarFieldNowLowers$' -count=1 -v -timeout 90s`: PASS, new top-level leaf 0.08s.
- `timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 570s -args -update-counts`: PASS, 280.171s. Two new rows; no existing count changes.

Both source mutants in boxed-mutants.json ran `timeout 150 go test -overlay <name>.overlay.json ./internal/oracle -run <pattern> -count=1 -v -timeout 90s` and returned semantic test failure (exit 1). `boxed-check` replaces the failure condition with false: native completes with 1 while Node and JavaScript print changed1, and the expected inserted panic is absent. `boxed-observation` removes strict-equality observation handling: the positive source's stale equality read panics instead of observing its actual stored tag. Full sources use .go.txt names.

Not claimed: string mixed-field checks (next scout slice), callable-union checks, optional writes, delete, hidden bytes or whole-corpus correctness. Step 17 advances through checked scalar unboxing only.

Integration lane checks passed: `lane checks 3.1 s: gofmt and tools on 15 Go files, t.Parallel on 4 test packages; vet 4 packages`.

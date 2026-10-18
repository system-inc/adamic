# Nullable-string residual

Rebuilt the executable residual of `8cc2a49a` on main's existing tagged representation. Ordinary narrowed fields now check their stored tag, union typeof uses the stored null sentinel, and native coalescing converts a tested union to a string. The obsolete nullable-number/boolean refusal assertions were omitted because main admits those representations. Source probes moved into oracle testdata; no cohere code was copied.

Commands ran from the repository root after sourcing `/workspace/adamic-tools/env.sh`. Output is in the adjacent logs.

- `timeout 180 go test ./internal/oracle -run 'TestNullableStringFieldCheckMutant|TestNativeAgreesWithNode/internal/oracle/testdata/scout_(nullable_string|hir_optional|tsc_)' -count=1 -v -timeout 90s`: PASS, 8.499s. Five Node/JavaScript/native/release/sanitizer witnesses; longest new fixture leaf 7.52s. Check mutant test 6.96s.
- `timeout 180 go test ./internal/lower -count=1 -timeout 90s`: PASS, 83.505s.
- `timeout 150 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s`: PASS, 50.252s.
- `timeout 150 go test ./internal/lower -run '^TestNullableStringAdmission$' -count=1 -v -timeout 90s`: PASS, new top-level test 0.41s.
- `timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 570s -args -update-counts`: PASS, 223.300s. Five new fixture rows; no existing row moved. The longer package deadline permits the full counts collection; individual new fixture leaves remain below 60 seconds.

Mutants: dropping the IR field check completes with undefined rather than the checked panic and is caught by TestNullableStringFieldCheckMutant. The source overlays in nullable-string-mutants.json each ran `timeout 150 go test -overlay <name>.overlay.json <package> -run <pattern> -count=1 -v -timeout 90s`; all returned test failure rather than a compile failure. Null boxing as undefined and a forced native typeof null hint disagree with Node output. Restoring the lowering typeof hint refuses the nullish source probe. Refusing nullable-string admission fails TestNullableStringAdmission. Removing the equality observation exception panics on an ordinary tag observation in the positive fixture. Their complete sources use .go.txt names.

Not claimed: whole-corpus compiler correctness, hidden byte credit, never-array lowering, class fields, scalar/string mixed-field checks, or nullable-object boundary correctness. Those residuals or dependencies are recorded in commits.md. Step 17 advances here through nullable-string checks and backend agreement only.

Integration lane checks passed: `lane checks 11.5 s: gofmt and tools on 6 Go files, t.Parallel on 2 test packages; vet 2 packages`.

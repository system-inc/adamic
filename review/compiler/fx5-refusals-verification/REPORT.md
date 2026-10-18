Verified merged fd27bb0e; fixed six fixture interactions introduced by combining items 127 and 130.
Fixture fix commit: eecbf001 on compiler/fx5-refusals-verification.
Final checks: full lower PASS 37.201s; reader guard PASS 1.604s; counts PASS 80.828s; lane PASS 2.4s.
Initial full lower FAILED 87.507s because the literal guard refused six fixtures before the intended narrowing checks.
No compiler implementation, test Go files, or counts table changed; no new guards or mutants were added.

Changed only cat, forward, inferred, property, readonly, and some_negated fixtures. Each now allocates mixed union storage, then pops the unwanted member. The readonly fixture takes a readonly view after the pop. This lets the every fixtures reach their original exact-location refusal checks and lets the negated-some control still exercise actual union storage. All 15 array-narrowing leaves passed in 1.151s.

Commands ran from the repository root after sourcing /workspace/adamic-tools/env.sh, with output sent directly to the saved logs:
- timeout 420 go test ./internal/lower -count=1 -timeout 6m
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
- timeout 780 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 12m
- timeout 180 go test ./internal/lower -run '^TestArrayNarrowing' -count=1 -v -timeout 90s
- timeout 180 bash -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'

The counts check ran on the requested merged compiler and passed without updating counts.md. The subsequent repairs change only lower testdata sources; no registered oracle fixture or compiler code changed. The full lower package and reader guard were rerun after those repairs.

Final committed-fixture lane output:
lane checks 2.4 s: gofmt and tools on 3 Go files, t.Parallel on 1 test packages; vet 1 packages

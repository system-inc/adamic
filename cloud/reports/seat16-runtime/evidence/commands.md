# Mutant commands actually run

All commands source /workspace/adamic-tools/env.sh. Go overlays replace the named build input with the retained scratch mutant; they do not edit repository files. Source copies are compressed to avoid creating Go packages in report directories.

```sh
go test -overlay /workspace/scratch/seat16-runtime/mutants/clobbers.json -count=1 -timeout 30m -run '^TestFreshWriteProbesStayRefused$/clobber_link_loop.a$' -v ./internal/oracle > mutants/clobbers.log 2>&1
go test -overlay /workspace/scratch/seat16-runtime/mutants/escaped-before.json -count=1 -timeout 30m -run '^TestFreshWriteProbesStayRefused$/escaped_before.a$' -v ./internal/oracle > mutants/escaped-before.log 2>&1
go test -overlay /workspace/scratch/seat16-runtime/mutants/memory-count.json -count=1 -timeout 30m -run '^TestMemoryExampleCountsMatchDocumentation$/list.a$' -v ./internal/oracle > mutants/memory-count.log 2>&1
go test -overlay /workspace/scratch/seat16-runtime/mutants/tree-refusal.json -count=1 -timeout 30m -run '^TestMemoryExamplesRefused$/tree$' -v ./internal/oracle > mutants/tree-refusal.log 2>&1
go test -overlay /workspace/scratch/seat16-runtime/mutants/closure-refusal.json -count=1 -timeout 30m -run '^TestMemoryExamplesRefused$/closure$' -v ./internal/oracle > mutants/closure-refusal.log 2>&1
go test -overlay /workspace/scratch/seat16-runtime/mutants/uncaught-name.json -count=1 -timeout 30m -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/coverage_uncaught_empty_changed.a$' -v ./internal/oracle > mutants/uncaught-name.log 2>&1
go test -count=1 -timeout 30m -run '^(TestReleaseSharedValueAndUnsafeMutant|TestFreedValuesAreCaughtWithSlabs)$' -v ./internal/native > mutants/release-slab.log 2>&1
go test -overlay /workspace/scratch/seat16-runtime/mutants/map-clear-fix.json -count=1 -timeout 30m -run '^TestEveryMutationIsInItsRange$/programs/../oracle/testdata/release_fast_graph.a$' -v ./internal/flow > map-clear-fix.log 2>&1
```

The first six commands exit 1 for the concrete check failures in their logs. The built-in release/slab mutant test exits 0 only because each deliberate faulty use is caught and its control succeeds. The MapClear fix overlay exits 0; the unmodified missing-entry version failed in 13-packages.log before correction.

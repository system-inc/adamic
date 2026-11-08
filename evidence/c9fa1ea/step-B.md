# Steps A and B at c9fa1ea7129678df8e3e01875830808ec87c9b4a

- gofmt -l internal/lower/: exit 0, empty output.
- go vet ./internal/lower/: exit 0.
- go test -count=1 -run 'TestAGenericMethod|TestNodeWatch' -v ./internal/lower/: exit 0, every test and subtest passed, including o2.a/o2.ts and all NodeWatch tests.
- Mutant removes exactly ` || l.hostDeclaredSymbol(inside)`; rerun `go test -count=1 -run 'TestNodeWatch' -v ./internal/lower/`: exit 1.
- Red: TestNodeWatchInheritedOnMutant, TestNodeWatchTscShape, TestNodeWatchInheritedOnStaysNotYet/view (and its parent).
- Restoration: byte-for-byte comparison exit 0; whole worktree Git diff exit 0.

All raw output, the mutant patch, and exit files are adjacent. Steps A/B were reported immediately before starting C.

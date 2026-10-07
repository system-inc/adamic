# Landing the passing ports

Registered: `@typescript-eslint/no-dupe-class-members` and `react/no-invalid-html-attribute`. Four other owned rules are retained outside registry discovery, with exact blockers and reproducers in `parked/README.md`. The parked fragment's repair tag and unescaped-entity suggestion bridge are retained for restoration. Shared files were not edited.

Base: origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. Go cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.

Commands and observations:

- `go run ./cmd/lint-registry`: exit 0; log `evidence/landing-passing/wave109-landing-registry.log`.
- `gofmt -w` on the edited owned adapters and `go vet ./stage1/cohere/lint`: exit 0; vet log is empty.
- Full uncached `go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=20m`: TestRulesAgree passed (52.34s), TestOwnedWitnesses passed (21.35s); 41 mutants passed and one survived. Its witness did not exercise the mutant's unknown-token branch.
- Added `<a rel="not-a-relation" />` to the owned invalid-attribute witness, then uncached `go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout=10m`: PASS, 69.463s. The final owned comparison is 98,446 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native.
- Uncached `go test ./stage1/cohere/lint -run '^TestMutants/react-no-invalid-html-attribute_listener_suppressed$' -count=1 -v -timeout=10m`: PASS, 20.095s. The repaired witness catches the mutant on source Node, emitted JavaScript and sanitized native. All 42 registered mutants are therefore caught across the full run and focused rerun; the other 41 were unchanged.

The owned duplicate-member mutant suppresses duplicate findings; the invalid-attribute mutant suppresses unknown relation findings. Both are caught only by comparison with unchanged Go, after clean compilation and execution.

The shared upstream gate retains its explicit malformed-parser-recovery refusals; this report does not claim parser recovery. No full repository gate or additional performance run was executed in this landing pass. Earlier profiles remain historical, not new timing measurements.

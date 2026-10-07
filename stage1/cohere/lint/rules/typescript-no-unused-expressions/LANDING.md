# Landing with parser-blocked rule parked

Base: origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. Five retained owned rules remain registered. no-async-promise-executor is excluded; claims/wave1-03-parked.md names its reproducer and preserved implementation SHA. No shared harness file changed.

`go run ./cmd/lint-registry` and `go vet ./stage1/cohere/lint` pass. `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=30m` passes in 945.628s. TestRulesAgree PASS 53.73s; all 45 registered mutants pass; TestOwnedWitnesses PASS 21.89s, 101,062 identical bytes on Go, source Node, emitted JavaScript and sanitized native. Raw output is under evidence/landing/isolated.

Full repository gate, 17 required external-input checks, full compiler source corpus and throughput are not run.

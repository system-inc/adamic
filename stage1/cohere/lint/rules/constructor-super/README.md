# Constructor super

This listener ports `core.ConstructorSuper` from the pinned Go cohere rule. The registry hands it a constructor node. The rule follows Go's structural never/sometimes/always lattice, acceptable exits, superclass whitelist, and separate may-call duplicate walk. It uses the delivered core `isSeparateEvaluationContext` helper with an identity-preserving kind projection. No checker facts or control-flow graph are needed.

The witness reports missingAll, missingSome, duplicate and badSuper. `constructor-super-conditional-call-guaranteed` changes the join of unlike branch states from sometimes to always, suppressing the conditional-call finding. `testdata/native_mutant.py` uses a scratch Go overlay to select this rule as the native mutant canary and compare all 120 distinct captured upstream cases plus its witness against unchanged Go on source Node, emitted JavaScript and sanitized native. It does not edit the shared harness.

The other unfinished wave1-07 claim, `consistent-return`, stays out. Its 75 distinct Go source/file/options cases require `control_flow_graph.Build(node, control_flow_graph.Hooks[struct{}]{})`, followed by `graph.EndReachable` in `cohere/internal/lint/ecmascript/consistentreturn/judgment.go:355`. For example, `function f(x) { if (x) return 1; }` needs a reachable function-end judgment. The pending CFG cannot be replaced by a private structural approximation.

Validation on wave2 `95968dd93ad0876245f181af63931c3134e14b3d`, merged with area `334509eea8a49b8085187e206c495cc6aa24c5c4`, and Go cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`:

- All 120 upstream cases and the witness matched Go on source Node, emitted JavaScript and ASan/UBSan native: 33,233 output bytes. The join mutant was caught by output comparison on Node and emitted JavaScript, and sanitized native matched that changed Node output byte for byte, proving the same disagreement with Go.
- The complete lint package passed in 1496.767 seconds: 147 passing tests/subtests, zero failures, one named skip. Top-level counts are 41 pass, zero fail, one skip. `TestRulesAgree`, `TestOwnedWitnesses`, `TestCompilerAndStage1Agree`, `TestProfileArtifacts`, `TestProfileCompilation` and `TestProfileSnapshotsAgree` all passed.
- The inherited skip is `TestCheckerBridgeRefusalPending`, at `checker_pending_test.go:51`: `tsgoInspect` must return `TSGoError` from the C error buffer, and the base prelude does not declare that type. This is not an omitted input. No shared test or guard was changed.
- `gofmt -l` was empty for the own oracle and all Go files brought in by the area merge. `git diff --check` was empty.

The full package command was `go test -json -count=1 -timeout=90m ./stage1/cohere/lint`, with `ADAMIC_GATE_UNCACHED=1`, `ADAMIC_TYPESCRIPT_SOURCE` pointing at Microsoft TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`, `ADAMIC_LINT_BENCH=1`, and both `ADAMIC_LINT_PROFILE_DIR` and `ADAMIC_LINT_PROFILE_SNAPSHOTS` pointing at the same fresh scratch directory. `GOPROXY=https://proxy.golang.org|direct` was set.

Logs are in `testdata/evidence/`; `lint-package.jsonl.gz` preserves the complete package run. The first two parity findings are preserved too: the repair category had to be empty, and binary operators had to be read from their token child. Both fixes changed the port, with every expectation and comparison retained.
